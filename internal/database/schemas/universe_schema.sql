-- Universe Database Schema
-- Single source of truth for universe.db
-- This schema represents the final state after all migrations

-- Securities table: investment universe (ISIN as PRIMARY KEY)
-- Note: Legacy columns yahoo_symbol and alphavantage_symbol removed.
-- Client-specific symbols are now stored in client_symbols table.
CREATE TABLE IF NOT EXISTS securities (
    isin TEXT PRIMARY KEY,
    symbol TEXT NOT NULL,
    name TEXT NOT NULL,
    product_type TEXT,
    industry TEXT,
    country TEXT,
    fullExchangeName TEXT,
    priority_multiplier REAL DEFAULT 1.0,
    min_lot INTEGER DEFAULT 1,
    active INTEGER DEFAULT 1,  -- Boolean: 1 = active, 0 = inactive (soft delete)
    allow_buy INTEGER DEFAULT 1,
    allow_sell INTEGER DEFAULT 1,
    currency TEXT,
    last_synced INTEGER,               -- Unix timestamp (seconds since epoch)
    min_portfolio_target REAL,
    max_portfolio_target REAL,
    created_at INTEGER NOT NULL,      -- Unix timestamp (seconds since epoch)
    updated_at INTEGER NOT NULL      -- Unix timestamp (seconds since epoch)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_securities_active ON securities(active);
CREATE INDEX IF NOT EXISTS idx_securities_country ON securities(country);
CREATE INDEX IF NOT EXISTS idx_securities_industry ON securities(industry);
CREATE INDEX IF NOT EXISTS idx_securities_symbol ON securities(symbol);

-- Country groups: custom groupings for allocation strategies
CREATE TABLE IF NOT EXISTS country_groups (
    group_name TEXT NOT NULL,
    country_name TEXT NOT NULL,  -- '__EMPTY__' is special marker for empty groups
    created_at INTEGER NOT NULL,      -- Unix timestamp (seconds since epoch)
    updated_at INTEGER NOT NULL,     -- Unix timestamp (seconds since epoch)
    PRIMARY KEY (group_name, country_name)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_country_groups_group ON country_groups(group_name);

-- Industry groups: custom groupings for diversification strategies
CREATE TABLE IF NOT EXISTS industry_groups (
    group_name TEXT NOT NULL,
    industry_name TEXT NOT NULL,  -- '__EMPTY__' is special marker for empty groups
    created_at INTEGER NOT NULL,      -- Unix timestamp (seconds since epoch)
    updated_at INTEGER NOT NULL,     -- Unix timestamp (seconds since epoch)
    PRIMARY KEY (group_name, industry_name)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_industry_groups_group ON industry_groups(group_name);

-- Tags table: tag definitions with ID and human-readable name
CREATE TABLE IF NOT EXISTS tags (
    id TEXT PRIMARY KEY,  -- e.g., 'value-opportunity', 'volatile', 'stable'
    name TEXT NOT NULL,   -- e.g., 'Value Opportunity', 'Volatile', 'Stable'
    created_at INTEGER NOT NULL,      -- Unix timestamp (seconds since epoch)
    updated_at INTEGER NOT NULL       -- Unix timestamp (seconds since epoch)
) STRICT;

-- Security tags junction table: links securities to tags (many-to-many, ISIN-based)
CREATE TABLE IF NOT EXISTS security_tags (
    isin TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    created_at INTEGER NOT NULL,      -- Unix timestamp (seconds since epoch)
    updated_at INTEGER NOT NULL,     -- Unix timestamp (seconds since epoch)
    PRIMARY KEY (isin, tag_id),
    FOREIGN KEY (isin) REFERENCES securities(isin) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
) STRICT;

CREATE INDEX IF NOT EXISTS idx_security_tags_isin ON security_tags(isin);
CREATE INDEX IF NOT EXISTS idx_security_tags_tag_id ON security_tags(tag_id);

-- Insert default tags (from migrations 028 and 032)
-- Quality Gate Tags
INSERT OR IGNORE INTO tags (id, name, created_at, updated_at) VALUES
    ('quality-gate-pass', 'Quality Gate Pass', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('quality-gate-fail', 'Quality Gate Fail', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('quality-value', 'Quality Value', (strftime('%s', 'now')), (strftime('%s', 'now')));

-- Bubble Detection Tags
INSERT OR IGNORE INTO tags (id, name, created_at, updated_at) VALUES
    ('bubble-risk', 'Bubble Risk', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('quality-high-cagr', 'Quality High CAGR', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('poor-risk-adjusted', 'Poor Risk-Adjusted', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('high-sharpe', 'High Sharpe', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('high-sortino', 'High Sortino', (strftime('%s', 'now')), (strftime('%s', 'now')));

-- Value Trap Tags
INSERT OR IGNORE INTO tags (id, name, created_at, updated_at) VALUES
    ('value-trap', 'Value Trap', (strftime('%s', 'now')), (strftime('%s', 'now')));

-- Total Return Tags
INSERT OR IGNORE INTO tags (id, name, created_at, updated_at) VALUES
    ('high-total-return', 'High Total Return', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('excellent-total-return', 'Excellent Total Return', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('dividend-total-return', 'Dividend Total Return', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('moderate-total-return', 'Moderate Total Return', (strftime('%s', 'now')), (strftime('%s', 'now')));

-- Optimizer Alignment Tags
INSERT OR IGNORE INTO tags (id, name, created_at, updated_at) VALUES
    ('underweight', 'Underweight', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('target-aligned', 'Target Aligned', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('needs-rebalance', 'Needs Rebalance', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('slightly-overweight', 'Slightly Overweight', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('slightly-underweight', 'Slightly Underweight', (strftime('%s', 'now')), (strftime('%s', 'now')));

-- Regime-Specific Tags
INSERT OR IGNORE INTO tags (id, name, created_at, updated_at) VALUES
    ('regime-bear-safe', 'Bear Market Safe', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('regime-bull-growth', 'Bull Market Growth', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('regime-sideways-value', 'Sideways Value', (strftime('%s', 'now')), (strftime('%s', 'now'))),
    ('regime-volatile', 'Regime Volatile', (strftime('%s', 'now')), (strftime('%s', 'now')));

-- Client symbols table: maps ISINs to client-specific symbol formats
-- Used for brokers (tradernet, ibkr, schwab) and data providers (yahoo, alphavantage, etc.)
CREATE TABLE IF NOT EXISTS client_symbols (
    isin TEXT NOT NULL,
    client_name TEXT NOT NULL,  -- e.g., 'tradernet', 'ibkr', 'yahoo', 'alphavantage'
    client_symbol TEXT NOT NULL,  -- The client-specific symbol format
    PRIMARY KEY (isin, client_name),
    FOREIGN KEY (isin) REFERENCES securities(isin) ON DELETE CASCADE
) STRICT;

CREATE INDEX IF NOT EXISTS idx_client_symbols_isin ON client_symbols(isin);
CREATE INDEX IF NOT EXISTS idx_client_symbols_client ON client_symbols(client_name);
CREATE INDEX IF NOT EXISTS idx_client_symbols_symbol ON client_symbols(client_symbol);
