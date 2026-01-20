# Coding Conventions

**Analysis Date:** 2026-01-20

## Naming Patterns

**Files:**
- `snake_case.go` for implementation files (e.g., `trade_execution_service.go`, `market_hours.go`)
- `*_test.go` for test files (e.g., `adapter_test.go`, `sharpe_test.go`)
- Special suffixes: `_mock.go` for mocks, `_helper.go` for test helpers
- Domain files by concept: `broker_types.go`, `interfaces.go`

**Functions:**
- Exported (public): `PascalCase` (e.g., `CalculateSharpeRatio`, `NewMarketHoursService`, `IsMarketOpen`)
- Unexported (private): `camelCase` (e.g., `scoreDividendYield`, `nullFloat64`, `isHoliday`)
- Constructors: Always `New` prefix (e.g., `NewTradernetBrokerAdapter`, `NewDividendScorer`)

**Variables:**
- Exported: `PascalCase` (e.g., `Version`, `OrderSideBuy`)
- Unexported: `camelCase` (e.g., `dividendYield`, `marketTime`, `lastTradeTime`)
- Acronyms: Use proper casing (e.g., `userID`, `isinCode`, not `userID`, `ISINCode`)

**Types:**
- Exported structs: `PascalCase` (e.g., `TradeExecutionService`, `BrokerPosition`, `MarketHoursService`)
- Unexported structs: `camelCase` (rare, typically for internal helpers)
- Interfaces: `PascalCase` ending in `Interface` or describing capability (e.g., `TradeRepositoryInterface`, `MarketHoursChecker`, `BrokerClient`)

**Constants:**
- Exported: `PascalCase` (e.g., `OrderSideBuy`, `OrderSideSell`)
- Magic numbers replaced with named constants

**Packages:**
- Short, lowercase, single word preferred (e.g., `domain`, `events`, `formulas`)
- Underscores when necessary for clarity: `cash_flows`, `market_hours`, `market_regime`
- Package name matches directory name

## Code Style

**Formatting:**
- Tool: `gofmt` (enforced via pre-commit hook)
- No manual formatting - let gofmt handle it
- Stage fixed: Yes (lefthook auto-formats and stages changes)

**Linting:**
- Tool: `golangci-lint` with custom config (`.golangci.yml`)
- Enabled linters: `errcheck`, `govet`, `gofmt`, `goimports`, `misspell`, `revive`, `gosec`, `bodyclose`, `gocritic`
- Disabled: `gosimple`, `ineffassign`, `staticcheck` (too many style suggestions)
- Timeout: 5 minutes
- Tests: Linted (some linters relaxed for test files)

**Line Length:**
- No hard limit, but keep reasonable (gofmt wraps naturally)

**Brace Style:**
- K&R style (enforced by gofmt)
- Opening brace on same line, closing brace on new line

## Import Organization

**Order (enforced by goimports):**
1. Standard library (e.g., `database/sql`, `fmt`, `time`)
2. Third-party packages (e.g., `github.com/rs/zerolog`, `github.com/stretchr/testify`)
3. Local packages (e.g., `github.com/aristath/sentinel/internal/domain`)

**Path Aliases:**
- Rare, used only when necessary to avoid conflicts
- Example: `planningdomain "github.com/aristath/sentinel/internal/modules/planning/domain"`

**Blank Imports:**
- Acceptable for SQL drivers and side-effects
- Not linted (blank-imports rule disabled)

**Grouping:**
- Automatic via goimports
- No manual grouping needed

## Error Handling

**Patterns:**
- Return errors, never panic (except in main initialization failures)
- Wrap errors with context: `fmt.Errorf("failed to fetch security: %w", err)`
- Use `%w` verb for error wrapping (enables `errors.Is` and `errors.As`)
- Check errors explicitly: `if err != nil { return nil, err }`

**Logging Errors:**
- Use structured logging with zerolog
- Include context fields: `log.Error().Err(err).Str("symbol", symbol).Msg("failed to place order")`
- Never log and return error - choose one approach per layer

**nil Returns:**
- Prefer returning nil pointer for "not found" cases: `func GetByID(id int64) (*Security, error)`
- Return empty slice instead of nil for collections: `return []Trade{}, nil`

**SQL Errors:**
- `sql.ErrNoRows` handled explicitly (not wrapped, checked with `errors.Is`)
- Acceptable to ignore Scan errors for PRAGMA queries

**Deferred Cleanup:**
- Rollback errors in defer typically ignored: `defer tx.Rollback()` (no error check)
- Close errors ignored for io.Closer, Rows, Stmt (excluded in errcheck config)

## Logging

**Framework:** zerolog (structured, high-performance JSON logging)

**Patterns:**
- Always use structured fields: `log.Info().Str("symbol", symbol).Float64("price", price).Msg("order placed")`
- Never use string formatting in messages: Use `.Msgf()` sparingly, prefer fields
- Create child loggers with context: `log := log.With().Str("service", "events").Logger()`

**Log Levels:**
- `Debug()`: Verbose diagnostic information
- `Info()`: Normal operational messages
- `Warn()`: Unexpected but recoverable situations
- `Error()`: Errors that need attention
- `Fatal()`: Unrecoverable errors (exits program)

**Testing:**
- Disable logging in tests: `zerolog.New(nil).Level(zerolog.Disabled)`

## Comments

**When to Comment:**
- Package documentation: Every package has a package-level comment
- Public API: All exported types, functions, constants documented with godoc-style comments
- Complex logic: Non-obvious algorithms explained inline
- Business rules: Why code exists, not what it does

**Style:**
- Godoc format: `// FunctionName does X.` (starts with function name, ends with period)
- Block comments for package docs and complex explanations: `/** ... */`
- Inline comments for clarification: `// Market is closed during lunch break [start, end)`

**JSDoc/TSDoc:**
- Not applicable (Go codebase)
- Use standard Go doc comments

**Examples:**
```go
// CalculateSharpeRatio calculates the Sharpe ratio for a series of returns.
// Returns nil if there is insufficient data or zero volatility.
func CalculateSharpeRatio(returns []float64, riskFreeRate float64, periodsPerYear int) *float64

/**
 * Package domain provides broker-agnostic types for portfolio management.
 *
 * These types abstract away broker-specific implementations (Tradernet, IBKR, etc.).
 */
package domain
```

## Function Design

**Size:**
- Keep functions focused and testable
- No hard line limit, but 100+ lines suggests refactoring opportunity
- Extract helper functions for complex logic

**Parameters:**
- Use pointers for optional values: `func Calculate(yield *float64, payout *float64) Score`
- Prefer structs for 4+ parameters
- Avoid `interface{}` when possible, use concrete types or generics

**Return Values:**
- Return pointers for "optional" or "not found" semantics: `func GetByID(id int64) (*Security, error)`
- Return error as last value: `func DoWork() (Result, error)`
- Multiple named returns for clarity: `func GetMetrics() (mean, stdDev float64, err error)` (rare)

**Receiver Types:**
- Use pointer receivers for mutations: `func (s *Service) Update()`
- Use value receivers for immutable methods on small structs
- Be consistent within a type (prefer pointer receivers by default)

## Module Design

**Exports:**
- Export only what's needed by other packages
- Keep internal helpers unexported
- Use internal/ directory for truly private packages

**Barrel Files:**
- Not a Go pattern (no index.go files)
- Each package is self-contained

**Package Structure:**
- One concept per package (e.g., `events`, `formulas`, `domain`)
- Group related functionality: `internal/modules/scoring/`, `internal/modules/portfolio/`
- Avoid circular dependencies (use interfaces in domain layer)

## Dependency Injection

**Pattern:**
- Constructor injection only
- Create via DI container: `internal/di/` (wire.go, services.go, repositories.go, databases.go)
- Services hold interfaces, not concrete implementations

**Example:**
```go
// Service definition
type TradeExecutionService struct {
    brokerClient domain.BrokerClient  // Interface, not concrete type
    tradeRepo    TradeRepositoryInterface
    log          zerolog.Logger
}

// Constructor
func NewTradeExecutionService(
    brokerClient domain.BrokerClient,
    tradeRepo TradeRepositoryInterface,
    log zerolog.Logger,
) *TradeExecutionService {
    return &TradeExecutionService{
        brokerClient: brokerClient,
        tradeRepo:    tradeRepo,
        log:          log.With().Str("service", "trade_execution").Logger(),
    }
}

// DI container initialization (internal/di/services.go)
func InitializeServices(container *Container) error {
    container.TradeExecutionService = NewTradeExecutionService(
        container.BrokerClient,
        container.TradeRepository,
        container.Log,
    )
    return nil
}
```

## Architecture Patterns

**Clean Architecture:**
- Domain layer is pure (no external dependencies): `internal/domain/`
- Dependency flow: Handlers → Services → Repositories → Domain
- Never import infrastructure from domain

**Repository Pattern:**
- All database access through repository interfaces
- Repositories in module directories: `internal/modules/portfolio/position_repository.go`
- Repository interfaces defined where used or in domain

**Broker Abstraction:**
- All broker interactions through `domain.BrokerClient` interface
- Broker-specific code in `internal/clients/<broker>/adapter.go`
- Transformers map broker types to domain types: `transformers_domain.go`

**Event-Driven:**
- Events published via `events.Bus`
- Subscribers register handlers: `bus.Subscribe(EventTypeTradeExecuted, handler)`
- Events used for cross-module communication

## Special Conventions

**Null Handling (SQL):**
- Use `sql.NullString`, `sql.NullFloat64` for nullable columns
- Helper functions for conversion: `nullFloat64(v float64) sql.NullFloat64`
- Zero values treated as NULL: `if v == 0.0 { return sql.NullFloat64{Valid: false} }`

**Time Handling:**
- Always use `time.Time` for timestamps
- Store as Unix timestamps in database
- Convert to market timezone for market hours checks
- Use UTC for storage: `time.Now().UTC()`

**Currency:**
- All portfolio values normalized to EUR
- Use `CurrencyExchangeService` for conversions
- Store original currency and exchange rate

**Magic Strings:**
- Replace with constants: `OrderSideBuy`, `OrderSideSell` instead of "BUY", "SELL"
- Broker-specific codes documented in transformers

**Compile-Time Interface Checks:**
- Verify interface implementation: `var _ domain.BrokerClient = (*TradernetBrokerAdapter)(nil)`
- Place at end of file

---

*Convention analysis: 2026-01-20*
