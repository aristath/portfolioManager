/**
 * Package di provides dependency injection for cash flows service initialization.
 *
 * Step 7: Initialize Cash Flows Services
 * Cash flows services handle dividends, deposits, and cash flow processing.
 */
package di

import (
	"github.com/aristath/sentinel/internal/modules/cash_flows"
	"github.com/aristath/sentinel/internal/modules/display"
	"github.com/aristath/sentinel/internal/modules/dividends"
	"github.com/rs/zerolog"
)

// initializeCashFlowsServices initializes cash flow related services.
func initializeCashFlowsServices(container *Container, cashManager *cash_flows.CashManager, displayManager *display.StateManager, log zerolog.Logger) error {
	// Dividend service implementation (adapter - uses existing dividendRepo)
	// Provides dividend-related operations (create, get, list)
	container.DividendService = cash_flows.NewDividendServiceImpl(container.DividendRepo, log)

	// Dividend creator
	// Creates dividend transactions from broker data
	container.DividendCreator = cash_flows.NewDividendCreator(container.DividendService, log)

	// Dividend yield calculator (uses ledger.db dividend transactions for yield calculation)
	// Calculates dividend yield based on dividend history and position values
	// Adapter for PositionRepo to implement PositionValueProvider interface
	positionValueAdapter := &positionValueProviderAdapter{positionRepo: container.PositionRepo}
	container.DividendYieldCalculator = dividends.NewDividendYieldCalculator(
		container.DividendRepo, // DividendRepository already implements DividendRepositoryInterface
		positionValueAdapter,
		log,
	)

	// Deposit processor (uses CashManager)
	// Processes deposit transactions and updates cash balances
	container.DepositProcessor = cash_flows.NewDepositProcessor(cashManager, log)

	// Tradernet adapter (adapts tradernet.Client to cash_flows.TradernetClient)
	// Bridges interface mismatch between broker client and cash flows service
	tradernetAdapter := cash_flows.NewTradernetAdapter(container.BrokerClient)

	// Cash flows sync job (created but not stored - used by service)
	// Syncs cash flows (deposits, dividends) from broker API
	syncJob := cash_flows.NewSyncJob(
		container.CashFlowsRepo,
		container.DepositProcessor,
		container.DividendCreator,
		tradernetAdapter,
		displayManager,
		container.EventManager,
		log,
	)

	// Cash flows service
	// Orchestrates cash flow operations (sync, processing)
	container.CashFlowsService = cash_flows.NewCashFlowsService(syncJob, log)

	return nil
}
