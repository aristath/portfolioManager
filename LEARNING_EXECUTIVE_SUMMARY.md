# Learning Geopolitical→Market Patterns: Executive Summary

## The Question
Can Sentinel learn relationships between geopolitical events and market outcomes WITHOUT hardcoded rules?

## The Answer
**Yes, but with important caveats.**

---

## Viability by Approach

### DEFINITELY VIABLE (Start here)
These approaches work well on 2GB ARM hardware, with reasonable data requirements:

#### 1. **Causal Inference via 2SLS (Two-Stage Least Squares)**
- **What it does**: Discovers causal coefficients ("War → -2.5% returns with 95% CI")
- **Resource cost**: Trivial (< 1MB RAM, < 1 second CPU)
- **Interpretability**: EXCELLENT (you can explain every coefficient)
- **Time to deploy**: 2-3 weeks
- **Data needed**: 30-50 documented events with observed impacts
- **Verdict**: Most rigorous. Start here. High priority.

#### 2. **Change Point Detection (PELT Algorithm)**
- **What it does**: Detects when market regime shifts abruptly (coincides with events)
- **Resource cost**: Minimal (100MB history, O(n) algorithm)
- **Interpretability**: EXCELLENT (exact dates of breaks)
- **Time to deploy**: 1-2 weeks
- **Data needed**: Just historical prices (you have this)
- **Verdict**: Simplest entry point. Unsupervised. Low risk.

#### 3. **Bayesian Linear Regression with Spike-and-Slab Priors**
- **What it does**: Automatic feature selection (P(Event Type Matters | data) = ?)
- **Resource cost**: Low (Gibbs sampling, ~30 seconds monthly)
- **Interpretability**: EXCELLENT (probability each feature is real)
- **Time to deploy**: 2-3 weeks
- **Data needed**: 30-50 events
- **Verdict**: Complement to causal inference. Best for "what matters?"

#### 4. **Anomaly Detection + Causal Attribution**
- **What it does**: "Market moved unexpectedly. Was it the event? How much?"
- **Resource cost**: Low (IF algorithm + regression)
- **Interpretability**: EXCELLENT (shows attribution percentages)
- **Time to deploy**: 2 weeks
- **Data needed**: Historical prices + event catalog
- **Verdict**: Build on top of CPD. Opportunity detection.

#### 5. **Active Learning via Uncertainty Sampling**
- **What it does**: "These 10 historical events confuse us—help us label them"
- **Resource cost**: Minimal (ranking + user input)
- **Interpretability**: EXCELLENT (user provides ground truth)
- **Time to deploy**: 1-2 weeks
- **Data needed**: 30-50 initial labeled events
- **Verdict**: How to grow training set efficiently. High ROI.

#### 6. **Symbolic Regression Extension** (You have this!)
- **What it does**: Discovers formulas like "Impact = -0.01×Sentiment + 0.5×Sentiment² - 0.0003×DaysSince"
- **Resource cost**: Moderate (genetic algorithm is slow, 30+ minutes monthly)
- **Interpretability**: EXCELLENT (mathematical formulas)
- **Time to deploy**: 2-3 weeks (extension of existing)
- **Data needed**: 50-100 training examples with features
- **Verdict**: Most expressive. Builds on infrastructure you have.

---

### PROBABLY VIABLE (If Phase 1 succeeds)
Harder to implement, but possible on your hardware:

#### 7. **Hidden Markov Models**
- Upgrade regime detection → probabilistic state framework
- 2 week implementation, monthly updates
- Not essential if CPD works well

#### 8. **Gaussian Process Regression**
- Non-parametric, soft uncertainty bounds
- 2 week implementation, scales to ~500 events
- Nice-to-have, simpler Bayesian approach sufficient initially

#### 9. **Knowledge Graphs**
- Explicit reasoning: "War → Oil up → Energy stocks up"
- Manual but transparent, light weight
- Medium-term: helps with interpretation, not learning

---

### NOT VIABLE (Avoid)
These require resources you don't have:

#### ✗ **Deep Learning / Neural Networks**
- Requires TensorFlow/PyTorch (no mature Go support)
- Slow on ARM, needs GPU
- Interpretability: Terrible for risk management
- **Verdict**: Skip entirely

#### ✗ **Meta-Learning / Few-Shot Learning**
- Requires expensive gradient-based optimization
- You don't have massive pre-training geopolitical datasets
- **Verdict**: Theoretically interesting, practically infeasible

#### ✗ **Reinforcement Learning**
- Portfolio management is already a solved optimization problem
- RL adds complexity without benefit
- **Verdict**: Wrong tool for the job

#### ✗ **Graph Neural Networks**
- Memory-heavy on embedded hardware
- Slow inference
- Opaque reasoning
- **Verdict**: Skip

---

## The Honest Assessment

### What Will Succeed
1. **Causal inference** (2SLS) to discover relationships
2. **Change point detection** to find event impact dates
3. **Bayesian methods** to quantify what matters
4. **Active learning** to grow training data efficiently
5. **Symbolic regression** to find interpretable formulas

### What Will Struggle
1. **Small dataset problem**: Only ~50-100 major geopolitical events documented
   - Solutions: Curate event catalog carefully, use active learning to grow it

2. **Non-stationarity**: Geopolitical relationships change over time
   - Solutions: Quarterly retraining, regime-aware models, drift detection

3. **Confounding**: Multiple things happen at once (hard to isolate causes)
   - Solutions: Careful DAG specification, instrumental variables, domain expertise

4. **Black swan events**: Training on history misses novel situations
   - Solutions: Always maintain baseline predictions, don't over-rely on models

### What Will Be Surprising
- **Simplicity**: Statistical methods (causal inference, Bayesian) outperform fancy ML
- **Interpretability**: You can actually explain why the model predicted X
- **Feedback loops**: Uncertainty sampling → user labels → model improves (accelerating)
- **Emergent insights**: Discovering which event types matter (that you didn't expect)

---

## Recommended Implementation Plan

### Timeline: 8 weeks to "Phase 1 complete"

**Weeks 1-2: Foundation**
- Event catalog + feature extraction
- Deliverable: Ability to represent 50 historical events as training data

**Weeks 2-3: Change Point Detection**
- Detect market regime shifts daily
- Deliverable: Daily anomaly report with 10-20% false positive rate

**Weeks 3-4: Causal Inference**
- Discover causal coefficients via 2SLS
- Deliverable: "War → -2.5% (95% CI: -3.1% to -1.9%, p=0.001)"

**Weeks 4-5: Anomaly Attribution**
- Use causal models to explain surprises
- Deliverable: Weekly report: "XYZ% anomaly explained by events"

**Weeks 5-6: Active Learning**
- User labels uncertain events
- Deliverable: "Help Us Learn" UI tab

**Weeks 6-8: Bayesian Feature Selection**
- Which event types matter? Automatic feature selection
- Deliverable: "P(War matters | data) = 0.94, P(Tension matters | data) = 0.42"

### Resource Requirements
- **Developer time**: 1 FTE × 8 weeks (or 2 people × 4 weeks)
- **Skills needed**: Go, SQL, basic statistics/linear algebra
- **Hardware**: Your current machine (deploy to ARM later)
- **Data curation**: 20-30 hours research to build event catalog

### Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|-----------|
| Event catalog incomplete | MEDIUM | Start small (20 events), grow via active learning |
| Causality is hard | MEDIUM | Use multiple methods, require agreement, domain review |
| Patterns are non-stationary | MEDIUM | Quarterly retraining, regime-aware models |
| Over-fitting on small dataset | HIGH | Cross-validation, regularization, active learning to grow data |
| Models perform poorly initially | MEDIUM | Expected! Use as "weak signal" first, integrate gradually |

---

## Expected Outcomes

### By Week 8 (Phase 1)
- Causal effects estimated for 10-20 event types
- Can explain 50-70% of major market moves
- Active learning has collected 20+ additional labels
- Monthly retraining cycle established
- User confidence: Medium (models are decent, not magic)

### By Week 12 (Phase 2, if pursuing)
- Can explain 70-80% of major market moves
- Symbolic regression has discovered interpretable formulas
- Multiple models ensemble voting
- Opportunity detection working (2-3 per month)
- User confidence: High (models validate consistently)

### By Month 6
- Causal relationships incorporated into allocation decisions
- Event risk included in portfolio dashboard
- Market regime forecasting improves planning
- Model confidence intervals narrow as data accumulates
- User trust: Validated through live trading results

---

## Key Decisions You Need to Make

### 1. Event Catalog Strategy
**Question**: How much time will you invest in curating events?

- **Option A (Light)**: 20-30 major events, 20 hours research → Start immediately, limited patterns
- **Option B (Medium)**: 50-100 events, 50 hours research → Better patterns, reasonable effort
- **Option C (Heavy)**: 200+ events, automated collection → Best patterns, ongoing work

**Recommendation**: Start with Option B. Grow via active learning.

### 2. Confidence Threshold
**Question**: How confident do models need to be before influencing decisions?

- **Conservative**: Only use models when p < 0.01 and 95% CI width < 1% → Slower learning, safer decisions
- **Moderate**: Use models when p < 0.05 and 95% CI width < 2% → Balance learning + safety
- **Aggressive**: Use models when p < 0.10 → Faster learning, higher risk

**Recommendation**: Start moderate, adjust quarterly based on validation.

### 3. Feature Engineering Effort
**Question**: Should you manually engineer geopolitical features, or keep them simple?

- **Simple**: Just event type + date → Fast, interpretable, limited patterns
- **Rich**: Type + intensity + affected regions + sentiment + ... → Complex, more patterns, harder to engineer

**Recommendation**: Start simple. Add features only when simple version saturates.

### 4. Deployment Strategy
**Question**: How do you deploy models to the Arduino Uno Q?

- **Binary files**: Pre-compute causal effects, store in agents.db, no recomputation → Fast, no learning
- **Lite engine**: Ship regression engine, recompute monthly → Slow (30+ min), but learns
- **Hybrid**: Pre-compute monthly on powerful machine, deploy to ARM → Best of both

**Recommendation**: Hybrid. Monthly re-run on dev machine, push results to ARM.

---

## Success Metrics

Track these to know if learning is working:

### Weekly
- Anomaly detection false positive rate (target: < 20%)
- User event label submissions (target: > 5 per month)

### Monthly
- Causal effect confidence interval width (target: < 2% on significant effects)
- Model prediction accuracy on held-out test events (target: > 70% directional accuracy)
- CPU time for retraining (target: < 5 minutes)

### Quarterly
- Major market moves explained by models (target: > 60%)
- Feature importance stability (are rankings consistent month-to-month?)
- Discovery of new relationships (# of new significant effects found)
- User feedback & confidence (subjective but important)

---

## Ethical & Risk Considerations

### This Is Real Money
- You're managing your retirement fund
- Models influence allocation decisions
- Mistakes have long-term consequences

**Mitigations**:
1. Always maintain baseline heuristics (don't fully trust models)
2. Require human review of major allocation changes
3. Use confidence intervals, not point predictions
4. Quarterly accuracy audits (compare predictions to outcomes)
5. Document all assumptions (causal DAG, feature engineering choices)

### Known Limitations
1. Models trained on past geopolitical events may not generalize to future black swans
2. Causality is inferred, not proven (assumptions required)
3. With 50 events, confidence intervals are naturally wide
4. Patterns may be due to confounding, not direct effects

**You are not building "AGI for geopolitics."** You're building a statistically rigorous tool to extract signal from noisy historical data. Good judgment from you is essential.

---

## Go/No-Go Decision Points

### Before Week 1: Do you have enough events?
- Poll your history for 30+ documented geopolitical events + their market impacts
- If NO: Abort Phase 1, spend 2-3 weeks curating first
- If YES: Proceed

### After Week 3: Is Change Point Detection working?
- Are detected breaks correlated with event calendar?
- If NO: Debug, may indicate event calendar is too sparse
- If YES: Continue to causal inference

### After Week 4: Is Causal Inference finding anything?
- Are any coefficients statistically significant (p < 0.10)?
- If NO: Likely too few events or bad feature engineering—iterate
- If YES: Proceed to active learning + Bayesian methods

### After Week 6: Is Active Learning producing good labels?
- Are user labels reducing model uncertainty?
- If NO: Users may not understand the task—retool UX
- If YES: System is improving—proceed to Phase 2

---

## Final Verdict

**This is achievable, non-trivial, and valuable work.**

You can build a statistically rigorous, interpretable system to learn geopolitical→market patterns on your embedded hardware. The roadmap is clear, the approaches are proven, and the implementation is feasible in 2-3 months.

**Key success factors**:
1. Invest in event catalog curation upfront (50+ events)
2. Start with simple methods (causal inference, change points)
3. Let active learning grow your dataset over time
4. Build incrementally, validate quarterly
5. Maintain healthy skepticism (models are tools, not truth)

**If you execute this well, you'll have the most interpretable, statistically rigorous geopolitical risk assessment tool in your domain.** That's rare and valuable.

Let's build it.
