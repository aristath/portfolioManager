package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/aristath/sentinel/internal/modules/settings"
	"github.com/aristath/sentinel/internal/reliability"
	"github.com/rs/zerolog"
)

// R2BackupJobConfig holds configuration for R2 backup job
type R2BackupJobConfig struct {
	Log             zerolog.Logger
	Service         *reliability.R2BackupService
	SettingsService *settings.Service
}

// R2BackupJob performs R2 cloud backups
type R2BackupJob struct {
	JobBase
	log             zerolog.Logger
	service         *reliability.R2BackupService
	settingsService *settings.Service
	lastBackup      time.Time // Track last backup time for schedule checking
}

// NewR2BackupJob creates a new R2 backup job
func NewR2BackupJob(cfg R2BackupJobConfig) *R2BackupJob {
	return &R2BackupJob{
		log:             cfg.Log.With().Str("job", "r2_backup").Logger(),
		service:         cfg.Service,
		settingsService: cfg.SettingsService,
		lastBackup:      time.Time{}, // Zero time, will trigger first backup
	}
}

// Name returns the job name
func (j *R2BackupJob) Name() string {
	return "r2_backup"
}

// Run executes the R2 backup job
func (j *R2BackupJob) Run() error {
	// Check if R2 backups are enabled
	enabledValue, err := j.settingsService.Get("r2_backup_enabled")
	if err != nil || enabledValue == nil {
		j.log.Debug().Msg("R2 backups not enabled, skipping")
		return nil
	}

	enabled := false
	if floatVal, ok := enabledValue.(float64); ok {
		enabled = floatVal == 1.0
	}

	if !enabled {
		j.log.Debug().Msg("R2 backups not enabled, skipping")
		return nil
	}

	// Check if backup should run based on schedule
	if !j.shouldRunNow() {
		j.log.Debug().Msg("Not scheduled to run now, skipping")
		return nil
	}

	j.log.Info().Msg("Starting R2 backup job")

	ctx := context.Background()

	if err := j.service.CreateAndUploadBackup(ctx); err != nil {
		j.log.Error().Err(err).Msg("R2 backup failed")
		return fmt.Errorf("r2 backup failed: %w", err)
	}

	// Update last backup time
	j.lastBackup = time.Now()

	j.log.Info().Msg("R2 backup job completed successfully")
	return nil
}

// shouldRunNow determines if backup should run based on schedule setting
func (j *R2BackupJob) shouldRunNow() bool {
	scheduleValue, err := j.settingsService.Get("r2_backup_schedule")
	if err != nil || scheduleValue == nil {
		// Default to daily if not set
		return j.shouldRunDaily()
	}

	schedule, ok := scheduleValue.(string)
	if !ok {
		// Default to daily if invalid
		return j.shouldRunDaily()
	}

	now := time.Now()

	switch schedule {
	case "daily":
		return j.shouldRunDaily()
	case "weekly":
		// Run on Sundays
		return now.Weekday() == time.Sunday && j.hasBeenMoreThan(24*time.Hour)
	case "monthly":
		// Run on the 1st of the month
		return now.Day() == 1 && j.hasBeenMoreThan(24*time.Hour)
	default:
		// Unknown schedule, default to daily
		j.log.Warn().Str("schedule", schedule).Msg("Unknown backup schedule, defaulting to daily")
		return j.shouldRunDaily()
	}
}

// shouldRunDaily checks if daily backup should run (once per day)
func (j *R2BackupJob) shouldRunDaily() bool {
	return j.hasBeenMoreThan(23 * time.Hour) // Allow some margin
}

// hasBeenMoreThan checks if enough time has passed since last backup
func (j *R2BackupJob) hasBeenMoreThan(duration time.Duration) bool {
	if j.lastBackup.IsZero() {
		return true // Never backed up, should run
	}
	return time.Since(j.lastBackup) > duration
}
