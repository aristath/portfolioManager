package queue

import (
	"database/sql"
	"fmt"
	"time"
)

// History tracks job execution history
type History struct {
	db *sql.DB
}

// NewHistory creates a new job history tracker
func NewHistory(db *sql.DB) *History {
	return &History{db: db}
}

// ShouldRun checks if a job should run based on its last execution time and interval
func (h *History) ShouldRun(jobType JobType, interval time.Duration) bool {
	if h.db == nil {
		// No database - should run (fallback behavior)
		return true
	}

	var lastRunAtUnix sql.NullInt64
	err := h.db.QueryRow(
		"SELECT last_run_at FROM job_history WHERE job_type = ?",
		string(jobType),
	).Scan(&lastRunAtUnix)

	if err == sql.ErrNoRows {
		// Never run before - should run
		return true
	}
	if err != nil {
		// Error querying - err on side of running
		return true
	}

	if !lastRunAtUnix.Valid {
		// NULL value - should run
		return true
	}

	lastRunAt := time.Unix(lastRunAtUnix.Int64, 0).UTC()

	nextRun := lastRunAt.Add(interval)
	return time.Now().After(nextRun)
}

// RecordExecution records a job execution
func (h *History) RecordExecution(jobType JobType, timestamp time.Time, status string) error {
	if h.db == nil {
		// No database - silently succeed (for testing)
		return nil
	}

	lastRunAt := timestamp.Unix()

	_, err := h.db.Exec(`
		INSERT INTO job_history (job_type, last_run_at, last_status)
		VALUES (?, ?, ?)
		ON CONFLICT(job_type) DO UPDATE SET
			last_run_at = excluded.last_run_at,
			last_status = excluded.last_status
	`, string(jobType), lastRunAt, status)

	if err != nil {
		return fmt.Errorf("failed to record job execution: %w", err)
	}

	return nil
}

// JobHistoryEntry represents a job history entry
type JobHistoryEntry struct {
	JobType   string    `json:"job_type"`
	LastRunAt time.Time `json:"last_run_at"`
	Status    string    `json:"status"`
}

// GetAllHistory returns all job history entries
func (h *History) GetAllHistory() ([]JobHistoryEntry, error) {
	if h.db == nil {
		return []JobHistoryEntry{}, nil
	}

	rows, err := h.db.Query(`
		SELECT job_type, last_run_at, last_status
		FROM job_history
		ORDER BY last_run_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query job history: %w", err)
	}
	defer rows.Close()

	var entries []JobHistoryEntry
	for rows.Next() {
		var entry JobHistoryEntry
		var lastRunAtUnix int64
		if err := rows.Scan(&entry.JobType, &lastRunAtUnix, &entry.Status); err != nil {
			return nil, fmt.Errorf("failed to scan job history row: %w", err)
		}
		entry.LastRunAt = time.Unix(lastRunAtUnix, 0).UTC()
		entries = append(entries, entry)
	}

	return entries, nil
}
