/**
 * Package di provides dependency injection for deployment work registration.
 *
 * Deployment work types handle checking for and executing system updates.
 */
package di

import (
	"fmt"
	"time"

	"github.com/aristath/sentinel/internal/work"
	"github.com/rs/zerolog"
)

// Deployment work type adapters
type deploymentCheckAdapter struct {
	container *Container
}

func (a *deploymentCheckAdapter) CheckForDeployment() error {
	// If deployment is disabled, skip
	if a.container.DeploymentManager == nil {
		return nil
	}

	// Call deployment manager's Deploy method
	result, err := a.container.DeploymentManager.Deploy()
	if err != nil {
		return fmt.Errorf("deployment check failed: %w", err)
	}

	// Log result if deployment actually happened
	// (Deploy() method already logs details, so no need to log "no changes")
	_ = result // Avoid unused variable warning

	return nil
}

func (a *deploymentCheckAdapter) GetCheckInterval() time.Duration {
	// If deployment is disabled, return a long interval (effectively disabled)
	if a.container.DeploymentManager == nil {
		return 24 * time.Hour
	}

	// Read interval from settings database
	intervalMinutes, err := a.container.SettingsRepo.GetFloat("job_auto_deploy_minutes", 5.0)
	if err != nil {
		// GetFloat already logs parse errors, this handles DB errors
		// Fall back to default 5 minutes
		return 5 * time.Minute
	}

	// Convert minutes to duration
	return time.Duration(intervalMinutes) * time.Minute
}

func registerDeploymentWork(registry *work.Registry, container *Container, log zerolog.Logger) {
	deps := &work.DeploymentDeps{
		DeploymentService: &deploymentCheckAdapter{container: container},
	}

	work.RegisterDeploymentWorkTypes(registry, deps)
	log.Debug().Msg("Deployment work types registered")
}
