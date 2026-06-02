package backtest

import (
	"testing"
	"time"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/trading"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// thresholdStrategy enters when index > threshold and exits after a fixed number of bars.
// Higher threshold = fewer trades.
type thresholdStrategy struct {
	threshold int
}

func (s *thresholdStrategy) ShouldEnter(index int, record *trading.TradingRecord) bool {
	return index > s.threshold && record.CurrentPosition().IsNew()
}

func (s *thresholdStrategy) ShouldExit(index int, record *trading.TradingRecord) bool {
	return record.CurrentPosition().IsOpen() && index-s.threshold > 5
}

func createOptimizerTestSeries() *series.TimeSeries {
	s := series.NewTimeSeries()
	base := time.Now().Add(-50 * time.Hour)
	for i := 0; i < 50; i++ {
		closePrice := decimal.New(float64(100 + i))
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

func thresholdStrategyFactory(params map[string]float64) trading.Strategy {
	return &thresholdStrategy{threshold: int(params["threshold"])}
}

func defaultOptimizerBTConfig() BacktestConfig {
	return BacktestConfig{
		InitialCapital: decimal.New(10000),
		PositionSize:   decimal.New(10),
		AllowLong:      true,
		AllowShort:     false,
	}
}

func TestOptimizer_GridSearch(t *testing.T) {
	ts := createOptimizerTestSeries()
	config := OptimizationConfig{
		Method: OptMethodGridSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "threshold", Min: 5, Max: 15, Step: 5},
		},
		ObjectiveFunc: ObjectiveNetProfit,
		MaxWorkers:    1,
	}

	opt, err := NewOptimizer(config)
	require.NoError(t, err)

	result, err := opt.Optimize(ts, thresholdStrategyFactory, defaultOptimizerBTConfig())
	require.NoError(t, err)
	require.NotNil(t, result)

	// threshold=5,10,15 → 3 combinations
	assert.Equal(t, 3, result.TotalRuns)
	assert.Equal(t, 3, len(result.AllResults))

	// Results should be sorted by score descending
	for i := 1; i < len(result.AllResults); i++ {
		assert.True(t, result.AllResults[i-1].Score >= result.AllResults[i].Score)
	}

	// Best config should match the top result
	assert.Equal(t, result.AllResults[0].Params, result.BestConfig)
	assert.Equal(t, result.AllResults[0].Score, result.BestScore)
}

func TestOptimizer_RandomSearch(t *testing.T) {
	ts := createOptimizerTestSeries()
	config := OptimizationConfig{
		Method: OptMethodRandomSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "threshold", Min: 0, Max: 20},
		},
		ObjectiveFunc: ObjectiveNetProfit,
		RandomSamples: 10,
		MaxWorkers:    1,
		Seed:          42,
	}

	opt, err := NewOptimizer(config)
	require.NoError(t, err)

	result, err := opt.Optimize(ts, thresholdStrategyFactory, defaultOptimizerBTConfig())
	require.NoError(t, err)
	assert.Equal(t, 10, result.TotalRuns)
	assert.Equal(t, 10, len(result.AllResults))
}

func TestOptimizer_ReproducibleWithSeed(t *testing.T) {
	ts := createOptimizerTestSeries()
	config := OptimizationConfig{
		Method: OptMethodRandomSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "threshold", Min: 0, Max: 20},
		},
		ObjectiveFunc: ObjectiveNetProfit,
		RandomSamples: 5,
		MaxWorkers:    1,
		Seed:          123,
	}

	opt1, _ := NewOptimizer(config)
	r1, _ := opt1.Optimize(ts, thresholdStrategyFactory, defaultOptimizerBTConfig())

	opt2, _ := NewOptimizer(config)
	r2, _ := opt2.Optimize(ts, thresholdStrategyFactory, defaultOptimizerBTConfig())

	assert.Equal(t, r1.TotalRuns, r2.TotalRuns)
	assert.Equal(t, r1.BestScore, r2.BestScore)
	for i := range r1.AllResults {
		assert.Equal(t, r1.AllResults[i].Params, r2.AllResults[i].Params)
	}
}

func TestOptimizer_DifferentSeeds(t *testing.T) {
	ts := createOptimizerTestSeries()
	config := OptimizationConfig{
		Method: OptMethodRandomSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "threshold", Min: 0, Max: 20},
		},
		ObjectiveFunc: ObjectiveNetProfit,
		RandomSamples: 5,
		MaxWorkers:    1,
		Seed:          111,
	}

	opt1, _ := NewOptimizer(config)
	r1, _ := opt1.Optimize(ts, thresholdStrategyFactory, defaultOptimizerBTConfig())

	config.Seed = 222
	opt2, _ := NewOptimizer(config)
	r2, _ := opt2.Optimize(ts, thresholdStrategyFactory, defaultOptimizerBTConfig())

	// Different seeds should produce different parameter samples (with high probability)
	different := false
	for i := range r1.AllResults {
		if r1.AllResults[i].Params["threshold"] != r2.AllResults[i].Params["threshold"] {
			different = true
			break
		}
	}
	assert.True(t, different, "different seeds should produce different samples")
}

func TestOptimizer_ParallelCorrectness(t *testing.T) {
	ts := createOptimizerTestSeries()
	config := OptimizationConfig{
		Method: OptMethodGridSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "threshold", Min: 0, Max: 10, Step: 2},
		},
		ObjectiveFunc: ObjectiveNetProfit,
		MaxWorkers:    1,
	}

	optSeq, _ := NewOptimizer(config)
	resultSeq, _ := optSeq.Optimize(ts, thresholdStrategyFactory, defaultOptimizerBTConfig())

	config.MaxWorkers = 4
	optPar, _ := NewOptimizer(config)
	resultPar, _ := optPar.Optimize(ts, thresholdStrategyFactory, defaultOptimizerBTConfig())

	assert.Equal(t, resultSeq.TotalRuns, resultPar.TotalRuns)
	assert.Equal(t, resultSeq.BestScore, resultPar.BestScore)
	assert.Equal(t, resultSeq.BestConfig, resultPar.BestConfig)
}

func TestOptimizer_EmptyTimeSeries(t *testing.T) {
	config := OptimizationConfig{
		Method: OptMethodGridSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "threshold", Min: 0, Max: 10, Step: 5},
		},
		ObjectiveFunc: ObjectiveNetProfit,
	}

	opt, _ := NewOptimizer(config)
	result, err := opt.Optimize(series.NewTimeSeries(), thresholdStrategyFactory, defaultOptimizerBTConfig())
	require.NoError(t, err)
	assert.Equal(t, 0, result.TotalRuns)
}

func TestOptimizer_Validation(t *testing.T) {
	// No parameter spaces
	_, err := NewOptimizer(OptimizationConfig{
		Method:        OptMethodGridSearch,
		ObjectiveFunc: ObjectiveNetProfit,
	})
	assert.Error(t, err)

	// Grid search with step = 0
	_, err = NewOptimizer(OptimizationConfig{
		Method: OptMethodGridSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "x", Min: 0, Max: 10, Step: 0},
		},
		ObjectiveFunc: ObjectiveNetProfit,
	})
	assert.Error(t, err)

	// Random search with RandomSamples = 0
	_, err = NewOptimizer(OptimizationConfig{
		Method: OptMethodRandomSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "x", Min: 0, Max: 10},
		},
		ObjectiveFunc: ObjectiveNetProfit,
	})
	assert.Error(t, err)

	// Nil objective function
	_, err = NewOptimizer(OptimizationConfig{
		Method: OptMethodGridSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "x", Min: 0, Max: 10, Step: 1},
		},
	})
	assert.Error(t, err)

	// Min > Max
	_, err = NewOptimizer(OptimizationConfig{
		Method: OptMethodGridSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "x", Min: 10, Max: 0, Step: 1},
		},
		ObjectiveFunc: ObjectiveNetProfit,
	})
	assert.Error(t, err)

	// Empty parameter name
	_, err = NewOptimizer(OptimizationConfig{
		Method: OptMethodGridSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "", Min: 0, Max: 10, Step: 1},
		},
		ObjectiveFunc: ObjectiveNetProfit,
	})
	assert.Error(t, err)

	// Unsupported method
	_, err = NewOptimizer(OptimizationConfig{
		Method: "genetic",
		ParameterSpaces: []ParameterSpace{
			{Name: "x", Min: 0, Max: 10, Step: 1},
		},
		ObjectiveFunc: ObjectiveNetProfit,
	})
	assert.Error(t, err)
}

func TestOptimizer_ProgressCallback(t *testing.T) {
	ts := createOptimizerTestSeries()
	var progressCalls []int
	config := OptimizationConfig{
		Method: OptMethodGridSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "threshold", Min: 0, Max: 10, Step: 5},
		},
		ObjectiveFunc: ObjectiveNetProfit,
		MaxWorkers:    1,
		ProgressFunc: func(completed, total int) {
			progressCalls = append(progressCalls, completed)
		},
	}

	opt, _ := NewOptimizer(config)
	_, err := opt.Optimize(ts, thresholdStrategyFactory, defaultOptimizerBTConfig())
	require.NoError(t, err)

	// 3 combinations: 0,5,10 → 3 progress calls
	assert.Equal(t, 3, len(progressCalls))
	assert.Equal(t, 1, progressCalls[0])
	assert.Equal(t, 3, progressCalls[2])
}

func TestOptimizer_MultiParameterGrid(t *testing.T) {
	ts := createOptimizerTestSeries()
	config := OptimizationConfig{
		Method: OptMethodGridSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "a", Min: 0, Max: 1, Step: 1},
			{Name: "b", Min: 0, Max: 1, Step: 1},
		},
		ObjectiveFunc: ObjectiveNetProfit,
		MaxWorkers:    1,
	}

	opt, _ := NewOptimizer(config)
	result, _ := opt.Optimize(
		ts,
		func(params map[string]float64) trading.Strategy {
			return &thresholdStrategy{threshold: int(params["a"] + params["b"])}
		},
		defaultOptimizerBTConfig(),
	)

	// 2x2 = 4 combinations
	assert.Equal(t, 4, result.TotalRuns)
}

func TestObjectiveNetProfit(t *testing.T) {
	result := BacktestResult{NetProfit: decimal.New(100)}
	assert.Equal(t, 100.0, ObjectiveNetProfit(result))
}

func TestObjectiveProfitFactor(t *testing.T) {
	result := BacktestResult{ProfitFactor: decimal.New(2.5)}
	assert.Equal(t, 2.5, ObjectiveProfitFactor(result))
}

func TestObjectiveWinRate(t *testing.T) {
	result := BacktestResult{WinRate: decimal.New(0.6)}
	assert.Equal(t, 0.6, ObjectiveWinRate(result))
}

func TestObjectiveMaxDrawdown(t *testing.T) {
	result := BacktestResult{MaxDrawdown: decimal.New(50)}
	assert.Equal(t, -50.0, ObjectiveMaxDrawdown(result))
}

func TestObjectiveSharpeRatio(t *testing.T) {
	// Low variance, positive mean
	result := BacktestResult{
		Trades: []Trade{
			{ProfitPercent: decimal.New(0.01)},
			{ProfitPercent: decimal.New(0.02)},
			{ProfitPercent: decimal.New(0.015)},
		},
	}
	sharpe := ObjectiveSharpeRatio(result)
	assert.True(t, sharpe > 0)

	// Single trade → zero
	result2 := BacktestResult{Trades: []Trade{{ProfitPercent: decimal.New(0.01)}}}
	assert.Equal(t, 0.0, ObjectiveSharpeRatio(result2))

	// No trades → zero
	assert.Equal(t, 0.0, ObjectiveSharpeRatio(BacktestResult{}))
}

func TestOptimizer_BestConfig(t *testing.T) {
	ts := createOptimizerTestSeries()
	config := OptimizationConfig{
		Method: OptMethodGridSearch,
		ParameterSpaces: []ParameterSpace{
			{Name: "threshold", Min: 0, Max: 20, Step: 5},
		},
		ObjectiveFunc: ObjectiveNetProfit,
		MaxWorkers:    1,
	}

	opt, _ := NewOptimizer(config)
	result, _ := opt.Optimize(ts, thresholdStrategyFactory, defaultOptimizerBTConfig())

	// Best config should be among the evaluated configs
	found := false
	for _, r := range result.AllResults {
		if r.Params["threshold"] == result.BestConfig["threshold"] {
			found = true
			break
		}
	}
	assert.True(t, found, "best config must be one of the evaluated configs")
}
