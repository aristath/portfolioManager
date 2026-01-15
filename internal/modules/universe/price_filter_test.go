package universe

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPriceFilter(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)
	require.NotNil(t, filter)
}

func TestPriceFilter_Filter_EmptyInput(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	result := filter.Filter([]DailyPrice{})
	assert.Empty(t, result)

	result = filter.Filter(nil)
	assert.Empty(t, result)
}

func TestPriceFilter_Filter_ValidOHLCPassesThrough(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	prices := []DailyPrice{
		{Date: "2025-01-15", Open: 50.0, High: 55.0, Low: 48.0, Close: 52.0},
		{Date: "2025-01-16", Open: 52.0, High: 54.0, Low: 51.0, Close: 53.0},
		{Date: "2025-01-17", Open: 53.0, High: 56.0, Low: 52.0, Close: 55.0},
	}

	result := filter.Filter(prices)
	require.Len(t, result, 3)
	assert.Equal(t, 52.0, result[0].Close)
	assert.Equal(t, 53.0, result[1].Close)
	assert.Equal(t, 55.0, result[2].Close)
}

func TestPriceFilter_Filter_ZeroCloseExcluded(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	prices := []DailyPrice{
		{Date: "2025-01-15", Open: 50.0, High: 55.0, Low: 48.0, Close: 52.0},
		{Date: "2025-01-16", Open: 52.0, High: 54.0, Low: 51.0, Close: 0}, // Invalid
		{Date: "2025-01-17", Open: 53.0, High: 56.0, Low: 52.0, Close: 55.0},
	}

	result := filter.Filter(prices)
	require.Len(t, result, 2)
	assert.Equal(t, 52.0, result[0].Close)
	assert.Equal(t, 55.0, result[1].Close)
}

func TestPriceFilter_Filter_NegativeCloseExcluded(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	prices := []DailyPrice{
		{Date: "2025-01-15", Open: 50.0, High: 55.0, Low: 48.0, Close: 52.0},
		{Date: "2025-01-16", Open: 52.0, High: 54.0, Low: 51.0, Close: -5.0}, // Invalid
		{Date: "2025-01-17", Open: 53.0, High: 56.0, Low: 52.0, Close: 55.0},
	}

	result := filter.Filter(prices)
	require.Len(t, result, 2)
	assert.Equal(t, 52.0, result[0].Close)
	assert.Equal(t, 55.0, result[1].Close)
}

func TestPriceFilter_Filter_HighBelowLowExcluded(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	prices := []DailyPrice{
		{Date: "2025-01-15", Open: 50.0, High: 55.0, Low: 48.0, Close: 52.0},
		{Date: "2025-01-16", Open: 52.0, High: 45.0, Low: 51.0, Close: 53.0}, // Invalid: High < Low
		{Date: "2025-01-17", Open: 53.0, High: 56.0, Low: 52.0, Close: 55.0},
	}

	result := filter.Filter(prices)
	require.Len(t, result, 2)
	assert.Equal(t, 52.0, result[0].Close)
	assert.Equal(t, 55.0, result[1].Close)
}

func TestPriceFilter_Filter_HighBelowCloseExcluded(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	prices := []DailyPrice{
		{Date: "2025-01-15", Open: 50.0, High: 55.0, Low: 48.0, Close: 52.0},
		{Date: "2025-01-16", Open: 52.0, High: 50.0, Low: 48.0, Close: 55.0}, // Invalid: High < Close
		{Date: "2025-01-17", Open: 53.0, High: 56.0, Low: 52.0, Close: 55.0},
	}

	result := filter.Filter(prices)
	require.Len(t, result, 2)
	assert.Equal(t, 52.0, result[0].Close)
	assert.Equal(t, 55.0, result[1].Close)
}

func TestPriceFilter_Filter_HighBelowOpenExcluded(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	prices := []DailyPrice{
		{Date: "2025-01-15", Open: 50.0, High: 55.0, Low: 48.0, Close: 52.0},
		{Date: "2025-01-16", Open: 55.0, High: 50.0, Low: 48.0, Close: 49.0}, // Invalid: High < Open
		{Date: "2025-01-17", Open: 53.0, High: 56.0, Low: 52.0, Close: 55.0},
	}

	result := filter.Filter(prices)
	require.Len(t, result, 2)
	assert.Equal(t, 52.0, result[0].Close)
	assert.Equal(t, 55.0, result[1].Close)
}

func TestPriceFilter_Filter_LowAboveCloseExcluded(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	prices := []DailyPrice{
		{Date: "2025-01-15", Open: 50.0, High: 55.0, Low: 48.0, Close: 52.0},
		{Date: "2025-01-16", Open: 52.0, High: 56.0, Low: 55.0, Close: 53.0}, // Invalid: Low > Close
		{Date: "2025-01-17", Open: 53.0, High: 56.0, Low: 52.0, Close: 55.0},
	}

	result := filter.Filter(prices)
	require.Len(t, result, 2)
	assert.Equal(t, 52.0, result[0].Close)
	assert.Equal(t, 55.0, result[1].Close)
}

func TestPriceFilter_Filter_LowAboveOpenExcluded(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	prices := []DailyPrice{
		{Date: "2025-01-15", Open: 50.0, High: 55.0, Low: 48.0, Close: 52.0},
		{Date: "2025-01-16", Open: 45.0, High: 56.0, Low: 50.0, Close: 53.0}, // Invalid: Low > Open
		{Date: "2025-01-17", Open: 53.0, High: 56.0, Low: 52.0, Close: 55.0},
	}

	result := filter.Filter(prices)
	require.Len(t, result, 2)
	assert.Equal(t, 52.0, result[0].Close)
	assert.Equal(t, 55.0, result[1].Close)
}

func TestPriceFilter_Filter_ExtremeHighExcluded(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	// ITH.EU case: High=46137.2 while Close=1.922 (24000x ratio)
	prices := []DailyPrice{
		{Date: "2025-01-15", Open: 1.8, High: 1.9, Low: 1.75, Close: 1.85},
		{Date: "2025-01-16", Open: 1.856, High: 46137.2, Low: 1.8, Close: 1.922}, // Invalid: High > Close * 100
		{Date: "2025-01-17", Open: 1.9, High: 2.0, Low: 1.85, Close: 1.95},
	}

	result := filter.Filter(prices)
	require.Len(t, result, 2)
	assert.Equal(t, 1.85, result[0].Close)
	assert.Equal(t, 1.95, result[1].Close)
}

func TestPriceFilter_Filter_ExtremeLowExcluded(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	prices := []DailyPrice{
		{Date: "2025-01-15", Open: 100.0, High: 105.0, Low: 98.0, Close: 102.0},
		{Date: "2025-01-16", Open: 100.0, High: 105.0, Low: 0.001, Close: 102.0}, // Invalid: Low < Close * 0.01
		{Date: "2025-01-17", Open: 100.0, High: 105.0, Low: 98.0, Close: 103.0},
	}

	result := filter.Filter(prices)
	require.Len(t, result, 2)
	assert.Equal(t, 102.0, result[0].Close)
	assert.Equal(t, 103.0, result[1].Close)
}

func TestPriceFilter_Filter_SpikeDetected(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	// >1000% change from previous day should be excluded
	prices := []DailyPrice{
		{Date: "2025-01-15", Open: 50.0, High: 55.0, Low: 48.0, Close: 50.0},
		{Date: "2025-01-16", Open: 600.0, High: 610.0, Low: 590.0, Close: 600.0}, // Invalid: 1100% increase
		{Date: "2025-01-17", Open: 51.0, High: 53.0, Low: 50.0, Close: 52.0},
	}

	result := filter.Filter(prices)
	require.Len(t, result, 2)
	assert.Equal(t, 50.0, result[0].Close)
	assert.Equal(t, 52.0, result[1].Close)
}

func TestPriceFilter_Filter_CrashDetected(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	// <-90% change from previous day should be excluded
	prices := []DailyPrice{
		{Date: "2025-01-15", Open: 100.0, High: 105.0, Low: 98.0, Close: 100.0},
		{Date: "2025-01-16", Open: 4.0, High: 5.0, Low: 3.0, Close: 4.0}, // Invalid: -96% decrease
		{Date: "2025-01-17", Open: 98.0, High: 102.0, Low: 97.0, Close: 99.0},
	}

	result := filter.Filter(prices)
	require.Len(t, result, 2)
	assert.Equal(t, 100.0, result[0].Close)
	assert.Equal(t, 99.0, result[1].Close)
}

func TestPriceFilter_Filter_ContextBasedOutlierHigh(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	// Build 30-day context of ~50 EUR prices, then insert a 550 EUR outlier (11x average)
	prices := make([]DailyPrice, 35)
	for i := 0; i < 30; i++ {
		prices[i] = DailyPrice{
			Date:  "2025-01-" + padDay(i+1),
			Open:  50.0,
			High:  52.0,
			Low:   48.0,
			Close: 50.0,
		}
	}
	// Normal prices at end to sandwich the outlier
	prices[30] = DailyPrice{Date: "2025-01-31", Open: 50.0, High: 52.0, Low: 48.0, Close: 50.0}
	prices[31] = DailyPrice{Date: "2025-02-01", Open: 50.0, High: 52.0, Low: 48.0, Close: 50.0}
	prices[32] = DailyPrice{Date: "2025-02-02", Open: 550.0, High: 560.0, Low: 540.0, Close: 550.0} // Invalid: 11x average
	prices[33] = DailyPrice{Date: "2025-02-03", Open: 50.0, High: 52.0, Low: 48.0, Close: 50.0}
	prices[34] = DailyPrice{Date: "2025-02-04", Open: 51.0, High: 53.0, Low: 49.0, Close: 51.0}

	result := filter.Filter(prices)
	// The outlier at index 32 should be excluded
	require.Len(t, result, 34)

	// Verify the outlier is not in the result
	for _, p := range result {
		assert.NotEqual(t, 550.0, p.Close, "550.0 outlier should be excluded")
	}
}

func TestPriceFilter_Filter_ContextBasedOutlierLow(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	// Build 30-day context of ~50 EUR prices, then insert a 4 EUR outlier (0.08x average)
	prices := make([]DailyPrice, 35)
	for i := 0; i < 30; i++ {
		prices[i] = DailyPrice{
			Date:  "2025-01-" + padDay(i+1),
			Open:  50.0,
			High:  52.0,
			Low:   48.0,
			Close: 50.0,
		}
	}
	// Normal prices at end to sandwich the outlier
	prices[30] = DailyPrice{Date: "2025-01-31", Open: 50.0, High: 52.0, Low: 48.0, Close: 50.0}
	prices[31] = DailyPrice{Date: "2025-02-01", Open: 50.0, High: 52.0, Low: 48.0, Close: 50.0}
	prices[32] = DailyPrice{Date: "2025-02-02", Open: 4.0, High: 5.0, Low: 3.0, Close: 4.0} // Invalid: 0.08x average
	prices[33] = DailyPrice{Date: "2025-02-03", Open: 50.0, High: 52.0, Low: 48.0, Close: 50.0}
	prices[34] = DailyPrice{Date: "2025-02-04", Open: 51.0, High: 53.0, Low: 49.0, Close: 51.0}

	result := filter.Filter(prices)
	// The outlier at index 32 should be excluded
	require.Len(t, result, 34)

	// Verify the outlier is not in the result
	for _, p := range result {
		assert.NotEqual(t, 4.0, p.Close, "4.0 outlier should be excluded")
	}
}

func TestPriceFilter_Filter_AllInvalidReturnsEmpty(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	prices := []DailyPrice{
		{Date: "2025-01-15", Open: 50.0, High: 45.0, Low: 48.0, Close: 52.0}, // Invalid: High < Low
		{Date: "2025-01-16", Open: 52.0, High: 54.0, Low: 51.0, Close: 0},    // Invalid: Zero close
		{Date: "2025-01-17", Open: 53.0, High: 56.0, Low: 52.0, Close: -5.0}, // Invalid: Negative close
	}

	result := filter.Filter(prices)
	assert.Empty(t, result)
}

func TestPriceFilter_Filter_PreservesVolumeAndDate(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	vol := int64(1000000)
	prices := []DailyPrice{
		{Date: "2025-01-15", Open: 50.0, High: 55.0, Low: 48.0, Close: 52.0, Volume: &vol},
	}

	result := filter.Filter(prices)
	require.Len(t, result, 1)
	assert.Equal(t, "2025-01-15", result[0].Date)
	require.NotNil(t, result[0].Volume)
	assert.Equal(t, int64(1000000), *result[0].Volume)
}

func TestPriceFilter_Filter_LDOEUAnomalyCase(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	// Real-world LDO.EU case: normal prices around 47, then spike to 44,000
	prices := []DailyPrice{
		{Date: "2025-08-09", Open: 46.0, High: 48.0, Low: 45.0, Close: 47.0},
		{Date: "2025-08-10", Open: 47.0, High: 48.0, Low: 46.0, Close: 47.0},
		{Date: "2025-08-11", Open: 44050.53, High: 44497.59, Low: 44050.53, Close: 44458.62}, // Anomaly
		{Date: "2025-08-12", Open: 47.2, High: 47.3, Low: 46.6, Close: 46.7},
	}

	result := filter.Filter(prices)
	require.Len(t, result, 3)
	assert.Equal(t, 47.0, result[0].Close)
	assert.Equal(t, 47.0, result[1].Close)
	assert.Equal(t, 46.7, result[2].Close)
}

func TestPriceFilter_Filter_MultipleConsecutiveAnomalies(t *testing.T) {
	log := zerolog.Nop()
	filter := NewPriceFilter(log)

	prices := []DailyPrice{
		{Date: "2025-08-09", Open: 46.0, High: 48.0, Low: 45.0, Close: 47.0},
		{Date: "2025-08-10", Open: 44050.53, High: 44497.59, Low: 44050.53, Close: 44458.62}, // Anomaly 1
		{Date: "2025-08-11", Open: 44050.53, High: 44497.59, Low: 44050.53, Close: 44458.62}, // Anomaly 2
		{Date: "2025-08-12", Open: 47.2, High: 47.3, Low: 46.6, Close: 46.7},
	}

	result := filter.Filter(prices)
	require.Len(t, result, 2)
	assert.Equal(t, 47.0, result[0].Close)
	assert.Equal(t, 46.7, result[1].Close)
}

// Helper function to pad day numbers
func padDay(day int) string {
	if day < 10 {
		return "0" + string(rune('0'+day))
	}
	tens := day / 10
	ones := day % 10
	return string(rune('0'+tens)) + string(rune('0'+ones))
}
