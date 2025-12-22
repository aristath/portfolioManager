"""Settings API endpoints."""

import aiosqlite
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

from app.config import settings

router = APIRouter()


class MinTradeSizeUpdate(BaseModel):
    value: float


async def get_setting(key: str, default: str = None) -> str | None:
    """Get a setting value from the database."""
    async with aiosqlite.connect(settings.database_path) as db:
        cursor = await db.execute(
            "SELECT value FROM settings WHERE key = ?",
            (key,)
        )
        row = await cursor.fetchone()
        return row[0] if row else default


async def set_setting(key: str, value: str) -> None:
    """Set a setting value in the database."""
    async with aiosqlite.connect(settings.database_path) as db:
        await db.execute(
            """
            INSERT INTO settings (key, value) VALUES (?, ?)
            ON CONFLICT(key) DO UPDATE SET value = excluded.value
            """,
            (key, value)
        )
        await db.commit()


async def get_min_trade_size() -> float:
    """Get the minimum trade size, checking DB first then falling back to config."""
    db_value = await get_setting("min_trade_size")
    if db_value:
        return float(db_value)
    return settings.min_trade_size


@router.get("")
async def get_settings():
    """Get all configurable settings."""
    min_trade_size = await get_setting("min_trade_size")

    return {
        "min_trade_size": float(min_trade_size) if min_trade_size else settings.min_trade_size,
    }


@router.put("/min_trade_size")
async def update_min_trade_size(data: MinTradeSizeUpdate):
    """Update the minimum trade size setting."""
    if data.value <= 0:
        raise HTTPException(status_code=400, detail="Value must be positive")

    await set_setting("min_trade_size", str(data.value))

    return {"min_trade_size": data.value}


@router.post("/restart")
async def restart_system():
    """Trigger system reboot."""
    import subprocess
    subprocess.Popen(["sudo", "reboot"])
    return {"status": "rebooting"}
