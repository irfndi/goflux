package indicators

import (
	"math"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
)

type hmaIndicator struct {
	Indicator
	series      *series.TimeSeries
	window      int
	rawHMACache []decimal.Decimal
}

func NewHMAIndicator(s *series.TimeSeries, window int) Indicator {
	return &hmaIndicator{
		series:      s,
		window:      window,
		rawHMACache: make([]decimal.Decimal, 0),
	}
}

func (h *hmaIndicator) Calculate(index int) decimal.Decimal {
	if index < 0 || index >= len(h.series.Candles) {
		return decimal.ZERO
	}

	if index < h.window-1 {
		return decimal.ZERO
	}

	h.fillRawHMACache(index)

	sqrtWindow := int(math.Sqrt(float64(h.window)))
	if sqrtWindow < 1 {
		sqrtWindow = 1
	}

	if index < h.window+sqrtWindow-2 {
		return h.rawHMACache[index]
	}

	numerator := decimal.ZERO
	denominator := decimal.New(float64(sqrtWindow * (sqrtWindow + 1) / 2))

	for i := 0; i < sqrtWindow; i++ {
		idx := index - i
		if idx < 0 {
			break
		}
		weight := decimal.New(float64(sqrtWindow - i))
		value := h.rawHMACache[idx]
		numerator = numerator.Add(value.Mul(weight))
	}

	return numerator.Div(denominator)
}

func (h *hmaIndicator) fillRawHMACache(index int) {
	halfWindow := h.window / 2
	if halfWindow < 1 {
		halfWindow = 1
	}

	wmaHalf := NewWMAIndicator(h.series, halfWindow)
	wmaFull := NewWMAIndicator(h.series, h.window)

	for i := len(h.rawHMACache); i <= index; i++ {
		valHalf := wmaHalf.Calculate(i)
		valFull := wmaFull.Calculate(i)
		rawHMA := valHalf.Mul(decimal.New(2)).Sub(valFull)
		h.rawHMACache = append(h.rawHMACache, rawHMA)
	}
}

type wmaIndicator struct {
	Indicator
	series *series.TimeSeries
	window int
}

func NewWMAIndicator(s *series.TimeSeries, window int) Indicator {
	return &wmaIndicator{
		series: s,
		window: window,
	}
}

func (w *wmaIndicator) Calculate(index int) decimal.Decimal {
	if index < 0 || index >= len(w.series.Candles) {
		return decimal.ZERO
	}

	if index < w.window-1 {
		return decimal.ZERO
	}

	numerator := decimal.ZERO
	denominator := decimal.New(float64(w.window * (w.window + 1) / 2))

	close := NewClosePriceIndicator(w.series)
	for i := 0; i < w.window; i++ {
		idx := index - i
		if idx < 0 {
			break
		}
		weight := decimal.New(float64(w.window - i))
		value := close.Calculate(idx)
		numerator = numerator.Add(value.Mul(weight))
	}

	return numerator.Div(denominator)
}
