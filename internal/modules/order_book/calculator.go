package order_book

import (
	"fmt"

	"github.com/aristath/sentinel/internal/domain"
)

// calculateOptimalLimitPrice is the main method for calculating optimal limit price
// Uses bid-ask midpoint strategy for better execution than simple best price
func (s *Service) calculateOptimalLimitPrice(symbol, side string, buffer float64) (float64, error) {
	// 1. Fetch Level 1 quote (best bid + best ask)
	quote, err := s.brokerClient.GetLevel1Quote(symbol)
	if err != nil {
		s.log.Error().Err(err).Str("symbol", symbol).Msg("Failed to fetch quote data")
		return 0, fmt.Errorf("quote data unavailable - cannot trade: %w", err)
	}

	// 2. Calculate bid-ask midpoint
	midpoint, err := s.calculateMidpoint(quote)
	if err != nil {
		s.log.Error().Err(err).Str("symbol", symbol).Msg("Failed to calculate midpoint")
		return 0, err
	}

	// 3. Calculate limit with buffer
	limitPrice := s.calculateLimitWithBuffer(midpoint, buffer, side)

	s.log.Info().
		Str("symbol", symbol).
		Str("side", side).
		Float64("midpoint", midpoint).
		Float64("buffer_pct", buffer*100).
		Float64("limit_price", limitPrice).
		Msg("Calculated optimal limit price using bid-ask midpoint")

	return limitPrice, nil
}

// calculateMidpoint calculates the midpoint between best bid and best ask
// This is a well-established trading strategy for optimal execution
func (s *Service) calculateMidpoint(quote *domain.BrokerOrderBook) (float64, error) {
	// Extract best bid and best ask
	if len(quote.Bids) == 0 || len(quote.Asks) == 0 {
		return 0, fmt.Errorf("incomplete quote data: missing bid or ask")
	}

	bestBid := quote.Bids[0].Price
	bestAsk := quote.Asks[0].Price

	// Validate prices
	if bestBid <= 0 || bestAsk <= 0 {
		return 0, fmt.Errorf("invalid prices: bid=%.2f, ask=%.2f", bestBid, bestAsk)
	}

	// Sanity check: ask should be >= bid
	if bestAsk < bestBid {
		s.log.Warn().
			Float64("bid", bestBid).
			Float64("ask", bestAsk).
			Msg("Inverted market: ask < bid (unusual but can happen)")
	}

	// Calculate midpoint
	midpoint := (bestBid + bestAsk) / 2.0

	s.log.Debug().
		Str("symbol", quote.Symbol).
		Float64("bid", bestBid).
		Float64("ask", bestAsk).
		Float64("midpoint", midpoint).
		Float64("spread_pct", ((bestAsk-bestBid)/midpoint)*100).
		Msg("Calculated bid-ask midpoint")

	return midpoint, nil
}

// calculateLimitWithBuffer applies buffer to price
func (s *Service) calculateLimitWithBuffer(price, buffer float64, side string) float64 {
	if side == "BUY" {
		// For BUY: willing to pay up to midpoint + buffer
		// Example: midpoint=$100, buffer=5% → limit=$105
		return price * (1 + buffer)
	}

	// For SELL: willing to accept down to midpoint - buffer
	// Example: midpoint=$100, buffer=5% → limit=$95
	return price * (1 - buffer)
}
