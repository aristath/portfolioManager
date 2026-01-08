package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBrokerPosition verifies BrokerPosition struct
func TestBrokerPosition(t *testing.T) {
	position := BrokerPosition{
		Symbol:         "AAPL",
		Quantity:       10.0,
		AvgPrice:       150.0,
		CurrentPrice:   155.0,
		MarketValue:    1550.0,
		MarketValueEUR: 1450.0,
		UnrealizedPnL:  50.0,
		Currency:       "USD",
		CurrencyRate:   0.93,
	}

	assert.Equal(t, "AAPL", position.Symbol)
	assert.Equal(t, 10.0, position.Quantity)
	assert.Equal(t, 150.0, position.AvgPrice)
	assert.Equal(t, 155.0, position.CurrentPrice)
	assert.Equal(t, 1550.0, position.MarketValue)
	assert.Equal(t, 1450.0, position.MarketValueEUR)
	assert.Equal(t, 50.0, position.UnrealizedPnL)
	assert.Equal(t, "USD", position.Currency)
	assert.Equal(t, 0.93, position.CurrencyRate)
}

// TestBrokerCashBalance verifies BrokerCashBalance struct
func TestBrokerCashBalance(t *testing.T) {
	balance := BrokerCashBalance{
		Currency: "EUR",
		Amount:   1000.50,
	}

	assert.Equal(t, "EUR", balance.Currency)
	assert.Equal(t, 1000.50, balance.Amount)
}

// TestBrokerOrderResult verifies BrokerOrderResult struct
func TestBrokerOrderResult(t *testing.T) {
	result := BrokerOrderResult{
		OrderID:  "order-123",
		Symbol:   "MSFT",
		Side:     "BUY",
		Quantity: 5.0,
		Price:    300.0,
	}

	assert.Equal(t, "order-123", result.OrderID)
	assert.Equal(t, "MSFT", result.Symbol)
	assert.Equal(t, "BUY", result.Side)
	assert.Equal(t, 5.0, result.Quantity)
	assert.Equal(t, 300.0, result.Price)
}

// TestBrokerTrade verifies BrokerTrade struct
func TestBrokerTrade(t *testing.T) {
	trade := BrokerTrade{
		OrderID:    "order-456",
		Symbol:     "TSLA",
		Side:       "SELL",
		Quantity:   2.0,
		Price:      250.0,
		ExecutedAt: "2025-01-08T10:00:00Z",
	}

	assert.Equal(t, "order-456", trade.OrderID)
	assert.Equal(t, "TSLA", trade.Symbol)
	assert.Equal(t, "SELL", trade.Side)
	assert.Equal(t, 2.0, trade.Quantity)
	assert.Equal(t, 250.0, trade.Price)
	assert.Equal(t, "2025-01-08T10:00:00Z", trade.ExecutedAt)
}

// TestBrokerQuote verifies BrokerQuote struct
func TestBrokerQuote(t *testing.T) {
	quote := BrokerQuote{
		Symbol:    "GOOGL",
		Price:     140.50,
		Change:    2.5,
		ChangePct: 1.8,
		Volume:    1000000,
		Timestamp: "2025-01-08T15:30:00Z",
	}

	assert.Equal(t, "GOOGL", quote.Symbol)
	assert.Equal(t, 140.50, quote.Price)
	assert.Equal(t, 2.5, quote.Change)
	assert.Equal(t, 1.8, quote.ChangePct)
	assert.Equal(t, int64(1000000), quote.Volume)
	assert.Equal(t, "2025-01-08T15:30:00Z", quote.Timestamp)
}

// TestBrokerPendingOrder verifies BrokerPendingOrder struct
func TestBrokerPendingOrder(t *testing.T) {
	order := BrokerPendingOrder{
		OrderID:  "pending-789",
		Symbol:   "AMZN",
		Side:     "BUY",
		Quantity: 3.0,
		Price:    175.0,
		Currency: "USD",
	}

	assert.Equal(t, "pending-789", order.OrderID)
	assert.Equal(t, "AMZN", order.Symbol)
	assert.Equal(t, "BUY", order.Side)
	assert.Equal(t, 3.0, order.Quantity)
	assert.Equal(t, 175.0, order.Price)
	assert.Equal(t, "USD", order.Currency)
}

// TestBrokerSecurityInfo verifies BrokerSecurityInfo struct
func TestBrokerSecurityInfo(t *testing.T) {
	name := "Apple Inc."
	isin := "US0378331005"
	currency := "USD"
	market := "NASDAQ"
	exchangeCode := "XNAS"

	security := BrokerSecurityInfo{
		Symbol:       "AAPL",
		Name:         &name,
		ISIN:         &isin,
		Currency:     &currency,
		Market:       &market,
		ExchangeCode: &exchangeCode,
	}

	assert.Equal(t, "AAPL", security.Symbol)
	assert.Equal(t, "Apple Inc.", *security.Name)
	assert.Equal(t, "US0378331005", *security.ISIN)
	assert.Equal(t, "USD", *security.Currency)
	assert.Equal(t, "NASDAQ", *security.Market)
	assert.Equal(t, "XNAS", *security.ExchangeCode)
}

// TestBrokerSecurityInfo_NullFields verifies BrokerSecurityInfo with nil fields
func TestBrokerSecurityInfo_NullFields(t *testing.T) {
	security := BrokerSecurityInfo{
		Symbol:       "TEST",
		Name:         nil,
		ISIN:         nil,
		Currency:     nil,
		Market:       nil,
		ExchangeCode: nil,
	}

	assert.Equal(t, "TEST", security.Symbol)
	assert.Nil(t, security.Name)
	assert.Nil(t, security.ISIN)
	assert.Nil(t, security.Currency)
	assert.Nil(t, security.Market)
	assert.Nil(t, security.ExchangeCode)
}

// TestBrokerCashMovement verifies BrokerCashMovement struct
func TestBrokerCashMovement(t *testing.T) {
	movement := BrokerCashMovement{
		TotalWithdrawals: 5000.0,
		Withdrawals: []map[string]interface{}{
			{"amount": 2000.0, "date": "2025-01-01"},
			{"amount": 3000.0, "date": "2025-01-05"},
		},
		Note: "Test withdrawals",
	}

	assert.Equal(t, 5000.0, movement.TotalWithdrawals)
	assert.Len(t, movement.Withdrawals, 2)
	assert.Equal(t, "Test withdrawals", movement.Note)
}

// TestBrokerCashFlow verifies BrokerCashFlow struct
func TestBrokerCashFlow(t *testing.T) {
	cashFlow := BrokerCashFlow{
		ID:              "cf-123",
		TransactionID:   "tx-456",
		TypeDocID:       1,
		Type:            "deposit",
		TransactionType: "wire_transfer",
		DT:              "2025-01-08",
		Date:            "2025-01-08T10:00:00Z",
		SM:              1000.0,
		Amount:          1000.0,
		Curr:            "EUR",
		Currency:        "EUR",
		SMEUR:           1000.0,
		AmountEUR:       1000.0,
		Status:          "completed",
		StatusC:         1,
		Description:     "Monthly deposit",
		Params:          map[string]interface{}{"source": "bank"},
	}

	assert.Equal(t, "cf-123", cashFlow.ID)
	assert.Equal(t, "tx-456", cashFlow.TransactionID)
	assert.Equal(t, 1, cashFlow.TypeDocID)
	assert.Equal(t, "deposit", cashFlow.Type)
	assert.Equal(t, "wire_transfer", cashFlow.TransactionType)
	assert.Equal(t, "2025-01-08", cashFlow.DT)
	assert.Equal(t, "2025-01-08T10:00:00Z", cashFlow.Date)
	assert.Equal(t, 1000.0, cashFlow.SM)
	assert.Equal(t, 1000.0, cashFlow.Amount)
	assert.Equal(t, "EUR", cashFlow.Curr)
	assert.Equal(t, "EUR", cashFlow.Currency)
	assert.Equal(t, 1000.0, cashFlow.SMEUR)
	assert.Equal(t, 1000.0, cashFlow.AmountEUR)
	assert.Equal(t, "completed", cashFlow.Status)
	assert.Equal(t, 1, cashFlow.StatusC)
	assert.Equal(t, "Monthly deposit", cashFlow.Description)
	assert.NotNil(t, cashFlow.Params)
}

// TestBrokerHealthResult verifies BrokerHealthResult struct
func TestBrokerHealthResult(t *testing.T) {
	result := BrokerHealthResult{
		Connected: true,
		Timestamp: "2025-01-08T12:00:00Z",
	}

	assert.True(t, result.Connected)
	assert.Equal(t, "2025-01-08T12:00:00Z", result.Timestamp)
}

// TestBrokerHealthResult_Disconnected verifies BrokerHealthResult when disconnected
func TestBrokerHealthResult_Disconnected(t *testing.T) {
	result := BrokerHealthResult{
		Connected: false,
		Timestamp: "2025-01-08T12:00:00Z",
	}

	assert.False(t, result.Connected)
	assert.Equal(t, "2025-01-08T12:00:00Z", result.Timestamp)
}
