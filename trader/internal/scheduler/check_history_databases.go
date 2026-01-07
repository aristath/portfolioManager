package scheduler

import (
	"database/sql"
	"fmt"

	"github.com/aristath/sentinel/internal/database"
	"github.com/rs/zerolog"
)

// CheckHistoryDatabasesJob verifies integrity of consolidated history database
type CheckHistoryDatabasesJob struct {
	log       zerolog.Logger
	historyDB *database.DB
}

// NewCheckHistoryDatabasesJob creates a new CheckHistoryDatabasesJob
func NewCheckHistoryDatabasesJob(historyDB *database.DB) *CheckHistoryDatabasesJob {
	return &CheckHistoryDatabasesJob{
		log:       zerolog.Nop(),
		historyDB: historyDB,
	}
}

// SetLogger sets the logger for the job
func (j *CheckHistoryDatabasesJob) SetLogger(log zerolog.Logger) {
	j.log = log
}

// Name returns the job name
func (j *CheckHistoryDatabasesJob) Name() string {
	return "check_history_databases"
}

// Run executes the check history databases job
func (j *CheckHistoryDatabasesJob) Run() error {
	if j.historyDB == nil {
		j.log.Debug().Msg("History database not configured, skipping history database checks")
		return nil
	}

	// Check integrity of consolidated history.db
	db := j.historyDB.Conn()
	if err := j.checkDatabaseIntegrity("history", db); err != nil {
		// Consolidated history database corruption is critical - log error but don't delete
		// The database contains all historical data and should not be auto-deleted
		j.log.Error().
			Err(err).
			Msg("History database integrity check failed - manual intervention required")
		return fmt.Errorf("history database integrity check failed: %w", err)
	}

	j.log.Info().Msg("History database integrity OK")
	return nil
}

// checkDatabaseIntegrity runs SQLite's PRAGMA integrity_check
func (j *CheckHistoryDatabasesJob) checkDatabaseIntegrity(name string, db *sql.DB) error {
	var result string
	err := db.QueryRow("PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return fmt.Errorf("integrity check failed: %w", err)
	}

	if result != "ok" {
		return fmt.Errorf("integrity check returned: %s", result)
	}

	return nil
}
