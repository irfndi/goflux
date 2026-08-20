package indicators

import (
	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
)

type zigzagIndicator struct {
	Indicator
	series    *series.TimeSeries
	percent   decimal.Decimal
	cache     []decimal.Decimal
	cachePeak []bool // true if it's a peak or trough
}

// NewZigZagIndicator returns an indicator that calculates the ZigZag.
// It requires a percentage change (e.g. 0.05 for 5%) to form a new leg.
func NewZigZagIndicator(s *series.TimeSeries, percent float64) Indicator {
	return &zigzagIndicator{
		series:    s,
		percent:   decimal.New(percent),
		cache:     make([]decimal.Decimal, 0),
		cachePeak: make([]bool, 0),
	}
}

func (z *zigzagIndicator) Calculate(index int) decimal.Decimal {
	if z == nil || z.series == nil || index < 0 || index >= z.series.Length() {
		return decimal.ZERO
	}

	values := make([]decimal.Decimal, index+1)
	first := z.series.GetCandle(0).ClosePrice
	values[0] = first
	if z.percent.IsZero() || z.percent.IsNegative() {
		for i := 1; i <= index; i++ {
			values[i] = z.series.GetCandle(i).ClosePrice
		}
		return values[index]
	}

	direction := 0
	pivot := first
	extreme := first
	for i := 1; i <= index; i++ {
		price := z.series.GetCandle(i).ClosePrice
		if direction == 0 {
			if changedBy(price, pivot, z.percent) {
				direction = signOf(price.Sub(pivot))
				extreme = price
			} else {
				extreme = price
			}
			values[i] = extreme
			continue
		}

		if direction > 0 {
			if price.GT(extreme) {
				extreme = price
			}
			if changedBy(extreme.Sub(price), extreme, z.percent) {
				pivot = extreme
				direction = -1
				extreme = price
			}
		} else {
			if price.LT(extreme) {
				extreme = price
			}
			if changedBy(price.Sub(extreme), extreme, z.percent) {
				pivot = extreme
				direction = 1
				extreme = price
			}
		}
		values[i] = extreme
	}

	return values[index]
}

func changedBy(delta, base, threshold decimal.Decimal) bool {
	if base.IsZero() {
		return delta.Abs().GTE(threshold)
	}
	return delta.Abs().Div(base.Abs()).GTE(threshold)
}

func signOf(value decimal.Decimal) int {
	if value.IsNegative() {
		return -1
	}
	if value.IsPositive() {
		return 1
	}
	return 0
}
