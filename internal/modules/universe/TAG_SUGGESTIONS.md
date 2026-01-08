# Security Tag Suggestions

Based on analysis of internal scoring and opportunity identification tools, here are suggested tags for securities.

## Tag Categories

### Opportunity Tags (Positive Signals)

#### Value Opportunities
- **`value-opportunity`** - "Value Opportunity"
  - Trigger: Opportunity score > 0.7 AND (below 52W high > 20% OR P/E ratio < market avg - 20%)
  - Indicates: Security is trading at a discount relative to recent highs or market valuation

- **`deep-value`** - "Deep Value"
  - Trigger: Below 52W high > 30% AND P/E ratio < market avg - 20%
  - Indicates: Significant discount, potential bargain

- **`below-52w-high`** - "Below 52-Week High"
  - Trigger: Current price < 52W high by > 10%
  - Indicates: Trading below recent peak, potential entry point

- **`undervalued-pe`** - "Undervalued P/E"
  - Trigger: P/E ratio < market avg - 20%
  - Indicates: Trading at discount to market valuation

#### Quality Opportunities
- **`high-quality`** - "High Quality"
  - Trigger: Fundamentals score > 0.8 AND long-term score > 0.75
  - Indicates: Strong financials and consistent long-term performance

- **`stable`** - "Stable"
  - Trigger: Fundamentals score > 0.75 AND volatility < 0.20 AND consistency score > 0.8
  - Indicates: Low volatility, consistent performance, strong fundamentals

- **`consistent-grower`** - "Consistent Grower"
  - Trigger: Consistency score > 0.8 AND CAGR > 8%
  - Indicates: Steady, predictable growth over multiple timeframes

- **`strong-fundamentals`** - "Strong Fundamentals"
  - Trigger: Fundamentals score > 0.8
  - Indicates: Excellent financial strength (profit margin, debt/equity, current ratio)

#### Technical Opportunities
- **`oversold`** - "Oversold"
  - Trigger: RSI < 30
  - Indicates: Technical indicator suggests oversold condition, potential bounce

- **`below-ema`** - "Below 200-Day EMA"
  - Trigger: Current price < 200-day EMA by > 5%
  - Indicates: Trading below long-term trend, potential mean reversion

- **`bollinger-oversold`** - "Near Bollinger Lower Band"
  - Trigger: Bollinger position < 0.2
  - Indicates: Near lower Bollinger band, potential support level

#### Dividend Opportunities
- **`high-dividend`** - "High Dividend Yield"
  - Trigger: Dividend yield > 6%
  - Indicates: Attractive income opportunity

- **`dividend-opportunity`** - "Dividend Opportunity"
  - Trigger: Dividend score > 0.7 AND dividend yield > 3%
  - Indicates: Good dividend yield with consistent payout

- **`dividend-grower`** - "Dividend Grower"
  - Trigger: Dividend consistency score > 0.8 AND 5-year avg yield > current yield
  - Indicates: History of increasing dividends

#### Momentum Opportunities
- **`positive-momentum`** - "Positive Momentum"
  - Trigger: Short-term momentum score > 0.7 AND momentum between 5-15%
  - Indicates: Healthy upward price movement

- **`recovery-candidate`** - "Recovery Candidate"
  - Trigger: Recent momentum negative but fundamentals strong AND below 52W high > 15%
  - Indicates: Quality company experiencing temporary weakness

#### Score-Based Opportunities
- **`high-score`** - "High Overall Score"
  - Trigger: Total score > 0.75
  - Indicates: Strong across multiple scoring dimensions

- **`good-opportunity`** - "Good Opportunity"
  - Trigger: Total score > 0.7 AND opportunity score > 0.7
  - Indicates: High overall quality with strong opportunity metrics

### Danger Tags (Warning Signals)

#### Volatility Warnings
- **`volatile`** - "Volatile"
  - Trigger: Volatility > 0.30
  - Indicates: High price volatility, higher risk

- **`volatility-spike`** - "Volatility Spike"
  - Trigger: Current volatility > historical volatility * 1.5
  - Indicates: Unusual volatility increase, potential instability

- **`high-volatility`** - "High Volatility"
  - Trigger: Volatility > 0.40
  - Indicates: Very high volatility, significant risk

#### Overvaluation Warnings
- **`overvalued`** - "Overvalued"
  - Trigger: P/E ratio > market avg + 20% AND near 52W high (< 5% below)
  - Indicates: Trading at premium to market and recent highs

- **`near-52w-high`** - "Near 52-Week High"
  - Trigger: Current price > 52W high * 0.95
  - Indicates: Trading near all-time high, potential overextension

- **`above-ema`** - "Above 200-Day EMA"
  - Trigger: Current price > 200-day EMA by > 10%
  - Indicates: Extended above long-term trend, potential pullback

- **`overbought`** - "Overbought"
  - Trigger: RSI > 70
  - Indicates: Technical indicator suggests overbought condition

#### Instability Warnings
- **`instability-warning`** - "Instability Warning"
  - Trigger: Instability score > 0.7 (from sell scorer)
  - Indicates: Signs of unsustainable gains or bubble conditions

- **`unsustainable-gains`** - "Unsustainable Gains"
  - Trigger: Annualized return > 50% AND volatility spike > 1.5x
  - Indicates: Extremely high returns with increasing volatility

- **`valuation-stretch`** - "Valuation Stretch"
  - Trigger: Distance from MA200 > 30%
  - Indicates: Price significantly extended above moving average

#### Underperformance Warnings
- **`underperforming`** - "Underperforming"
  - Trigger: Annualized return < 0% AND held > 180 days
  - Indicates: Negative returns over extended period

- **`stagnant`** - "Stagnant"
  - Trigger: Annualized return < 5% AND held > 365 days
  - Indicates: Low returns, may need to free up capital

- **`high-drawdown`** - "High Drawdown"
  - Trigger: Max drawdown > 30%
  - Indicates: Significant price decline from peak

#### Portfolio Risk Warnings
- **`overweight`** - "Overweight in Portfolio"
  - Trigger: Position weight > target weight + 2% OR position > 10% of portfolio
  - Indicates: Concentration risk, may need rebalancing

- **`concentration-risk`** - "Concentration Risk"
  - Trigger: Position > 15% of portfolio OR country/industry overweight
  - Indicates: High concentration in single position or sector

### Characteristic Tags (Descriptive)

#### Risk Profile
- **`low-risk`** - "Low Risk"
  - Trigger: Volatility < 0.15 AND fundamentals > 0.7 AND drawdown < 20%
  - Indicates: Lower risk profile suitable for conservative allocation

- **`medium-risk`** - "Medium Risk"
  - Trigger: Volatility between 0.15-0.30 AND fundamentals > 0.6
  - Indicates: Moderate risk profile

- **`high-risk`** - "High Risk"
  - Trigger: Volatility > 0.30 OR fundamentals < 0.5
  - Indicates: Higher risk profile

#### Growth Profile
- **`growth`** - "Growth"
  - Trigger: CAGR > 15% AND fundamentals > 0.7
  - Indicates: High growth potential

- **`value`** - "Value"
  - Trigger: P/E < market avg AND opportunity score > 0.7
  - Indicates: Value-oriented investment

- **`dividend-focused`** - "Dividend Focused"
  - Trigger: Dividend yield > 4% AND dividend score > 0.7
  - Indicates: Income-focused investment

#### Time Horizon
- **`long-term`** - "Long-Term Promise"
  - Trigger: Long-term score > 0.75 AND consistency > 0.8
  - Indicates: Suitable for long-term holding

- **`short-term-opportunity`** - "Short-Term Opportunity"
  - Trigger: Technical score > 0.7 AND opportunity score > 0.7 AND momentum positive
  - Indicates: Near-term trading opportunity

## Tag Assignment Logic

Tags should be assigned based on the following priority:

1. **Primary Tags** (most important signal):
   - Opportunity tags (value-opportunity, high-quality, etc.)
   - Danger tags (volatile, overvalued, etc.)

2. **Characteristic Tags** (descriptive):
   - Risk profile (low-risk, medium-risk, high-risk)
   - Growth profile (growth, value, dividend-focused)
   - Time horizon (long-term, short-term-opportunity)

3. **Multiple Tags**:
   - A security can have multiple tags simultaneously
   - Example: A security can be both "stable" and "value-opportunity"
   - Example: A security can be "high-quality" and "dividend-opportunity" and "long-term"

## Implementation Notes

- Tags are assigned automatically after security scoring
- Tags should be updated when scores are recalculated
- Tags help the planner identify opportunities and dangers programmatically
- Tags are used internally, not exposed to users for editing

## Tag ID Format

All tag IDs use kebab-case (lowercase with hyphens):
- `value-opportunity` (not `value_opportunity` or `ValueOpportunity`)
- `high-dividend` (not `highDividend` or `HighDividend`)

## Tag Name Format

Tag names are human-readable with proper capitalization:
- "Value Opportunity" (not "value opportunity" or "VALUE OPPORTUNITY")
- "High Dividend Yield" (not "high dividend yield")
