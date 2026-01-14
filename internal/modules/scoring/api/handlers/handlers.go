// Package handlers provides HTTP handlers for scoring API.
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aristath/sentinel/internal/modules/scoring/domain"
	"github.com/aristath/sentinel/internal/modules/scoring/scorers"
	"github.com/aristath/sentinel/pkg/formulas"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// Handlers provides HTTP handlers for scoring module
type Handlers struct {
	scorer *scorers.SecurityScorer
	log    zerolog.Logger
}

// NewHandlers creates a new scoring handlers instance
func NewHandlers(log zerolog.Logger) *Handlers {
	return &Handlers{
		scorer: scorers.NewSecurityScorer(),
		log:    log.With().Str("module", "scoring_handlers").Logger(),
	}
}

// ScoreRequest represents a request to score a security
// Note: External data fields (P/E ratio, profit margin, etc.) removed - Sentinel uses internal data only
type ScoreRequest struct {
	FiveYearAvgDivYield *float64                `json:"five_year_avg_div_yield,omitempty"`
	MaxDrawdown         *float64                `json:"max_drawdown,omitempty"`
	SortinoRatio        *float64                `json:"sortino_ratio,omitempty"`
	DividendYield       *float64                `json:"dividend_yield,omitempty"`
	PayoutRatio         *float64                `json:"payout_ratio,omitempty"`
	Symbol              string                  `json:"symbol"`
	ProductType         string                  `json:"product_type,omitempty"` // Product type: EQUITY, ETF, MUTUALFUND, ETC, UNKNOWN
	DailyPrices         []float64               `json:"daily_prices"`
	MonthlyPrices       []formulas.MonthlyPrice `json:"monthly_prices"`
	TargetAnnualReturn  float64                 `json:"target_annual_return,omitempty"`
}

// ScoreResponse represents the response from scoring
type ScoreResponse struct {
	Score *domain.CalculatedSecurityScore `json:"score,omitempty"`
	Error *string                         `json:"error,omitempty"`
}

// HandleScoreSecurity handles POST /api/scoring/score
// Calculates the complete security score
func (h *Handlers) HandleScoreSecurity(w http.ResponseWriter, r *http.Request) {
	var req ScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error().Err(err).Msg("Failed to decode score request")
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Symbol == "" {
		h.writeError(w, "Symbol is required", http.StatusBadRequest)
		return
	}

	if len(req.DailyPrices) == 0 {
		h.writeError(w, "Daily prices are required", http.StatusBadRequest)
		return
	}

	// Build scorer input
	productType := req.ProductType
	if productType == "" {
		productType = "UNKNOWN" // Default if not provided
	}

	input := scorers.ScoreSecurityInput{
		Symbol:              req.Symbol,
		ProductType:         productType,
		DailyPrices:         req.DailyPrices,
		MonthlyPrices:       req.MonthlyPrices,
		TargetAnnualReturn:  req.TargetAnnualReturn,
		DividendYield:       req.DividendYield,
		PayoutRatio:         req.PayoutRatio,
		FiveYearAvgDivYield: req.FiveYearAvgDivYield,
		SortinoRatio:        req.SortinoRatio,
		MaxDrawdown:         req.MaxDrawdown,
	}

	// Calculate score
	score := h.scorer.ScoreSecurityWithDefaults(input)

	h.writeJSON(w, ScoreResponse{Score: score})
}

// HandleGetScoreComponents handles GET /api/scoring/components/{isin}
func (h *Handlers) HandleGetScoreComponents(w http.ResponseWriter, r *http.Request) {
	isin := chi.URLParam(r, "isin")
	if isin == "" {
		h.writeErrorV2(w, "ISIN is required", http.StatusBadRequest)
		return
	}

	// Return 501 Not Implemented - requires database integration
	h.writeJSONV2(w, http.StatusNotImplemented, map[string]interface{}{
		"error": map[string]interface{}{
			"message": "Score component breakdown not yet implemented",
			"code":    "NOT_IMPLEMENTED",
			"details": map[string]string{
				"reason": "Requires database integration for historical score components",
				"isin":   isin,
			},
		},
	})
}

// HandleGetAllScoreComponents handles GET /api/scoring/components/all
func (h *Handlers) HandleGetAllScoreComponents(w http.ResponseWriter, r *http.Request) {
	// Return 501 Not Implemented - requires database integration
	h.writeJSONV2(w, http.StatusNotImplemented, map[string]interface{}{
		"error": map[string]interface{}{
			"message": "All score components endpoint not yet implemented",
			"code":    "NOT_IMPLEMENTED",
			"details": map[string]string{
				"reason": "Requires database integration for all securities with scores",
			},
		},
	})
}

// HandleGetCurrentWeights handles GET /api/scoring/weights/current
func (h *Handlers) HandleGetCurrentWeights(w http.ResponseWriter, r *http.Request) {
	// Return current scoring weights from the actual ScoreWeights constant
	response := map[string]interface{}{
		"data": map[string]interface{}{
			"base_weights":     scorers.ScoreWeights,
			"adaptive_enabled": false,
			"adaptive_weights": map[string]interface{}{
				"note": "Adaptive weights adjust based on market regime",
			},
			"effective_weights": scorers.ScoreWeights,
		},
		"metadata": map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	h.writeJSONV2(w, http.StatusOK, response)
}

// HandleGetAdaptiveWeightHistory handles GET /api/scoring/weights/adaptive-history
func (h *Handlers) HandleGetAdaptiveWeightHistory(w http.ResponseWriter, r *http.Request) {
	// Return 501 Not Implemented - requires time-series storage
	h.writeJSONV2(w, http.StatusNotImplemented, map[string]interface{}{
		"error": map[string]interface{}{
			"message": "Adaptive weight history not yet implemented",
			"code":    "NOT_IMPLEMENTED",
			"details": map[string]string{
				"reason": "Requires time-series database integration for historical weight changes",
			},
		},
	})
}

// HandleGetActiveFormula handles GET /api/scoring/formulas/active
func (h *Handlers) HandleGetActiveFormula(w http.ResponseWriter, r *http.Request) {
	// Return actual scoring components from the ScoreWeights constant
	components := make([]string, 0, len(scorers.ScoreWeights))
	for component := range scorers.ScoreWeights {
		components = append(components, component+"_score")
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"formula":     "default",
			"components":  components,
			"weights":     scorers.ScoreWeights,
			"description": "Quality-focused weights for 15-20 year retirement fund strategy",
			"note":        "Symbolic regression can discover alternative formulas",
		},
		"metadata": map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	h.writeJSONV2(w, http.StatusOK, response)
}

// HandleWhatIfScore handles POST /api/scoring/score/what-if
func (h *Handlers) HandleWhatIfScore(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ISIN    string             `json:"isin"`
		Weights map[string]float64 `json:"weights"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorV2(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ISIN == "" {
		h.writeErrorV2(w, "ISIN is required", http.StatusBadRequest)
		return
	}

	// Validate weights sum to 1.0
	var sum float64
	for _, weight := range req.Weights {
		sum += weight
	}

	if sum < 0.99 || sum > 1.01 {
		h.writeErrorV2(w, "Weights must sum to 1.0", http.StatusBadRequest)
		return
	}

	// Note: Full implementation would recalculate score with custom weights
	response := map[string]interface{}{
		"data": map[string]interface{}{
			"isin":           req.ISIN,
			"custom_weights": req.Weights,
			"original_score": 0.0,
			"custom_score":   0.0,
			"delta":          0.0,
			"note":           "What-if scoring requires database integration for security data",
		},
		"metadata": map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	h.writeJSONV2(w, http.StatusOK, response)
}

// writeJSON writes a JSON response
func (h *Handlers) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

// writeError writes an error response
func (h *Handlers) writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	errMsg := message
	h.writeJSON(w, ScoreResponse{Error: &errMsg})
}

// writeJSONV2 writes a JSON response with status code
func (h *Handlers) writeJSONV2(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

// writeErrorV2 writes an error response (v2 format)
func (h *Handlers) writeErrorV2(w http.ResponseWriter, message string, status int) {
	h.writeJSONV2(w, status, map[string]string{"error": message})
}
