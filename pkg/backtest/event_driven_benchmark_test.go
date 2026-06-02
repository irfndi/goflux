package backtest

import (
	"testing"
	"time"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/trading"
)

// edBenchStrategy enters at index > 0 when position is new and exits after 5 bars.
type edBenchStrategy struct{}

func (s *edBenchStrategy) ShouldEnter(index int, record *trading.TradingRecord) bool {
	return index > 0 && record.CurrentPosition().IsNew()
}

func (s *edBenchStrategy) ShouldExit(index int, record *trading.TradingRecord) bool {
	if !record.CurrentPosition().IsOpen() {
		return false
	}
	entryOrder := record.CurrentPosition().EntranceOrder()
	// Approximate 5-bar hold using creation time as index proxy
	return index > int(entryOrder.CreationTime.Unix())+5
}

func benchmarkEventSeries(n int) []*series.Candle {
	candles := make([]*series.Candle, n)
	base := time.Now()
	for i := 0; i < n; i++ {
		closePrice := decimal.New(float64(100 + i%50))
		candles[i] = &series.Candle{
			OpenPrice:  closePrice.Sub(decimal.New(1)),
			ClosePrice: closePrice,
			MaxPrice:   closePrice.Add(decimal.New(2)),
			MinPrice:   closePrice.Sub(decimal.New(2)),
			Volume:     decimal.New(1000),
			Period:     series.NewTimePeriod(base.Add(time.Duration(i)*time.Hour), time.Hour),
		}
	}
	return candles
}

func BenchmarkEventDrivenBacktester_Small(b *testing.B) {
	candles := benchmarkEventSeries(100)
	events := createTestEvents("TEST", candles)
	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &edBenchStrategy{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := edb.Run(events)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEventDrivenBacktester_Medium(b *testing.B) {
	candles := benchmarkEventSeries(1000)
	events := createTestEvents("TEST", candles)
	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &edBenchStrategy{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := edb.Run(events)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEventDrivenBacktester_Large(b *testing.B) {
	candles := benchmarkEventSeries(5000)
	events := createTestEvents("TEST", candles)
	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &edBenchStrategy{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := edb.Run(events)
		if err != nil {
			b.Fatal(err)
		}
	}
}
