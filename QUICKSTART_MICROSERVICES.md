# Planning Microservices - Quick Start Guide

This guide helps you get the Planning microservices up and running.

## Prerequisites

- Docker and Docker Compose installed
- Python 3.9+ with venv activated
- All dependencies installed (`pip install -r requirements.txt`)

## Architecture Overview

The Planning service has been split into 4 microservices:

```
┌─────────────┐
│ Coordinator │ (Port 8011)
└──────┬──────┘
       │
       ├─────────────┬─────────────┬─────────────┐
       │             │             │             │
       ▼             ▼             ▼             ▼
┌─────────────┐ ┌─────────┐ ┌───────────┐ ┌───────────┐
│ Opportunity │ │Generator│ │Evaluator-1│ │Evaluator-2│
│  (8008)     │ │ (8009)  │ │  (8010)   │ │  (8020)   │
└─────────────┘ └─────────┘ └───────────┘ └───────────┘
                                            ┌───────────┐
                                            │Evaluator-3│
                                            │  (8030)   │
                                            └───────────┘
```

**Performance Target:** 2.5-3× speedup (86s → 32-34s)

## Quick Start

### 1. Start Services

```bash
# Build and start all services
./scripts/start_planning_services.sh build
./scripts/start_planning_services.sh up

# Verify services are healthy
./scripts/start_planning_services.sh health
```

### 2. Run Tests

```bash
# Run unit tests (no services required)
./scripts/test_planning_services.sh unit

# Run integration tests (requires services running)
./scripts/test_planning_services.sh integration

# Run equivalence test (verifies microservices match monolithic)
./scripts/test_planning_services.sh equivalence

# Run performance test (verifies 2.5× speedup)
./scripts/test_planning_services.sh performance

# Run all tests
./scripts/test_planning_services.sh all
```

### 3. Manual Testing

#### Health Checks
```bash
curl http://localhost:8008/opportunity/health
curl http://localhost:8009/generator/health
curl http://localhost:8010/evaluator/health
curl http://localhost:8020/evaluator/health
curl http://localhost:8030/evaluator/health
curl http://localhost:8011/coordinator/health
```

#### Create a Plan via Coordinator
```bash
curl -X POST http://localhost:8011/coordinator/create-plan \
  -H "Content-Type: application/json" \
  -d @tests/fixtures/sample_plan_request.json
```

### 4. View Logs

```bash
# All services
./scripts/start_planning_services.sh logs

# Specific service
docker-compose logs -f opportunity
docker-compose logs -f generator
docker-compose logs -f evaluator-1
docker-compose logs -f coordinator
```

### 5. Stop Services

```bash
./scripts/start_planning_services.sh down
```

## Service Details

### Opportunity Service (Port 8008)
**Purpose:** Identify trading opportunities from portfolio state

**Endpoint:** `POST /opportunity/identify`

**Features:**
- Weight-based opportunity identification
- Heuristic-based fallback
- 5 opportunity categories

### Generator Service (Port 9009)
**Purpose:** Generate and filter action sequences

**Endpoint:** `POST /generator/generate` (streaming)

**Features:**
- 10 pattern types
- Combinatorial generation
- Batch streaming (500-1000 sequences per batch)
- Feasibility filtering

### Evaluator Service (Ports 8010/8020/8030)
**Purpose:** Simulate and score action sequences

**Endpoint:** `POST /evaluator/evaluate`

**Features:**
- Portfolio simulation
- Beam search (top K sequences)
- 3 parallel instances for distributed workload

### Coordinator Service (Port 8011)
**Purpose:** Orchestrate the complete planning workflow

**Endpoint:** `POST /coordinator/create-plan`

**Features:**
- Round-robin load balancing across evaluators
- Global beam aggregation
- Error handling with partial results
- Execution statistics

## Configuration

### Service Discovery
Services are configured in `app/config/services.yaml`:

```yaml
services:
  opportunity:
    mode: "local"
    port: 8008

  generator:
    mode: "local"
    port: 8009

  evaluator:
    instances:
      - port: 8010
      - port: 8020
      - port: 8030
    load_balancing: "round_robin"

  coordinator:
    mode: "local"
    port: 8011
```

### Docker Compose
Services are defined in `docker-compose.yml`. Each service:
- Has its own container
- Mounts `app/config` for shared configuration
- Connected via `arduino-trader` network

## Troubleshooting

### Service Not Starting
```bash
# Check logs
docker-compose logs <service-name>

# Rebuild service
docker-compose build <service-name>
docker-compose up -d <service-name>
```

### Service Not Healthy
```bash
# Check health endpoint
curl http://localhost:<port>/<service>/health

# Check if port is already in use
lsof -i :<port>
```

### Slow Performance
- Verify all 3 Evaluator instances are running
- Check CPU usage (should see 2-3 cores utilized)
- Monitor batch sizes in logs

### Tests Failing
```bash
# Ensure services are running
./scripts/start_planning_services.sh health

# Run tests with verbose output
./venv/bin/python3 -m pytest tests/integration/services/planning/ -vvs
```

## Development Workflow

### Adding New Features

1. **Update Domain Logic** in `holistic_planner.py`
2. **Services Use New Logic** automatically (thin wrapper)
3. **Add Tests** for new functionality
4. **Restart Services** to pick up changes

### Scaling Evaluators

Add more instances in `docker-compose.yml`:
```yaml
evaluator-4:
  build:
    context: .
    dockerfile: services/evaluator/Dockerfile
  ports:
    - "8040:8040"
  environment:
    - EVALUATOR_PORT=8040
```

Update `services.yaml`:
```yaml
evaluator:
  instances:
    - port: 8010
    - port: 8020
    - port: 8030
    - port: 8040  # New instance
```

## Performance Monitoring

### Metrics to Watch
- Request latency per service
- Batch processing time (Evaluator)
- Round-robin distribution fairness
- Beam aggregation time (Coordinator)
- Error rates

### Log Examples

**Successful Plan Creation:**
```
INFO:coordinator:Identified 15 opportunities
INFO:coordinator:Generated 2500 sequences in 3 batches
INFO:coordinator:Evaluated 2500 sequences using 3 evaluators
INFO:coordinator:Created plan with 8 steps in 34.2 seconds
```

**Round-Robin Distribution:**
```
INFO:coordinator:Batch 0 -> Evaluator-1
INFO:coordinator:Batch 1 -> Evaluator-2
INFO:coordinator:Batch 2 -> Evaluator-3
INFO:coordinator:Batch 3 -> Evaluator-1
```

## Next Steps

- **Phase 2:** Test parallel evaluators, measure actual speedup
- **Phase 3:** Production deployment across multiple devices
- **Enhancements:**
  - Full scoring integration in Evaluator
  - Narrative generation in Coordinator
  - Correlation-aware filtering in Generator
  - Monte Carlo evaluation (optional)

## Documentation

- **Architecture Details:** `services/README.md`
- **API Documentation:** Each service's `routes.py`
- **Test Examples:** `tests/unit/services/*/`

## Support

For issues or questions:
1. Check `services/README.md` for detailed documentation
2. Review logs: `./scripts/start_planning_services.sh logs`
3. Run health checks: `./scripts/start_planning_services.sh health`
4. Check test output for specific errors
