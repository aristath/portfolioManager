package portfolio

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSecurityProvider provides securities with overrides for tests
type testSecurityProvider struct {
	db  *sql.DB
	log zerolog.Logger
}

func newTestSecurityProvider(db *sql.DB, log zerolog.Logger) *testSecurityProvider {
	return &testSecurityProvider{db: db, log: log}
}

func (p *testSecurityProvider) GetAllActive() ([]SecurityInfo, error) {
	// Query securities
	query := `
		SELECT s.isin, s.symbol, s.name, s.geography, s.fullExchangeName, s.industry, s.currency
		FROM securities s
		WHERE s.active = 1
	`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query securities: %w", err)
	}
	defer rows.Close()

	var securities []SecurityInfo
	for rows.Next() {
		var sec SecurityInfo
		var geography, fullExchangeName, industry, currency sql.NullString

		if err := rows.Scan(&sec.ISIN, &sec.Symbol, &sec.Name, &geography, &fullExchangeName, &industry, &currency); err != nil {
			return nil, fmt.Errorf("failed to scan security: %w", err)
		}

		if geography.Valid {
			sec.Geography = geography.String
		}
		if fullExchangeName.Valid {
			sec.FullExchangeName = fullExchangeName.String
		}
		if industry.Valid {
			sec.Industry = industry.String
		}
		if currency.Valid {
			sec.Currency = currency.String
		}
		sec.AllowSell = true // default

		securities = append(securities, sec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating securities: %w", err)
	}

	// Apply overrides
	for i := range securities {
		overrides, err := p.getOverrides(securities[i].ISIN)
		if err != nil {
			p.log.Warn().Str("isin", securities[i].ISIN).Err(err).Msg("Failed to fetch overrides")
			continue
		}

		if len(overrides) > 0 {
			p.applyOverrides(&securities[i], overrides)
		}
	}

	return securities, nil
}

func (p *testSecurityProvider) GetAllActiveTradable() ([]SecurityInfo, error) {
	return p.GetAllActive()
}

func (p *testSecurityProvider) getOverrides(isin string) (map[string]string, error) {
	overrides := make(map[string]string)
	query := "SELECT field, value FROM security_overrides WHERE isin = ?"

	rows, err := p.db.Query(query, isin)
	if err != nil {
		return nil, fmt.Errorf("failed to query overrides: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var field, value string
		if err := rows.Scan(&field, &value); err != nil {
			return nil, fmt.Errorf("failed to scan override: %w", err)
		}
		overrides[field] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating overrides: %w", err)
	}

	return overrides, nil
}

func (p *testSecurityProvider) applyOverrides(sec *SecurityInfo, overrides map[string]string) {
	for field, value := range overrides {
		switch field {
		case "name":
			sec.Name = value
		case "geography":
			sec.Geography = value
		case "industry":
			sec.Industry = value
		case "currency":
			sec.Currency = value
		case "allow_sell":
			sec.AllowSell = value == "true" || value == "1"
		}
	}
}

// setupTestDBWithOverrides creates test databases with security_overrides table
func setupTestDBWithOverrides(t *testing.T) (*sql.DB, *sql.DB) {
	t.Helper()

	// Create portfolio database
	portfolioDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { portfolioDB.Close() })

	// Create positions table
	_, err = portfolioDB.Exec(`
		CREATE TABLE positions (
			isin TEXT PRIMARY KEY,
			symbol TEXT NOT NULL,
			quantity REAL NOT NULL,
			avg_price REAL NOT NULL,
			current_price REAL,
			currency TEXT NOT NULL DEFAULT 'EUR',
			currency_rate REAL NOT NULL DEFAULT 1.0,
			market_value_eur REAL,
			cost_basis_eur REAL,
			unrealized_pnl REAL,
			unrealized_pnl_pct REAL,
			last_updated INTEGER,
			first_bought INTEGER,
			last_sold INTEGER
		)
	`)
	require.NoError(t, err)

	// Create universe database
	universeDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { universeDB.Close() })

	// Create securities table
	_, err = universeDB.Exec(`
		CREATE TABLE securities (
			isin TEXT PRIMARY KEY,
			symbol TEXT NOT NULL,
			name TEXT NOT NULL,
			product_type TEXT,
			industry TEXT,
			geography TEXT,
			fullExchangeName TEXT,
			market_code TEXT,
			active INTEGER NOT NULL DEFAULT 1,
			currency TEXT,
			last_synced INTEGER,
			min_portfolio_target REAL,
			max_portfolio_target REAL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	require.NoError(t, err)

	// Create security_overrides table
	_, err = universeDB.Exec(`
		CREATE TABLE security_overrides (
			isin TEXT NOT NULL,
			field TEXT NOT NULL,
			value TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			PRIMARY KEY (isin, field),
			FOREIGN KEY (isin) REFERENCES securities(isin) ON DELETE CASCADE
		)
	`)
	require.NoError(t, err)

	return portfolioDB, universeDB
}

// insertSecurityWithoutOverrideColumns inserts a security without override columns
func insertSecurityWithoutOverrideColumns(t *testing.T, db *sql.DB, isin, symbol, name string) {
	t.Helper()

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO securities (isin, symbol, name, active, product_type, created_at, updated_at)
		VALUES (?, ?, ?, 1, 'STOCK', ?, ?)
	`, isin, symbol, name, now, now)
	require.NoError(t, err)
}

// insertOverride inserts an override record
func insertOverride(t *testing.T, db *sql.DB, isin, field, value string) {
	t.Helper()

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO security_overrides (isin, field, value, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(isin, field) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at
	`, isin, field, value, now, now)
	require.NoError(t, err)
}

func TestGetWithSecurityInfo_AppliesGeographyOverride(t *testing.T) {
	portfolioDB, universeDB := setupTestDBWithOverrides(t)

	// Insert security with US geography
	insertSecurityWithoutOverrideColumns(t, universeDB, "US0378331005", "AAPL.US", "Apple Inc.")
	_, err := universeDB.Exec(`UPDATE securities SET geography = ? WHERE isin = ?`, "US", "US0378331005")
	require.NoError(t, err)

	// Override geography to WORLD
	insertOverride(t, universeDB, "US0378331005", "geography", "WORLD")

	// Insert position for this security
	now := time.Now().Unix()
	_, err = portfolioDB.Exec(`INSERT INTO positions (isin, symbol, quantity, avg_price, currency, currency_rate, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, "US0378331005", "AAPL.US", 100.0, 150.0, "USD", 1.1, now)
	require.NoError(t, err)

	// Create repos with override support
	securityProvider := newTestSecurityProvider(universeDB, zerolog.Nop())
	positionRepo := NewPositionRepository(portfolioDB, universeDB, securityProvider, zerolog.Nop())

	// Execute
	positions, err := positionRepo.GetWithSecurityInfo()
	require.NoError(t, err)
	require.Len(t, positions, 1)

	// Assert override applied
	assert.Equal(t, "WORLD", positions[0].Geography, "Geography override not applied")
}

func TestGetWithSecurityInfo_AppliesIndustryOverride(t *testing.T) {
	portfolioDB, universeDB := setupTestDBWithOverrides(t)

	// Insert security with Technology industry
	insertSecurityWithoutOverrideColumns(t, universeDB, "US0378331005", "AAPL.US", "Apple Inc.")
	_, err := universeDB.Exec(`UPDATE securities SET industry = ? WHERE isin = ?`, "Technology", "US0378331005")
	require.NoError(t, err)

	// Override industry to Finance
	insertOverride(t, universeDB, "US0378331005", "industry", "Finance")

	// Insert position for this security
	now := time.Now().Unix()
	_, err = portfolioDB.Exec(`INSERT INTO positions (isin, symbol, quantity, avg_price, currency, currency_rate, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, "US0378331005", "AAPL.US", 100.0, 150.0, "USD", 1.1, now)
	require.NoError(t, err)

	// Create repos with override support
	securityProvider := newTestSecurityProvider(universeDB, zerolog.Nop())
	positionRepo := NewPositionRepository(portfolioDB, universeDB, securityProvider, zerolog.Nop())

	// Execute
	positions, err := positionRepo.GetWithSecurityInfo()
	require.NoError(t, err)
	require.Len(t, positions, 1)

	// Assert override applied
	assert.Equal(t, "Finance", positions[0].Industry, "Industry override not applied")
}

func TestGetWithSecurityInfo_AppliesNameOverride(t *testing.T) {
	portfolioDB, universeDB := setupTestDBWithOverrides(t)

	// Insert security with original name
	insertSecurityWithoutOverrideColumns(t, universeDB, "US0378331005", "AAPL.US", "Apple Inc.")

	// Override name to custom name
	insertOverride(t, universeDB, "US0378331005", "name", "Apple Custom Name")

	// Insert position for this security
	now := time.Now().Unix()
	_, err := portfolioDB.Exec(`INSERT INTO positions (isin, symbol, quantity, avg_price, currency, currency_rate, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, "US0378331005", "AAPL.US", 100.0, 150.0, "USD", 1.1, now)
	require.NoError(t, err)

	// Create repos with override support
	securityProvider := newTestSecurityProvider(universeDB, zerolog.Nop())
	positionRepo := NewPositionRepository(portfolioDB, universeDB, securityProvider, zerolog.Nop())

	// Execute
	positions, err := positionRepo.GetWithSecurityInfo()
	require.NoError(t, err)
	require.Len(t, positions, 1)

	// Assert override applied
	assert.Equal(t, "Apple Custom Name", positions[0].StockName, "Name override not applied")
}

func TestGetWithSecurityInfo_WithoutSecurityRepo_UsesFallback(t *testing.T) {
	portfolioDB, universeDB := setupTestDBWithOverrides(t)

	// Insert security with US geography
	insertSecurityWithoutOverrideColumns(t, universeDB, "US0378331005", "AAPL.US", "Apple Inc.")
	_, err := universeDB.Exec(`UPDATE securities SET geography = ? WHERE isin = ?`, "US", "US0378331005")
	require.NoError(t, err)

	// Override geography to WORLD (should be ignored without SecurityRepo)
	insertOverride(t, universeDB, "US0378331005", "geography", "WORLD")

	// Insert position for this security
	now := time.Now().Unix()
	_, err = portfolioDB.Exec(`INSERT INTO positions (isin, symbol, quantity, avg_price, currency, currency_rate, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, "US0378331005", "AAPL.US", 100.0, 150.0, "USD", 1.1, now)
	require.NoError(t, err)

	// Create repo WITHOUT override support (nil SecurityRepo)
	positionRepo := NewPositionRepository(portfolioDB, universeDB, nil, zerolog.Nop())

	// Execute
	positions, err := positionRepo.GetWithSecurityInfo()
	require.NoError(t, err)
	require.Len(t, positions, 1)

	// Assert override NOT applied (fallback to direct query)
	assert.Equal(t, "US", positions[0].Geography, "Should use base value when SecurityRepo is nil")
}
