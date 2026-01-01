# Deployment Guide

This directory contains deployment configurations and scripts for the Arduino Trader microservices architecture.

## Deployment Modes

### Single Device (Local)
All services run on one Arduino Uno Q in local mode (in-process, no gRPC overhead).

**Configuration:** `configs/single-device/`
- `device.yaml`: Device configuration
- `services.yaml`: All services in `mode: local`

**Use case:** Development, testing, or single-device production deployment.

### Dual Device (Distributed)
Services distributed across two Arduino Uno Q devices via gRPC.

**Configuration:** `configs/dual-device/`
- `device1.yaml`: Core services (planning, scoring, universe, gateway)
- `device2.yaml`: Execution services (portfolio, trading, optimization)
- `services.yaml`: Services configured with `mode: remote` and IP addresses

**Use case:** Load distribution, fault tolerance, or separating concerns.

## Quick Start

### Single Device Deployment

1. **Copy configuration files:**
   ```bash
   cp deploy/configs/single-device/*.yaml app/config/
   ```

2. **Start all services:**
   ```bash
   chmod +x deploy/scripts/*.sh
   ./deploy/scripts/start-all-services.sh
   ```

3. **Check status:**
   ```bash
   ./deploy/scripts/check-services-status.sh
   ```

4. **Stop services:**
   ```bash
   ./deploy/scripts/stop-all-services.sh
   ```

### Dual Device Deployment

#### Device 1 (Core Services)

1. **Copy configuration:**
   ```bash
   cp deploy/configs/dual-device/device1.yaml app/config/device.yaml
   cp deploy/configs/dual-device/services.yaml app/config/services.yaml
   ```

2. **Update IP addresses in** `app/config/services.yaml`

3. **Start services for this device:**
   ```bash
   # Start only services that run on device1
   python3 -m services.planning.main &
   python3 -m services.scoring.main &
   python3 -m services.universe.main &
   python3 -m services.gateway.main &
   ```

#### Device 2 (Execution Services)

1. **Copy configuration:**
   ```bash
   cp deploy/configs/dual-device/device2.yaml app/config/device.yaml
   cp deploy/configs/dual-device/services.yaml app/config/services.yaml
   ```

2. **Update IP addresses in** `app/config/services.yaml`

3. **Start services for this device:**
   ```bash
   # Start only services that run on device2
   python3 -m services.portfolio.main &
   python3 -m services.trading.main &
   python3 -m services.optimization.main &
   ```

## Docker Compose Deployment

For local testing with all services in containers:

```bash
docker-compose up -d
```

This starts all 7 services in separate containers with gRPC communication.

## Service Ports

| Service       | Port  |
|---------------|-------|
| Planning      | 50051 |
| Scoring       | 50052 |
| Optimization  | 50053 |
| Portfolio     | 50054 |
| Trading       | 50055 |
| Universe      | 50056 |
| Gateway       | 50057 |

## Logs

Service logs are stored in `logs/` directory:
- `logs/planning.log`
- `logs/scoring.log`
- etc.

PID files are also stored in `logs/` for process management.

## Health Checks

Check if services are healthy:

```bash
# Using grpcurl
grpcurl -plaintext localhost:50051 PlanningService/HealthCheck
grpcurl -plaintext localhost:50052 ScoringService/HealthCheck
# etc.

# Or via Gateway service
grpcurl -plaintext localhost:50057 GatewayService/GetSystemStatus
```

## Troubleshooting

### Service won't start
1. Check logs in `logs/<service>.log`
2. Ensure ports are not in use: `lsof -i :<port>`
3. Verify configuration files are correct
4. Check Python environment: `source venv/bin/activate`

### Cannot connect to remote service
1. Verify IP addresses in `services.yaml`
2. Check network connectivity: `ping <device-ip>`
3. Ensure firewall allows gRPC ports (50051-50057)
4. Verify service is running on remote device

### gRPC errors
1. Ensure protocol buffers are generated: `./scripts/generate_protos.sh`
2. Check gRPC dependencies: `pip list | grep grpc`
3. Verify service configuration mode (`local` vs `remote`)

## Production Considerations

### Security
- Enable TLS/mTLS for gRPC communication (Phase 7)
- Use strong authentication between services
- Configure firewalls to restrict service ports

### Monitoring
- Implement Prometheus metrics export (Phase 7)
- Set up distributed tracing
- Configure alerting for service failures

### Reliability
- Implement circuit breakers (Phase 7)
- Configure retry logic with exponential backoff
- Set up health check monitoring
- Consider service mesh (e.g., Istio) for advanced features

## Configuration Reference

### device.yaml
```yaml
device:
  id: "<unique-device-id>"
  name: "<human-readable-name>"
  roles: ["service1", "service2"]  # or ["all"] for single device
  network:
    bind_address: "0.0.0.0"
    advertise_address: "<ip-or-hostname>"
  resources:
    max_workers: 10
    max_memory_mb: 2048
```

### services.yaml
```yaml
deployment:
  mode: "local" | "distributed"

services:
  service_name:
    mode: "local" | "remote"
    device_id: "<device-id>"
    address: "<ip-address>"  # Only for remote mode
    port: <port-number>
    client:
      timeout_seconds: <timeout>
      max_retries: <retries>
```
