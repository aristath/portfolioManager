# Expected Return Analysis & Missing Components

## Executive Summary

The system uses a dual-hypothesis approach:
- **Wavelet-based analysis**: Uses multi-scale wavelet decomposition to compute expected return from CAGR, Sharpe ratio, cycle position, momentum, and consistency
- **ML prediction**: Neural network + XGBoost ensemble trained on 20 features with 14-day forward returns

Scores are blended with configurable weights (0-1 per security) to generate trade recommendations.

---

## Critical Missing Components

### 1. Downside Risk (Max Drawdown) ⚠️ CRITICAL

**Problem**: Neither model accounts for maximum drawdown. A security could have 15% expected return but 50% drawdown, yet receive high allocation.

**Impact**: System can overweight securities with massive downside risk.

**Current**: No drawdown consideration in score calculation or position sizing.

**Solution**:
- Add drawdown to wavelet formula as penalty term
- Cap allocations based on drawdown thresholds
- Include drawdown as ML training target (optional)

---

### 2. Model Uncertainty/Confidence ⚠️ CRITICAL

**Problem**: System blends scores but has no measure of confidence. If ML overfits or wavelet becomes stale, system confidently trades in wrong directions.

**Impact**: No signal about reliability of predictions. Low-confidence high-score recommendations get same weight as high-confidence.

**Current**: No uncertainty quantification.

**Solution**:
- Compute prediction variance via ensemble (ML) or bootstrap (wavelet)
- Compute model agreement/disagreement (wavelet vs ML)
- Reduce allocation weight when uncertainty is high
- Confidence threshold for trades (e.g., require confidence > 0.6)

---

### 3. Adaptive Blend Weighting ⚠️ CRITICAL

**Problem**: Blend ratio is fixed per security (0-1). Should adapt based on:
- Model uncertainty
- Recent performance
- Signal consistency

**Impact**: Cannot optimize blend per market regime or per security characteristics.

**Current**: Fixed blend ratio per security in database.

**Solution**:
- Dynamic blend based on model uncertainty
- Weight optimization based on recent validation performance
- Time-varying blend (increase ML during volatility, wavelet during trends)

---

### 4. Robust Disagreement Handling ⚠️ CRITICAL

**Problem**: When models disagree strongly (e.g., wavelet=0.9, ML=0.2), no graceful fallback. System should reduce allocation, not blindly blend.

**Impact**: Extreme disagreements can cause large, unhedged trades.

**Current**: Linear blend with fixed weights.

**Solution**:
- Disagreement threshold (e.g., |wavelet - ML| > 0.4 = warning)
- Down-weight allocation during disagreement
- Independent validation on out-of-sample data
- Separate confidence score for blended output

---

### 5. Drawdown-Aware Position Sizing ⚠️ CRITICAL

**Problem**: Current max position (20%) doesn't account for individual security drawdowns. A volatile security could get 20% with massive downside risk.

**Impact**: Over-exposure to volatile securities during drawdown periods.

**Current**: Single max position constraint for all securities.

**Solution**:
- Dynamic max position based on individual drawdown history
- Consider average drawdown over lookback period
- Reduce max position during market stress
- Implement "drawdown buffers" (don't buy if security > X% below ATH)

---

### 6. Volatility-Normalized Expected Return ⚠️ CRITICAL

**Problem**: Scores are 0-1 without adjusting for security volatility. High-volatility securities get same allocation as low-volatility, despite different risk profiles.

**Impact**: System may overweight volatile assets with similar expected returns.

**Current**: No volatility adjustment in scoring.

**Solution**:
- Normalize scores by security volatility (Sharpe-like adjustment)
- Use risk-adjusted return in allocation
- Cap scores based on volatility tier
- Implement volatility targeting in portfolio construction

---

### 7. Performance Validation ⚠️ HIGH PRIORITY

**Problem**: No systematic comparison of wavelet vs ML performance. Which approach actually works?

**Impact**: Blind trust in methodology without evidence.

**Current**: No validation framework.

**Solution**:
- Track both approaches separately on out-of-sample data
- Compare validation metrics (MAE, RMSE, Sharpe, max drawdown)
- A/B test blend ratios on held-out data
- Automatic model deprecation if performance degrades

---

### 8. Extreme Prediction Protection ⚠️ HIGH PRIORITY

**Problem**: ML predictions can be extreme (>10%, <-10%) with no floor/ceiling. These propagate through blend without guardrails.

**Impact**: Unrealistic allocations based on overfitted predictions.

**Current**: Linear normalization (-10% → 0.0, 0% → 0.5, +10% → 1.0) without validation.

**Solution**:
- Hard caps on normalized scores (e.g., 0.1 to 0.9)
- Prediction quality filters (discard low-confidence extreme predictions)
- Historical validation of extreme predictions (do they actually happen?)
- Rolling sanity checks (scores > 0.9 for 3+ months = warning)

---

## Wavelet-Specific Improvements

### 1. Volatility Penalty (Explicit)

**Current**: Sharpe ratio included in quality, but no further volatility adjustment.

**Proposal**: Add volatility penalty term to expected return formula:

```
expected_return = base_quality × (1 - volatility_penalty)
volatility_penalty = max(0, (volatility - threshold) / scaling_factor)
```

**Rationale**: Explicit control over volatility sensitivity.

---

### 2. Max Drawdown Integration

**Current**: Drawdown calculated but not used.

**Proposal**: Include drawdown in quality calculation:

```
quality = 0.30 × cagr + 0.20 × sharpe - 0.15 × max_drawdown
```

**Rationale**: Penalize poor tail risk behavior.

---

### 3. Adaptive Weights

**Current**: Static weights (30/40/20/10) seem arbitrary.

**Proposal**: Dynamic weights based on market regime:

```
# Bull market: Emphasize momentum
if regime == "bull":
    weights = [0.25, 0.35, 0.25, 0.15]

# Bear market: Emphasize mean reversion
elif regime == "bear":
    weights = [0.35, 0.25, 0.25, 0.15]

# Sideways: Emphasize consistency
else:
    weights = [0.25, 0.25, 0.30, 0.20]
```

**Rationale**: Weights should reflect market dynamics.

---

### 4. Adaptive Lookback Period

**Current**: Fixed 10-year lookback.

**Proposal**: Adapts based on security maturity:

```
if listing_date > "2015-01-01":
    lookback_years = 5
else:
    lookback_years = 10
```

**Rationale**: Newer securities have insufficient data for 10-year analysis.

---

### 5. Sector/Industry Context

**Current**: Single-security only.

**Proposal**: Add sector benchmark comparisons:

```
# Compute security vs sector performance
quality = baseline_quality + sector_alpha

sector_alpha = security_return - sector_return
```

**Rationale**: Some securities deserve premium for sector-relative performance.

---

### 6. Sophisticated Depth (Option 4)

**Current**: Only 3 timeframes (cycle, momentum, consistency).

**Proposal**: Add quarterly cycle and monthly oscillations:

```
• Quarterly cycle (longer-term trend)
• Monthly oscillations (short-term reversals)
• Multi-scale resonance (do multiple timeframes agree?)
```

**Rationale**: Capture more detailed market structure.

**Implementation**:
```python
# Wavelet decomposition at multiple scales
db4_wavelets = pywt.Wavelet('db4')
coeffs = pywt.wavedec(price_data, wavelet='db4', level=4)

# Analyze each scale
cycle_quality = analyze_scale(coeffs[0])    # Coarse scale (quarterly)
momentum_quality = analyze_scale(coeffs[1]) # Level 1 (monthly)
oscillation_quality = analyze_scale(coeffs[2])  # Level 2 (weekly)

# Multi-scale agreement
resonance = similarity_score([cycle_quality, momentum_quality, oscillation_quality])
```

---

## ML-Specific Improvements

### 1. Uncertainty Quantification

**Current**: Single prediction per input.

**Proposal**: Add prediction variance:

```python
# Monte Carlo dropout
predictions = [model.forward(X) for _ in range(100)]
std_dev = np.std(predictions)
confidence = 1 / (1 + std_dev)
```

---

### 2. Regime-Specific Models

**Current**: Single model for all regimes.

**Proposal**: Separate models for different market regimes:

```python
# During volatility
vol_model.predict(X_vol)

# During trends
trend_model.predict(X_trend)

# Blend models based on current regime
final_pred = 0.7 × vol_model(X) + 0.3 × trend_model(X)
```

---

### 3. Feature Importance Monitoring

**Current**: Static feature set.

**Proposal**: Track feature importance changes over time:

```python
# Detect drift
old_importance = model.feature_importance
new_importance = extract_from_trained_model()
drift = jensen_shannon_divergence(old_importance, new_importance)

if drift > threshold:
    alert("Feature importance changed significantly")
```

---

### 4. Out-of-Sample Validation Loop

**Current**: One-time training on historical data.

**Proposal**: Rolling validation:

```python
# Every week
train_data = get_last_200_days()
val_data = get_last_7_days()

val_score = model.evaluate(val_data)
if val_score < threshold:
    trigger_retraining()
```

---

### 5. Ensemble Weight Optimization

**Current**: Fixed 50/50 weights.

**Proposal**: Optimize weights based on recent performance:

```python
# Compute validation scores
nn_val = evaluate_neural_network(val_data)
xgb_val = evaluate_xgboost(val_data)

# Optimize weights
optimal_weights = argmax(weights) subject to weights sum to 1
# e.g., if nn_val=0.8, xgb_val=0.6 → weights = [0.57, 0.43]
```

---

## Implementation Priority

### Must Have (Critical) - Implement First
1. Downside risk (max drawdown)
2. Model uncertainty/confidence
3. Adaptive blend weighting
4. Robust disagreement handling
5. Drawdown-aware position sizing
6. Volatility-normalized expected return

### High Priority - Implement Next
7. Performance validation framework
8. Extreme prediction protection

### Medium Priority - Implement Later
9. Wavelet volatility penalty
10. Max drawdown integration
11. Adaptive weights
12. Adaptive lookback period
13. Sector/industry context
14. Sophisticated depth wavelet analysis

### Low Priority - Nice to Have
15. ML uncertainty quantification
16. Regime-specific models
17. Feature importance monitoring
18. Ensemble weight optimization

---

## Testing Strategy

### Unit Tests
- Test drawdown calculation accuracy
- Test uncertainty quantification bounds
- Test disagreement handling thresholds
- Test extreme prediction protection

### Integration Tests
- Test blend ratio adaptation over time
- Test validation framework end-to-end
- Test drawdown-aware position sizing

### Backtesting
- Backtest with drawdown constraints
- Compare wavelet vs ML performance separately
- Compare adaptive vs fixed blend ratios
- Stress test with extreme market scenarios

### Validation Framework
```python
class ModelValidator:
    def __init__(self):
        self.wavelet_history = []
        self.ml_history = []

    def validate_wavelet(self, actual_returns):
        self.wavelet_history.append(self.compute_score(actual_returns))

    def validate_ml(self, actual_returns):
        self.ml_history.append(self.compute_score(actual_returns))

    def compare_performance(self):
        wavelet_metrics = self.evaluate_metrics(self.wavelet_history)
        ml_metrics = self.evaluate_metrics(self.ml_history)

        return {
            "wavelet": wavelet_metrics,
            "ml": ml_metrics,
            "improvement": (ml_metrics - wavelet_metrics) / wavelet_metrics
        }
```

---

## Data Requirements

### New Data Needed
- Max drawdown history for all securities
- Rolling volatility windows (10d, 30d, 60d)
- Regime classification (volatility, trend, flat)
- Sector benchmark returns
- Out-of-sample validation sets

### Database Schema Updates
```sql
ALTER TABLE securities ADD COLUMN max_drawdown_52w REAL;
ALTER TABLE securities ADD COLUMN volatility_adjusted_score REAL;
ALTER TABLE ml_predictions ADD COLUMN confidence_score REAL;
ALTER TABLE validation_history ADD COLUMN wavelet_score REAL;
ALTER TABLE validation_history ADD COLUMN ml_score REAL;
```

---

## Risk Management Integration

### New Risk Constraints
1. Max drawdown per security (e.g., 30% from ATH)
2. Volatility-adjusted max position
3. Confidence threshold for trades (e.g., >0.6)
4. Disagreement penalty (reduce allocation when |wavelet - ML| > 0.4)
5. Extreme prediction floor/ceiling (0.1 to 0.9)

### Risk Dashboard
- Display max drawdown per security
- Show model confidence scores
- Highlight disagreements between wavelet and ML
- Track extreme predictions
- Monitor validation performance over time

---

## Next Steps

1. **Immediate**: Implement model uncertainty quantification and disagreement handling (Priority 1)
2. **Short-term**: Add max drawdown to scoring and position sizing (Priority 2)
3. **Medium-term**: Build validation framework and adaptive blend weights (Priority 3)
4. **Long-term**: Implement sophisticated wavelet analysis and ML optimization (Priority 4)

---

## References

- Wavelet implementation: `sentinel/analyzer.py:598-701`
- ML predictor: `sentinel/ml_predictor.py`
- ML ensemble: `sentinel/ml_ensemble.py`
- ML features: `sentinel/ml_features.py`
- Trade planner: `sentinel/planner.py:129-622`

---

*Document created: 2026-02-02*
*Last updated: 2026-02-02*
