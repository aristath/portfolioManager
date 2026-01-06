package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aristath/arduino-trader/internal/clients/tradernet/sdk"
	"github.com/aristath/arduino-trader/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// SDKClient interface for dependency injection in tests
type SDKClient interface {
	// User & Account
	UserInfo() (interface{}, error)
	AccountSummary() (interface{}, error)
	GetUserData() (interface{}, error)

	// Trading
	Buy(symbol string, quantity int, price float64, duration string, useMargin bool, customOrderID *int) (interface{}, error)
	Sell(symbol string, quantity int, price float64, duration string, useMargin bool, customOrderID *int) (interface{}, error)
	GetPlaced(active bool) (interface{}, error)
	Cancel(orderID int) (interface{}, error)
	CancelAll() (interface{}, error)
	Stop(symbol string, price float64) (interface{}, error)
	TrailingStop(symbol string, percent int) (interface{}, error)
	TakeProfit(symbol string, price float64) (interface{}, error)
	GetHistorical(start, end time.Time) (interface{}, error)

	// Transactions & History
	GetTradesHistory(start, end string, tradeID, limit *int, symbol, currency *string) (interface{}, error)
	GetClientCpsHistory(dateFrom, dateTo string, cpsDocID, id, limit, offset, cpsStatus *int) (interface{}, error)
	GetOrderFiles(orderID, internalID *int) (interface{}, error)
	GetBrokerReport(start, end time.Time, period time.Time, dataBlockType *string) (interface{}, error)

	// Market Data
	GetQuotes(symbols []string) (interface{}, error)
	GetCandles(symbol string, start, end time.Time, timeframeSeconds int) (interface{}, error)
	GetMarketStatus(market string, mode *string) (interface{}, error)
	GetMostTraded(instrumentType, exchange string, gainers bool, limit int) (interface{}, error)
	ExportSecurities(symbols []string, fields []string) (interface{}, error)
	GetNews(query string, symbol *string, storyID *string, limit int) (interface{}, error)

	// Securities
	FindSymbol(symbol string, exchange *string) (interface{}, error)
	SecurityInfo(symbol string, sup bool) (interface{}, error)
	Symbol(symbol string, lang string) (interface{}, error)
	Symbols(exchange *string) (interface{}, error)
	GetOptions(underlying, exchange string) (interface{}, error)
	CorporateActions(reception int) (interface{}, error)
	GetAll(filters map[string]interface{}, showExpired bool) (interface{}, error)

	// Price Alerts
	GetPriceAlerts(symbol *string) (interface{}, error)
	AddPriceAlert(symbol string, price interface{}, triggerType, quoteType, sendTo string, frequency, expire int) (interface{}, error)
	DeletePriceAlert(alertID int) (interface{}, error)

	// User Management
	NewUser(login, reception, phone, lastname, firstname string, password *string, utmCampaign *string, tariffID *int) (interface{}, error)
	CheckMissingFields(step int, office string) (interface{}, error)
	GetProfileFields(reception int) (interface{}, error)

	// Other
	GetTariffsList() (interface{}, error)
}

// newSDKClient creates a new SDK client (can be overridden in tests)
var newSDKClient = func(apiKey, apiSecret string, log zerolog.Logger) SDKClient {
	return sdk.NewClient(apiKey, apiSecret, log)
}

func main() {
	// Initialize logger
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	log := logger.New(logger.Config{
		Level:  logLevel,
		Pretty: true,
	})

	log.Info().Msg("Starting Tradernet SDK microservice")

	// Get port from environment or use default
	port := os.Getenv("TRADERNET_SDK_PORT")
	if port == "" {
		port = "9001"
	}

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Routes
	r.Get("/health", healthHandler)

	// User & Account
	r.Get("/user-info", userInfoHandler(log))
	r.Get("/account-summary", accountSummaryHandler(log))
	r.Get("/user-data", userDataHandler(log))

	// Trading
	r.Post("/buy", buyHandler(log))
	r.Post("/sell", sellHandler(log))
	r.Get("/pending-orders", pendingOrdersHandler(log))
	r.Post("/cancel-order", cancelOrderHandler(log))
	r.Post("/cancel-all", cancelAllHandler(log))
	r.Post("/stop", stopHandler(log))
	r.Post("/trailing-stop", trailingStopHandler(log))
	r.Post("/take-profit", takeProfitHandler(log))
	r.Get("/orders-history", ordersHistoryHandler(log))

	// Portfolio & Transactions
	r.Get("/trades-history", tradesHistoryHandler(log))
	r.Get("/cash-movements", cashMovementsHandler(log))
	r.Get("/order-files", orderFilesHandler(log))
	r.Get("/broker-report", brokerReportHandler(log))

	// Market Data
	r.Post("/quotes", quotesHandler(log))
	r.Post("/candles", candlesHandler(log))
	r.Get("/market-status", marketStatusHandler(log))
	r.Get("/most-traded", mostTradedHandler(log))
	r.Post("/export-securities", exportSecuritiesHandler(log))
	r.Get("/news", newsHandler(log))

	// Securities
	r.Get("/find-symbol", findSymbolHandler(log))
	r.Get("/security-info", securityInfoHandler(log))
	r.Get("/symbol", symbolHandler(log))
	r.Get("/symbols", symbolsHandler(log))
	r.Get("/options", optionsHandler(log))
	r.Get("/corporate-actions", corporateActionsHandler(log))
	r.Post("/get-all", getAllHandler(log))

	// Price Alerts
	r.Get("/price-alerts", priceAlertsHandler(log))
	r.Post("/add-price-alert", addPriceAlertHandler(log))
	r.Post("/delete-price-alert", deletePriceAlertHandler(log))

	// User Management
	r.Post("/new-user", newUserHandler(log))
	r.Get("/check-missing-fields", checkMissingFieldsHandler(log))
	r.Get("/profile-fields", profileFieldsHandler(log))

	// Other
	r.Get("/tariffs-list", tariffsListHandler(log))

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Info().Str("port", port).Msg("Server starting")
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

// healthHandler handles health check requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "ok",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// userInfoHandler handles user info requests
// Credentials are passed via headers: X-Tradernet-API-Key and X-Tradernet-API-Secret
func userInfoHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract credentials from headers
		apiKey := r.Header.Get("X-Tradernet-API-Key")
		apiSecret := r.Header.Get("X-Tradernet-API-Secret")

		if apiKey == "" || apiSecret == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.",
			})
			return
		}

		// Create SDK client with credentials from request
		sdkClient := newSDKClient(apiKey, apiSecret, log)

		// Call SDK method
		result, err := sdkClient.UserInfo()
		if err != nil {
			log.Error().Err(err).Msg("SDK call failed")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    result,
			"error":   nil,
		})
	}
}

// accountSummaryHandler handles account summary requests
// Credentials are passed via headers: X-Tradernet-API-Key and X-Tradernet-API-Secret
func accountSummaryHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract credentials from headers
		apiKey := r.Header.Get("X-Tradernet-API-Key")
		apiSecret := r.Header.Get("X-Tradernet-API-Secret")

		if apiKey == "" || apiSecret == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers.",
			})
			return
		}

		// Create SDK client with credentials from request
		sdkClient := newSDKClient(apiKey, apiSecret, log)

		// Call SDK method
		result, err := sdkClient.AccountSummary()
		if err != nil {
			log.Error().Err(err).Msg("SDK call failed")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    result,
			"error":   nil,
		})
	}
}
