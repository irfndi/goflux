package indicators

import (
	"strconv"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/telemetry"
)

// FibonacciRetracement holds the swing high and low for calculating
// standard Fibonacci retracement levels.
type FibonacciRetracement struct {
	High decimal.Decimal
	Low  decimal.Decimal
}

// FibonacciRetracementResult holds all standard retracement levels.
type FibonacciRetracementResult struct {
	Level0   decimal.Decimal
	Level236 decimal.Decimal
	Level382 decimal.Decimal
	Level500 decimal.Decimal
	Level618 decimal.Decimal
	Level786 decimal.Decimal
	Level100 decimal.Decimal
}

// NewFibonacciRetracement creates a retracement calculator from explicit
// swing high and low values.
func NewFibonacciRetracement(high, low decimal.Decimal) *FibonacciRetracement {
	return &FibonacciRetracement{High: high, Low: low}
}

// NewFibonacciRetracementFromSeries auto-detects the swing high and low
// over the lookback window ending at index.
func NewFibonacciRetracementFromSeries(s *series.TimeSeries, lookback int, index int) *FibonacciRetracement {
	if s == nil || lookback < 1 {
		return &FibonacciRetracement{}
	}
	length := s.Length()
	if index < 0 || index >= length {
		return &FibonacciRetracement{}
	}
	if index < lookback-1 {
		return &FibonacciRetracement{}
	}
	return &FibonacciRetracement{
		High: highestHigh(s, index, lookback),
		Low:  lowestLow(s, index, lookback),
	}
}

// Levels computes all standard Fibonacci retracement levels.
func (f *FibonacciRetracement) Levels() FibonacciRetracementResult {
	range_ := f.High.Sub(f.Low)
	return FibonacciRetracementResult{
		Level0:   f.Low,
		Level236: f.Low.Add(range_.Mul(decimal.New(0.236))),
		Level382: f.Low.Add(range_.Mul(decimal.New(0.382))),
		Level500: f.Low.Add(range_.Mul(decimal.New(0.5))),
		Level618: f.Low.Add(range_.Mul(decimal.New(0.618))),
		Level786: f.Low.Add(range_.Mul(decimal.New(0.786))),
		Level100: f.High,
	}
}

// fibonacciRetracementIndicator implements the Indicator interface for
// a specific retracement level.
type fibonacciRetracementIndicator struct {
	series   *series.TimeSeries
	lookback int
	level    float64
}

// NewFibonacciRetracementIndicator returns an Indicator that calculates
// a specific Fibonacci retracement level (e.g. 0.618) at each index.
func NewFibonacciRetracementIndicator(s *series.TimeSeries, lookback int, level float64) Indicator {
	if lookback < 1 {
		panic("goflux: Fibonacci retracement lookback must be >= 1")
	}
	telemetry.ReportUsage("FibonacciRetracement", map[string]string{
		"lookback": strconv.Itoa(lookback),
		"level":    strconv.FormatFloat(level, 'f', -1, 64),
	})
	return fibonacciRetracementIndicator{series: s, lookback: lookback, level: level}
}

func (f fibonacciRetracementIndicator) Calculate(index int) decimal.Decimal {
	if f.series == nil {
		return decimal.ZERO
	}
	length := f.series.Length()
	if index < 0 || index >= length || index < f.lookback-1 {
		return decimal.ZERO
	}
	high := highestHigh(f.series, index, f.lookback)
	low := lowestLow(f.series, index, f.lookback)
	range_ := high.Sub(low)
	return low.Add(range_.Mul(decimal.New(f.level)))
}

// FibonacciExtension holds the swing high and low for calculating
// standard Fibonacci extension levels above the high or below the low.
type FibonacciExtension struct {
	High decimal.Decimal
	Low  decimal.Decimal
}

// FibonacciExtensionResult holds standard extension levels.
type FibonacciExtensionResult struct {
	Level1272 decimal.Decimal
	Level1618 decimal.Decimal
	Level2000 decimal.Decimal
	Level2618 decimal.Decimal
	Level3000 decimal.Decimal
	Level4236 decimal.Decimal
}

// NewFibonacciExtension creates an extension calculator from explicit
// swing high and low values.
func NewFibonacciExtension(high, low decimal.Decimal) *FibonacciExtension {
	return &FibonacciExtension{High: high, Low: low}
}

// NewFibonacciExtensionFromSeries auto-detects the swing high and low
// over the lookback window ending at index.
func NewFibonacciExtensionFromSeries(s *series.TimeSeries, lookback int, index int) *FibonacciExtension {
	if s == nil || lookback < 1 {
		return &FibonacciExtension{}
	}
	length := s.Length()
	if index < 0 || index >= length {
		return &FibonacciExtension{}
	}
	if index < lookback-1 {
		return &FibonacciExtension{}
	}
	return &FibonacciExtension{
		High: highestHigh(s, index, lookback),
		Low:  lowestLow(s, index, lookback),
	}
}

// LevelsUp computes extension levels above the swing high.
func (f *FibonacciExtension) LevelsUp() FibonacciExtensionResult {
	range_ := f.High.Sub(f.Low)
	return FibonacciExtensionResult{
		Level1272: f.High.Add(range_.Mul(decimal.New(0.272))),
		Level1618: f.High.Add(range_.Mul(decimal.New(0.618))),
		Level2000: f.High.Add(range_.Mul(decimal.New(1.0))),
		Level2618: f.High.Add(range_.Mul(decimal.New(1.618))),
		Level3000: f.High.Add(range_.Mul(decimal.New(2.0))),
		Level4236: f.High.Add(range_.Mul(decimal.New(3.236))),
	}
}

// LevelsDown computes extension levels below the swing low.
func (f *FibonacciExtension) LevelsDown() FibonacciExtensionResult {
	range_ := f.High.Sub(f.Low)
	return FibonacciExtensionResult{
		Level1272: f.Low.Sub(range_.Mul(decimal.New(0.272))),
		Level1618: f.Low.Sub(range_.Mul(decimal.New(0.618))),
		Level2000: f.Low.Sub(range_.Mul(decimal.New(1.0))),
		Level2618: f.Low.Sub(range_.Mul(decimal.New(1.618))),
		Level3000: f.Low.Sub(range_.Mul(decimal.New(2.0))),
		Level4236: f.Low.Sub(range_.Mul(decimal.New(3.236))),
	}
}
