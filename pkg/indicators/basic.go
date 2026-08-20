package indicators

import (
	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
)

type volumeIndicator struct {
	series *series.TimeSeries
}

// NewVolumeIndicator returns an indicator which returns the volume of a candle for a given index
func NewVolumeIndicator(series *series.TimeSeries) Indicator {
	return volumeIndicator{series: series}
}

func (vi volumeIndicator) Calculate(index int) decimal.Decimal {
	candle := candleAt(vi.series, index)
	if candle == nil {
		return decimal.ZERO
	}
	return candle.Volume
}

type closePriceIndicator struct {
	series *series.TimeSeries
}

// NewClosePriceIndicator returns an Indicator which returns the close price of a candle for a given index
func NewClosePriceIndicator(series *series.TimeSeries) Indicator {
	return closePriceIndicator{series: series}
}

func (cpi closePriceIndicator) Calculate(index int) decimal.Decimal {
	candle := candleAt(cpi.series, index)
	if candle == nil {
		return decimal.ZERO
	}
	return candle.ClosePrice
}

type highPriceIndicator struct {
	series *series.TimeSeries
}

// NewHighPriceIndicator returns an Indicator which returns the high price of a candle for a given index
func NewHighPriceIndicator(series *series.TimeSeries) Indicator {
	return highPriceIndicator{
		series: series,
	}
}

func (hpi highPriceIndicator) Calculate(index int) decimal.Decimal {
	candle := candleAt(hpi.series, index)
	if candle == nil {
		return decimal.ZERO
	}
	return candle.MaxPrice
}

type lowPriceIndicator struct {
	series *series.TimeSeries
}

// NewLowPriceIndicator returns an Indicator which returns the low price of a candle for a given index
func NewLowPriceIndicator(series *series.TimeSeries) Indicator {
	return lowPriceIndicator{
		series: series,
	}
}

func (lpi lowPriceIndicator) Calculate(index int) decimal.Decimal {
	candle := candleAt(lpi.series, index)
	if candle == nil {
		return decimal.ZERO
	}
	return candle.MinPrice
}

type openPriceIndicator struct {
	series *series.TimeSeries
}

// NewOpenPriceIndicator returns an Indicator which returns the open price of a candle for a given index
func NewOpenPriceIndicator(series *series.TimeSeries) Indicator {
	return openPriceIndicator{
		series: series,
	}
}

func (opi openPriceIndicator) Calculate(index int) decimal.Decimal {
	candle := candleAt(opi.series, index)
	if candle == nil {
		return decimal.ZERO
	}
	return candle.OpenPrice
}

type typicalPriceIndicator struct {
	series *series.TimeSeries
}

// NewTypicalPriceIndicator returns an Indicator which returns the typical price of a candle for a given index.
// The typical price is an average of the high, low, and close prices for a given candle.
func NewTypicalPriceIndicator(series *series.TimeSeries) Indicator {
	return typicalPriceIndicator{series: series}
}

func (tpi typicalPriceIndicator) Calculate(index int) decimal.Decimal {
	candle := candleAt(tpi.series, index)
	if candle == nil {
		return decimal.ZERO
	}
	numerator := candle.MaxPrice.Add(candle.MinPrice).Add(candle.ClosePrice)
	return numerator.Div(decimal.NewFromString("3"))
}

type averagePriceIndicator struct {
	series *series.TimeSeries
}

func NewAveragePriceIndicator(series *series.TimeSeries) Indicator {
	return averagePriceIndicator{series: series}
}

func (api averagePriceIndicator) Calculate(index int) decimal.Decimal {
	candle := candleAt(api.series, index)
	if candle == nil {
		return decimal.ZERO
	}
	return candle.OpenPrice.Add(candle.MaxPrice).Add(candle.MinPrice).Add(candle.ClosePrice).Div(decimal.New(4))
}

type medianPriceIndicator struct {
	series *series.TimeSeries
}

func NewMedianPriceIndicator(series *series.TimeSeries) Indicator {
	return medianPriceIndicator{series: series}
}

func (mpi medianPriceIndicator) Calculate(index int) decimal.Decimal {
	candle := candleAt(mpi.series, index)
	if candle == nil {
		return decimal.ZERO
	}
	return candle.MaxPrice.Add(candle.MinPrice).Div(decimal.New(2))
}

type weightedCloseIndicator struct {
	series *series.TimeSeries
}

func NewWeightedCloseIndicator(series *series.TimeSeries) Indicator {
	return weightedCloseIndicator{series: series}
}

func (wci weightedCloseIndicator) Calculate(index int) decimal.Decimal {
	candle := candleAt(wci.series, index)
	if candle == nil {
		return decimal.ZERO
	}
	return candle.MaxPrice.Add(candle.MinPrice).Add(candle.ClosePrice).Add(candle.ClosePrice).Div(decimal.New(4))
}

func candleAt(s *series.TimeSeries, index int) *series.Candle {
	if s == nil {
		return nil
	}
	return s.GetCandle(index)
}
