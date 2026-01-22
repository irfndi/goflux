package indicators

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
)

func TestVWAP(t *testing.T) {
	s := series.NewTimeSeries()

	// Price 100, Vol 10 -> PV=1000, V=10 -> VWAP=100
	s.AddCandle(&series.Candle{
		MaxPrice:   decimal.New(105),
		MinPrice:   decimal.New(95),
		ClosePrice: decimal.New(100),
		Volume:     decimal.New(10),
	})

	// Price 110, Vol 20 -> PV=2200, V=20 -> Total PV=3200, V=30 -> VWAP=106.66
	s.AddCandle(&series.Candle{
		MaxPrice:   decimal.New(115),
		MinPrice:   decimal.New(105),
		ClosePrice: decimal.New(110),
		Volume:     decimal.New(20),
	})

	vwap := NewVWAPIndicator(s)

	val0 := vwap.Calculate(0)
	if val0.Float() != 100 {
		t.Errorf("VWAP(0) = %v, want 100", val0)
	}

	val1 := vwap.Calculate(1)
	if val1.FormattedString(2) != "106.67" {
		t.Errorf("VWAP(1) = %v, want 106.67", val1)
	}
}

func TestSuperTrend(t *testing.T) {
	s := series.NewTimeSeries()

	// Constant prices -> ATR will be small
	for i := 0; i < 20; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(100 + float64(i)),
			MinPrice:   decimal.New(90 + float64(i)),
			ClosePrice: decimal.New(95 + float64(i)),
			Volume:     decimal.New(1000),
		})
	}

	st := NewSuperTrendIndicator(s, 10, 3.0)
	val := st.Calculate(19)

	if val.IsZero() {
		t.Errorf("SuperTrend should not be zero")
	}
}
