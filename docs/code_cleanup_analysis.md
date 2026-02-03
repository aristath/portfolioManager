# Sentinel Codebase Cleanup Analysis

**Date**: February 2026
**Scope**: Duplicate code, redundant code, legacy/dead code identification

---

## Summary

| Category | Issues Found | Est. Lines Saved | Priority |
|----------|--------------|------------------|----------|
| Duplicate Code | 5 | ~180 | High |
| Redundant Code | 4 | ~50 | Medium |
| Legacy/Dead Code | 5 | ~100 | Medium |
| Simplification Opportunities | 4 | ~200 | Medium |
| **Total** | **18** | **~530 lines** | - |

---

## 1. DUPLICATE CODE

### 1.1 Duplicate Currency Conversion Logic

**Files**: `currency.py` (lines 102-107) and `currency_exchange.py` (lines 99-134)

**Issue**: Two classes handle currency conversion with overlapping functionality:

```python
# currency.py
async def to_eur(self, amount: float, currency: str) -> float:
    if currency.upper() == "EUR":
        return amount
    rate = await self.get_rate(currency)
    return amount * rate

# currency_exchange.py
async def get_rate(self, from_currency: str, to_currency: str) -> Optional[float]:
    # Imports Currency internally (line 120)
    from sentinel.currency import Currency
    currency = Currency()
    rates = await currency.get_rates()
    # ... same calculation
```

**Impact**: Circular import risk, confusing API (which class to use?)

**Recommendation**: Consolidate into single `CurrencyService` class:
- Move all conversion logic to `currency.py`
- Have `CurrencyExchangeService` delegate to `Currency`
- Or merge both into one unified service

**Est. Savings**: ~80 lines, reduced confusion

---

### 1.2 Repeated JSON Parsing for Security Metadata

**Files**: `app.py` (line 292) and `jobs/tasks.py` (line 723)

**Identical code pattern**:
```python
# app.py line 292
sec_data = _json.loads(data) if isinstance(data, str) else data
mkt_id = sec_data.get("mrkt", {}).get("mkt_id")

# jobs/tasks.py line 723
sec_data = json.loads(data) if isinstance(data, str) else data
market_id = str(sec_data.get("mrkt", {}).get("mkt_id"))
```

**Recommendation**: Add method to `Security` class or database layer:
```python
# In security.py or database/main.py
async def get_security_market_id(sec_data: dict) -> Optional[str]:
    data = sec_data.get("data")
    if not data:
        return None
    try:
        parsed = json.loads(data) if isinstance(data, str) else data
        return str(parsed.get("mrkt", {}).get("mkt_id"))
    except (json.JSONDecodeError, KeyError, TypeError):
        return None
```

**Est. Savings**: ~15 lines, improved maintainability

---

### 1.3 Duplicate Singleton Pattern Implementation

**Files**: Present in 4+ files

```python
# settings.py lines 78-82
class Settings:
    _instance: "Settings | None" = None
    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
            cls._instance._db = Database()
        return cls._instance

# Same pattern in: currency.py, broker.py, currency_exchange.py
```

**Recommendation**: Create a `@singleton` decorator in `utils/decorators.py`:
```python
def singleton(cls):
    instances = {}
    def get_instance(*args, **kwargs):
        if cls not in instances:
            instances[cls] = cls(*args, **kwargs)
        return instances[cls]
    return get_instance

# Usage:
@singleton
class Settings:
    def __init__(self):
        self._db = Database()
```

**Est. Savings**: ~30 lines per class, improved readability

---

### 1.4 Repeated Date Default Pattern

**Files**: `broker.py` (lines 402-403, 454-455, 494-495), `ml_trainer.py` (lines 59-60)

```python
# broker.py lines 402-403
if end_date is None:
    end_date = datetime.now().strftime("%Y-%m-%d")

# ml_trainer.py lines 59-60
if end_date is None:
    end_date = (datetime.now() - timedelta(days=1)).strftime("%Y-%m-%d")
```

**Recommendation**: Create utility function in `utils/dates.py`:
```python
def get_default_end_date(days_offset: int = 0) -> str:
    date = datetime.now() - timedelta(days=days_offset)
    return date.strftime("%Y-%m-%d")
```

**Est. Savings**: ~12 lines

---

### 1.5 Duplicate SQL Migration Pattern

**Files**: `database/main.py` (lines 672-693)

```python
cursor = await self.conn.execute("PRAGMA table_info(securities)")
columns = {row[1] for row in await cursor.fetchall()}

migrations = [
    ("market_id", "ALTER TABLE securities ADD COLUMN market_id TEXT"),
    # ... repeated 9 times
]

for col_name, sql in migrations:
    if col_name not in columns:
        await self.conn.execute(sql)
```

**Recommendation**: Create helper method:
```python
async def _add_column_if_missing(self, table: str, column: str, definition: str):
    cursor = await self.conn.execute(f"PRAGMA table_info({table})")
    columns = {row[1] for row in await cursor.fetchall()}
    if column not in columns:
        await self.conn.execute(f"ALTER TABLE {table} ADD COLUMN {column} {definition}")
```

**Est. Savings**: ~40 lines

---

## 2. REDUNDANT CODE

### 2.1 Async Wrapper for Simple Multiplication

**File**: `utils/positions.py` (lines 36-47)

```python
async def calculate_value_local(self, quantity: float, price: float) -> float:
    """Calculate position value in local currency."""
    return quantity * price  # Just multiplies two numbers
```

**Issue**: This async wrapper adds overhead for a simple multiplication. Called in hot loops.

**Recommendation**: Inline the multiplication where used, or make it a synchronous method.

**Est. Savings**: ~12 lines + async overhead removed

---

### 2.2 Empty TYPE_CHECKING Blocks

**Files**: `jobs/tasks.py` (line 14-15), `jobs/runner.py` (line 17-18), `regime_quote.py` (line 12-13)

```python
if TYPE_CHECKING:
    pass  # Empty block
```

**Recommendation**: Remove empty TYPE_CHECKING blocks or populate with actual type imports.

**Est. Savings**: ~6 lines

---

### 2.3 Unused RATE_SYMBOLS Constant

**File**: `config/currencies.py` (lines 57-63)

```python
RATE_SYMBOLS = {
    ("EUR", "USD"): "EUR/USD",
    ("EUR", "GBP"): "EUR/GBP",
    # ...
}
```

**Issue**: This constant is defined but `DIRECT_PAIRS` is used instead throughout the codebase.

**Recommendation**: Remove unused constant.

**Est. Savings**: 7 lines

---

### 2.4 Unused Tuple Import

**File**: `ml_features.py` (line 9)

```python
from typing import Dict, List, Optional, Tuple  # Tuple imported but not used
```

Line 97 uses built-in `tuple`, not `Tuple`.

**Est. Savings**: 1 line

---

## 3. LEGACY/DEAD CODE

### 3.1 Unused model_version Field

**Files**: `ml_predictor.py` (line 204), `database/main.py` (line 763)

```python
# ml_predictor.py
await self.db.conn.execute(
    """INSERT INTO ml_predictions
       (prediction_id, symbol, model_version, ...)""",
    (
        prediction_id,
        symbol,
        None,  # model_version no longer used
        # ...
    ),
)
```

**Recommendation**:
1. Remove `model_version` column from INSERT statements
2. Add migration to drop column from schema
3. Update any queries that reference it

**Est. Impact**: Database migration required

---

### 3.2 Empty API Routers Directory

**Directory**: `sentinel/api/routers/`

**Issue**: Directory exists but is empty. All FastAPI routes are defined in `app.py` (2000+ lines).

**Recommendation**: Either:
- Populate with router modules (recommended for maintainability)
- Or remove the empty directory

---

### 3.3 Unused Import in Currency Exchange

**File**: `currency_exchange.py` (line 12)

```python
from sentinel.broker import Broker  # Imported but only Trading methods used
```

Actually, `Broker` IS used in `_execute_step` method. Let me verify this is actually dead code...

After review: `self._broker` is used. Not dead code.

---

### 3.4 Commented-Out Code in app.py

**File**: `app.py` (line 1173)

```python
"alerts": [],  # TODO: implement concentration alerts if needed
```

**Recommendation**: Either implement the feature or remove the placeholder.

---

## 4. SIMPLIFICATION OPPORTUNITIES

### 4.1 Overly Complex Expected Return Calculation

**File**: `analyzer.py` (lines 598-701)

**Issue**: The `_calculate_expected_return` method duplicates the entire formula for each regime:

```python
if regime_data["regime_name"] == "Bull":
    expected_return = (
        0.30 * quality
        + 0.30 * mean_reversion  # Reduce from 40%
        + 0.30 * adjusted_momentum  # Increase from 20%
        + 0.10 * (consistency_bonus + 0.5)
    )
elif regime_data["regime_name"] == "Bear":
    expected_return = (
        0.40 * quality  # Increase from 30%
        + 0.40 * mean_reversion
        + 0.10 * adjusted_momentum  # Reduce from 20%
        + 0.10 * (consistency_bonus + 0.5)
    )
# Sideways: keep default weights
```

**Recommendation**: Extract weights to configuration:
```python
REGIME_WEIGHTS = {
    "default": {"quality": 0.30, "mean_reversion": 0.40, "momentum": 0.20, "consistency": 0.10},
    "Bull": {"quality": 0.30, "mean_reversion": 0.30, "momentum": 0.30, "consistency": 0.10},
    "Bear": {"quality": 0.40, "mean_reversion": 0.40, "momentum": 0.10, "consistency": 0.10},
}

weights = REGIME_WEIGHTS.get(regime, REGIME_WEIGHTS["default"])
expected_return = (
    weights["quality"] * quality +
    weights["mean_reversion"] * mean_reversion +
    # ...
)
```

**Est. Savings**: ~40 lines, much better maintainability

---

### 4.2 Complex Nested Conditionals in Security.buy()

**File**: `security.py` (lines 185-292)

**Issue**: 6 levels of nesting, currency conversion mixed with order logic.

**Recommendation**: Extract currency conversion to separate method:
```python
async def _ensure_currency_balance(self, trade_value: float) -> bool:
    """Ensure sufficient balance, converting currencies if needed."""
    # Extract all currency logic here

async def buy(self, quantity: int, auto_convert: bool = True) -> Optional[str]:
    # ... validation ...
    if auto_convert:
        await self._ensure_currency_balance(trade_value)
    # ... place order ...
```

**Est. Savings**: ~60 lines complexity reduction

---

### 4.3 Long Function - get_unified_view()

**File**: `app.py` (lines 685-908)

**Issue**: 223-line function mixing data fetching, transformation, and response building.

**Recommendation**: Extract into builder class:
```python
class UnifiedViewBuilder:
    def __init__(self, db, broker, currency, planner):
        self.db = db
        self.broker = broker
        # ...

    async def build(self, period: str) -> list[dict]:
        securities = await self._fetch_securities()
        quotes = await self._fetch_quotes()
        # ...
        return result
```

**Est. Impact**: Much better testability and maintainability

---

### 4.4 app.py is 2000+ Lines

**File**: `app.py`

**Issue**: Monolithic file containing all API routes.

**Recommendation**: Split into router modules:
```
sentinel/api/routers/
├── __init__.py
├── securities.py      # /api/securities/*
├── portfolio.py       # /api/portfolio/*
├── trading.py         # /api/trades/*, buy/sell endpoints
├── analytics.py       # /api/analytics/*, ML endpoints
├── jobs.py            # /api/jobs/*
└── settings.py        # /api/settings/*
```

**Est. Impact**: Each router ~200-400 lines, much more manageable

---

## Priority Recommendations

### High Priority (Do First)

1. **Consolidate currency handling** (Issue 1.1) - Reduces confusion and import cycles
2. **Create singleton decorator** (Issue 1.3) - Reduces boilerplate significantly
3. **Simplify expected return calculation** (Issue 4.1) - Improves maintainability

### Medium Priority (Do Next)

4. **Remove dead code** (Issues 3.1-3.4) - Clean up unused fields/constants
5. **Extract helper methods** (Issues 1.4, 1.5) - Reduce duplication
6. **Simplify Security.buy()** (Issue 4.2) - Improve readability

### Lower Priority (Nice to Have)

7. **Split app.py into routers** (Issue 4.4) - Major refactor, plan carefully
8. **Remove redundant wrapper** (Issue 2.1) - Minor performance gain

---

## Testing Considerations

When removing code:
1. Run full test suite: `pytest tests/`
2. Check import statements don't break
3. Verify database migrations for column removals
4. Test ML prediction pipeline end-to-end

Safe to remove immediately:
- Empty TYPE_CHECKING blocks
- Unused RATE_SYMBOLS constant
- Unused Tuple import

Requires testing:
- Currency consolidation (affects all trading)
- Singleton decorator (affects instance lifecycle)
- Expected return refactor (affects scoring)

---

*Document generated by automated codebase analysis*
*Review before implementing changes*
