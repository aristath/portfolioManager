# Codebase Structure

**Analysis Date:** 2026-01-20

## Directory Layout

```
sentinel/
├── cmd/                    # Application entry points
│   └── server/             # Main HTTP server application
├── internal/               # Private application code (not importable by external packages)
│   ├── clients/            # External API clients
│   │   └── tradernet/      # Tradernet broker integration (SDK + adapter)
│   ├── config/             # Configuration management
│   ├── database/           # Database connection and schema management
│   ├── deployment/         # Automated deployment system
│   ├── di/                 # Dependency injection container
│   ├── domain/             # Pure domain models and interfaces
│   ├── events/             # Event bus system (pub/sub)
│   ├── evaluation/         # Sequence evaluation (worker pool)
│   ├── market_regime/      # Market regime detection
│   ├── modules/            # Business modules (30+ modules)
│   ├── reliability/        # Backup and health check services
│   ├── server/             # HTTP server and routing
│   ├── services/           # Core cross-cutting services
│   ├── testing/            # Test utilities and fixtures
│   ├── ticker/             # LED ticker content generation
│   ├── utils/              # Utility functions
│   ├── version/            # Version information (auto-generated)
│   └── work/               # Background job processor
├── pkg/                    # Public reusable packages (importable by external code)
│   ├── embedded/           # Embedded frontend assets
│   ├── formulas/           # Financial formulas (HRP, Sharpe, CVaR, etc.)
│   └── logger/             # Structured logging
├── frontend/               # React frontend (Vite)
│   ├── src/                # Frontend source code
│   └── dist/               # Built frontend assets (embedded in Go binary)
├── display/                # LED display system
│   ├── sketch/             # Arduino C++ sketch
│   └── app/                # Python display app (Arduino App Framework)
├── scripts/                # Utility scripts
├── data/                   # Database files (runtime, not committed)
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── lefthook.yml            # Git hooks configuration (Lefthook)
├── CLAUDE.md               # Project-specific Claude instructions
└── README.md               # Main documentation
```

## Directory Purposes

**cmd/server:**
- Purpose: Application entry point (main.go)
- Contains: Main function, startup orchestration, shutdown handling
- Key files: `main.go` - wires dependencies, starts server and monitors

**internal/clients/tradernet:**
- Purpose: Tradernet broker integration
- Contains: SDK client (`sdk/`), adapter (implements domain.BrokerClient), transformers (broker → domain types), WebSocket client
- Key files: `adapter.go` (broker abstraction), `transformers_domain.go` (type mapping), `sdk_client.go` (Tradernet SDK), `websocket_client.go` (market status WebSocket)

**internal/config:**
- Purpose: Configuration management
- Contains: Config struct, environment variable loading, settings database integration
- Key files: `config.go` - two-tier config (env vars + settings DB)

**internal/database:**
- Purpose: SQLite database management
- Contains: Database connection wrapper, schema initialization, profile-based PRAGMA settings, transaction support
- Key files: `db.go` (connection management), `schemas/*.sql` (embedded schemas), `migrations/` (schema migrations)

**internal/deployment:**
- Purpose: Automated deployment system
- Contains: Deployment manager, git monitoring, binary downloads, service restarts
- Key files: `manager.go` - orchestrates deployments from GitHub Actions

**internal/di:**
- Purpose: Dependency injection container
- Contains: Container definition, initialization functions (databases, repositories, services, work processor)
- Key files: `wire.go` (orchestration), `types.go` (Container struct), `databases.go`, `repositories.go`, `services.go` (initialization steps)

**internal/domain:**
- Purpose: Pure domain models and interfaces
- Contains: Broker-agnostic types, core interfaces (no infrastructure dependencies)
- Key files: `broker_types.go` (BrokerPosition, BrokerTrade, etc.), `interfaces.go` (BrokerClient, CashManager)

**internal/events:**
- Purpose: Event bus system for decoupled communication
- Contains: Event bus (pub/sub), event manager (high-level API), typed event data
- Key files: `manager.go` (EventManager), `types.go` (event type definitions), `bus.go` (pub/sub implementation)

**internal/evaluation:**
- Purpose: Sequence evaluation with worker pool
- Contains: Evaluation service, worker pool, evaluation models
- Key files: `service.go` (evaluation orchestration), `workers/` (worker pool implementation)

**internal/market_regime:**
- Purpose: Market regime detection
- Contains: Market state detector, regime persistence, index management, correlation analysis
- Key files: `detector.go` (regime detection), `persistence.go` (regime history), `index_service.go` (market index sync)

**internal/modules:**
- Purpose: Business modules (30+ modules, one per domain area)
- Contains: Module-specific repositories, services, handlers
- Structure: Each module has `repository.go`, `service.go`, `handlers/` subdirectory
- Key modules:
  - `adaptation/` - Adaptive Market Hypothesis
  - `allocation/` - Allocation targets and alerts
  - `analytics/` - Factor exposure analytics
  - `calculations/` - Calculation cache
  - `cash_flows/` - Cash flow processing (deposits, withdrawals)
  - `charts/` - Chart data and visualization
  - `currency/` - Currency exchange
  - `display/` - LED display management
  - `dividends/` - Dividend processing
  - `evaluation/` - Sequence evaluation
  - `market_hours/` - Market hours and holidays
  - `opportunities/` - Opportunity identification
  - `optimization/` - Portfolio optimization (HRP, MV, BL)
  - `planning/` - Planning and recommendations
  - `portfolio/` - Portfolio management
  - `quantum/` - Quantum probability models
  - `rebalancing/` - Rebalancing logic
  - `scoring/` - Security scoring
  - `sequences/` - Trade sequence generation
  - `settings/` - Settings management
  - `trading/` - Trade execution
  - `universe/` - Security universe management

**internal/modules/planning:**
- Purpose: Core planning system (recommendation generation)
- Contains: Planner service, state monitor, universe monitor, evaluation service, repositories
- Key files: `planner/planner.go` (sequence generation), `state_monitor/service.go` (state change detection), `repository/in_memory_planner_repository.go` (in-memory storage)

**internal/reliability:**
- Purpose: Backup and health check services
- Contains: Local backup service, R2 cloud backup, restore service, database health checks
- Key files: `backup_service.go` (local backups), `r2_service.go` (cloud backups), `restore_service.go` (database restore), `health_service.go` (health checks)

**internal/server:**
- Purpose: HTTP server and routing
- Contains: Server setup, middleware, route registration, handler factories
- Key files: `server.go` (main server), `routes.go` (route registration), `system_handlers.go` (system endpoints)

**internal/services:**
- Purpose: Core cross-cutting services
- Contains: Currency exchange, price conversion, security service, trade execution
- Key files: `currency_exchange_service.go`, `trade_execution_service.go`, `security_service.go`

**internal/work:**
- Purpose: Background job processor
- Contains: Work processor (job queue), work registry (job types), market timing checker, completion cache
- Key files: `processor.go` (job execution), `registry.go` (job registration), `market_timing.go` (market-aware scheduling), `cache.go` (completion tracking)

**pkg/embedded:**
- Purpose: Embedded frontend assets
- Contains: Embedded filesystem with built frontend files
- Key files: `embedded.go` - embeds `frontend/dist/` into Go binary

**pkg/formulas:**
- Purpose: Financial formulas and calculations
- Contains: HRP, Sharpe ratio, CVaR, covariance, correlation, etc.
- Key files: `hrp.go`, `sharpe.go`, `cvar.go`, `covariance.go`

**pkg/logger:**
- Purpose: Structured logging wrapper
- Contains: Logger configuration, zerolog wrapper
- Key files: `logger.go` - zerolog wrapper with config support

**frontend/src:**
- Purpose: React frontend source code
- Contains: Components, views, stores (state management), API client, hooks
- Structure: `components/` (reusable UI), `views/` (page views), `stores/` (state), `api/` (backend client)

**display/sketch:**
- Purpose: Arduino C++ sketch for LED display
- Contains: Arduino sketch code for Uno Q
- Key files: Arduino .ino files for LED control

**display/app:**
- Purpose: Python display app (Arduino App Framework)
- Contains: Python app for display content updates
- Key files: Python scripts for display management

## Key File Locations

**Entry Points:**
- `cmd/server/main.go`: Main application entry point

**Configuration:**
- `.env`: Infrastructure settings (data directory, port, etc.)
- `config.db`: Application settings (credentials, allocation targets)
- `lefthook.yml`: Git hooks configuration

**Core Logic:**
- `internal/di/wire.go`: Dependency injection orchestration
- `internal/server/server.go`: HTTP server setup and routing
- `internal/work/processor.go`: Background job processor
- `internal/events/manager.go`: Event bus management

**Testing:**
- `*_test.go`: Unit and integration tests (co-located with source)
- `internal/testing/`: Test utilities and fixtures

**Database Schemas:**
- `internal/database/schemas/*.sql`: Embedded SQL schemas

**Frontend:**
- `frontend/dist/index.html`: Frontend entry point (embedded)
- `frontend/src/App.tsx`: Frontend React app root

## Naming Conventions

**Files:**
- `*_service.go`: Service implementations (business logic)
- `*_repository.go`: Repository implementations (data access)
- `*_test.go`: Test files (co-located with source)
- `handlers/*.go`: HTTP request handlers
- `types.go`: Type definitions for package
- `models.go`: Domain models for package

**Directories:**
- `internal/modules/<module>/`: One directory per business module
- `internal/modules/<module>/handlers/`: HTTP handlers for module
- `pkg/<package>/`: Public reusable packages
- `cmd/<app>/`: Application entry points

**Go Packages:**
- Package name matches directory name (e.g., `package portfolio` in `internal/modules/portfolio/`)
- Handlers use aliased imports when multiple modules have same package name: `import portfoliohandlers "github.com/aristath/sentinel/internal/modules/portfolio/handlers"`

**Variables:**
- Services: `<name>Service` (e.g., `portfolioService`, `tradingService`)
- Repositories: `<name>Repo` (e.g., `positionRepo`, `securityRepo`)
- Handlers: `<name>Handler` (e.g., `portfolioHandler`, `tradingHandler`)

## Where to Add New Code

**New Business Module:**
- Implementation: `internal/modules/<module_name>/`
- Structure: Create `repository.go`, `service.go`, `handlers/handler.go`
- Registration: Add to `internal/di/repositories.go`, `internal/di/services.go`, register routes in `internal/server/server.go`
- Container: Add service to `internal/di/types.go` (Container struct)

**New API Endpoint:**
- Handler: `internal/modules/<module>/handlers/handler.go`
- Route registration: `internal/server/server.go` (in `setupRoutes()`)
- Pattern: Handlers implement `RegisterRoutes(r chi.Router)` method

**New Background Job:**
- Job type: Define in `internal/work/types.go` or module-specific file
- Job registration: `internal/di/work.go` (in `RegisterWork()`)
- Job implementation: Module service method (called by work processor)

**New Event Type:**
- Type definition: `internal/events/types.go`
- Event data struct: `internal/events/types.go` (implement EventData interface)
- Emit event: `eventManager.Emit(eventType, eventData)`
- Listen to event: Work processor registers listener in `internal/di/work.go`

**New Domain Type:**
- Pure domain model: `internal/domain/<type>_types.go`
- Ensure no infrastructure dependencies (no database, no API clients)

**New Repository:**
- Implementation: `internal/modules/<module>/<entity>_repository.go`
- Interface (if needed): `internal/domain/interfaces.go` (for breaking circular dependencies)
- Registration: `internal/di/repositories.go` (in `InitializeRepositories()`)
- Container: Add to `internal/di/types.go` (Container struct)

**New Service:**
- Implementation: `internal/modules/<module>/<purpose>_service.go` or `internal/services/<service>.go`
- Registration: `internal/di/services.go` (in appropriate initialization step)
- Container: Add to `internal/di/types.go` (Container struct)
- Dependencies: Inject via constructor (all dependencies from container)

**Utilities:**
- Shared helpers: `internal/utils/` (for internal use) or `pkg/` (for external use)
- Financial formulas: `pkg/formulas/`

**Frontend Component:**
- Component: `frontend/src/components/<category>/<Component>.tsx`
- View (page): `frontend/src/views/<ViewName>.tsx`
- API client: `frontend/src/api/<module>.ts`

## Special Directories

**data/**
- Purpose: Database files (runtime)
- Generated: Yes (by application at runtime)
- Committed: No (in .gitignore)
- Contains: `universe.db`, `config.db`, `ledger.db`, `portfolio.db`, `history.db`, `cache.db`, `client_data.db`, `calculations.db`

**frontend/dist/**
- Purpose: Built frontend assets
- Generated: Yes (by Vite during build)
- Committed: Yes (embedded in Go binary)
- Contains: `index.html`, `assets/*.js`, `assets/*.css`

**internal/database/schemas/**
- Purpose: SQL schema files
- Generated: No (manually maintained)
- Committed: Yes (embedded in Go binary)
- Contains: `*.sql` schema definitions for each database

**internal/version/**
- Purpose: Version information
- Generated: Yes (by git hook during commit)
- Committed: Yes
- Contains: `version.go` with build metadata

**migrations/**
- Purpose: Database schema migrations (archived)
- Generated: No
- Committed: Yes
- Contains: Historical migrations (no longer used, schemas are now in `internal/database/schemas/`)

**scripts/**
- Purpose: Utility scripts for development and operations
- Generated: No
- Committed: Yes
- Contains: Database management, deployment, testing scripts

## Module Organization Pattern

Each business module follows this structure:

```
internal/modules/<module>/
├── repository.go          # Data access (implements repository interface)
├── service.go             # Business logic (coordinates repositories)
├── models.go              # Domain models specific to module
├── types.go               # Type definitions
├── handlers/              # HTTP handlers
│   └── handler.go         # HTTP endpoint handlers
└── <submodules>/          # Sub-modules if complex
```

**Example: Portfolio Module**

```
internal/modules/portfolio/
├── position_repository.go  # Position data access
├── service.go              # Portfolio business logic
├── cash_manager.go         # Cash balance management (implements domain.CashManager)
├── types.go                # Portfolio type definitions
└── handlers/               # HTTP handlers
    └── handler.go          # Portfolio API endpoints
```

**Example: Planning Module (Complex)**

```
internal/modules/planning/
├── repository/             # Repository implementations
│   ├── config_repository.go         # Planner config repo
│   └── in_memory_planner_repository.go  # In-memory plan storage
├── planner/                # Core planner
│   └── planner.go          # Sequence generation
├── state_monitor/          # State change detection
│   └── service.go          # Monitors state hash changes
├── evaluation/             # Sequence evaluation
│   └── service.go          # Evaluates trade sequences
├── handlers/               # HTTP handlers
│   └── handler.go          # Planning API endpoints
└── recommendation_repository.go  # Recommendation storage
```

## Import Path Guidelines

**Internal Packages:**
- Use full import paths: `github.com/aristath/sentinel/internal/domain`
- Use aliases for handlers: `import portfoliohandlers "github.com/aristath/sentinel/internal/modules/portfolio/handlers"`
- Never import `internal/` packages from `pkg/` (breaks encapsulation)

**Public Packages:**
- Use full import paths: `github.com/aristath/sentinel/pkg/logger`
- Can be imported by external code

**Third-Party:**
- Group separately from stdlib and local imports
- Examples: `github.com/go-chi/chi/v5`, `github.com/rs/zerolog`, `modernc.org/sqlite`

---

*Structure analysis: 2026-01-20*
