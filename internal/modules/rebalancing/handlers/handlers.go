// Package handlers provides HTTP handlers for rebalancing operations.
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aristath/sentinel/internal/modules/rebalancing"
	"github.com/rs/zerolog"
)

// Handler handles rebalancing HTTP requests
type Handler struct {
	service *rebalancing.Service
	log     zerolog.Logger
}

// NewHandler creates a new rebalancing handler
func NewHandler(
	service *rebalancing.Service,
	log zerolog.Logger,
) *Handler {
	return &Handler{
		service: service,
		log:     log.With().Str("handler", "rebalancing").Logger(),
	}
}

// CalculateRebalanceRequest represents a request to calculate rebalancing trades
type CalculateRebalanceRequest struct {
	AvailableCash float64 `json:"available_cash"`
}

// CalculateTargetWeightsRequest represents a request to rebalance to specific target weights
type CalculateTargetWeightsRequest struct {
	TargetWeights map[string]float64 `json:"target_weights"`
	AvailableCash float64            `json:"available_cash"`
}

// SimulateRebalanceRequest represents a request to simulate rebalancing
type SimulateRebalanceRequest struct {
	Trades []map[string]interface{} `json:"trades"`
}

// NegativeBalanceCheckRequest represents a request to check for negative balance scenarios
type NegativeBalanceCheckRequest struct {
	Trades       []map[string]interface{} `json:"trades"`
	CashBalances map[string]float64       `json:"cash_balances"`
}

// HandleCalculateRebalance handles POST /api/rebalancing/calculate
func (h *Handler) HandleCalculateRebalance(w http.ResponseWriter, r *http.Request) {
	var req CalculateRebalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error().Err(err).Msg("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AvailableCash <= 0 {
		http.Error(w, "available_cash must be greater than 0", http.StatusBadRequest)
		return
	}

	// Calculate rebalancing trades
	recommendations, err := h.service.CalculateRebalanceTrades(req.AvailableCash)

	if err != nil {
		h.log.Warn().Err(err).Msg("Failed to calculate rebalancing trades - returning placeholder")
		// Return placeholder response indicating full implementation requires services
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"recommendations": []interface{}{},
				"count":           0,
				"available_cash":  req.AvailableCash,
				"note":            "Full rebalancing calculation requires portfolio state and planning service access",
			},
			"metadata": map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"error":     err.Error(),
			},
		}
		h.writeJSON(w, http.StatusOK, response)
		return
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"recommendations": recommendations,
			"count":           len(recommendations),
			"available_cash":  req.AvailableCash,
		},
		"metadata": map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"note":      "Dry-run calculation - no trades executed",
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleCalculateTargetWeights handles POST /api/rebalancing/calculate/target-weights
func (h *Handler) HandleCalculateTargetWeights(w http.ResponseWriter, r *http.Request) {
	var req CalculateTargetWeightsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error().Err(err).Msg("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.TargetWeights) == 0 {
		http.Error(w, "target_weights is required and must not be empty", http.StatusBadRequest)
		return
	}

	if req.AvailableCash <= 0 {
		http.Error(w, "available_cash must be greater than 0", http.StatusBadRequest)
		return
	}

	// Note: This is a placeholder - full implementation requires custom allocation targets
	// For now, we calculate standard rebalancing and note the custom targets
	recommendations, err := h.service.CalculateRebalanceTrades(req.AvailableCash)

	if err != nil {
		h.log.Warn().Err(err).Msg("Failed to calculate target weight rebalancing - returning placeholder")
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"recommendations": []interface{}{},
				"count":           0,
				"target_weights":  req.TargetWeights,
				"available_cash":  req.AvailableCash,
				"note":            "Custom target weights require full allocation override and portfolio state access",
			},
			"metadata": map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"error":     err.Error(),
			},
		}
		h.writeJSON(w, http.StatusOK, response)
		return
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"recommendations": recommendations,
			"count":           len(recommendations),
			"target_weights":  req.TargetWeights,
			"available_cash":  req.AvailableCash,
		},
		"metadata": map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"note":      "Custom target weights require full allocation override - using standard rebalancing",
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleGetTriggers handles GET /api/rebalancing/triggers
func (h *Handler) HandleGetTriggers(w http.ResponseWriter, r *http.Request) {
	// Check triggers using current portfolio state
	result, err := h.service.CheckTriggers()
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to check rebalancing triggers")
		http.Error(w, "Failed to check rebalancing triggers", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"should_rebalance": result.ShouldRebalance,
			"reason":           result.Reason,
		},
		"metadata": map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleGetMinTradeAmount handles GET /api/rebalancing/min-trade-amount
func (h *Handler) HandleGetMinTradeAmount(w http.ResponseWriter, r *http.Request) {
	// Query parameters with defaults
	transactionCostFixed := 2.0     // €2.00 fixed cost
	transactionCostPercent := 0.002 // 0.2% variable cost
	maxCostRatio := 0.01            // 1% max cost ratio

	// Allow override via query params
	if r.URL.Query().Get("fixed_cost") != "" {
		var val float64
		if _, err := fmt.Sscanf(r.URL.Query().Get("fixed_cost"), "%f", &val); err == nil {
			transactionCostFixed = val
		}
	}
	if r.URL.Query().Get("percent_cost") != "" {
		var val float64
		if _, err := fmt.Sscanf(r.URL.Query().Get("percent_cost"), "%f", &val); err == nil {
			transactionCostPercent = val
		}
	}
	if r.URL.Query().Get("max_cost_ratio") != "" {
		var val float64
		if _, err := fmt.Sscanf(r.URL.Query().Get("max_cost_ratio"), "%f", &val); err == nil {
			maxCostRatio = val
		}
	}

	minTradeAmount := rebalancing.CalculateMinTradeAmount(
		transactionCostFixed,
		transactionCostPercent,
		maxCostRatio,
	)

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"min_trade_amount":         minTradeAmount,
			"transaction_cost_fixed":   transactionCostFixed,
			"transaction_cost_percent": transactionCostPercent,
			"max_cost_ratio":           maxCostRatio,
			"explanation":              "Minimum trade amount where transaction costs are acceptable",
		},
		"metadata": map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleSimulateRebalance handles POST /api/rebalancing/simulate-rebalance
func (h *Handler) HandleSimulateRebalance(w http.ResponseWriter, r *http.Request) {
	var req SimulateRebalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error().Err(err).Msg("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Trades) == 0 {
		http.Error(w, "trades is required and must not be empty", http.StatusBadRequest)
		return
	}

	// Convert request trades to service trades
	trades := make([]rebalancing.SimulateTrade, 0, len(req.Trades))
	for _, tradeMap := range req.Trades {
		// Extract trade fields
		isin, _ := tradeMap["isin"].(string)
		symbol, _ := tradeMap["symbol"].(string)
		side, _ := tradeMap["side"].(string)
		quantity, _ := tradeMap["quantity"].(float64)
		price, _ := tradeMap["price"].(float64)
		currency, _ := tradeMap["currency"].(string)

		// Validate required fields
		if side == "" || quantity <= 0 || price <= 0 {
			http.Error(w, "Each trade must have side, quantity > 0, and price > 0", http.StatusBadRequest)
			return
		}

		trades = append(trades, rebalancing.SimulateTrade{
			ISIN:     isin,
			Symbol:   symbol,
			Side:     side,
			Quantity: quantity,
			Price:    price,
			Currency: currency,
		})
	}

	// Simulate rebalance
	result, err := h.service.SimulateRebalance(trades)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to simulate rebalance")
		http.Error(w, "Failed to simulate rebalance", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"simulated_positions": result.Positions,
			"simulated_cash":      result.CashBalances,
			"total_value":         result.TotalValue,
			"total_cost":          result.TotalCost,
			"trades_applied":      result.TradesApplied,
		},
		"metadata": map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleNegativeBalanceCheck handles POST /api/rebalancing/negative-balance-check
func (h *Handler) HandleNegativeBalanceCheck(w http.ResponseWriter, r *http.Request) {
	var req NegativeBalanceCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error().Err(err).Msg("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Trades) == 0 {
		http.Error(w, "trades is required and must not be empty", http.StatusBadRequest)
		return
	}

	// Simulate cash balances after trades
	simulatedBalances := make(map[string]float64)
	for currency, balance := range req.CashBalances {
		simulatedBalances[currency] = balance
	}

	hasNegativeBalance := false
	negativeBalances := make(map[string]float64)

	for _, trade := range req.Trades {
		side, _ := trade["side"].(string)
		quantity, _ := trade["quantity"].(float64)
		price, _ := trade["price"].(float64)
		currency, _ := trade["currency"].(string)

		if currency == "" {
			currency = "EUR" // Default
		}

		value := quantity * price
		cost := 2.0 + (value * 0.002) // Fixed + variable commission

		if side == "BUY" {
			simulatedBalances[currency] -= (value + cost)
		} else if side == "SELL" {
			simulatedBalances[currency] += (value - cost)
		}

		// Check for negative balance
		if simulatedBalances[currency] < 0 {
			hasNegativeBalance = true
			negativeBalances[currency] = simulatedBalances[currency]
		}
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"has_negative_balance": hasNegativeBalance,
			"negative_balances":    negativeBalances,
			"simulated_balances":   simulatedBalances,
			"original_balances":    req.CashBalances,
		},
		"metadata": map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}
