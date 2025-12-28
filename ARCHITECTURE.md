# Arduino Trader Architecture

## Overview

The project follows a **Clean Architecture** pattern with clear separation between domain logic, infrastructure, and application layers.

## Architecture Layers

```
app/
├── domain/              # Pure business logic (no dependencies on infrastructure)
│   ├── models.py        # Domain entities (Stock, Position, Trade, etc.)
│   ├── analytics/       # Portfolio analytics (attribution, metrics, reconstruction)
│   ├── events/          # Domain events (position, stock, trade, recommendation)
│   ├── factories/       # Domain factories (stock, trade, recommendation)
│   ├── planning/        # Rebalancing strategies (holistic planner, opportunities)
│   ├── repositories/   # Repository interfaces (protocols)
│   ├── responses/       # Response models (calculation, list, score, service)
│   ├── scoring/        # Stock scoring system (8 groups, calculations, scorers)
│   ├── services/        # Domain services (PriorityCalculator, SettingsService, etc.)
│   ├── utils/           # Shared utilities
│   ├── value_objects/   # Value objects (Currency, Money, Price, TradeSide, etc.)
│   └── exceptions.py    # Domain-specific exceptions
│
├── infrastructure/      # External concerns
│   ├── database/        # Database manager, schemas, queue
│   ├── external/        # API clients (Tradernet, Yahoo Finance)
│   ├── hardware/        # LED display controller
│   ├── cache.py         # Caching layer
│   ├── events.py        # Event bus implementation
│   ├── dependencies.py # FastAPI dependency injection
│   └── ...              # Other infrastructure concerns (locking, rate limiting, etc.)
│
├── application/         # Application services (orchestration)
│   ├── services/        # Use cases
│   │   ├── optimization/    # Portfolio optimizer (Mean-Variance, HRP)
│   │   ├── recommendation/ # Recommendation services
│   │   ├── trade_execution/ # Trade execution services
│   │   ├── portfolio_service.py
│   │   ├── rebalancing_service.py
│   │   ├── scoring_service.py
│   │   └── trade_execution_service.py
│   └── dto/            # Data transfer objects
│
├── repositories/        # Repository implementations (SQLite)
│   ├── stock.py
│   ├── position.py
│   ├── trade.py
│   ├── portfolio.py
│   └── ...
│
├── api/                 # FastAPI routes (thin controllers)
│   ├── portfolio.py
│   ├── stocks.py
│   ├── trades.py
│   ├── allocation.py
│   ├── cash_flows.py
│   ├── charts.py
│   ├── optimizer.py
│   ├── recommendations.py
│   ├── settings.py
│   └── status.py
│
├── services/            # Legacy services (backward compatibility)
│   ├── allocator.py     # Position sizing, priority calculation
│   ├── tradernet.py     # Tradernet API client (legacy, use infrastructure/external)
│   └── yahoo.py         # Yahoo Finance integration (legacy, use infrastructure/external)
│
└── jobs/               # Background jobs (APScheduler)
    ├── scheduler.py
    ├── daily_sync.py
    ├── cash_rebalance.py
    └── ...
```

## Key Principles

### 1. Domain Layer (Pure Business Logic)
- **No dependencies** on infrastructure (database, APIs, hardware)
- Contains core business rules and calculations
- Fully testable without mocks
- Examples:
  - `domain/services/priority_calculator.py` - Priority calculation logic
  - `domain/utils/priority_helpers.py` - Shared utility functions

### 2. Repository Pattern
- **Interfaces** defined in `domain/repositories/protocols.py` (using Protocol)
- **Implementations** in `app/repositories/` (SQLite-based)
- All database access goes through repositories
- Easy to swap implementations (e.g., PostgreSQL instead of SQLite)

### 3. Dependency Injection
- FastAPI dependencies in `infrastructure/dependencies.py`
- Repositories injected via `Depends()`
- Makes testing easier (can inject mocks)

### 4. Application Services
- Orchestrate domain services and repositories
- Handle transactions and coordination
- **No business logic** (that's in domain layer)
- Examples:
  - `PortfolioService` - Portfolio operations and analytics
  - `RebalancingService` - Rebalancing use cases
  - `ScoringService` - Stock scoring orchestration
  - `TradeExecutionService` - Trade execution coordination
  - `PortfolioOptimizer` - Portfolio optimization (Mean-Variance, HRP)

### 5. API Layer (Thin Controllers)
- API endpoints are thin - just request/response handling
- Delegate to application services
- No business logic in API layer

## Current Status

### ✅ Architecture Complete
- Domain layer with pure business logic (scoring, analytics, planning, events)
- Repository pattern with Protocol-based interfaces
- Dependency injection via FastAPI `Depends()`
- Application services for orchestration
- External API clients in `infrastructure/external/` (Tradernet, Yahoo Finance)
- LED display in `infrastructure/hardware/`
- Portfolio optimizer with Mean-Variance and HRP algorithms
- Holistic planner for rebalancing recommendations
- Event-driven architecture with domain events
- Comprehensive test suite (unit and integration tests)

### Legacy Code
- `app/services/allocator.py` - Still used by some jobs (backward compatible)
- `app/services/tradernet.py` and `app/services/yahoo.py` - Legacy clients, prefer `infrastructure/external/`

### Known Architecture Violations

The codebase has some pragmatic violations of Clean Architecture principles. These are documented here for transparency and future refactoring:

#### 1. Domain Layer Importing Infrastructure/Repositories

**Violation:** Some domain services import repositories or infrastructure directly.

**Affected Files:**
- `domain/services/rebalancing_triggers.py` - Imports `SettingsRepository`
- `domain/analytics/market_regime.py` - Imports `TradernetClient`, `get_recommendation_cache`
- `domain/services/stock_discovery.py` - Imports `TradernetClient`
- `domain/services/ticker_content_service.py` - Imports `cache`, `TradernetClient`
- `domain/planning/holistic_planner.py` - Imports `yahoo_finance`, `SettingsRepository`, `TradeRepository`
- `domain/scoring/*` (multiple files) - Import `CalculationsRepository`
- `domain/services/exchange_rate_service.py` - Imports `DatabaseManager`

**Justification:**
- These violations exist for pragmatic reasons (performance, convenience)
- Domain logic needs access to cached calculations and external data
- Full dependency injection would require significant refactoring

**Migration Path:**
1. Short-term: Document violations and add comments to affected files
2. Medium-term: Use dependency injection for repositories in domain services
3. Long-term: Refactor to pass data/context objects instead of repositories

#### 2. Direct Database Access Outside Repositories

**Violation:** Some code calls `get_db_manager()` directly instead of using repositories.

**Affected Areas:**
- `app/jobs/*` - Multiple jobs call `get_db_manager()` directly
- `app/repositories/stock.py:111` - `get_with_scores()` method accesses multiple databases
- `app/infrastructure/recommendation_cache.py` - Direct database access

**Justification:**
- Jobs need to coordinate multiple repositories
- Some operations span multiple databases
- Repository pattern would require composite repositories

**Migration Path:**
1. Create composite repositories for multi-database operations
2. Inject repositories into jobs via dependency injection
3. Create repository for recommendation cache operations

#### 3. Repository Pattern Inconsistency

**Violation:** Some repositories access multiple databases directly.

**Example:** `StockRepository.get_with_scores()` accesses both `config.db` and `state.db`.

**Justification:**
- Composite queries need data from multiple databases
- Creating separate repositories for each database would complicate queries

**Migration Path:**
- Create composite repository or service that orchestrates multiple repositories
- Or inject multiple repositories as dependencies

## Benefits

1. **Testability**: Domain logic can be tested without database
2. **Maintainability**: Clear separation of concerns
3. **Flexibility**: Easy to swap implementations (e.g., different database)
4. **No Duplication**: Shared utilities centralized
5. **Scalability**: Architecture supports growth

## Testing

- **Unit Tests**: `tests/unit/domain/` - Test domain logic in isolation
- **Integration Tests**: `tests/integration/` - Test repository implementations
- Run tests: `pytest`

## Usage Examples

### Using Repositories

```python
from app.infrastructure.dependencies import StockRepositoryDep

@router.get("/stocks")
async def get_stocks(
    stock_repo: StockRepositoryDep,
):
    stocks = await stock_repo.get_all_active()
    return stocks
```

### Using Application Services

```python
from app.infrastructure.dependencies import PortfolioServiceDep

@router.get("/portfolio/summary")
async def get_summary(
    portfolio_service: PortfolioServiceDep,
):
    summary = await portfolio_service.get_portfolio_summary()
    return summary
```

### Using Domain Services

```python
from app.domain.services.priority_calculator import PriorityCalculator

results = PriorityCalculator.calculate_priorities(
    priority_inputs,
    geo_weights,
    industry_weights,
)
```
