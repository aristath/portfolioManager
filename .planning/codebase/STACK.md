# Technology Stack

**Analysis Date:** 2026-01-20

## Languages

**Primary:**
- Go 1.24.0 - Backend application, all business logic, API server
- JavaScript (ES6+) - Frontend application (React)

**Secondary:**
- C++ (Arduino) - LED display firmware (`display/sketch/sketch.ino`)
- SQL - Database schemas and migrations

## Runtime

**Environment:**
- Go 1.24.0 (ARM64 for production on Arduino Uno Q, any platform for development)

**Package Manager:**
- Go modules (go.mod)
- npm (frontend/package.json)
- Lockfile: go.sum present, package-lock.json expected

**Platform:**
- Production: Arduino Uno Q (ARM64 Linux)
- Development: Any platform supporting Go 1.24+

## Frameworks

**Core:**
- Chi v5.0.11 - HTTP router (stdlib-based, lightweight)
- React 18.2.0 - Frontend UI framework
- Vite 5.0.0 - Frontend build tool and dev server

**UI Components:**
- Mantine v7.0.0 - React component library (core, hooks, dates, form, notifications)
- Tabler Icons v3.0.0 - Icon set for React
- lightweight-charts v4.1.0 - Financial charting library
- zustand v4.4.0 - State management

**Testing:**
- testify v1.11.1 - Go test assertions and mocking
- vitest v1.0.0 - Frontend unit testing framework
- @testing-library/react v14.1.0 - React component testing utilities
- @testing-library/jest-dom v6.1.0 - DOM matchers for tests
- jsdom v23.0.0 - DOM implementation for Node.js testing

**Build/Dev:**
- Vite v5.0.0 - Frontend bundler with HMR
- @vitejs/plugin-react v4.2.0 - React plugin for Vite
- golangci-lint - Go linting (configured via `.golangci.yml`)
- Lefthook v1.5.0+ - Git hooks (configured via `lefthook.yml`, no stashing)
- gofmt - Go code formatting
- goimports - Go import formatting
- ESLint v8.0.0 - JavaScript/React linting

## Key Dependencies

**Critical:**
- modernc.org/sqlite v1.28.0 - Pure Go SQLite driver (no CGo, preferred for production)
- mattn/go-sqlite3 v1.14.16 - CGo-based SQLite driver (alternative/compatibility)
- github.com/aws/aws-sdk-go-v2 v1.41.0 - AWS SDK for Cloudflare R2 backups
- github.com/rs/zerolog v1.31.0 - Structured logging (high-performance)
- nhooyr.io/websocket v1.8.11 - WebSocket client for Tradernet streaming

**Infrastructure:**
- github.com/joho/godotenv v1.5.1 - Environment variable loading from .env files
- github.com/go-chi/cors v1.2.1 - CORS middleware for HTTP server
- github.com/google/uuid v1.3.0 - UUID generation
- github.com/vmihailenco/msgpack/v5 v5.4.1 - MessagePack serialization
- github.com/shirou/gopsutil/v3 v3.24.5 - System monitoring (CPU, memory, disk)

**Financial/Analytics:**
- gonum.org/v1/gonum v0.16.0 - Numerical computing (portfolio optimization, statistics)
- github.com/markcheno/go-talib v0.0.0-20250114000313-ec55a20c902f - Technical analysis indicators

**AWS SDK Modules:**
- github.com/aws/aws-sdk-go-v2/config v1.32.6 - SDK configuration
- github.com/aws/aws-sdk-go-v2/credentials v1.19.6 - Credentials provider
- github.com/aws/aws-sdk-go-v2/service/s3 v1.95.0 - S3 client (R2-compatible)
- github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.20.18 - Multipart upload/download

## Configuration

**Environment:**
- Configuration loaded from `.env` file (optional, via godotenv)
- CLI flags override environment variables (`--data-dir`)
- Settings database (`config.db`) takes precedence over environment variables
- Configuration managed via `internal/config/config.go`

**Key configs required:**
- `TRADER_DATA_DIR` - Database directory (default: `/home/arduino/data`)
- `GO_PORT` - HTTP server port (default: 8001)
- `TRADERNET_API_KEY` - Broker API key (can be set via Settings UI)
- `TRADERNET_API_SECRET` - Broker API secret (can be set via Settings UI)
- `GITHUB_TOKEN` - GitHub token for artifact downloads (can be set via Settings UI)
- `LOG_LEVEL` - Log level (default: info)
- `DEV_MODE` - Development mode flag (default: false)

**Build:**
- `vite.config.js` - Frontend build configuration
- `.golangci.yml` - Go linting rules
- `lefthook.yml` - Git hooks configuration

## Platform Requirements

**Development:**
- Go 1.24.0+
- Node.js (for frontend development)
- npm (for frontend package management)
- golangci-lint (optional, for linting)
- Lefthook (optional, for git hooks)

**Production:**
- Arduino Uno Q (ARM64 Linux)
- Systemd (for service management)
- SQLite 3 (embedded, provided by modernc.org/sqlite)
- Serial port access for LED display (`/dev/ttyACM0`)

**Build Commands:**
```bash
# Development
go mod download
air                              # Auto-reload dev server (requires air)
cd frontend && npm install && npm run dev

# Production build
GOOS=linux GOARCH=arm64 go build -o sentinel-arm64 ./cmd/server
cd frontend && npm run build     # Output to frontend/dist/
```

---

*Stack analysis: 2026-01-20*
