package indicators_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/testutils"
)

// --- T3 Tests ---

func TestT3Convergence(t *testing.T) {
	tsValues := make([]float64, 250)
	for i := range tsValues {
		tsValues[i] = 100
	}
	ts := testutils.MockTimeSeriesFl(tsValues...)
	closeInd := indicators.NewClosePriceIndicator(ts)
	t3 := indicators.NewT3Indicator(closeInd, 5, 0.7)

	val := t3.Calculate(249)
	assert.NotNil(t, val)
	assert.True(t, val.Sub(decimal.New(100)).Abs().LT(decimal.New(0.0001)))
}

func TestT3RisingPrices(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38)
	closeInd := indicators.NewClosePriceIndicator(ts)
	t3 := indicators.NewT3Indicator(closeInd, 3, 0.7)

	// At index 4, T3 should be non-zero and below the current price (smoothed lag)
	assert.True(t, t3.Calculate(4).GT(decimal.ZERO))
	// T3 should smooth the trend, so it should be less than the latest price
	assert.True(t, t3.Calculate(14).LT(decimal.New(38)), "T3 should smooth below latest price")
}

func TestT3InsufficientData(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20)
	closeInd := indicators.NewClosePriceIndicator(ts)
	t3 := indicators.NewT3Indicator(closeInd, 5, 0.7)

	// EMA returns ZERO for index < window-1 (4)
	assert.True(t, t3.Calculate(0).IsZero())
	assert.True(t, t3.Calculate(3).IsZero())
	// At index 4, first EMA value is available
	assert.False(t, t3.Calculate(4).IsZero())
}

func TestT3Default(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20, 22)
	closeInd := indicators.NewClosePriceIndicator(ts)
	t3 := indicators.NewDefaultT3Indicator(closeInd)

	// Default window=6, so first non-zero at index 5
	assert.True(t, t3.Calculate(4).IsZero())
	assert.False(t, t3.Calculate(5).IsZero())
}

func TestT3FromSeries(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20, 22)
	t3 := indicators.NewT3IndicatorFromSeries(ts, 3, 0.7)

	assert.False(t, t3.Calculate(4).IsZero())
}

func TestT3NilIndicatorPanics(t *testing.T) {
	assert.Panics(t, func() { indicators.NewT3Indicator(nil, 5, 0.7) })
}

func TestT3InvalidWindowPanics(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14)
	closeInd := indicators.NewClosePriceIndicator(ts)
	assert.Panics(t, func() { indicators.NewT3Indicator(closeInd, 0, 0.7) })
	assert.Panics(t, func() { indicators.NewT3Indicator(closeInd, -1, 0.7) })
}

func TestT3NilSeriesPanics(t *testing.T) {
	assert.Panics(t, func() { indicators.NewT3IndicatorFromSeries(nil, 5, 0.7) })
}

// --- ALMA Tests ---

func TestALMAConvergence(t *testing.T) {
	tsValues := make([]float64, 50)
	for i := range tsValues {
		tsValues[i] = 100
	}
	ts := testutils.MockTimeSeriesFl(tsValues...)
	closeInd := indicators.NewClosePriceIndicator(ts)
	alma := indicators.NewALMAIndicator(closeInd, 5, 0.85, 6.0)

	val := alma.Calculate(49)
	assert.NotNil(t, val)
	assert.True(t, val.Sub(decimal.New(100)).Abs().LT(decimal.New(0.0001)))
}

func TestALMARisingPrices(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38)
	closeInd := indicators.NewClosePriceIndicator(ts)
	alma := indicators.NewALMAIndicator(closeInd, 5, 0.85, 6.0)

	// ALMA should follow the trend
	assert.True(t, alma.Calculate(14).GT(alma.Calculate(5)), "ALMA should rise with prices")
}

func TestALMAInsufficientData(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20)
	closeInd := indicators.NewClosePriceIndicator(ts)
	alma := indicators.NewALMAIndicator(closeInd, 5, 0.85, 6.0)

	// Should return ZERO for index < window-1 (4)
	assert.True(t, alma.Calculate(0).IsZero())
	assert.True(t, alma.Calculate(3).IsZero())
	assert.False(t, alma.Calculate(4).IsZero())
}

func TestALMADefault(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20, 22, 24, 26, 28)
	closeInd := indicators.NewClosePriceIndicator(ts)
	alma := indicators.NewDefaultALMAIndicator(closeInd)

	// Default window=9, so first non-zero at index 8
	assert.True(t, alma.Calculate(7).IsZero())
	assert.False(t, alma.Calculate(8).IsZero())
}

func TestALMAFromSeries(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20, 22)
	alma := indicators.NewALMAIndicatorFromSeries(ts, 3, 0.85, 6.0)

	assert.False(t, alma.Calculate(4).IsZero())
}

func TestALMANilIndicatorPanics(t *testing.T) {
	assert.Panics(t, func() { indicators.NewALMAIndicator(nil, 5, 0.85, 6.0) })
}

func TestALMAInvalidWindowPanics(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14)
	closeInd := indicators.NewClosePriceIndicator(ts)
	assert.Panics(t, func() { indicators.NewALMAIndicator(closeInd, 0, 0.85, 6.0) })
	assert.Panics(t, func() { indicators.NewALMAIndicator(closeInd, -1, 0.85, 6.0) })
}

func TestALMANilSeriesPanics(t *testing.T) {
	assert.Panics(t, func() { indicators.NewALMAIndicatorFromSeries(nil, 5, 0.85, 6.0) })
}

// --- VIDYA Tests ---

func TestVIDYAConvergence(t *testing.T) {
	tsValues := make([]float64, 50)
	for i := range tsValues {
		tsValues[i] = 100
	}
	ts := testutils.MockTimeSeriesFl(tsValues...)
	closeInd := indicators.NewClosePriceIndicator(ts)
	vidya := indicators.NewVIDYAIndicator(closeInd, 5)

	val := vidya.Calculate(49)
	assert.NotNil(t, val)
	assert.True(t, val.EQ(decimal.New(100)))
}

func TestVIDYARisingPrices(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38)
	closeInd := indicators.NewClosePriceIndicator(ts)
	vidya := indicators.NewVIDYAIndicator(closeInd, 5)

	// VIDYA should follow the trend
	assert.True(t, vidya.Calculate(14).GT(vidya.Calculate(5)), "VIDYA should rise with prices")
}

func TestVIDYAInsufficientData(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20)
	closeInd := indicators.NewClosePriceIndicator(ts)
	vidya := indicators.NewVIDYAIndicator(closeInd, 5)

	// For i < window, VIDYA returns raw price (initialization)
	assert.True(t, vidya.Calculate(0).EQ(decimal.New(10)))
	assert.True(t, vidya.Calculate(2).EQ(decimal.New(14)))
}

func TestVIDYADefault(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38)
	closeInd := indicators.NewClosePriceIndicator(ts)
	vidya := indicators.NewDefaultVIDYAIndicator(closeInd)

	// Default window=14, so adaptive kicks in at index 14
	assert.True(t, vidya.Calculate(0).EQ(decimal.New(10)))
	assert.False(t, vidya.Calculate(14).IsZero())
}

func TestVIDYAFromSeries(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20, 22)
	vidya := indicators.NewVIDYAIndicatorFromSeries(ts, 3)

	assert.True(t, vidya.Calculate(0).EQ(decimal.New(10)))
	assert.False(t, vidya.Calculate(4).IsZero())
}

func TestVIDYANilIndicatorPanics(t *testing.T) {
	assert.Panics(t, func() { indicators.NewVIDYAIndicator(nil, 5) })
}

func TestVIDYAInvalidWindowPanics(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14)
	closeInd := indicators.NewClosePriceIndicator(ts)
	assert.Panics(t, func() { indicators.NewVIDYAIndicator(closeInd, 0) })
	assert.Panics(t, func() { indicators.NewVIDYAIndicator(closeInd, -1) })
}

func TestVIDYANilSeriesPanics(t *testing.T) {
	assert.Panics(t, func() { indicators.NewVIDYAIndicatorFromSeries(nil, 5) })
}

func TestVIDYAConcurrentAccess(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38)
	closeInd := indicators.NewClosePriceIndicator(ts)
	vidya := indicators.NewVIDYAIndicator(closeInd, 5)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = vidya.Calculate(idx)
		}(i)
	}
	wg.Wait()

	// Should not panic and cache should be consistent
	assert.False(t, vidya.Calculate(14).IsZero())
}

func TestVIDYANegativeIndex(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(10, 12, 14)
	closeInd := indicators.NewClosePriceIndicator(ts)
	vidya := indicators.NewVIDYAIndicator(closeInd, 5)

	assert.True(t, vidya.Calculate(-1).IsZero())
}

// --- MAMA / FAMA Tests (unchanged) ---

func TestMAMA(t *testing.T) {
	tsValues := make([]float64, 50)
	for i := range tsValues {
		tsValues[i] = 100
	}
	ts := testutils.MockTimeSeriesFl(tsValues...)
	closeInd := indicators.NewClosePriceIndicator(ts)
	mama := indicators.NewMAMAIndicator(closeInd, 0.5, 0.05)

	val := mama.Calculate(49)
	assert.NotNil(t, val)
	assert.True(t, val.EQ(decimal.New(100)))
}

func TestFAMA(t *testing.T) {
	tsValues := make([]float64, 50)
	for i := range tsValues {
		tsValues[i] = 100
	}
	ts := testutils.MockTimeSeriesFl(tsValues...)
	closeInd := indicators.NewClosePriceIndicator(ts)
	fama := indicators.NewFAMAIndicator(closeInd, 0.5, 0.05)

	famaVal := fama.Calculate(49)
	assert.True(t, famaVal.EQ(decimal.New(100)))
}

func TestFAMAFollowsMAMA(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12)
	closeInd := indicators.NewClosePriceIndicator(ts)
	mama := indicators.NewMAMAIndicator(closeInd, 0.5, 0.05)
	fama := indicators.NewFAMAIndicator(closeInd, 0.5, 0.05)

	mamaVal := mama.Calculate(11)
	famaVal := fama.Calculate(11)
	assert.True(t, famaVal.LT(mamaVal))
}
