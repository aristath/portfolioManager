"""Pytest configuration and fixtures."""

import pytest
import aiosqlite
import tempfile
import os
from pathlib import Path

from app.database import init_db, SCHEMA
from app.infrastructure.database.repositories import (
    SQLiteStockRepository,
    SQLitePositionRepository,
    SQLitePortfolioRepository,
    SQLiteAllocationRepository,
    SQLiteScoreRepository,
    SQLiteTradeRepository,
)


@pytest.fixture
async def db():
    """Create a temporary in-memory database for testing."""
    # Create temporary database file
    fd, db_path = tempfile.mkstemp(suffix='.db')
    os.close(fd)

    try:
        # Initialize database
        async with aiosqlite.connect(db_path) as db:
            db.row_factory = aiosqlite.Row
            await db.executescript(SCHEMA)
            await db.commit()
            yield db
    finally:
        # Clean up
        if os.path.exists(db_path):
            os.unlink(db_path)


@pytest.fixture
async def stock_repo(db):
    """Create a stock repository instance."""
    return SQLiteStockRepository(db)


@pytest.fixture
async def position_repo(db):
    """Create a position repository instance."""
    return SQLitePositionRepository(db)


@pytest.fixture
async def portfolio_repo(db):
    """Create a portfolio repository instance."""
    return SQLitePortfolioRepository(db)


@pytest.fixture
async def allocation_repo(db):
    """Create an allocation repository instance."""
    return SQLiteAllocationRepository(db)


@pytest.fixture
async def score_repo(db):
    """Create a score repository instance."""
    return SQLiteScoreRepository(db)


@pytest.fixture
async def trade_repo(db):
    """Create a trade repository instance."""
    return SQLiteTradeRepository(db)
