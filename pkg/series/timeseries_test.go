package series_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
)

func TestTimeSeries_AddCandle(t *testing.T) {
	t.Run("Returns false if nil candle passed", func(t *testing.T) {
		ts := series.NewTimeSeries()
		assert.False(t, ts.AddCandle(nil))
	})

	t.Run("Adds candle if last is nil", func(t *testing.T) {
		ts := series.NewTimeSeries()

		candle := series.NewCandle(series.NewTimePeriod(time.Now(), time.Minute))
		candle.ClosePrice = decimal.New(1)

		ts.AddCandle(candle)

		assert.Len(t, ts.Candles, 1)
	})

	t.Run("Does not add candle if before last candle", func(t *testing.T) {
		ts := series.NewTimeSeries()

		now := time.Now()
		candle := series.NewCandle(series.NewTimePeriod(now, time.Minute))
		candle.ClosePrice = decimal.New(1)

		ts.AddCandle(candle)
		then := now.Add(-time.Minute * 10)

		nextCandle := series.NewCandle(series.NewTimePeriod(then, time.Minute))
		candle.ClosePrice = decimal.New(2)

		ts.AddCandle(nextCandle)

		assert.Len(t, ts.Candles, 1)
		assert.EqualValues(t, now.UnixNano(), ts.Candles[0].Period.Start.UnixNano())
	})
}

func TestTimeSeries_AddCandleErr(t *testing.T) {
	t.Run("Returns error if nil candle passed", func(t *testing.T) {
		ts := series.NewTimeSeries()
		err := ts.AddCandleErr(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("Adds candle if last is nil", func(t *testing.T) {
		ts := series.NewTimeSeries()

		candle := series.NewCandle(series.NewTimePeriod(time.Now(), time.Minute))
		candle.ClosePrice = decimal.New(1)

		err := ts.AddCandleErr(candle)

		assert.NoError(t, err)
		assert.Len(t, ts.Candles, 1)
	})

	t.Run("Returns error if candle is before last candle", func(t *testing.T) {
		ts := series.NewTimeSeries()

		now := time.Now()
		candle := series.NewCandle(series.NewTimePeriod(now, time.Minute))
		candle.ClosePrice = decimal.New(1)

		err := ts.AddCandleErr(candle)

		assert.NoError(t, err)

		then := now.Add(-time.Minute * 10)
		nextCandle := series.NewCandle(series.NewTimePeriod(then, time.Minute))
		nextCandle.ClosePrice = decimal.New(2)

		err = ts.AddCandleErr(nextCandle)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not after last candle")
		assert.Len(t, ts.Candles, 1)
	})
}

func TestTimeSeries_LastCandle(t *testing.T) {
	ts := series.NewTimeSeries()

	now := time.Now()
	candle := series.NewCandle(series.NewTimePeriod(now, time.Minute))
	candle.ClosePrice = decimal.New(1)

	ts.AddCandle(candle)

	assert.EqualValues(t, now.UnixNano(), ts.LastCandle().Period.Start.UnixNano())
	assert.EqualValues(t, 1, ts.LastCandle().ClosePrice.Float())

	next := time.Now().Add(time.Minute)
	newCandle := series.NewCandle(series.NewTimePeriod(next, time.Minute))
	newCandle.ClosePrice = decimal.New(2)

	ts.AddCandle(newCandle)

	assert.Len(t, ts.Candles, 2)

	assert.EqualValues(t, next.UnixNano(), ts.LastCandle().Period.Start.UnixNano())
	assert.EqualValues(t, 2, ts.LastCandle().ClosePrice.Float())
}

func TestTimeSeries_LastIndex(t *testing.T) {
	ts := series.NewTimeSeries()

	candle := series.NewCandle(series.NewTimePeriod(time.Now(), time.Minute))
	ts.AddCandle(candle)

	assert.EqualValues(t, 0, ts.LastIndex())

	candle = series.NewCandle(series.NewTimePeriod(time.Now().Add(time.Minute), time.Minute))
	ts.AddCandle(candle)

	assert.EqualValues(t, 1, ts.LastIndex())
}

func TestTimeSeries_GetCandlePair(t *testing.T) {
	ts := series.NewTimeSeries()
	first := series.NewCandle(series.NewTimePeriod(time.Now(), time.Minute))
	second := series.NewCandle(series.NewTimePeriod(time.Now().Add(time.Minute), time.Minute))
	ts.AddCandle(first)
	ts.AddCandle(second)

	current, previous := ts.GetCandlePair(1)
	assert.Same(t, second, current)
	assert.Same(t, first, previous)

	current, previous = ts.GetCandlePair(0)
	assert.Same(t, first, current)
	assert.Nil(t, previous)

	current, previous = ts.GetCandlePair(2)
	assert.Nil(t, current)
	assert.Nil(t, previous)
}

func TestTimeSeries_CandleRange(t *testing.T) {
	ts := series.NewTimeSeries()
	first := series.NewCandle(series.NewTimePeriod(time.Now(), time.Minute))
	second := series.NewCandle(series.NewTimePeriod(time.Now().Add(time.Minute), time.Minute))
	ts.AddCandle(first)
	ts.AddCandle(second)

	rangeCandles := ts.CandleRange(-1, 10)
	assert.Len(t, rangeCandles, 2)
	assert.Same(t, first, rangeCandles[0])
	assert.Same(t, second, rangeCandles[1])
	assert.Empty(t, ts.CandleRange(2, 3))

	highest, lowest, ok := ts.HighLow(0, 2)
	assert.True(t, ok)
	assert.Equal(t, first.MaxPrice, highest)
	assert.Equal(t, first.MinPrice, lowest)

	highest, lowest, close, ok := ts.HighLowClose(0, 2)
	assert.True(t, ok)
	assert.Equal(t, highest, first.MaxPrice)
	assert.Equal(t, lowest, first.MinPrice)
	assert.Equal(t, second.ClosePrice, close)

	_, _, _, ok = ts.HighLowClose(0, 3)
	assert.False(t, ok)
}
