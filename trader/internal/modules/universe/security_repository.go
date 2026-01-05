package universe

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// SecurityRepository handles security database operations
// Faithful translation from Python: app/modules/universe/database/security_repository.py
type SecurityRepository struct {
	universeDB *sql.DB // universe.db - securities table
	tagRepo    *TagRepository
	log        zerolog.Logger
}

// securitiesColumns is the list of columns for the securities table
// Used to avoid SELECT * which can break when schema changes
// Column order must match the table schema (matches SELECT * order)
// After migration: isin is PRIMARY KEY, column order is isin, symbol, yahoo_symbol, ...
const securitiesColumns = `isin, symbol, yahoo_symbol, name, product_type, industry, country, fullExchangeName,
priority_multiplier, min_lot, active, allow_buy, allow_sell, currency, last_synced,
min_portfolio_target, max_portfolio_target, created_at, updated_at`

// NewSecurityRepository creates a new security repository
func NewSecurityRepository(universeDB *sql.DB, log zerolog.Logger) *SecurityRepository {
	return &SecurityRepository{
		universeDB: universeDB,
		tagRepo:    NewTagRepository(universeDB, log),
		log:        log.With().Str("repo", "security").Logger(),
	}
}

// GetBySymbol returns a security by symbol
// Faithful translation of Python: async def get_by_symbol(self, symbol: str) -> Optional[Security]
func (r *SecurityRepository) GetBySymbol(symbol string) (*Security, error) {
	query := "SELECT " + securitiesColumns + " FROM securities WHERE symbol = ?"

	rows, err := r.universeDB.Query(query, strings.ToUpper(strings.TrimSpace(symbol)))
	if err != nil {
		return nil, fmt.Errorf("failed to query security by symbol: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil // Security not found
	}

	security, err := r.scanSecurity(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan security: %w", err)
	}

	return &security, nil
}

// GetByISIN returns a security by ISIN
// Faithful translation of Python: async def get_by_isin(self, isin: str) -> Optional[Security]
func (r *SecurityRepository) GetByISIN(isin string) (*Security, error) {
	query := "SELECT " + securitiesColumns + " FROM securities WHERE isin = ?"

	rows, err := r.universeDB.Query(query, strings.ToUpper(strings.TrimSpace(isin)))
	if err != nil {
		return nil, fmt.Errorf("failed to query security by ISIN: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil // Security not found
	}

	security, err := r.scanSecurity(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan security: %w", err)
	}

	return &security, nil
}

// GetByIdentifier returns a security by symbol or ISIN (smart lookup)
// Faithful translation of Python: async def get_by_identifier(self, identifier: str) -> Optional[Security]
func (r *SecurityRepository) GetByIdentifier(identifier string) (*Security, error) {
	identifier = strings.ToUpper(strings.TrimSpace(identifier))

	// Check if it looks like an ISIN (12 chars, starts with 2 letters)
	if len(identifier) == 12 && len(identifier) >= 2 {
		firstTwo := identifier[:2]
		if (firstTwo[0] >= 'A' && firstTwo[0] <= 'Z') && (firstTwo[1] >= 'A' && firstTwo[1] <= 'Z') {
			// Try ISIN lookup first
			sec, err := r.GetByISIN(identifier)
			if err != nil {
				return nil, err
			}
			if sec != nil {
				return sec, nil
			}
		}
	}

	// Fall back to symbol lookup
	return r.GetBySymbol(identifier)
}

// GetAllActive returns all active securities
// Faithful translation of Python: async def get_all_active(self) -> List[Security]
func (r *SecurityRepository) GetAllActive() ([]Security, error) {
	query := "SELECT " + securitiesColumns + " FROM securities WHERE active = 1"

	rows, err := r.universeDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active securities: %w", err)
	}
	defer rows.Close()

	var securities []Security
	for rows.Next() {
		security, err := r.scanSecurity(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan security: %w", err)
		}
		securities = append(securities, security)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating securities: %w", err)
	}

	return securities, nil
}

// GetAllActiveTradable returns all active securities excluding cash
// Used for scoring and trading operations where cash should not be included
func (r *SecurityRepository) GetAllActiveTradable() ([]Security, error) {
	query := "SELECT " + securitiesColumns + " FROM securities WHERE active = 1 AND product_type != 'CASH'"

	rows, err := r.universeDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tradable securities: %w", err)
	}
	defer rows.Close()

	var securities []Security
	for rows.Next() {
		security, err := r.scanSecurity(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan security: %w", err)
		}
		securities = append(securities, security)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating securities: %w", err)
	}

	return securities, nil
}

// GetAll returns all securities (active and inactive)
// Faithful translation of Python: async def get_all(self) -> List[Security]
func (r *SecurityRepository) GetAll() ([]Security, error) {
	query := "SELECT " + securitiesColumns + " FROM securities"

	rows, err := r.universeDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all securities: %w", err)
	}
	defer rows.Close()

	var securities []Security
	for rows.Next() {
		security, err := r.scanSecurity(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan security: %w", err)
		}
		securities = append(securities, security)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating securities: %w", err)
	}

	return securities, nil
}

// Create creates a new security
// Faithful translation of Python: async def create(self, security: Security) -> None
func (r *SecurityRepository) Create(security Security) error {
	now := time.Now().Format(time.RFC3339)

	// Normalize symbol
	security.Symbol = strings.ToUpper(strings.TrimSpace(security.Symbol))

	// Begin transaction
	tx, err := r.universeDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		INSERT INTO securities
		(isin, symbol, yahoo_symbol, name, product_type, industry, country, fullExchangeName,
		 priority_multiplier, min_lot, active, allow_buy, allow_sell,
		 currency, min_portfolio_target, max_portfolio_target,
		 created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// ISIN is required (PRIMARY KEY)
	if security.ISIN == "" {
		return fmt.Errorf("ISIN is required for security creation")
	}

	_, err = tx.Exec(query,
		strings.ToUpper(strings.TrimSpace(security.ISIN)),
		security.Symbol,
		nullString(security.YahooSymbol),
		security.Name,
		nullString(security.ProductType),
		nullString(security.Industry),
		nullString(security.Country),
		nullString(security.FullExchangeName),
		security.PriorityMultiplier,
		security.MinLot,
		boolToInt(security.Active),
		boolToInt(security.AllowBuy),
		boolToInt(security.AllowSell),
		nullString(security.Currency),
		nullFloat64(security.MinPortfolioTarget),
		nullFloat64(security.MaxPortfolioTarget),
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert security: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.log.Info().Str("isin", security.ISIN).Str("symbol", security.Symbol).Msg("Security created")
	return nil
}

// Update updates security fields by ISIN
// Changed from symbol to ISIN as primary identifier
func (r *SecurityRepository) Update(isin string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Whitelist of allowed update fields
	allowedFields := map[string]bool{
		"active": true, "allow_buy": true, "allow_sell": true,
		"name": true, "product_type": true, "sector": true, "industry": true,
		"country": true, "fullExchangeName": true, "currency": true,
		"exchange": true, "market_cap": true, "pe_ratio": true,
		"dividend_yield": true, "beta": true, "52w_high": true, "52w_low": true,
		"min_portfolio_target": true, "max_portfolio_target": true,
		"isin": true, "min_lot": true, "priority_multiplier": true,
		"yahoo_symbol": true, "symbol": true,
	}

	// Validate all keys are in whitelist
	for key := range updates {
		if !allowedFields[key] {
			return fmt.Errorf("invalid update field: %s", key)
		}
	}

	// Add updated_at
	now := time.Now().Format(time.RFC3339)
	updates["updated_at"] = now

	// Convert booleans to integers
	for _, boolField := range []string{"active", "allow_buy", "allow_sell"} {
		if val, ok := updates[boolField]; ok {
			if boolVal, isBool := val.(bool); isBool {
				updates[boolField] = boolToInt(boolVal)
			}
		}
	}

	// Build SET clause
	var setClauses []string
	var values []interface{}
	for key, val := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", key))
		values = append(values, val)
	}
	values = append(values, strings.ToUpper(strings.TrimSpace(isin)))

	// Begin transaction
	tx, err := r.universeDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Safe: all keys are validated against whitelist above, values use parameterized query
	//nolint:gosec // G201: Field names are whitelisted, values are parameterized
	query := fmt.Sprintf("UPDATE securities SET %s WHERE isin = ?", strings.Join(setClauses, ", "))
	result, err := tx.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to update security: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	r.log.Info().Str("isin", isin).Int64("rows_affected", rowsAffected).Msg("Security updated")
	return nil
}

// Delete soft deletes a security by ISIN (sets active=0)
// Changed from symbol to ISIN as primary identifier
func (r *SecurityRepository) Delete(isin string) error {
	return r.Update(isin, map[string]interface{}{"active": false})
}

// GetWithScores returns all active securities with their scores and positions
// Faithful translation of Python: async def get_with_scores(self) -> List[dict]
// Note: This method accesses multiple databases (universe.db and portfolio.db) - architecture violation
func (r *SecurityRepository) GetWithScores(portfolioDB *sql.DB) ([]SecurityWithScore, error) {
	// Fetch securities from universe.db (excluding cash - cash doesn't have scores)
	securityRows, err := r.universeDB.Query("SELECT " + securitiesColumns + " FROM securities WHERE active = 1 AND product_type != 'CASH'")
	if err != nil {
		return nil, fmt.Errorf("failed to query securities: %w", err)
	}
	defer securityRows.Close()

	securitiesMap := make(map[string]SecurityWithScore)
	for securityRows.Next() {
		security, err := r.scanSecurity(securityRows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan security: %w", err)
		}

		// Convert to SecurityWithScore
		sws := SecurityWithScore{
			Symbol:             security.Symbol,
			Name:               security.Name,
			ISIN:               security.ISIN,
			YahooSymbol:        security.YahooSymbol,
			ProductType:        security.ProductType,
			Country:            security.Country,
			FullExchangeName:   security.FullExchangeName,
			Industry:           security.Industry,
			PriorityMultiplier: security.PriorityMultiplier,
			MinLot:             security.MinLot,
			Active:             security.Active,
			AllowBuy:           security.AllowBuy,
			AllowSell:          security.AllowSell,
			Currency:           security.Currency,
			LastSynced:         security.LastSynced,
			MinPortfolioTarget: security.MinPortfolioTarget,
			MaxPortfolioTarget: security.MaxPortfolioTarget,
			Tags:               security.Tags,
		}
		// Use ISIN as map key (primary identifier)
		securitiesMap[security.ISIN] = sws
	}

	if err := securityRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating securities: %w", err)
	}

	// Fetch scores from portfolio.db
	scoreRows, err := portfolioDB.Query("SELECT " + scoresColumns + " FROM scores")
	if err != nil {
		return nil, fmt.Errorf("failed to query scores: %w", err)
	}
	defer scoreRows.Close()

	scoresMap := make(map[string]SecurityScore)
	scoreRepo := NewScoreRepository(portfolioDB, r.log)
	for scoreRows.Next() {
		score, err := scoreRepo.scanScore(scoreRows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan score: %w", err)
		}
		// Convert symbol to ISIN for map key
		// Lookup ISIN from securities table using symbol
		security, err := r.GetBySymbol(score.Symbol)
		if err == nil && security != nil && security.ISIN != "" {
			scoresMap[security.ISIN] = score
		} else {
			// Fallback to symbol if ISIN lookup fails (shouldn't happen after migration)
			scoresMap[score.Symbol] = score
		}
	}

	if err := scoreRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating scores: %w", err)
	}

	// Fetch positions from portfolio.db
	positionRows, err := portfolioDB.Query(`SELECT symbol, quantity, avg_price, current_price, currency,
		currency_rate, market_value_eur, cost_basis_eur, unrealized_pnl,
		unrealized_pnl_pct, last_updated, first_bought, last_sold, isin
		FROM positions`)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer positionRows.Close()

	positionsMap := make(map[string]struct {
		marketValueEUR float64
		quantity       float64
	})

	for positionRows.Next() {
		var symbol string
		var quantity, marketValueEUR sql.NullFloat64
		// We only need symbol, quantity, and market_value_eur - scan minimal fields
		var avgPrice, currentPrice, currencyRate sql.NullFloat64
		var currency, lastUpdated sql.NullString
		var costBasis, unrealizedPnL, unrealizedPnLPct sql.NullFloat64
		var firstBought, lastSold, isin sql.NullString

		err := positionRows.Scan(
			&symbol, &quantity, &avgPrice, &currentPrice, &currency, &currencyRate,
			&marketValueEUR, &costBasis, &unrealizedPnL, &unrealizedPnLPct,
			&lastUpdated, &firstBought, &lastSold, &isin,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}

		// Convert symbol to ISIN for map key
		// Use ISIN from position if available, otherwise lookup
		var mapKey string
		if isin.Valid && isin.String != "" {
			mapKey = isin.String
		} else {
			// Lookup ISIN from securities table using symbol
			security, err := r.GetBySymbol(symbol)
			if err == nil && security != nil && security.ISIN != "" {
				mapKey = security.ISIN
			} else {
				// Fallback to symbol if ISIN lookup fails (for CASH positions)
				mapKey = symbol
			}
		}

		positionsMap[mapKey] = struct {
			marketValueEUR float64
			quantity       float64
		}{
			marketValueEUR: marketValueEUR.Float64,
			quantity:       quantity.Float64,
		}
	}

	if err := positionRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating positions: %w", err)
	}

	// Merge data
	var result []SecurityWithScore
	for isin, sws := range securitiesMap {
		// Add score data (scoresMap now uses ISIN as key)
		if score, found := scoresMap[isin]; found {
			sws.TotalScore = &score.TotalScore
			sws.QualityScore = &score.QualityScore
			sws.OpportunityScore = &score.OpportunityScore
			sws.AnalystScore = &score.AnalystScore
			sws.AllocationFitScore = &score.AllocationFitScore
			sws.Volatility = &score.Volatility
			sws.CAGRScore = &score.CAGRScore
			sws.ConsistencyScore = &score.ConsistencyScore
			sws.HistoryYears = &score.HistoryYears
			sws.TechnicalScore = &score.TechnicalScore
			sws.FundamentalScore = &score.FundamentalScore
		}

		// Add position data (positionsMap now uses ISIN as key)
		if pos, found := positionsMap[isin]; found {
			sws.PositionValue = &pos.marketValueEUR
			sws.PositionQuantity = &pos.quantity
		} else {
			zero := 0.0
			sws.PositionValue = &zero
			sws.PositionQuantity = &zero
		}

		result = append(result, sws)
	}

	return result, nil
}

// scanSecurity scans a database row into a Security struct
func (r *SecurityRepository) scanSecurity(rows *sql.Rows) (Security, error) {
	var security Security
	var yahooSymbol, isin, productType, country, fullExchangeName sql.NullString
	var industry, currency, lastSynced sql.NullString
	var minPortfolioTarget, maxPortfolioTarget sql.NullFloat64
	var active, allowBuy, allowSell sql.NullInt64
	var createdAt, updatedAt sql.NullString

	// Table schema after migration: isin, symbol, yahoo_symbol, name, product_type, industry, country, fullExchangeName,
	// priority_multiplier, min_lot, active, allow_buy, allow_sell, currency, last_synced,
	// min_portfolio_target, max_portfolio_target, created_at, updated_at
	var symbol sql.NullString
	err := rows.Scan(
		&isin,           // isin (PRIMARY KEY)
		&symbol,         // symbol
		&yahooSymbol,    // yahoo_symbol
		&security.Name,  // name
		&productType,    // product_type
		&industry,       // industry
		&country,        // country
		&fullExchangeName, // fullExchangeName
		&security.PriorityMultiplier, // priority_multiplier
		&security.MinLot, // min_lot
		&active,         // active
		&allowBuy,       // allow_buy
		&allowSell,      // allow_sell
		&currency,       // currency
		&lastSynced,     // last_synced
		&minPortfolioTarget, // min_portfolio_target
		&maxPortfolioTarget, // max_portfolio_target
		&createdAt,      // created_at
		&updatedAt,      // updated_at
	)
	if err != nil {
		return security, err
	}

	// Handle nullable fields
	if isin.Valid {
		security.ISIN = isin.String
	}
	if symbol.Valid {
		security.Symbol = symbol.String
	}
	if yahooSymbol.Valid {
		security.YahooSymbol = yahooSymbol.String
	}
	if productType.Valid {
		security.ProductType = productType.String
	}
	if country.Valid {
		security.Country = country.String
	}
	if fullExchangeName.Valid {
		security.FullExchangeName = fullExchangeName.String
	}
	if industry.Valid {
		security.Industry = industry.String
	}
	if currency.Valid {
		security.Currency = currency.String
	}
	if lastSynced.Valid {
		security.LastSynced = lastSynced.String
	}
	if minPortfolioTarget.Valid {
		security.MinPortfolioTarget = minPortfolioTarget.Float64
	}
	if maxPortfolioTarget.Valid {
		security.MaxPortfolioTarget = maxPortfolioTarget.Float64
	}

	// Handle boolean fields (stored as integers in SQLite)
	if active.Valid {
		security.Active = active.Int64 != 0
	}
	if allowBuy.Valid {
		security.AllowBuy = allowBuy.Int64 != 0
	} else {
		security.AllowBuy = true // Default
	}
	if allowSell.Valid {
		security.AllowSell = allowSell.Int64 != 0
	}

	// Normalize symbol
	security.Symbol = strings.ToUpper(strings.TrimSpace(security.Symbol))

	// Default values
	if security.PriorityMultiplier == 0 {
		security.PriorityMultiplier = 1.0
	}
	if security.MinLot == 0 {
		security.MinLot = 1
	}

	// Load tags for the security
	tagIDs, err := r.getTagsForSecurity(security.Symbol)
	if err != nil {
		// Log error but don't fail - tags are optional
		r.log.Warn().Str("symbol", security.Symbol).Err(err).Msg("Failed to load tags for security")
		security.Tags = []string{} // Initialize to empty slice
	} else {
		security.Tags = tagIDs
	}

	return security, nil
}

// getTagsForSecurity loads tag IDs for a security
func (r *SecurityRepository) getTagsForSecurity(symbol string) ([]string, error) {
	query := "SELECT tag_id FROM security_tags WHERE symbol = ? ORDER BY tag_id"

	rows, err := r.universeDB.Query(query, strings.ToUpper(strings.TrimSpace(symbol)))
	if err != nil {
		return nil, fmt.Errorf("failed to query tags for security: %w", err)
	}
	defer rows.Close()

	var tagIDs []string
	for rows.Next() {
		var tagID string
		err := rows.Scan(&tagID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag ID: %w", err)
		}
		tagIDs = append(tagIDs, tagID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	return tagIDs, nil
}

// SetTagsForSecurity replaces all tags for a security (deletes existing, inserts new)
func (r *SecurityRepository) SetTagsForSecurity(symbol string, tagIDs []string) error {
	// Normalize symbol
	symbol = strings.ToUpper(strings.TrimSpace(symbol))

	// Ensure all tag IDs exist (create with default names if missing)
	if err := r.tagRepo.EnsureTagsExist(tagIDs); err != nil {
		return fmt.Errorf("failed to ensure tags exist: %w", err)
	}

	// Begin transaction
	tx, err := r.universeDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Delete all existing tags for this security
	_, err = tx.Exec("DELETE FROM security_tags WHERE symbol = ?", symbol)
	if err != nil {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}

	// Insert new tags
	now := time.Now().Format(time.RFC3339)
	for _, tagID := range tagIDs {
		// Skip empty tag IDs
		tagID = strings.ToLower(strings.TrimSpace(tagID))
		if tagID == "" {
			continue
		}

		_, err = tx.Exec(`
			INSERT INTO security_tags (symbol, tag_id, created_at, updated_at)
			VALUES (?, ?, ?, ?)
		`, symbol, tagID, now, now)
		if err != nil {
			return fmt.Errorf("failed to insert tag %s: %w", tagID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.log.Debug().Str("symbol", symbol).Int("tag_count", len(tagIDs)).Msg("Tags updated for security")
	return nil
}

// Helper functions

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullFloat64(f float64) sql.NullFloat64 {
	if f == 0 {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: f, Valid: true}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// GetGroupedByExchange returns active securities grouped by exchange
func (r *SecurityRepository) GetGroupedByExchange() (map[string][]Security, error) {
	securities, err := r.GetAllActive()
	if err != nil {
		return nil, fmt.Errorf("failed to get active securities: %w", err)
	}

	grouped := make(map[string][]Security)
	for _, security := range securities {
		// Use FullExchangeName as the grouping key
		exchange := security.FullExchangeName
		if exchange == "" {
			exchange = "UNKNOWN"
		}
		grouped[exchange] = append(grouped[exchange], security)
	}

	r.log.Debug().Int("exchanges", len(grouped)).Msg("Grouped securities by exchange")
	return grouped, nil
}
