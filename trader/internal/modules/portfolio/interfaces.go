package portfolio

// PositionRepositoryInterface defines the contract for position repository operations
// Used by PortfolioService to enable testing with mocks
type PositionRepositoryInterface interface {
	GetAll() ([]Position, error)
	GetWithSecurityInfo() ([]PositionWithSecurity, error)
	GetBySymbol(symbol string) (*Position, error)
	Upsert(position Position) error
	Delete(symbol string) error
}

// Note: TradernetClientInterface and CurrencyExchangeServiceInterface have been moved to domain/interfaces.go
// They are now available as domain.TradernetClientInterface and domain.CurrencyExchangeServiceInterface
