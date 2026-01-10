package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/aristath/sentinel/internal/modules/opportunities"
	planningdomain "github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory SQLite database with minimal schema
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Minimal schema for testing (opportunities don't require much DB access)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS securities (isin TEXT PRIMARY KEY)`)
	require.NoError(t, err)

	return db
}

// mockTagFilter implements TagFilter interface for testing
type mockTagFilter struct{}

func (m *mockTagFilter) GetOpportunityCandidates(ctx *planningdomain.OpportunityContext, config *planningdomain.PlannerConfiguration) ([]string, error) {
	return []string{}, nil
}

func (m *mockTagFilter) GetSellCandidates(ctx *planningdomain.OpportunityContext, config *planningdomain.PlannerConfiguration) ([]string, error) {
	return []string{}, nil
}

func (m *mockTagFilter) IsMarketVolatile(ctx *planningdomain.OpportunityContext, config *planningdomain.PlannerConfiguration) bool {
	return false
}

// mockSecurityRepo implements SecurityRepository interface for testing
type mockSecurityRepo struct{}

func (m *mockSecurityRepo) GetAllActive() ([]domain.Security, error) {
	return []domain.Security{}, nil
}

func (m *mockSecurityRepo) GetByTags(tags []string) ([]domain.Security, error) {
	return []domain.Security{}, nil
}

func (m *mockSecurityRepo) GetPositionsByTags(positionSymbols []string, tags []string) ([]domain.Security, error) {
	return []domain.Security{}, nil
}

func (m *mockSecurityRepo) GetTagsForSecurity(symbol string) ([]string, error) {
	return []string{}, nil
}

// mockPositionRepo implements PositionRepository interface for testing
type mockPositionRepo struct{}

func (m *mockPositionRepo) GetAll() ([]interface{}, error) {
	return []interface{}{}, nil
}

// mockAllocRepo implements AllocationRepository interface for testing
type mockAllocRepo struct{}

func (m *mockAllocRepo) GetAll() (map[string]float64, error) {
	return map[string]float64{}, nil
}

// mockCashManager implements CashManager interface for testing
type mockCashManager struct{}

func (m *mockCashManager) GetAllCashBalances() (map[string]float64, error) {
	return map[string]float64{"EUR": 1000.0}, nil
}

func (m *mockCashManager) GetCashBalance(currency string) (float64, error) {
	return 1000.0, nil
}

func (m *mockCashManager) UpdateCashPosition(currency string, balance float64) error {
	return nil
}

// mockConfigRepo implements ConfigRepository interface for testing
type mockConfigRepo struct{}

func (m *mockConfigRepo) GetDefaultConfig() (*planningdomain.PlannerConfiguration, error) {
	return &planningdomain.PlannerConfiguration{
		MaxOpportunitiesPerCategory: 10,
	}, nil
}

func TestHandleGetAll(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	db := setupTestDB(t)
	defer db.Close()

	tagFilter := &mockTagFilter{}
	securityRepo := &mockSecurityRepo{}
	service := opportunities.NewService(tagFilter, securityRepo, logger)

	handler := NewHandler(service, nil, nil, nil, nil, nil, db, logger)

	req := httptest.NewRequest("GET", "/api/opportunities/all", nil)
	w := httptest.NewRecorder()

	handler.HandleGetAll(w, req)

	// With nil dependencies, expect 500 error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleGetProfitTaking(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	db := setupTestDB(t)
	defer db.Close()

	tagFilter := &mockTagFilter{}
	securityRepo := &mockSecurityRepo{}
	service := opportunities.NewService(tagFilter, securityRepo, logger)

	handler := NewHandler(service, nil, nil, nil, nil, nil, db, logger)

	req := httptest.NewRequest("GET", "/api/opportunities/profit-taking", nil)
	w := httptest.NewRecorder()

	handler.HandleGetProfitTaking(w, req)

	// With nil dependencies, expect 500 error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleGetAveragingDown(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	db := setupTestDB(t)
	defer db.Close()

	tagFilter := &mockTagFilter{}
	securityRepo := &mockSecurityRepo{}
	service := opportunities.NewService(tagFilter, securityRepo, logger)

	handler := NewHandler(service, nil, nil, nil, nil, nil, db, logger)

	req := httptest.NewRequest("GET", "/api/opportunities/averaging-down", nil)
	w := httptest.NewRecorder()

	handler.HandleGetAveragingDown(w, req)

	// With nil dependencies, expect 500 error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleGetOpportunityBuys(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	db := setupTestDB(t)
	defer db.Close()

	tagFilter := &mockTagFilter{}
	securityRepo := &mockSecurityRepo{}
	service := opportunities.NewService(tagFilter, securityRepo, logger)

	handler := NewHandler(service, nil, nil, nil, nil, nil, db, logger)

	req := httptest.NewRequest("GET", "/api/opportunities/opportunity-buys", nil)
	w := httptest.NewRecorder()

	handler.HandleGetOpportunityBuys(w, req)

	// With nil dependencies, expect 500 error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleGetRebalanceBuys(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	db := setupTestDB(t)
	defer db.Close()

	tagFilter := &mockTagFilter{}
	securityRepo := &mockSecurityRepo{}
	service := opportunities.NewService(tagFilter, securityRepo, logger)

	handler := NewHandler(service, nil, nil, nil, nil, nil, db, logger)

	req := httptest.NewRequest("GET", "/api/opportunities/rebalance-buys", nil)
	w := httptest.NewRecorder()

	handler.HandleGetRebalanceBuys(w, req)

	// With nil dependencies, expect 500 error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleGetRebalanceSells(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	db := setupTestDB(t)
	defer db.Close()

	tagFilter := &mockTagFilter{}
	securityRepo := &mockSecurityRepo{}
	service := opportunities.NewService(tagFilter, securityRepo, logger)

	handler := NewHandler(service, nil, nil, nil, nil, nil, db, logger)

	req := httptest.NewRequest("GET", "/api/opportunities/rebalance-sells", nil)
	w := httptest.NewRecorder()

	handler.HandleGetRebalanceSells(w, req)

	// With nil dependencies, expect 500 error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleGetWeightBased(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	db := setupTestDB(t)
	defer db.Close()

	tagFilter := &mockTagFilter{}
	securityRepo := &mockSecurityRepo{}
	service := opportunities.NewService(tagFilter, securityRepo, logger)

	handler := NewHandler(service, nil, nil, nil, nil, nil, db, logger)

	req := httptest.NewRequest("GET", "/api/opportunities/weight-based", nil)
	w := httptest.NewRecorder()

	handler.HandleGetWeightBased(w, req)

	// With nil dependencies, expect 500 error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleGetRegistry(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	db := setupTestDB(t)
	defer db.Close()

	tagFilter := &mockTagFilter{}
	securityRepo := &mockSecurityRepo{}
	service := opportunities.NewService(tagFilter, securityRepo, logger)

	handler := NewHandler(service, nil, nil, nil, nil, nil, db, logger)

	req := httptest.NewRequest("GET", "/api/opportunities/registry", nil)
	w := httptest.NewRecorder()

	handler.HandleGetRegistry(w, req)

	// Registry endpoint works without other dependencies
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Contains(t, data, "calculators")
	assert.Contains(t, data, "count")

	calculators := data["calculators"].([]interface{})
	assert.Greater(t, len(calculators), 0) // Should have registered calculators
}

func TestRouteIntegration(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	db := setupTestDB(t)
	defer db.Close()

	tagFilter := &mockTagFilter{}
	securityRepo := &mockSecurityRepo{}
	service := opportunities.NewService(tagFilter, securityRepo, logger)

	handler := NewHandler(service, nil, nil, nil, nil, nil, db, logger)

	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{"get all opportunities", "GET", "/opportunities/all", http.StatusInternalServerError},
		{"get profit taking", "GET", "/opportunities/profit-taking", http.StatusInternalServerError},
		{"get averaging down", "GET", "/opportunities/averaging-down", http.StatusInternalServerError},
		{"get opportunity buys", "GET", "/opportunities/opportunity-buys", http.StatusInternalServerError},
		{"get rebalance buys", "GET", "/opportunities/rebalance-buys", http.StatusInternalServerError},
		{"get rebalance sells", "GET", "/opportunities/rebalance-sells", http.StatusInternalServerError},
		{"get weight based", "GET", "/opportunities/weight-based", http.StatusInternalServerError},
		{"get registry", "GET", "/opportunities/registry", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
