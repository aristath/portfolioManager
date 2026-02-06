"""Jobs API routes for Sentinel ML scheduler."""

from datetime import datetime
from typing import Optional

from fastapi import APIRouter, Depends, HTTPException
from typing_extensions import Annotated

from sentinel_ml.api.dependencies import CommonDependencies, get_common_deps
from sentinel_ml.jobs import get_status, reschedule, run_now

router = APIRouter(prefix="/jobs", tags=["jobs"])

MARKET_TIMING_LABELS = {
    0: "Any time",
    1: "After market close",
    2: "During market open",
    3: "All markets closed",
}

_scheduler = None


def set_scheduler(scheduler):
    global _scheduler
    _scheduler = scheduler


@router.get("")
async def get_jobs() -> dict:
    return await get_status()


@router.post("/{job_type:path}/run")
async def run_job_endpoint(job_type: str) -> dict:
    result = await run_now(job_type)
    if result.get("status") == "failed" and "Unknown job type" in result.get("error", ""):
        raise HTTPException(status_code=404, detail=result["error"])
    return result


@router.post("/refresh-all")
async def refresh_all(
    deps: Annotated[CommonDependencies, Depends(get_common_deps)],
) -> dict:
    await deps.ml_db.conn.execute("UPDATE ml_job_schedules SET last_run = 0")
    await deps.ml_db.conn.commit()
    schedules = await deps.ml_db.get_job_schedules()
    for s in schedules:
        await reschedule(s["job_type"], deps.ml_db)
    return {"status": "ok", "message": "All jobs rescheduled"}


@router.get("/schedules")
async def get_job_schedules(deps: Annotated[CommonDependencies, Depends(get_common_deps)]) -> dict:
    schedules = await deps.ml_db.get_job_schedules()

    next_run_times = {}
    if _scheduler:
        for job in _scheduler.get_jobs():
            if job.next_run_time:
                next_run_times[job.id] = job.next_run_time.isoformat()

    result = []
    for s in schedules:
        job_type = s["job_type"]
        history = await deps.ml_db.get_job_history_for_type(job_type, limit=1)
        if history:
            last_run = datetime.fromtimestamp(history[0]["executed_at"]).isoformat()
            last_status = history[0]["status"]
        else:
            last_run = None
            last_status = None

        result.append(
            {
                "job_type": s["job_type"],
                "interval_minutes": s["interval_minutes"],
                "interval_market_open_minutes": s.get("interval_market_open_minutes"),
                "market_timing": s["market_timing"],
                "market_timing_label": MARKET_TIMING_LABELS.get(s["market_timing"], "Unknown"),
                "description": s.get("description"),
                "category": s.get("category"),
                "last_run": last_run,
                "last_status": last_status,
                "next_run": next_run_times.get(job_type),
            }
        )

    return {"schedules": result}


@router.put("/schedules/{job_type:path}")
async def update_job_schedule(
    job_type: str,
    data: dict,
    deps: Annotated[CommonDependencies, Depends(get_common_deps)],
) -> dict:
    existing = await deps.ml_db.get_job_schedule(job_type)
    if not existing:
        raise HTTPException(status_code=404, detail=f"Unknown job type: {job_type}")

    if "interval_minutes" in data:
        val = data["interval_minutes"]
        if not isinstance(val, int) or val < 1 or val > 10080:
            raise HTTPException(status_code=400, detail="interval_minutes must be between 1 and 10080")

    if "interval_market_open_minutes" in data:
        val = data["interval_market_open_minutes"]
        if val is not None and (not isinstance(val, int) or val < 1 or val > 10080):
            raise HTTPException(status_code=400, detail="interval_market_open_minutes must be between 1 and 10080")

    if "market_timing" in data:
        val = data["market_timing"]
        if not isinstance(val, int) or val < 0 or val > 3:
            raise HTTPException(status_code=400, detail="market_timing must be 0, 1, 2, or 3")

    await deps.ml_db.upsert_job_schedule(
        job_type,
        interval_minutes=data.get("interval_minutes"),
        interval_market_open_minutes=data.get("interval_market_open_minutes"),
        market_timing=data.get("market_timing"),
    )

    await reschedule(job_type, deps.ml_db)

    return {"status": "ok"}


@router.get("/history")
async def get_job_history(
    deps: Annotated[CommonDependencies, Depends(get_common_deps)],
    job_type: Optional[str] = None,
    limit: int = 50,
) -> dict:
    if job_type:
        history = await deps.ml_db.get_job_history_for_type(job_type, limit=limit)
    else:
        history = await deps.ml_db.get_job_history(limit=limit)

    return {"history": history}
