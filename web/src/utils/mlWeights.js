const WEIGHT_KEYS = ['wavelet', 'xgboost', 'ridge', 'rf', 'svr'];

export function normalizeWeights(rawWeights) {
  const safe = {};
  WEIGHT_KEYS.forEach((key) => {
    const value = Number(rawWeights?.[key] ?? 0);
    safe[key] = Number.isFinite(value) && value > 0 ? value : 0;
  });

  const total = Object.values(safe).reduce((sum, value) => sum + value, 0);
  if (total <= 0) {
    return null;
  }

  const normalized = {};
  WEIGHT_KEYS.forEach((key) => {
    normalized[key] = safe[key] / total;
  });
  return normalized;
}

export function computeCombinedPoint(point, pastPoint, normalizedWeights) {
  if (!normalizedWeights) return null;

  const waveletProjected =
    pastPoint?.actual_ann_return != null && pastPoint?.wavelet_ann_return != null
      ? pastPoint.actual_ann_return + pastPoint.wavelet_ann_return * 100
      : null;

  const components = {
    wavelet: waveletProjected,
    xgboost: point?.ml_xgboost ?? null,
    ridge: point?.ml_ridge ?? null,
    rf: point?.ml_rf ?? null,
    svr: point?.ml_svr ?? null,
  };

  let weighted = 0;
  let availableWeight = 0;
  Object.entries(components).forEach(([key, value]) => {
    if (value == null) return;
    const weight = Number(normalizedWeights[key] ?? 0);
    if (!Number.isFinite(weight) || weight <= 0) return;
    weighted += value * weight;
    availableWeight += weight;
  });

  if (availableWeight <= 0) return null;
  return weighted / availableWeight;
}

