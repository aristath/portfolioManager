// Package charts provides services for generating chart data from historical prices.
package charts

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/aristath/sentinel/internal/utils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
)

// ChartDataPoint represents a single point on a chart
type ChartDataPoint struct {
	Time  string  `json:"time"`  // YYYY-MM-DD format
	Value float64 `json:"value"` // Close price
}

// Service provides chart data operations
type Service struct {
	historyDB    *sql.DB // Consolidated history.db connection
	securityRepo *universe.SecurityRepository
	universeDB   *sql.DB // For querying securities (universe.db)
	log          zerolog.Logger
}

// NewService creates a new charts service
func NewService(
	historyDB *sql.DB,
	securityRepo *universe.SecurityRepository,
	universeDB *sql.DB,
	log zerolog.Logger,
) *Service {
	return &Service{
		historyDB:    historyDB,
		securityRepo: securityRepo,
		universeDB:   universeDB,
		log:          log.With().Str("service", "charts").Logger(),
	}
}

// GetSparklinesAggregated returns sparkline data with specified aggregation
// Replaces the old GetSparklines() method - supports 1Y (weekly) or 5Y (monthly) periods
func (s *Service) GetSparklinesAggregated(period string) (map[string][]ChartDataPoint, error) {
	var startDate string
	var groupBy string

	switch period {
	case "1Y":
		startDate = time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
		groupBy = "week" // Weekly aggregation
	case "5Y":
		startDate = time.Now().AddDate(-5, 0, 0).Format("2006-01-02")
		groupBy = "month" // Monthly aggregation
	default:
		return nil, fmt.Errorf("invalid period: %s (must be 1Y or 5Y)", period)
	}

	// Get all active tradable securities with ISINs (excludes indices)
	rows, err := s.universeDB.Query("SELECT symbol, isin FROM securities WHERE active = 1 AND isin != '' AND (product_type IS NULL OR product_type != ?)", string(domain.ProductTypeIndex))
	if err != nil {
		return nil, fmt.Errorf("failed to get active securities: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]ChartDataPoint)

	for rows.Next() {
		var symbol string
		var isin sql.NullString
		if err := rows.Scan(&symbol, &isin); err != nil {
			s.log.Warn().Err(err).Msg("Failed to scan symbol")
			continue
		}

		// Skip securities without ISIN
		if !isin.Valid || isin.String == "" {
			s.log.Debug().Str("symbol", symbol).Msg("Skipping security without ISIN")
			continue
		}

		// Get aggregated prices
		prices, err := s.getAggregatedPrices(isin.String, startDate, groupBy)
		if err != nil {
			s.log.Debug().
				Err(err).
				Str("symbol", symbol).
				Str("isin", isin.String).
				Msg("Failed to get aggregated prices for symbol")
			continue
		}

		if len(prices) > 0 {
			result[symbol] = prices
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating securities: %w", err)
	}

	return result, nil
}

// getAggregatedPrices fetches price data aggregated by week or month
func (s *Service) getAggregatedPrices(isin string, startDate string, groupBy string) ([]ChartDataPoint, error) {
	startUnix, err := utils.DateToUnix(startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}

	var query string
	var args []interface{}

	if groupBy == "week" {
		// Weekly aggregation: use strftime to group by ISO week
		query = `
			SELECT
				strftime('%Y-W%W', date, 'unixepoch') as period,
				AVG(close) as avg_close
			FROM daily_prices
			WHERE isin = ? AND date >= ?
			GROUP BY period
			ORDER BY MIN(date) ASC
		`
		args = []interface{}{isin, startUnix}
	} else {
		// Monthly aggregation: use existing monthly_prices table
		query = `
			SELECT
				year_month as period,
				avg_close
			FROM monthly_prices
			WHERE isin = ? AND year_month >= strftime('%Y-%m', ?, 'unixepoch')
			ORDER BY year_month ASC
		`
		args = []interface{}{isin, startUnix}
	}

	rows, err := s.historyDB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query aggregated prices: %w", err)
	}
	defer rows.Close()

	var prices []ChartDataPoint
	for rows.Next() {
		var period string
		var avgClose sql.NullFloat64

		if err := rows.Scan(&period, &avgClose); err != nil {
			s.log.Warn().Err(err).Msg("Failed to scan aggregated price row")
			continue
		}

		if !avgClose.Valid {
			continue
		}

		prices = append(prices, ChartDataPoint{
			Time:  period, // Format: "2024-W01" or "2024-01"
			Value: avgClose.Float64,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating aggregated prices: %w", err)
	}

	return prices, nil
}

// GetSecurityChart returns historical price data for a specific security
// Faithful translation from Python: app/api/charts.py -> get_security_chart()
func (s *Service) GetSecurityChart(isin string, dateRange string) ([]ChartDataPoint, error) {
	// Validate ISIN is not empty
	if isin == "" {
		return nil, fmt.Errorf("ISIN cannot be empty")
	}

	// Parse date range
	startDate := parseDateRange(dateRange)

	// Get prices from database using ISIN directly
	prices, err := s.getPricesFromDB(isin, startDate, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get prices: %w", err)
	}

	return prices, nil
}

// getPricesFromDB fetches price data from the consolidated history database using ISIN
func (s *Service) getPricesFromDB(isin string, startDate string, endDate string) ([]ChartDataPoint, error) {
	// Build query with ISIN filter
	var query string
	var args []interface{}

	if startDate != "" && endDate != "" {
		// Convert YYYY-MM-DD strings to Unix timestamps
		startUnix, err := utils.DateToUnix(startDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format (expected YYYY-MM-DD): %w", err)
		}
		// End date should be end of day (23:59:59)
		endTime, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format (expected YYYY-MM-DD): %w", err)
		}
		endUnix := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, time.UTC).Unix()

		query = `
			SELECT date, close
			FROM daily_prices
			WHERE isin = ? AND date >= ? AND date <= ?
			ORDER BY date ASC
		`
		args = []interface{}{isin, startUnix, endUnix}
	} else if startDate != "" {
		// Convert YYYY-MM-DD string to Unix timestamp
		startUnix, err := utils.DateToUnix(startDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format (expected YYYY-MM-DD): %w", err)
		}

		query = `
			SELECT date, close
			FROM daily_prices
			WHERE isin = ? AND date >= ?
			ORDER BY date ASC
		`
		args = []interface{}{isin, startUnix}
	} else {
		query = `
			SELECT date, close
			FROM daily_prices
			WHERE isin = ?
			ORDER BY date ASC
		`
		args = []interface{}{isin}
	}

	rows, err := s.historyDB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily prices: %w", err)
	}
	defer rows.Close()

	var prices []ChartDataPoint
	for rows.Next() {
		var dateUnix sql.NullInt64
		var closePrice sql.NullFloat64

		if err := rows.Scan(&dateUnix, &closePrice); err != nil {
			s.log.Warn().Err(err).Msg("Failed to scan price row")
			continue
		}

		// Skip rows with null prices or dates
		if !closePrice.Valid || !dateUnix.Valid {
			continue
		}

		// Convert Unix timestamp to YYYY-MM-DD string for JSON response
		dateStr := utils.UnixToDate(dateUnix.Int64)

		prices = append(prices, ChartDataPoint{
			Time:  dateStr,
			Value: closePrice.Float64,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating prices: %w", err)
	}

	return prices, nil
}

// parseDateRange converts a range string to a start date
// Faithful translation from Python: app/api/charts.py -> _parse_date_range()
func parseDateRange(rangeStr string) string {
	if rangeStr == "all" || rangeStr == "" {
		return ""
	}

	now := time.Now()
	var startDate time.Time

	switch rangeStr {
	case "1M":
		startDate = now.AddDate(0, -1, 0)
	case "3M":
		startDate = now.AddDate(0, -3, 0)
	case "6M":
		startDate = now.AddDate(0, -6, 0)
	case "1Y":
		startDate = now.AddDate(-1, 0, 0)
	case "5Y":
		startDate = now.AddDate(-5, 0, 0)
	case "10Y":
		startDate = now.AddDate(-10, 0, 0)
	default:
		return ""
	}

	return startDate.Format("2006-01-02")
}
