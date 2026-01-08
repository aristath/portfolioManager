package scheduler

import (
	"context"
	"fmt"

	"github.com/aristath/sentinel/internal/modules/settings"
	"github.com/aristath/sentinel/internal/reliability"
	"github.com/rs/zerolog"
)

// R2BackupRotationJobConfig holds configuration for R2 backup rotation job
type R2BackupRotationJobConfig struct {
	Log             zerolog.Logger
	Service         *reliability.R2BackupService
	SettingsService *settings.Service
}

// R2BackupRotationJob performs R2 backup rotation (deletes old backups)
type R2BackupRotationJob struct {
	log             zerolog.Logger
	service         *reliability.R2BackupService
	settingsService *settings.Service
}

// NewR2BackupRotationJob creates a new R2 backup rotation job
func NewR2BackupRotationJob(cfg R2BackupRotationJobConfig) *R2BackupRotationJob {
	return &R2BackupRotationJob{
		log:             cfg.Log.With().Str("job", "r2_backup_rotation").Logger(),
		service:         cfg.Service,
		settingsService: cfg.SettingsService,
	}
}

// Name returns the job name
func (j *R2BackupRotationJob) Name() string {
	return "r2_backup_rotation"
}

// Run executes the R2 backup rotation job
func (j *R2BackupRotationJob) Run() error {
	// Check if R2 backups are enabled
	enabledValue, err := j.settingsService.Get("r2_backup_enabled")
	if err != nil || enabledValue == nil {
		j.log.Debug().Msg("R2 backups not enabled, skipping rotation")
		return nil
	}

	enabled := false
	if floatVal, ok := enabledValue.(float64); ok {
		enabled = floatVal == 1.0
	}

	if !enabled {
		j.log.Debug().Msg("R2 backups not enabled, skipping rotation")
		return nil
	}

	j.log.Info().Msg("Starting R2 backup rotation job")

	ctx := context.Background()

	// Get retention days from settings
	retentionValue, err := j.settingsService.Get("r2_backup_retention_days")
	if err != nil {
		j.log.Warn().Err(err).Msg("Failed to get retention days, using default 90")
		retentionValue = 90.0
	}

	retentionDays := 90 // default
	if floatVal, ok := retentionValue.(float64); ok {
		retentionDays = int(floatVal)
	}

	if err := j.service.RotateOldBackups(ctx, retentionDays); err != nil {
		j.log.Error().Err(err).Msg("R2 backup rotation failed")
		return fmt.Errorf("r2 backup rotation failed: %w", err)
	}

	j.log.Info().Msg("R2 backup rotation job completed successfully")
	return nil
}
