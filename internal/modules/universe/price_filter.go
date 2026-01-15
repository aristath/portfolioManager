package universe

import (
	"github.com/rs/zerolog"
)

const (
	// Filter thresholds (same as previous validator)
	filterMaxPriceMultiplier    = 10.0   // Price > 10x average is abnormal
	filterMinPriceMultiplier    = 0.1    // Price < 0.1x average is abnormal
	filterMaxPriceChangePercent = 1000.0 // >1000% change is a spike
	filterMinPriceChangePercent = -90.0  // <-90% change is a crash
	filterContextWindowDays     = 30     // Use last 30 days for context

	// Component-level thresholds
	filterHighCloseMaxRatio = 100.0 // High ≤ Close × 100
	filterLowCloseMinRatio  = 0.01  // Low ≥ Close × 0.01
)

// PriceFilter filters out anomalous price data on read
// Unlike PriceValidator, it simply excludes invalid prices rather than interpolating
type PriceFilter struct {
	log zerolog.Logger
}

// NewPriceFilter creates a new price filter
func NewPriceFilter(log zerolog.Logger) *PriceFilter {
	return &PriceFilter{
		log: log.With().Str("component", "price_filter").Logger(),
	}
}

// Filter returns only valid prices, excluding anomalies
// No interpolation - simply exclude bad data points
// Input prices are expected in chronological order (oldest first)
func (f *PriceFilter) Filter(prices []DailyPrice) []DailyPrice {
	if len(prices) == 0 {
		return []DailyPrice{}
	}

	result := make([]DailyPrice, 0, len(prices))

	for i, price := range prices {
		// Build context from already-validated prices
		var context []DailyPrice
		if len(result) > 0 {
			contextSize := len(result)
			if contextSize > filterContextWindowDays {
				contextSize = filterContextWindowDays
			}
			// Context is newest-first for average calculation
			context = make([]DailyPrice, contextSize)
			for j := 0; j < contextSize; j++ {
				context[j] = result[len(result)-1-j]
			}
		}

		// Get previous valid price for day-over-day checks
		var previousPrice *DailyPrice
		if len(result) > 0 {
			previousPrice = &result[len(result)-1]
		}

		if f.isValid(price, previousPrice, context) {
			result = append(result, price)
		} else {
			f.log.Debug().
				Str("date", price.Date).
				Float64("close", price.Close).
				Int("index", i).
				Msg("Filtered out invalid price")
		}
	}

	return result
}

// isValid checks if a single price is valid given context
func (f *PriceFilter) isValid(price DailyPrice, previousPrice *DailyPrice, context []DailyPrice) bool {
	// 1. Zero/negative Close indicates missing or invalid data
	if price.Close <= 0 {
		return false
	}

	// 2. OHLC consistency checks
	if price.High < price.Low {
		return false
	}
	if price.High < price.Open {
		return false
	}
	if price.High < price.Close {
		return false
	}
	if price.Low > price.Open {
		return false
	}
	if price.Low > price.Close {
		return false
	}

	// 3. Extreme High/Low relative to Close
	if price.Close > 0 && price.High > price.Close*filterHighCloseMaxRatio {
		return false
	}
	if price.Close > 0 && price.Low > 0 && price.Low < price.Close*filterLowCloseMinRatio {
		return false
	}

	// 4. Day-over-day change detection (spike/crash)
	if previousPrice != nil && previousPrice.Close > 0 {
		changePercent := ((price.Close - previousPrice.Close) / previousPrice.Close) * 100.0
		if changePercent > filterMaxPriceChangePercent {
			return false
		}
		if changePercent < filterMinPriceChangePercent {
			return false
		}
	}

	// 5. Context-based validation (requires sufficient context)
	if len(context) >= filterContextWindowDays {
		var sum float64
		for i := 0; i < filterContextWindowDays; i++ {
			sum += context[i].Close
		}
		avgPrice := sum / float64(filterContextWindowDays)

		if price.Close > avgPrice*filterMaxPriceMultiplier {
			return false
		}
		if price.Close < avgPrice*filterMinPriceMultiplier {
			return false
		}
	}

	return true
}
