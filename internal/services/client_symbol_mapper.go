// Package services provides core business services shared across multiple modules.
package services

import "github.com/aristath/sentinel/internal/domain"

// ClientSymbolMapper provides client symbol mapping services.
// Converts between ISINs (internal identifier) and client-specific symbol formats.
// Used for brokers (tradernet, ibkr, schwab) and data providers.
type ClientSymbolMapper struct {
	repo domain.ClientSymbolRepositoryInterface
}

// NewClientSymbolMapper creates a new client symbol mapper service.
func NewClientSymbolMapper(repo domain.ClientSymbolRepositoryInterface) *ClientSymbolMapper {
	return &ClientSymbolMapper{
		repo: repo,
	}
}

// GetClientSymbol converts an ISIN to a client-specific symbol.
// Returns error if mapping doesn't exist (fail-fast approach).
func (m *ClientSymbolMapper) GetClientSymbol(isin, clientName string) (string, error) {
	return m.repo.GetClientSymbol(isin, clientName)
}

// SetClientSymbol creates or updates a client symbol mapping.
func (m *ClientSymbolMapper) SetClientSymbol(isin, clientName, symbol string) error {
	return m.repo.SetClientSymbol(isin, clientName, symbol)
}
