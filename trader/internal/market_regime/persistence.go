package market_regime

import (
	"database/sql"
	"time"

	"github.com/rs/zerolog"
)

// RegimePersistence handles regime score history and smoothing
type RegimePersistence struct {
	db  *sql.DB
	log zerolog.Logger
	// Smoothing parameters
	smoothingAlpha float64 // Exponential moving average alpha (default 0.1)
}

// RegimeHistoryEntry represents a single regime score record
type RegimeHistoryEntry struct {
	ID            int64
	RecordedAt    time.Time
	RawScore      float64 // Raw score before smoothing
	SmoothedScore float64 // Smoothed score (EMA)
}

// NewRegimePersistence creates a new regime persistence manager
func NewRegimePersistence(db *sql.DB, log zerolog.Logger) *RegimePersistence {
	return &RegimePersistence{
		db:             db,
		log:            log.With().Str("component", "regime_persistence").Logger(),
		smoothingAlpha: 0.1, // Default: slow adaptation (matches slow-growth strategy)
	}
}

// GetCurrentRegimeScore returns the current smoothed regime score
func (rp *RegimePersistence) GetCurrentRegimeScore() (MarketRegimeScore, error) {
	// Order by id DESC to ensure we get the most recent record
	// This is more reliable than ordering by recorded_at when stored as TEXT
	query := `SELECT smoothed_score FROM market_regime_history
	          ORDER BY id DESC LIMIT 1`

	var score float64
	err := rp.db.QueryRow(query).Scan(&score)
	if err == sql.ErrNoRows {
		// No history yet, return neutral
		return NeutralScore, nil
	}
	if err != nil {
		return NeutralScore, err
	}

	return MarketRegimeScore(score), nil
}

// RecordRegimeScore records a raw regime score and calculates smoothed score
func (rp *RegimePersistence) RecordRegimeScore(rawScore MarketRegimeScore) error {
	// Get last smoothed score
	lastSmoothed, err := rp.GetCurrentRegimeScore()
	if err != nil {
		return err
	}

	// Apply smoothing
	// If lastSmoothed is NeutralScore (0.0) and we have no history, use current score directly
	// Otherwise, apply exponential moving average
	var smoothed float64
	if lastSmoothed == NeutralScore {
		// Check if we actually have history
		var count int
		err := rp.db.QueryRow("SELECT COUNT(*) FROM market_regime_history").Scan(&count)
		if err == nil && count == 0 {
			// No history - use raw score directly (first entry)
			smoothed = float64(rawScore)
		} else {
			// We have history but last smoothed is 0.0 - apply smoothing
			smoothed = rp.ApplySmoothing(float64(rawScore), float64(lastSmoothed), rp.smoothingAlpha)
		}
	} else {
		// We have a previous smoothed score - apply EMA
		// Formula: smoothed = alpha * current + (1 - alpha) * lastSmoothed
		smoothed = rp.smoothingAlpha*float64(rawScore) + (1.0-rp.smoothingAlpha)*float64(lastSmoothed)
		// Ensure we're actually applying smoothing (debug check)
		if smoothed == float64(lastSmoothed) && float64(rawScore) != float64(lastSmoothed) {
			// This should never happen if smoothing is working correctly
			rp.log.Warn().
				Float64("rawScore", float64(rawScore)).
				Float64("lastSmoothed", float64(lastSmoothed)).
				Float64("alpha", rp.smoothingAlpha).
				Float64("calculated", smoothed).
				Msg("Smoothing calculation resulted in same value as lastSmoothed")
		}
	}

	// Discrete regime support is intentionally removed. Keep column populated for schema compatibility.
	discrete := "n/a"

	// Insert into database
	query := `INSERT INTO market_regime_history
	          (recorded_at, raw_score, smoothed_score, discrete_regime)
	          VALUES (?, ?, ?, ?)`

	_, err = rp.db.Exec(query, time.Now().Unix(), float64(rawScore), smoothed, discrete)
	if err != nil {
		return err
	}

	rp.log.Debug().
		Float64("raw_score", float64(rawScore)).
		Float64("smoothed_score", smoothed).
		Str("discrete_regime", discrete).
		Msg("Recorded regime score")

	return nil
}

// GetSmoothedScore returns the exponentially smoothed score
// This is a convenience method that wraps GetCurrentRegimeScore
func (rp *RegimePersistence) GetSmoothedScore() (MarketRegimeScore, error) {
	return rp.GetCurrentRegimeScore()
}

// GetScoreChange calculates the change in regime score from the last check
func (rp *RegimePersistence) GetScoreChange() (float64, error) {
	query := `SELECT smoothed_score FROM market_regime_history
	          ORDER BY recorded_at DESC LIMIT 2`

	rows, err := rp.db.Query(query)
	if err != nil {
		return 0.0, err
	}
	defer rows.Close()

	var scores []float64
	for rows.Next() {
		var score float64
		if err := rows.Scan(&score); err != nil {
			return 0.0, err
		}
		scores = append(scores, score)
	}

	if len(scores) < 2 {
		// Not enough history
		return 0.0, nil
	}

	// Calculate change (current - previous)
	change := scores[0] - scores[1]
	return change, nil
}

// GetRegimeHistory returns recent regime history entries
func (rp *RegimePersistence) GetRegimeHistory(limit int) ([]RegimeHistoryEntry, error) {
	query := `SELECT id, recorded_at, raw_score, smoothed_score
	          FROM market_regime_history
	          ORDER BY recorded_at DESC
	          LIMIT ?`

	rows, err := rp.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []RegimeHistoryEntry
	for rows.Next() {
		var entry RegimeHistoryEntry
		var recordedAtUnix sql.NullInt64

		if err := rows.Scan(
			&entry.ID,
			&recordedAtUnix,
			&entry.RawScore,
			&entry.SmoothedScore,
		); err != nil {
			return nil, err
		}

		if recordedAtUnix.Valid {
			entry.RecordedAt = time.Unix(recordedAtUnix.Int64, 0).UTC()
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// ApplySmoothing applies exponential moving average smoothing
// This is a helper method that can be used independently
func (rp *RegimePersistence) ApplySmoothing(currentScore, lastSmoothed, alpha float64) float64 {
	detector := NewMarketRegimeDetector(rp.log)
	return detector.ApplySmoothing(currentScore, lastSmoothed, alpha)
}


