package trading

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
)

func TestDonchianBreakoutUpperRule(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 9; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(10),
			MinPrice:   decimal.New(5),
			ClosePrice: decimal.New(7),
		})
	}
	s.AddCandle(&series.Candle{
		MaxPrice:   decimal.New(15),
		MinPrice:   decimal.New(10),
		ClosePrice: decimal.New(15),
	})

	rule := NewDonchianBreakoutUpperRule(s, 5)
	record := NewTradingRecord()

	// Upper band at index 9 = highest high in [5..9] = 15; Close = 15
	if !rule.IsSatisfied(9, record) {
		t.Errorf("DonchianBreakoutUpperRule should be satisfied when close >= upper")
	}
}

func TestDonchianBreakoutLowerRule(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 9; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(20),
			MinPrice:   decimal.New(15),
			ClosePrice: decimal.New(17),
		})
	}
	s.AddCandle(&series.Candle{
		MaxPrice:   decimal.New(10),
		MinPrice:   decimal.New(5),
		ClosePrice: decimal.New(5),
	})

	rule := NewDonchianBreakoutLowerRule(s, 5)
	record := NewTradingRecord()

	// Lower band at index 9 = lowest low in [5..9] = 5; Close = 5
	if !rule.IsSatisfied(9, record) {
		t.Errorf("DonchianBreakoutLowerRule should be satisfied when close <= lower")
	}
}

func TestDonchianChannelWidthRule(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 10; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(float64(110)),
			MinPrice:   decimal.New(float64(100)),
			ClosePrice: decimal.New(float64(105)),
		})
	}

	rule := NewDonchianChannelWidthRule(s, 5, 5)
	record := NewTradingRecord()

	// Width = 111 - 99 = 12 > 5
	if !rule.IsSatisfied(9, record) {
		t.Errorf("DonchianChannelWidthRule should be satisfied when width > threshold")
	}
}

func TestDonchianChannelWidthRuleNotSatisfied(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 10; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(float64(102)),
			MinPrice:   decimal.New(float64(100)),
			ClosePrice: decimal.New(float64(101)),
		})
	}

	rule := NewDonchianChannelWidthRule(s, 5, 5)
	record := NewTradingRecord()

	// Width = 103 - 99 = 4 < 5
	if rule.IsSatisfied(9, record) {
		t.Errorf("DonchianChannelWidthRule should not be satisfied when width < threshold")
	}
}
