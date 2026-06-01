package indicators

import (
	"strconv"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/telemetry"
)

// --- Ease of Movement (EOM) ---

// NewEaseOfMovementIndicator returns the Ease of Movement indicator smoothed
// with a simple moving average over the given window.
// Panics if s is nil or window <= 0.
func NewEaseOfMovementIndicator(s *series.TimeSeries, window int) Indicator {
	if s == nil {
		panic("goflux: EOM series cannot be nil")
	}
	if window <= 0 {
		panic("goflux: EOM window must be > 0")
	}
	telemetry.ReportUsage("EOM", map[string]string{"window": strconv.Itoa(window)})

	return NewSimpleMovingAverage(NewRawEaseOfMovementIndicator(s), window)
}

// NewDefaultEaseOfMovementIndicator returns an EOM with default window=14.
// Panics if s is nil.
func NewDefaultEaseOfMovementIndicator(s *series.TimeSeries) Indicator {
	return NewEaseOfMovementIndicator(s, 14)
}

type rawEaseOfMovementIndicator struct {
	series *series.TimeSeries
}

// NewRawEaseOfMovementIndicator returns the raw 1-period Ease of Movement.
func NewRawEaseOfMovementIndicator(s *series.TimeSeries) Indicator {
	return &rawEaseOfMovementIndicator{series: s}
}

func (eom *rawEaseOfMovementIndicator) Calculate(index int) decimal.Decimal {
	if index <= 0 {
		return decimal.ZERO
	}

	currCandle := eom.series.GetCandle(index)
	prevCandle := eom.series.GetCandle(index - 1)

	distanceMoved := currCandle.MaxPrice.Add(currCandle.MinPrice).Div(decimal.New(2)).
		Sub(prevCandle.MaxPrice.Add(prevCandle.MinPrice).Div(decimal.New(2)))

	highLowDiff := currCandle.MaxPrice.Sub(currCandle.MinPrice)
	if highLowDiff.IsZero() {
		return decimal.ZERO
	}

	boxRatio := currCandle.Volume.Div(decimal.New(100000000)).Div(highLowDiff)
	if boxRatio.IsZero() {
		return decimal.ZERO
	}

	return distanceMoved.Div(boxRatio)
}

// --- Force Index ---

// NewForceIndexIndicator returns the Force Index smoothed with an EMA over
// the given window.  Panics if s is nil or window <= 0.
func NewForceIndexIndicator(s *series.TimeSeries, window int) Indicator {
	if s == nil {
		panic("goflux: ForceIndex series cannot be nil")
	}
	if window <= 0 {
		panic("goflux: ForceIndex window must be > 0")
	}
	telemetry.ReportUsage("ForceIndex", map[string]string{"window": strconv.Itoa(window)})

	return NewEMAIndicator(NewRawForceIndexIndicator(s), window)
}

// NewDefaultForceIndexIndicator returns a Force Index with default window=13.
// Panics if s is nil.
func NewDefaultForceIndexIndicator(s *series.TimeSeries) Indicator {
	return NewForceIndexIndicator(s, 13)
}

type rawForceIndexIndicator struct {
	series *series.TimeSeries
}

// NewRawForceIndexIndicator returns the raw 1-period Force Index.
func NewRawForceIndexIndicator(s *series.TimeSeries) Indicator {
	return &rawForceIndexIndicator{series: s}
}

func (fi *rawForceIndexIndicator) Calculate(index int) decimal.Decimal {
	if index <= 0 {
		return decimal.ZERO
	}

	currCandle := fi.series.GetCandle(index)
	prevCandle := fi.series.GetCandle(index - 1)

	closeDiff := currCandle.ClosePrice.Sub(prevCandle.ClosePrice)
	return closeDiff.Mul(currCandle.Volume)
}
