# Geopolitical Event Learning: Implementation Roadmap

## Quick Decision Matrix

**Are you ready to learn geopolitical→market patterns?**

```
YES if:
✓ You have 3+ months of development time
✓ You're willing to curate event catalog (50-100 historical events)
✓ You want interpretable, explainable models
✓ You accept uncertainty (data is sparse, patterns shift)
✓ You'll monitor predictions vs realized impacts quarterly

NO if:
✗ You expect "black-box magic learning" without domain knowledge
✗ You can't spare developer time
✗ You need predictions better than causal domain experts
✗ You want to avoid all data curation/labeling
✗ You need models to work immediately without retraining
```

---

## Phase 1: Foundation (8 weeks)

### Week 1-2: Event Infrastructure
**Goal**: Build the plumbing for learning

```
New files:
  internal/modules/events_learning/
    ├── catalog.go           # Event catalog management
    ├── feature_extractor.go # Event → numerical features
    └── models.go            # Training example types

Changes:
  agents.db: Add event catalog tables
  internal/events/types.go: Add new event types for learning feedback
```

**Deliverable**:
- Event catalog table schema
- Feature extraction pipeline (Intensity, Sentiment, DaysSince, etc.)
- Ability to associate historical events with market reactions

**Effort**: 1-2 weeks
**Team**: 1 developer

---

### Week 2-3: Change Point Detection (CPD)
**Goal**: Detect when market regime shifts coincide with events

```
New files:
  internal/modules/time_series/
    ├── change_point_detector.go
    └── change_point_detector_test.go

Changes:
  history.db: Add detected_change_points table
  internal/scheduler/detect_changes.go: Daily job
```

**Algorithm**: PELT (Pruned Exact Linear Time)

**Example output**:
```
Date       Signal        Change      Magnitude
2024-01-15 VIX           Spike       +42%
2024-01-16 S&P 500       Shift       -1.2%
2024-02-01 Oil Prices    Trend       +0.5/day
```

**Monthly storage**: ~1KB (10-20 detected breaks)

**Deliverable**:
- Change point detector running daily
- Matches detected breaks to event calendar
- Stored in history.db

**Effort**: 2 weeks
**Team**: 1 developer
**Status**: Can start immediately

---

### Week 3-4: Causal Inference (2SLS)
**Goal**: Discover causal coefficients: "War declaration → -X% returns"

```
New files:
  internal/modules/causal/
    ├── inference_service.go       # 2SLS orchestrator
    ├── instrumental_variables.go  # IV regression
    ├── regression.go              # Matrix math
    └── types.go

Changes:
  agents.db: Add event_causal_effects table
  internal/scheduler/causal_inference.go: Monthly job
```

**Algorithm**: Two-Stage Least Squares (2SLS)

**Expected output**:
```
Event Type         Coefficient   95% CI              P-value
War                -0.025        [-0.035, -0.015]   0.001
Sanctions Broad    -0.012        [-0.018, -0.006]   0.002
Trade Tension      -0.003        [-0.008, +0.002]   0.350
Central Bank Hike  +0.008        [+0.002, +0.014]   0.008
```

**Dependencies**: `gonum/matrix` for matrix operations

**Deliverable**:
- Monthly causal effect estimation
- Statistical significance testing
- Stored in agents.db with confidence intervals

**Effort**: 2-3 weeks (matrix implementation is the bulk)
**Team**: 1 developer (requires statistics experience)
**Status**: Highest priority (most rigorous)

---

### Week 4-5: Anomaly Detection + Attribution
**Goal**: Detect unexpected market moves, explain them

```
New files:
  internal/modules/time_series/
    ├── isolation_anomaly.go       # Anomaly detection
    └── anomaly_attribution.go     # Explain via causal methods

Changes:
  agents.db: Add anomaly_explanations table
  internal/scheduler/explain_anomalies.go: Weekly job
```

**Algorithm**: Isolation Forest for detection, 2SLS for attribution

**Example workflow**:
```
1. [Detect] VIX jumps 25% on Day X (anomalous)
2. [Event?] "FOMC meeting happened"
3. [Attribute] Run 2SLS: FOMC → VIX change
   Result: FOMC explains 60% of jump (p=0.02)
   Residual: 40% unexplained
4. [Action] "Most anomaly is explained. Log confidence."
```

**Deliverable**:
- Weekly anomaly detection + attribution report
- Confidence scores on explanations
- Opportunity flagging (unexplained anomalies)

**Effort**: 2 weeks
**Team**: 1 developer
**Status**: Builds on CPD + Causal modules

---

### Week 5-6: Active Learning Integration
**Goal**: User labels uncertain events → model improves

```
New files:
  internal/server/active_learning_routes.go

Changes:
  frontend: New "Help Us Learn" tab
  agents.db: Add user_labels table
  events/types.go: New event types for feedback
```

**Workflow**:
```
1. Model scores all historical events for uncertainty
2. UI shows: "These 10 events confuse us - what actually happened?"
3. User provides: Observed impact + notes
4. System stores labels + triggers monthly retraining
5. Show: "Confidence improved from 0.73→0.78!"
```

**Deliverable**:
- UI for uncertainty sampling
- Backend job to surface uncertain events
- Tracking of user labels → model accuracy correlation

**Effort**: 1-2 weeks
**Team**: 1 developer (frontend + backend)
**Status**: Depends on Phase 1a infrastructure

---

### Week 6-8: Bayesian Linear Regression
**Goal**: Automatic feature selection (which event types matter?)

```
New files:
  internal/modules/bayesian/
    ├── spike_slab_regression.go   # Spike-and-slab Bayesian model
    ├── gibbs_sampler.go           # MCMC implementation
    └── types.go

Changes:
  agents.db: Add feature_significance table
  internal/scheduler/bayesian_feature_selection.go: Monthly job
```

**Algorithm**: Spike-and-slab priors with Gibbs sampling

**Expected output**:
```
Feature              P(≠0)   Mean    95% CI              Interpretation
War declaration      0.94    -2.5%   [-3.1%, -1.9%]     Strong evidence
Sanctions broad      0.87    -1.2%   [-1.8%, -0.6%]     Evidence
Political tension    0.42    -0.05%  [-0.4%, +0.3%]     No evidence
Central bank action  0.96    +0.8%   [+0.4%, +1.1%]     Strong evidence
```

**Monthly computational cost**: ~30 seconds (30 iterations × 200ms each)

**Deliverable**:
- Monthly feature significance estimation
- P(coefficient ≠ 0 | data) for each event type
- Credible intervals on coefficients
- Stored in agents.db

**Effort**: 2-3 weeks (Gibbs sampler implementation)
**Team**: 1 developer (statistics/MCMC experience)
**Status**: Can start after Week 4

---

### Phase 1 Summary

| Week | Component | Lines | Complexity | Risk |
|------|-----------|-------|-----------|------|
| 1-2 | Event Catalog | ~500 | Low | Low |
| 2-3 | Change Point Detection | ~400 | Low-Med | Low |
| 3-4 | Causal Inference | ~600 | High | Med |
| 4-5 | Anomaly Attribution | ~300 | Med | Low |
| 5-6 | Active Learning | ~400 | Med | Low |
| 6-8 | Bayesian Regression | ~700 | High | Med |
| **Total** | | **~3000** | **Medium** | **Medium** |

**Total effort**: 8 developer-weeks (or 2 developers × 4 weeks)
**Timeline**: 2 months with 1 full-time developer
**Hardware**: All methods fit easily in 2GB RAM
**Monthly cost**: ~2-3 minutes of CPU for retraining

---

## Phase 2: Enhancement (4 weeks)

### Option A: Hidden Markov Models
**When**: If Phase 1 CPD/regime detection is working well
**Effort**: 2 weeks
**Benefit**: Principled state transitions, probabilistic regime framework

### Option B: Symbolic Regression for Events
**When**: If Phase 1 shows consistent patterns
**Effort**: 2-3 weeks (reuse existing genetic algorithm)
**Benefit**: Discovered formulas like "Impact = -0.01×Sentiment + 0.5×Sentiment²"

### Option C: Gaussian Processes
**When**: If you want non-linear impact functions
**Effort**: 2 weeks
**Benefit**: Non-parametric, uncertainty quantification via confidence bands

**Recommendation**: Do A + B in parallel

---

## Phase 3: Polish (ongoing)

- Knowledge graphs (optional, for transparency)
- Event embeddings (optional, if text data rich)
- Multi-model ensemble voting
- Dashboard visualizations
- Quarterly accuracy reports

---

## Success Criteria

### By End of Phase 1 (8 weeks)
- ✓ Can explain 50%+ of major market moves to geopolitical events
- ✓ Event catalog has 50+ events with documented impacts
- ✓ User labeling process is smooth and produces 10+/month new labels
- ✓ Models show confidence intervals < 2% on test events
- ✓ Monthly retraining runs in < 5 minutes

### By End of Phase 2 (12 weeks)
- ✓ Can explain 70%+ of market moves
- ✓ Multiple models agree on top 5 most impactful event types
- ✓ Discovered formulas are interpretable and stable
- ✓ Opportunity detection flags 2-3 per month (some validate as real)

### By Month 6
- ✓ Models inform allocation decisions
- ✓ Event sensitivity included in risk dashboard
- ✓ User trust in system increases (validates predictions)

---

## Resource Planning

### Hardware
- **Development**: Your current machine (RAM: OK, CPU: OK)
- **Deployment**: Arduino Uno Q (2GB, ARM64)
  - All Phase 1 methods: Fit easily
  - Phase 2 methods: Fit, but slower retraining
  - Max sustainable: 200-300 training examples + 3-4 monthly models

### Personnel
- **Phase 1**: 1 full-time developer (8 weeks)
  - OR: 2 developers (4 weeks) with pair programming on hard parts
  - Required skills: Go, SQL, statistics/MCMC
- **Phase 2**: 0.5 FTE (part-time) for 4 weeks
- **Ongoing**: 1-2 hours/month for event curation + labeling

### Data
- **Event catalog**: Start with manual curation (you + research)
  - 50 events: 20-30 hours research
  - 100 events: 40-50 hours
  - Automate over time (news APIs, RSS)

### Monitoring
- Monthly: Model accuracy check (predictions vs realized)
- Quarterly: Feature importance review (did relative weights shift?)
- Yearly: Full validation on held-out 1-year data

---

## Critical Unknowns (Design Review Items)

Ask yourself before committing:

1. **Data availability**: Do you have good historical event catalog?
   - If NO: Budget 100 hours for curation
   - If MAYBE: Spend 1 week prototyping to assess feasibility

2. **Acceptable latency**: Can you afford 30-60 minute monthly retraining?
   - If NO: Simplify to CPD + Bayesian only (5 minutes)

3. **Confidence in domain expertise**: Should humans override model predictions?
   - If YES: Need explicit "domain expert veto" UI
   - If NO: Need very high model confidence thresholds

4. **Risk tolerance**: Can you deploy models that are "70% accurate"?
   - If NO: Use models as "weak signal" only, not decisions
   - If YES: Can integrate directly into planning module

5. **Feedback loop speed**: Can you get monthly user labels on uncertain events?
   - If NO: Models won't improve much (small data problem)
   - If YES: Active learning will compound improvements

---

## Go Forth, But Carefully

This research document is honest about:
- What's actually feasible (most things in Phase 1)
- What's risky (overfitting with 50 events)
- What's impossible (deep learning on ARM)
- What matters most (causal inference + change point detection)

**Start with Change Point Detection** (Week 1). It's simple, fast, and immediately informative.
**Then add Causal Inference** (Week 3). Most rigorous path to learning.
**Don't rush to symbolic regression or fancy ML**. The simpler statistical methods will teach you more.

Good luck. This is ambitious, but achievable.
