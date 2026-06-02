package backtest

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/metrics"
)

// MCMethod defines the Monte Carlo simulation method.
type MCMethod string

const (
	// MCMethodTradeShuffle randomly reorders trades.
	MCMethodTradeShuffle MCMethod = "trade_shuffle"
	// MCMethodBootstrap resamples trades with replacement.
	MCMethodBootstrap MCMethod = "bootstrap"
	// MCMethodRandomStart takes a random contiguous subset of trades.
	MCMethodRandomStart MCMethod = "random_start"
)

// MCSimulationConfig configures the Monte Carlo simulation.
type MCSimulationConfig struct {
	// Simulations is the number of Monte Carlo runs. Default 10000.
	Simulations int
	// ConfidenceLevel is the confidence level for intervals, e.g. 0.95.
	ConfidenceLevel float64
	// Method is the simulation method.
	Method MCMethod
	// Seed is the random seed. Zero means time-based.
	Seed int64
}

// DefaultMCSimulationConfig returns a default configuration.
func DefaultMCSimulationConfig() MCSimulationConfig {
	return MCSimulationConfig{
		Simulations:     10000,
		ConfidenceLevel: 0.95,
		Method:          MCMethodTradeShuffle,
	}
}

// MCStats holds aggregate statistics across simulations.
type MCStats struct {
	Mean   decimal.Decimal
	Median decimal.Decimal
	StdDev decimal.Decimal
	Min    decimal.Decimal
	Max    decimal.Decimal
}

// MCSimulationResult holds the outcome of a Monte Carlo simulation.
type MCSimulationResult struct {
	SimulatedEquityCurves          [][]metrics.EquityPoint
	FinalEquityStats               MCStats
	MaxDrawdownStats               MCStats
	SharpeStats                    MCStats
	BelowInitialCapitalProbability float64
	Percentiles                    map[string]map[float64]decimal.Decimal
}

// MonteCarloSimulator runs Monte Carlo simulations on backtest results.
type MonteCarloSimulator struct {
	config MCSimulationConfig
	rng    *rand.Rand
}

// NewMonteCarloSimulator creates a new simulator with the given config.
func NewMonteCarloSimulator(config MCSimulationConfig) *MonteCarloSimulator {
	seed := config.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &MonteCarloSimulator{
		config: config,
		// #nosec G404 -- math/rand is sufficient and significantly faster than
		// crypto/rand for Monte Carlo simulation of backtest trades.
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Run executes the Monte Carlo simulation on the provided backtest result.
func (mc *MonteCarloSimulator) Run(result BacktestResult) (*MCSimulationResult, error) {
	simulations := mc.config.Simulations
	if simulations <= 0 {
		simulations = 10000
	}

	finalEquities := make([]decimal.Decimal, simulations)
	maxDrawdowns := make([]decimal.Decimal, simulations)
	sharpeRatios := make([]decimal.Decimal, simulations)
	equityCurves := make([][]metrics.EquityPoint, simulations)

	ruinCount := 0

	for i := 0; i < simulations; i++ {
		simTrades := mc.generateSimulatedTrades(result.Trades)
		equityCurve := mc.buildEquityCurve(simTrades, result.InitialCapital)
		equityCurves[i] = equityCurve

		finalEquity := equityCurve[len(equityCurve)-1].Equity
		finalEquities[i] = finalEquity

		if finalEquity.LT(result.InitialCapital) {
			ruinCount++
		}

		maxDrawdowns[i] = mc.calculateMaxDrawdownFromCurve(equityCurve)
		sharpeRatios[i] = mc.calculateSharpeFromTrades(simTrades)
	}

	percentiles := mc.computePercentiles(finalEquities, maxDrawdowns, sharpeRatios)

	return &MCSimulationResult{
		SimulatedEquityCurves:          equityCurves,
		FinalEquityStats:               computeStats(finalEquities),
		MaxDrawdownStats:               computeStats(maxDrawdowns),
		SharpeStats:                    computeStats(sharpeRatios),
		BelowInitialCapitalProbability: float64(ruinCount) / float64(simulations),
		Percentiles:                    percentiles,
	}, nil
}

func (mc *MonteCarloSimulator) generateSimulatedTrades(trades []Trade) []Trade {
	switch mc.config.Method {
	case MCMethodBootstrap:
		return mc.bootstrapTrades(trades)
	case MCMethodRandomStart:
		return mc.randomStartTrades(trades)
	default:
		return mc.shuffleTrades(trades)
	}
}

func (mc *MonteCarloSimulator) shuffleTrades(trades []Trade) []Trade {
	if len(trades) == 0 {
		return nil
	}
	shuffled := make([]Trade, len(trades))
	copy(shuffled, trades)
	mc.rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}

func (mc *MonteCarloSimulator) bootstrapTrades(trades []Trade) []Trade {
	if len(trades) == 0 {
		return nil
	}
	n := len(trades)
	resampled := make([]Trade, n)
	for i := 0; i < n; i++ {
		resampled[i] = trades[mc.rng.Intn(n)]
	}
	return resampled
}

func (mc *MonteCarloSimulator) randomStartTrades(trades []Trade) []Trade {
	if len(trades) == 0 {
		return nil
	}
	n := len(trades)
	start := mc.rng.Intn(n)
	end := start + 1 + mc.rng.Intn(n-start)
	subset := make([]Trade, end-start)
	copy(subset, trades[start:end])
	return subset
}

func (mc *MonteCarloSimulator) buildEquityCurve(trades []Trade, initialCapital decimal.Decimal) []metrics.EquityPoint {
	if len(trades) == 0 {
		return []metrics.EquityPoint{
			{Equity: initialCapital, Drawdown: decimal.ZERO, DrawdownPct: decimal.ZERO},
		}
	}

	curve := make([]metrics.EquityPoint, len(trades))
	equity := initialCapital
	peak := initialCapital

	for i, trade := range trades {
		equity = equity.Add(trade.Profit)
		if equity.GT(peak) {
			peak = equity
		}
		drawdown := peak.Sub(equity)
		var drawdownPct decimal.Decimal
		if !peak.IsZero() {
			drawdownPct = drawdown.Div(peak)
		}
		curve[i] = metrics.EquityPoint{
			Equity:      equity,
			Drawdown:    drawdown,
			DrawdownPct: drawdownPct,
		}
	}

	return curve
}

func (mc *MonteCarloSimulator) calculateMaxDrawdownFromCurve(curve []metrics.EquityPoint) decimal.Decimal {
	maxDD := decimal.ZERO
	for _, point := range curve {
		if point.Drawdown.GT(maxDD) {
			maxDD = point.Drawdown
		}
	}
	return maxDD
}

func (mc *MonteCarloSimulator) calculateSharpeFromTrades(trades []Trade) decimal.Decimal {
	if len(trades) < 2 {
		return decimal.ZERO
	}

	sum := 0.0
	for _, trade := range trades {
		sum += trade.ProfitPercent.Float()
	}
	mean := sum / float64(len(trades))

	variance := 0.0
	for _, trade := range trades {
		diff := trade.ProfitPercent.Float() - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(trades)-1))
	if stdDev == 0 {
		return decimal.ZERO
	}

	return decimal.New(mean / stdDev)
}

func (mc *MonteCarloSimulator) computePercentiles(finalEquities, maxDrawdowns, sharpeRatios []decimal.Decimal) map[string]map[float64]decimal.Decimal {
	result := map[string]map[float64]decimal.Decimal{
		"final_equity": make(map[float64]decimal.Decimal),
		"max_drawdown": make(map[float64]decimal.Decimal),
		"sharpe_ratio": make(map[float64]decimal.Decimal),
	}

	// Standard quintiles
	percentileKeys := []float64{0.05, 0.25, 0.50, 0.75, 0.95}
	for _, p := range percentileKeys {
		result["final_equity"][p] = percentile(finalEquities, p)
		result["max_drawdown"][p] = percentile(maxDrawdowns, p)
		result["sharpe_ratio"][p] = percentile(sharpeRatios, p)
	}

	// Confidence interval bounds derived from config.ConfidenceLevel
	cl := mc.config.ConfidenceLevel
	if cl > 0 && cl < 1 {
		lower := (1 - cl) / 2
		upper := 1 - lower
		result["final_equity"][lower] = percentile(finalEquities, lower)
		result["final_equity"][upper] = percentile(finalEquities, upper)
		result["max_drawdown"][lower] = percentile(maxDrawdowns, lower)
		result["max_drawdown"][upper] = percentile(maxDrawdowns, upper)
		result["sharpe_ratio"][lower] = percentile(sharpeRatios, lower)
		result["sharpe_ratio"][upper] = percentile(sharpeRatios, upper)
	}

	return result
}

func computeStats(values []decimal.Decimal) MCStats {
	if len(values) == 0 {
		return MCStats{}
	}

	sorted := make([]decimal.Decimal, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Float() < sorted[j].Float()
	})

	min := sorted[0]
	max := sorted[len(sorted)-1]

	var sum float64
	for _, v := range values {
		sum += v.Float()
	}
	mean := decimal.New(sum / float64(len(values)))

	median := sorted[len(sorted)/2]
	if len(sorted)%2 == 0 {
		mid1 := sorted[len(sorted)/2-1]
		mid2 := sorted[len(sorted)/2]
		median = decimal.New((mid1.Float() + mid2.Float()) / 2)
	}

	var variance float64
	for _, v := range values {
		diff := v.Float() - mean.Float()
		variance += diff * diff
	}
	// Use sample standard deviation (divide by n-1) for consistency with calculateSharpeFromTrades.
	var stdDev decimal.Decimal
	if len(values) > 1 {
		stdDev = decimal.New(math.Sqrt(variance / float64(len(values)-1)))
	}

	return MCStats{
		Mean:   mean,
		Median: median,
		StdDev: stdDev,
		Min:    min,
		Max:    max,
	}
}

func percentile(values []decimal.Decimal, p float64) decimal.Decimal {
	if len(values) == 0 {
		return decimal.ZERO
	}
	sorted := make([]decimal.Decimal, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Float() < sorted[j].Float()
	})

	index := p * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	if lower == upper {
		return sorted[lower]
	}
	weight := index - float64(lower)
	lowerVal := sorted[lower].Float()
	upperVal := sorted[upper].Float()
	return decimal.New(lowerVal + weight*(upperVal-lowerVal))
}
