# Tag-Based Optimization Strategy

## Overview

This document outlines the strategy for using the security tags system to dramatically improve planner performance through intelligent pre-filtering and focused calculations.

**Note**: This document has been updated to incorporate the enhanced tag system (see `TAG_SYSTEM_ENHANCEMENT.md`). The enhanced tags support both performance optimization (fast filtering) and algorithm improvements (quality gates, bubble detection, value trap detection, total return, optimizer alignment, regime adaptation).

## Integration with Enhanced Tags

The tag-based optimization strategy now leverages **20+ new enhanced tags** that support algorithm improvements:

- **Quality Gate Tags**: `quality-gate-pass`, `quality-gate-fail`, `quality-value`
- **Bubble Detection Tags**: `bubble-risk`, `quality-high-cagr`, `high-sharpe`, `high-sortino`, `poor-risk-adjusted`
- **Value Trap Tags**: `value-trap`
- **Total Return Tags**: `high-total-return`, `excellent-total-return`, `dividend-total-return`, `moderate-total-return`
- **Optimizer Alignment Tags**: `underweight`, `target-aligned`, `needs-rebalance`, `slightly-overweight`, `slightly-underweight`
- **Regime-Specific Tags**: `regime-bear-safe`, `regime-bull-growth`, `regime-sideways-value`, `regime-volatile`

These enhanced tags enable:
1. **Fast Quality Gates**: Instant filtering using `quality-gate-pass` (no expensive calculations)
2. **Fast Bubble Detection**: Instant rejection using `bubble-risk` tag
3. **Fast Value Trap Detection**: Instant rejection using `value-trap` tag
4. **Fast Total Return**: Instant prioritization using `high-total-return` tag
5. **Fast Optimizer Alignment**: Instant scoring using `target-aligned` tag
6. **Fast Regime Filtering**: Instant adaptation using regime tags

**Result**: 5-7x faster planning + better decisions (algorithm improvements) = best of both worlds.

## Tag Update Frequency Strategy

Tags are categorized by volatility and update frequency to optimize both freshness and performance:

### Very Dynamic Tags (5-15 minutes)
**Price/Technical Tags** - Change intraday based on market movements

- `oversold`, `overbought` (RSI-based)
- `below-ema`, `above-ema` (price vs EMA200)
- `bollinger-oversold` (Bollinger position)
- `volatility-spike` (current vs historical volatility)
- `near-52w-high`, `below-52w-high` (price-based)
- `valuation-stretch` (distance from EMA)

**Update Frequency:** Every 10 minutes during market hours

### Dynamic Tags (Hourly)
**Opportunity/Risk Tags** - Change with market conditions

- `value-opportunity`, `deep-value`, `undervalued-pe` (price + P/E)
- `positive-momentum`, `recovery-candidate` (momentum-based)
- `overvalued` (P/E + price)
- `overweight`, `concentration-risk` (portfolio position weight)
- `unsustainable-gains` (return + volatility)

**Update Frequency:** Every hour

### Stable Tags (Daily)
**Quality/Characteristic Tags** - Change slowly

- `high-quality`, `stable`, `strong-fundamentals` (fundamentals-based)
- `consistent-grower` (CAGR + consistency)
- `high-dividend`, `dividend-opportunity`, `dividend-grower` (dividend-based)
- `high-score`, `good-opportunity` (score-based)
- `volatile`, `high-volatility` (volatility-based)
- `underperforming`, `stagnant`, `high-drawdown` (performance-based)
- `low-risk`, `medium-risk`, `high-risk` (risk profile)
- `growth`, `value`, `dividend-focused` (growth profile)
- `short-term-opportunity` (technical + opportunity)

**Update Frequency:** Daily (3:00 AM)

### Very Stable Tags (Weekly)
**Long-term Characteristics** - Rarely change

- `long-term` (long-term score + consistency)

**Update Frequency:** Weekly (Sunday 3:00 AM)

## Hybrid Calculator Architecture

### Design Philosophy

**Tag-Based Pre-Filtering → Focused Calculations → Smart Prioritization**

1. **Fast Pre-Filtering**: Use tags to reduce candidate set from 100+ securities to 10-20
2. **Focused Calculations**: Run detailed calculations only on filtered candidates
3. **Smart Prioritization**: Use tag combinations to intelligently boost priority

### Architecture Components

#### 1. TagBasedFilter Service

Smart tag selection based on context (market conditions, cash availability, etc.)

```go
type TagBasedFilter struct {
    securityRepo *universe.SecurityRepository
    log          zerolog.Logger
}

// GetOpportunityCandidates uses tags to quickly identify candidates
func (f *TagBasedFilter) GetOpportunityCandidates(ctx *domain.OpportunityContext) ([]string, error)

// selectOpportunityTags intelligently selects tags based on context
func (f *TagBasedFilter) selectOpportunityTags(ctx *domain.OpportunityContext) []string
```

**Key Features:**
- Adapts tag selection based on available cash
- Considers market volatility conditions
- Balances value, quality, and dividend opportunities
- No configuration needed - fully automatic

#### 2. Hybrid Calculators

Replace existing calculators with hybrid versions that:
- Use tags for fast pre-filtering (10-50ms)
- Run focused calculations on filtered set (100-500ms vs 2-5s)
- Apply smart priority boosting based on tag combinations

**Calculators to Convert:**
- `HybridOpportunityBuysCalculator`
- `HybridProfitTakingCalculator`
- `HybridAveragingDownCalculator`
- `HybridRebalanceCalculator`

#### 3. Tag Update Scheduler

Per-tag frequency updates with smart caching:

```go
type TagUpdateScheduler struct {
    tagAssigner  *universe.TagAssigner
    securityRepo *universe.SecurityRepository
    log          zerolog.Logger
}

// UpdateTagsForSecurity updates only tags that need updating
func (s *TagUpdateScheduler) UpdateTagsForSecurity(
    security universe.Security,
    requiredTags []string, // Tags needed for current planning cycle
) error
```

**Key Features:**
- Only updates tags that need updating (efficiency)
- Respects per-tag update frequencies
- Smart caching to avoid redundant updates

## Implementation Plan

### Phase 1: Foundation (Core Infrastructure)

#### Step 1.1: Add Tag Query Methods to SecurityRepository

```go
// GetByTags - fast SQL query with indexed tags
// Returns securities matching any of the provided tags
func (r *SecurityRepository) GetByTags(tagIDs []string) ([]Security, error)

// GetPositionsByTags - get portfolio positions with specific tags
// Returns securities that are in portfolio AND have these tags
func (r *SecurityRepository) GetPositionsByTags(tagIDs []string) ([]Security, error)

// GetTagsForSecurity - get all tags for a security
func (r *SecurityRepository) GetTagsForSecurity(symbol string) ([]string, error)

// GetByTagsWithScores - get securities by tags with scores pre-loaded
// Optimized for planner use
func (r *SecurityRepository) GetByTagsWithScores(tagIDs []string) ([]SecurityWithScore, error)
```

**SQL Implementation:**
```sql
-- GetByTags query (uses indexed security_tags table)
SELECT DISTINCT s.*
FROM securities s
INNER JOIN security_tags st ON s.symbol = st.symbol
WHERE st.tag_id IN (?, ?, ...)
AND s.active = 1
ORDER BY s.symbol;
```

#### Step 1.2: Create TagBasedFilter Service

```go
package opportunities

type TagBasedFilter struct {
    securityRepo *universe.SecurityRepository
    log          zerolog.Logger
}

func NewTagBasedFilter(securityRepo *universe.SecurityRepository, log zerolog.Logger) *TagBasedFilter {
    return &TagBasedFilter{
        securityRepo: securityRepo,
        log:          log.With().Str("component", "tag_filter").Logger(),
    }
}

// GetOpportunityCandidates uses tags to quickly identify candidates
func (f *TagBasedFilter) GetOpportunityCandidates(ctx *domain.OpportunityContext) ([]string, error) {
    tags := f.selectOpportunityTags(ctx)
    
    candidates, err := f.securityRepo.GetByTags(tags)
    if err != nil {
        return nil, err
    }
    
    symbols := make([]string, len(candidates))
    for i, c := range candidates {
        symbols[i] = c.Symbol
    }
    
    return symbols, nil
}

// selectOpportunityTags intelligently selects tags based on context
func (f *TagBasedFilter) selectOpportunityTags(ctx *domain.OpportunityContext) []string {
    tags := []string{}
    
    // Always include quality gates
    tags = append(tags, "high-quality", "good-opportunity")
    
    // Add value opportunities if we have cash
    if ctx.AvailableCashEUR > 1000 {
        tags = append(tags, "value-opportunity", "deep-value")
    }
    
    // Add technical opportunities if market is volatile
    if f.isMarketVolatile(ctx) {
        tags = append(tags, "oversold", "below-ema", "recovery-candidate")
    }
    
    // Add dividend opportunities
    tags = append(tags, "dividend-opportunity", "high-dividend")
    
    return tags
}

// isMarketVolatile determines if market conditions are volatile
func (f *TagBasedFilter) isMarketVolatile(ctx *domain.OpportunityContext) bool {
    // Check if many securities have volatility-spike tag
    volatileSecurities, _ := f.securityRepo.GetByTags([]string{"volatility-spike"})
    return len(volatileSecurities) > 5 // Threshold for "volatile market"
}
```

### Phase 2: Hybrid Calculators

#### Step 2.1: HybridOpportunityBuysCalculator

```go
package calculators

type HybridOpportunityBuysCalculator struct {
    *BaseCalculator
    tagFilter    *TagBasedFilter
    securityRepo *universe.SecurityRepository
}

func NewHybridOpportunityBuysCalculator(
    tagFilter *TagBasedFilter,
    securityRepo *universe.SecurityRepository,
    log zerolog.Logger,
) *HybridOpportunityBuysCalculator {
    return &HybridOpportunityBuysCalculator{
        BaseCalculator: NewBaseCalculator(log, "hybrid_opportunity_buys"),
        tagFilter:      tagFilter,
        securityRepo:   securityRepo,
    }
}

func (c *HybridOpportunityBuysCalculator) Calculate(
    ctx *domain.OpportunityContext,
    params map[string]interface{},
) ([]domain.ActionCandidate, error) {
    // Step 1: Fast tag-based pre-filtering (10-50ms)
    candidateSymbols, err := c.tagFilter.GetOpportunityCandidates(ctx)
    if err != nil {
        return nil, err
    }
    
    if len(candidateSymbols) == 0 {
        return nil, nil
    }
    
    c.log.Debug().
        Int("tag_candidates", len(candidateSymbols)).
        Msg("Tag-based pre-filtering complete")
    
    // Step 2: Focused calculations on filtered set (100-500ms vs 2-5s)
    var candidates []domain.ActionCandidate
    
    for _, symbol := range candidateSymbols {
        // Get security info
        security, ok := ctx.StocksBySymbol[symbol]
        if !ok {
            continue
        }
        
        // Skip if recently bought
        if ctx.RecentlyBought[symbol] {
            continue
        }
        
        // Get current price
        currentPrice, ok := ctx.CurrentPrices[symbol]
        if !ok || currentPrice <= 0 {
            continue
        }
        
        // Get score (already calculated, just lookup)
        score, ok := ctx.SecurityScores[symbol]
        if !ok || score < 0.7 {
            continue
        }
        
        // Focused calculation: precise quantity and value
        targetValue := 500.0
        if targetValue > ctx.AvailableCashEUR {
            targetValue = ctx.AvailableCashEUR
        }
        
        quantity := int(targetValue / currentPrice)
        if quantity == 0 {
            quantity = 1
        }
        
        valueEUR := float64(quantity) * currentPrice
        transactionCost := ctx.TransactionCostFixed + (valueEUR * ctx.TransactionCostPercent)
        totalCostEUR := valueEUR + transactionCost
        
        if totalCostEUR > ctx.AvailableCashEUR {
            continue
        }
        
        // Get tags for this security to boost priority
        tags, _ := c.securityRepo.GetTagsForSecurity(symbol)
        priority := c.calculatePriority(score, tags)
        
        candidate := domain.ActionCandidate{
            Side:     "BUY",
            Symbol:   symbol,
            Name:     security.Name,
            Quantity: quantity,
            Price:    currentPrice,
            ValueEUR: totalCostEUR,
            Currency: string(security.Currency),
            Priority: priority,
            Reason:   fmt.Sprintf("Tag-filtered opportunity: score %.2f", score),
            Tags:     tags,
        }
        
        candidates = append(candidates, candidate)
    }
    
    // Step 3: Sort by priority
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].Priority > candidates[j].Priority
    })
    
    // Step 4: Limit to top N
    maxPositions := GetIntParam(params, "max_positions", 5)
    if maxPositions > 0 && len(candidates) > maxPositions {
        candidates = candidates[:maxPositions]
    }
    
    c.log.Info().
        Int("candidates", len(candidates)).
        Int("filtered_from", len(candidateSymbols)).
        Msg("Hybrid opportunity buys calculated")
    
    return candidates, nil
}

// calculatePriority intelligently boosts priority based on tag combinations
func (c *HybridOpportunityBuysCalculator) calculatePriority(
    score float64,
    tags []string,
) float64 {
    priority := score
    
    // High-quality value opportunities get boost
    if contains(tags, "high-quality") && contains(tags, "value-opportunity") {
        priority *= 1.3
    }
    
    // Deep value gets boost
    if contains(tags, "deep-value") {
        priority *= 1.2
    }
    
    // Oversold high-quality gets boost
    if contains(tags, "oversold") && contains(tags, "high-quality") {
        priority *= 1.15
    }
    
    // Recovery candidates get moderate boost
    if contains(tags, "recovery-candidate") {
        priority *= 1.1
    }
    
    return math.Min(1.0, priority)
}
```

#### Step 2.2: HybridProfitTakingCalculator

Enhanced with optimizer alignment tags:

```go
func (c *HybridProfitTakingCalculator) Calculate(ctx, params) {
    // Use tags to identify profit-taking candidates
    sellTags := []string{
        "overvalued", "near-52w-high", "overbought", // Price-based
        "overweight", "concentration-risk",           // Portfolio-based
        "needs-rebalance",                           // Optimizer alignment - NEW
        "bubble-risk",                               // Bubble detection - NEW
    }
    
    candidates := c.securityRepo.GetPositionsByTags(sellTags)
    
    // Filter: Keep only positions with gains (profit-taking)
    // Focused calculation on filtered set
}
```

**Enhanced Tags Used:**
- `needs-rebalance` - Optimizer alignment (NEW)
- `bubble-risk` - Sell bubbles before they pop (NEW)

#### Step 2.3: HybridAveragingDownCalculator

Enhanced with quality gates and value trap detection:

```go
func (c *HybridAveragingDownCalculator) Calculate(ctx, params) {
    // Use tags to identify averaging-down opportunities
    avgDownTags := []string{
        "recovery-candidate",      // Quality company, temporary weakness
        "value-opportunity",        // Cheap price
        "quality-gate-pass",        // Quality gate - NEW
        "quality-value",           // Quality + value - NEW
    }
    
    candidates := c.securityRepo.GetPositionsByTags(avgDownTags)
    
    // CRITICAL: Exclude value traps
    candidates = c.filterOutNegativeTags(candidates, []string{"value-trap"})
    
    // Focused calculation: Only on quality positions with losses
}
```

**Enhanced Tags Used:**
- `quality-gate-pass` - Only average down on quality (NEW)
- `quality-value` - Best averaging-down candidates (NEW)
- Exclude `value-trap` - Don't average down into traps (NEW)

#### Step 2.4: HybridRebalanceCalculator

Enhanced with optimizer alignment tags:

```go
func (c *HybridRebalanceCalculator) Calculate(ctx, params) {
    // SELL tags: Overweight positions
    sellTags := []string{
        "overweight", "concentration-risk",
        "needs-rebalance",           // Optimizer alignment - NEW
        "slightly-overweight",        // Optimizer alignment - NEW
    }
    
    // BUY tags: Underweight quality positions
    buyTags := []string{
        "underweight",                // Optimizer alignment - NEW
        "slightly-underweight",       // Optimizer alignment - NEW
        "high-quality",               // Quality requirement
        "quality-gate-pass",          // Quality gate - NEW
        "target-aligned",             // Already aligned (skip) - NEW
    }
    
    sellCandidates := c.securityRepo.GetPositionsByTags(sellTags)
    buyCandidates := c.securityRepo.GetByTags(buyTags)
    
    // Focused calculation: Rebalance toward optimizer targets
}
```

**Enhanced Tags Used:**
- `underweight`, `slightly-underweight` - Optimizer alignment (NEW)
- `needs-rebalance`, `slightly-overweight` - Rebalancing triggers (NEW)
- `target-aligned` - Skip already-aligned positions (NEW)
- `quality-gate-pass` - Only rebalance into quality (NEW)

### Phase 3: Tag Update Scheduler

#### Step 3.1: Per-Tag Update Frequencies

```go
package scheduler

type TagUpdateFrequency struct {
    TagIDs      []string
    Frequency   time.Duration
    Description string
}

var TagUpdateFrequencies = []TagUpdateFrequency{
    // Very dynamic: 10 minutes
    {
        TagIDs: []string{
            "oversold", "overbought", "below-ema", "above-ema",
            "bollinger-oversold", "volatility-spike", "near-52w-high",
            "below-52w-high", "valuation-stretch",
            "regime-volatile", // NEW
        },
        Frequency:   10 * time.Minute,
        Description: "Price/technical tags",
    },
    // Dynamic: Hourly
    {
        TagIDs: []string{
            "value-opportunity", "deep-value", "undervalued-pe",
            "value-trap", "quality-value", // NEW
            "positive-momentum", "recovery-candidate", "overvalued",
            "overweight", "underweight", "concentration-risk", // ENHANCED
            "needs-rebalance", "slightly-overweight", "slightly-underweight", // NEW
            "unsustainable-gains",
            "regime-sideways-value", // NEW
        },
        Frequency:   1 * time.Hour,
        Description: "Opportunity/risk tags",
    },
    // Stable: Daily
    {
        TagIDs: []string{
            // Quality tags
            "high-quality", "stable", "strong-fundamentals",
            "quality-gate-pass", "quality-gate-fail", // NEW
            "consistent-grower",
            
            // Bubble detection tags
            "bubble-risk", "quality-high-cagr", // NEW
            "high-sharpe", "high-sortino", "poor-risk-adjusted", // NEW
            
            // Total return tags
            "high-total-return", "excellent-total-return", // NEW
            "dividend-total-return", "moderate-total-return", // NEW
            
            // Dividend tags
            "high-dividend", "dividend-opportunity", "dividend-grower",
            
            // Score tags
            "high-score", "good-opportunity",
            
            // Risk tags
            "volatile", "high-volatility", "underperforming", "stagnant",
            "high-drawdown", "low-risk", "medium-risk", "high-risk",
            
            // Profile tags
            "growth", "value", "dividend-focused", "short-term-opportunity",
            
            // Regime tags
            "regime-bear-safe", "regime-bull-growth", // NEW
            
            // Optimizer tags
            "target-aligned", // NEW
        },
        Frequency:   24 * time.Hour,
        Description: "Quality/characteristic tags",
    },
    // Very stable: Weekly
    {
        TagIDs: []string{
            "long-term",
        },
        Frequency:   7 * 24 * time.Hour,
        Description: "Long-term characteristics",
    },
}
```

#### Step 3.2: Smart Tag Update Job

```go
type SmartTagUpdateJob struct {
    tagAssigner  *universe.TagAssigner
    securityRepo *universe.SecurityRepository
    log          zerolog.Logger
    lastUpdate   map[string]time.Time // symbol -> last update time
}

// UpdateTagsForSecurity updates only tags that need updating
func (j *SmartTagUpdateJob) UpdateTagsForSecurity(
    security universe.Security,
    requiredTags []string, // Tags needed for current planning cycle
) error {
    // Determine which tags need updating based on frequency
    tagsToUpdate := j.getTagsNeedingUpdate(security.Symbol, requiredTags)
    
    if len(tagsToUpdate) == 0 {
        // All tags are fresh, skip update
        return nil
    }
    
    // Update only the tags that need it
    return j.updateSpecificTags(security, tagsToUpdate)
}
```

## Expected Performance Improvements

| Component | Current | With Tags | Improvement |
|-----------|---------|-----------|-------------|
| Opportunity identification | 2-5s | 100-300ms | 10-20x faster |
| Danger detection | 1-2s | 50-100ms | 10-20x faster |
| Sequence generation | 5-10s | 1-2s | 5x faster |
| **Total planning time** | **10-15s** | **2-3s** | **5-7x faster** |

## Key Benefits

1. **Fast Pre-Filtering**: Tags reduce candidate set from 100+ to 10-20 securities
2. **Focused Calculations**: Detailed work only on filtered candidates
3. **Smart Prioritization**: Tag combinations boost priority intelligently
4. **Efficient Updates**: Only update tags that need updating
5. **No Configuration**: System adapts automatically to market conditions
6. **Algorithm Support**: Enhanced tags enable quality gates, bubble detection, value trap detection
7. **Optimizer Alignment**: Tags support optimizer-planner integration
8. **Regime Adaptation**: Tags enable regime-based risk adjustments

## Migration Strategy

1. **Phase 1**: Add tag query methods (non-breaking)
2. **Phase 2**: Create hybrid calculators alongside existing ones
3. **Phase 3**: Switch planner to use hybrid calculators
4. **Phase 4**: Remove old calculators after validation

## Testing Strategy

1. **Unit Tests**: Tag query methods, tag filter logic
2. **Integration Tests**: Hybrid calculators produce same results as original
3. **Performance Tests**: Verify 5-7x speedup
4. **Validation**: Run both systems in parallel for 1 week, compare outputs

## Enhanced Tag Usage Examples

### Quality Gates (Algorithm Improvement)

```go
// Fast quality gate check - no expensive calculations needed
func hasQualityGate(security Security) bool {
    tags, _ := securityRepo.GetTagsForSecurity(security.Symbol)
    return contains(tags, "quality-gate-pass")
}

// Reject value traps instantly
func isValueTrap(security Security) bool {
    tags, _ := securityRepo.GetTagsForSecurity(security.Symbol)
    return contains(tags, "value-trap")
}
```

### Bubble Detection (Algorithm Improvement)

```go
// Fast bubble detection - no expensive risk calculations needed
func isBubble(security Security) bool {
    tags, _ := securityRepo.GetTagsForSecurity(security.Symbol)
    return contains(tags, "bubble-risk")
}

// Quality high CAGR gets monotonic scoring
func isQualityHighCAGR(security Security) bool {
    tags, _ := securityRepo.GetTagsForSecurity(security.Symbol)
    return contains(tags, "quality-high-cagr")
}
```

### Total Return (Algorithm Improvement)

```go
// Fast total return check - no expensive CAGR + dividend calculations needed
func hasHighTotalReturn(security Security) bool {
    tags, _ := securityRepo.GetTagsForSecurity(security.Symbol)
    return contains(tags, "high-total-return") || 
           contains(tags, "excellent-total-return")
}

// Your example: 5% growth + 10% dividend = 15% total
func isDividendTotalReturn(security Security) bool {
    tags, _ := securityRepo.GetTagsForSecurity(security.Symbol)
    return contains(tags, "dividend-total-return")
}
```

### Optimizer Alignment (Algorithm Improvement)

```go
// Fast optimizer alignment check
func getOptimizerAlignmentScore(security Security) float64 {
    tags, _ := securityRepo.GetTagsForSecurity(security.Symbol)
    if contains(tags, "target-aligned") {
        return 1.0
    } else if contains(tags, "slightly-overweight") || contains(tags, "slightly-underweight") {
        return 0.8
    } else if contains(tags, "needs-rebalance") {
        return 0.5
    }
    return 0.5 // Neutral
}
```

### Regime-Based Filtering (Algorithm Improvement)

```go
// Fast regime-based filtering
func getRegimeFilterTags(regime MarketRegime) []string {
    switch regime {
    case MarketRegimeBear:
        return []string{"regime-bear-safe", "quality-gate-pass", "low-risk"}
    case MarketRegimeBull:
        return []string{"regime-bull-growth", "quality-high-cagr"}
    case MarketRegimeSideways:
        return []string{"regime-sideways-value", "quality-value"}
    }
    return []string{}
}
```

## Future Enhancements

1. **Tag-Based Pattern Matching**: Use tags to match trading patterns
2. **Tag-Based Evaluation Shortcuts**: Skip expensive calculations for tagged securities
3. **Tag-Based Caching**: Cache tag queries for even faster lookups
4. **Machine Learning**: Learn optimal tag combinations from historical performance
5. **Tag-Based Regime Detection**: Use tag distributions to detect market regimes
6. **Tag-Based Risk Scoring**: Use risk tags to calculate portfolio risk metrics faster

