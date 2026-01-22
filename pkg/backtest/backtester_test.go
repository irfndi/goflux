package backtest

import (
	"testing"
	"time"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/trading"
	"github.com/stretchr/testify/assert"
)

func createTestSeries() *series.TimeSeries {
	s := series.NewTimeSeries()

	baseTime := time.Now().Add(-100 * time.Hour)

	for i := 0; i < 100; i++ {
		trend := decimal.New(float64(i) * 0.5)
		noise := decimal.New(float64(i%10) * 0.1)
		close := trend.Add(noise).Add(decimal.New(100))

		s.AddCandle(&series.Candle{
			OpenPrice:  close.Sub(decimal.New(1)),
			MaxPrice:   close.Add(decimal.New(2)),
			MinPrice:   close.Sub(decimal.New(2)),
			ClosePrice: close,
			Volume:     decimal.New(1000),
			Period:     series.NewTimePeriod(baseTime.Add(time.Duration(i)*time.Hour), time.Hour),
		})
	}

	return s
}

type simpleStrategy struct {
	trading.Strategy
}

func (s *simpleStrategy) ShouldEnter(index int, record *trading.TradingRecord) bool {
	return (index == 0 && record.CurrentPosition().IsNew()) || (index > 20 && index%10 == 0 && record.CurrentPosition().IsNew())
}

func (s *simpleStrategy) ShouldExit(index int, record *trading.TradingRecord) bool {
	return index > 30 && index%15 == 0 && record.CurrentPosition().IsOpen()
}

func TestBacktesterBasic(t *testing.T) {
	s := createTestSeries()
	strategy := &simpleStrategy{}

	backtester := NewBacktester(s, strategy)

	config := BacktestConfig{
		InitialCapital: decimal.New(10000),
		PositionSize:   decimal.New(100),
		AllowLong:      true,
		AllowShort:     false,
	}

	result := backtester.Run(config)

	assert.NotEqual(t, 0, result.TotalTrades)
	assert.True(t, result.FinalEquity.GT(decimal.ZERO))
}

func TestBacktesterNoTrades(t *testing.T) {
	s := series.NewTimeSeries()

	for i := 0; i < 10; i++ {
		s.AddCandle(&series.Candle{
			OpenPrice:  decimal.New(100),
			MaxPrice:   decimal.New(105),
			MinPrice:   decimal.New(95),
			ClosePrice: decimal.New(100),
			Volume:     decimal.New(1000),
			Period:     series.NewTimePeriod(time.Now().Add(time.Duration(i)*time.Hour), time.Hour),
		})
	}

	noSignalStrategy := &noSignalStrategy{}

	backtester := NewBacktester(s, noSignalStrategy)

	config := BacktestConfig{
		InitialCapital: decimal.New(10000),
		PositionSize:   decimal.New(100),
		AllowLong:      true,
	}

	result := backtester.Run(config)

	assert.Equal(t, 0, result.TotalTrades)
	assert.Equal(t, config.InitialCapital.Float(), result.FinalEquity.Float())
}

type noSignalStrategy struct {
	trading.Strategy
}

func (s *noSignalStrategy) ShouldEnter(index int, record *trading.TradingRecord) bool {
	return false
}

func (s *noSignalStrategy) ShouldExit(index int, record *trading.TradingRecord) bool {
	return false
}

func TestBacktestResultCalculations(t *testing.T) {
	bt := &Backtester{}

	trades := []Trade{
		{Profit: decimal.New(100)},
		{Profit: decimal.New(-50)},
		{Profit: decimal.New(200)},
	}

	equityCurve := []decimal.Decimal{decimal.New(1000), decimal.New(1100), decimal.New(1050), decimal.New(1250)}

	result := bt.calculateResults(trades, equityCurve, decimal.New(1000), decimal.New(1250))

	assert.Equal(t, 3, result.TotalTrades)
	assert.Equal(t, 2, result.WinningTrades)
	assert.Equal(t, 1, result.LosingTrades)
	assert.Equal(t, 250.0, result.NetProfit.Float())
}

type stopLossStrategy struct {
	trading.Strategy
}

func (s *stopLossStrategy) ShouldEnter(index int, record *trading.TradingRecord) bool {
	return index == 0 && record.CurrentPosition().IsNew()
}

func (s *stopLossStrategy) ShouldExit(index int, record *trading.TradingRecord) bool {
	return index > 5 && index%5 == 0 && record.CurrentPosition().IsOpen()
}

func TestBacktesterWithStopLoss(t *testing.T) {
	s := series.NewTimeSeries()

	for i := 0; i < 50; i++ {
		price := decimal.New(float64(100 + i))
		s.AddCandle(&series.Candle{
			OpenPrice:  price.Sub(decimal.New(1)),
			MaxPrice:   price.Add(decimal.New(2)),
			MinPrice:   price.Sub(decimal.New(3)),
			ClosePrice: price,
			Volume:     decimal.New(1000),
			Period:     series.NewTimePeriod(time.Now().Add(time.Duration(i)*time.Hour), time.Hour),
		})
	}

	strategy := &stopLossStrategy{}

	backtester := NewBacktester(s, strategy)

	config := BacktestConfig{
		InitialCapital: decimal.New(10000),
		PositionSize:   decimal.New(50),
		AllowLong:      true,
	}

	result := backtester.Run(config)

	assert.True(t, result.TotalTrades > 0)
}
