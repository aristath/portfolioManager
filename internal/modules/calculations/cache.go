// Package calculations provides expiration-based caching for expensive calculations.
// This includes per-security technical indicators (EMA, RSI, Sharpe) and
// portfolio-wide optimizer calculations (covariance matrix, HRP dendrogram).
package calculations

import (
	"database/sql"
	"time"
)

// Cache provides expiration-based caching for calculations.
// It uses an expires_at timestamp to determine cache validity,
// making staleness checks simple and efficient.
type Cache struct {
	db *sql.DB
}

// NewCache creates a new calculation cache.
func NewCache(db *sql.DB) *Cache {
	return &Cache{db: db}
}

// GetTechnical retrieves a cached technical metric if not expired.
// Returns the value and true if found and valid, or 0 and false otherwise.
func (c *Cache) GetTechnical(isin, metric string, period int) (float64, bool) {
	var value float64
	var expiresAt int64

	err := c.db.QueryRow(
		"SELECT value, expires_at FROM technical_cache WHERE isin = ? AND metric = ? AND period = ?",
		isin, metric, period,
	).Scan(&value, &expiresAt)

	if err != nil {
		return 0, false
	}

	// Check if expired
	if time.Now().Unix() >= expiresAt {
		return 0, false
	}

	return value, true
}

// SetTechnical stores a technical metric with the specified TTL.
// Uses upsert to update existing entries.
func (c *Cache) SetTechnical(isin, metric string, period int, value float64, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl).Unix()

	_, err := c.db.Exec(`
		INSERT INTO technical_cache (isin, metric, period, value, expires_at) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(isin, metric, period) DO UPDATE SET value = excluded.value, expires_at = excluded.expires_at
	`, isin, metric, period, value, expiresAt)

	return err
}

// GetOptimizer retrieves a cached optimizer result if not expired.
// Returns the raw bytes and true if found and valid, or nil and false otherwise.
func (c *Cache) GetOptimizer(cacheType, isinHash string) ([]byte, bool) {
	var value []byte
	var expiresAt int64

	err := c.db.QueryRow(
		"SELECT value, expires_at FROM optimizer_cache WHERE cache_type = ? AND isin_hash = ?",
		cacheType, isinHash,
	).Scan(&value, &expiresAt)

	if err != nil {
		return nil, false
	}

	// Check if expired
	if time.Now().Unix() >= expiresAt {
		return nil, false
	}

	return value, true
}

// SetOptimizer stores an optimizer result with the specified TTL.
// The value is stored as raw bytes (typically JSON-encoded).
// Uses upsert to update existing entries.
func (c *Cache) SetOptimizer(cacheType, isinHash string, value []byte, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl).Unix()

	_, err := c.db.Exec(`
		INSERT INTO optimizer_cache (cache_type, isin_hash, value, expires_at) VALUES (?, ?, ?, ?)
		ON CONFLICT(cache_type, isin_hash) DO UPDATE SET value = excluded.value, expires_at = excluded.expires_at
	`, cacheType, isinHash, value, expiresAt)

	return err
}

// GetISINsNeedingCalculation returns ISINs where the specified metric is expired or missing.
// This is used by the idle processor to find securities that need work.
func (c *Cache) GetISINsNeedingCalculation(isins []string, metric string, period int) []string {
	now := time.Now().Unix()
	needs := make([]string, 0)

	for _, isin := range isins {
		var expiresAt int64
		err := c.db.QueryRow(
			"SELECT expires_at FROM technical_cache WHERE isin = ? AND metric = ? AND period = ?",
			isin, metric, period,
		).Scan(&expiresAt)

		// Needs calculation if not found or expired
		if err != nil || expiresAt <= now {
			needs = append(needs, isin)
		}
	}

	return needs
}

// Cleanup removes all expired entries from both tables.
// Returns the total number of entries removed.
func (c *Cache) Cleanup() (int64, error) {
	now := time.Now().Unix()

	result1, err := c.db.Exec("DELETE FROM technical_cache WHERE expires_at < ?", now)
	if err != nil {
		return 0, err
	}
	count1, _ := result1.RowsAffected()

	result2, err := c.db.Exec("DELETE FROM optimizer_cache WHERE expires_at < ?", now)
	if err != nil {
		return count1, err
	}
	count2, _ := result2.RowsAffected()

	return count1 + count2, nil
}

// DeleteTechnicalForISIN removes all cached technical metrics for an ISIN.
// Useful when a security's data needs to be recalculated from scratch.
func (c *Cache) DeleteTechnicalForISIN(isin string) error {
	_, err := c.db.Exec("DELETE FROM technical_cache WHERE isin = ?", isin)
	return err
}

// DeleteOptimizerByType removes all cached optimizer results of a given type.
// Useful when optimizer settings change and all cached results are invalid.
func (c *Cache) DeleteOptimizerByType(cacheType string) error {
	_, err := c.db.Exec("DELETE FROM optimizer_cache WHERE cache_type = ?", cacheType)
	return err
}
