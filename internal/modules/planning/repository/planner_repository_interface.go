package repository

import (
	"github.com/aristath/sentinel/internal/modules/planning/domain"
)

// PlannerRepositoryInterface defines the contract for planner repository operations
type PlannerRepositoryInterface interface {
	// InsertSequence inserts a new sequence into the database
	InsertSequence(
		portfolioHash string,
		sequence domain.ActionSequence,
	) error

	// GetSequence retrieves a sequence by sequence hash and portfolio hash
	GetSequence(sequenceHash, portfolioHash string) (*domain.ActionSequence, error)

	// ListSequencesByPortfolioHash retrieves all sequences for a portfolio hash
	ListSequencesByPortfolioHash(
		portfolioHash string,
		limit int,
	) ([]SequenceRecord, error)

	// GetPendingSequences retrieves sequences that haven't been evaluated yet
	GetPendingSequences(
		portfolioHash string,
		limit int,
	) ([]SequenceRecord, error)

	// MarkSequenceCompleted marks a sequence as completed
	MarkSequenceCompleted(sequenceHash, portfolioHash string) error

	// DeleteSequencesByPortfolioHash deletes all sequences for a portfolio hash
	DeleteSequencesByPortfolioHash(portfolioHash string) error

	// InsertEvaluation inserts a new evaluation into the database
	InsertEvaluation(
		evaluation domain.EvaluationResult,
	) error

	// GetEvaluation retrieves an evaluation by sequence hash and portfolio hash
	GetEvaluation(sequenceHash, portfolioHash string) (*domain.EvaluationResult, error)

	// ListEvaluationsByPortfolioHash retrieves all evaluations for a portfolio hash
	ListEvaluationsByPortfolioHash(
		portfolioHash string,
	) ([]EvaluationRecord, error)

	// DeleteEvaluationsByPortfolioHash deletes all evaluations for a portfolio hash
	DeleteEvaluationsByPortfolioHash(portfolioHash string) error

	// UpsertBestResult inserts or updates the best result for a portfolio hash
	UpsertBestResult(
		portfolioHash string,
		result domain.EvaluationResult,
		sequence domain.ActionSequence,
	) error

	// GetBestResult retrieves the best result for a portfolio hash
	GetBestResult(portfolioHash string) (*domain.HolisticPlan, error)

	// DeleteBestResult deletes the best result for a portfolio hash
	DeleteBestResult(portfolioHash string) error

	// CountSequences returns the total number of sequences for a portfolio hash
	CountSequences(portfolioHash string) (int, error)

	// CountPendingSequences returns the number of pending sequences for a portfolio hash
	CountPendingSequences(portfolioHash string) (int, error)

	// CountEvaluations returns the total number of evaluations for a portfolio hash
	CountEvaluations(portfolioHash string) (int, error)
}

// Compile-time check that PlannerRepository implements PlannerRepositoryInterface
var _ PlannerRepositoryInterface = (*PlannerRepository)(nil)
