import { computeCombinedPoint, normalizeWeights } from './mlWeights';

describe('normalizeWeights', () => {
  it('normalizes proportionally', () => {
    const normalized = normalizeWeights({
      wavelet: 0,
      svr: 0.8,
      rf: 0.3,
      ridge: 0.4,
      xgboost: 0.5,
    });
    expect(normalized.svr).toBeCloseTo(0.4, 5);
    expect(normalized.rf).toBeCloseTo(0.15, 5);
    expect(normalized.ridge).toBeCloseTo(0.2, 5);
    expect(normalized.xgboost).toBeCloseTo(0.25, 5);
  });

  it('returns null for zero-sum', () => {
    expect(normalizeWeights({ wavelet: 0, svr: 0, rf: 0, ridge: 0, xgboost: 0 })).toBeNull();
  });
});

describe('computeCombinedPoint', () => {
  it('handles missing components with availability-aware renormalization', () => {
    const normalized = normalizeWeights({ wavelet: 0.5, xgboost: 0.5, ridge: 0, rf: 0, svr: 0 });
    const point = { ml_xgboost: null };
    const past = { actual_ann_return: 10, wavelet_ann_return: 0.2 };
    // wavelet projected = 30, xgboost missing => uses wavelet only
    expect(computeCombinedPoint(point, past, normalized)).toBeCloseTo(30, 5);
  });
});

