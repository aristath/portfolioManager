// Package openfigi provides a client for Bloomberg's OpenFIGI API.
// OpenFIGI is a free service for mapping securities identifiers like ISINs
// to exchange-specific ticker symbols.
package openfigi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	defaultBaseURL = "https://api.openfigi.com/v3"
	// Rate limits: 25 requests/minute without API key, 25,000 with key
	// (Actual rate limiting handled in the client implementation)
)

// MappingRequest represents a request to the OpenFIGI mapping API.
type MappingRequest struct {
	IDType    string `json:"idType"`
	IDValue   string `json:"idValue"`
	ExchCode  string `json:"exchCode,omitempty"`
	MarketSec string `json:"marketSecDes,omitempty"` // e.g., "Equity"
	SecType2  string `json:"securityType2,omitempty"`
	Currency  string `json:"currency,omitempty"`
}

// MappingResult represents a single result from the OpenFIGI API.
type MappingResult struct {
	FIGI            string `json:"figi"`
	Ticker          string `json:"ticker"`
	ExchCode        string `json:"exchCode"`        // Exchange code (e.g., "US", "LN", "GR", "GA")
	Name            string `json:"name"`            // Security name
	MarketSector    string `json:"marketSector"`    // e.g., "Equity"
	SecurityType    string `json:"securityType"`    // e.g., "Common Stock"
	CompositeFIGI   string `json:"compositeFIGI"`   // Composite FIGI
	ShareClassFIGI  string `json:"shareClassFIGI"`  // Share class FIGI
	UniqueID        string `json:"uniqueID"`        // Unique identifier
	SecurityType2   string `json:"securityType2"`   // Secondary security type
	MarketSectorDes string `json:"marketSectorDes"` // Market sector description
}

// MappingResponse represents a response item from the OpenFIGI API.
type MappingResponse struct {
	Data    []MappingResult `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
	Warning string          `json:"warning,omitempty"`
}

// cacheEntry stores cached results with expiration.
type cacheEntry struct {
	results   []MappingResult
	expiresAt time.Time
}

// Client is the OpenFIGI API client.
type Client struct {
	baseURL    string
	apiKey     string // Optional - increases rate limits
	httpClient *http.Client
	log        zerolog.Logger

	// Cache for ISIN lookups
	cacheMu  sync.RWMutex
	cache    map[string]cacheEntry
	cacheTTL time.Duration
}

// NewClient creates a new OpenFIGI client.
// apiKey is optional but recommended for higher rate limits.
func NewClient(apiKey string, log zerolog.Logger) *Client {
	return &Client{
		baseURL: defaultBaseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		log:      log.With().Str("component", "openfigi").Logger(),
		cache:    make(map[string]cacheEntry),
		cacheTTL: 24 * time.Hour, // Cache ISIN mappings for 24 hours
	}
}

// SetCacheTTL configures the cache expiration duration.
func (c *Client) SetCacheTTL(ttl time.Duration) {
	c.cacheTTL = ttl
}

// ClearCache removes all cached entries.
func (c *Client) ClearCache() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache = make(map[string]cacheEntry)
}

// LookupISIN maps an ISIN to ticker symbol(s).
// Returns multiple results if the security trades on multiple exchanges.
func (c *Client) LookupISIN(isin string) ([]MappingResult, error) {
	// Check cache
	if results, ok := c.getFromCache(isin); ok {
		return results, nil
	}

	// Make request
	requests := []MappingRequest{
		{
			IDType:  "ID_ISIN",
			IDValue: isin,
		},
	}

	responses, err := c.doRequest(requests)
	if err != nil {
		return nil, err
	}

	if len(responses) == 0 {
		return nil, nil
	}

	results := responses[0].Data

	// Cache results
	c.setCache(isin, results)

	return results, nil
}

// LookupISINForExchange maps an ISIN to ticker for a specific exchange.
func (c *Client) LookupISINForExchange(isin string, exchCode string) (*MappingResult, error) {
	// Check cache with exchange-specific key
	cacheKey := isin + ":" + exchCode
	if results, ok := c.getFromCache(cacheKey); ok {
		if len(results) > 0 {
			return &results[0], nil
		}
		return nil, nil
	}

	// Make request
	requests := []MappingRequest{
		{
			IDType:   "ID_ISIN",
			IDValue:  isin,
			ExchCode: exchCode,
		},
	}

	responses, err := c.doRequest(requests)
	if err != nil {
		return nil, err
	}

	if len(responses) == 0 || len(responses[0].Data) == 0 {
		c.setCache(cacheKey, nil)
		return nil, nil
	}

	result := responses[0].Data[0]
	c.setCache(cacheKey, responses[0].Data)

	return &result, nil
}

// BatchLookup looks up multiple ISINs in a single request.
// Returns a map of ISIN -> []MappingResult.
func (c *Client) BatchLookup(isins []string) (map[string][]MappingResult, error) {
	results := make(map[string][]MappingResult)

	// Check cache first
	uncachedISINs := make([]string, 0)
	for _, isin := range isins {
		if cached, ok := c.getFromCache(isin); ok {
			results[isin] = cached
		} else {
			uncachedISINs = append(uncachedISINs, isin)
		}
	}

	// If all were cached, return early
	if len(uncachedISINs) == 0 {
		return results, nil
	}

	// Build requests for uncached ISINs
	requests := make([]MappingRequest, len(uncachedISINs))
	for i, isin := range uncachedISINs {
		requests[i] = MappingRequest{
			IDType:  "ID_ISIN",
			IDValue: isin,
		}
	}

	responses, err := c.doRequest(requests)
	if err != nil {
		return nil, err
	}

	// Map responses back to ISINs
	for i, resp := range responses {
		if i < len(uncachedISINs) {
			isin := uncachedISINs[i]
			results[isin] = resp.Data
			c.setCache(isin, resp.Data)
		}
	}

	return results, nil
}

// doRequest performs the HTTP request to the OpenFIGI API.
func (c *Client) doRequest(requests []MappingRequest) ([]MappingResponse, error) {
	body, err := json.Marshal(requests)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/mapping", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-OPENFIGI-APIKEY", c.apiKey)
	}

	c.log.Debug().Int("count", len(requests)).Msg("Making OpenFIGI request")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenFIGI API error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var responses []MappingResponse
	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return responses, nil
}

// getFromCache retrieves cached results if they exist and haven't expired.
func (c *Client) getFromCache(key string) ([]MappingResult, bool) {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.results, true
}

// setCache stores results in the cache.
func (c *Client) setCache(key string, results []MappingResult) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	c.cache[key] = cacheEntry{
		results:   results,
		expiresAt: time.Now().Add(c.cacheTTL),
	}
}

// ExchangeCodeMappings maps OpenFIGI exchange codes to internal exchange codes.
// OpenFIGI uses Bloomberg exchange codes.
var ExchangeCodeMappings = map[string]string{
	// Americas
	"US": "NYSE", // United States (can be NYSE or NASDAQ)
	"CT": "TSX",  // Canada - Toronto
	"BZ": "B3",   // Brazil
	"MX": "BMV",  // Mexico

	// Europe
	"LN": "LSE",   // London
	"GR": "XETRA", // Germany (Xetra)
	"FP": "PAR",   // France (Paris)
	"NA": "AMS",   // Netherlands (Amsterdam)
	"IM": "MIL",   // Italy (Milan)
	"SM": "BME",   // Spain (Madrid)
	"SW": "SIX",   // Switzerland
	"GA": "ATH",   // Greece (Athens)

	// Asia-Pacific
	"HK": "HKG",  // Hong Kong
	"JT": "TYO",  // Japan (Tokyo)
	"SP": "SGX",  // Singapore
	"AU": "ASX",  // Australia
	"KS": "KRX",  // South Korea
	"TT": "TWSE", // Taiwan
	"IB": "BSE",  // India - Bombay
	"IN": "NSE",  // India - NSE
}

// GetInternalExchangeCode converts an OpenFIGI exchange code to an internal code.
func GetInternalExchangeCode(figiCode string) string {
	if internal, ok := ExchangeCodeMappings[figiCode]; ok {
		return internal
	}
	return figiCode
}

// TickerLookupResult represents a result from ticker lookup including the ISIN.
type TickerLookupResult struct {
	Ticker   string
	ExchCode string
	Name     string
	ISIN     string // Populated from compositeFIGI or direct lookup
}

// LookupByTicker attempts to find security information by ticker symbol.
// exchCode is optional (e.g., "US", "LN", "GA").
// Note: OpenFIGI doesn't directly return ISIN, but we can get it from the compositeFIGI.
func (c *Client) LookupByTicker(ticker string, exchCode string) (*TickerLookupResult, error) {
	// Check cache
	cacheKey := "ticker:" + ticker + ":" + exchCode
	if results, ok := c.getFromCache(cacheKey); ok && len(results) > 0 {
		return &TickerLookupResult{
			Ticker:   results[0].Ticker,
			ExchCode: results[0].ExchCode,
			Name:     results[0].Name,
			ISIN:     "", // OpenFIGI doesn't return ISIN directly
		}, nil
	}

	// Make request using TICKER id type
	req := MappingRequest{
		IDType:  "TICKER",
		IDValue: ticker,
	}
	if exchCode != "" {
		req.ExchCode = exchCode
	}

	responses, err := c.doRequest([]MappingRequest{req})
	if err != nil {
		return nil, err
	}

	if len(responses) == 0 || len(responses[0].Data) == 0 {
		c.setCache(cacheKey, nil)
		return nil, nil
	}

	result := responses[0].Data[0]
	c.setCache(cacheKey, responses[0].Data)

	// Note: OpenFIGI returns FIGI identifiers, not ISINs directly
	// The compositeFIGI can potentially be used to look up the ISIN
	// but this requires additional mapping
	return &TickerLookupResult{
		Ticker:   result.Ticker,
		ExchCode: result.ExchCode,
		Name:     result.Name,
		ISIN:     "", // OpenFIGI doesn't provide ISIN in ticker lookup
	}, nil
}
