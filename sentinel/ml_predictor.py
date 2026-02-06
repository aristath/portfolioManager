"""Production ML predictor with per-symbol models and wavelet score blending.

Uses 4 model types (xgboost, ridge, rf, svr) with per-model predictions
stored in separate tables via MLDatabase. Blending happens at query time.
"""

import json
import logging
import time
from typing import Any, Dict, Optional

import numpy as np

from sentinel.database import Database
from sentinel.database.ml import MODEL_TYPES, MLDatabase
from sentinel.ml_ensemble import EnsembleBlender
from sentinel.ml_features import (
    DEFAULT_FEATURES,
    features_to_array,
)
from sentinel.regime_quote import get_regime_adjusted_return
from sentinel.settings import Settings

logger = logging.getLogger(__name__)


class MLPredictor:
    """Production ML predictor with per-symbol model caching and per-model predictions."""

    def __init__(self, db=None, ml_db=None, settings=None):
        # Cache models per symbol: {symbol: EnsembleBlender}
        self._models: Dict[str, EnsembleBlender] = {}
        self._load_times: Dict[str, float] = {}
        self._cache_duration = 43200  # 12 hours

        self.db = db or Database()
        self.ml_db = ml_db or MLDatabase()
        self.settings = settings or Settings()

    async def predict_and_blend(
        self,
        symbol: str,
        date: str,
        wavelet_score: float,
        ml_enabled: bool,
        ml_blend_ratio: float,
        features: Optional[Dict[str, float]] = None,
        quote_data: Optional[Dict[str, Any]] = None,
        predicted_at_ts: Optional[int] = None,
        skip_cache: bool = False,
    ) -> Dict[str, Any]:
        """
        Predict return using all 4 ML models and return per-model results.

        Args:
            symbol: Security ticker
            date: Current date
            wavelet_score: Wavelet-based score (0-1)
            ml_enabled: Whether ML is enabled for this security
            ml_blend_ratio: Blend ratio (0.0 = pure wavelet, 1.0 = pure ML)
            features: Pre-computed features (optional, uses defaults if None)
            quote_data: Optional regime quote_data; if set, no DB load for regime
            predicted_at_ts: Optional unix ts for stored prediction (e.g. end-of-day for backfill)
            skip_cache: If True, do not read or write prediction cache (for backfill)

        Returns:
            {
                'predictions': {
                    'xgboost': {'predicted_return': float, 'ml_score': float, ...},
                    'ridge': {...}, 'rf': {...}, 'svr': {...},
                },
                'regime_score': float,
                'wavelet_score': float,
            }
        """
        await self.db.connect()

        if not ml_enabled:
            return self._fallback_to_wavelet(wavelet_score)

        # Check cache first (unless skip_cache e.g. for backfill)
        cache_key = f"ml:prediction:{symbol}:{date}"
        if not skip_cache:
            cached = await self.db.cache_get(cache_key)
            if cached is not None:
                logger.debug(f"{symbol}: Using cached ML prediction")
                return json.loads(cached)

        # Get model for this symbol
        ensemble = await self._get_model(symbol)

        if ensemble is None:
            return self._fallback_to_wavelet(wavelet_score)

        # Use pre-computed features or defaults
        if features is None:
            logger.debug(f"{symbol}: No features provided, using defaults")
            features = DEFAULT_FEATURES.copy()

        # Convert to array using centralized function (ensures correct order)
        feature_array = features_to_array(features).reshape(1, -1)

        # Predict using all 4 models (time this)
        start_time = time.time()
        try:
            per_model_raw = ensemble.predict(feature_array)  # dict[str, ndarray]
            inference_time_ms = (time.time() - start_time) * 1000

            if inference_time_ms > 100:
                logger.warning(f"{symbol}: Slow inference ({inference_time_ms:.1f}ms)")

        except Exception as e:
            logger.error(f"{symbol}: ML prediction failed: {e}")
            return self._fallback_to_wavelet(wavelet_score)

        # Apply regime-based dampening per model
        predictions = {}
        for mt in MODEL_TYPES:
            raw_return = float(per_model_raw[mt][0])

            adjusted_return, regime_score, dampening = await get_regime_adjusted_return(
                symbol, raw_return, self.db, quote_data=quote_data
            )

            ml_score = self._normalize_return_to_score(adjusted_return)

            predictions[mt] = {
                "predicted_return": adjusted_return,
                "raw_return": raw_return,
                "ml_score": ml_score,
                "regime_score": regime_score,
                "regime_dampening": dampening,
            }

            # Store prediction in per-model table
            ts = predicted_at_ts if predicted_at_ts is not None else int(time.time())
            prediction_id = f"{symbol}_{mt}_{ts}"

            await self._store_prediction(
                model_type=mt,
                prediction_id=prediction_id,
                symbol=symbol,
                features=features,
                predicted_return=adjusted_return,
                ml_score=ml_score,
                regime_score=regime_score,
                regime_dampening=dampening,
                inference_time_ms=inference_time_ms / len(MODEL_TYPES),
                predicted_at_ts=ts,
            )

        # Get regime_score from first model (all same since same quote_data)
        regime_score = predictions[MODEL_TYPES[0]]["regime_score"]

        # Compute blended final_score: weighted avg of per-model ml_scores, then blend with wavelet
        weights = {}
        for mt in MODEL_TYPES:
            weights[mt] = await self.settings.get(f"ml_weight_{mt}", 0.25)
        total_weight = sum(weights.values())
        if total_weight > 0:
            blended_ml_score = sum(predictions[mt]["ml_score"] * weights[mt] for mt in MODEL_TYPES) / total_weight
        else:
            blended_ml_score = wavelet_score
        final_score = (1 - ml_blend_ratio) * wavelet_score + ml_blend_ratio * blended_ml_score

        result = {
            "predictions": predictions,
            "final_score": float(final_score),
            "blended_ml_score": float(blended_ml_score),
            "regime_score": float(regime_score),
            "wavelet_score": float(wavelet_score),
        }

        if not skip_cache:
            await self.db.cache_set(cache_key, json.dumps(result), ttl_seconds=43200)
        return result

    def _normalize_return_to_score(self, predicted_return: float) -> float:
        """
        Normalize predicted return to 0-1 score.

        Mapping:
        - -10% return -> 0.0
        - 0% return -> 0.5
        - +10% return -> 1.0
        """
        score = 0.5 + (predicted_return * 5.0)
        return float(np.clip(score, 0.0, 1.0))

    async def _get_model(self, symbol: str) -> Optional[EnsembleBlender]:
        """Get model for a symbol, loading if needed."""
        current_time = time.time()

        if symbol in self._models:
            if current_time - self._load_times.get(symbol, 0) < self._cache_duration:
                return self._models[symbol]

        if not EnsembleBlender.model_exists(symbol):
            logger.debug(f"{symbol}: No trained model available")
            return None

        try:
            ensemble = EnsembleBlender()
            ensemble.load(symbol)

            self._models[symbol] = ensemble
            self._load_times[symbol] = current_time

            logger.info(f"{symbol}: ML model loaded")
            return ensemble

        except Exception as e:
            logger.error(f"{symbol}: Failed to load ML model: {e}")
            return None

    def _fallback_to_wavelet(self, wavelet_score: float) -> Dict[str, Any]:
        """Fallback to wavelet score if ML unavailable."""
        return {
            "predictions": {},
            "final_score": float(wavelet_score),
            "blended_ml_score": 0.0,
            "regime_score": 0.0,
            "wavelet_score": float(wavelet_score),
        }

    async def _store_prediction(
        self,
        model_type: str,
        prediction_id: str,
        symbol: str,
        features: dict,
        predicted_return: float,
        ml_score: float,
        regime_score: float,
        regime_dampening: float,
        inference_time_ms: float,
        predicted_at_ts: int,
    ):
        """Store prediction in per-model table via MLDatabase."""
        try:
            await self.ml_db.connect()
            await self.ml_db.store_prediction(
                model_type=model_type,
                prediction_id=prediction_id,
                symbol=symbol,
                predicted_at=predicted_at_ts,
                features=json.dumps(features),
                predicted_return=predicted_return,
                ml_score=ml_score,
                regime_score=regime_score,
                regime_dampening=regime_dampening,
                inference_time_ms=inference_time_ms,
            )
        except Exception as e:
            logger.error(f"Failed to store {model_type} prediction for {symbol}: {e}")

    def clear_cache(self, symbol: str | None = None):
        """Clear model cache for a symbol or all symbols."""
        if symbol:
            self._models.pop(symbol, None)
            self._load_times.pop(symbol, None)
        else:
            self._models.clear()
            self._load_times.clear()
