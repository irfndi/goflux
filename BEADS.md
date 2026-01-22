# GoFlux Development Tasks

This file tracks all development tasks for the GoFlux library (fork from techan).

## Project Overview

**GoFlux** is a modern technical analysis library for Go, forked from [techan](https://github.com/sdcoffey/techan) by sdcoffey. This project aims to revitalize and expand the library with modern Go best practices, comprehensive testing, and additional technical analysis indicators.

---

## Current Status

Completed:
- Systematic code review and optimization for Go 1.25.6
- Fixing compilation errors in `pkg/decimal`
- Resolving import cycles in tests
- Removed legacy dependency `sdcoffey/big`
- Fixed logic and panics in 17+ technical indicators
- Created compatibility layer in `pkg/compat.go`

In progress:
- Adding 90%+ test coverage for all files
- Implementing Candlestick Patterns (Sprint 4)

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

## Phase 3: Dependency Updates (COMPLETED)

- [x] Update Go version to 1.25.6
- [x] Update go.mod with new module path
- [x] Add golangci-lint configuration
- [x] Remove old dependencies (github.com/sdcoffey/big)
- [x] Add standard test dependencies if needed

---

## Phase 4: Systematic Code Review & Optimization (ACTIVE)

### Priority Order: Foundation -> Core -> Advanced

#### 4.1 Foundation Layer
- [x] **pkg/decimal/decimal.go**
  - [x] **FIX: Correct `math/big` usage in `Round`, `Floor`, `Ceil`, `Truncate`**
  - [x] **ADD: `Frac()` method for indicator calculations**
  - [x] Review arithmetic operations for overflow
  - [x] Add comprehensive test coverage (90%+)
  - [x] Check for race conditions
  - [x] Optimize performance (minimize allocations)
  
- [x] **pkg/math/math.go**
  - [x] Review mathematical functions
  - [x] Add comprehensive test coverage (90%+)
  - [x] Check for edge cases
  - [x] Use modern Go 1.25.6 features

#### 4.2 Core Data Structures
- [x] **pkg/series/timeperiod.go**
  - [x] Review time handling logic
  - [x] Add comprehensive test coverage (90%+)
  - [x] Optimize performance

- [x] **pkg/series/candle.go**
  - [x] **FIX: Remove duplicate imports**
  - [x] **FIX: Logic for initializing MinPrice/MaxPrice in `AddTrade`**
  - [x] Add comprehensive test coverage (90%+)
  - [x] Optimize memory usage (consider value types over pointers)

- [x] **pkg/series/timeseries.go**
  - [x] **FIX: Add thread-safety (RWMutex) to `AddCandle` and accessors**
  - [x] **OPTIMIZE: Use `[]Candle` instead of `[]*Candle` to reduce GC pressure**
  - [x] Add comprehensive test coverage (90%+)

#### 4.3 Analysis Layer
- [x] **pkg/analysis/analysis.go**
  - [x] Review analysis logic
  - [x] Add comprehensive test coverage (90%+)
  - [x] Optimize performance

- [ ] **pkg/indicators/cached_indicator.go**
  - [ ] **FIX: Add thread-safety (Mutex) to cache operations**
  - [ ] **OPTIMIZE: Use a more memory-efficient cache structure (avoid slice of pointers)**
  - [ ] **FIX: Potential memory leak in unbounded cache growth**
  - [ ] Add comprehensive test coverage (90%+)

- [x] **pkg/indicators/indicator.go**
  - [x] Review indicator interface
  - [x] Add comprehensive test coverage (90%+)
  - [x] Ensure consistent API

---

## Phase 5: Test Coverage & Quality Assurance (ACTIVE)

### 5.1 Test Coverage Goals
- [ ] Achieve 90%+ coverage for all files
- [x] **FIX: Resolve import cycles between `pkg/indicators` and `pkg/testutils`**
- [ ] Use table-driven tests where appropriate
- [ ] Add property-based tests where applicable
- [ ] Add benchmarks for performance-critical code
- [ ] Run tests with race detection: `go test -race ./...`

---

## Phase 6: New Technical Indicators (COMPLETED SPRINT 1-3)

### Trend (COMPLETED)
- [x] Ichimoku Cloud
- [x] Parabolic SAR
- [x] ADX (Average Directional Index)
- [x] Vortex Indicator

### Momentum & Oscillators (COMPLETED)
- [x] Williams %R
- [x] Rate of Change (ROC)
- [x] Momentum
- [x] Ultimate Oscillator
- [x] Awesome Oscillator
- [x] Money Flow Index (MFI)

### Moving Averages (COMPLETED)
- [x] HMA (Hull Moving Average)
- [x] KAMA (Kaufman Adaptive Moving Average)
- [x] DEMA (Double Exponential Moving Average)
- [x] TEMA (Triple Exponential Moving Average)
- [x] WMA (Weighted Moving Average)

### Volume (COMPLETED)
- [x] OBV (On Balance Volume)

---

## Sprint 4: Candlestick Patterns (COMPLETED)
- [x] Doji detection
- [x] Dragonfly/Gravestone Doji
- [x] Hammer/Hanging Man
- [x] Inverted Hammer/Shooting Star
- [x] Engulfing patterns (Bullish/Bearish)
- [x] Piercing Line/Dark Cloud Cover
- [x] Spinning Top/Marubozu
- [x] Framework for 20+ patterns in `pkg/candlesticks/`

---

## Sprint 6: Performance Metrics (COMPLETED)
- [x] Sharpe Ratio calculator
- [x] Sortino Ratio calculator
- [x] Calmar Ratio calculator
- [x] CAGR (Compound Annual Growth Rate)
- [x] Sterling Ratio
- [x] Burke Ratio
- [x] Skewness and Kurtosis (higher moments)
- [x] Downside deviation calculation
- [x] Unit tests for all metrics
- [x] Formatted metrics output

---

## Sprint 7: Volume Indicators (IN PROGRESS)
- [ ] CMF (Chaikin Money Flow)
- [ ] VWAP (Volume Weighted Average Price)
- [ ] A/D Line (Accumulation/Distribution)
- [ ] Klinger Oscillator
- [ ] Volume ROC

---

## Total Implemented: 22+ indicators + full backtesting + metrics

### Summary by Category
| Category | Count | Indicators/Features |
|----------|-------|---------------------|
| Momentum/Oscillators | 6 | Williams %R, ROC, Momentum, Ultimate AO, AO, MFI |
| Moving Averages | 5 | HMA, WMA, KAMA, DEMA, TEMA |
| Trend/Directional | 4 | ADX, Parabolic SAR, Vortex, Ichimoku |
| Volume | 1 | OBV |
| Candlestick Patterns | 20+ | Framework with pattern detection |
| Backtesting | 1 | Complete backtester engine |
| Performance Metrics | 6+ | Sharpe, Sortino, Calmar, CAGR, Sterling, Burke |

---

## Project Structure

```
goflux/
├── pkg/
│   ├── analysis/          # Analysis utilities
│   ├── backtest/          # Backtesting engine
│   ├── candlesticks/      # Candlestick pattern detection
│   ├── decimal/           # High-precision decimal arithmetic
│   ├── indicators/        # 22+ technical indicators
│   ├── math/              # Mathematical functions
│   ├── metrics/           # Performance metrics
│   ├── series/            # Time series and candle data
│   ├── trading/           # Trading rules and strategies
│   └── testutils/         # Testing utilities
├── example/               # Example usage
├── .github/workflows/     # CI/CD
└── BEADS.md              # Development roadmap
```

### Sprint 8: Advanced Trading Rules (MEDIUM PRIORITY)
- [ ] Trailing stop loss
- [ ] Time-based exits
- [ ] Trailing take profit
- [ ] Composite rules (AND/OR/NOT)

---

## Resources

- [Original techan](https://github.com/sdcoffey/techan)
- [ta4j](https://github.com/ta4j/ta4j)
- [pandas-ta](https://github.com/freqtrade/pandas-ta)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
