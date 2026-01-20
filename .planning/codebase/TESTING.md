# Testing Patterns

**Analysis Date:** 2026-01-20

## Test Framework

**Runner:**
- Go's built-in testing framework (`go test`)
- Version: Go 1.24.0
- Config: None required (uses Go conventions)

**Assertion Library:**
- `github.com/stretchr/testify/assert` (assertions)
- `github.com/stretchr/testify/require` (fatal assertions)

**Run Commands:**
```bash
go test ./...                          # Run all tests
go test -v ./...                       # Verbose output
go test -run TestSpecificName ./...    # Run specific test
go test -short ./...                   # Skip long-running tests
go test -timeout=10m ./...             # Set timeout
go test -json ./...                    # JSON output
go test -cover ./...                   # Coverage report
go test -coverprofile=coverage.out ./...  # Coverage file
```

## Test File Organization

**Location:**
- Co-located with source files (same directory)
- Pattern: `<filename>_test.go` for `<filename>.go`
- Examples:
  - `internal/clients/tradernet/adapter.go` → `adapter_test.go`
  - `pkg/formulas/sharpe.go` → `sharpe_test.go`
  - `internal/modules/portfolio/position_repository.go` → `position_repository_helper_test.go`

**Naming:**
- Test files: `*_test.go` (e.g., `adapter_test.go`)
- Test functions: `Test<FunctionName>_<Scenario>` (e.g., `TestCalculateSharpeRatio_InsufficientData`)
- Helper tests: `*_helper_test.go` for testing unexported helper functions
- Integration tests: `*_integration_test.go` (e.g., `optimization_integration_test.go`)
- Smoke tests: `*_smoke_test.go` (e.g., `adapter_smoke_test.go`)

**Structure:**
```
internal/modules/scoring/
├── scorers/
│   ├── dividend.go              # Implementation
│   ├── dividend_test.go         # Unit tests
│   ├── utils.go                 # Helpers
│   └── utils_test.go            # Helper tests
```

## Test Structure

**Suite Organization:**
```go
// Table-driven tests (preferred pattern)
func TestCalculateSharpeRatio(t *testing.T) {
    tests := []struct {
        name           string
        returns        []float64
        riskFreeRate   float64
        periodsPerYear int
        expectedNil    bool
        description    string
    }{
        {
            name:           "insufficient data",
            returns:        []float64{0.01},
            riskFreeRate:   0.02,
            periodsPerYear: 252,
            expectedNil:    true,
            description:    "Need at least 2 returns",
        },
        {
            name:           "valid daily returns",
            returns:        []float64{0.01, -0.005, 0.02, -0.01, 0.015},
            riskFreeRate:   0.02,
            periodsPerYear: 252,
            expectedNil:    false,
            description:    "Valid returns should calculate Sharpe",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateSharpeRatio(tt.returns, tt.riskFreeRate, tt.periodsPerYear)
            if (result == nil) != tt.expectedNil {
                t.Errorf("CalculateSharpeRatio() = %v, expected nil: %v - %s",
                    result, tt.expectedNil, tt.description)
                return
            }
            if result != nil {
                // Additional validations
                if math.IsNaN(*result) || math.IsInf(*result, 0) {
                    t.Errorf("Sharpe ratio is NaN or Inf: %v", *result)
                }
            }
        })
    }
}

// Subtests for different scenarios
func TestTradernetBrokerAdapter_GetPortfolio(t *testing.T) {
    log := zerolog.New(nil).Level(zerolog.Disabled)

    t.Run("success", func(t *testing.T) {
        mockSDK := &mockSDKClient{
            accountSummaryResult: map[string]interface{}{
                "result": map[string]interface{}{
                    "ps": map[string]interface{}{
                        "pos": []interface{}{
                            map[string]interface{}{
                                "i":            "AAPL",
                                "q":            10.0,
                                "bal_price_a":  150.0,
                            },
                        },
                    },
                },
            },
        }

        client := NewClientWithSDK(mockSDK, log)
        adapter := &TradernetBrokerAdapter{client: client}

        positions, err := adapter.GetPortfolio()
        require.NoError(t, err)
        assert.Len(t, positions, 1)
        assert.Equal(t, "AAPL", positions[0].Symbol)
    })

    t.Run("sdk error", func(t *testing.T) {
        mockSDK := &mockSDKClient{
            accountSummaryError: errors.New("sdk error"),
        }

        client := NewClientWithSDK(mockSDK, log)
        adapter := &TradernetBrokerAdapter{client: client}

        positions, err := adapter.GetPortfolio()
        assert.Error(t, err)
        assert.Nil(t, positions)
    })
}
```

**Patterns:**
- Use `t.Run()` for subtests (enables parallel execution and better output)
- Table-driven tests for multiple scenarios
- Descriptive test names: `Test<Function>_<Scenario>` or `Test<Function>/<subtest_name>`
- Group related tests with section comments: `// ============================================================================`

## Mocking

**Framework:** Manual mocks (no mocking library)

**Patterns:**
```go
// Mock structure with configurable behavior
type mockSDKClient struct {
    // Results to return
    accountSummaryResult map[string]interface{}
    buyResult            map[string]interface{}
    sellResult           map[string]interface{}

    // Errors to return
    accountSummaryError error
    buyError            error
    sellError           error

    // Capture arguments for verification
    lastLimitPrice float64
}

// Implement interface method
func (m *mockSDKClient) AccountSummary() (map[string]interface{}, error) {
    if m.accountSummaryError != nil {
        return nil, m.accountSummaryError
    }
    return m.accountSummaryResult, nil
}

// Usage in tests
mockSDK := &mockSDKClient{
    accountSummaryResult: map[string]interface{}{
        "result": map[string]interface{}{"data": "value"},
    },
}
```

**Mock Repository Pattern:**
```go
// MockPositionRepository in internal/testing/mocks.go
type MockPositionRepository struct {
    mu        sync.RWMutex
    positions []portfolio.Position
    err       error
}

func NewMockPositionRepository() *MockPositionRepository {
    return &MockPositionRepository{
        positions: make([]portfolio.Position, 0),
    }
}

// Setters for test setup
func (m *MockPositionRepository) SetPositions(positions []portfolio.Position) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.positions = positions
}

func (m *MockPositionRepository) SetError(err error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.err = err
}

// Interface implementation
func (m *MockPositionRepository) GetAll() ([]portfolio.Position, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    if m.err != nil {
        return nil, m.err
    }
    return m.positions, nil
}
```

**What to Mock:**
- External APIs (broker clients, HTTP clients)
- Database repositories (use mock implementations)
- Time-dependent behavior (use fixed timestamps)
- File system operations
- Network operations

**What NOT to Mock:**
- Pure functions (formulas, calculations)
- Simple data transformations
- Domain logic (test with real objects)
- Standard library (time.Time is fine to use)

## Fixtures and Factories

**Test Data:**
```go
// Inline test data (preferred for simple cases)
func TestNullFloat64(t *testing.T) {
    tests := []struct {
        name     string
        input    float64
        expected sql.NullFloat64
    }{
        {
            name:     "zero value",
            input:    0.0,
            expected: sql.NullFloat64{Valid: false},
        },
        {
            name:     "positive value",
            input:    3.14,
            expected: sql.NullFloat64{Float64: 3.14, Valid: true},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := nullFloat64(tt.input)
            assert.Equal(t, tt.expected.Valid, result.Valid)
            if tt.expected.Valid {
                assert.InDelta(t, tt.expected.Float64, result.Float64, 0.0001)
            }
        })
    }
}

// Helper functions for complex fixtures
func makeReturns(value float64, count int) []float64 {
    returns := make([]float64, count)
    for i := range returns {
        returns[i] = value
    }
    return returns
}
```

**Location:**
- Inline in test files (most common)
- Test-specific helpers: Unexported functions in `*_test.go` files
- Shared mocks: `internal/testing/mocks.go`

**Database Fixtures:**
- Use in-memory SQLite for integration tests
- Create schema with migrations
- Populate with minimal realistic data
- Clean up after tests (defer cleanup)

## Coverage

**Requirements:** No enforced minimum (but aim for high coverage on critical paths)

**View Coverage:**
```bash
go test -cover ./...                              # Quick coverage summary
go test -coverprofile=coverage.out ./...          # Generate coverage file
go tool cover -html=coverage.out                  # View in browser
go tool cover -func=coverage.out                  # Function-level coverage
```

**Coverage Exclusions:**
- Test files automatically excluded
- Main package often has no tests (entry point only)
- Generated code (e.g., `internal/version/version.go`)

## Test Types

**Unit Tests:**
- Scope: Single function or method
- No external dependencies (DB, network, filesystem)
- Use mocks for dependencies
- Fast execution (milliseconds)
- Examples: `pkg/formulas/sharpe_test.go`, `internal/modules/scoring/scorers/dividend_test.go`

**Integration Tests:**
- Scope: Multiple components working together
- May use real database (in-memory SQLite)
- Test repository interactions, service coordination
- Slower execution (seconds)
- File naming: `*_integration_test.go`
- Examples: `internal/market_regime/index_service_integration_test.go`, `internal/modules/optimization/optimization_integration_test.go`

**Smoke Tests:**
- Scope: Basic "does it work at all" validation
- Quick sanity checks
- File naming: `*_smoke_test.go`
- Example: `internal/clients/tradernet/adapter_smoke_test.go`

**End-to-End Tests:**
- Not used (no E2E framework)
- System testing done manually or via deployment tests

## Common Patterns

**Async Testing:**
```go
// Not common in this codebase (synchronous tests preferred)
// When needed, use channels and timeouts
func TestAsyncOperation(t *testing.T) {
    done := make(chan bool, 1)

    go func() {
        // Async work
        done <- true
    }()

    select {
    case <-done:
        // Success
    case <-time.After(5 * time.Second):
        t.Fatal("timeout waiting for async operation")
    }
}
```

**Error Testing:**
```go
// Test error cases explicitly
t.Run("sdk error", func(t *testing.T) {
    mockSDK := &mockSDKClient{
        buyError: errors.New("sdk error"),
    }

    client := NewClientWithSDK(mockSDK, log)
    adapter := &TradernetBrokerAdapter{client: client}

    result, err := adapter.PlaceOrder("AAPL", "BUY", 5.0, 155.0)
    assert.Error(t, err)
    assert.Nil(t, result)
})

// Test nil returns for "not found"
positions, err := adapter.GetPortfolio()
require.NoError(t, err)
assert.Nil(t, positions)
```

**Floating Point Comparisons:**
```go
// Use InDelta for floating point assertions
assert.InDelta(t, expected, actual, 0.0001)  // Within 0.0001
assert.InDelta(t, 3.14159, result, 0.001)    // ~3 decimal places

// For exact equality (when appropriate)
assert.Equal(t, 1.0, result)
```

**Boundary Testing:**
```go
// Test exact boundaries
func TestDividendScorer_YieldThresholds(t *testing.T) {
    testCases := []struct {
        name        string
        yield       float64
        expectedMin float64
        expectedMax float64
    }{
        {"Zero yield", 0.00, 0.30, 0.30},
        {"Low threshold (1%)", 0.01, 0.40, 0.45},
        {"Mid threshold (3%)", 0.03, 0.70, 0.75},
        {"High threshold (6%)", 0.06, 1.00, 1.00},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            score := scoreDividendYield(&tc.yield)
            assert.GreaterOrEqual(t, score, tc.expectedMin)
            assert.LessOrEqual(t, score, tc.expectedMax)
        })
    }
}
```

**Testing with Disabled Logging:**
```go
// Standard pattern for test setup
func TestSomething(t *testing.T) {
    log := zerolog.New(nil).Level(zerolog.Disabled)

    service := NewService(log)
    // Test service...
}
```

**Compile-Time Interface Checks:**
```go
// Verify mock implements interface
var _ domain.BrokerClient = (*TradernetBrokerAdapter)(nil)
```

## Test Execution in CI/CD

**Pre-commit Hooks (Lefthook):**
- Tests NOT run on pre-commit (too slow)
- Only formatting, linting, and building
- To enable: Uncomment `pre-push` section in `lefthook.yml`

**Manual Test Runs:**
```bash
go test -timeout=10m -short ./...    # Quick tests only
go test -timeout=30m ./...           # All tests
```

**Test Timeout:**
- Default: 10 minutes
- Set via `-timeout` flag
- Individual tests should complete in seconds

## Testing Utilities

**Location:** `internal/testing/`

**Available Utilities:**
- `MockPositionRepository` - Mock for position repository
- `NewMockPositionRepository()` - Factory for position mocks
- Database test helpers (if needed)

**Usage:**
```go
import (
    "testing"
    sentineltesting "github.com/aristath/sentinel/internal/testing"
)

func TestWithMockRepo(t *testing.T) {
    mockRepo := sentineltesting.NewMockPositionRepository()
    mockRepo.SetPositions([]portfolio.Position{
        {Symbol: "AAPL", Quantity: 10, AvgPrice: 150},
    })

    // Use mockRepo in test...
}
```

## Best Practices

**DO:**
- Write table-driven tests for multiple scenarios
- Use `t.Run()` for subtests
- Test edge cases and boundary conditions
- Test error paths explicitly
- Disable logging in tests (`zerolog.Disabled`)
- Use `require.NoError()` when error should fail test immediately
- Use `assert.*()` for non-fatal assertions
- Name tests descriptively: `TestFunction_Scenario`

**DON'T:**
- Skip writing tests for critical business logic
- Use `t.Fatal()` in subtests (breaks parallel execution)
- Test implementation details (test behavior, not internals)
- Create complex test fixtures (keep tests simple and readable)
- Use sleeps for synchronization (use channels or mocks)
- Decrease coverage (aim to maintain or increase)

---

*Testing analysis: 2026-01-20*
