package trading

import (
	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/series"
)

// trixOverLevelRule is satisfied when TRIX exceeds a given level.
type trixOverLevelRule struct {
	trix  indicators.Indicator
	level decimal.Decimal
}

// NewTRIXOverLevelRule returns a rule that is satisfied when TRIX exceeds the given level.
func NewTRIXOverLevelRule(trix indicators.Indicator, level float64) Rule {
	if trix == nil {
		panic("goflux: TRIXOverLevelRule requires non-nil indicator")
	}
	return trixOverLevelRule{
		trix:  trix,
		level: decimal.New(level),
	}
}

func (r trixOverLevelRule) IsSatisfied(index int, record *TradingRecord) bool {
	return r.trix.Calculate(index).GT(r.level)
}

// trixUnderLevelRule is satisfied when TRIX falls below a given level.
type trixUnderLevelRule struct {
	trix  indicators.Indicator
	level decimal.Decimal
}

// NewTRIXUnderLevelRule returns a rule that is satisfied when TRIX falls below the given level.
func NewTRIXUnderLevelRule(trix indicators.Indicator, level float64) Rule {
	if trix == nil {
		panic("goflux: TRIXUnderLevelRule requires non-nil indicator")
	}
	return trixUnderLevelRule{
		trix:  trix,
		level: decimal.New(level),
	}
}

func (r trixUnderLevelRule) IsSatisfied(index int, record *TradingRecord) bool {
	return r.trix.Calculate(index).LT(r.level)
}

// NewTRIXBullishRule returns a convenience rule using a TRIX
// constructed from the given time series. It triggers when TRIX is above zero.
func NewTRIXBullishRule(s *series.TimeSeries, window int) Rule {
	return NewTRIXOverLevelRule(indicators.NewTRIXIndicatorFromSeries(s, window), 0)
}

// NewTRIXBearishRule returns a convenience rule using a TRIX
// constructed from the given time series. It triggers when TRIX is below zero.
func NewTRIXBearishRule(s *series.TimeSeries, window int) Rule {
	return NewTRIXUnderLevelRule(indicators.NewTRIXIndicatorFromSeries(s, window), 0)
}
