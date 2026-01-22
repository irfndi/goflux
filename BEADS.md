# GoFlux Development Tasks

This file tracks all development tasks for the GoFlux library (fork from techan).

## Project Overview

**GoFlux** is a modern technical analysis library for Go, forked from [techan](https://github.com/sdcoffey/techan) by sdcoffey. This project aims to revitalize and expand the library with modern Go best practices, comprehensive testing, and additional technical analysis indicators.

---

## Current Status

In progress:
- Systematic code review and optimization for Go 1.25.6
- Fixing compilation errors in `pkg/decimal`
- Resolving import cycles in tests
- Addressing race conditions in caching and timeseries
- Adding 90%+ test coverage for all files

---

## Phase 1: Foundation & Rebranding (COMPLETED)

- [x] Update module name from `github.com/sdcoffey/techan` to `github.com/irfndi/goflux`
- [x] Rename package from `techan` to `goflux`
- [x] Update README.md with GoFlux branding and fork attribution
- [x] Update LICENSE with new maintainer info
- [x] Create CONTRIBUTING.md
- [x] Update CHANGELOG for GoFlux
- [x] Remove old CI/CD configs (travis) and set up GitHub Actions

---

## Phase 2: Modern Project Structure (COMPLETED)

- [x] Consolidate code in pkg/ structure
- [x] Create pkg/analysis, pkg/indicators, pkg/series, pkg/trading, pkg/math
- [x] Move tests to appropriate directories
- [x] Create compatibility layer in pkg/goflux
- [x] Clean up legacy directories

---

## Phase 3: Dependency Updates (IN PROGRESS)

- [x] Update Go version to 1.25.6
- [x] Update go.mod with new module path
- [x] Add golangci-lint configuration
- [ ] Remove old dependencies (github.com/sdcoffey/big)
- [ ] Add standard test dependencies if needed

---

## Phase 4: Systematic Code Review & Optimization (ACTIVE)

### Priority Order: Foundation -> Core -> Advanced

#### 4.1 Foundation Layer
- [ ] **pkg/decimal/decimal.go**
  - [ ] **FIX: Correct `math/big` usage in `Round`, `Floor`, `Ceil`, `Truncate`**
  - [ ] **ADD: `Frac()` method for indicator calculations**
  - [ ] Review arithmetic operations for overflow
  - [ ] Add comprehensive test coverage (90%+)
  - [ ] Check for race conditions
  - [ ] Optimize performance (minimize allocations)
  
- [ ] **pkg/math/math.go**
  - [ ] Review mathematical functions
  - [ ] Add comprehensive test coverage (90%+)
  - [ ] Check for edge cases
  - [ ] Use modern Go 1.25.6 features

#### 4.2 Core Data Structures
- [ ] **pkg/series/timeperiod.go**
  - [ ] Review time handling logic
  - [ ] Add comprehensive test coverage (90%+)
  - [ ] Optimize performance

- [ ] **pkg/series/candle.go**
  - [ ] **FIX: Remove duplicate imports**
  - [ ] **FIX: Logic for initializing MinPrice/MaxPrice in `AddTrade`**
  - [ ] Add comprehensive test coverage (90%+)
  - [ ] Optimize memory usage (consider value types over pointers)

- [ ] **pkg/series/timeseries.go**
  - [ ] **FIX: Add thread-safety (RWMutex) to `AddCandle` and accessors**
  - [ ] **OPTIMIZE: Use `[]Candle` instead of `[]*Candle` to reduce GC pressure**
  - [ ] Add comprehensive test coverage (90%+)

#### 4.3 Analysis Layer
- [ ] **pkg/analysis/analysis.go**
  - [ ] Review analysis logic
  - [ ] Add comprehensive test coverage (90%+)
  - [ ] Optimize performance

- [ ] **pkg/indicators/cached_indicator.go**
  - [ ] **FIX: Add thread-safety (Mutex) to cache operations**
  - [ ] **OPTIMIZE: Use a more memory-efficient cache structure (avoid slice of pointers)**
  - [ ] **FIX: Potential memory leak in unbounded cache growth**
  - [ ] Add comprehensive test coverage (90%+)

- [ ] **pkg/indicators/indicator.go**
  - [ ] Review indicator interface
  - [ ] Add comprehensive test coverage (90%+)
  - [ ] Ensure consistent API

#### 4.4 Basic Indicators
- [ ] **pkg/indicators/basic.go** - Close, Open, High, Low, Volume
- [ ] **pkg/indicators/constant.go** - Constant values
- [ ] **pkg/indicators/fixed.go** - Fixed values

#### 4.5 Average Indicators
- [ ] **pkg/indicators/average.go** - Average calculation
- [ ] **pkg/indicators/simple_moving_average.go** - SMA
- [ ] **pkg/indicators/exponential_moving_average.go** - EMA
- [ ] **pkg/indicators/modified_moving_average.go** - MMA
- [ ] **pkg/indicators/gains.go** - Gain/Loss calculation

#### 4.6 Volatility Indicators
- [ ] **pkg/indicators/true_range.go** - True Range
- [ ] **pkg/indicators/average_true_range.go** - ATR
- [ ] **pkg/indicators/standard_deviation.go** - Std Dev
- [ ] **pkg/indicators/mean_deviation.go** - Mean Deviation
- [ ] **pkg/indicators/variance.go** - Variance
- [ ] **pkg/indicators/windowed_standard_deviation.go** - Windowed Std Dev

#### 4.7 Oscillator Indicators
- [ ] **pkg/indicators/stochastic_oscillator.go** - Stochastic
- [ ] **pkg/indicators/relative_strength.go** - RSI
- [ ] **pkg/indicators/relative_vigor_index.go** - RVI
- [ ] **pkg/indicators/macd.go** - MACD

#### 4.8 Channel Indicators
- [ ] **pkg/indicators/bollinger_band.go** - Bollinger Bands
- [ ] **pkg/indicators/keltner_channel.go** - Keltner Channel
- [ ] **pkg/indicators/cci.go** - CCI
- [ ] **pkg/indicators/aroon.go** - Aroon

#### 4.9 Helper Indicators
- [ ] **pkg/indicators/difference.go** - Difference between indicators
- [ ] **pkg/indicators/derivative.go** - Rate of change
- [ ] **pkg/indicators/trend.go** - Trend detection
- [ ] **pkg/indicators/maximum_value.go** - Max value
- [ ] **pkg/indicators/minimum_value.go** - Min value
- [ ] **pkg/indicators/maximum_drawdown.go** - Max drawdown

#### 4.10 Trading Layer
- [ ] **pkg/trading/rule.go** - Rule interface and base
- [ ] **pkg/trading/cross.go** - Cross-over rules
- [ ] **pkg/trading/increase_decrease.go** - Increase/decrease rules
- [ ] **pkg/trading/stop.go** - Stop-loss rules
- [ ] **pkg/trading/strategy.go** - Strategy interface
- [ ] **pkg/trading/order.go** - Order management
- [ ] **pkg/trading/position.go** - Position tracking
- [ ] **pkg/trading/tradingrecord.go** - Trading history

#### 4.11 Compatibility Layer
- [ ] **pkg/goflux/** - Re-export all types
- [ ] Ensure backward compatibility

---

## Phase 5: Test Coverage & Quality Assurance (ACTIVE)

### 5.1 Test Coverage Goals
- [ ] Achieve 90%+ coverage for all files
- [ ] **FIX: Resolve import cycles between `pkg/indicators` and `pkg/testutils`**
- [ ] Use table-driven tests where appropriate
- [ ] Add property-based tests where applicable
- [ ] Add benchmarks for performance-critical code
- [ ] Run tests with race detection: `go test -race ./...`

### 5.2 Code Quality
- [ ] Run golangci-lint: `golangci-lint run`
- [ ] Fix all linting issues
- [ ] Use go vet: `go vet ./...`
- [ ] Use go fmt: `go fmt ./...`

---

## Phase 6: New Technical Indicators (FUTURE)

### Trend
- [ ] Ichimoku Cloud
- [ ] Parabolic SAR
- [ ] ADX (Average Directional Index)
- [ ] TEMA, TMA, VMA, HMA, KAMA (Moving averages)
- [ ] ZigZag Indicator
- [ ] SuperTrend

### Momentum
- [ ] Stochastic RSI
- [ ] Williams %R
- [ ] MFI (Money Flow Index)
- [ ] Chaikin Oscillator
- [ ] ROC (Rate of Change)
- [ ] Ultimate Oscillator
- [ ] Awesome Oscillator
- [ ] Fisher Transform

### Volatility
- [ ] Donchian Channel
- [ ] Standard Error Bands
- [ ] Historical Volatility
- [ ] Chaikin Volatility

### Volume
- [ ] OBV (On Balance Volume)
- [ ] A/D Line (Accumulation/Distribution)
- [ ] CMF (Chaikin Money Flow)
- [ ] VWAP (Volume Weighted Average Price)
- [ ] Klinger Oscillator

### Candlestick Patterns
- [ ] Doji detection
- [ ] Hammer/Hanging Man
- [ ] Engulfing patterns
- [ ] Morning/Evening Star

---

## Phase 7: Enhanced Trading System (FUTURE)

- [ ] Position sizing algorithms (Fixed Fractional, Fixed Ratio, Kelly Criterion)
- [ ] Risk management (Max drawdown limits, circuit breakers)
- [ ] Limit/stop-limit orders
- [ ] Trailing stops (Fixed, ATR-based)
- [ ] Slippage and Commission modeling for backtesting
- [ ] Monte Carlo simulation for strategy robustness

---

## Phase 8: Data Management & Visualization (FUTURE)

- [ ] CSV/JSON data loaders
- [ ] Database integration (InfluxDB, TimescaleDB)
- [ ] Time series resampling (e.g., 1m to 5m, 1h)
- [ ] **NEW: Heikin Ashi candle generation**
- [ ] **NEW: Renko chart generation**
- [ ] Export results to plotting libraries (e.g., gonum/plot)

---

## Phase 9: CI/CD & Automation (COMPLETED)

- [x] GitHub Actions CI workflow
- [x] GitHub Actions release workflow
- [x] Issue templates
- [x] PR template
- [ ] Dependabot configuration
- [ ] Automated CHANGELOG

---

## Priority Tasks (Current Sprint)

1. **Fix `pkg/decimal/decimal.go` compilation errors and add `Frac()` method.**
2. **Resolve import cycle between `pkg/indicators` and `pkg/testutils`.**
3. **Add thread-safety to `TimeSeries` and `cachedIndicator`.**
4. **Remove duplicate imports and fix `MinPrice` logic in `Candle`.**
5. **Migrate all indicators from `sdcoffey/big` to `pkg/decimal`.**

---

## Resources

- [Original techan](https://github.com/sdcoffey/techan)
- [ta4j](https://github.com/ta4j/ta4j)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
