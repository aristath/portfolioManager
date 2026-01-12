package repository

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/rs/zerolog"
)

type InMemoryPlannerRepository struct {
	sequences   map[string]*SequenceRecord
	evaluations map[string]*EvaluationRecord
	bestResults map[string]*BestResultRecord
	mu          sync.RWMutex
	log         zerolog.Logger
}

func NewInMemoryPlannerRepository(log zerolog.Logger) *InMemoryPlannerRepository {
	return &InMemoryPlannerRepository{
		sequences:   make(map[string]*SequenceRecord),
		evaluations: make(map[string]*EvaluationRecord),
		bestResults: make(map[string]*BestResultRecord),
		log:         log.With().Str("component", "planner_repository_inmemory").Logger(),
	}
}

func makeKey(sequenceHash, portfolioHash string) string {
	return fmt.Sprintf("%s:%s", sequenceHash, portfolioHash)
}

func (r *InMemoryPlannerRepository) GetSequence(sequenceHash, portfolioHash string) (*domain.ActionSequence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := makeKey(sequenceHash, portfolioHash)
	record, exists := r.sequences[key]
	if !exists {
		return nil, nil
	}

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

func (r *InMemoryPlannerRepository) ListSequencesByPortfolioHash(portfolioHash string, limit int) ([]SequenceRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records := make([]SequenceRecord, 0)
	for _, seq := range r.sequences {
		if seq.PortfolioHash == portfolioHash {
			records = append(records, *seq)
		}
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].Priority == records[j].Priority {
			return records[i].CreatedAt.After(records[j].CreatedAt)
		}
		return records[i].Priority > records[j].Priority
	})

	if limit > 0 && len(records) > limit {
		records = records[:limit]
	}

	return records, nil
}

func (r *InMemoryPlannerRepository) GetPendingSequences(portfolioHash string, limit int) ([]SequenceRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pending := make([]SequenceRecord, 0)
	for _, seq := range r.sequences {
		if seq.PortfolioHash == portfolioHash && !seq.Completed {
			pending = append(pending, *seq)
		}
	}

	sort.Slice(pending, func(i, j int) bool {
		if pending[i].Priority == pending[j].Priority {
			return pending[i].CreatedAt.Before(pending[j].CreatedAt)
		}
		return pending[i].Priority > pending[j].Priority
	})

	if limit > 0 && len(pending) > limit {
		pending = pending[:limit]
	}

	return pending, nil
}

func (r *InMemoryPlannerRepository) MarkSequenceCompleted(sequenceHash, portfolioHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := makeKey(sequenceHash, portfolioHash)
	seq, exists := r.sequences[key]
	if !exists {
		return fmt.Errorf("sequence not found: %s:%s", sequenceHash, portfolioHash)
	}

	now := time.Now().UTC()
	seq.Completed = true
	seq.EvaluatedAt = &now

	return nil
}

func (r *InMemoryPlannerRepository) DeleteSequencesByPortfolioHash(portfolioHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for key, seq := range r.sequences {
		if seq.PortfolioHash == portfolioHash {
			delete(r.sequences, key)
			count++
		}
	}

	r.log.Info().Str("portfolio_hash", portfolioHash).Int("rows_deleted", count).Msg("Deleted sequences")
	return nil
}

func (r *InMemoryPlannerRepository) DeleteAllSequences() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := len(r.sequences)
	r.sequences = make(map[string]*SequenceRecord)

	r.log.Info().Int("rows_deleted", count).Msg("Deleted all sequences")
	return nil
}

func (r *InMemoryPlannerRepository) CountSequences(portfolioHash string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, seq := range r.sequences {
		if seq.PortfolioHash == portfolioHash {
			count++
		}
	}

	return count, nil
}

func (r *InMemoryPlannerRepository) CountPendingSequences(portfolioHash string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, seq := range r.sequences {
		if seq.PortfolioHash == portfolioHash && !seq.Completed {
			count++
		}
	}

	return count, nil
}

func (r *InMemoryPlannerRepository) GetEvaluation(sequenceHash, portfolioHash string) (*domain.EvaluationResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := makeKey(sequenceHash, portfolioHash)
	record, exists := r.evaluations[key]
	if !exists {
		return nil, nil
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

func (r *InMemoryPlannerRepository) ListEvaluationsByPortfolioHash(portfolioHash string) ([]EvaluationRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records := make([]EvaluationRecord, 0)
	for _, eval := range r.evaluations {
		if eval.PortfolioHash == portfolioHash {
			records = append(records, *eval)
		}
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].EndScore == records[j].EndScore {
			return records[i].EvaluatedAt.After(records[j].EvaluatedAt)
		}
		return records[i].EndScore > records[j].EndScore
	})

	return records, nil
}

func (r *InMemoryPlannerRepository) DeleteEvaluationsByPortfolioHash(portfolioHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for key, eval := range r.evaluations {
		if eval.PortfolioHash == portfolioHash {
			delete(r.evaluations, key)
			count++
		}
	}

	r.log.Info().Str("portfolio_hash", portfolioHash).Int("rows_deleted", count).Msg("Deleted evaluations")
	return nil
}

func (r *InMemoryPlannerRepository) DeleteAllEvaluations() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := len(r.evaluations)
	r.evaluations = make(map[string]*EvaluationRecord)

	r.log.Info().Int("rows_deleted", count).Msg("Deleted all evaluations")
	return nil
}

func (r *InMemoryPlannerRepository) CountEvaluations(portfolioHash string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, eval := range r.evaluations {
		if eval.PortfolioHash == portfolioHash {
			count++
		}
	}

	return count, nil
}

func (r *InMemoryPlannerRepository) GetBestResult(portfolioHash string) (*domain.HolisticPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	record, exists := r.bestResults[portfolioHash]
	if !exists {
		return nil, nil
	}

	var plan domain.HolisticPlan
	if err := json.Unmarshal([]byte(record.PlanData), &plan); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plan: %w", err)
	}

	return &plan, nil
}

func (r *InMemoryPlannerRepository) DeleteBestResult(portfolioHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.bestResults, portfolioHash)

	r.log.Info().Str("portfolio_hash", portfolioHash).Msg("Deleted best result")
	return nil
}

func (r *InMemoryPlannerRepository) DeleteAllBestResults() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := len(r.bestResults)
	r.bestResults = make(map[string]*BestResultRecord)

	r.log.Info().Int("rows_deleted", count).Msg("Deleted all best results")
	return nil
}
