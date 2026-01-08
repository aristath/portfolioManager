package queue

import (
	"database/sql"
	"errors"
	"sync"
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
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)

	registry.Register(JobTypePlannerBatch, func(job *Job) error {
		mu.Lock()
		executed = true
		mu.Unlock()
		wg.Done()
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

	// Wait for handler to complete
	wg.Wait()
	// Give a bit more time for RecordExecution to complete
	time.Sleep(50 * time.Millisecond)
	pool.Stop()

	mu.Lock()
	assert.True(t, executed)
	mu.Unlock()

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

	var attempts int
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1) // We'll wait for the final successful attempt

	registry.Register(JobTypePlannerBatch, func(job *Job) error {
		mu.Lock()
		attempts++
		currentAttempt := attempts
		mu.Unlock()

		if currentAttempt < 2 {
			return errors.New("temporary error")
		}
		// Final successful attempt
		wg.Done()
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

	// Wait for final successful attempt
	wg.Wait()
	// Give a bit more time for retry logic to complete
	time.Sleep(100 * time.Millisecond)
	pool.Stop()

	// Should have retried and eventually succeeded
	mu.Lock()
	assert.GreaterOrEqual(t, attempts, 2)
	mu.Unlock()

	// Verify success was recorded
	var lastStatus string
	err := db.QueryRow("SELECT last_status FROM job_history WHERE job_type = ?", JobTypePlannerBatch).
		Scan(&lastStatus)
	require.NoError(t, err)
	assert.Equal(t, "success", lastStatus)
}
