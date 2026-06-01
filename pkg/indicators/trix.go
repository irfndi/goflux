package indicators

import (
	"strconv"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/telemetry"
)

// TRIX calculates the Triple Exponential Moving Average oscillator.
// It shows the percent rate of change of a triple-smoothed EMA.
// Values oscillate around zero; positive indicates upward momentum,
// negative indicates downward momentum.
type trixIndicator struct {
	tripleEMA Indicator
	window    int
}

// NewTRIXIndicator returns a TRIX indicator.
// The indicator is calculated as:
//
//	TRIX = ((current TripleEMA - previous TripleEMA) / previous TripleEMA) * 100
//
// Panics if window < 1.
func NewTRIXIndicator(indicator Indicator, window int) Indicator {
	if window < 1 {
		panic("goflux: TRIX window must be >= 1")
	}
	telemetry.ReportUsage("TRIX", map[string]string{"window": strconv.Itoa(window)})
	ema1 := NewEMAIndicator(indicator, window)
	ema2 := NewEMAIndicator(ema1, window)
	ema3 := NewEMAIndicator(ema2, window)
	return trixIndicator{tripleEMA: ema3, window: window}
}

// NewTRIXIndicatorFromSeries returns a TRIX built from a time series.
// Panics if window < 1.
func NewTRIXIndicatorFromSeries(s *series.TimeSeries, window int) Indicator {
	return NewTRIXIndicator(NewClosePriceIndicator(s), window)
}

func (t trixIndicator) Calculate(index int) decimal.Decimal {
	if index < 3*t.window {
		return decimal.ZERO
	}

	current := t.tripleEMA.Calculate(index)
	previous := t.tripleEMA.Calculate(index - 1)

	if previous.IsZero() {
		return decimal.ZERO
	}

	return current.Sub(previous).Div(previous).Mul(decimal.New(100))
}
