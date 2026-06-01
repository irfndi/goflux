package indicators

import (
	"math"
	"strconv"
	"sync"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/telemetry"
)

type vwmaIndicator struct {
	indicator Indicator
	volume    Indicator
	window    int
}

// NewVWMAIndicator returns a Volume Weighted Moving Average
func NewVWMAIndicator(indicator, volume Indicator, window int) Indicator {
	return vwmaIndicator{indicator, volume, window}
}

// NewVWMAIndicatorFromSeries is a helper to create VWMA from series
func NewVWMAIndicatorFromSeries(s *series.TimeSeries, window int) Indicator {
	return NewVWMAIndicator(
		NewClosePriceIndicator(s),
		NewVolumeIndicator(s),
		window,
	)
}

func (vwma vwmaIndicator) Calculate(index int) decimal.Decimal {
	if index < vwma.window-1 {
		return decimal.ZERO
	}

	priceVolumeSum := decimal.ZERO
	volumeSum := decimal.ZERO

	for i := index - vwma.window + 1; i <= index; i++ {
		price := vwma.indicator.Calculate(i)
		vol := vwma.volume.Calculate(i)
		priceVolumeSum = priceVolumeSum.Add(price.Mul(vol))
		volumeSum = volumeSum.Add(vol)
	}

	if volumeSum.IsZero() {
		return decimal.ZERO
	}

	return priceVolumeSum.Div(volumeSum)
}

// TimeSeriesIndicator is a helper to access TimeSeries from within indicators if needed
type TimeSeriesIndicator struct {
	series *series.TimeSeries
}

func NewTimeSeriesIndicator(s *series.TimeSeries) *TimeSeriesIndicator {
	return &TimeSeriesIndicator{s}
}

func (tsi *TimeSeriesIndicator) Calculate(index int) decimal.Decimal {
	return tsi.series.GetCandle(index).ClosePrice
}

// RMAIndicator is Running Moving Average (used in RSI, also known as SMMA)
type rmaIndicator struct {
	indicator   Indicator
	window      int
	alpha       decimal.Decimal
	resultCache resultCache
	cacheMu     sync.RWMutex
}

func NewRMAIndicator(indicator Indicator, window int) Indicator {
	return &rmaIndicator{
		indicator:   indicator,
		window:      window,
		alpha:       decimal.ONE.Div(decimal.New(float64(window))),
		resultCache: make([]*decimal.Decimal, 0, defaultCacheSize),
	}
}

func (rma *rmaIndicator) Calculate(index int) decimal.Decimal {
	if cachedValue := returnIfCached(rma, index, func(i int) decimal.Decimal {
		return NewSimpleMovingAverage(rma.indicator, rma.window).Calculate(i)
	}); cachedValue != nil {
		return *cachedValue
	}

	todayVal := rma.indicator.Calculate(index)
	lastVal := rma.Calculate(index - 1)

	// RMA = alpha * today + (1 - alpha) * last
	result := todayVal.Mul(rma.alpha).Add(lastVal.Mul(decimal.ONE.Sub(rma.alpha)))
	cacheResult(rma, index, result)

	return result
}

func (rma *rmaIndicator) cache() resultCache        { return rma.resultCache }
func (rma *rmaIndicator) setCache(c resultCache)    { rma.resultCache = c }
func (rma *rmaIndicator) windowSize() int           { return rma.window }
func (rma *rmaIndicator) cacheMutex() *sync.RWMutex { return &rma.cacheMu }
func (rma *rmaIndicator) maxCacheSize() int         { return defaultMaxCacheSize }

type trimaIndicator struct {
	indicator Indicator
	window    int
}

func NewTRIMAIndicator(indicator Indicator, window int) Indicator {
	return &trimaIndicator{
		indicator: indicator,
		window:    window,
	}
}

func (t *trimaIndicator) Calculate(index int) decimal.Decimal {
	if index < t.window-1 {
		return decimal.ZERO
	}

	half := (t.window + 1) / 2
	var denomInt int
	if t.window%2 == 0 {
		h := t.window / 2
		denomInt = h * (h + 1)
	} else {
		denomInt = half * half
	}
	denom := decimal.New(float64(denomInt))

	numerator := decimal.ZERO
	start := index - t.window + 1
	for i := 0; i < t.window; i++ {
		pos := i + 1
		weightInt := pos
		if pos > half {
			weightInt = t.window - pos + 1
		}
		weight := decimal.New(float64(weightInt))
		numerator = numerator.Add(t.indicator.Calculate(start + i).Mul(weight))
	}

	return numerator.Div(denom)
}

// wmaIndicator is the Weighted Moving Average
type wmaIndicator struct {
	indicator Indicator
	window    int
}

// NewWMAIndicator returns a new Weighted Moving Average
func NewWMAIndicator(indicator Indicator, window int) Indicator {
	return &wmaIndicator{indicator, window}
}

func (wma wmaIndicator) Calculate(index int) decimal.Decimal {
	if index < wma.window-1 {
		return decimal.ZERO
	}

	numerator := decimal.ZERO
	denominator := decimal.New(float64(wma.window * (wma.window + 1) / 2))

	for i := 0; i < wma.window; i++ {
		weight := decimal.New(float64(wma.window - i))
		numerator = numerator.Add(wma.indicator.Calculate(index - i).Mul(weight))
	}

	return numerator.Div(denominator)
}

// --- T3 (Tillson T3 Moving Average) ---

type t3Indicator struct {
	window int
	c1     decimal.Decimal
	c2     decimal.Decimal
	c3     decimal.Decimal
	c4     decimal.Decimal
	e3     Indicator
	e4     Indicator
	e5     Indicator
	e6     Indicator
}

// NewT3Indicator returns a new Tillson T3 Moving Average.
// Panics if indicator is nil or window <= 0.
func NewT3Indicator(indicator Indicator, window int, vFactor float64) Indicator {
	if indicator == nil {
		panic("goflux: T3 indicator cannot be nil")
	}
	if window <= 0 {
		panic("goflux: T3 window must be > 0")
	}
	telemetry.ReportUsage("T3", map[string]string{
		"window":        strconv.Itoa(window),
		"volume_factor": strconv.FormatFloat(vFactor, 'f', -1, 64),
	})

	e1 := NewEMAIndicator(indicator, window)
	e2 := NewEMAIndicator(e1, window)
	e3 := NewEMAIndicator(e2, window)
	e4 := NewEMAIndicator(e3, window)
	e5 := NewEMAIndicator(e4, window)
	e6 := NewEMAIndicator(e5, window)

	v := decimal.New(vFactor)
	v2 := v.Mul(v)
	v3 := v2.Mul(v)

	c1 := v3.Neg()
	c2 := v2.Mul(decimal.New(3)).Add(v3.Mul(decimal.New(3)))
	c3 := v2.Mul(decimal.New(-6)).Sub(v.Mul(decimal.New(3))).Sub(v3.Mul(decimal.New(3)))
	c4 := decimal.ONE.Add(v.Mul(decimal.New(3))).Add(v2.Mul(decimal.New(3))).Add(v3)

	return &t3Indicator{
		window: window,
		c1:     c1,
		c2:     c2,
		c3:     c3,
		c4:     c4,
		e3:     e3,
		e4:     e4,
		e5:     e5,
		e6:     e6,
	}
}

// NewDefaultT3Indicator returns a T3 with default period=6 and volumeFactor=0.7.
// Panics if indicator is nil.
func NewDefaultT3Indicator(indicator Indicator) Indicator {
	return NewT3Indicator(indicator, 6, 0.7)
}

// NewT3IndicatorFromSeries returns a T3 from a TimeSeries using close prices.
// Panics if s is nil.
func NewT3IndicatorFromSeries(s *series.TimeSeries, window int, vFactor float64) Indicator {
	if s == nil {
		panic("goflux: T3 series cannot be nil")
	}
	return NewT3Indicator(NewClosePriceIndicator(s), window, vFactor)
}

func (t3 *t3Indicator) Calculate(index int) decimal.Decimal {
	return t3.c1.Mul(t3.e6.Calculate(index)).
		Add(t3.c2.Mul(t3.e5.Calculate(index))).
		Add(t3.c3.Mul(t3.e4.Calculate(index))).
		Add(t3.c4.Mul(t3.e3.Calculate(index)))
}

// --- ALMA (Arnaud Legoux Moving Average) ---

type almaIndicator struct {
	indicator Indicator
	window    int
	weights   []decimal.Decimal
	norm      decimal.Decimal
}

// NewALMAIndicator returns a new Arnaud Legoux Moving Average.
// Panics if indicator is nil or window <= 0.
func NewALMAIndicator(indicator Indicator, window int, offset, sigma float64) Indicator {
	if indicator == nil {
		panic("goflux: ALMA indicator cannot be nil")
	}
	if window <= 0 {
		panic("goflux: ALMA window must be > 0")
	}
	telemetry.ReportUsage("ALMA", map[string]string{
		"window": strconv.Itoa(window),
		"offset": strconv.FormatFloat(offset, 'f', -1, 64),
		"sigma":  strconv.FormatFloat(sigma, 'f', -1, 64),
	})

	m := offset * float64(window-1)
	s := float64(window) / sigma

	weights := make([]decimal.Decimal, window)
	norm := 0.0
	for i := 0; i < window; i++ {
		weight := math.Exp(-((float64(i) - m) * (float64(i) - m)) / (2 * s * s))
		weights[i] = decimal.New(weight)
		norm += weight
	}

	return &almaIndicator{
		indicator: indicator,
		window:    window,
		weights:   weights,
		norm:      decimal.New(norm),
	}
}

// NewDefaultALMAIndicator returns an ALMA with default period=9, offset=0.85, sigma=6.0.
// Panics if indicator is nil.
func NewDefaultALMAIndicator(indicator Indicator) Indicator {
	return NewALMAIndicator(indicator, 9, 0.85, 6.0)
}

// NewALMAIndicatorFromSeries returns an ALMA from a TimeSeries using close prices.
// Panics if s is nil.
func NewALMAIndicatorFromSeries(s *series.TimeSeries, window int, offset, sigma float64) Indicator {
	if s == nil {
		panic("goflux: ALMA series cannot be nil")
	}
	return NewALMAIndicator(NewClosePriceIndicator(s), window, offset, sigma)
}

func (alma *almaIndicator) Calculate(index int) decimal.Decimal {
	if index < alma.window-1 {
		return decimal.ZERO
	}

	sum := decimal.ZERO
	start := index - alma.window + 1
	for i := 0; i < alma.window; i++ {
		sum = sum.Add(alma.indicator.Calculate(start + i).Mul(alma.weights[i]))
	}

	return sum.Div(alma.norm)
}

// --- VIDYA (Variable Index Dynamic Average) ---

type vidyaIndicator struct {
	indicator Indicator
	window    int
	alpha     decimal.Decimal
	cmo       Indicator
	cache     []decimal.Decimal
	cacheMu   sync.RWMutex
}

// NewVIDYAIndicator returns a new Variable Index Dynamic Average.
// Panics if indicator is nil or window <= 0.
func NewVIDYAIndicator(indicator Indicator, window int) Indicator {
	if indicator == nil {
		panic("goflux: VIDYA indicator cannot be nil")
	}
	if window <= 0 {
		panic("goflux: VIDYA window must be > 0")
	}
	telemetry.ReportUsage("VIDYA", map[string]string{"window": strconv.Itoa(window)})

	return &vidyaIndicator{
		indicator: indicator,
		window:    window,
		alpha:     decimal.New(2).Div(decimal.New(float64(window + 1))),
		cmo:       NewChandeMomentumOscillatorIndicator(indicator, window),
		cache:     make([]decimal.Decimal, 0),
	}
}

// NewDefaultVIDYAIndicator returns a VIDYA with default period=14.
// Panics if indicator is nil.
func NewDefaultVIDYAIndicator(indicator Indicator) Indicator {
	return NewVIDYAIndicator(indicator, 14)
}

// NewVIDYAIndicatorFromSeries returns a VIDYA from a TimeSeries using close prices.
// Panics if s is nil.
func NewVIDYAIndicatorFromSeries(s *series.TimeSeries, window int) Indicator {
	if s == nil {
		panic("goflux: VIDYA series cannot be nil")
	}
	return NewVIDYAIndicator(NewClosePriceIndicator(s), window)
}

func (vidya *vidyaIndicator) Calculate(index int) decimal.Decimal {
	if index < 0 {
		return decimal.ZERO
	}

	vidya.cacheMu.RLock()
	if index < len(vidya.cache) {
		val := vidya.cache[index]
		vidya.cacheMu.RUnlock()
		return val
	}
	vidya.cacheMu.RUnlock()

	vidya.cacheMu.Lock()
	defer vidya.cacheMu.Unlock()

	// Double-check after acquiring write lock
	if index < len(vidya.cache) {
		return vidya.cache[index]
	}

	start := len(vidya.cache)
	if start == 0 {
		vidya.cache = append(vidya.cache, vidya.indicator.Calculate(0))
		start = 1
	}

	for i := start; i <= index; i++ {
		if i < vidya.window {
			vidya.cache = append(vidya.cache, vidya.indicator.Calculate(i))
			continue
		}

		k := vidya.cmo.Calculate(i).Abs().Div(decimal.New(100))
		ak := vidya.alpha.Mul(k)
		val := vidya.indicator.Calculate(i)
		prev := vidya.cache[i-1]
		vidya.cache = append(vidya.cache, ak.Mul(val).Add(decimal.ONE.Sub(ak).Mul(prev)))
	}

	return vidya.cache[index]
}
