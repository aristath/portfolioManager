# LLM Advisor Workflow Research

## Overview

This document outlines a proposed workflow for an LLM-based system that monitors news for securities, analyzes market developments, and provides weekly advisory reports. The system runs in parallel with Sentinel and outputs recommendations for human review.

## Design Principles

1. **No single point of trust** — Every significant claim requires multiple sources or multiple model agreements
2. **Preserve provenance** — Every output traces back to raw sources
3. **Fail toward silence** — When uncertain, report "no significant findings" rather than guess
4. **Separate fact from interpretation** — Humans see both, can disagree with interpretation
5. **Adversarial by design** — Include a stage that actively tries to disprove conclusions

## Model Tier Strategy

| Task | Complexity | Volume | Model Tier |
|------|------------|--------|------------|
| Fact extraction from articles | Low | High (per article) | Small/cheap (Haiku-class) |
| Clustering & deduplication | Low | Medium | Small/cheap or algorithmic |
| Change detection (diff) | Low-Medium | Medium (per security) | Small/cheap |
| Significance assessment | High | Low (per delta) | Capable (Sonnet-class) |
| External corroboration | High | Very low (only flagged items) | Capable (Sonnet-class) |
| Adversarial review | High | Very low | Capable (Sonnet-class) |
| Final synthesis | High | Once per run | Most capable (Opus-class) |

### Model Assignment Rules

**Small models (Haiku-class) can:**
- Extract entities, dates, numbers
- Classify article type from a fixed list
- Detect presence/absence of keywords
- Compare structured data for changes

**Small models should NOT:**
- Assess significance
- Make predictions
- Reason about implications
- Handle ambiguity

**Information flows one direction:**
```
Small → Medium → Large → Output
         ↓
       (never back)
```

If large model output feeds back to small models, the small models will amplify errors or hallucinations.

---

## Workflow Stages

### Stage 1: Collection (No LLM)

```
┌─────────────────────────────────────────────────────────────────┐
│                     STAGE 1: COLLECTION                         │
│                        (No LLM)                                 │
├─────────────────────────────────────────────────────────────────┤
│ • Fetch all RSS feeds                                           │
│ • Store raw articles with: URL, title, body, date, source       │
│ • Deduplicate by URL                                            │
│ • Filter: only articles from past 7 days                        │
│ • Output: article records in database                           │
└─────────────────────────────────────────────────────────────────┘
```

**Purpose:** Archive raw data before any AI processing. This ensures we always have the source of truth.

**Output:** Article records with fields:
- `article_id` (generated)
- `url`
- `title`
- `body`
- `published_date`
- `source_domain`
- `fetched_at`

---

### Stage 2: Extraction (Haiku)

```
┌─────────────────────────────────────────────────────────────────┐
│                   STAGE 2: EXTRACTION                           │
│                      (Haiku × N articles)                       │
├─────────────────────────────────────────────────────────────────┤
│ Per article:                                                    │
│ • Extract: securities mentioned, event type, entities,          │
│   numbers with context, direct quotes, dates                    │
│ • Classify: earnings, legal, leadership, product, market,       │
│   regulatory, partnership, other                                │
│ • Output: structured fact record linked to source article_id    │
│                                                                 │
│ Validation:                                                     │
│ • JSON schema check                                             │
│ • Reject if security attribution is ambiguous                   │
│ • Retry up to 2x on failure, then skip with warning             │
└─────────────────────────────────────────────────────────────────┘
```

**Purpose:** Convert unstructured articles into structured, queryable facts.

**Output schema:**
```json
{
  "article_id": "string",
  "securities": ["AAPL", "MSFT"],
  "event_type": "earnings | legal | leadership | product | market | regulatory | partnership | other",
  "entities": ["Apple Inc", "Tim Cook", "iPhone"],
  "numbers": [
    {
      "value": 2000,
      "unit": "employees",
      "context": "layoffs announced"
    }
  ],
  "direct_quotes": [
    {
      "speaker": "Tim Cook",
      "quote": "We are restructuring..."
    }
  ],
  "dates_mentioned": ["2024-01-15"],
  "sentiment": "positive | negative | neutral | mixed"
}
```

**Validation rules:**
- Must have at least one security
- Event type must be from allowed list
- JSON must parse correctly

---

### Stage 3: Clustering & Deduplication (Haiku or Algorithmic)

```
┌─────────────────────────────────────────────────────────────────┐
│                STAGE 3: CLUSTERING & DEDUP                      │
│                     (Haiku or algorithmic)                      │
├─────────────────────────────────────────────────────────────────┤
│ • Group facts by security                                       │
│ • Identify same event reported by multiple sources              │
│ • Merge duplicates, preserve all source_ids                     │
│ • Flag: events with only 1 source vs 2+ sources                 │
│ • Output: clustered fact records with source count              │
└─────────────────────────────────────────────────────────────────┘
```

**Purpose:** Identify when multiple articles report the same event. Multi-source events have higher credibility.

**Output additions:**
```json
{
  "cluster_id": "string",
  "source_article_ids": ["article_001", "article_002", "article_003"],
  "source_count": 3,
  "source_domains": ["reuters.com", "wsj.com", "bloomberg.com"]
}
```

**Rules:**
- Same security + same event type + overlapping dates/entities = likely same event
- Preserve all source references when merging
- Single-source events get flagged for extra scrutiny later

---

### Stage 4: Change Detection (Haiku)

```
┌─────────────────────────────────────────────────────────────────┐
│                  STAGE 4: CHANGE DETECTION                      │
│                        (Haiku)                                  │
├─────────────────────────────────────────────────────────────────┤
│ Per security:                                                   │
│ • Load last week's fact baseline                                │
│ • Compare: new events, changed situations, resolved issues      │
│ • Output: delta record (what changed, what's new)               │
│ • Store current facts as next week's baseline                   │
│                                                                 │
│ No change = explicitly record "no significant news"             │
└─────────────────────────────────────────────────────────────────┘
```

**Purpose:** Identify what's actually new this week vs. ongoing situations.

**Output:**
```json
{
  "security": "AAPL",
  "period": "2024-01-08 to 2024-01-15",
  "new_events": [...],
  "ongoing_situations": [...],
  "resolved_situations": [...],
  "no_news": false
}
```

**Baseline management:**
- After each run, current facts become next week's baseline
- Enables "this is new" vs. "this was already known" distinction

---

### Stage 5: Significance Assessment (Sonnet × 3)

```
┌─────────────────────────────────────────────────────────────────┐
│              STAGE 5: SIGNIFICANCE ASSESSMENT                   │
│               (Sonnet × 3 independent passes)                   │
├─────────────────────────────────────────────────────────────────┤
│ Three independent assessors evaluate all deltas:                │
│                                                                 │
│ Assessor A: "What could negatively impact this security?"       │
│ Assessor B: "What could positively impact this security?"       │
│ Assessor C: "What would a skeptical analyst flag for review?"   │
│                                                                 │
│ Each outputs:                                                   │
│ • Significance score (1-5)                                      │
│ • Reasoning (must cite source_ids)                              │
│ • Confidence level                                              │
│ • What information is missing                                   │
│                                                                 │
│ Consensus rule:                                                 │
│ • Score ≥4 from 2+ assessors → "high significance"              │
│ • Score ≥3 from 2+ assessors → "moderate significance"          │
│ • Otherwise → "low/no significance"                             │
└─────────────────────────────────────────────────────────────────┘
```

**Purpose:** Multiple perspectives reduce hallucinated significance. Consensus requirement catches outlier assessments.

**Why three different prompts:**
- Assessor A looks for risks (bearish lens)
- Assessor B looks for opportunities (bullish lens)
- Assessor C is skeptical of both (neutral lens)

If all three agree something is significant, it probably is. If only one flags it, it's likely noise or hallucination.

**Output per assessor:**
```json
{
  "security": "AAPL",
  "assessor": "A",
  "score": 4,
  "reasoning": "Workforce reduction of 2000 in cloud division could signal margin pressure. Similar pattern preceded 2023 earnings miss. [reuters_001, wsj_042]",
  "confidence": "medium",
  "missing_info": ["Severance costs not disclosed", "No guidance update yet"]
}
```

---

### Stage 6: External Corroboration (Sonnet with Web Search)

```
┌─────────────────────────────────────────────────────────────────┐
│              STAGE 6: EXTERNAL CORROBORATION                    │
│          (Sonnet with web search — high significance only)      │
├─────────────────────────────────────────────────────────────────┤
│ For each high/moderate significance item:                       │
│                                                                 │
│ • Extract core factual claims                                   │
│ • Web search: find corroborating sources                        │
│ • Web search: explicitly seek contradicting sources             │
│ • Check: are original RSS sources reputable?                    │
│                                                                 │
│ Output per item:                                                │
│ • Corroboration status: confirmed / contested / unverifiable    │
│ • Supporting sources found (with URLs)                          │
│ • Contradicting sources found (with URLs)                       │
│ • Source reliability assessment                                 │
│                                                                 │
│ Rule: If claim cannot be corroborated externally AND            │
│       only 1 original source → downgrade to "unverified"        │
└─────────────────────────────────────────────────────────────────┘
```

**Purpose:** Verify facts against sources outside the RSS feeds. Actively look for contradictions.

**Key behavior:** The prompt explicitly asks to find contradicting sources, not just confirming ones. This combats confirmation bias.

**Output:**
```json
{
  "cluster_id": "cluster_001",
  "core_claims": [
    "ACME laying off 2000 employees",
    "Effective Q2 2024",
    "Cloud division affected"
  ],
  "corroboration_status": "confirmed",
  "supporting_sources": [
    {"url": "https://reuters.com/...", "snippet": "..."},
    {"url": "https://company.com/press/...", "snippet": "..."}
  ],
  "contradicting_sources": [],
  "source_reliability": {
    "reuters.com": "high",
    "wsj.com": "high"
  }
}
```

**Downgrade rules:**
- Single original source + no external corroboration = "unverified"
- Contradicting sources found = "contested" (surface both sides)

---

### Stage 7: Adversarial Review (Sonnet)

```
┌─────────────────────────────────────────────────────────────────┐
│                 STAGE 7: ADVERSARIAL REVIEW                     │
│                        (Sonnet)                                 │
├─────────────────────────────────────────────────────────────────┤
│ For each item reaching this stage:                              │
│                                                                 │
│ Prompt: "You are a skeptical reviewer. Your job is to find      │
│ problems with this assessment. Challenge the conclusions.       │
│ What could be wrong? What alternative explanations exist?       │
│ What would need to be true for this to be a false alarm?"       │
│                                                                 │
│ Output:                                                         │
│ • Identified weaknesses                                         │
│ • Alternative interpretations                                   │
│ • Confidence adjustment recommendation                          │
│                                                                 │
│ If adversarial review finds critical flaw → flag for human      │
│ review before inclusion, don't auto-include                     │
└─────────────────────────────────────────────────────────────────┘
```

**Purpose:** Deliberately try to break the conclusion before presenting it to humans.

**Prompt design:** The adversarial reviewer is told its job is to find problems, not to agree. This creates productive tension.

**Output:**
```json
{
  "cluster_id": "cluster_001",
  "weaknesses": [
    "Layoff numbers not confirmed by company directly",
    "Cloud division headcount baseline unclear"
  ],
  "alternative_interpretations": [
    "Could be routine optimization post-hiring-spree, not distress signal",
    "May be reallocation rather than elimination"
  ],
  "confidence_adjustment": "reduce",
  "critical_flaw": false
}
```

**Escalation:** If `critical_flaw: true`, the item goes to "human review required" bucket rather than final synthesis.

---

### Stage 8: Final Synthesis (Opus)

```
┌─────────────────────────────────────────────────────────────────┐
│                  STAGE 8: FINAL SYNTHESIS                       │
│                         (Opus)                                  │
├─────────────────────────────────────────────────────────────────┤
│ Input: All verified items + adversarial notes + raw facts       │
│                                                                 │
│ Task:                                                           │
│ • Write weekly advisory for human consumption                   │
│ • Structure: per-security summaries + portfolio-wide themes     │
│ • Every claim must have inline citation [source_id]             │
│ • Explicit confidence levels per item                           │
│ • Section: "What we don't know / couldn't verify"               │
│ • Section: "Items flagged for human review"                     │
│                                                                 │
│ Tone: Analyst report, not hype. Uncertainty is stated clearly.  │
└─────────────────────────────────────────────────────────────────┘
```

**Purpose:** Combine all verified findings into a coherent, human-readable advisory.

**Why Opus:** This is the output humans will read and trust. The most capable model ensures coherent reasoning and proper integration of all the verification work done earlier.

**Required sections:**
1. Executive summary
2. Per-security findings (with citations)
3. Portfolio-wide themes/risks
4. What we couldn't verify
5. Items requiring human review
6. Methodology notes

---

### Stage 9: Storage (No LLM)

```
┌─────────────────────────────────────────────────────────────────┐
│                    STAGE 9: STORAGE                             │
│                       (No LLM)                                  │
├─────────────────────────────────────────────────────────────────┤
│ Store in Sentinel database:                                     │
│ • Final advisory (rendered for UI)                              │
│ • All intermediate artifacts (for audit trail)                  │
│ • Source articles referenced                                    │
│ • Processing metadata (models used, timestamps, versions)       │
└─────────────────────────────────────────────────────────────────┘
```

**Purpose:** Preserve everything for audit trail and debugging.

**Stored artifacts:**
- Raw articles
- Extracted facts
- Clustered facts
- Change deltas
- All three assessor outputs
- Corroboration reports
- Adversarial reviews
- Final synthesis
- Processing metadata (which models, when, prompt versions)

---

## Artifact Minimization Checkpoints

| Stage | Check | What Gets Caught |
|-------|-------|------------------|
| 2 | Schema validation | Malformed extraction |
| 3 | Source count | Single-source claims flagged |
| 5 | Multi-assessor consensus | Hallucinated significance |
| 6 | External corroboration | Fabricated or misread facts |
| 6 | Contradiction search | One-sided interpretation |
| 7 | Adversarial review | Flawed reasoning, overconfidence |
| 8 | Citation requirement | Claims without backing |

---

## Example Output Format

```markdown
# Weekly Advisory: January 8-15, 2024

## Executive Summary

Significant developments in 3 of 30 monitored securities this week.
One high-confidence alert (ACME workforce reduction), two moderate
items requiring monitoring.

---

## ACME Corp (ACME)

**Significance: High** | **Confidence: Medium** | **Sources: 3**

### What Happened
ACME announced workforce reduction of 2,000 employees in its
cloud division, effective Q2. [reuters_001, wsj_042]

### Why It Matters
This represents ~8% of cloud division headcount. Previous
reductions in 2023 preceded margin compression. [analyst note]

### Corroboration
- Confirmed by Reuters, WSJ, company press release
- No contradicting sources found

### Uncertainties
- Severance cost impact not yet disclosed
- Unclear if this affects product roadmap

### Adversarial Notes
- Could be routine optimization rather than distress signal
- Cloud division was overstaffed post-2021 hiring spree

---

## Beta Inc (BETA)

**Significance: Moderate** | **Confidence: Low** | **Sources: 1**

### What Happened
CFO departure reported. [smallcap_017]

### Corroboration
- Could not verify externally
- Original source (smallcapnews.com) has unknown reliability

### Recommendation
Verify manually before acting.

---

## Items Requiring Human Review

1. **BETA CFO departure** — Single unverified source
2. **GAMMA acquisition rumor** — Contradicting reports found

---

## No Significant News

The following securities had no significant developments:
AAPL, MSFT, GOOGL, ... (27 securities)

---

## Methodology

- Articles processed: 142
- Extraction failures: 3 (2.1%)
- Models used: Haiku (extraction), Sonnet (assessment), Opus (synthesis)
- Processing period: 2024-01-15 02:00 - 2024-01-15 08:30 UTC
```

---

## Failure Modes and Handling

| Failure | Detection | Response |
|---------|-----------|----------|
| RSS feed down | Collection stage | Log warning, proceed with available feeds |
| Extraction fails repeatedly | >2 retries | Skip article, include in "gaps" section |
| No consensus at Stage 5 | 3-way disagreement | Include as "mixed signals" for human review |
| Cannot corroborate | Stage 6 finds nothing | Downgrade confidence, flag for human |
| Adversarial finds critical flaw | Stage 7 | Don't auto-include, human must approve |
| Opus synthesis fails | Stage 8 | Retry once, then output raw verified items without synthesis |

---

## Open Questions

1. **RSS source quality** — Are these established financial news sources, or mixed quality? This affects how much we trust single-source items.

2. **Historical depth** — Should the system reference older history ("this is their third layoff announcement this year") or only week-over-week changes?

3. **Cross-security analysis** — Should it identify portfolio-wide risks ("3 of your holdings depend on Taiwan semiconductor supply")?

4. **Output format** — Markdown report? Structured JSON that Sentinel renders? Both?

---

## Key Anti-Hallucination Patterns Used

1. **Retrieval-Augmented Generation** — LLM never uses parametric knowledge for current events
2. **Structured output** — JSON schemas prevent freeform fabrication
3. **Multi-assessor consensus** — Single hallucinated judgment gets outvoted
4. **External verification** — Claims checked against sources outside the input set
5. **Adversarial review** — Explicit attempt to break conclusions before presenting
6. **Citation requirement** — Every claim must link to source_id
7. **Explicit uncertainty** — Required sections for unknowns and limitations
