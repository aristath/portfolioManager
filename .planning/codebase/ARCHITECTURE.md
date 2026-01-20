# Architecture

**Analysis Date:** 2026-01-20

## Pattern Overview

**Overall:** Clean Architecture with Dependency Injection

**Key Characteristics:**
- Strict layered architecture with dependency inversion (domain layer has no infrastructure imports)
- Repository pattern abstracts all data access behind interfaces
- Service layer coordinates business logic between repositories and domain models
- Constructor-based dependency injection via centralized DI container (`internal/di`)
- Event-driven background job system with market-aware scheduling
- Broker abstraction allows switching trading providers without changing business logic

## Layers

**Domain Layer:**
- Purpose: Pure domain models and interfaces with zero infrastructure dependencies
- Location: `internal/domain`
- Contains: Broker-agnostic types, core interfaces (BrokerClient, CashManager, CurrencyExchangeServiceInterface)
- Depends on: Nothing (pure Go types only)
- Used by: All layers (repositories, services, handlers)

**Repository Layer:**
- Purpose: Data access abstraction via interfaces
- Location: `internal/modules/*/repository.go`, `internal/modules/*/*_repository.go`
- Contains: Repository implementations for each data entity (positions, securities, scores, trades, dividends, etc.)
- Depends on: Domain types, database connections
- Used by: Service layer

**Service Layer:**
- Purpose: Business logic coordination and orchestration
- Location: `internal/modules/*/service.go`, `internal/services`
- Contains: Business logic services (trading, portfolio, planning, optimization, scoring, etc.)
- Depends on: Repositories (via interfaces), domain types
- Used by: HTTP handlers, background jobs

**Adapter Layer:**
- Purpose: External system integration with domain type mapping
- Location: `internal/clients/tradernet/adapter.go`, `internal/clients/tradernet/transformers_domain.go`
- Contains: Broker API client implementations, WebSocket clients, transformers (broker types → domain types)
- Depends on: Domain interfaces (implements BrokerClient)
- Used by: Services (via domain.BrokerClient interface)

**Handler Layer:**
- Purpose: HTTP request handling and API endpoints
- Location: `internal/modules/*/handlers`, `internal/server`
- Contains: HTTP handlers for each module, request validation, response serialization
- Depends on: Services (via DI container), repositories
- Used by: HTTP server router

**Database Layer:**
- Purpose: SQLite database management with profile-based optimization
- Location: `internal/database`
- Contains: Database connection management, schema initialization, transaction support, health checks
- Depends on: Embedded schema files (`internal/database/schemas/*.sql`)
- Used by: Repositories

**Work Processor Layer:**
- Purpose: Event-driven background job execution with dependency resolution
- Location: `internal/work`
- Contains: Work processor (job queue), work registry (job types), market timing checker, completion cache
- Depends on: Services (executes jobs), event bus (triggers)
- Used by: Main application loop, event listeners

**Event Bus Layer:**
- Purpose: Decoupled event-driven communication between modules
- Location: `internal/events`
- Contains: Event bus (pub/sub), event manager (high-level API), typed event data structures
- Depends on: Nothing (pure event passing)
- Used by: All layers (emit and listen to events)

## Data Flow

**Trade Execution Flow:**

1. User submits trade via frontend → HTTP handler (`internal/modules/trading/handlers`)
2. Handler validates request → TradingService (`internal/modules/trading/service.go`)
3. TradingService checks safety → TradeSafetyService (`internal/modules/trading/safety_service.go`)
4. TradingService executes via broker → TradeExecutionService → BrokerClient (adapter)
5. BrokerClient places order → Tradernet API → Transformer maps response to domain.BrokerOrderResult
6. TradingService persists trade → TradeRepository → LedgerDB
7. TradingService emits TradeExecuted event → EventBus
8. Event triggers background work → Work Processor queues SyncPortfolio job
9. SyncPortfolio job syncs positions → PortfolioService → PositionRepository → PortfolioDB

**Portfolio Sync Flow:**

1. Time-based trigger (scheduler) or event (TradeExecuted) → Work Processor
2. Work Processor executes SyncPortfolio job → PortfolioService.SyncFromBroker()
3. PortfolioService fetches positions → BrokerClient.GetPortfolio()
4. BrokerClient retrieves positions → Tradernet API → Transformer maps to domain.BrokerPosition
5. PortfolioService updates positions → PositionRepository.SavePosition() → PortfolioDB
6. PortfolioService updates cash → CashManager.UpdateCashPosition() → PortfolioDB
7. PortfolioService emits PortfolioChanged event → EventBus
8. StateMonitor detects state hash change → Emits StateChanged event
9. StateChanged event triggers planning jobs → OpportunitiesService, SequencesService, PlannerService

**Planning Recommendation Flow:**

1. StateChanged event triggers work → Work Processor queues planning jobs
2. GetOptimizerWeights job → OptimizerService.GetHRPWeights() → CalculationsDB (cache)
3. BuildOpportunityContext job → OpportunitiesService → OpportunityContextBuilder
4. CreateTradePlan job → PlannerService.GeneratePlan()
5. PlannerService generates sequences → SequencesService → EvaluationService (worker pool)
6. EvaluationService evaluates sequences → Simulates trades, calculates metrics
7. PlannerService selects best plan → Stores recommendations → RecommendationRepo (in-memory)
8. PlannerService emits RecommendationsReady event → EventBus
9. Frontend polls /api/planning/recommendations → Returns stored recommendations

**State Management:**

- Portfolio state is source of truth in PortfolioDB (positions, cash, scores, metrics)
- State hash calculated from positions + scores + allocation targets (`StateHashService`)
- StateMonitor polls state hash every minute, emits StateChanged when hash changes
- StateChanged event triggers planning system to generate new recommendations
- Recommendations stored in-memory (RecommendationRepo) for fast access, not persisted

## Key Abstractions

**BrokerClient Interface:**
- Purpose: Broker-agnostic trading and portfolio operations
- Examples: `internal/clients/tradernet/adapter.go` (Tradernet implementation)
- Pattern: Adapter pattern with transformer layer for type mapping
- Design: All broker-specific quirks isolated in adapter/transformers, services never reference broker internals

**CashManager Interface:**
- Purpose: Cash balance management (breaks circular dependencies)
- Examples: `internal/modules/cash_flows/cash_manager.go`
- Pattern: Interface segregation (multiple packages depend on it without coupling)

**Repository Interfaces:**
- Purpose: Data access abstraction with testable contracts
- Examples: `planning.RecommendationRepositoryInterface`, `planningrepo.PlannerRepositoryInterface`
- Pattern: Repository pattern with interface-first design (can swap DB/in-memory implementations)

**DI Container:**
- Purpose: Single source of truth for all service instances
- Examples: `internal/di/types.go` (Container struct), `internal/di/services.go` (InitializeServices)
- Pattern: Dependency injection container with constructor injection only
- Design: All dependencies wired in correct order (databases → repositories → services → work processor)

**Work Item:**
- Purpose: Background job abstraction with dependency resolution
- Examples: `internal/work/types.go` (WorkItem), individual job implementations in various services
- Pattern: Job pattern with registry, processor handles dependency resolution and market timing

**Event:**
- Purpose: Typed event data for pub/sub communication
- Examples: `internal/events/types.go` (Event types), `internal/events/manager.go` (EventManager)
- Pattern: Event-driven architecture with typed events (TradeExecuted, PortfolioChanged, StateChanged, etc.)

## Entry Points

**HTTP Server Entry Point:**
- Location: `cmd/server/main.go`
- Triggers: Application startup
- Responsibilities: Loads config, wires dependencies via DI, starts HTTP server, starts background monitors, waits for shutdown signal

**HTTP Server:**
- Location: `internal/server/server.go`
- Triggers: Incoming HTTP requests
- Responsibilities: Routes requests to module handlers, serves embedded frontend, provides health check endpoint

**Work Processor:**
- Location: `internal/work/processor.go`
- Triggers: Events (from EventBus), time-based schedules (ticker), manual API calls
- Responsibilities: Executes background jobs, resolves dependencies, respects market timing, tracks completion history

**State Monitor:**
- Location: `internal/modules/planning/state_monitor/service.go`
- Triggers: Timer (every 60 seconds)
- Responsibilities: Calculates state hash, emits StateChanged event when hash changes

**Market Status WebSocket:**
- Location: `internal/clients/tradernet/websocket_client.go`
- Triggers: WebSocket connection established
- Responsibilities: Receives real-time market status updates, triggers work processor when markets open/close

## Error Handling

**Strategy:** Graceful degradation with structured logging

**Patterns:**
- Return errors from functions using `fmt.Errorf` with `%w` verb for wrapping
- Log errors with structured logging (zerolog) including context fields
- Degrade gracefully - partial results over total failure (e.g., portfolio sync continues even if some positions fail)
- Repository layer returns `nil, error` for not found cases (caller decides if error or valid state)
- Service layer wraps repository errors with business context
- Handler layer maps errors to HTTP status codes (400 for validation, 500 for internal errors)
- Background jobs retry failed work items with exponential backoff (Work Processor manages retry queue)

## Cross-Cutting Concerns

**Logging:** Structured logging with zerolog, logger passed via DI container to all services

**Validation:** Request validation in handlers (before calling services), business validation in services (before persistence)

**Authentication:** Settings stored in `config.db`, API credentials configured via Settings UI or Settings API, broker credentials passed to BrokerClient adapter

**Transaction Management:** Database layer provides transaction support (`db.BeginTx()`), repositories use transactions for multi-step operations, ledger writes use transactions for atomicity

**Configuration:** Two-tier config system - infrastructure settings from `.env` (via godotenv), application settings from `config.db` (via SettingsRepo), settings database takes precedence over environment variables

**Deployment:** Automated deployment via `internal/deployment` (monitors git, downloads binaries from GitHub Actions, restarts services), optional R2 cloud backup for databases

**Testing:** Unit tests for domain logic (no DB/network), integration tests for repositories and API clients, test utilities in `internal/testing` for database fixtures and test helpers

---

*Architecture analysis: 2026-01-20*
