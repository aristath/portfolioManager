package universe

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/rs/zerolog"
)

// ClientSymbolRepository handles client symbol mapping database operations.
// Implements domain.ClientSymbolRepositoryInterface following clean architecture.
// Used for mapping ISINs to client-specific symbols (brokers, data providers, etc.).
type ClientSymbolRepository struct {
	universeDB *sql.DB // universe.db - client_symbols table
	log        zerolog.Logger
}

// NewClientSymbolRepository creates a new client symbol repository.
func NewClientSymbolRepository(universeDB *sql.DB, log zerolog.Logger) *ClientSymbolRepository {
	return &ClientSymbolRepository{
		universeDB: universeDB,
		log:        log.With().Str("repo", "client_symbol").Logger(),
	}
}

// Compile-time check that ClientSymbolRepository implements ClientSymbolRepositoryInterface
var _ domain.ClientSymbolRepositoryInterface = (*ClientSymbolRepository)(nil)

// GetClientSymbol returns the client-specific symbol for a given ISIN and client name.
// Returns error if mapping doesn't exist (fail-fast approach).
func (r *ClientSymbolRepository) GetClientSymbol(isin, clientName string) (string, error) {
	query := "SELECT client_symbol FROM client_symbols WHERE isin = ? AND client_name = ?"

	isin = strings.ToUpper(strings.TrimSpace(isin))
	clientName = strings.ToLower(strings.TrimSpace(clientName))

	var symbol string
	err := r.universeDB.QueryRow(query, isin, clientName).Scan(&symbol)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("client symbol mapping not found for ISIN %s and client %s", isin, clientName)
		}
		return "", fmt.Errorf("failed to query client symbol: %w", err)
	}

	return symbol, nil
}

// SetClientSymbol creates or updates a client symbol mapping.
// Replaces existing mapping if present (upsert operation).
func (r *ClientSymbolRepository) SetClientSymbol(isin, clientName, symbol string) error {
	isin = strings.ToUpper(strings.TrimSpace(isin))
	clientName = strings.ToLower(strings.TrimSpace(clientName))
	symbol = strings.TrimSpace(symbol)

	if symbol == "" {
		return fmt.Errorf("client symbol cannot be empty")
	}

	query := `
		INSERT INTO client_symbols (isin, client_name, client_symbol)
		VALUES (?, ?, ?)
		ON CONFLICT(isin, client_name) DO UPDATE SET client_symbol = excluded.client_symbol
	`

	_, err := r.universeDB.Exec(query, isin, clientName, symbol)
	if err != nil {
		return fmt.Errorf("failed to set client symbol: %w", err)
	}

	r.log.Debug().
		Str("isin", isin).
		Str("client_name", clientName).
		Str("client_symbol", symbol).
		Msg("Client symbol mapping set")

	return nil
}

// GetAllClientSymbols returns all client symbols for a given ISIN.
// Returns a map of client_name -> client_symbol.
func (r *ClientSymbolRepository) GetAllClientSymbols(isin string) (map[string]string, error) {
	query := "SELECT client_name, client_symbol FROM client_symbols WHERE isin = ?"

	isin = strings.ToUpper(strings.TrimSpace(isin))

	rows, err := r.universeDB.Query(query, isin)
	if err != nil {
		return nil, fmt.Errorf("failed to query client symbols: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var clientName, symbol string
		if err := rows.Scan(&clientName, &symbol); err != nil {
			return nil, fmt.Errorf("failed to scan client symbol: %w", err)
		}
		result[clientName] = symbol
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating client symbols: %w", err)
	}

	return result, nil
}

// GetISINByClientSymbol performs reverse lookup: finds ISIN by client symbol.
// Returns error if mapping doesn't exist.
func (r *ClientSymbolRepository) GetISINByClientSymbol(clientName, clientSymbol string) (string, error) {
	query := "SELECT isin FROM client_symbols WHERE client_name = ? AND client_symbol = ?"

	clientName = strings.ToLower(strings.TrimSpace(clientName))
	clientSymbol = strings.TrimSpace(clientSymbol)

	var isin string
	err := r.universeDB.QueryRow(query, clientName, clientSymbol).Scan(&isin)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("ISIN not found for client %s symbol %s", clientName, clientSymbol)
		}
		return "", fmt.Errorf("failed to query ISIN by client symbol: %w", err)
	}

	return isin, nil
}

// DeleteClientSymbol removes a client symbol mapping for an ISIN/client pair.
func (r *ClientSymbolRepository) DeleteClientSymbol(isin, clientName string) error {
	query := "DELETE FROM client_symbols WHERE isin = ? AND client_name = ?"

	isin = strings.ToUpper(strings.TrimSpace(isin))
	clientName = strings.ToLower(strings.TrimSpace(clientName))

	result, err := r.universeDB.Exec(query, isin, clientName)
	if err != nil {
		return fmt.Errorf("failed to delete client symbol: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		r.log.Debug().
			Str("isin", isin).
			Str("client_name", clientName).
			Msg("Client symbol mapping deleted")
	}

	return nil
}
