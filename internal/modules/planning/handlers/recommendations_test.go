package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aristath/sentinel/internal/modules/planning"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecommendationsHandler_Get_RawList(t *testing.T) {
	// Setup: Create mock repository with test data
	repo := planning.NewInMemoryRecommendationRepository(zerolog.Nop())

	// Add test recommendations
	_, err := repo.CreateOrUpdate(planning.Recommendation{
		Symbol:                "AAPL",
		Name:                  "Apple Inc.",
		Side:                  "BUY",
		Quantity:              10,
		EstimatedPrice:        150.00,
		EstimatedValue:        1500.00,
		Currency:              "USD",
		Reason:                "Underweight vs target allocation",
		Priority:              0,
		CurrentPortfolioScore: 75.5,
		NewPortfolioScore:     78.2,
		ScoreChange:           2.7,
		Status:                "pending",
		PortfolioHash:         "test-hash",
	})
	require.NoError(t, err)

	// Create handler
	handler := NewRecommendationsHandler(nil, repo, zerolog.Nop())

	// Make GET request
	req := httptest.NewRequest(http.MethodGet, "/planning/recommendations", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	recommendations, ok := response["recommendations"].([]interface{})
	require.True(t, ok, "Response should contain recommendations array")
	assert.Len(t, recommendations, 1)

	// Verify recommendation data (JSON keys are capitalized)
	rec1 := recommendations[0].(map[string]interface{})
	assert.Equal(t, "AAPL", rec1["Symbol"])
	assert.Equal(t, "Apple Inc.", rec1["Name"])
	assert.Equal(t, "BUY", rec1["Side"])
	assert.Equal(t, float64(10), rec1["Quantity"])
}

func TestRecommendationsHandler_Get_PlanView(t *testing.T) {
	// Setup: Create mock repository with test data
	repo := planning.NewInMemoryRecommendationRepository(zerolog.Nop())

	// Add test recommendation
	_, err := repo.CreateOrUpdate(planning.Recommendation{
		Symbol:                "AAPL",
		Name:                  "Apple Inc.",
		Side:                  "BUY",
		Quantity:              10,
		EstimatedPrice:        150.00,
		EstimatedValue:        1500.00,
		Currency:              "USD",
		Reason:                "Underweight vs target allocation",
		Priority:              0,
		CurrentPortfolioScore: 75.5,
		NewPortfolioScore:     78.2,
		ScoreChange:           2.7,
		Status:                "pending",
		PortfolioHash:         "test-hash",
	})
	require.NoError(t, err)

	// Create handler
	handler := NewRecommendationsHandler(nil, repo, zerolog.Nop())

	// Make GET request with plan=true
	req := httptest.NewRequest(http.MethodGet, "/planning/recommendations?plan=true", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify plan structure
	steps, ok := response["steps"].([]interface{})
	require.True(t, ok, "Response should contain steps array")
	assert.Len(t, steps, 1)

	// Verify step data (JSON keys are lowercase for plan steps)
	step1 := steps[0].(map[string]interface{})
	assert.Equal(t, float64(1), step1["step"])
	assert.Equal(t, "AAPL", step1["symbol"])
	assert.Equal(t, "BUY", step1["side"])
}

func TestRecommendationsHandler_Get_WithLimit(t *testing.T) {
	// Setup: Create mock repository with multiple recommendations
	repo := planning.NewInMemoryRecommendationRepository(zerolog.Nop())

	// Add 10 test recommendations with different symbols to avoid merging
	for i := 0; i < 10; i++ {
		_, err := repo.CreateOrUpdate(planning.Recommendation{
			Symbol:                "TEST" + string(rune('0'+i)), // TEST0, TEST1, ..., TEST9
			Name:                  "Test Security",
			Side:                  "BUY",
			Quantity:              1,
			EstimatedPrice:        100.00,
			EstimatedValue:        100.00,
			Currency:              "USD",
			Reason:                "Test",
			Priority:              float64(i),
			CurrentPortfolioScore: 75.0,
			NewPortfolioScore:     76.0,
			ScoreChange:           1.0,
			Status:                "pending",
			PortfolioHash:         "test-hash",
		})
		require.NoError(t, err)
	}

	// Create handler
	handler := NewRecommendationsHandler(nil, repo, zerolog.Nop())

	// Make GET request with limit=5
	req := httptest.NewRequest(http.MethodGet, "/planning/recommendations?limit=5", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	recommendations := response["recommendations"].([]interface{})
	assert.Len(t, recommendations, 5, "Should return max 5 recommendations")
}

func TestRecommendationsHandler_Get_EmptyRepository(t *testing.T) {
	// Setup: Create empty repository
	repo := planning.NewInMemoryRecommendationRepository(zerolog.Nop())

	// Create handler
	handler := NewRecommendationsHandler(nil, repo, zerolog.Nop())

	// Make GET request
	req := httptest.NewRequest(http.MethodGet, "/planning/recommendations", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	recommendations := response["recommendations"].([]interface{})
	assert.Len(t, recommendations, 0, "Should return empty array")
}

func TestRecommendationsHandler_Post_CreatePlan(t *testing.T) {
	t.Skip("Skipping POST test - requires complex service setup. POST functionality is tested in integration tests.")
	// This test would verify that POST still works (existing functionality)
	// Full testing of POST requires a properly initialized planning service
	// which has complex dependencies. The route registration test verifies
	// that POST is registered correctly.
}

func TestRecommendationsHandler_MethodNotAllowed(t *testing.T) {
	// Test unsupported methods (PUT, DELETE, etc.)
	repo := planning.NewInMemoryRecommendationRepository(zerolog.Nop())
	handler := NewRecommendationsHandler(nil, repo, zerolog.Nop())

	testCases := []string{
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodHead,
		http.MethodOptions,
	}

	for _, method := range testCases {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/planning/recommendations", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, "Method %s should not be allowed", method)
		})
	}
}
