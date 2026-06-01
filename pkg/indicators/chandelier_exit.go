package indicators

import (
	"strconv"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/telemetry"
)

// ChandelierExitLong calculates the long chandelier exit level:
// Highest High(period) - multiplier * ATR(atrWindow)
type chandelierExitLongIndicator struct {
	series     *series.TimeSeries
	period     int
	atrWindow  int
	multiplier decimal.Decimal
}

// NewChandelierExitLong returns a long Chandelier Exit indicator.
// The long exit is calculated as: Highest High(period) - multiplier * ATR(atrWindow).
// Panics if period < 1 or atrWindow < 2.
func NewChandelierExitLong(s *series.TimeSeries, period int, atrWindow int, multiplier float64) Indicator {
	if period < 1 {
		panic("goflux: Chandelier Exit period must be >= 1")
	}
	if atrWindow < 2 {
		panic("goflux: Chandelier Exit ATR window must be >= 2")
	}
	telemetry.ReportUsage("ChandelierExitLong", map[string]string{
		"period":     strconv.Itoa(period),
		"atrWindow":  strconv.Itoa(atrWindow),
		"multiplier": strconv.FormatFloat(multiplier, 'f', -1, 64),
	})
	return chandelierExitLongIndicator{
		series:     s,
		period:     period,
		atrWindow:  atrWindow,
		multiplier: decimal.New(multiplier),
	}
}

// NewDefaultChandelierExitLong returns a long Chandelier Exit indicator with
// standard defaults: period=22, atrWindow=22, multiplier=3.0.
func NewDefaultChandelierExitLong(s *series.TimeSeries) Indicator {
	return NewChandelierExitLong(s, 22, 22, 3.0)
}

func (ce chandelierExitLongIndicator) Calculate(index int) decimal.Decimal {
	if index < ce.period-1 || index < ce.atrWindow-1 {
		return decimal.ZERO
	}

	highestHigh := highestHigh(ce.series, index, ce.period)
	atr := NewAverageTrueRangeIndicator(ce.series, ce.atrWindow).Calculate(index)

	return highestHigh.Sub(atr.Mul(ce.multiplier))
}

// ChandelierExitShort calculates the short chandelier exit level:
// Lowest Low(period) + multiplier * ATR(atrWindow)
type chandelierExitShortIndicator struct {
	series     *series.TimeSeries
	period     int
	atrWindow  int
	multiplier decimal.Decimal
}

// NewChandelierExitShort returns a short Chandelier Exit indicator.
// The short exit is calculated as: Lowest Low(period) + multiplier * ATR(atrWindow).
// Panics if period < 1 or atrWindow < 2.
func NewChandelierExitShort(s *series.TimeSeries, period int, atrWindow int, multiplier float64) Indicator {
	if period < 1 {
		panic("goflux: Chandelier Exit period must be >= 1")
	}
	if atrWindow < 2 {
		panic("goflux: Chandelier Exit ATR window must be >= 2")
	}
	telemetry.ReportUsage("ChandelierExitShort", map[string]string{
		"period":     strconv.Itoa(period),
		"atrWindow":  strconv.Itoa(atrWindow),
		"multiplier": strconv.FormatFloat(multiplier, 'f', -1, 64),
	})
	return chandelierExitShortIndicator{
		series:     s,
		period:     period,
		atrWindow:  atrWindow,
		multiplier: decimal.New(multiplier),
	}
}

// NewDefaultChandelierExitShort returns a short Chandelier Exit indicator with
// standard defaults: period=22, atrWindow=22, multiplier=3.0.
func NewDefaultChandelierExitShort(s *series.TimeSeries) Indicator {
	return NewChandelierExitShort(s, 22, 22, 3.0)
}

func (ce chandelierExitShortIndicator) Calculate(index int) decimal.Decimal {
	if index < ce.period-1 || index < ce.atrWindow-1 {
		return decimal.ZERO
	}

	lowestLow := lowestLow(ce.series, index, ce.period)
	atr := NewAverageTrueRangeIndicator(ce.series, ce.atrWindow).Calculate(index)

	return lowestLow.Add(atr.Mul(ce.multiplier))
}

// highestHigh returns the highest high price over the given period ending at index.
func highestHigh(s *series.TimeSeries, index int, period int) decimal.Decimal {
	maxPrice := s.GetCandle(index - period + 1).MaxPrice
	for i := index - period + 2; i <= index; i++ {
		price := s.GetCandle(i).MaxPrice
		if price.GT(maxPrice) {
			maxPrice = price
		}
	}
	return maxPrice
}

// lowestLow returns the lowest low price over the given period ending at index.
func lowestLow(s *series.TimeSeries, index int, period int) decimal.Decimal {
	minPrice := s.GetCandle(index - period + 1).MinPrice
	for i := index - period + 2; i <= index; i++ {
		price := s.GetCandle(i).MinPrice
		if price.LT(minPrice) {
			minPrice = price
		}
	}
	return minPrice
}
