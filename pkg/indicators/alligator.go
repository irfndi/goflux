package indicators

import (
	"strconv"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/telemetry"
)

// NewAlligatorIndicators returns the three lines of the Williams Alligator:
// Jaw (blue), Teeth (red), and Lips (green).
// Each line is a smoothed moving average (SMMA/MMA) of the median price,
// displaced forward by a fixed number of bars.
//
// Defaults: Jaw(13,8), Teeth(8,5), Lips(5,3).
// Panics if s is nil.
func NewAlligatorIndicators(s *series.TimeSeries) (jaw, teeth, lips Indicator) {
	if s == nil {
		panic("goflux: Alligator series cannot be nil")
	}
	telemetry.ReportUsage("Alligator", nil)
	median := NewMedianPriceIndicator(s)
	return newAlligatorJaw(median), newAlligatorTeeth(median), newAlligatorLips(median)
}

// NewAlligatorIndicatorsCustom returns Alligator lines with custom periods and shifts.
// Panics if s is nil or if any period <= 0 or shift < 0.
func NewAlligatorIndicatorsCustom(
	s *series.TimeSeries,
	jawPeriod, jawShift int,
	teethPeriod, teethShift int,
	lipsPeriod, lipsShift int,
) (jaw, teeth, lips Indicator) {
	if s == nil {
		panic("goflux: Alligator series cannot be nil")
	}
	telemetry.ReportUsage("AlligatorCustom", map[string]string{
		"jaw_period":   strconv.Itoa(jawPeriod),
		"jaw_shift":    strconv.Itoa(jawShift),
		"teeth_period": strconv.Itoa(teethPeriod),
		"teeth_shift":  strconv.Itoa(teethShift),
		"lips_period":  strconv.Itoa(lipsPeriod),
		"lips_shift":   strconv.Itoa(lipsShift),
	})
	median := NewMedianPriceIndicator(s)
	return newShiftedSMMA(median, jawPeriod, jawShift),
		newShiftedSMMA(median, teethPeriod, teethShift),
		newShiftedSMMA(median, lipsPeriod, lipsShift)
}

// NewGatorOscillatorIndicators returns the upper and lower histogram bars
// of the Gator Oscillator derived from the Alligator lines.
// Upper = |Jaw - Teeth|, Lower = -|Teeth - Lips|.
// Panics if s is nil.
func NewGatorOscillatorIndicators(s *series.TimeSeries) (upper, lower Indicator) {
	if s == nil {
		panic("goflux: Gator Oscillator series cannot be nil")
	}
	telemetry.ReportUsage("GatorOscillator", nil)
	jaw, teeth, lips := NewAlligatorIndicators(s)
	return NewGatorOscillatorIndicatorsFromAlligator(jaw, teeth, lips)
}

// NewGatorOscillatorIndicatorsFromAlligator creates Gator Oscillator bars
// from pre-computed Alligator indicators, avoiding redundant SMMA computation.
func NewGatorOscillatorIndicatorsFromAlligator(jaw, teeth, lips Indicator) (upper, lower Indicator) {
	return gatorUpper{jaw: jaw, teeth: teeth},
		gatorLower{teeth: teeth, lips: lips}
}

// --- Alligator lines ---

type shiftedSMMA struct {
	smma  Indicator
	shift int
}

func newShiftedSMMA(indicator Indicator, period, shift int) Indicator {
	if period <= 0 {
		panic("goflux: Alligator SMMA period must be > 0")
	}
	if shift < 0 {
		panic("goflux: Alligator shift must be >= 0")
	}
	return shiftedSMMA{smma: NewMMAIndicator(indicator, period), shift: shift}
}

func newAlligatorJaw(indicator Indicator) Indicator {
	return newShiftedSMMA(indicator, 13, 8)
}

func newAlligatorTeeth(indicator Indicator) Indicator {
	return newShiftedSMMA(indicator, 8, 5)
}

func newAlligatorLips(indicator Indicator) Indicator {
	return newShiftedSMMA(indicator, 5, 3)
}

func (s shiftedSMMA) Calculate(index int) decimal.Decimal {
	shifted := index - s.shift
	if shifted < 0 {
		return decimal.ZERO
	}
	return s.smma.Calculate(shifted)
}

// --- Gator Oscillator ---

type gatorUpper struct {
	jaw   Indicator
	teeth Indicator
}

type gatorLower struct {
	teeth Indicator
	lips  Indicator
}

func (g gatorUpper) Calculate(index int) decimal.Decimal {
	return g.jaw.Calculate(index).Sub(g.teeth.Calculate(index)).Abs()
}

func (g gatorLower) Calculate(index int) decimal.Decimal {
	return g.teeth.Calculate(index).Sub(g.lips.Calculate(index)).Abs().Neg()
}
