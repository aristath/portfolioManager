package scheduler

import (
	"fmt"

	"github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/rs/zerolog"
)

// CreateDividendRecommendationsJob creates recommendations for high-yield dividend reinvestments
type CreateDividendRecommendationsJob struct {
	log              zerolog.Logger
	securityRepo     SecurityRepositoryForDividendsInterface
	yahooClient      YahooClientForDividendsInterface
	highYieldSymbols map[string]SymbolDividendInfoForGroup
	minTradeSize     float64
	recommendations  []domain.HolisticStep
	dividendsToMark  map[string][]int
}

// NewCreateDividendRecommendationsJob creates a new CreateDividendRecommendationsJob
func NewCreateDividendRecommendationsJob(
	securityRepo SecurityRepositoryForDividendsInterface,
	yahooClient YahooClientForDividendsInterface,
	minTradeSize float64,
) *CreateDividendRecommendationsJob {
	return &CreateDividendRecommendationsJob{
		log:             zerolog.Nop(),
		securityRepo:    securityRepo,
		yahooClient:     yahooClient,
		minTradeSize:    minTradeSize,
		recommendations: make([]domain.HolisticStep, 0),
		dividendsToMark: make(map[string][]int),
	}
}

// SetLogger sets the logger for the job
func (j *CreateDividendRecommendationsJob) SetLogger(log zerolog.Logger) {
	j.log = log
}

// SetHighYieldSymbols sets the high-yield symbols to create recommendations for
func (j *CreateDividendRecommendationsJob) SetHighYieldSymbols(symbols map[string]SymbolDividendInfoForGroup) {
	j.highYieldSymbols = symbols
}

// GetRecommendations returns the created recommendations
func (j *CreateDividendRecommendationsJob) GetRecommendations() []domain.HolisticStep {
	return j.recommendations
}

// GetDividendsToMark returns the dividend IDs to mark as reinvested
func (j *CreateDividendRecommendationsJob) GetDividendsToMark() map[string][]int {
	return j.dividendsToMark
}

// Name returns the job name
func (j *CreateDividendRecommendationsJob) Name() string {
	return "create_dividend_recommendations"
}

// Run executes the create dividend recommendations job
func (j *CreateDividendRecommendationsJob) Run() error {
	if len(j.highYieldSymbols) == 0 {
		j.log.Info().Msg("No high-yield symbols to create recommendations for")
		return nil
	}

	recommendations := make([]domain.HolisticStep, 0)
	dividendsToMark := make(map[string][]int)

	for symbol, info := range j.highYieldSymbols {
		// Check if total meets minimum trade size
		if info.TotalAmount < j.minTradeSize {
			j.log.Info().
				Str("symbol", symbol).
				Float64("total", info.TotalAmount).
				Float64("min_trade_size", j.minTradeSize).
				Msg("Total below min trade size, skipping recommendation")
			continue
		}

		step, err := j.createSameSecurityReinvestment(symbol, info)
		if err != nil {
			j.log.Error().
				Err(err).
				Str("symbol", symbol).
				Msg("Failed to create same-security reinvestment")
			continue
		}

		if step != nil {
			recommendations = append(recommendations, *step)
			dividendsToMark[symbol] = info.DividendIDs
		}
	}

	j.recommendations = recommendations
	j.dividendsToMark = dividendsToMark

	j.log.Info().
		Int("recommendations", len(recommendations)).
		Int("symbols", len(j.highYieldSymbols)).
		Msg("Successfully created dividend recommendations")

	return nil
}

// createSameSecurityReinvestment creates a BUY step for reinvesting in the same security
func (j *CreateDividendRecommendationsJob) createSameSecurityReinvestment(
	symbol string,
	info SymbolDividendInfoForGroup,
) (*domain.HolisticStep, error) {
	// Get security info for name and other details
	security, err := j.securityRepo.GetBySymbol(symbol)
	if err != nil || security == nil {
		j.log.Warn().
			Str("symbol", symbol).
			Msg("Security not found in universe, skipping")
		return nil, fmt.Errorf("security %s not found", symbol)
	}

	// Get current security price from Yahoo Finance
	yahooSymbol := security.YahooSymbol
	if yahooSymbol == "" {
		yahooSymbol = symbol
	}

	pricePtr, err := j.yahooClient.GetCurrentPrice(yahooSymbol, nil, 3)
	if err != nil || pricePtr == nil || *pricePtr <= 0 {
		j.log.Warn().
			Str("symbol", symbol).
			Str("yahoo_symbol", yahooSymbol).
			Msg("Could not get current price, skipping")
		return nil, fmt.Errorf("invalid price for %s", symbol)
	}

	price := *pricePtr

	// Calculate shares to buy
	quantity := int(info.TotalAmount / price)
	if quantity <= 0 {
		j.log.Warn().
			Str("symbol", symbol).
			Int("quantity", quantity).
			Msg("Calculated quantity is invalid, skipping")
		return nil, fmt.Errorf("invalid quantity for %s", symbol)
	}

	// Adjust for min_lot
	if security.MinLot > 1 {
		quantity = (quantity / security.MinLot) * security.MinLot
		if quantity == 0 {
			quantity = security.MinLot
		}
	}

	estimatedValue := float64(quantity) * price

	// Create BUY step
	step := &domain.HolisticStep{
		Symbol:         symbol,
		Name:           security.Name,
		Side:           "BUY",
		Quantity:       quantity,
		EstimatedPrice: price,
		EstimatedValue: estimatedValue,
		Currency:       security.Currency,
		Reason: fmt.Sprintf("Dividend reinvestment (high yield): %.2f EUR from %d dividend(s)",
			info.TotalAmount, info.DividendCount),
		Narrative: fmt.Sprintf("Reinvest %.2f EUR dividend in %s (high yield security)",
			info.TotalAmount, symbol),
	}

	return step, nil
}
