package indicators

import (
	"testing"

	"github.com/irfndi/goflux/pkg/testutils"
)

const benchSize = 10000

// benchmarkIndicator is a helper that benchmarks a single Calculate call
// on an indicator constructed from a random time series.
func benchmarkIndicator(b *testing.B, ind Indicator) {
	b.Helper()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ind.Calculate(benchSize - 1)
	}
}

// benchmarkIndicatorConstruction benchmarks constructor + single Calculate.
func benchmarkIndicatorConstruction(b *testing.B, factory func() Indicator) {
	b.Helper()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		factory().Calculate(benchSize - 1)
	}
}

// --- Moving Averages ---

func BenchmarkSimpleMovingAverage(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewSimpleMovingAverage(NewClosePriceIndicator(ts), 20)
	})
}

func BenchmarkExponentialMovingAverage(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewEMAIndicator(NewClosePriceIndicator(ts), 20)
	})
}

func BenchmarkHullMovingAverage(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewHMAIndicator(NewClosePriceIndicator(ts), 16)
	})
}

func BenchmarkKaufmanAdaptiveMA(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewKAMAIndicator(ts, 10)
	})
}

func BenchmarkModifiedMovingAverage(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewMMAIndicator(NewClosePriceIndicator(ts), 20)
	})
}

// --- Momentum / Oscillators ---

func BenchmarkRelativeStrengthIndex(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewRelativeStrengthIndexIndicator(NewClosePriceIndicator(ts), 14)
	})
}

func BenchmarkMACD(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewMACDIndicator(NewClosePriceIndicator(ts), 12, 26)
	})
}

func BenchmarkStochasticOscillator_Fast(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewFastStochasticIndicator(ts, 14)
	})
}

func BenchmarkStochasticOscillator_Slow(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewSlowStochasticIndicator(NewFastStochasticIndicator(ts, 14), 3)
	})
}

func BenchmarkCCI(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewCCIIndicator(ts, 20)
	})
}

func BenchmarkADX(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewADXIndicator(ts, 14)
	})
}

func BenchmarkAwesomeOscillator(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewDefaultAwesomeOscillatorIndicator(ts)
	})
}

func BenchmarkWilliamsR(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewWilliamsRIndicator(ts, 14)
	})
}

func BenchmarkRateOfChange(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewROCIndicator(ts, 14)
	})
}

func BenchmarkMoneyFlowIndex(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewMFIIndicator(ts, 14)
	})
}

func BenchmarkUltimateOscillator(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewUltimateOscillatorIndicator(ts, 7, 14, 28)
	})
}

// --- Volatility ---

func BenchmarkBollingerUpperBand(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewBollingerUpperBandIndicator(NewClosePriceIndicator(ts), 20, 2.0)
	})
}

func BenchmarkBollingerLowerBand(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewBollingerLowerBandIndicator(NewClosePriceIndicator(ts), 20, 2.0)
	})
}

func BenchmarkAverageTrueRange(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewAverageTrueRangeIndicator(ts, 14)
	})
}

func BenchmarkKeltnerChannelUpper(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewKeltnerChannelUpperIndicator(ts, 20)
	})
}

// --- Volume ---

func BenchmarkOnBalanceVolume(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewOBVIndicator(ts)
	})
}

func BenchmarkVWAP(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewVWAPIndicator(ts)
	})
}

func BenchmarkAccumulationDistribution(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewADLineIndicator(ts)
	})
}

// --- Trend ---

func BenchmarkParabolicSAR(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewParabolicSARIndicator(ts)
	})
}

func BenchmarkIchimoku(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewIchimokuIndicator(ts)
	})
}

func BenchmarkSuperTrend(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewSuperTrendIndicator(ts, 10, 3.0)
	})
}

func BenchmarkVortex(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	benchmarkIndicatorConstruction(b, func() Indicator {
		return NewVortexIndicator(ts, 14)
	})
}

// --- Cached vs Uncached ---

func BenchmarkSimpleMovingAverage_Cached(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	sma := NewSimpleMovingAverage(NewClosePriceIndicator(ts), 20)
	benchmarkIndicator(b, sma)
}

func BenchmarkRelativeStrengthIndex_Cached(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	rsi := NewRelativeStrengthIndexIndicator(NewClosePriceIndicator(ts), 14)
	benchmarkIndicator(b, rsi)
}

func BenchmarkMACD_Cached(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	macd := NewMACDIndicator(NewClosePriceIndicator(ts), 12, 26)
	benchmarkIndicator(b, macd)
}

func BenchmarkBollingerUpperBand_Cached(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	bb := NewBollingerUpperBandIndicator(NewClosePriceIndicator(ts), 20, 2.0)
	benchmarkIndicator(b, bb)
}

func BenchmarkAverageTrueRange_Cached(b *testing.B) {
	ts := testutils.RandomTimeSeries(benchSize)
	atr := NewAverageTrueRangeIndicator(ts, 14)
	benchmarkIndicator(b, atr)
}
