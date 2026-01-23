package indicators_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/testutils"
)

func TestIndicatorBuilder(t *testing.T) {
	ts := testutils.MockTimeSeriesFl(100, 101, 102, 103, 104, 105)

	indicator := indicators.NewIndicatorBuilder(ts).
		SMA(2).
		EMA(2).
		Build()

	assert.NotNil(t, indicator)
	val := indicator.Calculate(5)
	assert.True(t, val.IsPositive())
}
