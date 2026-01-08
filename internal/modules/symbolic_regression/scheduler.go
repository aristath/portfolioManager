package symbolic_regression

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

// Scheduler handles periodic re-discovery of formulas
type Scheduler struct {
	discovery     *DiscoveryService
	dataPrep      *DataPrep
	storage       *FormulaStorage
	log           zerolog.Logger
	lastDiscovery map[string]time.Time // Track last discovery per formula type
}

// NewScheduler creates a new formula discovery scheduler
func NewScheduler(
	discovery *DiscoveryService,
	dataPrep *DataPrep,
	storage *FormulaStorage,
	log zerolog.Logger,
) *Scheduler {
	return &Scheduler{
		discovery:     discovery,
		dataPrep:      dataPrep,
		storage:       storage,
		log:           log.With().Str("component", "formula_scheduler").Logger(),
		lastDiscovery: make(map[string]time.Time),
	}
}

// ShouldRunDiscovery checks if discovery should run based on schedule
func (s *Scheduler) ShouldRunDiscovery(formulaType FormulaType, securityType SecurityType, intervalMonths int) bool {
	key := string(formulaType) + "_" + string(securityType)
	lastRun, exists := s.lastDiscovery[key]

	if !exists {
		return true // Never run, should run now
	}

	// Check if interval has passed
	nextRun := lastRun.AddDate(0, intervalMonths, 0)
	return time.Now().After(nextRun)
}

// RunScheduledDiscovery runs discovery if it's time
func (s *Scheduler) RunScheduledDiscovery(
	formulaType FormulaType,
	securityType SecurityType,
	intervalMonths int,
	forwardMonths int,
) error {
	if !s.ShouldRunDiscovery(formulaType, securityType, intervalMonths) {
		s.log.Debug().
			Str("formula_type", string(formulaType)).
			Str("security_type", string(securityType)).
			Msg("Discovery not due yet, skipping")
		return nil
	}

	s.log.Info().
		Str("formula_type", string(formulaType)).
		Str("security_type", string(securityType)).
		Msg("Running scheduled formula discovery")

	// Calculate date range (use last 3 years of data)
	endDate := time.Now()
	startDate := endDate.AddDate(-3, 0, 0)

	// Use default regime ranges for regime-specific discovery
	regimeRanges := DefaultRegimeRanges()

	var discoveredFormulas []*DiscoveredFormula
	var err error

	if formulaType == FormulaTypeExpectedReturn {
		discoveredFormulas, err = s.discovery.DiscoverExpectedReturnFormula(
			securityType,
			startDate,
			endDate,
			forwardMonths,
			regimeRanges, // Discover separate formulas for each regime
		)
	} else if formulaType == FormulaTypeScoring {
		discoveredFormulas, err = s.discovery.DiscoverScoringFormula(
			securityType,
			startDate,
			endDate,
			forwardMonths,
			regimeRanges, // Discover separate formulas for each regime
		)
	} else {
		return fmt.Errorf("invalid formula type: %s", formulaType)
	}

	if err != nil {
		s.log.Error().Err(err).Msg("Scheduled discovery failed")
		return err
	}

	if len(discoveredFormulas) > 0 {
		// Update last discovery time
		key := string(formulaType) + "_" + string(securityType)
		s.lastDiscovery[key] = time.Now()

		s.log.Info().
			Str("formula_type", string(formulaType)).
			Str("security_type", string(securityType)).
			Int("formulas_discovered", len(discoveredFormulas)).
			Msg("Scheduled discovery completed successfully")

		// Log each discovered formula
		for _, formula := range discoveredFormulas {
			regimeInfo := "all regimes"
			if formula.RegimeRangeMin != nil && formula.RegimeRangeMax != nil {
				regimeInfo = fmt.Sprintf("regime [%.2f, %.2f]", *formula.RegimeRangeMin, *formula.RegimeRangeMax)
			}
			s.log.Info().
				Str("formula", formula.FormulaExpression).
				Str("regime", regimeInfo).
				Msg("Discovered formula")
		}
	}

	return nil
}

// RunAllScheduledDiscoveries runs all scheduled discoveries
func (s *Scheduler) RunAllScheduledDiscoveries(intervalMonths int, forwardMonths int) error {
	// Discover expected return formulas
	err1 := s.RunScheduledDiscovery(FormulaTypeExpectedReturn, SecurityTypeStock, intervalMonths, forwardMonths)
	err2 := s.RunScheduledDiscovery(FormulaTypeExpectedReturn, SecurityTypeETF, intervalMonths, forwardMonths)

	// Discover scoring formulas
	err3 := s.RunScheduledDiscovery(FormulaTypeScoring, SecurityTypeStock, intervalMonths, forwardMonths)
	err4 := s.RunScheduledDiscovery(FormulaTypeScoring, SecurityTypeETF, intervalMonths, forwardMonths)

	// Return first error if any
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	if err3 != nil {
		return err3
	}
	if err4 != nil {
		return err4
	}

	return nil
}
