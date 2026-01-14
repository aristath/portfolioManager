// Package handlers provides HTTP handlers for portfolio allocation management.
package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/aristath/sentinel/internal/events"
	"github.com/aristath/sentinel/internal/modules/allocation"
	"github.com/rs/zerolog"
)

// Handler handles allocation HTTP requests
type Handler struct {
	allocRepo                *allocation.Repository
	alertService             *allocation.ConcentrationAlertService
	portfolioSummaryProvider allocation.PortfolioSummaryProvider
	eventManager             *events.Manager
	log                      zerolog.Logger
}

// NewHandler creates a new allocation handler
func NewHandler(
	allocRepo *allocation.Repository,
	alertService *allocation.ConcentrationAlertService,
	portfolioSummaryProvider allocation.PortfolioSummaryProvider,
	eventManager *events.Manager,
	log zerolog.Logger,
) *Handler {
	return &Handler{
		allocRepo:                allocRepo,
		alertService:             alertService,
		portfolioSummaryProvider: portfolioSummaryProvider,
		eventManager:             eventManager,
		log:                      log.With().Str("handler", "allocation").Logger(),
	}
}

// HandleGetTargets returns allocation targets for geography and industry
func (h *Handler) HandleGetTargets(w http.ResponseWriter, r *http.Request) {
	// Get allocation targets (may be empty)
	geographyTargets, err := h.allocRepo.GetGeographyTargets()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	industryTargets, err := h.allocRepo.GetIndustryTargets()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"geography": geographyTargets,
		"industry":  industryTargets,
	})
}

// HandleGetAvailableGeographies returns list of all available geographies from securities
func (h *Handler) HandleGetAvailableGeographies(w http.ResponseWriter, r *http.Request) {
	geographies, err := h.allocRepo.GetAvailableGeographies()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"geographies": geographies,
	})
}

// HandleGetAvailableIndustries returns list of all available industries from securities
func (h *Handler) HandleGetAvailableIndustries(w http.ResponseWriter, r *http.Request) {
	industries, err := h.allocRepo.GetAvailableIndustries()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"industries": industries,
	})
}

// HandleSetGeographyTargets updates geography allocation targets
func (h *Handler) HandleSetGeographyTargets(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Targets map[string]float64 `json:"targets"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Targets) == 0 {
		h.writeError(w, http.StatusBadRequest, "No targets provided")
		return
	}

	// Validate weights
	for name, weight := range req.Targets {
		if weight < 0 || weight > 1 {
			h.writeError(w, http.StatusBadRequest, fmt.Sprintf("Weight for %s must be between 0 and 1", name))
			return
		}
	}

	// Store targets
	if err := h.allocRepo.SetGeographyTargets(req.Targets); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Emit ALLOCATION_TARGETS_CHANGED event
	if h.eventManager != nil {
		h.eventManager.Emit(events.AllocationTargetsChanged, "allocation", map[string]interface{}{
			"type":  "geography",
			"count": len(req.Targets),
		})
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"weights": req.Targets,
		"count":   len(req.Targets),
	})
}

// HandleSetIndustryTargets updates industry allocation targets
func (h *Handler) HandleSetIndustryTargets(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Targets map[string]float64 `json:"targets"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Targets) == 0 {
		h.writeError(w, http.StatusBadRequest, "No targets provided")
		return
	}

	// Validate weights
	for name, weight := range req.Targets {
		if weight < 0 || weight > 1 {
			h.writeError(w, http.StatusBadRequest, fmt.Sprintf("Weight for %s must be between 0 and 1", name))
			return
		}
	}

	// Store targets
	if err := h.allocRepo.SetIndustryTargets(req.Targets); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Emit ALLOCATION_TARGETS_CHANGED event
	if h.eventManager != nil {
		h.eventManager.Emit(events.AllocationTargetsChanged, "allocation", map[string]interface{}{
			"type":  "industry",
			"count": len(req.Targets),
		})
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"weights": req.Targets,
		"count":   len(req.Targets),
	})
}

// HandleGetCurrentAllocation returns current allocation vs targets
func (h *Handler) HandleGetCurrentAllocation(w http.ResponseWriter, r *http.Request) {
	// Get portfolio summary
	summary, err := h.portfolioSummaryProvider.GetPortfolioSummary()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Detect concentration alerts
	alerts, err := h.alertService.DetectAlerts(summary)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Build response
	response := map[string]interface{}{
		"total_value":  summary.TotalValue,
		"cash_balance": summary.CashBalance,
		"geography":    buildAllocationArray(summary.GeographyAllocations),
		"industry":     buildAllocationArray(summary.IndustryAllocations),
		"alerts":       buildAlertsArray(alerts),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleGetDeviations returns allocation deviation scores
func (h *Handler) HandleGetDeviations(w http.ResponseWriter, r *http.Request) {
	// Get portfolio summary
	summary, err := h.portfolioSummaryProvider.GetPortfolioSummary()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Calculate deviations
	response := map[string]interface{}{
		"geography": calculateDeviationMap(summary.GeographyAllocations),
		"industry":  calculateDeviationMap(summary.IndustryAllocations),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// Helper methods

// calculateDeviationMap converts allocations to deviation status map
func calculateDeviationMap(allocations []allocation.PortfolioAllocation) map[string]interface{} {
	result := make(map[string]interface{})

	for _, a := range allocations {
		status := "balanced"
		if a.Deviation < -0.02 {
			status = "underweight"
		} else if a.Deviation > 0.02 {
			status = "overweight"
		}

		result[a.Name] = map[string]interface{}{
			"deviation": a.Deviation,
			"need":      math.Max(0, -a.Deviation),
			"status":    status,
		}
	}

	return result
}

// buildAllocationArray converts PortfolioAllocation slice to response format
func buildAllocationArray(allocations []allocation.PortfolioAllocation) []map[string]interface{} {
	result := make([]map[string]interface{}, len(allocations))
	for i, a := range allocations {
		result[i] = map[string]interface{}{
			"name":          a.Name,
			"target_pct":    a.TargetPct,
			"current_pct":   a.CurrentPct,
			"current_value": a.CurrentValue,
			"deviation":     a.Deviation,
		}
	}
	return result
}

// buildAlertsArray converts ConcentrationAlert slice to response format
func buildAlertsArray(alerts []allocation.ConcentrationAlert) []map[string]interface{} {
	result := make([]map[string]interface{}, len(alerts))
	for i, alert := range alerts {
		result[i] = map[string]interface{}{
			"type":                alert.Type,
			"name":                alert.Name,
			"current_pct":         alert.CurrentPct,
			"limit_pct":           alert.LimitPct,
			"alert_threshold_pct": alert.AlertThresholdPct,
			"severity":            alert.Severity,
		}
	}
	return result
}

// HandleGetAllocationVsTargets handles GET /api/allocation/vs-targets
func (h *Handler) HandleGetAllocationVsTargets(w http.ResponseWriter, r *http.Request) {
	// Get portfolio summary
	summary, err := h.portfolioSummaryProvider.GetPortfolioSummary()
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to get portfolio summary")
		h.writeError(w, http.StatusInternalServerError, "Failed to get portfolio summary")
		return
	}

	// Combine geography and industry allocations
	allocations := append(summary.GeographyAllocations, summary.IndustryAllocations...)

	// Build detailed comparison - geography allocations first, then industry
	comparison := make([]map[string]interface{}, 0)
	var totalDeviation float64
	var overweightCount int
	var underweightCount int

	for _, alloc := range allocations {
		deviation := alloc.CurrentPct - alloc.TargetPct
		totalDeviation += abs(deviation)

		status := "on_target"
		if deviation > 1.0 {
			status = "overweight"
			overweightCount++
		} else if deviation < -1.0 {
			status = "underweight"
			underweightCount++
		}

		// Determine type based on which list this came from
		allocType := "geography"
		if len(summary.GeographyAllocations) > 0 && len(comparison) >= len(summary.GeographyAllocations) {
			allocType = "industry"
		}

		comparison = append(comparison, map[string]interface{}{
			"group":         alloc.Name,
			"type":          allocType,
			"target_pct":    alloc.TargetPct,
			"current_pct":   alloc.CurrentPct,
			"deviation":     deviation,
			"status":        status,
			"current_value": alloc.CurrentValue,
		})
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"comparison":        comparison,
			"total_deviation":   totalDeviation,
			"overweight_count":  overweightCount,
			"underweight_count": underweightCount,
			"on_target_count":   len(allocations) - overweightCount - underweightCount,
		},
		"metadata": map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleGetRebalanceNeeds handles GET /api/allocation/rebalance-needs
func (h *Handler) HandleGetRebalanceNeeds(w http.ResponseWriter, r *http.Request) {
	// Get portfolio summary
	summary, err := h.portfolioSummaryProvider.GetPortfolioSummary()
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to get portfolio summary")
		h.writeError(w, http.StatusInternalServerError, "Failed to get portfolio summary")
		return
	}

	// Combine geography and industry allocations
	allocations := append(summary.GeographyAllocations, summary.IndustryAllocations...)

	// Calculate rebalancing needs
	rebalanceNeeds := make([]map[string]interface{}, 0)
	var totalRebalanceValue float64
	processed := 0

	for _, alloc := range allocations {
		deviation := alloc.CurrentPct - alloc.TargetPct

		// Only include groups that need rebalancing (>1% deviation)
		if abs(deviation) > 1.0 {
			// Calculate value needed to rebalance
			// Assumes total portfolio value from current_value and current_pct
			totalValue := 0.0
			if alloc.CurrentPct > 0 {
				totalValue = alloc.CurrentValue / (alloc.CurrentPct / 100.0)
			}
			targetValue := totalValue * (alloc.TargetPct / 100.0)
			valueChange := targetValue - alloc.CurrentValue

			totalRebalanceValue += abs(valueChange)

			// Determine type based on position in combined list
			allocType := "geography"
			if len(summary.GeographyAllocations) > 0 && processed >= len(summary.GeographyAllocations) {
				allocType = "industry"
			}

			rebalanceNeeds = append(rebalanceNeeds, map[string]interface{}{
				"group":         alloc.Name,
				"type":          allocType,
				"current_value": alloc.CurrentValue,
				"target_value":  targetValue,
				"value_change":  valueChange,
				"action":        getRebalanceAction(valueChange),
				"priority":      getPriority(abs(deviation)),
			})
		}
		processed++
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"rebalance_needs":                rebalanceNeeds,
			"total_groups_needing_rebalance": len(rebalanceNeeds),
			"total_rebalance_value":          totalRebalanceValue,
			"note":                           "Rebalancing requires trading module integration",
		},
		"metadata": map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleGetGroupContribution handles GET /api/allocation/groups/contribution
func (h *Handler) HandleGetGroupContribution(w http.ResponseWriter, r *http.Request) {
	// Get portfolio summary
	summary, err := h.portfolioSummaryProvider.GetPortfolioSummary()
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to get portfolio summary")
		h.writeError(w, http.StatusInternalServerError, "Failed to get portfolio summary")
		return
	}

	// Calculate diversification metrics by type
	geographicContribution := make(map[string]float64)
	industryContribution := make(map[string]float64)

	for _, alloc := range summary.GeographyAllocations {
		geographicContribution[alloc.Name] = alloc.CurrentPct
	}
	for _, alloc := range summary.IndustryAllocations {
		industryContribution[alloc.Name] = alloc.CurrentPct
	}

	// Calculate Herfindahl-Hirschman Index for each type
	geographicHHI := calculateHHI(geographicContribution)
	industryHHI := calculateHHI(industryContribution)

	// Calculate effective number of groups (1/HHI)
	effectiveGeographicGroups := 1.0 / geographicHHI
	effectiveIndustryGroups := 1.0 / industryHHI

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"geographic": map[string]interface{}{
				"contributions":         geographicContribution,
				"hhi":                   geographicHHI,
				"effective_groups":      effectiveGeographicGroups,
				"diversification_score": (1.0 - geographicHHI) * 100,
			},
			"industry": map[string]interface{}{
				"contributions":         industryContribution,
				"hhi":                   industryHHI,
				"effective_groups":      effectiveIndustryGroups,
				"diversification_score": (1.0 - industryHHI) * 100,
			},
			"interpretation": map[string]string{
				"hhi":                   "Lower is more diversified (range: 0-1)",
				"effective_groups":      "Number of equally-weighted groups equivalent to current allocation",
				"diversification_score": "Higher is better (range: 0-100)",
			},
		},
		"metadata": map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// Helper functions

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func getRebalanceAction(valueChange float64) string {
	if valueChange > 0 {
		return "BUY"
	} else if valueChange < 0 {
		return "SELL"
	}
	return "HOLD"
}

func getPriority(deviation float64) string {
	if deviation >= 5.0 {
		return "high"
	} else if deviation >= 2.0 {
		return "medium"
	}
	return "low"
}

func calculateHHI(weights map[string]float64) float64 {
	var hhi float64
	for _, weight := range weights {
		// Convert percentage to decimal
		decimal := weight / 100.0
		hhi += decimal * decimal
	}
	return hhi
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
