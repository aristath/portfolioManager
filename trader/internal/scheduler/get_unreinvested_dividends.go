package scheduler

import (
	"fmt"

	"github.com/aristath/sentinel/internal/modules/dividends"
	"github.com/rs/zerolog"
)

// GetUnreinvestedDividendsJob retrieves all unreinvested dividends
type GetUnreinvestedDividendsJob struct {
	log          zerolog.Logger
	dividendRepo DividendRepositoryInterface
	dividends    []dividends.DividendRecord
	minAmountEUR float64
}

// NewGetUnreinvestedDividendsJob creates a new GetUnreinvestedDividendsJob
func NewGetUnreinvestedDividendsJob(
	dividendRepo DividendRepositoryInterface,
	minAmountEUR float64,
) *GetUnreinvestedDividendsJob {
	return &GetUnreinvestedDividendsJob{
		log:          zerolog.Nop(),
		dividendRepo: dividendRepo,
		minAmountEUR: minAmountEUR,
	}
}

// SetLogger sets the logger for the job
func (j *GetUnreinvestedDividendsJob) SetLogger(log zerolog.Logger) {
	j.log = log
}

// GetDividends returns the retrieved dividends
func (j *GetUnreinvestedDividendsJob) GetDividends() []dividends.DividendRecord {
	return j.dividends
}

// Name returns the job name
func (j *GetUnreinvestedDividendsJob) Name() string {
	return "get_unreinvested_dividends"
}

// Run executes the get unreinvested dividends job
func (j *GetUnreinvestedDividendsJob) Run() error {
	if j.dividendRepo == nil {
		return fmt.Errorf("dividend repository not available")
	}

	dividendsInterface, err := j.dividendRepo.GetUnreinvestedDividends(j.minAmountEUR)
	if err != nil {
		j.log.Error().Err(err).Msg("Failed to get unreinvested dividends")
		return fmt.Errorf("failed to get unreinvested dividends: %w", err)
	}

	// Convert interface{} to []dividends.DividendRecord
	dividendRecords := make([]dividends.DividendRecord, 0, len(dividendsInterface))
	for _, d := range dividendsInterface {
		if div, ok := d.(dividends.DividendRecord); ok {
			dividendRecords = append(dividendRecords, div)
		} else if divPtr, ok := d.(*dividends.DividendRecord); ok {
			dividendRecords = append(dividendRecords, *divPtr)
		}
	}

	j.dividends = dividendRecords

	j.log.Info().
		Int("count", len(j.dividends)).
		Float64("min_amount_eur", j.minAmountEUR).
		Msg("Successfully retrieved unreinvested dividends")

	return nil
}
