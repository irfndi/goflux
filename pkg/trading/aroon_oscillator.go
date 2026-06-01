package trading

import (
	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/series"
)

// aroonOscillatorOverLevelRule is satisfied when the Aroon Oscillator
// is above a given level (e.g., +50 for strong uptrend confirmation).
type aroonOscillatorOverLevelRule struct {
	aroonOsc indicators.Indicator
	level    decimal.Decimal
}

// NewAroonOscillatorOverLevelRule returns a rule that is satisfied when
// the Aroon Oscillator exceeds the given level.
func NewAroonOscillatorOverLevelRule(aroonOsc indicators.Indicator, level float64) Rule {
	return aroonOscillatorOverLevelRule{
		aroonOsc: aroonOsc,
		level:    decimal.New(level),
	}
}

func (ar aroonOscillatorOverLevelRule) IsSatisfied(index int, record *TradingRecord) bool {
	return ar.aroonOsc.Calculate(index).GT(ar.level)
}

// aroonOscillatorUnderLevelRule is satisfied when the Aroon Oscillator
// is below a given level (e.g., -50 for strong downtrend confirmation).
type aroonOscillatorUnderLevelRule struct {
	aroonOsc indicators.Indicator
	level    decimal.Decimal
}

// NewAroonOscillatorUnderLevelRule returns a rule that is satisfied when
// the Aroon Oscillator falls below the given level.
func NewAroonOscillatorUnderLevelRule(aroonOsc indicators.Indicator, level float64) Rule {
	return aroonOscillatorUnderLevelRule{
		aroonOsc: aroonOsc,
		level:    decimal.New(level),
	}
}

func (ar aroonOscillatorUnderLevelRule) IsSatisfied(index int, record *TradingRecord) bool {
	return ar.aroonOsc.Calculate(index).LT(ar.level)
}

// NewAroonOscillatorBullishRule returns a convenience rule using an Aroon Oscillator
// constructed from the given time series. It triggers when the oscillator is above zero.
func NewAroonOscillatorBullishRule(s *series.TimeSeries, window int) Rule {
	return NewAroonOscillatorOverLevelRule(indicators.NewAroonOscillatorFromSeries(s, window), 0)
}

// NewAroonOscillatorBearishRule returns a convenience rule using an Aroon Oscillator
// constructed from the given time series. It triggers when the oscillator is below zero.
func NewAroonOscillatorBearishRule(s *series.TimeSeries, window int) Rule {
	return NewAroonOscillatorUnderLevelRule(indicators.NewAroonOscillatorFromSeries(s, window), 0)
}
