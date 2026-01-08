package tradernet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTransformPositionsToDomain tests position transformation
func TestTransformPositionsToDomain(t *testing.T) {
	tnPositions := []Position{
		{
			Symbol:         "AAPL",
			Quantity:       10.0,
			AvgPrice:       150.0,
			CurrentPrice:   155.0,
			MarketValue:    1550.0,
			MarketValueEUR: 1450.0,
			UnrealizedPnL:  50.0,
			Currency:       "USD",
			CurrencyRate:   0.93,
		},
	}

	result := transformPositionsToDomain(tnPositions)

	assert.Len(t, result, 1)
	assert.Equal(t, "AAPL", result[0].Symbol)
	assert.Equal(t, 10.0, result[0].Quantity)
	assert.Equal(t, 150.0, result[0].AvgPrice)
	assert.Equal(t, 155.0, result[0].CurrentPrice)
	assert.Equal(t, 1550.0, result[0].MarketValue)
	assert.Equal(t, 1450.0, result[0].MarketValueEUR)
	assert.Equal(t, 50.0, result[0].UnrealizedPnL)
	assert.Equal(t, "USD", result[0].Currency)
	assert.Equal(t, 0.93, result[0].CurrencyRate)
}

// TestTransformCashBalancesToDomain tests cash balance transformation
func TestTransformCashBalancesToDomain(t *testing.T) {
	tnBalances := []CashBalance{
		{Currency: "EUR", Amount: 1000.0},
		{Currency: "USD", Amount: 500.0},
	}

	result := transformCashBalancesToDomain(tnBalances)

	assert.Len(t, result, 2)
	assert.Equal(t, "EUR", result[0].Currency)
	assert.Equal(t, 1000.0, result[0].Amount)
	assert.Equal(t, "USD", result[1].Currency)
	assert.Equal(t, 500.0, result[1].Amount)
}

// TestTransformOrderResultToDomain tests order result transformation
func TestTransformOrderResultToDomain(t *testing.T) {
	t.Run("non-nil result", func(t *testing.T) {
		tnResult := &OrderResult{
			OrderID:  "order-123",
			Symbol:   "MSFT",
			Side:     "BUY",
			Quantity: 5.0,
			Price:    300.0,
		}

		result := transformOrderResultToDomain(tnResult)

		assert.NotNil(t, result)
		assert.Equal(t, "order-123", result.OrderID)
		assert.Equal(t, "MSFT", result.Symbol)
		assert.Equal(t, "BUY", result.Side)
		assert.Equal(t, 5.0, result.Quantity)
		assert.Equal(t, 300.0, result.Price)
	})

	t.Run("nil result", func(t *testing.T) {
		result := transformOrderResultToDomain(nil)
		assert.Nil(t, result)
	})
}

// TestTransformTradesToDomain tests trade transformation
func TestTransformTradesToDomain(t *testing.T) {
	tnTrades := []Trade{
		{
			OrderID:    "trade-1",
			Symbol:     "TSLA",
			Side:       "SELL",
			Quantity:   2.0,
			Price:      250.0,
			ExecutedAt: "2025-01-08T10:00:00Z",
		},
	}

	result := transformTradesToDomain(tnTrades)

	assert.Len(t, result, 1)
	assert.Equal(t, "trade-1", result[0].OrderID)
	assert.Equal(t, "TSLA", result[0].Symbol)
	assert.Equal(t, "SELL", result[0].Side)
	assert.Equal(t, 2.0, result[0].Quantity)
	assert.Equal(t, 250.0, result[0].Price)
	assert.Equal(t, "2025-01-08T10:00:00Z", result[0].ExecutedAt)
}

// TestTransformQuoteToDomain tests quote transformation
func TestTransformQuoteToDomain(t *testing.T) {
	t.Run("non-nil quote", func(t *testing.T) {
		tnQuote := &Quote{
			Symbol:    "GOOGL",
			Price:     140.50,
			Change:    2.5,
			ChangePct: 1.8,
			Volume:    1000000,
			Timestamp: "2025-01-08T15:30:00Z",
		}

		result := transformQuoteToDomain(tnQuote)

		assert.NotNil(t, result)
		assert.Equal(t, "GOOGL", result.Symbol)
		assert.Equal(t, 140.50, result.Price)
		assert.Equal(t, 2.5, result.Change)
		assert.Equal(t, 1.8, result.ChangePct)
		assert.Equal(t, int64(1000000), result.Volume)
		assert.Equal(t, "2025-01-08T15:30:00Z", result.Timestamp)
	})

	t.Run("nil quote", func(t *testing.T) {
		result := transformQuoteToDomain(nil)
		assert.Nil(t, result)
	})
}

// TestTransformPendingOrdersToDomain tests pending order transformation
func TestTransformPendingOrdersToDomain(t *testing.T) {
	tnOrders := []PendingOrder{
		{
			OrderID:  "pending-1",
			Symbol:   "AMZN",
			Side:     "BUY",
			Quantity: 3.0,
			Price:    175.0,
			Currency: "USD",
		},
	}

	result := transformPendingOrdersToDomain(tnOrders)

	assert.Len(t, result, 1)
	assert.Equal(t, "pending-1", result[0].OrderID)
	assert.Equal(t, "AMZN", result[0].Symbol)
	assert.Equal(t, "BUY", result[0].Side)
	assert.Equal(t, 3.0, result[0].Quantity)
	assert.Equal(t, 175.0, result[0].Price)
	assert.Equal(t, "USD", result[0].Currency)
}

// TestTransformSecurityInfoToDomain tests security info transformation
func TestTransformSecurityInfoToDomain(t *testing.T) {
	name := "Apple Inc."
	isin := "US0378331005"
	currency := "USD"
	market := "NASDAQ"
	exchangeCode := "XNAS"

	tnSecurities := []SecurityInfo{
		{
			Symbol:       "AAPL",
			Name:         &name,
			ISIN:         &isin,
			Currency:     &currency,
			Market:       &market,
			ExchangeCode: &exchangeCode,
		},
	}

	result := transformSecurityInfoToDomain(tnSecurities)

	assert.Len(t, result, 1)
	assert.Equal(t, "AAPL", result[0].Symbol)
	assert.NotNil(t, result[0].Name)
	assert.Equal(t, "Apple Inc.", *result[0].Name)
	assert.NotNil(t, result[0].ISIN)
	assert.Equal(t, "US0378331005", *result[0].ISIN)
	assert.NotNil(t, result[0].Currency)
	assert.Equal(t, "USD", *result[0].Currency)
	assert.NotNil(t, result[0].Market)
	assert.Equal(t, "NASDAQ", *result[0].Market)
	assert.NotNil(t, result[0].ExchangeCode)
	assert.Equal(t, "XNAS", *result[0].ExchangeCode)
}

// TestTransformSecurityInfoToDomain_NullFields tests security info transformation with nil fields
func TestTransformSecurityInfoToDomain_NullFields(t *testing.T) {
	tnSecurities := []SecurityInfo{
		{
			Symbol:       "TEST",
			Name:         nil,
			ISIN:         nil,
			Currency:     nil,
			Market:       nil,
			ExchangeCode: nil,
		},
	}

	result := transformSecurityInfoToDomain(tnSecurities)

	assert.Len(t, result, 1)
	assert.Equal(t, "TEST", result[0].Symbol)
	assert.Nil(t, result[0].Name)
	assert.Nil(t, result[0].ISIN)
	assert.Nil(t, result[0].Currency)
	assert.Nil(t, result[0].Market)
	assert.Nil(t, result[0].ExchangeCode)
}

// TestTransformCashFlowsToDomain tests cash flow transformation
func TestTransformCashFlowsToDomain(t *testing.T) {
	tnFlows := []CashFlowTransaction{
		{
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
		},
	}

	result := transformCashFlowsToDomain(tnFlows)

	assert.Len(t, result, 1)
	assert.Equal(t, "cf-123", result[0].ID)
	assert.Equal(t, "tx-456", result[0].TransactionID)
	assert.Equal(t, 1, result[0].TypeDocID)
	assert.Equal(t, "deposit", result[0].Type)
	assert.Equal(t, "wire_transfer", result[0].TransactionType)
	assert.Equal(t, "2025-01-08", result[0].DT)
	assert.Equal(t, "2025-01-08T10:00:00Z", result[0].Date)
	assert.Equal(t, 1000.0, result[0].SM)
	assert.Equal(t, 1000.0, result[0].Amount)
	assert.Equal(t, "EUR", result[0].Curr)
	assert.Equal(t, "EUR", result[0].Currency)
	assert.Equal(t, 1000.0, result[0].SMEUR)
	assert.Equal(t, 1000.0, result[0].AmountEUR)
	assert.Equal(t, "completed", result[0].Status)
	assert.Equal(t, 1, result[0].StatusC)
	assert.Equal(t, "Monthly deposit", result[0].Description)
	assert.NotNil(t, result[0].Params)
}

// TestTransformCashMovementsToDomain tests cash movements transformation
func TestTransformCashMovementsToDomain(t *testing.T) {
	t.Run("non-nil movements", func(t *testing.T) {
		tnMovements := &CashMovementsResponse{
			TotalWithdrawals: 5000.0,
			Withdrawals: []map[string]interface{}{
				{"amount": 2000.0, "date": "2025-01-01"},
				{"amount": 3000.0, "date": "2025-01-05"},
			},
			Note: "Test withdrawals",
		}

		result := transformCashMovementsToDomain(tnMovements)

		assert.NotNil(t, result)
		assert.Equal(t, 5000.0, result.TotalWithdrawals)
		assert.Len(t, result.Withdrawals, 2)
		assert.Equal(t, "Test withdrawals", result.Note)
	})

	t.Run("nil movements", func(t *testing.T) {
		result := transformCashMovementsToDomain(nil)
		assert.Nil(t, result)
	})
}

// TestTransformHealthResultToDomain tests health result transformation
func TestTransformHealthResultToDomain(t *testing.T) {
	t.Run("connected", func(t *testing.T) {
		tnHealth := &HealthCheckResult{
			Connected: true,
			Timestamp: "2025-01-08T12:00:00Z",
		}

		result := transformHealthResultToDomain(tnHealth)

		assert.NotNil(t, result)
		assert.True(t, result.Connected)
		assert.Equal(t, "2025-01-08T12:00:00Z", result.Timestamp)
	})

	t.Run("disconnected", func(t *testing.T) {
		tnHealth := &HealthCheckResult{
			Connected: false,
			Timestamp: "2025-01-08T12:00:00Z",
		}

		result := transformHealthResultToDomain(tnHealth)

		assert.NotNil(t, result)
		assert.False(t, result.Connected)
		assert.Equal(t, "2025-01-08T12:00:00Z", result.Timestamp)
	})

	t.Run("nil health", func(t *testing.T) {
		result := transformHealthResultToDomain(nil)
		assert.Nil(t, result)
	})
}
