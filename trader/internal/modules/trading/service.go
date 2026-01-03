package trading

import (
	"fmt"
	"strings"
	"time"

	"github.com/aristath/arduino-trader/internal/clients/tradernet"
	"github.com/rs/zerolog"
)

// TradingService handles trade-related business logic
type TradingService struct {
	log             zerolog.Logger
	tradeRepo       *TradeRepository
	tradernetClient *tradernet.Client
}

// NewTradingService creates a new trading service
func NewTradingService(
	tradeRepo *TradeRepository,
	tradernetClient *tradernet.Client,
	log zerolog.Logger,
) *TradingService {
	return &TradingService{
		log:             log.With().Str("service", "trading").Logger(),
		tradeRepo:       tradeRepo,
		tradernetClient: tradernetClient,
	}
}

// SyncFromTradernet synchronizes trade history from Tradernet microservice
// Returns count of newly synced trades
func (s *TradingService) SyncFromTradernet() error {
	s.log.Info().Msg("Syncing trades from Tradernet")

	// Get recent trades from Tradernet (last 100 trades)
	trades, err := s.tradernetClient.GetExecutedTrades(100)
	if err != nil {
		return fmt.Errorf("failed to get trades from Tradernet: %w", err)
	}

	// Sync trades to database
	syncedCount := 0
	for _, trade := range trades {
		// Parse trade side
		side, err := TradeSideFromString(trade.Side)
		if err != nil {
			s.log.Error().
				Err(err).
				Str("order_id", trade.OrderID).
				Str("side", trade.Side).
				Msg("Invalid trade side")
			continue
		}

		// Parse executed_at timestamp
		executedAt, err := time.Parse(time.RFC3339, trade.ExecutedAt)
		if err != nil {
			s.log.Error().
				Err(err).
				Str("order_id", trade.OrderID).
				Str("executed_at", trade.ExecutedAt).
				Msg("Invalid executed_at timestamp")
			continue
		}

		// Convert tradernet.Trade to trading.Trade domain model
		dbTrade := Trade{
			OrderID:    trade.OrderID,
			Symbol:     trade.Symbol,
			Side:       side,
			Quantity:   trade.Quantity,
			Price:      trade.Price,
			ExecutedAt: executedAt,
			Source:     "tradernet",
			Currency:   "EUR", // Default, should be from trade data
			BucketID:   "",    // Empty for automatic sync
		}

		// Insert trade to database (idempotent via order_id unique constraint)
		if err := s.tradeRepo.Create(dbTrade); err != nil {
			// Skip if already exists (duplicate order_id)
			if strings.Contains(err.Error(), "UNIQUE constraint") {
				continue
			}
			s.log.Error().
				Err(err).
				Str("order_id", trade.OrderID).
				Msg("Failed to insert trade")
			continue
		}

		syncedCount++
	}

	s.log.Info().
		Int("total", len(trades)).
		Int("synced", syncedCount).
		Msg("Trade sync completed")

	return nil
}
