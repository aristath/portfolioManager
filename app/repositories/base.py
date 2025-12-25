"""Base repository utilities for common operations."""

import logging
from datetime import datetime
from typing import Any, Optional

logger = logging.getLogger(__name__)


def safe_get(row: Any, key: str, default: Any = None) -> Any:
    """
    Safely get a value from a database row.
    
    Args:
        row: Database row (dict-like object)
        key: Key to retrieve
        default: Default value if key not found
        
    Returns:
        Value from row or default
    """
    try:
        if hasattr(row, 'keys') and key in row.keys():
            return row[key]
        elif hasattr(row, '__getitem__'):
            return row[key]
        else:
            return default
    except (KeyError, IndexError, AttributeError):
        return default


def safe_get_datetime(row: Any, key: str) -> Optional[datetime]:
    """
    Safely parse a datetime from a database row.
    
    Args:
        row: Database row (dict-like object)
        key: Key containing datetime string
        
    Returns:
        Parsed datetime or None if invalid/missing
    """
    value = safe_get(row, key)
    if not value:
        return None
    
    if isinstance(value, datetime):
        return value
    
    try:
        if isinstance(value, str):
            return datetime.fromisoformat(value)
    except (ValueError, TypeError) as e:
        logger.warning(f"Failed to parse datetime from {key}: {e}")
        return None
    
    return None


def safe_get_bool(row: Any, key: str, default: bool = False) -> bool:
    """
    Safely get a boolean value from a database row.
    
    Handles integer 0/1 values and boolean values.
    
    Args:
        row: Database row (dict-like object)
        key: Key to retrieve
        default: Default value if key not found
        
    Returns:
        Boolean value
    """
    value = safe_get(row, key, default)
    
    if isinstance(value, bool):
        return value
    
    if isinstance(value, int):
        return bool(value)
    
    if isinstance(value, str):
        return value.lower() in ('true', '1', 'yes', 'on')
    
    return bool(value) if value is not None else default


def safe_get_float(row: Any, key: str, default: Optional[float] = None) -> Optional[float]:
    """
    Safely get a float value from a database row.
    
    Args:
        row: Database row (dict-like object)
        key: Key to retrieve
        default: Default value if key not found or invalid
        
    Returns:
        Float value or default
    """
    value = safe_get(row, key, default)
    
    if value is None:
        return default
    
    if isinstance(value, (int, float)):
        return float(value)
    
    if isinstance(value, str):
        try:
            return float(value)
        except (ValueError, TypeError):
            return default
    
    return default


def safe_get_int(row: Any, key: str, default: Optional[int] = None) -> Optional[int]:
    """
    Safely get an integer value from a database row.
    
    Args:
        row: Database row (dict-like object)
        key: Key to retrieve
        default: Default value if key not found or invalid
        
    Returns:
        Integer value or default
    """
    value = safe_get(row, key, default)
    
    if value is None:
        return default
    
    if isinstance(value, int):
        return value
    
    if isinstance(value, float):
        return int(value)
    
    if isinstance(value, str):
        try:
            return int(float(value))  # Handle "123.0" strings
        except (ValueError, TypeError):
            return default
    
    return default

