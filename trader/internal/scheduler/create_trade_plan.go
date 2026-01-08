package scheduler

import (
	"fmt"

	planningdomain "github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/rs/zerolog"
)

// CreateTradePlanJob creates a holistic trade plan from opportunity context
type CreateTradePlanJob struct {
	log                zerolog.Logger
	plannerService     PlannerServiceInterface
	configRepo         ConfigRepositoryInterface
	opportunityContext *planningdomain.OpportunityContext
	plan               *planningdomain.HolisticPlan
}

// NewCreateTradePlanJob creates a new CreateTradePlanJob
func NewCreateTradePlanJob(
	plannerService PlannerServiceInterface,
	configRepo ConfigRepositoryInterface,
) *CreateTradePlanJob {
	return &CreateTradePlanJob{
		log:            zerolog.Nop(),
		plannerService: plannerService,
		configRepo:     configRepo,
	}
}

// SetLogger sets the logger for the job
func (j *CreateTradePlanJob) SetLogger(log zerolog.Logger) {
	j.log = log
}

// SetOpportunityContext sets the opportunity context for plan creation
func (j *CreateTradePlanJob) SetOpportunityContext(ctx *planningdomain.OpportunityContext) {
	j.opportunityContext = ctx
}

// GetPlan returns the created plan
func (j *CreateTradePlanJob) GetPlan() *planningdomain.HolisticPlan {
	return j.plan
}

// Name returns the job name
func (j *CreateTradePlanJob) Name() string {
	return "create_trade_plan"
}

// Run executes the create trade plan job
func (j *CreateTradePlanJob) Run() error {
	if j.plannerService == nil {
		return fmt.Errorf("planner service not available")
	}

	if j.opportunityContext == nil {
		return fmt.Errorf("opportunity context not set")
	}

	// Load planner configuration
	config, err := j.loadPlannerConfig()
	if err != nil {
		j.log.Warn().Err(err).Msg("Failed to load planner config, using defaults")
		config = planningdomain.NewDefaultConfiguration()
	}

	// Create plan (planner service returns interface{}, we type assert to HolisticPlan)
	planInterface, err := j.plannerService.CreatePlan(j.opportunityContext, config)
	if err != nil {
		j.log.Error().Err(err).Msg("Failed to create plan")
		return fmt.Errorf("failed to create plan: %w", err)
	}

	// Type assert to HolisticPlan
	plan, ok := planInterface.(*planningdomain.HolisticPlan)
	if !ok {
		return fmt.Errorf("plan has invalid type: expected *planningdomain.HolisticPlan")
	}

	j.plan = plan

	j.log.Info().
		Msg("Successfully created trade plan")

	return nil
}

// loadPlannerConfig loads the planner configuration from the repository or uses defaults
func (j *CreateTradePlanJob) loadPlannerConfig() (*planningdomain.PlannerConfiguration, error) {
	// Try to load default config from repository
	if j.configRepo != nil {
		configInterface, err := j.configRepo.GetDefaultConfig()
		if err != nil {
			j.log.Warn().Err(err).Msg("Failed to load default config from repository, using defaults")
		} else if configInterface != nil {
			if config, ok := configInterface.(*planningdomain.PlannerConfiguration); ok {
				j.log.Debug().Str("config_name", config.Name).Msg("Loaded planner config from repository")
				return config, nil
			}
		}
	}

	// Use default config
	j.log.Debug().Msg("Using default planner configuration")
	return planningdomain.NewDefaultConfiguration(), nil
}
