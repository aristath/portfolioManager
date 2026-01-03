# Python-to-Go Migration Discrepancy Report
**Date:** 2026-01-02
**Python Commit:** e58839f3b8deaf2fc4ac12561fee67e4ea6a1dd6
**Go State:** Current (main branch + 28 recent commits)

## Summary Statistics
- **Modules Reviewed:** 10 of 10
- **Endpoint Migration:** 68 of 111 (61%)
- **Operational Capability:** ~40%
- **Critical Blockers:** 2 (Planning/Recommendations, Emergency Rebalancing)

---

## Incomplete Modules & Critical Gaps

### DIVIDENDS Module ✅ COMPLETE (100%)

**Completed**:
- ✅ DRIP background job implemented in Go (dividend_reinvestment.go, 446 lines)
- ✅ Comprehensive tests (178 lines, all passing)
- ✅ High-yield dividend reinvestment (>=3% yield)
- ✅ Low-yield dividend handling (pending bonuses)
- ✅ Yahoo Finance integration for prices and yields

**Impact:** Fully autonomous dividend reinvestment operational

---

### CASH FLOWS Module ⚠️ NEEDS VERIFICATION

**Status:** Completely rewritten (2,034 lines added)

**Original Bugs** (need verification in new code):
1. Deposit currency bug (was passing `created.Currency` instead of `"EUR"`)
2. Missing fallback to core bucket
3. Deduplication performance (O(N) vs O(1))

**Action Required:** Test new implementation

---

### TRADING Module ⚠️ NEEDS VERIFICATION

**Status:** Service layer added (103 lines)

**Original Issues** (need verification):
1. No trade recording to database
2. Missing 7-layer validation
3. Missing safety checks

**Action Required:** Verify service includes validation and recording

---

### OPTIMIZATION Module ⚠️ MICROSERVICE-DEPENDENT

**Status:** Functional but requires external service

**Dependencies**:
- ⚠️ pyportfolioopt microservice (critical dependency)
- Recently added: Real-time price fetching, cash balance, dividend bonuses

**Impact:** Cannot run optimization without pyportfolioopt microservice

---

### SCORING Module ✅ COMPLETE (Enhanced)

**Status:** 100% complete + enhanced with windfall scorer

---

### PLANNING Module ⚠️ HANDLERS EXIST, ROUTES NOT REGISTERED

**Status:** Domain complete, handlers implemented, routes not wired

**Architecture:**
- ✅ Opportunities module (1,888 LOC) - Complete
- ✅ Sequences module (1,786 LOC) - Complete
- ⚠️ Evaluation delegated to evaluator-go microservice
- ✅ **Handlers implemented** but commented out in server.go (line 167)

**Implemented but not registered**:
- ✅ Planner config CRUD (list, get, create, update, delete)
- ✅ Config validation endpoint
- ✅ Config history endpoint
- ✅ Batch generation (POST /api/planning/batch)
- ✅ Plan execution (POST /api/planning/execute)
- ✅ Status checking (GET /api/planning/status)
- ✅ SSE streaming (GET /api/planning/stream)

**Missing**:
- ❌ GET /api/trades/recommendations (handler not found)
- ❌ POST /api/trades/recommendations/execute (handler not found)
- ❌ Route registration (commented out in server.go)

**Impact:** **BLOCKS AUTONOMOUS TRADING** - Handlers ready but routes not accessible via HTTP

---

### UNIVERSE Module ✅ NEARLY COMPLETE (99%)

**Status:** Functional with minor proxy dependencies

**Dependencies**:
- ⚠️ Create operations proxy to Python (Yahoo Finance integration)
- ⚠️ Some refresh operations proxied

**Impact:** Core functionality works, complex writes require Python

---

## Python-Only Modules (NOT Migrated)

### SYSTEM Module ✅ COMPLETE (100%)

**Completed**:
- ✅ Status & monitoring (7 endpoints)
- ✅ Logs (3 endpoints)
- ✅ Background jobs: sync_cycle, health_check, dividend_reinvestment
- ✅ Sync operation triggers (4 endpoints registered):
  - POST /api/system/sync/prices
  - POST /api/system/sync/historical
  - POST /api/system/sync/rebuild-universe
  - POST /api/system/sync/securities-data
  - POST /api/system/sync/portfolio
  - POST /api/system/sync/daily-pipeline
  - POST /api/system/sync/recommendations
- ✅ Maintenance job triggers (5 endpoints):
  - POST /api/system/jobs/sync-cycle
  - POST /api/system/jobs/weekly-maintenance
  - POST /api/system/jobs/dividend-reinvestment
  - POST /api/system/jobs/planner-batch
  - POST /api/system/maintenance/daily
- ✅ Lock management:
  - POST /api/system/locks/clear

**Missing** (low priority):
- ❌ GET /api/system/deploy/status (deployment tooling)
- ❌ POST /api/system/deploy/trigger (deployment tooling)

**Impact:** All core system operations accessible, only deployment tooling missing

---

### ANALYTICS Module ⚠️ PARTIAL (80%)

**Completed** (implemented in portfolio module):
- ✅ Portfolio reconstruction from trades (attribution.go)
- ✅ Performance metrics: Sharpe, Sortino, Calmar, Volatility, Max Drawdown (service.go)
- ✅ Attribution analysis by country and industry (attribution.go)
- ✅ Market regime detection (market_regime.go, 200 lines + comprehensive tests)

**Missing**:
- ❌ Position risk metrics (beta, correlation matrix)

**Impact:** Core analytics complete, position-level risk metrics unavailable (low priority)

---

### REBALANCING Module ✅ COMPLETE (commits 8c5b1c70, 1dba55ec)

**Implemented in Go (trader-go)**:
- ✅ Portfolio drift detection (triggers.go with 7 tests)
- ✅ Cash accumulation detection (triggers.go with tests)
- ✅ Automatic rebalancing triggers (event-driven)
- ✅ Rebalancing patterns (rebalance.go, deep_rebalance.go)
- ✅ Opportunity calculators (rebalance_buys.go, rebalance_sells.go)
- ✅ NegativeBalanceRebalancer (negative_balance_rebalancer.go, 95 LOC)
- ✅ RebalancingService (service.go, 133 LOC)
- ✅ HTTP handlers (handlers.go, 237 LOC, 4 endpoints)
- ✅ Routes registered in server.go

**Status:** Fully migrated with 12 passing tests

---

### SATELLITES Module ⚠️ NEARLY COMPLETE (95%)

**Completed** (commits 44379a86 through 6adcb65f):
- ✅ Domain models (complete with tests)
- ✅ Repositories (complete with 44 tests)
- ✅ Domain logic (complete with 108+ tests)
  - Aggression calculator
  - Win cooldown
  - Graduated reawakening
  - Strategy presets (4 strategies)
  - Parameter mapper
- ✅ **Services layer (6 services, ~2,400 lines)**
  - BucketService (lifecycle, hibernation, circuit breaker)
  - BalanceService (deposits, transfers, reallocation)
  - ReconciliationService (balance sync, auto-correction)
  - DividendRouter (3 routing modes)
  - PerformanceMetrics (Sharpe, Sortino, Calmar, win rate)
  - MetaAllocator (performance-based rebalancing)
- ✅ **API endpoints (21 endpoints, 820 lines)**
  - Bucket CRUD (4 endpoints)
  - Lifecycle management (4 endpoints)
  - Settings (2 endpoints)
  - Balance management (4 endpoints)
  - Transactions (1 endpoint)
  - Reconciliation (3 endpoints)
  - Allocation settings (2 endpoints)
  - Strategy presets (2 endpoints)
- ✅ **Database schema (satellites.db with 5 tables)**
- ✅ **Event system (13 satellite event types)**

**Missing**:
- ❌ Background jobs (maintenance, reconciliation, evaluation)
- ❌ Planner integration

**Impact:** API operational, multi-bucket strategies available via API, background automation pending

---

### SETTINGS Module ❌ MISSING (0%)

**Missing All 9 Endpoints**:
- `GET /api/settings`
- `PUT /api/settings/{key}`
- `POST /api/settings/restart-service`
- `POST /api/settings/restart`
- `POST /api/settings/reset-cache`
- `GET /api/settings/cache-stats`
- `POST /api/settings/reschedule-jobs`
- `GET /api/settings/trading-mode`
- `POST /api/settings/trading-mode`

**Impact:** Cannot modify system settings or trading mode

---

### CHARTS Module ❌ MISSING (0%)

**Missing Both Endpoints**:
- `GET /api/charts/sparklines`
- `GET /api/charts/securities/{isin}`

**Impact:** Dashboard sparklines unavailable

---

### GATEWAY Module ✅ N/A

**Status:** Empty stub in Python, no action needed

---

## Critical Blockers

### P0 - BLOCKING AUTONOMOUS OPERATION

**1. Planning/Recommendations Module (12 of 14 handlers ready)**
- **Impact:** Blocks autonomous trading
- **Estimated Work:** 3-5 days
- **Status:** Handlers implemented but routes commented out in server.go
- **Quick win:** Uncomment route registration + implement 2 missing endpoints:
  - ❌ `GET /api/trades/recommendations` - Get recommendations (needs handler)
  - ❌ `POST /api/trades/recommendations/execute` - Execute (needs handler)
  - ✅ All planner config endpoints already implemented
  - ✅ Status, batch generation, execution, SSE streaming ready

**2. Emergency Rebalancing Migration from Python**
- **Impact:** Emergency handler works in Python, needs Go migration
- **Estimated Work:** 1 week
- **Status:** Fully implemented in Python, needs porting to Go
- **Components to migrate:**
  - Negative balance auto-correction (3-step workflow)
  - Currency exchange orchestration
  - Emergency position sales
  - Sync cycle integration (currently stubbed)

---

## High Priority (P1)

**3. System Job Triggers (11 handlers ready, 4 missing)**
- **Impact:** Cannot manually trigger operations
- **Estimated Work:** 1-2 days
- **Status:** 11 handlers implemented but not registered in routes
- **Quick win:** Register existing handlers in setupSystemRoutes()
- **Need implementation:** 4 endpoints (sync/portfolio, sync/daily-pipeline, sync/recommendations, locks/clear)

**4. Market Regime Detection**
- **Impact:** Broken feature (referenced in Go planning code)
- **Estimated Work:** 2-3 days
- **URGENT:** Already referenced but not implemented

---

## Medium Priority (P2)

**5. Satellites Module Completion**
- **Impact:** Background automation and planner integration remaining
- **Estimated Work:** 3-5 days
- Background jobs (maintenance, reconciliation) + planner integration

**6. Settings Module (9 endpoints)**
- **Impact:** Cannot modify settings
- **Estimated Work:** 3-5 days

**7. Analytics Module Completion**
- **Impact:** Missing market regime detection and position risk metrics (beta, correlation)
- **Estimated Work:** 1 week

**8. Charts Module (2 endpoints)**
- **Impact:** Dashboard enhancements
- **Estimated Work:** 2-3 days

---

## Verification Required

### Cash Flows - Needs Testing
- [ ] Verify deposit currency handling fixed
- [ ] Verify fallback to core bucket fixed
- [ ] Verify deduplication performance fixed

### Trading - Needs Testing
- [ ] Verify trade recording to database works
- [ ] Verify 7-layer validation implemented
- [ ] Verify safety checks in place

---

## Migration Roadmap

### Phase 1: Unblock Auto-Trading (1-2 weeks) - P0
1. Planning/Recommendations module - 3-5 days
   - Uncomment route registration
   - Implement 2 missing recommendation endpoints
2. Emergency Rebalancing migration - 1 week
   - Port Python handler to Go
   - Integrate with sync cycle
3. Verify cash flows + trading fixes - 2-3 days

**Deliverable:** Autonomous trading operational

### Phase 2: Operational Control (1 week) - P1
1. System job triggers - 1-2 days
   - Register 11 existing handlers
   - Implement 4 missing endpoints
2. Settings module (9 endpoints) - 3-5 days
3. Market regime detection - 2-3 days

**Deliverable:** Full manual control + settings management

### Phase 3: Feature Complete (1-2 weeks) - P2
1. Satellites completion (jobs + planner integration) - 3-5 days
2. Analytics module completion (regime detection, beta/correlation) - 1 week
3. Charts module - 2-3 days

**Deliverable:** 100% feature parity

### Phase 4: Independence (1 week) - P3
1. Remove universe proxies - 1 week
2. Documentation and optimization - 2-3 days

**Deliverable:** Zero Python dependencies

---

## Total Effort Estimate

**3-5 weeks to full Python independence**

| Phase | Weeks | Priority |
|-------|-------|----------|
| Phase 1 | 1-2 | P0 - Critical |
| Phase 2 | 1 | P1 - High |
| Phase 3 | 1-2 | P2 - Medium |
| Phase 4 | 1 | P3 - Low |

---

## Recommendation

**PROCEED WITH PHASE 1** - Implement Planning/Recommendations module and Emergency Rebalancing to unblock autonomous trading.

**Current Status:** 61% endpoint migration, ~40% operational capability

**Key Discovery:** Many handlers already implemented but routes not registered - quick wins available

**After Phase 1:** Autonomous trading operational (1-2 weeks vs original 3-4 weeks)
**After All Phases:** Full independence from Python (3-5 weeks total, down from 8-12 weeks)

---

*Last Updated: 2026-01-02*
