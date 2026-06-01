package trading

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/series"
)

func enterLong(record *TradingRecord, price decimal.Decimal) {
	order := NewOrderDetail(BUY, MarketOrder, "TEST", decimal.New(1))
	order.Fill(price, decimal.New(1))
	record.Operate(*order)
}

func enterShort(record *TradingRecord, price decimal.Decimal) {
	order := NewOrderDetail(SELL, MarketOrder, "TEST", decimal.New(1))
	order.Fill(price, decimal.New(1))
	record.Operate(*order)
}

func TestChandelierExitLongRuleNotSatisfiedWhenPositionClosed(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 25; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(105),
			MinPrice:   decimal.New(95),
			ClosePrice: decimal.New(100),
		})
	}

	rule := NewDefaultChandelierExitLongRule(s)
	record := NewTradingRecord()

	if rule.IsSatisfied(24, record) {
		t.Errorf("ChandelierExitLongRule should not be satisfied when no position is open")
	}
}

func TestChandelierExitLongRuleSatisfied(t *testing.T) {
	s := series.NewTimeSeries()
	// Create 25 constant candles
	for i := 0; i < 25; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(105),
			MinPrice:   decimal.New(95),
			ClosePrice: decimal.New(100),
		})
	}

	// The long Chandelier Exit with defaults on this series is around 75
	// If price drops to 70, rule should trigger
	s.AddCandle(&series.Candle{
		MaxPrice:   decimal.New(75),
		MinPrice:   decimal.New(65),
		ClosePrice: decimal.New(70),
	})

	rule := NewDefaultChandelierExitLongRule(s)
	record := NewTradingRecord()
	enterLong(record, decimal.New(100))

	if !rule.IsSatisfied(25, record) {
		t.Errorf("ChandelierExitLongRule should be satisfied when close drops below exit level")
	}
}

func TestChandelierExitLongRuleNotSatisfiedAboveExit(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 25; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(105),
			MinPrice:   decimal.New(95),
			ClosePrice: decimal.New(100),
		})
	}

	rule := NewDefaultChandelierExitLongRule(s)
	record := NewTradingRecord()
	enterLong(record, decimal.New(100))

	if rule.IsSatisfied(24, record) {
		t.Errorf("ChandelierExitLongRule should not be satisfied when close is above exit level")
	}
}

func TestChandelierExitShortRuleNotSatisfiedWhenPositionClosed(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 25; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(105),
			MinPrice:   decimal.New(95),
			ClosePrice: decimal.New(100),
		})
	}

	rule := NewDefaultChandelierExitShortRule(s)
	record := NewTradingRecord()

	if rule.IsSatisfied(24, record) {
		t.Errorf("ChandelierExitShortRule should not be satisfied when no position is open")
	}
}

func TestChandelierExitShortRuleSatisfied(t *testing.T) {
	s := series.NewTimeSeries()
	// Create 25 constant candles
	for i := 0; i < 25; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(105),
			MinPrice:   decimal.New(95),
			ClosePrice: decimal.New(100),
		})
	}

	// The short Chandelier Exit with defaults on this series is around 125
	// If price rises to 130, rule should trigger
	s.AddCandle(&series.Candle{
		MaxPrice:   decimal.New(135),
		MinPrice:   decimal.New(125),
		ClosePrice: decimal.New(130),
	})

	rule := NewDefaultChandelierExitShortRule(s)
	record := NewTradingRecord()
	enterShort(record, decimal.New(100))

	if !rule.IsSatisfied(25, record) {
		t.Errorf("ChandelierExitShortRule should be satisfied when close rises above exit level")
	}
}

func TestChandelierExitShortRuleNotSatisfiedBelowExit(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 25; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(105),
			MinPrice:   decimal.New(95),
			ClosePrice: decimal.New(100),
		})
	}

	rule := NewDefaultChandelierExitShortRule(s)
	record := NewTradingRecord()
	enterShort(record, decimal.New(100))

	if rule.IsSatisfied(24, record) {
		t.Errorf("ChandelierExitShortRule should not be satisfied when close is below exit level")
	}
}

func TestChandelierExitLongRuleWithCustomParams(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 15; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(105),
			MinPrice:   decimal.New(95),
			ClosePrice: decimal.New(100),
		})
	}

	rule := NewChandelierExitLongRule(s, 10, 10, 2.0)
	record := NewTradingRecord()
	enterLong(record, decimal.New(100))

	// With period=10, atrWindow=10, we need at least 10 candles for validity
	if rule.IsSatisfied(9, record) {
		t.Errorf("ChandelierExitLongRule should not be satisfied at boundary index")
	}
}

func TestChandelierExitIntegrationWithStrategy(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 30; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(105),
			MinPrice:   decimal.New(95),
			ClosePrice: decimal.New(100),
		})
	}

	entryRule := NewOverIndicatorRule(
		indicators.NewClosePriceIndicator(s),
		indicators.NewConstantIndicator(0),
	)
	exitRule := NewDefaultChandelierExitLongRule(s)

	strategy, err := NewRuleStrategy(entryRule, exitRule, 0)
	if err != nil {
		t.Fatalf("NewRuleStrategy failed: %v", err)
	}

	record := NewTradingRecord()
	for i := 0; i < 30; i++ {
		strategy.ShouldEnter(i, record)
		strategy.ShouldExit(i, record)
	}

	// Strategy should be constructible and not panic
	if strategy.EntryRule == nil || strategy.ExitRule == nil {
		t.Errorf("Strategy with ChandelierExit should have valid rules")
	}
}
