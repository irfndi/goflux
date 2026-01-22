package indicators_test

import (
	"testing"

	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func TestNewMACDIndicator(t *testing.T) {
	ts := testutils.RandomTimeSeries(100)

	macd := indicators.NewMACDIndicator(indicators.NewClosePriceIndicator(ts), 12, 26)

	assert.NotNil(t, macd)
}

func TestNewMACDHistogramIndicator(t *testing.T) {
	ts := testutils.RandomTimeSeries(100)

	macd := indicators.NewMACDIndicator(indicators.NewClosePriceIndicator(ts), 12, 26)
	macdHistogram := indicators.NewMACDHistogramIndicator(macd, 9)

	assert.NotNil(t, macdHistogram)
}
