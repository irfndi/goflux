package trading_test

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/testutils"
	"github.com/irfndi/goflux/pkg/trading"
	"github.com/stretchr/testify/assert"
)

func TestStopLossRule(t *testing.T) {
	t.Run("Returns false when position is new or closed", func(t *testing.T) {
		record := trading.NewTradingRecord()

		series := testutils.MockTimeSeriesFl(1, 2, 3, 4)

		slr := trading.NewStopLossRule(series, -0.1)

		assert.False(t, slr.IsSatisfied(3, record))
	})

	t.Run("Returns true when losses exceed tolerance", func(t *testing.T) {
		record := trading.NewTradingRecord()
		record.Operate(trading.Order{
			Side:   BUY,
			Amount: decimal.NewFromString("10"),
			Price:  decimal.ONE,
		})

		series := testutils.MockTimeSeriesFl(10, 9) // Lose 10%

		slr := trading.NewStopLossRule(series, -0.05)

		assert.True(t, slr.IsSatisfied(1, record))
	})

	t.Run("Returns false when losses do not exceed tolerance", func(t *testing.T) {
		record := trading.NewTradingRecord()

		record.Operate(trading.Order{
			Side:   BUY,
			Amount: decimal.NewFromString("10"),
			Price:  decimal.ONE,
		})

		series := testutils.MockTimeSeriesFl(10, 10.1) // Gain 1%

		slr := trading.NewStopLossRule(series, -0.05)

		assert.False(t, slr.IsSatisfied(1, record))
	})
}
