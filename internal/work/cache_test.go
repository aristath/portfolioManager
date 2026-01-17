package work

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestCache_SetAndGet(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create table
	_, err = db.Exec(`
		CREATE TABLE cache (
			key TEXT PRIMARY KEY,
			value TEXT,
			expires_at INTEGER
		) STRICT
	`)
	if err != nil {
		t.Fatal(err)
	}

	cache := NewCache(db)

	// Set cache entry that expires in 1 hour
	expiresAt := time.Now().Add(1 * time.Hour).Unix()
	err = cache.Set("test:key", expiresAt)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Get expiration
	retrieved := cache.GetExpiresAt("test:key")
	if retrieved != expiresAt {
		t.Errorf("Expected %d, got %d", expiresAt, retrieved)
	}
}

func TestCache_GetExpiresAt_NotFound(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create table
	_, err = db.Exec(`
		CREATE TABLE cache (
			key TEXT PRIMARY KEY,
			value TEXT,
			expires_at INTEGER
		) STRICT
	`)
	if err != nil {
		t.Fatal(err)
	}

	cache := NewCache(db)

	// Get non-existent key should return 0
	expiresAt := cache.GetExpiresAt("nonexistent")
	if expiresAt != 0 {
		t.Errorf("Expected 0, got %d", expiresAt)
	}
}

func TestCache_Update(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create table
	_, err = db.Exec(`
		CREATE TABLE cache (
			key TEXT PRIMARY KEY,
			value TEXT,
			expires_at INTEGER
		) STRICT
	`)
	if err != nil {
		t.Fatal(err)
	}

	cache := NewCache(db)

	// Set initial value
	expiresAt1 := time.Now().Add(1 * time.Hour).Unix()
	err = cache.Set("test:key", expiresAt1)
	if err != nil {
		t.Fatal(err)
	}

	// Update with new expiration
	expiresAt2 := time.Now().Add(2 * time.Hour).Unix()
	err = cache.Set("test:key", expiresAt2)
	if err != nil {
		t.Fatal(err)
	}

	// Should have new value
	retrieved := cache.GetExpiresAt("test:key")
	if retrieved != expiresAt2 {
		t.Errorf("Expected %d, got %d", expiresAt2, retrieved)
	}
}
