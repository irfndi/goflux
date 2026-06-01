package indicators

import (
	"math"
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
)

func TestLinearRegressionForecast(t *testing.T) {
	s := series.NewTimeSeries()
	values := []float64{10, 15, 30, 35, 50}
	for _, v := range values {
		s.AddCandle(&series.Candle{ClosePrice: decimal.New(v)})
	}

	lr := NewLinearRegressionIndicator(NewClosePriceIndicator(s), 5)
	// Slope=10, Intercept=8, Forecast at x=4 = 10*4+8 = 48
	val := lr.Calculate(4)
	expected := decimal.New(48)
	if val.Sub(expected).Abs().GT(decimal.New(0.001)) {
		t.Errorf("LinearRegression(4) = %v, want ~48", val)
	}
}

func TestLinearRegressionSlope(t *testing.T) {
	s := series.NewTimeSeries()
	values := []float64{10, 15, 30, 35, 50}
	for _, v := range values {
		s.AddCandle(&series.Candle{ClosePrice: decimal.New(v)})
	}

	lr := NewLinearRegressionSlopeIndicator(NewClosePriceIndicator(s), 5)
	val := lr.Calculate(4)
	expected := decimal.New(10)
	if val.Sub(expected).Abs().GT(decimal.New(0.001)) {
		t.Errorf("LinearRegressionSlope(4) = %v, want ~10", val)
	}
}

func TestLinearRegressionIntercept(t *testing.T) {
	s := series.NewTimeSeries()
	values := []float64{10, 15, 30, 35, 50}
	for _, v := range values {
		s.AddCandle(&series.Candle{ClosePrice: decimal.New(v)})
	}

	lr := NewLinearRegressionInterceptIndicator(NewClosePriceIndicator(s), 5)
	val := lr.Calculate(4)
	expected := decimal.New(8)
	if val.Sub(expected).Abs().GT(decimal.New(0.001)) {
		t.Errorf("LinearRegressionIntercept(4) = %v, want ~8", val)
	}
}

func TestLinearRegressionAngle(t *testing.T) {
	s := series.NewTimeSeries()
	values := []float64{10, 15, 30, 35, 50}
	for _, v := range values {
		s.AddCandle(&series.Candle{ClosePrice: decimal.New(v)})
	}

	lr := NewLinearRegressionAngleIndicator(NewClosePriceIndicator(s), 5)
	val := lr.Calculate(4)
	// atan(10) * 180/pi ≈ 84.289
	expected := decimal.New(84.289)
	if val.Sub(expected).Abs().GT(decimal.New(0.001)) {
		t.Errorf("LinearRegressionAngle(4) = %v, want ~84.289", val)
	}
}

func TestLinearRegressionStandardError(t *testing.T) {
	s := series.NewTimeSeries()
	values := []float64{10, 15, 30, 35, 50}
	for _, v := range values {
		s.AddCandle(&series.Candle{ClosePrice: decimal.New(v)})
	}

	lr := NewStandardErrorIndicator(NewClosePriceIndicator(s), 5)
	val := lr.Calculate(4)
	// SE = sqrt(30/3) = sqrt(10) ≈ 3.1623
	expected := decimal.New(math.Sqrt(10))
	if val.Sub(expected).Abs().GT(decimal.New(0.001)) {
		t.Errorf("StandardError(4) = %v, want ~3.162", val)
	}
}

func TestLinearRegressionChannel(t *testing.T) {
	s := series.NewTimeSeries()
	values := []float64{10, 15, 30, 35, 50}
	for _, v := range values {
		s.AddCandle(&series.Candle{ClosePrice: decimal.New(v)})
	}

	mid, upper, lower := NewLinearRegressionChannel(NewClosePriceIndicator(s), 5, 2.0)
	// Forecast = 48, SE = sqrt(10) ≈ 3.162
	// Upper = 48 + 2*3.162 = 54.325
	// Lower = 48 - 2*3.162 = 41.675

	midVal := mid.Calculate(4)
	upperVal := upper.Calculate(4)
	lowerVal := lower.Calculate(4)

	if midVal.Sub(decimal.New(48)).Abs().GT(decimal.New(0.001)) {
		t.Errorf("LRC mid(4) = %v, want ~48", midVal)
	}

	expectedUpper := decimal.New(48 + 2*math.Sqrt(10))
	if upperVal.Sub(expectedUpper).Abs().GT(decimal.New(0.001)) {
		t.Errorf("LRC upper(4) = %v, want ~54.325", upperVal)
	}

	expectedLower := decimal.New(48 - 2*math.Sqrt(10))
	if lowerVal.Sub(expectedLower).Abs().GT(decimal.New(0.001)) {
		t.Errorf("LRC lower(4) = %v, want ~41.675", lowerVal)
	}
}

func TestLinearRegressionInsufficientData(t *testing.T) {
	s := series.NewTimeSeries()
	s.AddCandle(&series.Candle{ClosePrice: decimal.New(10)})

	lr := NewLinearRegressionIndicator(NewClosePriceIndicator(s), 5)
	if !lr.Calculate(0).EQ(decimal.ZERO) {
		t.Errorf("LinearRegression with insufficient data should be ZERO")
	}
}

func TestLinearRegressionNilIndicator(t *testing.T) {
	lr := NewLinearRegressionIndicator(nil, 5)
	if !lr.Calculate(0).EQ(decimal.ZERO) {
		t.Errorf("LinearRegression with nil indicator should be ZERO")
	}
}

func TestLinearRegressionPanicInvalidWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewLinearRegressionIndicator with window=1 should panic")
		}
	}()
	_ = NewLinearRegressionIndicator(nil, 1)
}

func TestLinearRegressionChannelPanicInvalidWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewLinearRegressionChannel with window=1 should panic")
		}
	}()
	_, _, _ = NewLinearRegressionChannel(nil, 1, 2.0)
}
