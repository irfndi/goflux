package indicators

import (
	"github.com/irfndi/goflux/pkg/decimal"
	"sync"
)

const (
	defaultCacheSize    = 1000
	defaultMaxCacheSize = 10000
)

type resultCache []*decimal.Decimal

type cachedIndicator interface {
	Indicator
	cache() resultCache
	setCache(cache resultCache)
	windowSize() int
	cacheMutex() *sync.RWMutex
	maxCacheSize() int
}

type cacheManager struct {
	mu      sync.RWMutex
	items   resultCache
	maxSize int
}

func newCacheManager(initialSize int) *cacheManager {
	size := initialSize
	if size <= 0 {
		size = defaultCacheSize
	}
	return &cacheManager{
		items:   make(resultCache, size),
		maxSize: defaultMaxCacheSize,
	}
}

func (c *cacheManager) Get(index int) *decimal.Decimal {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if index < 0 || index >= len(c.items) {
		return nil
	}
	return c.items[index]
}

func (c *cacheManager) Set(index int, val decimal.Decimal) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if index < 0 {
		return
	}

	if index >= len(c.items) {
		if len(c.items) >= c.maxSize {
			return
		}
		newSize := index + 1
		if newSize > c.maxSize {
			newSize = c.maxSize
		}
		expansion := make(resultCache, newSize-len(c.items))
		c.items = append(c.items, expansion...)
	}
	c.items[index] = &val
}

func (c *cacheManager) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *cacheManager) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(resultCache, defaultCacheSize)
}

func cacheResult(indicator cachedIndicator, index int, val decimal.Decimal) {
	mu := indicator.cacheMutex()
	mu.Lock()
	defer mu.Unlock()

	c := indicator.cache()
	if index < len(c) {
		c[index] = &val
	} else if index == len(c) {
		if len(c) >= indicator.maxCacheSize() {
			return
		}
		indicator.setCache(append(c, &val))
	} else {
		expandResultCacheInternal(indicator, index+1)
		if index < len(indicator.cache()) {
			indicator.cache()[index] = &val
		}
	}
}

func expandResultCacheInternal(indicator cachedIndicator, newSize int) {
	c := indicator.cache()
	if newSize > indicator.maxCacheSize() {
		newSize = indicator.maxCacheSize()
	}
	sizeDiff := newSize - len(c)
	if sizeDiff <= 0 {
		return
	}

	expansion := make([]*decimal.Decimal, sizeDiff)
	indicator.setCache(append(c, expansion...))
}

func returnIfCached(indicator cachedIndicator, index int, firstValueFallback func(int) decimal.Decimal) *decimal.Decimal {
	mu := indicator.cacheMutex()
	mu.RLock()
	c := indicator.cache()
	if index < len(c) && index >= 0 {
		if val := c[index]; val != nil {
			mu.RUnlock()
			return val
		}
	}
	mu.RUnlock()

	if index < indicator.windowSize()-1 {
		return nil
	}

	if index == indicator.windowSize()-1 {
		value := firstValueFallback(index)
		cacheResult(indicator, index, value)
		return &value
	}

	return nil
}

func ClearCache(c *cacheManager) {
	c.Clear()
}

func GetCacheSize(c *cacheManager) int {
	return c.Len()
}

func GetCacheCapacity(c *cacheManager) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return cap(c.items)
}