package trading

import (
	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/series"
)

// chaikinMoneyFlowOverLevelRule is satisfied when the Chaikin Money Flow
// is above a given level (e.g., +0.05 for buying pressure confirmation).
type chaikinMoneyFlowOverLevelRule struct {
	cmf   indicators.Indicator
	level decimal.Decimal
}

// NewChaikinMoneyFlowOverLevelRule returns a rule that is satisfied when
// the CMF exceeds the given level.
func NewChaikinMoneyFlowOverLevelRule(cmf indicators.Indicator, level float64) Rule {
	if cmf == nil {
		panic("goflux: ChaikinMoneyFlowOverLevelRule requires non-nil indicator")
	}
	return chaikinMoneyFlowOverLevelRule{
		cmf:   cmf,
		level: decimal.New(level),
	}
}

func (r chaikinMoneyFlowOverLevelRule) IsSatisfied(index int, record *TradingRecord) bool {
	return r.cmf.Calculate(index).GT(r.level)
}

// chaikinMoneyFlowUnderLevelRule is satisfied when the Chaikin Money Flow
// is below a given level (e.g., -0.05 for selling pressure confirmation).
type chaikinMoneyFlowUnderLevelRule struct {
	cmf   indicators.Indicator
	level decimal.Decimal
}

// NewChaikinMoneyFlowUnderLevelRule returns a rule that is satisfied when
// the CMF falls below the given level.
func NewChaikinMoneyFlowUnderLevelRule(cmf indicators.Indicator, level float64) Rule {
	if cmf == nil {
		panic("goflux: ChaikinMoneyFlowUnderLevelRule requires non-nil indicator")
	}
	return chaikinMoneyFlowUnderLevelRule{
		cmf:   cmf,
		level: decimal.New(level),
	}
}

func (r chaikinMoneyFlowUnderLevelRule) IsSatisfied(index int, record *TradingRecord) bool {
	return r.cmf.Calculate(index).LT(r.level)
}

// NewChaikinMoneyFlowBullishRule returns a convenience rule using a CMF
// constructed from the given time series. It triggers when CMF is above zero.
func NewChaikinMoneyFlowBullishRule(s *series.TimeSeries, window int) Rule {
	return NewChaikinMoneyFlowOverLevelRule(indicators.NewChaikinMoneyFlowIndicator(s, window), 0)
}

// NewChaikinMoneyFlowBearishRule returns a convenience rule using a CMF
// constructed from the given time series. It triggers when CMF is below zero.
func NewChaikinMoneyFlowBearishRule(s *series.TimeSeries, window int) Rule {
	return NewChaikinMoneyFlowUnderLevelRule(indicators.NewChaikinMoneyFlowIndicator(s, window), 0)
}
