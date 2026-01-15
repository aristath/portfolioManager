package tradernet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAndInterpolateCandles_OHLCConsistency(t *testing.T) {
	candles := []OHLCV{
		{Timestamp: 1000, Open: 50.0, High: 55.0, Low: 48.0, Close: 52.0, Volume: 1000},
		{Timestamp: 2000, Open: 50.0, High: 45.0, Low: 48.0, Close: 52.0, Volume: 2000}, // Invalid: High < Low
		{Timestamp: 3000, Open: 50.0, High: 55.0, Low: 48.0, Close: 52.0, Volume: 3000},
	}

	result := validateAndInterpolateCandles(candles)

	assert.Len(t, result, 3, "should return same number of candles")
	// Middle candle should be interpolated
	assert.GreaterOrEqual(t, result[1].High, result[1].Low, "interpolated candle should have valid OHLC")
	assert.GreaterOrEqual(t, result[1].High, result[1].Open, "interpolated candle should have valid OHLC")
	assert.GreaterOrEqual(t, result[1].High, result[1].Close, "interpolated candle should have valid OHLC")
	assert.Equal(t, int64(2000), result[1].Volume, "volume should be preserved")
}

func TestValidateAndInterpolateCandles_LargeSpikeDetection(t *testing.T) {
	// Test that a massive spike (29900%) is detected and interpolated
	// when the next candle returns to normal
	candles := []OHLCV{
		{Timestamp: 1000, Open: 50.0, High: 55.0, Low: 48.0, Close: 50.0, Volume: 1000},
		{Timestamp: 2000, Open: 15000.0, High: 15100.0, Low: 14900.0, Close: 15000.0, Volume: 2000}, // 29900% spike
		{Timestamp: 3000, Open: 50.0, High: 55.0, Low: 48.0, Close: 52.0, Volume: 3000},             // Returns to normal
	}

	result := validateAndInterpolateCandles(candles)

	assert.Len(t, result, 3, "should return same number of candles")
	// Middle candle should be interpolated to reasonable value between 50 and 52
	assert.InDelta(t, 51.0, result[1].Close, 1.5, "interpolated close should be between neighbors")
}

func TestValidateAndInterpolateCandles_SpikeDetection(t *testing.T) {
	candles := []OHLCV{
		{Timestamp: 1000, Open: 50.0, High: 55.0, Low: 48.0, Close: 50.0, Volume: 1000},
		{Timestamp: 2000, Open: 600.0, High: 610.0, Low: 590.0, Close: 600.0, Volume: 2000}, // 1100% spike
		{Timestamp: 3000, Open: 52.0, High: 55.0, Low: 48.0, Close: 52.0, Volume: 3000},     // Normal
	}

	result := validateAndInterpolateCandles(candles)

	assert.Len(t, result, 3, "should return same number of candles")
	// Middle candle should be interpolated (spike detected, next is normal)
	assert.InDelta(t, 51.0, result[1].Close, 1.0, "interpolated close should be between 50 and 52")
}

func TestValidateAndInterpolateCandles_ValidPricesPassThrough(t *testing.T) {
	candles := []OHLCV{
		{Timestamp: 1000, Open: 50.0, High: 55.0, Low: 48.0, Close: 52.0, Volume: 1000},
		{Timestamp: 2000, Open: 51.0, High: 56.0, Low: 49.0, Close: 53.0, Volume: 2000},
		{Timestamp: 3000, Open: 52.0, High: 57.0, Low: 50.0, Close: 54.0, Volume: 3000},
	}

	result := validateAndInterpolateCandles(candles)

	assert.Equal(t, candles, result, "valid prices should pass through unchanged")
}

func TestInterpolateOHLCVFull_Linear(t *testing.T) {
	allCandles := []OHLCV{
		{Timestamp: 1000, Open: 47.0, High: 48.0, Low: 46.0, Close: 47.0, Volume: 1000},
		{Timestamp: 2000, Open: 0, High: 0, Low: 0, Close: 0, Volume: 2000}, // To be interpolated (Close=0 is invalid)
		{Timestamp: 3000, Open: 46.7, High: 47.3, Low: 46.6, Close: 46.7, Volume: 3000},
	}

	result := interpolateOHLCVFull(allCandles[1], 1, allCandles)

	// Should interpolate between 47.0 and 46.7
	assert.InDelta(t, 46.85, result.Close, 0.1, "should interpolate close price")
	assert.GreaterOrEqual(t, result.High, result.Close, "high should be >= close")
	assert.LessOrEqual(t, result.Low, result.Close, "low should be <= close")
	assert.Equal(t, int64(2000), result.Volume, "volume should be preserved")
}

func TestInterpolateOHLCVFull_ForwardFill(t *testing.T) {
	allCandles := []OHLCV{
		{Timestamp: 1000, Open: 47.0, High: 48.0, Low: 46.0, Close: 47.0, Volume: 1000},
		{Timestamp: 2000, Open: 0, High: 0, Low: 0, Close: 0, Volume: 2000}, // To be interpolated (Close=0 is invalid)
	}

	result := interpolateOHLCVFull(allCandles[1], 1, allCandles)

	// Should use before price (forward fill)
	assert.Equal(t, 47.0, result.Close, "should forward fill from before")
	assert.Equal(t, 47.0, result.Open, "should forward fill from before")
	assert.Equal(t, int64(2000), result.Volume, "volume should be preserved")
}

func TestInterpolateOHLCVFull_BackwardFill(t *testing.T) {
	allCandles := []OHLCV{
		{Timestamp: 1000, Open: 0, High: 0, Low: 0, Close: 0, Volume: 1000}, // To be interpolated (Close=0 is invalid)
		{Timestamp: 2000, Open: 46.7, High: 47.3, Low: 46.6, Close: 46.7, Volume: 2000},
	}

	result := interpolateOHLCVFull(allCandles[0], 0, allCandles)

	// Should use after price (backward fill)
	assert.Equal(t, 46.7, result.Close, "should backward fill from after")
	assert.Equal(t, 46.7, result.Open, "should backward fill from after")
	assert.Equal(t, int64(1000), result.Volume, "volume should be preserved")
}

// Component-level validation tests

func TestValidateOHLCComponents(t *testing.T) {
	tests := []struct {
		name     string
		candle   OHLCV
		expected OHLCValidation
	}{
		{
			name:   "valid candle passes all checks",
			candle: OHLCV{Open: 100, High: 105, Low: 98, Close: 102},
			expected: OHLCValidation{
				OpenValid: true, HighValid: true, LowValid: true, CloseValid: true,
			},
		},
		{
			name:   "zero close triggers full invalid",
			candle: OHLCV{Open: 100, High: 105, Low: 98, Close: 0},
			expected: OHLCValidation{
				OpenValid: false, HighValid: false, LowValid: false, CloseValid: false,
				Reason: "close_zero_or_negative",
			},
		},
		{
			name:   "negative close triggers full invalid",
			candle: OHLCV{Open: 100, High: 105, Low: 98, Close: -5},
			expected: OHLCValidation{
				OpenValid: false, HighValid: false, LowValid: false, CloseValid: false,
				Reason: "close_zero_or_negative",
			},
		},
		{
			name:   "ITH.EU actual case - extreme high detected",
			candle: OHLCV{Open: 1.856, High: 46137.2, Low: 1.8, Close: 1.922},
			expected: OHLCValidation{
				OpenValid: true, HighValid: false, LowValid: true, CloseValid: true,
				Reason: "high_extreme_relative_to_close",
			},
		},
		{
			name:   "extreme low detected",
			candle: OHLCV{Open: 100, High: 105, Low: 0.001, Close: 102},
			expected: OHLCValidation{
				OpenValid: true, HighValid: true, LowValid: false, CloseValid: true,
				Reason: "low_extreme_relative_to_close",
			},
		},
		{
			name:   "OHLC consistency - high below low",
			candle: OHLCV{Open: 100, High: 90, Low: 95, Close: 102},
			expected: OHLCValidation{
				OpenValid: true, HighValid: false, LowValid: false, CloseValid: true,
				Reason: "ohlc_inconsistent_high_below_low",
			},
		},
		{
			name:   "OHLC consistency - high below close",
			candle: OHLCV{Open: 95, High: 98, Low: 94, Close: 102}, // High >= Open but High < Close
			expected: OHLCValidation{
				OpenValid: true, HighValid: false, LowValid: true, CloseValid: true,
				Reason: "ohlc_inconsistent_high_below_close",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateOHLCComponents(tt.candle, nil, nil)
			assert.Equal(t, tt.expected.OpenValid, result.OpenValid, "OpenValid mismatch")
			assert.Equal(t, tt.expected.HighValid, result.HighValid, "HighValid mismatch")
			assert.Equal(t, tt.expected.LowValid, result.LowValid, "LowValid mismatch")
			assert.Equal(t, tt.expected.CloseValid, result.CloseValid, "CloseValid mismatch")
			if tt.expected.Reason != "" {
				assert.Equal(t, tt.expected.Reason, result.Reason, "Reason mismatch")
			}
		})
	}
}

func TestValidateOHLCComponents_SpikeDetection(t *testing.T) {
	prev := &OHLCV{Close: 50.0}
	next := &OHLCV{Close: 52.0} // Normal next price

	// 1100% spike - should be invalid since next is normal
	candle := OHLCV{Open: 600.0, High: 610.0, Low: 590.0, Close: 600.0}
	result := validateOHLCComponents(candle, prev, next)

	assert.False(t, result.CloseValid, "Close should be invalid for spike")
	assert.Equal(t, "close_spike_or_crash", result.Reason)
}

func TestValidateOHLCComponents_RealMoveNotSpike(t *testing.T) {
	prev := &OHLCV{Close: 50.0}
	next := &OHLCV{Close: 580.0} // Also moved up - real move

	// 1100% spike - but next also moved, so it's a real move
	candle := OHLCV{Open: 600.0, High: 610.0, Low: 590.0, Close: 600.0}
	result := validateOHLCComponents(candle, prev, next)

	// Should be valid since next price also moved (confirming it's real)
	assert.True(t, result.CloseValid, "Close should be valid for real move")
}

func TestInterpolateOHLCVSelective_PreservesValidComponents(t *testing.T) {
	candles := []OHLCV{
		{Timestamp: 1000, Open: 1.8, High: 1.9, Low: 1.75, Close: 1.85, Volume: 1000},
		{Timestamp: 2000, Open: 1.856, High: 46137.2, Low: 1.8, Close: 1.922, Volume: 2000}, // Bad High
		{Timestamp: 3000, Open: 1.9, High: 2.0, Low: 1.85, Close: 1.95, Volume: 3000},
	}

	validation := OHLCValidation{
		OpenValid: true, HighValid: false, LowValid: true, CloseValid: true,
		Reason: "high_extreme_relative_to_close",
	}

	result := interpolateOHLCVSelective(candles[1], 1, candles, validation)

	// Close should be preserved (was valid)
	assert.Equal(t, 1.922, result.Close, "Close should be preserved")
	// Open should be preserved (was valid)
	assert.Equal(t, 1.856, result.Open, "Open should be preserved")
	// Low should be preserved (was valid)
	assert.Equal(t, 1.8, result.Low, "Low should be preserved")
	// High should be interpolated to reasonable value
	assert.Less(t, result.High, 10.0, "High should be interpolated to reasonable value")
	assert.GreaterOrEqual(t, result.High, result.Close, "High should be >= Close")
	// Volume preserved
	assert.Equal(t, int64(2000), result.Volume, "Volume should be preserved")
}

func TestInterpolateOHLCVSelective_FullInterpolationWhenCloseInvalid(t *testing.T) {
	// Test full interpolation when Close is marked invalid (e.g., detected as spike)
	candles := []OHLCV{
		{Timestamp: 1000, Open: 47.0, High: 48.0, Low: 46.0, Close: 47.0, Volume: 1000},
		{Timestamp: 2000, Open: 44000.0, High: 45000.0, Low: 43000.0, Close: 44000.0, Volume: 2000}, // Spike
		{Timestamp: 3000, Open: 46.7, High: 47.3, Low: 46.6, Close: 46.7, Volume: 3000},
	}

	validation := OHLCValidation{
		OpenValid: false, HighValid: false, LowValid: false, CloseValid: false,
		Reason: "close_spike_or_crash",
	}

	result := interpolateOHLCVSelective(candles[1], 1, candles, validation)

	// Should do full interpolation since Close is invalid
	assert.InDelta(t, 46.85, result.Close, 0.2, "Close should be interpolated")
	assert.Less(t, result.Close, 100.0, "Close should be reasonable")
}

func TestValidateAndInterpolateCandlesPreservesGoodData(t *testing.T) {
	// Test that good O/L/C values are preserved when only High is bad
	candles := []OHLCV{
		{Timestamp: 1000, Open: 1.8, High: 1.9, Low: 1.75, Close: 1.85, Volume: 1000},
		{Timestamp: 2000, Open: 1.856, High: 46137.2, Low: 1.8, Close: 1.922, Volume: 2000},
		{Timestamp: 3000, Open: 1.9, High: 2.0, Low: 1.85, Close: 1.95, Volume: 3000},
	}

	result := validateAndInterpolateCandles(candles)

	assert.Len(t, result, 3)
	// Middle candle should have preserved O/L/C
	assert.Equal(t, 1.922, result[1].Close, "Close should be preserved")
	assert.Equal(t, 1.856, result[1].Open, "Open should be preserved")
	assert.Equal(t, 1.8, result[1].Low, "Low should be preserved")
	// High should be fixed
	assert.Less(t, result[1].High, 100.0, "High should be fixed")
	assert.GreaterOrEqual(t, result[1].High, result[1].Close, "High should be >= Close")
}

func TestOHLCValidation_Methods(t *testing.T) {
	// Test AllValid
	v1 := OHLCValidation{OpenValid: true, HighValid: true, LowValid: true, CloseValid: true}
	assert.True(t, v1.AllValid())

	v2 := OHLCValidation{OpenValid: true, HighValid: false, LowValid: true, CloseValid: true}
	assert.False(t, v2.AllValid())

	// Test NeedsFullInterpolation
	assert.False(t, v2.NeedsFullInterpolation(), "Should not need full interpolation when Close is valid")

	v3 := OHLCValidation{OpenValid: false, HighValid: false, LowValid: false, CloseValid: false}
	assert.True(t, v3.NeedsFullInterpolation(), "Should need full interpolation when Close is invalid")

	// Test NeedsInterpolation
	assert.False(t, v1.NeedsInterpolation(), "Should not need interpolation when all valid")
	assert.True(t, v2.NeedsInterpolation(), "Should need interpolation when High is invalid")
}
