/**
 * Package di provides dependency injection for trading work registration.
 *
 * Trading work types handle trade execution and retry of failed trades.
 */
package di

import (
	"github.com/aristath/sentinel/internal/modules/trading"
	"github.com/aristath/sentinel/internal/work"
	"github.com/rs/zerolog"
)

// Trading adapters
type tradingExecutionAdapter struct {
	container *Container
}

func (a *tradingExecutionAdapter) ExecutePendingTrades() error {
	// Get pending recommendations (limit to 1 for throttling)
	recommendations, err := a.container.RecommendationRepo.GetPendingRecommendations(1)
	if err != nil {
		return err
	}

	if len(recommendations) == 0 {
		return nil
	}

	rec := recommendations[0]

	// Execute the trade via trading service
	tradeRequest := trading.TradeRequest{
		Symbol:   rec.Symbol,
		Side:     rec.Side,
		Quantity: int(rec.Quantity),
		Reason:   rec.Reason,
	}

	result, err := a.container.TradingService.ExecuteTrade(tradeRequest)
	if err != nil {
		// Record failed attempt
		_ = a.container.RecommendationRepo.RecordFailedAttempt(rec.UUID, err.Error())
		return err
	}

	if !result.Success {
		_ = a.container.RecommendationRepo.RecordFailedAttempt(rec.UUID, result.Reason)
		return nil
	}

	// Mark recommendation as executed
	return a.container.RecommendationRepo.MarkExecuted(rec.UUID)
}

func (a *tradingExecutionAdapter) HasPendingTrades() bool {
	buyCount, sellCount, _ := a.container.RecommendationRepo.CountPendingBySide()
	return buyCount > 0 || sellCount > 0
}

type tradingRetryAdapter struct {
	container *Container
}

func (a *tradingRetryAdapter) RetryFailedTrades() error {
	// Get pending retries from the trade repository
	retries, err := a.container.TradeRepo.GetPendingRetries()
	if err != nil {
		return err
	}

	for _, retry := range retries {
		// Try to execute the retry
		tradeRequest := trading.TradeRequest{
			Symbol:   retry.Symbol,
			Side:     retry.Side,
			Quantity: int(retry.Quantity), // Convert float64 to int
			Reason:   retry.Reason,
		}

		result, err := a.container.TradingService.ExecuteTrade(tradeRequest)
		if err != nil {
			// Increment retry attempt
			_ = a.container.TradeRepo.IncrementRetryAttempt(retry.ID)
			continue
		}

		if result.Success {
			// Mark as completed
			_ = a.container.TradeRepo.UpdateRetryStatus(retry.ID, "completed")
		} else {
			// Increment retry attempt
			_ = a.container.TradeRepo.IncrementRetryAttempt(retry.ID)
		}
	}

	return nil
}

func (a *tradingRetryAdapter) HasFailedTrades() bool {
	retries, err := a.container.TradeRepo.GetPendingRetries()
	if err != nil {
		return false
	}
	return len(retries) > 0
}

func registerTradingWork(registry *work.Registry, container *Container, log zerolog.Logger) {
	deps := &work.TradingDeps{
		ExecutionService: &tradingExecutionAdapter{container: container},
		RetryService:     &tradingRetryAdapter{container: container},
	}

	work.RegisterTradingWorkTypes(registry, deps)
	log.Debug().Msg("Trading work types registered")
}
