package optimization

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aristath/sentinel/internal/domain"
	planningdomain "github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/aristath/sentinel/internal/modules/portfolio"
	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/rs/zerolog"
)

// PriceClientForWeights defines interface for getting batch prices
// This adapts domain.BrokerClient.GetQuotes() to the signature needed
type PriceClientForWeights interface {
	GetBatchQuotes(symbolMap map[string]*string) (map[string]*float64, error)
}

// MarketHoursForWeights defines interface for market hours checking
type MarketHoursForWeights interface {
	AnyMajorMarketOpen(t time.Time) bool
}

// brokerClientAdapter adapts domain.BrokerClient to PriceClientForWeights
type brokerClientAdapter struct {
	client domain.BrokerClient
}

func (a *brokerClientAdapter) GetBatchQuotes(symbolMap map[string]*string) (map[string]*float64, error) {
	if a.client == nil {
		return nil, fmt.Errorf("broker client not available")
	}
	// Extract symbols from map
	symbols := make([]string, 0, len(symbolMap))
	for symbol := range symbolMap {
		symbols = append(symbols, symbol)
	}

	// Get quotes from broker
	quotes, err := a.client.GetQuotes(symbols)
	if err != nil {
		return nil, fmt.Errorf("failed to get broker quotes: %w", err)
	}

	// Convert to price map
	prices := make(map[string]*float64)
	for symbol, quote := range quotes {
		if quote != nil && quote.Price > 0 {
			price := quote.Price
			prices[symbol] = &price
		}
	}

	return prices, nil
}

// RepositoryForWeights defines interface for allocation repository operations
type RepositoryForWeights interface {
	GetAll() (map[string]float64, error)
}

// OptimizerWeightsService calculates optimizer target weights for the current portfolio
type OptimizerWeightsService struct {
	log                    zerolog.Logger
	positionRepo           portfolio.PositionRepositoryInterface // Use interface for testability
	securityRepo           universe.SecurityRepositoryInterface  // Use interface for testability
	allocRepo              RepositoryForWeights                  // Minimal interface for testability
	cashManager            domain.CashManager
	priceClient            PriceClientForWeights // Adapter for broker client (exported for testing)
	optimizerService       OptimizerForWeights   // Interface for testability
	priceConversionService PriceConversionForWeights
	plannerConfigRepo      PlannerConfigForWeights
	clientDataRepo         ClientDataForWeights
	marketHoursService     MarketHoursForWeights // Interface for market hours checking
}

// OptimizerForWeights defines interface for optimizer operations
type OptimizerForWeights interface {
	Optimize(state PortfolioState, settings Settings) (*Result, error)
}

// PriceConversionForWeights defines interface for price conversion
type PriceConversionForWeights interface {
	ConvertPricesToEUR(prices map[string]float64, securities []universe.Security) map[string]float64
}

// PlannerConfigForWeights defines interface for planner config repository
type PlannerConfigForWeights interface {
	GetDefaultConfig() (*planningdomain.PlannerConfiguration, error)
}

// ClientDataForWeights defines interface for client data cache
type ClientDataForWeights interface {
	GetIfFresh(table, key string) (json.RawMessage, error)
	Get(table, key string) (json.RawMessage, error)
	Store(table, key string, data interface{}, ttl time.Duration) error
}

// NewOptimizerWeightsService creates a new OptimizerWeightsService
// If brokerClient is provided, it will be adapted to PriceClientForWeights automatically
// For testing, you can pass nil for brokerClient and provide priceClient directly via NewOptimizerWeightsServiceWithPriceClient
func NewOptimizerWeightsService(
	positionRepo portfolio.PositionRepositoryInterface,
	securityRepo universe.SecurityRepositoryInterface,
	allocRepo RepositoryForWeights,
	cashManager domain.CashManager,
	brokerClient domain.BrokerClient,
	optimizerService OptimizerForWeights,
	priceConversionService PriceConversionForWeights,
	plannerConfigRepo PlannerConfigForWeights,
	clientDataRepo ClientDataForWeights,
	marketHoursService MarketHoursForWeights,
) *OptimizerWeightsService {
	// Create adapter for broker client
	var priceClient PriceClientForWeights
	if brokerClient != nil {
		priceClient = &brokerClientAdapter{client: brokerClient}
	}

	return &OptimizerWeightsService{
		log:                    zerolog.Nop(),
		positionRepo:           positionRepo,
		securityRepo:           securityRepo,
		allocRepo:              allocRepo,
		cashManager:            cashManager,
		priceClient:            priceClient,
		optimizerService:       optimizerService,
		priceConversionService: priceConversionService,
		plannerConfigRepo:      plannerConfigRepo,
		clientDataRepo:         clientDataRepo,
		marketHoursService:     marketHoursService,
	}
}

// CalculateWeights calculates optimizer target weights for the current portfolio
func (s *OptimizerWeightsService) CalculateWeights(ctx context.Context) (map[string]float64, error) {
	if s.optimizerService == nil {
		return nil, fmt.Errorf("optimizer service not available")
	}

	// Step 1: Get positions
	positions, err := s.positionRepo.GetAll()
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get positions")
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	// Step 2: Get securities
	securities, err := s.securityRepo.GetAllActive()
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get securities")
		return nil, fmt.Errorf("failed to get securities: %w", err)
	}

	// Step 3: Get cash balances
	cashBalances := make(map[string]float64)
	if s.cashManager != nil {
		balances, err := s.cashManager.GetAllCashBalances()
		if err != nil {
			s.log.Warn().Err(err).Msg("Failed to get cash balances, using empty")
		} else {
			cashBalances = balances
		}
	}

	// Step 4: Fetch current prices (returns ISIN-keyed map)
	currentPrices := s.fetchCurrentPrices(securities)

	// Step 5: Calculate portfolio value using ISIN lookup
	portfolioValue := cashBalances["EUR"]
	for _, pos := range positions {
		if pos.ISIN == "" {
			s.log.Warn().
				Str("symbol", pos.Symbol).
				Msg("Position missing ISIN, skipping in portfolio value")
			continue
		}
		if price, ok := currentPrices[pos.ISIN]; ok {
			portfolioValue += price * float64(pos.Quantity)
		}
	}

	// Step 6: Get allocation targets
	allocations, err := s.allocRepo.GetAll()
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get allocations")
		return nil, fmt.Errorf("failed to get allocations: %w", err)
	}

	// Step 7: Extract geography and industry targets (raw - normalization happens in constraints)
	geographyTargets := make(map[string]float64)
	industryTargets := make(map[string]float64)
	for key, value := range allocations {
		if strings.HasPrefix(key, "geography:") {
			geography := strings.TrimPrefix(key, "geography:")
			geographyTargets[geography] = value
		} else if strings.HasPrefix(key, "industry:") {
			industry := strings.TrimPrefix(key, "industry:")
			industryTargets[industry] = value
		}
	}

	// Step 8: Convert positions to optimizer format (ISIN-keyed map)
	optimizerPositions := make(map[string]Position)
	for _, pos := range positions {
		isin := pos.ISIN
		if isin == "" {
			s.log.Warn().
				Str("symbol", pos.Symbol).
				Msg("Position missing ISIN, skipping")
			continue
		}

		valueEUR := 0.0
		if price, ok := currentPrices[isin]; ok {
			valueEUR = price * float64(pos.Quantity)
		}
		optimizerPositions[isin] = Position{
			ISIN:     isin,
			Quantity: float64(pos.Quantity),
			ValueEUR: valueEUR,
		}
	}

	// Step 9: Convert securities to optimizer format
	optimizerSecurities := make([]Security, 0, len(securities))
	for _, sec := range securities {
		if sec.ISIN == "" {
			s.log.Warn().
				Str("symbol", sec.Symbol).
				Msg("Security missing ISIN, skipping")
			continue
		}
		optimizerSecurities = append(optimizerSecurities, Security{
			ISIN:               sec.ISIN,
			Symbol:             sec.Symbol,
			Geography:          sec.Geography,
			Industry:           sec.Industry,
			MinPortfolioTarget: 0.0,
			MaxPortfolioTarget: 1.0,
			AllowBuy:           sec.AllowBuy,
			AllowSell:          true,
			MinLot:             1.0,
			PriorityMultiplier: 1.0,
			TargetPriceEUR:     0.0,
		})
	}

	// Step 10: Build portfolio state
	state := PortfolioState{
		Securities:       optimizerSecurities,
		Positions:        optimizerPositions,
		PortfolioValue:   portfolioValue,
		CurrentPrices:    currentPrices,
		CashBalance:      cashBalances["EUR"],
		GeographyTargets: geographyTargets,
		IndustryTargets:  industryTargets,
		DividendBonuses:  make(map[string]float64),
	}

	// Step 11: Get optimizer settings from planner config
	plannerConfig, err := s.getPlannerConfig()
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get planner config, using defaults")
		// Fallback to defaults
		plannerConfig = &Settings{
			Blend:                    0.5,
			TargetReturn:             0.11,
			MinCashReserve:           500.0,
			MinTradeAmount:           0.0,
			TransactionCostPct:       0.002,
			MaxConcentration:         0.25,
			TargetReturnThresholdPct: 0.80,
		}
	}

	settings := *plannerConfig

	// Step 12: Run optimization
	result, err := s.optimizerService.Optimize(state, settings)
	if err != nil {
		s.log.Error().Err(err).Msg("Optimizer failed")
		return nil, fmt.Errorf("optimizer failed: %w", err)
	}

	if !result.Success {
		s.log.Error().Msg("Optimizer returned unsuccessful result")
		return nil, fmt.Errorf("optimizer returned unsuccessful result")
	}

	s.log.Info().
		Int("target_count", len(result.TargetWeights)).
		Msg("Successfully retrieved optimizer target weights")

	return result.TargetWeights, nil
}

// fetchCurrentPrices fetches current prices for all securities and converts them to EUR
// Uses cache with market-aware TTL: 30min when markets open, 24h when closed
// Returns ISIN-keyed map (not Symbol-keyed)
func (s *OptimizerWeightsService) fetchCurrentPrices(securities []universe.Security) map[string]float64 {
	prices := make(map[string]float64)

	if len(securities) == 0 {
		return prices
	}

	// Determine TTL based on market hours
	ttl := 24 * time.Hour // Default: markets closed
	if s.marketHoursService != nil && s.marketHoursService.AnyMajorMarketOpen(time.Now()) {
		ttl = 30 * time.Minute // Markets open: shorter TTL
	}

	// Step 1: Check cache for each security
	needFetch := make([]universe.Security, 0)
	for _, sec := range securities {
		if sec.ISIN == "" {
			continue
		}

		// Try to get cached price
		if s.clientDataRepo != nil {
			if data, err := s.clientDataRepo.GetIfFresh("current_prices", sec.ISIN); err == nil && data != nil {
				var cachedPrice float64
				if json.Unmarshal(data, &cachedPrice) == nil {
					prices[sec.ISIN] = cachedPrice
					continue
				}
			}
		}

		// Cache miss - need to fetch
		needFetch = append(needFetch, sec)
	}

	cacheHits := len(securities) - len(needFetch)
	if cacheHits > 0 {
		s.log.Debug().
			Int("cache_hits", cacheHits).
			Int("need_fetch", len(needFetch)).
			Msg("Price cache hits")
	}

	// Step 2: If all prices were cached, return early
	if len(needFetch) == 0 {
		s.log.Info().
			Int("total", len(prices)).
			Msg("All prices served from cache")
		return prices
	}

	// Step 3: Fetch missing prices from API
	if s.priceClient == nil {
		s.log.Warn().Msg("Price client not available, cannot fetch missing prices")
		// Fall back to stale cache for missing prices
		return s.fallbackToStaleCache(securities, prices)
	}

	// Build symbol map for price API (only for securities that need fetching)
	symbolMap := make(map[string]*string)
	for _, security := range needFetch {
		symbolMap[security.Symbol] = nil
	}

	// Fetch batch quotes
	quotes, err := s.priceClient.GetBatchQuotes(symbolMap)
	if err != nil {
		s.log.Warn().Err(err).Msg("Failed to fetch batch quotes, falling back to stale cache")
		return s.fallbackToStaleCache(securities, prices)
	}

	// Convert quotes to price map (native currencies)
	nativePrices := make(map[string]float64)
	for symbol, pricePtr := range quotes {
		if pricePtr != nil {
			nativePrices[symbol] = *pricePtr
		}
	}

	// Convert all prices to EUR
	var eurPricesSymbol map[string]float64
	if s.priceConversionService != nil {
		eurPricesSymbol = s.priceConversionService.ConvertPricesToEUR(nativePrices, needFetch)
	} else {
		s.log.Warn().Msg("Price conversion service not available, using native currency prices")
		eurPricesSymbol = nativePrices
	}

	// Build Symbol → ISIN mapping for the fetched securities
	symbolToISIN := make(map[string]string)
	for _, sec := range needFetch {
		if sec.ISIN != "" && sec.Symbol != "" {
			symbolToISIN[sec.Symbol] = sec.ISIN
		}
	}

	// Transform Symbol keys → ISIN keys and cache the results
	fetchedCount := 0
	for symbol, price := range eurPricesSymbol {
		if isin, ok := symbolToISIN[symbol]; ok {
			prices[isin] = price
			fetchedCount++

			// Cache the fetched price
			if s.clientDataRepo != nil {
				if err := s.clientDataRepo.Store("current_prices", isin, price, ttl); err != nil {
					s.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache price")
				}
			}
		} else {
			s.log.Warn().
				Str("symbol", symbol).
				Msg("No ISIN mapping found for symbol, skipping price")
		}
	}

	s.log.Info().
		Int("from_cache", cacheHits).
		Int("fetched", fetchedCount).
		Int("total", len(prices)).
		Dur("ttl", ttl).
		Msg("Fetched and cached prices")

	return prices
}

// fallbackToStaleCache attempts to get stale cached prices when API fails
func (s *OptimizerWeightsService) fallbackToStaleCache(securities []universe.Security, prices map[string]float64) map[string]float64 {
	if s.clientDataRepo == nil {
		return prices
	}

	staleCount := 0
	for _, sec := range securities {
		if sec.ISIN == "" {
			continue
		}

		// Skip if we already have this price (from fresh cache)
		if _, ok := prices[sec.ISIN]; ok {
			continue
		}

		// Try to get stale cached price
		if data, err := s.clientDataRepo.Get("current_prices", sec.ISIN); err == nil && data != nil {
			var cachedPrice float64
			if json.Unmarshal(data, &cachedPrice) == nil {
				prices[sec.ISIN] = cachedPrice
				staleCount++
			}
		}
	}

	if staleCount > 0 {
		s.log.Info().
			Int("stale_fallback", staleCount).
			Msg("Used stale cached prices as fallback")
	}

	return prices
}

// getPlannerConfig fetches optimizer settings from planner configuration
func (s *OptimizerWeightsService) getPlannerConfig() (*Settings, error) {
	if s.plannerConfigRepo == nil {
		return nil, fmt.Errorf("planner config repository not available")
	}

	config, err := s.plannerConfigRepo.GetDefaultConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get planner config: %w", err)
	}

	// Convert planner config to optimizer settings
	settings := &Settings{
		Blend:                    config.OptimizerBlend,
		TargetReturn:             config.OptimizerTargetReturn,
		MinCashReserve:           config.MinCashReserve,
		MinTradeAmount:           0.0,
		TransactionCostPct:       config.TransactionCostPercent,
		MaxConcentration:         0.25,
		TargetReturnThresholdPct: 0.80,
	}

	return settings, nil
}
