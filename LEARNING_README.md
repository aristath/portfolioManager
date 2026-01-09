# Learning Geopolitical→Market Patterns: Research & Implementation Guide

## Overview

This directory contains comprehensive research on learning relationships between geopolitical events and market outcomes WITHOUT hardcoded rules. It covers 10 adaptive learning paradigms, evaluates viability on constrained ARM hardware (2GB), and provides a concrete implementation roadmap.

## Documents

### 1. **LEARNING_EXECUTIVE_SUMMARY.md** ← START HERE
**Best for**: Quick understanding, decision-making
- What's viable, what's not
- Success criteria and expected outcomes
- Timeline: 8 weeks to Phase 1
- Key trade-offs and risk assessment
- **Read time: 15 minutes**

### 2. **RESEARCH_ADAPTIVE_LEARNING.md** ← DETAILED ANALYSIS
**Best for**: Understanding the approaches deeply
- Comprehensive evaluation of 10 learning paradigms
- Resource requirements, complexity, interpretability for each
- Detailed trade-offs and limitations
- Architecture integration patterns
- **Sections**:
  1. Causal Inference & Discovery
  2. Time Series Analysis + Event Detection
  3. Symbolic Regression & Formula Discovery
  4. Few-Shot / Meta-Learning
  5. Active Learning Systems
  6. Probabilistic / Bayesian Approaches
  7. Graph-Based Learning
  8. Embedding-Based Similarity
  9. Reinforcement Learning
  10. Anomaly Detection + Opportunity Detection
- **Read time: 60-90 minutes**

### 3. **LEARNING_IMPLEMENTATION_ROADMAP.md** ← EXECUTION PLAN
**Best for**: Planning implementation, resource allocation
- Phase 1 (8 weeks): 6 core components
  - Week-by-week breakdown
  - Deliverables per component
  - Effort estimates
  - Risk assessment
- Phase 2 (4 weeks): Enhancement options
- Phase 3: Polish & optional advanced methods
- Success criteria by milestone
- Resource planning (personnel, hardware, data)
- **Read time: 30-40 minutes**

### 4. **LEARNING_CODE_STUBS.md** ← IMPLEMENTATION REFERENCE
**Best for**: Developers starting implementation
- Concrete Go code examples for each approach
- File structure and module organization
- Algorithm implementations (PELT, 2SLS, Gibbs sampling, Isolation Forest, etc.)
- Integration patterns with DI container
- Testing stubs
- **Read time: 45-60 minutes** (reference; read as needed)

## Quick Start

### For Decision-Makers
1. Read: **LEARNING_EXECUTIVE_SUMMARY.md** (15 min)
2. Skim: **RESEARCH_ADAPTIVE_LEARNING.md** (10 min, focus on "Viability Overview" section)
3. Review: **LEARNING_IMPLEMENTATION_ROADMAP.md** (15 min)
4. Decide: Can you commit 1 FTE × 8 weeks + 50+ event curation effort?

### For Architects
1. Read: **RESEARCH_ADAPTIVE_LEARNING.md** (full, 90 min)
2. Study: "Architecture Integration" section (20 min)
3. Review: **LEARNING_IMPLEMENTATION_ROADMAP.md** (40 min)
4. Design: agents.db schema and scheduler integration

### For Developers
1. Review: **LEARNING_IMPLEMENTATION_ROADMAP.md** (Phase 1 details, 30 min)
2. Study: **LEARNING_CODE_STUBS.md** (30 min)
3. Start: Week 1 tasks (event catalog + change point detection)
4. Reference: RESEARCH document for algorithm details as needed

## Key Findings

### What Will Work (HIGH CONFIDENCE)
✓ **Causal Inference (2SLS)** - Most rigorous approach to discovering causality
✓ **Change Point Detection (PELT)** - Simplest, deployable immediately
✓ **Bayesian Linear Regression** - Automatic feature selection via spike-and-slab priors
✓ **Symbolic Regression** - Extend existing genetic algorithm for event→impact formulas
✓ **Active Learning** - User labeling grows training dataset efficiently
✓ **Anomaly Detection + Attribution** - Explain market surprises via causal inference

### What's Questionable (MEDIUM CONFIDENCE)
? **Hidden Markov Models** - Nice-to-have upgrade to regime detection
? **Gaussian Processes** - Works but scales to ~500 events (limits if dataset grows)
? **Knowledge Graphs** - Transparent but requires manual ontology

### What Won't Work (LOW CONFIDENCE)
✗ **Deep Learning** - No GPU, limited TensorFlow/PyTorch Go support
✗ **Meta-Learning** - Too expensive, you don't have pre-training datasets
✗ **Graph Neural Networks** - Memory/CPU prohibitive on ARM
✗ **Reinforcement Learning** - Over-engineered for this problem

## Timeline

### Phase 1: Foundation (8 weeks)
- **Week 1-2**: Event catalog infrastructure
- **Week 2-3**: Change point detection (PELT algorithm)
- **Week 3-4**: Causal inference (2SLS implementation)
- **Week 4-5**: Anomaly detection + attribution
- **Week 5-6**: Active learning integration
- **Week 6-8**: Bayesian linear regression (spike-and-slab)

**Outcome**: 50-70% of major market moves explained, active feedback loop

### Phase 2: Enhancement (4 weeks)
- **Option A**: Hidden Markov Models (regime detection upgrade)
- **Option B**: Symbolic regression extension (event→impact formulas)
- **Option C**: Gaussian processes (non-linear impact functions)

**Outcome**: 70-80% of market moves explained, multiple models ensemble

### Phase 3: Polish (ongoing)
- Knowledge graphs (transparency)
- Event embeddings (similarity-based prediction)
- Dashboard visualizations
- Quarterly accuracy audits

## Success Metrics

### By Week 8 (Phase 1)
- Can explain 50-70% of major market moves to geopolitical events
- Event catalog: 50+ events with documented impacts
- Active learning: 20+ user labels collected
- Monthly retraining: < 5 minutes CPU

### By Week 12 (Phase 2)
- Can explain 70-80% of market moves
- 5-7 independent learned models working
- Opportunity detection: 2-3 flagged per month
- User confidence: High (predictions validate regularly)

### By Month 6+
- Models inform allocation decisions
- Event sensitivity included in risk dashboard
- Causal relationships stable and documented

## Architecture Overview

```
Event Data Curation
    ↓
Feature Extraction (Intensity, Sentiment, Days, Regions, etc.)
    ↓
Training Examples (Event + Market Reaction)
    ↓
├─→ Causal Inference (2SLS)
│   └─→ Causal coefficients, statistical tests
│
├─→ Change Point Detection (PELT)
│   └─→ Event impact dates, anomalies
│
├─→ Bayesian Regression (Spike-and-Slab)
│   └─→ Feature significance, P(≠0)
│
├─→ Symbolic Regression (Genetic Algorithm)
│   └─→ Discovered formulas
│
├─→ Anomaly Attribution
│   └─→ Explain market surprises
│
└─→ Active Learning
    └─→ Identify uncertain events for labeling
        ↓
    User Labels → Trigger Retraining
        ↓
    Model Improves
        ↓
    Confidence Increases

All outputs → agents.db (discovered causal laws, model predictions)
            → Cache.db (recommendations, opportunities)
```

## Database Schema (agents.db additions)

```sql
-- Event catalog
CREATE TABLE geopolitical_events (
    id INTEGER PRIMARY KEY,
    type TEXT,  -- "war", "sanctions", "trade_tension", etc.
    date DATE,
    description TEXT,
    intensity FLOAT,
    affected_regions TEXT,  -- JSON array
    observed_impact FLOAT,  -- Optional: user-labeled actual return
    user_labeled BOOLEAN,
    created_at TIMESTAMP
);

-- Causal effects discovered
CREATE TABLE event_causal_effects (
    event_type TEXT,
    coefficient FLOAT,
    std_error FLOAT,
    t_statistic FLOAT,
    p_value FLOAT,
    ci_lower FLOAT,
    ci_upper FLOAT,
    model_version INTEGER,
    updated_at TIMESTAMP
);

-- Feature importance
CREATE TABLE feature_significance (
    feature_name TEXT,
    p_not_zero FLOAT,  -- P(coefficient ≠ 0 | data)
    posterior_mean FLOAT,
    credible_interval_lower FLOAT,
    credible_interval_upper FLOAT,
    model_version INTEGER,
    updated_at TIMESTAMP
);

-- Discovered formulas
CREATE TABLE discovered_event_formulas (
    id INTEGER PRIMARY KEY,
    formula TEXT,
    event_type TEXT,
    regime TEXT,
    fitness FLOAT,
    complexity INTEGER,
    accuracy FLOAT,
    model_version INTEGER,
    updated_at TIMESTAMP
);

-- Model predictions for auditing
CREATE TABLE event_impact_predictions (
    event_id INTEGER,
    model_type TEXT,  -- "causal", "symbolic", "gp", "hmm"
    predicted_impact FLOAT,
    uncertainty FLOAT,
    actual_impact FLOAT,  -- Filled later
    accurate BOOLEAN,
    prediction_date TIMESTAMP
);
```

## Implementation Phases & Effort

| Phase | Duration | Components | FTE | Risk | ROI |
|-------|----------|-----------|-----|------|-----|
| **Phase 1** | 8 weeks | Event catalog, CPD, Causal, Anomaly, AL, Bayesian | 1.0 | Medium | High |
| **Phase 2** | 4 weeks | HMM, Symbolic Reg, GP | 0.5 | Medium | Medium |
| **Phase 3** | Ongoing | Knowledge graphs, embeddings, polish | 0.2 | Low | Low |

## Critical Unknowns (Design Review)

Before committing, clarify:

1. **Event catalog**: Do you have/can you curate 50+ historical events?
2. **Feedback loop**: Can you label 5-10 uncertain events per month?
3. **Latency**: Can you afford 30-60 minute monthly retraining?
4. **Risk tolerance**: How confident must models be before influencing decisions?
5. **Domain expertise**: Should humans override model predictions? How often?

## Risks & Mitigations

| Risk | Severity | Mitigation |
|------|----------|-----------|
| Small dataset (50 events) | MEDIUM | Active learning grows data, accept wide confidence intervals initially |
| Causality is hard | MEDIUM | Use multiple methods, require agreement, domain expert review |
| Non-stationary relationships | MEDIUM | Quarterly retraining, regime-aware models, drift detection |
| Over-fitting | HIGH | Cross-validation, regularization, hold-out test set |
| Black swan events | MEDIUM | Always use baseline heuristics, don't trust extrapolation |

## Recommended Reading Order

1. **First**: LEARNING_EXECUTIVE_SUMMARY.md (decide if this is worth pursuing)
2. **Second**: RESEARCH_ADAPTIVE_LEARNING.md "Viability by Approach" section
3. **Third**: LEARNING_IMPLEMENTATION_ROADMAP.md (commit to timeline)
4. **Fourth**: RESEARCH_ADAPTIVE_LEARNING.md "Synthesis & Recommendation Matrix"
5. **Finally**: LEARNING_CODE_STUBS.md (when ready to implement)

## Key Papers & References

For those wanting deeper understanding:

- **Causal Inference**: Angrist & Pischke "Mostly Harmless Econometrics" (2SLS, IV)
- **Change Point Detection**: Killick et al. "Optimal detection of changepoints with a linear computational cost" (PELT)
- **Bayesian Feature Selection**: O'Hara & Sillanpää "A review of Bayesian variable selection methods" (spike-and-slab)
- **Symbolic Regression**: La Cava et al. "Contemporary Symbolic Regression Methods"
- **Active Learning**: Freeman "Active Learning for Deep Object Detection" (uncertainty sampling)
- **Anomaly Detection**: Liu et al. "Isolation Forest" (isolation-based anomaly)

## Questions?

This research covers:
- What's possible on ARM embedded hardware ✓
- What's interpretable & explainable ✓
- What's realistically implementable in 2-3 months ✓
- What's actually useful for portfolio management ✓

It does NOT cover:
- Pre-trained large language models (overkill, slow on ARM)
- Distributed ML systems (single device, no cluster)
- Real-time streaming models (use batch processing, monthly retraining)
- Predicting specific prices (focus on impacts & causality, not point predictions)

---

## Document Maintenance

These documents are living. Update when:
- Phase 1 implementation reveals new insights
- Hardware constraints change
- New approaches become viable
- Lessons learned from Phase 1

---

**Last Updated**: 2026-01-09
**Status**: Complete research phase, ready for Phase 1 implementation
**Next Step**: Architect Phase 1, design agents.db schema, begin Week 1 coding
