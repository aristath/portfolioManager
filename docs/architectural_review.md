# Sentinel Architectural Review

**Date:** 2025-02-03
**Scope:** Complete codebase analysis for architectural patterns, coupling, and refactoring opportunities

---

## Executive Summary

The codebase has **solid foundational patterns** (singletons, clean database layer, well-organized ML modules) but suffers from **monolithic modules** and **missing layered architecture**. The biggest risks are maintainability and testability as the system grows.

**Key Metrics:**
- Total Python files: 40+ core modules
- app.py: 2,228 lines (73 API routes)
- planner.py: 1,216 lines (13 methods)
- ML modules: 2,999 lines (well-organized)
- Test coverage: 610 tests (all passing)

---

## Critical Issues (Fix First)

### 1. God Module: app.py (2,228 lines, 73 API routes)

**Problem:** All API routes crammed into one file mixing 10+ domains:
- Settings, LED, Exchange Rates, Markets, Portfolio, Securities, Trading
- Allocation Targets, Scores, Planner, Jobs, Cache, Backtest, ML, Backup

**Current Endpoints by Domain:**
| Domain | Endpoints |
|--------|-----------|
| Settings | /api/settings, /api/led/* |
| Exchange Rates | /api/exchange-rates, /api/exchange-rates/sync |
| Markets | /api/markets/status, /api/meta/categories |
| Portfolio | /api/portfolio, /api/portfolio/sync, /api/portfolio/allocations |
| Securities | /api/securities, /api/securities/{symbol}, /api/securities/*/prices |
| Trading | /api/trades, /api/cashflows, /api/securities/*/buy, /api/securities/*/sell |
| Allocation | /api/allocation-targets, /api/allocation/* |
| Analysis | /api/scores/calculate, /api/unified |
| Jobs | /api/jobs/*, /api/backup/* |
| ML | /api/ml/*, /api/regime/* |
| System | /api/health, /api/cache/*, /api/backtest |

**Recommended Structure:**
```
sentinel/api/routers/
├── __init__.py
├── settings.py      # /api/settings, /api/led
├── portfolio.py     # /api/portfolio, /api/allocation*
├── securities.py    # /api/securities*, /api/prices*
├── trading.py       # /api/trades*, /api/cashflows*, buy/sell
├── analysis.py      # /api/scores*, /api/unified
├── jobs.py          # /api/jobs*, /api/backup
├── ml.py            # /api/ml/* endpoints
└── system.py        # /api/health, /api/cache, /api/backtest
```

**Migration Strategy:**
```python
# sentinel/app.py becomes:
from sentinel.api.routers import (
    settings, portfolio, securities, trading,
    analysis, jobs, ml, system
)

app.include_router(settings.router, prefix="/api")
app.include_router(portfolio.router, prefix="/api")
# ... etc
```

**Effort:** 2-3 days
**Impact:** High - Maintainability, testability

---

### 2. God Class: Planner (1,216 lines, 13 methods)

**Problem:** Planner handles too many responsibilities:
- Ideal portfolio calculation (`calculate_ideal_portfolio`)
- Recommendation generation (`get_recommendations`)
- Cash constraint application (`_apply_cash_constraint`)
- Fee calculation (`_calculate_transaction_cost`)
- Cool-off checking (`_check_cooloff_violation`)
- Diversification scoring (`_calculate_diversification_score`)
- Deficit sell generation (`_generate_deficit_sells`)
- Classic allocation (`_classic_allocation`)
- Rebalance summary (`get_rebalance_summary`)

**Current Dependencies Created in __init__:**
```python
self._db = db or Database()
self._broker = broker or Broker()
self._portfolio = portfolio or Portfolio()
self._analyzer = Analyzer(db=self._db)
self._settings = Settings()
self._currency = Currency()
self._ml_predictor = MLPredictor(db=self._db, settings=self._settings)
self._feature_extractor = FeatureExtractor(db=self._db)
```

**Recommended Structure:**
```
sentinel/planning/
├── __init__.py
├── planner.py              # Main facade (orchestrates only)
├── ideal_calculator.py     # calculate_ideal_portfolio()
├── recommendation_engine.py # get_recommendations()
├── constraint_applier.py   # _apply_cash_constraint()
└── models.py               # TradeRecommendation dataclass
```

**Refactoring Approach:**
```python
# planner.py becomes orchestrator only
class Planner:
    def __init__(
        self,
        calculator: IdealPortfolioCalculator,
        engine: RecommendationEngine,
        constraint_applier: CashConstraintApplier
    ):
        self._calculator = calculator
        self._engine = engine
        self._constraint_applier = constraint_applier

    async def get_recommendations(self) -> list[TradeRecommendation]:
        ideal = await self._calculator.calculate()
        recommendations = await self._engine.generate(ideal)
        return await self._constraint_applier.apply(recommendations)
```

**Effort:** 1-2 days
**Impact:** High - Testability, SRP compliance

---

### 3. Late Import Anti-Patterns

**Problem:** Deferred imports indicate circular dependency risks:

| File | Line | Late Import | Risk Level |
|------|------|-------------|------------|
| `analyzer.py` | 671 | `from sentinel.settings import Settings` | Medium |
| `analyzer.py` | 677 | `from sentinel.regime_hmm import RegimeDetector` | High |
| `security.py` | 213 | `from sentinel.currency_exchange import CurrencyExchangeService` | Medium |
| `planner.py` | 177 | `from sentinel.utils.scoring import adjust_score_for_conviction` | Low |
| `planner.py` | 281 | `from sentinel.utils.positions import PositionCalculator` | Low |
| `planner.py` | 387 | `from sentinel.utils.scoring import adjust_score_for_conviction` | Low |
| `planner.py` | 647 | `from sentinel.utils.fees import FeeCalculator` | Low |

**Analysis:**
- `analyzer.py` → `regime_hmm.py` is the most dangerous (RegimeDetector imports analyzer components)
- Multiple late imports in `planner.py` suggest tight coupling to utils

**Fix:** Create a proper service layer to break cycles:
```python
# sentinel/services/regime_service.py
class RegimeService:
    """Breaks analyzer ↔ regime_hmm cycle"""
    async def adjust_for_regime(self, symbol, expected_return):
        from sentinel.regime_hmm import RegimeDetector  # Now isolated
        ...
```

**Effort:** 1 day
**Impact:** Medium - Stability, prevent circular imports

---

## High-Priority Issues

### 4. Dependency Injection Inconsistency

**Problem:** Planner creates 7 dependencies in `__init__` with inconsistent patterns:

```python
# Some use "or" pattern (allows injection but creates default)
self._db = db or Database()

# Some always create new instances
self._analyzer = Analyzer(db=self._db)

# Some use singletons directly
self._settings = Settings()
```

**Issues:**
- Hard to test (can't easily mock dependencies)
- Hidden dependencies (caller doesn't know what's needed)
- Multiple initialization patterns confusing

**Fix:** Use proper DI container or factory pattern:
```python
@dataclass
class PlannerServices:
    db: Database
    broker: Broker
    portfolio: Portfolio
    analyzer: Analyzer
    settings: Settings
    currency: Currency
    ml_predictor: MLPredictor

class Planner:
    def __init__(self, services: PlannerServices):
        self._services = services
```

**Effort:** 1-2 days
**Impact:** Medium - Testability

---

### 5. Duplicate Cash/Fee Logic

**Problem:** Cash balance calculation exists in multiple places:

**Location 1: portfolio.py:109-128**
```python
async def get_cash_balances(self, include_simulated: bool = False) -> dict[str, float]:
    """Get cash balances from broker with optional simulated cash override."""
    # Implementation with currency conversion
```

**Location 2: tasks.py:444-540**
```python
async def trading_balance_fix_task():
    """Balance fix task with currency conversion logic."""
    # Independent implementation
```

**Fix:** Create a `CashManager` service:
```python
# sentinel/cash.py
@dataclass
class CashBalances:
    balances: dict[str, float]  # By currency
    total_eur: float

class CashManager:
    def __init__(self, db: Database, broker: Broker, currency: Currency):
        self._db = db
        self._broker = broker
        self._currency = currency

    async def get_balances(self, include_simulated: bool = False) -> CashBalances
    async def ensure_balance(self, currency: str, amount: float) -> bool
    async def convert(self, from_curr: str, to_curr: str, amount: float) -> ConversionResult
    async def get_total_eur(self) -> float
```

**Effort:** 4 hours
**Impact:** Medium - DRY principle

---

### 6. Empty API Routers Directory

**Current State:** `sentinel/api/routers/` exists but is empty

**Problem:** Directory structure implies intent but no implementation. Creates confusion for new developers.

**Fix Options:**
- Option A: Populate with extracted routers (recommended - see Issue #1)
- Option B: Remove directory to avoid confusion

**Effort:** 1 hour (if removing) or 2-3 days (if implementing)
**Impact:** Low - Code clarity

---

## Medium-Priority Issues

### 7. ML Module Organization

**State:** 2,999 lines across 7 files - actually well-organized!

| Module | Lines | Purpose |
|--------|-------|---------|
| ml_ensemble.py | 1,037 | Neural network + XGBoost ensemble |
| ml_features.py | 461 | Feature definitions and extraction |
| ml_monitor.py | 380 | Per-symbol performance tracking |
| ml_trainer.py | 380 | Training data generation |
| ml_predictor.py | 225 | Production prediction with caching |
| ml_retrainer.py | 263 | Weekly retraining pipeline |
| ml_reset.py | 253 | ML reset functionality |

**Minor Issue:** `ml_ensemble.py` (1,037 lines) is getting large

**Potential Split:**
```
sentinel/ml/
├── __init__.py
├── models/
│   ├── neural_network.py
│   ├── xgboost_model.py
│   └── ensemble.py
├── features.py
├── training.py
├── prediction.py
├── monitoring.py
└── reset.py
```

**Effort:** 4 hours
**Impact:** Low - Current structure is functional

---

### 8. Database Migration Pattern

**Current:** Inline migrations in `_apply_migrations()`
```python
# database/main.py
migrations = [
    ("market_id", "TEXT"),
    ("data", "TEXT"),
    # ... 9 more
]
```

**Problems:**
- Hard to track migration history
- No rollback capability
- Difficult to review changes

**Better:** Use migration files
```
sentinel/database/migrations/
├── 001_initial.sql
├── 002_add_market_id.sql
├── 003_add_ml_columns.sql
└── __init__.py  # Migration runner
```

**Migration Runner:**
```python
class MigrationManager:
    async def migrate(self):
        applied = await self._get_applied_migrations()
        for migration in self._load_pending():
            if migration.id not in applied:
                await self._apply(migration)
                await self._record(migration)
```

**Effort:** 4 hours
**Impact:** Low - Maintainability

---

### 9. Global State in app.py

```python
_scheduler = None
_led_controller = None
_led_task: asyncio.Task | None = None
```

**Problems:**
- Global mutable state
- Hard to test
- No encapsulation

**Fix:** Use FastAPI's `app.state` or a state class:
```python
class ApplicationState:
    def __init__(self):
        self.scheduler: AsyncIOScheduler | None = None
        self.led_controller: LEDController | None = None
        self.led_task: asyncio.Task | None = None

# In lifespan:
app.state.services = ApplicationState()
app.state.services.scheduler = await init_jobs(...)
```

**Effort:** 2 hours
**Impact:** Low - Testability

---

## Low-Priority (Code Quality)

### 10. Import Organization

Some files have inconsistent import ordering. Use `ruff check --select I --fix`

### 11. Type Coverage

Many functions lack return type annotations, especially in `app.py`.

---

## Dependency Analysis

### Current Import Graph

```
app.py
├── broker, cache, currency, database
├── jobs/*, portfolio, price_validator
├── security, settings, utils/fees
└── analyzer, ml_monitor, ml_retrainer, planner, regime_hmm  # Late imports

planner.py
├── analyzer, broker, currency, database
├── ml_features, ml_predictor, portfolio
├── price_validator, settings
└── utils/scoring, utils/positions, utils/fees  # Late imports

analyzer.py
├── cache, database, security
└── settings, regime_hmm  # Late imports (circular risk)

security.py
├── broker, database, settings
└── currency_exchange, currency  # Late imports
```

### Circular Import Risk Assessment

| Risk | Pair | Status |
|------|------|--------|
| High | analyzer.py ↔ regime_hmm.py | Currently broken by late import |
| Medium | analyzer.py → security.py → analyzer.py | Not circular yet |
| Low | planner.py → utils/* | Just disorganized |

---

## Recommended Refactoring Roadmap

### Phase 1: API Layer Extraction (Priority: Critical)
**Effort:** 2-3 days
**Impact:** High
**Files:** `sentinel/app.py`, `sentinel/api/routers/*`

1. Create router files with no logic changes
2. Move endpoint handlers one domain at a time
3. Update `app.py` to use `include_router()`
4. Ensure all tests pass after each router move

### Phase 2: Planner Decomposition (Priority: High)
**Effort:** 1-2 days
**Impact:** High
**Files:** `sentinel/planner.py`, `sentinel/planning/*`

1. Extract `TradeRecommendation` to `planning/models.py`
2. Create `IdealPortfolioCalculator` class
3. Create `RecommendationEngine` class
4. Create `CashConstraintApplier` class
5. Refactor `Planner` to use composition

### Phase 3: Service Layer Creation (Priority: High)
**Effort:** 1 day
**Impact:** Medium
**Files:** `sentinel/services/*`, `sentinel/cash.py`

1. Create `RegimeService` to break analyzer cycle
2. Create `CashManager` to consolidate cash logic
3. Move late imports to services

### Phase 4: Dependency Injection (Priority: Medium)
**Effort:** 1-2 days
**Impact:** Medium
**Files:** `sentinel/planner.py`, `sentinel/analyzer.py`

1. Define service dataclasses
2. Update constructors to accept services
3. Update callers to provide services

### Phase 5: Database Migrations (Priority: Low)
**Effort:** 4 hours
**Impact:** Low
**Files:** `sentinel/database/migrations/*`

1. Extract current migrations to SQL files
2. Create migration runner
3. Add migration tracking table

### Phase 6: Global State Cleanup (Priority: Low)
**Effort:** 2 hours
**Impact:** Low
**Files:** `sentinel/app.py`

1. Create `ApplicationState` class
2. Move globals to `app.state`
3. Update all references

---

## Architecture After Refactoring

```
┌─────────────────────────────────────────┐
│  API Layer (sentinel/api/routers/)      │
│  portfolio.py | securities.py | ml.py   │
└─────────────────────────────────────────┘
                   │
┌─────────────────────────────────────────┐
│  Service Layer (sentinel/services/)     │
│  planning/ | cash.py | regime.py        │
└─────────────────────────────────────────┘
                   │
┌─────────────────────────────────────────┐
│  Domain Layer                           │
│  Portfolio | Security | Analyzer        │
│  MLPredictor | Planner (orchestrator)   │
└─────────────────────────────────────────┘
                   │
┌─────────────────────────────────────────┐
│  Infrastructure Layer                   │
│  Database | Broker | Cache | Settings   │
└─────────────────────────────────────────┘
```

---

## Success Metrics

After refactoring:
- `app.py` < 300 lines (from 2,228)
- `planner.py` < 200 lines (from 1,216)
- Zero late imports (except in service layer)
- All routers < 300 lines each
- Test coverage maintained at 610+ tests
- No circular import risks

---

## Appendix: File Size Audit

| File | Lines | Category | Target |
|------|-------|----------|--------|
| app.py | 2,228 | God Module | < 300 |
| planner.py | 1,216 | God Class | < 200 |
| analyzer.py | 782 | Large | < 500 |
| ml_ensemble.py | 1,037 | Large | < 500 (if split) |
| ml_features.py | 461 | OK | - |
| security.py | 352 | OK | - |
| portfolio.py | 280 | OK | - |
| broker.py | 559 | OK | - |
| database/main.py | 1,100 | Large | < 500 (if migrations moved) |

Total Python lines: ~15,000
Target after refactoring: Same functionality, better organization
