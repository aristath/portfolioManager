"""APScheduler-based ML job runner."""

from __future__ import annotations

import asyncio
import logging
from datetime import datetime
from typing import Any, Callable

from apscheduler.executors.asyncio import AsyncIOExecutor
from apscheduler.jobstores.memory import MemoryJobStore
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.interval import IntervalTrigger

from sentinel_ml.jobs import tasks

logger = logging.getLogger(__name__)

_scheduler: AsyncIOScheduler | None = None
_deps: dict[str, Any] = {}
_current_job: str | None = None

JOB_TIMEOUT = 15 * 60

TASK_REGISTRY: dict[str, tuple[Callable, list[str]]] = {
    "analytics:regime": (tasks.analytics_regime, ["detector", "monolith"]),
    "ml:retrain": (tasks.ml_retrain, ["monolith", "retrainer"]),
    "ml:monitor": (tasks.ml_monitor, ["monolith", "monitor"]),
}


async def init(ml_db, monolith, detector, retrainer, monitor) -> AsyncIOScheduler:
    global _scheduler, _deps, _current_job

    _deps = {
        "ml_db": ml_db,
        "monolith": monolith,
        "detector": detector,
        "retrainer": retrainer,
        "monitor": monitor,
    }
    _current_job = None

    await ml_db.seed_default_job_schedules()

    _scheduler = AsyncIOScheduler(
        jobstores={"default": MemoryJobStore()},
        executors={"default": AsyncIOExecutor()},
        job_defaults={"coalesce": True, "max_instances": 1, "misfire_grace_time": 60},
    )

    schedules = await ml_db.get_job_schedules()
    schedule_map = {s["job_type"]: s for s in schedules}

    for job_type in TASK_REGISTRY:
        schedule = schedule_map.get(job_type) or {"job_type": job_type, "interval_minutes": 60}
        _add_job(job_type, schedule)

    _scheduler.start()
    logger.info("ML APScheduler started with %d jobs", len(TASK_REGISTRY))
    return _scheduler


async def stop() -> None:
    global _scheduler, _current_job
    if _scheduler:
        _scheduler.shutdown(wait=False)
        _scheduler = None
    _current_job = None


async def reschedule(job_type: str, ml_db) -> None:
    global _scheduler
    if not _scheduler:
        return

    schedule = await ml_db.get_job_schedule(job_type)
    if not schedule:
        return

    try:
        _scheduler.reschedule_job(job_type, trigger=IntervalTrigger(minutes=schedule["interval_minutes"]))
    except Exception as exc:  # noqa: BLE001
        logger.error("Failed to reschedule %s: %s", job_type, exc)


async def run_now(job_type: str) -> dict:
    if job_type not in TASK_REGISTRY:
        return {"status": "failed", "error": f"Unknown job type: {job_type}", "duration_ms": 0}

    ml_db = _deps.get("ml_db")
    schedule = await ml_db.get_job_schedule(job_type) if ml_db else {"job_type": job_type, "interval_minutes": 60}

    start = datetime.now()
    try:
        await _run_task(job_type, schedule)
        duration_ms = int((datetime.now() - start).total_seconds() * 1000)
        return {"status": "completed", "duration_ms": duration_ms}
    except Exception as exc:  # noqa: BLE001
        duration_ms = int((datetime.now() - start).total_seconds() * 1000)
        return {"status": "failed", "error": str(exc), "duration_ms": duration_ms}


async def get_status() -> dict:
    global _scheduler, _current_job

    status = {"current": _current_job, "upcoming": [], "recent": []}
    if _scheduler:
        for job in _scheduler.get_jobs():
            if job.next_run_time:
                status["upcoming"].append({"job_type": job.id, "next_run": job.next_run_time.isoformat()})
        status["upcoming"] = sorted(status["upcoming"], key=lambda x: x["next_run"])[:3]

    ml_db = _deps.get("ml_db")
    if ml_db:
        history = await ml_db.get_job_history(limit=20)
        seen_types = set()
        recent = []
        for entry in history:
            jt = entry["job_type"]
            if jt in seen_types:
                continue
            seen_types.add(jt)
            recent.append(
                {
                    "job_type": jt,
                    "status": entry["status"],
                    "executed_at": datetime.fromtimestamp(entry["executed_at"]).isoformat(),
                }
            )
            if len(recent) >= 3:
                break
        status["recent"] = recent

    return status


def _add_job(job_type: str, schedule: dict) -> None:
    global _scheduler
    if not _scheduler:
        return

    _scheduler.add_job(
        _job_executor,
        trigger=IntervalTrigger(minutes=schedule.get("interval_minutes", 60)),
        id=job_type,
        name=job_type,
        args=[job_type, schedule],
        replace_existing=True,
    )


async def _job_executor(job_type: str, schedule: dict) -> None:
    await _run_task(job_type, schedule)


async def _run_task(job_type: str, schedule: dict) -> dict | None:
    global _current_job

    if job_type not in TASK_REGISTRY:
        return {"skipped": True, "reason": "unknown_job_type"}

    ml_db = _deps.get("ml_db")
    task_func, dep_keys = TASK_REGISTRY[job_type]
    args = []
    for key in dep_keys:
        dep = _deps.get(key)
        if dep is None:
            logger.error("Missing dependency %s for job %s", key, job_type)
            return None
        args.append(dep)

    _current_job = job_type
    start = datetime.now()
    try:
        await asyncio.wait_for(task_func(*args), timeout=JOB_TIMEOUT)
        duration_ms = int((datetime.now() - start).total_seconds() * 1000)
        if ml_db:
            await ml_db.mark_job_completed(job_type)
            await ml_db.log_job_execution(job_type, job_type, "completed", None, duration_ms, 0)
        return {"status": "completed"}
    except TimeoutError:
        duration_ms = int((datetime.now() - start).total_seconds() * 1000)
        msg = f"Job {job_type} timed out after {JOB_TIMEOUT}s"
        if ml_db:
            await ml_db.mark_job_failed(job_type)
            await ml_db.log_job_execution(job_type, job_type, "failed", msg, duration_ms, 0)
        raise
    except Exception as exc:  # noqa: BLE001
        duration_ms = int((datetime.now() - start).total_seconds() * 1000)
        if ml_db:
            await ml_db.mark_job_failed(job_type)
            await ml_db.log_job_execution(job_type, job_type, "failed", str(exc), duration_ms, 0)
        raise
    finally:
        _current_job = None
