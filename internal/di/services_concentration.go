/**
 * Package di provides dependency injection for concentration alert service initialization.
 *
 * Step 16: Initialize Concentration Alert Service
 * Concentration alert service detects portfolio concentration breaches.
 */
package di

import (
	"github.com/aristath/sentinel/internal/modules/allocation"
	"github.com/rs/zerolog"
)

// initializeConcentrationAlertService initializes the concentration alert service.
func initializeConcentrationAlertService(container *Container, log zerolog.Logger) error {
	container.ConcentrationAlertService = allocation.NewConcentrationAlertService(
		container.PortfolioDB.Conn(),
		log,
	)

	return nil
}
