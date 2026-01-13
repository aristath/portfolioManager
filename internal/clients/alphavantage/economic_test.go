package alphavantage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEconomicParseEconomicData(t *testing.T) {
	jsonData := `{
		"name": "Real Gross Domestic Product",
		"interval": "quarterly",
		"unit": "billions of dollars",
		"data": [
			{"date": "2023-12-31", "value": "25000.5"},
			{"date": "2023-09-30", "value": "24500.2"},
			{"date": "2023-06-30", "value": "24000.8"}
		]
	}`

	data, err := parseEconomicData([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, "Real Gross Domestic Product", data.Name)
	assert.Equal(t, "quarterly", data.Interval)
	assert.Equal(t, "billions of dollars", data.Unit)
	require.Len(t, data.Data, 3)

	assert.Equal(t, 25000.5, data.Data[0].Value)
	assert.Equal(t, 24500.2, data.Data[1].Value)
}

func TestEconomicParseEconomicData_Annual(t *testing.T) {
	jsonData := `{
		"name": "Real GDP",
		"interval": "annual",
		"unit": "billions of dollars",
		"data": [
			{"date": "2023-01-01", "value": "25000.0"},
			{"date": "2022-01-01", "value": "24000.0"}
		]
	}`

	data, err := parseEconomicData([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, "annual", data.Interval)
	require.Len(t, data.Data, 2)
}

func TestEconomicParseEconomicData_WithMissingValues(t *testing.T) {
	jsonData := `{
		"name": "Unemployment Rate",
		"interval": "monthly",
		"unit": "percent",
		"data": [
			{"date": "2024-01-01", "value": "3.7"},
			{"date": "2023-12-01", "value": "."},
			{"date": "2023-11-01", "value": "3.9"}
		]
	}`

	data, err := parseEconomicData([]byte(jsonData))
	require.NoError(t, err)
	// Entry with "." should be skipped
	require.Len(t, data.Data, 2)
}

func TestEconomicParseEconomicData_InvalidJSON(t *testing.T) {
	_, err := parseEconomicData([]byte("not json"))
	assert.Error(t, err)
}

func TestEconomicParseEconomicData_EmptyData(t *testing.T) {
	jsonData := `{
		"name": "CPI",
		"interval": "monthly",
		"unit": "index",
		"data": []
	}`

	data, err := parseEconomicData([]byte(jsonData))
	require.NoError(t, err)
	assert.Empty(t, data.Data)
}

// TestClientEconomicMethods verifies all economic indicator methods exist.
func TestClientEconomicMethods(t *testing.T) {
	client := newTestClient("test-key")

	var _ func(string, string) (*EconomicIndicator, error) = client.GetEconomicIndicator
	var _ func(string) (*EconomicIndicator, error) = client.GetRealGDP
	var _ func() (*EconomicIndicator, error) = client.GetRealGDPPerCapita
	var _ func() (*EconomicIndicator, error) = client.GetUnemployment
	var _ func(string) (*EconomicIndicator, error) = client.GetCPI
	var _ func() (*EconomicIndicator, error) = client.GetInflation
	var _ func(string) (*EconomicIndicator, error) = client.GetFederalFundsRate
	var _ func(string, string) (*EconomicIndicator, error) = client.GetTreasuryYield
	var _ func() (*EconomicIndicator, error) = client.GetRetailSales
	var _ func() (*EconomicIndicator, error) = client.GetDurableGoodsOrders
	var _ func() (*EconomicIndicator, error) = client.GetNonfarmPayroll
}
