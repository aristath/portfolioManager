# ML Historical Predictions – Investigation

**Goal:** Enable historical ML predictions (and optional backfill) in the same spirit as historical scores: store one row per (symbol, date), support “as-of” reads, and backfill from price history.

**No code changes in this phase – investigation only.**

---

## 1. Current State

### 1.1 Where ML predictions are written

- **Table:** `ml_predictions` (see schema below).
- **Write path:** Only when the **planner** runs (`RebalanceEngine.get_recommendations()` in `planner/rebalance.py`).
  - Scoring job (`scoring:calculate`) calls `analyzer.update_scores()` and writes only **wavelet** scores to `scores`; it does **not** run ML or write to `ml_predictions`.
  - So ML rows are created on demand when someone (or a job) triggers the planner, not on a fixed schedule for “all symbols, latest date”.

### 1.2 Where ML predictions are read

- **API:** `api/routers/securities.py` (securities list / portfolio view).
  - Fetches “latest” ML prediction per symbol via:
    - `SELECT ... FROM ml_predictions WHERE predicted_at = (SELECT MAX(predicted_at) FROM ml_predictions p2 WHERE p2.symbol = ml_predictions.symbol)`.
  - Uses `final_score` (and wavelet/ml breakdown) for expected return; falls back to wavelet `expected_return` from `scores` when no ML row exists.

### 1.3 Current schema: `ml_predictions`

```text
prediction_id TEXT PRIMARY KEY,   -- "{symbol}_{unix_ts}"
symbol TEXT NOT NULL,
predicted_at INTEGER NOT NULL,   -- Unix timestamp when prediction was run
features TEXT,                    -- JSON
predicted_return REAL,
ml_score REAL,
wavelet_score REAL,
blend_ratio REAL,
final_score REAL,
inference_time_ms REAL
```

- Index: `(symbol, predicted_at DESC)`.
- **Semantics:** One row per prediction run; `prediction_id` is unique per run (symbol + now()). So the table is already **append-only** in practice. There is no “replace latest per symbol” – each planner run inserts new rows. “Latest” is resolved in read queries with `MAX(predicted_at)` per symbol.

### 1.4 End-to-end flow (live, single “today”)

1. Planner calls `db.get_scores(all_symbols)` → latest **wavelet** score per symbol (from `scores`).
2. For each symbol: get historical prices (no `end_date` in rebalance; effectively “up to now”), build `hist_rows`.
3. `FeatureExtractor.extract_features(symbol, date=today, price_data, ...)` → 20 features (price_data must have ≥200 rows; last row is “as of” `date`).
4. `MLPredictor.predict_and_blend(symbol, date=today, wavelet_score, ml_enabled, ml_blend_ratio, features)`:
   - Loads current model from disk: `EnsembleBlender.load(symbol)` (no date/version).
   - Runs ensemble on `features` → `predicted_return`.
   - `get_regime_adjusted_return(symbol, predicted_return, db)` → uses **current** `security.quote_data` (live quotes) for regime; no historical quote data.
   - Blends with `wavelet_score` → `final_score`.
   - `_store_prediction(...)` → INSERT into `ml_predictions` with `predicted_at = datetime.now().isoformat()`.

So: **live path** uses “today”, current quotes for regime, and current wavelet score from `scores`.

---

## 2. Comparison with historical scores (already done)

| Aspect | Scores (wavelet) | ML predictions |
|--------|-------------------|-----------------|
| **Schema** | Append-only: `id`, `symbol`, `score`, `components`, `calculated_at` | Already append-only; `prediction_id` (symbol+timestamp), `predicted_at` |
| **Write** | `Security.set_score()` → INSERT (no replace) | INSERT in `_store_prediction()` (no replace) |
| **Read “latest”** | `get_score(symbol)` / `get_scores(symbols)` → ORDER BY calculated_at DESC, id DESC LIMIT 1 | API: MAX(predicted_at) per symbol |
| **Read “as-of”** | Not in DB API yet; backfill script doesn’t need it | Not in DB/API; would need for backtester/UI |
| **Inputs for a date T** | Prices up to T → `analyzer.analyze_prices(symbol, prices, use_regime=False)` | Prices up to T → features; wavelet score as of T; model inference; regime = N/A for backfill |
| **Backfill script** | Loop dates/symbols; `get_prices(..., end_date=date_str)`; `analyze_prices(..., use_regime=False)`; INSERT scores | Would loop dates/symbols; prices + wavelet as of T; features; predict; INSERT ml_predictions with predicted_at = T |

---

## 3. What “historical ML” would mean

- **Stored:** One row per (symbol, as-of date): same schema, but `predicted_at` used as the “as-of” date (e.g. end-of-day date string or timestamp).
- **Backfill:** For each past date T (and each symbol), compute features using only data up to T, get wavelet score as of T from `scores`, run **current** model on those features, optionally skip regime, write one row with `predicted_at = T`.
- **Read “latest”:** Unchanged: MAX(predicted_at) per symbol.
- **Read “as-of T”:** New: query row(s) with `predicted_at <= T` ORDER BY predicted_at DESC LIMIT 1 per symbol (or equivalent), e.g. for backtester or time-series UI.

---

## 4. Dependencies and blockers

### 4.1 Wavelet score as of date T (for blending)

- **Current:** Planner uses `get_scores(all_symbols)` → latest score per symbol (no date).
- **For backfill:** We need the wavelet score **as of date T** to blend with ML.
- **Gap:** `get_score(symbol)` / `get_scores(symbols)` do not accept an `as_of` date. We have historical rows in `scores` with `calculated_at`.
- **Required:** Add something like `get_score(symbol, as_of_date=None)` / `get_scores(symbols, as_of_date=None)` that:
  - If `as_of_date` is None: current behavior (latest by calculated_at, id).
  - If `as_of_date` is set: return the score row where `calculated_at <= end_of_day_ts(as_of_date)` ORDER BY calculated_at DESC, id DESC LIMIT 1 (and similarly for bulk).
- **Note:** Backfill script can do this with a direct query if we don’t want to extend the public DB API yet.

### 4.2 Features at date T

- **FeatureExtractor.extract_features(symbol, date, price_data, ...)** already takes `date` and `price_data`; it uses the last row of `price_data` as “current” and computes all 20 features (including aggregates).
- **Aggregates:** `_extract_aggregate_features(security_data, date)` → `_compute_aggregate_features(agg_symbol, date, ...)` loads aggregate prices via `get_prices_bulk([agg_symbol])` (no `end_date`), then filters `price_data["date"] <= date`. So it already supports “as of date” for a given `date`.
- **For backfill:** Per symbol, call `get_prices(symbol, days=LOOKBACK, end_date=date_str)` (we have `end_date` on `get_prices`), convert to DataFrame, pass `date=date_str` and security metadata (geography, industry) so aggregates work. Need ≥200 rows (existing requirement).
- **Conclusion:** Feature extraction is ready for historical dates once we have price data up to T and optional `get_prices_bulk(..., end_date)` if we want to restrict aggregate fetches; in-memory filter already supports T.

### 4.3 Regime adjustment

- **Current:** `get_regime_adjusted_return(symbol, ml_return, db)` uses `db.get_security(symbol)` → `quote_data` (live). Regime is from real-time quotes (chg5, chg22, ltp, x_max, x_min, etc.).
- **Historical:** We do not have quote_data for past dates. Options:
  - **Option A (recommended for v1):** In backfill, skip regime: use raw ML return (like `use_regime=False` in score backfill). Requires a way to call the predictor without regime (e.g. `predict_and_blend(..., use_regime=False)` or an internal “store only” path that doesn’t call `get_regime_adjusted_return`).
  - **Option B:** Derive a proxy “regime” from historical prices (e.g. momentum / position in range) and use that in backfill. More work, consistent with “point-in-time” but not required for a first backfill.
  - **Option C:** Store historical regime elsewhere and use it; out of scope for this investigation.
- **Conclusion:** For a first historical ML backfill, skip regime (Option A). Live path keeps current behavior.

### 4.4 Model version (“point-in-time” vs “current model”)

- **Current:** Models are stored under `data/ml_models/{symbol}/` and loaded with `EnsembleBlender.load(symbol)`. There is no versioning by date; retraining overwrites.
- **Backfill options:**
  - **Use current model (recommended for v1):** For every backfill date T, run the **current** model on features computed at T. Interpretation: “What would our current model predict if we had run it at T?” Good for backtesting current strategy and for filling history quickly. No model versioning needed.
  - **Point-in-time models:** For each T, load a model trained only on data up to T. Would require training and storing models per (symbol, training_date) and loading by date – much heavier and out of scope for this doc.
- **Conclusion:** Backfill uses the current model only; document this clearly so “historical ML” is interpreted as “current model applied to historical features (and historical wavelet).”

### 4.5 prediction_id and predicted_at

- **Current:** `prediction_id = f"{symbol}_{datetime.now().isoformat()}"` – unique per run, can collide if two runs in same second for same symbol.
- **For backfill:** We want one row per (symbol, date). Options:
  - Use `prediction_id = f"{symbol}_{date_str}"` (e.g. YYYY-MM-DD) so backfill is idempotent per day; then “latest” is still MAX(predicted_at). If we ever run ML twice per day (e.g. live + backfill), we need a policy (e.g. backfill uses date only; live uses datetime).
  - Or add an `id INTEGER AUTOINCREMENT` and make the natural key `(symbol, predicted_at)` for “latest” / “as-of” queries; keep `prediction_id` for backward compatibility or drop it. Schema change.
- **Recommendation:** For backfill-only rows, use `prediction_id = f"{symbol}_{date_str}"` and `predicted_at = date_str" 23:59:59"` (or normalized format). Resumability: before insert, check `SELECT 1 FROM ml_predictions WHERE symbol = ? AND predicted_at = ?` (or same for prediction_id) and skip if present.

### 4.6 Where planner gets wavelet score

- Rebalance calls `scores_map = await self._db.get_scores(all_symbols)` and then `wavelet_score = scores_map.get(symbol, 0)`. So it uses the **latest** wavelet score from `scores`. For live, that’s correct. For backfill we’d call a “score as of T” instead (once available).

### 4.7 get_scores / scores API and “as-of”

- **API** (securities list): Currently uses `SELECT * FROM scores` and builds `scores_map` by symbol – this will break or be ambiguous once there are multiple rows per symbol unless the query is updated to “latest” (e.g. same pattern as scores backfill: latest by calculated_at, id). Ensure the API uses the same “latest score” logic as `get_scores()` (e.g. one row per symbol, ORDER BY calculated_at DESC, id DESC).
- **As-of:** No current requirement for the API to show “score as of date X”. Backtester or analytics could use a dedicated `get_score(symbol, as_of_date=T)` / `get_scores(symbols, as_of_date=T)` or raw query.

---

## 5. Summary: what’s needed for historical ML (and backfill)

| Item | Effort | Notes |
|------|--------|--------|
| **Schema** | None | `ml_predictions` already append-only; optional: add `id` and index on `(symbol, predicted_at)` for cleaner “as-of” and dedup. |
| **Read “latest”** | None | Already MAX(predicted_at) per symbol. |
| **Read “as-of T”** | Small | New DB helper or query: per symbol, row with predicted_at <= T, ORDER BY predicted_at DESC LIMIT 1. Only needed if backtester/UI need it. |
| **Wavelet as of T** | Small | `get_score(symbol, as_of_date=T)` / `get_scores(symbols, as_of_date=T)` or equivalent (or inline query in backfill script). |
| **Features at T** | None | Already supported: `get_prices(..., end_date=T)`, DataFrame, `extract_features(symbol, T, price_data, ...)`. Aggregate path filters by date. |
| **Regime for backfill** | Small | Add a way to run prediction without regime (e.g. `use_regime=False` in predictor or internal path that skips `get_regime_adjusted_return`) and use raw ML return. |
| **Model** | None | Use current model for all backfill dates; document. |
| **prediction_id / resumability** | Small | Backfill uses stable id (e.g. symbol + date_str) and checks for existing (symbol, predicted_at) before insert. |
| **Backfill script** | Medium | Loop over date range and symbols; for each (date, symbol): get prices to date, get wavelet score as of date, extract features, predict (no regime), insert row with predicted_at = date. Skip if already present. |
| **get_prices_bulk end_date** | Optional | FeatureExtractor uses get_prices_bulk for aggregates and then filters by date; adding end_date would be consistent and could reduce memory for backfill. |

---

## 6. Suggested order of work (when implementing)

1. **Wavelet as-of:** Implement `get_score(symbol, as_of_date=None)` and `get_scores(symbols, as_of_date=None)` (or equivalent) using `calculated_at` (and id) so backfill can get the wavelet score for date T.
2. **Regime bypass:** In ML predictor, support “no regime” (e.g. `predict_and_blend(..., use_regime=False)` or internal flag) so backfill can store predictions without live quote_data.
3. **Backfill script:** New script (e.g. `scripts/backfill_ml_predictions.py`): iterate dates (and symbols); for each (date, symbol): get prices with `end_date=date`, get wavelet score as of date, extract features for date, run current model, apply blend with wavelet (no regime), INSERT into `ml_predictions` with `predicted_at = date`, skip if row exists. Reuse MIN_DAYS / LOOKBACK style constants; require models to exist for symbols that are ML-enabled.
4. **Optional:** DB helper for “ML prediction as of date T” for backtester/UI; optional `get_prices_bulk(..., end_date)` for aggregates.
5. **API:** Confirm securities list (and any other consumers) resolve “latest score” from `scores` in the same way as `get_scores()` (one row per symbol, latest by calculated_at/id), so historical scores don’t break the UI.

---

## 7. Files to touch (when implementing)

- **Database:** `sentinel/database/base.py` (and possibly `main.py`) – `get_score` / `get_scores` with optional `as_of_date`; any new “get ML prediction as of” helper.
- **Predictor:** `sentinel/ml_predictor.py` – regime bypass, and optionally `_store_prediction` accepting explicit `predicted_at` / `prediction_id` for backfill.
- **Feature extractor:** `sentinel/ml_features.py` – no change for backfill; optional support for `get_prices_bulk(..., end_date)` in aggregate path.
- **Planner:** No change for live path; backfill doesn’t go through planner.
- **API:** `sentinel/api/routers/securities.py` – ensure “latest score” from `scores` is well-defined (one row per symbol); optionally add “as-of” ML read later.
- **New:** `scripts/backfill_ml_predictions.py` – backfill loop as above.
- **Docs:** `docs/ML_PREDICTION.md` – document “historical ML” and “current model on historical features” semantics.

This document is the investigation baseline; no code changes have been made.
