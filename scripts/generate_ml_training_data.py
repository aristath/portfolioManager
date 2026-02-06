"""Generate initial ML training data."""

import asyncio
import sys
from pathlib import Path

# Add parent directory to path
sys.path.insert(0, str(Path(__file__).parent.parent))

from sentinel.database import Database
from sentinel_ml.database.ml import MLDatabase
from sentinel_ml.ml_trainer import TrainingDataGenerator


async def main():
    db = Database()
    await db.connect()
    ml_db = MLDatabase()
    await ml_db.connect()

    generator = TrainingDataGenerator(db=db, ml_db=ml_db)

    print("=" * 70)
    print("ML Training Data Generation")
    print("=" * 70)
    print("\nThis will generate training samples from historical data.")
    print("Training data stored in: data/ml.db")
    print("\nPress Ctrl+C to cancel\n")

    try:
        df = await generator.generate_training_data(
            prediction_horizon_days=14,
        )

        print("\n" + "=" * 70)
        print("Training Data Generation Complete!")
        print("=" * 70)
        print(f"Total samples: {len(df)}")
        print(f"Symbols: {df['symbol'].nunique()}")
        print("\nSample distribution by symbol:")
        print(df.groupby("symbol").size().describe())
        print("\nReturn distribution:")
        print(df["future_return"].describe())

    except KeyboardInterrupt:
        print("\n\nCancelled by user")
    except Exception as e:
        print(f"\n\nError: {e}")
        import traceback

        traceback.print_exc()
    finally:
        await ml_db.close()
        await db.close()


if __name__ == "__main__":
    asyncio.run(main())
