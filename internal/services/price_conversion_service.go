// Package services provides core business services shared across multiple modules.
//
// This file contains PriceConversionService which handles currency conversion for price data.
package services

import (
	"github.com/aristath/sentinel/internal/domain"
	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/rs/zerolog"
)

/**
 * PriceConversionService handles currency conversion for price data.
 *
 * Converts prices from securities' native currencies to EUR for portfolio calculations.
 * This ensures all prices are in a consistent currency (EUR) for comparison and aggregation.
 */
type PriceConversionService struct {
	exchangeService domain.CurrencyExchangeServiceInterface // Currency exchange service
	log             zerolog.Logger                          // Structured logger
}

/**
 * NewPriceConversionService creates a new price conversion service.
 *
 * @param exchangeService - Currency exchange service for rate fetching
 * @param log - Structured logger
 * @returns *PriceConversionService - New price conversion service instance
 */
func NewPriceConversionService(
	exchangeService domain.CurrencyExchangeServiceInterface,
	log zerolog.Logger,
) *PriceConversionService {
	return &PriceConversionService{
		exchangeService: exchangeService,
		log:             log.With().Str("service", "price_conversion").Logger(),
	}
}

/**
 * ConvertPricesToEUR converts a map of prices from native currencies to EUR.
 *
 * This function:
 * 1. Builds a currency lookup map from securities
 * 2. For each price, determines the security's native currency
 * 3. Converts non-EUR prices to EUR using exchange rates
 * 4. Returns all prices in EUR for consistent portfolio calculations
 *
 * Prices already in EUR are passed through unchanged. If exchange service is unavailable
 * or conversion fails, the native price is used (with warning logged).
 *
 * Parameters:
 * - prices: map[symbol]price in native currency
 * - securities: securities with currency information
 *
 * Returns: map[symbol]price in EUR
 *
 * @param prices - Map of symbol to price in native currency
 * @param securities - List of securities with currency information
 * @returns map[string]float64 - Map of symbol to price in EUR
 */
func (s *PriceConversionService) ConvertPricesToEUR(
	prices map[string]float64,
	securities []universe.Security,
) map[string]float64 {
	if len(prices) == 0 {
		return prices
	}

	// Build currency lookup map from securities
	currencyMap := make(map[string]string)
	for _, security := range securities {
		currency := security.Currency
		if currency == "" {
			currency = "EUR" // Default to EUR if currency not specified
		}
		currencyMap[security.Symbol] = currency
	}

	convertedPrices := make(map[string]float64)
	convertedCount := 0
	skippedCount := 0

	for symbol, nativePrice := range prices {
		currency, hasCurrency := currencyMap[symbol]
		if !hasCurrency {
			// No currency info, assume EUR (pass through)
			convertedPrices[symbol] = nativePrice
			skippedCount++
			continue
		}

		if currency == "EUR" || currency == "" {
			// Already in EUR, pass through unchanged
			convertedPrices[symbol] = nativePrice
			continue
		}

		// Convert to EUR using exchange service
		if s.exchangeService == nil {
			s.log.Warn().
				Str("symbol", symbol).
				Str("currency", currency).
				Float64("native_price", nativePrice).
				Msg("Exchange service not available, cannot convert price - using native price")
			convertedPrices[symbol] = nativePrice
			skippedCount++
			continue
		}

		rate, err := s.exchangeService.GetRate(currency, "EUR")
		if err != nil || rate <= 0 {
			s.log.Warn().
				Err(err).
				Str("symbol", symbol).
				Str("currency", currency).
				Float64("native_price", nativePrice).
				Msg("Failed to get exchange rate - using native price")
			convertedPrices[symbol] = nativePrice
			skippedCount++
			continue
		}

		// Convert: native_price × rate = priceEUR
		priceEUR := nativePrice * rate
		convertedPrices[symbol] = priceEUR
		convertedCount++

		s.log.Debug().
			Str("symbol", symbol).
			Str("currency", currency).
			Float64("native_price", nativePrice).
			Float64("rate", rate).
			Float64("price_eur", priceEUR).
			Msg("Converted price to EUR")
	}

	s.log.Info().
		Int("total", len(prices)).
		Int("converted", convertedCount).
		Int("skipped_conversion", skippedCount).
		Msg("Converted prices to EUR")

	return convertedPrices
}
