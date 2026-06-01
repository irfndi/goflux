package indicators

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/testutils"
)

func TestAlligatorJawShift(t *testing.T) {
	s := testutils.MockTimeSeriesFl(
		10, 12, 14, 16, 18,
		20, 22, 24, 26, 28,
		30, 32, 34, 36, 38,
		40, 42, 44, 46, 48,
		50, 52, 54, 56, 58,
	)

	jaw, _, _ := NewAlligatorIndicators(s)

	// With shift=8, index < 8 should be ZERO (negative shifted index).
	for i := 0; i < 8; i++ {
		assert.True(t, jaw.Calculate(i).IsZero(), "jaw at %d should be ZERO", i)
	}

	// MMA returns ZERO for indices < window-1 (12). With shift=8, first
	// non-zero jaw value is at index 8+12 = 20.
	assert.True(t, jaw.Calculate(19).IsZero(), "jaw at 19 should be ZERO")
	assert.False(t, jaw.Calculate(20).IsZero(), "jaw at 20 should be non-zero")
}

func TestAlligatorTeethShift(t *testing.T) {
	s := testutils.MockTimeSeriesFl(
		10, 12, 14, 16, 18,
		20, 22, 24, 26, 28,
		30, 32, 34, 36, 38,
	)

	_, teeth, _ := NewAlligatorIndicators(s)

	// With shift=5, index < 5 should be ZERO.
	for i := 0; i < 5; i++ {
		assert.True(t, teeth.Calculate(i).IsZero(), "teeth at %d should be ZERO", i)
	}

	// MMA returns ZERO for indices < window-1 (7). With shift=5, first
	// non-zero teeth value is at index 5+7 = 12.
	assert.True(t, teeth.Calculate(11).IsZero(), "teeth at 11 should be ZERO")
	assert.False(t, teeth.Calculate(12).IsZero(), "teeth at 12 should be non-zero")
}

func TestAlligatorLipsShift(t *testing.T) {
	s := testutils.MockTimeSeriesFl(
		10, 12, 14, 16, 18,
		20, 22, 24, 26, 28,
	)

	_, _, lips := NewAlligatorIndicators(s)

	// With shift=3, index < 3 should be ZERO.
	for i := 0; i < 3; i++ {
		assert.True(t, lips.Calculate(i).IsZero(), "lips at %d should be ZERO", i)
	}

	// MMA returns ZERO for indices < window-1 (4). With shift=3, first
	// non-zero lips value is at index 3+4 = 7.
	assert.True(t, lips.Calculate(6).IsZero(), "lips at 6 should be ZERO")
	assert.False(t, lips.Calculate(7).IsZero(), "lips at 7 should be non-zero")
}

func TestAlligatorOrdering(t *testing.T) {
	// Rising prices: lips (fastest) > teeth > jaw (slowest)
	s := testutils.MockTimeSeriesFl(
		10, 12, 14, 16, 18, 20, 22, 24, 26, 28,
		30, 32, 34, 36, 38, 40, 42, 44, 46, 48,
		50, 52, 54, 56, 58, 60, 62, 64, 66, 68,
	)

	jaw, teeth, lips := NewAlligatorIndicators(s)

	// At a late index where all lines are established
	idx := 25
	j := jaw.Calculate(idx)
	te := teeth.Calculate(idx)
	l := lips.Calculate(idx)

	// In a strong uptrend, lips should be above teeth, teeth above jaw
	// (because shorter MAs react faster and the shift is smaller)
	assert.True(t, l.GT(te), "lips should be above teeth in uptrend")
	assert.True(t, te.GT(j), "teeth should be above jaw in uptrend")
}

func TestAlligatorCustom(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20, 22, 24, 26, 28)

	jaw, teeth, lips := NewAlligatorIndicatorsCustom(s, 5, 2, 3, 1, 2, 0)

	// Custom: jaw(5,2), teeth(3,1), lips(2,0)
	// jaw: shift=2, window=5 → first non-zero at 2+4=6
	assert.True(t, jaw.Calculate(5).IsZero(), "jaw at 5 should be ZERO")
	assert.False(t, jaw.Calculate(6).IsZero(), "jaw at 6 should be non-zero")

	// teeth: shift=1, window=3 → first non-zero at 1+2=3
	assert.True(t, teeth.Calculate(2).IsZero(), "teeth at 2 should be ZERO")
	assert.False(t, teeth.Calculate(3).IsZero(), "teeth at 3 should be non-zero")

	// lips: shift=0, window=2 → first non-zero at 0+1=1
	assert.True(t, lips.Calculate(0).IsZero(), "lips at 0 should be ZERO")
	assert.False(t, lips.Calculate(1).IsZero(), "lips at 1 should be non-zero")
}

func TestGatorOscillatorUpper(t *testing.T) {
	s := testutils.MockTimeSeriesFl(
		10, 12, 14, 16, 18,
		20, 22, 24, 26, 28,
		30, 32, 34, 36, 38,
		40, 42, 44, 46, 48,
		50, 52, 54, 56, 58,
	)

	upper, _ := NewGatorOscillatorIndicators(s)

	// Upper = |Jaw - Teeth|, always >= 0
	for i := 0; i < s.Length(); i++ {
		val := upper.Calculate(i)
		assert.True(t, val.GTE(decimal.ZERO), "gator upper at %d should be >= 0", i)
	}
}

func TestGatorOscillatorLower(t *testing.T) {
	s := testutils.MockTimeSeriesFl(
		10, 12, 14, 16, 18,
		20, 22, 24, 26, 28,
		30, 32, 34, 36, 38,
		40, 42, 44, 46, 48,
		50, 52, 54, 56, 58,
	)

	_, lower := NewGatorOscillatorIndicators(s)

	// Lower = -|Teeth - Lips|, always <= 0
	for i := 0; i < s.Length(); i++ {
		val := lower.Calculate(i)
		assert.True(t, val.LTE(decimal.ZERO), "gator lower at %d should be <= 0", i)
	}
}

func TestGatorOscillatorValues(t *testing.T) {
	// Use a series where median price is constant = 10
	s := series.NewTimeSeries()
	for i := 0; i < 30; i++ {
		c := series.NewCandle(series.NewTimePeriod(time.Unix(int64(i), 0), time.Second))
		c.OpenPrice = decimal.New(9)
		c.ClosePrice = decimal.New(11)
		c.MaxPrice = decimal.New(11)
		c.MinPrice = decimal.New(9)
		s.AddCandle(c)
	}

	upper, lower := NewGatorOscillatorIndicators(s)

	// With constant price, all Alligator lines converge to the same value (10).
	// When they converge, upper and lower should both approach 0.
	idx := 25
	upVal := upper.Calculate(idx)
	lowVal := lower.Calculate(idx)

	// Both should be very close to zero after convergence
	assert.True(t, upVal.LT(decimal.New(0.1)), "upper should be near zero: got %v", upVal)
	assert.True(t, lowVal.Abs().LT(decimal.New(0.1)), "lower should be near zero: got %v", lowVal)
}

func TestAlligatorInsufficientData(t *testing.T) {
	s := series.NewTimeSeries()
	s.AddCandle(&series.Candle{ClosePrice: decimal.New(10), MaxPrice: decimal.New(11), MinPrice: decimal.New(9)})

	jaw, teeth, lips := NewAlligatorIndicators(s)
	assert.True(t, jaw.Calculate(0).IsZero())
	assert.True(t, teeth.Calculate(0).IsZero())
	assert.True(t, lips.Calculate(0).IsZero())
}
