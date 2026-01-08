package trading

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// TestCreate_ValidatesPrice tests that Create() validates price before insertion
func TestCreate_ValidatesPrice(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	// Create in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create trades table
	_, err = db.Exec(`
		CREATE TABLE trades (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			symbol TEXT NOT NULL,
			isin TEXT,
			side TEXT NOT NULL,
			quantity REAL NOT NULL,
			price REAL NOT NULL CHECK(price > 0),
			executed_at TEXT NOT NULL,
			order_id TEXT UNIQUE,
			currency TEXT,
			value_eur REAL,
			source TEXT NOT NULL,
			mode TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	repo := NewTradeRepository(db, log)

	testCases := []struct {
		name        string
		price       float64
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Valid positive price",
			price:       100.0,
			shouldError: false,
		},
		{
			name:        "Zero price should fail",
			price:       0.0,
			shouldError: true,
			errorMsg:    "price must be positive",
		},
		{
			name:        "Negative price should fail",
			price:       -10.0,
			shouldError: true,
			errorMsg:    "price must be positive",
		},
		{
			name:        "Small positive price should pass",
			price:       0.01,
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trade := Trade{
				OrderID:    "ORDER-" + tc.name,
				Symbol:     "AAPL",
				Side:       TradeSideBuy,
				Quantity:   10,
				Price:      tc.price,
				ExecutedAt: time.Now(),
				Source:     "test",
				Currency:   "EUR",
				Mode:       "test",
			}

			err := repo.Create(trade)

			if tc.shouldError {
				assert.Error(t, err, "Create should return error for invalid price")
				if tc.errorMsg != "" {
					// Validation should catch it before database constraint
					assert.Contains(t, err.Error(), tc.errorMsg, "Error should mention price validation: %s", err.Error())
				}
			} else {
				assert.NoError(t, err, "Create should succeed for valid price")
			}
		})
	}
}
