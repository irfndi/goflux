package trading

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/series"
)

func TestAroonOscillatorOverZeroRule(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 20; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(float64(10 + i)),
			MinPrice:   decimal.New(float64(5 + i)),
			ClosePrice: decimal.New(float64(7 + i)),
		})
	}

	osc := indicators.NewAroonOscillatorFromSeries(s, 5)
	rule := NewAroonOscillatorOverLevelRule(osc, 0)
	record := NewTradingRecord()

	if !rule.IsSatisfied(19, record) {
		t.Errorf("AroonOscillatorOverLevelRule with level=0 should be satisfied in uptrend")
	}
}

func TestAroonOscillatorOverLevelRule(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 20; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(float64(10 + i)),
			MinPrice:   decimal.New(float64(5 + i)),
			ClosePrice: decimal.New(float64(7 + i)),
		})
	}

	osc := indicators.NewAroonOscillatorFromSeries(s, 5)
	rule := NewAroonOscillatorOverLevelRule(osc, 50)
	record := NewTradingRecord()

	if !rule.IsSatisfied(19, record) {
		t.Errorf("AroonOscillatorOverLevelRule should be satisfied when oscillator > 50")
	}
}

func TestAroonOscillatorUnderLevelRule(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 20; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(float64(20 - i)),
			MinPrice:   decimal.New(float64(15 - i)),
			ClosePrice: decimal.New(float64(17 - i)),
		})
	}

	osc := indicators.NewAroonOscillatorFromSeries(s, 5)
	rule := NewAroonOscillatorUnderLevelRule(osc, -50)
	record := NewTradingRecord()

	if !rule.IsSatisfied(19, record) {
		t.Errorf("AroonOscillatorUnderLevelRule should be satisfied when oscillator < -50")
	}
}

func TestAroonOscillatorBullishRule(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 20; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(float64(10 + i)),
			MinPrice:   decimal.New(float64(5 + i)),
			ClosePrice: decimal.New(float64(7 + i)),
		})
	}

	rule := NewAroonOscillatorBullishRule(s, 5)
	record := NewTradingRecord()

	if !rule.IsSatisfied(19, record) {
		t.Errorf("AroonOscillatorBullishRule should be satisfied in uptrend")
	}
}

func TestAroonOscillatorBearishRule(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 20; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(float64(20 - i)),
			MinPrice:   decimal.New(float64(15 - i)),
			ClosePrice: decimal.New(float64(17 - i)),
		})
	}

	rule := NewAroonOscillatorBearishRule(s, 5)
	record := NewTradingRecord()

	if !rule.IsSatisfied(19, record) {
		t.Errorf("AroonOscillatorBearishRule should be satisfied in downtrend")
	}
}

func TestAroonOscillatorAtThreshold(t *testing.T) {
	// Flat trend: oscillator == 0 at index 9.
	s := series.NewTimeSeries()
	for i := 0; i < 10; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(6),
			MinPrice:   decimal.New(4),
			ClosePrice: decimal.New(5),
		})
	}

	osc := indicators.NewAroonOscillatorFromSeries(s, 5)
	record := NewTradingRecord()

	over := NewAroonOscillatorOverLevelRule(osc, 0)
	if over.IsSatisfied(9, record) {
		t.Errorf("AroonOscillatorOverLevelRule(level=0) should not be satisfied when oscillator == 0")
	}

	under := NewAroonOscillatorUnderLevelRule(osc, 0)
	if under.IsSatisfied(9, record) {
		t.Errorf("AroonOscillatorUnderLevelRule(level=0) should not be satisfied when oscillator == 0")
	}
}

func TestAroonOscillatorInsufficientData(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 3; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(float64(10 + i)),
			MinPrice:   decimal.New(float64(5 + i)),
			ClosePrice: decimal.New(float64(7 + i)),
		})
	}

	record := NewTradingRecord()

	bullish := NewAroonOscillatorBullishRule(s, 5)
	if bullish.IsSatisfied(2, record) {
		t.Errorf("AroonOscillatorBullishRule should not be satisfied with insufficient data")
	}

	bearish := NewAroonOscillatorBearishRule(s, 5)
	if bearish.IsSatisfied(2, record) {
		t.Errorf("AroonOscillatorBearishRule should not be satisfied with insufficient data")
	}
}

func TestAroonOscillatorFlatTrend(t *testing.T) {
	// Flat trend yields oscillator ≈ 0.
	s := series.NewTimeSeries()
	for i := 0; i < 10; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(6),
			MinPrice:   decimal.New(4),
			ClosePrice: decimal.New(5),
		})
	}

	record := NewTradingRecord()

	bullish := NewAroonOscillatorBullishRule(s, 5)
	if bullish.IsSatisfied(9, record) {
		t.Errorf("AroonOscillatorBullishRule should not be satisfied in flat trend")
	}

	bearish := NewAroonOscillatorBearishRule(s, 5)
	if bearish.IsSatisfied(9, record) {
		t.Errorf("AroonOscillatorBearishRule should not be satisfied in flat trend")
	}

	over := NewAroonOscillatorOverLevelRule(indicators.NewAroonOscillatorFromSeries(s, 5), 50)
	if over.IsSatisfied(9, record) {
		t.Errorf("AroonOscillatorOverLevelRule(level=50) should not be satisfied in flat trend")
	}

	under := NewAroonOscillatorUnderLevelRule(indicators.NewAroonOscillatorFromSeries(s, 5), -50)
	if under.IsSatisfied(9, record) {
		t.Errorf("AroonOscillatorUnderLevelRule(level=-50) should not be satisfied in flat trend")
	}
}
