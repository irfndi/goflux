package indicators

import (
	"strconv"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/math"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/telemetry"
)

// flatStochasticValue is returned when the price range is zero (min == max),
// meaning the candle is flat. The midpoint (50) is used because a flat candle
// provides no directional information, so the neutral midpoint of the 0–100
// stochastic range is the least misleading assumption. This was changed from
// +Inf to prevent propagation of non-finite values through downstream SMA/EMA
// calculations. Note: this is a behavioral change from the previous +Inf return.
const flatStochasticValue = 50

type kIndicator struct {
	closePrice Indicator
	minValue   Indicator
	maxValue   Indicator
	window     int
	series     *series.TimeSeries
}

// NewFastStochasticIndicator returns a derivative Indicator which returns the fast stochastic indicator (%K) for the
// given window.
// https://www.investopedia.com/terms/s/stochasticoscillator.asp
func NewFastStochasticIndicator(series *series.TimeSeries, timeframe int) Indicator {
	telemetry.ReportUsage("Stochastic", map[string]string{"timeframe": strconv.Itoa(timeframe), "type": "fast"})
	return kIndicator{
		closePrice: NewClosePriceIndicator(series),
		minValue:   NewMinimumValueIndicator(NewLowPriceIndicator(series), timeframe),
		maxValue:   NewMaximumValueIndicator(NewHighPriceIndicator(series), timeframe),
		window:     timeframe,
		series:     series,
	}
}

func (k kIndicator) Calculate(index int) decimal.Decimal {
	if k.series != nil {
		start := 0
		if k.window > 0 {
			start = math.Max(index-k.window+1, 0)
		}
		maxVal, minVal, closeVal, ok := k.series.HighLowClose(start, index+1)
		if !ok {
			return decimal.ZERO
		}
		if minVal.EQ(maxVal) {
			return decimal.New(flatStochasticValue)
		}
		return closeVal.Sub(minVal).Div(maxVal.Sub(minVal)).Mul(decimal.New(100))
	}

	closeVal := k.closePrice.Calculate(index)
	minVal := k.minValue.Calculate(index)
	maxVal := k.maxValue.Calculate(index)

	if minVal.EQ(maxVal) {
		return decimal.New(flatStochasticValue)
	}

	return closeVal.Sub(minVal).Div(maxVal.Sub(minVal)).Mul(decimal.New(100))
}

type dIndicator struct {
	k      Indicator
	window int
}

// NewSlowStochasticIndicator returns a derivative Indicator which returns the slow stochastic indicator (%D) for the
// given window.
// https://www.investopedia.com/terms/s/stochasticoscillator.asp
func NewSlowStochasticIndicator(k Indicator, window int) Indicator {
	return dIndicator{k, window}
}

func (d dIndicator) Calculate(index int) decimal.Decimal {
	return NewSimpleMovingAverage(d.k, d.window).Calculate(index)
}
