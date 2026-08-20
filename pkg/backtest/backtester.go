package backtest

import (
	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/metrics"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/trading"
)

type Trade struct {
	EntryTime     int
	EntryPrice    decimal.Decimal
	ExitTime      int
	ExitPrice     decimal.Decimal
	Direction     string
	Quantity      decimal.Decimal
	Profit        decimal.Decimal
	ProfitPercent decimal.Decimal
	Duration      int
}

type Position struct {
	EntryTime  int
	EntryPrice decimal.Decimal
	Direction  string
	Quantity   decimal.Decimal
	StopLoss   decimal.Decimal
	TakeProfit decimal.Decimal
}

type BacktestResult struct {
	TotalTrades          int
	WinningTrades        int
	LosingTrades         int
	WinRate              decimal.Decimal
	TotalProfit          decimal.Decimal
	TotalLoss            decimal.Decimal
	NetProfit            decimal.Decimal
	GrossProfit          decimal.Decimal
	GrossLoss            decimal.Decimal
	ProfitFactor         decimal.Decimal
	AverageWin           decimal.Decimal
	AverageLoss          decimal.Decimal
	AverageTrade         decimal.Decimal
	MaxConsecutiveWins   int
	MaxConsecutiveLosses int
	MaxDrawdown          decimal.Decimal
	MaxDrawdownPercent   decimal.Decimal
	RecoveryFactor       decimal.Decimal
	RiskRewardRatio      decimal.Decimal
	CalmarRatio          decimal.Decimal
	SortinoRatio         decimal.Decimal
	SharpeRatio          decimal.Decimal
	CAGR                 decimal.Decimal
	FinalEquity          decimal.Decimal
	InitialCapital       decimal.Decimal
	Trades               []Trade
	Analysis             AnalysisResult
}

type BacktestConfig struct {
	InitialCapital decimal.Decimal
	PositionSize   decimal.Decimal
	RiskPerTrade   decimal.Decimal
	Commission     decimal.Decimal
	Slippage       decimal.Decimal
	AllowShort     bool
	AllowLong      bool
}

type Backtester struct {
	series    *series.TimeSeries
	strategy  trading.Strategy
	analyzers *AnalyzerRegistry
}

func NewBacktester(s *series.TimeSeries, strategy trading.Strategy) *Backtester {
	return &Backtester{
		series:    s,
		strategy:  strategy,
		analyzers: NewAnalyzerRegistry(),
	}
}

// AddAnalyzer adds an analyzer to the backtester.
func (b *Backtester) AddAnalyzer(a Analyzer) {
	if b == nil {
		return
	}
	if b.analyzers == nil {
		b.analyzers = NewAnalyzerRegistry()
	}
	b.analyzers.Add(a)
}

func (b *Backtester) Run(config BacktestConfig) BacktestResult {
	if b == nil || b.series == nil || b.strategy == nil {
		return BacktestResult{InitialCapital: config.InitialCapital, FinalEquity: config.InitialCapital}
	}
	positions := make([]Position, 0)
	trades := make([]Trade, 0)
	equityCurve := make([]decimal.Decimal, len(b.series.Candles))
	equity := config.InitialCapital
	mark := equityMark{value: decimal.ZERO}

	record := trading.NewTradingRecord()

	for i := 0; i < len(b.series.Candles); i++ {
		b.step(i, &positions, &trades, equityCurve, &equity, &mark, record, config)
	}

	b.finalizeOpenPositions(&positions, &trades, &equity, config)
	if len(equityCurve) > 0 {
		equityCurve[len(equityCurve)-1] = equity
	}

	result := b.calculateResults(trades, equityCurve, config.InitialCapital, equity)

	// Run analyzers
	metricsTrades := make([]metrics.Trade, len(trades))
	for i, t := range trades {
		metricsTrades[i] = metrics.Trade{
			Profit:    t.Profit,
			ProfitPct: t.ProfitPercent,
			Duration:  t.Duration,
			IsWin:     t.Profit.IsPositive(),
		}
	}

	metricsEquityCurve := make([]metrics.EquityPoint, len(equityCurve))
	peak := config.InitialCapital
	for i, eq := range equityCurve {
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

	if b.analyzers != nil {
		result.Analysis = b.analyzers.Run(metricsTrades, metricsEquityCurve)
	}

	return result
}

func (b *Backtester) step(
	index int,
	positions *[]Position,
	trades *[]Trade,
	equityCurve []decimal.Decimal,
	equity *decimal.Decimal,
	mark *equityMark,
	record *trading.TradingRecord,
	config BacktestConfig,
) {
	candle := b.series.Candles[index]
	if candle == nil {
		equityCurve[index] = *equity
		return
	}

	currentPrice := candle.ClosePrice
	mark.advance(currentPrice)

	b.closePositionsByStops(index, currentPrice, positions, trades, equity, mark, record, config)
	b.applyStrategy(index, currentPrice, positions, trades, equity, mark, record, config)
	equityCurve[index] = equity.Add(mark.value)
}

func (b *Backtester) closePositionsByStops(
	index int,
	currentPrice decimal.Decimal,
	positions *[]Position,
	trades *[]Trade,
	equity *decimal.Decimal,
	mark *equityMark,
	record *trading.TradingRecord,
	config BacktestConfig,
) {
	for j := len(*positions) - 1; j >= 0; j-- {
		pos := (*positions)[j]
		if !b.exitTriggered(pos, currentPrice) {
			continue
		}

		mark.remove(pos.Direction, pos.EntryPrice, pos.Quantity, currentPrice)
		exitPrice := b.exitPrice(currentPrice, pos.Direction, config)
		profit := b.positionProfit(pos, exitPrice)
		*trades = append(*trades, b.makeTrade(pos, index, exitPrice, profit))
		*equity = equity.Add(profit)
		*equity = equity.Sub(config.Commission)

		exitSide := trading.SELL
		if pos.Direction == "short" {
			exitSide = trading.BUY
		}
		record.Operate(trading.Order{
			Side:   exitSide,
			Price:  exitPrice,
			Amount: pos.Quantity,
		})

		*positions = append((*positions)[:j], (*positions)[j+1:]...)
	}
}

func (b *Backtester) applyStrategy(
	index int,
	currentPrice decimal.Decimal,
	positions *[]Position,
	trades *[]Trade,
	equity *decimal.Decimal,
	mark *equityMark,
	record *trading.TradingRecord,
	config BacktestConfig,
) {
	if b.strategy.ShouldEnter(index, record) {
		if config.AllowLong {
			b.openLong(index, currentPrice, positions, equity, mark, record, config)
		} else if config.AllowShort {
			b.openShort(index, currentPrice, positions, equity, mark, record, config)
		}
		return
	}

	if !b.strategy.ShouldExit(index, record) || len(*positions) == 0 {
		return
	}

	b.closeAllPositions(index, currentPrice, positions, trades, equity, mark, record, config)
}

func (b *Backtester) openLong(
	index int,
	currentPrice decimal.Decimal,
	positions *[]Position,
	equity *decimal.Decimal,
	mark *equityMark,
	record *trading.TradingRecord,
	config BacktestConfig,
) {
	quantity := config.PositionSize
	if quantity.IsZero() {
		quantity = b.positionQuantity(*equity, currentPrice, config)
	}
	entryPrice := b.entryPrice(currentPrice, "long", config)

	*positions = append(*positions, Position{
		EntryTime:  index,
		EntryPrice: entryPrice,
		Direction:  "long",
		Quantity:   quantity,
	})
	mark.add("long", entryPrice, quantity, currentPrice)

	record.Operate(trading.Order{
		Side:   trading.BUY,
		Price:  entryPrice,
		Amount: quantity,
	})
	*equity = equity.Sub(config.Commission)
}

func (b *Backtester) openShort(
	index int,
	currentPrice decimal.Decimal,
	positions *[]Position,
	equity *decimal.Decimal,
	mark *equityMark,
	record *trading.TradingRecord,
	config BacktestConfig,
) {
	quantity := config.PositionSize
	if quantity.IsZero() {
		quantity = b.positionQuantity(*equity, currentPrice, config)
	}
	entryPrice := b.entryPrice(currentPrice, "short", config)
	*positions = append(*positions, Position{
		EntryTime: index, EntryPrice: entryPrice, Direction: "short", Quantity: quantity,
	})
	mark.add("short", entryPrice, quantity, currentPrice)
	record.Operate(trading.Order{Side: trading.SELL, Price: entryPrice, Amount: quantity})
	*equity = equity.Sub(config.Commission)
}

func (b *Backtester) closeAllPositions(
	exitTime int,
	exitPrice decimal.Decimal,
	positions *[]Position,
	trades *[]Trade,
	equity *decimal.Decimal,
	mark *equityMark,
	record *trading.TradingRecord,
	config BacktestConfig,
) {
	for j := len(*positions) - 1; j >= 0; j-- {
		pos := (*positions)[j]
		mark.remove(pos.Direction, pos.EntryPrice, pos.Quantity, exitPrice)
		adjustedExitPrice := b.exitPrice(exitPrice, pos.Direction, config)
		profit := b.positionProfit(pos, adjustedExitPrice)
		*trades = append(*trades, b.makeTrade(pos, exitTime, adjustedExitPrice, profit))
		*equity = equity.Add(profit)
		*equity = equity.Sub(config.Commission)

		exitSide := trading.SELL
		if pos.Direction == "short" {
			exitSide = trading.BUY
		}
		record.Operate(trading.Order{
			Side:   exitSide,
			Price:  adjustedExitPrice,
			Amount: pos.Quantity,
		})

		*positions = append((*positions)[:j], (*positions)[j+1:]...)
	}
}

func (b *Backtester) finalizeOpenPositions(positions *[]Position, trades *[]Trade, equity *decimal.Decimal, config BacktestConfig) {
	if len(*positions) == 0 || len(b.series.Candles) == 0 {
		return
	}

	lastIndex := len(b.series.Candles) - 1
	lastCandle := b.series.Candles[lastIndex]
	if lastCandle == nil {
		return
	}
	exitPrice := lastCandle.ClosePrice

	for _, pos := range *positions {
		adjustedExitPrice := b.exitPrice(exitPrice, pos.Direction, config)
		profit := b.positionProfit(pos, adjustedExitPrice)
		*trades = append(*trades, b.makeTrade(pos, lastIndex, adjustedExitPrice, profit))
		*equity = equity.Add(profit)
		*equity = equity.Sub(config.Commission)
	}
}

func (b *Backtester) positionQuantity(equity, price decimal.Decimal, config BacktestConfig) decimal.Decimal {
	if config.RiskPerTrade.IsPositive() {
		return equity.Mul(config.RiskPerTrade).Div(price)
	}
	return equity.Div(price)
}

func (b *Backtester) entryPrice(price decimal.Decimal, direction string, config BacktestConfig) decimal.Decimal {
	if direction == "short" {
		return price.Sub(config.Slippage)
	}
	return price.Add(config.Slippage)
}

func (b *Backtester) exitPrice(price decimal.Decimal, direction string, config BacktestConfig) decimal.Decimal {
	if direction == "short" {
		return price.Add(config.Slippage)
	}
	return price.Sub(config.Slippage)
}

func (b *Backtester) exitTriggered(pos Position, currentPrice decimal.Decimal) bool {
	if pos.Direction == "long" {
		return b.exitTriggeredLong(pos, currentPrice)
	}
	if pos.Direction == "short" {
		return b.exitTriggeredShort(pos, currentPrice)
	}
	return false
}

func (b *Backtester) exitTriggeredLong(pos Position, currentPrice decimal.Decimal) bool {
	if !pos.StopLoss.IsZero() && currentPrice.LTE(pos.StopLoss) {
		return true
	}
	if !pos.TakeProfit.IsZero() && currentPrice.GTE(pos.TakeProfit) {
		return true
	}
	return false
}

func (b *Backtester) exitTriggeredShort(pos Position, currentPrice decimal.Decimal) bool {
	if !pos.StopLoss.IsZero() && currentPrice.GTE(pos.StopLoss) {
		return true
	}
	if !pos.TakeProfit.IsZero() && currentPrice.LTE(pos.TakeProfit) {
		return true
	}
	return false
}

func (b *Backtester) positionProfit(pos Position, exitPrice decimal.Decimal) decimal.Decimal {
	if pos.Direction == "long" {
		return exitPrice.Sub(pos.EntryPrice).Mul(pos.Quantity)
	}
	return pos.EntryPrice.Sub(exitPrice).Mul(pos.Quantity)
}

func (b *Backtester) makeTrade(pos Position, exitTime int, exitPrice, profit decimal.Decimal) Trade {
	trade := Trade{
		EntryTime:  pos.EntryTime,
		EntryPrice: pos.EntryPrice,
		ExitTime:   exitTime,
		ExitPrice:  exitPrice,
		Direction:  pos.Direction,
		Quantity:   pos.Quantity,
		Profit:     profit,
	}
	trade.ProfitPercent = profit.Div(pos.EntryPrice.Mul(pos.Quantity))
	trade.Duration = exitTime - pos.EntryTime
	return trade
}

func (b *Backtester) calculateResults(trades []Trade, equityCurve []decimal.Decimal, initialCapital, finalEquity decimal.Decimal) BacktestResult {
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

	drawdown, drawdownPercent := b.calculateMaxDrawdown(equityCurve, initialCapital)
	result.MaxDrawdown = drawdown
	result.MaxDrawdownPercent = drawdownPercent

	return result
}

func (b *Backtester) calculateMaxDrawdown(equityCurve []decimal.Decimal, initialCapital decimal.Decimal) (decimal.Decimal, decimal.Decimal) {
	maxDrawdown := decimal.ZERO
	maxDrawdownPercent := decimal.ZERO
	peak := initialCapital

	for _, equity := range equityCurve {
		if equity.GT(peak) {
			peak = equity
		}

		drawdown := peak.Sub(equity)
		if drawdown.GT(maxDrawdown) {
			maxDrawdown = drawdown
			if !peak.IsZero() {
				maxDrawdownPercent = drawdown.Div(peak)
			}
		}
	}

	return maxDrawdown, maxDrawdownPercent
}
