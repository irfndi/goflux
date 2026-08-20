package series

import (
	"fmt"
	"sync"

	"github.com/irfndi/goflux/pkg/decimal"
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

// CandleRange returns a shallow copy of the candle references in [start, end).
// The range is captured under one read lock, making it safe for callers that
// need to inspect several adjacent candles while avoiding repeated locking.
func (ts *TimeSeries) CandleRange(start, end int) []*Candle {
	if ts == nil {
		return nil
	}
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	if start < 0 {
		start = 0
	}
	if end > len(ts.Candles) {
		end = len(ts.Candles)
	}
	if start >= end {
		return nil
	}

	rangeCandles := make([]*Candle, end-start)
	copy(rangeCandles, ts.Candles[start:end])
	return rangeCandles
}

// HighLow returns the highest high and lowest low in [start, end).
func (ts *TimeSeries) HighLow(start, end int) (highest, lowest decimal.Decimal, ok bool) {
	if ts == nil {
		return decimal.ZERO, decimal.ZERO, false
	}
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	start, end, ok = normalizeStrictRange(start, end, len(ts.Candles))
	if !ok {
		return decimal.ZERO, decimal.ZERO, false
	}
	return highLowUnsafe(ts.Candles[start:end])
}

// HighLowClose returns the highest high, lowest low, and closing price of the
// final candle in [start, end). All values are read under one lock.
func (ts *TimeSeries) HighLowClose(start, end int) (highest, lowest, close decimal.Decimal, ok bool) {
	if ts == nil {
		return decimal.ZERO, decimal.ZERO, decimal.ZERO, false
	}
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	start, end, ok = normalizeStrictRange(start, end, len(ts.Candles))
	if !ok || ts.Candles[end-1] == nil {
		return decimal.ZERO, decimal.ZERO, decimal.ZERO, false
	}
	highest, lowest, ok = highLowUnsafe(ts.Candles[start:end])
	if !ok {
		return decimal.ZERO, decimal.ZERO, decimal.ZERO, false
	}
	return highest, lowest, ts.Candles[end-1].ClosePrice, true
}

func normalizeStrictRange(start, end, length int) (int, int, bool) {
	if start < 0 {
		start = 0
	}
	return start, end, start < end && end <= length
}

func highLowUnsafe(candles []*Candle) (highest, lowest decimal.Decimal, ok bool) {
	for _, candle := range candles {
		if candle == nil {
			continue
		}
		if !ok {
			highest = candle.MaxPrice
			lowest = candle.MinPrice
			ok = true
			continue
		}
		if candle.MaxPrice.GT(highest) {
			highest = candle.MaxPrice
		}
		if candle.MinPrice.LT(lowest) {
			lowest = candle.MinPrice
		}
	}
	return highest, lowest, ok
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
