# Quick Start - New Architecture

## Overview

The arduino-trader project now follows **Clean Architecture** principles. This guide helps you understand and work with the new structure.

## Key Concepts

### 1. Repository Pattern
All database access goes through repository interfaces:

```python
# ✅ Good - Use repository
from app.infrastructure.dependencies import StockRepositoryDep

@router.get("/stocks/{symbol}")
async def get_stock(
    stock_repo: StockRepositoryDep,
):
    return await stock_repo.get_by_symbol(symbol)

# ❌ Bad - Direct database access
@router.get("/stocks/{symbol}")
async def get_stock(db: aiosqlite.Connection = Depends(get_db)):
    cursor = await db.execute("SELECT * FROM stocks WHERE symbol = ?", (symbol,))
    # ...
```

### 2. Application Services
Orchestrate business logic using application services:

```python
# ✅ Good - Use application service
portfolio_service = PortfolioService(portfolio_repo, position_repo, allocation_repo)
summary = await portfolio_service.get_portfolio_summary()

# ❌ Bad - Direct repository calls in API
positions = await position_repo.get_all()
# ... complex logic in API endpoint ...
```

### 3. Domain Services
Pure business logic with no infrastructure dependencies:

```python
# ✅ Good - Domain service (testable without database)
result = PriorityCalculator.calculate_priority(input, geo_weights, industry_weights)

# ❌ Bad - Business logic in API or infrastructure
# ... calculations mixed with database queries ...
```

## Common Tasks

### Adding a New Endpoint

1. **Define what data you need** - Which repositories?
2. **Check if an application service exists** - Use it if available
3. **Create thin controller** - Just delegate to service/repository

```python
from app.infrastructure.dependencies import StockRepositoryDep

@router.get("/new-endpoint")
async def new_endpoint(
    stock_repo: StockRepositoryDep,
):
    # Simple delegation
    stocks = await stock_repo.get_all_active()
    return stocks
```

### Adding a New Repository Method

1. **Add to interface** (`domain/repositories/protocols.py`):
```python
class StockRepository(Protocol):
    async def get_by_industry(self, industry: str) -> List[Stock]:
        ...
```

2. **Implement** (`app/repositories/stock.py`):
```python
class StockRepository:
    async def get_by_industry(self, industry: str) -> List[Stock]:
        # SQLite implementation
        pass
```

3. **Use in API**:
```python
stocks = await stock_repo.get_by_industry("Technology")
```

### Creating an Application Service

```python
# application/services/new_service.py
class NewService:
    def __init__(
        self,
        repo1: Repository1,
        repo2: Repository2,
    ):
        self._repo1 = repo1
        self._repo2 = repo2

    async def do_something(self, param: str) -> Result:
        # Orchestrate repositories
        data1 = await self._repo1.get(param)
        data2 = await self._repo2.get(param)

        # Business logic
        result = self._calculate(data1, data2)

        return result
```

## File Locations

### Where to Put Code

| What | Where |
|------|-------|
| Business logic (no DB) | `app/domain/services/` |
| Data access interface | `app/domain/repositories/protocols.py` |
| Database implementation | `app/repositories/` |
| Orchestration logic | `app/application/services/` |
| API endpoints | `app/api/` |
| Hardware/External | `app/infrastructure/` |

### Examples

- **Priority calculation** → `app/domain/services/priority_calculator.py`
- **Stock repository interface** → `app/domain/repositories/protocols.py` (StockRepository protocol)
- **SQLite stock repository** → `app/repositories/stock.py`
- **Portfolio operations** → `app/application/services/portfolio_service.py`
- **Stock API endpoints** → `app/api/stocks.py`

## Testing

### Unit Test (Domain Logic)
```python
def test_priority_calculator():
    # No database needed!
    input = PriorityInput(stock_score=80, ...)
    result = PriorityCalculator.calculate_priority(input, {}, {})
    assert result.combined_priority > 0
```

### Integration Test (Repository)
```python
@pytest.mark.asyncio
async def test_stock_repository(stock_repo):
    # Uses test database
    stock = await stock_repo.get_by_symbol("AAPL")
    assert stock is not None
```

## Benefits

✅ **Testable** - Domain logic testable without database
✅ **Maintainable** - Clear where code belongs
✅ **Flexible** - Easy to swap implementations
✅ **Scalable** - Easy to add features
✅ **Clean** - No duplication, clear structure

## Need Help?

- See `ARCHITECTURE.md` for detailed architecture docs
- See `MIGRATION_NOTES.md` for migrating existing code
- See `README_ARCHITECTURE.md` for more examples
