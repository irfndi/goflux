package indicators

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/testutils"
)

const benchSize = 10000

var (
	benchmarkResult  decimal.Decimal
	sharedTimeSeries *series.TimeSeries
)

func init() {
	sharedTimeSeries = testutils.RandomTimeSeries(benchSize)
}

// benchmarkIndicator is a helper that benchmarks a single Calculate call
// on an indicator constructed from a random time series.
func benchmarkIndicator(b *testing.B, ind Indicator) {
	b.Helper()
	benchmarkResult = ind.Calculate(benchSize - 1) // warm up cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkResult = ind.Calculate(benchSize - 1)
	}
}

// benchmarkIndicatorConstruction benchmarks constructor + single Calculate.
func benchmarkIndicatorConstruction(b *testing.B, factory func() Indicator) {
	b.Helper()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkResult = factory().Calculate(benchSize - 1)
	}
}

// --- Moving Averages ---

func BenchmarkSimpleMovingAverage(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewSimpleMovingAverage(NewClosePriceIndicator(sharedTimeSeries), 20)
	})
}

func BenchmarkExponentialMovingAverage(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewEMAIndicator(NewClosePriceIndicator(sharedTimeSeries), 20)
	})
}

func BenchmarkHullMovingAverage(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewHMAIndicator(NewClosePriceIndicator(sharedTimeSeries), 16)
	})
}

func BenchmarkKaufmanAdaptiveMA(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewKAMAIndicator(sharedTimeSeries, 10)
	})
}

func BenchmarkModifiedMovingAverage(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewMMAIndicator(NewClosePriceIndicator(sharedTimeSeries), 20)
	})
}

// --- Momentum / Oscillators ---

func BenchmarkRelativeStrengthIndex(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewRelativeStrengthIndexIndicator(NewClosePriceIndicator(sharedTimeSeries), 14)
	})
}

func BenchmarkMACD(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewMACDIndicator(NewClosePriceIndicator(sharedTimeSeries), 12, 26)
	})
}

func BenchmarkStochasticOscillator_Fast(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewFastStochasticIndicator(sharedTimeSeries, 14)
	})
}

func BenchmarkStochasticOscillator_Slow(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewSlowStochasticIndicator(NewFastStochasticIndicator(sharedTimeSeries, 14), 3)
	})
}

func BenchmarkCCI(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewCCIIndicator(sharedTimeSeries, 20)
	})
}

func BenchmarkADX(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewADXIndicator(sharedTimeSeries, 14)
	})
}

func BenchmarkAwesomeOscillator(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewDefaultAwesomeOscillatorIndicator(sharedTimeSeries)
	})
}

func BenchmarkWilliamsR(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewWilliamsRIndicator(sharedTimeSeries, 14)
	})
}

func BenchmarkRateOfChange(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewROCIndicator(sharedTimeSeries, 14)
	})
}

func BenchmarkMoneyFlowIndex(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewMFIIndicator(sharedTimeSeries, 14)
	})
}

func BenchmarkUltimateOscillator(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewUltimateOscillatorIndicator(sharedTimeSeries, 7, 14, 28)
	})
}

// --- Volatility ---

func BenchmarkBollingerUpperBand(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewBollingerUpperBandIndicator(NewClosePriceIndicator(sharedTimeSeries), 20, 2.0)
	})
}

func BenchmarkBollingerLowerBand(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewBollingerLowerBandIndicator(NewClosePriceIndicator(sharedTimeSeries), 20, 2.0)
	})
}

func BenchmarkAverageTrueRange(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewAverageTrueRangeIndicator(sharedTimeSeries, 14)
	})
}

func BenchmarkKeltnerChannelUpper(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewKeltnerChannelUpperIndicator(sharedTimeSeries, 20)
	})
}

// --- Volume ---

func BenchmarkOnBalanceVolume(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewOBVIndicator(sharedTimeSeries)
	})
}

func BenchmarkVWAP(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewVWAPIndicator(sharedTimeSeries)
	})
}

func BenchmarkAccumulationDistribution(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewADLineIndicator(sharedTimeSeries)
	})
}

// --- Trend ---

func BenchmarkParabolicSAR(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewParabolicSARIndicator(sharedTimeSeries)
	})
}

func BenchmarkIchimoku(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewIchimokuIndicator(sharedTimeSeries)
	})
}

func BenchmarkSuperTrend(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewSuperTrendIndicator(sharedTimeSeries, 10, 3.0)
	})
}

func BenchmarkVortex(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewVortexIndicator(sharedTimeSeries, 14)
	})
}

// --- Cached vs Uncached ---

func BenchmarkSimpleMovingAverage_Cached(b *testing.B) {
	sma := NewSimpleMovingAverage(NewClosePriceIndicator(sharedTimeSeries), 20)
	benchmarkIndicator(b, sma)
}

func BenchmarkRelativeStrengthIndex_Cached(b *testing.B) {
	rsi := NewRelativeStrengthIndexIndicator(NewClosePriceIndicator(sharedTimeSeries), 14)
	benchmarkIndicator(b, rsi)
}

func BenchmarkMACD_Cached(b *testing.B) {
	macd := NewMACDIndicator(NewClosePriceIndicator(sharedTimeSeries), 12, 26)
	benchmarkIndicator(b, macd)
}

func BenchmarkBollingerUpperBand_Cached(b *testing.B) {
	bb := NewBollingerUpperBandIndicator(NewClosePriceIndicator(sharedTimeSeries), 20, 2.0)
	benchmarkIndicator(b, bb)
}

func BenchmarkAverageTrueRange_Cached(b *testing.B) {
	atr := NewAverageTrueRangeIndicator(sharedTimeSeries, 14)
	benchmarkIndicator(b, atr)
}

// --- Exit / Risk Management ---

func BenchmarkChandelierExitLong(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewChandelierExitLong(sharedTimeSeries, 22, 22, 3.0)
	})
}

func BenchmarkChandelierExitShort(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewChandelierExitShort(sharedTimeSeries, 22, 22, 3.0)
	})
}

func BenchmarkAroonOscillator(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewAroonOscillatorFromSeries(sharedTimeSeries, 14)
	})
}

func BenchmarkChaikinMoneyFlow(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewChaikinMoneyFlowIndicator(sharedTimeSeries, 20)
	})
}

func BenchmarkTRIX(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewTRIXIndicatorFromSeries(sharedTimeSeries, 14)
	})
}

func BenchmarkDonchianUpperBand(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewDonchianUpperBandIndicator(sharedTimeSeries, 20)
	})
}

func BenchmarkDonchianLowerBand(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewDonchianLowerBandIndicator(sharedTimeSeries, 20)
	})
}

func BenchmarkDonchianMiddleBand(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewDonchianMiddleBandIndicator(sharedTimeSeries, 20)
	})
}

// --- Fibonacci ---

func BenchmarkFibonacciRetracementIndicator(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewFibonacciRetracementIndicator(sharedTimeSeries, 20, 0.618)
	})
}

// --- Linear Regression ---

func BenchmarkLinearRegression(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewLinearRegressionIndicator(NewClosePriceIndicator(sharedTimeSeries), 20)
	})
}

func BenchmarkLinearRegressionChannel(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		mid, upper, _ := NewLinearRegressionChannel(NewClosePriceIndicator(sharedTimeSeries), 20, 2.0)
		_ = mid
		return upper
	})
}

// --- Alligator ---

func BenchmarkAlligatorJaw(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		jaw, _, _ := NewAlligatorIndicators(sharedTimeSeries)
		return jaw
	})
}

func BenchmarkGatorOscillatorUpper(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		upper, _ := NewGatorOscillatorIndicators(sharedTimeSeries)
		return upper
	})
}

// --- Moving Averages Extended ---

func BenchmarkT3(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewT3Indicator(NewClosePriceIndicator(sharedTimeSeries), 6, 0.7)
	})
}

func BenchmarkALMA(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewALMAIndicator(NewClosePriceIndicator(sharedTimeSeries), 9, 0.85, 6.0)
	})
}

func BenchmarkVIDYA(b *testing.B) {
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewVIDYAIndicator(NewClosePriceIndicator(sharedTimeSeries), 14)
	})
}
