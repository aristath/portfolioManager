"""Tests for file-based locking infrastructure.

These tests validate file-based locking for preventing concurrent execution
of critical operations.
"""

import asyncio

import pytest


class TestFileLock:
    """Test file_lock context manager."""

    @pytest.mark.asyncio
    async def test_acquires_and_releases_lock(self):
        """Test that lock is acquired and released."""
        from app.infrastructure.locking import file_lock

        async with file_lock("test", timeout=1.0):
            # Lock should be held - can't verify file system directly
            # but should complete without error
            pass

        # Lock should be released - can acquire again
        async with file_lock("test", timeout=1.0):
            pass

    @pytest.mark.asyncio
    async def test_prevents_concurrent_execution(self):
        """Test that concurrent execution is prevented."""
        from app.infrastructure.locking import file_lock

        execution_order = []
        lock_acquired = asyncio.Event()

        async def task1():
            async with file_lock("test", timeout=1.0):
                execution_order.append("task1_start")
                lock_acquired.set()
                await asyncio.sleep(0.1)
                execution_order.append("task1_end")

        async def task2():
            await lock_acquired.wait()  # Wait for task1 to acquire lock
            try:
                async with file_lock("test", timeout=0.05):  # Short timeout
                    execution_order.append("task2_start")
                    execution_order.append("task2_end")
            except TimeoutError:
                execution_order.append("task2_timeout")

        # Run both tasks concurrently
        await asyncio.gather(task1(), task2(), return_exceptions=True)

        # task1 should complete before task2 can start
        assert "task1_start" in execution_order
        assert "task1_end" in execution_order
        # task2 should timeout
        assert "task2_timeout" in execution_order

    @pytest.mark.asyncio
    async def test_handles_timeout_gracefully(self):
        """Test that lock timeout is handled gracefully."""
        from app.infrastructure.locking import file_lock

        async def hold_lock_long():
            async with file_lock("test", timeout=1.0):
                await asyncio.sleep(0.2)  # Hold lock longer

        async def try_lock_short():
            try:
                async with file_lock("test", timeout=0.05):  # Very short timeout
                    pass
            except TimeoutError:
                return True
            return False

        # Start first task to hold lock
        lock_task = asyncio.create_task(hold_lock_long())

        # Try to acquire lock with short timeout
        timed_out = await try_lock_short()

        await lock_task  # Wait for lock to be released

        # Should have timed out
        assert timed_out

    @pytest.mark.asyncio
    async def test_allows_sequential_execution(self):
        """Test that sequential execution is allowed."""
        from app.infrastructure.locking import file_lock

        execution_order = []

        async def task1():
            async with file_lock("test", timeout=1.0):
                execution_order.append("task1")
                await asyncio.sleep(0.05)

        async def task2():
            await asyncio.sleep(0.1)  # Wait for task1 to complete
            async with file_lock("test", timeout=1.0):
                execution_order.append("task2")

        await task1()
        await task2()

        assert execution_order == ["task1", "task2"]

    @pytest.mark.asyncio
    async def test_uses_different_locks_for_different_names(self):
        """Test that different lock names don't interfere."""
        from app.infrastructure.locking import file_lock

        execution_order = []

        async def task1():
            async with file_lock("lock1", timeout=1.0):
                execution_order.append("lock1_start")
                await asyncio.sleep(0.1)
                execution_order.append("lock1_end")

        async def task2():
            async with file_lock("lock2", timeout=1.0):
                execution_order.append("lock2_start")
                await asyncio.sleep(0.05)
                execution_order.append("lock2_end")

        # Both should be able to run concurrently
        await asyncio.gather(task1(), task2())

        # Both should have executed
        assert "lock1_start" in execution_order
        assert "lock1_end" in execution_order
        assert "lock2_start" in execution_order
        assert "lock2_end" in execution_order

    @pytest.mark.asyncio
    async def test_handles_exceptions_and_releases_lock(self):
        """Test that lock is released even when exception occurs."""
        from app.infrastructure.locking import file_lock

        async def task_with_exception():
            try:
                async with file_lock("test", timeout=1.0):
                    raise ValueError("Test exception")
            except ValueError:
                pass

        async def task_after_exception():
            await asyncio.sleep(0.1)  # Wait for exception to be handled
            async with file_lock("test", timeout=1.0):
                return True

        await task_with_exception()
        result = await task_after_exception()

        # Should be able to acquire lock after exception
        assert result is True

    @pytest.mark.asyncio
    async def test_timeout_error_includes_lock_name(self):
        """Test that timeout error message includes lock name."""
        from app.infrastructure.locking import file_lock

        async def hold_lock():
            async with file_lock("test_timeout", timeout=1.0):
                await asyncio.sleep(0.2)

        async def try_lock():
            try:
                async with file_lock("test_timeout", timeout=0.05):
                    pass
            except TimeoutError as e:
                return str(e)
            return None

        lock_task = asyncio.create_task(hold_lock())
        error_message = await try_lock()
        await lock_task

        assert error_message is not None
        assert "test_timeout" in error_message
