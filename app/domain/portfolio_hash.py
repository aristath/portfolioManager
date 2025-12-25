"""
Portfolio Hash Generation

Generates a deterministic hash from current portfolio positions.
Used to identify when recommendations apply to the same portfolio state.
"""

import hashlib
from typing import List, Dict, Any


def generate_portfolio_hash(positions: List[Dict[str, Any]]) -> str:
    """
    Generate a deterministic hash from current portfolio positions.

    Args:
        positions: List of position dicts with 'symbol' and 'quantity' keys

    Returns:
        8-character hex hash (first 8 chars of MD5)

    Example:
        positions = [{"symbol": "AAPL", "quantity": 10}, {"symbol": "MSFT", "quantity": 5}]
        hash = generate_portfolio_hash(positions)  # e.g., "a1b2c3d4"
    """
    # Filter out positions with zero or no quantity
    active_positions = [
        p for p in positions
        if p.get('quantity', 0) and p.get('quantity', 0) > 0
    ]

    # Sort positions by symbol for deterministic ordering
    sorted_positions = sorted(active_positions, key=lambda p: p['symbol'])

    # Build canonical string: "SYMBOL:QUANTITY,SYMBOL:QUANTITY,..."
    # Round quantity to integer to avoid float precision issues
    parts = [
        f"{p['symbol'].upper()}:{int(p['quantity'])}"
        for p in sorted_positions
    ]
    canonical = ",".join(parts)

    # Generate hash and return first 8 characters
    full_hash = hashlib.md5(canonical.encode()).hexdigest()
    return full_hash[:8]
