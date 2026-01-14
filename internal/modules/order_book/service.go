// Package order_book provides order book analysis for optimal trade execution
// This module validates liquidity and calculates optimal limit prices based on bid-ask spread
package order_book

import (
	"fmt"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/rs/zerolog"
)

// SettingsServiceInterface defines the contract for settings operations
// This interface follows Dependency Inversion Principle and maintains consistency with other services
type SettingsServiceInterface interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) (bool, error)
}

// Service provides order book analysis for trade execution
// Validates liquidity and calculates optimal limit prices using bid-ask midpoint strategy
type Service struct {
	brokerClient    domain.BrokerClient
	priceValidator  PriceValidator
	settingsService SettingsServiceInterface
	log             zerolog.Logger
}

// NewService creates a new order book service
func NewService(broker domain.BrokerClient, priceValidator PriceValidator, settingsService SettingsServiceInterface, log zerolog.Logger) *Service {
	return &Service{
		brokerClient:    broker,
		priceValidator:  priceValidator,
		settingsService: settingsService,
		log:             log,
	}
}

// IsEnabled checks if order book analysis is enabled
// Returns true if enabled, false if disabled (falls back to Tradernet-only mode)
func (s *Service) IsEnabled() bool {
	val, err := s.settingsService.Get("enable_order_book_analysis")
	if err != nil {
		s.log.Warn().Err(err).Msg("Failed to get enable_order_book_analysis setting, defaulting to enabled")
		return true
	}

	enabled, ok := val.(float64)
	if !ok {
		s.log.Warn().Msg("enable_order_book_analysis setting is not a float64, defaulting to enabled")
		return true
	}

	return enabled >= 1.0
}

// ValidateLiquidity checks if sufficient liquidity exists for the trade
// This method delegates to validator.go
func (s *Service) ValidateLiquidity(symbol, side string, quantity float64) error {
	if !s.IsEnabled() {
		s.log.Debug().Msg("Order book analysis disabled, skipping liquidity validation")
		return nil
	}

	// Fetch Level 1 quote (best bid/ask)
	orderBook, err := s.brokerClient.GetLevel1Quote(symbol)
	if err != nil {
		return fmt.Errorf("failed to fetch quote data: %w", err)
	}

	// Validate sufficient liquidity
	return s.validateSufficientLiquidity(orderBook, side, quantity)
}

// CalculateOptimalLimit calculates optimal limit price using bid-ask midpoint + price validation
// Uses asymmetric validation: blocks BAD trades only (overpaying on BUY, underselling on SELL)
// This method delegates to calculator.go
func (s *Service) CalculateOptimalLimit(symbol, side string, buffer float64) (float64, error) {
	if !s.IsEnabled() {
		s.log.Debug().Msg("Order book analysis disabled")
		return 0, fmt.Errorf("order book analysis is disabled")
	}

	// Calculate optimal limit with asymmetric validation
	// This calls methods in calculator.go
	return s.calculateOptimalLimitPrice(symbol, side, buffer)
}

// getSettingFloat retrieves a float64 setting with default value
func (s *Service) getSettingFloat(key string, defaultValue float64) float64 {
	val, err := s.settingsService.Get(key)
	if err != nil {
		return defaultValue
	}

	floatVal, ok := val.(float64)
	if !ok {
		return defaultValue
	}

	return floatVal
}
