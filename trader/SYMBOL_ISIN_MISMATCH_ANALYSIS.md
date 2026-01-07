# Symbol/ISIN Mismatch Analysis

**Date:** 2026-01-07
**Status:** Critical Issues Identified
**Impact:** System cannot generate recommendations, optimizer fails, data cleanup deletes all daily prices

## Quick Reference

**Total Issues Found:** 11
**Critical (Must Fix):** 6
**Needs Verification:** 0
**Correct (Keep As-Is):** 5

### Critical Issues Summary:
1. ⚠️ **History Cleanup Job** - Deletes all daily prices (ISIN vs symbol comparison)
2. ⚠️ **RiskModelBuilder.fetchPriceHistory** - Queries with symbols, data stored with ISINs
3. ⚠️ **Optimizer Service** - Passes symbols instead of ISINs
4. ⚠️ **Cleanup Job cleanupPortfolioData** - Uses symbol WHERE on ISIN PRIMARY KEY tables
5. ⚠️ **Market Index Service** - Queries with symbols but indices stored with ISINs
6. ⚠️ **BuildRegimeAwareCovarianceMatrix** - Same issue as #2, for market indices

## Executive Summary

The codebase has multiple critical symbol/ISIN mismatches that prevent the system from functioning correctly. The architecture principle states: **ISIN should be used internally for everything, symbols only for 3rd-party API calls (Tradernet/Yahoo)**. However, several components violate this principle, causing data loss and query failures.

## Architecture Principle

**Correct Pattern:**
- **Internal operations:** Use ISIN as primary identifier
- **External APIs:** Convert ISIN → Symbol for Tradernet/Yahoo calls
- **Database storage:** ISIN in PRIMARY KEY columns, symbol in indexed columns for display/API conversion

**Current Violations:**
- History database stores ISINs in `symbol` column (legacy schema)
- Some queries use symbols where ISINs are stored
- Cleanup jobs compare ISINs to symbols (mismatch)

---

## Critical Issues Found

### 1. **History Cleanup Job - Deletes All Daily Prices** ⚠️ CRITICAL

**File:** `trader/internal/modules/cleanup/history_cleanup_job.go`

**Problem:**
- `findOrphanedSymbols()` (lines 75-118) compares:
  - `daily_prices.symbol` (contains ISINs) vs `securities.symbol` (contains Tradernet symbols)
- All ISINs are treated as "orphans" because they don't match Tradernet symbols
- Result: **All daily_prices data is deleted every night**

**Code:**
```go
// Line 79: Gets ISINs from daily_prices
rows, err := j.historyDB.Conn().Query("SELECT DISTINCT symbol FROM daily_prices")

// Line 95: Gets Tradernet symbols from securities
rows2, err := j.universeDB.Conn().Query("SELECT symbol FROM securities")

// Line 111-114: Compares ISINs to symbols - ALWAYS MISMATCHES!
for symbol := range historySymbols {
    if !activeSymbols[symbol] {
        orphaned = append(orphaned, symbol)
    }
}
```

**Impact:**
- Daily prices deleted every night
- Optimizer cannot build covariance matrix
- System cannot generate recommendations

**Fix Required:**
- Query `securities.isin` instead of `securities.symbol`
- Compare ISINs to ISINs

---

### 2. **RiskModelBuilder.fetchPriceHistory - Symbol/ISIN Mismatch** ⚠️ CRITICAL

**File:** `trader/internal/modules/optimization/risk.go`

**Problem:**
- `fetchPriceHistory()` (line 240) receives symbols (e.g., "AAPL.US")
- Queries `daily_prices` WHERE `symbol IN (...)` (line 256)
- But `daily_prices.symbol` contains ISINs (e.g., "US0378331005")
- Result: **Query returns 0 rows**

**Code:**
```go
// Line 91-94: OptimizerService passes symbols
symbols := make([]string, len(activeSecurities))
for i, sec := range activeSecurities {
    symbols[i] = sec.Symbol  // "AAPL.US"
}

// Line 135: Passes symbols to RiskModelBuilder
covMatrix, _, correlations, err := os.riskBuilder.BuildCovarianceMatrix(symbols, DefaultLookbackDays)

// Line 256: Queries with symbols, but data is stored with ISINs
WHERE symbol IN (...)  // Looking for "AAPL.US" but data has "US0378331005"
```

**Impact:**
- Optimizer cannot build covariance matrix
- "insufficient price history: only 0 days available"
- No optimizer target weights
- No recommendations generated

**Fix Required:**
- Convert symbols to ISINs before querying
- Or: Query `securities` to get ISINs, then query `daily_prices` with ISINs

---

### 3. **Cleanup Job cleanupPortfolioData - Uses Symbol Instead of ISIN** ⚠️ CRITICAL

**File:** `trader/internal/modules/cleanup/history_cleanup_job.go`

**Problem:**
- `cleanupPortfolioData()` (line 155) receives an ISIN (from orphaned check)
- But deletes from `positions` and `scores` using `WHERE symbol = ?`
- These tables use ISIN as PRIMARY KEY, not symbol
- Result: **Deletions fail or delete wrong records**

**Code:**
```go
// Line 130: Receives ISIN (from daily_prices.symbol which contains ISINs)
result, err := j.historyDB.Conn().Exec("DELETE FROM daily_prices WHERE symbol = ?", symbol)

// Line 157: Tries to delete from positions using symbol, but PRIMARY KEY is ISIN!
_, err := j.portfolioDB.Conn().Exec("DELETE FROM positions WHERE symbol = ?", symbol)

// Line 163: Tries to delete from scores using symbol, but PRIMARY KEY is ISIN!
_, err = j.portfolioDB.Conn().Exec("DELETE FROM scores WHERE symbol = ?", symbol)
```

**Impact:**
- Cleanup doesn't work correctly
- May delete wrong records or fail silently
- Orphaned data accumulates

**Fix Required:**
- Use `WHERE isin = ?` instead of `WHERE symbol = ?` for positions and scores

**Note:** Test file `history_cleanup_job_test.go` also uses symbol WHERE clauses (lines 111, 115) - needs updating for tests to pass

---

### 4. **Charts Service - Correctly Uses ISIN but Query Comment is Misleading**

**File:** `trader/internal/modules/charts/service.go`

**Status:** ✅ **CORRECT IMPLEMENTATION** (but comment is confusing)

**Code:**
```go
// Line 123: Query correctly uses ISIN parameter
WHERE symbol = ? AND date >= ?  // Parameter is ISIN, column name is "symbol" (legacy)
```

**Note:** This is actually correct! The `daily_prices.symbol` column contains ISINs, so passing ISIN works. The column name is misleading (legacy schema).

**Recommendation:**
- Add comment clarifying that `symbol` column contains ISINs
- Or: Migrate schema to rename column to `isin`

---

### 5. **Market Index Service - Symbol/ISIN Mismatch** ⚠️ CRITICAL

**File:** `trader/internal/modules/portfolio/market_index_service.go`

**Problem:**
- `getIndexReturns()` (line 188) queries `daily_prices` with symbol "SPX.US"
- But indices are created with ISIN = "INDEX-SPX.US" (line 93)
- If indices are synced via historical sync, they're stored with ISIN "INDEX-SPX.US" in `daily_prices.symbol` column
- Query looks for "SPX.US" but data is stored as "INDEX-SPX.US"
- Result: **No index data found**

**Code:**
```go
// Line 93: Creates index with ISIN = "INDEX-SPX.US"
isin := fmt.Sprintf("INDEX-%s", idx.Symbol)  // "INDEX-SPX.US"

// Line 192: Queries with symbol "SPX.US"
WHERE symbol = ?  // Looking for "SPX.US" but data has "INDEX-SPX.US"
```

**Root Cause:**
- Indices are created with ISIN = "INDEX-SPX.US" in `securities` table
- If synced via `SyncAllHistoricalData()`, they're stored with ISIN in `daily_prices.symbol` column
- But `getIndexReturns()` queries with the symbol "SPX.US", not the ISIN

**Impact:**
- Market index returns cannot be calculated
- Regime detection fails (depends on market returns)
- Regime-aware covariance matrix cannot be built

**Fix Required:**
- Option A: Query with ISIN instead of symbol
  ```go
  // Get ISIN from securities table first
  var isin string
  err := s.universeDB.QueryRow("SELECT isin FROM securities WHERE symbol = ? AND product_type = 'INDEX'", symbol).Scan(&isin)
  // Then query with ISIN
  WHERE symbol = ?  // Use isin ("INDEX-SPX.US")
  ```

- Option B: Store indices with symbol as ISIN (change creation logic)
  ```go
  // Line 93: Use symbol as ISIN for indices
  isin := idx.Symbol  // "SPX.US" instead of "INDEX-SPX.US"
  ```

**Status:** ✅ **VERIFIED - CRITICAL ISSUE**

---

### 6. **BuildRegimeAwareCovarianceMatrix - Uses Index Symbols** ⚠️ CRITICAL

**File:** `trader/internal/modules/optimization/risk.go`

**Problem:**
- `BuildRegimeAwareCovarianceMatrix()` (line 106) calls `fetchPriceHistory()` with index symbols
- Line 161: `indexPriceData, err := rb.fetchPriceHistory(indexSymbols, lookbackDays)`
- Same issue as #2 and #5 - queries with symbols but data stored with ISINs
- Indices are stored with ISIN "INDEX-SPX.US" but query uses "SPX.US"
- Result: **No index price data found**

**Code:**
```go
// Line 152-159: Creates index symbols
indices := []marketIndexSpec{
    {Symbol: "SPX.US", Weight: 0.20},
    {Symbol: "STOXX600.EU", Weight: 0.50},
    {Symbol: "MSCIASIA.ASIA", Weight: 0.30},
}
indexSymbols := make([]string, 0, len(indices))
for _, idx := range indices {
    indexSymbols = append(indexSymbols, idx.Symbol)  // "SPX.US"
}

// Line 161: Passes symbols to fetchPriceHistory
indexPriceData, err := rb.fetchPriceHistory(indexSymbols, lookbackDays)
// fetchPriceHistory queries WHERE symbol IN ("SPX.US", ...)
// But data is stored with ISIN "INDEX-SPX.US"
```

**Impact:**
- Regime-aware covariance matrix cannot be built
- Market returns cannot be calculated
- Regime detection fails
- Optimizer falls back to basic covariance matrix

**Fix Required:**
- Convert index symbols to ISINs before calling `fetchPriceHistory()`
- Or: Fix `fetchPriceHistory()` to accept ISINs and convert symbols to ISINs internally
- Same fix as #2 (RiskModelBuilder.fetchPriceHistory)

**Status:** ✅ **VERIFIED - CRITICAL ISSUE**

---

## Additional Issues Found

### 7. **Optimizer Service - Passes Symbols Instead of ISINs**

**File:** `trader/internal/modules/optimization/service.go`

**Problem:**
- Line 91-94: Extracts symbols from securities
- Line 135: Passes symbols to `BuildCovarianceMatrix()`
- Should pass ISINs instead

**Fix Required:**
```go
// Current (WRONG):
symbols := make([]string, len(activeSecurities))
for i, sec := range activeSecurities {
    symbols[i] = sec.Symbol  // "AAPL.US"
}

// Should be:
isins := make([]string, len(activeSecurities))
for i, sec := range activeSecurities {
    isins[i] = sec.ISIN  // "US0378331005"
}
```

---

### 8. **ReturnsCalculator.getCAGRAndDividend - Uses Symbol via JOIN**

**File:** `trader/internal/modules/optimization/returns.go`

**Status:** ✅ **CORRECT** (uses JOIN to map symbol → ISIN)

**Code:**
```go
// Line 397-398: Correctly uses JOIN to map symbol to ISIN
INNER JOIN positions p ON s.isin = p.isin
WHERE p.symbol = ?
```

**Note:** This pattern is acceptable - it correctly maps symbol → ISIN via positions table.

---

### 9. **PlannerBatchJob.populateCAGRs - Uses Symbol via JOIN**

**File:** `trader/internal/scheduler/planner_batch.go`

**Status:** ✅ **CORRECT** (uses JOIN to map symbol → ISIN)

**Code:**
```go
// Line 527: Correctly uses JOIN
INNER JOIN positions p ON s.isin = p.isin
WHERE s.cagr_score IS NOT NULL
```

**Note:** This pattern is acceptable.

---

### 10. **Symbolic Regression Data Prep - Uses Symbol via JOIN**

**File:** `trader/internal/modules/symbolic_regression/data_prep.go`

**Status:** ✅ **CORRECT** (uses JOIN to map symbol → ISIN)

**Code:**
```go
// Line 355-356: Correctly uses JOIN
INNER JOIN positions p ON s.isin = p.isin
WHERE p.symbol = ?
```

**Note:** This pattern is acceptable.

---

### 11. **Rebalancing Service - Uses Symbol via JOIN**

**File:** `trader/internal/modules/rebalancing/service.go`

**Status:** ✅ **CORRECT** (uses JOIN to map symbol → ISIN)

**Code:**
```go
// Line 403: Correctly uses JOIN
INNER JOIN positions p ON s.isin = p.isin
WHERE s.cagr_score IS NOT NULL
```

**Note:** This pattern is acceptable.

---

## Summary of Issues

### Critical (Must Fix Immediately)

1. **History Cleanup Job** - Deletes all daily prices due to ISIN/symbol comparison mismatch
2. **RiskModelBuilder.fetchPriceHistory** - Queries with symbols but data stored with ISINs
3. **Optimizer Service** - Passes symbols instead of ISINs to RiskModelBuilder
4. **Cleanup Job cleanupPortfolioData** - Uses symbol WHERE clause on ISIN PRIMARY KEY tables

### Additional Critical Issues

5. **Market Index Service** - Queries with symbols but indices stored with ISINs (VERIFIED)
6. **BuildRegimeAwareCovarianceMatrix** - Same issue as #2, for market indices (VERIFIED)

### Correct Implementations (Keep As-Is)

7. **Charts Service** - Correctly uses ISIN (despite misleading column name)
8. **ReturnsCalculator** - Correctly uses JOIN pattern (symbol → ISIN)
9. **PlannerBatchJob.populateCAGRs** - Correctly uses JOIN pattern
10. **Symbolic Regression** - Correctly uses JOIN pattern
11. **Rebalancing Service** - Correctly uses JOIN pattern

---

## Root Cause Analysis

### History Database Schema Issue

The `history.db` schema uses `symbol` as the column name but stores ISIN values:

```sql
CREATE TABLE daily_prices (
    symbol TEXT NOT NULL,  -- Column name says "symbol" but stores ISIN!
    ...
)
```

**Migration Note:** There's a TODO comment in `history_db.go:139`:
```go
// TODO: Migrate history database to use 'isin' column name for consistency
```

This legacy schema causes confusion throughout the codebase.

### Pattern: Symbol → ISIN Mapping via JOIN

Many components correctly use this pattern:
```sql
SELECT ... FROM scores s
INNER JOIN positions p ON s.isin = p.isin
WHERE p.symbol = ?
```

This is acceptable because:
- Positions table has both `isin` (PRIMARY KEY) and `symbol` (indexed)
- JOIN correctly maps symbol → ISIN
- Then queries scores by ISIN (PRIMARY KEY)

**However:** This pattern is inefficient. Better to:
1. Look up ISIN from symbol first
2. Query directly by ISIN

---

## Recommended Fixes

### Priority 1: Fix History Cleanup Job

**File:** `trader/internal/modules/cleanup/history_cleanup_job.go`

**Change:**
```go
// Line 95: Change from symbol to isin
rows2, err := j.universeDB.Conn().Query("SELECT isin FROM securities WHERE active = 1")
```

**Also fix cleanupPortfolioData:**
```go
// Line 157: Change from symbol to isin
_, err := j.portfolioDB.Conn().Exec("DELETE FROM positions WHERE isin = ?", isin)

// Line 163: Change from symbol to isin
_, err = j.portfolioDB.Conn().Exec("DELETE FROM scores WHERE isin = ?", isin)
```

---

### Priority 2: Fix RiskModelBuilder

**File:** `trader/internal/modules/optimization/risk.go`

**Option A: Convert symbols to ISINs before querying**
```go
func (rb *RiskModelBuilder) fetchPriceHistory(symbols []string, days int) (TimeSeriesData, error) {
    // Convert symbols to ISINs
    isins := make([]string, len(symbols))
    for i, symbol := range symbols {
        // Lookup ISIN from universe.db
        var isin string
        err := rb.universeDB.QueryRow("SELECT isin FROM securities WHERE symbol = ?", symbol).Scan(&isin)
        if err != nil {
            return TimeSeriesData{}, fmt.Errorf("failed to get ISIN for symbol %s: %w", symbol, err)
        }
        isins[i] = isin
    }

    // Query with ISINs
    query := `SELECT symbol, date, close FROM daily_prices WHERE symbol IN (...)`
    // Use isins instead of symbols
}
```

**Option B: Accept ISINs directly (RECOMMENDED)**
```go
// Change OptimizerService to pass ISINs
func (os *OptimizerService) Optimize(...) {
    isins := make([]string, len(activeSecurities))
    for i, sec := range activeSecurities {
        isins[i] = sec.ISIN
    }
    covMatrix, _, correlations, err := os.riskBuilder.BuildCovarianceMatrix(isins, DefaultLookbackDays)
}

// RiskModelBuilder.fetchPriceHistory already works with ISINs
// (daily_prices.symbol column contains ISINs, despite misleading name)
```

**Note:** RiskModelBuilder only has `historyDB` connection, not `universeDB`. Converting symbols to ISINs in OptimizerService is cleaner than adding universeDB to RiskModelBuilder.

---

### Priority 3: Fix Optimizer Service

**File:** `trader/internal/modules/optimization/service.go`

**Change:**
```go
// Line 91-94: Extract ISINs instead of symbols
isins := make([]string, len(activeSecurities))
for i, sec := range activeSecurities {
    isins[i] = sec.ISIN
}

// Line 135: Pass ISINs
covMatrix, _, correlations, err := os.riskBuilder.BuildCovarianceMatrix(isins, DefaultLookbackDays)
```

---

### Priority 4: Fix Market Index Service

**File:** `trader/internal/modules/portfolio/market_index_service.go`

**Change:**
```go
// getIndexReturns gets daily returns for a specific index
func (s *MarketIndexService) getIndexReturns(symbol string, days int) ([]float64, error) {
    // Get ISIN from securities table
    var isin string
    err := s.universeDB.QueryRow(`
        SELECT isin FROM securities
        WHERE symbol = ? AND product_type = 'INDEX'
    `, symbol).Scan(&isin)
    if err != nil {
        return nil, fmt.Errorf("failed to get ISIN for index %s: %w", symbol, err)
    }

    // Query with ISIN (stored in symbol column)
    query := `
        SELECT date, close
        FROM daily_prices
        WHERE symbol = ?  -- Use ISIN ("INDEX-SPX.US")
        ORDER BY date DESC
        LIMIT ?
    `
    rows, err := s.historyDB.Query(query, isin, days+1)
    // ... rest of function
}
```

**Alternative:** Change index creation to use symbol as ISIN:
```go
// Line 93: Use symbol as ISIN for indices
isin := idx.Symbol  // "SPX.US" instead of "INDEX-SPX.US"
```

**Recommendation:** Use first approach (query with ISIN) to maintain consistency with other securities.

---

## Testing Checklist

After fixes, verify:

1. ✅ History cleanup job doesn't delete valid daily prices
2. ✅ Optimizer can build covariance matrix (252 days of data)
3. ✅ Planner batch job generates sequences
4. ✅ Recommendations are generated
5. ✅ Charts service still works
6. ✅ All database queries use correct identifiers

---

## Long-Term Recommendations

1. **Migrate History Database Schema:**
   - Rename `daily_prices.symbol` → `daily_prices.isin`
   - Rename `monthly_prices.symbol` → `monthly_prices.isin`
   - Update all queries accordingly

2. **Eliminate Symbol → ISIN JOIN Pattern:**
   - Look up ISIN from symbol once at API boundary
   - Use ISIN internally throughout
   - Only convert ISIN → Symbol for external API calls

3. **Add Validation:**
   - Verify all securities have ISINs before operations
   - Log warnings when symbol → ISIN lookups fail
   - Add integration tests for symbol/ISIN flows

---

## Files Requiring Changes

### Critical Fixes Required:

1. `trader/internal/modules/cleanup/history_cleanup_job.go` - Lines 79, 95, 130, 157, 163
2. `trader/internal/modules/optimization/risk.go` - Line 240-337 (fetchPriceHistory), Line 161 (BuildRegimeAwareCovarianceMatrix)
3. `trader/internal/modules/optimization/service.go` - Lines 91-94, 135, 191, 203
4. `trader/internal/modules/portfolio/market_index_service.go` - Line 188-237 (getIndexReturns)

### Documentation Updates:

5. `trader/internal/modules/universe/history_db.go` - Add comments clarifying symbol column contains ISINs
6. `trader/internal/modules/charts/service.go` - Add comments clarifying symbol column contains ISINs

---

## Conclusion

The symbol/ISIN mismatch is a systemic issue affecting multiple components. The most critical issues are:

1. **History cleanup job deleting all daily prices** (prevents optimizer from working)
2. **RiskModelBuilder querying with wrong identifiers** (prevents optimizer from working)
3. **Optimizer service passing symbols instead of ISINs** (prevents optimizer from working)
4. **Cleanup job using wrong WHERE clauses** (prevents proper cleanup)
5. **Market index service querying with wrong identifiers** (prevents regime detection)
6. **BuildRegimeAwareCovarianceMatrix using wrong identifiers** (prevents regime-aware optimization)

**All 6 critical issues must be fixed** to restore the system's ability to:
- Generate recommendations
- Build covariance matrices
- Detect market regimes
- Perform proper data cleanup

The issues are interconnected - fixing one without fixing others will leave the system partially broken. A comprehensive fix addressing all symbol/ISIN mismatches is required.
