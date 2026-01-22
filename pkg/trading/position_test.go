package trading_test

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/trading"
	"github.com/stretchr/testify/assert"
)

func TestPositionNewRule(t *testing.T) {
	t.Run("returns true when position new", func(t *testing.T) {
		record := trading.NewTradingRecord()
		rule := trading.PositionNewRule{}

		assert.True(t, rule.IsSatisfied(0, record))
	})

	t.Run("returns false when position open", func(t *testing.T) {
		record := trading.NewTradingRecord()

		record.Operate(trading.Order{
			Side:   trading.BUY,
			Amount: decimal.ONE,
			Price:  decimal.ONE,
		})

		rule := trading.PositionNewRule{}

		assert.False(t, rule.IsSatisfied(0, record))
	})
}

func TestPositionOpenRule(t *testing.T) {
	t.Run("returns false when position new", func(t *testing.T) {
		record := trading.NewTradingRecord()

		rule := trading.PositionOpenRule{}

		assert.False(t, rule.IsSatisfied(0, record))
	})

	t.Run("returns true when position open", func(t *testing.T) {
		record := trading.NewTradingRecord()

		record.Operate(trading.Order{
			Side:   trading.BUY,
			Amount: decimal.ONE,
			Price:  decimal.ONE,
		})

		rule := trading.PositionOpenRule{}

		assert.True(t, rule.IsSatisfied(0, record))
	})
}
