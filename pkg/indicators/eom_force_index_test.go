package indicators_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/testutils"
)

// --- EOM Tests ---

func TestRawEOM(t *testing.T) {
	// Use OCHL to control H, L, C, and Volume precisely
	// Each element: [Open, Close, High, Low]
	s := testutils.MockTimeSeriesOCHL(
		[]float64{10, 10, 12, 8},  // idx 0: H=12, L=8, C=10, V=0
		[]float64{11, 11, 13, 9},  // idx 1: H=13, L=9, C=11, V=1
		[]float64{12, 12, 14, 10}, // idx 2: H=14, L=10, C=12, V=2
		[]float64{11, 11, 13, 9},  // idx 3: H=13, L=9, C=11, V=3
		[]float64{10, 10, 12, 8},  // idx 4: H=12, L=8, C=10, V=4
	)

	rawEOM := indicators.NewRawEaseOfMovementIndicator(s)

	// At index 0, no previous candle
	assert.True(t, rawEOM.Calculate(0).IsZero())

	// At index 1:
	// Distance Moved = ((13+9)/2) - ((12+8)/2) = 11 - 10 = 1
	// Box Ratio = (1 / 100,000,000) / (13-9) = 1e-8 / 4 = 2.5e-9
	// EOM = 1 / 2.5e-9 = 400,000,000
	val1 := rawEOM.Calculate(1)
	assert.True(t, val1.GT(decimal.ZERO))

	// At index 4, price contracted back
	// Distance Moved = ((12+8)/2) - ((13+9)/2) = 10 - 11 = -1
	// Box Ratio = (4 / 100,000,000) / (12-8) = 4e-8 / 4 = 1e-8
	// EOM = -1 / 1e-8 = -100,000,000
	val4 := rawEOM.Calculate(4)
	assert.True(t, val4.LT(decimal.ZERO))
}

func TestRawEOMInsufficientData(t *testing.T) {
	s := testutils.MockTimeSeriesOCHL(
		[]float64{10, 10, 12, 8},
	)
	rawEOM := indicators.NewRawEaseOfMovementIndicator(s)
	assert.True(t, rawEOM.Calculate(0).IsZero())
}

func TestEOMSmoothed(t *testing.T) {
	s := testutils.MockTimeSeriesOCHL(
		[]float64{10, 10, 12, 8},
		[]float64{11, 11, 13, 9},
		[]float64{12, 12, 14, 10},
		[]float64{11, 11, 13, 9},
		[]float64{10, 10, 12, 8},
	)

	eom := indicators.NewEaseOfMovementIndicator(s, 3)

	// SMA(3) at index 2 averages raw EOM at 0,1,2
	// Raw EOM at 0 is ZERO, at 1 and 2 are non-zero
	assert.True(t, eom.Calculate(1).IsZero())
	assert.False(t, eom.Calculate(2).IsZero())
}

func TestEOMDefault(t *testing.T) {
	s := testutils.MockTimeSeriesOCHL(
		[]float64{10, 10, 12, 8},
		[]float64{11, 11, 13, 9},
	)
	eom := indicators.NewDefaultEaseOfMovementIndicator(s)
	assert.True(t, eom.Calculate(1).IsZero())
}

func TestEOMNilSeriesPanics(t *testing.T) {
	assert.Panics(t, func() { indicators.NewEaseOfMovementIndicator(nil, 14) })
}

func TestEOMInvalidWindowPanics(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12, 14)
	assert.Panics(t, func() { indicators.NewEaseOfMovementIndicator(s, 0) })
	assert.Panics(t, func() { indicators.NewEaseOfMovementIndicator(s, -1) })
}

// --- Force Index Tests ---

func TestRawForceIndex(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18)
	// MockTimeSeriesFl: Volume = Close value
	// idx 0: C=10, V=10
	// idx 1: C=12, V=12
	// idx 2: C=14, V=14
	// idx 3: C=16, V=16
	// idx 4: C=18, V=18

	fi := indicators.NewRawForceIndexIndicator(s)

	// At index 0, no previous close
	assert.True(t, fi.Calculate(0).IsZero())

	// At index 1: (12 - 10) * 12 = 2 * 12 = 24
	val1 := fi.Calculate(1)
	assert.True(t, val1.EQ(decimal.New(24)))

	// At index 2: (14 - 12) * 14 = 2 * 14 = 28
	val2 := fi.Calculate(2)
	assert.True(t, val2.EQ(decimal.New(28)))
}

func TestRawForceIndexDecliningPrices(t *testing.T) {
	s := testutils.MockTimeSeriesFl(20, 18, 16, 14, 12)
	fi := indicators.NewRawForceIndexIndicator(s)

	// At index 1: (18 - 20) * 18 = -2 * 18 = -36
	val1 := fi.Calculate(1)
	assert.True(t, val1.EQ(decimal.New(-36)))
}

func TestRawForceIndexInsufficientData(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10)
	fi := indicators.NewRawForceIndexIndicator(s)
	assert.True(t, fi.Calculate(0).IsZero())
}

func TestForceIndexSmoothed(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20, 22, 24, 26, 28)
	fi := indicators.NewForceIndexIndicator(s, 3)

	// Raw FI available from index 1; EMA(3) returns ZERO for index < 2
	assert.True(t, fi.Calculate(0).IsZero())
	assert.True(t, fi.Calculate(1).IsZero())
	// At index 2, EMA starts being non-zero
	assert.False(t, fi.Calculate(2).IsZero())
	assert.False(t, fi.Calculate(3).IsZero())
}

func TestForceIndexDefault(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18)
	fi := indicators.NewDefaultForceIndexIndicator(s)

	// Default window=13, but with only 5 candles most will be zero
	assert.True(t, fi.Calculate(0).IsZero())
}

func TestForceIndexNilSeriesPanics(t *testing.T) {
	assert.Panics(t, func() { indicators.NewForceIndexIndicator(nil, 13) })
}

func TestForceIndexInvalidWindowPanics(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12, 14)
	assert.Panics(t, func() { indicators.NewForceIndexIndicator(s, 0) })
	assert.Panics(t, func() { indicators.NewForceIndexIndicator(s, -1) })
}
