package queue

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupSchedulerTest(t *testing.T) (*Scheduler, *Manager, *sql.DB) {
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

	scheduler := NewScheduler(manager)

	return scheduler, manager, db
}

func TestScheduler_EnqueueTimeBasedJob(t *testing.T) {
	scheduler, manager, db := setupSchedulerTest(t)
	defer db.Close()

	// Enqueue a job that should run (never run before)
	enqueued := scheduler.enqueueTimeBasedJob(JobTypeHealthCheck, PriorityMedium, 24*time.Hour)
	assert.True(t, enqueued)
	assert.Equal(t, 1, manager.Size())

	// Record execution
	err := manager.RecordExecution(JobTypeHealthCheck, "success")
	require.NoError(t, err)

	// Try again - should not enqueue (interval not passed)
	enqueued = scheduler.enqueueTimeBasedJob(JobTypeHealthCheck, PriorityMedium, 24*time.Hour)
	assert.False(t, enqueued)
	assert.Equal(t, 1, manager.Size()) // Still 1 from before
}

func TestScheduler_StartStop(t *testing.T) {
	scheduler, _, _ := setupSchedulerTest(t)

	// Should start without error
	scheduler.Start()

	// Give it a moment
	time.Sleep(100 * time.Millisecond)

	// Should stop without error
	scheduler.Stop()

	// Give it time to stop
	time.Sleep(100 * time.Millisecond)
}
