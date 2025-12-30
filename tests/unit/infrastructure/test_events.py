"""Tests for event system infrastructure.

These tests validate the event bus functionality for publishing and subscribing to events.
"""

import pytest

from app.infrastructure.events import SystemEvent, emit, get_event_bus, subscribe


class TestEventBus:
    """Test event bus functionality."""

    def test_get_event_bus_returns_singleton(self):
        """Test that get_event_bus returns the same instance."""
        bus1 = get_event_bus()
        bus2 = get_event_bus()

        assert bus1 is bus2

    def test_emit_publishes_event(self):
        """Test that emit publishes an event."""
        bus = get_event_bus()
        received_events = []

        def handler(event, **kwargs):
            received_events.append((event, kwargs))

        subscribe(SystemEvent.ERROR_OCCURRED, handler)
        emit(SystemEvent.ERROR_OCCURRED, message="test message")

        # Events are processed synchronously in this implementation
        # Check that handler was called
        assert len(received_events) >= 1
        assert received_events[-1][0] == SystemEvent.ERROR_OCCURRED
        assert received_events[-1][1].get("message") == "test message"

    def test_subscribe_adds_handler(self):
        """Test that subscribe adds a handler for an event."""
        bus = get_event_bus()
        call_count = [0]  # Use list to allow modification in nested function

        def handler(event, **kwargs):
            call_count[0] += 1

        subscribe(SystemEvent.ERROR_OCCURRED, handler)
        emit(SystemEvent.ERROR_OCCURRED)

        assert call_count[0] >= 1

    def test_multiple_handlers_for_same_event(self):
        """Test that multiple handlers can subscribe to the same event."""
        bus = get_event_bus()
        call_counts = [0, 0]

        def handler1(event, **kwargs):
            call_counts[0] += 1

        def handler2(event, **kwargs):
            call_counts[1] += 1

        subscribe(SystemEvent.ERROR_OCCURRED, handler1)
        subscribe(SystemEvent.ERROR_OCCURRED, handler2)

        emit(SystemEvent.ERROR_OCCURRED)

        assert call_counts[0] >= 1
        assert call_counts[1] >= 1

    def test_handler_receives_event_and_kwargs(self):
        """Test that handlers receive both event and keyword arguments."""
        bus = get_event_bus()
        received_data = []

        def handler(event, **kwargs):
            received_data.append({"event": event, "kwargs": kwargs})

        subscribe(SystemEvent.ERROR_OCCURRED, handler)
        emit(SystemEvent.ERROR_OCCURRED, message="test", code=123)

        assert len(received_data) >= 1
        assert received_data[-1]["event"] == SystemEvent.ERROR_OCCURRED
        assert received_data[-1]["kwargs"]["message"] == "test"
        assert received_data[-1]["kwargs"]["code"] == 123

    def test_different_events_are_separate(self):
        """Test that different events are handled separately."""
        bus = get_event_bus()
        event1_count = [0]
        event2_count = [0]

        def handler1(event, **kwargs):
            event1_count[0] += 1

        def handler2(event, **kwargs):
            event2_count[0] += 1

        subscribe(SystemEvent.ERROR_OCCURRED, handler1)
        subscribe(SystemEvent.REBALANCE_START, handler2)

        emit(SystemEvent.ERROR_OCCURRED)
        emit(SystemEvent.REBALANCE_START)

        assert event1_count[0] >= 1
        assert event2_count[0] >= 1

    def test_handler_called_for_each_emit(self):
        """Test that handlers are called each time an event is emitted."""
        bus = get_event_bus()
        call_count = [0]

        def handler(event, **kwargs):
            call_count[0] += 1

        subscribe(SystemEvent.ERROR_OCCURRED, handler)

        emit(SystemEvent.ERROR_OCCURRED)
        emit(SystemEvent.ERROR_OCCURRED)
        emit(SystemEvent.ERROR_OCCURRED)

        assert call_count[0] >= 3

    def test_emit_without_subscribers_does_not_error(self):
        """Test that emitting an event without subscribers does not error."""
        # This is fine - events can be emitted even if no one is listening
        emit(SystemEvent.ERROR_OCCURRED, message="test")

    def test_subscribe_multiple_times_same_handler(self):
        """Test that subscribing the same handler multiple times is allowed."""
        bus = get_event_bus()
        call_count = [0]

        def handler(event, **kwargs):
            call_count[0] += 1

        subscribe(SystemEvent.ERROR_OCCURRED, handler)
        subscribe(SystemEvent.ERROR_OCCURRED, handler)  # Subscribe again

        emit(SystemEvent.ERROR_OCCURRED)

        # Handler may be called once or twice depending on implementation
        assert call_count[0] >= 1

    def test_emit_with_no_kwargs(self):
        """Test that emit works with no keyword arguments."""
        bus = get_event_bus()
        received = []

        def handler(event, **kwargs):
            received.append({"event": event, "kwargs": kwargs})

        subscribe(SystemEvent.ERROR_OCCURRED, handler)
        emit(SystemEvent.ERROR_OCCURRED)

        assert len(received) >= 1
        assert received[-1]["event"] == SystemEvent.ERROR_OCCURRED
        assert received[-1]["kwargs"] == {}

