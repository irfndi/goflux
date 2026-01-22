package indicators

import "github.com/irfndi/goflux/pkg/decimal"

type resultCache []*decimal.Decimal

type cachedIndicator interface {
	Indicator
	cache() resultCache
	setCache(cache resultCache)
	windowSize() int
}

func cacheResult(indicator cachedIndicator, index int, val decimal.Decimal) {
	if index < len(indicator.cache()) {
		indicator.cache()[index] = &val
	} else if index == len(indicator.cache()) {
		indicator.setCache(append(indicator.cache(), &val))
	} else {
		expandResultCache(indicator, index+1)
		cacheResult(indicator, index, val)
	}
}

func expandResultCache(indicator cachedIndicator, newSize int) {
	sizeDiff := newSize - len(indicator.cache())

	expansion := make([]*decimal.Decimal, sizeDiff)
	indicator.setCache(append(indicator.cache(), expansion...))
}

func returnIfCached(indicator cachedIndicator, index int, firstValueFallback func(int) decimal.Decimal) *decimal.Decimal {
	if index >= len(indicator.cache()) {
		expandResultCache(indicator, index+1)
	} else if index < indicator.windowSize()-1 {
		return &decimal.ZERO
	} else if val := indicator.cache()[index]; val != nil {
		return val
	} else if index == indicator.windowSize()-1 {
		value := firstValueFallback(index)
		cacheResult(indicator, index, value)
		return &value
	}

	return nil
}
