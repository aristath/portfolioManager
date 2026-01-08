package tradernet

import (
	"github.com/aristath/sentinel/internal/domain"
	"github.com/rs/zerolog"
)

// TradernetBrokerAdapter adapts tradernet.Client to domain.BrokerClient
// This adapter owns the Tradernet client internally and provides a broker-agnostic interface
type TradernetBrokerAdapter struct {
	client *Client
}

// NewTradernetBrokerAdapter creates a new Tradernet broker adapter
// The adapter owns the Tradernet client internally
func NewTradernetBrokerAdapter(apiKey, apiSecret string, log zerolog.Logger) *TradernetBrokerAdapter {
	client := NewClient(apiKey, apiSecret, log)
	return &TradernetBrokerAdapter{
		client: client,
	}
}

// GetPortfolio implements domain.BrokerClient
func (a *TradernetBrokerAdapter) GetPortfolio() ([]domain.BrokerPosition, error) {
	tnPositions, err := a.client.GetPortfolio()
	if err != nil {
		return nil, err
	}
	return transformPositionsToDomain(tnPositions), nil
}

// GetCashBalances implements domain.BrokerClient
func (a *TradernetBrokerAdapter) GetCashBalances() ([]domain.BrokerCashBalance, error) {
	tnBalances, err := a.client.GetCashBalances()
	if err != nil {
		return nil, err
	}
	return transformCashBalancesToDomain(tnBalances), nil
}

// PlaceOrder implements domain.BrokerClient
func (a *TradernetBrokerAdapter) PlaceOrder(symbol, side string, quantity float64) (*domain.BrokerOrderResult, error) {
	tnResult, err := a.client.PlaceOrder(symbol, side, quantity)
	if err != nil {
		return nil, err
	}
	return transformOrderResultToDomain(tnResult), nil
}

// GetExecutedTrades implements domain.BrokerClient
func (a *TradernetBrokerAdapter) GetExecutedTrades(limit int) ([]domain.BrokerTrade, error) {
	tnTrades, err := a.client.GetExecutedTrades(limit)
	if err != nil {
		return nil, err
	}
	return transformTradesToDomain(tnTrades), nil
}

// GetPendingOrders implements domain.BrokerClient
func (a *TradernetBrokerAdapter) GetPendingOrders() ([]domain.BrokerPendingOrder, error) {
	tnOrders, err := a.client.GetPendingOrders()
	if err != nil {
		return nil, err
	}
	return transformPendingOrdersToDomain(tnOrders), nil
}

// GetQuote implements domain.BrokerClient
func (a *TradernetBrokerAdapter) GetQuote(symbol string) (*domain.BrokerQuote, error) {
	tnQuote, err := a.client.GetQuote(symbol)
	if err != nil {
		return nil, err
	}
	return transformQuoteToDomain(tnQuote), nil
}

// FindSymbol implements domain.BrokerClient
func (a *TradernetBrokerAdapter) FindSymbol(symbol string, exchange *string) ([]domain.BrokerSecurityInfo, error) {
	tnSecurities, err := a.client.FindSymbol(symbol, exchange)
	if err != nil {
		return nil, err
	}
	return transformSecurityInfoToDomain(tnSecurities), nil
}

// GetAllCashFlows implements domain.BrokerClient
func (a *TradernetBrokerAdapter) GetAllCashFlows(limit int) ([]domain.BrokerCashFlow, error) {
	tnFlows, err := a.client.GetAllCashFlows(limit)
	if err != nil {
		return nil, err
	}
	return transformCashFlowsToDomain(tnFlows), nil
}

// GetCashMovements implements domain.BrokerClient
func (a *TradernetBrokerAdapter) GetCashMovements() (*domain.BrokerCashMovement, error) {
	tnMovements, err := a.client.GetCashMovements()
	if err != nil {
		return nil, err
	}
	return transformCashMovementsToDomain(tnMovements), nil
}

// IsConnected implements domain.BrokerClient
func (a *TradernetBrokerAdapter) IsConnected() bool {
	return a.client.IsConnected()
}

// HealthCheck implements domain.BrokerClient
func (a *TradernetBrokerAdapter) HealthCheck() (*domain.BrokerHealthResult, error) {
	tnHealth, err := a.client.HealthCheck()
	if err != nil {
		return nil, err
	}
	return transformHealthResultToDomain(tnHealth), nil
}

// SetCredentials implements domain.BrokerClient
func (a *TradernetBrokerAdapter) SetCredentials(apiKey, apiSecret string) {
	a.client.SetCredentials(apiKey, apiSecret)
}
