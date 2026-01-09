// Package exchangerate provides currency exchange rate fetching and caching functionality.
package exchangerate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Client for exchangerate-api.com
type Client struct {
	baseURL string
	cache   *RateCache
	client  *http.Client
	log     zerolog.Logger
}

// RateCache provides 1-hour in-memory cache
type RateCache struct {
	mu    sync.RWMutex
	rates map[string]cachedRate
}

type cachedRate struct {
	rate      float64
	expiresAt time.Time
}

// NewClient creates a new exchangerate-api.com client
func NewClient(log zerolog.Logger) *Client {
	return &Client{
		baseURL: "https://api.exchangerate-api.com/v4/latest",
		cache:   &RateCache{rates: make(map[string]cachedRate)},
		client:  &http.Client{Timeout: 10 * time.Second},
		log:     log.With().Str("client", "exchangerate-api").Logger(),
	}
}

// GetRate fetches exchange rate with 1-hour cache
func (c *Client) GetRate(fromCurrency, toCurrency string) (float64, error) {
	if fromCurrency == toCurrency {
		return 1.0, nil
	}

	// Check cache
	cacheKey := fromCurrency + ":" + toCurrency
	c.cache.mu.RLock()
	if cached, exists := c.cache.rates[cacheKey]; exists {
		if time.Now().Before(cached.expiresAt) {
			c.cache.mu.RUnlock()
			c.log.Debug().
				Str("from", fromCurrency).
				Str("to", toCurrency).
				Float64("rate", cached.rate).
				Msg("Cache hit")
			return cached.rate, nil
		}
	}
	c.cache.mu.RUnlock()

	// Fetch from API
	url := fmt.Sprintf("%s/%s", c.baseURL, fromCurrency)
	c.log.Debug().Str("url", url).Msg("Fetching rates")

	resp, err := c.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	rate, exists := result.Rates[toCurrency]
	if !exists {
		return 0, fmt.Errorf("rate not found for %s->%s", fromCurrency, toCurrency)
	}

	// Cache for 1 hour
	c.cache.mu.Lock()
	c.cache.rates[cacheKey] = cachedRate{
		rate:      rate,
		expiresAt: time.Now().Add(time.Hour),
	}
	c.cache.mu.Unlock()

	c.log.Info().
		Str("from", fromCurrency).
		Str("to", toCurrency).
		Float64("rate", rate).
		Msg("Fetched rate")

	return rate, nil
}
