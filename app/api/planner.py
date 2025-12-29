"""Planner API endpoints for sequence regeneration."""

import logging

from fastapi import APIRouter, HTTPException

from app.domain.portfolio_hash import generate_portfolio_hash
from app.infrastructure.dependencies import PositionRepositoryDep, StockRepositoryDep
from app.repositories.planner_repository import PlannerRepository

logger = logging.getLogger(__name__)
router = APIRouter()


@router.post("/regenerate-sequences")
async def regenerate_sequences(
    position_repo: PositionRepositoryDep,
    stock_repo: StockRepositoryDep,
):
    """
    Regenerate sequences for current portfolio with current settings.

    This:
    1. Deletes existing sequences (keeps evaluations)
    2. Clears best_result (since best sequence may no longer exist)
    3. Returns success status

    Next batch run will detect no sequences and regenerate them.
    """
    try:
        # Get current portfolio state
        positions = await position_repo.get_all()
        stocks = await stock_repo.get_all_active()

        # Generate portfolio hash
        position_dicts = [
            {"symbol": p.symbol, "quantity": p.quantity} for p in positions
        ]
        portfolio_hash = generate_portfolio_hash(position_dicts, stocks)

        # Delete sequences only (keep evaluations)
        planner_repo = PlannerRepository()
        await planner_repo.delete_sequences_only(portfolio_hash)

        # Clear best_result (since best sequence may no longer exist)
        db = await planner_repo._get_db()
        await db.execute(
            "DELETE FROM best_result WHERE portfolio_hash = ?", (portfolio_hash,)
        )
        await db.commit()

        logger.info(
            f"Regenerated sequences for portfolio {portfolio_hash[:8]} - sequences deleted, evaluations preserved"
        )

        return {
            "status": "success",
            "message": "Sequences regenerated. New sequences will be generated on next batch run.",
            "portfolio_hash": portfolio_hash[:8],
        }

    except Exception as e:
        logger.error(f"Error regenerating sequences: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=str(e))
