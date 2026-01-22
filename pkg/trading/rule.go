package trading

import (
	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/series"
)

// Rule is an interface describing an algorithm by which a set of criteria may be satisfied
type Rule interface {
	IsSatisfied(index int, record *TradingRecord) bool
}

// And returns a new rule whereby BOTH of the passed-in rules must be satisfied for the rule to be satisfied
func And(r1, r2 Rule) Rule {
	return andRule{r1, r2}
}

// Or returns a new rule whereby ONE OF the passed-in rules must be satisfied for the rule to be satisfied
func Or(r1, r2 Rule) Rule {
	return orRule{r1, r2}
}

// Not returns a new rule whereby the passed-in rule must NOT be satisfied for the rule to be satisfied
func Not(r Rule) Rule {
	return notRule{r}
}

type andRule struct {
	r1 Rule
	r2 Rule
}

func (ar andRule) IsSatisfied(index int, record *TradingRecord) bool {
	return ar.r1.IsSatisfied(index, record) && ar.r2.IsSatisfied(index, record)
}

type orRule struct {
	r1 Rule
	r2 Rule
}

func (or orRule) IsSatisfied(index int, record *TradingRecord) bool {
	return or.r1.IsSatisfied(index, record) || or.r2.IsSatisfied(index, record)
}

type notRule struct {
	r Rule
}

func (nr notRule) IsSatisfied(index int, record *TradingRecord) bool {
	return !nr.r.IsSatisfied(index, record)
}

// PositionNewRule is a rule that is satisfied when the current position is new
type PositionNewRule struct{}

func (pnr PositionNewRule) IsSatisfied(index int, record *TradingRecord) bool {
	return record.CurrentPosition().IsNew()
}

// PositionOpenRule is a rule that is satisfied when the current position is open
type PositionOpenRule struct{}

func (por PositionOpenRule) IsSatisfied(index int, record *TradingRecord) bool {
	return record.CurrentPosition().IsOpen()
}

// OverIndicatorRule is a rule where the First indicators.Indicator must be greater than the Second indicators.Indicator to be Satisfied
type OverIndicatorRule struct {
	First  indicators.Indicator
	Second indicators.Indicator
}

// IsSatisfied returns true when the First indicators.Indicator is greater than the Second indicators.Indicator
func (oir OverIndicatorRule) IsSatisfied(index int, record *TradingRecord) bool {
	return oir.First.Calculate(index).GT(oir.Second.Calculate(index))
}

// UnderIndicatorRule is a rule where the First indicators.Indicator must be less than the Second indicators.Indicator to be Satisfied
type UnderIndicatorRule struct {
	First  indicators.Indicator
	Second indicators.Indicator
}

// IsSatisfied returns true when the First indicators.Indicator is less than the Second indicators.Indicator
func (uir UnderIndicatorRule) IsSatisfied(index int, record *TradingRecord) bool {
	return uir.First.Calculate(index).LT(uir.Second.Calculate(index))
}

type percentChangeRule struct {
	indicator indicators.Indicator
	percent   decimal.Decimal
}

func (pgr percentChangeRule) IsSatisfied(index int, record *TradingRecord) bool {
	return pgr.indicator.Calculate(index).Abs().GT(pgr.percent.Abs())
}

// NewPercentChangeRule returns a rule whereby the given indicators.Indicator must have changed by a given percentage to be satisfied.
// You should specify percent as a float value between -1 and 1
func NewPercentChangeRule(indicator indicators.Indicator, percent float64) Rule {
	return percentChangeRule{
		indicator: indicators.NewPercentChangeIndicator(indicator),
		percent:   decimal.New(percent),
	}
}

// FixedBarExitRule is a rule that is satisfied when the position has been open for a fixed number of bars.
// This requires the TimeSeries to be passed in to find the entry index.
type FixedBarExitRule struct {
	Series *series.TimeSeries
	Bars   int
}

func (fbe FixedBarExitRule) IsSatisfied(index int, record *TradingRecord) bool {
	if !record.CurrentPosition().IsOpen() {
		return false
	}

	pos := record.CurrentPosition()
	entryTime := pos.EntranceOrder().ExecutionTime
	entryIndex := -1
	// Find entry index
	for i := index; i >= 0; i-- {
		if fbe.Series.GetCandle(i).Period.End.Equal(entryTime) {
			entryIndex = i
			break
		}
	}

	if entryIndex == -1 {
		return false
	}

	return index-entryIndex >= fbe.Bars
}
