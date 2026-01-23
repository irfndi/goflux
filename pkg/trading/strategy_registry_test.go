package trading_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/irfndi/goflux/pkg/testutils"
	"github.com/irfndi/goflux/pkg/trading"
)

func TestStrategyRegistry(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(100, 101, 102, 103, 104, 105)
	registry := trading.NewStrategyRegistry()

	t.Run("List strategies", func(t *testing.T) {
		list := registry.List()
		assert.Contains(t, list, "sma_cross_fast")
		assert.Contains(t, list, "ema_cross_fast")
		assert.Contains(t, list, "rsi_overbought_oversold")
	})

	t.Run("Instantiate sma_cross_fast", func(t *testing.T) {
		params := map[string]interface{}{
			"fast_period": 2,
			"slow_period": 5,
		}
		strategy, err := registry.Instantiate("sma_cross_fast", ts, params)
		assert.NoError(t, err)
		assert.NotNil(t, strategy)
	})

	t.Run("Serialization", func(t *testing.T) {
		params := map[string]interface{}{
			"fast_period": 2,
			"slow_period": 5,
		}
		data, err := trading.SerializeStrategy("sma_cross_fast", params)
		assert.NoError(t, err)

		strategy, err := trading.DeserializeStrategy(data, ts)
		assert.NoError(t, err)
		assert.NotNil(t, strategy)
	})
}

func TestVoteRule(t *testing.T) {
	record := trading.NewTradingRecord()

	r1 := trading.PositionNewRule{}
	r2 := trading.PositionNewRule{}
	r3 := trading.PositionOpenRule{}

	vote := trading.Vote(2, r1, r2, r3)

	// Since r1 and r2 are satisfied (Position is New), count = 2 >= threshold 2
	assert.True(t, vote.IsSatisfied(0, record))

	vote3 := trading.Vote(3, r1, r2, r3)
	assert.False(t, vote3.IsSatisfied(0, record))
}
