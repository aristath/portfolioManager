"""Train initial ML models — 4 models per symbol (XGBoost, Ridge, RF, SVR)."""

import asyncio
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from sentinel.database import Database
from sentinel.settings import Settings
from sentinel_ml.database.ml import MODEL_TYPES, MLDatabase
from sentinel_ml.ml_ensemble import EnsembleBlender


async def main():
    print("=" * 70)
    print("Per-Symbol ML Model Training (XGBoost + Ridge + RF + SVR)")
    print("=" * 70)

    db = Database()
    await db.connect()
    ml_db = MLDatabase()
    await ml_db.connect()
    settings = Settings()

    # Get minimum samples requirement
    min_samples = await settings.get("ml_min_samples_per_symbol", 100)
    print(f"\nMinimum samples per symbol: {min_samples}")

    # Get symbols with sufficient data from ml_db
    symbols_data = await ml_db.get_symbols_with_sufficient_data(min_samples)

    if not symbols_data:
        print("\nERROR: No symbols with sufficient training samples!")
        print("Please run generate_ml_training_data.py first.")
        return

    symbols = list(symbols_data.items())
    print(f"\nFound {len(symbols)} symbols with sufficient data:")
    for symbol, count in symbols[:10]:
        print(f"  {symbol}: {count} samples")
    if len(symbols) > 10:
        print(f"  ... and {len(symbols) - 10} more")

    # Train model for each symbol
    print("\n" + "=" * 70)
    print("Training Models")
    print("=" * 70)

    trained = 0
    failed = 0

    for i, (symbol, sample_count) in enumerate(symbols):
        print(f"\n[{i + 1}/{len(symbols)}] {symbol} ({sample_count} samples)")

        X, y = await ml_db.load_training_data(symbol)

        if len(X) == 0:
            print("  SKIP: No valid training data")
            failed += 1
            continue

        print(f"  Features: {X.shape}, Labels: {y.shape}")
        print(f"  Return stats: mean={y.mean():.4f}, std={y.std():.4f}")

        try:
            # Train ensemble (4 models)
            ensemble = EnsembleBlender()
            metrics = ensemble.train(X, y, validation_split=0.2)

            # Save model files
            ensemble.save(symbol)

            # Print per-model metrics
            for mt in MODEL_TYPES:
                mt_metrics = metrics.get(f"{mt}_metrics", {})
                mae = mt_metrics.get("val_mae", 0)
                r2 = mt_metrics.get("val_r2", 0)
                print(f"  {mt:8s}: MAE={mae:.4f}, R²={r2:.4f}")

                # Register in ml_db
                await ml_db.update_model_record(
                    model_type=mt,
                    symbol=symbol,
                    training_samples=len(X),
                    metrics={
                        "validation_rmse": mt_metrics.get("val_rmse", 0),
                        "validation_mae": mae,
                        "validation_r2": r2,
                    },
                )

            trained += 1

        except Exception as e:
            print(f"  ERROR: {e}")
            failed += 1
            continue

    print("\n" + "=" * 70)
    print("Training Complete!")
    print("=" * 70)
    print(f"\nModels trained: {trained}")
    print(f"Models failed: {failed}")
    print("\nModels saved to: data/ml_models/<symbol>/")

    # Show summary
    all_status = await ml_db.get_all_model_status()
    for mt in MODEL_TYPES:
        count = len(all_status.get(mt, []))
        print(f"  {mt}: {count} models")

    await ml_db.close()
    await db.close()


if __name__ == "__main__":
    asyncio.run(main())
