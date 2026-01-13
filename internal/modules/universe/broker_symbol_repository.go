package universe

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/rs/zerolog"
)

// BrokerSymbolRepository handles broker symbol mapping database operations.
// Implements domain.BrokerSymbolRepositoryInterface following clean architecture.
type BrokerSymbolRepository struct {
	universeDB *sql.DB // universe.db - broker_symbols table
	log        zerolog.Logger
}

// NewBrokerSymbolRepository creates a new broker symbol repository.
func NewBrokerSymbolRepository(universeDB *sql.DB, log zerolog.Logger) *BrokerSymbolRepository {
	return &BrokerSymbolRepository{
		universeDB: universeDB,
		log:        log.With().Str("repo", "broker_symbol").Logger(),
	}
}

// Compile-time check that BrokerSymbolRepository implements BrokerSymbolRepositoryInterface
var _ domain.BrokerSymbolRepositoryInterface = (*BrokerSymbolRepository)(nil)

// GetBrokerSymbol returns the broker-specific symbol for a given ISIN and broker name.
// Returns error if mapping doesn't exist (fail-fast approach).
func (r *BrokerSymbolRepository) GetBrokerSymbol(isin, brokerName string) (string, error) {
	query := "SELECT broker_symbol FROM broker_symbols WHERE isin = ? AND broker_name = ?"

	isin = strings.ToUpper(strings.TrimSpace(isin))
	brokerName = strings.ToLower(strings.TrimSpace(brokerName))

	var symbol string
	err := r.universeDB.QueryRow(query, isin, brokerName).Scan(&symbol)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("broker symbol mapping not found for ISIN %s and broker %s", isin, brokerName)
		}
		return "", fmt.Errorf("failed to query broker symbol: %w", err)
	}

	return symbol, nil
}

// SetBrokerSymbol creates or updates a broker symbol mapping.
// Replaces existing mapping if present (upsert operation).
func (r *BrokerSymbolRepository) SetBrokerSymbol(isin, brokerName, symbol string) error {
	isin = strings.ToUpper(strings.TrimSpace(isin))
	brokerName = strings.ToLower(strings.TrimSpace(brokerName))
	symbol = strings.TrimSpace(symbol)

	if symbol == "" {
		return fmt.Errorf("broker symbol cannot be empty")
	}

	query := `
		INSERT INTO broker_symbols (isin, broker_name, broker_symbol)
		VALUES (?, ?, ?)
		ON CONFLICT(isin, broker_name) DO UPDATE SET broker_symbol = excluded.broker_symbol
	`

	_, err := r.universeDB.Exec(query, isin, brokerName, symbol)
	if err != nil {
		return fmt.Errorf("failed to set broker symbol: %w", err)
	}

	r.log.Debug().
		Str("isin", isin).
		Str("broker_name", brokerName).
		Str("broker_symbol", symbol).
		Msg("Broker symbol mapping set")

	return nil
}

// GetAllBrokerSymbols returns all broker symbols for a given ISIN.
// Returns a map of broker_name -> broker_symbol.
func (r *BrokerSymbolRepository) GetAllBrokerSymbols(isin string) (map[string]string, error) {
	query := "SELECT broker_name, broker_symbol FROM broker_symbols WHERE isin = ?"

	isin = strings.ToUpper(strings.TrimSpace(isin))

	rows, err := r.universeDB.Query(query, isin)
	if err != nil {
		return nil, fmt.Errorf("failed to query broker symbols: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var brokerName, symbol string
		if err := rows.Scan(&brokerName, &symbol); err != nil {
			return nil, fmt.Errorf("failed to scan broker symbol: %w", err)
		}
		result[brokerName] = symbol
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating broker symbols: %w", err)
	}

	return result, nil
}

// GetISINByBrokerSymbol performs reverse lookup: finds ISIN by broker symbol.
// Returns error if mapping doesn't exist.
func (r *BrokerSymbolRepository) GetISINByBrokerSymbol(brokerName, brokerSymbol string) (string, error) {
	query := "SELECT isin FROM broker_symbols WHERE broker_name = ? AND broker_symbol = ?"

	brokerName = strings.ToLower(strings.TrimSpace(brokerName))
	brokerSymbol = strings.TrimSpace(brokerSymbol)

	var isin string
	err := r.universeDB.QueryRow(query, brokerName, brokerSymbol).Scan(&isin)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("ISIN not found for broker %s symbol %s", brokerName, brokerSymbol)
		}
		return "", fmt.Errorf("failed to query ISIN by broker symbol: %w", err)
	}

	return isin, nil
}

// DeleteBrokerSymbol removes a broker symbol mapping for an ISIN/broker pair.
func (r *BrokerSymbolRepository) DeleteBrokerSymbol(isin, brokerName string) error {
	query := "DELETE FROM broker_symbols WHERE isin = ? AND broker_name = ?"

	isin = strings.ToUpper(strings.TrimSpace(isin))
	brokerName = strings.ToLower(strings.TrimSpace(brokerName))

	result, err := r.universeDB.Exec(query, isin, brokerName)
	if err != nil {
		return fmt.Errorf("failed to delete broker symbol: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		r.log.Debug().
			Str("isin", isin).
			Str("broker_name", brokerName).
			Msg("Broker symbol mapping deleted")
	}

	return nil
}
