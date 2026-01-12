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
	"sync/atomic"
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
	apiKeys    []string // Multiple API keys for round-robin rotation
	keyIndex   uint32   // Atomic counter for round-robin key selection
	httpClient *http.Client
	log        zerolog.Logger
	cacheTTL   CacheTTL

	// Rate limiting (per-key)
	mu           sync.Mutex
	keyCounters  []int // Daily counter per key
	counterReset time.Time

	// Response cache
	cacheMu sync.RWMutex
	cache   map[string]cacheEntry
}

// NewClient creates a new Alpha Vantage client.
// apiKeysCSV can be a single key or multiple comma-separated keys for round-robin rotation.
func NewClient(apiKeysCSV string, log zerolog.Logger) *Client {
	// Parse comma-separated keys, trim whitespace, filter empty
	var keys []string
	for _, k := range strings.Split(apiKeysCSV, ",") {
		k = strings.TrimSpace(k)
		if k != "" {
			keys = append(keys, k)
		}
	}

	return &Client{
		apiKeys: keys,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		log:          log.With().Str("component", "alphavantage").Logger(),
		cacheTTL:     DefaultCacheTTL(),
		keyCounters:  make([]int, max(len(keys), 1)), // At least 1 slot for empty key case
		counterReset: nextMidnightUTC(),
		cache:        make(map[string]cacheEntry),
	}
}

// SetCacheTTL configures custom cache expiration durations.
func (c *Client) SetCacheTTL(ttl CacheTTL) {
	c.cacheTTL = ttl
}

// GetRemainingRequests returns the total number of remaining API requests for today across all keys.
func (c *Client) GetRemainingRequests() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.checkDailyReset()

	// Sum remaining requests across all keys
	numKeys := len(c.apiKeys)
	if numKeys == 0 {
		return 0
	}

	totalRemaining := 0
	for i := 0; i < numKeys; i++ {
		remaining := freeRequestLimit - c.keyCounters[i]
		if remaining > 0 {
			totalRemaining += remaining
		}
	}
	return totalRemaining
}

// ResetDailyCounter resets the daily request counters for all keys (for testing).
func (c *Client) ResetDailyCounter() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.keyCounters {
		c.keyCounters[i] = 0
	}
	c.counterReset = nextMidnightUTC()
}

// checkDailyReset resets all key counters if a new day has started.
// Must be called with mutex held.
func (c *Client) checkDailyReset() {
	if time.Now().UTC().After(c.counterReset) {
		for i := range c.keyCounters {
			c.keyCounters[i] = 0
		}
		c.counterReset = nextMidnightUTC()
	}
}

// getNextKeyIndex returns the next key index using round-robin rotation.
// Returns the index atomically for thread safety.
func (c *Client) getNextKeyIndex() int {
	if len(c.apiKeys) == 0 {
		return 0
	}
	idx := atomic.AddUint32(&c.keyIndex, 1) - 1
	return int(idx % uint32(len(c.apiKeys)))
}

// checkRateLimitForKey checks if the specified key has hit the daily limit and increments its counter.
// Must be called with mutex held.
func (c *Client) checkRateLimitForKey(keyIdx int) error {
	c.checkDailyReset()

	if keyIdx >= len(c.keyCounters) {
		return ErrRateLimitExceeded{}
	}

	remaining := freeRequestLimit - c.keyCounters[keyIdx]
	totalRemaining := c.getTotalRemainingLocked()

	if totalRemaining <= 5 && totalRemaining > 0 {
		c.log.Warn().Int("remaining", totalRemaining).Int("key_index", keyIdx).Msg("Daily limit approaching")
	}
	if remaining <= 0 {
		return ErrRateLimitExceeded{}
	}
	c.keyCounters[keyIdx]++
	return nil
}

// getTotalRemainingLocked returns total remaining requests across all keys.
// Must be called with mutex held.
func (c *Client) getTotalRemainingLocked() int {
	total := 0
	for i := 0; i < len(c.apiKeys); i++ {
		remaining := freeRequestLimit - c.keyCounters[i]
		if remaining > 0 {
			total += remaining
		}
	}
	return total
}

// checkRateLimit checks if we've hit the daily limit across all keys.
// This is used for backward compatibility - actual rate checking is done per-key in doRequest.
func (c *Client) checkRateLimit() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.checkDailyReset()

	// Check if any key has remaining capacity
	for i := range c.keyCounters {
		if c.keyCounters[i] < freeRequestLimit {
			c.keyCounters[i]++
			return nil
		}
	}
	return ErrRateLimitExceeded{}
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
// It uses round-robin key selection for load balancing across multiple API keys.
func (c *Client) doRequest(function string, params map[string]string) ([]byte, error) {
	if len(c.apiKeys) == 0 {
		return nil, ErrInvalidAPIKey{}
	}

	// Get next key index using round-robin
	keyIdx := c.getNextKeyIndex()

	// Check rate limit for this specific key
	c.mu.Lock()
	err := c.checkRateLimitForKey(keyIdx)
	c.mu.Unlock()
	if err != nil {
		return nil, err
	}

	apiKey := c.apiKeys[keyIdx]

	// Build URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	q := u.Query()
	q.Set("function", function)
	q.Set("apikey", apiKey)
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	c.log.Debug().Str("function", function).Int("key_index", keyIdx).Msg("Making API request")

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
	if err := c.checkAPIErrorForKey(body, keyIdx); err != nil {
		return nil, err
	}

	return body, nil
}

// checkAPIErrorForKey checks for common API error responses and adjusts the counter for the specific key.
func (c *Client) checkAPIErrorForKey(body []byte, keyIdx int) error {
	// Check for rate limit error
	if strings.Contains(string(body), "Thank you for using Alpha Vantage") {
		// Decrement counter since this request didn't count
		c.mu.Lock()
		if keyIdx < len(c.keyCounters) && c.keyCounters[keyIdx] > 0 {
			c.keyCounters[keyIdx]--
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
			// Decrement counter since this request was rate limited
			c.mu.Lock()
			if keyIdx < len(c.keyCounters) && c.keyCounters[keyIdx] > 0 {
				c.keyCounters[keyIdx]--
			}
			c.mu.Unlock()
			return ErrRateLimitExceeded{}
		}
	}

	return nil
}

// checkAPIError checks for common API error responses (backward compatibility).
func (c *Client) checkAPIError(body []byte) error {
	return c.checkAPIErrorForKey(body, 0)
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
