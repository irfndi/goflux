package indicators

import (
	"strconv"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/telemetry"
)

// ChaikinMoneyFlow calculates the Chaikin Money Flow indicator.
// It measures the amount of Money Flow Volume over a given window.
// Values range between -1 and +1.
// Positive values indicate buying pressure, negative values selling pressure.
type chaikinMoneyFlow struct {
	series *series.TimeSeries
	window int
}

// NewChaikinMoneyFlowIndicator returns a Chaikin Money Flow indicator.
// The CMF is calculated as: Sum(Money Flow Volume, window) / Sum(Volume, window).
// Panics if window < 1.
func NewChaikinMoneyFlowIndicator(s *series.TimeSeries, window int) Indicator {
	if window < 1 {
		panic("goflux: Chaikin Money Flow window must be >= 1")
	}
	telemetry.ReportUsage("ChaikinMoneyFlow", map[string]string{"window": strconv.Itoa(window)})
	return chaikinMoneyFlow{series: s, window: window}
}

func (cmf chaikinMoneyFlow) Calculate(index int) decimal.Decimal {
	if cmf.series == nil {
		return decimal.ZERO
	}

	length := cmf.series.Length()
	if index < 0 || index >= length {
		return decimal.ZERO
	}

	if index < cmf.window-1 {
		return decimal.ZERO
	}

	sumMFV := decimal.ZERO
	sumVolume := decimal.ZERO

	start := index - cmf.window + 1

	for i := start; i <= index; i++ {
		candle := cmf.series.GetCandle(i)
		if candle == nil {
			continue
		}
		high := candle.MaxPrice
		low := candle.MinPrice
		close := candle.ClosePrice
		volume := candle.Volume

		highLow := high.Sub(low)
		var mfm decimal.Decimal
		if highLow.IsZero() {
			mfm = decimal.ZERO
		} else {
			mfm = close.Sub(low).Sub(high.Sub(close)).Div(highLow)
		}

		mfv := mfm.Mul(volume)
		sumMFV = sumMFV.Add(mfv)
		sumVolume = sumVolume.Add(volume)
	}

	if sumVolume.IsZero() {
		return decimal.ZERO
	}

	return sumMFV.Div(sumVolume)
}
