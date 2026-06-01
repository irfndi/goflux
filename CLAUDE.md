# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## RTK (Rust Token Killer)

Always prefix commands with `rtk` for 60-90% token savings. Even in chains:

```bash
# Correct
rtk git status && rtk git diff && rtk go test ./...
```

See the full RTK reference at the bottom of this file.

## Common Commands

```bash
# Run all tests
rtk go test ./...
# Or with verbose output
rtk go test -v ./...

# Run a single test
rtk go test -run TestName ./pkg/indicators

# Format code (goimports + go fmt)
make fmt

# Run linter (golangci-lint)
make lint

# Run all checks (fmt, test, lint)
make check

# Run benchmarks
make bench

# Run tests with coverage
make test-with-coverage

# Build examples
make build-examples
```

## Architecture

GoFlux is a Go technical analysis library (forked from techan). Module: `github.com/irfndi/goflux`, Go 1.21.

### Package Layout

```
pkg/
  decimal/        # High-precision arithmetic wrapping big.Float
  series/         # TimeSeries (thread-safe candle array), Candle, TimePeriod
  indicators/     # All TA indicators implementing Indicator interface
  trading/        # Rules, strategies, position/record management
  backtest/       # Backtesting engine
  analysis/       # Performance metrics (Sharpe, drawdown, etc.)
  candlesticks/   # Pattern detection
  metrics/        # Risk/performance metrics
  math/           # Math utilities
  telemetry/      # Opt-in usage reporting
example/          # Example applications
```

### Core Patterns

**Indicator Interface** (`pkg/indicators/indicator.go`):
```go
type Indicator interface {
    Calculate(int) decimal.Decimal
}
```

**Cached Indicator** (`pkg/indicators/cached_indicator.go`):
- Indicators cache results via `resultCache` to avoid recalculation.
- Use `returnIfCached()` and `cacheResult()` in `Calculate()`.
- Thread-safe with `sync.RWMutex`.

**Constructor Pattern**:
- Constructors validate inputs (nil checks, window > 0) and panic with `goflux: <message>`.
- Call `telemetry.ReportUsage("Name", params)` in constructors (no-op unless user opts in).
- Example:
```go
func NewEMAIndicator(indicator Indicator, window int) Indicator {
    if indicator == nil { panic("goflux: EMA indicator cannot be nil") }
    if window <= 0 { panic("goflux: EMA window must be > 0") }
    telemetry.ReportUsage("EMA", map[string]string{"window": strconv.Itoa(window)})
    return &emaIndicator{...}
}
```

**Decimal Type** (`pkg/decimal`):
- Wraps `math/big.Float`. Use `decimal.New(val)` for float64, `decimal.NewFromString(s)` for strings.
- Compare with `.LT()`, `.GT()`, `.EQ()`, `.IsZero()`.
- `decimal.ZERO` is the zero constant.

**TimeSeries** (`pkg/series/timeseries.go`):
- Thread-safe candle storage. `GetCandle(index)` returns the candle at index.
- `MockTimeSeriesFl(v1, v2, ...)` creates candles with Close=v, Max=v+1, Min=v-1, Volume=v.
- `MockTimeSeriesOCHL([O,C,H,L], ...)` creates candles with exact O,C,H,L and Volume=index.

### Testing

- Tests live in `*_test.go` files, usually in the same package or `_test` suffix package.
- Use `testutils.MockTimeSeriesFl(...)` and `testutils.MockTimeSeriesOCHL(...)` for fixtures.
- Test constructor panics with `assert.Panics(t, func() { ... })`.
- Test insufficient data by asserting `.IsZero()` on early indices.
- Benchmarks use `benchmarkIndicatorConstruction()` in `benchmarks_test.go`.

## Development Workflow

### Issue Tracking (Beads)

This project uses Beads for issue tracking:
```bash
bd ready          # See available work
bd show <id>      # View issue details
bd sync           # Sync issues with git
```

### Git Workflow

1. Create feature branch from `main`
2. Implement with tests, run `make check`
3. Commit with Conventional Commits (`feat:`, `fix:`, `test:`, etc.)
4. Push and create PR
5. Address ALL review bot comments (Kilo, CodeRabbit, Gemini, Codex) until none remain
6. **CodSpeed benchmark check is known flaky** — do not let it block merging
7. Squash merge when approved, delete branch

### Telemetry

- Telemetry is **opt-in only**. Never enable by default.
- `telemetry.ReportUsage()` in constructors is a no-op when disabled.
- The example app shows the opt-in pattern (commented out).

## RTK Full Reference

### Build & Compile
```bash
rtk cargo build         # Cargo build output
rtk cargo check         # Cargo check output
rtk cargo clippy        # Clippy warnings grouped by file (80%)
rtk tsc                 # TypeScript errors grouped by file/code (83%)
rtk lint                # ESLint/Biome violations grouped (84%)
rtk prettier --check    # Files needing format only (70%)
rtk next build          # Next.js build with route metrics (87%)
```

### Test
```bash
rtk cargo test          # Cargo test failures only (90%)
rtk go test             # Go test failures only (90%)
rtk jest                # Jest failures only (99.5%)
rtk vitest              # Vitest failures only (99.5%)
rtk playwright test     # Playwright failures only (94%)
rtk pytest              # Python test failures only (90%)
rtk rspec               # RSpec failures only (60%)
rtk test <cmd>          # Generic test wrapper - failures only
```

### Git
```bash
rtk git status          # Compact status
rtk git log             # Compact log (works with all git flags)
rtk git diff            # Compact diff (80%)
rtk git show            # Compact show (80%)
rtk git add             # Ultra-compact confirmations (59%)
rtk git commit          # Ultra-compact confirmations (59%)
rtk git push            # Ultra-compact confirmations
rtk git pull            # Ultra-compact confirmations
rtk git branch          # Compact branch list
rtk git fetch           # Compact fetch
rtk git stash           # Compact stash
rtk git worktree        # Compact worktree
```

### GitHub
```bash
rtk gh pr view <num>    # Compact PR view (87%)
rtk gh pr checks        # Compact PR checks (79%)
rtk gh run list         # Compact workflow runs (82%)
rtk gh issue list       # Compact issue list (80%)
rtk gh api              # Compact API responses (26%)
```

### JavaScript/TypeScript Tooling
```bash
rtk pnpm list           # Compact dependency tree (70%)
rtk pnpm outdated       # Compact outdated packages (80%)
rtk pnpm install        # Compact install output (90%)
rtk npm run <script>    # Compact npm script output
rtk npx <cmd>           # Compact npx command output
rtk prisma              # Prisma without ASCII art (88%)
```

### Files & Search
```bash
rtk ls <path>           # Tree format, compact (65%)
rtk read <file>         # Code reading with filtering (60%)
rtk grep <pattern>      # Search grouped by file (75%). Format flags (-c, -l, -L, -o, -Z) run raw.
rtk find <pattern>      # Find grouped by directory (70%)
```

### Analysis & Debug
```bash
rtk err <cmd>           # Filter errors only from any command
rtk log <file>          # Deduplicated logs with counts
rtk json <file>         # JSON structure without values
rtk deps                # Dependency overview
rtk env                 # Environment variables compact
rtk summary <cmd>       # Smart summary of command output
rtk diff                # Ultra-compact diffs
```

### Infrastructure
```bash
rtk docker ps           # Compact container list
rtk docker images       # Compact image list
rtk docker logs <c>     # Deduplicated logs
rtk kubectl get         # Compact resource list
rtk kubectl logs        # Deduplicated pod logs
```

### Network
```bash
rtk curl <url>          # Compact HTTP responses (70%)
rtk wget <url>          # Compact download output (65%)
```

### Meta Commands
```bash
rtk gain                # View token savings statistics
rtk gain --history      # View command history with savings
rtk discover            # Analyze Claude Code sessions for missed RTK usage
rtk proxy <cmd>         # Run command without filtering (for debugging)
rtk init                # Add RTK instructions to CLAUDE.md
rtk init --global       # Add RTK to ~/.claude/CLAUDE.md
```

| Category | Commands | Typical Savings |
|----------|----------|-----------------|
| Tests | vitest, playwright, cargo test | 90-99% |
| Build | next, tsc, lint, prettier | 70-87% |
| Git | status, log, diff, add, commit | 59-80% |
| GitHub | gh pr, gh run, gh issue | 26-87% |
| Package Managers | pnpm, npm, npx | 70-90% |
| Files | ls, read, grep, find | 60-75% |
| Infrastructure | docker, kubectl | 85% |
| Network | curl, wget | 65-70% |

Overall average: **60-90% token reduction** on common development operations.
