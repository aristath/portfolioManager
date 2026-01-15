package display

import (
	"fmt"
	"math"
	"time"

	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/rs/zerolog"
)

// SecurityPerformanceService calculates individual security performance metrics for display
type SecurityPerformanceService struct {
	historyDBClient universe.HistoryDBInterface // Filtered and cached price access
	log             zerolog.Logger
}

// NewSecurityPerformanceService creates a new security performance service
func NewSecurityPerformanceService(historyDBClient universe.HistoryDBInterface, log zerolog.Logger) *SecurityPerformanceService {
	return &SecurityPerformanceService{
		historyDBClient: historyDBClient,
		log:             log.With().Str("service", "security_performance").Logger(),
	}
}

// CalculateTrailing12MoCAGR calculates trailing 12-month CAGR for a specific security using ISIN
// Uses filtered price data from HistoryDB to exclude anomalies
func (s *SecurityPerformanceService) CalculateTrailing12MoCAGR(isin string) (*float64, error) {
	startDateStr := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")

	// Get filtered prices from HistoryDB (already cached and filtered)
	dailyPrices, err := s.historyDBClient.GetDailyPrices(isin, 0) // 0 = no limit
	if err != nil {
		return nil, fmt.Errorf("failed to get price history for %s: %w", isin, err)
	}

	// Filter to last 12 months and collect prices in chronological order
	// dailyPrices comes in descending order (most recent first)
	var prices []struct {
		Date  string
		Close float64
	}

	for i := len(dailyPrices) - 1; i >= 0; i-- {
		p := dailyPrices[i]
		if p.Date >= startDateStr {
			prices = append(prices, struct {
				Date  string
				Close float64
			}{Date: p.Date, Close: p.Close})
		}
	}

	if len(prices) < 2 {
		s.log.Debug().Str("isin", isin).Msg("Insufficient price data for trailing 12mo calculation")
		return nil, nil
	}

	// Use first and last price
	startPrice := prices[0].Close
	endPrice := prices[len(prices)-1].Close

	if startPrice <= 0 {
		s.log.Warn().Str("isin", isin).Msg("Invalid start price for trailing 12mo calculation")
		return nil, nil
	}

	// Calculate days between first and last price
	startDt, _ := time.Parse("2006-01-02", prices[0].Date)
	endDt, _ := time.Parse("2006-01-02", prices[len(prices)-1].Date)
	days := endDt.Sub(startDt).Hours() / 24

	if days < 30 {
		s.log.Debug().Str("isin", isin).Msg("Insufficient time period for trailing 12mo calculation")
		return nil, nil
	}

	years := days / 365.0

	var cagr float64
	if years >= 0.25 {
		// Use CAGR formula for periods >= 3 months
		cagr = math.Pow(endPrice/startPrice, 1/years) - 1
	} else {
		// Simple return for very short periods
		cagr = (endPrice / startPrice) - 1
	}

	s.log.Debug().
		Str("isin", isin).
		Float64("cagr", cagr).
		Float64("start_price", startPrice).
		Float64("end_price", endPrice).
		Float64("days", days).
		Msg("Calculated trailing 12mo CAGR")

	return &cagr, nil
}

// GetPerformanceVsTarget gets security performance difference vs target
func (s *SecurityPerformanceService) GetPerformanceVsTarget(isin string, target float64) (*float64, error) {
	cagr, err := s.CalculateTrailing12MoCAGR(isin)
	if err != nil {
		return nil, err
	}

	if cagr == nil {
		return nil, nil
	}

	difference := *cagr - target

	s.log.Debug().
		Str("isin", isin).
		Float64("difference", difference).
		Float64("cagr", *cagr).
		Float64("target", target).
		Msg("Calculated performance vs target")

	return &difference, nil
}
