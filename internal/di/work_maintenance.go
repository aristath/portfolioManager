/**
 * Package di provides dependency injection for maintenance work registration.
 *
 * Maintenance work types handle backups, R2 cloud backups, database vacuuming,
 * health checks, and cleanup operations.
 */
package di

import (
	"context"
	"time"

	"github.com/aristath/sentinel/internal/database"
	"github.com/aristath/sentinel/internal/work"
	"github.com/rs/zerolog"
)

// Maintenance adapters
type maintenanceBackupAdapter struct {
	container *Container
}

func (a *maintenanceBackupAdapter) RunDailyBackup() error {
	return a.container.BackupService.DailyBackup()
}

func (a *maintenanceBackupAdapter) BackedUpToday() bool {
	// Check if the last backup was today
	// The BackupService tracks this internally
	return false // Always run to be safe
}

type maintenanceR2BackupAdapter struct {
	container *Container
}

func (a *maintenanceR2BackupAdapter) UploadBackup() error {
	if a.container.R2BackupService != nil {
		return a.container.R2BackupService.CreateAndUploadBackup(context.Background())
	}
	return nil
}

func (a *maintenanceR2BackupAdapter) RotateBackups() error {
	if a.container.R2BackupService != nil {
		// Use default retention of 90 days
		return a.container.R2BackupService.RotateOldBackups(context.Background(), 90)
	}
	return nil
}

type maintenanceVacuumAdapter struct {
	container *Container
	log       zerolog.Logger
}

func (a *maintenanceVacuumAdapter) VacuumDatabases() error {
	// Vacuum ephemeral databases: cache, history, portfolio (following WeeklyMaintenanceJob pattern)
	// Ledger is append-only and should not be vacuumed
	dbs := []struct {
		name string
		db   *database.DB
	}{
		{"cache", a.container.CacheDB},
		{"history", a.container.HistoryDB},
		{"portfolio", a.container.PortfolioDB},
	}

	for _, dbInfo := range dbs {
		if dbInfo.db == nil {
			continue
		}

		// Get size before vacuum
		var pageCount, pageSize int
		dbInfo.db.Conn().QueryRow("PRAGMA page_count").Scan(&pageCount)
		dbInfo.db.Conn().QueryRow("PRAGMA page_size").Scan(&pageSize)
		sizeBefore := float64(pageCount*pageSize) / 1024 / 1024

		// Run VACUUM
		_, err := dbInfo.db.Conn().Exec("VACUUM")
		if err != nil {
			a.log.Error().Err(err).Str("database", dbInfo.name).Msg("VACUUM failed")
			continue // Don't fail the entire operation
		}

		// Get size after vacuum
		dbInfo.db.Conn().QueryRow("PRAGMA page_count").Scan(&pageCount)
		sizeAfter := float64(pageCount*pageSize) / 1024 / 1024
		spaceReclaimed := sizeBefore - sizeAfter

		a.log.Info().
			Str("database", dbInfo.name).
			Float64("size_before_mb", sizeBefore).
			Float64("size_after_mb", sizeAfter).
			Float64("space_reclaimed_mb", spaceReclaimed).
			Msg("VACUUM completed")
	}

	return nil
}

type maintenanceHealthAdapter struct {
	container *Container
}

func (a *maintenanceHealthAdapter) RunHealthChecks() error {
	// Run health checks using CheckAndRecover
	for _, hs := range a.container.HealthServices {
		if err := hs.CheckAndRecover(); err != nil {
			return err
		}
	}
	return nil
}

type maintenanceCleanupAdapter struct {
	container *Container
	log       zerolog.Logger
}

func (a *maintenanceCleanupAdapter) CleanupHistory() error {
	// History database cleanup: run WAL checkpoint to prevent bloat
	// Full history deletion is not implemented as we want to preserve historical data
	if a.container.HistoryDB != nil {
		_, err := a.container.HistoryDB.Conn().Exec("PRAGMA wal_checkpoint(TRUNCATE)")
		if err != nil {
			a.log.Warn().Err(err).Msg("History WAL checkpoint failed")
		} else {
			a.log.Debug().Msg("History WAL checkpoint completed")
		}
	}
	return nil
}

func (a *maintenanceCleanupAdapter) CleanupCache() error {
	// Cache database cleanup: run WAL checkpoint to prevent bloat
	if a.container.CacheDB != nil {
		_, err := a.container.CacheDB.Conn().Exec("PRAGMA wal_checkpoint(TRUNCATE)")
		if err != nil {
			a.log.Warn().Err(err).Msg("Cache WAL checkpoint failed")
		} else {
			a.log.Debug().Msg("Cache WAL checkpoint completed")
		}
	}
	return nil
}

func (a *maintenanceCleanupAdapter) CleanupRecommendations() error {
	// Delete recommendations older than 7 days
	_, err := a.container.RecommendationRepo.DeleteOlderThan(7 * 24 * time.Hour)
	return err
}

func (a *maintenanceCleanupAdapter) CleanupClientData() error {
	_, err := a.container.ClientDataRepo.DeleteAllExpired()
	return err
}

func registerMaintenanceWork(registry *work.Registry, container *Container, log zerolog.Logger) {
	deps := &work.MaintenanceDeps{
		BackupService:      &maintenanceBackupAdapter{container: container},
		R2BackupService:    &maintenanceR2BackupAdapter{container: container},
		VacuumService:      &maintenanceVacuumAdapter{container: container, log: log},
		HealthCheckService: &maintenanceHealthAdapter{container: container},
		CleanupService:     &maintenanceCleanupAdapter{container: container, log: log},
	}

	work.RegisterMaintenanceWorkTypes(registry, deps)
	log.Debug().Msg("Maintenance work types registered")
}
