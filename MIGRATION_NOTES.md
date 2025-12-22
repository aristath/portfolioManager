# Migration Notes

## Backward Compatibility

The refactoring maintains backward compatibility. The following functions in `app/services/allocator.py` are still available for existing code:

- `get_portfolio_summary(db)` - Still used by jobs
- `calculate_rebalance_trades(db, cash)` - Still used by jobs
- `execute_trades(db, trades)` - Still used by jobs and API (now uses repository internally)

## Migration Path for Jobs

Jobs (`app/jobs/cash_rebalance.py`, `app/jobs/monthly_rebalance.py`) can be incrementally migrated to use application services:

### Current (Backward Compatible)
```python
from app.services.allocator import calculate_rebalance_trades, execute_trades

trades = await calculate_rebalance_trades(db, cash_balance)
results = await execute_trades(db, trades)
```

### Future (Using Application Services)
```python
from app.infrastructure.dependencies import (
    get_stock_repository,
    get_position_repository,
    get_allocation_repository,
    get_portfolio_repository,
    get_trade_repository,
)
from app.application.services.rebalancing_service import RebalancingService
from app.application.services.trade_execution_service import TradeExecutionService

# Get repositories
stock_repo = await get_stock_repository()
position_repo = await get_position_repository()
allocation_repo = await get_allocation_repository()
portfolio_repo = await get_portfolio_repository()
trade_repo = await get_trade_repository()

# Use application services
rebalancing_service = RebalancingService(
    stock_repo, position_repo, allocation_repo, portfolio_repo
)
trades = await rebalancing_service.calculate_rebalance_trades(cash_balance)

trade_execution_service = TradeExecutionService(trade_repo)
results = await trade_execution_service.execute_trades(trades)
```

## What's New

### Using Repositories Directly
```python
from app.infrastructure.dependencies import get_stock_repository

@router.get("/stocks")
async def get_stocks(
    stock_repo: StockRepository = Depends(get_stock_repository)
):
    stocks = await stock_repo.get_all_active()
    return stocks
```

### Using Application Services
```python
from app.application.services.portfolio_service import PortfolioService

portfolio_service = PortfolioService(
    portfolio_repo, position_repo, allocation_repo
)
summary = await portfolio_service.get_portfolio_summary()
```

### Using Domain Services
```python
from app.domain.services.priority_calculator import PriorityCalculator

results = PriorityCalculator.calculate_priorities(
    priority_inputs, geo_weights, industry_weights
)
```

## Testing

All new code is testable:

```python
# Unit test domain logic (no database needed)
def test_priority_calculator():
    result = PriorityCalculator.calculate_priority(...)
    assert result.combined_priority > 0

# Integration test repositories (with test database)
@pytest.mark.asyncio
async def test_stock_repository(stock_repo):
    stock = await stock_repo.get_by_symbol("AAPL")
    assert stock is not None
```
