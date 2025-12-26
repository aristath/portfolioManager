"""Tests for job scheduler.

These tests validate job scheduling, failure tracking,
and scheduler management.
"""

from datetime import datetime, timedelta
from unittest.mock import AsyncMock, MagicMock, patch

import pytest


class TestJobListener:
    """Test job event listener."""

    def test_tracks_job_failure(self):
        """Test that job failures are tracked."""
        from app.jobs.scheduler import _job_failures, job_listener

        # Clear any existing failures
        _job_failures.clear()

        # Create a mock event with an exception
        event = MagicMock()
        event.job_id = "test_job"
        event.exception = Exception("Test error")

        with patch("app.jobs.scheduler.settings") as mock_settings:
            mock_settings.job_failure_window_hours = 1
            mock_settings.job_failure_threshold = 3
            job_listener(event)

        assert "test_job" in _job_failures
        assert len(_job_failures["test_job"]) == 1

    def test_clears_failures_on_success(self):
        """Test that failures are cleared on successful execution."""
        from app.jobs.scheduler import _job_failures, job_listener

        # Add some failures
        _job_failures["test_job"] = [datetime.now()]

        # Create a mock event without exception
        event = MagicMock()
        event.job_id = "test_job"
        event.exception = None

        job_listener(event)

        assert len(_job_failures["test_job"]) == 0

    def test_removes_old_failures(self):
        """Test that old failures outside window are removed."""
        from app.jobs.scheduler import _job_failures, job_listener

        _job_failures.clear()

        # Add an old failure
        old_time = datetime.now() - timedelta(hours=2)
        _job_failures["test_job"] = [old_time]

        # Create new failure
        event = MagicMock()
        event.job_id = "test_job"
        event.exception = Exception("New error")

        with patch("app.jobs.scheduler.settings") as mock_settings:
            mock_settings.job_failure_window_hours = 1
            mock_settings.job_failure_threshold = 3
            job_listener(event)

        # Old failure should be removed, only new one remains
        assert len(_job_failures["test_job"]) == 1


class TestGetJobHealthStatus:
    """Test job health status retrieval."""

    def test_returns_job_status(self):
        """Test that job status is returned correctly."""
        from app.jobs.scheduler import get_job_health_status

        with patch("app.jobs.scheduler.scheduler") as mock_scheduler:
            mock_job = MagicMock()
            mock_job.id = "portfolio_sync"
            mock_job.next_run_time = datetime.now() + timedelta(minutes=5)

            mock_scheduler.get_jobs.return_value = [mock_job]

            result = get_job_health_status()

        assert "portfolio_sync" in result
        assert "next_run" in result["portfolio_sync"]

    def test_handles_no_scheduler(self):
        """Test handling when scheduler is not initialized."""
        from app.jobs.scheduler import get_job_health_status

        with patch("app.jobs.scheduler.scheduler", None):
            result = get_job_health_status()

        assert result == {}


class TestRescheduleAllJobs:
    """Test job rescheduling."""

    @pytest.mark.asyncio
    async def test_reschedules_jobs(self):
        """Test that jobs are rescheduled with new intervals."""
        from app.jobs.scheduler import reschedule_all_jobs

        with patch("app.jobs.scheduler.scheduler") as mock_scheduler:
            with patch(
                "app.jobs.scheduler._get_all_job_settings",
                new_callable=AsyncMock,
                return_value={
                    "job_portfolio_sync_minutes": 5,
                    "job_price_sync_minutes": 10,
                    "job_score_refresh_minutes": 60,
                    "job_health_check_minutes": 30,
                    "job_rebalance_minutes": 15,
                    "job_metrics_hour": 6,
                    "job_historical_sync_hour": 4,
                    "job_maintenance_hour": 3,
                },
            ):
                await reschedule_all_jobs()

            mock_scheduler.reschedule_job.assert_called()
