"""Analytics API routes for regime information."""

from fastapi import APIRouter, Depends
from typing_extensions import Annotated

from sentinel_ml.api.dependencies import CommonDependencies, get_common_deps
from sentinel_ml.regime_hmm import RegimeDetector

router = APIRouter(prefix="/analytics", tags=["analytics"])


@router.get("/regime/{symbol}")
async def get_regime_status(symbol: str) -> dict:
    detector = RegimeDetector()
    regime = await detector.detect_current_regime(symbol)
    history = await detector.get_regime_history(symbol, days=90)
    return {"current": regime, "history": history}


@router.get("/regimes")
async def get_all_regimes(deps: Annotated[CommonDependencies, Depends(get_common_deps)]) -> dict:
    detector = RegimeDetector()

    securities = await deps.monolith.get_securities(active_only=True, ml_enabled_only=False)
    results = {}
    for sec in securities:
        regime = await detector.detect_current_regime(sec["symbol"])
        results[sec["symbol"]] = regime
    return results
