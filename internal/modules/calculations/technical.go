package calculations

import (
	"time"

	"github.com/aristath/sentinel/pkg/formulas"
	"github.com/rs/zerolog"
)

// TTL constants for cache expiration
const (
	TTLTechnical        = 24 * time.Hour     // Per-security metrics (EMA, RSI, Sharpe)
	TTLOptimizer        = 24 * time.Hour     // Covariance, HRP, returns
	TTLRegimeCovariance = 4 * time.Hour      // Regime-aware covariance (shorter TTL due to regime sensitivity)
	TTLConstraints      = 7 * 24 * time.Hour // Allocation targets rarely change

	// MinPricesRequired is the minimum number of prices needed to calculate any metric
	MinPricesRequired = 50
)

// PriceProvider fetches historical prices for a security
type PriceProvider interface {
	GetDailyPrices(isin string, days int) ([]float64, error)
}

// TechnicalCalculator computes and caches technical indicators for securities.
// It calculates all metrics at once for a given ISIN and stores them in the cache.
type TechnicalCalculator struct {
	cache         *Cache
	priceProvider PriceProvider
	log           zerolog.Logger
}

// NewTechnicalCalculator creates a new technical calculator
func NewTechnicalCalculator(cache *Cache, priceProvider PriceProvider, log zerolog.Logger) *TechnicalCalculator {
	return &TechnicalCalculator{
		cache:         cache,
		priceProvider: priceProvider,
		log:           log.With().Str("component", "technical_calculator").Logger(),
	}
}

// CalculateForISIN computes and caches all technical metrics for an ISIN.
// This includes:
// - EMA-50 and EMA-200
// - RSI-14
// - Sharpe Ratio
// - Max Drawdown
// - 52-week High/Low
//
// Metrics are only calculated if sufficient price data is available.
// Insufficient data is not an error - the metric is simply not cached.
func (tc *TechnicalCalculator) CalculateForISIN(isin string) error {
	prices, err := tc.priceProvider.GetDailyPrices(isin, 365)
	if err != nil {
		return err
	}

	if len(prices) < MinPricesRequired {
		tc.log.Debug().
			Str("isin", isin).
			Int("prices", len(prices)).
			Int("required", MinPricesRequired).
			Msg("Insufficient price data for calculation")
		return nil
	}

	// EMA-50
	if len(prices) >= 50 {
		if ema := formulas.CalculateEMA(prices, 50); ema != nil {
			if err := tc.cache.SetTechnical(isin, "ema", 50, *ema, TTLTechnical); err != nil {
				tc.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache EMA-50")
			}
		}
	}

	// EMA-200
	if len(prices) >= 200 {
		if ema := formulas.CalculateEMA(prices, 200); ema != nil {
			if err := tc.cache.SetTechnical(isin, "ema", 200, *ema, TTLTechnical); err != nil {
				tc.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache EMA-200")
			}
		}
	}

	// RSI-14
	if rsi := formulas.CalculateRSI(prices, 14); rsi != nil {
		if err := tc.cache.SetTechnical(isin, "rsi", 14, *rsi, TTLTechnical); err != nil {
			tc.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache RSI-14")
		}
	}

	// Sharpe Ratio (risk-free rate 2%)
	if sharpe := formulas.CalculateSharpeFromPrices(prices, 0.02); sharpe != nil {
		if err := tc.cache.SetTechnical(isin, "sharpe", 0, *sharpe, TTLTechnical); err != nil {
			tc.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache Sharpe ratio")
		}
	}

	// Max Drawdown
	if mdd := formulas.CalculateMaxDrawdown(prices); mdd != nil {
		if err := tc.cache.SetTechnical(isin, "max_drawdown", 0, *mdd, TTLTechnical); err != nil {
			tc.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache max drawdown")
		}
	}

	// 52-week High/Low (requires 252 trading days)
	if len(prices) >= 252 {
		if high := formulas.Calculate52WeekHigh(prices); high != nil {
			if err := tc.cache.SetTechnical(isin, "52w_high", 0, *high, TTLTechnical); err != nil {
				tc.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache 52-week high")
			}
		}
		if low := formulas.Calculate52WeekLow(prices); low != nil {
			if err := tc.cache.SetTechnical(isin, "52w_low", 0, *low, TTLTechnical); err != nil {
				tc.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache 52-week low")
			}
		}
	}

	tc.log.Debug().Str("isin", isin).Int("prices", len(prices)).Msg("Calculated technical metrics")
	return nil
}
