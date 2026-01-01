"""Local Generator Service - Domain service wrapper for sequence generation."""

from typing import AsyncIterator, List

from app.modules.planning.domain.models import ActionCandidate
from services.generator.models import (
    ActionCandidateModel,
    GenerateSequencesRequest,
    SequenceBatch,
)


class LocalGeneratorService:
    """
    Service for generating and filtering action sequences.

    Wraps the sequence generation logic from holistic_planner.py
    for use by the Generator microservice.
    """

    def __init__(self):
        """Initialize the service."""
        pass

    async def generate_sequences_batched(
        self, request: GenerateSequencesRequest
    ) -> AsyncIterator[SequenceBatch]:
        """
        Generate action sequences from opportunities and yield in batches.

        Uses combinatorial generation with adaptive patterns, then applies
        filters (correlation-aware, feasibility), and yields results in
        batches for streaming to evaluators.

        Args:
            request: Opportunities, settings, and batch size

        Yields:
            SequenceBatch objects containing sequences

        TODO: Extract logic from holistic_planner.py:
            - generate_action_sequences() (lines 2136-2242) - MAIN FUNCTION
            - _generate_patterns_at_depth() (lines 1927-2133)
            - _generate_combinations() (lines 1863-1925)
            - _generate_weighted_combinations() (lines 1730-1860)
            - _generate_adaptive_patterns() (lines 1076-1338)
            - _filter_correlation_aware_sequences() (lines 1339-1420)
            - _generate_partial_execution_scenarios() (lines 1422-1490)
            - _generate_constraint_relaxation_scenarios() (lines 1492-1560)
            - _hash_sequence() (lines 71-89)
            - Helper functions (lines 1562-1650)
            - Feasibility filtering (lines 3114-3205 from create_holistic_plan)
        """
        # TODO: Implement sequence generation logic
        # For now, return empty batch
        all_sequences: List[List[ActionCandidate]] = []

        # Yield in batches
        batch_size = request.batch_size
        total_batches = max(1, (len(all_sequences) + batch_size - 1) // batch_size)

        for batch_number in range(total_batches):
            start_idx = batch_number * batch_size
            end_idx = min(start_idx + batch_size, len(all_sequences))
            batch_sequences = all_sequences[start_idx:end_idx]

            # Convert domain models to Pydantic
            pydantic_sequences = [
                [self._action_candidate_to_model(action) for action in sequence]
                for sequence in batch_sequences
            ]

            yield SequenceBatch(
                batch_number=batch_number,
                sequences=pydantic_sequences,
                total_batches=total_batches,
                more_available=batch_number < total_batches - 1,
            )

    def _action_candidate_to_model(
        self, action: ActionCandidate
    ) -> ActionCandidateModel:
        """
        Convert domain ActionCandidate to Pydantic model.

        Args:
            action: Domain ActionCandidate

        Returns:
            ActionCandidateModel for API response
        """
        return ActionCandidateModel(
            side=action.side.value if hasattr(action.side, "value") else action.side,
            symbol=action.symbol,
            name=action.name,
            quantity=action.quantity,
            price=action.price,
            value_eur=action.value_eur,
            currency=action.currency,
            priority=action.priority,
            reason=action.reason,
            tags=action.tags if hasattr(action, "tags") else [],
        )
