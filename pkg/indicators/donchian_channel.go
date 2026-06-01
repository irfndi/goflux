package indicators

import (
	"strconv"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/telemetry"
)

// donchianUpperBand returns the highest high over the window.
type donchianUpperBand struct {
	series *series.TimeSeries
	window int
}

// NewDonchianUpperBandIndicator returns the Donchian Channel upper band.
// Panics if window < 1.
func NewDonchianUpperBandIndicator(s *series.TimeSeries, window int) Indicator {
	if window < 1 {
		panic("goflux: Donchian Channel window must be >= 1")
	}
	telemetry.ReportUsage("DonchianUpperBand", map[string]string{"window": strconv.Itoa(window)})
	return donchianUpperBand{series: s, window: window}
}

func (d donchianUpperBand) Calculate(index int) decimal.Decimal {
	if d.series == nil {
		return decimal.ZERO
	}
	length := d.series.Length()
	if index < 0 || index >= length || index < d.window-1 {
		return decimal.ZERO
	}
	return highestHigh(d.series, index, d.window)
}

// donchianLowerBand returns the lowest low over the window.
type donchianLowerBand struct {
	series *series.TimeSeries
	window int
}

// NewDonchianLowerBandIndicator returns the Donchian Channel lower band.
// Panics if window < 1.
func NewDonchianLowerBandIndicator(s *series.TimeSeries, window int) Indicator {
	if window < 1 {
		panic("goflux: Donchian Channel window must be >= 1")
	}
	telemetry.ReportUsage("DonchianLowerBand", map[string]string{"window": strconv.Itoa(window)})
	return donchianLowerBand{series: s, window: window}
}

func (d donchianLowerBand) Calculate(index int) decimal.Decimal {
	if d.series == nil {
		return decimal.ZERO
	}
	length := d.series.Length()
	if index < 0 || index >= length || index < d.window-1 {
		return decimal.ZERO
	}
	return lowestLow(d.series, index, d.window)
}

// donchianMiddleBand returns the midpoint between highest high and lowest low.
type donchianMiddleBand struct {
	upper Indicator
	lower Indicator
}

// NewDonchianMiddleBandIndicator returns the Donchian Channel middle band.
func NewDonchianMiddleBandIndicator(s *series.TimeSeries, window int) Indicator {
	return donchianMiddleBand{
		upper: NewDonchianUpperBandIndicator(s, window),
		lower: NewDonchianLowerBandIndicator(s, window),
	}
}

func (d donchianMiddleBand) Calculate(index int) decimal.Decimal {
	up := d.upper.Calculate(index)
	lo := d.lower.Calculate(index)
	return up.Add(lo).Div(decimal.New(2))
}
