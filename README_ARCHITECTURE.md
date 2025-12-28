# Arduino Trader - Architecture Overview

## Quick Start

This project follows **Clean Architecture** principles with clear separation of concerns.

### Project Structure

```
app/
├── domain/              # Pure business logic (no infrastructure dependencies)
│   ├── repositories/    # Repository interfaces (Protocol-based)
│   ├── services/        # Domain services (PriorityCalculator, SettingsService)
│   ├── scoring/         # Stock scoring system (8 groups)
│   ├── analytics/       # Portfolio analytics
│   ├── planning/        # Rebalancing strategies
│   ├── events/          # Domain events
│   └── value_objects/   # Value objects (Currency, Money, etc.)
│
├── infrastructure/      # External concerns
│   ├── database/        # Database manager, schemas
│   ├── external/        # API clients (Tradernet, Yahoo Finance)
│   ├── hardware/        # LED display
│   └── dependencies.py  # FastAPI dependency injection
│
├── application/         # Orchestration layer
│   └── services/        # Application services
│       ├── optimization/    # Portfolio optimizer
│       ├── recommendation/  # Recommendation services
│       ├── PortfolioService
│       ├── RebalancingService
│       ├── ScoringService
│       └── TradeExecutionService
│
├── repositories/        # Repository implementations (SQLite)
│
└── api/                 # Thin controllers (delegation only)
    ├── stocks.py
    ├── portfolio.py
    ├── trades.py
    ├── allocation.py
    ├── cash_flows.py
    ├── charts.py
    ├── optimizer.py
    ├── recommendations.py
    ├── settings.py
    └── status.py
```

## Key Principles

### 1. Repository Pattern
All database access goes through repository interfaces:

```python
# Domain layer defines interface (Protocol-based)
from typing import Protocol

class StockRepository(Protocol):
    async def get_by_symbol(self, symbol: str) -> Optional[Stock]:
        ...

# Implementation in app/repositories/stock.py
class StockRepository:
    async def get_by_symbol(self, symbol: str) -> Optional[Stock]:
        # SQLite implementation
        pass

# API uses it via dependency injection
from app.infrastructure.dependencies import StockRepositoryDep

@router.get("/stocks/{symbol}")
async def get_stock(
    stock_repo: StockRepositoryDep,
):
    return await stock_repo.get_by_symbol(symbol)
```

### 2. Dependency Injection
FastAPI dependencies provide repository instances:

```python
# infrastructure/dependencies.py defines dependency types
# Repositories are injected via FastAPI Depends()

# Usage in API
from app.infrastructure.dependencies import StockRepositoryDep

@router.get("/stocks")
async def get_stocks(
    stock_repo: StockRepositoryDep,
):
    return await stock_repo.get_all_active()
```

### 3. Application Services
Orchestrate domain services and repositories:

```python
class PortfolioService:
    def __init__(
        self,
        portfolio_repo: PortfolioRepository,
        position_repo: PositionRepository,
        allocation_repo: AllocationRepository,
    ):
        self._portfolio_repo = portfolio_repo
        self._position_repo = position_repo
        self._allocation_repo = allocation_repo

    async def get_portfolio_summary(self) -> PortfolioSummary:
        # Orchestrates multiple repositories
        positions = await self._position_repo.get_all()
        latest = await self._portfolio_repo.get_latest()
        # ... business logic ...
        return summary
```

### 4. Domain Services
Pure business logic with no infrastructure dependencies:

```python
class PriorityCalculator:
    @staticmethod
    def calculate_priority(
        input: PriorityInput,
        geo_weights: Dict[str, float],
        industry_weights: Dict[str, float],
    ) -> PriorityResult:
        # Pure business logic - no database, no I/O
        quality_score = input.stock_score * 0.4
        # ... calculations ...
        return PriorityResult(...)
```

## Usage Examples

### Adding a New Repository

1. **Define interface** in `domain/repositories/protocols.py`:
```python
from typing import Protocol

class NewRepository(Protocol):
    async def get_by_id(self, id: int) -> Optional[NewEntity]:
        ...
```

2. **Implement** in `app/repositories/new.py`:
```python
class NewRepository:
    async def get_by_id(self, id: int) -> Optional[NewEntity]:
        # SQLite implementation
        pass
```

3. **Add dependency** in `infrastructure/dependencies.py`:
```python
NewRepositoryDep = Annotated[NewRepository, Depends(get_new_repository)]
```

4. **Use in API**:
```python
from app.infrastructure.dependencies import NewRepositoryDep

@router.get("/new/{id}")
async def get_new(
    new_repo: NewRepositoryDep,
):
    return await new_repo.get_by_id(id)
```

### Creating an Application Service

```python
# application/services/new_service.py
class NewService:
    def __init__(
        self,
        new_repo: NewRepository,
        other_repo: OtherRepository,
    ):
        self._new_repo = new_repo
        self._other_repo = other_repo

    async def do_something(self, id: int) -> Result:
        # Orchestrate repositories and domain services
        entity = await self._new_repo.get_by_id(id)
        # ... business logic ...
        return result
```

## Testing

### Unit Tests (Domain Logic)
Test domain services without database:

```python
def test_priority_calculator():
    input = PriorityInput(stock_score=80, ...)
    result = PriorityCalculator.calculate_priority(input, {}, {})
    assert result.combined_priority > 0
```

### Integration Tests (Repositories)
Test repository implementations with test database:

```python
@pytest.mark.asyncio
async def test_stock_repository(stock_repo):
    stock = await stock_repo.get_by_symbol("AAPL")
    assert stock is not None
```

## Benefits

✅ **Testable** - Domain logic testable without database
✅ **Maintainable** - Clear separation of concerns
✅ **Flexible** - Easy to swap implementations
✅ **Scalable** - Easy to add new features
✅ **Clean** - No code duplication
✅ **Type-safe** - Full type hints throughout

## Migration Notes

See `MIGRATION_NOTES.md` for details on migrating existing code to use the new architecture.

## Documentation

- `ARCHITECTURE.md` - Detailed architecture documentation
- `MIGRATION_NOTES.md` - How to migrate existing code
- `QUICK_START.md` - Quick start guide with examples
