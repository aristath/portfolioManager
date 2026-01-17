package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aristath/sentinel/internal/work"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemHandlers_HandleJobsStatus(t *testing.T) {
	tests := []struct {
		name           string
		setupProcessor func() *work.Processor
		expectedCount  int
		validate       func(t *testing.T, response JobsStatusResponse)
	}{
		{
			name: "returns all work types from registry",
			setupProcessor: func() *work.Processor {
				registry := work.NewRegistry()
				market := work.NewMarketTimingChecker(&MockMarketChecker{allMarketsClosed: true})

				// Register multiple work types
				registry.Register(&work.WorkType{
					ID:           "sync:portfolio",
					MarketTiming: work.AnyTime,
					Interval:     5 * time.Minute,
					DependsOn:    []string{},
					FindSubjects: func() []string {
						return []string{""}
					},
					Execute: func(ctx context.Context, subject string, progress *work.ProgressReporter) error {
						return nil
					},
				})

				registry.Register(&work.WorkType{
					ID:           "planner:weights",
					MarketTiming: work.AnyTime,
					Interval:     0, // On-demand
					DependsOn:    []string{},
					FindSubjects: func() []string {
						return []string{""}
					},
					Execute: func(ctx context.Context, subject string, progress *work.ProgressReporter) error {
						return nil
					},
				})

				return work.NewProcessor(registry, market, nil)
			},
			expectedCount: 2,
			validate: func(t *testing.T, response JobsStatusResponse) {
				assert.Len(t, response.WorkTypes, 2)
				// Should be ordered by registration order (FIFO)
				assert.Equal(t, "sync:portfolio", response.WorkTypes[0].ID)
				assert.Equal(t, "planner:weights", response.WorkTypes[1].ID)
				assert.Equal(t, "5m", response.WorkTypes[0].Interval)
				assert.Equal(t, "0", response.WorkTypes[1].Interval)
			},
		},
		{
			name: "works with empty registry",
			setupProcessor: func() *work.Processor {
				registry := work.NewRegistry()
				market := work.NewMarketTimingChecker(&MockMarketChecker{allMarketsClosed: true})
				return work.NewProcessor(registry, market, nil)
			},
			expectedCount: 0,
			validate: func(t *testing.T, response JobsStatusResponse) {
				assert.Len(t, response.WorkTypes, 0)
			},
		},
		{
			name: "includes all work type metadata",
			setupProcessor: func() *work.Processor {
				registry := work.NewRegistry()
				market := work.NewMarketTimingChecker(&MockMarketChecker{allMarketsClosed: true})

				registry.Register(&work.WorkType{
					ID:           "planner:context",
					MarketTiming: work.DuringMarketOpen,
					Interval:     10 * time.Minute,
					DependsOn:    []string{"planner:weights"},
					FindSubjects: func() []string {
						return []string{""}
					},
					Execute: func(ctx context.Context, subject string, progress *work.ProgressReporter) error {
						return nil
					},
				})

				return work.NewProcessor(registry, market, nil)
			},
			expectedCount: 1,
			validate: func(t *testing.T, response JobsStatusResponse) {
				require.Len(t, response.WorkTypes, 1)
				wt := response.WorkTypes[0]
				assert.Equal(t, "planner:context", wt.ID)
				assert.Equal(t, "DuringMarketOpen", wt.MarketTiming)
				assert.Equal(t, "10m", wt.Interval)
				assert.Equal(t, []string{"planner:weights"}, wt.DependsOn)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := tt.setupProcessor()

			// Create minimal SystemHandlers for testing
			log := zerolog.Nop()
			handlers := &SystemHandlers{
				log:           log,
				workProcessor: processor,
			}

			req := httptest.NewRequest(http.MethodGet, "/api/system/jobs", nil)
			rec := httptest.NewRecorder()

			handlers.HandleJobsStatus(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			var response JobsStatusResponse
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Len(t, response.WorkTypes, tt.expectedCount)
			tt.validate(t, response)
		})
	}
}

// MockMarketChecker is a simple mock for market timing checks
type MockMarketChecker struct {
	isOpen           bool
	allMarketsClosed bool
	isSecurityOpen   map[string]bool
}

func (m *MockMarketChecker) IsAnyMarketOpen() bool {
	return m.isOpen
}

func (m *MockMarketChecker) IsSecurityMarketOpen(isin string) bool {
	if m.isSecurityOpen != nil {
		if open, ok := m.isSecurityOpen[isin]; ok {
			return open
		}
	}
	return m.isOpen
}

func (m *MockMarketChecker) AreAllMarketsClosed() bool {
	return m.allMarketsClosed
}
