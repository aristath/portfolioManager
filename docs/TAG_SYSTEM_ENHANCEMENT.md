# Tag System Enhancement for Algorithm Improvements

## Overview

This document outlines enhancements to the tag system to support both performance optimization (fast filtering) and algorithm improvements (quality gates, bubble detection, value trap detection, etc.).

## Tag System Goals

1. **Fast Filtering**: Enable quick pre-filtering of securities (10-20 candidates vs 100+)
2. **Quality Gates**: Support quality thresholds for opportunity scoring
3. **Bubble Detection**: Identify potential bubbles for CAGR scoring
4. **Value Trap Detection**: Flag cheap but declining securities
5. **Total Return Focus**: Support dividend + growth total return calculation
6. **Optimizer Alignment**: Support target weight deviations
7. **Regime Adaptation**: Support regime-based risk adjustments

---

## Current Tag Analysis

### ✅ Existing Tags (Keep and Enhance)

**Quality Tags:**
- `high-quality` - Fundamentals > 0.8 AND long-term > 0.75 ✅
- `strong-fundamentals` - Fundamentals > 0.8 ✅
- `stable` - Fundamentals > 0.75 AND volatility < 0.20 ✅
- `consistent-grower` - Consistency > 0.8 AND CAGR > 8% ✅

**Opportunity Tags:**
- `value-opportunity` - Opportunity > 0.7 AND (below 52W > 20% OR P/E < market - 20%) ✅
- `deep-value` - Below 52W > 30% AND P/E < market - 20% ✅
- `oversold` - RSI < 30 ✅
- `recovery-candidate` - Negative momentum BUT fundamentals > 0.7 ✅

**Danger Tags:**
- `unsustainable-gains` - Annualized return > 50% AND volatility spike ✅
- `volatility-spike` - Current volatility > historical * 1.5 ✅
- `overvalued` - P/E > market + 20% AND near 52W high ✅
- `overweight` - Position weight > target + 2% OR > 10% ✅

**Dividend Tags:**
- `high-dividend` - Dividend yield > 6% ✅
- `dividend-opportunity` - Dividend score > 0.7 AND yield > 3% ✅
- `dividend-grower` - Dividend consistency > 0.8 AND 5Y avg > current ✅

### ❌ Missing Tags (Add for Algorithm Support)

**Quality Gate Tags:**
- `quality-gate-pass` - Meets minimum quality thresholds (fundamentals ≥ 0.6, long-term ≥ 0.5)
- `quality-gate-fail` - Fails minimum quality thresholds

**Bubble Detection Tags:**
- `bubble-risk` - High CAGR (> 16.5%) with poor risk metrics (Sharpe < 0.5 OR Sortino < 0.5 OR volatility > 0.40 OR fundamentals < 0.6)
- `quality-high-cagr` - High CAGR (> 15%) with good risk metrics (opposite of bubble)

**Value Trap Tags:**
- `value-trap` - Cheap (P/E < market - 20%) BUT (declining fundamentals OR negative momentum OR high volatility)
- `quality-value` - Cheap AND high quality (value-opportunity + high-quality)

**Total Return Tags:**
- `high-total-return` - Total return (CAGR + dividend) ≥ 15%
- `excellent-total-return` - Total return ≥ 18%
- `dividend-total-return` - Dividend yield ≥ 8% AND CAGR ≥ 5% (your 5% + 10% = 15% example)

**Optimizer Alignment Tags:**
- `underweight` - Position weight < target - 2% (for optimizer alignment)
- `target-aligned` - Position weight within ±1% of target
- `needs-rebalance` - Position weight deviation > 3% from target

**Regime-Specific Tags:**
- `regime-bear-safe` - Low volatility (< 0.20) AND strong fundamentals (> 0.75) AND low drawdown (< 20%)
- `regime-bull-growth` - High CAGR (> 12%) AND good fundamentals (> 0.7) AND positive momentum
- `regime-sideways-value` - Value opportunity AND stable fundamentals

**Risk-Adjusted Tags:**
- `high-sharpe` - Sharpe ratio ≥ 1.5
- `high-sortino` - Sortino ratio ≥ 1.5
- `poor-risk-adjusted` - Sharpe < 0.5 OR Sortino < 0.5 (for bubble detection)

---

## Enhanced Tag Categories

### 1. Quality Gate Tags (NEW - for Algorithm Support)

**Purpose**: Support quality gates in opportunity scoring

| Tag ID | Name | Trigger | Update Frequency |
|--------|------|---------|------------------|
| `quality-gate-pass` | Quality Gate Pass | Fundamentals ≥ 0.6 AND long-term ≥ 0.5 | Daily |
| `quality-gate-fail` | Quality Gate Fail | Fundamentals < 0.6 OR long-term < 0.5 | Daily |
| `quality-value` | Quality Value | `value-opportunity` AND `high-quality` | Hourly |

**Usage:**
- Fast filtering: Only consider `quality-gate-pass` for opportunity buys
- Quality gates: Reject `quality-gate-fail` even if opportunity score is high
- Priority boost: `quality-value` gets 1.3x priority boost

### 2. Bubble Detection Tags (NEW - for CAGR Scoring)

**Purpose**: Support bubble detection in CAGR scoring

| Tag ID | Name | Trigger | Update Frequency |
|--------|------|---------|------------------|
| `bubble-risk` | Bubble Risk | CAGR > 16.5% AND (Sharpe < 0.5 OR Sortino < 0.5 OR volatility > 0.40 OR fundamentals < 0.6) | Daily |
| `quality-high-cagr` | Quality High CAGR | CAGR > 15% AND Sharpe ≥ 0.5 AND Sortino ≥ 0.5 AND volatility ≤ 0.40 AND fundamentals ≥ 0.6 | Daily |
| `poor-risk-adjusted` | Poor Risk-Adjusted | Sharpe < 0.5 OR Sortino < 0.5 | Daily |
| `high-sharpe` | High Sharpe | Sharpe ratio ≥ 1.5 | Daily |
| `high-sortino` | High Sortino | Sortino ratio ≥ 1.5 | Daily |

**Usage:**
- CAGR scoring: Cap score at 0.6 for `bubble-risk`
- CAGR scoring: Reward `quality-high-cagr` with monotonic scoring
- Fast filtering: Exclude `bubble-risk` from opportunity buys

### 3. Value Trap Tags (NEW - for Opportunity Scoring)

**Purpose**: Support value trap detection

| Tag ID | Name | Trigger | Update Frequency |
|--------|------|---------|------------------|
| `value-trap` | Value Trap | Cheap (P/E < market - 20%) AND (fundamentals < 0.6 OR long-term < 0.5 OR momentum < -0.05 OR volatility > 0.35) | Hourly |
| `quality-value` | Quality Value | `value-opportunity` AND `high-quality` | Hourly |

**Usage:**
- Opportunity scoring: Penalize `value-trap` by 30%
- Fast filtering: Exclude `value-trap` from opportunity buys
- Priority boost: `quality-value` gets 1.3x priority boost

### 4. Total Return Tags (NEW - for Dividend Scoring)

**Purpose**: Support total return calculation (growth + dividend)

| Tag ID | Name | Trigger | Update Frequency |
|--------|------|---------|------------------|
| `high-total-return` | High Total Return | (CAGR + dividend yield) ≥ 15% | Daily |
| `excellent-total-return` | Excellent Total Return | (CAGR + dividend yield) ≥ 18% | Daily |
| `dividend-total-return` | Dividend Total Return | Dividend yield ≥ 8% AND CAGR ≥ 5% | Daily |
| `moderate-total-return` | Moderate Total Return | (CAGR + dividend yield) ≥ 12% | Daily |

**Usage:**
- Dividend scoring: Boost score for `high-total-return` and `excellent-total-return`
- Fast filtering: Prioritize `high-total-return` for opportunity buys
- Your example: 5% growth + 10% dividend = 15% → `high-total-return` tag

### 5. Optimizer Alignment Tags (NEW - for Optimizer Integration)

**Purpose**: Support optimizer-planner alignment

| Tag ID | Name | Trigger | Update Frequency |
|--------|------|---------|------------------|
| `underweight` | Underweight | Position weight < target weight - 2% | Hourly |
| `target-aligned` | Target Aligned | Position weight within ±1% of target | Hourly |
| `needs-rebalance` | Needs Rebalance | Position weight deviation > 3% from target | Hourly |
| `slightly-overweight` | Slightly Overweight | Position weight > target + 1% AND ≤ target + 3% | Hourly |
| `slightly-underweight` | Slightly Underweight | Position weight < target - 1% AND ≥ target - 3% | Hourly |

**Usage:**
- Evaluation: Boost score for `target-aligned` positions
- Rebalancing: Prioritize `needs-rebalance` positions
- Optimizer alignment: Weight alignment score based on these tags

### 6. Regime-Specific Tags (NEW - for Regime-Based Adjustments)

**Purpose**: Support regime-based risk adjustments

| Tag ID | Name | Trigger | Update Frequency |
|--------|------|---------|------------------|
| `regime-bear-safe` | Bear Market Safe | Volatility < 0.20 AND fundamentals > 0.75 AND max drawdown < 20% | Daily |
| `regime-bull-growth` | Bull Market Growth | CAGR > 12% AND fundamentals > 0.7 AND momentum > 0 | Daily |
| `regime-sideways-value` | Sideways Value | `value-opportunity` AND fundamentals > 0.75 | Hourly |
| `regime-volatile` | Regime Volatile | Volatility > 0.30 OR volatility spike | 10 minutes |

**Usage:**
- Bear market: Favor `regime-bear-safe` positions
- Bull market: Favor `regime-bull-growth` positions
- Sideways: Favor `regime-sideways-value` positions
- Risk reduction: Reduce exposure to `regime-volatile` in bear markets

---

## Enhanced Tag Assignment Logic

### Updated Tag Assigner

```go
// Enhanced tag assignment with algorithm support
func (ta *TagAssigner) AssignTagsForSecurity(input AssignTagsInput) ([]string, error) {
    var tags []string
    
    // ... existing tag assignment logic ...
    
    // === NEW: QUALITY GATE TAGS ===
    
    // Quality gate pass/fail
    fundamentalsScore := getScore(input.GroupScores, "fundamentals")
    longTermScore := getScore(input.GroupScores, "long_term")
    
    if fundamentalsScore >= 0.6 && longTermScore >= 0.5 {
        tags = append(tags, "quality-gate-pass")
    } else {
        tags = append(tags, "quality-gate-fail")
    }
    
    // Quality value (high quality + value opportunity)
    if contains(tags, "high-quality") && contains(tags, "value-opportunity") {
        tags = append(tags, "quality-value")
    }
    
    // === NEW: BUBBLE DETECTION TAGS ===
    
    cagrValue := getSubScore(input.SubScores, "long_term", "cagr")
    sharpeRatio := getSubScore(input.SubScores, "long_term", "sharpe_raw")
    sortinoRatio := getSubScore(input.SubScores, "long_term", "sortino")
    volatility := 0.0
    if input.Volatility != nil {
        volatility = *input.Volatility
    }
    
    // Bubble risk: High CAGR with poor risk metrics
    if cagrValue > 0.165 { // 16.5% for 11% target
        isBubble := false
        if sharpeRatio < 0.5 || sortinoRatio < 0.5 || volatility > 0.40 || fundamentalsScore < 0.6 {
            isBubble = true
        }
        
        if isBubble {
            tags = append(tags, "bubble-risk")
        } else {
            tags = append(tags, "quality-high-cagr")
        }
    }
    
    // Risk-adjusted tags
    if sharpeRatio >= 1.5 {
        tags = append(tags, "high-sharpe")
    }
    if sortinoRatio >= 1.5 {
        tags = append(tags, "high-sortino")
    }
    if sharpeRatio < 0.5 || sortinoRatio < 0.5 {
        tags = append(tags, "poor-risk-adjusted")
    }
    
    // === NEW: VALUE TRAP TAGS ===
    
    peVsMarket := 0.0
    if input.MarketAvgPE > 0 && input.PERatio != nil {
        peVsMarket = (*input.PERatio - input.MarketAvgPE) / input.MarketAvgPE
    }
    
    momentumScore := getSubScore(input.SubScores, "short_term", "momentum")
    
    // Value trap: Cheap but declining
    if peVsMarket < -0.20 {
        isValueTrap := false
        if fundamentalsScore < 0.6 || longTermScore < 0.5 || momentumScore < -0.05 || volatility > 0.35 {
            isValueTrap = true
        }
        
        if isValueTrap {
            tags = append(tags, "value-trap")
        }
    }
    
    // === NEW: TOTAL RETURN TAGS ===
    
    dividendYield := 0.0
    if input.DividendYield != nil {
        dividendYield = *input.DividendYield
    }
    
    totalReturn := cagrValue + dividendYield
    
    if totalReturn >= 0.18 {
        tags = append(tags, "excellent-total-return")
    } else if totalReturn >= 0.15 {
        tags = append(tags, "high-total-return")
    } else if totalReturn >= 0.12 {
        tags = append(tags, "moderate-total-return")
    }
    
    // Dividend total return (your example: 5% growth + 10% dividend)
    if dividendYield >= 0.08 && cagrValue >= 0.05 {
        tags = append(tags, "dividend-total-return")
    }
    
    // === NEW: OPTIMIZER ALIGNMENT TAGS ===
    
    positionWeight := 0.0
    if input.PositionWeight != nil {
        positionWeight = *input.PositionWeight
    }
    
    targetWeight := 0.0
    if input.TargetWeight != nil {
        targetWeight = *input.TargetWeight
    }
    
    if targetWeight > 0 {
        deviation := positionWeight - targetWeight
        
        if math.Abs(deviation) <= 0.01 {
            tags = append(tags, "target-aligned")
        } else if deviation > 0.03 {
            tags = append(tags, "needs-rebalance")
            if deviation > 0.02 {
                tags = append(tags, "overweight") // Keep existing
            }
        } else if deviation < -0.03 {
            tags = append(tags, "needs-rebalance")
            tags = append(tags, "underweight")
        } else if deviation > 0.01 {
            tags = append(tags, "slightly-overweight")
        } else if deviation < -0.01 {
            tags = append(tags, "slightly-underweight")
        }
    }
    
    // === NEW: REGIME-SPECIFIC TAGS ===
    
    maxDrawdown := 0.0
    if input.MaxDrawdown != nil {
        maxDrawdown = *input.MaxDrawdown
    }
    
    // Bear market safe
    if volatility < 0.20 && fundamentalsScore > 0.75 && maxDrawdown < 20.0 {
        tags = append(tags, "regime-bear-safe")
    }
    
    // Bull market growth
    if cagrValue > 0.12 && fundamentalsScore > 0.7 && momentumScore > 0 {
        tags = append(tags, "regime-bull-growth")
    }
    
    // Sideways value
    if contains(tags, "value-opportunity") && fundamentalsScore > 0.75 {
        tags = append(tags, "regime-sideways-value")
    }
    
    // Regime volatile
    if volatility > 0.30 || contains(tags, "volatility-spike") {
        tags = append(tags, "regime-volatile")
    }
    
    // Remove duplicates
    tags = removeDuplicates(tags)
    
    return tags, nil
}
```

---

## Tag Usage in Algorithm Improvements

### 1. Quality Gates (Opportunity Scoring)

```go
// Fast quality gate check using tags
func hasQualityGate(security Security) bool {
    return security.HasTag("quality-gate-pass")
}

// Reject value traps
func isValueTrap(security Security) bool {
    return security.HasTag("value-trap")
}

// Quality value gets priority boost
func getPriorityBoost(security Security) float64 {
    if security.HasTag("quality-value") {
        return 1.3
    }
    return 1.0
}
```

### 2. Bubble Detection (CAGR Scoring)

```go
// Fast bubble detection using tags
func isBubble(security Security) bool {
    return security.HasTag("bubble-risk")
}

// Quality high CAGR gets monotonic scoring
func isQualityHighCAGR(security Security) bool {
    return security.HasTag("quality-high-cagr")
}
```

### 3. Total Return (Dividend Scoring)

```go
// Fast total return check using tags
func hasHighTotalReturn(security Security) bool {
    return security.HasTag("high-total-return") || 
           security.HasTag("excellent-total-return")
}

// Your example: 5% growth + 10% dividend
func isDividendTotalReturn(security Security) bool {
    return security.HasTag("dividend-total-return")
}
```

### 4. Optimizer Alignment (Evaluation)

```go
// Fast optimizer alignment check
func getOptimizerAlignmentScore(security Security) float64 {
    if security.HasTag("target-aligned") {
        return 1.0
    } else if security.HasTag("slightly-overweight") || security.HasTag("slightly-underweight") {
        return 0.8
    } else if security.HasTag("needs-rebalance") {
        return 0.5
    }
    return 0.5 // Neutral
}
```

### 5. Regime-Based Adjustments (Evaluation)

```go
// Fast regime-based filtering
func getRegimeTags(regime MarketRegime) []string {
    switch regime {
    case MarketRegimeBear:
        return []string{"regime-bear-safe"}
    case MarketRegimeBull:
        return []string{"regime-bull-growth"}
    case MarketRegimeSideways:
        return []string{"regime-sideways-value"}
    }
    return []string{}
}
```

---

## Tag Update Frequency (Enhanced)

### Very Dynamic (10 minutes)
- `oversold`, `overbought`, `below-ema`, `above-ema`
- `bollinger-oversold`, `volatility-spike`
- `near-52w-high`, `below-52w-high`
- `regime-volatile`

### Dynamic (Hourly)
- `value-opportunity`, `deep-value`, `undervalued-pe`
- `value-trap`, `quality-value`
- `positive-momentum`, `recovery-candidate`
- `overvalued`, `overweight`, `underweight`
- `needs-rebalance`, `slightly-overweight`, `slightly-underweight`
- `regime-sideways-value`

### Stable (Daily)
- `high-quality`, `strong-fundamentals`, `stable`
- `quality-gate-pass`, `quality-gate-fail`
- `bubble-risk`, `quality-high-cagr`
- `high-sharpe`, `high-sortino`, `poor-risk-adjusted`
- `high-total-return`, `excellent-total-return`
- `dividend-total-return`, `moderate-total-return`
- `regime-bear-safe`, `regime-bull-growth`
- All existing quality/characteristic tags

### Very Stable (Weekly)
- `long-term` (unchanged)

---

## Summary

### New Tags Added: 20

**Quality Gate (3):**
- `quality-gate-pass`, `quality-gate-fail`, `quality-value`

**Bubble Detection (4):**
- `bubble-risk`, `quality-high-cagr`, `poor-risk-adjusted`, `high-sharpe`, `high-sortino`

**Value Trap (1):**
- `value-trap` (plus `quality-value` above)

**Total Return (4):**
- `high-total-return`, `excellent-total-return`, `dividend-total-return`, `moderate-total-return`

**Optimizer Alignment (4):**
- `underweight`, `target-aligned`, `needs-rebalance`, `slightly-overweight`, `slightly-underweight`

**Regime-Specific (4):**
- `regime-bear-safe`, `regime-bull-growth`, `regime-sideways-value`, `regime-volatile`

### Benefits

1. **Fast Quality Gates**: `quality-gate-pass` enables instant filtering
2. **Fast Bubble Detection**: `bubble-risk` enables instant rejection
3. **Fast Value Trap Detection**: `value-trap` enables instant rejection
4. **Fast Total Return**: `high-total-return` enables instant prioritization
5. **Fast Optimizer Alignment**: `target-aligned` enables instant scoring
6. **Fast Regime Filtering**: Regime tags enable instant adaptation

### Implementation Priority

1. **Phase 1**: Quality gate tags + value trap tags (enables quality gates)
2. **Phase 2**: Bubble detection tags + total return tags (enables CAGR/dividend improvements)
3. **Phase 3**: Optimizer alignment tags (enables optimizer integration)
4. **Phase 4**: Regime-specific tags (enables regime-based adjustments)

---

## Migration Strategy

1. **Add new tags to database** (migration)
2. **Update TagAssigner** with new tag logic
3. **Update tag update frequencies** (per-tag scheduling)
4. **Test tag assignment** (validate triggers)
5. **Update calculators** to use new tags
6. **Update evaluation** to use new tags

All existing tags remain unchanged - this is purely additive.

