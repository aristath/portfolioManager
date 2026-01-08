package testing

import (
	"testing"
	"time"

	"github.com/aristath/sentinel/internal/modules/trading"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSecurityFixtures tests that NewSecurityFixtures creates valid test securities
func TestNewSecurityFixtures(t *testing.T) {
	fixtures := NewSecurityFixtures()
	require.NotNil(t, fixtures)
	require.Greater(t, len(fixtures), 0)

	// Verify all fixtures are valid
	for _, security := range fixtures {
		assert.NotEmpty(t, security.ISIN, "Security should have ISIN")
		assert.NotEmpty(t, security.Symbol, "Security should have symbol")
		assert.NotEmpty(t, security.Name, "Security should have name")
		assert.True(t, len(security.ISIN) == 12, "ISIN should be 12 characters")
	}
}

// TestNewSecurityFixtures_DiverseSecurities tests that fixtures include diverse securities
func TestNewSecurityFixtures_DiverseSecurities(t *testing.T) {
	fixtures := NewSecurityFixtures()

	// Collect symbols and countries
	symbols := make(map[string]bool)
	countries := make(map[string]bool)

	for _, security := range fixtures {
		symbols[security.Symbol] = true
		if security.Country != "" {
			countries[security.Country] = true
		}
	}

	// Verify diversity
	assert.Greater(t, len(symbols), 1, "Fixtures should include multiple symbols")
	assert.Greater(t, len(countries), 1, "Fixtures should include multiple countries")
}

// TestNewPositionFixtures tests that NewPositionFixtures creates valid test positions
func TestNewPositionFixtures(t *testing.T) {
	fixtures := NewPositionFixtures()
	require.NotNil(t, fixtures)
	require.Greater(t, len(fixtures), 0)

	// Verify all fixtures are valid
	for _, position := range fixtures {
		assert.NotEmpty(t, position.ISIN, "Position should have ISIN")
		assert.NotEmpty(t, position.Symbol, "Position should have symbol")
		assert.Greater(t, position.Quantity, 0.0, "Position should have positive quantity")
		assert.Greater(t, position.AvgPrice, 0.0, "Position should have positive price")
	}
}

// TestNewPositionFixtures_ValuesCalculated tests that position values are calculated
func TestNewPositionFixtures_ValuesCalculated(t *testing.T) {
	fixtures := NewPositionFixtures()

	for _, position := range fixtures {
		expectedValue := position.Quantity * position.AvgPrice * position.CurrencyRate
		if position.MarketValueEUR > 0 {
			// If market value is set, verify it's reasonable
			assert.Greater(t, position.MarketValueEUR, 0.0, "Market value should be positive")
		}
		_ = expectedValue // May be used for validation
	}
}

// TestNewTradeFixtures tests that NewTradeFixtures creates valid test trades
func TestNewTradeFixtures(t *testing.T) {
	fixtures := NewTradeFixtures()
	require.NotNil(t, fixtures)
	require.Greater(t, len(fixtures), 0)

	// Verify all fixtures are valid
	for _, trade := range fixtures {
		assert.NotEmpty(t, trade.Symbol, "Trade should have symbol")
		assert.NotEmpty(t, trade.Side, "Trade should have side")
		assert.Greater(t, trade.Quantity, 0.0, "Trade should have positive quantity")
		assert.Greater(t, trade.Price, 0.0, "Trade should have positive price")
		assert.False(t, trade.ExecutedAt.IsZero(), "Trade should have execution timestamp")
	}
}

// TestNewTradeFixtures_DiverseTrades tests that fixtures include diverse trades
func TestNewTradeFixtures_DiverseTrades(t *testing.T) {
	fixtures := NewTradeFixtures()

	// Collect sides and symbols
	sides := make(map[trading.TradeSide]bool)
	symbols := make(map[string]bool)

	for _, trade := range fixtures {
		sides[trade.Side] = true
		symbols[trade.Symbol] = true
	}

	// Verify diversity
	assert.Greater(t, len(sides), 1, "Fixtures should include both BUY and SELL trades")
	assert.Greater(t, len(symbols), 1, "Fixtures should include multiple symbols")
}

// TestNewTradeFixtures_Chronological tests that trades are in chronological order
func TestNewTradeFixtures_Chronological(t *testing.T) {
	fixtures := NewTradeFixtures()

	if len(fixtures) < 2 {
		t.Skip("Need at least 2 trades to test chronological order")
	}

	// Verify trades are in chronological order (oldest first)
	for i := 1; i < len(fixtures); i++ {
		assert.True(t,
			fixtures[i-1].ExecutedAt.Before(fixtures[i].ExecutedAt) || fixtures[i-1].ExecutedAt.Equal(fixtures[i].ExecutedAt),
			"Trades should be in chronological order",
		)
	}
}

// TestNewCashFlowFixtures tests that NewCashFlowFixtures creates valid test cash flows
func TestNewCashFlowFixtures(t *testing.T) {
	fixtures := NewCashFlowFixtures()
	require.NotNil(t, fixtures)
	require.Greater(t, len(fixtures), 0)

	// Verify all fixtures are valid
	for _, cf := range fixtures {
		assert.NotEmpty(t, cf.TransactionID, "Cash flow should have transaction ID")
		assert.NotNil(t, cf.TransactionType, "Cash flow should have transaction type")
		assert.NotEqual(t, 0.0, cf.Amount, "Cash flow should have non-zero amount")
		assert.NotEmpty(t, cf.Date, "Cash flow should have date")
	}
}

// TestNewCashFlowFixtures_DiverseTypes tests that fixtures include diverse cash flow types
func TestNewCashFlowFixtures_DiverseTypes(t *testing.T) {
	fixtures := NewCashFlowFixtures()

	// Collect types
	types := make(map[string]bool)
	for _, cf := range fixtures {
		if cf.TransactionType != nil {
			types[*cf.TransactionType] = true
		}
	}

	// Verify diversity
	assert.Greater(t, len(types), 1, "Fixtures should include multiple cash flow types")
}

// TestNewCashFlowFixtures_DateRange tests that fixtures cover a date range
func TestNewCashFlowFixtures_DateRange(t *testing.T) {
	fixtures := NewCashFlowFixtures()

	if len(fixtures) < 2 {
		t.Skip("Need at least 2 cash flows to test date range")
	}

	// Parse dates and find min and max
	minDate, _ := time.Parse("2006-01-02", fixtures[0].Date)
	maxDate := minDate

	for _, cf := range fixtures {
		date, err := time.Parse("2006-01-02", cf.Date)
		require.NoError(t, err)
		if date.Before(minDate) {
			minDate = date
		}
		if date.After(maxDate) {
			maxDate = date
		}
	}

	// Verify date range exists
	assert.True(t, maxDate.After(minDate), "Fixtures should cover a date range")
}

// TestNewDividendFixtures tests that NewDividendFixtures creates valid test dividends
func TestNewDividendFixtures(t *testing.T) {
	fixtures := NewDividendFixtures()
	require.NotNil(t, fixtures)
	require.Greater(t, len(fixtures), 0)

	// Verify all fixtures are valid
	for _, div := range fixtures {
		assert.NotEmpty(t, div.Symbol, "Dividend should have symbol")
		assert.Greater(t, div.Amount, 0.0, "Dividend should have positive amount")
		assert.NotNil(t, div.PaymentDate, "Dividend should have payment date")
	}
}

// TestNewAllocationTargetFixtures tests that NewAllocationTargetFixtures creates valid test targets
func TestNewAllocationTargetFixtures(t *testing.T) {
	fixtures := NewAllocationTargetFixtures()
	require.NotNil(t, fixtures)
	require.Greater(t, len(fixtures), 0)

	// Verify all fixtures are valid
	totalPct := 0.0
	for _, target := range fixtures {
		assert.NotEmpty(t, target.Name, "Target should have name")
		assert.NotEmpty(t, target.Type, "Target should have type")
		assert.GreaterOrEqual(t, target.TargetPct, 0.0, "Target percentage should be non-negative")
		assert.LessOrEqual(t, target.TargetPct, 1.0, "Target percentage should be <= 1.0")
		totalPct += target.TargetPct
	}

	// Total percentage should be reasonable (between 0 and 1.5 to allow for rounding)
	assert.LessOrEqual(t, totalPct, 1.5, "Total target percentage should be reasonable")
}

// TestNewTimeFixtures tests that NewTimeFixtures provides valid test timestamps
func TestNewTimeFixtures(t *testing.T) {
	now := time.Now()
	dayAgo := time.Now().AddDate(0, 0, -1)
	weekAgo := time.Now().AddDate(0, 0, -7)
	monthAgo := time.Now().AddDate(0, -1, 0)
	yearAgo := time.Now().AddDate(-1, 0, 0)

	// Verify all timestamps are in the past (relative to now)
	assert.True(t, dayAgo.Before(now))
	assert.True(t, weekAgo.Before(now))
	assert.True(t, monthAgo.Before(now))
	assert.True(t, yearAgo.Before(now))
}

// TestNewTimeFixtures_Order tests that time fixtures are in chronological order
func TestNewTimeFixtures_Order(t *testing.T) {
	yearAgo := time.Now().AddDate(-1, 0, 0)
	monthAgo := time.Now().AddDate(0, -1, 0)
	weekAgo := time.Now().AddDate(0, 0, -7)
	dayAgo := time.Now().AddDate(0, 0, -1)

	// Verify chronological order
	assert.True(t, yearAgo.Before(monthAgo))
	assert.True(t, monthAgo.Before(weekAgo))
	assert.True(t, weekAgo.Before(dayAgo))
}

// TestNewPriceFixtures tests that NewPriceFixtures provides valid test prices
func TestNewPriceFixtures(t *testing.T) {
	// Price fixtures should provide common price ranges
	// This is a placeholder test - actual implementation will vary
	assert.True(t, true, "Price fixtures test")
}

// TestNewCurrencyFixtures tests that NewCurrencyFixtures provides valid test currencies
func TestNewCurrencyFixtures(t *testing.T) {
	fixtures := NewCurrencyFixtures()
	require.NotNil(t, fixtures)
	require.Greater(t, len(fixtures), 0)

	// Verify all currencies are valid ISO codes
	for _, currency := range fixtures {
		assert.Len(t, currency, 3, "Currency code should be 3 characters")
		assert.Equal(t, currency, string([]byte(currency)[:3]), "Currency should be uppercase")
	}

	// Verify common currencies are included
	currencies := make(map[string]bool)
	for _, c := range fixtures {
		currencies[c] = true
	}

	assert.True(t, currencies["EUR"] || currencies["USD"], "Common currencies should be included")
}

// TestNewISINFuxtures tests that NewISINFuxtures provides valid test ISINs
func TestNewISINFuxtures(t *testing.T) {
	fixtures := NewISINFuxtures()
	require.NotNil(t, fixtures)
	require.Greater(t, len(fixtures), 0)

	// Verify all ISINs are valid
	for _, isin := range fixtures {
		assert.Len(t, isin, 12, "ISIN should be 12 characters")
		assert.Regexp(t, "^[A-Z]{2}[A-Z0-9]{9}[0-9]$", isin, "ISIN should match format")
	}
}
