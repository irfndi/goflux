package indicators

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/testutils"
)

func TestDonchianUpperBand(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12, 11, 15, 13, 14, 16, 14, 15, 17)

	upper := NewDonchianUpperBandIndicator(s, 5)

	// At index 4, highest high in window [0..4] is 15 (index 3)
	val := upper.Calculate(4)
	if !val.EQ(decimal.New(16)) {
		t.Errorf("DonchianUpperBand(4) = %v, want 16", val)
	}

	// At index 9, highest high in window [5..9] is 17 (index 9)
	val = upper.Calculate(9)
	if !val.EQ(decimal.New(18)) {
		t.Errorf("DonchianUpperBand(9) = %v, want 18", val)
	}
}

func TestDonchianLowerBand(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 8, 11, 9, 13, 14, 12, 15, 16, 14)

	lower := NewDonchianLowerBandIndicator(s, 5)

	// At index 4, lowest low in window [0..4] is 7 (index 1: MinPrice=8-1=7)
	val := lower.Calculate(4)
	if !val.EQ(decimal.New(7)) {
		t.Errorf("DonchianLowerBand(4) = %v, want 7", val)
	}
}

func TestDonchianMiddleBand(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12, 11, 15, 13, 14, 16, 14, 15, 17)

	middle := NewDonchianMiddleBandIndicator(s, 5)

	// Upper at 4 = 16, Lower at 4 = 10-1=9 → Middle = (16+9)/2 = 12.5
	val := middle.Calculate(4)
	expected := decimal.New(12.5)
	if val.Sub(expected).Abs().GT(decimal.New(0.0001)) {
		t.Errorf("DonchianMiddleBand(4) = %v, want ~12.5", val)
	}
}

func TestDonchianBandsInsufficientData(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 3; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice: decimal.New(float64(10 + i)),
			MinPrice: decimal.New(float64(5 + i)),
		})
	}

	upper := NewDonchianUpperBandIndicator(s, 5)
	lower := NewDonchianLowerBandIndicator(s, 5)

	if !upper.Calculate(2).EQ(decimal.ZERO) {
		t.Errorf("DonchianUpperBand with insufficient data should be ZERO")
	}
	if !lower.Calculate(2).EQ(decimal.ZERO) {
		t.Errorf("DonchianLowerBand with insufficient data should be ZERO")
	}
}

func TestDonchianBandsPanicInvalidWindow(t *testing.T) {
	s := series.NewTimeSeries()
	s.AddCandle(&series.Candle{MaxPrice: decimal.New(105), MinPrice: decimal.New(95)})

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewDonchianUpperBandIndicator with window=0 should panic")
		}
	}()
	_ = NewDonchianUpperBandIndicator(s, 0)
}

func TestDonchianBandsOutOfBounds(t *testing.T) {
	s := series.NewTimeSeries()
	s.AddCandle(&series.Candle{MaxPrice: decimal.New(105), MinPrice: decimal.New(95)})

	upper := NewDonchianUpperBandIndicator(s, 5)
	lower := NewDonchianLowerBandIndicator(s, 5)

	if !upper.Calculate(-1).EQ(decimal.ZERO) {
		t.Errorf("DonchianUpperBand(-1) should be ZERO")
	}
	if !lower.Calculate(10).EQ(decimal.ZERO) {
		t.Errorf("DonchianLowerBand(10) out of bounds should be ZERO")
	}
}

func TestDonchianNilSeries(t *testing.T) {
	upper := NewDonchianUpperBandIndicator(nil, 5)
	lower := NewDonchianLowerBandIndicator(nil, 5)

	if !upper.Calculate(0).EQ(decimal.ZERO) {
		t.Errorf("DonchianUpperBand with nil series should be ZERO")
	}
	if !lower.Calculate(0).EQ(decimal.ZERO) {
		t.Errorf("DonchianLowerBand with nil series should be ZERO")
	}
}
