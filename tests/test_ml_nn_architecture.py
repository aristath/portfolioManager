import numpy as np

from sentinel.ml_ensemble import EnsembleBlender, NeuralNetReturnPredictor
from sentinel.ml_features import NUM_FEATURES


def test_neural_net_predictor_is_bounded_by_max_return():
    predictor = NeuralNetReturnPredictor()
    predictor.build_model(input_dim=NUM_FEATURES, max_return=0.12)

    # Force huge activations to test bounding behavior.
    assert predictor.model is not None
    predictor.model.weights[0][:] = 1000.0
    predictor.model.weights[1][:] = 1000.0

    predictor.scaler = None
    # Bypass scaler by calling model.forward directly with large inputs.
    X = (np.ones((5, NUM_FEATURES), dtype=np.float32) * 1000.0).astype(np.float32)
    out = predictor.model.forward(X).flatten()

    assert np.all(np.isfinite(out))
    assert np.all(out <= 0.12 + 1e-6)
    assert np.all(out >= -0.12 - 1e-6)


def test_ensemble_predict_is_bounded_by_nn_max_return_when_xgb_zero():
    ensemble = EnsembleBlender(nn_weight=1.0, xgb_weight=0.0)
    ensemble.nn_predictor.build_model(input_dim=NUM_FEATURES, max_return=0.07)

    assert ensemble.nn_predictor.model is not None
    ensemble.nn_predictor.model.weights[0][:] = 1000.0
    ensemble.nn_predictor.model.weights[1][:] = 1000.0

    # Set scaler to identity transform by fitting on zeros.
    X_fit = np.zeros((10, NUM_FEATURES), dtype=np.float32)
    ensemble.nn_predictor.scaler = None

    # If scaler is missing, predict() would assert; so wire a trivial scaler.
    from sklearn.preprocessing import StandardScaler

    scaler = StandardScaler()
    scaler.fit(X_fit)
    ensemble.nn_predictor.scaler = scaler

    # XGB isn't trained/loaded; swap predictor with stub returning zeros.
    class _ZeroXGB:
        def predict(self, X):
            return np.zeros((X.shape[0],), dtype=np.float32)

    ensemble.xgb_predictor = _ZeroXGB()  # type: ignore[assignment]

    X = (np.ones((3, NUM_FEATURES), dtype=np.float32) * 1000.0).astype(np.float32)
    preds = ensemble.predict(X)

    assert np.all(preds <= 0.07 + 1e-6)
    assert np.all(preds >= -0.07 - 1e-6)
