/**
 * Package di provides dependency injection for planner work registration.
 *
 * Planner work types handle portfolio optimization, opportunity context building,
 * plan creation, and recommendation storage.
 */
package di

import (
	"fmt"

	"github.com/aristath/sentinel/internal/events"
	planningdomain "github.com/aristath/sentinel/internal/modules/planning/domain"
	planninghash "github.com/aristath/sentinel/internal/modules/planning/hash"
	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/aristath/sentinel/internal/scheduler"
	"github.com/aristath/sentinel/internal/work"
	"github.com/rs/zerolog"
)

// Planner adapters - wrap existing planner jobs
type plannerOptimizerAdapter struct {
	container *Container
	cache     *workCache
	log       zerolog.Logger
}

func (a *plannerOptimizerAdapter) CalculateWeights() (map[string]float64, error) {
	// Create adapters for the optimizer weights job
	positionAdapter := scheduler.NewPositionRepositoryAdapter(a.container.PositionRepo)
	securityAdapter := scheduler.NewSecurityRepositoryAdapter(a.container.SecurityRepo)
	allocAdapter := scheduler.NewAllocationRepositoryAdapter(a.container.AllocRepo)
	priceAdapter := scheduler.NewPriceClientAdapter(&workBrokerPriceAdapter{client: a.container.BrokerClient})
	optimizerAdapter := scheduler.NewOptimizerServiceAdapter(a.container.OptimizerService)
	priceConversionAdapter := scheduler.NewPriceConversionServiceAdapter(a.container.PriceConversionService)
	plannerConfigAdapter := scheduler.NewPlannerConfigRepositoryAdapter(a.container.PlannerConfigRepo)
	clientDataAdapter := scheduler.NewClientDataRepositoryAdapter(a.container.ClientDataRepo)
	marketHoursAdapter := scheduler.NewMarketHoursServiceAdapter(a.container.MarketHoursService)

	// Create and run the optimizer weights job
	job := scheduler.NewGetOptimizerWeightsJob(
		positionAdapter,
		securityAdapter,
		allocAdapter,
		a.container.CashManager,
		priceAdapter,
		optimizerAdapter,
		priceConversionAdapter,
		plannerConfigAdapter,
		clientDataAdapter,
		marketHoursAdapter,
	)
	job.SetLogger(a.log)

	if err := job.Run(); err != nil {
		return nil, err
	}

	return job.GetTargetWeights(), nil
}

type plannerContextBuilderAdapter struct {
	container *Container
	cache     *workCache
}

func (a *plannerContextBuilderAdapter) Build() (interface{}, error) {
	// Get optimizer weights from cache (set by planner:weights work type)
	weightsInterface := a.cache.Get("optimizer_weights")
	var weights map[string]float64
	if weightsInterface != nil {
		if w, ok := weightsInterface.(map[string]float64); ok {
			weights = w
		}
	}

	return a.container.OpportunityContextBuilder.Build(weights)
}

type plannerServiceAdapter struct {
	container *Container
	cache     *workCache
}

func (a *plannerServiceAdapter) CreatePlan(ctx interface{}) (interface{}, error) {
	// Get opportunity context from cache
	opportunityContext := a.cache.Get("opportunity_context")
	if opportunityContext == nil {
		return nil, nil
	}

	// Get planner configuration
	config, err := a.container.PlannerConfigRepo.GetDefaultConfig()
	if err != nil {
		return nil, err
	}

	// Use the existing planner service with type assertion
	ctxTyped, ok := opportunityContext.(*planningdomain.OpportunityContext)
	if !ok {
		return nil, nil
	}

	return a.container.PlannerService.CreatePlan(ctxTyped, config)
}

type plannerRecommendationRepoAdapter struct {
	container *Container
	log       zerolog.Logger
}

func (a *plannerRecommendationRepoAdapter) Store(recommendations interface{}) error {
	// Type assert the plan to HolisticPlan
	plan, ok := recommendations.(*planningdomain.HolisticPlan)
	if !ok {
		return fmt.Errorf("invalid plan type: expected *HolisticPlan")
	}

	// Generate portfolio hash for tracking using the hash package
	portfolioHash := a.generatePortfolioHash()

	// Store the plan using the recommendation repository
	err := a.container.RecommendationRepo.StorePlan(plan, portfolioHash)
	if err != nil {
		return fmt.Errorf("failed to store plan: %w", err)
	}

	a.log.Info().
		Str("portfolio_hash", portfolioHash).
		Int("steps", len(plan.Steps)).
		Msg("Successfully stored recommendations")

	return nil
}

func (a *plannerRecommendationRepoAdapter) generatePortfolioHash() string {
	// Get positions
	positions, err := a.container.PositionRepo.GetAll()
	if err != nil {
		a.log.Warn().Err(err).Msg("Failed to get positions for hash")
		return ""
	}

	// Get securities
	securities, err := a.container.SecurityRepo.GetAllActive()
	if err != nil {
		a.log.Warn().Err(err).Msg("Failed to get securities for hash")
		return ""
	}

	// Get cash balances
	cashBalances := make(map[string]float64)
	if a.container.CashManager != nil {
		balances, err := a.container.CashManager.GetAllCashBalances()
		if err != nil {
			a.log.Warn().Err(err).Msg("Failed to get cash balances for hash")
		} else {
			cashBalances = balances
		}
	}

	// Convert to hash format
	hashPositions := make([]planninghash.Position, 0, len(positions))
	for _, pos := range positions {
		hashPositions = append(hashPositions, planninghash.Position{
			Symbol:   pos.Symbol,
			Quantity: int(pos.Quantity),
		})
	}

	hashSecurities := make([]*universe.Security, 0, len(securities))
	for i := range securities {
		hashSecurities = append(hashSecurities, &securities[i])
	}

	pendingOrders := []planninghash.PendingOrder{}

	return planninghash.GeneratePortfolioHash(
		hashPositions,
		hashSecurities,
		cashBalances,
		pendingOrders,
	)
}

type plannerEventManagerAdapter struct {
	container *Container
}

func (a *plannerEventManagerAdapter) Emit(event string, data interface{}) {
	a.container.EventManager.Emit(events.JobProgress, event, nil)
}

func registerPlannerWork(registry *work.Registry, container *Container, cache *workCache, log zerolog.Logger) {
	deps := &work.PlannerDeps{
		Cache:              cache,
		OptimizerService:   &plannerOptimizerAdapter{container: container, cache: cache, log: log},
		ContextBuilder:     &plannerContextBuilderAdapter{container: container, cache: cache},
		PlannerService:     &plannerServiceAdapter{container: container, cache: cache},
		RecommendationRepo: &plannerRecommendationRepoAdapter{container: container, log: log},
		EventManager:       &plannerEventManagerAdapter{container: container},
	}

	work.RegisterPlannerWorkTypes(registry, deps)
	log.Debug().Msg("Planner work types registered")
}
