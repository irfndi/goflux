package indicators

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/testutils"
)

func TestTRIXInsufficientData(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 5; i++ {
		s.AddCandle(&series.Candle{
			ClosePrice: decimal.New(float64(100 + i)),
		})
	}

	trix := NewTRIXIndicatorFromSeries(s, 10)
	if !trix.Calculate(4).EQ(decimal.ZERO) {
		t.Errorf("TRIX with insufficient data should return ZERO")
	}
}

func TestTRIXIncreasingPrices(t *testing.T) {
	// Strong uptrend → TRIX should be positive (need ≥6*window data for warmup)
	s := testutils.MockTimeSeriesFl(100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130)

	trix := NewTRIXIndicatorFromSeries(s, 5)
	val := trix.Calculate(30)

	if !val.GT(decimal.ZERO) {
		t.Errorf("TRIX in uptrend should be positive, got %v", val)
	}
}

func TestTRIXDecreasingPrices(t *testing.T) {
	// Strong downtrend → TRIX should be negative (need ≥6*window data for warmup)
	s := testutils.MockTimeSeriesFl(130, 129, 128, 127, 126, 125, 124, 123, 122, 121, 120, 119, 118, 117, 116, 115, 114, 113, 112, 111, 110, 109, 108, 107, 106, 105, 104, 103, 102, 101, 100)

	trix := NewTRIXIndicatorFromSeries(s, 5)
	val := trix.Calculate(30)

	if !val.LT(decimal.ZERO) {
		t.Errorf("TRIX in downtrend should be negative, got %v", val)
	}
}

func TestTRIXFlatPrices(t *testing.T) {
	// Flat trend → TRIX should be near zero (need ≥6*window data for warmup)
	s := testutils.MockTimeSeriesFl(100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100)

	trix := NewTRIXIndicatorFromSeries(s, 5)
	val := trix.Calculate(35)

	if val.Abs().GT(decimal.New(1)) {
		t.Errorf("TRIX in flat trend should be ~0, got %v", val)
	}
}

func TestTRIXFirstIndexIsZero(t *testing.T) {
	s := testutils.MockTimeSeriesFl(100, 101, 102, 103, 104)

	trix := NewTRIXIndicatorFromSeries(s, 3)

	if !trix.Calculate(0).EQ(decimal.ZERO) {
		t.Errorf("TRIX(0) should be ZERO")
	}
}

func TestTRIXPanicInvalidWindow(t *testing.T) {
	s := series.NewTimeSeries()
	s.AddCandle(&series.Candle{ClosePrice: decimal.New(100)})

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewTRIXIndicatorFromSeries with window=0 should panic")
		}
	}()
	_ = NewTRIXIndicatorFromSeries(s, 0)
}

func TestTRIXFromIndicator(t *testing.T) {
	s := testutils.MockTimeSeriesFl(100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120)

	close := NewClosePriceIndicator(s)
	trix := NewTRIXIndicator(close, 3)

	// Should calculate without panic (need ≥6*window = 18 data points)
	_ = trix.Calculate(18)
}
