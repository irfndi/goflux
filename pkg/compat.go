package goflux

import (
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

// Re-export functions
var (
	NewDecimal           = decimal.New
	NewDecimalFromInt    = decimal.NewFromInt
	NewDecimalFromString = decimal.NewFromString

	NewTimeSeries = series.NewTimeSeries
	NewCandle     = series.NewCandle
	NewTimePeriod = series.NewTimePeriod

	NewClosePriceIndicator = indicators.NewClosePriceIndicator
	NewEMAIndicator        = indicators.NewEMAIndicator
	NewSMAIndicator        = indicators.NewSimpleMovingAverage
	NewConstantIndicator   = indicators.NewConstantIndicator

	NewTradingRecord = trading.NewTradingRecord
	NewPosition      = trading.NewPosition

	And = trading.And
	Or  = trading.Or
	Not = trading.Not

	NewCrossUpIndicatorRule   = trading.NewCrossUpIndicatorRule
	NewCrossDownIndicatorRule = trading.NewCrossDownIndicatorRule
)

// PositionNewRule re-export
type PositionNewRule = trading.PositionNewRule
type PositionOpenRule = trading.PositionOpenRule

// Analysis re-exports
type (
	Analysis            = analysis.Analysis
	TotalProfitAnalysis = analysis.TotalProfitAnalysis
)
