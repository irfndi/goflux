package indicators_test

import (
	"testing"
	"time"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/stretchr/testify/assert"
)

func TestNewVolumeIndicator(t *testing.T) {
	assert.NotNil(t, indicators.NewVolumeIndicator(Newseries.TimeSeries()))
}

func TestVolumeIndicator_Calculate(t *testing.T) {
	series := Newseries.TimeSeries()

	candle := Newseries.Candle(series.TimePeriod{
		Start: time.Now(),
		End:   time.Now().Add(time.Minute),
	})
	candle.Volume = decimal.NewFromString("1.2080")

	series.Addseries.Candle(candle)

	indicator := indicators.NewVolumeIndicator(series)
	assert.EqualValues(t, "1.208", indicator.Calculate(0).FormattedString(3))
}

func TestTypicalPriceIndicator_Calculate(t *testing.T) {
	series := Newseries.TimeSeries()

	candle := Newseries.Candle(series.TimePeriod{
		Start: time.Now(),
		End:   time.Now().Add(time.Minute),
	})
	candle.MinPrice = decimal.NewFromString("1.2080")
	candle.MaxPrice = decimal.NewFromString("1.22")
	candle.ClosePrice = decimal.NewFromString("1.215")

	series.Addseries.Candle(candle)

	typicalPrice := indicators.NewTypicalPriceIndicator(series).Calculate(0)

	assert.EqualValues(t, "1.2143", typicalPrice.FormattedString(4))
}
