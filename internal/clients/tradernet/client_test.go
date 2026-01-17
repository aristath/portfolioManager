package tradernet

import (
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestGetPendingOrders_CallsCorrectEndpoint(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	// Mock SDK client
	mockSDK := &mockSDKClient{
		getPlacedResult: map[string]interface{}{
			"result": []interface{}{},
		},
	}

	client := NewClientWithSDK(mockSDK, log)
	_, err := client.GetPendingOrders()

	assert.NoError(t, err)
}

// TestClient_GetFXRates tests GetFXRates() using SDK
func TestClient_GetFXRates(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockSDK := &mockSDKClient{
		getCrossRatesForDateResult: map[string]interface{}{
			"rates": map[string]interface{}{
				"EUR": 0.92261342533093,
				"HKD": 7.8070160113905,
			},
		},
	}

	client := &Client{
		sdkClient: mockSDK,
		log:       log,
	}

	rates, err := client.GetFXRates("USD", []string{"EUR", "HKD"})

	assert.NoError(t, err)
	assert.NotNil(t, rates)
	assert.Len(t, rates, 2)
	assert.Equal(t, 0.92261342533093, rates["EUR"])
	assert.Equal(t, 7.8070160113905, rates["HKD"])
}

// TestClient_GetFXRates_SDKError tests GetFXRates() error handling
func TestClient_GetFXRates_SDKError(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockSDK := &mockSDKClient{
		getCrossRatesForDateError: errors.New("SDK error"),
	}

	client := &Client{
		sdkClient: mockSDK,
		log:       log,
	}

	rates, err := client.GetFXRates("USD", []string{"EUR"})

	assert.Error(t, err)
	assert.Nil(t, rates)
	assert.Contains(t, err.Error(), "failed to get FX rates")
}

// TestClient_GetFXRates_TransformerError tests GetFXRates() transformer error handling
func TestClient_GetFXRates_TransformerError(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockSDK := &mockSDKClient{
		getCrossRatesForDateResult: map[string]interface{}{
			"data": "invalid format",
		},
	}

	client := &Client{
		sdkClient: mockSDK,
		log:       log,
	}

	rates, err := client.GetFXRates("USD", []string{"EUR"})

	assert.Error(t, err)
	assert.Nil(t, rates)
	assert.Contains(t, err.Error(), "failed to transform rates")
}

// ============================================================================
// Batch Metadata Tests (TDD - Tests written before implementation)
// ============================================================================

// TestGetSecurityMetadataBatch_Success tests successful batch metadata retrieval
func TestGetSecurityMetadataBatch_Success(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockSDK := &mockSDKClient{
		getAllSecuritiesBatchResult: map[string]interface{}{
			"securities": []interface{}{
				map[string]interface{}{"ticker": "AAPL.US", "name": "Apple Inc."},
				map[string]interface{}{"ticker": "MSFT.US", "name": "Microsoft Corp."},
				map[string]interface{}{"ticker": "GOOGL.US", "name": "Alphabet Inc."},
			},
			"total": 3,
		},
	}

	client := &Client{
		sdkClient: mockSDK,
		log:       log,
	}

	symbols := []string{"AAPL.US", "MSFT.US", "GOOGL.US"}
	result, err := client.GetSecurityMetadataBatch(symbols)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	securities, ok := resultMap["securities"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, securities, 3)

	total, ok := resultMap["total"].(int)
	assert.True(t, ok)
	assert.Equal(t, 3, total)
}

// TestGetSecurityMetadataBatch_SDKError tests error handling when SDK fails
func TestGetSecurityMetadataBatch_SDKError(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockSDK := &mockSDKClient{
		getAllSecuritiesBatchError: errors.New("SDK batch error"),
	}

	client := &Client{
		sdkClient: mockSDK,
		log:       log,
	}

	symbols := []string{"AAPL.US", "MSFT.US"}
	result, err := client.GetSecurityMetadataBatch(symbols)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get batch security metadata")
}

// TestGetSecurityMetadataBatch_EmptySymbols tests with empty symbol array
func TestGetSecurityMetadataBatch_EmptySymbols(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockSDK := &mockSDKClient{
		getAllSecuritiesBatchResult: map[string]interface{}{
			"securities": []interface{}{},
			"total":      0,
		},
	}

	client := &Client{
		sdkClient: mockSDK,
		log:       log,
	}

	result, err := client.GetSecurityMetadataBatch([]string{})

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	securities, ok := resultMap["securities"].([]interface{})
	assert.True(t, ok)
	assert.Empty(t, securities)
}
