package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aristath/sentinel/internal/database"
	"github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/rs/zerolog"
)

// PlannerRepository handles database operations for planning module.
// Database: agents.db (sequences, evaluations, best_result tables)
type PlannerRepository struct {
	db  *database.DB // agents.db
	log zerolog.Logger
}

// NewPlannerRepository creates a new planner repository.
// db parameter should be agents.db connection
func NewPlannerRepository(db *database.DB, log zerolog.Logger) *PlannerRepository {
	return &PlannerRepository{
		db:  db,
		log: log.With().Str("component", "planner_repository").Logger(),
	}
}

// SequenceRecord represents a sequence in the database.
type SequenceRecord struct {
	SequenceHash  string
	PortfolioHash string
	SequenceJSON  string // JSON serialized ActionSequence
	PatternType   string
	Depth         int
	Priority      float64
	Completed     bool
	EvaluatedAt   *time.Time // Nullable
	CreatedAt     time.Time
}

// EvaluationRecord represents an evaluation in the database.
type EvaluationRecord struct {
	SequenceHash            string
	PortfolioHash           string
	EndScore                float64
	BreakdownJSON           string // JSON serialized score breakdown
	EndCash                 float64
	EndContextPositionsJSON string // JSON serialized positions
	DiversificationScore    float64
	TotalValue              float64
	EvaluatedAt             time.Time
}

// BestResultRecord represents the best result in the database.
type BestResultRecord struct {
	PortfolioHash string
	SequenceHash  string
	PlanData      string // JSON serialized HolisticPlan
	Score         float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// GetSequence retrieves a sequence by sequence hash and portfolio hash.
func (r *PlannerRepository) GetSequence(sequenceHash, portfolioHash string) (*domain.ActionSequence, error) {
	var record SequenceRecord
	var evaluatedAtUnix sql.NullInt64
	var createdAtUnix sql.NullInt64
	err := r.db.QueryRow(`
		SELECT sequence_hash, portfolio_hash, sequence_json, pattern_type, depth, priority, completed, evaluated_at, created_at
		FROM sequences
		WHERE sequence_hash = ? AND portfolio_hash = ?
	`, sequenceHash, portfolioHash).Scan(
		&record.SequenceHash,
		&record.PortfolioHash,
		&record.SequenceJSON,
		&record.PatternType,
		&record.Depth,
		&record.Priority,
		&record.Completed,
		&evaluatedAtUnix,
		&createdAtUnix,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get sequence: %w", err)
	}

	// Note: EvaluatedAt and CreatedAt are scanned but not used in domain.ActionSequence
	// They're kept in SequenceRecord struct for potential future use

	// Unmarshal actions from sequence_json
	var actions []domain.ActionCandidate
	if err := json.Unmarshal([]byte(record.SequenceJSON), &actions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sequence actions: %w", err)
	}

	sequence := &domain.ActionSequence{
		Actions:      actions,
		Priority:     record.Priority,
		Depth:        record.Depth,
		PatternType:  record.PatternType,
		SequenceHash: record.SequenceHash,
	}

	return sequence, nil
}

// ListSequencesByPortfolioHash retrieves all sequences for a portfolio hash.
func (r *PlannerRepository) ListSequencesByPortfolioHash(
	portfolioHash string,
	limit int,
) ([]SequenceRecord, error) {
	query := `
		SELECT sequence_hash, portfolio_hash, sequence_json, pattern_type, depth, priority, completed, evaluated_at, created_at
		FROM sequences
		WHERE portfolio_hash = ?
		ORDER BY priority DESC, created_at DESC
	`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.Query(query, portfolioHash)
	if err != nil {
		return nil, fmt.Errorf("failed to list sequences: %w", err)
	}
	defer rows.Close()

	var records []SequenceRecord
	for rows.Next() {
		var record SequenceRecord
		var evaluatedAtUnix, createdAtUnix sql.NullInt64
		if err := rows.Scan(
			&record.SequenceHash,
			&record.PortfolioHash,
			&record.SequenceJSON,
			&record.PatternType,
			&record.Depth,
			&record.Priority,
			&record.Completed,
			&evaluatedAtUnix,
			&createdAtUnix,
		); err != nil {
			return nil, fmt.Errorf("failed to scan sequence: %w", err)
		}
		if evaluatedAtUnix.Valid {
			t := time.Unix(evaluatedAtUnix.Int64, 0).UTC()
			record.EvaluatedAt = &t
		}
		if createdAtUnix.Valid {
			record.CreatedAt = time.Unix(createdAtUnix.Int64, 0).UTC()
		}
		records = append(records, record)
	}

	return records, nil
}

// GetPendingSequences retrieves sequences that haven't been evaluated yet.
func (r *PlannerRepository) GetPendingSequences(
	portfolioHash string,
	limit int,
) ([]SequenceRecord, error) {
	query := `
		SELECT sequence_hash, portfolio_hash, sequence_json, pattern_type, depth, priority, completed, evaluated_at, created_at
		FROM sequences
		WHERE portfolio_hash = ? AND completed = 0
		ORDER BY priority DESC, created_at ASC
	`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.Query(query, portfolioHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending sequences: %w", err)
	}
	defer rows.Close()

	var records []SequenceRecord
	for rows.Next() {
		var record SequenceRecord
		var evaluatedAtUnix, createdAtUnix sql.NullInt64
		if err := rows.Scan(
			&record.SequenceHash,
			&record.PortfolioHash,
			&record.SequenceJSON,
			&record.PatternType,
			&record.Depth,
			&record.Priority,
			&record.Completed,
			&evaluatedAtUnix,
			&createdAtUnix,
		); err != nil {
			return nil, fmt.Errorf("failed to scan sequence: %w", err)
		}
		if evaluatedAtUnix.Valid {
			t := time.Unix(evaluatedAtUnix.Int64, 0).UTC()
			record.EvaluatedAt = &t
		}
		if createdAtUnix.Valid {
			record.CreatedAt = time.Unix(createdAtUnix.Int64, 0).UTC()
		}
		records = append(records, record)
	}

	return records, nil
}

// MarkSequenceCompleted marks a sequence as completed.
func (r *PlannerRepository) MarkSequenceCompleted(sequenceHash, portfolioHash string) error {
	now := time.Now().Unix()
	_, err := r.db.Exec(`
		UPDATE sequences
		SET completed = 1, evaluated_at = ?
		WHERE sequence_hash = ? AND portfolio_hash = ?
	`, now, sequenceHash, portfolioHash)
	if err != nil {
		return fmt.Errorf("failed to mark sequence completed: %w", err)
	}

	r.log.Debug().
		Str("sequence_hash", sequenceHash).
		Str("portfolio_hash", portfolioHash).
		Msg("Marked sequence as completed")
	return nil
}

// DeleteSequencesByPortfolioHash deletes all sequences for a portfolio hash.
func (r *PlannerRepository) DeleteSequencesByPortfolioHash(portfolioHash string) error {
	result, err := r.db.Exec(`DELETE FROM sequences WHERE portfolio_hash = ?`, portfolioHash)
	if err != nil {
		return fmt.Errorf("failed to delete sequences: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	r.log.Info().
		Str("portfolio_hash", portfolioHash).
		Int64("rows_deleted", rowsAffected).
		Msg("Deleted sequences")

	return nil
}

// DeleteAllSequences deletes all sequences regardless of portfolio hash.
func (r *PlannerRepository) DeleteAllSequences() error {
	result, err := r.db.Exec(`DELETE FROM sequences`)
	if err != nil {
		return fmt.Errorf("failed to delete all sequences: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	r.log.Info().
		Int64("rows_deleted", rowsAffected).
		Msg("Deleted all sequences")

	return nil
}

// GetEvaluation retrieves an evaluation by sequence hash and portfolio hash.
func (r *PlannerRepository) GetEvaluation(sequenceHash, portfolioHash string) (*domain.EvaluationResult, error) {
	var record EvaluationRecord
	var evaluatedAtUnix sql.NullInt64
	err := r.db.QueryRow(`
		SELECT sequence_hash, portfolio_hash, end_score, breakdown_json, end_cash, end_context_positions_json, div_score, total_value, evaluated_at
		FROM evaluations
		WHERE sequence_hash = ? AND portfolio_hash = ?
	`, sequenceHash, portfolioHash).Scan(
		&record.SequenceHash,
		&record.PortfolioHash,
		&record.EndScore,
		&record.BreakdownJSON,
		&record.EndCash,
		&record.EndContextPositionsJSON,
		&record.DiversificationScore,
		&record.TotalValue,
		&evaluatedAtUnix,
	)

	// Note: EvaluatedAt is scanned but not used in domain.Evaluation
	// It's kept in EvaluationRecord struct for potential future use

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get evaluation: %w", err)
	}

	var scoreBreakdown map[string]float64
	if err := json.Unmarshal([]byte(record.BreakdownJSON), &scoreBreakdown); err != nil {
		return nil, fmt.Errorf("failed to unmarshal score breakdown: %w", err)
	}

	var endContextPositions map[string]float64
	if err := json.Unmarshal([]byte(record.EndContextPositionsJSON), &endContextPositions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal end context positions: %w", err)
	}

	evaluation := &domain.EvaluationResult{
		SequenceHash:         record.SequenceHash,
		PortfolioHash:        record.PortfolioHash,
		EndScore:             record.EndScore,
		ScoreBreakdown:       scoreBreakdown,
		EndCash:              record.EndCash,
		EndContextPositions:  endContextPositions,
		DiversificationScore: record.DiversificationScore,
		TotalValue:           record.TotalValue,
	}

	return evaluation, nil
}

// ListEvaluationsByPortfolioHash retrieves all evaluations for a portfolio hash.
func (r *PlannerRepository) ListEvaluationsByPortfolioHash(
	portfolioHash string,
) ([]EvaluationRecord, error) {
	rows, err := r.db.Query(`
		SELECT sequence_hash, portfolio_hash, end_score, breakdown_json, end_cash, end_context_positions_json, div_score, total_value, evaluated_at
		FROM evaluations
		WHERE portfolio_hash = ?
		ORDER BY end_score DESC, evaluated_at DESC
	`, portfolioHash)

	if err != nil {
		return nil, fmt.Errorf("failed to list evaluations: %w", err)
	}
	defer rows.Close()

	var records []EvaluationRecord
	for rows.Next() {
		var record EvaluationRecord
		var evaluatedAtUnix sql.NullInt64
		if err := rows.Scan(
			&record.SequenceHash,
			&record.PortfolioHash,
			&record.EndScore,
			&record.BreakdownJSON,
			&record.EndCash,
			&record.EndContextPositionsJSON,
			&record.DiversificationScore,
			&record.TotalValue,
			&evaluatedAtUnix,
		); err != nil {
			return nil, fmt.Errorf("failed to scan evaluation: %w", err)
		}
		if evaluatedAtUnix.Valid {
			record.EvaluatedAt = time.Unix(evaluatedAtUnix.Int64, 0).UTC()
		}
		records = append(records, record)
	}

	return records, nil
}

// DeleteEvaluationsByPortfolioHash deletes all evaluations for a portfolio hash.
func (r *PlannerRepository) DeleteEvaluationsByPortfolioHash(portfolioHash string) error {
	result, err := r.db.Exec(`DELETE FROM evaluations WHERE portfolio_hash = ?`, portfolioHash)
	if err != nil {
		return fmt.Errorf("failed to delete evaluations: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	r.log.Info().
		Str("portfolio_hash", portfolioHash).
		Int64("rows_deleted", rowsAffected).
		Msg("Deleted evaluations")

	return nil
}

// DeleteAllEvaluations deletes all evaluations regardless of portfolio hash.
func (r *PlannerRepository) DeleteAllEvaluations() error {
	result, err := r.db.Exec(`DELETE FROM evaluations`)
	if err != nil {
		return fmt.Errorf("failed to delete all evaluations: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	r.log.Info().
		Int64("rows_deleted", rowsAffected).
		Msg("Deleted all evaluations")

	return nil
}

// GetBestResult retrieves the best result for a portfolio hash.
func (r *PlannerRepository) GetBestResult(portfolioHash string) (*domain.HolisticPlan, error) {
	var record BestResultRecord
	var createdAtUnix, updatedAtUnix sql.NullInt64
	err := r.db.QueryRow(`
		SELECT portfolio_hash, sequence_hash, plan_data, score, created_at, updated_at
		FROM best_result
		WHERE portfolio_hash = ?
	`, portfolioHash).Scan(
		&record.PortfolioHash,
		&record.SequenceHash,
		&record.PlanData,
		&record.Score,
		&createdAtUnix,
		&updatedAtUnix,
	)

	// Note: CreatedAt and UpdatedAt are scanned but not used in domain.HolisticPlan
	// They're kept in BestResultRecord struct for potential future use

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get best result: %w", err)
	}

	var plan domain.HolisticPlan
	if err := json.Unmarshal([]byte(record.PlanData), &plan); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plan: %w", err)
	}

	return &plan, nil
}

// DeleteBestResult deletes the best result for a portfolio hash.
func (r *PlannerRepository) DeleteBestResult(portfolioHash string) error {
	_, err := r.db.Exec(`DELETE FROM best_result WHERE portfolio_hash = ?`, portfolioHash)
	if err != nil {
		return fmt.Errorf("failed to delete best result: %w", err)
	}

	r.log.Info().
		Str("portfolio_hash", portfolioHash).
		Msg("Deleted best result")

	return nil
}

// DeleteAllBestResults deletes all best results regardless of portfolio hash.
func (r *PlannerRepository) DeleteAllBestResults() error {
	result, err := r.db.Exec(`DELETE FROM best_result`)
	if err != nil {
		return fmt.Errorf("failed to delete all best results: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	r.log.Info().
		Int64("rows_deleted", rowsAffected).
		Msg("Deleted all best results")

	return nil
}

// CountSequences returns the total number of sequences for a portfolio hash.
func (r *PlannerRepository) CountSequences(portfolioHash string) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM sequences WHERE portfolio_hash = ?
	`, portfolioHash).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count sequences: %w", err)
	}

	return count, nil
}

// CountPendingSequences returns the number of pending sequences for a portfolio hash.
func (r *PlannerRepository) CountPendingSequences(portfolioHash string) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM sequences WHERE portfolio_hash = ? AND completed = 0
	`, portfolioHash).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count pending sequences: %w", err)
	}

	return count, nil
}

// CountEvaluations returns the total number of evaluations for a portfolio hash.
func (r *PlannerRepository) CountEvaluations(portfolioHash string) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM evaluations WHERE portfolio_hash = ?
	`, portfolioHash).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count evaluations: %w", err)
	}

	return count, nil
}
