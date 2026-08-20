package indicators

import (
	"sync"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
)

type parabolicSARIndicator struct {
	Indicator
	series      *series.TimeSeries
	high        Indicator
	low         Indicator
	af          decimal.Decimal
	maxAF       decimal.Decimal
	prevSAR     decimal.Decimal
	prevEP      decimal.Decimal
	prevAF      decimal.Decimal
	trend       int
	initialized bool
	results     []decimal.Decimal
	mu          sync.Mutex
}

func NewParabolicSARIndicator(s *series.TimeSeries) Indicator {
	return &parabolicSARIndicator{
		series: s,
		high:   NewHighPriceIndicator(s),
		low:    NewLowPriceIndicator(s),
		af:     decimal.New(0.02),
		maxAF:  decimal.New(0.2),
		trend:  0,
	}
}

func (ps *parabolicSARIndicator) Calculate(index int) decimal.Decimal {
	if ps == nil || ps.series == nil || index < 0 || index >= ps.series.Length() {
		return decimal.ZERO
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if index < len(ps.results) {
		return ps.results[index]
	}
	if len(ps.results) == 0 {
		ps.results = append(ps.results, ps.high.Calculate(0))
	}
	if !ps.initialized && index >= 1 {
		ps.initialize()
		ps.results = append(ps.results, ps.prevSAR)
	}
	for i := len(ps.results); i <= index; i++ {
		ps.step(i)
		ps.results = append(ps.results, ps.prevSAR)
	}
	return ps.results[index]
}

func (ps *parabolicSARIndicator) step(index int) {
	currentHigh := ps.high.Calculate(index)
	currentLow := ps.low.Calculate(index)
	previousHigh := ps.high.Calculate(index - 1)
	previousLow := ps.low.Calculate(index - 1)

	switch ps.trend {
	case 1:
		if currentHigh.GT(ps.prevEP) {
			ps.prevEP = currentHigh
			ps.prevAF = ps.prevAF.Add(ps.af)
			if ps.prevAF.GT(ps.maxAF) {
				ps.prevAF = ps.maxAF
			}
		}
		ps.prevSAR = ps.prevSAR.Add(ps.prevAF.Mul(ps.prevEP.Sub(ps.prevSAR)))

		if currentLow.LT(ps.prevSAR) {
			ps.trend = -1
			ps.prevAF = ps.af
			ps.prevEP = previousLow
			ps.prevSAR = ps.prevEP
		}
	case -1:
		if currentLow.LT(ps.prevEP) {
			ps.prevEP = currentLow
			ps.prevAF = ps.prevAF.Add(ps.af)
			if ps.prevAF.GT(ps.maxAF) {
				ps.prevAF = ps.maxAF
			}
		}
		ps.prevSAR = ps.prevSAR.Add(ps.prevAF.Mul(ps.prevEP.Sub(ps.prevSAR)))

		if currentHigh.GT(ps.prevSAR) {
			ps.trend = 1
			ps.prevAF = ps.af
			ps.prevEP = previousHigh
			ps.prevSAR = ps.prevEP
		}
	}

}

func (ps *parabolicSARIndicator) initialize() {
	if ps.series == nil || ps.series.Length() < 2 {
		return
	}
	firstHigh := ps.high.Calculate(0)
	firstLow := ps.low.Calculate(0)
	secondHigh := ps.high.Calculate(1)
	secondLow := ps.low.Calculate(1)

	if secondHigh.GT(firstHigh) && secondLow.GT(firstLow) {
		ps.trend = 1
		ps.prevEP = secondHigh
		ps.prevSAR = firstLow
	} else if secondHigh.LT(firstHigh) && secondLow.LT(firstLow) {
		ps.trend = -1
		ps.prevEP = secondLow
		ps.prevSAR = firstHigh
	} else if secondHigh.GT(firstHigh) {
		ps.trend = 1
		ps.prevEP = secondHigh
		ps.prevSAR = secondLow
	} else {
		ps.trend = -1
		ps.prevEP = secondLow
		ps.prevSAR = secondHigh
	}

	ps.prevAF = ps.af
	ps.initialized = true
}

func (ps *parabolicSARIndicator) Trend() int {
	return ps.trend
}

func (ps *parabolicSARIndicator) EP() decimal.Decimal {
	return ps.prevEP
}

func (ps *parabolicSARIndicator) AF() decimal.Decimal {
	return ps.prevAF
}
