package example

import (
	"strconv"
	"time"

	goflux "github.com/irfndi/goflux/pkg"
	"github.com/irfndi/goflux/pkg/decimal"
)

// BasicEma is an example of how to create a basic Exponential moving average indicator
// based on close prices of a timeseries from your exchange of choice.
func BasicEma() goflux.Indicator {
	series := goflux.NewTimeSeries()

	// fetch this from your preferred exchange
	dataset := [][]string{
		// Timestamp, Open, Close, High, Low, volume
		{"1234567", "1", "2", "3", "5", "6"},
	}

	for _, datum := range dataset {
		start, _ := strconv.ParseInt(datum[0], 10, 64)
		period := goflux.NewTimePeriod(time.Unix(start, 0), time.Hour*24)

		candle := goflux.NewCandle(period)
		candle.OpenPrice = decimal.NewFromString(datum[1])
		candle.ClosePrice = decimal.NewFromString(datum[2])
		candle.MaxPrice = decimal.NewFromString(datum[3])
		candle.MinPrice = decimal.NewFromString(datum[4])

		series.AddCandle(candle)
	}

	closePrices := goflux.NewClosePriceIndicator(series)
	movingAverage := goflux.NewEMAIndicator(closePrices, 10)

	return movingAverage
}
