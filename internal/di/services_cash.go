/**
 * Package di provides dependency injection for cash manager initialization.
 *
 * Step 3: Initialize Cash Manager
 * Cash manager implements cash-as-balances architecture.
 */
package di

import (
	"github.com/aristath/sentinel/internal/modules/cash_flows"
	"github.com/rs/zerolog"
)

// initializeCashManager initializes the cash manager.
// Returns the concrete cash manager instance for use in other services.
func initializeCashManager(container *Container, log zerolog.Logger) *cash_flows.CashManager {
	// Cash manager (cash-as-balances architecture)
	// Manages cash balances with dual-write to both CashRepo and PositionRepo
	// This implements domain.CashManager interface
	cashManager := cash_flows.NewCashManagerWithDualWrite(container.CashRepo, container.PositionRepo, log)
	container.CashManager = cashManager // Store as interface
	return cashManager
}
