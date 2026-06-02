package backtest

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/trading"
)

// OptimizationMethod defines the parameter search strategy.
type OptimizationMethod string

const (
	// OptMethodGridSearch exhaustively searches the discretized parameter space.
	OptMethodGridSearch OptimizationMethod = "grid_search"
	// OptMethodRandomSearch randomly samples within parameter bounds.
	OptMethodRandomSearch OptimizationMethod = "random_search"
)

// ParameterSpace defines the range and step for a single parameter.
type ParameterSpace struct {
	Name string
	Min  float64
	Max  float64
	Step float64 // For grid search; must be > 0
}

// Validate checks that the parameter space is valid.
func (ps ParameterSpace) Validate() error {
	if ps.Name == "" {
		return errors.New("parameter space name cannot be empty")
	}
	if ps.Min > ps.Max {
		return fmt.Errorf("parameter %q: min (%g) cannot be greater than max (%g)", ps.Name, ps.Min, ps.Max)
	}
	return nil
}

// OptimizationConfig configures the optimization run.
type OptimizationConfig struct {
	Method          OptimizationMethod
	ParameterSpaces []ParameterSpace
	ObjectiveFunc   ObjectiveFunction
	RandomSamples   int // Total random samples for random search
	MaxWorkers      int // Parallelism; 0 means sequential
	ProgressFunc    func(completed, total int)
	Seed            int64 // For reproducible random search; 0 means time-based
}

// Validate checks that the optimization configuration is valid.
func (c OptimizationConfig) Validate() error {
	if c.Method != OptMethodGridSearch && c.Method != OptMethodRandomSearch {
		return fmt.Errorf("unsupported optimization method: %q", c.Method)
	}
	if len(c.ParameterSpaces) == 0 {
		return errors.New("at least one parameter space is required")
	}
	for _, ps := range c.ParameterSpaces {
		if err := ps.Validate(); err != nil {
			return err
		}
	}
	if c.Method == OptMethodGridSearch {
		for _, ps := range c.ParameterSpaces {
			if ps.Step <= 0 {
				return fmt.Errorf("parameter %q: step must be > 0 for grid search", ps.Name)
			}
		}
	}
	if c.Method == OptMethodRandomSearch && c.RandomSamples <= 0 {
		return errors.New("random_samples must be > 0 for random search")
	}
	if c.ObjectiveFunc == nil {
		return errors.New("objective function is required")
	}
	return nil
}

// ObjectiveFunction scores a backtest result. Higher is better.
type ObjectiveFunction func(result BacktestResult) float64

// OptimizationResult holds the outcome of a parameter optimization run.
type OptimizationResult struct {
	BestConfig map[string]float64
	BestScore  float64
	AllResults []ParameterSetResult
	Duration   time.Duration
	TotalRuns  int
}

// ParameterSetResult holds the score for a single parameter combination.
type ParameterSetResult struct {
	Params map[string]float64
	Score  float64
	Result BacktestResult
}

// Optimizer runs parameter optimization over a backtest strategy.
type Optimizer struct {
	config OptimizationConfig
	rng    *rand.Rand
}

// NewOptimizer creates a new optimizer with the given configuration.
func NewOptimizer(config OptimizationConfig) (*Optimizer, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	seed := config.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &Optimizer{
		config: config,
		// #nosec G404 -- math/rand is sufficient and significantly faster than
		// crypto/rand for parameter optimization sampling.
		rng: rand.New(rand.NewSource(seed)),
	}, nil
}

// Optimize searches for the best strategy parameters on the given time series.
func (o *Optimizer) Optimize(
	ts *series.TimeSeries,
	strategyFactory func(params map[string]float64) trading.Strategy,
	btConfig BacktestConfig,
) (*OptimizationResult, error) {
	if ts == nil || ts.Length() == 0 {
		return &OptimizationResult{}, nil
	}

	startTime := time.Now()
	combinations := o.generateCombinations()
	total := len(combinations)
	if total == 0 {
		return &OptimizationResult{}, nil
	}

	results := make([]ParameterSetResult, 0, total)
	var bestScore float64
	var bestConfig map[string]float64

	workers := o.config.MaxWorkers
	if workers <= 0 {
		workers = 1
	}
	if workers > total {
		workers = total
	}

	if workers == 1 {
		// Sequential execution avoids goroutine overhead.
		for i, params := range combinations {
			result := o.runBacktest(ts, strategyFactory, btConfig, params)
			score := o.config.ObjectiveFunc(result)
			results = append(results, ParameterSetResult{Params: params, Score: score, Result: result})
			if i == 0 || score > bestScore {
				bestScore = score
				bestConfig = copyParams(params)
			}
			if o.config.ProgressFunc != nil {
				o.config.ProgressFunc(i+1, total)
			}
		}
	} else {
		// Worker pool for parallel execution.
		jobs := make(chan map[string]float64)
		resCh := make(chan ParameterSetResult)
		var wg sync.WaitGroup
		var completed atomic.Int32

		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for params := range jobs {
					result := o.runBacktest(ts, strategyFactory, btConfig, params)
					score := o.config.ObjectiveFunc(result)
					resCh <- ParameterSetResult{Params: params, Score: score, Result: result}
					if o.config.ProgressFunc != nil {
						c := int(completed.Add(1))
						o.config.ProgressFunc(c, total)
					}
				}
			}()
		}

		// Close results channel when all workers are done.
		go func() {
			wg.Wait()
			close(resCh)
		}()

		// Send jobs in a goroutine so main can collect results concurrently.
		go func() {
			for _, params := range combinations {
				jobs <- params
			}
			close(jobs)
		}()

		// Collect results.
		for r := range resCh {
			results = append(results, r)
			if len(results) == 1 || r.Score > bestScore {
				bestScore = r.Score
				bestConfig = copyParams(r.Params)
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return &OptimizationResult{
		BestConfig: bestConfig,
		BestScore:  bestScore,
		AllResults: results,
		Duration:   time.Since(startTime),
		TotalRuns:  total,
	}, nil
}

func (o *Optimizer) generateCombinations() []map[string]float64 {
	switch o.config.Method {
	case OptMethodRandomSearch:
		return o.generateRandomSamples()
	default:
		return o.generateGridCombinations()
	}
}

func (o *Optimizer) generateGridCombinations() []map[string]float64 {
	spaces := o.config.ParameterSpaces
	// Generate values for each parameter space.
	valueLists := make([][]float64, len(spaces))
	for i, ps := range spaces {
		values := make([]float64, 0)
		for v := ps.Min; v <= ps.Max+1e-9; v += ps.Step {
			values = append(values, v)
		}
		valueLists[i] = values
	}

	// Compute total combinations.
	total := 1
	for _, values := range valueLists {
		total *= len(values)
	}
	if total == 0 {
		return nil
	}

	// Cartesian product.
	combinations := make([]map[string]float64, 0, total)
	indices := make([]int, len(spaces))
	for {
		params := make(map[string]float64, len(spaces))
		for i, ps := range spaces {
			params[ps.Name] = valueLists[i][indices[i]]
		}
		combinations = append(combinations, params)

		// Increment indices.
		pos := len(spaces) - 1
		for pos >= 0 {
			indices[pos]++
			if indices[pos] < len(valueLists[pos]) {
				break
			}
			indices[pos] = 0
			pos--
		}
		if pos < 0 {
			break
		}
	}
	return combinations
}

func (o *Optimizer) generateRandomSamples() []map[string]float64 {
	spaces := o.config.ParameterSpaces
	count := o.config.RandomSamples
	combinations := make([]map[string]float64, 0, count)
	for i := 0; i < count; i++ {
		params := make(map[string]float64, len(spaces))
		for _, ps := range spaces {
			params[ps.Name] = ps.Min + o.rng.Float64()*(ps.Max-ps.Min)
		}
		combinations = append(combinations, params)
	}
	return combinations
}

func (o *Optimizer) runBacktest(
	ts *series.TimeSeries,
	strategyFactory func(params map[string]float64) trading.Strategy,
	btConfig BacktestConfig,
	params map[string]float64,
) BacktestResult {
	strategy := strategyFactory(params)
	backtester := NewBacktester(ts, strategy)
	return backtester.Run(btConfig)
}

func copyParams(src map[string]float64) map[string]float64 {
	dst := make(map[string]float64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// --- Built-in objective functions ---

// ObjectiveNetProfit returns the net profit as a score.
func ObjectiveNetProfit(result BacktestResult) float64 {
	return result.NetProfit.Float()
}

// ObjectiveProfitFactor returns the profit factor as a score.
func ObjectiveProfitFactor(result BacktestResult) float64 {
	return result.ProfitFactor.Float()
}

// ObjectiveWinRate returns the win rate as a score.
func ObjectiveWinRate(result BacktestResult) float64 {
	return result.WinRate.Float()
}

// ObjectiveMaxDrawdown returns a score where lower drawdown is better.
// The raw drawdown is negated so that maximizing the score minimizes drawdown.
func ObjectiveMaxDrawdown(result BacktestResult) float64 {
	return -result.MaxDrawdown.Float()
}

// ObjectiveSharpeRatio computes a simple Sharpe-like ratio from trade profit percentages.
func ObjectiveSharpeRatio(result BacktestResult) float64 {
	if len(result.Trades) < 2 {
		return 0
	}

	sum := 0.0
	for _, trade := range result.Trades {
		sum += trade.ProfitPercent.Float()
	}
	mean := sum / float64(len(result.Trades))

	variance := 0.0
	for _, trade := range result.Trades {
		diff := trade.ProfitPercent.Float() - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(result.Trades)-1))
	if stdDev == 0 {
		return 0
	}

	return mean / stdDev
}

// NumCPU returns the number of logical CPUs available.
func NumCPU() int {
	return runtime.NumCPU()
}
