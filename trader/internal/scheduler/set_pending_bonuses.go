package scheduler

import (
	"fmt"

	"github.com/aristath/sentinel/internal/modules/dividends"
	"github.com/rs/zerolog"
)

// SetPendingBonusesJob sets pending bonuses for dividends that are too small to reinvest
type SetPendingBonusesJob struct {
	log             zerolog.Logger
	dividendRepo    DividendRepositoryInterface
	dividendRecords []dividends.DividendRecord
}

// NewSetPendingBonusesJob creates a new SetPendingBonusesJob
func NewSetPendingBonusesJob(
	dividendRepo DividendRepositoryInterface,
) *SetPendingBonusesJob {
	return &SetPendingBonusesJob{
		log:          zerolog.Nop(),
		dividendRepo: dividendRepo,
	}
}

// SetLogger sets the logger for the job
func (j *SetPendingBonusesJob) SetLogger(log zerolog.Logger) {
	j.log = log
}

// SetDividends sets the dividends to set as pending bonuses
func (j *SetPendingBonusesJob) SetDividends(dividends []dividends.DividendRecord) {
	j.dividendRecords = dividends
}

// Name returns the job name
func (j *SetPendingBonusesJob) Name() string {
	return "set_pending_bonuses"
}

// Run executes the set pending bonuses job
func (j *SetPendingBonusesJob) Run() error {
	if j.dividendRepo == nil {
		return fmt.Errorf("dividend repository not available")
	}

	if len(j.dividendRecords) == 0 {
		j.log.Info().Msg("No dividends to set as pending bonuses")
		return nil
	}

	setCount := 0
	for _, dividend := range j.dividendRecords {
		if err := j.dividendRepo.SetPendingBonus(dividend.ID, dividend.AmountEUR); err != nil {
			j.log.Error().
				Err(err).
				Int("dividend_id", dividend.ID).
				Str("symbol", dividend.Symbol).
				Float64("amount", dividend.AmountEUR).
				Msg("Failed to set pending bonus")
			continue
		}
		setCount++
	}

	j.log.Info().
		Int("set_count", setCount).
		Int("total", len(j.dividendRecords)).
		Msg("Successfully set pending bonuses")

	return nil
}
