package dividends

import (
	"time"

	"github.com/rs/zerolog"
)

// PositionValueProvider provides market value for yield calculation
// Minimal interface to avoid coupling with full PositionRepository
type PositionValueProvider interface {
	GetMarketValueByISIN(isin string) (float64, error)
}

// DividendYieldCalculator computes dividend yield from actual receipts
// Uses ledger.db dividend transactions instead of external data sources
type DividendYieldCalculator struct {
	dividendRepo DividendRepositoryInterface
	positionRepo PositionValueProvider
	log          zerolog.Logger
}

// YieldResult contains calculated dividend metrics
type YieldResult struct {
	CurrentYield     float64 // Last 12 months dividends / current market value
	FiveYearAvgYield float64 // 5-year average yield
	GrowthRate       float64 // YoY dividend growth rate
}

// NewDividendYieldCalculator creates a new yield calculator
func NewDividendYieldCalculator(
	dividendRepo DividendRepositoryInterface,
	positionRepo PositionValueProvider,
	log zerolog.Logger,
) *DividendYieldCalculator {
	return &DividendYieldCalculator{
		dividendRepo: dividendRepo,
		positionRepo: positionRepo,
		log:          log.With().Str("service", "dividend_yield_calculator").Logger(),
	}
}

// CalculateYield computes dividend metrics from actual receipts
func (c *DividendYieldCalculator) CalculateYield(isin string) (*YieldResult, error) {
	// Get all dividends for this ISIN
	dividends, err := c.dividendRepo.GetByISIN(isin)
	if err != nil {
		c.log.Error().Err(err).Str("isin", isin).Msg("Failed to get dividends")
		return nil, err
	}

	// Get current market value
	marketValue, err := c.positionRepo.GetMarketValueByISIN(isin)
	if err != nil {
		c.log.Error().Err(err).Str("isin", isin).Msg("Failed to get market value")
		return nil, err
	}

	// Initialize result with zero values
	result := &YieldResult{
		CurrentYield:     0.0,
		FiveYearAvgYield: 0.0,
		GrowthRate:       0.0,
	}

	// Can't calculate yield without market value
	if marketValue <= 0 {
		c.log.Debug().Str("isin", isin).Msg("No market value, cannot calculate yield")
		return result, nil
	}

	// Can't calculate yield without dividends
	if len(dividends) == 0 {
		c.log.Debug().Str("isin", isin).Msg("No dividends found")
		return result, nil
	}

	// Calculate time boundaries
	now := time.Now()
	oneYearAgo := now.AddDate(-1, 0, 0)
	fiveYearsAgo := now.AddDate(-5, 0, 0)

	// Group dividends by year (relative to now)
	yearlyDividends := make(map[int]float64) // key: years ago (0=current, 1=last year, etc.)

	for _, div := range dividends {
		if div.PaymentDate == nil {
			continue
		}

		paymentTime := time.Unix(*div.PaymentDate, 0)
		yearsAgo := yearsSince(paymentTime, now)

		// Only include dividends from last 5 years
		if yearsAgo >= 0 && yearsAgo < 5 {
			yearlyDividends[yearsAgo] += div.AmountEUR
		}
	}

	// Calculate current yield (last 12 months)
	last12MonthsDividends := sumDividendsSince(dividends, oneYearAgo)
	result.CurrentYield = last12MonthsDividends / marketValue

	// Calculate 5-year average yield
	last5YearsDividends := sumDividendsSince(dividends, fiveYearsAgo)
	yearsWithData := countYearsWithData(yearlyDividends, 5)
	if yearsWithData > 0 {
		avgAnnualDividends := last5YearsDividends / float64(yearsWithData)
		result.FiveYearAvgYield = avgAnnualDividends / marketValue
	}

	// Calculate YoY growth rate
	currentYearDividends := yearlyDividends[0]
	previousYearDividends := yearlyDividends[1]
	if previousYearDividends > 0 {
		result.GrowthRate = (currentYearDividends - previousYearDividends) / previousYearDividends
	}

	c.log.Debug().
		Str("isin", isin).
		Float64("current_yield", result.CurrentYield).
		Float64("five_year_avg_yield", result.FiveYearAvgYield).
		Float64("growth_rate", result.GrowthRate).
		Msg("Calculated dividend yield")

	return result, nil
}

// yearsSince calculates how many complete 12-month periods have passed between a date and now
// Returns 0 for dates within the last 12 months, 1 for dates 12-24 months ago, etc.
func yearsSince(date time.Time, now time.Time) int {
	if date.After(now) {
		return -1 // Future date
	}

	// Calculate the difference in days and convert to years
	daysDiff := now.Sub(date).Hours() / 24
	years := int(daysDiff / 365.25) // Use 365.25 to account for leap years

	return years
}

// sumDividendsSince sums all dividends paid since a given date
func sumDividendsSince(dividends []DividendRecord, since time.Time) float64 {
	var total float64
	sinceUnix := since.Unix()

	for _, div := range dividends {
		if div.PaymentDate == nil {
			continue
		}
		if *div.PaymentDate >= sinceUnix {
			total += div.AmountEUR
		}
	}

	return total
}

// countYearsWithData counts how many years (0 to maxYears-1) have dividend data
func countYearsWithData(yearlyDividends map[int]float64, maxYears int) int {
	count := 0
	for i := 0; i < maxYears; i++ {
		if yearlyDividends[i] > 0 {
			count++
		}
	}
	return count
}
