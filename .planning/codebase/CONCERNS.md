# Codebase Concerns

**Analysis Date:** 2026-01-20

## Tech Debt

**Panic on Timezone Load Failure:**
- Issue: `internal/modules/market_hours/exchanges.go:578` uses `panic()` on timezone load failure instead of returning error
- Files: `internal/modules/market_hours/exchanges.go`
- Impact: Application crashes if timezone data is corrupt or missing, preventing graceful degradation
- Fix approach: Convert to error return and handle gracefully in initialization with fallback to UTC

**Log.Fatal Usage in HTTP Server:**
- Issue: `internal/server/server.go:473` uses `log.Fatal()` during route setup for schema initialization failure
- Files: `internal/server/server.go`
- Impact: Crashes entire application if cash_flows schema init fails during server startup, no graceful recovery
- Fix approach: Return error from server initialization and handle in main.go with proper cleanup

**Fatal Logging in Main:**
- Issue: `cmd/server/main.go` uses `log.Fatal()` in 4 locations (lines 102, 120, 186)
- Files: `cmd/server/main.go`
- Impact: Some failures occur after partial initialization (e.g., line 186 after resources allocated), preventing cleanup
- Fix approach: Use error returns with deferred cleanup where possible, reserve Fatal only for early init failures

**Deployment Manager Fatal Logging:**
- Issue: `internal/deployment/manager.go:95,102` uses `log.Fatal()` for git repository access failures
- Files: `internal/deployment/manager.go`
- Impact: Crashes entire application if git operations fail, even though deployment is optional feature
- Fix approach: Return errors instead of Fatal, allow application to run without deployment feature

**interface{} Instead of any:**
- Issue: Codebase uses `interface{}` (165 files, 1423+ occurrences) instead of Go 1.18+ `any` type alias
- Files: Throughout codebase (see grep results - 165 files affected)
- Impact: Less readable code, not using modern Go idioms despite requiring Go 1.23+
- Fix approach: Systematic replacement with `any` type alias (automated refactor safe)

**Large Files Indicate High Complexity:**
- Issue: Several files exceed 2000 lines indicating potential complexity
- Files:
  - `internal/utils/temperament_config.go` (2143 lines) - configuration mappings
  - `internal/testing/mocks.go` (2131 lines) - mock implementations
  - `internal/modules/universe/security_repository.go` (2095 lines) - repository logic
  - `internal/clients/tradernet/sdk/methods.go` (1747 lines) - SDK methods
- Impact: Difficult to navigate, test, and maintain; high cognitive load
- Fix approach: Split by responsibility (e.g., separate mock types to individual files, split repository by concern)

**TODO Comments in Universe Repository:**
- Issue: Multiple TODO comments in `internal/modules/universe/security_repository_json_test.go`
- Files: `internal/modules/universe/security_repository_json_test.go:40,45,96,237`
- Impact: Tests are skipped or incomplete for JSON repository implementation
- Fix approach: Implement or remove JSON repository tests if no longer needed

**SELECT * Queries:**
- Issue: Three files use `SELECT *` instead of explicit column lists
- Files:
  - `internal/modules/dividends/dividend_repository.go`
  - `internal/modules/trading/trade_repository.go`
  - `internal/modules/universe/score_repository.go`
- Impact: Schema changes break queries silently, performance overhead from fetching unused columns
- Fix approach: Replace with explicit column lists matching struct fields

**DEBUG Comment in Production Code:**
- Issue: `internal/clients/tradernet/client.go:152` contains DEBUG comment for logging raw API response
- Files: `internal/clients/tradernet/client.go`
- Impact: Indicates investigation code may still be present
- Fix approach: Review if debug logging is still needed or can be removed

## Known Bugs

**No Critical Bugs Identified:**
- No BUG, FIXME, or HACK comments found in codebase
- No error patterns indicating known issues

## Security Considerations

**log.Fatal in Production Code:**
- Risk: Application can be crashed by external failures (timezone data, git operations, schema init)
- Files: `cmd/server/main.go`, `internal/server/server.go`, `internal/deployment/manager.go`, `internal/modules/market_hours/exchanges.go`
- Current mitigation: Limited - relies on external resources being stable
- Recommendations: Convert to error returns with graceful degradation, especially for optional features like deployment

**Panic in Production Code:**
- Risk: Single timezone load failure crashes entire application
- Files: `internal/modules/market_hours/exchanges.go:578`
- Current mitigation: Assumes timezone database is always available
- Recommendations: Return errors instead of panic, provide fallback timezone handling

**Database Busy Timeout:**
- Risk: SQLite database locks can occur under high concurrent load
- Files: `internal/database/db.go:204` sets 5-second busy timeout
- Current mitigation: 5-second busy timeout, 25 max concurrent connections
- Recommendations: Monitor for "database is locked" errors in production, consider reducing concurrent connection limit if locks occur frequently

## Performance Bottlenecks

**Large Mock File:**
- Problem: `internal/testing/mocks.go` at 2131 lines may slow compilation
- Files: `internal/testing/mocks.go`
- Cause: All mock implementations in single file
- Improvement path: Split into separate files per interface, reduce compilation time

**Temperament Config Size:**
- Problem: `internal/utils/temperament_config.go` at 2143 lines is configuration data
- Files: `internal/utils/temperament_config.go`
- Cause: All temperament mappings in single map structure
- Improvement path: Consider splitting by category or moving to data file if dynamic config needed

**Connection Pool Limits:**
- Problem: 25 max concurrent connections may bottleneck under high API load
- Files: `internal/database/db.go:220`
- Cause: Conservative limit for embedded device (Arduino Uno Q)
- Improvement path: Monitor connection pool utilization, increase if needed while watching memory

## Fragile Areas

**Tradernet API Field Mapping:**
- Files: `internal/clients/tradernet/transformers_domain.go`, `internal/clients/tradernet/transformers.go`
- Why fragile: Tradernet uses inconsistent field names across endpoints ("i" vs "instr_nm" vs "instr_name" for same field)
- Safe modification: Always test with live API responses, maintain comprehensive field mapping documentation
- Test coverage: Transformer tests exist but rely on Tradernet API stability

**Database Schema Migration:**
- Files: `internal/server/server.go:472-474` (schema init during server startup)
- Why fragile: Schema initialization happens during HTTP server setup, failure causes Fatal crash
- Safe modification: Test schema changes thoroughly, ensure migrations are idempotent
- Test coverage: Limited - relies on manual testing

**Context.Background() Usage:**
- Files: 46 occurrences across 20 files
- Why fragile: Some contexts created with Background() should use request contexts for cancellation
- Safe modification: Review each usage, prefer context propagation from request handlers
- Test coverage: Limited - potential timeout/cancellation issues hard to test

**Goroutine Spawning:**
- Files: 76 occurrences across 22 files
- Why fragile: Multiple goroutines without explicit lifecycle management in some areas
- Safe modification: Ensure proper context cancellation, wait groups for cleanup
- Test coverage: Some test coverage for worker pools, limited for background monitors

## Scaling Limits

**SQLite Concurrent Writes:**
- Current capacity: 25 concurrent connections, 5-second busy timeout
- Limit: SQLite write serialization under high concurrent write load
- Scaling path: Monitor database lock contention, consider write queueing or migration to PostgreSQL if needed

**In-Memory Recommendation Repository:**
- Current capacity: All recommendations held in memory with mutex protection
- Limit: Memory exhausted if recommendation volume exceeds available RAM on Arduino Uno Q
- Scaling path: Migrate to persistent storage with LRU cache, or implement recommendation cleanup/archival

**Worker Pool Size:**
- Current capacity: Built-in worker pool for sequence evaluation
- Limit: Fixed pool size may bottleneck under high evaluation load
- Scaling path: Make pool size configurable, monitor queue depth and evaluation latency

## Dependencies at Risk

**modernc.org/sqlite (Pure Go SQLite):**
- Risk: Less battle-tested than cgo-based mattn/go-sqlite3
- Impact: Potential performance or compatibility issues with complex queries
- Migration plan: Switch to mattn/go-sqlite3 if performance issues arise (requires cgo)

**No Deprecated Dependencies Detected:**
- All dependencies appear current based on go.mod analysis

## Missing Critical Features

**No Circuit Breakers:**
- Problem: No circuit breaker pattern for external API calls (Tradernet, Yahoo Finance)
- Blocks: Cascading failures if external services degrade
- Priority: Medium - system handles API failures but doesn't prevent retry storms

**Limited Observability:**
- Problem: No distributed tracing, limited metrics collection beyond logs
- Blocks: Difficult to diagnose performance issues in production
- Priority: Low - current logging adequate for single-service architecture

**No Rate Limiting on API Endpoints:**
- Problem: HTTP API has no rate limiting for external clients
- Blocks: Potential abuse or resource exhaustion
- Priority: Low - single-user system on private network

## Test Coverage Gaps

**Integration Tests Missing:**
- What's not tested: End-to-end workflows across multiple modules
- Files: No integration test suite exists beyond unit tests
- Risk: Module interactions may fail in production despite passing unit tests
- Priority: Medium

**Deployment Manager:**
- What's not tested: GitHub artifact download, binary deployment, sketch upload
- Files: `internal/deployment/manager.go`, `internal/deployment/github_artifact.go`, `internal/deployment/sketch.go`
- Risk: Deployment failures could brick system requiring manual intervention
- Priority: High - manages system updates

**Backup/Restore System:**
- What's not tested: Full restore workflow from R2 backup
- Files: `internal/reliability/restore_service.go`, `internal/reliability/r2_backup_service.go`
- Risk: Backups may be corrupt or incomplete, discovered only during disaster recovery
- Priority: High - critical for data recovery

**Market Regime Detection:**
- What's not tested: Regime transitions, correlation calculations with live data
- Files: `internal/market_regime/detector.go`, `internal/market_regime/index_service.go`
- Risk: Regime detection failures could lead to poor allocation decisions
- Priority: Medium - affects portfolio optimization quality

**Database Transaction Handling:**
- What's not tested: Concurrent transaction conflicts, rollback scenarios
- Files: Database repositories lack comprehensive transaction tests
- Risk: Data corruption or inconsistency under concurrent writes
- Priority: Medium - SQLite handles atomicity but app logic untested

**WebSocket Client Reconnection:**
- What's not tested: Connection loss/recovery, backoff strategies
- Files: `internal/clients/tradernet/websocket_client.go`
- Risk: Market status updates may stop after connection loss
- Priority: Medium - affects market hours detection

**Cash-as-Security Sync:**
- What's not tested: Cash position synchronization edge cases
- Files: `internal/modules/cash_flows/cash_security_manager.go`
- Risk: Cash balance drift between broker and local state
- Priority: High - incorrect cash balances block trades

---

*Concerns audit: 2026-01-20*
