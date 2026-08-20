package indicators

import (
	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
)

type williamsRIndicator struct {
	Indicator
	series *series.TimeSeries
	high   Indicator
	low    Indicator
	close  Indicator
	window int
}

func NewWilliamsRIndicator(s *series.TimeSeries, window int) Indicator {
	return &williamsRIndicator{
		series: s,
		high:   NewHighPriceIndicator(s),
		low:    NewLowPriceIndicator(s),
		close:  NewClosePriceIndicator(s),
		window: window,
	}
}

func (wi *williamsRIndicator) Calculate(index int) decimal.Decimal {
	if wi == nil || wi.series == nil || index < wi.window-1 {
		return decimal.ZERO
	}

	highestHigh, lowestLow, closePrice, ok := wi.series.HighLowClose(index-wi.window+1, index+1)
	if !ok {
		return decimal.ZERO
	}
	rangeVal := highestHigh.Sub(lowestLow)

	if rangeVal.Zero() {
		return decimal.ZERO
	}

	numerator := highestHigh.Sub(closePrice)
	result := numerator.Div(rangeVal).Mul(decimal.New(-100))

	return result
}
