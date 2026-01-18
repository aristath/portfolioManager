/**
 * Package di provides dependency injection for callback initialization.
 *
 * Step 18: Initialize Callbacks (for jobs)
 * Callbacks are functions that jobs can call to trigger actions.
 */
package di

import (
	"fmt"

	"github.com/aristath/sentinel/internal/modules/display"
	"github.com/rs/zerolog"
)

// initializeCallbacks initializes callback functions for jobs.
func initializeCallbacks(container *Container, displayManager *display.StateManager, log zerolog.Logger) {
	// Display ticker update callback (called by sync cycle)
	// Updates LED ticker text with current portfolio information
	container.UpdateDisplayTicker = func() error {
		text, err := container.TickerContentService.GenerateTickerText()
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate ticker text")
			return err
		}

		displayManager.SetText(text)

		log.Debug().
			Str("ticker_text", text).
			Msg("Updated display ticker")

		return nil
	}

	// Emergency rebalance callback (called when negative balance detected)
	// Triggers emergency rebalancing to correct negative cash balances
	container.EmergencyRebalance = func() error {
		log.Warn().Msg("EMERGENCY: Executing negative balance rebalancing")

		success, err := container.NegativeBalanceRebalancer.RebalanceNegativeBalances()
		if err != nil {
			return fmt.Errorf("emergency rebalancing failed: %w", err)
		}

		if !success {
			log.Warn().Msg("Emergency rebalancing completed but some issues may remain")
		} else {
			log.Info().Msg("Emergency rebalancing completed successfully")
		}

		return nil
	}
}
