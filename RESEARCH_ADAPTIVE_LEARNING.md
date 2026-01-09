# Adaptive Learning for Geopolitical Events → Market Outcomes
## Comprehensive Research & Viability Analysis for ARM Embedded Systems

**Target Environment**: Arduino Uno Q (2GB RAM, ARM64, Linux)
**Execution Context**: Autonomous portfolio management system
**Goal**: Learn relationships between geopolitical events and market impacts WITHOUT hardcoded rules

---

## Executive Summary

For your embedded, resource-constrained system, **NOT all approaches are viable**. This analysis categorizes ten learning paradigms by actual feasibility on 2GB ARM hardware, focusing on interpretable, deployable solutions that work within Sentinel's clean architecture.

**Verdict**: 3-4 approaches are genuinely practical. Several others are theoretically interesting but require careful engineering or are fundamentally mismatched to your constraints.

---

## 1. CAUSAL INFERENCE & DISCOVERY METHODS

### Overview
Causal inference determines whether event X *causes* market outcome Y, not just correlation. Essential for distinguishing genuine geopolitical impacts from spurious patterns.

### Key Approaches

#### 1.1 Instrumental Variables (IV) / Two-Stage Least Squares (2SLS)
**What it does**: Uses a variable that affects the outcome only through the treatment to isolate causal effects.
```
Example: Political sentiment → VIX → Stock Returns
Find instrument that affects sentiment but not returns directly
```

**Viable for your system?** ✓ YES (constrained scope)
- **Resource requirements**: Minimal. Pure statistical computation.
  - Memory: ~50-100MB for 5 years price + event data
  - CPU: Single-threaded matrix operations
  - Suitable for monthly/weekly batch processing

**Implementation complexity**: Medium
- Requires: Data validation, statistical test implementation
- Don't need: External ML libraries
- Go implementation: Hand-rolled regression, eigenvalue decomposition

**Interpretability**: EXCELLENT
- Direct causal coefficients: "10% sentiment increase → 2.3% VIX rise → -0.8% return"
- Statistical significance tests included
- Matches your symbolic regression architecture (formula-based)

**For "event → market impact"**: Excellent match
- Events are natural experiments (discrete treatment)
- Can use geopolitical event dates as instruments
- Captures both immediate and lagged effects

**Learning data requirements**: Moderate
- Need: 3-5 years price history + event catalog + market indicators
- Can start with 20-50 documented geopolitical events
- Quality > quantity: Well-documented events matter more

**Feedback loop mechanism**: Strong
- Run IV estimation weekly
- Compare predicted vs. actual market moves
- Update event severity classification based on realized impacts
- Surface surprising causal relationships

**Trade-offs**:
- ✓ Statistically rigorous, interpretable
- ✓ Incremental learning (re-estimate as new data arrives)
- ✗ Requires strong instruments (hard to find for geopolitics)
- ✗ Assumes linear relationships (markets aren't always linear)
- ✗ Struggles with simultaneous causality (both affect each other)

**Recommended implementation**:
```go
// Pseudo-code structure for your DI container
type CausalInferenceService struct {
    instrumentRepository    RepositoryInterface  // Event catalogs
    priceHistoryRepository  RepositoryInterface  // Market data
    statisticsEngine        StatisticsEngine     // 2SLS implementation
    log                     zerolog.Logger
}

// Run monthly
func (s *CausalInferenceService) EstimateEventCausalImpact(
    eventType string,
    lookbackMonths int,
) (*CausalEffect, error) {
    // 1. Fetch events of type
    // 2. Construct lagged price changes
    // 3. Run 2SLS
    // 4. Return causal coefficients + significance tests
    // 5. Store in agents.db (treatment effect size)
}
```

**Go libraries needed**: None. Implement matrix operations directly (or use `gonum/matrix`).

---

#### 1.2 Directed Acyclic Graphs (DAGs) + Backdoor Criterion
**What it does**: Visualizes causal structures, identifies which variables must be controlled to isolate effects.

**Viable for your system?** ✓ YES (emerging)
- **Resource requirements**: Minimal
  - Memory: <50MB (graph structure + statistics)
  - CPU: Topological sorting, path enumeration (fast)

**Implementation complexity**: Low-Medium
- Represent as adjacency matrix or edge list
- Run backdoor/frontdoor criterion checks
- Standard graph algorithms

**Interpretability**: EXCELLENT
- Visual representation of assumptions
- Shows why you're controlling variables
- Falsifiable: "Does this DAG match reality?"

**For "event → market impact"**: Very good
```
Proposed DAG:
  Event Severity ──→ Media Sentiment ──→ VIX ──→ Returns
                 │                              │
                 └──→ Central Bank ──────────→ Rates

Control for: Media Sentiment (blocks backdoor)
Leave open: VIX (captures mechanism)
```

**Learning data requirements**: Moderate
- 3-5 variables per analysis path
- Causal relationships discovered via IV (then encoded in DAG)
- 20-50 examples per causal path

**Feedback loop mechanism**: Medium
- Store DAGs in agents.db
- Compare predictions under different DAGs
- Refine structure quarterly

**Trade-offs**:
- ✓ Transparent assumptions
- ✓ Guides what data to collect
- ✓ Integrates with IV/causal methods
- ✗ Requires domain expertise to specify
- ✗ Sensitive to misspecification
- ✗ No automatic structure learning (on resource budget)

**Recommended implementation**:
```go
// In internal/modules/causal (new module)
type CausalDAG struct {
    nodes map[string]Variable
    edges []CausalEdge  // source → target
    assumptions []string
}

func (dag *CausalDAG) HasBackdoorPath(
    treatment, outcome string,
    controlSet []string,
) bool {
    // Check if control set blocks all backdoor paths
}

func (dag *CausalDAG) MinimalAdjustmentSet(
    treatment, outcome string,
) ([]string, error) {
    // Return variables to control for causal identification
}
```

---

#### 1.3 Causal Forests (Athey & Wager)
**What it does**: Machine learning approach that estimates heterogeneous treatment effects (does event affect different securities differently?).

**Viable for your system?** ✗ MARGINAL
- **Resource requirements**: Moderate-High
  - Memory: 200-500MB for 1000+ securities
  - CPU: Recursive tree building (parallel-friendly but intensive)
  - Training time: 5-30 minutes for full forest

**Implementation complexity**: High
- Requires: Random forest infrastructure, sample splitting, cross-fitting
- Tricky details: MSE shrinkage, honest trees, orthogonalization
- Go implementation: Custom or GoML libraries

**Interpretability**: Good (but opaque compared to IV)
- Outputs: Heterogeneous treatment effects per security
- Can extract: "Tech stocks respond 1.2x to geopolitical events, utilities 0.3x"
- Not transparent on mechanisms

**For "event → market impact"**: Excellent
- Handles: Different securities react differently
- Captures: Non-linear threshold effects
- Learns: Which features matter for response magnitude

**Learning data requirements**: High
- Need: 500+ observations (events × securities)
- Training time intensive: Monthly re-training challenging

**Feedback loop mechanism**: Medium-Hard
- Re-train monthly (expensive)
- Monitor OOB error and prediction accuracy
- Drift detection challenging

**Trade-offs**:
- ✓ Captures heterogeneity (realistic)
- ✓ Non-parametric (flexible)
- ✗ Black-box relative to IV
- ✗ Computationally expensive
- ✗ Requires large sample sizes
- ✗ Monthly re-training is resource-heavy

**Verdict**: Skip unless event→impact heterogeneity is critical. Use 2SLS first.

---

### Recommendation for Causal Methods
**Start with**: Instrumental Variables (2SLS) + DAG encoding
- Lightweight, interpretable, integrates well with symbolic regression
- Monthly batch process: ~30 seconds on 2GB hardware
- Store findings in `agents.db` as discovered causal laws
- Use DAG output to guide feature engineering for symbolic regression

**Architecture**:
```
Event Catalog + Price History
    ↓
CausalInferenceService (2SLS)
    ↓
Discovered Causal Effects (e.g., "War → -1.8% returns in 2 weeks")
    ↓
CausalDAG Encoder
    ↓
agents.db (treatment_effects table)
    ↓
Feedback: Monitor accuracy of predictions weekly
```

---

## 2. TIME SERIES ANALYSIS WITH EVENT DETECTION

### Overview
Detect when something unusual happened in time series (event detection), then correlate with geopolitical calendar.

### Key Approaches

#### 2.1 Change Point Detection (CPD)
**What it does**: Identifies abrupt shifts in time series statistics (mean, volatility, trend).

**Viable for your system?** ✓ YES
- **Resource requirements**: Minimal
  - Memory: 100MB for 5 years of daily data
  - CPU: Single pass (O(n)) or O(n log n) with PELT
  - Runs continuously

**Implementation complexity**: Low-Medium
- Algorithm: PELT (Pruned Exact Linear Time) is straightforward
- Go implementation: ~500 lines of code
- No external dependencies

**Interpretability**: EXCELLENT
- Shows exact dates of regime changes
- Direct: "Price volatility jumped 45% on this date"
- Maps directly to event calendar

**For "event → market impact"**: Excellent foundation
```
Example workflow:
1. Detect: VIX volatility jump on day X
2. Check: Is there a geopolitical event on day X?
3. If yes: Log correlation (event → volatility)
4. If no: Investigation (data quality? sparse event catalog?)
```

**Learning data requirements**: Low
- Just historical prices (you have this)
- No event labels needed initially
- Unsupervised: Discover structural breaks automatically

**Feedback loop mechanism**: Strong
- Run daily on portfolio returns, VIX, sector indices
- Match detected change points to event calendar
- Build event severity estimates from magnitude of break
- Monthly: Update event catalog with confidence scores

**Trade-offs**:
- ✓ Computationally cheap
- ✓ Unsupervised (no labels needed)
- ✓ Real-time capable
- ✓ Transparent output
- ✗ Detects breaks, not causes
- ✗ Blind to events without market impact
- ✗ Sensitive to parameter tuning (penalty term)

**Recommended implementation**:
```go
// In internal/modules/time_series (new module)
type ChangePointDetector struct {
    minSegmentLength int
    penalty          float64  // Tuned via cross-validation
    log              zerolog.Logger
}

// Returns: list of change point dates
func (d *ChangePointDetector) DetectChangePoints(
    timeSeries []float64,
    timestamps []time.Time,
) ([]ChangePoint, error) {
    // PELT algorithm
    // Return: {Date, OriginalValue, ChangeType: "spike/shift/trend", Magnitude}
}

// Integrate with events.EventManager
func DetectEventImpactDates(
    priceHistory []DailyReturn,
    eventCalendar []GeopoliticalEvent,
) ([]EventMarketImpact, error) {
    // 1. Run CPD on returns
    // 2. For each detected break:
    //    - Search event calendar in ±3 day window
    //    - Calculate impact magnitude
    // 3. Return annotated list
}
```

**Storage**:
- `history.db`: Detected change points with confidence
- `agents.db`: Event → impact magnitude mappings

---

#### 2.2 Anomaly Detection (Isolation Forests, LOF)
**What it does**: Identifies unusual data points (outliers = potential event impacts).

**Viable for your system?** ✓ YES (lightweight variant)
- **Resource requirements**: Low-Medium
  - Memory: 50-150MB for 5 years daily data
  - CPU: O(n log n) for Isolation Forest
  - Real-time scoring: <1ms per observation

**Implementation complexity**: Low
- Algorithm: Isolation Forest is simple and elegant
- Go implementation: ~400 lines
- No heavy math library needed

**Interpretability**: Good
- Shows: "This day's returns are 3.2 sigma outliers"
- Connects to: Event calendar
- Limitation: Doesn't explain *why* (but CPD + this tells the story)

**For "event → market impact"**: Good as auxiliary method
- Complements CPD: CPD finds shifts, anomaly finds spikes
- Use together: "Event caused both a spike (anomaly) and shift (CPD)"

**Learning data requirements**: None
- Unsupervised
- Just historical returns

**Feedback loop mechanism**: Strong
- Daily anomaly scores → event impact magnitude
- Build distributions: P(anomaly size | event type)
- Quarterly: Refine decision threshold

**Trade-offs**:
- ✓ Fast, lightweight
- ✓ Complements CPD well
- ✓ No labels needed
- ✗ Doesn't distinguish causes
- ✗ Sensitive to "normal" regime definition
- ✗ Outliers ≠ events (some outliers are data errors)

**Recommended variant**: **Isolation Anomaly Detection**
```go
type IsolationForest struct {
    trees []*IsolationTree
    sampleSize int
}

// Much simpler than Random Forests
// Key insight: Anomalies are easy to isolate (fewer splits needed)
// Memory-efficient: Only store split points, not full data
```

---

#### 2.3 Hidden Markov Models (HMM)
**What it does**: Assumes market operates in hidden states (calm, volatile, crisis) that emit observable prices. Detects state transitions.

**Viable for your system?** ✓ YES
- **Resource requirements**: Minimal
  - Memory: <50MB (state transition matrix + emissions)
  - CPU: Forward-backward algorithm O(n×k²) where k=3-5 states

**Implementation complexity**: Medium
- Algorithm: Forward-backward (Baum-Welch) is well-known
- Go implementation: ~600 lines
- Tricky: Numerical stability (log-space computation)

**Interpretability**: Good
- States map to market regimes: "Quiet, Stressed, Chaotic"
- Transition probabilities: P(Quiet→Stressed|event) = ?
- Viterbi: Most likely state sequence given observations

**For "event → market impact"**: Good
```
Example:
Event: War declaration
Day 0: P(Chaotic) jumps from 0.1 to 0.7
Days 1-7: Expected duration in Chaotic state
Day 8: P(Stressed) = 0.6, P(Quiet) = 0.4
```

**Learning data requirements**: Low
- Just prices (unsupervised state discovery)
- Can initialize with regime-based splits (calm/volatile/crisis)

**Feedback loop mechanism**: Strong
- Daily state inference
- Correlate state transitions with event calendar
- Learn: P(state transition | event type)
- Update: Transition matrix monthly

**Trade-offs**:
- ✓ Principled probabilistic model
- ✓ Fast inference (Viterbi, forward)
- ✓ Natural regime framework
- ✓ Integrates with Bayesian approaches
- ✗ Markov assumption (no long-term memory)
- ✗ State interpretation is manual
- ✗ Number of states must be chosen

**Recommended implementation**:
```go
type MarketHMM struct {
    states    []string  // e.g., ["calm", "volatile", "crisis"]
    transitionProb  [3][3]float64  // P(s_t | s_t-1)
    emissionProb    interface{}     // P(price_change | state)
}

func (m *MarketHMM) InferState(
    returns []float64,
) ([]StateInference, error) {
    // Forward-backward algorithm
    // Return: Most likely state sequence + confidence
}

// Integrate with events
func CorrelateEventStateTransitions(
    events []GeopoliticalEvent,
    stateSequence []StateInference,
) map[string]StateTransitionFrequency {
    // Build: P(Calm→Chaotic | "War") = 0.8, etc.
}
```

**Integration with existing code**: Fits naturally with `market_regime/detector.go` (upgrade current system)

---

### Recommendation for Time Series
**Start with**: Change Point Detection + Isolation Anomaly
- Unsupervised, fast, transparent
- **Weekly process**: Scan historical returns for breaks
- **Daily process**: Score new market data for anomalies
- Correlate detected breaks with event calendar

**Next phase**: Hidden Markov Model
- Replaces existing regime detection
- More principled state discovery
- Integrates with causal methods (stratify IV estimates by state)

**Architecture**:
```
Daily Returns/VIX/Sectors
    ↓
Change Point Detector + Isolation Forest
    ↓
Detected breaks + anomalies (high-confidence)
    ↓
Correlate with Event Calendar
    ↓
Compute: P(impact | event type, market state)
    ↓
agents.db (event_impacts table)
    ↓
Feedback: Compare predictions vs realized
```

---

## 3. SYMBOLIC REGRESSION & FORMULA DISCOVERY

### Overview
You already have this partially implemented. This section details extensions for geopolitical learning.

**Current Status**:
- Genetic programming engine exists (`internal/modules/symbolic_regression/`)
- Discovers formulas for expected returns (price → features → formula)
- Regime-aware: Can discover per-regime formulas

**Extension opportunity**: Discover formulas mapping geopolitical features → market impact

#### 3.1 Adding Geopolitical Feature Engineering
**What it does**: Extract numerical geopolitical features, feed to symbolic regression.

**Viable for your system?** ✓ YES (natural extension)
- **Resource requirements**: Modest
  - Memory: 100-200MB (training examples + population)
  - CPU: Genetic algorithm continues to be your bottleneck
  - Training time: 30-60 minutes per regime (acceptable monthly)

**Implementation complexity**: Medium
- Extend `DataPrep` in symbolic_regression to accept event features
- Define geopolitical feature extraction functions
- Reuse existing genetic algorithm + validation

**Interpretability**: EXCELLENT
- Formula output: e.g., "Returns = -0.01 × SentimentScore + 0.5 × SentimentScore² - 0.0003 × TensionDays"
- Direct causal interpretation
- Matches symbolic regression philosophy

**For "event → market impact"**: Perfect match
- Events → numerical features (sentiment, days since, intensity, category)
- Genetic algorithm discovers relations
- Can be regime-specific

**Learning data requirements**: Moderate-High
- Need: Historical event catalog + market reactions
- Start with: 50-100 events with documented impacts
- Build over time: Collect retrospective calibrations

**Feedback loop mechanism**: Strong
- Monthly: Re-train with new events
- Weekly: Score predictions vs realized
- Track: Which formula types perform best

**Trade-offs**:
- ✓ Interpretable formulas
- ✓ Integrates naturally with existing code
- ✓ Regime-aware
- ✓ Computationally feasible
- ✗ Feature engineering is manual (hard to automate)
- ✗ Genetic algorithm is slow for high-dimensional feature space
- ✗ Requires event catalog curation

**Recommended implementation**:
```go
// Extend internal/modules/symbolic_regression

// New feature types for events
type GeopoliticalFeature struct {
    EventType    string    // "War", "SanctionsBroad", "Tension", etc.
    Intensity    float64   // 0-1 scale (extracted from categorical)
    DaysSinceStart int     // Event aging
    Category     string    // "Conflict", "Economic", "Policy"
    Sentiment    float64   // -1 to 1 (optional: from news sentiment API)
    AffectedRegions int    // Number of affected regions
    SecurityCorrelation float64 // Does security correlate with affected region?
}

type EventTrainingExample struct {
    Date time.Time
    Features []GeopoliticalFeature  // Multiple concurrent events possible
    AggregatedFeatures map[string]float64  // Reduced to scalar features
    Target float64  // Market impact (% return 1-30 days post-event)
}

// Extend DataPrep
func (dp *DataPrep) ExtractEventTrainingExamples(
    startDate, endDate time.Time,
    eventCatalog []GeopoliticalEvent,
    lookforwardDays int,
) ([]EventTrainingExample, error) {
    // 1. For each event, compute feature vector
    // 2. Look forward lookforwardDays
    // 3. Compute actual market reaction
    // 4. Return training examples
}

// Feature aggregation (reduce multiple events to scalars for formula input)
func AggregateEventFeatures(
    features []GeopoliticalFeature,
) map[string]float64 {
    // Output: {
    //   "max_intensity": 0.8,
    //   "avg_sentiment": -0.3,
    //   "days_since_earliest": 5,
    //   "num_events": 2,
    //   "num_affected_regions": 3,
    // }
}

// Reuse existing genetic algorithm with new features
func (ds *DiscoveryService) DiscoverEventImpactFormula(
    examples []EventTrainingExample,
    regimeRange *RegimeRange,
) (*DiscoveredFormula, error) {
    // 1. Extract features and targets
    // 2. Normalize features
    // 3. Run genetic algorithm (existing)
    // 4. Return: formula string + fitness
}
```

**Storage**:
- `agents.db`: Add table for discovered event→impact formulas
- `history.db`: Store event feature vectors (for retraining)

**Integration with causal methods**:
```
Symbolic Regression Path:
Discovered formula: "Impact = f(Intensity, Sentiment, DaysSince)"
    ↓
If formula coefficients are stable across time:
    → Suggests causal relationship (not just correlation)
    → Feed to causal inference as feature engineering guide

Causal Inference Path:
Identified: "Intensity → Market Return" (via 2SLS)
    ↓
Use as constraint for symbolic regression:
    → Genetic algorithm should prefer formulas where Intensity has positive/negative coefficient
    → Faster convergence
```

---

#### 3.2 Multi-Objective Optimization
**What it does**: Discover formulas that balance accuracy AND interpretability (Pareto frontier).

**Viable for your system?** ✓ YES (modification of existing)
- **Resource requirements**: Same as current
  - Just track multiple objectives simultaneously

**Implementation complexity**: Low-Medium
- Existing: Single fitness metric (RMSE)
- New: Multiple metrics (RMSE + complexity + parameter count)
- Use: NSGA-II (non-dominated sorting genetic algorithm)

**For "event → market impact"**: Good enhancement
- Pareto frontier shows: Simple interpretable formulas vs. complex accurate ones
- Decision: "Is 0.05 RMSE improvement worth doubling formula complexity?"

**Recommended change**:
```go
// Extend existing EvolutionConfig
type MultiObjectiveConfig struct {
    Objectives []ObjectiveType  // ACCURACY, COMPLEXITY, INTERPRETABILITY
    Weights    []float64        // Weight per objective
}

type ParetoFront struct {
    solutions []*FormulaWithFitness  // None dominates others
}

// Replace simple fitness ranking with Pareto ranking
func RankByPareto(population []*FormulaWithFitness) *ParetoFront {
    // Non-dominated sorting
}
```

---

### Recommendation for Symbolic Regression
**Phase 1** (immediate):
- Extend feature engineering to accept geopolitical features
- Reuse existing genetic algorithm
- Monthly discovery: "Event Type → Impact Formula"
- Store discovered formulas in agents.db

**Phase 2** (optional):
- Add multi-objective optimization (accuracy vs. complexity)
- Display Pareto frontier to user
- User selects preferred trade-off

**Architecture**:
```
Manual Event Catalog (curated)
    ↓
Feature Extraction (Intensity, Sentiment, Days, etc.)
    ↓
Training Examples (Event + Market Reaction)
    ↓
Existing Genetic Algorithm + DataPrep
    ↓
Discovered Formulas: "Impact = f(Event Features)"
    ↓
agents.db (discovered_event_impact_formulas)
    ↓
Prediction: Apply formula to new events
    ↓
Feedback: Compare predictions vs realized impact
```

---

## 4. FEW-SHOT / META-LEARNING APPROACHES

### Overview
Learn from few examples, adapt quickly to new situations.

**Important caveat**: Meta-learning is the hardest paradigm for embedded systems. Requires:
- Large pre-training dataset (geopolitical events × securities × markets)
- Expensive gradient-based optimization
- Stateful models that carry parameters across tasks

### 4.1 Few-Shot Learning via Prototypical Networks
**What it does**: Learn to classify/predict by comparing to prototype examples in learned embedding space.

**Viable for your system?** ✗ NO (too heavy)
- **Resource requirements**: High
  - Memory: 200-500MB (embedding model + prototypes)
  - CPU: Neural network inference (expensive)
  - Training: Days of data + expensive gradient updates

**Implementation complexity**: Very High
- Requires: Deep learning framework (TensorFlow Lite, ONNX Runtime)
- Go bindings: Immature, unreliable
- Debugging: Black-box training process

**Interpretability**: Poor
- Embeddings are opaque
- Can visualize via t-SNE but doesn't explain decisions

**Verdict**: Skip. Too heavy, not interpretable enough for risk management.

---

### 4.2 Bayesian Meta-Learning (MAML / Hierarchical Bayes)
**What it does**: Learn a prior over model parameters; quickly adapt to new task (new event type) with few examples.

**Viable for your system?** ✗ MARGINAL
- **Resource requirements**: Moderate-High
  - Memory: 100-200MB (prior + posterior samples)
  - CPU: MCMC sampling is slow (hours per update)

**Implementation complexity**: Very High
- Requires: Probabilistic programming (Stan, PyMC)
- Go: Limited libraries (`go-echarts` for visualization only)
- Debugging: Convergence diagnostics are complex

**Interpretability**: Excellent (theoretical)
- Outputs: Probability distributions over parameters
- But: Difficult to understand in practice

**Verdict**: Interesting theoretically, but implementation burden is too high for marginal benefit over simpler Bayesian methods.

---

### 4.3 Transfer Learning via Pre-trained Models (Practical Alternative)
**What it does**: Train model on large geopolitical dataset, adapt to your specific portfolio.

**Viable for your system?** ✗ NO (requires external resources)
- **Blocker**: No publicly available geopolitical→market impact models
- You would need to: Build model offline, deploy as ONNX/TFLite
- Go inference: Possible but complex

**Verdict**: Interesting for future, but not immediately viable.

---

### Recommendation for Few-Shot / Meta-Learning
**Skip meta-learning entirely.** Reasons:
1. **Data availability**: You don't have massive geopolitical event datasets
2. **Training cost**: Not affordable on embedded hardware
3. **Interpretability**: Black-box models risky for managing retirement funds
4. **Simpler alternatives exist**: Symbolic regression + causal inference are better suited

**Alternative (if you want "adapt quickly to new situations")**:
- Use **Online Bayesian learning** (next section) instead
- Faster updates, interpretable, lighter-weight
- Same goal: Learn from recent data without full retraining

---

## 5. ACTIVE LEARNING SYSTEMS

### Overview
Asks "which data points should I collect next?" to maximize learning efficiency. Critical for expensive data collection (e.g., geopolitical event metadata).

### 5.1 Uncertainty Sampling
**What it does**: Ask human to label examples where model is most uncertain.

**Viable for your system?** ✓ YES
- **Resource requirements**: Minimal
  - Computation: Identify uncertain predictions (O(n))
  - No retraining needed (just sampling strategy)

**Implementation complexity**: Low
- Simple: Score predictions by uncertainty
- Rank by uncertainty
- Ask user to label top N

**Interpretability**: Excellent
- Shows: Which events is system unsure about?
- User feedback: Refines event catalog

**For "event → market impact"**: Good strategy
```
Example workflow:
1. Train model on 50 events with known impacts
2. Apply to 500 unlabeled historical events
3. Identify: Events where model predicts impact = [0.4, 0.6] (very uncertain)
4. Ask user: "What actually happened after this event?"
5. User provides labels (5-10 high-uncertainty events)
6. Retrain with new labels (big improvement)
7. Repeat
```

**Learning data requirements**: Low (that's the point!)
- Start: 30-50 labeled events
- Active learning finds: Which unlabeled events are most informative
- Each iteration adds 5-10 high-value labels

**Feedback loop mechanism**: Strong
- Weekly: Score all historical events for uncertainty
- Monthly: User labels top 10 uncertain events
- Retrain: Updated model with new labels
- Measure: Confidence on held-out test set improves

**Trade-offs**:
- ✓ Efficient label collection
- ✓ Minimal computation
- ✓ Integrates with any learning method (causal, regression, etc.)
- ✗ Requires user interaction (not autonomous)
- ✗ Depends on user labeling quality
- ✗ Can't distinguish "data error" from "genuine uncertainty"

**Recommended implementation**:
```go
// In internal/modules/causal or symbolic_regression

type UncertaintySampler struct {
    model interface{}  // Existing causal model or regression formula
    repository interface{}  // Event repository
}

type UncertainEvent struct {
    EventID int64
    EventDescription string
    ModelPrediction float64
    PredictionUncertainty float64  // std dev or confidence interval width
    Rank int  // 1 = most uncertain
}

func (us *UncertaintySampler) SampleUncertainEvents(
    topK int,
) ([]UncertainEvent, error) {
    // 1. Get all unlabeled events
    // 2. Score each with model
    // 3. Compute uncertainty (e.g., width of 95% CI)
    // 4. Return top K most uncertain
}

// User labels the uncertain event via API
func (us *UncertaintySampler) RecordEventLabel(
    eventID int64,
    actualImpact float64,  // User provides observed impact
) error {
    // Store label in agents.db
    // Trigger model retraining
}
```

**Integration with UI**:
- New tab: "Uncertain Events"
- Shows: Top 10 uncertain historical events
- User clicks → enters observed impact
- System auto-retrains and shows improved confidence

**Storage**:
- `agents.db`: Add column to events table for user-labeled impacts
- Track: Which events were user-labeled (vs. auto-labeled)

---

### 5.2 Query-by-Committee
**What it does**: Train multiple models, ask human to label examples where models disagree most.

**Viable for your system?** ✓ YES
- **Resource requirements**: Modest
  - Train 3-5 diverse models (slightly higher computation)
  - Scoring: Find disagreement points (O(n))

**Implementation complexity**: Medium
- Train multiple independent models (causal + symbolic regression + HMM)
- Compute disagreement metric
- Rank by disagreement

**For "event → market impact"**: Excellent
```
Example:
Model A (Causal): "War → -2.5% impact"
Model B (Symbolic Regression): "War → -1.8% impact"
Model C (HMM-based): "War → -3.1% impact"
Disagreement = high → Ask user: "What actually happened?"
```

**Trade-offs**:
- ✓ Finds examples where models conflict
- ✓ Focuses labeling on hardest cases
- ✓ More sophisticated than uncertainty sampling
- ✗ Requires training multiple models
- ✗ Higher computation cost
- ✗ Still requires user interaction

**Recommended implementation**: Medium priority (Phase 2)
- Start with uncertainty sampling (simpler)
- Add committee later when multiple models mature

---

### Recommendation for Active Learning
**Priority: High** (Phase 1)

**Start with**: Uncertainty sampling
- Minimal computation cost
- Integrates with existing learning methods
- User-friendly: "These 10 events are unclear—help us learn!"
- Monthly process: 30 minutes of user time → significant model improvement

**Architecture**:
```
Trained Models (Causal + Symbolic Regression)
    ↓
Score all historical events
    ↓
Identify top K most uncertain
    ↓
UI: Show uncertain events to user
    ↓
User provides: Observed impact, additional context
    ↓
agents.db: Store labels + confidence
    ↓
Trigger model retraining
    ↓
Measured outcome: Confidence on test set improves
```

**User experience**:
```
Dashboard Tab: "Help Us Learn"
- "We're unsure about these 10 historical events"
- List shows: Date, event description, our prediction, our uncertainty
- User clicks → Modals to enter observed impact + notes
- System auto-retrains and thanks user
- Show: "Confidence improved from 0.73 to 0.78!"
```

---

## 6. PROBABILISTIC / BAYESIAN APPROACHES

### Overview
Model uncertainty explicitly. Perfect fit for portfolio management (risk is your business).

### 6.1 Bayesian Linear Regression with Spike-and-Slab Priors
**What it does**: Discover which features (event types) actually matter via automatic feature selection.

**Viable for your system?** ✓ YES
- **Resource requirements**: Low
  - Memory: <50MB (prior + posterior)
  - CPU: Gibbs sampling is tractable (few minutes)

**Implementation complexity**: Medium
- Algorithm: Gibbs sampling for spike-and-slab posteriors
- Go implementation: ~400 lines of MCMC
- Tricky: Convergence diagnostics

**Interpretability**: EXCELLENT
- Output: P(coefficient ≠ 0 | data) for each feature
- Simple threshold: Features with P(≠ 0) > 0.95 are "real"
- Uncertainty intervals: "Impact is -2.5% ± 0.8% (95% CI)"

**For "event → market impact"**: Perfect match
```
Example output:
Feature              P(≠0)   Mean      95% CI
War declaration      0.94    -2.50%    [-3.10%, -1.95%]
Sanctions (broad)    0.87    -1.20%    [-1.85%, -0.55%]
Political tension    0.42    -0.05%    [-0.45%, +0.35%]
Central bank action  0.96    +0.80%    [+0.45%, +1.15%]
```
→ Automatically discovered: "War and sanctions matter, tension doesn't"

**Learning data requirements**: Moderate
- Start: 30-50 events with observed impacts
- Gibbs sampler converges quickly (fewer data points needed than frequentist)
- Uncertainty increases naturally as N decreases

**Feedback loop mechanism**: Strong
- Monthly: Add new events to dataset
- Rerun Gibbs sampler (30 seconds)
- Compare posterior to prior (did we learn?)
- Track: P(≠ 0) changes over time

**Trade-offs**:
- ✓ Automatic feature selection
- ✓ Uncertainty quantified naturally
- ✓ Lightweight computation
- ✓ Integrates with causal methods
- ✗ Assumes linear relationships
- ✗ MCMC convergence can be fussy
- ✗ Requires careful prior specification

**Recommended implementation**:
```go
// In internal/modules/bayesian (new module)

type SpikeSlabPrior struct {
    spikeProbability float64  // P(feature is active)
    slabMean         float64  // N(0, σ²) for active features
    slabVariance     float64
}

type BayesianEventModel struct {
    features []string  // Feature names (event types)
    priors   []SpikeSlabPrior
    posteriorSamples []float64  // MCMC samples
    log zerolog.Logger
}

func (bem *BayesianEventModel) FitGibbs(
    X [][]float64,  // Feature matrix (events)
    y []float64,    // Impacts
    numIterations int,
) error {
    // Gibbs sampling for spike-and-slab
    // Store posterior samples
}

type FeatureSignificance struct {
    Feature string
    ProbabilityNotZero float64
    PosteriorMean float64
    CredibleInterval [2]float64
}

func (bem *BayesianEventModel) GetFeatureSignificance() []FeatureSignificance {
    // Summarize posterior: P(≠ 0), mean, CI for each feature
}
```

**Storage**:
- `agents.db`: Store posterior summaries (means, intervals)
- `history.db`: MCMC trace (for convergence diagnostics)

**Integration with symbolic regression**:
```
Bayesian Linear Regression Path:
Discovered: P(War≠0) = 0.94, P(Tension≠0) = 0.42
    ↓
Use to initialize symbolic regression:
    → Genetic algorithm should prioritize War, deprioritize Tension
    → Faster convergence
```

---

### 6.2 Gaussian Process Regression (GPR)
**What it does**: Non-parametric Bayesian method. Learns function shape (smooth/wiggly) and uncertainty directly from data.

**Viable for your system?** ✓ YES (constrained)
- **Resource requirements**: Moderate
  - Memory: 50-150MB (covariance matrix for 100-200 events)
  - CPU: Matrix inversion O(n³) — don't scale beyond 500 events

**Implementation complexity**: Medium-High
- Algorithm: Straightforward (compute K matrix, invert, mean/variance)
- Tricky: Numerical stability (use Cholesky decomposition)
- Hyperparameter tuning: Manual or cheap grid search

**Interpretability**: Good
- Function posterior: Mean (prediction) + std dev (uncertainty)
- Can visualize: Learned function with confidence bounds
- Limitation: Doesn't explain *why* (black-box function)

**For "event → market impact"**: Good fit
```
Example:
Input: Sentiment score (-1 to 1)
Output: Expected impact distribution
Prediction at Sentiment=0.3:
  Expected: -0.8%
  95% CI: [-1.4%, -0.2%]
  (Very confident; we've seen similar events)

Prediction at Sentiment=0.9:
  Expected: -2.1%
  95% CI: [-4.5%, +0.3%]
  (Uncertain; few similar extreme events in training)
```

**Learning data requirements**: Moderate
- 50-100 examples sufficient (Bayesian ≠ data-hungry)
- Uncertainty naturally reflects data scarcity
- Can add incrementally (one event per month)

**Feedback loop mechanism**: Strong
- Monthly: Add new event (recompute posterior)
- Weekly: Score unlabeled events (get uncertainty)
- Active learning: Query user on most uncertain events
- Quarterly: Review fitted function shape

**Trade-offs**:
- ✓ Non-parametric (flexible, no assumptions on functional form)
- ✓ Uncertainty quantified properly
- ✓ Incremental learning easy
- ✓ Interpretable as learned function
- ✗ Cubic scaling (slow beyond 500 events)
- ✗ Requires kernel choice (RBF? Matérn? Periodic?)
- ✗ Hyperparameter tuning needed
- ✗ Doesn't identify *which* features matter (unlike spike-and-slab)

**Recommended implementation**:
```go
// In internal/modules/bayesian

type GaussianProcessRegressor struct {
    trainingX [][]float64  // Event feature vectors
    trainingY []float64    // Observed impacts

    kernel interface{}  // Kernel function (e.g., RBF)
    noise  float64      // Measurement noise variance

    // Cached for efficiency
    covarianceMatrix [][]float64  // K matrix
    alpha            []float64    // (K + σ²I)⁻¹ y
}

type GPPrediction struct {
    Mean     float64
    StdDev   float64
    ConfidenceInterval [2]float64
}

func (gp *GaussianProcessRegressor) Fit(
    X [][]float64,
    y []float64,
) error {
    // Compute covariance matrix
    // Invert (K + σ²I)
    // Solve for α
}

func (gp *GaussianProcessRegressor) Predict(
    x []float64,  // New event feature vector
) GPPrediction {
    // Compute k(x, X)
    // Mean = k^T α
    // Var = k(x, x) - k^T (K + σ²I)⁻¹ k
}

func (gp *GaussianProcessRegressor) AddObservation(
    x []float64,
    y float64,
) error {
    // Incremental update (recompute from scratch)
    // Efficient enough for monthly updates
}
```

**Kernel choice recommendation**:
- **RBF kernel** (default): Works well when you don't know structure
- **Matérn kernel**: Better for underlying smooth functions
- Start with RBF

**Hyperparameter tuning**:
- Noise variance σ²: Start at 0.01² (1% of typical impact)
- Kernel lengthscale: Use 50% of feature range
- Refine monthly with marginal likelihood maximization (cheap grid search)

---

### 6.3 Dirichlet Process Mixture Models (Nonparametric Mixture)
**What it does**: Discover unknown number of event types/clusters automatically.

**Viable for your system?** ✗ MARGINAL
- **Resource requirements**: Moderate-High
  - MCMC for DP is expensive (slow convergence)
  - Memory: 100MB+ for posterior samples

**Implementation complexity**: Very High
- Algorithm: Chinese Restaurant Process MCMC is intricate
- Go implementation: Research-level code
- Convergence diagnostics: Complex

**Verdict**: Skip initially. Use simpler approaches. Revisit if needed.

---

### Recommendation for Bayesian Methods
**Priority: Medium-High** (Phase 1)

**Start with**: Bayesian Linear Regression (spike-and-slab)
- Automatic feature selection (find which events matter)
- Fast computation, easy to implement
- Interpretable uncertainty quantification
- Monthly retraining: ~30 seconds

**Phase 2**: Add Gaussian Process Regression
- For non-linear relationships
- Soft prediction bounds (uncertainty increases with extrapolation)
- More flexible than linear, still interpretable

**Architecture**:
```
Event Training Data (Event Type/Features + Observed Impact)
    ↓
Split into: {War, Sanctions, Tension, Central Bank, ...}
    ↓
Bayesian Linear Regression
    → Discovers: P(each event type matters)
    → Output: P(≠0), mean, credible interval
    ↓
agents.db (event_feature_significance)
    ↓
For each significant event type:
    Gaussian Process Regression
    → Learn: non-linear impact function
    → Output: predictions with uncertainty
    ↓
agents.db (gp_posteriors for prediction)
    ↓
Feedback loop:
    - Weekly: Score unlabeled historical events
    - Monthly: User labels uncertain events + retrain
    - Quarterly: Review feature significance changes
```

---

## 7. GRAPH-BASED LEARNING (KNOWLEDGE GRAPHS, GNNs)

### Overview
Represent geopolitical world as graph: Events → Countries → Markets → Securities. Learn patterns on this structure.

### 7.1 Knowledge Graph Construction + Symbolic Query
**What it does**: Build explicit graph of relationships. Query via SPARQL-like syntax or graph traversal.

**Viable for your system?** ✓ YES (lightweight variant)
- **Resource requirements**: Low-Medium
  - Memory: 50-100MB (small graphs are efficient)
  - CPU: Graph traversal (BFS/DFS) is cheap

**Implementation complexity**: Medium
- Build: Manual ontology (Event, Country, Market, Security)
- Store: Triples (subject, predicate, object)
- Query: Simple rule engine or graph database (SQLite works!)

**Interpretability**: EXCELLENT
- Visual: Draw the graph, inspect edges
- Transparent: "War in Ukraine affects Oil prices affects Energy stocks"
- Debuggable: Trace reasoning path

**For "event → market impact"**: Very good
```
Example graph:
  [War in Ukraine]
  ├─ affects_commodity → [Oil]
  ├─ affects_country → [Ukraine, Russia]
  └─ affects_region → [Europe]

  [Oil price increase]
  ├─ affects_sector → [Energy, Transportation]
  └─ affects_country_negatively → [Oil importers]

  [Energy stocks]
  ├─ sector_of → [Energy]
  ├─ affected_by → [Oil price]
  └─ in_portfolio? → Yes (e.g., BP, Shell)

  → Inference: War in Ukraine → Oil up → Energy stocks up → Portfolio impact +2.5%
```

**Learning data requirements**: Low
- Start: Manual ontology + 50-100 events
- Rules: Human-specified relationships
- Grow over time: Add new event types, relations

**Feedback loop mechanism**: Strong
- Monthly: New events added to graph
- Yearly: Add new relationship types (learn new causal patterns)
- Quarterly: Compare predictions (following graph) vs realized impacts
- Refine: Update edge weights (strengthen/weaken relationships)

**Trade-offs**:
- ✓ Transparent reasoning
- ✓ Lightweight computation
- ✓ Interpretable by design
- ✓ Integrates with causal reasoning
- ✗ Requires manual knowledge encoding
- ✗ Hard to scale to 1000s of entities
- ✗ Doesn't learn structure (humans define it)
- ✗ Brittle if ontology is wrong

**Recommended implementation**:
```go
// In internal/modules/knowledge_graph (new module)

type Triple struct {
    Subject   string  // e.g., "war_ukraine_2022"
    Predicate string  // e.g., "affects_commodity"
    Object    string  // e.g., "oil"
    Weight    float64 // e.g., 0.85 (confidence in relationship)
}

type KnowledgeGraph struct {
    triples   []Triple
    entities  map[string]Entity  // Entity metadata
    log       zerolog.Logger
}

type Entity struct {
    ID              string
    Type            string  // "Event", "Commodity", "Market", "Security"
    Name            string
    Metadata        map[string]interface{}
    ImpactOnReturns float64  // If known
}

// Rule-based inference
type InferenceRule struct {
    Antecedents []string  // e.g., ["War in Ukraine", "affects_commodity", "Oil"]
    Consequents []Triple  // e.g., [{Oil, increases_price, Energy}]
    Confidence  float64
}

func (kg *KnowledgeGraph) QueryPath(
    source string,
    target string,
    relationTypes []string,  // Filter edges
) ([][]Triple, error) {
    // Find all paths from source to target
    // Following specified relation types
}

func (kg *KnowledgeGraph) InferImpact(
    event *GeopoliticalEvent,
) ([]ImpactPrediction, error) {
    // 1. Add event to graph
    // 2. Run inference rules
    // 3. Forward-chain to reach portfolio securities
    // 4. Return: Securities that should be affected + predicted impact
}

type ImpactPrediction struct {
    Security   string
    PredictedImpact float64
    ReasoningPath []Triple  // Show how we got here
    Confidence float64
}
```

**Storage**:
- `agents.db`: New tables for triples, entities, rules
- Or: Lightweight RDF store (SPARQL endpoint)

**Integration with causal methods**:
```
Knowledge Graph + Causal Inference:
1. Use graph to identify candidate causal relationships
2. Test via 2SLS (instrumental variables)
3. Update graph edge weights based on causal estimates
4. Cycle: Better graph → better hypotheses → better causal testing
```

---

### 7.2 Graph Neural Networks (GNNs)
**What it does**: Learn node embeddings by aggregating neighbor information. Discover patterns in graph structure.

**Viable for your system?** ✗ MARGINAL (too heavy)
- **Resource requirements**: Moderate-High
  - Memory: 150-300MB (node embeddings + model)
  - CPU: Graph convolutions (expensive)
  - Training: Requires automatic differentiation (TensorFlow)

**Implementation complexity**: Very High
- Requires: Deep learning framework
- Go support: Weak (TensorFlow Go bindings are outdated)
- Training: Expensive, requires GPU (not available on ARM)

**Verdict**: Skip for now. Knowledge graph + symbolic query is lighter and more interpretable.

---

### Recommendation for Graph-Based Approaches
**Priority: Medium** (Phase 2)

**Start with**: Knowledge Graph + symbolic query
- Manual but transparent
- Build incrementally (start small, grow)
- Integrate with causal inference
- Reasonable effort for significant interpretability gain

**Architecture**:
```
Manual Ontology (Event, Country, Commodity, Sector, Security)
    ↓
Event Data Curation
    ↓
Build Knowledge Graph (Triples)
    ↓
Inference Rules (manually specified)
    ↓
Given new event:
  1. Add to graph
  2. Run inference rules
  3. Forward-chain to securities
  4. Predict impact
    ↓
agents.db (knowledge_graph tables)
    ↓
Feedback:
  - Compare predictions to realized impact
  - Update edge weights (strengthen/weaken)
  - Add new relation types (learn new patterns)
```

**Example initial ontology** (simplified):
```
Entities:
- Event types: War, Sanctions, TradeTension, MonetaryPolicy, NaturalDisaster
- Geographies: Country, Region
- Markets: Commodity (Oil, Gold, Wheat), Currency
- Portfolio: Sector, Security

Relations:
- affects_commodity (strength: 0-1)
- affects_currency (strength)
- affects_sector (strength)
- correlates_with_security (strength)
- increases / decreases (for price impacts)
```

---

## 8. EMBEDDING-BASED SIMILARITY APPROACHES

### Overview
Represent events as vectors in learned space. Similar events should have similar impacts.

### 8.1 Event Embedding (Word2Vec-like)
**What it does**: Train embeddings so that events with similar market impacts cluster together.

**Viable for your system?** ✓ YES (constrained)
- **Resource requirements**: Low
  - Memory: 20-50MB (embedding vectors)
  - CPU: Skip-gram training is efficient

**Implementation complexity**: Medium
- Algorithm: Skip-gram with negative sampling
- Go implementation: ~300-400 lines
- Or: Use existing word2vec libraries (goml/nlp)

**Interpretability**: Medium
- Embeddings are opaque but can visualize (t-SNE projection)
- Cluster analysis: "These 5 events cluster together → similar impact?"

**For "event → market impact"**: Decent fit
```
Example:
1. Train embeddings where context = "observed market return"
2. Events with similar embeddings had similar returns
3. New event → find k-nearest neighbors → average their impacts

This is basically: "Find historical events similar to new event, predict new event has similar impact"
```

**Learning data requirements**: Low-Moderate
- 100-200 historical events minimum
- Each event needs: Description + observed impact

**Feedback loop mechanism**: Medium
- Quarterly: Retrain embeddings with new events
- Longer feedback cycle (embeddings are somewhat stable)
- Compare: Neighbors' impacts vs predicted impact

**Trade-offs**:
- ✓ Lightweight computation
- ✓ Can find similarity automatically
- ✓ Works with small data
- ✗ Opaque (embeddings don't explain causation)
- ✗ Assumes similar events → similar impacts (may not be true)
- ✗ Requires text descriptions (harder to automate)
- ✗ Breaks under distribution shift (new event types)

**Recommended implementation**:
```go
// In internal/modules/embeddings (new module)

type EventEmbedding struct {
    EventID int64
    Vector  []float64  // Embedding vector (128-256 dims)
}

type EmbeddingModel struct {
    embeddings map[int64][]float64
    vocabSize int
    embeddingDim int
    log zerolog.Logger
}

func (em *EmbeddingModel) TrainSkipGram(
    events []Event,
    contextWindowSize int,
    numIterations int,
) error {
    // Skip-gram training
    // Context = market return in following week
}

type SimilarEvent struct {
    EventID int64
    Description string
    CosineSimilarity float64
    ObservedImpact float64
}

func (em *EmbeddingModel) FindSimilarEvents(
    newEvent *Event,
    k int,  // Return top k
) ([]SimilarEvent, error) {
    // Find k-nearest neighbors in embedding space
}

func (em *EmbeddingModel) PredictImpact(
    newEvent *Event,
) (predictedImpact float64, confidence float64, err error) {
    // 1. Embed new event
    // 2. Find k-nearest neighbors
    // 3. Return: weighted average of neighbor impacts
    // 4. Confidence: inverse of variance in neighbor impacts
}
```

**Storage**:
- `agents.db`: Store embeddings + trained model parameters

---

### 8.2 Semantic Similarity via Textual Features
**What it does**: Use text descriptions to compute similarity (no neural training required).

**Viable for your system?** ✓ YES
- **Resource requirements**: Minimal
  - Memory: <10MB
  - CPU: String similarity (O(n²) but fast for small n)

**Implementation complexity**: Low
- Algorithm: TF-IDF or simple cosine similarity on word counts
- Go implementation: ~200 lines

**For "event → market impact"**: Same logic as embeddings
```
Example:
New event: "Russia announces sanctions against US tech companies"
Historical similar events:
  1. "China announces sanctions against US" (observed: -1.5% impact)
  2. "Russia restricts tech imports" (observed: -1.8% impact)
  3. "US responds with counter-sanctions" (observed: -2.1% impact)

Predicted impact: Weighted average ≈ -1.8% (with wide confidence band)
```

**Trade-offs**:
- ✓ No training required
- ✓ Extremely lightweight
- ✓ Works with any event description
- ✗ Text similarity ≠ causal similarity
- ✗ Brittle to paraphrasing
- ✗ Misses latent structure (different phrasings)

**Recommended implementation**: Combine with embeddings
- Start with TF-IDF (no training, quick prototype)
- Upgrade to embeddings if needed

---

### Recommendation for Embedding Approaches
**Priority: Low** (Phase 3, optional)

**Only pursue if**:
- You have many historical event descriptions
- Causal methods (Section 1) plateau
- Want to try "similarity-based prediction" as baseline

**Best path**: TF-IDF similarity (zero training cost)
```go
func ComputeEventSimilarity(event1, event2 string) float64 {
    // Tokenize, count words, compute cosine similarity
}

func PredictImpactBySimilarity(
    newEvent *Event,
    historicalEvents []*Event,
) float64 {
    // Find similar events, average their impacts
}
```

**Advanced (if worthwhile)**: Event embeddings via skip-gram
- 2-3 weeks of implementation
- Quarterly retraining
- Likely outperforms TF-IDF

---

## 9. REINFORCEMENT LEARNING FOR STRATEGY DISCOVERY

### Overview
Learn trading/allocation strategies by interacting with simulated market. RL agent discovers what to do when events occur.

### 9.1 Model-Based RL (MBRL)
**What it does**: Learn forward model (event → market state), then plan optimal actions via model predictive control.

**Viable for your system?** ✗ NO
- **Resource requirements**: High
  - Memory: 200-500MB (value function approximation)
  - CPU: Planning via search trees is expensive
  - Training: Requires simulation (data-inefficient)

**Implementation complexity**: Very High
- Requires: Reinforcement learning framework, simulator, solver
- Go: No mature RL libraries
- Debugging: Complex (credit assignment problem)

**Interpretability**: Poor
- RL policies are black-box (hard to explain decisions)
- Risky for managing retirement funds

**Verdict**: Skip. Portfolio management is already a solved control problem (optimization). RL adds complexity without benefit.

---

### 9.2 Imitation Learning (Learning from Expert Demonstrations)
**What it does**: Observe expert portfolio manager's actions after events, learn to imitate.

**Viable for your system?** ✗ MARGINAL
- **Resource requirements**: Moderate
  - You're the only "expert" (your trading history)
  - Need: Data on your past event-response decisions

**Implementation complexity**: Medium-High
- Requires: Behavioral cloning or inverse RL setup
- Go: Possible but limited libraries
- Problem: You have limited demonstrations

**Verdict**: Skip. Philosophically doesn't align with your goal (you're not the expert you're trying to learn from; geopolitics→markets is the expert).

---

### Recommendation for RL Approaches
**Verdict: SKIP ENTIRELY**

Reasons:
1. **Misaligned problem**: RL is for learning sequential decision-making under uncertainty. But you have:
   - Clear objective: Optimize portfolio returns
   - Known constraints: Risk limits, rebalancing rules, etc.
   - Better solved by: Optimization + causal reasoning (not learning)

2. **Interpretability**: RL policies are opaque. Risk management requires transparent decision logic.

3. **Data efficiency**: You'd need thousands of simulated "event interactions" to train RL. Too expensive.

4. **Existing solutions**: Your planning module (planning/) already does this via optimization. Causal inference tells you what markets will do. RL is redundant.

---

## 10. ANOMALY DETECTION + OPPORTUNITY DETECTION

### Overview
Detect unusual patterns (anomalies). When paired with causal inference, anomalies become opportunities.

**You've already seen** this in Section 2 (Time Series). This section focuses on integration and opportunity detection.

### 10.1 Anomaly Detection + Causal Root Cause Analysis

**What it does**: Detect anomaly, then ask "why?" via causal methods.

**Viable for your system?** ✓ YES (already partially implemented)
- **Resource requirements**: Low
  - Anomaly detection: ~50MB
  - Causal analysis: ~20MB

**Implementation complexity**: Low-Medium
- Combine existing change point detection + IV estimation
- New: Attribution (which causal factors explain the anomaly?)

**For "event → market impact"**: Excellent
```
Example workflow:
1. [Anomaly Detection] VIX jumps 25% on Day X (unusual)
2. [Check Event Calendar] "FOMC Meeting announced 3 hours before jump"
3. [Causal Inference] Run IV regression: "FOMC announcements → VIX change"
4. [Attribution] Estimate: How much of the 25% jump is FOMC-driven?
   (vs. mechanical rebalancing, technical breakout, etc.)
5. [Action] "This anomaly is explained by FOMC. Update VIX forecast."
```

**Implementation**:
```go
// Combine existing modules

func (cis *CausalInferenceService) ExplainAnomaly(
    anomalyDate time.Time,
    anomalyMagnitude float64,
    associatedEvents []GeopoliticalEvent,
) (*AnomalyExplanation, error) {
    // 1. Look back 1 month from anomaly date
    // 2. Collect events and price changes
    // 3. Run 2SLS with event dummies
    // 4. Attribution: How much of anomaly is each event?
    // 5. Return: Explained % + p-value of explanation
}

type AnomalyExplanation struct {
    AnomalyDate time.Time
    AnomalyMagnitude float64
    ExplainedBy map[string]float64  // Event type → attributed impact
    ResidualUnexplained float64      // What's left (data error? latent factors?)
    Confidence float64              // P-value of explanation
}
```

**Storage**: `agents.db` (anomaly_explanations table)

**Feedback loop**: Strong
- Weekly: Detect anomalies
- Match to events + run attribution
- If residual is large: Investigate (missing events? data errors?)
- Refine event catalog

---

### 10.2 Opportunity Detection

**What it does**: When causal inference + anomaly detection find "mismatch" (event happened but market didn't react, or over-reacted), flag as opportunity.

**Viable for your system?** ✓ YES
- **Resource requirements**: Minimal
  - Comparison logic: O(1) per opportunity

**Implementation complexity**: Low
- Simple threshold: |predicted impact - realized impact| > threshold
- Advanced: Bayesian surprise (log likelihood ratio)

**For "event → market impact"**: Direct application
```
Examples:
1. War declared in distant region. Causal model predicts -2% impact.
   Realized: Market up 1%. Surprise: +3%. Opportunity: Underpriced?

2. Central bank hikes rates. Model predicts +1% bonds, -0.5% stocks.
   Realized: Bonds down 2%, stocks down 3%. Surprise: Tech panic.
   Opportunity: Are growth stocks now cheap?
```

**Implementation**:
```go
type Opportunity struct {
    EventID int64
    Date time.Time
    EventDescription string
    CausalPrediction float64
    RealizedImpact float64
    Surprise float64  // |Predicted - Realized|
    Type string  // "underreaction", "overreaction"
    Confidence float64
    RecommendedAction string  // "Buy", "Hold", "Sell"
}

func (cis *CausalInferenceService) DetectOpportunities(
    recentEvents []GeopoliticalEvent,
    lookforwardDays int,
) ([]Opportunity, error) {
    // 1. For each recent event:
    //    - Get causal prediction
    //    - Get realized market return in lookforwardDays
    //    - Compute surprise
    // 2. Filter high-confidence surprises
    // 3. Suggest actions (contrarian bets? cover shorts?)
}
```

**Storage**: `cache.db` (opportunities table with TTL)

**UI integration**:
- Dashboard widget: "Potential Opportunities"
- Shows: Events with unusual market responses
- User can: Mark as "acted on" or "ignored"
- Feedback: Track accuracy of opportunity flagging

---

### Recommendation for Anomaly Detection
**Priority: High** (Phase 1)

**Integrate**:
1. Existing change point detection + isolation forest
2. Causal inference (2SLS) to explain anomalies
3. Opportunity detection on top

**Minimal additional code**:
```go
// New module: internal/modules/anomaly_explanation

func ExplainAnomalyWithCausalInference(
    anomalyDate time.Time,
    anomalyMagnitude float64,
    eventCatalog []GeopoliticalEvent,
) (*AnomalyExplanation, error) {
    // Orchestrate: CPD output → 2SLS → Attribution
}

func DetectOpportunitiesBySurprise(
    predictions []CausalPrediction,
    realizedReturns []float64,
    lookforwardDays int,
) ([]Opportunity, error) {
    // Compute: Predicted vs Realized for each event
    // Flag: High-magnitude surprises
}
```

---

## SYNTHESIS & RECOMMENDATION MATRIX

### Viability Overview

| Approach | Feasibility | Resource Cost | Complexity | Interpretability | "Event→Impact" Fit | Data Needs | Priority |
|----------|-------------|----------------|------------|------------------|-------------------|-----------|----------|
| **1. Causal Inference (2SLS)** | ✓ YES | LOW | MED | EXCELLENT | EXCELLENT | MOD | HIGH |
| **1b. DAG Reasoning** | ✓ YES | LOW | LOW-MED | EXCELLENT | EXCELLENT | MOD | HIGH |
| **2. Change Point Detection** | ✓ YES | LOW | LOW-MED | EXCELLENT | EXCELLENT | LOW | HIGH |
| **2b. Anomaly Detection (IF)** | ✓ YES | LOW | LOW | GOOD | GOOD | LOW | HIGH |
| **2c. Hidden Markov Models** | ✓ YES | LOW | MED | GOOD | GOOD | LOW | MED |
| **3. Symbolic Regression** | ✓ YES (ext) | MOD | MED | EXCELLENT | EXCELLENT | MOD-HIGH | HIGH |
| **4. Few-Shot Learning** | ✗ NO | HIGH | VERY HIGH | POOR | GOOD | HIGH | - |
| **5. Active Learning** | ✓ YES | LOW | LOW | EXCELLENT | EXCELLENT | LOW | HIGH |
| **6a. Bayesian Linear Reg** | ✓ YES | LOW | MED | EXCELLENT | EXCELLENT | MOD | HIGH |
| **6b. Gaussian Processes** | ✓ YES | MOD | MED-HIGH | GOOD | GOOD | MOD | MED |
| **7. Knowledge Graphs** | ✓ YES | LOW | MED | EXCELLENT | EXCELLENT | LOW | MED |
| **7b. Graph Neural Nets** | ✗ NO | HIGH | VERY HIGH | POOR | GOOD | HIGH | - |
| **8. Event Embeddings** | ✓ YES (opt) | LOW | MED | MEDIUM | GOOD | MOD | LOW |
| **8b. Semantic Similarity** | ✓ YES | LOW | LOW | GOOD | GOOD | LOW | LOW |
| **9. Reinforcement Learning** | ✗ NO | HIGH | VERY HIGH | POOR | MEDIUM | HIGH | - |
| **10a. Anomaly+Attribution** | ✓ YES | LOW | LOW-MED | EXCELLENT | EXCELLENT | LOW | HIGH |
| **10b. Opportunity Detection** | ✓ YES | LOW | LOW | EXCELLENT | EXCELLENT | MOD | HIGH |

### Recommended Implementation Path

#### Phase 1 (Months 1-2): Foundation
Build these four approaches in parallel:

1. **Causal Inference (2SLS + Instrumental Variables)**
   - New module: `internal/modules/causal/`
   - Monthly job: Estimate event causal effects
   - Output: agents.db (event_causal_effects table)
   - Effort: 2-3 weeks

2. **Change Point Detection + Anomaly Detection**
   - Extend: `internal/modules/time_series/` (new)
   - Daily job: Scan for structural breaks and anomalies
   - Output: history.db + agents.db
   - Effort: 2 weeks (reuse PELT/IF algorithms)

3. **Active Learning Integration**
   - Extend: Existing uncertainty sampling
   - Weekly job: Identify uncertain events for labeling
   - UI: New "Help Us Learn" tab
   - Effort: 1 week

4. **Bayesian Linear Regression (Spike-and-Slab)**
   - New module: `internal/modules/bayesian/`
   - Monthly job: Feature selection for event types
   - Output: agents.db (feature_significance table)
   - Effort: 2 weeks (MCMC implementation)

**Phase 1 output**:
- 4 independent learned models of geopolitical→market impacts
- agents.db with causal effects, event impacts, feature importance
- Weekly feedback loop identifying uncertain events
- Monthly retraining as new events accumulate
- Total effort: ~8-10 weeks

#### Phase 2 (Months 3-4): Enhancement
Add these if Phase 1 is successful:

5. **Hidden Markov Models** (replace current regime detection)
   - Upgrade: `internal/market_regime/`
   - Effort: 2 weeks

6. **Symbolic Regression for Events**
   - Extend: `internal/modules/symbolic_regression/`
   - Effort: 2-3 weeks

7. **Gaussian Processes**
   - Add: `internal/modules/bayesian/gp.go`
   - Effort: 2 weeks

**Phase 2 output**:
- 7 independent models (ensemble)
- Better regime detection
- Interpretable learned formulas
- Principled uncertainty quantification
- Total effort: ~8-10 weeks

#### Phase 3 (Month 5+): Optional
- Knowledge graphs (if you want explicit reasoning)
- Event embeddings (if you have extensive text descriptions)
- Query-by-committee active learning (if Phase 1 user labeling goes well)

---

## ARCHITECTURE INTEGRATION

All approaches should integrate via:

1. **agents.db schema**:
   ```sql
   -- Causal effects
   CREATE TABLE event_causal_effects (
       event_id INTEGER,
       causal_coefficient FLOAT,
       confidence_interval_low FLOAT,
       confidence_interval_high FLOAT,
       updated_at TIMESTAMP
   );

   -- Feature significance
   CREATE TABLE feature_significance (
       feature_name TEXT,
       p_not_zero FLOAT,  -- P(coefficient ≠ 0 | data)
       posterior_mean FLOAT,
       credible_interval_low FLOAT,
       credible_interval_high FLOAT,
       updated_at TIMESTAMP
   );

   -- Discovered formulas
   CREATE TABLE discovered_event_formulas (
       formula_id INTEGER PRIMARY KEY,
       formula TEXT,  -- e.g., "-0.01*Sentiment + 0.5*Sentiment^2 - 0.0003*DaysSince"
       regime TEXT,
       fitness FLOAT,
       updated_at TIMESTAMP
   );

   -- Model predictions
   CREATE TABLE event_impact_predictions (
       event_id INTEGER,
       model_type TEXT,  -- "causal", "symbolic", "gp", "hmm"
       predicted_impact FLOAT,
       uncertainty FLOAT,
       model_version INTEGER,
       predicted_at TIMESTAMP
   );
   ```

2. **Event system integration**:
   ```go
   // Emit events when models update
   type EventModelUpdated struct {
       ModelType string
       Timestamp time.Time
       Accuracy float64
   }

   // Listeners react: Update predictions, refresh dashboard, log metrics
   ```

3. **Scheduler integration**:
   ```go
   // In di/jobs.go, register monthly/weekly jobs:
   scheduler.RegisterJob("causal_inference_monthly", ...)
   scheduler.RegisterJob("anomaly_detection_daily", ...)
   scheduler.RegisterJob("active_learning_weekly", ...)
   scheduler.RegisterJob("bayesian_update_monthly", ...)
   ```

4. **API/UI exposure**:
   ```go
   // New routes in server/routes.go:
   GET /api/causal-effects     // View discovered causal effects
   GET /api/feature-importance // View which event types matter
   GET /api/opportunities      // View flagged opportunities
   POST /api/label-event       // User labels uncertain event → triggers retraining
   ```

---

## HONEST TRADE-OFFS & LIMITATIONS

### What Will Actually Work on 2GB Hardware
- Causal inference (2SLS): YES, trivial
- Change point detection: YES, trivial
- Anomaly detection (IF): YES, trivial
- Bayesian regression: YES, 30-60 seconds monthly
- Symbolic regression (genetic): YES, but slow (30+ minutes monthly)
- Gaussian processes: YES, up to ~500 events (then scaling breaks)
- Knowledge graphs: YES, if small (<1000 triples)
- HMMs: YES, trivial

### What Will NOT Work
- Deep learning: No TensorFlow/PyTorch support, no GPU
- Large GNNs: Memory and CPU prohibitive
- Meta-learning: Requires expensive optimization
- Big RL: Simulation loop is data-inefficient

### What Will Teach You the Most About Geopolitical Patterns
1. **Causal Inference** (2SLS): Most rigorous causality discovery
2. **Change Point Detection**: Simplest → deployed first
3. **Symbolic Regression**: Most interpretable formulas
4. **Bayesian Methods**: Best uncertainty quantification

### What Risks Do You Face?
1. **Overfitting to historical events**: Only 50-100 documented "major" events → small dataset
   - Mitigation: Cross-validation, active learning to expand training data

2. **Non-stationarity**: Geopolitical patterns may shift
   - Mitigation: Quarterly retraining, regime-aware models (HMM), monitoring drift

3. **Selection bias**: Only major events are documented
   - Mitigation: Knowledge graph to surface overlooked patterns

4. **Causality is hard**: True causal effects masked by confounders
   - Mitigation: Multiple methods (agreement → confidence), domain expertise via DAGs

5. **Black swan events**: Methods trained on historical data fail on novel events
   - Mitigation: Always include baseline predictions, don't over-rely on learned models

---

## FINAL RECOMMENDATION

**Implement Phase 1 (in order)**:

1. **Change Point Detection** (Week 1)
   - Simplest, immediate insights
   - Lightweight foundation for other methods

2. **Causal Inference (2SLS)** (Week 2-3)
   - Rigorous causality discovery
   - Monthly retraining cycle

3. **Anomaly Detection + Attribution** (Week 3-4)
   - Use CPD output to explain surprises
   - Opportunity detection

4. **Active Learning** (Week 4-5)
   - User labels uncertain events
   - Grows training dataset

5. **Bayesian Regression** (Week 5-6)
   - Automatic feature selection
   - Quantified uncertainty

6. **Symbolic Regression Extension** (Week 6-8)
   - Discovered formulas for event impacts
   - Regime-aware variants

**Launch with**: Methods 1-5 (2 months)
**Then add**: Method 6 if computational budget allows

**Success metrics**:
- Can explain 60%+ of major market moves post-event with <10% confidence interval width
- User labeling reaches 80%+ agreement with model predictions
- Monthly retraining shows consistent improvement (fewer surprises)
- Opportunity detection flags average +2% outperformance opportunities (backtested)

---

## Appendix: Library Recommendations for Go Implementation

| Task | Library | Notes |
|------|---------|-------|
| Linear algebra | `gonum/matrix` | Matrix ops, eigenvalues (needed for 2SLS) |
| Statistics | `gonum/stat` | Distributions, hypothesis tests |
| MCMC/Bayesian | Custom Gibbs sampler | ~300 lines for spike-and-slab |
| Symbolic regression | Extend existing | Genetic algorithm already implemented |
| Graph algorithms | stdlib `container/list` | BFS/DFS for knowledge graphs |
| Timeseries | PELT implementation | ~200 lines, custom or adapt from R package |
| Logging | `zerolog` | Already in use |

**Avoid**: TensorFlow, PyTorch, scikit-learn (Python). Stick to pure Go or minimal dependencies.
