package services

import (
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// mockYahooClientForPriceValidator for testing PriceValidator
type mockYahooClientForPriceValidator struct {
	prices map[string]float64
	err    error
}

func (m *mockYahooClientForPriceValidator) GetCurrentPrice(symbol string) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	price, ok := m.prices[symbol]
	if !ok {
		return 0, errors.New("symbol not found")
	}
	return price, nil
}

func TestPriceValidator_ValidatePrice_TradernetOnly(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	// Yahoo unavailable - should return Tradernet price
	mockYahoo := &mockYahooClientForPriceValidator{err: errors.New("connection refused")}
	validator := NewPriceValidator(mockYahoo, log)

	result := validator.ValidatePrice("AAPL.US", "AAPL", 150.0)
	assert.Equal(t, 150.0, result)
}

func TestPriceValidator_ValidatePrice_TradernetValid(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	// Tradernet within 50% of Yahoo - should return Tradernet
	mockYahoo := &mockYahooClientForPriceValidator{prices: map[string]float64{"AAPL": 148.0}}
	validator := NewPriceValidator(mockYahoo, log)

	result := validator.ValidatePrice("AAPL.US", "AAPL", 150.0)
	assert.Equal(t, 150.0, result)
}

func TestPriceValidator_ValidatePrice_TradernetSuspicious(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	// Tradernet >50% higher than Yahoo - should return Yahoo
	// Example: MOH.GR showing 653 EUR when real price is ~30 EUR
	mockYahoo := &mockYahooClientForPriceValidator{prices: map[string]float64{"MOH.DE": 30.0}}
	validator := NewPriceValidator(mockYahoo, log)

	result := validator.ValidatePrice("MOH.GR", "MOH.DE", 653.0)
	assert.Equal(t, 30.0, result) // Should use Yahoo price instead
}

func TestPriceValidator_ValidatePrice_YahooZero(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	// Yahoo returns 0 - should return Tradernet
	mockYahoo := &mockYahooClientForPriceValidator{prices: map[string]float64{"AAPL": 0}}
	validator := NewPriceValidator(mockYahoo, log)

	result := validator.ValidatePrice("AAPL.US", "AAPL", 150.0)
	assert.Equal(t, 150.0, result)
}

func TestPriceValidator_ValidatePrice_TradernetZero(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	// Tradernet returns 0 - should return 0
	mockYahoo := &mockYahooClientForPriceValidator{prices: map[string]float64{"AAPL": 150.0}}
	validator := NewPriceValidator(mockYahoo, log)

	result := validator.ValidatePrice("AAPL.US", "AAPL", 0)
	assert.Equal(t, 0.0, result)
}

func TestPriceValidator_ValidatePrice_EmptyYahooSymbol(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	// No Yahoo symbol provided - should return Tradernet price
	mockYahoo := &mockYahooClientForPriceValidator{prices: map[string]float64{"AAPL": 150.0}}
	validator := NewPriceValidator(mockYahoo, log)

	result := validator.ValidatePrice("AAPL.US", "", 150.0)
	assert.Equal(t, 150.0, result)
}

func TestPriceValidator_ValidatePrice_AtThreshold(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	// Tradernet exactly at 150% of Yahoo - should be OK (threshold is >150%)
	mockYahoo := &mockYahooClientForPriceValidator{prices: map[string]float64{"AAPL": 100.0}}
	validator := NewPriceValidator(mockYahoo, log)

	result := validator.ValidatePrice("AAPL.US", "AAPL", 150.0)
	assert.Equal(t, 150.0, result) // 150% is OK, only >150% triggers Yahoo fallback
}

func TestPriceValidator_ValidatePrice_JustOverThreshold(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	// Tradernet at 151% of Yahoo - should return Yahoo
	mockYahoo := &mockYahooClientForPriceValidator{prices: map[string]float64{"AAPL": 100.0}}
	validator := NewPriceValidator(mockYahoo, log)

	result := validator.ValidatePrice("AAPL.US", "AAPL", 151.0)
	assert.Equal(t, 100.0, result) // >150% triggers Yahoo fallback
}

func TestPriceValidator_ValidatePrice_TradernetLowerThanYahoo(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	// Tradernet lower than Yahoo - should return Tradernet (we only check if TN is suspiciously HIGH)
	mockYahoo := &mockYahooClientForPriceValidator{prices: map[string]float64{"AAPL": 200.0}}
	validator := NewPriceValidator(mockYahoo, log)

	result := validator.ValidatePrice("AAPL.US", "AAPL", 150.0)
	assert.Equal(t, 150.0, result)
}
