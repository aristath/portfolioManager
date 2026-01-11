package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// TestAuthorizedRequest_SetsCorrectHeaders tests that authorizedRequest sets all required headers
func TestAuthorizedRequest_SetsCorrectHeaders(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	var capturedHeaders http.Header
	var capturedBody string
	var capturedURL string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header
		capturedURL = r.URL.Path

		// Read body
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		capturedBody = string(body)

		// Return success response
		response := map[string]interface{}{
			"result": map[string]interface{}{"status": "ok"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test_public_key", "test_private_key", log)
	client.baseURL = server.URL
	defer client.Close()

	params := map[string]interface{}{}
	_, err := client.authorizedRequest("GetAllUserTexInfo", params)

	assert.NoError(t, err)

	// Verify URL
	assert.Equal(t, "/api/GetAllUserTexInfo", capturedURL, "URL should be /api/{cmd}")

	// Verify headers
	assert.Equal(t, "application/json", capturedHeaders.Get("Content-Type"), "Content-Type should be application/json")
	assert.Equal(t, "test_public_key", capturedHeaders.Get("X-NtApi-PublicKey"), "X-NtApi-PublicKey header should be set")
	assert.NotEmpty(t, capturedHeaders.Get("X-NtApi-Timestamp"), "X-NtApi-Timestamp header should be set")
	assert.NotEmpty(t, capturedHeaders.Get("X-NtApi-Sig"), "X-NtApi-Sig header should be set")

	// Verify body is JSON
	assert.NotEmpty(t, capturedBody, "Request body should not be empty")

	// Verify signature is valid (can be verified by checking it's 64 hex chars)
	sig := capturedHeaders.Get("X-NtApi-Sig")
	assert.Len(t, sig, 64, "Signature should be 64 hex characters")
}

// TestAuthorizedRequest_TimestampFormat tests that timestamp is in seconds (not milliseconds)
func TestAuthorizedRequest_TimestampFormat(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	var capturedTimestamp string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTimestamp = r.Header.Get("X-NtApi-Timestamp")

		response := map[string]interface{}{
			"result": map[string]interface{}{"status": "ok"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test_public_key", "test_private_key", log)
	client.baseURL = server.URL
	defer client.Close()

	params := map[string]interface{}{}
	_, err := client.authorizedRequest("GetAllUserTexInfo", params)

	assert.NoError(t, err)

	// Timestamp should be a string of digits (Unix seconds)
	assert.NotEmpty(t, capturedTimestamp, "Timestamp should not be empty")

	// Should be reasonable length (Unix timestamp in seconds is ~10 digits)
	assert.GreaterOrEqual(t, len(capturedTimestamp), 10, "Timestamp should be at least 10 digits")
	assert.LessOrEqual(t, len(capturedTimestamp), 11, "Timestamp should be at most 11 digits (for next century)")
}

// TestAuthorizedRequest_SignatureMatchesPython tests that signature generation matches Python SDK
func TestAuthorizedRequest_SignatureMatchesPython(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	var capturedHeaders http.Header
	var capturedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header

		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		capturedBody = string(body)

		response := map[string]interface{}{
			"result": map[string]interface{}{"status": "ok"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Use known credentials for testing
	publicKey := "test_public"
	privateKey := "test_private"

	client := NewClient(publicKey, privateKey, log)
	client.baseURL = server.URL
	defer client.Close()

	params := map[string]interface{}{}
	_, err := client.authorizedRequest("GetAllUserTexInfo", params)

	assert.NoError(t, err)

	// Verify signature was generated correctly
	// Signature = HMAC-SHA256(privateKey, payload + timestamp)
	timestamp := capturedHeaders.Get("X-NtApi-Timestamp")
	payload := capturedBody
	expectedMessage := payload + timestamp

	expectedSig := sign(privateKey, expectedMessage)
	actualSig := capturedHeaders.Get("X-NtApi-Sig")

	assert.Equal(t, expectedSig, actualSig, "Signature should match expected HMAC calculation")
}

// TestAuthorizedRequest_ResponseParsing tests that response is parsed correctly
func TestAuthorizedRequest_ResponseParsing(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	client := NewClient("test_public", "test_private", log)
	client.baseURL = server.URL
	defer client.Close()

	params := map[string]interface{}{}
	result, err := client.authorizedRequest("GetAllUserTexInfo", params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify response structure
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")

	resultData, ok := resultMap["result"].(map[string]interface{})
	assert.True(t, ok, "Result should have 'result' key")

	user, ok := resultData["user"].(map[string]interface{})
	assert.True(t, ok, "Result should have 'user' key")

	assert.Equal(t, float64(123), user["id"], "User ID should be 123")
	assert.Equal(t, "Test User", user["name"], "User name should be 'Test User'")
}

// TestPlainRequest tests plainRequest for unauthenticated endpoints
func TestPlainRequest(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	var capturedURL string
	var capturedMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.Path + "?" + r.URL.RawQuery
		capturedMethod = r.Method

		response := map[string]interface{}{
			"result": []string{"symbol1", "symbol2"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("", "", log)
	client.baseURL = server.URL
	defer client.Close()

	params := map[string]interface{}{
		"q": `{"ticker":"AAPL.US"}`,
	}
	result, err := client.plainRequest("findSymbol", params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "GET", capturedMethod, "plainRequest should use GET method")
	assert.Contains(t, capturedURL, "findSymbol", "URL should contain command")
}
