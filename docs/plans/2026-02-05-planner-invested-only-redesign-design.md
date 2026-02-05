# Planner Invested-Only Targets + Funding Sells Redesign (Design)

**Goal:** Make planner/rebalancing semantics consistent and practical by using *invested-only* allocation weights (positions sum to 1.0) while still supporting *sell-to-fund-buys* and explicit cash deployment.

**Architecture:** Keep the current `Planner` facade, but evolve `PortfolioAnalyzer` to provide invested-only allocations and invested-value totals. Update `RebalanceEngine` to size deltas against invested value (not total value including cash), and add an explicit “funding sells” pass that can raise cash for high-priority buys when the cash constraint would otherwise prevent deployment.

**Tech stack:** Python async, existing planner modules (`sentinel/planner/*`), tests in `tests/test_planner.py`.

---

## Definitions (canonical semantics)

### Invested-only allocation (canonical)

- **Denominator:** sum of position EUR values (no cash).
- **Weights:** `w_i = position_value_eur_i / invested_value_eur`.
- **Properties:**
  - If cash increases (deposit), weights of positions don’t change.
  - Ideal portfolio weights naturally represent the desired *composition* of invested capital.

### Cash deployment

- Deposits create *cash* that must be deployed over time.
- Deployment is **not** achieved by including cash in the allocation denominator; it is a separate policy.

### Funding sells

- When a desired buy cannot be funded by current cash (plus proceeds from already-planned sells), the system may recommend additional sells to raise enough budget.
- Funding sells should be explicit (not accidental) so rules/guardrails can be reasoned about and tested.

---

## Current behavior gaps (what we’re fixing)

1) `PortfolioAnalyzer.get_current_allocations()` currently computes weights as `% of (positions + cash)`.
2) `Planner.get_recommendations()` sizes targets against `Portfolio.total_value()` (cash included).
3) Cash constraint `_apply_cash_constraint()` can downscale buys, but it doesn’t *choose* sells to fund specific high-priority buys.

---

## Proposed changes (high level)

### 1) PortfolioAnalyzer provides invested-only allocations

Update `sentinel/planner/analyzer.py` to compute:

- `position_values_eur: dict[symbol, float]`
- `invested_value_eur = sum(position_values_eur.values())`
- `allocations[symbol] = position_values_eur[symbol] / invested_value_eur` (if invested_value_eur > 0)

Caching remains acceptable (same 5-minute TTL), but the cache key should reflect the semantics (invested-only).

### 2) Planner passes invested value to RebalanceEngine

Update `sentinel/planner/planner.py` so `get_recommendations()` passes:

- `invested_value_eur` (positions only)
- `total_value_eur` (for informational context and possibly reserve-cash policy)

This avoids using cash-included totals to size composition rebalances.

### 3) RebalanceEngine sizes deltas against invested value

Modify `sentinel/planner/rebalance.py` to treat `total_value` parameter as *invested_value_eur* (or rename parameter, but keep backward compatibility if needed).

- `current_value_eur = current_alloc * invested_value_eur`
- `target_value_eur = target_alloc * invested_value_eur`

This makes recommendations represent *composition adjustments* independent of cash.

### 4) Add explicit cash deployment + funding-sells step

Introduce a two-stage flow inside `RebalanceEngine.get_recommendations()`:

1. **Composition recommendations** (existing `_build_recommendation` for deltas vs ideal).
2. **Budgeting / deployment**:
   - Start with available budget = `cash_eur - reserve_cash_eur` (plus proceeds from already-planned sells).
   - Ensure buys are feasible:
     - If high-priority buys are infeasible due to budget, add “funding sells” chosen from eligible positions.

Funding sell selection (minimal viable rules):

- Candidates: positions with `allow_sell=1`, not trade-blocked, not in cooloff.
- Rank candidates by *worst expected_return first*, then *most overweight* first.
- Sell just enough (respecting lot_size and min_trade_value) to fund the buy(s).

Guardrails (minimal, optional settings):

- `planner_reserve_cash_eur` (default 0)
- `planner_max_turnover_pct` (default None or 1.0)
- `planner_deploy_cash_pct` (default 1.0) or `planner_min_deploy_eur`

(If settings are too much for first pass, implement hardcoded conservative defaults and add settings later.)

### 5) Behavior with deposits (the $1000 deposit / $1400 stock case)

With invested-only weights, a deposit doesn’t create artificial “underweight” deltas. Deployment is handled by:

- If a desired buy is high priority but unaffordable, create funding sells (sell lower quality / overweight holdings) to raise required cash.

This keeps the “rotation + deployment” behavior without conflating semantics.

---

## Backtest / as-of-date note (out of scope for this slice)

This redesign does not fully fix `as_of_date` correctness (cash/positions totals are still live). It should be addressed as a separate audit item with dedicated tests.

---

## Testing strategy

Add/extend tests in `tests/test_planner.py`:

1) **Invested-only allocation**: with cash present, allocations remain 1.0 across positions and do not shrink.
2) **Sizing**: target value deltas computed against invested_value_eur.
3) **Funding sells**: when cash < required buy, engine adds sells (low-score / overweight first) sufficient to fund buy.

---

## Non-goals (for this iteration)

- Full per-currency budget enforcement.
- Full as-of-date correctness across the entire planning stack.
- Changing ML prediction write-to-DB behavior.
