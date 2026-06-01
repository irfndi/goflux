package backtest

import (
	"strconv"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/telemetry"
	"github.com/irfndi/goflux/pkg/trading"
)

// WFAConfig configures the walk-forward analysis parameters.
type WFAConfig struct {
	InSampleWindowSize    int // Number of candles for in-sample (training)
	OutOfSampleWindowSize int // Number of candles for out-of-sample (testing)
	StepSize              int // How many candles to roll forward each window
}

// Validate checks that the WFA configuration is sensible.
func (c WFAConfig) Validate() error {
	if c.InSampleWindowSize < 10 {
		return ErrInvalidWFAConfig("in-sample window must be >= 10")
	}
	if c.OutOfSampleWindowSize < 1 {
		return ErrInvalidWFAConfig("out-of-sample window must be >= 1")
	}
	if c.StepSize < 1 {
		return ErrInvalidWFAConfig("step size must be >= 1")
	}
	return nil
}

// ErrInvalidWFAConfig is returned when WFAConfig fails validation.
type ErrInvalidWFAConfig string

func (e ErrInvalidWFAConfig) Error() string { return string(e) }

// WFWindowResult holds the backtest results for a single WFA window.
type WFWindowResult struct {
	WindowIndex       int
	InSampleStart     int
	InSampleEnd       int
	OutOfSampleStart  int
	OutOfSampleEnd    int
	InSampleResult    BacktestResult
	OutOfSampleResult BacktestResult
}

// WFAAggregateMetrics summarizes performance across all WFA windows.
type WFAAggregateMetrics struct {
	TotalWindows             int
	AverageInSampleSharpe    decimal.Decimal
	AverageOutOfSampleSharpe decimal.Decimal
	DegradationRate          decimal.Decimal // (IS - OOS) / |IS|, lower is better
	WinningWindowsPercent    decimal.Decimal // % of OOS windows with positive net profit
	AverageInSampleProfit    decimal.Decimal
	AverageOutOfSampleProfit decimal.Decimal
}

// WFAResult is the complete output of a walk-forward analysis.
type WFAResult struct {
	Windows          []WFWindowResult
	AggregateMetrics WFAAggregateMetrics
}

// StrategyFactory creates a fresh strategy instance bound to the given series.
// Used by WalkForwardAnalyzer to prevent indicator state leakage between
// in-sample and out-of-sample runs.
type StrategyFactory func(ts *series.TimeSeries) trading.Strategy

// OptimizationFunc is a user-provided function that optimizes strategy
// configuration on the given in-sample time series and returns a factory
// that can build a fresh strategy for any sub-series.
type OptimizationFunc func(ts *series.TimeSeries) (StrategyFactory, BacktestConfig)

// WalkForwardAnalyzer runs walk-forward analysis on a time series.
type WalkForwardAnalyzer struct {
	config WFAConfig
}

// NewWalkForwardAnalyzer creates a new WFA analyzer.
func NewWalkForwardAnalyzer(config WFAConfig) (*WalkForwardAnalyzer, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	telemetry.ReportUsage("WalkForwardAnalysis", map[string]string{
		"in_sample":  strconv.Itoa(config.InSampleWindowSize),
		"out_sample": strconv.Itoa(config.OutOfSampleWindowSize),
		"step":       strconv.Itoa(config.StepSize),
	})
	return &WalkForwardAnalyzer{config: config}, nil
}

// Run executes walk-forward analysis on the provided time series.
func (wfa *WalkForwardAnalyzer) Run(
	ts *series.TimeSeries,
	optimize OptimizationFunc,
) (*WFAResult, error) {
	if ts == nil || ts.Length() == 0 {
		return &WFAResult{}, nil
	}

	n := ts.Length()
	cfg := wfa.config

	results := &WFAResult{
		Windows: make([]WFWindowResult, 0),
	}

	// Roll windows forward
	for isStart := 0; isStart+cfg.InSampleWindowSize+cfg.OutOfSampleWindowSize <= n; isStart += cfg.StepSize {
		isEnd := isStart + cfg.InSampleWindowSize
		oosStart := isEnd
		oosEnd := oosStart + cfg.OutOfSampleWindowSize

		isSeries := sliceTimeSeries(ts, isStart, isEnd)
		oosSeries := sliceTimeSeries(ts, oosStart, oosEnd)

		// Optimize on in-sample data; factory prevents indicator state leakage
		strategyFactory, btConfig := optimize(isSeries)

		// Run in-sample backtest with a fresh strategy bound to IS data
		isBacktester := NewBacktester(isSeries, strategyFactory(isSeries))
		isResult := isBacktester.Run(btConfig)

		// Run out-of-sample backtest with a fresh strategy bound to OOS data
		oosBacktester := NewBacktester(oosSeries, strategyFactory(oosSeries))
		oosResult := oosBacktester.Run(btConfig)

		results.Windows = append(results.Windows, WFWindowResult{
			WindowIndex:       len(results.Windows),
			InSampleStart:     isStart,
			InSampleEnd:       isEnd,
			OutOfSampleStart:  oosStart,
			OutOfSampleEnd:    oosEnd,
			InSampleResult:    isResult,
			OutOfSampleResult: oosResult,
		})
	}

	results.AggregateMetrics = wfa.computeAggregateMetrics(results.Windows)
	return results, nil
}

// computeAggregateMetrics calculates summary statistics across all windows.
func (wfa *WalkForwardAnalyzer) computeAggregateMetrics(windows []WFWindowResult) WFAAggregateMetrics {
	if len(windows) == 0 {
		return WFAAggregateMetrics{}
	}

	var (
		totalISSharpe  decimal.Decimal
		totalOOSSharpe decimal.Decimal
		totalISProfit  decimal.Decimal
		totalOOSProfit decimal.Decimal
		winningWindows int
	)

	for _, w := range windows {
		isSharpe := w.InSampleResult.SharpeRatio
		oosSharpe := w.OutOfSampleResult.SharpeRatio

		totalISSharpe = totalISSharpe.Add(isSharpe)
		totalOOSSharpe = totalOOSSharpe.Add(oosSharpe)
		totalISProfit = totalISProfit.Add(w.InSampleResult.NetProfit)
		totalOOSProfit = totalOOSProfit.Add(w.OutOfSampleResult.NetProfit)

		if w.OutOfSampleResult.NetProfit.IsPositive() {
			winningWindows++
		}
	}

	count := decimal.New(float64(len(windows)))
	avgISSharpe := totalISSharpe.Div(count)
	avgOOSSharpe := totalOOSSharpe.Div(count)
	avgISProfit := totalISProfit.Div(count)
	avgOOSProfit := totalOOSProfit.Div(count)

	// Degradation rate = (avg IS - avg OOS) / avg IS
	var degradation decimal.Decimal
	if !avgISSharpe.IsZero() {
		degradation = avgISSharpe.Sub(avgOOSSharpe).Div(avgISSharpe.Abs())
	}

	winPct := decimal.New(float64(winningWindows)).Div(count).Mul(decimal.New(100))

	return WFAAggregateMetrics{
		TotalWindows:             len(windows),
		AverageInSampleSharpe:    avgISSharpe,
		AverageOutOfSampleSharpe: avgOOSSharpe,
		DegradationRate:          degradation,
		WinningWindowsPercent:    winPct,
		AverageInSampleProfit:    avgISProfit,
		AverageOutOfSampleProfit: avgOOSProfit,
	}
}

// sliceTimeSeries creates a new TimeSeries containing candles [start, end).
// Accesses Candles directly because the source series is not modified
// concurrently during backtesting, avoiding per-iteration lock overhead.
func sliceTimeSeries(ts *series.TimeSeries, start, end int) *series.TimeSeries {
	result := series.NewTimeSeries()
	candles := ts.Candles
	length := len(candles)
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	for i := start; i < end; i++ {
		if c := candles[i]; c != nil {
			result.AddCandle(c)
		}
	}
	return result
}
