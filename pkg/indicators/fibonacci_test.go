package indicators

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
)

func TestFibonacciRetracementLevels(t *testing.T) {
	high := decimal.New(100)
	low := decimal.New(50)
	f := NewFibonacciRetracement(high, low)
	levels := f.Levels()

	// Range = 50
	// 0%   = 50
	// 23.6% = 50 + 50*0.236 = 61.8
	// 38.2% = 50 + 50*0.382 = 69.1
	// 50%   = 50 + 50*0.5   = 75
	// 61.8% = 50 + 50*0.618 = 80.9
	// 78.6% = 50 + 50*0.786 = 89.3
	// 100%  = 100

	if !levels.Level0.EQ(decimal.New(50)) {
		t.Errorf("Level0 = %v, want 50", levels.Level0)
	}
	if !levels.Level100.EQ(decimal.New(100)) {
		t.Errorf("Level100 = %v, want 100", levels.Level100)
	}

	expected618 := decimal.New(80.9)
	if levels.Level618.Sub(expected618).Abs().GT(decimal.New(0.001)) {
		t.Errorf("Level618 = %v, want ~80.9", levels.Level618)
	}

	expected500 := decimal.New(75)
	if !levels.Level500.EQ(expected500) {
		t.Errorf("Level500 = %v, want 75", levels.Level500)
	}
}

func TestFibonacciRetracementFromSeries(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 10; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice: decimal.New(float64(100 + i)),
			MinPrice: decimal.New(float64(50 + i)),
		})
	}

	f := NewFibonacciRetracementFromSeries(s, 5, 9)
	// High in [5..9] = max(105,106,107,108,109) = 109
	// Low  in [5..9] = min(55,56,57,58,59) = 55
	// Range = 54
	// 50% = 55 + 27 = 82
	levels := f.Levels()
	expected500 := decimal.New(82)
	if levels.Level500.Sub(expected500).Abs().GT(decimal.New(0.001)) {
		t.Errorf("Level500 from series = %v, want ~82", levels.Level500)
	}
}

func TestFibonacciRetracementFromSeriesInsufficientData(t *testing.T) {
	s := series.NewTimeSeries()
	s.AddCandle(&series.Candle{MaxPrice: decimal.New(100), MinPrice: decimal.New(50)})

	f := NewFibonacciRetracementFromSeries(s, 5, 0)
	levels := f.Levels()
	if !levels.Level0.EQ(decimal.ZERO) {
		t.Errorf("Level0 with insufficient data should be ZERO")
	}
}

func TestFibonacciRetracementFromSeriesNil(t *testing.T) {
	f := NewFibonacciRetracementFromSeries(nil, 5, 0)
	levels := f.Levels()
	if !levels.Level0.EQ(decimal.ZERO) {
		t.Errorf("Level0 with nil series should be ZERO")
	}
}

func TestFibonacciRetracementIndicator(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 10; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice: decimal.New(float64(100 + i)),
			MinPrice: decimal.New(float64(50 + i)),
		})
	}

	ind := NewFibonacciRetracementIndicator(s, 5, 0.5)
	// At index 9: high=109, low=55, range=54, 50% = 55 + 27 = 82
	val := ind.Calculate(9)
	expected := decimal.New(82)
	if val.Sub(expected).Abs().GT(decimal.New(0.001)) {
		t.Errorf("FibonacciRetracementIndicator(9) = %v, want ~82", val)
	}
}

func TestFibonacciRetracementIndicatorInsufficientData(t *testing.T) {
	s := series.NewTimeSeries()
	s.AddCandle(&series.Candle{MaxPrice: decimal.New(100), MinPrice: decimal.New(50)})

	ind := NewFibonacciRetracementIndicator(s, 5, 0.618)
	if !ind.Calculate(0).EQ(decimal.ZERO) {
		t.Errorf("FibonacciRetracementIndicator with insufficient data should be ZERO")
	}
}

func TestFibonacciRetracementIndicatorPanicInvalidLookback(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewFibonacciRetracementIndicator with lookback=0 should panic")
		}
	}()
	_ = NewFibonacciRetracementIndicator(nil, 0, 0.5)
}

func TestFibonacciExtensionLevelsUp(t *testing.T) {
	high := decimal.New(100)
	low := decimal.New(50)
	f := NewFibonacciExtension(high, low)
	levels := f.LevelsUp()

	// Range = 50
	// 127.2% = 100 + 50*0.272 = 113.6
	// 161.8% = 100 + 50*0.618 = 130.9
	// 200%   = 100 + 50*1.0   = 150
	// 261.8% = 100 + 50*1.618 = 180.9
	// 300%   = 100 + 50*2.0   = 200
	// 423.6% = 100 + 50*3.236 = 261.8

	expected2000 := decimal.New(150)
	if !levels.Level2000.EQ(expected2000) {
		t.Errorf("Extension Level2000 = %v, want 150", levels.Level2000)
	}

	expected1618 := decimal.New(130.9)
	if levels.Level1618.Sub(expected1618).Abs().GT(decimal.New(0.001)) {
		t.Errorf("Extension Level1618 = %v, want ~130.9", levels.Level1618)
	}
}

func TestFibonacciExtensionLevelsDown(t *testing.T) {
	high := decimal.New(100)
	low := decimal.New(50)
	f := NewFibonacciExtension(high, low)
	levels := f.LevelsDown()

	// Range = 50
	// 127.2% = 50 - 50*0.272 = 36.4
	// 161.8% = 50 - 50*0.618 = 19.1
	// 200%   = 50 - 50*1.0   = 0
	// 261.8% = 50 - 50*1.618 = -30.9
	// 300%   = 50 - 50*2.0   = -50
	// 423.6% = 50 - 50*3.236 = -111.8

	expected2000 := decimal.New(0)
	if !levels.Level2000.EQ(expected2000) {
		t.Errorf("Extension Down Level2000 = %v, want 0", levels.Level2000)
	}
}

func TestFibonacciExtensionFromSeries(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 10; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice: decimal.New(float64(100 + i)),
			MinPrice: decimal.New(float64(50 + i)),
		})
	}

	f := NewFibonacciExtensionFromSeries(s, 5, 9)
	levels := f.LevelsUp()
	// High = 109, Low = 55, Range = 54
	// 200% = 109 + 54 = 163
	expected2000 := decimal.New(163)
	if levels.Level2000.Sub(expected2000).Abs().GT(decimal.New(0.001)) {
		t.Errorf("Extension Up Level2000 from series = %v, want ~163", levels.Level2000)
	}
}

func TestFibonacciExtensionFromSeriesInsufficientData(t *testing.T) {
	s := series.NewTimeSeries()
	s.AddCandle(&series.Candle{MaxPrice: decimal.New(100), MinPrice: decimal.New(50)})

	f := NewFibonacciExtensionFromSeries(s, 5, 0)
	levels := f.LevelsUp()
	if !levels.Level1272.EQ(decimal.ZERO) {
		t.Errorf("Extension Level1272 with insufficient data should be ZERO")
	}
}

func TestFibonacciExtensionFromSeriesNil(t *testing.T) {
	f := NewFibonacciExtensionFromSeries(nil, 5, 0)
	levels := f.LevelsUp()
	if !levels.Level1272.EQ(decimal.ZERO) {
		t.Errorf("Extension Level1272 with nil series should be ZERO")
	}
}

func TestFibonacciRetracementInvertedHighLow(t *testing.T) {
	// When high < low, values should be swapped
	f := NewFibonacciRetracement(decimal.New(50), decimal.New(100))
	levels := f.Levels()

	if !levels.Level0.EQ(decimal.New(50)) {
		t.Errorf("Level0 after swap = %v, want 50", levels.Level0)
	}
	if !levels.Level100.EQ(decimal.New(100)) {
		t.Errorf("Level100 after swap = %v, want 100", levels.Level100)
	}
}

func TestFibonacciExtensionInvertedHighLow(t *testing.T) {
	// When high < low, values should be swapped
	f := NewFibonacciExtension(decimal.New(50), decimal.New(100))
	levels := f.LevelsUp()

	expected2000 := decimal.New(150)
	if !levels.Level2000.EQ(expected2000) {
		t.Errorf("Extension Level2000 after swap = %v, want 150", levels.Level2000)
	}
}
