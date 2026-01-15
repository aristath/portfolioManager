package universe

import (
	"fmt"
	"math"
	"time"

	"github.com/rs/zerolog"
)

const (
	// Validation thresholds
	// Note: No absolute price bounds - only relative checks. This system must work for 15+ years
	// and absolute bounds like "Close > 10,000" are arbitrary and will break over time.
	maxPriceMultiplier    = 10.0   // Price > 10x average is abnormal
	minPriceMultiplier    = 0.1    // Price < 0.1x average is abnormal
	maxPriceChangePercent = 1000.0 // >1000% change is a spike
	minPriceChangePercent = -90.0  // <-90% change is a crash
	contextWindowDays     = 30     // Use last 30 days for context

	// Component-level validation thresholds (lenient: 100x/0.01x)
	highCloseMaxRatio = 100.0 // High ≤ Close × 100
	lowCloseMinRatio  = 0.01  // Low ≥ Close × 0.01
)

// OHLCValidation tracks validity of each OHLC component
type OHLCValidation struct {
	OpenValid  bool
	HighValid  bool
	LowValid   bool
	CloseValid bool
	Reason     string
}

// NeedsFullInterpolation returns true when Close is invalid (Close is the anchor)
func (v OHLCValidation) NeedsFullInterpolation() bool {
	return !v.CloseValid
}

// NeedsInterpolation returns true when any component is invalid
func (v OHLCValidation) NeedsInterpolation() bool {
	return !v.OpenValid || !v.HighValid || !v.LowValid || !v.CloseValid
}

// AllValid returns true when all components are valid
func (v OHLCValidation) AllValid() bool {
	return v.OpenValid && v.HighValid && v.LowValid && v.CloseValid
}

// InterpolationLog records when a price was interpolated
type InterpolationLog struct {
	Date              string
	OriginalClose     float64
	InterpolatedClose float64
	OriginalHigh      float64
	InterpolatedHigh  float64
	OriginalLow       float64
	InterpolatedLow   float64
	Method            string // "linear", "forward_fill", "backward_fill", "selective"
	Reason            string
}

// PriceValidator validates and interpolates abnormal prices
type PriceValidator struct {
	log zerolog.Logger
}

// NewPriceValidator creates a new price validator
func NewPriceValidator(log zerolog.Logger) *PriceValidator {
	return &PriceValidator{
		log: log.With().Str("component", "price_validator").Logger(),
	}
}

// ValidatePrice validates each OHLC component independently
// Returns OHLCValidation with per-component validity status
func (v *PriceValidator) ValidatePrice(price DailyPrice, previousPrice *DailyPrice, context []DailyPrice) OHLCValidation {
	result := OHLCValidation{
		OpenValid: true, HighValid: true, LowValid: true, CloseValid: true,
	}

	// Zero/negative Close indicates missing or invalid data
	if price.Close <= 0 {
		result.CloseValid = false
		result.OpenValid = false
		result.HighValid = false
		result.LowValid = false
		result.Reason = "close_zero_or_negative"
		return result
	}

	// 1. OHLC consistency checks
	if price.High < price.Low {
		result.HighValid = false
		result.LowValid = false
		result.Reason = "high_below_low"
	}
	if price.High < price.Open {
		result.HighValid = false
		if result.Reason == "" {
			result.Reason = "high_below_open"
		}
	}
	if price.High < price.Close {
		result.HighValid = false
		if result.Reason == "" {
			result.Reason = "high_below_close"
		}
	}
	if price.Low > price.Open {
		result.LowValid = false
		if result.Reason == "" {
			result.Reason = "low_above_open"
		}
	}
	if price.Low > price.Close {
		result.LowValid = false
		if result.Reason == "" {
			result.Reason = "low_above_close"
		}
	}

	// 2. Extreme High/Low relative to Close (100x/0.01x thresholds)
	if price.Close > 0 && price.High > price.Close*highCloseMaxRatio {
		result.HighValid = false
		if result.Reason == "" {
			result.Reason = "high_extreme_relative_to_close"
		}
	}
	if price.Close > 0 && price.Low > 0 && price.Low < price.Close*lowCloseMinRatio {
		result.LowValid = false
		if result.Reason == "" {
			result.Reason = "low_extreme_relative_to_close"
		}
	}

	// 3. Day-over-day change detection (spike/crash affects all components)
	if previousPrice != nil && previousPrice.Close > 0 {
		changePercent := ((price.Close - previousPrice.Close) / previousPrice.Close) * 100.0
		if changePercent > maxPriceChangePercent {
			result.CloseValid = false
			result.OpenValid = false
			result.HighValid = false
			result.LowValid = false
			result.Reason = "spike_detected"
		}
		if changePercent < minPriceChangePercent {
			result.CloseValid = false
			result.OpenValid = false
			result.HighValid = false
			result.LowValid = false
			result.Reason = "crash_detected"
		}
	}

	// 4. Average-based validation (requires context, affects Close which anchors all)
	if len(context) > 0 && result.CloseValid {
		contextSize := len(context)
		if contextSize > contextWindowDays {
			contextSize = contextWindowDays
		}
		var sum float64
		for i := 0; i < contextSize; i++ {
			sum += context[i].Close
		}
		avgPrice := sum / float64(contextSize)

		if price.Close > avgPrice*maxPriceMultiplier {
			result.CloseValid = false
			result.OpenValid = false
			result.HighValid = false
			result.LowValid = false
			result.Reason = "price_too_high"
		}
		if price.Close < avgPrice*minPriceMultiplier {
			result.CloseValid = false
			result.OpenValid = false
			result.HighValid = false
			result.LowValid = false
			result.Reason = "price_too_low"
		}
	}

	// Note: No absolute bounds fallback. When there's no context, we rely on:
	// - OHLC consistency checks
	// - High/Low relative to Close (100x/0.01x thresholds)
	// - Day-over-day change detection (if previous price available)
	// This approach is more robust for a 15+ year system where absolute bounds become meaningless.

	return result
}

// InterpolatePrice interpolates invalid components of a price
// Only replaces components that are invalid, preserves valid ones
func (v *PriceValidator) InterpolatePrice(price DailyPrice, validation OHLCValidation, before, after []DailyPrice) (DailyPrice, string, error) {
	interpolated := price // Start with original - preserves date, volume, and valid components

	// If Close is invalid, we need full interpolation (Close is the anchor)
	if !validation.CloseValid {
		return v.interpolateFull(price, before, after)
	}

	// Selective interpolation - only fix invalid components
	method := "selective"

	if !validation.HighValid {
		ratio := v.getTypicalHighCloseRatio(before, after)
		interpolated.High = interpolated.Close * ratio
	}

	if !validation.LowValid {
		ratio := v.getTypicalLowCloseRatio(before, after)
		interpolated.Low = interpolated.Close * ratio
	}

	if !validation.OpenValid {
		if len(before) > 0 {
			interpolated.Open = before[0].Close
		} else if len(after) > 0 {
			interpolated.Open = after[0].Open
		} else {
			interpolated.Open = interpolated.Close
		}
	}

	ensureOHLCConsistency(&interpolated)
	return interpolated, method, nil
}

// interpolateFull performs full interpolation when Close is invalid
func (v *PriceValidator) interpolateFull(price DailyPrice, before, after []DailyPrice) (DailyPrice, string, error) {
	interpolated := price

	priceDate, err := time.Parse("2006-01-02", price.Date)
	if err != nil {
		return interpolated, "", fmt.Errorf("failed to parse price date: %w", err)
	}

	// Linear interpolation if both before and after available
	if len(before) > 0 && len(after) > 0 {
		beforePrice := before[0]
		afterPrice := after[0]

		beforeDate, err1 := time.Parse("2006-01-02", beforePrice.Date)
		afterDate, err2 := time.Parse("2006-01-02", afterPrice.Date)
		if err1 != nil || err2 != nil {
			return interpolated, "", fmt.Errorf("failed to parse before/after dates")
		}

		daysBetween := priceDate.Sub(beforeDate).Hours() / 24.0
		totalDays := afterDate.Sub(beforeDate).Hours() / 24.0

		if totalDays > 0 {
			interpolated.Close = beforePrice.Close + (afterPrice.Close-beforePrice.Close)*(daysBetween/totalDays)

			// Use ratios for other components
			openRatio := (beforePrice.Open/beforePrice.Close + afterPrice.Open/afterPrice.Close) / 2.0
			highRatio := (beforePrice.High/beforePrice.Close + afterPrice.High/afterPrice.Close) / 2.0
			lowRatio := (beforePrice.Low/beforePrice.Close + afterPrice.Low/afterPrice.Close) / 2.0

			interpolated.Open = interpolated.Close * openRatio
			interpolated.High = interpolated.Close * highRatio
			interpolated.Low = interpolated.Close * lowRatio

			ensureOHLCConsistency(&interpolated)
			return interpolated, "linear", nil
		}
	}

	// Forward fill
	if len(before) > 0 {
		interpolated.Close = before[0].Close
		interpolated.Open = before[0].Open
		interpolated.High = before[0].High
		interpolated.Low = before[0].Low
		return interpolated, "forward_fill", nil
	}

	// Backward fill
	if len(after) > 0 {
		interpolated.Close = after[0].Close
		interpolated.Open = after[0].Open
		interpolated.High = after[0].High
		interpolated.Low = after[0].Low
		return interpolated, "backward_fill", nil
	}

	ensureOHLCConsistency(&interpolated)
	return interpolated, "no_interpolation", nil
}

func (v *PriceValidator) getTypicalHighCloseRatio(before, after []DailyPrice) float64 {
	var sum float64
	var count int

	if len(before) > 0 && before[0].Close > 0 {
		sum += before[0].High / before[0].Close
		count++
	}
	if len(after) > 0 && after[0].Close > 0 {
		sum += after[0].High / after[0].Close
		count++
	}

	if count == 0 {
		return 1.02 // Default: High is 2% above Close
	}
	return sum / float64(count)
}

func (v *PriceValidator) getTypicalLowCloseRatio(before, after []DailyPrice) float64 {
	var sum float64
	var count int

	if len(before) > 0 && before[0].Close > 0 {
		sum += before[0].Low / before[0].Close
		count++
	}
	if len(after) > 0 && after[0].Close > 0 {
		sum += after[0].Low / after[0].Close
		count++
	}

	if count == 0 {
		return 0.98 // Default: Low is 2% below Close
	}
	return sum / float64(count)
}

// ValidateAndInterpolate validates all prices and interpolates abnormal ones
// Uses component-level validation to preserve valid data while fixing invalid components
func (v *PriceValidator) ValidateAndInterpolate(prices []DailyPrice, context []DailyPrice) ([]DailyPrice, []InterpolationLog, error) {
	if len(prices) == 0 {
		return prices, []InterpolationLog{}, nil
	}

	result := make([]DailyPrice, 0, len(prices))
	logs := []InterpolationLog{}

	for i, price := range prices {
		// Get previous price from result (already validated prices)
		var previousPrice *DailyPrice
		if len(result) > 0 {
			priceDate, err1 := time.Parse("2006-01-02", price.Date)
			for j := len(result) - 1; j >= 0; j-- {
				resultDate, err2 := time.Parse("2006-01-02", result[j].Date)
				if err1 == nil && err2 == nil && resultDate.Before(priceDate) {
					previousPrice = &result[j]
					break
				}
			}
		}
		// Fallback to previous valid price in original array
		if previousPrice == nil && i > 0 {
			for j := i - 1; j >= 0; j-- {
				prevValidation := v.ValidatePrice(prices[j], nil, context)
				if prevValidation.AllValid() {
					previousPrice = &prices[j]
					break
				}
			}
		}

		validation := v.ValidatePrice(price, previousPrice, context)

		if validation.AllValid() {
			result = append(result, price)
			continue
		}

		// Find before/after prices for interpolation
		before, after := v.findInterpolationContext(i, prices, result, context, price.Date)

		interpolated, method, err := v.InterpolatePrice(price, validation, before, after)
		if err != nil {
			v.log.Error().Err(err).Str("date", price.Date).Msg("Interpolation failed, using original")
			result = append(result, price)
			continue
		}

		logs = append(logs, InterpolationLog{
			Date:              price.Date,
			OriginalClose:     price.Close,
			InterpolatedClose: interpolated.Close,
			OriginalHigh:      price.High,
			InterpolatedHigh:  interpolated.High,
			OriginalLow:       price.Low,
			InterpolatedLow:   interpolated.Low,
			Method:            method,
			Reason:            validation.Reason,
		})

		v.log.Debug().
			Str("date", price.Date).
			Float64("original_close", price.Close).
			Float64("interpolated_close", interpolated.Close).
			Str("method", method).
			Str("reason", validation.Reason).
			Msg("Interpolated abnormal price")

		result = append(result, interpolated)
	}

	// Cascade detection warning
	if len(prices) > 0 {
		invalidRatio := float64(len(logs)) / float64(len(prices))
		if invalidRatio > 0.5 {
			v.log.Warn().
				Int("invalid_count", len(logs)).
				Int("total_count", len(prices)).
				Float64("invalid_ratio", invalidRatio).
				Msg("More than 50% of prices flagged invalid - possible data quality issue")
		}
	}

	return result, logs, nil
}

// findInterpolationContext finds valid before/after prices for interpolation
func (v *PriceValidator) findInterpolationContext(index int, prices, result []DailyPrice, context []DailyPrice, currentDate string) ([]DailyPrice, []DailyPrice) {
	var before, after []DailyPrice
	priceDate, _ := time.Parse("2006-01-02", currentDate)

	// Look for "before" price
	// 1. Check already validated results
	for j := len(result) - 1; j >= 0; j-- {
		resultDate, _ := time.Parse("2006-01-02", result[j].Date)
		if resultDate.Before(priceDate) {
			before = []DailyPrice{result[j]}
			break
		}
	}
	// 2. Check previous prices in array
	if len(before) == 0 {
		for j := index - 1; j >= 0; j-- {
			prevValidation := v.ValidatePrice(prices[j], nil, context)
			if prevValidation.AllValid() {
				before = []DailyPrice{prices[j]}
				break
			}
		}
	}
	// 3. Check context (most recent before current date)
	if len(before) == 0 && len(context) > 0 {
		for _, ctxPrice := range context {
			ctxDate, _ := time.Parse("2006-01-02", ctxPrice.Date)
			if ctxDate.Before(priceDate) {
				before = []DailyPrice{ctxPrice}
				break
			}
		}
	}

	// Look for "after" price
	// 1. Check subsequent prices in array
	for j := index + 1; j < len(prices); j++ {
		nextValidation := v.ValidatePrice(prices[j], nil, context)
		if nextValidation.AllValid() {
			after = []DailyPrice{prices[j]}
			break
		}
	}
	// 2. Check context (earliest after current date)
	if len(after) == 0 && len(context) > 0 {
		var earliestAfter *DailyPrice
		for i := range context {
			ctxDate, _ := time.Parse("2006-01-02", context[i].Date)
			if ctxDate.After(priceDate) {
				if earliestAfter == nil {
					earliestAfter = &context[i]
				} else {
					earliestDate, _ := time.Parse("2006-01-02", earliestAfter.Date)
					if ctxDate.Before(earliestDate) {
						earliestAfter = &context[i]
					}
				}
			}
		}
		if earliestAfter != nil {
			after = []DailyPrice{*earliestAfter}
		}
	}

	return before, after
}

// Helper function to ensure OHLC consistency
func ensureOHLCConsistency(price *DailyPrice) {
	// Ensure High >= all
	price.High = math.Max(price.High, price.Open)
	price.High = math.Max(price.High, price.Close)

	// Ensure Low <= all
	price.Low = math.Min(price.Low, price.Open)
	price.Low = math.Min(price.Low, price.Close)

	// Ensure High >= Low
	if price.High < price.Low {
		price.High = price.Low
	}
}
