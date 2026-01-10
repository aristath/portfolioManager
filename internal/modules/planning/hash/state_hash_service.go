package hash

import (
	"fmt"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/aristath/sentinel/internal/modules/allocation"
	"github.com/aristath/sentinel/internal/modules/portfolio"
	"github.com/aristath/sentinel/internal/modules/settings"
	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/rs/zerolog"
)

// StateHashService fetches all state and calculates unified hash
type StateHashService struct {
	positionRepo    *portfolio.PositionRepository
	securityRepo    *universe.SecurityRepository
	scoreRepo       universe.ScoreRepositoryInterface
	cashManager     domain.CashManager
	settingsRepo    *settings.Repository
	settingsService *settings.Service
	allocRepo       *allocation.Repository
	exchangeService domain.CurrencyExchangeServiceInterface
	brokerClient    domain.BrokerClient
	log             zerolog.Logger
}

// NewStateHashService creates new service
func NewStateHashService(
	positionRepo *portfolio.PositionRepository,
	securityRepo *universe.SecurityRepository,
	scoreRepo universe.ScoreRepositoryInterface,
	cashManager domain.CashManager,
	settingsRepo *settings.Repository,
	settingsService *settings.Service,
	allocRepo *allocation.Repository,
	exchangeService domain.CurrencyExchangeServiceInterface,
	brokerClient domain.BrokerClient,
	log zerolog.Logger,
) *StateHashService {
	return &StateHashService{
		positionRepo:    positionRepo,
		securityRepo:    securityRepo,
		scoreRepo:       scoreRepo,
		cashManager:     cashManager,
		settingsRepo:    settingsRepo,
		settingsService: settingsService,
		allocRepo:       allocRepo,
		exchangeService: exchangeService,
		brokerClient:    brokerClient,
		log:             log.With().Str("component", "state_hash_service").Logger(),
	}
}

// CalculateCurrentHash fetches all state and computes unified hash
func (s *StateHashService) CalculateCurrentHash() (string, error) {
	// 1. Get positions
	positions, err := s.positionRepo.GetAll()
	if err != nil {
		return "", fmt.Errorf("failed to get positions: %w", err)
	}

	// Convert to hash.Position format
	hashPositions := make([]Position, len(positions))
	for i, pos := range positions {
		hashPositions[i] = Position{
			Symbol:   pos.Symbol,
			Quantity: int(pos.Quantity),
		}
	}

	// 2. Get securities
	securitiesSlice, err := s.securityRepo.GetAll()
	if err != nil {
		return "", fmt.Errorf("failed to get securities: %w", err)
	}

	// Convert to pointer slice for hash function
	securities := make([]*universe.Security, len(securitiesSlice))
	for i := range securitiesSlice {
		securities[i] = &securitiesSlice[i]
	}

	// 3. Get scores
	scores, err := s.scoreRepo.GetAll()
	if err != nil {
		return "", fmt.Errorf("failed to get scores: %w", err)
	}

	// 4. Get cash balances
	cashBalances, err := s.cashManager.GetAllCashBalances()
	if err != nil {
		return "", fmt.Errorf("failed to get cash balances: %w", err)
	}

	// 5. Get pending orders from broker
	var pendingOrders []PendingOrder
	if s.brokerClient != nil && s.brokerClient.IsConnected() {
		orders, err := s.brokerClient.GetPendingOrders()
		if err == nil {
			for _, order := range orders {
				pendingOrders = append(pendingOrders, PendingOrder{
					Symbol:   order.Symbol,
					Side:     order.Side,
					Quantity: int(order.Quantity),
					Price:    order.Price,
					Currency: order.Currency,
				})
			}
		}
	}

	// 6. Get exchange rates for active currencies
	rates := s.fetchExchangeRates(cashBalances)

	// 7. Get relevant settings
	settingsMap := s.fetchRelevantSettings()

	// 8. Get allocations
	allocations, err := s.allocRepo.GetAll()
	if err != nil {
		s.log.Warn().Err(err).Msg("Failed to get allocations, using empty")
		allocations = make(map[string]float64)
	}

	// 9. Generate unified hash
	return GenerateUnifiedStateHash(
		hashPositions, securities, cashBalances, pendingOrders,
		scores, rates, settingsMap, allocations,
	), nil
}

// fetchExchangeRates gets rates for all currency pairs in portfolio
func (s *StateHashService) fetchExchangeRates(cashBalances map[string]float64) map[string]float64 {
	// Get currencies from cash balances
	currencies := make([]string, 0, len(cashBalances))
	for currency := range cashBalances {
		currencies = append(currencies, currency)
	}

	// Always include EUR (base currency)
	if len(currencies) == 0 {
		currencies = []string{"EUR"}
	}

	rates := make(map[string]float64)
	for _, from := range currencies {
		for _, to := range currencies {
			if from != to {
				rate, err := s.exchangeService.GetRate(from, to)
				if err == nil {
					rates[fmt.Sprintf("%s/%s", from, to)] = rate
				}
			}
		}
	}
	return rates
}

// fetchRelevantSettings gets settings that affect recommendations
func (s *StateHashService) fetchRelevantSettings() map[string]interface{} {
	relevantKeys := []string{
		"min_security_score",
		"min_hold_days",
		"sell_cooldown_days",
		"max_loss_threshold",
		"target_annual_return",
		"optimizer_blend",
		"optimizer_target_return",
		"transaction_cost_fixed",
		"transaction_cost_percent",
		"min_cash_reserve",
		"max_plan_depth",
	}

	settingsMap := make(map[string]interface{})
	for _, key := range relevantKeys {
		// Special handling for min_security_score: use adjusted value based on risk_tolerance
		if key == "min_security_score" && s.settingsService != nil {
			adjustedValue, err := s.settingsService.GetAdjustedMinSecurityScore()
			if err == nil {
				settingsMap[key] = fmt.Sprintf("%f", adjustedValue)
			} else {
				s.log.Warn().Err(err).Msg("Failed to get adjusted min_security_score, using raw value")
				// Fallback to raw value
				val, err := s.settingsRepo.Get(key)
				if err == nil && val != nil {
					settingsMap[key] = *val
				}
			}
		} else {
			val, err := s.settingsRepo.Get(key)
			if err == nil && val != nil {
				settingsMap[key] = *val
			}
		}
	}
	return settingsMap
}
