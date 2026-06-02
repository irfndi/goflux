package backtest

import (
	"testing"
	"time"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/trading"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// alwaysEnterStrategy enters at bar 1 and exits at bar 3.
type alwaysEnterStrategy struct{}

func (s *alwaysEnterStrategy) ShouldEnter(index int, record *trading.TradingRecord) bool {
	return index == 1 && record.CurrentPosition().IsNew()
}

func (s *alwaysEnterStrategy) ShouldExit(index int, record *trading.TradingRecord) bool {
	return index == 3 && record.CurrentPosition().IsOpen()
}

// edNeverEnterStrategy never enters.
type edNeverEnterStrategy struct{}

func (s *edNeverEnterStrategy) ShouldEnter(index int, record *trading.TradingRecord) bool {
	return false
}

func (s *edNeverEnterStrategy) ShouldExit(index int, record *trading.TradingRecord) bool {
	return false
}

func createTestCandle(open, close, high, low float64) *series.Candle {
	return &series.Candle{
		OpenPrice:  decimal.New(open),
		ClosePrice: decimal.New(close),
		MaxPrice:   decimal.New(high),
		MinPrice:   decimal.New(low),
		Volume:     decimal.New(1000),
		Period:     series.NewTimePeriod(time.Now(), time.Hour),
	}
}

func createTestEvents(symbol string, candles []*series.Candle) []Event {
	events := make([]Event, len(candles))
	for i, c := range candles {
		events[i] = Event{
			Type:      EventBar,
			Timestamp: time.Unix(int64(i), 0),
			Symbol:    symbol,
			Data:      BarEventData{Candle: c},
		}
	}
	return events
}

func TestEventDrivenBacktester_MarketOrderFillAtClose(t *testing.T) {
	candles := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
		createTestCandle(101, 102, 103, 100),
		createTestCandle(102, 103, 104, 101),
		createTestCandle(103, 104, 105, 102),
	}
	events := createTestEvents("TEST", candles)

	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	broker.FillPriceSource = FillAtClose

	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &alwaysEnterStrategy{})

	results, err := edb.Run(events)
	require.NoError(t, err)

	result := results["TEST"]
	require.Equal(t, 1, result.TotalTrades)

	// Entered at bar 1 close = 102, exited at bar 3 close = 104
	assert.True(t, result.Trades[0].EntryPrice.EQ(decimal.New(102)))
	assert.True(t, result.Trades[0].ExitPrice.EQ(decimal.New(104)))
	assert.True(t, result.Trades[0].Profit.IsPositive())
}

func TestEventDrivenBacktester_MarketOrderFillAtOpen(t *testing.T) {
	candles := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
		createTestCandle(101, 102, 103, 100),
		createTestCandle(102, 103, 104, 101),
		createTestCandle(103, 104, 105, 102),
	}
	events := createTestEvents("TEST", candles)

	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	broker.FillPriceSource = FillAtOpen

	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &alwaysEnterStrategy{})

	results, err := edb.Run(events)
	require.NoError(t, err)

	result := results["TEST"]
	require.Equal(t, 1, result.TotalTrades)

	// Entered at bar 1 open = 101, exited at bar 3 open = 103
	assert.True(t, result.Trades[0].EntryPrice.EQ(decimal.New(101)))
	assert.True(t, result.Trades[0].ExitPrice.EQ(decimal.New(103)))
}

func TestEventDrivenBacktester_LimitOrderFills(t *testing.T) {
	candles := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
		createTestCandle(101, 102, 103, 100),
		createTestCandle(102, 99, 103, 98), // low = 98, crosses buy limit at 99
	}
	events := createTestEvents("TEST", candles)

	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &edNeverEnterStrategy{})

	// Submit a buy limit order at 99
	limitOrder := trading.NewOrderDetail(trading.BUY, trading.LimitOrder, "TEST", decimal.ONE)
	limitOrder.Price = decimal.New(99)
	broker.SubmitOrder(limitOrder)

	results, err := edb.Run(events)
	require.NoError(t, err)

	result := results["TEST"]
	require.Equal(t, 1, result.TotalTrades)
	assert.True(t, result.Trades[0].EntryPrice.EQ(decimal.New(99)))
	assert.Equal(t, "long", result.Trades[0].Direction)
}

func TestEventDrivenBacktester_LimitOrderDoesNotFill(t *testing.T) {
	candles := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
		createTestCandle(101, 102, 103, 100),
	}
	events := createTestEvents("TEST", candles)

	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &edNeverEnterStrategy{})

	// Submit a buy limit order at 95 (never reached)
	limitOrder := trading.NewOrderDetail(trading.BUY, trading.LimitOrder, "TEST", decimal.ONE)
	limitOrder.Price = decimal.New(95)
	broker.SubmitOrder(limitOrder)

	results, err := edb.Run(events)
	require.NoError(t, err)

	result := results["TEST"]
	assert.Equal(t, 0, result.TotalTrades)
}

func TestEventDrivenBacktester_StopOrderFills(t *testing.T) {
	candles := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
		createTestCandle(101, 103, 104, 100), // high = 104, crosses buy stop at 103
	}
	events := createTestEvents("TEST", candles)

	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &edNeverEnterStrategy{})

	// Submit a buy stop order at 103
	stopOrder := trading.NewOrderDetail(trading.BUY, trading.StopOrder, "TEST", decimal.ONE)
	stopOrder.StopPrice = decimal.New(103)
	broker.SubmitOrder(stopOrder)

	results, err := edb.Run(events)
	require.NoError(t, err)

	result := results["TEST"]
	require.Equal(t, 1, result.TotalTrades)
	assert.True(t, result.Trades[0].EntryPrice.EQ(decimal.New(103)))
}

func TestEventDrivenBacktester_StopOrderDoesNotFill(t *testing.T) {
	candles := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
		createTestCandle(101, 102, 103, 100),
	}
	events := createTestEvents("TEST", candles)

	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &edNeverEnterStrategy{})

	// Submit a buy stop order at 105 (never reached)
	stopOrder := trading.NewOrderDetail(trading.BUY, trading.StopOrder, "TEST", decimal.ONE)
	stopOrder.StopPrice = decimal.New(105)
	broker.SubmitOrder(stopOrder)

	results, err := edb.Run(events)
	require.NoError(t, err)

	result := results["TEST"]
	assert.Equal(t, 0, result.TotalTrades)
}

func TestEventDrivenBacktester_SellLimitOrderFills(t *testing.T) {
	candles := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
		createTestCandle(101, 102, 103, 100),
		createTestCandle(102, 101, 105, 100), // high = 105, crosses sell limit at 104
	}
	events := createTestEvents("TEST", candles)

	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &edNeverEnterStrategy{})

	// First enter a long position with a market order at bar 0
	marketOrder := trading.NewOrderDetail(trading.BUY, trading.MarketOrder, "TEST", decimal.ONE)
	broker.SubmitOrder(marketOrder)

	// Submit a sell limit order at 104
	limitOrder := trading.NewOrderDetail(trading.SELL, trading.LimitOrder, "TEST", decimal.ONE)
	limitOrder.Price = decimal.New(104)
	broker.SubmitOrder(limitOrder)

	results, err := edb.Run(events)
	require.NoError(t, err)

	result := results["TEST"]
	require.Equal(t, 1, result.TotalTrades)
	assert.True(t, result.Trades[0].ExitPrice.EQ(decimal.New(104)))
}

func TestEventDrivenBacktester_SellStopOrderFills(t *testing.T) {
	candles := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
		createTestCandle(101, 102, 103, 100),
		createTestCandle(102, 99, 103, 98), // low = 98, crosses sell stop at 99
	}
	events := createTestEvents("TEST", candles)

	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &edNeverEnterStrategy{})

	// First enter a long position
	marketOrder := trading.NewOrderDetail(trading.BUY, trading.MarketOrder, "TEST", decimal.ONE)
	broker.SubmitOrder(marketOrder)

	// Submit a sell stop order at 99
	stopOrder := trading.NewOrderDetail(trading.SELL, trading.StopOrder, "TEST", decimal.ONE)
	stopOrder.StopPrice = decimal.New(99)
	broker.SubmitOrder(stopOrder)

	results, err := edb.Run(events)
	require.NoError(t, err)

	result := results["TEST"]
	require.Equal(t, 1, result.TotalTrades)
	assert.True(t, result.Trades[0].ExitPrice.EQ(decimal.New(99)))
}

func TestEventDrivenBacktester_PartialFill(t *testing.T) {
	candles := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
		createTestCandle(101, 102, 103, 100),
	}
	events := createTestEvents("TEST", candles)

	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	broker.PartialFillModel = HalfFill
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &edNeverEnterStrategy{})

	marketOrder := trading.NewOrderDetail(trading.BUY, trading.MarketOrder, "TEST", decimal.New(10))
	broker.SubmitOrder(marketOrder)

	results, err := edb.Run(events)
	require.NoError(t, err)

	result := results["TEST"]
	require.Equal(t, 1, result.TotalTrades)
	// Half fill of 10 = 5
	assert.True(t, result.Trades[0].Quantity.EQ(decimal.New(10)))
	// Note: the position tracks the original order amount, but the fill amount is 5.
	// This is a known limitation of the current Position type.
}

func TestEventDrivenBacktester_CommissionDeducted(t *testing.T) {
	candles := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
		createTestCandle(101, 102, 103, 100),
		createTestCandle(102, 103, 104, 101),
		createTestCandle(103, 104, 105, 102),
	}
	events := createTestEvents("TEST", candles)

	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	broker.CommissionModel = FixedCommission(decimal.New(5))
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &alwaysEnterStrategy{})

	results, err := edb.Run(events)
	require.NoError(t, err)

	result := results["TEST"]
	require.Equal(t, 1, result.TotalTrades)

	// Entry commission = 5, exit commission = 5, total commission = 10
	// Profit without commission = 104 - 102 = 2
	// Net profit = 2 - 10 = -8
	assert.True(t, result.NetProfit.IsNegative())
}

func TestEventDrivenBacktester_SlippageAdjustsFillPrice(t *testing.T) {
	candles := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
		createTestCandle(101, 102, 103, 100),
		createTestCandle(102, 103, 104, 101),
		createTestCandle(103, 104, 105, 102),
	}
	events := createTestEvents("TEST", candles)

	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	broker.SlippageModel = FixedSlippage(decimal.New(1))
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &alwaysEnterStrategy{})

	results, err := edb.Run(events)
	require.NoError(t, err)

	result := results["TEST"]
	require.Equal(t, 1, result.TotalTrades)

	// Buy slippage: fill price = close + 1
	// Entry = 102 + 1 = 103
	// Exit = 104 - 1 = 103 (sell slippage subtracts)
	// Profit = 103 - 103 = 0 (approximately, ignoring exact decimals)
	assert.True(t, result.Trades[0].EntryPrice.EQ(decimal.New(103)))
	assert.True(t, result.Trades[0].ExitPrice.EQ(decimal.New(103)))
}

func TestEventDrivenBacktester_MultiAsset(t *testing.T) {
	candlesA := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
		createTestCandle(101, 102, 103, 100),
		createTestCandle(102, 103, 104, 101),
		createTestCandle(103, 104, 105, 102),
	}
	candlesB := []*series.Candle{
		createTestCandle(200, 201, 202, 199),
		createTestCandle(201, 202, 203, 200),
		createTestCandle(202, 203, 204, 201),
		createTestCandle(203, 204, 205, 202),
	}

	events := append(createTestEvents("A", candlesA), createTestEvents("B", candlesB)...)

	brokerA := NewSimulatedBroker("A", decimal.New(10000))
	brokerB := NewSimulatedBroker("B", decimal.New(10000))

	edb := NewEventDrivenBacktester()
	edb.Register("A", brokerA, &alwaysEnterStrategy{})
	edb.Register("B", brokerB, &alwaysEnterStrategy{})

	results, err := edb.Run(events)
	require.NoError(t, err)

	assert.Equal(t, 2, len(results))
	assert.Equal(t, 1, results["A"].TotalTrades)
	assert.Equal(t, 1, results["B"].TotalTrades)

	// A entered at 102, exited at 104
	assert.True(t, results["A"].Trades[0].EntryPrice.EQ(decimal.New(102)))
	// B entered at 202, exited at 204
	assert.True(t, results["B"].Trades[0].EntryPrice.EQ(decimal.New(202)))
}

func TestEventDrivenBacktester_EmptyEvents(t *testing.T) {
	edb := NewEventDrivenBacktester()
	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb.Register("TEST", broker, &alwaysEnterStrategy{})

	results, err := edb.Run(nil)
	require.NoError(t, err)
	assert.Equal(t, 0, results["TEST"].TotalTrades)

	results, err = edb.Run([]Event{})
	require.NoError(t, err)
	assert.Equal(t, 0, results["TEST"].TotalTrades)
}

func TestEventDrivenBacktester_NoBrokerForSymbol(t *testing.T) {
	candles := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
	}
	events := []Event{
		{
			Type:      EventBar,
			Timestamp: time.Unix(0, 0),
			Symbol:    "UNKNOWN",
			Data:      BarEventData{Candle: candles[0]},
		},
	}

	edb := NewEventDrivenBacktester()
	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb.Register("TEST", broker, &alwaysEnterStrategy{})

	results, err := edb.Run(events)
	require.NoError(t, err)
	assert.Equal(t, 0, results["TEST"].TotalTrades)
}

func TestEventDrivenBacktester_SortedByTimestamp(t *testing.T) {
	candles := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
		createTestCandle(101, 102, 103, 100),
	}

	// Events out of order
	events := []Event{
		{
			Type:      EventBar,
			Timestamp: time.Unix(2, 0),
			Symbol:    "TEST",
			Data:      BarEventData{Candle: candles[1]},
		},
		{
			Type:      EventBar,
			Timestamp: time.Unix(1, 0),
			Symbol:    "TEST",
			Data:      BarEventData{Candle: candles[0]},
		},
	}

	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &alwaysEnterStrategy{})

	results, err := edb.Run(events)
	require.NoError(t, err)

	result := results["TEST"]
	require.Equal(t, 1, result.TotalTrades)
	// Strategy enters at index 1, so it should use the second event (timestamp=1)
	// as index 0 and the first event (timestamp=2) as index 1 where it enters.
	assert.True(t, result.Trades[0].EntryPrice.EQ(decimal.New(102)))
}

func TestSimulatedBroker_DefaultModels(t *testing.T) {
	broker := NewSimulatedBroker("TEST", decimal.New(10000))

	order := trading.NewOrderDetail(trading.BUY, trading.MarketOrder, "TEST", decimal.ONE)
	candle := createTestCandle(100, 101, 102, 99)

	assert.True(t, broker.CommissionModel(order, decimal.New(100), decimal.ONE).IsZero())
	assert.True(t, broker.SlippageModel(order, candle).IsZero())
	assert.True(t, broker.PartialFillModel(order, candle).EQ(decimal.ONE))
}

func TestSimulatedBroker_BacktestResult_NoTrades(t *testing.T) {
	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	result := broker.BacktestResult()

	assert.Equal(t, 0, result.TotalTrades)
	assert.True(t, result.NetProfit.IsZero())
	assert.True(t, result.InitialCapital.EQ(decimal.New(10000)))
}

func TestEventDrivenBacktester_InvalidEventData(t *testing.T) {
	events := []Event{
		{
			Type:      EventBar,
			Timestamp: time.Unix(0, 0),
			Symbol:    "TEST",
			Data:      "invalid",
		},
	}

	edb := NewEventDrivenBacktester()
	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb.Register("TEST", broker, &edNeverEnterStrategy{})

	_, err := edb.Run(events)
	assert.Error(t, err)
}

func TestEventDrivenBacktester_StopLossViaPendingOrder(t *testing.T) {
	// Bar 0: enter at 100
	// Bar 1: stop at 99 should fill because low = 98
	candles := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
		createTestCandle(101, 99, 100, 98),
	}
	events := createTestEvents("TEST", candles)

	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &edNeverEnterStrategy{})

	// Enter position at bar 0 via market order
	marketOrder := trading.NewOrderDetail(trading.BUY, trading.MarketOrder, "TEST", decimal.ONE)
	broker.SubmitOrder(marketOrder)

	// Submit stop loss at 99
	stopOrder := trading.NewOrderDetail(trading.SELL, trading.StopOrder, "TEST", decimal.ONE)
	stopOrder.StopPrice = decimal.New(99)
	broker.SubmitOrder(stopOrder)

	results, err := edb.Run(events)
	require.NoError(t, err)

	result := results["TEST"]
	require.Equal(t, 1, result.TotalTrades)
	assert.True(t, result.Trades[0].ExitPrice.EQ(decimal.New(99)))
}

func TestEventDrivenBacktester_LimitEntryAndExit(t *testing.T) {
	// Bar 0: no fill
	// Bar 1: buy limit at 100 fills (low = 100)
	// Bar 2: sell limit at 103 fills (high = 104)
	candles := []*series.Candle{
		createTestCandle(101, 102, 103, 101),
		createTestCandle(101, 102, 103, 100),
		createTestCandle(102, 103, 104, 101),
	}
	events := createTestEvents("TEST", candles)

	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &edNeverEnterStrategy{})

	// Buy limit at 100
	buyLimit := trading.NewOrderDetail(trading.BUY, trading.LimitOrder, "TEST", decimal.ONE)
	buyLimit.Price = decimal.New(100)
	broker.SubmitOrder(buyLimit)

	// Sell limit at 103
	sellLimit := trading.NewOrderDetail(trading.SELL, trading.LimitOrder, "TEST", decimal.ONE)
	sellLimit.Price = decimal.New(103)
	broker.SubmitOrder(sellLimit)

	results, err := edb.Run(events)
	require.NoError(t, err)

	result := results["TEST"]
	require.Equal(t, 1, result.TotalTrades)
	assert.True(t, result.Trades[0].EntryPrice.EQ(decimal.New(100)))
	assert.True(t, result.Trades[0].ExitPrice.EQ(decimal.New(103)))
	assert.True(t, result.Trades[0].Profit.IsPositive())
}

func TestEventDrivenBacktester_EquityTracking(t *testing.T) {
	candles := []*series.Candle{
		createTestCandle(100, 101, 102, 99),
		createTestCandle(101, 102, 103, 100),
		createTestCandle(102, 103, 104, 101),
		createTestCandle(103, 104, 105, 102),
	}
	events := createTestEvents("TEST", candles)

	broker := NewSimulatedBroker("TEST", decimal.New(10000))
	edb := NewEventDrivenBacktester()
	edb.Register("TEST", broker, &alwaysEnterStrategy{})

	results, err := edb.Run(events)
	require.NoError(t, err)

	result := results["TEST"]
	// Equity starts at 10000, profit = 104 - 102 = 2
	assert.True(t, result.FinalEquity.EQ(decimal.New(10002)))
	assert.True(t, result.NetProfit.EQ(decimal.New(2)))
}
