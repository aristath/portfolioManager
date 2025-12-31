"""Bucket service - lifecycle management for buckets."""

import logging
from datetime import datetime
from typing import List, Optional

from app.modules.satellites.database.balance_repository import BalanceRepository
from app.modules.satellites.database.bucket_repository import BucketRepository
from app.modules.satellites.domain.enums import BucketStatus, BucketType
from app.modules.satellites.domain.models import Bucket, SatelliteSettings

logger = logging.getLogger(__name__)


class BucketService:
    """Service for bucket lifecycle management.

    Handles creation, updates, status transitions, and retirement
    of both core and satellite buckets.
    """

    def __init__(
        self,
        bucket_repo: Optional[BucketRepository] = None,
        balance_repo: Optional[BalanceRepository] = None,
    ):
        self._bucket_repo = bucket_repo or BucketRepository()
        self._balance_repo = balance_repo or BalanceRepository()

    # Query methods

    async def get_bucket(self, bucket_id: str) -> Optional[Bucket]:
        """Get a bucket by ID."""
        return await self._bucket_repo.get_by_id(bucket_id)

    async def get_all_buckets(self) -> List[Bucket]:
        """Get all buckets."""
        return await self._bucket_repo.get_all()

    async def get_active_buckets(self) -> List[Bucket]:
        """Get all active buckets (not retired or paused)."""
        return await self._bucket_repo.get_active()

    async def get_satellites(self) -> List[Bucket]:
        """Get all satellite buckets."""
        return await self._bucket_repo.get_satellites()

    async def get_core(self) -> Optional[Bucket]:
        """Get the core bucket."""
        return await self._bucket_repo.get_core()

    async def get_settings(self, satellite_id: str) -> Optional[SatelliteSettings]:
        """Get settings for a satellite."""
        return await self._bucket_repo.get_settings(satellite_id)

    # Lifecycle methods

    async def create_satellite(
        self,
        satellite_id: str,
        name: str,
        notes: Optional[str] = None,
        start_in_research: bool = True,
    ) -> Bucket:
        """Create a new satellite bucket.

        New satellites start in research mode by default, allowing
        paper trading until the user is ready to activate.

        Args:
            satellite_id: Unique identifier for the satellite
            name: Human-readable name
            notes: Optional documentation/description
            start_in_research: If True, start in research mode (default)

        Returns:
            The created bucket

        Raises:
            ValueError: If satellite_id already exists
        """
        existing = await self._bucket_repo.get_by_id(satellite_id)
        if existing:
            raise ValueError(f"Bucket with id '{satellite_id}' already exists")

        # Get allocation settings for defaults
        settings = await self._balance_repo.get_all_allocation_settings()
        min_pct = settings.get("satellite_min_pct", 0.02)
        max_pct = settings.get("satellite_max_pct", 0.15)

        bucket = Bucket(
            id=satellite_id,
            name=name,
            type=BucketType.SATELLITE,
            status=(
                BucketStatus.RESEARCH
                if start_in_research
                else BucketStatus.ACCUMULATING
            ),
            notes=notes,
            target_pct=0.0,  # Starts with no allocation
            min_pct=min_pct,
            max_pct=max_pct,
        )

        created = await self._bucket_repo.create(bucket)
        logger.info(
            f"Created satellite '{satellite_id}' in "
            f"{'research' if start_in_research else 'accumulating'} mode"
        )
        return created

    async def activate_satellite(self, satellite_id: str) -> Bucket:
        """Activate a satellite from research or accumulating mode.

        The satellite must have reached minimum allocation threshold
        to become fully active.

        Args:
            satellite_id: ID of the satellite to activate

        Returns:
            Updated bucket

        Raises:
            ValueError: If satellite doesn't exist or cannot be activated
        """
        bucket = await self._bucket_repo.get_by_id(satellite_id)
        if not bucket:
            raise ValueError(f"Satellite '{satellite_id}' not found")

        if bucket.type != BucketType.SATELLITE:
            raise ValueError("Cannot activate non-satellite bucket")

        if bucket.status not in (BucketStatus.RESEARCH, BucketStatus.ACCUMULATING):
            raise ValueError(
                f"Cannot activate satellite in '{bucket.status.value}' status"
            )

        updated = await self._bucket_repo.update_status(
            satellite_id, BucketStatus.ACTIVE
        )
        if not updated:
            raise RuntimeError(
                f"Failed to update status for satellite '{satellite_id}' - "
                "bucket disappeared during operation"
            )
        logger.info(f"Activated satellite '{satellite_id}'")
        return updated

    async def pause_bucket(self, bucket_id: str) -> Bucket:
        """Pause a bucket, stopping all trading.

        Args:
            bucket_id: ID of the bucket to pause

        Returns:
            Updated bucket

        Raises:
            ValueError: If bucket doesn't exist or is already paused/retired
        """
        bucket = await self._bucket_repo.get_by_id(bucket_id)
        if not bucket:
            raise ValueError(f"Bucket '{bucket_id}' not found")

        if bucket.status == BucketStatus.RETIRED:
            raise ValueError("Cannot pause a retired bucket")

        if bucket.status == BucketStatus.PAUSED:
            raise ValueError("Bucket is already paused")

        updated = await self._bucket_repo.update_status(bucket_id, BucketStatus.PAUSED)
        if not updated:
            raise RuntimeError(
                f"Failed to pause bucket '{bucket_id}' - "
                "bucket disappeared during operation"
            )
        logger.info(f"Paused bucket '{bucket_id}'")
        return updated

    async def resume_bucket(self, bucket_id: str) -> Bucket:
        """Resume a paused bucket.

        Returns the bucket to its previous active state (active or accumulating
        depending on allocation).

        Args:
            bucket_id: ID of the bucket to resume

        Returns:
            Updated bucket

        Raises:
            ValueError: If bucket doesn't exist or is not paused
        """
        bucket = await self._bucket_repo.get_by_id(bucket_id)
        if not bucket:
            raise ValueError(f"Bucket '{bucket_id}' not found")

        if bucket.status != BucketStatus.PAUSED:
            raise ValueError("Bucket is not paused")

        # Determine appropriate status based on allocation
        min_pct = bucket.min_pct or 0.02
        target_pct = bucket.target_pct or 0.0

        if target_pct >= min_pct:
            new_status = BucketStatus.ACTIVE
        else:
            new_status = BucketStatus.ACCUMULATING

        updated = await self._bucket_repo.update_status(bucket_id, new_status)
        if not updated:
            raise RuntimeError(
                f"Failed to resume bucket '{bucket_id}' - "
                "bucket disappeared during operation"
            )
        logger.info(f"Resumed bucket '{bucket_id}' to '{new_status.value}' status")
        return updated

    async def hibernate_bucket(self, bucket_id: str) -> Bucket:
        """Put a bucket into hibernation.

        Used when a satellite falls below minimum allocation or
        during safety circuit breaker triggers.

        Args:
            bucket_id: ID of the bucket to hibernate

        Returns:
            Updated bucket

        Raises:
            ValueError: If bucket doesn't exist or cannot be hibernated
        """
        bucket = await self._bucket_repo.get_by_id(bucket_id)
        if not bucket:
            raise ValueError(f"Bucket '{bucket_id}' not found")

        if bucket.id == "core":
            raise ValueError("Cannot hibernate core bucket")

        if bucket.status in (BucketStatus.RETIRED, BucketStatus.RESEARCH):
            raise ValueError(
                f"Cannot hibernate bucket in '{bucket.status.value}' status"
            )

        updated = await self._bucket_repo.update_status(
            bucket_id, BucketStatus.HIBERNATING
        )
        if not updated:
            raise RuntimeError(
                f"Failed to hibernate bucket '{bucket_id}' - "
                "bucket disappeared during operation"
            )
        logger.info(f"Hibernated bucket '{bucket_id}'")
        return updated

    async def retire_satellite(self, satellite_id: str) -> Bucket:
        """Retire a satellite permanently.

        Prerequisites:
        - Satellite must be paused first
        - All positions should be reassigned or liquidated
        - Cash should be transferred to other buckets

        The satellite's data is preserved for historical reporting.

        Args:
            satellite_id: ID of the satellite to retire

        Returns:
            Updated bucket

        Raises:
            ValueError: If satellite doesn't exist, is core, or not paused
        """
        bucket = await self._bucket_repo.get_by_id(satellite_id)
        if not bucket:
            raise ValueError(f"Satellite '{satellite_id}' not found")

        if bucket.type != BucketType.SATELLITE:
            raise ValueError("Cannot retire non-satellite bucket")

        if bucket.status != BucketStatus.PAUSED:
            raise ValueError(
                "Satellite must be paused before retirement. "
                "Please pause first and ensure all positions are handled."
            )

        # Check if satellite still has cash balances
        balances = await self._balance_repo.get_all_balances(satellite_id)
        total_balance = sum(b.balance for b in balances)
        if total_balance > 0.01:  # Allow small rounding errors
            raise ValueError(
                f"Satellite still has {total_balance:.2f} in cash. "
                "Please transfer funds before retiring."
            )

        updated = await self._bucket_repo.update_status(
            satellite_id, BucketStatus.RETIRED
        )
        if not updated:
            raise RuntimeError(
                f"Failed to retire satellite '{satellite_id}' - "
                "bucket disappeared during operation"
            )
        logger.info(f"Retired satellite '{satellite_id}'")
        return updated

    # Settings methods

    async def save_settings(self, settings: SatelliteSettings) -> SatelliteSettings:
        """Save or update satellite settings.

        Args:
            settings: Settings to save

        Returns:
            Saved settings

        Raises:
            ValueError: If satellite doesn't exist or validation fails
        """
        bucket = await self._bucket_repo.get_by_id(settings.satellite_id)
        if not bucket:
            raise ValueError(f"Satellite '{settings.satellite_id}' not found")

        if bucket.type != BucketType.SATELLITE:
            raise ValueError("Settings can only be saved for satellites")

        settings.validate()
        saved = await self._bucket_repo.save_settings(settings)
        logger.info(f"Saved settings for satellite '{settings.satellite_id}'")
        return saved

    # Circuit breaker methods

    async def record_trade_result(
        self, bucket_id: str, is_win: bool, trade_pnl: float
    ) -> Bucket:
        """Record a trade result for circuit breaker tracking.

        Args:
            bucket_id: ID of the bucket
            is_win: Whether the trade was profitable
            trade_pnl: Profit/loss amount

        Returns:
            Updated bucket
        """
        bucket = await self._bucket_repo.get_by_id(bucket_id)
        if not bucket:
            raise ValueError(f"Bucket '{bucket_id}' not found")

        if is_win:
            # Reset consecutive losses on any win
            await self._bucket_repo.reset_consecutive_losses(bucket_id)
            logger.info(f"Reset consecutive losses for '{bucket_id}' after win")
        else:
            # Increment consecutive losses
            new_count = await self._bucket_repo.increment_consecutive_losses(bucket_id)
            logger.info(
                f"Consecutive losses for '{bucket_id}': {new_count}/"
                f"{bucket.max_consecutive_losses}"
            )

            # Check circuit breaker
            if new_count >= bucket.max_consecutive_losses:
                await self._bucket_repo.update(
                    bucket_id,
                    status=BucketStatus.PAUSED,
                    loss_streak_paused_at=datetime.now().isoformat(),
                )
                logger.warning(
                    f"Circuit breaker triggered for '{bucket_id}' "
                    f"after {new_count} consecutive losses"
                )

        result = await self._bucket_repo.get_by_id(bucket_id)
        if not result:
            raise RuntimeError(
                f"Failed to retrieve bucket '{bucket_id}' after recording trade result - "
                "bucket disappeared during operation"
            )
        return result

    async def update_high_water_mark(
        self, bucket_id: str, current_value: float
    ) -> Bucket:
        """Update high water mark if current value exceeds it.

        Args:
            bucket_id: ID of the bucket
            current_value: Current total value of the bucket

        Returns:
            Updated bucket
        """
        bucket = await self._bucket_repo.get_by_id(bucket_id)
        if not bucket:
            raise ValueError(f"Bucket '{bucket_id}' not found")

        if current_value > bucket.high_water_mark:
            updated = await self._bucket_repo.update_high_water_mark(
                bucket_id, current_value
            )
            if not updated:
                raise RuntimeError(
                    f"Failed to update high water mark for '{bucket_id}' - "
                    "bucket disappeared during operation"
                )
            logger.info(
                f"Updated high water mark for '{bucket_id}' to {current_value:.2f}"
            )
            return updated

        return bucket

    # Update methods

    async def update_bucket(self, bucket_id: str, **updates) -> Optional[Bucket]:
        """Update bucket fields.

        Args:
            bucket_id: ID of the bucket to update
            **updates: Fields to update

        Returns:
            Updated bucket or None if not found
        """
        return await self._bucket_repo.update(bucket_id, **updates)

    async def update_allocation(
        self, bucket_id: str, target_pct: float
    ) -> Optional[Bucket]:
        """Update bucket target allocation.

        Args:
            bucket_id: ID of the bucket
            target_pct: New target percentage (0.0-1.0)

        Returns:
            Updated bucket
        """
        bucket = await self._bucket_repo.get_by_id(bucket_id)
        if not bucket:
            raise ValueError(f"Bucket '{bucket_id}' not found")

        # Validate against min/max
        if bucket.min_pct and target_pct < bucket.min_pct:
            logger.warning(
                f"Target {target_pct:.2%} below min {bucket.min_pct:.2%} "
                f"for '{bucket_id}'"
            )

        if bucket.max_pct and target_pct > bucket.max_pct:
            raise ValueError(
                f"Target {target_pct:.2%} exceeds max {bucket.max_pct:.2%}"
            )

        return await self._bucket_repo.update(bucket_id, target_pct=target_pct)
