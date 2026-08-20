package indicators

import "github.com/irfndi/goflux/pkg/decimal"

// MultiCalculate calculates multiple indicators for a given index.
// Indicator implementations may maintain recursive caches, so calculations
// are deliberately ordered to keep the helper race-free for every Indicator.
func MultiCalculate(index int, indicators ...Indicator) []decimal.Decimal {
	results := make([]decimal.Decimal, len(indicators))
	for i, ind := range indicators {
		if ind != nil {
			results[i] = ind.Calculate(index)
		}
	}
	return results
}

// BatchCalculate calculates an indicator for a range of indices.
// NOTE: This only works for non-recursive indicators (like SMA, RSI, but NOT EMA)
// unless the cache is already populated or the indicator handles concurrency internally.
func BatchCalculate(ind Indicator, indices []int) []decimal.Decimal {
	results := make([]decimal.Decimal, len(indices))
	if ind == nil {
		return results
	}
	for i, idx := range indices {
		results[i] = ind.Calculate(idx)
	}
	return results
}
