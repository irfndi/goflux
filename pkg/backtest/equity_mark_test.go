package backtest

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/metrics"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/trading"
)

func TestEquityMarkTracksLongAndShortPositions(t *testing.T) {
	tests := []struct {
		name      string
		direction string
		entry     float64
		mark      float64
		quantity  float64
		want      float64
	}{
		{name: "long profit", direction: "long", entry: 100, mark: 110, quantity: 2, want: 20},
		{name: "long loss", direction: "long", entry: 100, mark: 90, quantity: 2, want: -20},
		{name: "short profit", direction: "short", entry: 100, mark: 90, quantity: 2, want: 20},
		{name: "short loss", direction: "short", entry: 100, mark: 110, quantity: 2, want: -20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := decimal.New(tt.entry)
			markPrice := decimal.New(tt.mark)
			quantity := decimal.New(tt.quantity)
			var mark equityMark

			mark.advance(entry)
			mark.add(tt.direction, entry, quantity, entry)
			mark.advance(markPrice)

			assert.True(t, mark.value.EQ(decimal.New(tt.want)))
			mark.remove(tt.direction, entry, quantity, markPrice)
			assert.True(t, mark.value.IsZero())
			assert.True(t, mark.netQuantity.IsZero())
		})
	}
}

func TestEquityMarkAggregatesPositions(t *testing.T) {
	var mark equityMark
	mark.advance(decimal.New(100))
	mark.add("long", decimal.New(100), decimal.New(2), decimal.New(100))
	mark.add("short", decimal.New(200), decimal.New(3), decimal.New(100))
	mark.advance(decimal.New(110))

	// Long: +20, short: +270, total: +290.
	assert.True(t, mark.value.EQ(decimal.New(290)))
	assert.True(t, mark.netQuantity.EQ(decimal.New(-1)))

	mark.advance(decimal.New(120))
	assert.True(t, mark.value.EQ(decimal.New(280)))
}

func TestBacktesterEquityCurveUsesMarkedOpenPositionValue(t *testing.T) {
	ts := series.NewTimeSeries()
	for _, closePrice := range []float64{100, 110, 90} {
		price := decimal.New(closePrice)
		ts.AddCandle(&series.Candle{ClosePrice: price, OpenPrice: price, MaxPrice: price, MinPrice: price})
	}

	backtester := NewBacktester(ts, &enterOnceNoExitStrategy{})
	backtester.AddAnalyzer(&EquityCurveAnalyzer{})
	result := backtester.Run(BacktestConfig{
		InitialCapital: decimal.New(1000),
		PositionSize:   decimal.ONE,
		AllowLong:      true,
	})

	curve := result.Analysis["EquityCurve"].([]metrics.EquityPoint)
	assert.Len(t, curve, 3)
	assert.True(t, curve[0].Equity.EQ(decimal.New(1000)))
	assert.True(t, curve[1].Equity.EQ(decimal.New(1010)))
	assert.True(t, curve[2].Equity.EQ(decimal.New(990)))
}

func TestSimulatedBrokerEquityCurveUsesMarkedOpenPositionValue(t *testing.T) {
	broker := NewSimulatedBroker("TEST", decimal.New(1000))
	order := trading.NewOrderDetail(trading.BUY, trading.MarketOrder, "TEST", decimal.ONE)
	broker.SubmitOrder(order)

	for index, closePrice := range []float64{100, 110, 90} {
		price := decimal.New(closePrice)
		broker.ProcessBar(index, &series.Candle{ClosePrice: price, OpenPrice: price, MaxPrice: price, MinPrice: price})
	}

	assert.Len(t, broker.equityHistory, 3)
	assert.True(t, broker.equityHistory[0].EQ(decimal.New(1000)))
	assert.True(t, broker.equityHistory[1].EQ(decimal.New(1010)))
	assert.True(t, broker.equityHistory[2].EQ(decimal.New(990)))
}
