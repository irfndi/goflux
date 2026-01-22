package indicators

import (
	"sync"
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
)

type mockCachedIndicator struct {
	*cacheManager
	calculateFunc func(int) decimal.Decimal
}

func (m *mockCachedIndicator) Calculate(index int) decimal.Decimal {
	return m.calculateFunc(index)
}

func (m *mockCachedIndicator) cache() resultCache {
	return m.items
}

func (m *mockCachedIndicator) setCache(c resultCache) {
	m.items = c
}

func (m *mockCachedIndicator) windowSize() int {
	return 5
}

func (m *mockCachedIndicator) cacheMutex() *sync.RWMutex {
	return &m.mu
}

func (m *mockCachedIndicator) maxCacheSize() int {
	return m.maxSize
}

func TestnewCacheManager(t *testing.T) {
	c := newCacheManager(5)

	if len(c.items) != 5 {
		t.Errorf("Expected items length 5, got %d", len(c.items))
	}

	if c.maxSize != defaultMaxCacheSize {
		t.Errorf("Expected maxSize %d, got %d", defaultMaxCacheSize, c.maxSize)
	}
}

func TestCacheSetGet(t *testing.T) {
	c := newCacheManager(5)

	val := decimal.New(42.5)
	c.Set(0, val)

	if result := c.Get(0); result == nil || result.String() != val.String() {
		t.Errorf("Expected %s, got %v", val.String(), result)
	}
}

func TestCacheSetGetMultiple(t *testing.T) {
	c := newCacheManager(5)

	vals := []decimal.Decimal{
		decimal.New(1),
		decimal.New(2),
		decimal.New(3),
	}

	for i, v := range vals {
		c.Set(i, v)
	}

	for i, expected := range vals {
		if result := c.Get(i); result == nil || result.String() != expected.String() {
			t.Errorf("Index %d: expected %s, got %v", i, expected.String(), result)
		}
	}
}

func TestCacheExpand(t *testing.T) {
	c := newCacheManager(5)

	c.Set(10, decimal.New(99))

	if len(c.items) <= 5 {
		t.Errorf("Expected cache to expand, got length %d", len(c.items))
	}

	if result := c.Get(10); result == nil || result.String() != "99" {
		t.Errorf("Expected 99 at index 10, got %v", result)
	}
}

func TestCacheEvictAndSet(t *testing.T) {
	c := newCacheManager(5)
	c.maxSize = 100

	for i := 0; i < 50; i++ {
		c.Set(i, decimal.New(float64(i)))
	}

	if c.Len() != 50 {
		t.Errorf("Expected size 50, got %d", c.Len())
	}

	c.Set(150, decimal.New(999))

	if result := c.Get(150); result == nil {
		t.Error("Expected result at index 150")
	}
}

func TestCacheClear(t *testing.T) {
	c := newCacheManager(5)

	c.Set(0, decimal.New(1))
	c.Set(1, decimal.New(2))

	c.Clear()

	if c.Len() != defaultCacheSize {
		t.Errorf("Expected size %d after clear, got %d", defaultCacheSize, c.Len())
	}

	if c.Get(0) != nil {
		t.Error("Expected nil after clear")
	}
}

func TestCacheThreadSafety(t *testing.T) {
	c := newCacheManager(100)

	var wg sync.WaitGroup
	numGoroutines := 100
	iterationsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterationsPerGoroutine; j++ {
				idx := id*iterationsPerGoroutine + j
				c.Set(idx, decimal.New(float64(idx)))
				c.Get(idx)
			}
		}(i)
	}

	wg.Wait()

	finalSize := c.Len()
	expectedSize := numGoroutines * iterationsPerGoroutine

	if finalSize > c.maxSize {
		t.Logf("Cache size %d exceeds max size %d (expected due to eviction)", finalSize, c.maxSize)
	}

	if finalSize > expectedSize {
		t.Errorf("Cache size %d exceeds expected %d", finalSize, expectedSize)
	}
}

func TestCacheResult(t *testing.T) {
	c := newCacheManager(5)
	ind := &mockCachedIndicator{
		cacheManager: c,
		calculateFunc: func(idx int) decimal.Decimal {
			return decimal.New(float64(idx))
		},
	}

	cacheResult(ind, 0, decimal.New(10))
	cacheResult(ind, 1, decimal.New(20))
	cacheResult(ind, 2, decimal.New(30))

	if result := ind.Get(1); result == nil || result.String() != "20" {
		t.Errorf("Expected 20, got %v", result)
	}
}

func TestExpandResultCache(t *testing.T) {
	c := newCacheManager(5)
	ind := &mockCachedIndicator{
		cacheManager: c,
		calculateFunc: func(idx int) decimal.Decimal {
			return decimal.ZERO
		},
	}

	expandResultCacheInternal(ind, 10)

	if c.Len() != 10 {
		t.Errorf("Expected size 10, got %d", c.Len())
	}
}

func TestReturnIfCached(t *testing.T) {
	c := newCacheManager(5)
	ind := &mockCachedIndicator{
		cacheManager: c,
		calculateFunc: func(idx int) decimal.Decimal {
			return decimal.New(float64(idx * 10))
		},
	}

	fallbackCalled := 0
	fallback := func(idx int) decimal.Decimal {
		fallbackCalled++
		return decimal.New(float64(idx * 100))
	}

	result := returnIfCached(ind, 0, fallback)
	if result != nil {
		t.Error("Expected nil for index below window size")
	}

	if fallbackCalled != 0 {
		t.Error("Fallback should not be called for index below window size")
	}

	c.Set(0, decimal.New(1))
	c.Set(1, decimal.New(2))
	c.Set(2, decimal.New(3))
	c.Set(3, decimal.New(4))
	c.Set(4, decimal.New(5))

	result = returnIfCached(ind, 4, fallback)
	if result == nil || result.String() != "5" {
		t.Errorf("Expected 5, got %v", result)
	}

	if fallbackCalled != 0 {
		t.Error("Fallback should not be called when value is cached")
	}

	result = returnIfCached(ind, 5, fallback)
	if result == nil {
		t.Error("Expected non-nil result")
	}

	if fallbackCalled != 1 {
		t.Errorf("Expected fallback to be called once, called %d times", fallbackCalled)
	}
}

func TestGetCacheSize(t *testing.T) {
	c := newCacheManager(5)

	if GetCacheSize(c) != 5 {
		t.Errorf("Expected size 5, got %d", GetCacheSize(c))
	}

	c.Set(0, decimal.New(1))
	c.Set(1, decimal.New(2))

	if GetCacheSize(c) != 5 {
		t.Errorf("Expected size 5 (initial), got %d", GetCacheSize(c))
	}
}

func TestGetCacheCapacity(t *testing.T) {
	c := newCacheManager(5)

	if GetCacheCapacity(c) != 5 {
		t.Errorf("Expected capacity 5, got %d", GetCacheCapacity(c))
	}
}

func TestCacheMaxSize(t *testing.T) {
	c := newCacheManager(5)
	c.maxSize = 20

	if c.maxSize != 20 {
		t.Errorf("Expected maxSize 20, got %d", c.maxSize)
	}
}

func TestCacheEviction(t *testing.T) {
	c := newCacheManager(10)
	c.maxSize = 50

	for i := 0; i < 100; i++ {
		c.Set(i, decimal.New(float64(i)))
	}

	if c.Len() > c.maxSize {
		t.Errorf("Cache size %d exceeds max size %d", c.Len(), c.maxSize)
	}
}

func TestCacheGetOutOfBounds(t *testing.T) {
	c := newCacheManager(5)

	if result := c.Get(100); result != nil {
		t.Error("Expected nil for out of bounds index")
	}
}

func BenchmarkCacheManagerSet(b *testing.B) {
	c := newCacheManager(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(i%1000, decimal.New(float64(i)))
	}
}

func BenchmarkCacheManagerGet(b *testing.B) {
	c := newCacheManager(1000)
	for i := 0; i < 1000; i++ {
		c.Set(i, decimal.New(float64(i)))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(i % 1000)
	}
}

func BenchmarkCacheResult(b *testing.B) {
	c := newCacheManager(1000)
	ind := &mockCachedIndicator{
		cacheManager: c,
		calculateFunc: func(idx int) decimal.Decimal {
			return decimal.New(float64(idx))
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cacheResult(ind, i%1000, decimal.New(float64(i)))
	}
}

func BenchmarkReturnIfCached(b *testing.B) {
	c := newCacheManager(100)
	ind := &mockCachedIndicator{
		cacheManager: c,
		calculateFunc: func(idx int) decimal.Decimal {
			return decimal.New(float64(idx))
		},
	}
	for i := 0; i < 1000; i++ {
		c.Set(i, decimal.New(float64(i)))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		returnIfCached(ind, i%1000, func(idx int) decimal.Decimal {
			return decimal.New(float64(idx * 10))
		})
	}
}
