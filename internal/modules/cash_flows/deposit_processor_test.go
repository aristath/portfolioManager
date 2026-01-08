package cash_flows

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// TestShouldProcessCashFlow tests deposit detection logic
func TestShouldProcessCashFlow(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	processor := NewDepositProcessor(nil, log)

	tests := []struct {
		name            string
		transactionType *string
		expected        bool
	}{
		{
			name:            "deposit type",
			transactionType: stringPtr("deposit"),
			expected:        true,
		},
		{
			name:            "DEPOSIT uppercase",
			transactionType: stringPtr("DEPOSIT"),
			expected:        true,
		},
		{
			name:            "refill type",
			transactionType: stringPtr("refill"),
			expected:        true,
		},
		{
			name:            "REFILL uppercase",
			transactionType: stringPtr("REFILL"),
			expected:        true,
		},
		{
			name:            "transfer_in type",
			transactionType: stringPtr("transfer_in"),
			expected:        true,
		},
		{
			name:            "TRANSFER_IN uppercase",
			transactionType: stringPtr("TRANSFER_IN"),
			expected:        true,
		},
		{
			name:            "dividend type should not process",
			transactionType: stringPtr("dividend"),
			expected:        false,
		},
		{
			name:            "withdrawal type should not process",
			transactionType: stringPtr("withdrawal"),
			expected:        false,
		},
		{
			name:            "nil transaction type",
			transactionType: nil,
			expected:        false,
		},
		{
			name:            "empty string",
			transactionType: stringPtr(""),
			expected:        false,
		},
		{
			name:            "unknown type",
			transactionType: stringPtr("unknown"),
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cashFlow := &CashFlow{
				TransactionType: tt.transactionType,
			}

			result := processor.ShouldProcessCashFlow(cashFlow)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProcessDeposit_NilCashManager tests deposit with nil cash manager (graceful degradation)
// This tests the fallback behavior when CashSecurityManager is not available
func TestProcessDeposit_NilCashManager(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	processor := NewDepositProcessor(nil, log)

	result, err := processor.ProcessDeposit(500.0, "EUR", stringPtr("tx-123"), stringPtr("Test deposit"))

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// When cashManager is nil, it returns the deposit amount as total (graceful degradation)
	assert.Equal(t, 500.0, result["total"])
}

// TestProcessDeposit_NilDescription tests deposit with nil description
func TestProcessDeposit_NilDescription(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	processor := NewDepositProcessor(nil, log)

	result, err := processor.ProcessDeposit(500.0, "EUR", stringPtr("tx-123"), nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 500.0, result["total"])
}

// TestProcessDeposit_ZeroAmount tests deposit with zero amount
func TestProcessDeposit_ZeroAmount(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	processor := NewDepositProcessor(nil, log)

	result, err := processor.ProcessDeposit(0.0, "EUR", stringPtr("tx-123"), stringPtr("Zero deposit"))

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0.0, result["total"])
}

// TestProcessDeposit_NegativeAmount tests deposit with negative amount (withdrawal scenario)
func TestProcessDeposit_NegativeAmount(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	processor := NewDepositProcessor(nil, log)

	result, err := processor.ProcessDeposit(-500.0, "EUR", stringPtr("tx-123"), stringPtr("Withdrawal"))

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Negative amounts are passed through (validation should happen elsewhere)
	assert.Equal(t, -500.0, result["total"])
}

// TestProcessDeposit_DifferentCurrencies tests deposit processing with various currencies
func TestProcessDeposit_DifferentCurrencies(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	processor := NewDepositProcessor(nil, log)

	currencies := []string{"EUR", "USD", "GBP", "JPY"}

	for _, currency := range currencies {
		t.Run(currency, func(t *testing.T) {
			result, err := processor.ProcessDeposit(100.0, currency, stringPtr("tx-"+currency), stringPtr("Test"))

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, 100.0, result["total"])
		})
	}
}

// TestProcessDeposit_SmallAmount tests deposit with very small amount (rounding edge case)
func TestProcessDeposit_SmallAmount(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	processor := NewDepositProcessor(nil, log)

	// Test very small amounts
	smallAmounts := []float64{0.01, 0.001, 0.0001}

	for i, amount := range smallAmounts {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			result, err := processor.ProcessDeposit(amount, "EUR", stringPtr("tx-small"), stringPtr("Small deposit"))

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.InDelta(t, amount, result["total"], 0.0001)
		})
	}
}

// TestProcessDeposit_LargeAmount tests deposit with large amount
func TestProcessDeposit_LargeAmount(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	processor := NewDepositProcessor(nil, log)

	largeAmount := 1000000.0
	result, err := processor.ProcessDeposit(largeAmount, "EUR", stringPtr("tx-large"), stringPtr("Large deposit"))

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, largeAmount, result["total"])
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
