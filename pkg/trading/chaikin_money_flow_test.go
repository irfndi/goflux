package trading

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/series"
)

func TestChaikinMoneyFlowOverLevelRule(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 20; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(110),
			MinPrice:   decimal.New(100),
			ClosePrice: decimal.New(109), // near high
			Volume:     decimal.New(1000),
		})
	}

	cmf := indicators.NewChaikinMoneyFlowIndicator(s, 10)
	rule := NewChaikinMoneyFlowOverLevelRule(cmf, 0)
	record := NewTradingRecord()

	if !rule.IsSatisfied(19, record) {
		t.Errorf("ChaikinMoneyFlowOverLevelRule should be satisfied with strong buying pressure")
	}
}

func TestChaikinMoneyFlowUnderLevelRule(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 20; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(110),
			MinPrice:   decimal.New(100),
			ClosePrice: decimal.New(101), // near low
			Volume:     decimal.New(1000),
		})
	}

	cmf := indicators.NewChaikinMoneyFlowIndicator(s, 10)
	rule := NewChaikinMoneyFlowUnderLevelRule(cmf, 0)
	record := NewTradingRecord()

	if !rule.IsSatisfied(19, record) {
		t.Errorf("ChaikinMoneyFlowUnderLevelRule should be satisfied with strong selling pressure")
	}
}

func TestChaikinMoneyFlowBullishRule(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 20; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(110),
			MinPrice:   decimal.New(100),
			ClosePrice: decimal.New(109),
			Volume:     decimal.New(1000),
		})
	}

	rule := NewChaikinMoneyFlowBullishRule(s, 10)
	record := NewTradingRecord()

	if !rule.IsSatisfied(19, record) {
		t.Errorf("ChaikinMoneyFlowBullishRule should be satisfied in uptrend")
	}
}

func TestChaikinMoneyFlowBearishRule(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 20; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(110),
			MinPrice:   decimal.New(100),
			ClosePrice: decimal.New(101),
			Volume:     decimal.New(1000),
		})
	}

	rule := NewChaikinMoneyFlowBearishRule(s, 10)
	record := NewTradingRecord()

	if !rule.IsSatisfied(19, record) {
		t.Errorf("ChaikinMoneyFlowBearishRule should be satisfied in downtrend")
	}
}

func TestChaikinMoneyFlowAtThreshold(t *testing.T) {
	// Neutral trend: close at midpoint → CMF ≈ 0.
	s := series.NewTimeSeries()
	for i := 0; i < 10; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(110),
			MinPrice:   decimal.New(100),
			ClosePrice: decimal.New(105), // midpoint
			Volume:     decimal.New(1000),
		})
	}

	cmf := indicators.NewChaikinMoneyFlowIndicator(s, 5)
	record := NewTradingRecord()

	over := NewChaikinMoneyFlowOverLevelRule(cmf, 0)
	if over.IsSatisfied(9, record) {
		t.Errorf("ChaikinMoneyFlowOverLevelRule(level=0) should not be satisfied when CMF == 0")
	}

	under := NewChaikinMoneyFlowUnderLevelRule(cmf, 0)
	if under.IsSatisfied(9, record) {
		t.Errorf("ChaikinMoneyFlowUnderLevelRule(level=0) should not be satisfied when CMF == 0")
	}
}

func TestChaikinMoneyFlowInsufficientData(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 3; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(110),
			MinPrice:   decimal.New(100),
			ClosePrice: decimal.New(109),
			Volume:     decimal.New(1000),
		})
	}

	record := NewTradingRecord()

	bullish := NewChaikinMoneyFlowBullishRule(s, 5)
	if bullish.IsSatisfied(2, record) {
		t.Errorf("ChaikinMoneyFlowBullishRule should not be satisfied with insufficient data")
	}

	bearish := NewChaikinMoneyFlowBearishRule(s, 5)
	if bearish.IsSatisfied(2, record) {
		t.Errorf("ChaikinMoneyFlowBearishRule should not be satisfied with insufficient data")
	}
}
