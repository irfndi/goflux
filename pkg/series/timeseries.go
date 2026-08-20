package series

import (
	"fmt"
	"sync"
)

// TimeSeries represents an array of candles with thread-safe operations
type TimeSeries struct {
	mu      sync.RWMutex
	Candles []*Candle
}

// NewTimeSeries returns a new, empty, TimeSeries
func NewTimeSeries() (t *TimeSeries) {
	t = new(TimeSeries)
	t.Candles = make([]*Candle, 0)

	return t
}

// AddCandle adds the given candle to this TimeSeries if it is not nil and after the last candle in this timeseries.
// If the candle is added, AddCandle will return true, otherwise it will return false.
// Thread-safe: uses write lock.
func (ts *TimeSeries) AddCandle(candle *Candle) bool {
	if ts == nil || candle == nil {
		return false
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()

	last := ts.lastCandleUnsafe()
	if last == nil || candle.Period.Since(last.Period) >= 0 {
		ts.Candles = append(ts.Candles, candle)
		return true
	}

	return false
}

// AddCandleErr adds given candle to this TimeSeries with error handling.
// Returns error if candle is nil or if candle cannot be added.
// Thread-safe: uses write lock.
func (ts *TimeSeries) AddCandleErr(candle *Candle) error {
	if ts == nil {
		return fmt.Errorf("time series cannot be nil")
	}
	if candle == nil {
		return fmt.Errorf("candle cannot be nil")
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()

	last := ts.lastCandleUnsafe()
	if last == nil || candle.Period.Since(last.Period) >= 0 {
		ts.Candles = append(ts.Candles, candle)
		return nil
	}

	return fmt.Errorf("candle period (%v) is not after last candle period (%v)", candle.Period, last.Period)
}

// LastCandle will return the lastCandle in this series, or nil if this series is empty
// Thread-safe: uses read lock.
func (ts *TimeSeries) LastCandle() *Candle {
	if ts == nil {
		return nil
	}
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.lastCandleUnsafe()
}

func (ts *TimeSeries) lastCandleUnsafe() *Candle {
	if len(ts.Candles) > 0 {
		return ts.Candles[len(ts.Candles)-1]
	}

	return nil
}

// LastIndex will return the index of the last candle in this series
// Thread-safe: uses read lock.
func (ts *TimeSeries) LastIndex() int {
	if ts == nil {
		return -1
	}
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.Candles) - 1
}

// GetCandle returns the candle at the given index, or nil if out of bounds
// Thread-safe: uses read lock.
func (ts *TimeSeries) GetCandle(index int) *Candle {
	if ts == nil {
		return nil
	}
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	if index < 0 || index >= len(ts.Candles) {
		return nil
	}
	return ts.Candles[index]
}

// GetCandlePair returns the candle at index and its immediate predecessor.
// Both lookups share one read lock so indicators that need adjacent candles
// do not pay for multiple lock acquisitions.
func (ts *TimeSeries) GetCandlePair(index int) (current, previous *Candle) {
	if ts == nil {
		return nil, nil
	}
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	if index < 0 || index >= len(ts.Candles) {
		return nil, nil
	}
	current = ts.Candles[index]
	if index > 0 {
		previous = ts.Candles[index-1]
	}
	return current, previous
}

// Length returns the number of candles in the series
// Thread-safe: uses read lock.
func (ts *TimeSeries) Length() int {
	if ts == nil {
		return 0
	}
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.Candles)
}

// CandlesSnapshot returns a shallowly immutable snapshot of the candles.
// The returned candle values are copies and can be safely rearranged by the
// caller without changing the series slice. Use AddCandle for series updates.
func (ts *TimeSeries) CandlesSnapshot() []*Candle {
	if ts == nil {
		return nil
	}
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	snapshot := make([]*Candle, len(ts.Candles))
	for i, candle := range ts.Candles {
		if candle == nil {
			continue
		}
		copy := *candle
		snapshot[i] = &copy
	}
	return snapshot
}
