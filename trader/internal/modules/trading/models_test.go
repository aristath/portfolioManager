package trading

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTradeSide_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		side     TradeSide
		expected bool
	}{
		{"BUY is valid", TradeSideBuy, true},
		{"SELL is valid", TradeSideSell, true},
		{"empty is invalid", TradeSide(""), false},
		{"invalid value is invalid", TradeSide("INVALID"), false},
		{"lowercase buy is invalid", TradeSide("buy"), false},
		{"lowercase sell is invalid", TradeSide("sell"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.side.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTradeSide_IsBuy(t *testing.T) {
	tests := []struct {
		name     string
		side     TradeSide
		expected bool
	}{
		{"BUY is buy", TradeSideBuy, true},
		{"SELL is not buy", TradeSideSell, false},
		{"empty is not buy", TradeSide(""), false},
		{"invalid is not buy", TradeSide("INVALID"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.side.IsBuy()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTradeSide_IsSell(t *testing.T) {
	tests := []struct {
		name     string
		side     TradeSide
		expected bool
	}{
		{"SELL is sell", TradeSideSell, true},
		{"BUY is not sell", TradeSideBuy, false},
		{"empty is not sell", TradeSide(""), false},
		{"invalid is not sell", TradeSide("INVALID"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.side.IsSell()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTradeSideFromString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    TradeSide
		expectedErr bool
	}{
		{"BUY uppercase", "BUY", TradeSideBuy, false},
		{"SELL uppercase", "SELL", TradeSideSell, false},
		{"buy lowercase", "buy", TradeSideBuy, false},
		{"sell lowercase", "sell", TradeSideSell, false},
		{"Buy mixed case", "Buy", TradeSideBuy, false},
		{"Sell mixed case", "Sell", TradeSideSell, false},
		{"bUy mixed case", "bUy", TradeSideBuy, false},
		{"sElL mixed case", "sElL", TradeSideSell, false},
		{"empty string", "", "", true},
		{"invalid value", "INVALID", "", true},
		{"partial match", "BU", "", true},
		{"whitespace around BUY", "  BUY  ", "", true}, // TradeSideFromString doesn't trim, only uppercases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TradeSideFromString(tt.input)
			if tt.expectedErr {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestTrade_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		trade       Trade
		expectedErr bool
		description string
	}{
		{
			name: "valid BUY trade",
			trade: Trade{
				ID:         1,
				Symbol:     "AAPL.US",
				Side:       TradeSideBuy,
				Quantity:   10.0,
				Price:      150.0,
				ExecutedAt: now,
				Mode:       "live",
				Source:     "manual",
			},
			expectedErr: false,
			description: "All required fields valid",
		},
		{
			name: "valid SELL trade",
			trade: Trade{
				ID:         2,
				Symbol:     "MSFT.US",
				Side:       TradeSideSell,
				Quantity:   5.0,
				Price:      200.0,
				ExecutedAt: now,
				Mode:       "live",
				Source:     "auto",
			},
			expectedErr: false,
			description: "SELL trade is valid",
		},
		{
			name: "empty symbol",
			trade: Trade{
				ID:         3,
				Symbol:     "",
				Side:       TradeSideBuy,
				Quantity:   10.0,
				Price:      150.0,
				ExecutedAt: now,
				Mode:       "live",
				Source:     "manual",
			},
			expectedErr: true,
			description: "Empty symbol should fail",
		},
		{
			name: "whitespace only symbol",
			trade: Trade{
				ID:         4,
				Symbol:     "   ",
				Side:       TradeSideBuy,
				Quantity:   10.0,
				Price:      150.0,
				ExecutedAt: now,
				Mode:       "live",
				Source:     "manual",
			},
			expectedErr: true,
			description: "Whitespace-only symbol should fail",
		},
		{
			name: "zero quantity",
			trade: Trade{
				ID:         5,
				Symbol:     "AAPL.US",
				Side:       TradeSideBuy,
				Quantity:   0.0,
				Price:      150.0,
				ExecutedAt: now,
				Mode:       "live",
				Source:     "manual",
			},
			expectedErr: true,
			description: "Zero quantity should fail",
		},
		{
			name: "negative quantity",
			trade: Trade{
				ID:         6,
				Symbol:     "AAPL.US",
				Side:       TradeSideBuy,
				Quantity:   -10.0,
				Price:      150.0,
				ExecutedAt: now,
				Mode:       "live",
				Source:     "manual",
			},
			expectedErr: true,
			description: "Negative quantity should fail",
		},
		{
			name: "zero price",
			trade: Trade{
				ID:         7,
				Symbol:     "AAPL.US",
				Side:       TradeSideBuy,
				Quantity:   10.0,
				Price:      0.0,
				ExecutedAt: now,
				Mode:       "live",
				Source:     "manual",
			},
			expectedErr: true,
			description: "Zero price should fail",
		},
		{
			name: "negative price",
			trade: Trade{
				ID:         8,
				Symbol:     "AAPL.US",
				Side:       TradeSideBuy,
				Quantity:   10.0,
				Price:      -150.0,
				ExecutedAt: now,
				Mode:       "live",
				Source:     "manual",
			},
			expectedErr: true,
			description: "Negative price should fail",
		},
		{
			name: "symbol normalization lowercase",
			trade: Trade{
				ID:         9,
				Symbol:     "aapl.us",
				Side:       TradeSideBuy,
				Quantity:   10.0,
				Price:      150.0,
				ExecutedAt: now,
				Mode:       "live",
				Source:     "manual",
			},
			expectedErr: false,
			description: "Symbol should be normalized to uppercase",
		},
		{
			name: "symbol normalization with whitespace",
			trade: Trade{
				ID:         10,
				Symbol:     "  aapl.us  ",
				Side:       TradeSideBuy,
				Quantity:   10.0,
				Price:      150.0,
				ExecutedAt: now,
				Mode:       "live",
				Source:     "manual",
			},
			expectedErr: false,
			description: "Symbol should be trimmed and uppercased",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.trade.Validate()
			if tt.expectedErr {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				// Verify symbol normalization
				if tt.trade.Symbol != "" {
					assert.Equal(t, strings.ToUpper(strings.TrimSpace(tt.trade.Symbol)), tt.trade.Symbol,
						"Symbol should be normalized to uppercase and trimmed")
				}
			}
		})
	}
}

func TestTrade_Validate_SymbolNormalization(t *testing.T) {
	now := time.Now()

	trade := Trade{
		ID:         1,
		Symbol:     "  aapl.us  ",
		Side:       TradeSideBuy,
		Quantity:   10.0,
		Price:      150.0,
		ExecutedAt: now,
		Mode:       "live",
		Source:     "manual",
	}

	err := trade.Validate()
	require.NoError(t, err)
	assert.Equal(t, "AAPL.US", trade.Symbol, "Symbol should be normalized to uppercase and trimmed")
}

func TestTradeSide_Constants(t *testing.T) {
	// Verify constants are defined correctly
	assert.Equal(t, TradeSide("BUY"), TradeSideBuy)
	assert.Equal(t, TradeSide("SELL"), TradeSideSell)
	assert.True(t, TradeSideBuy.IsValid())
	assert.True(t, TradeSideSell.IsValid())
}
