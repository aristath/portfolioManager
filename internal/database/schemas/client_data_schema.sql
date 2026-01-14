-- Client Data Cache Database Schema
-- Single source of truth for client_data.db
-- Stores cached data with expiration timestamps
-- Note: External API tables (Alpha Vantage, Yahoo, OpenFIGI) removed - only Tradernet is used as data source

-- ExchangeRate table (keyed by currency pair, e.g., "EUR:USD")
CREATE TABLE IF NOT EXISTS exchangerate (
    pair TEXT PRIMARY KEY,
    data TEXT NOT NULL,
    expires_at INTEGER NOT NULL
);

-- Current prices cache (short TTL - 10 minutes)
CREATE TABLE IF NOT EXISTS current_prices (
    isin TEXT PRIMARY KEY,
    data TEXT NOT NULL,
    expires_at INTEGER NOT NULL
);

-- Symbol to ISIN mapping cache (long TTL - mappings rarely change)
CREATE TABLE IF NOT EXISTS symbol_to_isin (
    symbol TEXT PRIMARY KEY,
    data TEXT NOT NULL,
    expires_at INTEGER NOT NULL
);

-- Indexes for expiration queries (cleanup, freshness checks)
CREATE INDEX IF NOT EXISTS idx_exchangerate_expires ON exchangerate(expires_at);
CREATE INDEX IF NOT EXISTS idx_prices_expires ON current_prices(expires_at);
CREATE INDEX IF NOT EXISTS idx_symbol_to_isin_expires ON symbol_to_isin(expires_at);
