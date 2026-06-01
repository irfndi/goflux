package indicators

import (
	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/telemetry"
)

var (
	decTwo         = decimal.New(2)
	decThree       = decimal.New(3)
	decFour        = decimal.New(4)
	decSix         = decimal.New(6)
	decTwelve      = decimal.New(12)
	decOnePointOne = decimal.New(1.1)
	decFib382      = decimal.New(0.382)
	decFib618      = decimal.New(0.618)
	decOne         = decimal.New(1.0)
)

// PivotPointResult holds all levels for standard pivot points.
type PivotPointResult struct {
	Pivot      decimal.Decimal
	R1, R2, R3 decimal.Decimal
	S1, S2, S3 decimal.Decimal
}

// pivotPointsIndicator calculates standard (classic) Pivot Points.
// Kept for backward compatibility; Calculate returns the Pivot value.
type pivotPointsIndicator struct {
	Indicator
	series *series.TimeSeries
}

// NewPivotPointsIndicator returns an indicator that calculates standard Pivot Points.
// It uses the previous candle's H, L, C to calculate today's levels.
// Kept for backward compatibility. Panics if s is nil.
func NewPivotPointsIndicator(s *series.TimeSeries) *pivotPointsIndicator {
	if s == nil {
		panic("goflux: PivotPoints series cannot be nil")
	}
	return &pivotPointsIndicator{series: s}
}

func (p *pivotPointsIndicator) Calculate(index int) decimal.Decimal {
	if p.series == nil || index < 0 || index >= len(p.series.Candles) {
		return decimal.ZERO
	}
	res := p.GetLevels(index)
	return res.Pivot
}

// GetLevels returns all standard pivot levels at the given index.
func (p *pivotPointsIndicator) GetLevels(index int) PivotPointResult {
	if p.series == nil || index <= 0 || index >= len(p.series.Candles) {
		return PivotPointResult{}
	}
	prev := p.series.Candles[index-1]
	return calculateStandardPivotPointResult(prev.MaxPrice, prev.MinPrice, prev.ClosePrice)
}

// --- Standard (Classic) Pivot Points ---

// NewPivotPointIndicators returns standard Pivot Point indicators (P, R1-R3, S1-S3).
// Each level is computed from the previous candle's H, L, C.
// Panics if s is nil.
func NewPivotPointIndicators(s *series.TimeSeries) (pp, r1, r2, r3, s1, s2, s3 Indicator) {
	if s == nil {
		panic("goflux: PivotPoints series cannot be nil")
	}
	telemetry.ReportUsage("PivotPoints", nil)
	levels := newPivotPointBase(s, func(h, l, c decimal.Decimal) pivotPointValues {
		res := calculateStandardPivotPointResult(h, l, c)
		return pivotPointValues{
			pp: res.Pivot,
			r1: res.R1, r2: res.R2, r3: res.R3,
			s1: res.S1, s2: res.S2, s3: res.S3,
		}
	})
	return levels.pp, levels.r1, levels.r2, levels.r3, levels.s1, levels.s2, levels.s3
}

// --- Camarilla Pivot Points ---

// NewCamarillaPivotPointIndicators returns Camarilla indicators (P, R1-R4, S1-S4).
// Panics if s is nil.
func NewCamarillaPivotPointIndicators(s *series.TimeSeries) (pp, r1, r2, r3, r4, s1, s2, s3, s4 Indicator) {
	if s == nil {
		panic("goflux: CamarillaPivotPoints series cannot be nil")
	}
	telemetry.ReportUsage("CamarillaPivotPoints", nil)
	base := newPivotPointBase(s, func(h, l, c decimal.Decimal) pivotPointValues {
		range_ := h.Sub(l)
		pp := h.Add(l).Add(c).Div(decThree)
		return pivotPointValues{
			pp: pp,
			r1: c.Add(range_.Mul(decOnePointOne.Div(decTwelve))),
			r2: c.Add(range_.Mul(decOnePointOne.Div(decSix))),
			r3: c.Add(range_.Mul(decOnePointOne.Div(decFour))),
			r4: c.Add(range_.Mul(decOnePointOne.Div(decTwo))),
			s1: c.Sub(range_.Mul(decOnePointOne.Div(decTwelve))),
			s2: c.Sub(range_.Mul(decOnePointOne.Div(decSix))),
			s3: c.Sub(range_.Mul(decOnePointOne.Div(decFour))),
			s4: c.Sub(range_.Mul(decOnePointOne.Div(decTwo))),
		}
	})
	return base.pp, base.r1, base.r2, base.r3, base.r4, base.s1, base.s2, base.s3, base.s4
}

// --- Woodie Pivot Points ---

// NewWoodiePivotPointIndicators returns Woodie indicators (P, R1-R3, S1-S3).
// Panics if s is nil.
func NewWoodiePivotPointIndicators(s *series.TimeSeries) (pp, r1, r2, r3, s1, s2, s3 Indicator) {
	if s == nil {
		panic("goflux: WoodiePivotPoints series cannot be nil")
	}
	telemetry.ReportUsage("WoodiePivotPoints", nil)
	levels := newPivotPointBase(s, func(h, l, c decimal.Decimal) pivotPointValues {
		pp := h.Add(l).Add(c.Mul(decTwo)).Div(decFour)
		return pivotPointValues{
			pp: pp,
			r1: pp.Mul(decTwo).Sub(l),
			r2: pp.Add(h.Sub(l)),
			r3: h.Add(pp.Sub(l).Mul(decTwo)),
			s1: pp.Mul(decTwo).Sub(h),
			s2: pp.Sub(h.Sub(l)),
			s3: l.Sub(h.Sub(pp).Mul(decTwo)),
		}
	})
	return levels.pp, levels.r1, levels.r2, levels.r3, levels.s1, levels.s2, levels.s3
}

// --- Fibonacci Pivot Points ---

// NewFibonacciPivotPointIndicators returns Fibonacci indicators (P, R1-R3, S1-S3).
// Panics if s is nil.
func NewFibonacciPivotPointIndicators(s *series.TimeSeries) (pp, r1, r2, r3, s1, s2, s3 Indicator) {
	if s == nil {
		panic("goflux: FibonacciPivotPoints series cannot be nil")
	}
	telemetry.ReportUsage("FibonacciPivotPoints", nil)
	levels := newPivotPointBase(s, func(h, l, c decimal.Decimal) pivotPointValues {
		pp := h.Add(l).Add(c).Div(decThree)
		range_ := h.Sub(l)
		return pivotPointValues{
			pp: pp,
			r1: pp.Add(range_.Mul(decFib382)),
			r2: pp.Add(range_.Mul(decFib618)),
			r3: pp.Add(range_.Mul(decOne)),
			s1: pp.Sub(range_.Mul(decFib382)),
			s2: pp.Sub(range_.Mul(decFib618)),
			s3: pp.Sub(range_.Mul(decOne)),
		}
	})
	return levels.pp, levels.r1, levels.r2, levels.r3, levels.s1, levels.s2, levels.s3
}

// --- internal helpers ---

type pivotPointLevels struct {
	pp         Indicator
	r1, r2, r3 Indicator
	r4         Indicator
	s1, s2, s3 Indicator
	s4         Indicator
}

type pivotPointValues struct {
	pp, r1, r2, r3, r4, s1, s2, s3, s4 decimal.Decimal
}

type pivotPointCalculator func(high, low, close decimal.Decimal) pivotPointValues

type pivotPointBase struct {
	series *series.TimeSeries
	calc   pivotPointCalculator
	cache  map[int]pivotPointValues
}

func (b *pivotPointBase) getValues(index int) pivotPointValues {
	if b == nil || b.series == nil || index <= 0 || index >= len(b.series.Candles) {
		return pivotPointValues{}
	}
	if v, ok := b.cache[index]; ok {
		return v
	}
	prev := b.series.Candles[index-1]
	v := b.calc(prev.MaxPrice, prev.MinPrice, prev.ClosePrice)
	if b.cache == nil {
		b.cache = make(map[int]pivotPointValues)
	}
	b.cache[index] = v
	return v
}

type levelField int

const (
	levelPP levelField = iota
	levelR1
	levelR2
	levelR3
	levelR4
	levelS1
	levelS2
	levelS3
	levelS4
)

func newPivotPointBase(s *series.TimeSeries, calc pivotPointCalculator) pivotPointLevels {
	base := &pivotPointBase{series: s, calc: calc}
	return pivotPointLevels{
		pp: pivotLevel{base: base, field: levelPP},
		r1: pivotLevel{base: base, field: levelR1},
		r2: pivotLevel{base: base, field: levelR2},
		r3: pivotLevel{base: base, field: levelR3},
		r4: pivotLevel{base: base, field: levelR4},
		s1: pivotLevel{base: base, field: levelS1},
		s2: pivotLevel{base: base, field: levelS2},
		s3: pivotLevel{base: base, field: levelS3},
		s4: pivotLevel{base: base, field: levelS4},
	}
}

type pivotLevel struct {
	base  *pivotPointBase
	field levelField
}

func (p pivotLevel) Calculate(index int) decimal.Decimal {
	levels := p.base.getValues(index)
	switch p.field {
	case levelPP:
		return levels.pp
	case levelR1:
		return levels.r1
	case levelR2:
		return levels.r2
	case levelR3:
		return levels.r3
	case levelR4:
		return levels.r4
	case levelS1:
		return levels.s1
	case levelS2:
		return levels.s2
	case levelS3:
		return levels.s3
	case levelS4:
		return levels.s4
	}
	return decimal.ZERO
}

func calculateStandardPivotPointResult(high, low, close decimal.Decimal) PivotPointResult {
	pivot := high.Add(low).Add(close).Div(decThree)
	r1 := pivot.Mul(decTwo).Sub(low)
	s1 := pivot.Mul(decTwo).Sub(high)
	r2 := pivot.Add(high.Sub(low))
	s2 := pivot.Sub(high.Sub(low))
	r3 := high.Add(pivot.Sub(low).Mul(decTwo))
	s3 := low.Sub(high.Sub(pivot).Mul(decTwo))
	return PivotPointResult{
		Pivot: pivot,
		R1:    r1, R2: r2, R3: r3,
		S1: s1, S2: s2, S3: s3,
	}
}
