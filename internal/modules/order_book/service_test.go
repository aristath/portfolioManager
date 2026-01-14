package order_book

import (
	"fmt"
	"testing"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/aristath/sentinel/internal/modules/settings"
	internalTesting "github.com/aristath/sentinel/internal/testing"
	"github.com/rs/zerolog"
)

// createTestSettingsService creates a properly initialized settings service for testing
func createTestSettingsService(t *testing.T) *settings.Service {
	t.Helper()
	// Create test config database
	testDB, _ := internalTesting.NewTestDB(t, "config")
	// Create settings repository
	settingsRepo := settings.NewRepository(testDB.Conn(), zerolog.Nop())
	// Create settings service
	return settings.NewService(settingsRepo, zerolog.Nop())
}

// mockPriceValidator is a simple mock implementing PriceValidator interface
type mockPriceValidator struct {
	prices map[string]*float64
}

func newMockPriceValidator() *mockPriceValidator {
	return &mockPriceValidator{
		prices: make(map[string]*float64),
	}
}

func (m *mockPriceValidator) SetPrice(symbol string, price *float64) {
	m.prices[symbol] = price
}

func (m *mockPriceValidator) GetValidationPrice(symbol string) (*float64, error) {
	if price, exists := m.prices[symbol]; exists {
		return price, nil
	}
	return nil, fmt.Errorf("price not found for %s", symbol)
}

// TestValidateLiquidity_SufficientLiquidity tests liquidity validation with sufficient liquidity
func TestValidateLiquidity_SufficientLiquidity(t *testing.T) {
	broker := internalTesting.NewMockBrokerClient()
	priceValidator := newMockPriceValidator()
	settingsService := createTestSettingsService(t)

	// Set defaults
	settingsService.Set("enable_order_book_analysis", 1.0)
	settingsService.Set("min_liquidity_multiple", 2.0)
	settingsService.Set("order_book_depth_levels", 5.0)

	// Configure mock to return order book with sufficient liquidity
	broker.SetOrderBook(&domain.BrokerOrderBook{
		Symbol: "AAPL.US",
		Bids: []domain.OrderBookLevel{
			{Price: 150.0, Quantity: 1000.0, Position: 1},
		},
		Asks: []domain.OrderBookLevel{
			{Price: 151.0, Quantity: 1000.0, Position: 1},
		},
		Timestamp: "2024-01-01T00:00:00Z",
	})

	service := NewService(broker, priceValidator, settingsService, zerolog.Nop())

	// Test BUY side - need 100 shares, have 1000 available (10x > 2x required)
	err := service.ValidateLiquidity("AAPL.US", "BUY", 100.0)
	if err != nil {
		t.Errorf("Expected no error for sufficient liquidity, got: %v", err)
	}

	// Test SELL side - need 100 shares, have 1000 available (10x > 2x required)
	err = service.ValidateLiquidity("AAPL.US", "SELL", 100.0)
	if err != nil {
		t.Errorf("Expected no error for sufficient liquidity, got: %v", err)
	}
}

// TestValidateLiquidity_InsufficientLiquidity tests liquidity validation with insufficient liquidity
func TestValidateLiquidity_InsufficientLiquidity(t *testing.T) {
	broker := internalTesting.NewMockBrokerClient()
	priceValidator := newMockPriceValidator()
	settingsService := createTestSettingsService(t)

	// Set defaults
	settingsService.Set("min_liquidity_multiple", 2.0)
	settingsService.Set("order_book_depth_levels", 5.0)

	// Configure mock to return order book with insufficient liquidity
	broker.SetOrderBook(&domain.BrokerOrderBook{
		Symbol: "AAPL.US",
		Bids: []domain.OrderBookLevel{
			{Price: 150.0, Quantity: 50.0, Position: 1},
		},
		Asks: []domain.OrderBookLevel{
			{Price: 151.0, Quantity: 50.0, Position: 1},
		},
		Timestamp: "2024-01-01T00:00:00Z",
	})

	service := NewService(broker, priceValidator, settingsService, zerolog.Nop())

	// Test BUY side - need 100 shares, only 50 available (0.5x < 2x required)
	err := service.ValidateLiquidity("AAPL.US", "BUY", 100.0)
	if err == nil {
		t.Error("Expected error for insufficient liquidity, got nil")
	}

	// Test SELL side - need 100 shares, only 50 available (0.5x < 2x required)
	err = service.ValidateLiquidity("AAPL.US", "SELL", 100.0)
	if err == nil {
		t.Error("Expected error for insufficient liquidity, got nil")
	}
}

// TestCalculateOptimalLimit_BuyWithCheaperOrderBook tests BUY when order book is cheaper than validation price
func TestCalculateOptimalLimit_BuyWithCheaperOrderBook(t *testing.T) {
	orderBookPrice := 90.0
	testPrice := 100.0

	broker := internalTesting.NewMockBrokerClient()
	priceValidator := newMockPriceValidator()
	settingsService := createTestSettingsService(t)

	// Set defaults
	settingsService.Set("price_discrepancy_threshold", 0.50)

	broker.SetOrderBook(&domain.BrokerOrderBook{
		Symbol: "AAPL.US",
		Bids: []domain.OrderBookLevel{
			{Price: 89.0, Quantity: 1000.0, Position: 1},
		},
		Asks: []domain.OrderBookLevel{
			{Price: orderBookPrice, Quantity: 1000.0, Position: 1},
		},
		Timestamp: "2024-01-01T00:00:00Z",
	})

	priceValidator.SetPrice("AAPL.US", &testPrice)

	service := NewService(broker, priceValidator, settingsService, zerolog.Nop())

	// BUY at cheaper price should be ALLOWED (10% cheaper is good!)
	limitPrice, err := service.CalculateOptimalLimit("AAPL.US", "BUY", 0.05)
	if err != nil {
		t.Errorf("Expected no error for cheaper buy, got: %v", err)
	}

	// Limit price should be midpoint + 5% buffer
	// Midpoint = (89 + 90) / 2 = 89.5
	expectedMidpoint := (89.0 + orderBookPrice) / 2.0
	expectedLimit := expectedMidpoint * 1.05
	if limitPrice != expectedLimit {
		t.Errorf("Expected limit price %.2f, got %.2f", expectedLimit, limitPrice)
	}
}

// TestCalculateOptimalLimit_BuyWithExpensiveOrderBook tests BUY when order book is significantly more expensive (API bug)
func TestCalculateOptimalLimit_BuyWithExpensiveOrderBook(t *testing.T) {
	orderBookPrice := 1000.0
	testPrice := 100.0

	broker := internalTesting.NewMockBrokerClient()
	priceValidator := newMockPriceValidator()
	settingsService := createTestSettingsService(t)

	// Set defaults
	settingsService.Set("price_discrepancy_threshold", 0.50)

	broker.SetOrderBook(&domain.BrokerOrderBook{
		Symbol: "AAPL.US",
		Bids: []domain.OrderBookLevel{
			{Price: 999.0, Quantity: 1000.0, Position: 1},
		},
		Asks: []domain.OrderBookLevel{
			{Price: orderBookPrice, Quantity: 1000.0, Position: 1},
		},
		Timestamp: "2024-01-01T00:00:00Z",
	})

	priceValidator.SetPrice("AAPL.US", &testPrice)

	service := NewService(broker, priceValidator, settingsService, zerolog.Nop())

	// BUY at 10x price should be BLOCKED (overpaying)
	_, err := service.CalculateOptimalLimit("AAPL.US", "BUY", 0.05)
	if err == nil {
		t.Error("Expected error for expensive buy (overpaying), got nil")
	}
}

// TestCalculateOptimalLimit_SellWithHigherOrderBook tests SELL when order book is higher than validation price
func TestCalculateOptimalLimit_SellWithHigherOrderBook(t *testing.T) {
	orderBookPrice := 110.0
	testPrice := 100.0

	broker := internalTesting.NewMockBrokerClient()
	priceValidator := newMockPriceValidator()
	settingsService := createTestSettingsService(t)

	// Set defaults
	settingsService.Set("price_discrepancy_threshold", 0.50)

	broker.SetOrderBook(&domain.BrokerOrderBook{
		Symbol: "AAPL.US",
		Bids: []domain.OrderBookLevel{
			{Price: orderBookPrice, Quantity: 1000.0, Position: 1},
		},
		Asks: []domain.OrderBookLevel{
			{Price: 111.0, Quantity: 1000.0, Position: 1},
		},
		Timestamp: "2024-01-01T00:00:00Z",
	})

	priceValidator.SetPrice("AAPL.US", &testPrice)

	service := NewService(broker, priceValidator, settingsService, zerolog.Nop())

	// SELL at higher price should be ALLOWED (10% higher is good!)
	limitPrice, err := service.CalculateOptimalLimit("AAPL.US", "SELL", 0.05)
	if err != nil {
		t.Errorf("Expected no error for higher sell price, got: %v", err)
	}

	// Limit price should be midpoint - 5% buffer
	// Midpoint = (110 + 111) / 2 = 110.5
	expectedMidpoint := (orderBookPrice + 111.0) / 2.0
	expectedLimit := expectedMidpoint * 0.95
	if limitPrice != expectedLimit {
		t.Errorf("Expected limit price %.2f, got %.2f", expectedLimit, limitPrice)
	}
}

// TestCalculateOptimalLimit_SellWithLowOrderBook tests SELL when order book is significantly lower (API bug)
func TestCalculateOptimalLimit_SellWithLowOrderBook(t *testing.T) {
	orderBookPrice := 10.0
	testPrice := 100.0

	broker := internalTesting.NewMockBrokerClient()
	priceValidator := newMockPriceValidator()
	settingsService := createTestSettingsService(t)

	// Set defaults
	settingsService.Set("price_discrepancy_threshold", 0.50)

	broker.SetOrderBook(&domain.BrokerOrderBook{
		Symbol: "AAPL.US",
		Bids: []domain.OrderBookLevel{
			{Price: orderBookPrice, Quantity: 1000.0, Position: 1},
		},
		Asks: []domain.OrderBookLevel{
			{Price: 11.0, Quantity: 1000.0, Position: 1},
		},
		Timestamp: "2024-01-01T00:00:00Z",
	})

	priceValidator.SetPrice("AAPL.US", &testPrice)

	service := NewService(broker, priceValidator, settingsService, zerolog.Nop())

	// SELL at 90% discount should be BLOCKED (underselling)
	_, err := service.CalculateOptimalLimit("AAPL.US", "SELL", 0.05)
	if err == nil {
		t.Error("Expected error for low sell price (underselling), got nil")
	}
}

// TestIsEnabled tests the IsEnabled method
func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		setting  interface{}
		expected bool
	}{
		{"Enabled", 1.0, true},
		{"Disabled", 0.0, false},
		{"Missing setting", nil, true}, // Default to enabled
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			broker := internalTesting.NewMockBrokerClient()
			priceValidator := newMockPriceValidator()
			settingsService := createTestSettingsService(t)

			if tt.setting != nil {
				settingsService.Set("enable_order_book_analysis", tt.setting)
			}

			service := NewService(broker, priceValidator, settingsService, zerolog.Nop())
			result := service.IsEnabled()

			if result != tt.expected {
				t.Errorf("Expected IsEnabled() = %v, got %v", tt.expected, result)
			}
		})
	}
}
