package indicators

import (
	"math"

	"github.com/irfndi/goflux/pkg/decimal"
)

// DominantCyclePeriod uses John Ehlers' algorithm to find the dominant cycle in price data.
type dominantCyclePeriod struct {
	indicator Indicator
}

func NewDominantCyclePeriod(indicator Indicator) Indicator {
	return dominantCyclePeriod{indicator}
}

func (dcp dominantCyclePeriod) Calculate(index int) decimal.Decimal {
	if dcp.indicator == nil || index < 7 {
		return decimal.ZERO
	}

	maxPeriod := index / 2
	if maxPeriod > 50 {
		maxPeriod = 50
	}
	if maxPeriod < 7 {
		return decimal.New(7)
	}

	bestPeriod := 7
	bestScore := math.Inf(-1)
	for period := 7; period <= maxPeriod; period++ {
		count := minInt(period, index-period+1)
		var sumX, sumY, sumXX, sumYY, sumXY float64
		for i := 0; i < count; i++ {
			x := dcp.indicator.Calculate(index - i).Float()
			y := dcp.indicator.Calculate(index - i - period).Float()
			sumX += x
			sumY += y
			sumXX += x * x
			sumYY += y * y
			sumXY += x * y
		}
		countFloat := float64(count)
		covariance := countFloat*sumXY - sumX*sumY
		variance := (countFloat*sumXX - sumX*sumX) * (countFloat*sumYY - sumY*sumY)
		score := 0.0
		if variance > 0 {
			score = covariance / math.Sqrt(variance)
		}
		if score > bestScore {
			bestScore = score
			bestPeriod = period
		}
	}
	return decimal.New(float64(bestPeriod))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// HilbertTransform provides the In-Phase and Quadrature components of a signal
type hilbertTransform struct {
	indicator Indicator
}

func NewHilbertTransform(indicator Indicator) Indicator {
	return hilbertTransform{indicator}
}

func (ht hilbertTransform) Calculate(index int) decimal.Decimal {
	if index < 7 {
		return ht.indicator.Calculate(index)
	}

	// Hilbert Transform: H(x) = (x(i) - x(i-6)) * 0.125 + (x(i-2) - x(i-4)) * 0.485
	// This is a simplified digital filter approximation
	val0 := ht.indicator.Calculate(index)
	val2 := ht.indicator.Calculate(index - 2)
	val4 := ht.indicator.Calculate(index - 4)
	val6 := ht.indicator.Calculate(index - 6)

	res := val0.Sub(val6).Mul(decimal.New(0.125)).Add(val2.Sub(val4).Mul(decimal.New(0.485)))
	return res
}

// HilbertTransformInstantaneousTrendline calculates Ehlers' Instantaneous Trendline
type htTrendline struct {
	indicator Indicator
}

func NewHTTrendline(indicator Indicator) Indicator {
	return htTrendline{indicator}
}

func (htt htTrendline) Calculate(index int) decimal.Decimal {
	if index < 12 {
		return htt.indicator.Calculate(index)
	}

	// WMA calculation for trendline
	// (4*p0 + 3*p1 + 2*p2 + p3) / 10
	p0 := htt.indicator.Calculate(index)
	p1 := htt.indicator.Calculate(index - 1)
	p2 := htt.indicator.Calculate(index - 2)
	p3 := htt.indicator.Calculate(index - 3)

	return p0.Mul(decimal.New(0.4)).Add(p1.Mul(decimal.New(0.3))).Add(p2.Mul(decimal.New(0.2))).Add(p3.Mul(decimal.New(0.1)))
}
