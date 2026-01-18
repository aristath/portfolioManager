// Package evaluation provides type aliases for evaluation models.
// This package delegates type definitions to internal/evaluation/models
// to maintain a single source of truth and eliminate duplication.
package evaluation

import (
	coremodels "github.com/aristath/sentinel/internal/evaluation/models"
)

// Type aliases - all types are now unified with core evaluation models
type TradeSide = coremodels.TradeSide
type ActionCandidate = coremodels.ActionCandidate
type Security = coremodels.Security
type Position = coremodels.Position
type PortfolioContext = coremodels.PortfolioContext
type EvaluationContext = coremodels.EvaluationContext
type SequenceEvaluationResult = coremodels.SequenceEvaluationResult
type BatchEvaluationRequest = coremodels.BatchEvaluationRequest
type BatchEvaluationResponse = coremodels.BatchEvaluationResponse
type MonteCarloRequest = coremodels.MonteCarloRequest
type MonteCarloResult = coremodels.MonteCarloResult
type StochasticRequest = coremodels.StochasticRequest
type StochasticResult = coremodels.StochasticResult
type BatchSimulationRequest = coremodels.BatchSimulationRequest
type SimulationResult = coremodels.SimulationResult
type BatchSimulationResponse = coremodels.BatchSimulationResponse

// Re-export constants for convenience
const (
	TradeSideBuy  = coremodels.TradeSideBuy
	TradeSideSell = coremodels.TradeSideSell
)
