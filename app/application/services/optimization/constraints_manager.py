"""
Constraints Manager for Portfolio Optimization.

Translates business rules into PyPortfolioOpt constraints:
- allow_buy/allow_sell flags
- min_lot constraints (can't partially sell if at min lot)
- Concentration limits (20% max per stock)
- Country/Industry sector constraints
"""

import logging
from dataclasses import dataclass
from typing import Dict, List, Optional, Tuple

from app.domain.models import Position, Stock
from app.domain.scoring.constants import (
    GEO_ALLOCATION_TOLERANCE,
    IND_ALLOCATION_TOLERANCE,
    MAX_CONCENTRATION,
    MAX_COUNTRY_CONCENTRATION,
    MAX_SECTOR_CONCENTRATION,
)

logger = logging.getLogger(__name__)


@dataclass
class WeightBounds:
    """Weight bounds for a single stock."""

    symbol: str
    lower: float  # Minimum weight (0.0 to 1.0)
    upper: float  # Maximum weight (0.0 to 1.0)
    reason: Optional[str] = None


@dataclass
class SectorConstraint:
    """Constraint for a sector (country or industry)."""

    name: str
    symbols: List[str]
    target: float  # Target weight
    lower: float  # Lower bound
    upper: float  # Upper bound


class ConstraintsManager:
    """Manage portfolio optimization constraints."""

    def __init__(
        self,
        max_concentration: float = MAX_CONCENTRATION,
        geo_tolerance: float = GEO_ALLOCATION_TOLERANCE,
        ind_tolerance: float = IND_ALLOCATION_TOLERANCE,
    ):
        self.max_concentration = max_concentration
        self.geo_tolerance = geo_tolerance
        self.ind_tolerance = ind_tolerance

    def calculate_weight_bounds(
        self,
        stocks: List[Stock],
        positions: Dict[str, Position],
        portfolio_value: float,
        current_prices: Dict[str, float],
    ) -> Dict[str, Tuple[float, float]]:
        """
        Calculate weight bounds for each stock.

        Args:
            stocks: List of Stock objects
            positions: Dict mapping symbol to Position
            portfolio_value: Total portfolio value in EUR
            current_prices: Dict mapping symbol to current price

        Returns:
            Dict mapping symbol to (lower_bound, upper_bound) tuple
        """
        bounds = {}

        logger.debug(
            f"Calculating weight bounds for {len(stocks)} stocks, "
            f"portfolio_value={portfolio_value:.2f} EUR"
        )

        for stock in stocks:
            symbol = stock.symbol
            position = positions.get(symbol)
            current_price = current_prices.get(symbol, 0)

            # Calculate current weight
            if (
                position is not None
                and position.market_value_eur is not None
                and portfolio_value > 0
            ):
                current_weight = position.market_value_eur / portfolio_value
            else:
                current_weight = 0.0

            # Track constraint application for diagnostics
            constraint_steps = []

            # Default bounds
            lower = 0.0
            upper = self.max_concentration
            constraint_steps.append(f"initial: lower={lower:.2%}, upper={upper:.2%}")

            # Apply user-defined portfolio targets (convert percentage to fraction)
            if stock.min_portfolio_target is not None:
                lower = stock.min_portfolio_target / 100.0
                constraint_steps.append(
                    f"min_portfolio_target={stock.min_portfolio_target}% → lower={lower:.2%}"
                )

            if stock.max_portfolio_target is not None:
                upper = stock.max_portfolio_target / 100.0
                constraint_steps.append(
                    f"max_portfolio_target={stock.max_portfolio_target}% → upper={upper:.2%}"
                )

            # Check allow_buy constraint
            if not stock.allow_buy:
                # Can't buy more, so upper bound = current weight
                old_upper = upper
                upper = min(upper, current_weight)
                constraint_steps.append(
                    f"allow_buy=False → upper=min({old_upper:.2%}, {current_weight:.2%})={upper:.2%}"
                )

            # Check allow_sell constraint
            if not stock.allow_sell:
                # Can't sell, so lower bound = current weight
                old_lower = lower
                lower = max(lower, current_weight)
                constraint_steps.append(
                    f"allow_sell=False → lower=max({old_lower:.2%}, {current_weight:.2%})={lower:.2%}"
                )

            # Check min_lot constraint
            if position and stock.min_lot > 0 and current_price > 0:
                if position.quantity <= stock.min_lot:
                    # Can't partially sell - it's all or nothing
                    # Set lower bound to current weight (can't reduce)
                    old_lower = lower
                    lower = max(lower, current_weight)
                    constraint_steps.append(
                        f"at min_lot (qty={position.quantity} <= {stock.min_lot}) → "
                        f"lower=max({old_lower:.2%}, {current_weight:.2%})={lower:.2%}"
                    )
                else:
                    # Can sell down to min_lot worth
                    min_lot_value = stock.min_lot * current_price
                    min_weight = (
                        min_lot_value / portfolio_value if portfolio_value > 0 else 0
                    )
                    old_lower = lower
                    lower = max(lower, min_weight)
                    constraint_steps.append(
                        f"min_lot constraint (min_lot_value={min_lot_value:.2f} EUR) → "
                        f"lower=max({old_lower:.2%}, {min_weight:.2%})={lower:.2%}"
                    )

            # Ensure lower <= upper
            if lower > upper:
                # Constraint conflict - keep current weight
                logger.warning(
                    f"{symbol}: constraint conflict detected! "
                    f"lower={lower:.2%} > upper={upper:.2%}, "
                    f"current_weight={current_weight:.2%}, "
                    f"portfolio_value={portfolio_value:.2f} EUR, "
                    f"position_value={position.market_value_eur if position and position.market_value_eur else 0:.2f} EUR, "
                    f"min_portfolio_target={stock.min_portfolio_target}, "
                    f"max_portfolio_target={stock.max_portfolio_target}, "
                    f"allow_sell={stock.allow_sell}, allow_buy={stock.allow_buy}, "
                    f"min_lot={stock.min_lot}, position_qty={position.quantity if position else 0}, "
                    f"current_price={current_price:.2f}. "
                    f"Constraint steps: {'; '.join(constraint_steps)}. "
                    f"Using current weight {current_weight:.2%} for both bounds."
                )
                lower = current_weight
                upper = current_weight
            elif lower == upper and lower > 0:
                # Locked position - log for diagnostics
                logger.debug(
                    f"{symbol}: locked position (lower=upper={lower:.2%}), "
                    f"constraint steps: {'; '.join(constraint_steps)}"
                )

            bounds[symbol] = (lower, upper)

        return bounds

    def build_sector_constraints(
        self,
        stocks: List[Stock],
        country_targets: Dict[str, float],
        ind_targets: Dict[str, float],
    ) -> Tuple[List[SectorConstraint], List[SectorConstraint]]:
        """
        Build country and industry sector constraints.

        Args:
            stocks: List of Stock objects
            country_targets: Dict mapping country name to target weight
            ind_targets: Dict mapping industry name to target weight

        Returns:
            Tuple of (country_constraints, industry_constraints)
        """
        # Group stocks by country
        country_groups: Dict[str, List[str]] = {}
        for stock in stocks:
            country = stock.country or "OTHER"
            if country not in country_groups:
                country_groups[country] = []
            country_groups[country].append(stock.symbol)

        # Group stocks by industry
        ind_groups: Dict[str, List[str]] = {}
        for stock in stocks:
            ind = stock.industry or "OTHER"
            if ind not in ind_groups:
                ind_groups[ind] = []
            ind_groups[ind].append(stock.symbol)

        # Build country constraints
        country_constraints = []
        for country, symbols in country_groups.items():
            target = country_targets.get(country, 0.0)
            if target > 0:
                # Calculate tolerance-based bounds
                tolerance_upper = min(1.0, target + self.geo_tolerance)
                # Enforce hard limit: cap at MAX_COUNTRY_CONCENTRATION
                hard_upper = min(tolerance_upper, MAX_COUNTRY_CONCENTRATION)
                country_constraints.append(
                    SectorConstraint(
                        name=country,
                        symbols=symbols,
                        target=target,
                        lower=max(0.0, target - self.geo_tolerance),
                        upper=hard_upper,
                    )
                )

        # Build industry constraints
        ind_constraints = []
        for ind, symbols in ind_groups.items():
            target = ind_targets.get(ind, 0.0)
            if target > 0:
                # Calculate tolerance-based bounds
                tolerance_upper = min(1.0, target + self.ind_tolerance)
                # Enforce hard limit: cap at MAX_SECTOR_CONCENTRATION
                hard_upper = min(tolerance_upper, MAX_SECTOR_CONCENTRATION)
                ind_constraints.append(
                    SectorConstraint(
                        name=ind,
                        symbols=symbols,
                        target=target,
                        lower=max(0.0, target - self.ind_tolerance),
                        upper=hard_upper,
                    )
                )

        logger.info(
            f"Built {len(country_constraints)} country constraints, "
            f"{len(ind_constraints)} industry constraints"
        )

        return country_constraints, ind_constraints

    def get_constraint_summary(
        self,
        bounds: Dict[str, Tuple[float, float]],
        country_constraints: List[SectorConstraint],
        ind_constraints: List[SectorConstraint],
    ) -> Dict:
        """
        Get a summary of all constraints for logging/debugging.

        Returns:
            Dict with constraint details
        """
        # Count constrained stocks
        locked = []  # lower == upper (can't change)
        buy_only = []  # lower == 0, upper < max (can only buy)
        sell_blocked = []  # lower > 0 (can't fully exit)

        for symbol, (lower, upper) in bounds.items():
            if lower == upper:
                locked.append(symbol)
            elif lower == 0 and upper < self.max_concentration:
                buy_only.append(symbol)
            elif lower > 0:
                sell_blocked.append(symbol)

        return {
            "total_stocks": len(bounds),
            "locked_positions": locked,
            "buy_only": buy_only,
            "sell_blocked": sell_blocked,
            "country_constraints": [
                {"name": c.name, "target": c.target, "range": (c.lower, c.upper)}
                for c in country_constraints
            ],
            "industry_constraints": [
                {"name": c.name, "target": c.target, "range": (c.lower, c.upper)}
                for c in ind_constraints
            ],
        }
