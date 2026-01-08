// Package handlers provides HTTP handlers for symbolic regression API.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/aristath/sentinel/internal/modules/symbolic_regression"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// Handlers provides HTTP handlers for symbolic regression API
type Handlers struct {
	storage   *symbolic_regression.FormulaStorage
	discovery *symbolic_regression.DiscoveryService
	dataPrep  *symbolic_regression.DataPrep
	log       zerolog.Logger
}

// NewHandlers creates new symbolic regression handlers
func NewHandlers(
	storage *symbolic_regression.FormulaStorage,
	discovery *symbolic_regression.DiscoveryService,
	dataPrep *symbolic_regression.DataPrep,
	log zerolog.Logger,
) *Handlers {
	return &Handlers{
		storage:   storage,
		discovery: discovery,
		dataPrep:  dataPrep,
		log:       log.With().Str("component", "symbolic_regression_handlers").Logger(),
	}
}

// HandleListFormulas lists all formulas (active and inactive)
// GET /api/symbolic-regression/formulas?formula_type=expected_return&security_type=stock
func (h *Handlers) HandleListFormulas(w http.ResponseWriter, r *http.Request) {
	formulaTypeStr := r.URL.Query().Get("formula_type")
	securityTypeStr := r.URL.Query().Get("security_type")

	if formulaTypeStr == "" || securityTypeStr == "" {
		h.respondError(w, http.StatusBadRequest, "formula_type and security_type are required")
		return
	}

	formulaType := symbolic_regression.FormulaType(formulaTypeStr)
	securityType := symbolic_regression.SecurityType(securityTypeStr)

	formulas, err := h.storage.GetAllFormulas(formulaType, securityType)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to get formulas")
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve formulas")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"formulas": formulas,
		"count":    len(formulas),
	})
}

// HandleGetActiveFormula gets the active formula for a type
// GET /api/symbolic-regression/formulas/active?formula_type=expected_return&security_type=stock&regime_score=0.3
func (h *Handlers) HandleGetActiveFormula(w http.ResponseWriter, r *http.Request) {
	formulaTypeStr := r.URL.Query().Get("formula_type")
	securityTypeStr := r.URL.Query().Get("security_type")
	regimeScoreStr := r.URL.Query().Get("regime_score")

	if formulaTypeStr == "" || securityTypeStr == "" {
		h.respondError(w, http.StatusBadRequest, "formula_type and security_type are required")
		return
	}

	formulaType := symbolic_regression.FormulaType(formulaTypeStr)
	securityType := symbolic_regression.SecurityType(securityTypeStr)

	var regimePtr *float64
	if regimeScoreStr != "" {
		regimeScore, err := strconv.ParseFloat(regimeScoreStr, 64)
		if err == nil {
			regimePtr = &regimeScore
		}
	}

	formula, err := h.storage.GetActiveFormula(formulaType, securityType, regimePtr)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to get active formula")
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve formula")
		return
	}

	if formula == nil {
		h.respondJSON(w, http.StatusOK, map[string]interface{}{
			"formula": nil,
			"message": "No active formula found",
		})
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"formula": formula,
	})
}

// HandleDeactivateFormula deactivates a formula
// POST /api/symbolic-regression/formulas/{id}/deactivate
func (h *Handlers) HandleDeactivateFormula(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid formula ID")
		return
	}

	err = h.storage.DeactivateFormula(id)
	if err != nil {
		h.log.Error().Err(err).Int64("id", id).Msg("Failed to deactivate formula")
		h.respondError(w, http.StatusInternalServerError, "Failed to deactivate formula")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Formula deactivated",
		"id":      id,
	})
}

// HandleRunDiscovery runs formula discovery
// POST /api/symbolic-regression/discover
func (h *Handlers) HandleRunDiscovery(w http.ResponseWriter, r *http.Request) {
	if h.discovery == nil || h.dataPrep == nil {
		h.respondError(w, http.StatusServiceUnavailable, "Discovery service not available")
		return
	}

	var req DiscoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.FormulaType == "" || req.SecurityType == "" {
		h.respondError(w, http.StatusBadRequest, "formula_type and security_type are required")
		return
	}

	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		h.respondError(w, http.StatusBadRequest, "start_date and end_date are required")
		return
	}

	if req.ForwardMonths <= 0 {
		req.ForwardMonths = 6 // Default to 6 months
	}

	// Run discovery
	// Use default regime ranges for regime-specific discovery
	regimeRanges := symbolic_regression.DefaultRegimeRanges()

	var discoveredFormulas []*symbolic_regression.DiscoveredFormula
	var err error

	securityType := symbolic_regression.SecurityType(req.SecurityType)

	if req.FormulaType == string(symbolic_regression.FormulaTypeExpectedReturn) {
		discoveredFormulas, err = h.discovery.DiscoverExpectedReturnFormula(
			securityType,
			req.StartDate,
			req.EndDate,
			req.ForwardMonths,
			regimeRanges, // Discover separate formulas for each regime
		)
	} else if req.FormulaType == string(symbolic_regression.FormulaTypeScoring) {
		discoveredFormulas, err = h.discovery.DiscoverScoringFormula(
			securityType,
			req.StartDate,
			req.EndDate,
			req.ForwardMonths,
			regimeRanges, // Discover separate formulas for each regime
		)
	} else {
		h.respondError(w, http.StatusBadRequest, "Invalid formula_type")
		return
	}

	if err != nil {
		h.log.Error().Err(err).Msg("Discovery failed")
		h.respondError(w, http.StatusInternalServerError, "Discovery failed: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"formulas": discoveredFormulas,
		"count":    len(discoveredFormulas),
		"message":  "Discovery completed",
	})
}

// DiscoveryRequest represents a discovery request
type DiscoveryRequest struct {
	FormulaType   string    `json:"formula_type"`  // "expected_return" or "scoring"
	SecurityType  string    `json:"security_type"` // "stock" or "etf"
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	ForwardMonths int       `json:"forward_months"` // 6 or 12
}

// Helper methods

func (h *Handlers) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

func (h *Handlers) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]interface{}{
		"error":   true,
		"message": message,
	})
}

// RegisterRoutes registers all symbolic regression routes
func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Route("/symbolic-regression", func(r chi.Router) {
		// Formula management
		r.Get("/formulas", h.HandleListFormulas)
		r.Get("/formulas/active", h.HandleGetActiveFormula)
		r.Post("/formulas/{id}/deactivate", h.HandleDeactivateFormula)

		// Discovery
		r.Post("/discover", h.HandleRunDiscovery)
	})
}
