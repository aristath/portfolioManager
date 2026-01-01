"""Local Evaluator Service - Domain service wrapper for sequence evaluation."""

from app.modules.planning.domain.models import ActionCandidate
from services.evaluator.models import (
    ActionCandidateModel,
    EvaluateSequencesRequest,
    EvaluateSequencesResponse,
)


class LocalEvaluatorService:
    """
    Service for evaluating action sequences.

    Wraps the simulation and evaluation logic from holistic_planner.py
    for use by the Evaluator microservice.
    """

    def __init__(self):
        """Initialize the service."""
        pass

    async def evaluate_sequences(
        self, request: EvaluateSequencesRequest
    ) -> EvaluateSequencesResponse:
        """
        Evaluate action sequences and return top K via beam search.

        Simulates each sequence to calculate portfolio end state, then
        scores using diversification, risk, and end-state metrics.
        Maintains beam of top K sequences during evaluation.

        Args:
            request: Sequences batch and evaluation settings

        Returns:
            Top K evaluated sequences with scores

        TODO: Extract logic from holistic_planner.py:
            - simulate_sequence() (lines 2245-2329) - Core simulation
            - Evaluation loop from create_holistic_plan (lines 3207-3649):
              - Metrics pre-fetching (lines 3207-3247)
              - _evaluate_sequence() helper (lines 3292-3400)
              - Batch evaluation with beam search (lines 3402-3649)
              - Monte Carlo evaluation (lines 3470-3570) - optional
              - Stochastic scenario evaluation (lines 3578-3633) - optional
            - _calculate_transaction_cost() (lines 46-68)
            - _update_beam() helper (lines 3418-3468)
            - calculate_portfolio_score() from diversification.py
            - calculate_portfolio_end_state_score() from end_state.py
        """
        # TODO: Implement evaluation logic
        # For now, return empty results
        return EvaluateSequencesResponse(
            top_sequences=[],
            total_evaluated=len(request.sequences),
            beam_width=request.settings.beam_width,
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

    def _action_candidate_from_model(
        self, model: ActionCandidateModel
    ) -> ActionCandidate:
        """
        Convert Pydantic model to domain ActionCandidate.

        Args:
            model: ActionCandidateModel from API

        Returns:
            Domain ActionCandidate
        """
        from app.domain.value_objects.trade_side import TradeSide

        return ActionCandidate(
            side=TradeSide(model.side),
            symbol=model.symbol,
            name=model.name,
            quantity=model.quantity,
            price=model.price,
            value_eur=model.value_eur,
            currency=model.currency,
            priority=model.priority,
            reason=model.reason,
            tags=model.tags,
        )
