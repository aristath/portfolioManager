// Package database provides database connection and initialization functionality.
// It manages SQLite database connections with production-grade configuration,
// including profile-based PRAGMA settings, connection pooling, schema migration,
// and health checks. Supports 8-database architecture with different profiles
// for different use cases (ledger for safety, cache for speed, standard for balance).
package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver (no CGo dependency)
)

// schemaFiles embeds all SQL schema files into the binary at compile time.
// This ensures schemas are always available regardless of deployment location.
//
//go:embed schemas/*.sql
var schemaFiles embed.FS

// DatabaseProfile defines different configuration profiles for databases.
// Each profile applies different PRAGMA settings optimized for specific use cases.
type DatabaseProfile string

const (
	// ProfileLedger - Maximum safety for immutable audit trail.
	// Uses FULL synchronous mode (fsync after every write) and disables auto-vacuum
	// to ensure data integrity for financial records. Never shrinks database.
	ProfileLedger DatabaseProfile = "ledger"
	// ProfileCache - Maximum speed for ephemeral data.
	// Uses OFF synchronous mode (no fsync) and FULL auto-vacuum for space reclamation.
	// Suitable for job history, temporary calculations, and other non-critical data.
	ProfileCache DatabaseProfile = "cache"
	// ProfileStandard - Balanced configuration for most databases.
	// Uses NORMAL synchronous mode (fsync at checkpoints) and INCREMENTAL auto-vacuum.
	// Suitable for universe, portfolio, history, and other operational databases.
	ProfileStandard DatabaseProfile = "standard"
)

// DB wraps the database connection with production-grade configuration.
// It provides a clean interface for database operations while managing
// connection pooling, PRAGMA settings, schema migration, and health checks.
type DB struct {
	conn    *sql.DB         // Underlying SQLite connection
	path    string          // Absolute path to database file
	profile DatabaseProfile // Configuration profile (ledger, cache, standard)
	name    string          // Database name for logging (e.g., "universe", "ledger")
}

// Config holds database configuration used when creating a new database connection.
type Config struct {
	Path    string          // Database file path (resolved to absolute)
	Profile DatabaseProfile // Configuration profile (defaults to ProfileStandard)
	Name    string          // Friendly name for logging (e.g., "universe", "ledger")
}

// New creates a new database connection with production-grade configuration.
// It handles path resolution, directory creation, connection string building,
// connection pool configuration, and connection testing.
//
// Parameters:
//   - cfg: Database configuration (path, profile, name)
//
// Returns:
//   - *DB: Configured database connection
//   - error: Error if connection fails or configuration is invalid
func New(cfg Config) (*DB, error) {
	// Handle file: URIs (used for in-memory databases) - skip filepath operations
	// In-memory databases use "file::memory:?cache=shared" format
	if strings.HasPrefix(cfg.Path, "file:") {
		// For file: URIs, use as-is without filepath operations
		// This is used for in-memory databases in tests
	} else {
		// Ensure directory exists - resolve to absolute path to avoid relative path issues
		// This prevents problems when the working directory changes
		absPath, err := filepath.Abs(cfg.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve database path to absolute: %w", err)
		}
		dir := filepath.Dir(absPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
		// Use absolute path for database operations
		cfg.Path = absPath
	}

	// Default to standard profile if not specified
	if cfg.Profile == "" {
		cfg.Profile = ProfileStandard
	}

	// Build connection string with appropriate PRAGMAs based on profile
	// PRAGMAs are set via connection string parameters for immediate effect
	connStr := buildConnectionString(cfg.Path, cfg.Profile)

	// Open database connection using modernc.org/sqlite driver
	conn, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database %s: %w", cfg.Name, err)
	}

	// Configure connection pool for long-term operation
	// Sets max open/idle connections and connection lifetimes
	configureConnectionPool(conn, cfg.Profile)

	// Test connection with timeout to ensure database is accessible
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database %s: %w", cfg.Name, err)
	}

	db := &DB{
		conn:    conn,
		path:    cfg.Path,
		profile: cfg.Profile,
		name:    cfg.Name,
	}

	// Apply additional PRAGMAs that can't be set via connection string
	// Currently a no-op, but reserved for future runtime-only PRAGMAs
	if err := db.applyRuntimePragmas(); err != nil {
		return nil, fmt.Errorf("failed to apply runtime PRAGMAs for %s: %w", cfg.Name, err)
	}

	return db, nil
}

// getSchemaContent retrieves schema content from embedded files.
// This ensures schemas are always available regardless of where the binary is deployed.
// Schema files are embedded at compile time via //go:embed directive.
//
// Parameters:
//   - schemaFile: Name of schema file (e.g., "universe_schema.sql")
//
// Returns:
//   - []byte: Schema file content
//   - error: Error if file not found or read fails
func getSchemaContent(schemaFile string) ([]byte, error) {
	// Schema files are embedded in schemas/ directory
	path := "schemas/" + schemaFile
	content, err := schemaFiles.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded schema file %s: %w", schemaFile, err)
	}
	return content, nil
}

// buildConnectionString creates SQLite connection string with profile-specific PRAGMAs.
// PRAGMAs are set via connection string parameters using _pragma= syntax.
// This ensures optimal configuration for each database type (ledger, cache, standard).
//
// Parameters:
//   - path: Database file path
//   - profile: Database profile (ledger, cache, standard)
//
// Returns:
//   - string: Complete connection string with all PRAGMAs
func buildConnectionString(path string, profile DatabaseProfile) string {
	// Base connection string with WAL mode (all databases)
	// WAL (Write-Ahead Logging) provides better concurrency and performance
	connStr := path + "?_pragma=journal_mode(WAL)"

	// Profile-specific PRAGMAs
	switch profile {
	case ProfileLedger:
		// Maximum safety - audit trail for real money
		// FULL synchronous mode ensures every write is immediately persisted to disk
		// NONE auto-vacuum prevents database shrinking (append-only pattern)
		connStr += "&_pragma=synchronous(FULL)" // Fsync after every write
		connStr += "&_pragma=auto_vacuum(NONE)" // Never shrink (append-only)

	case ProfileCache:
		// Maximum speed - ephemeral data
		// OFF synchronous mode skips fsync for maximum performance
		// FULL auto-vacuum aggressively reclaims space
		// MEMORY temp_store keeps temporary tables in RAM
		connStr += "&_pragma=synchronous(OFF)"   // No fsync (it's cache!)
		connStr += "&_pragma=auto_vacuum(FULL)"  // Auto-reclaim space
		connStr += "&_pragma=temp_store(MEMORY)" // Temp tables in RAM

	case ProfileStandard:
		// Balanced - most databases
		// NORMAL synchronous mode fsyncs at checkpoints (good balance)
		// INCREMENTAL auto-vacuum gradually reclaims space
		// MEMORY temp_store keeps temporary tables in RAM
		connStr += "&_pragma=synchronous(NORMAL)"      // Fsync at checkpoints
		connStr += "&_pragma=auto_vacuum(INCREMENTAL)" // Gradual space reclamation
		connStr += "&_pragma=temp_store(MEMORY)"       // Temp tables in RAM
	}

	// Common PRAGMAs for all profiles
	connStr += "&_pragma=foreign_keys(1)"          // Enable foreign key constraints (data integrity)
	connStr += "&_pragma=wal_autocheckpoint(1000)" // Checkpoint every 1000 pages (WAL management)
	connStr += "&_pragma=cache_size(-64000)"       // 64MB cache (negative = KB, positive = pages)
	connStr += "&_pragma=busy_timeout(5000)"       // Wait up to 5 seconds if database is locked

	return connStr
}

// configureConnectionPool sets up connection pool for long-term operation.
// Connection pooling is essential for performance in a long-running application.
// Settings are tuned for embedded device operation (Arduino Uno Q).
//
// Parameters:
//   - conn: SQLite database connection
//   - profile: Database profile (cache gets fewer connections)
func configureConnectionPool(conn *sql.DB, profile DatabaseProfile) {
	// Connection pool limits for standard/cache databases
	// MaxOpenConns limits concurrent database operations
	// MaxIdleConns keeps connections warm to avoid connection overhead
	conn.SetMaxOpenConns(25) // Max concurrent connections
	conn.SetMaxIdleConns(5)  // Keep some connections warm

	// Connection lifecycle management (tuned for long-running embedded device)
	// Extended lifetimes prevent unnecessary reconnection during long operations
	// This is important for Arduino Uno Q which may have network interruptions
	conn.SetConnMaxLifetime(24 * time.Hour)   // Recycle connections after 24 hours
	conn.SetConnMaxIdleTime(30 * time.Minute) // Close idle connections after 30 minutes

	// Cache database can have fewer connections (less frequently accessed)
	// Reduces memory footprint for ephemeral data
	if profile == ProfileCache {
		conn.SetMaxOpenConns(10)
		conn.SetMaxIdleConns(2)
	}
}

// applyRuntimePragmas applies PRAGMAs that require a query execution.
// Most PRAGMAs can be set via connection string, but some require runtime execution.
// Currently a no-op, but reserved for future PRAGMAs that can't be set via connection string.
//
// Returns:
//   - error: Error if PRAGMA execution fails
func (db *DB) applyRuntimePragmas() error {
	// These PRAGMAs don't work via connection string, must be executed
	// Currently all critical PRAGMAs are handled via connection string
	// This method is here for future runtime-only PRAGMAs if needed
	return nil
}

// Close closes the database connection.
// Should be called during application shutdown to ensure proper cleanup.
//
// Returns:
//   - error: Error if close fails
func (db *DB) Close() error {
	return db.conn.Close()
}

// Conn returns the underlying sql.DB connection.
// Used by repositories to execute queries directly.
// This provides access to the standard database/sql interface.
//
// Returns:
//   - *sql.DB: Underlying database connection
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Name returns the database name for logging.
// Used in log messages to identify which database is being accessed.
//
// Returns:
//   - string: Database name (e.g., "universe", "ledger")
func (db *DB) Name() string {
	return db.name
}

// Profile returns the database profile.
// Indicates which configuration profile is active (ledger, cache, standard).
//
// Returns:
//   - DatabaseProfile: Active profile
func (db *DB) Profile() DatabaseProfile {
	return db.profile
}

// Path returns the database file path.
// Returns the absolute path to the database file.
//
// Returns:
//   - string: Absolute path to database file
func (db *DB) Path() string {
	return db.path
}

// Migrate applies the database schema from embedded schema files.
// This is the single source of truth for each database's schema.
// Schemas are embedded in the binary, ensuring they're always available.
//
// Migration is idempotent - if schema is already applied, it skips gracefully.
// This allows safe re-execution during application startup.
//
// Returns:
//   - error: Error if schema execution fails
func (db *DB) Migrate() error {
	// Map database names to their schema files
	// Each database has a corresponding schema file in schemas/ directory
	schemaFileMap := map[string]string{
		"universe":     "universe_schema.sql",
		"config":       "config_schema.sql",
		"ledger":       "ledger_schema.sql",
		"portfolio":    "portfolio_schema.sql",
		"history":      "history_schema.sql",
		"cache":        "cache_schema.sql",
		"client_data":  "client_data_schema.sql",
		"calculations": "calculations_schema.sql",
	}

	schemaFile, ok := schemaFileMap[db.name]
	if !ok {
		// Unknown database name, skip migration
		// This allows for test databases or custom databases without schemas
		return nil
	}

	// Read schema content from embedded files
	// Schema files are embedded at compile time via //go:embed
	content, err := getSchemaContent(schemaFile)
	if err != nil {
		return fmt.Errorf("failed to get schema content for %s: %w", db.name, err)
	}

	// Execute schema within a transaction for atomicity
	// If schema execution fails, transaction is rolled back
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for schema %s: %w", schemaFile, err)
	}

	if _, err := tx.Exec(string(content)); err != nil {
		_ = tx.Rollback()

		// If error indicates schema already applied, skip it
		// This makes migration idempotent - safe to run multiple times
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate column") ||
			strings.Contains(errStr, "already exists") {
			// Schema already applied, commit and continue
			_ = tx.Commit()
			return nil
		}

		return fmt.Errorf("failed to execute schema %s for %s: %w", schemaFile, db.name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit schema %s for %s: %w", schemaFile, db.name, err)
	}

	return nil
}

// Begin starts a new transaction.
// Returns a transaction handle that must be committed or rolled back.
//
// Returns:
//   - *sql.Tx: Transaction handle
//   - error: Error if transaction start fails
func (db *DB) Begin() (*sql.Tx, error) {
	return db.conn.Begin()
}

// BeginTx starts a new transaction with options.
// Allows specifying context and transaction options (isolation level, read-only).
//
// Parameters:
//   - ctx: Context for cancellation/timeout
//   - opts: Transaction options (nil for defaults)
//
// Returns:
//   - *sql.Tx: Transaction handle
//   - error: Error if transaction start fails
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.conn.BeginTx(ctx, opts)
}

// WithTransaction executes a function within a database transaction.
// It handles begin, commit, rollback, panic recovery, and error wrapping automatically.
// This is a convenience function that ensures proper transaction lifecycle management.
//
// Transaction lifecycle:
//   - If the function returns an error or panics, the transaction is rolled back
//   - If the function succeeds, the transaction is committed
//   - Panics are caught and converted to errors
//
// Parameters:
//   - db: Database connection
//   - fn: Function to execute within transaction (receives transaction handle)
//
// Returns:
//   - error: Error if transaction fails, function returns error, or panic occurs
func WithTransaction(db *sql.DB, fn func(*sql.Tx) error) (err error) {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer rollback with panic recovery
	// Use named return variable to capture panic value
	defer func() {
		if p := recover(); p != nil {
			// Panic occurred - rollback and convert panic to error
			_ = tx.Rollback()
			err = fmt.Errorf("panic in transaction: %v", p)
		} else if err != nil {
			// Function returned error - rollback
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = fmt.Errorf("transaction failed: %w (rollback also failed: %v)", err, rollbackErr)
			} else {
				err = fmt.Errorf("transaction failed: %w", err)
			}
		} else {
			// Function succeeded - commit
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("failed to commit transaction: %w", commitErr)
			}
		}
	}()

	// Execute function within transaction
	err = fn(tx)
	return err
}

// Exec executes a query without returning rows (INSERT, UPDATE, DELETE, etc.).
// This is a convenience wrapper around sql.DB.Exec.
//
// Parameters:
//   - query: SQL query string
//   - args: Query parameters
//
// Returns:
//   - sql.Result: Result with LastInsertId and RowsAffected
//   - error: Error if query execution fails
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

// ExecContext executes a query with context (for cancellation/timeout).
// This is a convenience wrapper around sql.DB.ExecContext.
//
// Parameters:
//   - ctx: Context for cancellation/timeout
//   - query: SQL query string
//   - args: Query parameters
//
// Returns:
//   - sql.Result: Result with LastInsertId and RowsAffected
//   - error: Error if query execution fails
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.conn.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows (SELECT).
// This is a convenience wrapper around sql.DB.Query.
// Caller must close the returned *sql.Rows.
//
// Parameters:
//   - query: SQL query string
//   - args: Query parameters
//
// Returns:
//   - *sql.Rows: Query results (must be closed by caller)
//   - error: Error if query execution fails
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// QueryContext executes a query with context (for cancellation/timeout).
// This is a convenience wrapper around sql.DB.QueryContext.
// Caller must close the returned *sql.Rows.
//
// Parameters:
//   - ctx: Context for cancellation/timeout
//   - query: SQL query string
//   - args: Query parameters
//
// Returns:
//   - *sql.Rows: Query results (must be closed by caller)
//   - error: Error if query execution fails
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns at most one row.
// This is a convenience wrapper around sql.DB.QueryRow.
// Use Scan() on the returned *sql.Row to extract values.
//
// Parameters:
//   - query: SQL query string
//   - args: Query parameters
//
// Returns:
//   - *sql.Row: Single row result (use Scan() to extract values)
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

// QueryRowContext executes a query with context (for cancellation/timeout).
// This is a convenience wrapper around sql.DB.QueryRowContext.
// Use Scan() on the returned *sql.Row to extract values.
//
// Parameters:
//   - ctx: Context for cancellation/timeout
//   - query: SQL query string
//   - args: Query parameters
//
// Returns:
//   - *sql.Row: Single row result (use Scan() to extract values)
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRowContext(ctx, query, args...)
}

// HealthCheck performs a comprehensive health check on the database.
// This includes connection ping and full integrity check.
// Integrity check can be expensive on large databases, so use QuickCheck for frequent checks.
//
// Parameters:
//   - ctx: Context for cancellation/timeout
//
// Returns:
//   - error: Error if health check fails (connection or integrity)
func (db *DB) HealthCheck(ctx context.Context) error {
	// 1. Test connection
	// Ping verifies the database is accessible and connection is alive
	if err := db.conn.PingContext(ctx); err != nil {
		return fmt.Errorf("ping failed for %s: %w", db.name, err)
	}

	// 2. Integrity check (comprehensive but expensive)
	// PRAGMA integrity_check verifies database file integrity
	// Returns "ok" if database is valid, error message otherwise
	var integrityResult string
	err := db.conn.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrityResult)
	if err != nil {
		return fmt.Errorf("integrity check query failed for %s: %w", db.name, err)
	}

	if integrityResult != "ok" {
		return fmt.Errorf("integrity check failed for %s: %s", db.name, integrityResult)
	}

	return nil
}

// QuickCheck performs a quick health check (just ping, no integrity check).
// Use this for frequent health checks where full integrity check is too expensive.
//
// Parameters:
//   - ctx: Context for cancellation/timeout
//
// Returns:
//   - error: Error if connection ping fails
func (db *DB) QuickCheck(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

// WALCheckpoint forces a WAL checkpoint to prevent WAL file bloat.
// WAL (Write-Ahead Log) files can grow large if not checkpointed regularly.
// This is typically called during maintenance windows.
//
// Modes:
//   - PASSIVE: Checkpoint if no other connection is using WAL
//   - FULL: Checkpoint even if other connections are using WAL (may block)
//   - RESTART: Like FULL, but also resets WAL file
//   - TRUNCATE: Like RESTART, but also truncates WAL file to minimal size (recommended)
//
// Parameters:
//   - mode: Checkpoint mode (defaults to "TRUNCATE" if empty)
//
// Returns:
//   - error: Error if checkpoint fails
func (db *DB) WALCheckpoint(mode string) error {
	// Modes: PASSIVE, FULL, RESTART, TRUNCATE
	// TRUNCATE is recommended for maintenance (resets WAL file to minimal size)
	if mode == "" {
		mode = "TRUNCATE"
	}

	query := fmt.Sprintf("PRAGMA wal_checkpoint(%s)", mode)
	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("WAL checkpoint failed for %s: %w", db.name, err)
	}

	return nil
}

// Vacuum runs VACUUM to reclaim space and reduce fragmentation.
// VACUUM rebuilds the database file, removing free pages and optimizing layout.
// This can be expensive on large databases and should only be run during maintenance windows.
//
// Note: VACUUM can take a long time on large databases and may lock the database.
// Consider running during low-traffic periods or maintenance windows.
//
// Returns:
//   - error: Error if vacuum fails
func (db *DB) Vacuum() error {
	// Note: VACUUM can be expensive on large databases
	// Should only be run during maintenance windows
	if _, err := db.conn.Exec("VACUUM"); err != nil {
		return fmt.Errorf("vacuum failed for %s: %w", db.name, err)
	}

	return nil
}

// Stats contains database statistics for monitoring and maintenance.
type Stats struct {
	SizeBytes     int64 // Database file size in bytes
	WALSizeBytes  int64 // WAL file size in bytes (may not exist if no writes)
	PageCount     int64 // Total number of pages in database
	PageSize      int64 // Page size in bytes (typically 4096)
	FreelistCount int64 // Number of free pages (available for reuse)
}

// GetStats retrieves database statistics for monitoring and maintenance.
// This includes file sizes, page counts, and free space information.
//
// Returns:
//   - *Stats: Database statistics
//   - error: Error if statistics retrieval fails
func (db *DB) GetStats() (*Stats, error) {
	stats := &Stats{}

	// Get file size from filesystem
	// If file doesn't exist (shouldn't happen), size remains 0
	if fileInfo, err := os.Stat(db.path); err == nil {
		stats.SizeBytes = fileInfo.Size()
	}

	// Get WAL file size from filesystem
	// WAL file may not exist if database hasn't been written to
	walPath := db.path + "-wal"
	if fileInfo, err := os.Stat(walPath); err == nil {
		stats.WALSizeBytes = fileInfo.Size()
	}

	// Get page count from database
	// Total number of pages currently allocated
	if err := db.conn.QueryRow("PRAGMA page_count").Scan(&stats.PageCount); err != nil {
		return nil, fmt.Errorf("failed to get page count: %w", err)
	}

	// Get page size from database
	// Page size is typically 4096 bytes, set at database creation
	if err := db.conn.QueryRow("PRAGMA page_size").Scan(&stats.PageSize); err != nil {
		return nil, fmt.Errorf("failed to get page size: %w", err)
	}

	// Get freelist count from database
	// Number of pages available for reuse (freed by DELETE operations)
	if err := db.conn.QueryRow("PRAGMA freelist_count").Scan(&stats.FreelistCount); err != nil {
		return nil, fmt.Errorf("failed to get freelist count: %w", err)
	}

	return stats, nil
}
