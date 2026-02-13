"""Build 40-part score arrays for the NeoPixel heatmap.

The MCU is responsible for animation and visualization. The MPU (Python)
only provides two 40-element arrays:
  - before scores
  - after scores

These arrays are sorted so index corresponds to a percentile in the score
distribution, which makes diffs meaningful without needing stable identities.
"""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class SecurityScore:
    symbol: str
    weight: float
    score: float


def _largest_remainder_counts(weights: list[float], total_parts: int) -> list[int]:
    if total_parts <= 0:
        return [0 for _ in weights]

    total_w = sum(w for w in weights if w > 0)
    if total_w <= 0:
        return [0 for _ in weights]

    scaled = [max(0.0, w) / total_w * total_parts for w in weights]
    floors = [int(x) for x in scaled]
    remainder = total_parts - sum(floors)
    fracs = sorted(((scaled[i] - floors[i], i) for i in range(len(weights))), reverse=True)

    counts = floors[:]
    for _, i in fracs[:remainder]:
        counts[i] += 1
    return counts


def build_sorted_parts(scores: list[SecurityScore], *, total_parts: int = 40) -> list[float]:
    """Return a sorted list of length total_parts with repeated per-security scores."""
    if total_parts <= 0:
        return []
    if not scores:
        return [0.0] * total_parts

    weights = [max(0.0, float(s.weight)) for s in scores]
    counts = _largest_remainder_counts(weights, total_parts)

    parts: list[float] = []
    for s, n in zip(scores, counts, strict=False):
        parts.extend([float(s.score)] * int(n))

    # Ensure exact length (guard against any floating rounding oddities)
    if len(parts) < total_parts:
        parts.extend([0.0] * (total_parts - len(parts)))
    elif len(parts) > total_parts:
        parts = parts[:total_parts]

    parts.sort()
    return parts


def clamp_score(score: float, *, clamp_abs: float = 0.5) -> float:
    """Clamp score to [-clamp_abs, +clamp_abs]."""
    lo = -abs(clamp_abs)
    hi = abs(clamp_abs)
    if score < lo:
        return lo
    if score > hi:
        return hi
    return score
