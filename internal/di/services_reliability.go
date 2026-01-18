/**
 * Package di provides dependency injection for reliability service initialization.
 *
 * Step 15: Initialize Reliability Services
 * Reliability services handle backups, health checks, and data integrity.
 */
package di

import (
	"github.com/aristath/sentinel/internal/config"
	"github.com/aristath/sentinel/internal/database"
	"github.com/aristath/sentinel/internal/reliability"
	"github.com/rs/zerolog"
)

// initializeReliabilityServices initializes reliability-related services.
func initializeReliabilityServices(container *Container, cfg *config.Config, log zerolog.Logger) error {
	// Create all database references map for reliability services
	// Health services monitor database integrity and file size
	databases := map[string]*database.DB{
		"universe":  container.UniverseDB,
		"config":    container.ConfigDB,
		"ledger":    container.LedgerDB,
		"portfolio": container.PortfolioDB,
		"history":   container.HistoryDB,
		"cache":     container.CacheDB,
	}

	// Initialize health services for each database
	// Health services check database integrity, file size, and corruption
	dataDir := cfg.DataDir
	container.HealthServices = make(map[string]*reliability.DatabaseHealthService)
	container.HealthServices["universe"] = reliability.NewDatabaseHealthService(container.UniverseDB, "universe", dataDir+"/universe.db", log)
	container.HealthServices["config"] = reliability.NewDatabaseHealthService(container.ConfigDB, "config", dataDir+"/config.db", log)
	container.HealthServices["ledger"] = reliability.NewDatabaseHealthService(container.LedgerDB, "ledger", dataDir+"/ledger.db", log)
	container.HealthServices["portfolio"] = reliability.NewDatabaseHealthService(container.PortfolioDB, "portfolio", dataDir+"/portfolio.db", log)
	container.HealthServices["history"] = reliability.NewDatabaseHealthService(container.HistoryDB, "history", dataDir+"/history.db", log)
	container.HealthServices["cache"] = reliability.NewDatabaseHealthService(container.CacheDB, "cache", dataDir+"/cache.db", log)

	// Initialize backup service
	// Creates local backups of all databases
	backupDir := dataDir + "/backups"
	container.BackupService = reliability.NewBackupService(databases, dataDir, backupDir, log)

	// Initialize R2 cloud backup services (optional - only if credentials are configured)
	// R2 backup provides cloud storage for database backups
	r2AccountID := ""
	r2AccessKeyID := ""
	r2SecretAccessKey := ""
	r2BucketName := ""

	if container.SettingsRepo != nil {
		if val, err := container.SettingsRepo.Get("r2_account_id"); err == nil && val != nil {
			r2AccountID = *val
		}
		if val, err := container.SettingsRepo.Get("r2_access_key_id"); err == nil && val != nil {
			r2AccessKeyID = *val
		}
		if val, err := container.SettingsRepo.Get("r2_secret_access_key"); err == nil && val != nil {
			r2SecretAccessKey = *val
		}
		if val, err := container.SettingsRepo.Get("r2_bucket_name"); err == nil && val != nil {
			r2BucketName = *val
		}
	}

	// Only initialize R2 services if all credentials are provided
	// R2 backup is optional - system works without it
	if r2AccountID != "" && r2AccessKeyID != "" && r2SecretAccessKey != "" && r2BucketName != "" {
		r2Client, err := reliability.NewR2Client(r2AccountID, r2AccessKeyID, r2SecretAccessKey, r2BucketName, log)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to initialize R2 client - R2 backup disabled")
		} else {
			container.R2Client = r2Client
			container.R2BackupService = reliability.NewR2BackupService(
				r2Client,
				container.BackupService,
				dataDir,
				log,
			)
			container.RestoreService = reliability.NewRestoreService(r2Client, dataDir, log)
			log.Info().Msg("R2 cloud backup services initialized")
		}
	} else {
		log.Debug().Msg("R2 credentials not configured - R2 backup disabled")
	}

	return nil
}
