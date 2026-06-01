package backtest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/trading"
)

type wfaTestStrategy struct{}

func (s *wfaTestStrategy) ShouldEnter(index int, record *trading.TradingRecord) bool {
	return index == 0 && record.CurrentPosition().IsNew()
}

func (s *wfaTestStrategy) ShouldExit(index int, record *trading.TradingRecord) bool {
	return index > 5 && index%5 == 0 && record.CurrentPosition().IsOpen()
}

type neverEnterStrategy struct{}

func (s *neverEnterStrategy) ShouldEnter(int, *trading.TradingRecord) bool { return false }
func (s *neverEnterStrategy) ShouldExit(int, *trading.TradingRecord) bool  { return false }

func makeTestSeries(n int) *series.TimeSeries {
	s := series.NewTimeSeries()
	for i := 0; i < n; i++ {
		close := decimal.New(float64(100 + i))
		s.AddCandle(&series.Candle{
			OpenPrice:  close.Sub(decimal.New(1)),
			MaxPrice:   close.Add(decimal.New(2)),
			MinPrice:   close.Sub(decimal.New(2)),
			ClosePrice: close,
			Volume:     decimal.New(1000),
			Period:     series.NewTimePeriod(time.Now().Add(time.Duration(i)*time.Hour), time.Hour),
		})
	}
	return s
}

func TestWFAConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config WFAConfig
		want   string
	}{
		{
			name:   "in-sample too small",
			config: WFAConfig{InSampleWindowSize: 5, OutOfSampleWindowSize: 10, StepSize: 5},
			want:   "in-sample window must be >= 10",
		},
		{
			name:   "out-of-sample zero",
			config: WFAConfig{InSampleWindowSize: 20, OutOfSampleWindowSize: 0, StepSize: 5},
			want:   "out-of-sample window must be >= 1",
		},
		{
			name:   "step size zero",
			config: WFAConfig{InSampleWindowSize: 20, OutOfSampleWindowSize: 10, StepSize: 0},
			want:   "step size must be >= 1",
		},
		{
			name:   "valid config",
			config: WFAConfig{InSampleWindowSize: 20, OutOfSampleWindowSize: 10, StepSize: 5},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.want == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.want)
			}
		})
	}
}

func TestWFANewAnalyzerInvalidConfig(t *testing.T) {
	_, err := NewWalkForwardAnalyzer(WFAConfig{InSampleWindowSize: 5})
	assert.Error(t, err)
}

func TestWFARunEmptySeries(t *testing.T) {
	analyzer, err := NewWalkForwardAnalyzer(WFAConfig{
		InSampleWindowSize:    20,
		OutOfSampleWindowSize: 10,
		StepSize:              5,
	})
	assert.NoError(t, err)

	result, err := analyzer.Run(nil, func(ts *series.TimeSeries) (trading.Strategy, BacktestConfig) {
		return &wfaTestStrategy{}, BacktestConfig{InitialCapital: decimal.New(10000), AllowLong: true}
	})
	assert.NoError(t, err)
	assert.Empty(t, result.Windows)
	assert.Equal(t, 0, result.AggregateMetrics.TotalWindows)

	emptySeries := series.NewTimeSeries()
	result, err = analyzer.Run(emptySeries, func(ts *series.TimeSeries) (trading.Strategy, BacktestConfig) {
		return &wfaTestStrategy{}, BacktestConfig{InitialCapital: decimal.New(10000), AllowLong: true}
	})
	assert.NoError(t, err)
	assert.Empty(t, result.Windows)
}

func TestWFARunSingleWindow(t *testing.T) {
	ts := makeTestSeries(40)
	analyzer, err := NewWalkForwardAnalyzer(WFAConfig{
		InSampleWindowSize:    20,
		OutOfSampleWindowSize: 10,
		StepSize:              30, // Only one window fits
	})
	assert.NoError(t, err)

	result, err := analyzer.Run(ts, func(s *series.TimeSeries) (trading.Strategy, BacktestConfig) {
		return &wfaTestStrategy{}, BacktestConfig{
			InitialCapital: decimal.New(10000),
			PositionSize:   decimal.New(10),
			AllowLong:      true,
		}
	})
	assert.NoError(t, err)
	assert.Len(t, result.Windows, 1)

	w := result.Windows[0]
	assert.Equal(t, 0, w.WindowIndex)
	assert.Equal(t, 0, w.InSampleStart)
	assert.Equal(t, 20, w.InSampleEnd)
	assert.Equal(t, 20, w.OutOfSampleStart)
	assert.Equal(t, 30, w.OutOfSampleEnd)

	assert.Equal(t, 1, result.AggregateMetrics.TotalWindows)
}

func TestWFARunMultipleWindows(t *testing.T) {
	ts := makeTestSeries(100)
	analyzer, err := NewWalkForwardAnalyzer(WFAConfig{
		InSampleWindowSize:    20,
		OutOfSampleWindowSize: 10,
		StepSize:              15,
	})
	assert.NoError(t, err)

	result, err := analyzer.Run(ts, func(s *series.TimeSeries) (trading.Strategy, BacktestConfig) {
		return &wfaTestStrategy{}, BacktestConfig{
			InitialCapital: decimal.New(10000),
			PositionSize:   decimal.New(10),
			AllowLong:      true,
		}
	})
	assert.NoError(t, err)

	// Windows: 0+20+10=30, 15+20+10=45, 30+20+10=60, 45+20+10=75, 60+20+10=90
	// Next would be 75+20+10=105 > 100, so 5 windows
	assert.GreaterOrEqual(t, len(result.Windows), 4)
	assert.LessOrEqual(t, len(result.Windows), 6)

	for i, w := range result.Windows {
		assert.Equal(t, i, w.WindowIndex)
		assert.Equal(t, w.InSampleEnd-w.InSampleStart, 20)
		assert.Equal(t, w.OutOfSampleEnd-w.OutOfSampleStart, 10)
		assert.Equal(t, w.OutOfSampleStart, w.InSampleEnd)
	}
}

func TestWFAAggregateMetrics(t *testing.T) {
	// Create two windows with known results manually
	windows := []WFWindowResult{
		{
			InSampleResult: BacktestResult{
				SharpeRatio: decimal.New(2.0),
				NetProfit:   decimal.New(1000),
			},
			OutOfSampleResult: BacktestResult{
				SharpeRatio: decimal.New(1.0),
				NetProfit:   decimal.New(500),
			},
		},
		{
			InSampleResult: BacktestResult{
				SharpeRatio: decimal.New(1.5),
				NetProfit:   decimal.New(2000),
			},
			OutOfSampleResult: BacktestResult{
				SharpeRatio: decimal.New(0.5),
				NetProfit:   decimal.New(-200),
			},
		},
	}

	analyzer := &WalkForwardAnalyzer{}
	metrics := analyzer.computeAggregateMetrics(windows)

	assert.Equal(t, 2, metrics.TotalWindows)

	// Average IS Sharpe = (2.0 + 1.5) / 2 = 1.75
	assert.True(t, metrics.AverageInSampleSharpe.Sub(decimal.New(1.75)).Abs().LT(decimal.New(0.001)))

	// Average OOS Sharpe = (1.0 + 0.5) / 2 = 0.75
	assert.True(t, metrics.AverageOutOfSampleSharpe.Sub(decimal.New(0.75)).Abs().LT(decimal.New(0.001)))

	// Degradation = (1.75 - 0.75) / 1.75 = 0.5714...
	expectedDegradation := decimal.New(1.75).Sub(decimal.New(0.75)).Div(decimal.New(1.75))
	assert.True(t, metrics.DegradationRate.Sub(expectedDegradation).Abs().LT(decimal.New(0.001)))

	// Winning windows = 1 out of 2 = 50%
	assert.True(t, metrics.WinningWindowsPercent.Sub(decimal.New(50)).Abs().LT(decimal.New(0.001)))

	// Average IS Profit = (1000 + 2000) / 2 = 1500
	assert.True(t, metrics.AverageInSampleProfit.Sub(decimal.New(1500)).Abs().LT(decimal.New(0.001)))

	// Average OOS Profit = (500 + (-200)) / 2 = 150
	assert.True(t, metrics.AverageOutOfSampleProfit.Sub(decimal.New(150)).Abs().LT(decimal.New(0.001)))
}

func TestWFAZeroWindowsMetrics(t *testing.T) {
	analyzer := &WalkForwardAnalyzer{}
	metrics := analyzer.computeAggregateMetrics(nil)
	assert.Equal(t, 0, metrics.TotalWindows)
	assert.True(t, metrics.AverageInSampleSharpe.IsZero())

	metrics = analyzer.computeAggregateMetrics([]WFWindowResult{})
	assert.Equal(t, 0, metrics.TotalWindows)
}

func TestWFADegradationWithZeroISSharpe(t *testing.T) {
	windows := []WFWindowResult{
		{
			InSampleResult: BacktestResult{
				SharpeRatio: decimal.ZERO,
				NetProfit:   decimal.New(100),
			},
			OutOfSampleResult: BacktestResult{
				SharpeRatio: decimal.New(0.5),
				NetProfit:   decimal.New(50),
			},
		},
	}

	analyzer := &WalkForwardAnalyzer{}
	metrics := analyzer.computeAggregateMetrics(windows)
	assert.True(t, metrics.DegradationRate.IsZero())
}

func TestWFASliceTimeSeries(t *testing.T) {
	ts := makeTestSeries(50)

	// Normal slice
	sliced := sliceTimeSeries(ts, 10, 20)
	assert.Equal(t, 10, sliced.Length())
	assert.Equal(t, ts.GetCandle(10).ClosePrice.Float(), sliced.GetCandle(0).ClosePrice.Float())
	assert.Equal(t, ts.GetCandle(19).ClosePrice.Float(), sliced.GetCandle(9).ClosePrice.Float())

	// Clamp start < 0
	clampedStart := sliceTimeSeries(ts, -5, 5)
	assert.Equal(t, 5, clampedStart.Length())

	// Clamp end > length
	clampedEnd := sliceTimeSeries(ts, 45, 100)
	assert.Equal(t, 5, clampedEnd.Length())
}

func TestWFAWithNeverEnterStrategy(t *testing.T) {
	ts := makeTestSeries(50)
	analyzer, err := NewWalkForwardAnalyzer(WFAConfig{
		InSampleWindowSize:    20,
		OutOfSampleWindowSize: 10,
		StepSize:              30,
	})
	assert.NoError(t, err)

	result, err := analyzer.Run(ts, func(s *series.TimeSeries) (trading.Strategy, BacktestConfig) {
		return &neverEnterStrategy{}, BacktestConfig{
			InitialCapital: decimal.New(10000),
			AllowLong:      true,
		}
	})
	assert.NoError(t, err)
	assert.Len(t, result.Windows, 1)

	// No trades should be made
	assert.Equal(t, 0, result.Windows[0].InSampleResult.TotalTrades)
	assert.Equal(t, 0, result.Windows[0].OutOfSampleResult.TotalTrades)
}
