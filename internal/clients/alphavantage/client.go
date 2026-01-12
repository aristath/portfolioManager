package alphavantage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	baseURL          = "https://www.alphavantage.co/query"
	freeRequestLimit = 25
)

// CacheTTL defines cache expiration durations for different data types.
type CacheTTL struct {
	Fundamentals        time.Duration
	TechnicalIndicators time.Duration
	PriceData           time.Duration
	EconomicIndicators  time.Duration
	Commodities         time.Duration
	ExchangeRates       time.Duration
}

// DefaultCacheTTL returns the default cache expiration durations.
func DefaultCacheTTL() CacheTTL {
	return CacheTTL{
		Fundamentals:        24 * time.Hour,
		TechnicalIndicators: 1 * time.Hour,
		PriceData:           15 * time.Minute,
		EconomicIndicators:  24 * time.Hour,
		Commodities:         1 * time.Hour,
		ExchangeRates:       15 * time.Minute,
	}
}

// cacheEntry stores a cached response with expiration.
type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

// Client is the Alpha Vantage API client.
type Client struct {
	apiKey     string
	httpClient *http.Client
	log        zerolog.Logger
	cacheTTL   CacheTTL

	// Rate limiting
	mu           sync.Mutex
	dailyCounter int
	counterReset time.Time

	// Response cache
	cacheMu sync.RWMutex
	cache   map[string]cacheEntry
}

// NewClient creates a new Alpha Vantage client.
func NewClient(apiKey string, log zerolog.Logger) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		log:          log.With().Str("component", "alphavantage").Logger(),
		cacheTTL:     DefaultCacheTTL(),
		dailyCounter: 0,
		counterReset: nextMidnightUTC(),
		cache:        make(map[string]cacheEntry),
	}
}

// SetCacheTTL configures custom cache expiration durations.
func (c *Client) SetCacheTTL(ttl CacheTTL) {
	c.cacheTTL = ttl
}

// GetRemainingRequests returns the number of remaining API requests for today.
func (c *Client) GetRemainingRequests() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.checkDailyReset()
	return freeRequestLimit - c.dailyCounter
}

// ResetDailyCounter resets the daily request counter (for testing).
func (c *Client) ResetDailyCounter() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dailyCounter = 0
	c.counterReset = nextMidnightUTC()
}

// checkDailyReset resets the counter if a new day has started.
// Must be called with mutex held.
func (c *Client) checkDailyReset() {
	if time.Now().UTC().After(c.counterReset) {
		c.dailyCounter = 0
		c.counterReset = nextMidnightUTC()
	}
}

// checkRateLimit checks if we've hit the daily limit and increments the counter.
func (c *Client) checkRateLimit() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.checkDailyReset()

	remaining := freeRequestLimit - c.dailyCounter
	if remaining <= 5 && remaining > 0 {
		c.log.Warn().Int("remaining", remaining).Msg("Daily limit approaching")
	}
	if remaining <= 0 {
		return ErrRateLimitExceeded{}
	}
	c.dailyCounter++
	return nil
}

// getFromCache retrieves a cached response if it exists and hasn't expired.
func (c *Client) getFromCache(key string) (interface{}, bool) {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.data, true
}

// setCache stores a response in the cache with the given TTL.
func (c *Client) setCache(key string, data interface{}, ttl time.Duration) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	c.cache[key] = cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}
}

// ClearCache removes all cached entries.
func (c *Client) ClearCache() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache = make(map[string]cacheEntry)
}

// buildCacheKey creates a cache key from function and parameters.
// Keys are sorted to ensure consistent cache keys across calls.
func buildCacheKey(function string, params map[string]string) string {
	var parts []string
	parts = append(parts, function)

	// Sort keys for consistent cache key generation
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "apikey" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	return strings.Join(parts, "&")
}

// doRequest performs an HTTP request to the Alpha Vantage API.
func (c *Client) doRequest(function string, params map[string]string) ([]byte, error) {
	// Check rate limit
	if err := c.checkRateLimit(); err != nil {
		return nil, err
	}

	// Build URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	q := u.Query()
	q.Set("function", function)
	q.Set("apikey", c.apiKey)
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	c.log.Debug().Str("function", function).Msg("Making API request")

	// Make request
	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for API error responses
	if err := c.checkAPIError(body); err != nil {
		return nil, err
	}

	return body, nil
}

// checkAPIError checks for common API error responses.
func (c *Client) checkAPIError(body []byte) error {
	// Check for rate limit error
	if strings.Contains(string(body), "Thank you for using Alpha Vantage") {
		// Decrement counter since this request didn't count
		c.mu.Lock()
		if c.dailyCounter > 0 {
			c.dailyCounter--
		}
		c.mu.Unlock()
		return ErrRateLimitExceeded{}
	}

	// Check for invalid API key
	if strings.Contains(string(body), "Invalid API call") ||
		strings.Contains(string(body), "apikey") {
		return ErrInvalidAPIKey{}
	}

	// Check for error message in JSON
	var errorResp struct {
		ErrorMessage string `json:"Error Message"`
		Note         string `json:"Note"`
		Information  string `json:"Information"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil {
		if errorResp.ErrorMessage != "" {
			return fmt.Errorf("API error: %s", errorResp.ErrorMessage)
		}
		if strings.Contains(errorResp.Note, "API call frequency") {
			return ErrRateLimitExceeded{}
		}
	}

	return nil
}

// nextMidnightUTC returns the next midnight in UTC.
func nextMidnightUTC() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
}

// =============================================================================
// Response Parsing Helpers
// =============================================================================

// parseFloat64 converts a string to float64, returning 0 for "None" or empty.
func parseFloat64(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "None" || s == "null" || s == "-" {
		return 0
	}
	// Remove percentage signs
	s = strings.TrimSuffix(s, "%")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// parseFloat64Ptr converts a string to *float64, returning nil for "None" or empty.
func parseFloat64Ptr(s string) *float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "None" || s == "null" || s == "-" {
		return nil
	}
	s = strings.TrimSuffix(s, "%")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

// parseInt64 converts a string to int64, returning 0 for "None" or empty.
func parseInt64(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "None" || s == "null" || s == "-" {
		return 0
	}
	// Handle scientific notation
	if strings.Contains(s, "E") || strings.Contains(s, "e") {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return int64(f)
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		// Try parsing as float and converting
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return int64(f)
	}
	return v
}

// parseInt converts a string to int, returning 0 for "None" or empty.
func parseInt(s string) int {
	return int(parseInt64(s))
}

// parseDate parses a date string in YYYY-MM-DD format.
func parseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// parseDateTime parses a date-time string.
func parseDateTime(s string) time.Time {
	// Try various formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC3339,
	}
	for _, f := range formats {
		t, err := time.Parse(f, s)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}

// Verify that Client implements ClientInterface at compile time.
var _ ClientInterface = (*Client)(nil)
