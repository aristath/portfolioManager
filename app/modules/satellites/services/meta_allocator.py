"""Meta-allocator service - adjusts satellite allocations based on performance.

Implements performance-based allocation rebalancing:
- Evaluates satellites quarterly (every 3 months)
- Calculates performance scores (Sharpe, Sortino, win rate, profit factor)
- Reallocates budget based on relative performance
- Applies dampening to avoid excessive churn
- Enforces min/max allocation constraints
"""

import logging
from dataclasses import dataclass
from datetime import datetime

from app.modules.satellites.services.bucket_service import BucketService
from app.modules.satellites.services.performance_metrics import (
    calculate_bucket_performance,
)
from app.repositories import SettingsRepository

logger = logging.getLogger(__name__)


@dataclass
class AllocationRecommendation:
    """Recommended allocation adjustment for a satellite."""

    bucket_id: str
    current_allocation_pct: float
    target_allocation_pct: float
    new_allocation_pct: float  # After dampening
    adjustment_pct: float  # Change amount
    performance_score: float
    reason: str


@dataclass
class ReallocationResult:
    """Result of meta-allocator rebalancing."""

    total_satellite_budget: float  # Total % allocated to satellites
    recommendations: list[AllocationRecommendation]
    dampening_factor: float
    evaluation_period_days: int
    satellites_evaluated: int
    satellites_improved: int
    satellites_reduced: int
    timestamp: str


class MetaAllocator:
    """Manages performance-based allocation rebalancing."""

    def __init__(self):
        self.bucket_service = BucketService()
        self.settings_repo = SettingsRepository()

    async def evaluate_and_reallocate(
        self,
        evaluation_months: int = 3,
        dampening_factor: float = 0.5,
        apply_changes: bool = False,
    ) -> ReallocationResult:
        """Evaluate satellite performance and recommend allocation changes.

        Args:
            evaluation_months: Months of history to evaluate (default 3 for quarterly)
            dampening_factor: How much to move toward target (0.5 = 50% of the way)
            apply_changes: Whether to actually apply the changes (default False for dry-run)

        Returns:
            ReallocationResult with recommendations
        """
        evaluation_days = evaluation_months * 30

        # Get all satellites (excluding core)
        all_buckets = await self.bucket_service.get_all_buckets()
        satellites = [b for b in all_buckets if b.id != "core"]

        if not satellites:
            logger.info("No satellites to evaluate")
            return ReallocationResult(
                total_satellite_budget=0.0,
                recommendations=[],
                dampening_factor=dampening_factor,
                evaluation_period_days=evaluation_days,
                satellites_evaluated=0,
                satellites_improved=0,
                satellites_reduced=0,
                timestamp=datetime.now().isoformat(),
            )

        # Get current satellite budget from settings
        satellite_budget_pct = await self.settings_repo.get_float(
            "satellite_budget_pct", 0.20
        )  # Default 20%

        # Calculate performance metrics for each satellite
        performance_scores = {}
        for satellite in satellites:
            try:
                metrics = await calculate_bucket_performance(
                    bucket_id=satellite.id,
                    period_days=evaluation_days,
                )

                if metrics:
                    performance_scores[satellite.id] = metrics.composite_score
                    logger.info(
                        f"{satellite.id}: Performance score = {metrics.composite_score:.3f} "
                        f"(Sharpe={metrics.sharpe_ratio:.2f}, "
                        f"Sortino={metrics.sortino_ratio:.2f}, "
                        f"Win rate={metrics.win_rate:.1f}%, "
                        f"Profit factor={metrics.profit_factor:.2f})"
                    )
                else:
                    # Insufficient data - use neutral score
                    performance_scores[satellite.id] = 0.0
                    logger.warning(
                        f"{satellite.id}: Insufficient data for evaluation, "
                        "using neutral score"
                    )

            except Exception as e:
                logger.error(
                    f"Error calculating performance for {satellite.id}: {e}",
                    exc_info=True,
                )
                performance_scores[satellite.id] = 0.0

        # Calculate new allocations based on relative performance
        recommendations = await self._calculate_allocations(
            satellites=satellites,
            performance_scores=performance_scores,
            total_budget_pct=satellite_budget_pct,
            dampening_factor=dampening_factor,
        )

        # Apply changes if requested
        improved = 0
        reduced = 0

        if apply_changes:
            for rec in recommendations:
                if rec.adjustment_pct > 0.01:  # Increased by >0.01%
                    improved += 1
                elif rec.adjustment_pct < -0.01:  # Decreased by >0.01%
                    reduced += 1

                # Update bucket target allocation
                await self.bucket_service.update_allocation(
                    bucket_id=rec.bucket_id,
                    target_allocation_pct=rec.new_allocation_pct,
                )

                logger.info(
                    f"{rec.bucket_id}: Allocation updated "
                    f"{rec.current_allocation_pct:.2%} → {rec.new_allocation_pct:.2%} "
                    f"({rec.adjustment_pct:+.2%})"
                )

        return ReallocationResult(
            total_satellite_budget=satellite_budget_pct,
            recommendations=recommendations,
            dampening_factor=dampening_factor,
            evaluation_period_days=evaluation_days,
            satellites_evaluated=len(satellites),
            satellites_improved=improved,
            satellites_reduced=reduced,
            timestamp=datetime.now().isoformat(),
        )

    async def _calculate_allocations(
        self,
        satellites: list,
        performance_scores: dict[str, float],
        total_budget_pct: float,
        dampening_factor: float,
    ) -> list[AllocationRecommendation]:
        """Calculate new allocations based on performance scores.

        Algorithm:
        1. Normalize scores to be non-negative (shift by min score)
        2. Allocate proportionally to normalized scores
        3. Apply min/max constraints
        4. Apply dampening to smooth changes

        Args:
            satellites: List of satellite buckets
            performance_scores: Dict of bucket_id -> score
            total_budget_pct: Total % to allocate to satellites
            dampening_factor: How much to move toward target (0-1)

        Returns:
            List of AllocationRecommendation objects
        """
        # Get constraints from settings
        min_satellite_pct = await self.settings_repo.get_float(
            "satellite_min_pct", 0.03
        )  # 3%
        max_satellite_pct = await self.settings_repo.get_float(
            "satellite_max_pct", 0.12
        )  # 12%

        # Normalize scores to be non-negative
        min_score = min(performance_scores.values())
        normalized_scores = {
            sat_id: score - min_score + 0.01  # Add small epsilon
            for sat_id, score in performance_scores.items()
        }

        total_score = sum(normalized_scores.values())

        recommendations = []

        for satellite in satellites:
            current_pct = satellite.target_allocation_pct
            sat_id = satellite.id
            score = normalized_scores.get(sat_id, 0.01)

            # Calculate raw target allocation (proportional to score)
            if total_score > 0:
                raw_target_pct = (score / total_score) * total_budget_pct
            else:
                # Equal allocation if no scores
                raw_target_pct = total_budget_pct / len(satellites)

            # Apply constraints
            constrained_target_pct = max(
                min_satellite_pct, min(max_satellite_pct, raw_target_pct)
            )

            # Apply dampening (move partway toward target)
            new_pct = (
                current_pct + (constrained_target_pct - current_pct) * dampening_factor
            )

            adjustment_pct = new_pct - current_pct

            # Determine reason
            if abs(adjustment_pct) < 0.005:  # <0.5% change
                reason = "No significant change"
            elif adjustment_pct > 0:
                if performance_scores[sat_id] > 0:
                    reason = f"Increased due to strong performance (score: {performance_scores[sat_id]:.2f})"
                else:
                    reason = "Increased to maintain minimum allocation"
            else:
                if performance_scores[sat_id] < 0:
                    reason = f"Reduced due to weak performance (score: {performance_scores[sat_id]:.2f})"
                else:
                    reason = "Reduced to enforce maximum allocation"

            recommendations.append(
                AllocationRecommendation(
                    bucket_id=sat_id,
                    current_allocation_pct=current_pct,
                    target_allocation_pct=constrained_target_pct,
                    new_allocation_pct=new_pct,
                    adjustment_pct=adjustment_pct,
                    performance_score=performance_scores[sat_id],
                    reason=reason,
                )
            )

        # Normalize to ensure total equals budget (due to constraints)
        total_new = sum(r.new_allocation_pct for r in recommendations)
        if total_new > 0 and abs(total_new - total_budget_pct) > 0.001:
            scale_factor = total_budget_pct / total_new
            for rec in recommendations:
                rec.new_allocation_pct *= scale_factor
                rec.adjustment_pct = rec.new_allocation_pct - rec.current_allocation_pct

        return recommendations

    async def preview_reallocation(
        self, evaluation_months: int = 3
    ) -> ReallocationResult:
        """Preview allocation changes without applying them.

        Args:
            evaluation_months: Months of history to evaluate

        Returns:
            ReallocationResult with recommendations (not applied)
        """
        return await self.evaluate_and_reallocate(
            evaluation_months=evaluation_months,
            apply_changes=False,
        )

    async def apply_reallocation(
        self, evaluation_months: int = 3, dampening_factor: float = 0.5
    ) -> ReallocationResult:
        """Evaluate and apply allocation changes.

        Args:
            evaluation_months: Months of history to evaluate
            dampening_factor: How much to move toward target

        Returns:
            ReallocationResult with applied changes
        """
        return await self.evaluate_and_reallocate(
            evaluation_months=evaluation_months,
            dampening_factor=dampening_factor,
            apply_changes=True,
        )
