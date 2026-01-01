# Deployment Considerations and Resource Limitations

## Arduino Uno Q Resource Constraints

The Arduino Trader is designed to run on an Arduino Uno Q (x86 SBC with limited resources). When deploying the microservices architecture, several resource constraints must be considered:

### Hardware Specifications
- **CPU**: Limited processing power compared to traditional servers
- **RAM**: Constrained memory requiring careful resource management
- **Storage**: Limited disk space for applications and data
- **Network**: Single network interface with bandwidth limitations

### Deployment Modes

#### 1. **Local Mode (Recommended for Single Arduino)**
**Resource Impact**: MINIMAL
- All services run in-process (no gRPC overhead)
- Shared memory space
- No network communication between services
- Latency: ~0.1ms (function calls)
- Memory: Single process footprint

**When to use**:
- Single Arduino Uno Q deployment
- Development and testing
- Resource-constrained environments

**Configuration**:
```yaml
# app/config/device.yaml
device:
  id: "device1"
  role: "standalone"

# app/config/services.yaml
services:
  planning:
    mode: "local"
  scoring:
    mode: "local"
  # ... all services set to local
```

#### 2. **Dual-Device Mode**
**Resource Impact**: MODERATE
- Services distributed across 2 Arduino devices
- gRPC network overhead
- Each device runs 3-4 services
- Latency: ~1-5ms (localhost gRPC)
- Memory: Split across devices

**When to use**:
- Two Arduino Uno Q devices available
- Want to isolate core planning from execution
- Need better fault tolerance

**Recommended split**:
- **Device 1** (Core): Planning, Scoring, Universe, Gateway
- **Device 2** (Execution): Portfolio, Trading, Optimization

#### 3. **Docker Mode (Testing Only)**
**Resource Impact**: HIGH
- Each service in separate container
- Docker daemon overhead
- Inter-container networking
- NOT recommended for Arduino production use

**When to use**:
- Local development on powerful machines
- Integration testing
- Service debugging

### Resource Analysis

#### Memory Footprint (Estimated)

| Component | Local Mode | gRPC Mode (per service) |
|-----------|------------|-------------------------|
| Python Runtime | ~50MB | ~50MB per container |
| Service Code | ~10MB | ~10MB per service |
| Circuit Breakers | ~1MB | ~1MB per service |
| Metrics Collection | ~2MB | ~2MB per service |
| gRPC Overhead | N/A | ~5MB per service |
| **Total (7 services)** | ~63MB | ~476MB |

#### CPU Utilization

**Local Mode**:
- Scoring/Planning: CPU-intensive during batch operations
- Other services: Minimal CPU usage
- **Peak CPU**: 80-90% during planning cycle
- **Idle CPU**: <10%

**gRPC Mode (per device)**:
- Additional CPU for serialization/deserialization
- Network I/O overhead
- **Peak CPU**: 90-100% (may cause slowdowns)
- **Idle CPU**: 15-20% (gRPC keepalives)

### Performance Characteristics

#### Local Mode Benchmarks
- Portfolio summary: <1ms
- Score single security: ~50ms
- Score all securities (batch): ~2s (50 securities)
- Generate holistic plan: ~5-10s
- Trade execution: ~100ms

#### gRPC Mode Overhead
- Add +1-2ms per service call
- Protobuf serialization: +0.5ms
- Network latency (localhost): +0.5-1ms
- Total overhead: ~2-3x local mode

### Mitigation Strategies

#### For Single Arduino Deployment
1. **Use Local Mode Exclusively**
   - No gRPC overhead
   - Minimal memory footprint
   - Best performance

2. **Optimize Batch Operations**
   ```python
   # Good: Batch processing
   scores = await scoring_service.batch_score_securities(isins)

   # Bad: Individual calls
   for isin in isins:
       score = await scoring_service.score_security(isin)
   ```

3. **Disable Unnecessary Features**
   - Metrics collection (if not monitoring)
   - Circuit breakers (in local mode, they're redundant)
   - Detailed logging (use ERROR level only)

#### For Dual Arduino Deployment
1. **Strategic Service Placement**
   - CPU-heavy services (Planning, Scoring) on more powerful device
   - I/O-heavy services (Trading, Portfolio) on device with better network

2. **Connection Pooling**
   - Reuse gRPC channels (already implemented)
   - Keep-alive settings optimized

3. **Circuit Breaker Configuration**
   ```python
   # Relaxed thresholds for resource-constrained environments
   CircuitBreakerConfig(
       failure_threshold=10,  # More tolerant
       timeout=120.0,         # Longer recovery time
   )
   ```

### Monitoring and Alerts

#### Key Metrics to Watch
1. **Memory Usage**: Alert if >80% of available RAM
2. **CPU Usage**: Alert if sustained >90% for >5 minutes
3. **Disk Space**: Alert if <500MB free
4. **Response Times**: Alert if P95 latency >1s

#### Health Check Strategy
```python
# Lightweight health checks
@app.get("/health")
async def health():
    return {"status": "healthy"}  # Don't check dependencies in local mode
```

### Production Recommendations

#### For Arduino Uno Q (Single Device)
✅ **DO**:
- Deploy in local mode
- Use TOML configuration for planner settings
- Enable basic logging (ERROR level)
- Schedule planning during off-peak hours
- Use database vacuuming regularly

❌ **DON'T**:
- Run gRPC services on single device
- Enable verbose DEBUG logging
- Run Docker containers
- Collect Prometheus metrics continuously
- Run all services simultaneously with full circuit breakers

#### For Dual Arduino Setup
✅ **DO**:
- Split by CPU intensity (see recommended split above)
- Use static IP addresses for service discovery
- Configure circuit breakers with relaxed thresholds
- Monitor network bandwidth usage
- Test failover scenarios

❌ **DON'T**:
- Put all CPU-heavy services on one device
- Rely on dynamic service discovery
- Use aggressive retry policies (causes cascading load)
- Deploy without testing resource usage

### Migration Path

If starting with single Arduino and considering dual-device:

1. **Phase 1**: Deploy in local mode, measure resource usage
2. **Phase 2**: If hitting resource limits, identify bottleneck services
3. **Phase 3**: Acquire second Arduino, test dual-device in staging
4. **Phase 4**: Gradual migration (one service at a time)
5. **Phase 5**: Monitor and optimize based on real usage

### Future Scalability

The microservices architecture supports:
- **Vertical scaling**: Upgrade to more powerful SBC (e.g., Raspberry Pi 5)
- **Horizontal scaling**: Add more devices for specific services
- **Cloud hybrid**: Move non-critical services to cloud (future)

However, for current Arduino Uno Q deployment:
- **Local mode is strongly recommended**
- Distributed mode should only be used with 2+ devices
- Resource monitoring is essential

## Conclusion

The microservices implementation provides **flexibility without forcing complexity**. For typical Arduino Uno Q deployments:
- Start with local mode
- Monitor resource usage
- Only distribute if necessary
- Always test configuration changes in staging

The architecture's value is in **optionality** - you can run distributed when beneficial, but aren't forced to pay the overhead cost.
