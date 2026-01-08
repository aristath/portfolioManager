package scheduler

import (
	"github.com/rs/zerolog"
)

// CheckNegativeBalancesJob checks for negative cash balances and triggers emergency rebalance
type CheckNegativeBalancesJob struct {
	log                zerolog.Logger
	balanceService     BalanceServiceInterface
	emergencyRebalance func() error
}

// CheckNegativeBalancesConfig holds configuration for check negative balances job
type CheckNegativeBalancesConfig struct {
	Log                zerolog.Logger
	BalanceService     BalanceServiceInterface
	EmergencyRebalance func() error
}

// NewCheckNegativeBalancesJob creates a new check negative balances job
func NewCheckNegativeBalancesJob(cfg CheckNegativeBalancesConfig) *CheckNegativeBalancesJob {
	return &CheckNegativeBalancesJob{
		log:                cfg.Log.With().Str("job", "check_negative_balances").Logger(),
		balanceService:     cfg.BalanceService,
		emergencyRebalance: cfg.EmergencyRebalance,
	}
}

// Name returns the job name
func (j *CheckNegativeBalancesJob) Name() string {
	return "check_negative_balances"
}

// Run executes the check negative balances job
func (j *CheckNegativeBalancesJob) Run() error {
	j.log.Debug().Msg("Checking for negative balances")

	if j.balanceService == nil {
		j.log.Warn().Msg("Balance service not available, skipping negative balance check")
		return nil
	}

	// Get all currencies that have balances
	currencies, err := j.balanceService.GetAllCurrencies()
	if err != nil {
		j.log.Error().Err(err).Msg("Failed to get currencies for negative balance check")
		return nil // Non-critical, don't fail
	}

	// Check each currency for negative balance
	hasNegativeBalance := false
	for _, currency := range currencies {
		total, err := j.balanceService.GetTotalByCurrency(currency)
		if err != nil {
			j.log.Error().
				Err(err).
				Str("currency", currency).
				Msg("Failed to get total balance for currency")
			continue
		}

		if total < 0 {
			hasNegativeBalance = true
			j.log.Error().
				Str("currency", currency).
				Float64("balance", total).
				Msg("CRITICAL: Negative cash balance detected")
		}
	}

	// Trigger emergency rebalance if negative balance detected
	if hasNegativeBalance {
		j.log.Error().Msg("CRITICAL: Negative balance detected, triggering emergency rebalance")

		if j.emergencyRebalance != nil {
			if err := j.emergencyRebalance(); err != nil {
				j.log.Error().Err(err).Msg("Emergency rebalance failed")
			} else {
				j.log.Info().Msg("Emergency rebalance completed successfully")
			}
		} else {
			j.log.Error().Msg("Emergency rebalance callback not configured")
		}
	}

	return nil
}
