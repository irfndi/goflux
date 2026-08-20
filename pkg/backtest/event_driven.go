package backtest

import (
	"errors"
	"sort"
	"time"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/metrics"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/trading"
)

// EventType defines the type of event.
type EventType string

const (
	// EventBar represents a bar (candle) event.
	EventBar EventType = "bar"
)

// Event represents a market event in chronological order.
type Event struct {
	Type      EventType
	Timestamp time.Time
	Symbol    string
	Data      any
}

// BarEventData contains the candle data for a bar event.
type BarEventData struct {
	Candle *series.Candle
}

// FillPriceSource determines which price to use for market order fills.
type FillPriceSource int

const (
	// FillAtOpen fills market orders at the bar's open price.
	FillAtOpen FillPriceSource = iota
	// FillAtClose fills market orders at the bar's close price.
	FillAtClose
)

// CommissionModel computes the commission for an order fill.
type CommissionModel func(order *trading.Order, fillPrice, fillAmount decimal.Decimal) decimal.Decimal

// SlippageModel computes the slippage for an order fill.
// Positive slippage worsens the fill (higher for buy, lower for sell).
type SlippageModel func(order *trading.Order, candle *series.Candle) decimal.Decimal

// PartialFillModel determines how much of an order fills.
// Returns the filled amount (must be <= order.Amount).
type PartialFillModel func(order *trading.Order, candle *series.Candle) decimal.Decimal

// NoCommission returns zero commission.
func NoCommission(order *trading.Order, fillPrice, fillAmount decimal.Decimal) decimal.Decimal {
	return decimal.ZERO
}

// FixedCommission returns a fixed commission per order fill.
func FixedCommission(amount decimal.Decimal) CommissionModel {
	return func(order *trading.Order, fillPrice, fillAmount decimal.Decimal) decimal.Decimal {
		return amount
	}
}

// PercentCommission returns commission as a percentage of fill value.
func PercentCommission(pct float64) CommissionModel {
	return func(order *trading.Order, fillPrice, fillAmount decimal.Decimal) decimal.Decimal {
		return fillPrice.Mul(fillAmount).Mul(decimal.New(pct))
	}
}

// NoSlippage returns zero slippage.
func NoSlippage(order *trading.Order, candle *series.Candle) decimal.Decimal {
	return decimal.ZERO
}

// FixedSlippage returns a fixed slippage amount.
func FixedSlippage(amount decimal.Decimal) SlippageModel {
	return func(order *trading.Order, candle *series.Candle) decimal.Decimal {
		return amount
	}
}

// FullFill always fills the entire order.
func FullFill(order *trading.Order, candle *series.Candle) decimal.Decimal {
	return order.Amount
}

// HalfFill fills 50% of the order (useful for testing partial fills).
func HalfFill(order *trading.Order, candle *series.Candle) decimal.Decimal {
	return order.Amount.Div(decimal.New(2))
}

// effectiveQty returns the filled amount if available, otherwise the order amount.
func effectiveQty(order *trading.Order) decimal.Decimal {
	if !order.FilledAmount.IsZero() {
		return order.FilledAmount
	}
	return order.Amount
}

// brokerPosition wraps a trading.Position with bar index tracking.
type brokerPosition struct {
	pos        *trading.Position
	entryIndex int
	remaining  decimal.Decimal
}

// brokerTrade tracks a closed position with entry/exit indices.
type brokerTrade struct {
	pos        *trading.Position
	entryIndex int
	exitIndex  int
}

// SimulatedBroker simulates order execution for a single asset.
type SimulatedBroker struct {
	Symbol           string
	InitialCapital   decimal.Decimal
	Equity           decimal.Decimal
	CommissionModel  CommissionModel
	SlippageModel    SlippageModel
	FillPriceSource  FillPriceSource
	PartialFillModel PartialFillModel
	AllowLong        bool
	AllowShort       bool

	pendingOrders []*trading.Order
	openPositions []*brokerPosition
	closedTrades  []brokerTrade
	record        *trading.TradingRecord
	equityHistory []decimal.Decimal
	currentIndex  int
	lastCandle    *series.Candle
	lastExecution *trading.Order
}

// NewSimulatedBroker creates a new simulated broker.
func NewSimulatedBroker(symbol string, initialCapital decimal.Decimal) *SimulatedBroker {
	return &SimulatedBroker{
		Symbol:           symbol,
		InitialCapital:   initialCapital,
		Equity:           initialCapital,
		CommissionModel:  NoCommission,
		SlippageModel:    NoSlippage,
		FillPriceSource:  FillAtClose,
		PartialFillModel: FullFill,
		AllowLong:        true,
		AllowShort:       false,
		record:           trading.NewTradingRecord(),
		equityHistory:    make([]decimal.Decimal, 0),
	}
}

// SubmitOrder submits an order to the broker. It becomes pending and is
// evaluated against subsequent bars.
func (b *SimulatedBroker) SubmitOrder(order *trading.Order) {
	if b == nil || order == nil {
		return
	}
	order.Status = trading.OrderStatusPending
	b.pendingOrders = append(b.pendingOrders, order)
}

// ProcessBar processes all pending orders against the given candle and
// records marked-to-market equity for the bar.
func (b *SimulatedBroker) ProcessBar(index int, candle *series.Candle) {
	if b == nil {
		return
	}
	b.currentIndex = index
	b.lastCandle = candle
	b.equityHistory = append(b.equityHistory, b.Equity)
	if candle == nil {
		return
	}

	var remaining []*trading.Order
	for _, order := range b.pendingOrders {
		filled := b.tryFill(order, candle)
		if filled {
			execution := b.lastExecution
			if execution == nil {
				execution = order
			}
			b.handleFilledOrder(execution, index)
		}
		if order.Status != trading.OrderStatusFilled {
			remaining = append(remaining, order)
		}
	}
	b.pendingOrders = remaining
	if len(b.equityHistory) > 0 {
		b.equityHistory[len(b.equityHistory)-1] = b.markedEquity(candle)
	}
}

// ProcessStrategySignal handles immediate market orders from strategy signals.
// Market orders fill at the configured FillPriceSource within the current bar.
// When both AllowLong and AllowShort are true, short entries take priority
// because the Strategy interface does not specify direction.
func (b *SimulatedBroker) ProcessStrategySignal(shouldEnter, shouldExit bool, index int, candle *series.Candle) {
	if b == nil || candle == nil {
		return
	}
	if shouldEnter && b.canEnterShort() {
		order := trading.NewOrderDetail(trading.SELL, trading.MarketOrder, b.Symbol, decimal.ONE)
		order.CreationTime = time.Unix(int64(index), 0)
		if b.fillMarketOrder(order, candle) {
			b.enterPosition(order, index)
		}
		return
	}

	if shouldEnter && b.canEnterLong() {
		order := trading.NewOrderDetail(trading.BUY, trading.MarketOrder, b.Symbol, decimal.ONE)
		order.CreationTime = time.Unix(int64(index), 0)
		if b.fillMarketOrder(order, candle) {
			b.enterPosition(order, index)
		}
		return
	}

	if shouldExit && b.hasOpenPosition() {
		b.closeAllPositions(index, candle)
	}
}

func (b *SimulatedBroker) canEnterLong() bool {
	return b.AllowLong && !b.hasOpenPosition()
}

func (b *SimulatedBroker) canEnterShort() bool {
	return b.AllowShort && !b.hasOpenPosition()
}

func (b *SimulatedBroker) hasOpenPosition() bool {
	return len(b.openPositions) > 0
}

func (b *SimulatedBroker) markedEquity(candle *series.Candle) decimal.Decimal {
	if candle == nil {
		return b.Equity
	}

	marked := b.Equity
	for _, bp := range b.openPositions {
		if bp == nil || bp.pos == nil || !bp.remaining.IsPositive() {
			continue
		}
		entryOrder := bp.pos.EntranceOrder()
		entryPrice := entryOrder.FilledPrice
		if entryPrice.IsZero() {
			entryPrice = entryOrder.Price
		}
		if bp.pos.IsLong() {
			marked = marked.Add(candle.ClosePrice.Sub(entryPrice).Mul(bp.remaining))
		} else if bp.pos.IsShort() {
			marked = marked.Add(entryPrice.Sub(candle.ClosePrice).Mul(bp.remaining))
		}
	}
	return marked
}

func (b *SimulatedBroker) tryFill(order *trading.Order, candle *series.Candle) bool {
	switch order.Type {
	case trading.MarketOrder:
		return b.fillMarketOrder(order, candle)
	case trading.LimitOrder:
		return b.fillLimitOrder(order, candle)
	case trading.StopOrder:
		return b.fillStopOrder(order, candle)
	default:
		return false
	}
}

func (b *SimulatedBroker) fillMarketOrder(order *trading.Order, candle *series.Candle) bool {
	if candle == nil {
		return false
	}
	var price decimal.Decimal
	switch b.FillPriceSource {
	case FillAtOpen:
		price = candle.OpenPrice
	default:
		price = candle.ClosePrice
	}
	return b.executeFill(order, price, candle)
}

func (b *SimulatedBroker) fillLimitOrder(order *trading.Order, candle *series.Candle) bool {
	if order == nil || candle == nil {
		return false
	}
	if order.Side == trading.BUY && candle.MinPrice.LTE(order.Price) {
		return b.executeFill(order, order.Price, candle)
	}
	if order.Side == trading.SELL && candle.MaxPrice.GTE(order.Price) {
		return b.executeFill(order, order.Price, candle)
	}
	return false
}

func (b *SimulatedBroker) fillStopOrder(order *trading.Order, candle *series.Candle) bool {
	if order == nil || candle == nil {
		return false
	}
	if order.Side == trading.BUY && candle.MaxPrice.GTE(order.StopPrice) {
		return b.executeFill(order, order.StopPrice, candle)
	}
	if order.Side == trading.SELL && candle.MinPrice.LTE(order.StopPrice) {
		return b.executeFill(order, order.StopPrice, candle)
	}
	return false
}

func (b *SimulatedBroker) executeFill(order *trading.Order, price decimal.Decimal, candle *series.Candle) bool {
	if order == nil || candle == nil {
		return false
	}
	remaining := order.Amount.Sub(order.FilledAmount)
	if !remaining.IsPositive() {
		return false
	}
	var slippage decimal.Decimal
	if order.Type != trading.LimitOrder && b.SlippageModel != nil {
		slippage = b.SlippageModel(order, candle)
	}
	var fillPrice decimal.Decimal
	if order.Side == trading.BUY {
		fillPrice = price.Add(slippage)
	} else {
		fillPrice = price.Sub(slippage)
	}

	fillRequest := *order
	fillRequest.Amount = remaining
	fillRequest.FilledAmount = decimal.ZERO
	fillModel := b.PartialFillModel
	if fillModel == nil {
		fillModel = FullFill
	}
	fillAmount := fillModel(&fillRequest, candle)
	if fillAmount.GT(remaining) {
		fillAmount = remaining
	}
	if fillAmount.IsZero() {
		return false
	}

	commission := decimal.ZERO
	if b.CommissionModel != nil {
		commission = b.CommissionModel(order, fillPrice, fillAmount)
	}

	order.Fill(fillPrice, fillAmount)
	execution := *order
	execution.Amount = fillAmount
	execution.FilledAmount = fillAmount
	execution.FilledPrice = fillPrice
	execution.Status = trading.OrderStatusFilled
	b.lastExecution = &execution
	b.Equity = b.Equity.Sub(commission)

	return true
}

func (b *SimulatedBroker) handleFilledOrder(order *trading.Order, index int) {
	if b.record.CurrentPosition().IsOpen() {
		if len(b.openPositions) > 0 {
			bp := b.openPositions[0]
			entrySide := bp.pos.EntranceOrder().Side
			if (entrySide == trading.BUY && order.Side == trading.SELL) ||
				(entrySide == trading.SELL && order.Side == trading.BUY) {
				b.exitPosition(bp, order, index)
			} else if order.Side == entrySide {
				// A partial entry may fill over multiple bars. Consolidate the
				// execution into the open position instead of dropping it.
				quantity := effectiveQty(order)
				entry := bp.pos.EntranceOrder()
				entry.Amount = entry.Amount.Add(quantity)
				entry.FilledAmount = entry.FilledAmount.Add(quantity)
				bp.remaining = bp.remaining.Add(quantity)
			}
		}
		return
	}
	if b.record.CurrentPosition().IsNew() {
		if (order.Side == trading.BUY && b.AllowLong) || (order.Side == trading.SELL && b.AllowShort) {
			b.enterPosition(order, index)
		}
	}
}

func (b *SimulatedBroker) enterPosition(order *trading.Order, index int) {
	if order == nil || order.FilledAmount.IsZero() {
		return
	}
	order.ExecutionTime = time.Unix(int64(index), 0)
	pos := trading.NewPosition(*order)
	b.openPositions = append(b.openPositions, &brokerPosition{
		pos: pos, entryIndex: index, remaining: effectiveQty(order),
	})

	b.record.Operate(*order)
}

func (b *SimulatedBroker) exitPosition(bp *brokerPosition, order *trading.Order, index int) {
	if bp == nil || bp.pos == nil || order == nil {
		return
	}

	entryOrder := bp.pos.EntranceOrder()
	qty := effectiveQty(order)
	if qty.GT(bp.remaining) {
		qty = bp.remaining
	}
	if qty.IsZero() {
		return
	}
	var profit decimal.Decimal
	if bp.pos.IsLong() {
		profit = order.FilledPrice.Sub(entryOrder.FilledPrice).Mul(qty)
	} else {
		profit = entryOrder.FilledPrice.Sub(order.FilledPrice).Mul(qty)
	}
	b.Equity = b.Equity.Add(profit)

	bp.remaining = bp.remaining.Sub(qty)
	partial := bp.remaining.IsPositive()
	tradeEntry := *entryOrder
	tradeEntry.Amount = qty
	tradeEntry.FilledAmount = qty
	tradeExit := *order
	tradeExit.Amount = qty
	tradeExit.FilledAmount = qty
	tradePos := trading.NewPosition(tradeEntry)
	tradePos.Exit(tradeExit)

	b.closedTrades = append(b.closedTrades, brokerTrade{
		pos:        tradePos,
		entryIndex: bp.entryIndex,
		exitIndex:  index,
	})

	order.ExecutionTime = time.Unix(int64(index), 0)
	if !partial {
		bp.pos.Exit(*order)
		for i, p := range b.openPositions {
			if p == bp {
				b.openPositions = append(b.openPositions[:i], b.openPositions[i+1:]...)
				break
			}
		}
		b.record.Operate(*order)
	}
}

func (b *SimulatedBroker) closeAllPositions(index int, candle *series.Candle) {
	for len(b.openPositions) > 0 {
		bp := b.openPositions[0]
		qty := bp.remaining
		exitSide := trading.SELL
		if bp.pos.IsShort() {
			exitSide = trading.BUY
		}
		order := trading.NewOrderDetail(exitSide, trading.MarketOrder, b.Symbol, qty)
		order.CreationTime = time.Unix(int64(index), 0)
		if !b.fillMarketOrderFull(order, candle) {
			// Prevent infinite loop if market order cannot fill.
			break
		}
		b.exitPosition(bp, order, index)
	}
}

// BacktestResult converts broker state to a BacktestResult.
func (b *SimulatedBroker) BacktestResult() BacktestResult {
	if b == nil {
		return BacktestResult{}
	}
	// Close any remaining open positions at the last known equity price.
	// Use the last equity value as the exit price basis.
	b.finalizeOpenPositions()
	if len(b.equityHistory) > 0 {
		b.equityHistory[len(b.equityHistory)-1] = b.Equity
	}

	trades := make([]Trade, 0, len(b.closedTrades))
	for _, bt := range b.closedTrades {
		trades = append(trades, b.brokerTradeToTrade(bt))
	}

	result := b.calculateResults(trades, b.equityHistory, b.InitialCapital, b.Equity)
	return result
}

func (b *SimulatedBroker) finalizeOpenPositions() {
	if len(b.openPositions) == 0 {
		return
	}
	if b.lastCandle == nil {
		b.openPositions = nil
		return
	}
	// Copy slice to avoid mutation during iteration by exitPosition.
	positions := make([]*brokerPosition, len(b.openPositions))
	copy(positions, b.openPositions)
	for _, bp := range positions {
		qty := bp.remaining
		exitSide := trading.SELL
		if bp.pos.IsShort() {
			exitSide = trading.BUY
		}
		order := trading.NewOrderDetail(exitSide, trading.MarketOrder, b.Symbol, qty)
		order.CreationTime = time.Unix(int64(b.currentIndex), 0)
		if !b.fillMarketOrderFull(order, b.lastCandle) {
			continue
		}
		b.exitPosition(bp, order, b.currentIndex)
	}
}

func (b *SimulatedBroker) fillMarketOrderFull(order *trading.Order, candle *series.Candle) bool {
	partialFillModel := b.PartialFillModel
	b.PartialFillModel = FullFill
	filled := b.fillMarketOrder(order, candle)
	b.PartialFillModel = partialFillModel
	return filled
}

func (b *SimulatedBroker) brokerTradeToTrade(bt brokerTrade) Trade {
	entry := bt.pos.EntranceOrder()
	exit := bt.pos.ExitOrder()

	qty := effectiveQty(entry)
	var profit decimal.Decimal
	if bt.pos.IsLong() {
		profit = exit.FilledPrice.Sub(entry.FilledPrice).Mul(qty)
	} else {
		profit = entry.FilledPrice.Sub(exit.FilledPrice).Mul(qty)
	}

	entryPrice := entry.FilledPrice
	if entryPrice.IsZero() {
		entryPrice = entry.Price
	}
	exitPrice := exit.FilledPrice
	if exitPrice.IsZero() {
		exitPrice = exit.Price
	}

	var profitPct decimal.Decimal
	cost := entryPrice.Mul(qty)
	if !cost.IsZero() {
		profitPct = profit.Div(cost)
	}

	return Trade{
		EntryTime:     bt.entryIndex,
		EntryPrice:    entryPrice,
		ExitTime:      bt.exitIndex,
		ExitPrice:     exitPrice,
		Direction:     directionFromPosition(bt.pos),
		Quantity:      qty,
		Profit:        profit,
		ProfitPercent: profitPct,
		Duration:      bt.exitIndex - bt.entryIndex,
	}
}

func directionFromPosition(pos *trading.Position) string {
	if pos.IsLong() {
		return "long"
	}
	if pos.IsShort() {
		return "short"
	}
	return ""
}

func (b *SimulatedBroker) calculateResults(trades []Trade, equityCurve []decimal.Decimal, initialCapital, finalEquity decimal.Decimal) BacktestResult {
	result := BacktestResult{
		TotalTrades:    len(trades),
		Trades:         trades,
		InitialCapital: initialCapital,
		FinalEquity:    finalEquity,
		GrossProfit:    decimal.ZERO,
		GrossLoss:      decimal.ZERO,
		TotalProfit:    decimal.ZERO,
	}

	if len(trades) == 0 {
		result.NetProfit = decimal.ZERO
		return result
	}

	for _, trade := range trades {
		if trade.Profit.IsPositive() {
			result.WinningTrades++
			result.GrossProfit = result.GrossProfit.Add(trade.Profit)
		} else if trade.Profit.IsNegative() {
			result.LosingTrades++
			result.GrossLoss = result.GrossLoss.Add(trade.Profit.Abs())
		}
		result.TotalProfit = result.TotalProfit.Add(trade.Profit)
	}

	if result.TotalTrades > 0 {
		result.WinRate = decimal.New(float64(result.WinningTrades)).Div(decimal.New(float64(result.TotalTrades)))
		result.AverageTrade = result.TotalProfit.Div(decimal.New(float64(result.TotalTrades)))
	}

	if !result.GrossLoss.IsZero() {
		result.ProfitFactor = result.GrossProfit.Div(result.GrossLoss)
	}

	result.NetProfit = finalEquity.Sub(initialCapital)

	peak := initialCapital
	for _, eq := range equityCurve {
		if eq.GT(peak) {
			peak = eq
		}
		drawdown := peak.Sub(eq)
		if drawdown.GT(result.MaxDrawdown) {
			result.MaxDrawdown = drawdown
			if !peak.IsZero() {
				result.MaxDrawdownPercent = drawdown.Div(peak)
			}
		}
	}

	return result
}

// EventDrivenBacktester processes market events bar-by-bar for realistic
// order simulation.
type EventDrivenBacktester struct {
	brokers    map[string]*SimulatedBroker
	strategies map[string]trading.Strategy
	analyzers  *AnalyzerRegistry
}

// NewEventDrivenBacktester creates a new event-driven backtester.
func NewEventDrivenBacktester() *EventDrivenBacktester {
	return &EventDrivenBacktester{
		brokers:    make(map[string]*SimulatedBroker),
		strategies: make(map[string]trading.Strategy),
		analyzers:  NewAnalyzerRegistry(),
	}
}

// Register associates a symbol with its broker and strategy.
func (edb *EventDrivenBacktester) Register(symbol string, broker *SimulatedBroker, strategy trading.Strategy) {
	if edb == nil || broker == nil || strategy == nil {
		return
	}
	if edb.brokers == nil {
		edb.brokers = make(map[string]*SimulatedBroker)
	}
	if edb.strategies == nil {
		edb.strategies = make(map[string]trading.Strategy)
	}
	edb.brokers[symbol] = broker
	edb.strategies[symbol] = strategy
}

// AddAnalyzer adds an analyzer to the backtester.
func (edb *EventDrivenBacktester) AddAnalyzer(a Analyzer) {
	if edb == nil {
		return
	}
	if edb.analyzers == nil {
		edb.analyzers = NewAnalyzerRegistry()
	}
	edb.analyzers.Add(a)
}

// Run processes events in chronological order and returns per-symbol results.
func (edb *EventDrivenBacktester) Run(events []Event) (map[string]BacktestResult, error) {
	if edb == nil {
		return nil, errors.New("event-driven backtester cannot be nil")
	}
	if len(events) == 0 {
		return make(map[string]BacktestResult), nil
	}

	sorted := make([]Event, len(events))
	copy(sorted, events)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	symbolIndex := make(map[string]int)

	for _, event := range sorted {
		switch event.Type {
		case EventBar:
			data, ok := event.Data.(BarEventData)
			if !ok {
				return nil, errors.New("invalid bar event data")
			}
			if err := edb.processBarEvent(event.Symbol, symbolIndex, data); err != nil {
				return nil, err
			}
		}
	}

	results := make(map[string]BacktestResult)
	for symbol, broker := range edb.brokers {
		result := broker.BacktestResult()

		metricsTrades := make([]metrics.Trade, len(result.Trades))
		for i, t := range result.Trades {
			metricsTrades[i] = metrics.Trade{
				Profit:    t.Profit,
				ProfitPct: t.ProfitPercent,
				Duration:  t.Duration,
				IsWin:     t.Profit.IsPositive(),
			}
		}

		metricsEquityCurve := make([]metrics.EquityPoint, len(broker.equityHistory))
		peak := broker.InitialCapital
		for i, eq := range broker.equityHistory {
			if eq.GT(peak) {
				peak = eq
			}
			drawdown := peak.Sub(eq)
			var drawdownPct decimal.Decimal
			if !peak.IsZero() {
				drawdownPct = drawdown.Div(peak)
			}
			metricsEquityCurve[i] = metrics.EquityPoint{
				Equity:      eq,
				Drawdown:    drawdown,
				DrawdownPct: drawdownPct,
			}
		}

		if edb.analyzers != nil {
			result.Analysis = edb.analyzers.Run(metricsTrades, metricsEquityCurve)
		}
		results[symbol] = result
	}

	return results, nil
}

func (edb *EventDrivenBacktester) processBarEvent(symbol string, symbolIndex map[string]int, data BarEventData) error {
	broker, ok := edb.brokers[symbol]
	if !ok {
		return nil
	}

	strategy, ok := edb.strategies[symbol]
	if !ok {
		return nil
	}

	candle := data.Candle
	if candle == nil {
		return nil
	}

	index := symbolIndex[symbol]
	symbolIndex[symbol] = index + 1

	broker.ProcessBar(index, candle)

	shouldEnter := strategy.ShouldEnter(index, broker.record)
	shouldExit := strategy.ShouldExit(index, broker.record)
	broker.ProcessStrategySignal(shouldEnter, shouldExit, index, candle)

	return nil
}
