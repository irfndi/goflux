package backtest

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleBacktestResult() BacktestResult {
	return BacktestResult{
		InitialCapital: decimal.New(1000),
		FinalEquity:    decimal.New(1050),
		Trades: []Trade{
			{Profit: decimal.New(10), ProfitPercent: decimal.New(0.01)},
			{Profit: decimal.New(-5), ProfitPercent: decimal.New(-0.005)},
			{Profit: decimal.New(20), ProfitPercent: decimal.New(0.02)},
			{Profit: decimal.New(-10), ProfitPercent: decimal.New(-0.01)},
			{Profit: decimal.New(15), ProfitPercent: decimal.New(0.015)},
			{Profit: decimal.New(-5), ProfitPercent: decimal.New(-0.005)},
			{Profit: decimal.New(25), ProfitPercent: decimal.New(0.025)},
			{Profit: decimal.New(-10), ProfitPercent: decimal.New(-0.01)},
			{Profit: decimal.New(10), ProfitPercent: decimal.New(0.01)},
			{Profit: decimal.New(-5), ProfitPercent: decimal.New(-0.005)},
		},
	}
}

func TestMonteCarloSimulator_TradeShuffle(t *testing.T) {
	result := sampleBacktestResult()
	config := DefaultMCSimulationConfig()
	config.Simulations = 1000
	config.Seed = 42

	mc := NewMonteCarloSimulator(config)
	simResult, err := mc.Run(result)
	require.NoError(t, err)
	require.NotNil(t, simResult)

	assert.Equal(t, 1000, len(simResult.SimulatedEquityCurves))
	assert.True(t, simResult.FinalEquityStats.Mean.GT(decimal.ZERO))
	assert.True(t, simResult.MaxDrawdownStats.Mean.GTE(decimal.ZERO))
	assert.True(t, simResult.SharpeStats.Mean.GT(decimal.ZERO))
	assert.GreaterOrEqual(t, simResult.RuinProbability, 0.0)
	assert.LessOrEqual(t, simResult.RuinProbability, 1.0)

	// Percentiles should exist
	assert.NotNil(t, simResult.Percentiles["final_equity"])
	assert.NotNil(t, simResult.Percentiles["max_drawdown"])
	assert.NotNil(t, simResult.Percentiles["sharpe_ratio"])

	// 5th percentile <= 50th percentile <= 95th percentile
	fe5 := simResult.Percentiles["final_equity"][0.05]
	fe50 := simResult.Percentiles["final_equity"][0.50]
	fe95 := simResult.Percentiles["final_equity"][0.95]
	assert.True(t, fe5.LTE(fe50), "5th percentile should be <= 50th")
	assert.True(t, fe50.LTE(fe95), "50th percentile should be <= 95th")
}

func TestMonteCarloSimulator_Bootstrap(t *testing.T) {
	result := sampleBacktestResult()
	config := DefaultMCSimulationConfig()
	config.Simulations = 1000
	config.Method = MCMethodBootstrap
	config.Seed = 42

	mc := NewMonteCarloSimulator(config)
	simResult, err := mc.Run(result)
	require.NoError(t, err)
	require.NotNil(t, simResult)

	assert.Equal(t, 1000, len(simResult.SimulatedEquityCurves))
	assert.True(t, simResult.FinalEquityStats.Mean.GT(decimal.ZERO))
}

func TestMonteCarloSimulator_RandomStart(t *testing.T) {
	result := sampleBacktestResult()
	config := DefaultMCSimulationConfig()
	config.Simulations = 1000
	config.Method = MCMethodRandomStart
	config.Seed = 42

	mc := NewMonteCarloSimulator(config)
	simResult, err := mc.Run(result)
	require.NoError(t, err)
	require.NotNil(t, simResult)

	assert.Equal(t, 1000, len(simResult.SimulatedEquityCurves))
	assert.True(t, simResult.FinalEquityStats.Mean.GT(decimal.ZERO))
}

func TestMonteCarloSimulator_EmptyTrades(t *testing.T) {
	result := BacktestResult{
		InitialCapital: decimal.New(1000),
		Trades:         []Trade{},
	}
	config := DefaultMCSimulationConfig()
	config.Simulations = 100
	config.Seed = 42

	mc := NewMonteCarloSimulator(config)
	simResult, err := mc.Run(result)
	require.NoError(t, err)
	require.NotNil(t, simResult)

	// With no trades, all simulations should have the same final equity
	assert.Equal(t, 0.0, simResult.RuinProbability)
	assert.True(t, simResult.FinalEquityStats.Mean.EQ(decimal.New(1000)))
	assert.True(t, simResult.FinalEquityStats.Min.EQ(decimal.New(1000)))
	assert.True(t, simResult.FinalEquityStats.Max.EQ(decimal.New(1000)))
}

func TestMonteCarloSimulator_RuinScenario(t *testing.T) {
	// All losing trades should guarantee ruin
	result := BacktestResult{
		InitialCapital: decimal.New(1000),
		Trades: []Trade{
			{Profit: decimal.New(-100), ProfitPercent: decimal.New(-0.1)},
			{Profit: decimal.New(-200), ProfitPercent: decimal.New(-0.2)},
			{Profit: decimal.New(-300), ProfitPercent: decimal.New(-0.3)},
		},
	}
	config := DefaultMCSimulationConfig()
	config.Simulations = 100
	config.Seed = 42

	mc := NewMonteCarloSimulator(config)
	simResult, err := mc.Run(result)
	require.NoError(t, err)

	// With all negative trades, any ordering results in ruin
	assert.Equal(t, 1.0, simResult.RuinProbability)
	assert.True(t, simResult.FinalEquityStats.Mean.LT(decimal.New(1000)))
}

func TestMonteCarloSimulator_StatsConsistency(t *testing.T) {
	result := sampleBacktestResult()
	config := DefaultMCSimulationConfig()
	config.Simulations = 5000
	config.Seed = 123

	mc := NewMonteCarloSimulator(config)
	simResult, err := mc.Run(result)
	require.NoError(t, err)

	// Mean should be between min and max
	stats := simResult.FinalEquityStats
	assert.True(t, stats.Min.LTE(stats.Mean), "min should be <= mean")
	assert.True(t, stats.Mean.LTE(stats.Max), "mean should be <= max")

	// Median should be between min and max
	assert.True(t, stats.Min.LTE(stats.Median), "min should be <= median")
	assert.True(t, stats.Median.LTE(stats.Max), "median should be <= max")

	// StdDev should be non-negative
	assert.True(t, stats.StdDev.GTE(decimal.ZERO), "stddev should be >= 0")
}

func TestMonteCarloSimulator_ReproducibleWithSeed(t *testing.T) {
	result := sampleBacktestResult()
	config := DefaultMCSimulationConfig()
	config.Simulations = 100
	config.Seed = 999

	mc1 := NewMonteCarloSimulator(config)
	r1, err := mc1.Run(result)
	require.NoError(t, err)

	mc2 := NewMonteCarloSimulator(config)
	r2, err := mc2.Run(result)
	require.NoError(t, err)

	// With same seed, results should be identical
	assert.Equal(t, r1.RuinProbability, r2.RuinProbability)
	assert.True(t, r1.FinalEquityStats.Mean.EQ(r2.FinalEquityStats.Mean))
	assert.True(t, r1.MaxDrawdownStats.Mean.EQ(r2.MaxDrawdownStats.Mean))
}

func TestMonteCarloSimulator_DifferentSeeds(t *testing.T) {
	result := sampleBacktestResult()
	config := DefaultMCSimulationConfig()
	config.Simulations = 100
	config.Seed = 111

	mc1 := NewMonteCarloSimulator(config)
	r1, _ := mc1.Run(result)

	config.Seed = 222
	mc2 := NewMonteCarloSimulator(config)
	r2, _ := mc2.Run(result)

	// Different seeds should produce different equity paths (with high probability).
	// Final equity for trade_shuffle is commutative, so we compare max drawdown
	// which depends on path ordering.
	assert.False(t, r1.MaxDrawdownStats.Mean.EQ(r2.MaxDrawdownStats.Mean))
}

func TestComputeStats(t *testing.T) {
	values := []decimal.Decimal{
		decimal.New(1),
		decimal.New(2),
		decimal.New(3),
		decimal.New(4),
		decimal.New(5),
	}

	stats := computeStats(values)
	assert.True(t, stats.Min.EQ(decimal.New(1)))
	assert.True(t, stats.Max.EQ(decimal.New(5)))
	assert.True(t, stats.Mean.EQ(decimal.New(3)))
	assert.True(t, stats.Median.EQ(decimal.New(3)))
	assert.True(t, stats.StdDev.GT(decimal.ZERO))
}

func TestPercentile(t *testing.T) {
	values := []decimal.Decimal{
		decimal.New(1),
		decimal.New(2),
		decimal.New(3),
		decimal.New(4),
		decimal.New(5),
	}

	assert.True(t, percentile(values, 0.0).EQ(decimal.New(1)))
	assert.True(t, percentile(values, 1.0).EQ(decimal.New(5)))
	assert.True(t, percentile(values, 0.5).EQ(decimal.New(3)))
}

func TestBuildEquityCurve(t *testing.T) {
	mc := NewMonteCarloSimulator(DefaultMCSimulationConfig())
	trades := []Trade{
		{Profit: decimal.New(10), ProfitPercent: decimal.New(0.01)},
		{Profit: decimal.New(-5), ProfitPercent: decimal.New(-0.005)},
		{Profit: decimal.New(20), ProfitPercent: decimal.New(0.02)},
	}

	curve := mc.buildEquityCurve(trades, decimal.New(100))
	require.Equal(t, 3, len(curve))
	assert.True(t, curve[0].Equity.EQ(decimal.New(110)))
	assert.True(t, curve[1].Equity.EQ(decimal.New(105)))
	assert.True(t, curve[2].Equity.EQ(decimal.New(125)))
	assert.True(t, curve[0].Drawdown.EQ(decimal.ZERO))
	assert.True(t, curve[1].Drawdown.EQ(decimal.New(5)))
}

func TestBuildEquityCurve_Empty(t *testing.T) {
	mc := NewMonteCarloSimulator(DefaultMCSimulationConfig())
	curve := mc.buildEquityCurve([]Trade{}, decimal.New(1000))
	require.Equal(t, 1, len(curve))
	assert.True(t, curve[0].Equity.EQ(decimal.New(1000)))
}

func TestCalculateSharpeFromTrades(t *testing.T) {
	mc := NewMonteCarloSimulator(DefaultMCSimulationConfig())

	// Low variance, positive mean
	trades1 := []Trade{
		{ProfitPercent: decimal.New(0.01)},
		{ProfitPercent: decimal.New(0.02)},
		{ProfitPercent: decimal.New(0.015)},
	}
	sharpe1 := mc.calculateSharpeFromTrades(trades1)
	assert.True(t, sharpe1.GT(decimal.ZERO))

	// Single trade should return zero
	trades2 := []Trade{{ProfitPercent: decimal.New(0.01)}}
	sharpe2 := mc.calculateSharpeFromTrades(trades2)
	assert.True(t, sharpe2.EQ(decimal.ZERO))
}

func TestCalculateMaxDrawdownFromCurve(t *testing.T) {
	mc := NewMonteCarloSimulator(DefaultMCSimulationConfig())
	curve := []metrics.EquityPoint{
		{Equity: decimal.New(100), Drawdown: decimal.ZERO},
		{Equity: decimal.New(120), Drawdown: decimal.ZERO},
		{Equity: decimal.New(90), Drawdown: decimal.New(30)},
		{Equity: decimal.New(110), Drawdown: decimal.New(10)},
		{Equity: decimal.New(80), Drawdown: decimal.New(40)},
	}
	maxDD := mc.calculateMaxDrawdownFromCurve(curve)
	assert.True(t, maxDD.EQ(decimal.New(40)))
}

func TestShuffleTrades(t *testing.T) {
	mc := NewMonteCarloSimulator(DefaultMCSimulationConfig())
	trades := []Trade{
		{Profit: decimal.New(1)},
		{Profit: decimal.New(2)},
		{Profit: decimal.New(3)},
	}
	shuffled := mc.shuffleTrades(trades)
	require.Equal(t, 3, len(shuffled))

	// Sum should be preserved
	sum := decimal.ZERO
	for _, t := range shuffled {
		sum = sum.Add(t.Profit)
	}
	assert.True(t, sum.EQ(decimal.New(6)))
}

func TestBootstrapTrades(t *testing.T) {
	mc := NewMonteCarloSimulator(DefaultMCSimulationConfig())
	trades := []Trade{
		{Profit: decimal.New(1)},
		{Profit: decimal.New(2)},
		{Profit: decimal.New(3)},
	}
	resampled := mc.bootstrapTrades(trades)
	require.Equal(t, 3, len(resampled))
}

func TestRandomStartTrades(t *testing.T) {
	mc := NewMonteCarloSimulator(DefaultMCSimulationConfig())
	trades := []Trade{
		{Profit: decimal.New(1)},
		{Profit: decimal.New(2)},
		{Profit: decimal.New(3)},
		{Profit: decimal.New(4)},
		{Profit: decimal.New(5)},
	}
	subset := mc.randomStartTrades(trades)
	require.True(t, len(subset) >= 1)
	require.True(t, len(subset) <= 5)
}

func TestDefaultMCSimulationConfig(t *testing.T) {
	cfg := DefaultMCSimulationConfig()
	assert.Equal(t, 10000, cfg.Simulations)
	assert.Equal(t, 0.95, cfg.ConfidenceLevel)
	assert.Equal(t, MCMethodTradeShuffle, cfg.Method)
	assert.Equal(t, int64(0), cfg.Seed)
}

func BenchmarkMonteCarloSimulator_TradeShuffle(b *testing.B) {
	result := sampleBacktestResult()
	config := DefaultMCSimulationConfig()
	config.Simulations = 1000
	config.Seed = 42
	mc := NewMonteCarloSimulator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := mc.Run(result); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMonteCarloSimulator_Bootstrap(b *testing.B) {
	result := sampleBacktestResult()
	config := DefaultMCSimulationConfig()
	config.Simulations = 1000
	config.Method = MCMethodBootstrap
	config.Seed = 42
	mc := NewMonteCarloSimulator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := mc.Run(result); err != nil {
			b.Fatal(err)
		}
	}
}
