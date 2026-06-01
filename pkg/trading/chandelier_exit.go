package trading

import (
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/series"
)

// chandelierExitLongRule is satisfied when the close price falls below
// the long Chandelier Exit level (a volatility-based trailing stop for longs).
type chandelierExitLongRule struct {
	closePrice indicators.Indicator
	exitLevel  indicators.Indicator
}

// NewChandelierExitLongRule returns a rule that is satisfied when the close price
// drops below the long Chandelier Exit level. This is used as a stop-loss exit
// for long positions.
func NewChandelierExitLongRule(s *series.TimeSeries, period int, atrWindow int, multiplier float64) Rule {
	return chandelierExitLongRule{
		closePrice: indicators.NewClosePriceIndicator(s),
		exitLevel:  indicators.NewChandelierExitLong(s, period, atrWindow, multiplier),
	}
}

// NewDefaultChandelierExitLongRule returns a rule using standard Chandelier Exit
// defaults: period=22, atrWindow=22, multiplier=3.0.
func NewDefaultChandelierExitLongRule(s *series.TimeSeries) Rule {
	return chandelierExitLongRule{
		closePrice: indicators.NewClosePriceIndicator(s),
		exitLevel:  indicators.NewDefaultChandelierExitLong(s),
	}
}

func (ce chandelierExitLongRule) IsSatisfied(index int, record *TradingRecord) bool {
	if !record.CurrentPosition().IsOpen() {
		return false
	}

	exitLevel := ce.exitLevel.Calculate(index)
	if exitLevel.IsZero() {
		return false
	}

	return ce.closePrice.Calculate(index).LTE(exitLevel)
}

// chandelierExitShortRule is satisfied when the close price rises above
// the short Chandelier Exit level (a volatility-based trailing stop for shorts).
type chandelierExitShortRule struct {
	closePrice indicators.Indicator
	exitLevel  indicators.Indicator
}

// NewChandelierExitShortRule returns a rule that is satisfied when the close price
// rises above the short Chandelier Exit level. This is used as a stop-loss exit
// for short positions.
func NewChandelierExitShortRule(s *series.TimeSeries, period int, atrWindow int, multiplier float64) Rule {
	return chandelierExitShortRule{
		closePrice: indicators.NewClosePriceIndicator(s),
		exitLevel:  indicators.NewChandelierExitShort(s, period, atrWindow, multiplier),
	}
}

// NewDefaultChandelierExitShortRule returns a rule using standard Chandelier Exit
// defaults: period=22, atrWindow=22, multiplier=3.0.
func NewDefaultChandelierExitShortRule(s *series.TimeSeries) Rule {
	return chandelierExitShortRule{
		closePrice: indicators.NewClosePriceIndicator(s),
		exitLevel:  indicators.NewDefaultChandelierExitShort(s),
	}
}

func (ce chandelierExitShortRule) IsSatisfied(index int, record *TradingRecord) bool {
	if !record.CurrentPosition().IsOpen() {
		return false
	}

	exitLevel := ce.exitLevel.Calculate(index)
	if exitLevel.IsZero() {
		return false
	}

	return ce.closePrice.Calculate(index).GTE(exitLevel)
}
