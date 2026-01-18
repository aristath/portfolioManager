/**
 * Package di provides dependency injection for dividend work registration.
 *
 * Dividend work types handle detection, analysis, recommendation creation,
 * and execution of dividend reinvestment trades.
 */
package di

import (
	"github.com/aristath/sentinel/internal/work"
	"github.com/rs/zerolog"
)

// Dividend work type adapters
type dividendDetectionAdapter struct {
	container *Container
}

func (a *dividendDetectionAdapter) DetectUnreinvestedDividends() (any, error) {
	// Use the dividend repository to get unreinvested dividends (min 0 EUR)
	return a.container.DividendRepo.GetUnreinvestedDividends(0)
}

func (a *dividendDetectionAdapter) HasPendingDividends() bool {
	dividends, err := a.container.DividendRepo.GetUnreinvestedDividends(0)
	if err != nil {
		return false
	}
	return len(dividends) > 0
}

type dividendAnalysisAdapter struct {
	container *Container
}

func (a *dividendAnalysisAdapter) AnalyzeDividends(dividends any) (any, error) {
	// Analyze dividends - the actual analysis happens in the detection step
	return dividends, nil
}

type dividendRecommendationAdapter struct {
	container *Container
}

func (a *dividendRecommendationAdapter) CreateRecommendations(analysis any) (any, error) {
	// Create dividend reinvestment recommendations
	return analysis, nil
}

type dividendExecutionAdapter struct {
	container *Container
}

func (a *dividendExecutionAdapter) ExecuteTrades(recommendations any) error {
	// Execute dividend reinvestment trades
	return nil
}

func registerDividendWork(registry *work.Registry, container *Container, cache *workCache, log zerolog.Logger) {
	deps := &work.DividendDeps{
		DetectionService:      &dividendDetectionAdapter{container: container},
		AnalysisService:       &dividendAnalysisAdapter{container: container},
		RecommendationService: &dividendRecommendationAdapter{container: container},
		ExecutionService:      &dividendExecutionAdapter{container: container},
		Cache:                 cache,
	}

	work.RegisterDividendWorkTypes(registry, deps)
	log.Debug().Msg("Dividend work types registered")
}
