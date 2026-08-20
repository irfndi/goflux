package indicators

import "github.com/irfndi/goflux/pkg/decimal"

// NewMaximumDrawdownIndicator returns a derivative Indicator which returns the maximum
// drawdown of the underlying indicator over a window. Maximum drawdown is defined as the
// maximum observed loss from peak of an underlying indicator in a given timeframe.
// Maximum drawdown is given as a percentage of the peak. Use a window value of -1 to include
// all values present in the underlying indicator.
// See: https://www.investopedia.com/terms/m/maximum-drawdown-mdd.asp
func NewMaximumDrawdownIndicator(ind Indicator, window int) Indicator {
	return maximumDrawdownIndicator{
		indicator: ind,
		window:    window,
	}
}

type maximumDrawdownIndicator struct {
	indicator Indicator
	window    int
}

func (mdi maximumDrawdownIndicator) Calculate(index int) decimal.Decimal {
	if mdi.indicator == nil || index < 0 {
		return decimal.ZERO
	}

	start := 0
	if mdi.window > 0 && index >= mdi.window {
		start = index - mdi.window + 1
	}
	peak := mdi.indicator.Calculate(start)
	maxDrawdown := decimal.ZERO
	for i := start; i <= index; i++ {
		value := mdi.indicator.Calculate(i)
		if value.GT(peak) {
			peak = value
		}
		if peak.IsZero() {
			continue
		}
		drawdown := value.Sub(peak).Div(peak)
		if drawdown.LT(maxDrawdown) {
			maxDrawdown = drawdown
		}
	}
	return maxDrawdown
}
