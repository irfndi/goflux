package example

import (
	"strconv"
	"time"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/series"
)

// BasicEma is an example of how to create a basic Exponential moving average indicator
// based on close prices of a timeseries from your exchange of choice.
func BasicEma() indicators.Indicator {
	// Optional: enable anonymized telemetry to help improve the library
	// telemetry.Enable("https://goflux-telemetry.irfndi.workers.dev/v1/telemetry", "")

	ts := series.NewTimeSeries()

	// fetch this from your preferred exchange
	// Each row: [Timestamp, Open, Close, High, Low, Volume]
	dataset := [][]string{
		{"1234567", "1", "2", "3", "1", "100"},
		{"1234568", "2", "3", "4", "2", "110"},
		{"1234569", "3", "2", "4", "2", "90"},
		{"1234570", "2", "3", "4", "2", "105"},
		{"1234571", "3", "4", "5", "3", "120"},
		{"1234572", "4", "5", "6", "4", "130"},
		{"1234573", "5", "4", "6", "4", "115"},
		{"1234574", "4", "5", "6", "4", "125"},
		{"1234575", "5", "6", "7", "5", "140"},
		{"1234576", "6", "7", "8", "6", "150"},
	}

	for _, datum := range dataset {
		start, _ := strconv.ParseInt(datum[0], 10, 64)
		period := series.NewTimePeriod(time.Unix(start, 0), time.Hour*24)

		candle := series.NewCandle(period)
		candle.OpenPrice = decimal.NewFromString(datum[1])
		candle.ClosePrice = decimal.NewFromString(datum[2])
		candle.MaxPrice = decimal.NewFromString(datum[3])
		candle.MinPrice = decimal.NewFromString(datum[4])
		candle.Volume = decimal.NewFromString(datum[5])

		ts.AddCandle(candle)
	}

	closePrices := indicators.NewClosePriceIndicator(ts)
	movingAverage := indicators.NewEMAIndicator(closePrices, 10)

	return movingAverage
}
