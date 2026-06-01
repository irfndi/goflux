package trading

import (
	"strconv"

	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/series"
	"github.com/irfndi/goflux/pkg/telemetry"
)

// --- Bullish Divergence ---

// bullishDivergenceRule is satisfied when price has made a lower low
// over the lookback window while the oscillator has made a higher low.
type bullishDivergenceRule struct {
	price    indicators.Indicator
	osc      indicators.Indicator
	lookback int
}

// NewBullishDivergenceRule returns a rule that detects bullish divergence
// between price and an oscillator over the given lookback window.
// Panics if price or osc is nil or lookback <= 0.
func NewBullishDivergenceRule(price, osc indicators.Indicator, lookback int) Rule {
	if price == nil || osc == nil {
		panic("goflux: BullishDivergenceRule requires non-nil indicators")
	}
	if lookback <= 0 {
		panic("goflux: BullishDivergenceRule lookback must be > 0")
	}
	telemetry.ReportUsage("BullishDivergenceRule", map[string]string{"lookback": strconv.Itoa(lookback)})
	return bullishDivergenceRule{price: price, osc: osc, lookback: lookback}
}

func (r bullishDivergenceRule) IsSatisfied(index int, record *TradingRecord) bool {
	if index < r.lookback {
		return false
	}

	startPrice := r.price.Calculate(index - r.lookback)
	endPrice := r.price.Calculate(index)
	startOsc := r.osc.Calculate(index - r.lookback)
	endOsc := r.osc.Calculate(index)

	// Price lower low + oscillator higher low = bullish divergence
	return endPrice.LT(startPrice) && endOsc.GT(startOsc)
}

// --- Bearish Divergence ---

// bearishDivergenceRule is satisfied when price has made a higher high
// over the lookback window while the oscillator has made a lower high.
type bearishDivergenceRule struct {
	price    indicators.Indicator
	osc      indicators.Indicator
	lookback int
}

// NewBearishDivergenceRule returns a rule that detects bearish divergence
// between price and an oscillator over the given lookback window.
// Panics if price or osc is nil or lookback <= 0.
func NewBearishDivergenceRule(price, osc indicators.Indicator, lookback int) Rule {
	if price == nil || osc == nil {
		panic("goflux: BearishDivergenceRule requires non-nil indicators")
	}
	if lookback <= 0 {
		panic("goflux: BearishDivergenceRule lookback must be > 0")
	}
	telemetry.ReportUsage("BearishDivergenceRule", map[string]string{"lookback": strconv.Itoa(lookback)})
	return bearishDivergenceRule{price: price, osc: osc, lookback: lookback}
}

func (r bearishDivergenceRule) IsSatisfied(index int, record *TradingRecord) bool {
	if index < r.lookback {
		return false
	}

	startPrice := r.price.Calculate(index - r.lookback)
	endPrice := r.price.Calculate(index)
	startOsc := r.osc.Calculate(index - r.lookback)
	endOsc := r.osc.Calculate(index)

	// Price higher high + oscillator lower high = bearish divergence
	return endPrice.GT(startPrice) && endOsc.LT(startOsc)
}

// --- Convenience constructors for RSI ---

// NewRSIBullishDivergenceRule returns a bullish divergence rule using close price
// and RSI constructed from the given time series.
// Panics if s is nil or lookback <= 0.
func NewRSIBullishDivergenceRule(s *series.TimeSeries, lookback int) Rule {
	if s == nil {
		panic("goflux: RSIBullishDivergenceRule series cannot be nil")
	}
	price := indicators.NewClosePriceIndicator(s)
	rsi := indicators.NewRelativeStrengthIndexIndicator(price, 14)
	return NewBullishDivergenceRule(price, rsi, lookback)
}

// NewRSIBearishDivergenceRule returns a bearish divergence rule using close price
// and RSI constructed from the given time series.
// Panics if s is nil or lookback <= 0.
func NewRSIBearishDivergenceRule(s *series.TimeSeries, lookback int) Rule {
	if s == nil {
		panic("goflux: RSIBearishDivergenceRule series cannot be nil")
	}
	price := indicators.NewClosePriceIndicator(s)
	rsi := indicators.NewRelativeStrengthIndexIndicator(price, 14)
	return NewBearishDivergenceRule(price, rsi, lookback)
}

// --- Convenience constructors for MACD ---

// NewMACDBullishDivergenceRule returns a bullish divergence rule using close price
// and MACD constructed from the given time series.
// Panics if s is nil or lookback <= 0.
func NewMACDBullishDivergenceRule(s *series.TimeSeries, lookback int) Rule {
	if s == nil {
		panic("goflux: MACDBullishDivergenceRule series cannot be nil")
	}
	price := indicators.NewClosePriceIndicator(s)
	macd := indicators.NewMACDIndicator(price, 12, 26)
	return NewBullishDivergenceRule(price, macd, lookback)
}

// NewMACDBearishDivergenceRule returns a bearish divergence rule using close price
// and MACD constructed from the given time series.
// Panics if s is nil or lookback <= 0.
func NewMACDBearishDivergenceRule(s *series.TimeSeries, lookback int) Rule {
	if s == nil {
		panic("goflux: MACDBearishDivergenceRule series cannot be nil")
	}
	price := indicators.NewClosePriceIndicator(s)
	macd := indicators.NewMACDIndicator(price, 12, 26)
	return NewBearishDivergenceRule(price, macd, lookback)
}
