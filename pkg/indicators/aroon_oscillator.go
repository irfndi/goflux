package indicators

import (
	"strconv"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/telemetry"
)

// AroonOscillator calculates the difference between Aroon Up and Aroon Down.
// Values range between -100 and +100.
// +100 indicates a strong uptrend, -100 a strong downtrend, 0 no trend.
type aroonOscillator struct {
	aroonUp   Indicator
	aroonDown Indicator
}

// NewAroonOscillator returns an Aroon Oscillator indicator.
// The oscillator is calculated as: AroonUp - AroonDown.
// Panics if window < 1.
func NewAroonOscillator(indicator Indicator, window int) Indicator {
	if window < 1 {
		panic("goflux: Aroon Oscillator window must be >= 1")
	}
	telemetry.ReportUsage("AroonOscillator", map[string]string{"window": strconv.Itoa(window)})
	return aroonOscillator{
		aroonUp:   NewAroonUpIndicator(indicator, window),
		aroonDown: NewAroonDownIndicator(indicator, window),
	}
}

// NewAroonOscillatorFromSeries returns an Aroon Oscillator built from a time series.
// It automatically creates the high and low price indicators internally.
// Panics if window < 1.
func NewAroonOscillatorFromSeries(s *series.TimeSeries, window int) Indicator {
	if window < 1 {
		panic("goflux: Aroon Oscillator window must be >= 1")
	}
	telemetry.ReportUsage("AroonOscillator", map[string]string{"window": strconv.Itoa(window)})
	return aroonOscillator{
		aroonUp:   NewAroonUpIndicator(NewHighPriceIndicator(s), window),
		aroonDown: NewAroonDownIndicator(NewLowPriceIndicator(s), window),
	}
}

func (ao aroonOscillator) Calculate(index int) decimal.Decimal {
	up := ao.aroonUp.Calculate(index)
	down := ao.aroonDown.Calculate(index)

	if up.IsZero() && down.IsZero() {
		return decimal.ZERO
	}

	return up.Sub(down)
}
