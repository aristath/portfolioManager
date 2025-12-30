"""Tests for simple cache infrastructure.

These tests validate the in-memory cache functionality for caching computed values.
"""

import pytest

from app.infrastructure.cache import SimpleCache


class TestSimpleCache:
    """Test SimpleCache class."""

    @pytest.fixture
    def cache(self):
        """Create a SimpleCache instance."""
        return SimpleCache()

    def test_init_creates_empty_cache(self, cache):
        """Test that cache is empty after initialization."""
        assert cache._cache == {}

    def test_set_stores_value(self, cache):
        """Test that set stores a value in the cache."""
        cache.set("key1", "value1")
        assert cache._cache["key1"] == "value1"

    def test_get_returns_stored_value(self, cache):
        """Test that get returns the stored value."""
        cache.set("key1", "value1")
        assert cache.get("key1") == "value1"

    def test_get_returns_none_for_missing_key(self, cache):
        """Test that get returns None for missing keys."""
        assert cache.get("nonexistent") is None

    def test_get_returns_default_for_missing_key(self, cache):
        """Test that get returns default value for missing keys."""
        assert cache.get("nonexistent", default="default_value") == "default_value"

    def test_invalidate_removes_key(self, cache):
        """Test that invalidate removes a key from the cache."""
        cache.set("key1", "value1")
        cache.set("key2", "value2")

        cache.invalidate("key1")

        assert cache.get("key1") is None
        assert cache.get("key2") == "value2"

    def test_invalidate_nonexistent_key_does_not_error(self, cache):
        """Test that invalidate on nonexistent key does not raise error."""
        cache.invalidate("nonexistent")  # Should not raise

    def test_invalidate_prefix_removes_matching_keys(self, cache):
        """Test that invalidate_prefix removes all keys with matching prefix."""
        cache.set("prefix:key1", "value1")
        cache.set("prefix:key2", "value2")
        cache.set("other:key1", "value3")
        cache.set("prefix_key3", "value4")  # No colon, should not match

        cache.invalidate_prefix("prefix:")

        assert cache.get("prefix:key1") is None
        assert cache.get("prefix:key2") is None
        assert cache.get("other:key1") == "value3"  # Not removed
        assert cache.get("prefix_key3") == "value4"  # Not removed (no colon)

    def test_invalidate_prefix_with_no_matches(self, cache):
        """Test that invalidate_prefix works when no keys match."""
        cache.set("key1", "value1")

        cache.invalidate_prefix("nonexistent:")

        assert cache.get("key1") == "value1"  # Still there

    def test_clear_removes_all_keys(self, cache):
        """Test that clear removes all keys from the cache."""
        cache.set("key1", "value1")
        cache.set("key2", "value2")
        cache.set("key3", "value3")

        cache.clear()

        assert cache.get("key1") is None
        assert cache.get("key2") is None
        assert cache.get("key3") is None
        assert len(cache._cache) == 0

    def test_overwrite_existing_key(self, cache):
        """Test that setting an existing key overwrites the value."""
        cache.set("key1", "value1")
        cache.set("key1", "value2")

        assert cache.get("key1") == "value2"

    def test_store_various_data_types(self, cache):
        """Test that cache can store various data types."""
        cache.set("string", "value")
        cache.set("int", 123)
        cache.set("float", 45.67)
        cache.set("bool", True)
        cache.set("list", [1, 2, 3])
        cache.set("dict", {"key": "value"})
        cache.set("none", None)

        assert cache.get("string") == "value"
        assert cache.get("int") == 123
        assert cache.get("float") == 45.67
        assert cache.get("bool") is True
        assert cache.get("list") == [1, 2, 3]
        assert cache.get("dict") == {"key": "value"}
        assert cache.get("none") is None

    def test_invalidate_prefix_with_empty_string(self, cache):
        """Test that invalidate_prefix with empty string removes all keys."""
        cache.set("key1", "value1")
        cache.set("key2", "value2")

        cache.invalidate_prefix("")

        # Empty prefix should match all keys (all strings start with empty string)
        assert cache.get("key1") is None
        assert cache.get("key2") is None

    def test_multiple_caches_are_independent(self):
        """Test that multiple cache instances are independent."""
        cache1 = SimpleCache()
        cache2 = SimpleCache()

        cache1.set("key1", "value1")
        cache2.set("key1", "value2")

        assert cache1.get("key1") == "value1"
        assert cache2.get("key1") == "value2"

