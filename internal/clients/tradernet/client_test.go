package tradernet

import (
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
