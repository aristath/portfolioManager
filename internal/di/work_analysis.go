/**
 * Package di provides dependency injection for analysis work registration.
 *
 * Analysis work types handle market regime analysis and other analytical tasks.
 */
package di

import (
	"github.com/aristath/sentinel/internal/work"
	"github.com/rs/zerolog"
)

// Analysis work type adapters
type marketRegimeAdapter struct {
	container *Container
}

func (a *marketRegimeAdapter) AnalyzeMarketRegime() error {
	if a.container.RegimeDetector == nil || a.container.RegimePersistence == nil {
		return nil
	}
	// Calculate and persist regime scores for all regions
	scores, err := a.container.RegimeDetector.CalculateAllRegionScores(90) // 90-day window
	if err != nil {
		return err
	}
	// Record each region's score
	for region, score := range scores {
		regimeScore := a.container.RegimeDetector.CalculateRegimeScore(score, 0, 0) // Basic score
		if err := a.container.RegimePersistence.RecordRegimeScoreForRegion(region, regimeScore); err != nil {
			return err
		}
	}
	return nil
}

func (a *marketRegimeAdapter) NeedsAnalysis() bool {
	// Check if regime detector is available
	return a.container.RegimeDetector != nil && a.container.RegimePersistence != nil
}

func registerAnalysisWork(registry *work.Registry, container *Container, log zerolog.Logger) {
	deps := &work.AnalysisDeps{
		MarketRegimeService: &marketRegimeAdapter{container: container},
	}

	work.RegisterAnalysisWorkTypes(registry, deps)
	log.Debug().Msg("Analysis work types registered")
}
