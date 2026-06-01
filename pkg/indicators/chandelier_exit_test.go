package indicators

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
)

func TestChandelierExitLongInsufficientData(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 10; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(105),
			MinPrice:   decimal.New(95),
			ClosePrice: decimal.New(100),
		})
	}

	ce := NewChandelierExitLong(s, 22, 22, 3.0)
	if !ce.Calculate(9).EQ(decimal.ZERO) {
		t.Errorf("ChandelierExitLong with insufficient data should return ZERO, got %v", ce.Calculate(9))
	}
}

func TestChandelierExitShortInsufficientData(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 10; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(105),
			MinPrice:   decimal.New(95),
			ClosePrice: decimal.New(100),
		})
	}

	ce := NewChandelierExitShort(s, 22, 22, 3.0)
	if !ce.Calculate(9).EQ(decimal.ZERO) {
		t.Errorf("ChandelierExitShort with insufficient data should return ZERO, got %v", ce.Calculate(9))
	}
}

func TestChandelierExitLongCalculation(t *testing.T) {
	s := series.NewTimeSeries()
	// Create 25 constant candles: high=105, low=95, close=100
	// ATR(22) will be 10 (true range is consistently 10)
	// Highest High over 22 periods = 105
	// Chandelier Exit Long = 105 - 3.0 * 10 = 75
	for i := 0; i < 25; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(105),
			MinPrice:   decimal.New(95),
			ClosePrice: decimal.New(100),
		})
	}

	ce := NewChandelierExitLong(s, 22, 22, 3.0)
	got := ce.Calculate(24)
	want := decimal.New(75)

	if got.Sub(want).Abs().GT(decimal.New(0.0001)) {
		t.Errorf("ChandelierExitLong(24) = %v, want %v", got, want)
	}
}

func TestChandelierExitShortCalculation(t *testing.T) {
	s := series.NewTimeSeries()
	// Create 25 constant candles: high=105, low=95, close=100
	// ATR(22) will be 10
	// Lowest Low over 22 periods = 95
	// Chandelier Exit Short = 95 + 3.0 * 10 = 125
	for i := 0; i < 25; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(105),
			MinPrice:   decimal.New(95),
			ClosePrice: decimal.New(100),
		})
	}

	ce := NewChandelierExitShort(s, 22, 22, 3.0)
	got := ce.Calculate(24)
	want := decimal.New(125)

	if got.Sub(want).Abs().GT(decimal.New(0.0001)) {
		t.Errorf("ChandelierExitShort(24) = %v, want %v", got, want)
	}
}

func TestChandelierExitLongWithTrendingPrices(t *testing.T) {
	s := series.NewTimeSeries()
	// Ascending prices: high = 100+i, low = 90+i, close = 95+i
	for i := 0; i < 30; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(float64(100 + i)),
			MinPrice:   decimal.New(float64(90 + i)),
			ClosePrice: decimal.New(float64(95 + i)),
		})
	}

	ce := NewChandelierExitLong(s, 10, 10, 2.0)
	val := ce.Calculate(29)

	// Highest high in last 10 periods = 129 (at index 29)
	// ATR should be positive
	// Exit should be less than highest high
	if val.IsZero() {
		t.Errorf("ChandelierExitLong should not be zero for valid index")
	}

	highestHigh := decimal.New(129)
	if val.GTE(highestHigh) {
		t.Errorf("ChandelierExitLong(%v) should be below highest high (%v)", val, highestHigh)
	}
}

func TestChandelierExitShortWithTrendingPrices(t *testing.T) {
	s := series.NewTimeSeries()
	// Ascending prices
	for i := 0; i < 30; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(float64(100 + i)),
			MinPrice:   decimal.New(float64(90 + i)),
			ClosePrice: decimal.New(float64(95 + i)),
		})
	}

	ce := NewChandelierExitShort(s, 10, 10, 2.0)
	val := ce.Calculate(29)

	// Lowest low in last 10 periods = 119 (at index 29)
	// Exit should be greater than lowest low
	if val.IsZero() {
		t.Errorf("ChandelierExitShort should not be zero for valid index")
	}

	lowestLow := decimal.New(119)
	if val.LTE(lowestLow) {
		t.Errorf("ChandelierExitShort(%v) should be above lowest low (%v)", val, lowestLow)
	}
}

func TestChandelierExitDefaults(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 25; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(105),
			MinPrice:   decimal.New(95),
			ClosePrice: decimal.New(100),
		})
	}

	long := NewDefaultChandelierExitLong(s)
	short := NewDefaultChandelierExitShort(s)

	if long.Calculate(24).IsZero() {
		t.Errorf("Default ChandelierExitLong should not be zero")
	}
	if short.Calculate(24).IsZero() {
		t.Errorf("Default ChandelierExitShort should not be zero")
	}
}

func TestChandelierExitPanicInvalidPeriod(t *testing.T) {
	s := series.NewTimeSeries()
	s.AddCandle(&series.Candle{MaxPrice: decimal.New(105), MinPrice: decimal.New(95), ClosePrice: decimal.New(100)})

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewChandelierExitLong with period=0 should panic")
		}
	}()
	_ = NewChandelierExitLong(s, 0, 10, 3.0)
}

func TestChandelierExitPanicInvalidATRWindow(t *testing.T) {
	s := series.NewTimeSeries()
	s.AddCandle(&series.Candle{MaxPrice: decimal.New(105), MinPrice: decimal.New(95), ClosePrice: decimal.New(100)})

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewChandelierExitLong with atrWindow=1 should panic")
		}
	}()
	_ = NewChandelierExitLong(s, 10, 1, 3.0)
}
