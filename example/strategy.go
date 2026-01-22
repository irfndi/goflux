package example

import goflux "github.com/irfndi/goflux/pkg"

// StrategyExample shows how to create a simple trading strategy. In this example, a position should
// be opened if the price moves above 70, and the position should be closed if a position moves below 30.
func StrategyExample() {
	indicator := BasicEma() // from basic.go

	// record trades on this object
	record := goflux.NewTradingRecord()

	entryConstant := goflux.NewConstantIndicator(30)
	exitConstant := goflux.NewConstantIndicator(10)

	entryRule := goflux.And(
		goflux.NewCrossUpIndicatorRule(entryConstant, indicator),
		goflux.PositionNewRule{}) // Is satisfied when the price ema moves above 30 and the current position is new

	exitRule := goflux.And(
		goflux.NewCrossDownIndicatorRule(indicator, exitConstant),
		goflux.PositionOpenRule{}) // Is satisfied when the price ema moves below 10 and the current position is open

	strategy := goflux.RuleStrategy{
		UnstablePeriod: 10,
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}

	strategy.ShouldEnter(0, record) // returns false
}
