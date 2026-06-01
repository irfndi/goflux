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
