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
