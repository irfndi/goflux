package trading

import (
	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/series"
)

type stopLossRule struct {
	indicators.Indicator
	tolerance decimal.Decimal
}

// NewStopLossRule returns a new rule that is satisfied when the given loss tolerance (a percentage) is met or exceeded.
// Loss tolerance should be a value between -1 and 1.
func NewStopLossRule(series *series.TimeSeries, lossTolerance float64) Rule {
	return stopLossRule{
		Indicator: indicators.NewClosePriceIndicator(series),
		tolerance: decimal.New(lossTolerance),
	}
}

func (slr stopLossRule) IsSatisfied(index int, record *TradingRecord) bool {
	if !record.CurrentPosition().IsOpen() {
		return false
	}

	openPrice := record.CurrentPosition().CostBasis()
	loss := slr.Indicator.Calculate(index).Div(openPrice).Sub(decimal.ONE)
	return loss.LTE(slr.tolerance)
}
