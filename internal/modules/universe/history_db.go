package universe

import (
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/aristath/sentinel/internal/utils"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/rs/zerolog"
)

// HistoryDB provides access to historical price data
// Includes in-memory caching and read-time filtering of anomalies
type HistoryDB struct {
	db          *sql.DB
	priceFilter *PriceFilter
	log         zerolog.Logger

	// In-memory cache for filtered prices
	cacheMu    sync.RWMutex
	priceCache map[string][]DailyPrice // keyed by ISIN
}

// Compile-time check that HistoryDB implements HistoryDBInterface
var _ HistoryDBInterface = (*HistoryDB)(nil)

// NewHistoryDB creates a new history database accessor with price filtering and caching
func NewHistoryDB(db *sql.DB, priceFilter *PriceFilter, log zerolog.Logger) *HistoryDB {
	return &HistoryDB{
		db:          db,
		priceFilter: priceFilter,
		log:         log.With().Str("component", "history_db").Logger(),
		priceCache:  make(map[string][]DailyPrice),
	}
}

// DailyPrice represents a daily OHLCV price point
type DailyPrice struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume *int64  `json:"volume,omitempty"`
}

// MonthlyPrice represents a monthly average price
type MonthlyPrice struct {
	YearMonth   string  `json:"year_month"`
	AvgAdjClose float64 `json:"avg_adj_close"`
}

// GetDailyPrices fetches daily price data for an ISIN
// Returns filtered, cached data with anomalies removed
func (h *HistoryDB) GetDailyPrices(isin string, limit int) ([]DailyPrice, error) {
	// Check cache first
	h.cacheMu.RLock()
	if cached, ok := h.priceCache[isin]; ok {
		h.cacheMu.RUnlock()
		return h.applyLimit(cached, limit), nil
	}
	h.cacheMu.RUnlock()

	// Fetch from DB, filter, and cache
	filtered, err := h.fetchAndFilter(isin)
	if err != nil {
		return nil, err
	}

	// Cache the result
	h.cacheMu.Lock()
	h.priceCache[isin] = filtered
	h.cacheMu.Unlock()

	return h.applyLimit(filtered, limit), nil
}

// GetRecentPrices fetches recent daily price data for an ISIN
// Returns prices from the last N days, ordered by date descending
// Uses the same cache as GetDailyPrices but filters by date
func (h *HistoryDB) GetRecentPrices(isin string, days int) ([]DailyPrice, error) {
	if days <= 0 {
		return []DailyPrice{}, nil
	}

	// Get all cached/filtered prices
	allPrices, err := h.GetDailyPrices(isin, 0) // 0 = no limit
	if err != nil {
		return nil, err
	}

	// Filter by date cutoff
	cutoffDate := time.Now().AddDate(0, 0, -days)
	cutoffStr := cutoffDate.Format("2006-01-02")

	var recent []DailyPrice
	for _, p := range allPrices {
		if p.Date >= cutoffStr {
			recent = append(recent, p)
		}
	}

	return recent, nil
}

// fetchAndFilter fetches all prices from DB, converts to chronological order,
// filters anomalies, then returns in descending order (most recent first)
func (h *HistoryDB) fetchAndFilter(isin string) ([]DailyPrice, error) {
	query := `
		SELECT date, close, high, low, open, volume
		FROM daily_prices
		WHERE isin = ?
		ORDER BY date ASC
	`

	rows, err := h.db.Query(query, isin)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily prices: %w", err)
	}
	defer rows.Close()

	var rawPrices []DailyPrice
	for rows.Next() {
		var p DailyPrice
		var volume sql.NullInt64
		var dateUnix sql.NullInt64

		err := rows.Scan(&dateUnix, &p.Close, &p.High, &p.Low, &p.Open, &volume)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily price: %w", err)
		}

		if dateUnix.Valid {
			t := time.Unix(dateUnix.Int64, 0).UTC()
			p.Date = t.Format("2006-01-02")
		}
		if volume.Valid {
			p.Volume = &volume.Int64
		}

		rawPrices = append(rawPrices, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating daily prices: %w", err)
	}

	// Filter anomalies (prices are in chronological order for filtering)
	var filtered []DailyPrice
	if h.priceFilter != nil {
		filtered = h.priceFilter.Filter(rawPrices)
	} else {
		filtered = rawPrices
	}

	// Sort descending (most recent first) for return
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Date > filtered[j].Date
	})

	return filtered, nil
}

// applyLimit returns the first n elements if limit > 0, otherwise all
func (h *HistoryDB) applyLimit(prices []DailyPrice, limit int) []DailyPrice {
	if limit <= 0 || len(prices) <= limit {
		return prices
	}
	return prices[:limit]
}

// InvalidateCache removes the cached data for a specific ISIN
func (h *HistoryDB) InvalidateCache(isin string) {
	h.cacheMu.Lock()
	delete(h.priceCache, isin)
	h.cacheMu.Unlock()
}

// InvalidateAllCaches removes all cached price data
func (h *HistoryDB) InvalidateAllCaches() {
	h.cacheMu.Lock()
	h.priceCache = make(map[string][]DailyPrice)
	h.cacheMu.Unlock()
}

// GetMonthlyPrices fetches monthly price data for an ISIN
func (h *HistoryDB) GetMonthlyPrices(isin string, limit int) ([]MonthlyPrice, error) {
	query := `
		SELECT year_month, avg_adj_close
		FROM monthly_prices
		WHERE isin = ?
		ORDER BY year_month DESC
		LIMIT ?
	`

	rows, err := h.db.Query(query, isin, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query monthly prices: %w", err)
	}
	defer rows.Close()

	var prices []MonthlyPrice
	for rows.Next() {
		var p MonthlyPrice

		err := rows.Scan(&p.YearMonth, &p.AvgAdjClose)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monthly price: %w", err)
		}

		prices = append(prices, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating monthly prices: %w", err)
	}

	return prices, nil
}

// HasMonthlyData checks if the history database has monthly price data for an ISIN
// Used to determine if initial 10-year seed has been done
func (h *HistoryDB) HasMonthlyData(isin string) (bool, error) {
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM monthly_prices WHERE isin = ?", isin).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check monthly data: %w", err)
	}

	return count > 0, nil
}

// SyncHistoricalPrices writes historical price data to the database
// Stores raw data from Tradernet without validation - filtering happens on read
//
// Inserts/replaces daily prices and aggregates to monthly prices in a single transaction
// The isin parameter is the ISIN (e.g., US0378331005), not the Tradernet symbol
func (h *HistoryDB) SyncHistoricalPrices(isin string, prices []DailyPrice) error {
	// Begin transaction
	tx, err := h.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Will be no-op if Commit succeeds

	// Insert/replace daily prices with ISIN
	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO daily_prices
		(isin, date, open, high, low, close, volume, adjusted_close)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, price := range prices {
		volume := sql.NullInt64{}
		if price.Volume != nil {
			volume.Int64 = *price.Volume
			volume.Valid = true
		}

		adjustedClose := price.Close // Use close as adjusted_close if not provided

		// Convert date string to Unix timestamp
		dateUnix, err := utils.DateToUnix(price.Date)
		if err != nil {
			return fmt.Errorf("failed to parse date %s: %w", price.Date, err)
		}

		_, err = stmt.Exec(
			isin,
			dateUnix,
			price.Open,
			price.High,
			price.Low,
			price.Close,
			volume,
			adjustedClose,
		)
		if err != nil {
			return fmt.Errorf("failed to insert daily price for %s: %w", price.Date, err)
		}
	}

	// Aggregate to monthly prices with ISIN filter
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO monthly_prices
		(isin, year_month, avg_close, avg_adj_close, source, created_at)
		SELECT
			? as isin,
			strftime('%Y-%m', datetime(date, 'unixepoch')) as year_month,
			AVG(close) as avg_close,
			AVG(adjusted_close) as avg_adj_close,
			'calculated',
			strftime('%s', 'now')
		FROM daily_prices
		WHERE isin = ?
		GROUP BY strftime('%Y-%m', datetime(date, 'unixepoch'))
	`, isin, isin)
	if err != nil {
		return fmt.Errorf("failed to aggregate monthly prices: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache for this ISIN after successful write
	h.InvalidateCache(isin)

	h.log.Info().
		Str("isin", isin).
		Int("count", len(prices)).
		Msg("Synced historical prices")

	return nil
}

// ExchangeRate represents a cached exchange rate
type ExchangeRate struct {
	FromCurrency string
	ToCurrency   string
	Date         time.Time
	Rate         float64
}

// UpsertExchangeRate inserts or replaces an exchange rate
// Uses current date at midnight UTC for the date field
func (h *HistoryDB) UpsertExchangeRate(fromCurrency, toCurrency string, rate float64) error {
	now := time.Now()
	dateUnix := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Unix()

	query := `
		INSERT OR REPLACE INTO exchange_rates (from_currency, to_currency, date, rate)
		VALUES (?, ?, ?, ?)
	`

	_, err := h.db.Exec(query, fromCurrency, toCurrency, dateUnix, rate)
	if err != nil {
		return fmt.Errorf("failed to upsert exchange rate: %w", err)
	}

	h.log.Debug().
		Str("from", fromCurrency).
		Str("to", toCurrency).
		Float64("rate", rate).
		Msg("Upserted exchange rate")

	return nil
}

// GetLatestExchangeRate fetches most recent rate for a currency pair
// Returns nil if no rate found (not an error)
func (h *HistoryDB) GetLatestExchangeRate(fromCurrency, toCurrency string) (*ExchangeRate, error) {
	query := `
		SELECT from_currency, to_currency, date, rate
		FROM exchange_rates
		WHERE from_currency = ? AND to_currency = ?
		ORDER BY date DESC
		LIMIT 1
	`

	var er ExchangeRate
	var dateUnix int64

	err := h.db.QueryRow(query, fromCurrency, toCurrency).Scan(
		&er.FromCurrency,
		&er.ToCurrency,
		&dateUnix,
		&er.Rate,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found (not an error)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	er.Date = time.Unix(dateUnix, 0).UTC()
	return &er, nil
}

// DeletePricesForSecurity removes all price history (daily and monthly) for a security
// Used when hard-deleting a security from the universe
func (h *HistoryDB) DeletePricesForSecurity(isin string) error {
	tx, err := h.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Delete daily prices
	dailyResult, err := tx.Exec("DELETE FROM daily_prices WHERE isin = ?", isin)
	if err != nil {
		return fmt.Errorf("failed to delete daily prices: %w", err)
	}

	// Delete monthly prices
	monthlyResult, err := tx.Exec("DELETE FROM monthly_prices WHERE isin = ?", isin)
	if err != nil {
		return fmt.Errorf("failed to delete monthly prices: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache for this ISIN
	h.InvalidateCache(isin)

	dailyRows, _ := dailyResult.RowsAffected()
	monthlyRows, _ := monthlyResult.RowsAffected()

	h.log.Info().
		Str("isin", isin).
		Int64("daily_deleted", dailyRows).
		Int64("monthly_deleted", monthlyRows).
		Msg("Deleted price history for security")

	return nil
}

// DeleteStaleRates removes exchange rates older than threshold
// Used by cleanup jobs to prevent unbounded table growth
func (h *HistoryDB) DeleteStaleRates(olderThan time.Time) error {
	dateUnix := olderThan.Unix()

	result, err := h.db.Exec("DELETE FROM exchange_rates WHERE date < ?", dateUnix)
	if err != nil {
		return fmt.Errorf("failed to delete stale rates: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		h.log.Info().
			Int64("rows_deleted", rowsAffected).
			Time("older_than", olderThan).
			Msg("Deleted stale exchange rates")
	}

	return nil
}
