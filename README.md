# GoFlux

[![Go Reference](https://pkg.go.dev/badge/github.com/irfndi/goflux/pkg.svg)](https://pkg.go.dev/github.com/irfndi/goflux/pkg)
[![Go Report Card](https://goreportcard.com/badge/github.com/irfndi/goflux)](https://goreportcard.com/report/github.com/irfndi/goflux)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**GoFlux** is a modern technical analysis library for Go, forked from [techan](https://github.com/sdcoffey/techan) by sdcoffey. This project aims to revitalize and expand the library with modern Go best practices, comprehensive testing, and additional technical analysis indicators.

## Features 

- 35+ technical analysis indicators (trend, momentum, volume, moving averages)
- Performance metrics (Sharpe, Sortino, Calmar, CAGR, drawdown, and more)
- Candlestick pattern detection (20+ patterns)
- Rule-based strategy engine (AND/OR/NOT, trailing stops, time-based exits)
- Backtesting engine with trade & equity analytics
- Time series utilities (resampling, Heikin Ashi, Renko)

### Installation

```sh
$ go get github.com/irfndi/goflux@latest
```

### Quickstart

```go
package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/irfndi/goflux/pkg"
)

func main() {
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
		candle.OpenPrice = goflux.NewDecimalFromString(datum[1])
		candle.ClosePrice = goflux.NewDecimalFromString(datum[2])
		candle.MaxPrice = goflux.NewDecimalFromString(datum[3])
		candle.MinPrice = goflux.NewDecimalFromString(datum[4])
		candle.Volume = goflux.NewDecimalFromString(datum[5])

		series.AddCandle(candle)
	}

	closePrices := goflux.NewClosePriceIndicator(series)
	movingAverage := goflux.NewEMAIndicator(closePrices, 10)

	fmt.Println(movingAverage.Calculate(0).FormattedString(2))
}
```

### Creating trading strategies

```go
indicator := goflux.NewClosePriceIndicator(series)

// record trades on this object
record := goflux.NewTradingRecord()

entryConstant := goflux.NewConstantIndicator(30)
exitConstant := goflux.NewConstantIndicator(10)

// Is satisfied when the price ema moves above 30 and the current position is new
entryRule := goflux.And(
	goflux.NewCrossUpIndicatorRule(entryConstant, indicator),
	goflux.PositionNewRule{})
	
// Is satisfied when the price ema moves below 10 and the current position is open
exitRule := goflux.And(
	goflux.NewCrossDownIndicatorRule(indicator, exitConstant),
	goflux.PositionOpenRule{})

strategy := goflux.RuleStrategy{
	UnstablePeriod: 10, // Period before which ShouldEnter and ShouldExit always return false
	EntryRule:      entryRule,
	ExitRule:       exitRule,
}

strategy.ShouldEnter(0, record) // returns false
```

## Roadmap

See [BEADS.md](BEADS.md) for a detailed roadmap of planned improvements including:

- Modern project structure
- Additional indicators and utilities
- Expanded trading and risk-management rules
- Comprehensive testing suite
- Improved documentation
- CI/CD pipeline with GitHub Actions

## Migration from Techan

GoFlux maintains backward compatibility with the original techan API. Simply update your imports:

```go
// Old
import "github.com/sdcoffey/techan"

// New
import "github.com/irfndi/goflux/pkg"
```

The package name is `goflux`, so you can use it directly:

```go
import "github.com/irfndi/goflux/pkg"

// Usage
series := goflux.NewTimeSeries()
```

## Contributing

Contributions are welcome! Please see [BEADS.md](BEADS.md) for planned improvements and areas where help is needed.

To contribute:

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

## Acknowledgments

GoFlux is heavily influenced by the great [ta4j](https://github.com/ta4j/ta4j). Many of the ideas and frameworks in this library owe their genesis to the great work done over there.

Special thanks to **sdcoffey** for creating the original [techan](https://github.com/sdcoffey/techan) library, which serves as the foundation for this project.

## License

GoFlux is released under the MIT license. See [LICENSE](./LICENSE) for details.
