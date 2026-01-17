package work

import (
	"database/sql"
)

// Cache provides simple key-value storage with expiration.
type Cache struct {
	db *sql.DB
}

// NewCache creates a new cache instance.
func NewCache(db *sql.DB) *Cache {
	return &Cache{db: db}
}

// GetExpiresAt returns the expiration timestamp for a key.
// Returns 0 if key doesn't exist or is expired.
func (c *Cache) GetExpiresAt(key string) int64 {
	var expiresAt int64
	err := c.db.QueryRow("SELECT expires_at FROM cache WHERE key = ?", key).Scan(&expiresAt)
	if err != nil {
		return 0
	}
	return expiresAt
}

// Set stores a key with expiration timestamp.
func (c *Cache) Set(key string, expiresAt int64) error {
	_, err := c.db.Exec(`
		INSERT INTO cache (key, value, expires_at)
		VALUES (?, '', ?)
		ON CONFLICT(key) DO UPDATE SET
			expires_at = excluded.expires_at
	`, key, expiresAt)
	return err
}
