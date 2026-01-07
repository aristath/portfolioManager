package scheduler

import (
	"github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/rs/zerolog"
)

// ExecuteDividendTradesJob executes dividend reinvestment trades and marks dividends as reinvested
type ExecuteDividendTradesJob struct {
	log                   zerolog.Logger
	dividendRepo          DividendRepositoryInterface
	tradeExecutionService TradeExecutionServiceInterface
	recommendations       []domain.HolisticStep
	dividendsToMark       map[string][]int
	executedCount         int
}

// NewExecuteDividendTradesJob creates a new ExecuteDividendTradesJob
func NewExecuteDividendTradesJob(
	dividendRepo DividendRepositoryInterface,
	tradeExecutionService TradeExecutionServiceInterface,
) *ExecuteDividendTradesJob {
	return &ExecuteDividendTradesJob{
		log:                   zerolog.Nop(),
		dividendRepo:          dividendRepo,
		tradeExecutionService: tradeExecutionService,
		dividendsToMark:       make(map[string][]int),
	}
}

// SetLogger sets the logger for the job
func (j *ExecuteDividendTradesJob) SetLogger(log zerolog.Logger) {
	j.log = log
}

// SetRecommendations sets the recommendations to execute
func (j *ExecuteDividendTradesJob) SetRecommendations(recommendations []domain.HolisticStep, dividendsToMark map[string][]int) {
	j.recommendations = recommendations
	j.dividendsToMark = dividendsToMark
}

// GetExecutedCount returns the number of successfully executed trades
func (j *ExecuteDividendTradesJob) GetExecutedCount() int {
	return j.executedCount
}

// Name returns the job name
func (j *ExecuteDividendTradesJob) Name() string {
	return "execute_dividend_trades"
}

// Run executes the execute dividend trades job
func (j *ExecuteDividendTradesJob) Run() error {
	if len(j.recommendations) == 0 {
		j.log.Info().Msg("No dividend reinvestment trades to execute")
		return nil
	}

	if j.tradeExecutionService == nil {
		j.log.Warn().Msg("Trade execution service not available, skipping trade execution")
		return nil
	}

	// Convert domain.HolisticStep to TradeRecommendationForDividends
	tradeRecs := make([]TradeRecommendationForDividends, 0, len(j.recommendations))
	for _, rec := range j.recommendations {
		tradeRecs = append(tradeRecs, TradeRecommendationForDividends{
			Symbol:         rec.Symbol,
			Side:           rec.Side,
			Quantity:       float64(rec.Quantity),
			EstimatedPrice: rec.EstimatedPrice,
			Currency:       rec.Currency,
			Reason:         rec.Reason,
		})
	}

	// Execute all trades via trade execution service
	results := j.tradeExecutionService.ExecuteTrades(tradeRecs)

	// Process results and mark dividends as reinvested
	executedCount := 0
	for i, result := range results {
		rec := j.recommendations[i]

		if result.Status == "success" {
			j.log.Info().
				Str("symbol", rec.Symbol).
				Str("side", rec.Side).
				Int("quantity", rec.Quantity).
				Float64("estimated_value", rec.EstimatedValue).
				Msg("Successfully executed dividend reinvestment trade")

			// Mark dividends as reinvested
			if dividendIDs, ok := j.dividendsToMark[rec.Symbol]; ok {
				for _, dividendID := range dividendIDs {
					if err := j.dividendRepo.MarkReinvested(dividendID, rec.Quantity); err != nil {
						j.log.Error().
							Err(err).
							Int("dividend_id", dividendID).
							Str("symbol", rec.Symbol).
							Msg("Failed to mark dividend as reinvested")
					}
				}

				j.log.Info().
					Str("symbol", rec.Symbol).
					Int("dividends_marked", len(dividendIDs)).
					Msg("Marked dividends as reinvested")

				executedCount++
			}
		} else {
			errorMsg := ""
			if result.Error != nil {
				errorMsg = *result.Error
			}
			j.log.Warn().
				Str("symbol", rec.Symbol).
				Str("status", result.Status).
				Str("error", errorMsg).
				Msg("Failed to execute dividend reinvestment trade")
		}
	}

	j.executedCount = executedCount

	j.log.Info().
		Int("executed", executedCount).
		Int("total", len(j.recommendations)).
		Msg("Dividend reinvestment trades execution completed")

	return nil
}
