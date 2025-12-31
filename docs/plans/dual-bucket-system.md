# Dual-Bucket Portfolio System

## Overview

A dual-mode portfolio management system that allocates funds between two separate "buckets":

- **Conservative Bucket (90%)**: Long-term retirement-focused, current strategy
- **Opportunistic Bucket (10%)**: Short-term gains, momentum-based, feeds profits to main fund

The opportunistic bucket operates in virtual isolation with its own rules, universe, and cash tracking. Gains are harvested to the main fund; losses are contained.

---

## Core Concepts

### Bucket Separation

| Aspect | Conservative | Opportunistic |
|--------|--------------|---------------|
| Target allocation | 90% | 10% (max) |
| Strategy | Long-term, dividend-focused | Momentum, regime-aware |
| Hold duration | Months to years | Days to weeks |
| Universe | Blue chips, stable | Higher volatility, uncorrelated |
| Risk tolerance | Low | Higher (but contained) |

### Universe Separation

Each security belongs to exactly one universe. No overlap.

```sql
ALTER TABLE stocks ADD COLUMN universe TEXT DEFAULT 'conservative';
-- Values: 'conservative', 'opportunistic'
```

Benefits:
- Clear position attribution
- No "which bucket owns this?" ambiguity
- Simple filtering: `WHERE universe = 'opportunistic'`

---

## Cash Bucket Infrastructure

### Database Schema

```sql
-- Bucket definitions
CREATE TABLE buckets (
    id TEXT PRIMARY KEY,              -- 'conservative', 'opportunistic'
    name TEXT NOT NULL,
    target_pct REAL NOT NULL,         -- 0.90 or 0.10
    min_pct REAL,                     -- hibernation threshold (0.05)
    max_pct REAL,                     -- harvest threshold (0.10)
    strategy TEXT,
    consecutive_losses INTEGER DEFAULT 0,
    max_consecutive_losses INTEGER DEFAULT 5,
    high_water_mark REAL DEFAULT 0,
    high_water_mark_date TEXT,
    loss_streak_paused_at TEXT,
    created_at TEXT NOT NULL
);

-- Virtual cash per bucket per currency
CREATE TABLE bucket_balances (
    bucket_id TEXT NOT NULL,
    currency TEXT NOT NULL,
    balance REAL NOT NULL DEFAULT 0,
    last_updated TEXT NOT NULL,
    PRIMARY KEY (bucket_id, currency),
    FOREIGN KEY (bucket_id) REFERENCES buckets(id)
);

-- Audit trail for all bucket transactions
CREATE TABLE bucket_transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bucket_id TEXT NOT NULL,
    type TEXT NOT NULL,               -- 'deposit', 'harvest', 'trade_buy', 'trade_sell', 'dividend'
    amount REAL NOT NULL,
    currency TEXT NOT NULL,
    description TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (bucket_id) REFERENCES buckets(id)
);
```

### Critical Invariant

```
SUM(bucket_balances for currency X) == Actual brokerage balance for currency X
```

This must be enforced on every operation. Any drift indicates a bug.

### Events That Update Bucket Balances

| Event | Effect |
|-------|--------|
| Deposit | Split per target allocation (90/10) |
| Buy trade | Deduct from trade's bucket |
| Sell trade | Add to trade's bucket |
| Dividend | Add to bucket that owns the stock |
| Harvest | Transfer excess from opportunistic → conservative |

---

## Allocation Rules

### New Deposits

```
On deposit of $X:
    if opportunistic_pct < 10%:
        opportunistic gets: X * 0.10
        conservative gets: X * 0.90
    else:
        conservative gets: X * 1.00
```

Opportunistic bucket only receives funds when below its target allocation.

### Harvest Mechanism

When opportunistic bucket exceeds 10% of total portfolio:

```
excess = opportunistic_value - (total_portfolio * 0.10)
transfer excess to conservative bucket
```

This is done via cash transfer, not forced position sales. The opportunistic bucket naturally generates cash through profit-taking, which then flows to conservative.

---

## Dynamic Aggression System

### Percentage-Based Aggression

| Bucket % of Total | Aggression Level | Behavior |
|-------------------|------------------|----------|
| 10%+ | Full | All opportunities, full position sizes |
| 8-10% | High | Slightly reduced sizes |
| 6-8% | Moderate | Only higher-conviction trades |
| 5-6% | Low | Minimal new positions |
| <5% | Hibernation | Hold only, no new trades |

### Drawdown-Based Aggression

Track high water mark (peak bucket value). Calculate drawdown:

```
drawdown = (high_water_mark - current_value) / high_water_mark
```

| Drawdown from Peak | Effect |
|-------------------|--------|
| 0-15% | Full aggression allowed |
| 15-25% | Reduced position sizes |
| 25-35% | High-conviction only |
| 35%+ | Hibernation |

### Combined Aggression

```python
aggression = min(
    aggression_from_percentage(bucket_pct),
    aggression_from_drawdown(drawdown)
)
```

Both conditions must allow trading. Either can trigger hibernation.

---

## Safety Mechanisms

### Consecutive Losses Circuit Breaker

Track consecutive losing trades. After 5 losses in a row, pause trading.

**What counts as a loss:**
- Closed position where `sell_price < buy_price - threshold`
- Threshold accounts for fees/spreads (e.g., -1%)

**On pause:**
- Stop opening new positions
- Hold existing positions (don't panic sell)
- Log event for review

**Reset conditions:**
- Any winning trade resets counter to 0
- Held position recovers and closes at profit
- Time-based: 30 days allows one small "test" trade
- Manual override after review

### Win Cooldown

After exceptional performance (+20% in a month), temporarily reduce aggression.

Rationale:
- Prevents overconfidence
- Locks in gains
- Mean reversion: hot streaks often precede cold ones

### Trailing Stops

For opportunistic positions, implement trailing stops to protect gains:

```
if position_gain > 15%:
    set trailing_stop at 10% below peak
```

Lets winners run while locking in minimum profit.

### Graduated Re-awakening

When exiting hibernation, don't immediately go full aggressive:

```
Hibernation → First trade: 25% normal size
            → Win? Second trade: 50% normal size
            → Win? Third trade: 75% normal size
            → Win? Resume normal sizing
```

Bucket must prove it can make money before getting full capital back.

---

## Startup Process

### Initial State

- Opportunistic bucket starts at 0%
- No special migration of existing positions
- No forced selling

### Gradual Build-up

1. Each new deposit splits 90/10
2. Opportunistic cash accumulates passively
3. At 5% of portfolio, bucket "wakes up"
4. Trading begins with graduated position sizing

### Timeline Example

```
Month 1: Deposit €1000 → €100 to opportunistic (0.5% of portfolio)
Month 2: Deposit €1000 → €100 to opportunistic (~1%)
...
Month 10: Opportunistic reaches ~5%, begins trading
Month 12: Opportunistic reaches 10%, fully active
```

---

## Opportunistic Strategy (TBD)

High-level approach:

### Momentum + Regime Awareness

- **Bull market**: Full momentum, ride winners
- **Sideways market**: Mean reversion, buy dips on quality
- **Bear market**: Stay cash-heavy, wait for opportunities

### Position Management

- Shorter hold periods (days to weeks)
- Trailing stops to protect gains
- Quicker profit-taking
- Smaller position sizes with higher conviction threshold

### Universe Selection Criteria

- Lower correlation with conservative holdings
- Higher volatility (within reason)
- Different sectors/geographies for diversification
- Sufficient liquidity for quick entry/exit

---

## Reconciliation

### Continuous Checks

On every operation:
```python
assert sum(bucket_balances[currency]) == actual_brokerage_balance[currency]
```

### Periodic Full Reconciliation

Daily job to:
1. Fetch actual balances from brokerage
2. Calculate expected balances from bucket_balances
3. If mismatch:
   - Log discrepancy
   - Alert for review
   - Auto-correct small drifts (< €1)
   - Block trades for large drifts until resolved

---

## UI Considerations

### Unified View with Visual Distinction

- Single stock list with badge/color per universe
- Single recommendations list, tagged by bucket
- Dashboard shows combined + per-bucket breakdown
- Optional filter toggle to focus on one bucket

### Bucket Health Display

Show:
- Current allocation %
- Drawdown from peak
- Aggression level
- Consecutive losses count
- Status (Active / Cautious / Hibernating / Paused)

---

## Implementation Phases

### Phase 1: Foundation
- Add `universe` column to stocks
- Create bucket tables
- Implement virtual cash tracking
- Reconciliation system

### Phase 2: Deposit Splitting
- Modify deposit handling to split 90/10
- Dividend attribution to correct bucket
- Trade settlement to correct bucket

### Phase 3: Opportunistic Planner
- New scoring algorithm for momentum
- Regime detection integration
- Position sizing based on aggression
- Trailing stop implementation

### Phase 4: Safety Systems
- Drawdown tracking
- Consecutive loss tracking
- Graduated re-awakening
- Win cooldown

### Phase 5: Harvest Mechanism
- Detect when opportunistic > 10%
- Automated cash transfer
- Transaction logging

---

## Open Questions

1. **Opportunistic universe selection**: What specific stocks/ETFs? Sector ETFs? Small caps? Different geographies?

2. **Exact momentum signals**: What indicators? Moving averages? RSI? Relative strength?

3. **Harvest frequency**: Check daily? Weekly? On every transaction?

4. **Win cooldown duration**: How long to reduce aggression after big win?

5. **Test period**: Run in "paper trading" mode first before committing real capital?

---

## Risks

1. **Complexity**: More code = more potential bugs in a system managing real money

2. **Strategy risk**: Momentum strategies can underperform in choppy markets

3. **Behavioral risk**: Temptation to increase the 10% allocation during winning streaks

4. **Correlation risk**: If both buckets crash together, diversification benefit is lost

---

## Success Metrics

- Opportunistic bucket contributes positive net returns over 12-month periods
- Harvest events occur (profits actually flow to conservative bucket)
- Drawdowns stay within defined limits
- System correctly hibernates during losing streaks
- Virtual cash always reconciles with actual brokerage balance
