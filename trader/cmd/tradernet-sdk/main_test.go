package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// TestHealthHandler_Returns200 tests that the health check endpoint returns 200
func TestHealthHandler_Returns200(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Health check should return 200")

	// Verify response format
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"], "Status should be 'ok'")
}

// TestUserInfoHandler_ExtractsCredentials tests that the handler extracts credentials from headers
func TestUserInfoHandler_ExtractsCredentials(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	var capturedAPIKey string
	var capturedAPISecret string

	// Create a mock SDK that captures credentials
	originalNewClient := newSDKClient
	newSDKClient = func(apiKey, apiSecret string, log zerolog.Logger) SDKClient {
		capturedAPIKey = apiKey
		capturedAPISecret = apiSecret
		return &mockSDKClient{
			userInfoResult: map[string]interface{}{
				"result": map[string]interface{}{
					"user": map[string]interface{}{
						"id": 123,
					},
				},
			},
		}
	}
	defer func() {
		newSDKClient = originalNewClient
	}()

	req := httptest.NewRequest("GET", "/user-info", nil)
	req.Header.Set("X-Tradernet-API-Key", "test_api_key")
	req.Header.Set("X-Tradernet-API-Secret", "test_api_secret")
	w := httptest.NewRecorder()

	userInfoHandler(log)(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 on success")
	assert.Equal(t, "test_api_key", capturedAPIKey, "Should extract API key from header")
	assert.Equal(t, "test_api_secret", capturedAPISecret, "Should extract API secret from header")

	// Verify response format
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool), "Response should have success=true")
	assert.NotNil(t, response["data"], "Response should have data")
}

// TestUserInfoHandler_MissingCredentials tests that the handler returns 400 when credentials are missing
func TestUserInfoHandler_MissingCredentials(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	req := httptest.NewRequest("GET", "/user-info", nil)
	// No credentials in headers
	w := httptest.NewRecorder()

	userInfoHandler(log)(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Should return 400 when credentials are missing")
	assert.Contains(t, w.Body.String(), "Missing credentials", "Error message should mention missing credentials")
}

// TestUserInfoHandler_MissingAPIKey tests that the handler returns 400 when API key is missing
func TestUserInfoHandler_MissingAPIKey(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	req := httptest.NewRequest("GET", "/user-info", nil)
	req.Header.Set("X-Tradernet-API-Secret", "test_secret")
	// Missing API key
	w := httptest.NewRecorder()

	userInfoHandler(log)(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Should return 400 when API key is missing")
}

// TestUserInfoHandler_MissingAPISecret tests that the handler returns 400 when API secret is missing
func TestUserInfoHandler_MissingAPISecret(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	req := httptest.NewRequest("GET", "/user-info", nil)
	req.Header.Set("X-Tradernet-API-Key", "test_key")
	// Missing API secret
	w := httptest.NewRecorder()

	userInfoHandler(log)(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Should return 400 when API secret is missing")
}

// TestUserInfoHandler_SDKError tests that the handler handles SDK errors correctly
func TestUserInfoHandler_SDKError(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	// Create a mock SDK that returns an error
	originalNewClient := newSDKClient
	newSDKClient = func(apiKey, apiSecret string, log zerolog.Logger) SDKClient {
		return &mockSDKClient{
			userInfoError: assert.AnError,
		}
	}
	defer func() {
		newSDKClient = originalNewClient
	}()

	req := httptest.NewRequest("GET", "/user-info", nil)
	req.Header.Set("X-Tradernet-API-Key", "test_key")
	req.Header.Set("X-Tradernet-API-Secret", "test_secret")
	w := httptest.NewRecorder()

	userInfoHandler(log)(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Should return 500 on SDK error")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool), "Response should have success=false")
	assert.NotNil(t, response["error"], "Response should have error")
}

// Mock SDK client for testing
type mockSDKClient struct {
	userInfoResult       interface{}
	userInfoError        error
	accountSummaryResult interface{}
	accountSummaryError  error
}

func (m *mockSDKClient) UserInfo() (interface{}, error) {
	return m.userInfoResult, m.userInfoError
}

func (m *mockSDKClient) AccountSummary() (interface{}, error) {
	return m.accountSummaryResult, m.accountSummaryError
}

// Implement all SDKClient interface methods with no-op implementations
func (m *mockSDKClient) GetUserData() (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) Buy(symbol string, quantity int, price float64, duration string, useMargin bool, customOrderID *int) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) Sell(symbol string, quantity int, price float64, duration string, useMargin bool, customOrderID *int) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetPlaced(active bool) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) Cancel(orderID int) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) CancelAll() (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) Stop(symbol string, price float64) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) TrailingStop(symbol string, percent int) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) TakeProfit(symbol string, price float64) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetHistorical(start, end time.Time) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetTradesHistory(start, end string, tradeID, limit *int, symbol, currency *string) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetClientCpsHistory(dateFrom, dateTo string, cpsDocID, id, limit, offset, cpsStatus *int) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetOrderFiles(orderID, internalID *int) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetBrokerReport(start, end time.Time, period time.Time, dataBlockType *string) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetQuotes(symbols []string) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetCandles(symbol string, start, end time.Time, timeframeSeconds int) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetMarketStatus(market string, mode *string) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetMostTraded(instrumentType, exchange string, gainers bool, limit int) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) ExportSecurities(symbols []string, fields []string) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetNews(query string, symbol *string, storyID *string, limit int) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) FindSymbol(symbol string, exchange *string) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) SecurityInfo(symbol string, sup bool) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) Symbol(symbol string, lang string) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) Symbols(exchange *string) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetOptions(underlying, exchange string) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) CorporateActions(reception int) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetAll(filters map[string]interface{}, showExpired bool) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetPriceAlerts(symbol *string) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) AddPriceAlert(symbol string, price interface{}, triggerType, quoteType, sendTo string, frequency, expire int) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) DeletePriceAlert(alertID int) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) NewUser(login, reception, phone, lastname, firstname string, password *string, utmCampaign *string, tariffID *int) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) CheckMissingFields(step int, office string) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetProfileFields(reception int) (interface{}, error) {
	return nil, nil
}

func (m *mockSDKClient) GetTariffsList() (interface{}, error) {
	return nil, nil
}
