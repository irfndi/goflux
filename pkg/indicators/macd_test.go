package indicators_test

import (
	"testing"

	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/stretchr/testify/assert"
)

func TestNewMACDIndicator(t *testing.T) {
	series := randomseries.TimeSeries(100)

	macd := indicators.NewMACDIndicator(indicators.NewClosePriceIndicator(series), 12, 26)

	assert.NotNil(t, macd)
}

func TestNewMACDHistogramIndicator(t *testing.T) {
	series := randomseries.TimeSeries(100)

	macd := indicators.NewMACDIndicator(indicators.NewClosePriceIndicator(series), 12, 26)
	macdHistogram := indicators.NewMACDHistogramIndicator(macd, 9)

	assert.NotNil(t, macdHistogram)
}
