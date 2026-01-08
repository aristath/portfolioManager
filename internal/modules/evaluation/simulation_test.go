package evaluation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimulateSequence_BuyAction(t *testing.T) {
	// Setup
	initialCash := 1000.0
	portfolioContext := PortfolioContext{
		Positions:       make(map[string]float64),
		TotalValue:      500.0,
		CountryWeights:  make(map[string]float64),
		IndustryWeights: make(map[string]float64),
	}

	securities := []Security{
		{
			Symbol:  "AAPL",
			Name:    "Apple Inc.",
			Country: stringPtr("United States"),
		},
	}

	// Create a BUY action
	buyAction := ActionCandidate{
		Side:     TradeSideBuy,
		Symbol:   "AAPL",
		Quantity: 10,
		Price:    150.0,
		ValueEUR: 1500.0,
		Currency: "USD",
	}

	sequence := []ActionCandidate{buyAction}

	// Execute
	endPortfolio, endCash := SimulateSequence(
		sequence,
		portfolioContext,
		initialCash,
		securities,
		nil, // No price adjustments
	)

	// Can't afford the full amount (only 1000 EUR available, need 1500)
	// Should skip the action
	assert.Equal(t, initialCash, endCash, "Cash should remain unchanged when buy is unaffordable")
	assert.Equal(t, 0.0, endPortfolio.Positions["AAPL"], "Position should not be created")
}

func TestSimulateSequence_AffordableBuy(t *testing.T) {
	// Setup
	initialCash := 2000.0
	portfolioContext := PortfolioContext{
		Positions:       make(map[string]float64),
		TotalValue:      500.0,
		CountryWeights:  make(map[string]float64),
		IndustryWeights: make(map[string]float64),
	}

	securities := []Security{
		{
			Symbol:  "AAPL",
			Name:    "Apple Inc.",
			Country: stringPtr("United States"),
		},
	}

	// Create an affordable BUY action
	buyAction := ActionCandidate{
		Side:     TradeSideBuy,
		Symbol:   "AAPL",
		Quantity: 10,
		Price:    150.0,
		ValueEUR: 1500.0,
		Currency: "USD",
	}

	sequence := []ActionCandidate{buyAction}

	// Execute
	endPortfolio, endCash := SimulateSequence(
		sequence,
		portfolioContext,
		initialCash,
		securities,
		nil,
	)

	// Assertions
	assert.Equal(t, 500.0, endCash, "Cash should decrease by buy value")
	assert.Equal(t, 1500.0, endPortfolio.Positions["AAPL"], "Position should be created with correct value")
	assert.Equal(t, "United States", endPortfolio.SecurityCountries["AAPL"], "Country should be set")
}

func TestSimulateSequence_SellAction(t *testing.T) {
	// Setup - portfolio with existing position
	initialCash := 1000.0
	portfolioContext := PortfolioContext{
		Positions: map[string]float64{
			"AAPL": 2000.0, // Existing position worth 2000 EUR
		},
		TotalValue:      3000.0,
		CountryWeights:  make(map[string]float64),
		IndustryWeights: make(map[string]float64),
		SecurityCountries: map[string]string{
			"AAPL": "United States",
		},
	}

	securities := []Security{
		{
			Symbol:  "AAPL",
			Name:    "Apple Inc.",
			Country: stringPtr("United States"),
		},
	}

	// Create a SELL action
	sellAction := ActionCandidate{
		Side:     TradeSideSell,
		Symbol:   "AAPL",
		Quantity: 5,
		Price:    150.0,
		ValueEUR: 750.0,
		Currency: "USD",
	}

	sequence := []ActionCandidate{sellAction}

	// Execute
	endPortfolio, endCash := SimulateSequence(
		sequence,
		portfolioContext,
		initialCash,
		securities,
		nil,
	)

	// Assertions
	assert.Equal(t, 1750.0, endCash, "Cash should increase by sell value")
	assert.Equal(t, 1250.0, endPortfolio.Positions["AAPL"], "Position should decrease by sell value")
}

func TestSimulateSequence_SellEntirePosition(t *testing.T) {
	// Setup
	initialCash := 1000.0
	portfolioContext := PortfolioContext{
		Positions: map[string]float64{
			"AAPL": 1500.0,
		},
		TotalValue:      2500.0,
		CountryWeights:  make(map[string]float64),
		IndustryWeights: make(map[string]float64),
	}

	securities := []Security{
		{
			Symbol:  "AAPL",
			Name:    "Apple Inc.",
			Country: stringPtr("United States"),
		},
	}

	// Sell entire position
	sellAction := ActionCandidate{
		Side:     TradeSideSell,
		Symbol:   "AAPL",
		Quantity: 10,
		Price:    150.0,
		ValueEUR: 1500.0,
		Currency: "USD",
	}

	sequence := []ActionCandidate{sellAction}

	// Execute
	endPortfolio, endCash := SimulateSequence(
		sequence,
		portfolioContext,
		initialCash,
		securities,
		nil,
	)

	// Assertions
	assert.Equal(t, 2500.0, endCash, "Cash should increase by full sell value")
	_, exists := endPortfolio.Positions["AAPL"]
	assert.False(t, exists, "Position should be removed when sold entirely")
}

func TestCheckSequenceFeasibility_Feasible(t *testing.T) {
	sequence := []ActionCandidate{
		{Side: TradeSideBuy, ValueEUR: 500.0},
		{Side: TradeSideBuy, ValueEUR: 300.0},
	}

	feasible := CheckSequenceFeasibility(
		sequence,
		1000.0, // Enough cash
		PortfolioContext{},
	)

	assert.True(t, feasible, "Sequence should be feasible with sufficient cash")
}

func TestCheckSequenceFeasibility_NotFeasible(t *testing.T) {
	sequence := []ActionCandidate{
		{Side: TradeSideBuy, ValueEUR: 500.0},
		{Side: TradeSideBuy, ValueEUR: 600.0},
	}

	feasible := CheckSequenceFeasibility(
		sequence,
		1000.0, // Not enough cash for both buys
		PortfolioContext{},
	)

	assert.False(t, feasible, "Sequence should be infeasible with insufficient cash")
}

func TestCheckSequenceFeasibility_SellThenBuy(t *testing.T) {
	sequence := []ActionCandidate{
		{Side: TradeSideSell, ValueEUR: 500.0}, // Adds 500
		{Side: TradeSideBuy, ValueEUR: 1200.0}, // Needs 1200, have 500+500=1000
	}

	feasible := CheckSequenceFeasibility(
		sequence,
		500.0, // Initial cash
		PortfolioContext{},
	)

	assert.False(t, feasible, "Sequence should be infeasible even with sell proceeds")
}

func TestCalculateSequenceCashFlow(t *testing.T) {
	sequence := []ActionCandidate{
		{Side: TradeSideSell, ValueEUR: 500.0},
		{Side: TradeSideBuy, ValueEUR: 300.0},
		{Side: TradeSideSell, ValueEUR: 200.0},
		{Side: TradeSideBuy, ValueEUR: 150.0},
	}

	cashFlow := CalculateSequenceCashFlow(sequence)

	assert.Equal(t, 700.0, cashFlow.CashGenerated, "Should sum all sells")
	assert.Equal(t, 450.0, cashFlow.CashRequired, "Should sum all buys")
	assert.Equal(t, 250.0, cashFlow.NetCashFlow, "Net flow should be positive")
}

// Helper function
func stringPtr(s string) *string {
	return &s
}

func TestCopyMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]float64
		expected map[string]float64
		desc     string
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: make(map[string]float64),
			desc:     "Nil map should return empty map",
		},
		{
			name:     "empty map",
			input:    make(map[string]float64),
			expected: make(map[string]float64),
			desc:     "Empty map should return empty map",
		},
		{
			name:     "map with values",
			input:    map[string]float64{"AAPL": 1000.0, "MSFT": 500.0},
			expected: map[string]float64{"AAPL": 1000.0, "MSFT": 500.0},
			desc:     "Map with values should be copied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := copyMap(tt.input)
			assert.Equal(t, tt.expected, result, tt.desc)

			// Verify it's a deep copy (modifying result shouldn't affect input)
			if len(tt.input) > 0 {
				result["NEW_KEY"] = 999.0
				_, exists := tt.input["NEW_KEY"]
				assert.False(t, exists, "Modifying copy should not affect original")
			}
		})
	}
}

func TestCopyStringMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected map[string]string
		desc     string
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: make(map[string]string),
			desc:     "Nil map should return empty map",
		},
		{
			name:     "empty map",
			input:    make(map[string]string),
			expected: make(map[string]string),
			desc:     "Empty map should return empty map",
		},
		{
			name:     "map with values",
			input:    map[string]string{"AAPL": "US", "MSFT": "US"},
			expected: map[string]string{"AAPL": "US", "MSFT": "US"},
			desc:     "Map with values should be copied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := copyStringMap(tt.input)
			assert.Equal(t, tt.expected, result, tt.desc)

			// Verify it's a deep copy (modifying result shouldn't affect input)
			if len(tt.input) > 0 {
				result["NEW_KEY"] = "NEW_VALUE"
				_, exists := tt.input["NEW_KEY"]
				assert.False(t, exists, "Modifying copy should not affect original")
			}
		})
	}
}

func TestMaxFloat64(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		b        float64
		expected float64
	}{
		{
			name:     "a greater than b",
			a:        10.0,
			b:        5.0,
			expected: 10.0,
		},
		{
			name:     "b greater than a",
			a:        5.0,
			b:        10.0,
			expected: 10.0,
		},
		{
			name:     "equal values",
			a:        5.0,
			b:        5.0,
			expected: 5.0,
		},
		{
			name:     "negative values",
			a:        -5.0,
			b:        -10.0,
			expected: -5.0,
		},
		{
			name:     "one negative",
			a:        -5.0,
			b:        10.0,
			expected: 10.0,
		},
		{
			name:     "zero values",
			a:        0.0,
			b:        0.0,
			expected: 0.0,
		},
		{
			name:     "decimal values",
			a:        3.14,
			b:        2.71,
			expected: 3.14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maxFloat64(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
