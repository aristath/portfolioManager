// Package services provides core business services shared across multiple modules.
package services

import "github.com/aristath/sentinel/internal/domain"

// BrokerSymbolMapper provides broker symbol mapping services.
// Converts between ISINs (internal identifier) and broker-specific symbol formats.
type BrokerSymbolMapper struct {
	repo domain.BrokerSymbolRepositoryInterface
}

// NewBrokerSymbolMapper creates a new broker symbol mapper service.
func NewBrokerSymbolMapper(repo domain.BrokerSymbolRepositoryInterface) *BrokerSymbolMapper {
	return &BrokerSymbolMapper{
		repo: repo,
	}
}

// GetBrokerSymbol converts an ISIN to a broker-specific symbol.
// Returns error if mapping doesn't exist (fail-fast approach).
func (m *BrokerSymbolMapper) GetBrokerSymbol(isin, brokerName string) (string, error) {
	return m.repo.GetBrokerSymbol(isin, brokerName)
}

// SetBrokerSymbol creates or updates a broker symbol mapping.
func (m *BrokerSymbolMapper) SetBrokerSymbol(isin, brokerName, symbol string) error {
	return m.repo.SetBrokerSymbol(isin, brokerName, symbol)
}
