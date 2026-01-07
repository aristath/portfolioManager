package scheduler

import "github.com/aristath/portfolioManager/internal/events"

// UniverseServiceInterface defines the contract for universe service operations
// Used by scheduler to enable testing with mocks
type UniverseServiceInterface interface {
	SyncPrices() error
}

// BalanceServiceInterface defines the contract for balance service operations
// Used by scheduler to enable testing with mocks
type BalanceServiceInterface interface {
	GetAllCurrencies() ([]string, error)
	GetTotalByCurrency(currency string) (float64, error)
}

// EventManagerInterface defines the contract for event emission
type EventManagerInterface interface {
	Emit(eventType events.EventType, module string, data map[string]interface{})
}
