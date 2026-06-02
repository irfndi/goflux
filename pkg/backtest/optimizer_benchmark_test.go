package backtest

import (
	"testing"
	"time"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/trading"
)

func benchmarkOptimizerSeries() *series.TimeSeries {
	s := series.NewTimeSeries()
	base := time.Now().Add(-200 * time.Hour)
	for i := 0; i < 200; i++ {
		closePrice := decimal.New(float64(100 + i%50))
		s.AddCandle(&series.Candle{
			OpenPrice:  closePrice.Sub(decimal.New(1)),
			MaxPrice:   closePrice.Add(decimal.New(2)),
			MinPrice:   closePrice.Sub(decimal.New(2)),
			ClosePrice: closePrice,
			Volume:     decimal.New(1000),
			Period:     series.NewTimePeriod(base.Add(time.Duration(i)*time.Hour), time.Hour),
		})
	}
	return s
}

func benchmarkStrategyFactory(params map[string]float64) trading.Strategy {
	return &thresholdStrategy{threshold: int(params["threshold"])}
}

func BenchmarkOptimizer_GridSearch_Sequential(b *testing.B) {
	ts := benchmarkOptimizerSeries()
	config := OptimizationConfig{
		Method: OptMethodGridSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "threshold", Min: 0, Max: 20, Step: 5},
		},
		ObjectiveFunc: ObjectiveNetProfit,
		MaxWorkers:    1,
	}
	opt, _ := NewOptimizer(config)
	btConfig := defaultOptimizerBTConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := opt.Optimize(ts, benchmarkStrategyFactory, btConfig); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOptimizer_GridSearch_Parallel(b *testing.B) {
	ts := benchmarkOptimizerSeries()
	config := OptimizationConfig{
		Method: OptMethodGridSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "threshold", Min: 0, Max: 20, Step: 5},
		},
		ObjectiveFunc: ObjectiveNetProfit,
		MaxWorkers:    4,
	}
	opt, _ := NewOptimizer(config)
	btConfig := defaultOptimizerBTConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := opt.Optimize(ts, benchmarkStrategyFactory, btConfig); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOptimizer_RandomSearch(b *testing.B) {
	ts := benchmarkOptimizerSeries()
	config := OptimizationConfig{
		Method: OptMethodRandomSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "threshold", Min: 0, Max: 50},
		},
		ObjectiveFunc: ObjectiveNetProfit,
		RandomSamples: 20,
		MaxWorkers:    4,
		Seed:          42,
	}
	opt, _ := NewOptimizer(config)
	btConfig := defaultOptimizerBTConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := opt.Optimize(ts, benchmarkStrategyFactory, btConfig); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOptimizer_LargeGrid(b *testing.B) {
	ts := benchmarkOptimizerSeries()
	config := OptimizationConfig{
		Method: OptMethodGridSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "a", Min: 0, Max: 4, Step: 1},
			{Name: "b", Min: 0, Max: 4, Step: 1},
		},
		ObjectiveFunc: ObjectiveNetProfit,
		MaxWorkers:    4,
	}
	opt, _ := NewOptimizer(config)
	btConfig := defaultOptimizerBTConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := opt.Optimize(ts, benchmarkStrategyFactory, btConfig); err != nil {
			b.Fatal(err)
		}
	}
}
