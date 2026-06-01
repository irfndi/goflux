package indicators

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/testutils"
)

func TestChaikinMoneyFlowInsufficientData(t *testing.T) {
	s := series.NewTimeSeries()
	for i := 0; i < 5; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(float64(10 + i)),
			MinPrice:   decimal.New(float64(5 + i)),
			ClosePrice: decimal.New(float64(7 + i)),
			Volume:     decimal.New(1000),
		})
	}

	cmf := NewChaikinMoneyFlowIndicator(s, 10)
	if !cmf.Calculate(4).EQ(decimal.ZERO) {
		t.Errorf("CMF with insufficient data should return ZERO")
	}
}

func TestChaikinMoneyFlowStrongBuyingPressure(t *testing.T) {
	// Close near high every day → positive MFM → positive CMF
	s := series.NewTimeSeries()
	for i := 0; i < 20; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(float64(110)),
			MinPrice:   decimal.New(float64(100)),
			ClosePrice: decimal.New(float64(109)), // near high
			Volume:     decimal.New(1000),
		})
	}

	cmf := NewChaikinMoneyFlowIndicator(s, 10)
	val := cmf.Calculate(19)

	if !val.GT(decimal.ZERO) {
		t.Errorf("CMF with strong buying pressure should be positive, got %v", val)
	}
	if val.GT(decimal.ONE) {
		t.Errorf("CMF should not exceed +1, got %v", val)
	}
}

func TestChaikinMoneyFlowStrongSellingPressure(t *testing.T) {
	// Close near low every day → negative MFM → negative CMF
	s := series.NewTimeSeries()
	for i := 0; i < 20; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(float64(110)),
			MinPrice:   decimal.New(float64(100)),
			ClosePrice: decimal.New(float64(101)), // near low
			Volume:     decimal.New(1000),
		})
	}

	cmf := NewChaikinMoneyFlowIndicator(s, 10)
	val := cmf.Calculate(19)

	if !val.LT(decimal.ZERO) {
		t.Errorf("CMF with strong selling pressure should be negative, got %v", val)
	}
	if val.LT(decimal.New(-1)) {
		t.Errorf("CMF should not exceed -1, got %v", val)
	}
}

func TestChaikinMoneyFlowNeutral(t *testing.T) {
	// Close at midpoint every day → MFM = 0 → CMF = 0
	s := series.NewTimeSeries()
	for i := 0; i < 20; i++ {
		s.AddCandle(&series.Candle{
			MaxPrice:   decimal.New(float64(110)),
			MinPrice:   decimal.New(float64(100)),
			ClosePrice: decimal.New(float64(105)), // midpoint
			Volume:     decimal.New(1000),
		})
	}

	cmf := NewChaikinMoneyFlowIndicator(s, 10)
	val := cmf.Calculate(19)

	if val.Abs().GT(decimal.New(0.0001)) {
		t.Errorf("CMF with neutral pressure should be ~0, got %v", val)
	}
}

func TestChaikinMoneyFlowRange(t *testing.T) {
	s := testutils.RandomTimeSeries(50)
	cmf := NewChaikinMoneyFlowIndicator(s, 14)

	for i := 13; i < 50; i++ {
		val := cmf.Calculate(i)
		if val.GT(decimal.ONE) {
			t.Errorf("CMF(%d) = %v, exceeds max +1", i, val)
		}
		if val.LT(decimal.New(-1)) {
			t.Errorf("CMF(%d) = %v, exceeds min -1", i, val)
		}
	}
}

func TestChaikinMoneyFlowPanicInvalidWindow(t *testing.T) {
	s := series.NewTimeSeries()
	s.AddCandle(&series.Candle{MaxPrice: decimal.New(105), MinPrice: decimal.New(95), ClosePrice: decimal.New(100), Volume: decimal.New(1000)})

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewChaikinMoneyFlowIndicator with window=0 should panic")
		}
	}()
	_ = NewChaikinMoneyFlowIndicator(s, 0)
}

func TestChaikinMoneyFlowOutOfBounds(t *testing.T) {
	s := series.NewTimeSeries()
	s.AddCandle(&series.Candle{MaxPrice: decimal.New(105), MinPrice: decimal.New(95), ClosePrice: decimal.New(100), Volume: decimal.New(1000)})

	cmf := NewChaikinMoneyFlowIndicator(s, 5)

	if !cmf.Calculate(-1).EQ(decimal.ZERO) {
		t.Errorf("CMF(-1) should be ZERO")
	}
	if !cmf.Calculate(10).EQ(decimal.ZERO) {
		t.Errorf("CMF(10) out of bounds should be ZERO")
	}
}
