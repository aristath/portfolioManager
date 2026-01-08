package scheduler

import (
	"fmt"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/aristath/sentinel/internal/modules/planning/hash"
	"github.com/aristath/sentinel/internal/modules/portfolio"
	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/rs/zerolog"
)

// PositionRepositoryForHash defines the interface for position repository operations needed for hash generation
type PositionRepositoryForHash interface {
	GetAll() ([]portfolio.Position, error)
}

// SecurityRepositoryForHash defines the interface for security repository operations needed for hash generation
type SecurityRepositoryForHash interface {
	GetAllActive() ([]universe.Security, error)
}

// AllocationRepositoryForHash defines the interface for allocation repository operations needed for hash generation
type AllocationRepositoryForHash interface {
	GetAll() (map[string]float64, error)
}

// GeneratePortfolioHashJob generates and stores portfolio hash
type GeneratePortfolioHashJob struct {
	log               zerolog.Logger
	positionRepo      PositionRepositoryForHash
	securityRepo      SecurityRepositoryForHash
	cashManager       domain.CashManager
	lastPortfolioHash string
}

// GeneratePortfolioHashConfig holds configuration for generate portfolio hash job
type GeneratePortfolioHashConfig struct {
	Log          zerolog.Logger
	PositionRepo PositionRepositoryForHash
	SecurityRepo SecurityRepositoryForHash
	CashManager  domain.CashManager
}

// NewGeneratePortfolioHashJob creates a new generate portfolio hash job
func NewGeneratePortfolioHashJob(cfg GeneratePortfolioHashConfig) *GeneratePortfolioHashJob {
	return &GeneratePortfolioHashJob{
		log:          cfg.Log.With().Str("job", "generate_portfolio_hash").Logger(),
		positionRepo: cfg.PositionRepo,
		securityRepo: cfg.SecurityRepo,
		cashManager:  cfg.CashManager,
	}
}

// Name returns the job name
func (j *GeneratePortfolioHashJob) Name() string {
	return "generate_portfolio_hash"
}

// GetLastPortfolioHash returns the last generated portfolio hash
func (j *GeneratePortfolioHashJob) GetLastPortfolioHash() string {
	return j.lastPortfolioHash
}

// Run executes the generate portfolio hash job
func (j *GeneratePortfolioHashJob) Run() error {
	j.log.Info().Msg("Starting portfolio hash generation")

	// Step 1: Get current portfolio state
	positions, err := j.positionRepo.GetAll()
	if err != nil {
		j.log.Error().Err(err).Msg("Failed to get positions")
		return fmt.Errorf("failed to get positions: %w", err)
	}

	securities, err := j.securityRepo.GetAllActive()
	if err != nil {
		j.log.Error().Err(err).Msg("Failed to get securities")
		return fmt.Errorf("failed to get securities: %w", err)
	}

	// Step 2: Convert positions to hash format
	hashPositions := make([]hash.Position, 0, len(positions))
	for _, pos := range positions {
		hashPositions = append(hashPositions, hash.Position{
			Symbol:   pos.Symbol,
			Quantity: int(pos.Quantity),
		})
	}

	// Step 3: Convert securities to hash format
	hashSecurities := make([]*universe.Security, 0, len(securities))
	for i := range securities {
		hashSecurities = append(hashSecurities, &securities[i])
	}

	// Step 4: Get cash balances from CashManager
	cashBalances := make(map[string]float64)
	if j.cashManager != nil {
		balances, err := j.cashManager.GetAllCashBalances()
		if err != nil {
			j.log.Warn().Err(err).Msg("Failed to get cash balances from CashManager, using empty")
		} else {
			cashBalances = balances
		}
	}

	// Step 5: Generate portfolio hash (no pending orders for now)
	pendingOrders := []hash.PendingOrder{}
	portfolioHash := hash.GeneratePortfolioHash(
		hashPositions,
		hashSecurities,
		cashBalances,
		pendingOrders,
	)

	// Step 6: Store hash
	j.lastPortfolioHash = portfolioHash

	j.log.Info().
		Str("portfolio_hash", portfolioHash).
		Int("positions", len(positions)).
		Int("securities", len(securities)).
		Msg("Portfolio hash generated successfully")

	return nil
}
