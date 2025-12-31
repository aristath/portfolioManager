"""Correlation-aware sequence filter.

Filters sequences to avoid highly correlated positions.
"""

import logging
from typing import Any, Dict, List

from app.domain.models import Security
from app.domain.value_objects.trade_side import TradeSide
from app.modules.planning.domain.calculations.filters.base import (
    SequenceFilter,
    sequence_filter_registry,
)
from app.modules.planning.domain.holistic_planner import ActionCandidate

logger = logging.getLogger(__name__)


class CorrelationAwareFilter(SequenceFilter):
    """Correlation-aware filter: Remove sequences with highly correlated buys."""

    @property
    def name(self) -> str:
        return "correlation_aware"

    def default_params(self) -> Dict[str, Any]:
        return {
            "correlation_threshold": 0.7,
            "securities": None,  # List[Security] for symbol lookup
            "max_steps": 5,
        }

    async def filter(
        self,
        sequences: List[List[ActionCandidate]],
        params: Dict[str, Any],
    ) -> List[List[ActionCandidate]]:
        """
        Filter sequences to avoid highly correlated positions.

        Uses correlation data from risk models to identify and filter out
        sequences that would create highly correlated positions.

        Args:
            sequences: List of candidate sequences to filter
            params: Filter parameters (must include securities)

        Returns:
            Filtered list of sequences with reduced correlation
        """
        securities: List[Security] = params.get("securities", [])
        correlation_threshold = params.get("correlation_threshold", 0.7)

        if not sequences or not securities:
            return sequences

        # Build correlation data
        try:
            from app.modules.optimization.services.risk_models import RiskModelBuilder

            # Get all buy symbols from sequences
            all_buy_symbols = set()
            for sequence in sequences:
                for action in sequence:
                    if action.side == TradeSide.BUY:
                        all_buy_symbols.add(action.symbol)

            if not all_buy_symbols:
                return sequences  # No buys to check

            # Build returns DataFrame for correlation calculation
            risk_builder = RiskModelBuilder()
            lookback_days = 252  # 1 year
            prices_df = await risk_builder._fetch_prices(
                list(all_buy_symbols), lookback_days
            )

            if prices_df.empty:
                return sequences  # No price data available

            # Calculate returns and correlation matrix
            returns_df = prices_df.pct_change().dropna()
            if returns_df.empty:
                return sequences

            corr_matrix = returns_df.corr()

            # Build correlation dict for quick lookup
            correlations: Dict[str, float] = {}
            symbols = list(corr_matrix.columns)
            for i, sym1 in enumerate(symbols):
                for sym2 in symbols[i + 1 :]:
                    corr = corr_matrix.loc[sym1, sym2]
                    # Store both directions
                    correlations[f"{sym1}:{sym2}"] = corr
                    correlations[f"{sym2}:{sym1}"] = corr

        except Exception as e:
            logger.warning(f"Failed to build correlations for filtering: {e}")
            return sequences  # Return all if correlation check fails

        # Build symbol set from securities
        stock_symbols = {s.symbol for s in securities}

        filtered: List[List[ActionCandidate]] = []

        for sequence in sequences:
            # Get buy symbols from sequence
            buy_symbols = [
                action.symbol
                for action in sequence
                if action.side == TradeSide.BUY and action.symbol in stock_symbols
            ]

            # Check if any pair of buys is highly correlated
            has_high_correlation = False
            for i, symbol1 in enumerate(buy_symbols):
                for symbol2 in buy_symbols[i + 1 :]:
                    # Check correlation (both directions)
                    corr_key = f"{symbol1}:{symbol2}"
                    correlation = correlations.get(corr_key)
                    if correlation and abs(correlation) > correlation_threshold:
                        has_high_correlation = True
                        logger.debug(
                            f"Filtering sequence due to high correlation ({correlation:.2f}) "
                            f"between {symbol1} and {symbol2}"
                        )
                        break
                if has_high_correlation:
                    break

            if not has_high_correlation:
                filtered.append(sequence)

        if len(filtered) < len(sequences):
            logger.info(
                f"Correlation filtering: {len(sequences)} -> {len(filtered)} sequences "
                f"(removed {len(sequences) - len(filtered)} with high correlation)"
            )

        return filtered


# Auto-register
_correlation_aware_filter = CorrelationAwareFilter()
sequence_filter_registry.register(
    _correlation_aware_filter.name, _correlation_aware_filter
)
