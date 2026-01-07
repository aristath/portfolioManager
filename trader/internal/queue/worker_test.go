package queue

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupWorkerTest(t *testing.T) (*WorkerPool, *Manager, *Registry, *sql.DB) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS job_history (
			job_type TEXT PRIMARY KEY,
			last_run_at TEXT NOT NULL,
			last_status TEXT NOT NULL DEFAULT 'success'
		)
	`)
	require.NoError(t, err)

	queue := NewMemoryQueue()
	history := NewHistory(db)
	manager := NewManager(queue, history)
	registry := NewRegistry()

	pool := NewWorkerPool(manager, registry, 2)

	return pool, manager, registry, db
}

func TestWorkerPool_ProcessJob(t *testing.T) {
	pool, manager, registry, db := setupWorkerTest(t)
	defer db.Close()

	var executed bool
	registry.Register(JobTypePlannerBatch, func(job *Job) error {
		executed = true
		return nil
	})

	job := &Job{
		ID:          "test-1",
		Type:        JobTypePlannerBatch,
		Priority:    PriorityHigh,
		AvailableAt: time.Now(),
	}

	manager.Enqueue(job)
	pool.Start()

	// Give worker time to process
	time.Sleep(200 * time.Millisecond)
	pool.Stop()

	assert.True(t, executed)

	// Verify execution was recorded
	var lastStatus string
	err := db.QueryRow("SELECT last_status FROM job_history WHERE job_type = ?", JobTypePlannerBatch).
		Scan(&lastStatus)
	require.NoError(t, err)
	assert.Equal(t, "success", lastStatus)
}

func TestWorkerPool_ProcessJobFailure(t *testing.T) {
	pool, manager, registry, db := setupWorkerTest(t)
	defer db.Close()

	registry.Register(JobTypePlannerBatch, func(job *Job) error {
		return errors.New("test error")
	})

	job := &Job{
		ID:          "test-1",
		Type:        JobTypePlannerBatch,
		Priority:    PriorityHigh,
		AvailableAt: time.Now(),
		Retries:     3, // Already at max retries, so it will record failure
		MaxRetries:  3,
	}

	manager.Enqueue(job)
	pool.Start()

	// Give worker time to process
	time.Sleep(200 * time.Millisecond)
	pool.Stop()

	// Verify failure was recorded
	var lastStatus string
	err := db.QueryRow("SELECT last_status FROM job_history WHERE job_type = ?", JobTypePlannerBatch).
		Scan(&lastStatus)
	require.NoError(t, err)
	assert.Equal(t, "failed", lastStatus)
}

func TestWorkerPool_RetryOnFailure(t *testing.T) {
	pool, manager, registry, db := setupWorkerTest(t)
	defer db.Close()

	attempts := 0
	registry.Register(JobTypePlannerBatch, func(job *Job) error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary error")
		}
		return nil
	})

	job := &Job{
		ID:          "test-1",
		Type:        JobTypePlannerBatch,
		Priority:    PriorityHigh,
		AvailableAt: time.Now(),
		Retries:     0,
		MaxRetries:  3,
	}

	manager.Enqueue(job)
	pool.Start()

	// Give worker time to process, retry, and succeed
	time.Sleep(2 * time.Second)
	pool.Stop()

	// Should have retried and eventually succeeded
	assert.GreaterOrEqual(t, attempts, 2)

	// Verify success was recorded
	var lastStatus string
	err := db.QueryRow("SELECT last_status FROM job_history WHERE job_type = ?", JobTypePlannerBatch).
		Scan(&lastStatus)
	require.NoError(t, err)
	assert.Equal(t, "success", lastStatus)
}
