package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

// getSDKClient extracts credentials from HTTP headers and creates a new SDK client.
// This ensures the microservice remains stateless - credentials are passed per request.
//
// Required headers:
//   - X-Tradernet-API-Key: Public API key
//   - X-Tradernet-API-Secret: Private API secret
//
// Returns:
//   - SDKClient: A new SDK client instance configured with the provided credentials
//   - error: If credentials are missing or invalid
func getSDKClient(r *http.Request, log zerolog.Logger) (SDKClient, error) {
	apiKey := r.Header.Get("X-Tradernet-API-Key")
	apiSecret := r.Header.Get("X-Tradernet-API-Secret")

	if apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("missing credentials")
	}

	return newSDKClient(apiKey, apiSecret, log), nil
}

// sendError sends a standardized error response in JSON format.
//
// Parameters:
//   - w: HTTP response writer
//   - statusCode: HTTP status code (e.g., 400, 500)
//   - message: Error message to return to the client
//
// Response format:
//
//	{
//	  "success": false,
//	  "error": "Error message"
//	}
func sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

// sendSuccess sends a standardized success response in JSON format.
//
// Parameters:
//   - w: HTTP response writer
//   - data: Response data to return to the client
//
// Response format:
//
//	{
//	  "success": true,
//	  "data": { ... },
//	  "error": null
//	}
func sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
		"error":   nil,
	})
}

// buyHandler handles POST /buy requests for placing buy orders.
//
// Request body:
//
//	{
//	  "symbol": "AAPL.US",
//	  "quantity": 10,
//	  "price": 150.0,          // Optional, 0.0 for market order
//	  "duration": "day",        // Optional, default: "day"
//	  "use_margin": true,       // Optional, default: true
//	  "custom_order_id": null   // Optional
//	}
//
// Response: Standard success/error format with order information in data field.
//
// See ENDPOINTS.md for complete documentation.
func buyHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		var req struct {
			Symbol        string  `json:"symbol"`
			Quantity      int     `json:"quantity"`
			Price         float64 `json:"price,omitempty"`
			Duration      string  `json:"duration,omitempty"`
			UseMargin     bool    `json:"use_margin,omitempty"`
			CustomOrderID *int    `json:"custom_order_id,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
			return
		}

		// Defaults
		if req.Duration == "" {
			req.Duration = "day"
		}
		if req.Price == 0 {
			req.Price = 0.0 // Market order
		}

		result, err := sdkClient.Buy(req.Symbol, req.Quantity, req.Price, req.Duration, req.UseMargin, req.CustomOrderID)
		if err != nil {
			log.Error().Err(err).Msg("Buy order failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// sellHandler handles POST /sell requests for placing sell orders.
//
// Request body: Same format as buyHandler
//
// Response: Standard success/error format with order information in data field.
//
// See ENDPOINTS.md for complete documentation.
func sellHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		var req struct {
			Symbol        string  `json:"symbol"`
			Quantity      int     `json:"quantity"`
			Price         float64 `json:"price,omitempty"`
			Duration      string  `json:"duration,omitempty"`
			UseMargin     bool    `json:"use_margin,omitempty"`
			CustomOrderID *int    `json:"custom_order_id,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
			return
		}

		// Defaults
		if req.Duration == "" {
			req.Duration = "day"
		}
		if req.Price == 0 {
			req.Price = 0.0 // Market order
		}

		result, err := sdkClient.Sell(req.Symbol, req.Quantity, req.Price, req.Duration, req.UseMargin, req.CustomOrderID)
		if err != nil {
			log.Error().Err(err).Msg("Sell order failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// pendingOrdersHandler handles GET /pending-orders requests for retrieving pending/active orders.
//
// Query parameters:
//   - active (bool, optional): If true, returns only active orders. If false, returns all orders.
//     Default: true
//
// Response: Standard success/error format with orders array in data field.
//
// See ENDPOINTS.md for complete documentation.
func pendingOrdersHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		active := true
		if activeStr := r.URL.Query().Get("active"); activeStr != "" {
			active, _ = strconv.ParseBool(activeStr)
		}

		result, err := sdkClient.GetPlaced(active)
		if err != nil {
			log.Error().Err(err).Msg("Get pending orders failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// cancelOrderHandler handles order cancellation requests
func cancelOrderHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		var req struct {
			OrderID int `json:"order_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
			return
		}

		result, err := sdkClient.Cancel(req.OrderID)
		if err != nil {
			log.Error().Err(err).Msg("Cancel order failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// tradesHistoryHandler handles trades history requests
func tradesHistoryHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		// Parse query parameters
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")
		if start == "" {
			start = "1970-01-01"
		}
		if end == "" {
			end = time.Now().Format("2006-01-02")
		}

		var tradeID, limit *int
		var symbol, currency *string

		if tradeIDStr := r.URL.Query().Get("trade_id"); tradeIDStr != "" {
			if id, err := strconv.Atoi(tradeIDStr); err == nil {
				tradeID = &id
			}
		}
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = &l
			}
		}
		if symbolStr := r.URL.Query().Get("symbol"); symbolStr != "" {
			symbol = &symbolStr
		}
		if currencyStr := r.URL.Query().Get("currency"); currencyStr != "" {
			currency = &currencyStr
		}

		result, err := sdkClient.GetTradesHistory(start, end, tradeID, limit, symbol, currency)
		if err != nil {
			log.Error().Err(err).Msg("Get trades history failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// cashMovementsHandler handles cash movements requests
func cashMovementsHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		// Parse query parameters
		dateFrom := r.URL.Query().Get("date_from")
		dateTo := r.URL.Query().Get("date_to")
		if dateFrom == "" {
			dateFrom = "2011-01-11T00:00:00"
		}
		if dateTo == "" {
			dateTo = time.Now().Format("2006-01-02T15:04:05")
		}

		var limit *int
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = &l
			}
		}

		result, err := sdkClient.GetClientCpsHistory(dateFrom, dateTo, nil, nil, limit, nil, nil)
		if err != nil {
			log.Error().Err(err).Msg("Get cash movements failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// quotesHandler handles quotes requests
func quotesHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		var req struct {
			Symbols []string `json:"symbols"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
			return
		}

		if len(req.Symbols) == 0 {
			sendError(w, http.StatusBadRequest, "symbols array cannot be empty")
			return
		}

		result, err := sdkClient.GetQuotes(req.Symbols)
		if err != nil {
			log.Error().Err(err).Msg("Get quotes failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// candlesHandler handles candles/historical data requests
func candlesHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		var req struct {
			Symbol           string `json:"symbol"`
			Start            string `json:"start"`             // ISO format: "2010-01-01T00:00:00Z"
			End              string `json:"end"`               // ISO format: "2024-01-01T00:00:00Z"
			TimeframeSeconds int    `json:"timeframe_seconds"` // Default: 86400 (1 day)
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
			return
		}

		// Parse dates
		startTime, err := time.Parse(time.RFC3339, req.Start)
		if err != nil {
			// Try alternative format
			startTime, err = time.Parse("2006-01-02T15:04:05", req.Start)
			if err != nil {
				startTime = time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
			}
		}

		endTime, err := time.Parse(time.RFC3339, req.End)
		if err != nil {
			// Try alternative format
			endTime, err = time.Parse("2006-01-02T15:04:05", req.End)
			if err != nil {
				endTime = time.Now()
			}
		}

		timeframe := req.TimeframeSeconds
		if timeframe == 0 {
			timeframe = 86400 // Default: 1 day
		}

		result, err := sdkClient.GetCandles(req.Symbol, startTime, endTime, timeframe)
		if err != nil {
			log.Error().Err(err).Msg("Get candles failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// findSymbolHandler handles symbol lookup requests
func findSymbolHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		symbol := r.URL.Query().Get("symbol")
		if symbol == "" {
			sendError(w, http.StatusBadRequest, "symbol parameter is required")
			return
		}

		var exchange *string
		if exchangeStr := r.URL.Query().Get("exchange"); exchangeStr != "" {
			exchange = &exchangeStr
		}

		result, err := sdkClient.FindSymbol(symbol, exchange)
		if err != nil {
			log.Error().Err(err).Msg("Find symbol failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// securityInfoHandler handles security info requests
func securityInfoHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		symbol := r.URL.Query().Get("symbol")
		if symbol == "" {
			sendError(w, http.StatusBadRequest, "symbol parameter is required")
			return
		}

		sup := true
		if supStr := r.URL.Query().Get("sup"); supStr != "" {
			sup, _ = strconv.ParseBool(supStr)
		}

		result, err := sdkClient.SecurityInfo(symbol, sup)
		if err != nil {
			log.Error().Err(err).Msg("Get security info failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// userDataHandler handles get user data requests
func userDataHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		result, err := sdkClient.GetUserData()
		if err != nil {
			log.Error().Err(err).Msg("Get user data failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// cancelAllHandler handles cancel all orders requests
func cancelAllHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		result, err := sdkClient.CancelAll()
		if err != nil {
			log.Error().Err(err).Msg("Cancel all orders failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// stopHandler handles stop loss order requests
func stopHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		var req struct {
			Symbol string  `json:"symbol"`
			Price  float64 `json:"price"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
			return
		}

		result, err := sdkClient.Stop(req.Symbol, req.Price)
		if err != nil {
			log.Error().Err(err).Msg("Stop order failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// trailingStopHandler handles trailing stop order requests
func trailingStopHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		var req struct {
			Symbol  string `json:"symbol"`
			Percent int    `json:"percent"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
			return
		}

		if req.Percent == 0 {
			req.Percent = 1 // Default
		}

		result, err := sdkClient.TrailingStop(req.Symbol, req.Percent)
		if err != nil {
			log.Error().Err(err).Msg("Trailing stop order failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// takeProfitHandler handles take profit order requests
func takeProfitHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		var req struct {
			Symbol string  `json:"symbol"`
			Price  float64 `json:"price"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
			return
		}

		result, err := sdkClient.TakeProfit(req.Symbol, req.Price)
		if err != nil {
			log.Error().Err(err).Msg("Take profit order failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// ordersHistoryHandler handles orders history requests
func ordersHistoryHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		// Parse query parameters
		startStr := r.URL.Query().Get("start")
		endStr := r.URL.Query().Get("end")

		startTime := time.Date(2011, 1, 11, 0, 0, 0, 0, time.UTC)
		endTime := time.Now()

		if startStr != "" {
			if t, err := time.Parse(time.RFC3339, startStr); err == nil {
				startTime = t
			} else if t, err := time.Parse("2006-01-02T15:04:05", startStr); err == nil {
				startTime = t
			}
		}

		if endStr != "" {
			if t, err := time.Parse(time.RFC3339, endStr); err == nil {
				endTime = t
			} else if t, err := time.Parse("2006-01-02T15:04:05", endStr); err == nil {
				endTime = t
			}
		}

		result, err := sdkClient.GetHistorical(startTime, endTime)
		if err != nil {
			log.Error().Err(err).Msg("Get orders history failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// orderFilesHandler handles order files requests
func orderFilesHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		var orderID, internalID *int
		if orderIDStr := r.URL.Query().Get("order_id"); orderIDStr != "" {
			if id, err := strconv.Atoi(orderIDStr); err == nil {
				orderID = &id
			}
		}
		if internalIDStr := r.URL.Query().Get("internal_id"); internalIDStr != "" {
			if id, err := strconv.Atoi(internalIDStr); err == nil {
				internalID = &id
			}
		}

		if orderID == nil && internalID == nil {
			sendError(w, http.StatusBadRequest, "Either order_id or internal_id parameter is required")
			return
		}

		result, err := sdkClient.GetOrderFiles(orderID, internalID)
		if err != nil {
			log.Error().Err(err).Msg("Get order files failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// brokerReportHandler handles broker report requests
func brokerReportHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		// Parse query parameters
		startStr := r.URL.Query().Get("start")
		endStr := r.URL.Query().Get("end")
		periodStr := r.URL.Query().Get("period")
		dataBlockType := r.URL.Query().Get("type")

		startTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		endTime := time.Now()
		periodTime := time.Date(0, 0, 0, 23, 59, 59, 0, time.UTC)

		if startStr != "" {
			if t, err := time.Parse("2006-01-02", startStr); err == nil {
				startTime = t
			}
		}
		if endStr != "" {
			if t, err := time.Parse("2006-01-02", endStr); err == nil {
				endTime = t
			}
		}
		if periodStr != "" {
			if t, err := time.Parse("15:04:05", periodStr); err == nil {
				periodTime = t
			}
		}

		var dataBlockTypePtr *string
		if dataBlockType != "" {
			dataBlockTypePtr = &dataBlockType
		}

		result, err := sdkClient.GetBrokerReport(startTime, endTime, periodTime, dataBlockTypePtr)
		if err != nil {
			log.Error().Err(err).Msg("Get broker report failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// marketStatusHandler handles market status requests
func marketStatusHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		market := r.URL.Query().Get("market")
		if market == "" {
			market = "*"
		}
		mode := r.URL.Query().Get("mode")

		var modePtr *string
		if mode != "" {
			modePtr = &mode
		}

		result, err := sdkClient.GetMarketStatus(market, modePtr)
		if err != nil {
			log.Error().Err(err).Msg("Get market status failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// mostTradedHandler handles most traded securities requests
func mostTradedHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// This is a plain request (no auth needed)
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			// For plain requests, we can still create a client without credentials
			// But for consistency, we'll require them
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		instrumentType := r.URL.Query().Get("instrument_type")
		if instrumentType == "" {
			instrumentType = "stocks"
		}
		exchange := r.URL.Query().Get("exchange")
		if exchange == "" {
			exchange = "usa"
		}
		gainers := true
		if gainersStr := r.URL.Query().Get("gainers"); gainersStr != "" {
			gainers, _ = strconv.ParseBool(gainersStr)
		}
		limit := 10
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		result, err := sdkClient.GetMostTraded(instrumentType, exchange, gainers, limit)
		if err != nil {
			log.Error().Err(err).Msg("Get most traded failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// exportSecuritiesHandler handles export securities requests
func exportSecuritiesHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		var req struct {
			Symbols []string `json:"symbols"`
			Fields  []string `json:"fields,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
			return
		}

		if len(req.Symbols) == 0 {
			sendError(w, http.StatusBadRequest, "symbols array cannot be empty")
			return
		}

		result, err := sdkClient.ExportSecurities(req.Symbols, req.Fields)
		if err != nil {
			log.Error().Err(err).Msg("Export securities failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// newsHandler handles news requests
func newsHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		query := r.URL.Query().Get("query")
		if query == "" {
			sendError(w, http.StatusBadRequest, "query parameter is required")
			return
		}

		symbol := r.URL.Query().Get("symbol")
		storyID := r.URL.Query().Get("story_id")
		limit := 30
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		var symbolPtr, storyIDPtr *string
		if symbol != "" {
			symbolPtr = &symbol
		}
		if storyID != "" {
			storyIDPtr = &storyID
		}

		result, err := sdkClient.GetNews(query, symbolPtr, storyIDPtr, limit)
		if err != nil {
			log.Error().Err(err).Msg("Get news failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// symbolHandler handles symbol (stock data) requests
func symbolHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		symbol := r.URL.Query().Get("symbol")
		if symbol == "" {
			sendError(w, http.StatusBadRequest, "symbol parameter is required")
			return
		}

		lang := r.URL.Query().Get("lang")
		if lang == "" {
			lang = "en"
		}

		result, err := sdkClient.Symbol(symbol, lang)
		if err != nil {
			log.Error().Err(err).Msg("Get symbol failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// symbolsHandler handles symbols (ready list) requests
func symbolsHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		exchange := r.URL.Query().Get("exchange")
		var exchangePtr *string
		if exchange != "" {
			exchangePtr = &exchange
		}

		result, err := sdkClient.Symbols(exchangePtr)
		if err != nil {
			log.Error().Err(err).Msg("Get symbols failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// optionsHandler handles options requests
func optionsHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		underlying := r.URL.Query().Get("underlying")
		exchange := r.URL.Query().Get("exchange")

		if underlying == "" || exchange == "" {
			sendError(w, http.StatusBadRequest, "underlying and exchange parameters are required")
			return
		}

		result, err := sdkClient.GetOptions(underlying, exchange)
		if err != nil {
			log.Error().Err(err).Msg("Get options failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// corporateActionsHandler handles corporate actions requests
func corporateActionsHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		reception := 35 // Default
		if receptionStr := r.URL.Query().Get("reception"); receptionStr != "" {
			if r, err := strconv.Atoi(receptionStr); err == nil {
				reception = r
			}
		}

		result, err := sdkClient.CorporateActions(reception)
		if err != nil {
			log.Error().Err(err).Msg("Get corporate actions failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// getAllHandler handles get all securities requests
func getAllHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		var req struct {
			Filters     map[string]interface{} `json:"filters,omitempty"`
			ShowExpired bool                   `json:"show_expired,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// If body is empty, use defaults
			req.Filters = make(map[string]interface{})
		}

		result, err := sdkClient.GetAll(req.Filters, req.ShowExpired)
		if err != nil {
			log.Error().Err(err).Msg("Get all securities failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// priceAlertsHandler handles price alerts requests
func priceAlertsHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		symbol := r.URL.Query().Get("symbol")
		var symbolPtr *string
		if symbol != "" {
			symbolPtr = &symbol
		}

		result, err := sdkClient.GetPriceAlerts(symbolPtr)
		if err != nil {
			log.Error().Err(err).Msg("Get price alerts failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// addPriceAlertHandler handles add price alert requests
func addPriceAlertHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		var req struct {
			Symbol      string      `json:"symbol"`
			Price       interface{} `json:"price"`
			TriggerType string      `json:"trigger_type,omitempty"`
			QuoteType   string      `json:"quote_type,omitempty"`
			SendTo      string      `json:"send_to,omitempty"`
			Frequency   int         `json:"frequency,omitempty"`
			Expire      int         `json:"expire,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
			return
		}

		if req.TriggerType == "" {
			req.TriggerType = "crossing"
		}
		if req.QuoteType == "" {
			req.QuoteType = "ltp"
		}
		if req.SendTo == "" {
			req.SendTo = "email"
		}

		result, err := sdkClient.AddPriceAlert(req.Symbol, req.Price, req.TriggerType, req.QuoteType, req.SendTo, req.Frequency, req.Expire)
		if err != nil {
			log.Error().Err(err).Msg("Add price alert failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// deletePriceAlertHandler handles delete price alert requests
func deletePriceAlertHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		var req struct {
			AlertID int `json:"alert_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
			return
		}

		result, err := sdkClient.DeletePriceAlert(req.AlertID)
		if err != nil {
			log.Error().Err(err).Msg("Delete price alert failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// newUserHandler handles new user registration requests
func newUserHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// This is a plain request (no auth needed), but for consistency we'll require headers
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		var req struct {
			Login       string  `json:"login"`
			Reception   string  `json:"reception"`
			Phone       string  `json:"phone"`
			Lastname    string  `json:"lastname"`
			Firstname   string  `json:"firstname"`
			Password    *string `json:"password,omitempty"`
			UtmCampaign *string `json:"utm_campaign,omitempty"`
			TariffID    *int    `json:"tariff_id,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
			return
		}

		result, err := sdkClient.NewUser(req.Login, req.Reception, req.Phone, req.Lastname, req.Firstname, req.Password, req.UtmCampaign, req.TariffID)
		if err != nil {
			log.Error().Err(err).Msg("New user registration failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// checkMissingFieldsHandler handles check missing fields requests
func checkMissingFieldsHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		stepStr := r.URL.Query().Get("step")
		office := r.URL.Query().Get("office")

		if stepStr == "" || office == "" {
			sendError(w, http.StatusBadRequest, "step and office parameters are required")
			return
		}

		step, err := strconv.Atoi(stepStr)
		if err != nil {
			sendError(w, http.StatusBadRequest, "step must be an integer")
			return
		}

		result, err := sdkClient.CheckMissingFields(step, office)
		if err != nil {
			log.Error().Err(err).Msg("Check missing fields failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// profileFieldsHandler handles profile fields requests
func profileFieldsHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		receptionStr := r.URL.Query().Get("reception")
		if receptionStr == "" {
			sendError(w, http.StatusBadRequest, "reception parameter is required")
			return
		}

		reception, err := strconv.Atoi(receptionStr)
		if err != nil {
			sendError(w, http.StatusBadRequest, "reception must be an integer")
			return
		}

		result, err := sdkClient.GetProfileFields(reception)
		if err != nil {
			log.Error().Err(err).Msg("Get profile fields failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}

// tariffsListHandler handles tariffs list requests
func tariffsListHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sdkClient, err := getSDKClient(r, log)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.")
			return
		}

		result, err := sdkClient.GetTariffsList()
		if err != nil {
			log.Error().Err(err).Msg("Get tariffs list failed")
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendSuccess(w, result)
	}
}
