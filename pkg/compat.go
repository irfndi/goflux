package goflux

import (
	"time"

	"github.com/irfndi/goflux/pkg/analysis"
	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/trading"
)

// Re-export types
type (
	Decimal       = decimal.Decimal
	Indicator     = indicators.Indicator
	TimeSeries    = series.TimeSeries
	Candle        = series.Candle
	TimePeriod    = series.TimePeriod
	TradingRecord = trading.TradingRecord
	Order         = trading.Order
	Position      = trading.Position
	Rule          = trading.Rule
	Strategy      = trading.Strategy
	RuleStrategy  = trading.RuleStrategy
	Side          = trading.OrderSide
)

// Re-export constants
const (
	BUY  = trading.BUY
	SELL = trading.SELL
)

var (
	ZERO = decimal.ZERO
	ONE  = decimal.ONE
)

// Re-export decimal functions
func NewDecimal(v float64) decimal.Decimal          { return decimal.New(v) }
func NewDecimalFromInt(v int64) decimal.Decimal     { return decimal.NewFromInt(v) }
func NewDecimalFromString(v string) decimal.Decimal { return decimal.NewFromString(v) }
func NewDecimalFromStringWithError(v string) (decimal.Decimal, error) {
	return decimal.NewFromStringWithError(v)
}

// Re-export series functions
func NewTimeSeries() *series.TimeSeries                 { return series.NewTimeSeries() }
func NewCandle(period series.TimePeriod) *series.Candle { return series.NewCandle(period) }
func NewTimePeriod(start time.Time, duration time.Duration) series.TimePeriod {
	return series.NewTimePeriod(start, duration)
}

// Re-export indicator functions
func NewClosePriceIndicator(s *series.TimeSeries) indicators.Indicator {
	return indicators.NewClosePriceIndicator(s)
}
func NewAveragePriceIndicator(s *series.TimeSeries) indicators.Indicator {
	return indicators.NewAveragePriceIndicator(s)
}
func NewMedianPriceIndicator(s *series.TimeSeries) indicators.Indicator {
	return indicators.NewMedianPriceIndicator(s)
}
func NewWeightedCloseIndicator(s *series.TimeSeries) indicators.Indicator {
	return indicators.NewWeightedCloseIndicator(s)
}
func NewBollingerBandwidthIndicator(ind indicators.Indicator, window int, k float64) indicators.Indicator {
	return indicators.NewBollingerBandwidthIndicator(ind, window, k)
}
func NewATRRatioIndicator(atr, price indicators.Indicator) indicators.Indicator {
	return indicators.NewATRRatioIndicator(atr, price)
}
func NewATRRatioIndicatorFromSeries(s *series.TimeSeries, atrWindow int) indicators.Indicator {
	return indicators.NewATRRatioIndicatorFromSeries(s, atrWindow)
}
func NewEMAIndicator(ind indicators.Indicator, window int) indicators.Indicator {
	return indicators.NewEMAIndicator(ind, window)
}
func NewSMAIndicator(ind indicators.Indicator, window int) indicators.Indicator {
	return indicators.NewSimpleMovingAverage(ind, window)
}
func NewTRIMAIndicator(ind indicators.Indicator, window int) indicators.Indicator {
	return indicators.NewTRIMAIndicator(ind, window)
}
func NewRMAIndicator(ind indicators.Indicator, window int) indicators.Indicator {
	return indicators.NewRMAIndicator(ind, window)
}
func NewT3Indicator(ind indicators.Indicator, window int, vFactor float64) indicators.Indicator {
	return indicators.NewT3Indicator(ind, window, vFactor)
}
func NewALMAIndicator(ind indicators.Indicator, window int, offset float64, sigma float64) indicators.Indicator {
	return indicators.NewALMAIndicator(ind, window, offset, sigma)
}
func NewVIDYAIndicator(ind indicators.Indicator, window int) indicators.Indicator {
	return indicators.NewVIDYAIndicator(ind, window)
}
func NewMAMAIndicator(ind indicators.Indicator, fastLimit, slowLimit float64) indicators.Indicator {
	return indicators.NewMAMAIndicator(ind, fastLimit, slowLimit)
}
func NewFAMAIndicator(ind indicators.Indicator, fastLimit, slowLimit float64) indicators.Indicator {
	return indicators.NewFAMAIndicator(ind, fastLimit, slowLimit)
}
func NewTimeSeriesIndicator(s *series.TimeSeries) indicators.Indicator {
	return indicators.NewTimeSeriesIndicator(s)
}
func NewVWMAIndicator(ind, volume indicators.Indicator, window int) indicators.Indicator {
	return indicators.NewVWMAIndicator(ind, volume, window)
}
func NewVWMAIndicatorFromSeries(s *series.TimeSeries, window int) indicators.Indicator {
	return indicators.NewVWMAIndicatorFromSeries(s, window)
}
func NewConstantIndicator(v float64) indicators.Indicator { return indicators.NewConstantIndicator(v) }
func Max(a, b int) int                                    { return indicators.Max(a, b) }
func Min(a, b int) int                                    { return indicators.Min(a, b) }

// Re-export trading functions
func NewTradingRecord() *trading.TradingRecord               { return trading.NewTradingRecord() }
func NewPosition(entryOrder trading.Order) *trading.Position { return trading.NewPosition(entryOrder) }

func And(r1, r2 trading.Rule) trading.Rule { return trading.And(r1, r2) }
func Or(r1, r2 trading.Rule) trading.Rule  { return trading.Or(r1, r2) }
func Not(rule trading.Rule) trading.Rule   { return trading.Not(rule) }
func NewCrossUpIndicatorRule(upper, lower indicators.Indicator) trading.Rule {
	return trading.NewCrossUpIndicatorRule(upper, lower)
}
func NewCrossDownIndicatorRule(upper, lower indicators.Indicator) trading.Rule {
	return trading.NewCrossDownIndicatorRule(upper, lower)
}

// PositionNewRule re-export
type PositionNewRule = trading.PositionNewRule
type PositionOpenRule = trading.PositionOpenRule

// Analysis re-exports
type (
	Analysis            = analysis.Analysis
	TotalProfitAnalysis = analysis.TotalProfitAnalysis
)
