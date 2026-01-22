package trading_test

import (
	"testing"

	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func TestIncreaseRule(t *testing.T) {
	t.Run("returns false when index == 0", func(t *testing.T) {
		rule := IncreaseRule{}

		assert.False(t, rule.IsSatisfied(0, nil))
	})

	t.Run("returns true when increase", func(t *testing.T) {
		series := testutils.MockTimeSeries("1", "2")
		rule := IncreaseRule{indicators.NewClosePriceIndicator(series)}

		assert.True(t, rule.IsSatisfied(1, nil))
	})

	t.Run("returns false when same", func(t *testing.T) {
		series := testutils.MockTimeSeries("1", "1")
		rule := IncreaseRule{indicators.NewClosePriceIndicator(series)}

		assert.False(t, rule.IsSatisfied(1, nil))
	})

	t.Run("returns false when decrease", func(t *testing.T) {
		series := testutils.MockTimeSeries("1", "0")
		rule := IncreaseRule{indicators.NewClosePriceIndicator(series)}

		assert.False(t, rule.IsSatisfied(1, nil))
	})
}

func TestDecreaseRule(t *testing.T) {
	t.Run("returns false when index == 0", func(t *testing.T) {
		rule := DecreaseRule{}

		assert.False(t, rule.IsSatisfied(0, nil))
	})

	t.Run("returns true when decrease", func(t *testing.T) {
		series := testutils.MockTimeSeries("1", "0")
		rule := DecreaseRule{indicators.NewClosePriceIndicator(series)}

		assert.True(t, rule.IsSatisfied(1, nil))
	})

	t.Run("returns false when  decrease", func(t *testing.T) {
		series := testutils.MockTimeSeries("1", "2")
		rule := DecreaseRule{indicators.NewClosePriceIndicator(series)}

		assert.False(t, rule.IsSatisfied(1, nil))
	})

	t.Run("returns false when same", func(t *testing.T) {
		series := testutils.MockTimeSeries("1", "1")
		rule := IncreaseRule{indicators.NewClosePriceIndicator(series)}

		assert.False(t, rule.IsSatisfied(1, nil))
	})
}
