package indicators

import (
	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
)

type trueRangeIndicator struct {
	series *series.TimeSeries
}

// NewTrueRangeIndicator returns a base indicator
// which calculates the true range at the current point in time for a series
// https://www.investopedia.com/terms/a/atr.asp
func NewTrueRangeIndicator(series *series.TimeSeries) Indicator {
	return trueRangeIndicator{
		series: series,
	}
}

func (tri trueRangeIndicator) Calculate(index int) decimal.Decimal {
	if tri.series == nil || index < 0 || index >= tri.series.Length() {
		return decimal.ZERO
	}

	candle := tri.series.GetCandle(index)
	if index == 0 {
		return candle.MaxPrice.Sub(candle.MinPrice).Abs()
	}
	previous := tri.series.GetCandle(index - 1)
	if previous == nil {
		return decimal.ZERO
	}
	previousClose := previous.ClosePrice

	trueHigh := candle.MaxPrice.Max(previousClose)
	trueLow := candle.MinPrice.Min(previousClose)

	return trueHigh.Sub(trueLow)
}
