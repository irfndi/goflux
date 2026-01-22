package indicators_test

import (
	"testing"
	"time"

	"github.com/irfndi/goflux/pkg/series"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/stretchr/testify/assert"
)

func TestNewVolumeIndicator(t *testing.T) {
	assert.NotNil(t, indicators.NewVolumeIndicator(series.NewTimeSeries()))
}

func TestVolumeIndicator_Calculate(t *testing.T) {
	ts := series.NewTimeSeries()

	period := series.NewTimePeriod(time.Now(), time.Minute)
	candle := series.NewCandle(period)
	candle.Volume = decimal.NewFromString("1.2080")

	ts.AddCandle(candle)

	indicator := indicators.NewVolumeIndicator(ts)
	assert.EqualValues(t, "1.208", indicator.Calculate(0).FormattedString(3))
}

func TestTypicalPriceIndicator_Calculate(t *testing.T) {
	ts := series.NewTimeSeries()

	period := series.NewTimePeriod(time.Now(), time.Minute)
	candle := series.NewCandle(period)
	candle.MinPrice = decimal.NewFromString("1.2080")
	candle.MaxPrice = decimal.NewFromString("1.22")
	candle.ClosePrice = decimal.NewFromString("1.215")

	ts.AddCandle(candle)

	typicalPrice := indicators.NewTypicalPriceIndicator(ts).Calculate(0)

	assert.EqualValues(t, "1.2143", typicalPrice.FormattedString(4))
}
