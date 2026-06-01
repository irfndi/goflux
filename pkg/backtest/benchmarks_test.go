package backtest

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/testutils"
	"github.com/irfndi/goflux/pkg/trading"
)

func BenchmarkWalkForwardAnalysis(b *testing.B) {
	ts := testutils.RandomTimeSeries(500)

	analyzer, _ := NewWalkForwardAnalyzer(WFAConfig{
		InSampleWindowSize:    100,
		OutOfSampleWindowSize: 50,
		StepSize:              50,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = analyzer.Run(ts, func(s *series.TimeSeries) (StrategyFactory, BacktestConfig) {
			return func(ts *series.TimeSeries) trading.Strategy { return &benchmarkStrategy{} }, BacktestConfig{
				InitialCapital: decimal.New(10000),
				PositionSize:   decimal.New(10),
				AllowLong:      true,
			}
		})
	}
}

type benchmarkStrategy struct{}

func (s *benchmarkStrategy) ShouldEnter(index int, record *trading.TradingRecord) bool {
	return index == 0 && record.CurrentPosition().IsNew()
}

func (s *benchmarkStrategy) ShouldExit(index int, record *trading.TradingRecord) bool {
	return index > 5 && index%5 == 0 && record.CurrentPosition().IsOpen()
}

func BenchmarkSliceTimeSeries(b *testing.B) {
	ts := testutils.RandomTimeSeries(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sliceTimeSeries(ts, 1000, 2000)
	}
}

func BenchmarkComputeAggregateMetrics(b *testing.B) {
	windows := make([]WFWindowResult, 20)
	for i := range windows {
		windows[i] = WFWindowResult{
			InSampleResult: BacktestResult{
				SharpeRatio: decimal.New(2.0),
				NetProfit:   decimal.New(1000),
			},
			OutOfSampleResult: BacktestResult{
				SharpeRatio: decimal.New(1.0),
				NetProfit:   decimal.New(500),
			},
		}
	}
	analyzer := &WalkForwardAnalyzer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.computeAggregateMetrics(windows)
	}
}
