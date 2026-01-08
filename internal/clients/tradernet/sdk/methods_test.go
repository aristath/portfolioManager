package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// TestUserInfo_CallsCorrectEndpoint tests that UserInfo() calls the correct API endpoint
func TestUserInfo_CallsCorrectEndpoint(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	var capturedURL string
	var capturedMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.Path
		capturedMethod = r.Method

		response := map[string]interface{}{
			"result": map[string]interface{}{
				"user": map[string]interface{}{
					"id":   123,
					"name": "Test User",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		publicKey:  "test_public",
		privateKey: "test_private",
		baseURL:    server.URL,
		httpClient: &http.Client{},
		log:        log,
	}

	result, err := client.UserInfo()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "POST", capturedMethod, "UserInfo should use POST method")
	assert.Equal(t, "/api/GetAllUserTexInfo", capturedURL, "UserInfo should call /api/GetAllUserTexInfo endpoint")
}

// TestUserInfo_UsesAuthorizedRequest tests that UserInfo() uses authorizedRequest
func TestUserInfo_UsesAuthorizedRequest(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	var capturedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header

		response := map[string]interface{}{
			"result": map[string]interface{}{
				"user": map[string]interface{}{
					"id": 123,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		publicKey:  "test_public",
		privateKey: "test_private",
		baseURL:    server.URL,
		httpClient: &http.Client{},
		log:        log,
	}

	_, err := client.UserInfo()

	assert.NoError(t, err)

	// Verify authentication headers are present
	assert.Equal(t, "test_public", capturedHeaders.Get("X-NtApi-PublicKey"), "Should include public key header")
	assert.NotEmpty(t, capturedHeaders.Get("X-NtApi-Timestamp"), "Should include timestamp header")
	assert.NotEmpty(t, capturedHeaders.Get("X-NtApi-Sig"), "Should include signature header")
}

// TestUserInfo_ResponseParsing tests that UserInfo() parses response correctly
func TestUserInfo_ResponseParsing(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"result": map[string]interface{}{
				"user": map[string]interface{}{
					"id":    456,
					"name":  "John Doe",
					"email": "john@example.com",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		publicKey:  "test_public",
		privateKey: "test_private",
		baseURL:    server.URL,
		httpClient: &http.Client{},
		log:        log,
	}

	result, err := client.UserInfo()

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify response structure
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")

	resultData, ok := resultMap["result"].(map[string]interface{})
	assert.True(t, ok, "Result should have 'result' key")

	user, ok := resultData["user"].(map[string]interface{})
	assert.True(t, ok, "Result should have 'user' key")

	assert.Equal(t, float64(456), user["id"], "User ID should be 456")
	assert.Equal(t, "John Doe", user["name"], "User name should be 'John Doe'")
}

// TestUserInfo_ErrorHandling tests that UserInfo() handles API errors correctly
func TestUserInfo_ErrorHandling(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return response with error message
		response := map[string]interface{}{
			"result": map[string]interface{}{},
			"errMsg": "Invalid credentials",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		publicKey:  "test_public",
		privateKey: "test_private",
		baseURL:    server.URL,
		httpClient: &http.Client{},
		log:        log,
	}

	result, err := client.UserInfo()

	// Should not return error (matches Python SDK behavior - logs but doesn't raise)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify error message is in response
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	errMsg, ok := resultMap["errMsg"].(string)
	assert.True(t, ok, "Response should have errMsg key")
	assert.Equal(t, "Invalid credentials", errMsg, "Error message should be 'Invalid credentials'")
}

// TestUserInfo_EmptyParams tests that UserInfo() works with empty params
func TestUserInfo_EmptyParams(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	var capturedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		capturedBody = string(body)

		response := map[string]interface{}{
			"result": map[string]interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		publicKey:  "test_public",
		privateKey: "test_private",
		baseURL:    server.URL,
		httpClient: &http.Client{},
		log:        log,
	}

	_, err := client.UserInfo()

	assert.NoError(t, err)

	// Verify empty params are sent as empty JSON object
	assert.Equal(t, "{}", capturedBody, "Empty params should be sent as '{}'")
}
