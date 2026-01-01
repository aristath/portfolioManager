# REST API Usage Examples

Complete guide to using the Arduino Trader REST APIs.

## Starting Services

### Option 1: Start Individual Services

Each service can be started independently from its own directory:

```bash
# Universe Service (Port 8001)
cd services/universe
../../venv/bin/python main.py

# Portfolio Service (Port 8002)
cd services/portfolio
../../venv/bin/python main.py

# Trading Service (Port 8003)
cd services/trading
../../venv/bin/python main.py

# Scoring Service (Port 8004)
cd services/scoring
../../venv/bin/python main.py

# Optimization Service (Port 8005)
cd services/optimization
../../venv/bin/python main.py

# Planning Service (Port 8006)
cd services/planning
../../venv/bin/python main.py

# Gateway Service (Port 8007)
cd services/gateway
../../venv/bin/python main.py
```

### Option 2: Start All Services (Background)

```bash
# Start all services in background
for service in universe portfolio trading scoring optimization planning gateway; do
    cd "services/$service"
    ../../venv/bin/python main.py > "/tmp/$service.log" 2>&1 &
    cd ../..
done

# Check all services are running
curl http://localhost:8001/universe/health
curl http://localhost:8002/portfolio/health
curl http://localhost:8003/trading/health
curl http://localhost:8004/scoring/health
curl http://localhost:8005/optimization/health
curl http://localhost:8006/planning/health
curl http://localhost:8007/gateway/health
```

## API Usage Examples

### Universe Service (Port 8001)

```bash
# List all tradable securities
curl http://localhost:8001/universe/securities?tradable_only=true

# Get specific security
curl http://localhost:8001/universe/securities/US0378331005

# Search securities
curl http://localhost:8001/universe/search?q=AAPL

# Sync price data
curl -X POST http://localhost:8001/universe/sync/prices

# Sync fundamental data
curl -X POST http://localhost:8001/universe/sync/fundamentals

# Get market data
curl http://localhost:8001/universe/market-data/US0378331005

# Add security
curl -X POST http://localhost:8001/universe/securities \
  -H "Content-Type: application/json" \
  -d '{
    "isin": "US0378331005",
    "symbol": "AAPL",
    "name": "Apple Inc.",
    "asset_class": "equity",
    "tradable": true
  }'

# Remove security
curl -X DELETE http://localhost:8001/universe/securities/US0378331005
```

### Portfolio Service (Port 8002)

```bash
# Get all positions
curl http://localhost:8002/portfolio/positions

# Get specific position
curl http://localhost:8002/portfolio/positions/AAPL

# Get portfolio summary
curl http://localhost:8002/portfolio/summary

# Get performance metrics
curl http://localhost:8002/portfolio/performance

# Get cash balance
curl http://localhost:8002/portfolio/cash

# Sync positions from broker
curl -X POST http://localhost:8002/portfolio/positions/sync
```

### Trading Service (Port 8003)

```bash
# Execute single trade
curl -X POST http://localhost:8003/trading/execute \
  -H "Content-Type: application/json" \
  -d '{
    "isin": "US0378331005",
    "symbol": "AAPL",
    "quantity": 10,
    "side": "buy",
    "order_type": "market"
  }'

# Batch execute trades
curl -X POST http://localhost:8003/trading/execute/batch \
  -H "Content-Type: application/json" \
  -d '{
    "trades": [
      {
        "isin": "US0378331005",
        "symbol": "AAPL",
        "quantity": 10,
        "side": "buy",
        "order_type": "market"
      },
      {
        "isin": "US5949181045",
        "symbol": "MSFT",
        "quantity": 5,
        "side": "buy",
        "order_type": "limit",
        "limit_price": 350.00
      }
    ]
  }'

# Get trade status
curl http://localhost:8003/trading/status/trade-12345

# Get trade history
curl http://localhost:8003/trading/history?limit=50

# Cancel trade
curl -X POST http://localhost:8003/trading/cancel/trade-12345

# Validate trade
curl -X POST http://localhost:8003/trading/validate \
  -H "Content-Type: application/json" \
  -d '{
    "isin": "US0378331005",
    "symbol": "AAPL",
    "quantity": 10,
    "side": "buy"
  }'
```

### Scoring Service (Port 8004)

```bash
# Score single security
curl -X POST http://localhost:8004/scoring/score \
  -H "Content-Type: application/json" \
  -d '{
    "isin": "US0378331005"
  }'

# Batch score securities
curl -X POST http://localhost:8004/scoring/score/batch \
  -H "Content-Type: application/json" \
  -d '{
    "isins": ["US0378331005", "US5949181045", "US02079K1079"]
  }'

# Score portfolio
curl -X POST http://localhost:8004/scoring/score/portfolio \
  -H "Content-Type: application/json" \
  -d '{
    "positions": [
      {
        "isin": "US0378331005",
        "symbol": "AAPL",
        "quantity": 100,
        "market_value": 17500.00
      },
      {
        "isin": "US5949181045",
        "symbol": "MSFT",
        "quantity": 50,
        "market_value": 17500.00
      }
    ]
  }'

# Get score history
curl http://localhost:8004/scoring/history/US0378331005
```

### Optimization Service (Port 8005)

```bash
# Optimize allocation
curl -X POST http://localhost:8005/optimization/allocation \
  -H "Content-Type: application/json" \
  -d '{
    "current_positions": [
      {
        "isin": "US0378331005",
        "symbol": "AAPL",
        "quantity": 100,
        "market_value": 17500.00
      }
    ],
    "targets": [
      {
        "isin": "US0378331005",
        "target_weight": 0.5
      },
      {
        "isin": "US5949181045",
        "target_weight": 0.5
      }
    ],
    "total_value": 35000.00,
    "available_cash": 17500.00
  }'

# Optimize execution
curl -X POST http://localhost:8005/optimization/execution \
  -H "Content-Type: application/json" \
  -d '{
    "trades": [
      {
        "isin": "US0378331005",
        "quantity": 10,
        "side": "buy"
      },
      {
        "isin": "US5949181045",
        "quantity": 5,
        "side": "sell"
      }
    ],
    "constraints": {
      "max_slippage_bps": 50,
      "urgency": "normal"
    }
  }'

# Calculate rebalancing
curl -X POST http://localhost:8005/optimization/rebalancing \
  -H "Content-Type: application/json" \
  -d '{
    "current_positions": [...],
    "target_weights": {...},
    "total_value": 100000.00,
    "available_cash": 5000.00
  }'
```

### Planning Service (Port 8006)

```bash
# Create plan
curl -X POST http://localhost:8006/planning/create \
  -H "Content-Type: application/json" \
  -d '{
    "portfolio_hash": "abc123",
    "deposit_amount": 1000.00,
    "max_plans": 5
  }'

# Get specific plan
curl http://localhost:8006/planning/plans/plan-uuid-here

# List plans for portfolio
curl http://localhost:8006/planning/plans?portfolio_hash=abc123

# Get best plan
curl http://localhost:8006/planning/best?portfolio_hash=abc123
```

### Gateway Service (Port 8007)

```bash
# Get system status
curl http://localhost:8007/gateway/status

# Trigger trading cycle
curl -X POST http://localhost:8007/gateway/trading-cycle \
  -H "Content-Type: application/json" \
  -d '{
    "force": false,
    "deposit_amount": 1000.00
  }'

# Process deposit
curl -X POST http://localhost:8007/gateway/deposit \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": "main",
    "amount": 1000.00
  }'

# Get service health
curl http://localhost:8007/gateway/services/universe/health
```

## Using HTTP Clients from Python Code

### Basic Usage

```python
from app.infrastructure.service_discovery import get_service_locator

# Get service locator
locator = get_service_locator()

# Create clients
universe_client = locator.create_http_client("universe")
portfolio_client = locator.create_http_client("portfolio")
trading_client = locator.create_http_client("trading")
scoring_client = locator.create_http_client("scoring")
optimization_client = locator.create_http_client("optimization")
planning_client = locator.create_http_client("planning")
gateway_client = locator.create_http_client("gateway")

# Use clients
securities = await universe_client.get_securities(tradable_only=True)
positions = await portfolio_client.get_positions()
```

### Universe Client Example

```python
from app.infrastructure.http_clients import UniverseHTTPClient

client = UniverseHTTPClient(
    base_url="http://localhost:8001",
    service_name="universe",
    timeout=30.0
)

# Get securities
securities = await client.get_securities(tradable_only=True)

# Search securities
results = await client.search_securities(query="AAPL")

# Get specific security
security = await client.get_security(isin="US0378331005")

# Sync prices
await client.sync_prices()

# Sync fundamentals
await client.sync_fundamentals()

# Get market data
market_data = await client.get_market_data(isin="US0378331005")

# Add security
await client.add_security({
    "isin": "US0378331005",
    "symbol": "AAPL",
    "name": "Apple Inc.",
    "asset_class": "equity",
    "tradable": True
})

# Remove security
await client.remove_security(isin="US0378331005")
```

### Portfolio Client Example

```python
from app.infrastructure.http_clients import PortfolioHTTPClient

client = PortfolioHTTPClient(
    base_url="http://localhost:8002",
    service_name="portfolio"
)

# Get all positions
positions = await client.get_positions()

# Get specific position
position = await client.get_position(symbol="AAPL")

# Get summary
summary = await client.get_summary()

# Get performance
performance = await client.get_performance()

# Get cash balance
cash = await client.get_cash()

# Sync positions
await client.sync_positions()
```

### Trading Client Example

```python
from app.infrastructure.http_clients import TradingHTTPClient

client = TradingHTTPClient(
    base_url="http://localhost:8003",
    service_name="trading"
)

# Execute trade
result = await client.execute_trade({
    "isin": "US0378331005",
    "symbol": "AAPL",
    "quantity": 10,
    "side": "buy",
    "order_type": "market"
})

# Batch execute
results = await client.batch_execute([
    {"isin": "US0378331005", "symbol": "AAPL", "quantity": 10, "side": "buy"},
    {"isin": "US5949181045", "symbol": "MSFT", "quantity": 5, "side": "buy"}
])

# Get trade status
status = await client.get_trade_status(trade_id="trade-12345")

# Get history
history = await client.get_trade_history(limit=50)

# Cancel trade
await client.cancel_trade(trade_id="trade-12345")

# Validate trade
validation = await client.validate_trade({
    "isin": "US0378331005",
    "symbol": "AAPL",
    "quantity": 10,
    "side": "buy"
})
```

### Scoring Client Example

```python
from app.infrastructure.http_clients import ScoringHTTPClient

client = ScoringHTTPClient(
    base_url="http://localhost:8004",
    service_name="scoring"
)

# Score single security
score = await client.score_security(isin="US0378331005")

# Batch score
scores = await client.batch_score_securities(
    isins=["US0378331005", "US5949181045"]
)

# Score portfolio
portfolio_score = await client.score_portfolio(positions=[...])

# Get history
history = await client.get_score_history(isin="US0378331005")
```

### Gateway Client Example

```python
from app.infrastructure.http_clients import GatewayHTTPClient

client = GatewayHTTPClient(
    base_url="http://localhost:8007",
    service_name="gateway"
)

# Get system status
status = await client.get_system_status()

# Trigger trading cycle
result = await client.trigger_trading_cycle(
    force=False,
    deposit_amount=1000.00
)

# Process deposit
deposit_result = await client.process_deposit(
    account_id="main",
    amount=1000.00
)

# Get service health
health = await client.get_service_health(service_name="universe")
```

## Common Workflows

### Monthly Deposit and Rebalancing

```python
async def process_monthly_deposit(amount: float):
    """Process monthly deposit and trigger portfolio rebalancing."""

    # 1. Process deposit
    deposit_result = await gateway_client.process_deposit(
        account_id="main",
        amount=amount
    )

    # 2. Trigger trading cycle
    cycle_result = await gateway_client.trigger_trading_cycle(
        force=False,
        deposit_amount=amount
    )

    return cycle_result
```

### Score and Buy Best Securities

```python
async def score_and_buy(isins: list[str], budget: float):
    """Score securities and buy the best ones within budget."""

    # 1. Score all securities
    scores = await scoring_client.batch_score_securities(isins=isins)

    # 2. Sort by score
    sorted_scores = sorted(scores, key=lambda x: x["total_score"], reverse=True)

    # 3. Calculate allocation
    allocation_request = {
        "current_positions": await portfolio_client.get_positions(),
        "targets": [
            {"isin": s["isin"], "target_weight": 1.0 / len(sorted_scores)}
            for s in sorted_scores[:5]  # Top 5 securities
        ],
        "total_value": budget,
        "available_cash": budget
    }

    allocation = await optimization_client.optimize_allocation(allocation_request)

    # 4. Execute trades
    trades = [
        {
            "isin": change["isin"],
            "symbol": change["symbol"],
            "quantity": change["quantity"],
            "side": change["action"]
        }
        for change in allocation["changes"]
    ]

    results = await trading_client.batch_execute(trades)

    return results
```

### Full Rebalancing Workflow

```python
async def full_rebalance():
    """Complete portfolio rebalancing workflow."""

    # 1. Get current portfolio
    positions = await portfolio_client.get_positions()
    summary = await portfolio_client.get_summary()

    # 2. Score all holdings
    isins = [p["isin"] for p in positions]
    scores = await scoring_client.batch_score_securities(isins=isins)

    # 3. Create rebalancing plan
    plan_request = {
        "portfolio_hash": summary["hash"],
        "deposit_amount": 0.00,
        "max_plans": 5
    }

    plan = await planning_client.create_plan(plan_request)

    # 4. Execute plan actions
    trades = [
        {
            "isin": action["isin"],
            "symbol": action["symbol"],
            "quantity": action["quantity"],
            "side": action["action"]
        }
        for action in plan["actions"]
    ]

    results = await trading_client.batch_execute(trades)

    return results
```

## Error Handling

All HTTP clients include built-in error handling with circuit breakers and retries:

```python
from httpx import HTTPStatusError

try:
    securities = await universe_client.get_securities()
except HTTPStatusError as e:
    if e.response.status_code == 404:
        print("Resource not found")
    elif e.response.status_code == 500:
        print("Server error")
    else:
        print(f"HTTP error: {e.response.status_code}")
except Exception as e:
    print(f"Unexpected error: {e}")
```

## API Documentation

Each service provides interactive API documentation at:

- Universe: http://localhost:8001/docs
- Portfolio: http://localhost:8002/docs
- Trading: http://localhost:8003/docs
- Scoring: http://localhost:8004/docs
- Optimization: http://localhost:8005/docs
- Planning: http://localhost:8006/docs
- Gateway: http://localhost:8007/docs

## Testing

Run the comprehensive test suite:

```bash
# Test all service startups
./test_all_services.sh

# Validate imports and structure
./venv/bin/python validate_rest_services.py
```

---

*Generated with [Claude Code](https://claude.com/claude-code)*
*Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>*
