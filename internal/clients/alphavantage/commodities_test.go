package alphavantage

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommoditiesParseCommodityData(t *testing.T) {
	jsonData := `{
		"name": "WTI Crude Oil",
		"interval": "daily",
		"unit": "dollars per barrel",
		"data": [
			{"date": "2024-01-15", "value": "75.50"},
			{"date": "2024-01-12", "value": "74.80"},
			{"date": "2024-01-11", "value": "73.90"}
		]
	}`

	prices, err := parseCommodityData([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, prices, 3)

	assert.Equal(t, 75.5, prices[0].Value)
	assert.Equal(t, 74.8, prices[1].Value)
	assert.Equal(t, 73.9, prices[2].Value)
}

func TestCommoditiesParseCommodityData_WithMissingValues(t *testing.T) {
	// Some entries may have "." as value indicating no data
	jsonData := `{
		"name": "WTI",
		"interval": "daily",
		"unit": "dollars per barrel",
		"data": [
			{"date": "2024-01-15", "value": "75.50"},
			{"date": "2024-01-14", "value": "."},
			{"date": "2024-01-13", "value": "74.80"}
		]
	}`

	prices, err := parseCommodityData([]byte(jsonData))
	require.NoError(t, err)
	// Entry with "." should be skipped
	require.Len(t, prices, 2)

	assert.Equal(t, 75.5, prices[0].Value)
	assert.Equal(t, 74.8, prices[1].Value)
}

func TestCommoditiesParseCommodityData_Monthly(t *testing.T) {
	jsonData := `{
		"name": "Natural Gas",
		"interval": "monthly",
		"unit": "dollars per million BTU",
		"data": [
			{"date": "2024-01-01", "value": "2.50"},
			{"date": "2023-12-01", "value": "2.80"}
		]
	}`

	prices, err := parseCommodityData([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, prices, 2)

	assert.Equal(t, 2.5, prices[0].Value)
}

func TestCommoditiesParseCommodityData_InvalidJSON(t *testing.T) {
	_, err := parseCommodityData([]byte("not json"))
	assert.Error(t, err)
}

func TestCommoditiesParseCommodityData_EmptyData(t *testing.T) {
	jsonData := `{
		"name": "WTI",
		"interval": "daily",
		"unit": "dollars per barrel",
		"data": []
	}`

	prices, err := parseCommodityData([]byte(jsonData))
	require.NoError(t, err)
	assert.Empty(t, prices)
}

// TestClientCommodityMethods verifies all commodity methods exist.
func TestClientCommodityMethods(t *testing.T) {
	client := NewClient("test-key", zerolog.Nop())

	var _ func(string, string) ([]CommodityPrice, error) = client.GetCommodity
	var _ func(string) ([]CommodityPrice, error) = client.GetWTI
	var _ func(string) ([]CommodityPrice, error) = client.GetBrent
	var _ func(string) ([]CommodityPrice, error) = client.GetNaturalGas
	var _ func(string) ([]CommodityPrice, error) = client.GetCopper
	var _ func(string) ([]CommodityPrice, error) = client.GetAluminum
	var _ func(string) ([]CommodityPrice, error) = client.GetWheat
	var _ func(string) ([]CommodityPrice, error) = client.GetCorn
	var _ func(string) ([]CommodityPrice, error) = client.GetCotton
	var _ func(string) ([]CommodityPrice, error) = client.GetSugar
	var _ func(string) ([]CommodityPrice, error) = client.GetCoffee
	var _ func(string) ([]CommodityPrice, error) = client.GetAllCommodities
}
