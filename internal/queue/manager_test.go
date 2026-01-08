package queue

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupManagerTest(t *testing.T) (*Manager, *sql.DB) {
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

	return manager, db
}

func TestManager_Enqueue(t *testing.T) {
	manager, db := setupManagerTest(t)
	defer db.Close()

	job := &Job{
		ID:          "test-1",
		Type:        JobTypePlannerBatch,
		Priority:    PriorityHigh,
		Payload:     map[string]interface{}{"test": "data"},
		CreatedAt:   time.Now(),
		AvailableAt: time.Now(),
		Retries:     0,
		MaxRetries:  3,
	}

	err := manager.Enqueue(job)
	assert.NoError(t, err)
	assert.Equal(t, 1, manager.queue.Size())
}

func TestManager_EnqueueWithHistoryCheck(t *testing.T) {
	manager, db := setupManagerTest(t)
	defer db.Close()

	// Record that job just ran
	err := manager.history.RecordExecution(JobTypePlannerBatch, time.Now(), "success")
	require.NoError(t, err)

	// Try to enqueue with interval check - should not enqueue
	enqueued := manager.EnqueueIfShouldRun(JobTypePlannerBatch, PriorityHigh, 15*time.Minute, map[string]interface{}{})
	assert.False(t, enqueued)
	assert.Equal(t, 0, manager.queue.Size())

	// Record old execution time
	oldTime := time.Now().Add(-16 * time.Minute)
	err = manager.history.RecordExecution(JobTypePlannerBatch, oldTime, "success")
	require.NoError(t, err)

	// Now should enqueue
	enqueued = manager.EnqueueIfShouldRun(JobTypePlannerBatch, PriorityHigh, 15*time.Minute, map[string]interface{}{})
	assert.True(t, enqueued)
	assert.Equal(t, 1, manager.queue.Size())
}

func TestManager_Dequeue(t *testing.T) {
	manager, db := setupManagerTest(t)
	defer db.Close()

	job := &Job{
		ID:          "test-1",
		Type:        JobTypePlannerBatch,
		Priority:    PriorityHigh,
		AvailableAt: time.Now(),
	}

	manager.Enqueue(job)

	dequeued, err := manager.Dequeue()
	assert.NoError(t, err)
	assert.NotNil(t, dequeued)
	assert.Equal(t, "test-1", dequeued.ID)
}
