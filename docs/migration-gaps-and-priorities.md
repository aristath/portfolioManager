# Python-to-Go Migration: Gaps & Priorities
**Date**: 2026-01-02
**Status**: 42% Complete (47 of 111 endpoints migrated)

---

## Executive Summary

**What Works**: 7 modules fully complete (allocation, cash_flows, display, dividends, optimization, scoring, trading)
**Critical Blockers**: Planning/Recommendations module + Emergency rebalancing
**Estimated Remaining Effort**: 8-12 weeks to full independence

---

## Critical Blockers (P0)

### 1. Planning/Recommendations Module - ❌ MISSING
**Impact**: Blocks autonomous trading

**Missing Endpoints (14)**:
- `GET /api/trades/recommendations` - Get trade recommendations
- `POST /api/trades/recommendations/execute` - Execute recommendation
- `GET /api/planner/status` - Planner progress
- `POST /api/planner/regenerate-sequences` - Force regeneration
- `GET /api/planner/status/stream` - SSE status stream
- `GET /api/planners` - List planner configs
- `GET /api/planners/{config_id}` - Get config
- `POST /api/planners` - Create config
- `PUT /api/planners/{config_id}` - Update config
- `DELETE /api/planners/{config_id}` - Delete config
- `POST /api/planners/validate` - Validate TOML
- `GET /api/planners/{config_id}/history` - Config history
- `POST /api/planners/{config_id}/apply` - Apply config
- `GET /api/trades/recommendations/stream` - SSE recommendation updates

**Estimated Work**: 3 weeks
- Recommendations API: 3-4 days
- Recommendation execution: 3-4 days
- Planner status: 1 day
- Planner config CRUD: 1 week
- Integration testing: 1 week

---

### 2. Emergency Rebalancing - ❌ MISSING
**Impact**: System can deadlock with negative currency balances

**Missing Components**:
- Negative balance auto-correction (3-step recovery)
- Currency exchange orchestration
- Portfolio drift detection
- Emergency position sales
- Minimum currency reserve management (€5 per currency)

**Estimated Work**: 1-2 weeks
- Port negative_balance_rebalancer.py (726 lines)
- Implement 3-step workflow (exchange → sell → exchange)
- Integration testing with scenarios

---

## High Priority (P1)

### 3. System Module - Job Triggers - ⚠️ 15 MISSING ENDPOINTS
**Impact**: Cannot manually trigger background operations

**Status**: 10 of 25 endpoints (40% complete)
- ✅ Status & monitoring (7 endpoints)
- ✅ Logs (3 endpoints)
- ❌ Sync operations (7 endpoints)
- ❌ Maintenance jobs (5 endpoints)
- ❌ Operations (3 endpoints)

**Missing Endpoints**:

**Sync Operations**:
- `POST /api/system/sync/portfolio` - Manual portfolio sync
- `POST /api/system/sync/prices` - Manual price sync
- `POST /api/system/sync/historical` - Historical data sync
- `POST /api/system/sync/rebuild-universe` - Rebuild universe
- `POST /api/system/sync/securities-data` - Securities data sync
- `POST /api/system/sync/daily-pipeline` - Daily pipeline
- `POST /api/system/sync/recommendations` - Manual recommendations

**Maintenance Jobs**:
- `POST /api/system/maintenance/daily` - Daily maintenance
- `POST /api/system/jobs/sync-cycle` - Trigger sync cycle
- `POST /api/system/jobs/weekly-maintenance` - Weekly maintenance
- `POST /api/system/jobs/dividend-reinvestment` - Trigger DRIP
- `POST /api/system/jobs/planner-batch` - Trigger planner

**Operations**:
- `POST /api/system/locks/clear` - Clear stuck locks
- `GET /api/system/deploy/status` - Deployment status
- `POST /api/system/deploy/trigger` - Trigger deployment

**Estimated Work**: 3-5 days (wrappers around existing jobs)

---

### 4. Analytics Module - ❌ MISSING
**Impact**: Market regime detection broken (referenced in Go code), no performance reporting

**Missing Components**:
- Market regime detection (bull/bear/sideways) - **URGENT** (referenced but not implemented)
- Portfolio reconstruction from trades
- Performance metrics (Sharpe, Sortino, max drawdown)
- Attribution analysis (by country, industry, factor)
- Position risk metrics (beta, correlation)

**Estimated Work**: 2-3 weeks
- Market regime detection: 2-3 days (CRITICAL)
- Performance metrics: 1 week
- Attribution analysis: 1 week
- PyFolio integration or Go alternative: TBD

---

## Medium Priority (P2)

### 5. Satellites Module - ⚠️ 70% COMPLETE
**Impact**: Multi-bucket portfolio strategies not available

**Status**:
- ✅ Domain models (complete with tests)
- ✅ Repositories (complete with 44 tests)
- ✅ Domain logic (complete with 108+ tests)
- ❌ Services (BucketService, BalanceService, MetaAllocator)
- ❌ API endpoints (24 endpoints)
- ❌ Background jobs (maintenance, reconciliation, evaluation)

**Missing Endpoints (24)**:
- Bucket CRUD (4)
- Lifecycle management (4)
- Settings (4)
- Balances (4)
- Transactions (1)
- Reconciliation (3)
- Allocation settings (2)
- Plus 2 additional endpoints

**Estimated Work**: 2 weeks
- Services layer: 1 week
- API registration: 3 days
- Background jobs: 3-4 days

---

### 6. Settings Module - ❌ MISSING
**Impact**: Cannot modify system settings, trading mode, cache

**Missing Endpoints (9)**:
- `GET /api/settings` - Get all settings
- `PUT /api/settings/{key}` - Update setting
- `POST /api/settings/restart-service` - Restart systemd
- `POST /api/settings/restart` - Restart app
- `POST /api/settings/reset-cache` - Reset cache
- `GET /api/settings/cache-stats` - Cache stats
- `POST /api/settings/reschedule-jobs` - Reschedule jobs
- `GET /api/settings/trading-mode` - Get trading mode
- `POST /api/settings/trading-mode` - Set trading mode

**Estimated Work**: 3-5 days (simple CRUD with validation)

---

### 7. Charts Module - ❌ MISSING
**Impact**: Dashboard sparklines and price history charts unavailable

**Missing Endpoints (2)**:
- `GET /api/charts/sparklines` - Dashboard sparklines
- `GET /api/charts/securities/{isin}` - Price history charts

**Estimated Work**: 2-3 days (price history aggregation)

---

## Issues Requiring Verification

### Cash Flows Module - ⚠️ NEEDS VERIFICATION
**Status**: Completely rewritten (2,034 lines added)

**Original Issues** (may be fixed):
1. Deposit currency bug (was passing wrong currency)
2. Missing fallback to core bucket
3. Deduplication performance (O(N) vs O(1))

**Action**: Test new implementation to verify fixes

---

### Trading Module - ⚠️ NEEDS VERIFICATION
**Status**: Service layer added (103 lines)

**Original Issues**:
1. No trade recording to database
2. Missing 7-layer validation
3. Missing safety checks

**Action**: Verify service layer includes validation and recording

---

## Partial Implementations

### Portfolio Module - ⚠️ 80% COMPLETE
**Missing**:
- Monthly aggregation for CAGR calculation
- Live cash balance calculation
- Native analytics (currently proxied to Python)

**Estimated Work**: 1 week

---

### Display Module - ⚠️ 60% COMPLETE
**Missing**:
- Hardware LED control (linux_leds.go)
- Event emission system
- Ticker update service

**Estimated Work**: 3-5 days

---

### Dividends Module - ⚠️ 90% COMPLETE
**Missing**:
- DRIP background job execution (Python-dependent)

**Note**: API is excellent, but job requires Python endpoints

---

### Securities/Universe Module - ⚠️ 70% NATIVE
**Status**: 10/10 endpoints present, 7 proxied to Python

**Proxied Operations** (complex writes):
- `POST /api/securities` - Create security
- `POST /api/securities/add-by-identifier` - Add by symbol
- `POST /api/securities/refresh-all` - Batch score refresh
- `POST /api/securities/{isin}/refresh-data` - Full refresh
- `POST /api/securities/{isin}/refresh` - Quick refresh
- `PUT /api/securities/{isin}` - Update security
- Several refresh operations

**Estimated Work**: 1-2 weeks to eliminate proxies

---

## Migration Roadmap

### Phase 1: Unblock Auto-Trading (3-4 weeks) - P0

**Goal**: Enable autonomous trading without Python

**Tasks**:
1. Planning/Recommendations module (14 endpoints) - 3 weeks
2. Emergency Rebalancing - 1-2 weeks
3. Verify cash flows bug fixes - 2-3 days
4. Verify trading validation - 2-3 days

**Deliverables**:
- System generates and executes trades autonomously
- Emergency safety system operational
- No Python dependency for trading

---

### Phase 2: Complete Operations (1-2 weeks) - P1

**Goal**: Full manual control and configuration

**Tasks**:
1. System job triggers (15 endpoints) - 3-5 days
2. Settings module (9 endpoints) - 3-5 days
3. Market regime detection (URGENT) - 2-3 days

**Deliverables**:
- Manual operations fully supported
- Settings manageable via API
- Market regime detection working

---

### Phase 3: Feature Completeness (3-4 weeks) - P2

**Goal**: 100% functional parity with Python

**Tasks**:
1. Satellites module completion - 2 weeks
2. Portfolio module completion - 1 week
3. Charts module - 2-3 days
4. Display module - 3-5 days

**Deliverables**:
- All Python features available in Go
- Multi-bucket strategies operational
- Complete analytics and reporting

---

### Phase 4: Python Deprecation (1-2 weeks) - P3

**Goal**: Eliminate all Python dependencies

**Tasks**:
1. Remove proxies in universe module - 1 week
2. Portfolio analytics native implementation - 3-5 days
3. Performance optimization - 3-5 days
4. Documentation - 2-3 days

**Deliverables**:
- Zero Python dependencies
- Fully autonomous Go system
- Complete documentation

---

## Total Effort Estimate

| Phase | Weeks | Priority | Blockers |
|-------|-------|----------|----------|
| Phase 1 | 3-4 | P0 | Auto-trading blocked |
| Phase 2 | 1-2 | P1 | Manual operations limited |
| Phase 3 | 3-4 | P2 | Feature incomplete |
| Phase 4 | 1-2 | P3 | Python dependency |
| **Total** | **8-12 weeks** | - | **Full migration** |

---

## Immediate Actions

### This Week
1. Verify cash flows bug fixes (deposit currency, fallback, deduplication)
2. Verify trading service validation and recording
3. Begin Planning/Recommendations module implementation

### Next Week
4. Complete recommendations API (GET + POST execute)
5. Implement planner status endpoint
6. Start emergency rebalancing handler

### Week 3-4
7. Complete planner config CRUD endpoints
8. Complete emergency rebalancing
9. Integration testing

---

## Strategic Decisions Needed

1. **Satellites Priority**
   - Is multi-bucket strategy required for production?
   - Can defer to Phase 3+?

2. **Analytics Approach**
   - Port PyFolio to Go?
   - Keep Python analytics service?
   - Implement Go-native analytics?

3. **Display Hardware**
   - Is LED control critical?
   - Can dashboard suffice?

4. **Charts Module**
   - Essential for dashboard?
   - Can client fetch price data directly?

---

## Success Criteria

### Phase 1 (Auto-Trading)
- [ ] Generate recommendations without Python
- [ ] Execute recommendations successfully
- [ ] Emergency rebalancing handles negative balances
- [ ] All safety checks operational

### Phase 2 (Operations)
- [ ] All background jobs manually triggerable
- [ ] Settings modifiable via API
- [ ] Market regime detection operational

### Phase 3 (Feature Complete)
- [ ] Satellite strategies operational
- [ ] Complete analytics and reporting
- [ ] Feature parity with Python

### Phase 4 (Independence)
- [ ] Zero Python dependencies
- [ ] Python codebase archived
- [ ] Complete documentation

---

## Recommendation

**PROCEED WITH PHASE 1** - Focus on Planning/Recommendations module to unblock autonomous trading. The Go architecture is solid, infrastructure is ready, and 28 recent commits show strong momentum.

**Estimated Time to Production**: 3-4 weeks (Phase 1 only) or 8-12 weeks (full independence)

---

*Last Updated: 2026-01-02*
