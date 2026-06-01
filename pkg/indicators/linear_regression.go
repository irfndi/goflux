package indicators

import (
	"math"
	"strconv"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/telemetry"
)

// linearRegressionBase holds the shared regression computation.
type linearRegressionBase struct {
	indicator Indicator
	window    int
}

// computeRegression calculates slope, intercept, forecast and standard error
// for the window ending at index.
func (l linearRegressionBase) computeRegression(index int) (slope, intercept, forecast, stdErr decimal.Decimal) {
	if l.indicator == nil || l.window < 2 || index < l.window-1 {
		return decimal.ZERO, decimal.ZERO, decimal.ZERO, decimal.ZERO
	}

	n := l.window
	start := index - n + 1

	// Cache indicator values to avoid redundant Calculate calls.
	values := make([]decimal.Decimal, n)
	sumY := decimal.ZERO
	sumXY := decimal.ZERO
	for i := 0; i < n; i++ {
		v := l.indicator.Calculate(start + i)
		values[i] = v
		sumY = sumY.Add(v)
		sumXY = sumXY.Add(v.Mul(decimal.New(float64(i))))
	}

	// sumX and sumX2 are constants for x = 0..n-1.
	sumX := n * (n - 1) / 2
	sumX2 := n * (n - 1) * (2*n - 1) / 6

	sumXDec := decimal.New(float64(sumX))
	sumX2Dec := decimal.New(float64(sumX2))
	nDec := decimal.New(float64(n))

	denominator := nDec.Mul(sumX2Dec).Sub(sumXDec.Mul(sumXDec))
	if denominator.Zero() {
		return decimal.ZERO, decimal.ZERO, decimal.ZERO, decimal.ZERO
	}

	slopeDec := nDec.Mul(sumXY).Sub(sumXDec.Mul(sumY)).Div(denominator)
	interceptDec := sumY.Sub(slopeDec.Mul(sumXDec)).Div(nDec)
	forecastDec := slopeDec.Mul(decimal.New(float64(n - 1))).Add(interceptDec)

	// Standard error using cached values.
	ssResidual := decimal.ZERO
	for i := 0; i < n; i++ {
		x := decimal.New(float64(i))
		yHat := slopeDec.Mul(x).Add(interceptDec)
		residual := values[i].Sub(yHat)
		ssResidual = ssResidual.Add(residual.Mul(residual))
	}

	var stdErrDec decimal.Decimal
	if n > 2 {
		stdErrDec = ssResidual.Div(decimal.New(float64(n - 2))).Sqrt()
	} else {
		stdErrDec = decimal.ZERO
	}

	return slopeDec, interceptDec, forecastDec, stdErrDec
}

// linearRegressionIndicator returns the forecasted value (linear regression
// estimate at the current index).
type linearRegressionIndicator struct {
	linearRegressionBase
}

// NewLinearRegressionIndicator returns an Indicator that calculates the
// linear regression forecast at each index.
func NewLinearRegressionIndicator(indicator Indicator, window int) Indicator {
	if window < 2 {
		panic("goflux: Linear Regression window must be >= 2")
	}
	telemetry.ReportUsage("LinearRegression", map[string]string{"window": strconv.Itoa(window)})
	return linearRegressionIndicator{linearRegressionBase{indicator: indicator, window: window}}
}

func (l linearRegressionIndicator) Calculate(index int) decimal.Decimal {
	_, _, forecast, _ := l.computeRegression(index)
	return forecast
}

// linearRegressionSlopeIndicator returns the slope of the regression line.
type linearRegressionSlopeIndicator struct {
	linearRegressionBase
}

// NewLinearRegressionSlopeIndicator returns an Indicator for the slope.
func NewLinearRegressionSlopeIndicator(indicator Indicator, window int) Indicator {
	if window < 2 {
		panic("goflux: Linear Regression Slope window must be >= 2")
	}
	telemetry.ReportUsage("LinearRegressionSlope", map[string]string{"window": strconv.Itoa(window)})
	return linearRegressionSlopeIndicator{linearRegressionBase{indicator: indicator, window: window}}
}

func (l linearRegressionSlopeIndicator) Calculate(index int) decimal.Decimal {
	slope, _, _, _ := l.computeRegression(index)
	return slope
}

// linearRegressionInterceptIndicator returns the y-intercept of the regression line.
type linearRegressionInterceptIndicator struct {
	linearRegressionBase
}

// NewLinearRegressionInterceptIndicator returns an Indicator for the intercept.
func NewLinearRegressionInterceptIndicator(indicator Indicator, window int) Indicator {
	if window < 2 {
		panic("goflux: Linear Regression Intercept window must be >= 2")
	}
	telemetry.ReportUsage("LinearRegressionIntercept", map[string]string{"window": strconv.Itoa(window)})
	return linearRegressionInterceptIndicator{linearRegressionBase{indicator: indicator, window: window}}
}

func (l linearRegressionInterceptIndicator) Calculate(index int) decimal.Decimal {
	_, intercept, _, _ := l.computeRegression(index)
	return intercept
}

// linearRegressionAngleIndicator returns the angle of the regression line in degrees.
type linearRegressionAngleIndicator struct {
	linearRegressionBase
}

// NewLinearRegressionAngleIndicator returns an Indicator for the angle in degrees.
func NewLinearRegressionAngleIndicator(indicator Indicator, window int) Indicator {
	if window < 2 {
		panic("goflux: Linear Regression Angle window must be >= 2")
	}
	telemetry.ReportUsage("LinearRegressionAngle", map[string]string{"window": strconv.Itoa(window)})
	return linearRegressionAngleIndicator{linearRegressionBase{indicator: indicator, window: window}}
}

func (l linearRegressionAngleIndicator) Calculate(index int) decimal.Decimal {
	slope, _, _, _ := l.computeRegression(index)
	if slope.Zero() {
		return decimal.ZERO
	}
	angle := math.Atan(slope.Float()) * 180.0 / math.Pi
	return decimal.New(angle)
}

// standardErrorIndicator returns the standard error of the regression estimate.
type standardErrorIndicator struct {
	linearRegressionBase
}

// NewStandardErrorIndicator returns an Indicator for the standard error.
func NewStandardErrorIndicator(indicator Indicator, window int) Indicator {
	if window < 2 {
		panic("goflux: Standard Error window must be >= 2")
	}
	telemetry.ReportUsage("StandardError", map[string]string{"window": strconv.Itoa(window)})
	return standardErrorIndicator{linearRegressionBase{indicator: indicator, window: window}}
}

func (l standardErrorIndicator) Calculate(index int) decimal.Decimal {
	_, _, _, stdErr := l.computeRegression(index)
	return stdErr
}

// linearRegressionChannelUpper returns the upper band (forecast + k * SE).
type linearRegressionChannelUpper struct {
	mid        Indicator
	stdErr     Indicator
	deviations decimal.Decimal
}

// linearRegressionChannelLower returns the lower band (forecast - k * SE).
type linearRegressionChannelLower struct {
	mid        Indicator
	stdErr     Indicator
	deviations decimal.Decimal
}

// NewLinearRegressionChannel returns mid, upper, and lower band indicators
// for a Linear Regression Channel.
func NewLinearRegressionChannel(indicator Indicator, window int, deviations float64) (mid, upper, lower Indicator) {
	if window < 2 {
		panic("goflux: Linear Regression Channel window must be >= 2")
	}
	mid = NewLinearRegressionIndicator(indicator, window)
	stdErr := NewStandardErrorIndicator(indicator, window)
	k := decimal.New(deviations)
	return mid,
		linearRegressionChannelUpper{mid: mid, stdErr: stdErr, deviations: k},
		linearRegressionChannelLower{mid: mid, stdErr: stdErr, deviations: k}
}

func (l linearRegressionChannelUpper) Calculate(index int) decimal.Decimal {
	return l.mid.Calculate(index).Add(l.stdErr.Calculate(index).Mul(l.deviations))
}

func (l linearRegressionChannelLower) Calculate(index int) decimal.Decimal {
	return l.mid.Calculate(index).Sub(l.stdErr.Calculate(index).Mul(l.deviations))
}
