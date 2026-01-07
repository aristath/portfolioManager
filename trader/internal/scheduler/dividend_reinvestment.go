package scheduler

import (
	"fmt"
	"time"

	"github.com/aristath/sentinel/internal/modules/dividends"
	planningdomain "github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/rs/zerolog"
)

// DividendReinvestmentJob orchestrates individual dividend jobs to automatically reinvest dividends
type DividendReinvestmentJob struct {
	log                              zerolog.Logger
	getUnreinvestedDividendsJob      Job
	groupDividendsBySymbolJob        Job
	checkDividendYieldsJob           Job
	createDividendRecommendationsJob Job
	setPendingBonusesJob             Job
	executeDividendTradesJob         Job
}

// DividendReinvestmentConfig holds configuration for dividend reinvestment job
type DividendReinvestmentConfig struct {
	Log                              zerolog.Logger
	GetUnreinvestedDividendsJob      Job
	GroupDividendsBySymbolJob        Job
	CheckDividendYieldsJob           Job
	CreateDividendRecommendationsJob Job
	SetPendingBonusesJob             Job
	ExecuteDividendTradesJob         Job
}

// NewDividendReinvestmentJob creates a new dividend reinvestment job
func NewDividendReinvestmentJob(cfg DividendReinvestmentConfig) *DividendReinvestmentJob {
	return &DividendReinvestmentJob{
		log:                              cfg.Log.With().Str("job", "dividend_reinvestment").Logger(),
		getUnreinvestedDividendsJob:      cfg.GetUnreinvestedDividendsJob,
		groupDividendsBySymbolJob:        cfg.GroupDividendsBySymbolJob,
		checkDividendYieldsJob:           cfg.CheckDividendYieldsJob,
		createDividendRecommendationsJob: cfg.CreateDividendRecommendationsJob,
		setPendingBonusesJob:             cfg.SetPendingBonusesJob,
		executeDividendTradesJob:         cfg.ExecuteDividendTradesJob,
	}
}

// Name returns the job name
func (j *DividendReinvestmentJob) Name() string {
	return "dividend_reinvestment"
}

// Run executes the dividend reinvestment job by orchestrating individual dividend jobs
// Note: Concurrent execution is prevented by the scheduler's SkipIfStillRunning wrapper
func (j *DividendReinvestmentJob) Run() error {
	j.log.Info().Msg("Starting automatic dividend reinvestment")
	startTime := time.Now()

	// Step 1: Get all unreinvested dividends
	if j.getUnreinvestedDividendsJob == nil {
		return fmt.Errorf("get unreinvested dividends job not available")
	}
	if err := j.getUnreinvestedDividendsJob.Run(); err != nil {
		return fmt.Errorf("failed to get unreinvested dividends: %w", err)
	}

	// Get dividends from job
	dividendsJob, ok := j.getUnreinvestedDividendsJob.(*GetUnreinvestedDividendsJob)
	if !ok {
		return fmt.Errorf("get unreinvested dividends job has wrong type")
	}
	unreinvestedDividends := dividendsJob.GetDividends()

	if len(unreinvestedDividends) == 0 {
		j.log.Info().Msg("No unreinvested dividends found")
		return nil
	}

	// Step 2: Group dividends by symbol
	if j.groupDividendsBySymbolJob == nil {
		return fmt.Errorf("group dividends by symbol job not available")
	}
	groupJob, ok := j.groupDividendsBySymbolJob.(*GroupDividendsBySymbolJob)
	if !ok {
		return fmt.Errorf("group dividends by symbol job has wrong type")
	}
	groupJob.SetDividends(unreinvestedDividends)
	if err := j.groupDividendsBySymbolJob.Run(); err != nil {
		return fmt.Errorf("failed to group dividends: %w", err)
	}
	groupedDividends := groupJob.GetGroupedDividends()

	// Step 3: Check dividend yields
	if j.checkDividendYieldsJob == nil {
		return fmt.Errorf("check dividend yields job not available")
	}
	yieldsJob, ok := j.checkDividendYieldsJob.(*CheckDividendYieldsJob)
	if !ok {
		return fmt.Errorf("check dividend yields job has wrong type")
	}
	yieldsJob.SetGroupedDividends(groupedDividends)
	if err := j.checkDividendYieldsJob.Run(); err != nil {
		return fmt.Errorf("failed to check dividend yields: %w", err)
	}
	highYieldSymbols := yieldsJob.GetHighYieldSymbols()
	lowYieldSymbols := yieldsJob.GetLowYieldSymbols()

	// Step 4: Create recommendations for high-yield dividends
	var recommendations []planningdomain.HolisticStep
	var dividendsToMark map[string][]int
	if len(highYieldSymbols) > 0 {
		if j.createDividendRecommendationsJob == nil {
			return fmt.Errorf("create dividend recommendations job not available")
		}
		recJob, ok := j.createDividendRecommendationsJob.(*CreateDividendRecommendationsJob)
		if !ok {
			return fmt.Errorf("create dividend recommendations job has wrong type")
		}
		recJob.SetHighYieldSymbols(highYieldSymbols)
		if err := j.createDividendRecommendationsJob.Run(); err != nil {
			return fmt.Errorf("failed to create dividend recommendations: %w", err)
		}
		recommendations = recJob.GetRecommendations()
		dividendsToMark = recJob.GetDividendsToMark()
	}

	// Step 5: Set pending bonuses for low-yield dividends
	if len(lowYieldSymbols) > 0 {
		if j.setPendingBonusesJob != nil {
			setJob, ok := j.setPendingBonusesJob.(*SetPendingBonusesJob)
			if ok {
				// Collect all low-yield dividends
				var allLowYieldDividends []dividends.DividendRecord
				for _, info := range lowYieldSymbols {
					allLowYieldDividends = append(allLowYieldDividends, info.Dividends...)
				}
				setJob.SetDividends(allLowYieldDividends)
				if err := j.setPendingBonusesJob.Run(); err != nil {
					j.log.Error().Err(err).Msg("Failed to set pending bonuses (non-critical)")
					// Continue - non-critical
				}
			}
		}
	}

	// Step 6: Execute trades if any
	if len(recommendations) > 0 {
		if j.executeDividendTradesJob == nil {
			return fmt.Errorf("execute dividend trades job not available")
		}
		execJob, ok := j.executeDividendTradesJob.(*ExecuteDividendTradesJob)
		if !ok {
			return fmt.Errorf("execute dividend trades job has wrong type")
		}
		execJob.SetRecommendations(recommendations, dividendsToMark)
		if err := j.executeDividendTradesJob.Run(); err != nil {
			return fmt.Errorf("failed to execute dividend trades: %w", err)
		}
	}

	duration := time.Since(startTime)
	j.log.Info().
		Dur("duration", duration).
		Int("recommendations", len(recommendations)).
		Msg("Dividend reinvestment job completed")

	return nil
}
