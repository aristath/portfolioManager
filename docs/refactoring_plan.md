# Sentinel Refactoring Implementation Plan

**Document:** Architectural Review Findings
**Plan Version:** 1.0
**Estimated Duration:** 5-7 days
**Risk Level:** Medium (requires careful testing)

---

## Overview

This plan addresses the critical architectural issues identified in `docs/architectural_review.md`. The refactoring is organized into 6 phases, each building on the previous.

**Key Principles:**
1. **No logic changes** - Only move and reorganize code
2. **Test-driven** - All 610 tests must pass after each phase
3. **Incremental** - Each phase is independently mergeable
4. **Documented** - Each change is tracked in commit messages

---

## Phase 1: API Router Extraction (Days 1-3)

**Goal:** Break `app.py` (2,228 lines, 73 routes) into focused routers.

### 1.1 Create Router Infrastructure

**Files to Create:**
```
sentinel/api/
├── __init__.py              # Export create_app() factory
├── dependencies.py          # Shared FastAPI dependencies
└── routers/
    ├── __init__.py
    ├── settings.py          # /api/settings, /api/led
    ├── portfolio.py         # /api/portfolio, /api/allocation*
    ├── securities.py        # /api/securities*, /api/prices*
    ├── trading.py           # /api/trades*, buy/sell
    ├── analysis.py          # /api/scores*, /api/unified
    ├── jobs.py              # /api/jobs*, /api/backup
    ├── ml.py                # /api/ml/* endpoints
    └── system.py            # /api/health, /api/cache, /api/backtest
```

**Tasks:**
- [ ] Create `sentinel/api/__init__.py` with app factory
- [ ] Create `sentinel/api/dependencies.py` with common deps (db, broker, settings)
- [ ] Create each router file with skeleton structure

**Verification:**
```bash
python -c "from sentinel.api.routers import settings, portfolio, ..."
```

### 1.2 Extract Settings & LED Routes

**Routes to Move:**
- `GET /api/settings` → `settings.py`
- `PUT /api/settings/{key}` → `settings.py`
- `GET /api/led/status` → `settings.py`
- `PUT /api/led/enabled` → `settings.py`
- `POST /api/led/refresh` → `settings.py`

**Tasks:**
- [ ] Copy handlers to `routers/settings.py`
- [ ] Update imports in new file
- [ ] Add router to main app with `app.include_router()`
- [ ] Comment out original handlers in `app.py`
- [ ] Run tests

**Verification:**
```bash
pytest tests/test_settings.py -v
```

### 1.3 Extract Portfolio Routes

**Routes to Move:**
- `GET /api/portfolio`
- `POST /api/portfolio/sync`
- `GET /api/portfolio/allocations`
- `GET /api/portfolio/pnl-history`
- `GET /api/allocation-targets`
- `PUT /api/allocation-targets/{target_type}/{name}`
- `GET /api/allocation/current`
- `GET /api/allocation/targets`
- `GET /api/allocation/available-geographies`
- `GET /api/allocation/available-industries`
- `PUT /api/allocation/targets/geography`
- `PUT /api/allocation/targets/industry`

**Tasks:**
- [ ] Move handlers to `routers/portfolio.py`
- [ ] Move helper `_backfill_portfolio_snapshots`
- [ ] Update imports
- [ ] Add router to main app
- [ ] Run tests

**Verification:**
```bash
pytest tests/ -k portfolio -v
```

### 1.4 Extract Securities Routes

**Routes to Move:**
- `GET /api/securities`
- `POST /api/securities`
- `DELETE /api/securities/{symbol}`
- `GET /api/securities/aliases`
- `GET /api/securities/{symbol}`
- `PUT /api/securities/{symbol}`
- `GET /api/securities/{symbol}/prices`
- `POST /api/securities/{symbol}/sync-prices`
- `POST /api/prices/sync-all`

**Tasks:**
- [ ] Move handlers to `routers/securities.py`
- [ ] Update imports
- [ ] Add router to main app
- [ ] Run tests

### 1.5 Extract Trading Routes

**Routes to Move:**
- `GET /api/trades`
- `POST /api/trades/sync`
- `GET /api/cashflows`
- `POST /api/cashflows/sync`
- `POST /api/securities/{symbol}/buy`
- `POST /api/securities/{symbol}/sell`

**Tasks:**
- [ ] Move handlers to `routers/trading.py`
- [ ] Update imports
- [ ] Add router to main app
- [ ] Run tests

### 1.6 Extract Analysis Routes

**Routes to Move:**
- `POST /api/scores/calculate`
- `GET /api/unified`

**Tasks:**
- [ ] Move handlers to `routers/analysis.py`
- [ ] Update imports
- [ ] Add router to main app
- [ ] Run tests

### 1.7 Extract Jobs Routes

**Routes to Move:**
- All `/api/jobs/*` endpoints
- `/api/backup/*` endpoints

**Tasks:**
- [ ] Move handlers to `routers/jobs.py`
- [ ] Update imports
- [ ] Add router to main app
- [ ] Run tests

### 1.8 Extract ML Routes

**Routes to Move:**
- All `/api/ml/*` endpoints
- All `/api/regime/*` endpoints

**Tasks:**
- [ ] Move handlers to `routers/ml.py`
- [ ] Update imports
- [ ] Add router to main app
- [ ] Run tests

### 1.9 Extract System Routes

**Routes to Move:**
- `GET /api/health`
- `GET /api/health/detailed`
- Cache endpoints
- Backtest endpoints
- Exchange rate endpoints
- Market endpoints

**Tasks:**
- [ ] Move handlers to `routers/system.py`
- [ ] Update imports
- [ ] Add router to main app
- [ ] Run tests

### 1.10 Cleanup app.py

**Tasks:**
- [ ] Remove all original handler functions (now in routers)
- [ ] Keep only:
  - `lifespan()` context manager
  - `create_app()` factory
  - Static file mounting
  - CORS middleware setup
- [ ] Run full test suite

**Target:** `app.py` < 300 lines

**Verification:**
```bash
pytest tests/ -v
wc -l sentinel/app.py  # Should be ~200-300
```

---

## Phase 2: Planner Decomposition (Days 4-5)

**Goal:** Break `Planner` god class (1,216 lines) into focused components.

### 2.1 Create Planning Module Structure

**Files to Create:**
```
sentinel/planning/
├── __init__.py              # Export Planner facade
├── models.py                # TradeRecommendation dataclass
├── ideal_calculator.py      # calculate_ideal_portfolio()
├── recommendation_engine.py # get_recommendations()
├── constraint_applier.py    # _apply_cash_constraint()
└── utils.py                 # Helper functions (_calculate_priority, etc.)
```

### 2.2 Extract Models

**Tasks:**
- [ ] Move `TradeRecommendation` to `planning/models.py`
- [ ] Update imports in `planner.py`
- [ ] Run tests

### 2.3 Extract IdealPortfolioCalculator

**Move to `ideal_calculator.py`:**
- `calculate_ideal_portfolio()`
- `_classic_allocation()`
- `_calculate_diversification_score()`

**Tasks:**
- [ ] Create `IdealPortfolioCalculator` class
- [ ] Move methods with minimal changes
- [ ] Update imports
- [ ] Run tests

### 2.4 Extract RecommendationEngine

**Move to `recommendation_engine.py`:**
- `get_recommendations()`
- `_calculate_priority()`
- `_generate_buy_reason()`
- `_generate_sell_reason()`
- `_check_cooloff_violation()`

**Tasks:**
- [ ] Create `RecommendationEngine` class
- [ ] Move methods
- [ ] Update imports
- [ ] Run tests

### 2.5 Extract CashConstraintApplier

**Move to `constraint_applier.py`:**
- `_apply_cash_constraint()`
- `_get_deficit_sells()`
- `_generate_deficit_sells()`
- `_calculate_transaction_cost()`

**Tasks:**
- [ ] Create `CashConstraintApplier` class
- [ ] Move methods
- [ ] Update imports
- [ ] Run tests

### 2.6 Refactor Planner Facade

**Update `Planner` class:**
```python
class Planner:
    def __init__(
        self,
        calculator: IdealPortfolioCalculator | None = None,
        engine: RecommendationEngine | None = None,
        constraint_applier: CashConstraintApplier | None = None,
    ):
        self._calculator = calculator or IdealPortfolioCalculator()
        self._engine = engine or RecommendationEngine()
        self._constraint_applier = constraint_applier or CashConstraintApplier()
```

**Tasks:**
- [ ] Update `__init__` to use composition
- [ ] Delegate methods to components
- [ ] Update all imports
- [ ] Run full test suite

**Target:** `planner.py` < 200 lines, planning/ module ~1,200 lines total

---

## Phase 3: Service Layer Creation (Day 6)

**Goal:** Break circular import risks and consolidate duplicate logic.

### 3.1 Create Services Directory

**Files to Create:**
```
sentinel/services/
├── __init__.py
├── regime.py                # Break analyzer ↔ regime_hmm cycle
└── cash.py                  # Consolidate cash logic
```

### 3.2 Create RegimeService

**Purpose:** Isolate `RegimeDetector` import to break circular dependency.

**File:** `sentinel/services/regime.py`

```python
class RegimeService:
    """Service for regime detection to avoid circular imports."""

    async def adjust_expected_return(
        self,
        symbol: str,
        base_return: float
    ) -> float:
        from sentinel.regime_hmm import RegimeDetector  # Late import OK here
        # ... adjustment logic
```

**Tasks:**
- [ ] Create `RegimeService` class
- [ ] Move regime adjustment logic from `analyzer.py`
- [ ] Update `analyzer.py` to use service
- [ ] Run tests

### 3.3 Create CashManager

**Purpose:** Consolidate cash balance logic from `portfolio.py` and `tasks.py`.

**File:** `sentinel/services/cash.py`

```python
@dataclass
class CashBalances:
    by_currency: dict[str, float]
    total_eur: float

class CashManager:
    def __init__(self, db: Database, broker: Broker, currency: Currency):
        self._db = db
        self._broker = broker
        self._currency = currency

    async def get_balances(self, include_simulated: bool = False) -> CashBalances
    async def convert(self, from_curr: str, to_curr: str, amount: float) -> float
    async def ensure_balance(self, currency: str, min_amount: float) -> bool
```

**Tasks:**
- [ ] Create `CashManager` class
- [ ] Extract logic from `portfolio.py:get_cash_balances()`
- [ ] Extract logic from `tasks.py:trading_balance_fix_task()`
- [ ] Update callers to use `CashManager`
- [ ] Run tests

### 3.4 Fix Late Imports in Planner

**File:** `sentinel/planner.py`

**Current late imports:**
- Line 177: `from sentinel.utils.scoring import adjust_score_for_conviction`
- Line 281: `from sentinel.utils.positions import PositionCalculator`
- Line 387: `from sentinel.utils.scoring import adjust_score_for_conviction`
- Line 647: `from sentinel.utils.fees import FeeCalculator`

**Tasks:**
- [ ] Move imports to top of file
- [ ] Verify no circular imports introduced
- [ ] Run tests

---

## Phase 4: Dependency Injection (Day 6-7)

**Goal:** Make components testable with proper DI.

### 4.1 Create Service Dataclasses

**File:** `sentinel/services/container.py`

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
    feature_extractor: FeatureExtractor

@dataclass
class PortfolioServices:
    db: Database
    broker: Broker
    settings: Settings
    currency: Currency
```

### 4.2 Update Planner DI

**Tasks:**
- [ ] Update `Planner.__init__()` to accept `PlannerServices`
- [ ] Update all callers to construct services
- [ ] Run tests

### 4.3 Update Portfolio DI

**Tasks:**
- [ ] Update `Portfolio.__init__()` to accept `PortfolioServices`
- [ ] Update all callers
- [ ] Run tests

### 4.4 Update Analyzer DI

**Tasks:**
- [ ] Update `Analyzer.__init__()` to accept dependencies explicitly
- [ ] Update all callers
- [ ] Run tests

---

## Phase 5: Database Migrations (Day 7)

**Goal:** Move inline migrations to file-based system.

### 5.1 Create Migration Infrastructure

**Files to Create:**
```
sentinel/database/migrations/
├── __init__.py              # Migration runner
├── 001_initial.sql          # Current schema
├── 002_add_market_id.sql
├── 003_add_securities_data.sql
├── 004_add_ml_columns.sql
└── ...
```

### 5.2 Create Migration Runner

**File:** `sentinel/database/migrations/__init__.py`

```python
class MigrationManager:
    def __init__(self, conn: aiosqlite.Connection):
        self._conn = conn

    async def migrate(self):
        await self._ensure_migrations_table()
        applied = await self._get_applied()

        for migration in self._load_all():
            if migration.id not in applied:
                await self._apply(migration)
                await self._record(migration)
```

### 5.3 Extract Current Migrations

**Tasks:**
- [ ] Create SQL files for each migration in `_apply_migrations()`
- [ ] Add migration tracking table schema
- [ ] Replace `_apply_migrations()` with `MigrationManager`
- [ ] Run tests

---

## Phase 6: Global State Cleanup (Day 7)

**Goal:** Remove global mutable state from `app.py`.

### 6.1 Create ApplicationState

**File:** `sentinel/api/state.py`

```python
class ApplicationState:
    """Encapsulates application-wide state."""

    def __init__(self):
        self.scheduler: AsyncIOScheduler | None = None
        self.led_controller: LEDController | None = None
        self.led_task: asyncio.Task | None = None
        self.market_checker: BrokerMarketChecker | None = None
```

### 6.2 Update Lifespan

**Tasks:**
- [ ] Create state instance in `lifespan()`
- [ ] Store in `app.state.services`
- [ ] Update all references from globals to `app.state.services`
- [ ] Run tests

---

## Testing Strategy

### Per-Phase Testing

**After each phase:**
```bash
# Run full test suite
pytest tests/ -v

# Check for import errors
python -c "from sentinel.app import app"

# Verify line counts
cd sentinel && find . -name "*.py" -exec wc -l {} + | sort -n
```

### Integration Testing

**After Phase 1 (API extraction):**
```bash
# Test each router
pytest tests/test_settings.py -v
pytest tests/test_api_job_schedules.py -v
pytest tests/test_api_ml_reset.py -v
pytest tests/jobs/ -v
```

**After Phase 2 (Planner):**
```bash
pytest tests/test_planner.py -v
```

### Regression Testing

**Final verification:**
```bash
# Full test suite
pytest tests/ -v --tb=short

# Smoke test
python -c "
from sentinel.app import app
from sentinel.api.routers import settings, portfolio, securities
print('All imports OK')
"
```

---

## Rollback Plan

**If issues arise:**

1. **Git Revert:** Each phase is a separate commit, can revert individually
2. **Feature Flags:** Keep old code commented until phase is stable
3. **Gradual Rollout:** Move one router at a time, verify, then remove old

**Emergency Rollback Commands:**
```bash
# Revert last phase
git revert HEAD

# Or restore from backup branch
git checkout -b refactor-backup main
git checkout main
git reset --hard refactor-backup
```

---

## Success Criteria

| Metric | Before | After |
|--------|--------|-------|
| app.py lines | 2,228 | < 300 |
| planner.py lines | 1,216 | < 200 |
| Late imports | 7 | 0 (in core code) |
| API routers | 0 | 8 |
| Test pass rate | 610/610 | 610/610 |
| Circular import risk | High | Low |

---

## Timeline

| Phase | Duration | Cumulative |
|-------|----------|------------|
| Phase 1: API Routers | 3 days | Day 3 |
| Phase 2: Planner | 2 days | Day 5 |
| Phase 3: Services | 1 day | Day 6 |
| Phase 4: DI | 1 day | Day 7 |
| Phase 5: Migrations | 0.5 day | Day 7.5 |
| Phase 6: State | 0.5 day | Day 8 |

**Buffer:** +1 day for unexpected issues
**Total:** 5-7 business days

---

## Checklist Summary

- [ ] Phase 1: Extract all 8 API routers
- [ ] Phase 2: Break Planner into 4 components
- [ ] Phase 3: Create RegimeService and CashManager
- [ ] Phase 4: Implement DI container pattern
- [ ] Phase 5: File-based database migrations
- [ ] Phase 6: Remove global state
- [ ] All 610 tests pass
- [ ] No new warnings from ruff/pyright
- [ ] Documentation updated

---

## Notes

1. **Preserve Git History:** Use `git mv` when moving files to preserve blame
2. **Commit Messages:** Use conventional commits:
   - `refactor(api): extract portfolio router`
   - `refactor(planning): decompose Planner into components`
3. **Code Review:** Each phase should be reviewed before proceeding
4. **Staging:** Test on staging environment before production deploy
