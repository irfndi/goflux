package trading

import (
	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/series"
)

// donchianBreakoutUpperRule is satisfied when the close price breaks
// above the previous period's Donchian Channel upper band.
type donchianBreakoutUpperRule struct {
	closePrice indicators.Indicator
	upperBand  indicators.Indicator
	window     int
}

// NewDonchianBreakoutUpperRule returns a rule that triggers when close
// price exceeds the previous period's Donchian upper band.
func NewDonchianBreakoutUpperRule(s *series.TimeSeries, window int) Rule {
	return donchianBreakoutUpperRule{
		closePrice: indicators.NewClosePriceIndicator(s),
		upperBand:  indicators.NewDonchianUpperBandIndicator(s, window),
		window:     window,
	}
}

func (r donchianBreakoutUpperRule) IsSatisfied(index int, record *TradingRecord) bool {
	if index < r.window {
		return false
	}
	return r.closePrice.Calculate(index).GT(r.upperBand.Calculate(index - 1))
}

// donchianBreakoutLowerRule is satisfied when the close price breaks
// below the previous period's Donchian Channel lower band.
type donchianBreakoutLowerRule struct {
	closePrice indicators.Indicator
	lowerBand  indicators.Indicator
	window     int
}

// NewDonchianBreakoutLowerRule returns a rule that triggers when close
// price falls below the previous period's Donchian lower band.
func NewDonchianBreakoutLowerRule(s *series.TimeSeries, window int) Rule {
	return donchianBreakoutLowerRule{
		closePrice: indicators.NewClosePriceIndicator(s),
		lowerBand:  indicators.NewDonchianLowerBandIndicator(s, window),
		window:     window,
	}
}

func (r donchianBreakoutLowerRule) IsSatisfied(index int, record *TradingRecord) bool {
	if index < r.window {
		return false
	}
	return r.closePrice.Calculate(index).LT(r.lowerBand.Calculate(index - 1))
}

// donchianChannelWidthRule is satisfied when the channel width
// (upper - lower) exceeds a given threshold, indicating volatility.
type donchianChannelWidthRule struct {
	upper     indicators.Indicator
	lower     indicators.Indicator
	threshold decimal.Decimal
}

// NewDonchianChannelWidthRule returns a rule that triggers when the
// Donchian Channel width exceeds the given threshold.
func NewDonchianChannelWidthRule(s *series.TimeSeries, window int, threshold float64) Rule {
	return donchianChannelWidthRule{
		upper:     indicators.NewDonchianUpperBandIndicator(s, window),
		lower:     indicators.NewDonchianLowerBandIndicator(s, window),
		threshold: decimal.New(threshold),
	}
}

func (r donchianChannelWidthRule) IsSatisfied(index int, record *TradingRecord) bool {
	width := r.upper.Calculate(index).Sub(r.lower.Calculate(index))
	return width.GT(r.threshold)
}
