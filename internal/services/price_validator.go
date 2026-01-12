package services

import (
	"github.com/rs/zerolog"
)

// YahooClientInterface defines the Yahoo client methods needed by PriceValidator
type YahooClientInterface interface {
	GetCurrentPrice(symbol string) (float64, error)
}

// PriceValidator validates prices using Tradernet as primary, Yahoo as sanity check
// Tradernet is the single source of truth. Yahoo is only used to detect anomalies.
type PriceValidator struct {
	yahooClient YahooClientInterface
	log         zerolog.Logger
}

// NewPriceValidator creates a new PriceValidator
func NewPriceValidator(yahooClient YahooClientInterface, log zerolog.Logger) *PriceValidator {
	return &PriceValidator{
		yahooClient: yahooClient,
		log:         log.With().Str("service", "price_validator").Logger(),
	}
}

// ValidatePrice returns the validated price for a symbol
// Primary: Tradernet price (tradernetPrice)
// Sanity check: If Tradernet > Yahoo * 1.5, use Yahoo instead
// If Yahoo is unavailable, use Tradernet price anyway
func (v *PriceValidator) ValidatePrice(symbol string, yahooSymbol string, tradernetPrice float64) float64 {
	// Tradernet price is 0 - nothing to validate
	if tradernetPrice <= 0 {
		return 0
	}

	// No Yahoo symbol - use Tradernet
	if yahooSymbol == "" {
		v.log.Debug().
			Str("symbol", symbol).
			Float64("tradernet_price", tradernetPrice).
			Msg("No Yahoo symbol configured, using Tradernet price")
		return tradernetPrice
	}

	// Try Yahoo sanity check (best effort, non-blocking)
	yahooPrice, err := v.yahooClient.GetCurrentPrice(yahooSymbol)
	if err != nil {
		v.log.Debug().
			Err(err).
			Str("symbol", symbol).
			Str("yahoo_symbol", yahooSymbol).
			Float64("tradernet_price", tradernetPrice).
			Msg("Yahoo unavailable, using Tradernet price")
		return tradernetPrice
	}

	// Yahoo returned 0 - use Tradernet
	if yahooPrice <= 0 {
		v.log.Debug().
			Str("symbol", symbol).
			Str("yahoo_symbol", yahooSymbol).
			Float64("tradernet_price", tradernetPrice).
			Msg("Yahoo returned 0, using Tradernet price")
		return tradernetPrice
	}

	// Sanity check: Tradernet should not be >50% higher than Yahoo
	// This catches cases like MOH.GR showing 653 EUR when real price is ~30 EUR
	if tradernetPrice > yahooPrice*1.5 {
		v.log.Warn().
			Str("symbol", symbol).
			Str("yahoo_symbol", yahooSymbol).
			Float64("tradernet_price", tradernetPrice).
			Float64("yahoo_price", yahooPrice).
			Float64("ratio", tradernetPrice/yahooPrice).
			Msg("Tradernet price suspicious (>50% higher than Yahoo), using Yahoo price")
		return yahooPrice
	}

	// Tradernet price looks valid
	return tradernetPrice
}
