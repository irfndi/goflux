package indicators

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/testutils"
)

func TestPivotPointsStandard(t *testing.T) {
	s := testutils.MockTimeSeriesFl(
		10, 12, 14, 16, 18,
		20, 22, 24, 26, 28,
	)
	// MockTimeSeriesFl: MaxPrice = val+1, MinPrice = val-1
	// So at index 4 (value 18): H=19, L=17, C=18

	pp, r1, r2, r3, s1, s2, s3 := NewPivotPointIndicators(s)

	// At index 5, using previous candle (index 4): H=19, L=17, C=18
	// PP = (19+17+18)/3 = 54/3 = 18
	// R1 = 2*18 - 17 = 19
	// S1 = 2*18 - 19 = 17
	// R2 = 18 + (19-17) = 20
	// S2 = 18 - (19-17) = 16
	// R3 = 19 + 2*(18-17) = 21
	// S3 = 17 - 2*(19-18) = 15

	assert.True(t, pp.Calculate(5).EQ(decimal.New(18)), "PP should be 18")
	assert.True(t, r1.Calculate(5).EQ(decimal.New(19)), "R1 should be 19")
	assert.True(t, s1.Calculate(5).EQ(decimal.New(17)), "S1 should be 17")
	assert.True(t, r2.Calculate(5).EQ(decimal.New(20)), "R2 should be 20")
	assert.True(t, s2.Calculate(5).EQ(decimal.New(16)), "S2 should be 16")
	assert.True(t, r3.Calculate(5).EQ(decimal.New(21)), "R3 should be 21")
	assert.True(t, s3.Calculate(5).EQ(decimal.New(15)), "S3 should be 15")
}

func TestPivotPointsIndexZero(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12)
	pp, r1, r2, r3, s1, s2, s3 := NewPivotPointIndicators(s)

	// At index 0 there is no previous candle
	assert.True(t, pp.Calculate(0).IsZero())
	assert.True(t, r1.Calculate(0).IsZero())
	assert.True(t, s1.Calculate(0).IsZero())
	assert.True(t, r2.Calculate(0).IsZero())
	assert.True(t, s2.Calculate(0).IsZero())
	assert.True(t, r3.Calculate(0).IsZero())
	assert.True(t, s3.Calculate(0).IsZero())
}

func TestPivotPointsBackwardCompat(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18)
	p := NewPivotPointsIndicator(s)

	// Calculate returns PP using previous candle (index 2: H=15, L=13, C=14)
	// PP = (15+13+14)/3 = 14
	assert.True(t, p.Calculate(3).EQ(decimal.New(14)), "backward compat PP")

	// GetLevels returns all levels
	levels := p.GetLevels(3)
	assert.True(t, levels.Pivot.EQ(decimal.New(14)))
	assert.True(t, levels.R1.GT(levels.Pivot))
	assert.True(t, levels.S1.LT(levels.Pivot))
}

func TestCamarillaPivotPoints(t *testing.T) {
	s := testutils.MockTimeSeriesFl(
		10, 12, 14, 16, 18, 20,
	)
	// At index 4: H=19, L=17, C=18

	pp, r1, r2, r3, r4, s1, s2, s3, s4 := NewCamarillaPivotPointIndicators(s)

	idx := 5
	ppVal := pp.Calculate(idx)
	r1Val := r1.Calculate(idx)
	s1Val := s1.Calculate(idx)

	// PP = (19+17+18)/3 = 18
	assert.True(t, ppVal.EQ(decimal.New(18)), "Camarilla PP should be 18")

	// R1 = 18 + (19-17)*1.1/12 = 18 + 2*1.1/12 = 18.1833...
	// S1 = 18 - (19-17)*1.1/12 = 17.8166...
	assert.True(t, r1Val.GT(ppVal), "Camarilla R1 should be above PP")
	assert.True(t, s1Val.LT(ppVal), "Camarilla S1 should be below PP")

	// Verify ordering
	assert.True(t, r4.Calculate(idx).GT(r3.Calculate(idx)), "R4 > R3")
	assert.True(t, r3.Calculate(idx).GT(r2.Calculate(idx)), "R3 > R2")
	assert.True(t, r2.Calculate(idx).GT(r1.Calculate(idx)), "R2 > R1")
	assert.True(t, s1.Calculate(idx).GT(s2.Calculate(idx)), "S1 > S2")
	assert.True(t, s2.Calculate(idx).GT(s3.Calculate(idx)), "S2 > S3")
	assert.True(t, s3.Calculate(idx).GT(s4.Calculate(idx)), "S3 > S4")
}

func TestWoodiePivotPoints(t *testing.T) {
	s := testutils.MockTimeSeriesFl(
		10, 12, 14, 16, 18, 20,
	)
	// At index 4: H=19, L=17, C=18

	pp, r1, r2, r3, s1, s2, s3 := NewWoodiePivotPointIndicators(s)

	idx := 5
	// PP = (19+17+2*18)/4 = (19+17+36)/4 = 72/4 = 18
	assert.True(t, pp.Calculate(idx).EQ(decimal.New(18)), "Woodie PP should be 18")

	// Verify ordering: R3 > R2 > R1 > PP > S1 > S2 > S3
	assert.True(t, r3.Calculate(idx).GT(r2.Calculate(idx)), "R3 > R2")
	assert.True(t, r2.Calculate(idx).GT(r1.Calculate(idx)), "R2 > R1")
	assert.True(t, r1.Calculate(idx).GT(pp.Calculate(idx)), "R1 > PP")
	assert.True(t, pp.Calculate(idx).GT(s1.Calculate(idx)), "PP > S1")
	assert.True(t, s1.Calculate(idx).GT(s2.Calculate(idx)), "S1 > S2")
	assert.True(t, s2.Calculate(idx).GT(s3.Calculate(idx)), "S2 > S3")
}

func TestFibonacciPivotPoints(t *testing.T) {
	s := testutils.MockTimeSeriesFl(
		10, 12, 14, 16, 18, 20,
	)
	// At index 4: H=19, L=17, C=18

	pp, r1, r2, r3, s1, s2, s3 := NewFibonacciPivotPointIndicators(s)

	idx := 5
	// PP = (19+17+18)/3 = 18
	assert.True(t, pp.Calculate(idx).EQ(decimal.New(18)), "Fib PP should be 18")

	// Verify ordering: R3 > R2 > R1 > PP > S1 > S2 > S3
	assert.True(t, r3.Calculate(idx).GT(r2.Calculate(idx)), "R3 > R2")
	assert.True(t, r2.Calculate(idx).GT(r1.Calculate(idx)), "R2 > R1")
	assert.True(t, r1.Calculate(idx).GT(pp.Calculate(idx)), "R1 > PP")
	assert.True(t, pp.Calculate(idx).GT(s1.Calculate(idx)), "PP > S1")
	assert.True(t, s1.Calculate(idx).GT(s2.Calculate(idx)), "S1 > S2")
	assert.True(t, s2.Calculate(idx).GT(s3.Calculate(idx)), "S2 > S3")
}

func TestPivotPointsOOBIndex(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12)
	pp, _, _, _, _, _, _ := NewPivotPointIndicators(s)

	// Index beyond series length should return ZERO
	assert.True(t, pp.Calculate(100).IsZero())
}

func TestPivotPointsNilSeriesPanics(t *testing.T) {
	assert.Panics(t, func() { NewPivotPointsIndicator(nil) })
	assert.Panics(t, func() { NewPivotPointIndicators(nil) })
	assert.Panics(t, func() { NewCamarillaPivotPointIndicators(nil) })
	assert.Panics(t, func() { NewWoodiePivotPointIndicators(nil) })
	assert.Panics(t, func() { NewFibonacciPivotPointIndicators(nil) })
}

func TestPivotPointsCacheShared(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20)

	pp, r1, _, _, s1, _, _ := NewPivotPointIndicators(s)

	idx := 5
	// First call populates cache
	_ = pp.Calculate(idx)
	// Subsequent calls on sibling levels should hit cache
	_ = r1.Calculate(idx)
	_ = s1.Calculate(idx)
	// No panic, correct values
	assert.True(t, pp.Calculate(idx).EQ(decimal.New(18)))
	assert.True(t, r1.Calculate(idx).EQ(decimal.New(19)))
	assert.True(t, s1.Calculate(idx).EQ(decimal.New(17)))
}
