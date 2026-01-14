package scheduler

import (
	"github.com/aristath/sentinel/internal/modules/dividends"
	"github.com/aristath/sentinel/internal/modules/scoring"
	"github.com/rs/zerolog"
)

// DividendYieldResult holds the dividend yield result for a symbol
type DividendYieldResult struct {
	Symbol      string
	Yield       float64
	IsHighYield bool
	IsAvailable bool
}

// CheckDividendYieldsJob checks dividend yields for symbols
type CheckDividendYieldsJob struct {
	JobBase
	log              zerolog.Logger
	securityRepo     SecurityRepositoryForDividendsInterface
	yieldCalculator  *dividends.DividendYieldCalculator
	symbols          []string
	yieldResults     map[string]DividendYieldResult
	highYieldSymbols map[string]SymbolDividendInfoForGroup
	lowYieldSymbols  map[string]SymbolDividendInfoForGroup
	groupedDividends map[string]SymbolDividendInfoForGroup
}

// NewCheckDividendYieldsJob creates a new CheckDividendYieldsJob
func NewCheckDividendYieldsJob(
	securityRepo SecurityRepositoryForDividendsInterface,
	yieldCalculator *dividends.DividendYieldCalculator,
) *CheckDividendYieldsJob {
	return &CheckDividendYieldsJob{
		log:              zerolog.Nop(),
		securityRepo:     securityRepo,
		yieldCalculator:  yieldCalculator,
		yieldResults:     make(map[string]DividendYieldResult),
		highYieldSymbols: make(map[string]SymbolDividendInfoForGroup),
		lowYieldSymbols:  make(map[string]SymbolDividendInfoForGroup),
	}
}

// SetLogger sets the logger for the job
func (j *CheckDividendYieldsJob) SetLogger(log zerolog.Logger) {
	j.log = log
}

// SetGroupedDividends sets the grouped dividends to check
func (j *CheckDividendYieldsJob) SetGroupedDividends(grouped map[string]SymbolDividendInfoForGroup) {
	j.groupedDividends = grouped
	j.symbols = make([]string, 0, len(grouped))
	for symbol := range grouped {
		j.symbols = append(j.symbols, symbol)
	}
}

// GetYieldResults returns the yield results
func (j *CheckDividendYieldsJob) GetYieldResults() map[string]DividendYieldResult {
	return j.yieldResults
}

// GetHighYieldSymbols returns symbols with high yield
func (j *CheckDividendYieldsJob) GetHighYieldSymbols() map[string]SymbolDividendInfoForGroup {
	return j.highYieldSymbols
}

// GetLowYieldSymbols returns symbols with low yield
func (j *CheckDividendYieldsJob) GetLowYieldSymbols() map[string]SymbolDividendInfoForGroup {
	return j.lowYieldSymbols
}

// Name returns the job name
func (j *CheckDividendYieldsJob) Name() string {
	return "check_dividend_yields"
}

// Run executes the check dividend yields job
func (j *CheckDividendYieldsJob) Run() error {
	if len(j.groupedDividends) == 0 {
		j.log.Info().Msg("No grouped dividends to check")
		return nil
	}

	for symbol, info := range j.groupedDividends {
		yield := j.getDividendYield(symbol)

		result := DividendYieldResult{
			Symbol:      symbol,
			Yield:       yield,
			IsAvailable: yield >= 0,
			IsHighYield: yield >= scoring.HighDividendReinvestmentThreshold,
		}

		j.yieldResults[symbol] = result

		if yield < 0 {
			// No yield data available, treat as low-yield (safer)
			j.log.Debug().
				Str("symbol", symbol).
				Msg("No dividend yield data, treating as low-yield")
			j.lowYieldSymbols[symbol] = info
		} else if yield >= scoring.HighDividendReinvestmentThreshold {
			// High-yield security (>=3%): reinvest in same security
			j.log.Info().
				Str("symbol", symbol).
				Float64("yield", yield*100).
				Float64("total", info.TotalAmount).
				Msg("High yield, will reinvest in same security")
			j.highYieldSymbols[symbol] = info
		} else {
			// Low-yield security (<3%): aggregate for best opportunities
			j.log.Info().
				Str("symbol", symbol).
				Float64("yield", yield*100).
				Float64("total", info.TotalAmount).
				Msg("Low yield, will aggregate for best opportunities")
			j.lowYieldSymbols[symbol] = info
		}
	}

	j.log.Info().
		Int("total_symbols", len(j.symbols)).
		Int("high_yield", len(j.highYieldSymbols)).
		Int("low_yield", len(j.lowYieldSymbols)).
		Msg("Successfully checked dividend yields")

	return nil
}

// getDividendYield gets the dividend yield for a symbol
// Returns -1.0 if not available
func (j *CheckDividendYieldsJob) getDividendYield(symbol string) float64 {
	if j.securityRepo == nil || j.yieldCalculator == nil {
		return -1.0
	}

	// Get the security to find the ISIN
	security, err := j.securityRepo.GetBySymbol(symbol)
	if err != nil || security == nil {
		j.log.Debug().
			Str("symbol", symbol).
			Msg("Security not found, cannot get dividend yield")
		return -1.0
	}

	if security.ISIN == "" {
		j.log.Debug().
			Str("symbol", symbol).
			Msg("Security has no ISIN, cannot get dividend yield")
		return -1.0
	}

	// Get dividend yield from internal calculator
	yieldResult, err := j.yieldCalculator.CalculateYield(security.ISIN)
	if err != nil || yieldResult == nil {
		j.log.Debug().
			Str("symbol", symbol).
			Str("isin", security.ISIN).
			Msg("Failed to calculate dividend yield")
		return -1.0
	}

	// CurrentYield is already a fraction (e.g., 0.03 for 3%)
	if yieldResult.CurrentYield > 0 {
		return yieldResult.CurrentYield
	}

	return -1.0
}
