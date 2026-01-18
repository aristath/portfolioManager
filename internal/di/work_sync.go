/**
 * Package di provides dependency injection for sync work registration.
 *
 * Sync work types handle synchronization of portfolio data, trades, cash flows,
 * prices, exchange rates, display updates, and negative balance checks.
 */
package di

import (
	"github.com/aristath/sentinel/internal/events"
	"github.com/aristath/sentinel/internal/work"
	"github.com/rs/zerolog"
)

// Sync adapters
type syncPortfolioAdapter struct {
	container *Container
}

func (a *syncPortfolioAdapter) SyncPortfolio() error {
	return a.container.PortfolioService.SyncFromTradernet()
}

type syncTradesAdapter struct {
	container *Container
}

func (a *syncTradesAdapter) SyncTrades() error {
	return a.container.TradingService.SyncFromTradernet()
}

type syncCashFlowsAdapter struct {
	container *Container
}

func (a *syncCashFlowsAdapter) SyncCashFlows() error {
	return a.container.CashFlowsService.SyncFromTradernet()
}

type syncPricesAdapter struct {
	container *Container
}

func (a *syncPricesAdapter) SyncPrices() error {
	return a.container.UniverseService.SyncPrices()
}

type syncExchangeRatesAdapter struct {
	container *Container
}

func (a *syncExchangeRatesAdapter) SyncExchangeRates() error {
	return a.container.ExchangeRateCacheService.SyncRates()
}

type syncDisplayAdapter struct {
	container *Container
}

func (a *syncDisplayAdapter) UpdateDisplay() error {
	if a.container.UpdateDisplayTicker != nil {
		return a.container.UpdateDisplayTicker()
	}
	return nil
}

type syncNegativeBalanceAdapter struct {
	container *Container
}

func (a *syncNegativeBalanceAdapter) CheckNegativeBalances() error {
	if a.container.EmergencyRebalance != nil {
		return a.container.EmergencyRebalance()
	}
	return nil
}

type syncEventManagerAdapter struct {
	container *Container
}

func (a *syncEventManagerAdapter) Emit(event string, data any) {
	a.container.EventManager.Emit(events.JobProgress, event, nil)
}

func registerSyncWork(registry *work.Registry, container *Container, log zerolog.Logger) {
	deps := &work.SyncDeps{
		PortfolioService:       &syncPortfolioAdapter{container: container},
		TradesService:          &syncTradesAdapter{container: container},
		CashFlowsService:       &syncCashFlowsAdapter{container: container},
		PricesService:          &syncPricesAdapter{container: container},
		ExchangeRateService:    &syncExchangeRatesAdapter{container: container},
		DisplayService:         &syncDisplayAdapter{container: container},
		NegativeBalanceService: &syncNegativeBalanceAdapter{container: container},
		EventManager:           &syncEventManagerAdapter{container: container},
	}

	work.RegisterSyncWorkTypes(registry, deps)
	log.Debug().Msg("Sync work types registered")
}
