package queue

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
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

	return db
}

func TestHistory_ShouldRun(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	history := NewHistory(db)

	// First run - should run
	shouldRun := history.ShouldRun(JobTypePlannerBatch, 15*time.Minute)
	assert.True(t, shouldRun)

	// Record execution
	err := history.RecordExecution(JobTypePlannerBatch, time.Now(), "success")
	require.NoError(t, err)

	// Just ran - should not run
	shouldRun = history.ShouldRun(JobTypePlannerBatch, 15*time.Minute)
	assert.False(t, shouldRun)

	// Wait for interval to pass (simulate by recording old time)
	oldTime := time.Now().Add(-16 * time.Minute)
	err = history.RecordExecution(JobTypePlannerBatch, oldTime, "success")
	require.NoError(t, err)

	// Interval passed - should run
	shouldRun = history.ShouldRun(JobTypePlannerBatch, 15*time.Minute)
	assert.True(t, shouldRun)
}

func TestHistory_RecordExecution(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	history := NewHistory(db)

	now := time.Now()
	err := history.RecordExecution(JobTypePlannerBatch, now, "success")
	require.NoError(t, err)

	// Verify it was recorded
	var lastStatus string
	var lastRunAtUnix int64
	err = db.QueryRow("SELECT last_run_at, last_status FROM job_history WHERE job_type = ?", JobTypePlannerBatch).
		Scan(&lastRunAtUnix, &lastStatus)
	require.NoError(t, err)

	assert.Equal(t, "success", lastStatus)

	// Convert Unix timestamp to time.Time and verify it's close
	parsed := time.Unix(lastRunAtUnix, 0).UTC()
	assert.WithinDuration(t, now, parsed, 1*time.Second)
}

func TestHistory_RecordFailure(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	history := NewHistory(db)

	err := history.RecordExecution(JobTypePlannerBatch, time.Now(), "failed")
	require.NoError(t, err)

	var lastStatus string
	err = db.QueryRow("SELECT last_status FROM job_history WHERE job_type = ?", JobTypePlannerBatch).
		Scan(&lastStatus)
	require.NoError(t, err)

	assert.Equal(t, "failed", lastStatus)
}

func TestHistory_DifferentJobTypes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	history := NewHistory(db)

	// Record different job types
	history.RecordExecution(JobTypePlannerBatch, time.Now(), "success")
	history.RecordExecution(JobTypeEventBasedTrading, time.Now().Add(-1*time.Hour), "success")

	// Each should track independently
	assert.False(t, history.ShouldRun(JobTypePlannerBatch, 15*time.Minute))
	assert.True(t, history.ShouldRun(JobTypeEventBasedTrading, 30*time.Minute))
}
