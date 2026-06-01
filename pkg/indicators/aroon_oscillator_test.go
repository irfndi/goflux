package indicators

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/testutils"
)

func TestAroonOscillatorInsufficientData(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 5; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(float64(10 + i)),
			MinPrice:   decimal.New(float64(5 + i)),
			ClosePrice: decimal.New(float64(7 + i)),
		})
	}

	osc := NewAroonOscillatorFromSeries(s, 10)
	if !osc.Calculate(4).EQ(decimal.ZERO) {
		t.Errorf("AroonOscillator with insufficient data should return ZERO")
	}
}

func TestAroonOscillatorStrongUptrend(t *testing.T) {
	// Prices always increasing: every candle has a new high
	// Aroon Up = 100 (new high at current index)
	// Aroon Down = 20 (lowest was 4 periods ago in window)
	// Oscillator = 100 - 20 = 80
	s := testutils.MockTimeSeriesFl(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	osc := NewAroonOscillatorFromSeries(s, 5)
	val := osc.Calculate(9)

	if val.Sub(decimal.New(80)).Abs().GT(decimal.New(0.0001)) {
		t.Errorf("AroonOscillator in strong uptrend = %v, want ~80", val)
	}
}

func TestAroonOscillatorStrongDowntrend(t *testing.T) {
	// Prices always decreasing: every candle has a new low
	// Aroon Up = 20 (highest max was 4 periods ago in window)
	// Aroon Down = 100 (new low at current index)
	// Oscillator = 20 - 100 = -80
	s := testutils.MockTimeSeriesFl(10, 9, 8, 7, 6, 5, 4, 3, 2, 1)

	osc := NewAroonOscillatorFromSeries(s, 5)
	val := osc.Calculate(9)

	if val.Sub(decimal.New(-80)).Abs().GT(decimal.New(0.0001)) {
		t.Errorf("AroonOscillator in strong downtrend = %v, want ~-80", val)
	}
}

func TestAroonOscillatorFlatTrend(t *testing.T) {
	// Constant prices: Aroon Up and Down will both decrease from 100
	// After window periods, both are 0, so oscillator is 0
	s := testutils.MockTimeSeriesFl(5, 5, 5, 5, 5, 5, 5, 5, 5, 5)

	osc := NewAroonOscillatorFromSeries(s, 5)
	val := osc.Calculate(9)

	if val.Sub(decimal.New(0)).Abs().GT(decimal.New(0.0001)) {
		t.Errorf("AroonOscillator in flat trend = %v, want ~0", val)
	}
}

func TestAroonOscillatorEqualsUpMinusDown(t *testing.T) {
	s := testutils.MockTimeSeriesFl(5, 4, 3, 2, 3, 4, 5)

	high := NewHighPriceIndicator(s)
	low := NewLowPriceIndicator(s)

	up := NewAroonUpIndicator(high, 4)
	down := NewAroonDownIndicator(low, 4)
	osc := NewAroonOscillator(high, 4)

	for i := 3; i < 7; i++ {
		expected := up.Calculate(i).Sub(down.Calculate(i))
		got := osc.Calculate(i)
		if got.Sub(expected).Abs().GT(decimal.New(0.0001)) {
			t.Errorf("AroonOscillator(%d) = %v, want %v (Up - Down)", i, got, expected)
		}
	}
}

func TestAroonOscillatorRange(t *testing.T) {
	s := testutils.RandomTimeSeries(50)
	osc := NewAroonOscillatorFromSeries(s, 14)

	for i := 13; i < 50; i++ {
		val := osc.Calculate(i)
		if val.GT(decimal.New(100)) {
			t.Errorf("AroonOscillator(%d) = %v, exceeds max +100", i, val)
		}
		if val.LT(decimal.New(-100)) {
			t.Errorf("AroonOscillator(%d) = %v, exceeds min -100", i, val)
		}
	}
}

func TestAroonOscillatorPanicInvalidWindow(t *testing.T) {
	s := series.NewTimeSeries()
	s.AddCandle(&series.Candle{MaxPrice: decimal.New(105), MinPrice: decimal.New(95), ClosePrice: decimal.New(100)})

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewAroonOscillator with window=0 should panic")
		}
	}()
	_ = NewAroonOscillatorFromSeries(s, 0)
}
