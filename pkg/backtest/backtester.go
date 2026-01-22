package backtest

import (
	"github.com/irfndi/goflux/pkg/decimal"
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
	series   *series.TimeSeries
	strategy trading.Strategy
}

func NewBacktester(s *series.TimeSeries, strategy trading.Strategy) *Backtester {
	return &Backtester{
		series:   s,
		strategy: strategy,
	}
}

func (b *Backtester) Run(config BacktestConfig) BacktestResult {
	positions := make([]Position, 0)
	trades := make([]Trade, 0)
	equityCurve := make([]decimal.Decimal, len(b.series.Candles))
	equity := config.InitialCapital

	record := trading.NewTradingRecord()

	for i := 0; i < len(b.series.Candles); i++ {
		equityCurve[i] = equity

		candle := b.series.Candles[i]
		if candle == nil {
			continue
		}

		currentPrice := candle.ClosePrice

		// Check internal StopLoss/TakeProfit for all open positions
		for j := len(positions) - 1; j >= 0; j-- {
			pos := &positions[j]
			shouldExit := false

			if pos.Direction == "long" {
				if !pos.StopLoss.IsZero() && currentPrice.LTE(pos.StopLoss) {
					shouldExit = true
				}
				if !pos.TakeProfit.IsZero() && currentPrice.GTE(pos.TakeProfit) {
					shouldExit = true
				}
			} else if pos.Direction == "short" {
				if !pos.StopLoss.IsZero() && currentPrice.GTE(pos.StopLoss) {
					shouldExit = true
				}
				if !pos.TakeProfit.IsZero() && currentPrice.LTE(pos.TakeProfit) {
					shouldExit = true
				}
			}

			if shouldExit {
				var profit decimal.Decimal
				if pos.Direction == "long" {
					profit = currentPrice.Sub(pos.EntryPrice).Mul(pos.Quantity)
				} else {
					profit = pos.EntryPrice.Sub(currentPrice).Mul(pos.Quantity)
				}

				trade := Trade{
					EntryTime:  pos.EntryTime,
					EntryPrice: pos.EntryPrice,
					ExitTime:   i,
					ExitPrice:  currentPrice,
					Direction:  pos.Direction,
					Quantity:   pos.Quantity,
					Profit:     profit,
				}
				trade.ProfitPercent = profit.Div(pos.EntryPrice.Mul(pos.Quantity))
				trade.Duration = i - pos.EntryTime
				trades = append(trades, trade)
				equity = equity.Add(profit)

				// Sync with trading.TradingRecord
				record.Operate(trading.Order{
					Side:   trading.SELL,
					Price:  currentPrice,
					Amount: pos.Quantity,
				})

				positions = append(positions[:j], positions[j+1:]...)
			}
		}

		// Strategy Entry
		if b.strategy.ShouldEnter(i, record) {
			if config.AllowLong {
				quantity := config.PositionSize
				if quantity.IsZero() {
					quantity = equity.Div(currentPrice)
				}

				position := Position{
					EntryTime:  i,
					EntryPrice: currentPrice,
					Direction:  "long",
					Quantity:   quantity,
				}

				positions = append(positions, position)
				record.Operate(trading.Order{
					Side:   trading.BUY,
					Price:  currentPrice,
					Amount: quantity,
				})
			}
		} else if b.strategy.ShouldExit(i, record) && len(positions) > 0 {
			// Strategy Exit
			for j := len(positions) - 1; j >= 0; j-- {
				pos := &positions[j]
				var profit decimal.Decimal
				if pos.Direction == "long" {
					profit = currentPrice.Sub(pos.EntryPrice).Mul(pos.Quantity)
				} else {
					profit = pos.EntryPrice.Sub(currentPrice).Mul(pos.Quantity)
				}

				trade := Trade{
					EntryTime:  pos.EntryTime,
					EntryPrice: pos.EntryPrice,
					ExitTime:   i,
					ExitPrice:  currentPrice,
					Direction:  pos.Direction,
					Quantity:   pos.Quantity,
					Profit:     profit,
				}
				trade.ProfitPercent = profit.Div(pos.EntryPrice.Mul(pos.Quantity))
				trade.Duration = i - pos.EntryTime
				trades = append(trades, trade)
				equity = equity.Add(profit)

				record.Operate(trading.Order{
					Side:   trading.SELL,
					Price:  currentPrice,
					Amount: pos.Quantity,
				})

				positions = append(positions[:j], positions[j+1:]...)
			}
		}
	}

	// Final close out
	for _, pos := range positions {
		candle := b.series.Candles[len(b.series.Candles)-1]
		currentPrice := candle.ClosePrice
		var profit decimal.Decimal
		if pos.Direction == "long" {
			profit = currentPrice.Sub(pos.EntryPrice).Mul(pos.Quantity)
		} else {
			profit = pos.EntryPrice.Sub(currentPrice).Mul(pos.Quantity)
		}

		trade := Trade{
			EntryTime:  pos.EntryTime,
			EntryPrice: pos.EntryPrice,
			ExitTime:   len(b.series.Candles) - 1,
			ExitPrice:  currentPrice,
			Direction:  pos.Direction,
			Quantity:   pos.Quantity,
			Profit:     profit,
		}
		trade.ProfitPercent = profit.Div(pos.EntryPrice.Mul(pos.Quantity))
		trade.Duration = trade.ExitTime - trade.EntryTime
		trades = append(trades, trade)
		equity = equity.Add(profit)
	}

	return b.calculateResults(trades, equityCurve, config.InitialCapital, equity)
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
