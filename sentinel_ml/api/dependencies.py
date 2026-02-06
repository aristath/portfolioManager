"""FastAPI dependencies for Sentinel ML service."""

from dataclasses import dataclass

from sentinel_ml.clients.monolith_client import MonolithDataClient
from sentinel_ml.database.ml import MLDatabase


@dataclass
class CommonDependencies:
    ml_db: MLDatabase
    monolith: MonolithDataClient


async def get_common_deps() -> CommonDependencies:
    ml_db = MLDatabase()
    await ml_db.connect()
    return CommonDependencies(
        ml_db=ml_db,
        monolith=MonolithDataClient(),
    )
