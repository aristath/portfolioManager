# External Integrations

**Analysis Date:** 2026-01-20

## APIs & External Services

**Broker:**
- Tradernet (Freedom24) - Primary broker for trading and market data
  - SDK/Client: Internal Go SDK (`internal/clients/tradernet`)
  - Auth: `TRADERNET_API_KEY`, `TRADERNET_API_SECRET` (env or Settings UI)
  - Base URL: `https://api.tradernet.com`
  - WebSocket: Real-time market data streaming (`nhooyr.io/websocket`)
  - Adapter: `internal/clients/tradernet/adapter.go` implements `domain.BrokerClient`
  - Capabilities: Portfolio sync, order placement, trade history, quotes, historical prices
  - Rate limiting: Implemented in SDK client

**Currency Exchange:**
- Tradernet FX Pairs - Exchange rates via broker FX instruments
  - Implementation: `internal/services/currency_exchange_service.go`
  - Supported: EUR, USD, GBP, HKD conversions
  - Caching: 5-minute in-memory cache + database fallback
  - Fallback: Cached rates from `history.db` when API unavailable

**GitHub Actions:**
- GitHub API - Automated deployment via artifact downloads
  - Auth: `GITHUB_TOKEN` (env or Settings UI)
  - Repository: `aristath/sentinel`
  - Workflow: `build-go.yml`
  - Artifact: `sentinel-arm64`
  - Implementation: `internal/deployment`
  - Purpose: Download pre-built ARM64 binaries for deployment

## Data Storage

**Databases:**
- SQLite (7-database architecture)
  - Connection: Local filesystem via `TRADER_DATA_DIR` (default: `/home/arduino/data`)
  - Client: modernc.org/sqlite v1.28.0 (pure Go, no CGo)
  - Alternative: mattn/go-sqlite3 v1.14.16 (CGo-based, for compatibility)
  - WAL mode enabled for all databases
  - Databases:
    - `universe.db` - Investment universe (securities, groups)
    - `config.db` - Application configuration (settings, allocation targets)
    - `ledger.db` - Immutable financial audit trail (trades, cash flows, dividends)
    - `portfolio.db` - Current portfolio state (positions, scores, metrics, snapshots)
    - `history.db` - Historical time-series data (prices, rates, cleanup tracking)
    - `cache.db` - Ephemeral operational data (job history)
    - `client_data.db` - Cache for exchange rates and current prices
    - `calculations.db` - Portfolio calculations and optimization results

**File Storage:**
- Cloudflare R2 - Cloud backup storage (S3-compatible)
  - Client: AWS SDK for Go v2 (`github.com/aws/aws-sdk-go-v2/service/s3`)
  - Implementation: `internal/reliability/r2_client.go`
  - Auth: R2 credentials (account ID, access key, secret key)
  - Endpoint: `https://{account_id}.r2.cloudflarestorage.com`
  - Purpose: Tiered database backups (hourly/daily/weekly/monthly)
- Local filesystem - Database backups
  - Location: `{data_dir}/backups/`
  - Retention: Hourly (24h), Daily (30d), Weekly (12w), Monthly (12m)
  - Implementation: `internal/reliability/backup_service.go`

**Caching:**
- In-memory - Currency exchange rates (5-minute TTL)
- SQLite (`client_data.db`) - Exchange rates and price cache
- SQLite (`cache.db`) - Job execution history

## Authentication & Identity

**Auth Provider:**
- Custom - Tradernet API key/secret authentication
  - Implementation: HTTP basic auth in `internal/clients/tradernet/sdk/client.go`
  - Credentials: API key and secret from Settings UI or environment variables
  - No OAuth/JWT - direct API key authentication

**Settings Management:**
- Settings Database (`config.db`) - Credential storage
  - Implementation: `internal/modules/settings`
  - UI: Settings tab in frontend
  - API: `/api/settings` endpoints
  - Priority: Settings DB overrides environment variables

## Monitoring & Observability

**Error Tracking:**
- zerolog - Structured logging to stdout/stderr
  - Implementation: `pkg/logger`
  - Format: JSON structured logs
  - Levels: debug, info, warn, error, fatal
  - Configuration: `LOG_LEVEL` environment variable

**Logs:**
- Console output - Structured JSON logs via zerolog
- No external log aggregation service
- System monitoring: CPU, memory, disk via gopsutil (`github.com/shirou/gopsutil/v3`)

**Metrics:**
- Internal - System health endpoints
  - Implementation: `internal/reliability`
  - Endpoints: Health checks, database integrity, backup status
  - LED Display: Visual status indicators on Arduino Uno Q

## CI/CD & Deployment

**Hosting:**
- Self-hosted on Arduino Uno Q (ARM64 Linux)
- Systemd service management (`sentinel.service`)

**CI Pipeline:**
- GitHub Actions - Automated builds
  - Workflow: `.github/workflows/build-go.yml`
  - Output: ARM64 binary artifact (`sentinel-arm64`)
  - Deployment: Automated via `internal/deployment` manager
  - Trigger: Git push to monitored branch

**Deployment:**
- Automated deployment system (`internal/deployment`)
  - Monitors: GitHub Actions for new builds
  - Downloads: Pre-built ARM64 binaries from GitHub artifacts
  - Deploys: Go service, frontend, display app, Arduino sketch
  - Restarts: Systemd service automatically
  - Health checks: Verify deployment success, rollback on failure
  - Lock timeout: 120 seconds (prevents concurrent deployments)

## Environment Configuration

**Required env vars:**
- `TRADER_DATA_DIR` - Database directory path
- `TRADERNET_API_KEY` - Tradernet API key (or via Settings UI)
- `TRADERNET_API_SECRET` - Tradernet API secret (or via Settings UI)
- `GITHUB_TOKEN` - GitHub token for artifact downloads (or via Settings UI)

**Optional env vars:**
- `GO_PORT` - HTTP server port (default: 8001)
- `DEV_MODE` - Development mode flag (default: false)
- `LOG_LEVEL` - Log level (default: info)
- `GITHUB_WORKFLOW_NAME` - GitHub Actions workflow name (default: build-go.yml)
- `GITHUB_ARTIFACT_NAME` - GitHub Actions artifact name (default: sentinel-arm64)
- `GITHUB_BRANCH` - Branch to deploy from (default: auto-detect)
- `GITHUB_REPO` - GitHub repository (default: aristath/sentinel)

**Secrets location:**
- Settings database (`config.db`) - Preferred for credentials (Tradernet API, GitHub token)
- `.env` file - Fallback for infrastructure settings (not committed)
- Environment variables - Lowest priority

## Webhooks & Callbacks

**Incoming:**
- None - No webhooks from external services
- HTTP API endpoints for manual triggers (planning, rebalancing, etc.)

**Outgoing:**
- None - No webhooks sent to external services
- All integrations are pull-based (poll broker API, GitHub API)

## Hardware Integration

**LED Display:**
- Arduino Uno Q - Serial communication for status display
  - Protocol: Serial UART (`/dev/ttyACM0`, 115200 baud)
  - Implementation: `display/sketch/sketch.ino` (Arduino C++)
  - Management: `internal/modules/display` (state manager)
  - Purpose: Visual status indicators (service health, portfolio status, planner actions)

---

*Integration audit: 2026-01-20*
