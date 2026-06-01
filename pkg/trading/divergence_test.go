package trading

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/testutils"
)

func TestBullishDivergenceRule(t *testing.T) {
	// Price declining but RSI rising = bullish divergence
	// Mock series: Close = val, Max = val+1, Min = val-1, Volume = val
	// Need enough data for RSI(14) to produce non-zero values
	s := testutils.MockTimeSeriesFl(
		50, 52, 51, 53, 52, 54, 53, 55, 54, 56,
		55, 57, 56, 58, 57, 56, 55, 54, 53, 52,
	)
	price := indicators.NewClosePriceIndicator(s)
	rsi := indicators.NewRelativeStrengthIndexIndicator(price, 14)

	rule := NewBullishDivergenceRule(price, rsi, 5)

	// At index 19: compare price[14]=57 with price[19]=52 (lower)
	// RSI at 14 vs 19: need to check if RSI rose while price fell
	// The rule should evaluate based on start/end of lookback window
	assert.False(t, rule.IsSatisfied(4, nil), "insufficient lookback")
	assert.False(t, rule.IsSatisfied(5, nil), "no divergence at early index")
}

func TestBullishDivergenceRuleManually(t *testing.T) {
	// Use a custom price indicator and custom osc indicator
	// to force a known bullish divergence scenario
	s := testutils.MockTimeSeriesFl(100, 90, 80, 70, 60)
	price := indicators.NewClosePriceIndicator(s)

	// Create a mock oscillator that rises while price falls
	// We'll use a simple increase/decrease indicator as proxy
	// Actually, let's just use NewHighestValueIndicator and NewLowestValueIndicator
	// to create a controllable scenario. Hmm, that's complex.

	// Simpler: use a raw indicator that we can reason about
	// Let's use ROC which will have known behavior
	roc := indicators.NewROCIndicator(s, 1)

	// Price: 100 -> 90 -> 80 -> 70 -> 60 (declining)
	// ROC at index 1: (90-100)/100 = -0.1
	// ROC at index 2: (80-90)/90 = -0.111
	// ROC at index 3: (70-80)/80 = -0.125
	// ROC at index 4: (60-70)/70 = -0.143
	// ROC is also declining, so no bullish divergence

	rule := NewBullishDivergenceRule(price, roc, 2)

	// Price at 4 (60) < Price at 2 (80), but ROC at 4 (-0.143) < ROC at 2 (-0.111)
	// So no bullish divergence
	assert.False(t, rule.IsSatisfied(4, nil))
}

func TestBearishDivergenceRule(t *testing.T) {
	s := testutils.MockTimeSeriesFl(
		50, 52, 51, 53, 52, 54, 53, 55, 54, 56,
		55, 57, 56, 58, 57, 59, 60, 61, 62, 63,
	)
	price := indicators.NewClosePriceIndicator(s)
	rsi := indicators.NewRelativeStrengthIndexIndicator(price, 14)

	rule := NewBearishDivergenceRule(price, rsi, 5)

	assert.False(t, rule.IsSatisfied(4, nil), "insufficient lookback")
}

func TestBullishDivergenceNilIndicatorsPanics(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12, 14)
	price := indicators.NewClosePriceIndicator(s)
	assert.Panics(t, func() { NewBullishDivergenceRule(nil, price, 5) })
	assert.Panics(t, func() { NewBullishDivergenceRule(price, nil, 5) })
}

func TestBullishDivergenceInvalidLookbackPanics(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12, 14)
	price := indicators.NewClosePriceIndicator(s)
	osc := indicators.NewClosePriceIndicator(s)
	assert.Panics(t, func() { NewBullishDivergenceRule(price, osc, 0) })
	assert.Panics(t, func() { NewBullishDivergenceRule(price, osc, -1) })
}

func TestBearishDivergenceNilIndicatorsPanics(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12, 14)
	price := indicators.NewClosePriceIndicator(s)
	assert.Panics(t, func() { NewBearishDivergenceRule(nil, price, 5) })
	assert.Panics(t, func() { NewBearishDivergenceRule(price, nil, 5) })
}

func TestBearishDivergenceInvalidLookbackPanics(t *testing.T) {
	s := testutils.MockTimeSeriesFl(10, 12, 14)
	price := indicators.NewClosePriceIndicator(s)
	osc := indicators.NewClosePriceIndicator(s)
	assert.Panics(t, func() { NewBearishDivergenceRule(price, osc, 0) })
	assert.Panics(t, func() { NewBearishDivergenceRule(price, osc, -1) })
}

func TestRSIBullishDivergenceRule(t *testing.T) {
	s := testutils.MockTimeSeriesFl(
		50, 52, 51, 53, 52, 54, 53, 55, 54, 56,
		55, 57, 56, 58, 57, 56, 55, 54, 53, 52,
	)
	rule := NewRSIBullishDivergenceRule(s, 5)
	assert.False(t, rule.IsSatisfied(4, nil), "insufficient lookback")
}

func TestRSIBearishDivergenceRule(t *testing.T) {
	s := testutils.MockTimeSeriesFl(
		50, 52, 51, 53, 52, 54, 53, 55, 54, 56,
		55, 57, 56, 58, 57, 59, 60, 61, 62, 63,
	)
	rule := NewRSIBearishDivergenceRule(s, 5)
	assert.False(t, rule.IsSatisfied(4, nil), "insufficient lookback")
}

func TestRSIDivergenceNilSeriesPanics(t *testing.T) {
	assert.Panics(t, func() { NewRSIBullishDivergenceRule(nil, 5) })
	assert.Panics(t, func() { NewRSIBearishDivergenceRule(nil, 5) })
}

func TestMACDBullishDivergenceRule(t *testing.T) {
	s := testutils.MockTimeSeriesFl(
		50, 52, 51, 53, 52, 54, 53, 55, 54, 56,
		55, 57, 56, 58, 57, 56, 55, 54, 53, 52,
		50, 48, 46, 44, 42, 40, 38, 36, 34, 32,
	)
	rule := NewMACDBullishDivergenceRule(s, 10)
	assert.False(t, rule.IsSatisfied(4, nil), "insufficient lookback")
}

func TestMACDBearishDivergenceRule(t *testing.T) {
	s := testutils.MockTimeSeriesFl(
		50, 52, 51, 53, 52, 54, 53, 55, 54, 56,
		55, 57, 56, 58, 57, 59, 60, 61, 62, 63,
		65, 67, 69, 71, 73, 75, 77, 79, 81, 83,
	)
	rule := NewMACDBearishDivergenceRule(s, 10)
	assert.False(t, rule.IsSatisfied(4, nil), "insufficient lookback")
}

func TestMACDDivergenceNilSeriesPanics(t *testing.T) {
	assert.Panics(t, func() { NewMACDBullishDivergenceRule(nil, 5) })
	assert.Panics(t, func() { NewMACDBearishDivergenceRule(nil, 5) })
}

func TestBullishDivergenceRuleContrived(t *testing.T) {
	// Create a scenario where we KNOW divergence occurs
	// by using two simple indicators: price declines, osc rises
	// We'll use MockTimeSeriesOCHL to control volume as a proxy oscillator.
	// We'll construct it by using the LowPriceIndicator which goes down
	// and then invert it... Actually, let's just create a helper.
	// The simplest way is to use a TimeSeries where we KNOW the values.
	// Let's use MockTimeSeriesOCHL to control volume (which we can use as a proxy).

	s2 := testutils.MockTimeSeriesOCHL(
		[]float64{0, 100, 101, 99},
		[]float64{0, 90, 91, 89},
		[]float64{0, 80, 81, 79},
		[]float64{0, 70, 71, 69},
		[]float64{0, 60, 61, 59},
	)
	// Volume at indices 0..4 = 0,1,2,3,4 (increasing)
	price2 := indicators.NewClosePriceIndicator(s2)
	vol := indicators.NewVolumeIndicator(s2)

	// Price: 100 -> 90 -> 80 -> 70 -> 60 (declining)
	// Volume: 0 -> 1 -> 2 -> 3 -> 4 (rising)
	// This is a bullish divergence: price lower, volume higher

	rule := NewBullishDivergenceRule(price2, vol, 4)
	assert.False(t, rule.IsSatisfied(3, nil), "lookback=4, index=3 < 4")
	assert.True(t, rule.IsSatisfied(4, nil), "price down, volume up = bullish divergence")
}

func TestBearishDivergenceRuleContrived(t *testing.T) {
	s := testutils.MockTimeSeriesOCHL(
		[]float64{0, 60, 61, 59},
		[]float64{0, 70, 71, 69},
		[]float64{0, 80, 81, 79},
		[]float64{0, 90, 91, 89},
		[]float64{0, 100, 101, 99},
	)
	// Volume at indices 0..4 = 0,1,2,3,4 (increasing)
	price := indicators.NewClosePriceIndicator(s)
	vol := indicators.NewVolumeIndicator(s)

	// Price: 60 -> 70 -> 80 -> 90 -> 100 (rising)
	// Volume: 0 -> 1 -> 2 -> 3 -> 4 (rising)
	// This is NOT a bearish divergence (both rising)

	rule := NewBearishDivergenceRule(price, vol, 4)
	assert.False(t, rule.IsSatisfied(4, nil), "price up, volume up != bearish divergence")

	// Now create actual bearish divergence using a custom oscillator
	// that declines while price rises.
	priceInd := &mockIndicator{values: []float64{60, 70, 80, 90, 100}}
	oscInd := &mockIndicator{values: []float64{100, 80, 60, 40, 20}}

	rule2 := NewBearishDivergenceRule(priceInd, oscInd, 4)
	assert.True(t, rule2.IsSatisfied(4, nil), "price up, osc down = bearish divergence")
}

// mockIndicator is a test helper that returns hardcoded values.
type mockIndicator struct {
	values []float64
}

func (m *mockIndicator) Calculate(index int) decimal.Decimal {
	if index < 0 || index >= len(m.values) {
		return decimal.ZERO
	}
	return decimal.New(m.values[index])
}
