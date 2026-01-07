package scheduler

import (
	"fmt"
	"time"

	"github.com/aristath/sentinel/internal/events"
	planningdomain "github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/rs/zerolog"
)

// PlannerBatchJob orchestrates individual planning jobs to generate trading recommendations
type PlannerBatchJob struct {
	log               zerolog.Logger
	eventManager      EventManagerInterface
	lastPortfolioHash string
	// Individual planning jobs
	generatePortfolioHashJob   Job
	buildOpportunityContextJob Job
	createTradePlanJob         Job
	storeRecommendationsJob    Job
}

// PlannerBatchConfig holds configuration for planner batch job
type PlannerBatchConfig struct {
	Log                        zerolog.Logger
	EventManager               EventManagerInterface
	GeneratePortfolioHashJob   Job
	BuildOpportunityContextJob Job
	CreateTradePlanJob         Job
	StoreRecommendationsJob    Job
}

// NewPlannerBatchJob creates a new planner batch job
func NewPlannerBatchJob(cfg PlannerBatchConfig) *PlannerBatchJob {
	return &PlannerBatchJob{
		log:                        cfg.Log.With().Str("job", "planner_batch").Logger(),
		eventManager:               cfg.EventManager,
		generatePortfolioHashJob:   cfg.GeneratePortfolioHashJob,
		buildOpportunityContextJob: cfg.BuildOpportunityContextJob,
		createTradePlanJob:         cfg.CreateTradePlanJob,
		storeRecommendationsJob:    cfg.StoreRecommendationsJob,
	}
}

// Name returns the job name
func (j *PlannerBatchJob) Name() string {
	return "planner_batch"
}

// Run executes the planner batch job by orchestrating individual planning jobs
func (j *PlannerBatchJob) Run() error {
	j.log.Info().Msg("Starting planner batch generation")
	startTime := time.Now()

	// Step 1: Generate portfolio hash
	if j.generatePortfolioHashJob == nil {
		return fmt.Errorf("generate portfolio hash job not available")
	}
	if err := j.generatePortfolioHashJob.Run(); err != nil {
		return fmt.Errorf("failed to generate portfolio hash: %w", err)
	}

	// Get portfolio hash from job
	hashJob, ok := j.generatePortfolioHashJob.(*GeneratePortfolioHashJob)
	if !ok {
		return fmt.Errorf("generate portfolio hash job has wrong type")
	}
	portfolioHash := hashJob.GetLastPortfolioHash()

	// Check if portfolio has changed
	if portfolioHash == j.lastPortfolioHash && j.lastPortfolioHash != "" {
		j.log.Info().
			Str("portfolio_hash", portfolioHash).
			Msg("Portfolio unchanged, skipping planning")
		return nil
	}

	j.log.Info().
		Str("portfolio_hash", portfolioHash).
		Str("prev_hash", j.lastPortfolioHash).
		Msg("Portfolio changed, generating new plan")

	// Step 2: Build opportunity context
	if j.buildOpportunityContextJob == nil {
		return fmt.Errorf("build opportunity context job not available")
	}
	if err := j.buildOpportunityContextJob.Run(); err != nil {
		return fmt.Errorf("failed to build opportunity context: %w", err)
	}

	// Get opportunity context from job
	contextJob, ok := j.buildOpportunityContextJob.(*BuildOpportunityContextJob)
	if !ok {
		return fmt.Errorf("build opportunity context job has wrong type")
	}
	opportunityContext := contextJob.GetOpportunityContext()
	if opportunityContext == nil {
		return fmt.Errorf("opportunity context is nil")
	}

	// Step 3: Create trade plan
	if j.createTradePlanJob == nil {
		return fmt.Errorf("create trade plan job not available")
	}
	planJob, ok := j.createTradePlanJob.(*CreateTradePlanJob)
	if !ok {
		return fmt.Errorf("create trade plan job has wrong type")
	}
	planJob.SetOpportunityContext(opportunityContext)
	if err := j.createTradePlanJob.Run(); err != nil {
		return fmt.Errorf("failed to create trade plan: %w", err)
	}

	// Get plan from job
	plan := planJob.GetPlan()
	if plan == nil {
		return fmt.Errorf("plan is nil")
	}

	// Step 4: Store recommendations
	if j.storeRecommendationsJob == nil {
		return fmt.Errorf("store recommendations job not available")
	}
	storeJob, ok := j.storeRecommendationsJob.(*StoreRecommendationsJob)
	if !ok {
		return fmt.Errorf("store recommendations job has wrong type")
	}
	storeJob.SetPlan(plan)
	storeJob.SetPortfolioHash(portfolioHash)
	if err := j.storeRecommendationsJob.Run(); err != nil {
		return fmt.Errorf("failed to store recommendations: %w", err)
	}

	// Update state
	j.lastPortfolioHash = portfolioHash

	// Emit events
	if j.eventManager != nil {
		planInterface, ok := plan.(*planningdomain.HolisticPlan)
		if ok {
			j.eventManager.Emit(events.PlanGenerated, "planner", map[string]interface{}{
				"portfolio_hash": portfolioHash,
				"steps":          len(planInterface.Steps),
				"end_score":      planInterface.EndStateScore,
				"improvement":    planInterface.Improvement,
				"feasible":       planInterface.Feasible,
			})

			if len(planInterface.Steps) > 0 {
				j.eventManager.Emit(events.RecommendationsReady, "planner", map[string]interface{}{
					"portfolio_hash": portfolioHash,
					"count":          len(planInterface.Steps),
				})
			}
		}
	}

	duration := time.Since(startTime)
	j.log.Info().
		Dur("duration", duration).
		Str("portfolio_hash", portfolioHash).
		Msg("Planner batch completed")

	return nil
}
