package order_book

import (
	"fmt"
	"math"

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

	// 3. Cross-validate with secondary price source (if available) - ASYMMETRIC validation
	if err := s.validatePriceAgainstSecondarySource(symbol, side, midpoint); err != nil {
		// BLOCKS if midpoint would cause bad trade
		return 0, err
	}

	// 4. Calculate limit with buffer
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

// validatePriceAgainstSecondarySource checks if midpoint price is reasonable
// Uses ASYMMETRIC validation: only blocks when midpoint would cause bad trades
// Implements "BUY CHEAP, SELL HIGH" principle
func (s *Service) validatePriceAgainstSecondarySource(symbol, side string, midpointPrice float64) error {
	// Try to fetch secondary price source for validation
	validationPrice, err := s.fetchValidationPrice(symbol)
	if err != nil {
		// Validation source unavailable - proceed with midpoint only
		s.log.Warn().
			Str("symbol", symbol).
			Err(err).
			Msg("Validation price unavailable - using midpoint without validation")
		return nil
	}

	// Get threshold (default 50%)
	threshold := s.getSettingFloat("price_discrepancy_threshold", 0.50)

	// ASYMMETRIC VALIDATION: Enforce "BUY CHEAP, SELL HIGH" principle
	// Only block when midpoint would cause BAD trade
	var shouldBlock bool
	var reason string

	if side == "BUY" {
		// BUY CHEAP: Block if midpoint is significantly HIGHER than validation price (overpaying)
		// Allow if midpoint is LOWER than validation price (buying cheap is good!)
		maxAllowedPrice := validationPrice * (1.0 + threshold)
		if midpointPrice > maxAllowedPrice {
			shouldBlock = true
			reason = fmt.Sprintf("overpaying: midpoint %.2f > validation %.2f * %.2f = %.2f",
				midpointPrice, validationPrice, 1.0+threshold, maxAllowedPrice)
		}
	} else if side == "SELL" {
		// SELL HIGH: Block if midpoint is significantly LOWER than validation price (underselling)
		// Allow if midpoint is HIGHER than validation price (selling high is good!)
		minAllowedPrice := validationPrice * (1.0 - threshold)
		if midpointPrice < minAllowedPrice {
			shouldBlock = true
			reason = fmt.Sprintf("underselling: midpoint %.2f < validation %.2f * %.2f = %.2f",
				midpointPrice, validationPrice, 1.0-threshold, minAllowedPrice)
		}
	}

	if shouldBlock {
		// BLOCK - API bug detected (bad trade would result)
		s.log.Error().
			Str("symbol", symbol).
			Str("side", side).
			Float64("midpoint_price", midpointPrice).
			Float64("validation_price", validationPrice).
			Float64("threshold_pct", threshold*100).
			Str("reason", reason).
			Msg("Price validation FAILED - BLOCKING trade (API bug suspected)")

		return fmt.Errorf("price validation failed (%s) - API bug suspected", reason)
	}

	// Log validation success
	discrepancyPct := math.Abs(midpointPrice-validationPrice) / validationPrice * 100
	s.log.Info().
		Str("symbol", symbol).
		Str("side", side).
		Float64("midpoint_price", midpointPrice).
		Float64("validation_price", validationPrice).
		Float64("discrepancy_pct", discrepancyPct).
		Msg("Price validation passed - using midpoint")

	return nil
}

// fetchValidationPrice gets price for validation using PriceValidator interface
func (s *Service) fetchValidationPrice(symbol string) (float64, error) {
	// Use PriceValidator interface (decouples from specific implementation)
	price, err := s.priceValidator.GetValidationPrice(symbol)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch validation price: %w", err)
	}

	if price == nil {
		return 0, fmt.Errorf("validation price is nil")
	}

	if *price <= 0 {
		return 0, fmt.Errorf("validation price is invalid: %.2f", *price)
	}

	return *price, nil
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
