/**
 * Package di provides dependency injection for work processor event triggers.
 *
 * Event triggers connect system events to work processor execution, ensuring
 * that work is triggered when relevant events occur (state changes, recommendations,
 * market status changes, dividend detection).
 */
package di

import (
	"github.com/aristath/sentinel/internal/events"
	"github.com/aristath/sentinel/internal/work"
	"github.com/rs/zerolog"
)

// registerTriggers registers event listeners that trigger work processor execution
func registerTriggers(container *Container, processor *work.Processor, workCache *work.Cache, cache *workCache, log zerolog.Logger) {
	bus := container.EventBus

	// StateChanged -> Clear planner cache and trigger
	bus.Subscribe(events.StateChanged, func(e *events.Event) {
		cache.DeletePrefix("planner:")
		cache.DeletePrefix("optimizer_weights")
		cache.DeletePrefix("opportunity_context")
		cache.DeletePrefix("trade_plan")
		if err := workCache.DeleteByPrefix("planner:"); err != nil {
			log.Warn().Err(err).Msg("Failed to clear planner work cache")
		}
		processor.Trigger()
	})

	// RecommendationsReady -> trigger trading
	bus.Subscribe(events.RecommendationsReady, func(e *events.Event) {
		processor.Trigger()
	})

	// MarketsStatusChanged -> Trigger to check market-timed work
	bus.Subscribe(events.MarketsStatusChanged, func(e *events.Event) {
		processor.Trigger()
	})

	// DividendDetected -> Clear dividend cache and trigger
	bus.Subscribe(events.DividendDetected, func(e *events.Event) {
		cache.DeletePrefix("dividend:")
		if err := workCache.DeleteByPrefix("dividend:"); err != nil {
			log.Warn().Err(err).Msg("Failed to clear dividend work cache")
		}
		processor.Trigger()
	})
}
