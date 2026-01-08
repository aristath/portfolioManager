# Market-Aware Job Scheduling Implementation

## Overview

Implemented dynamic job scheduling based on real-time market state, with intelligent trade execution throttling and robust error handling.

## Key Features

### 1. Market State Detection
- Automatically detects exchanges from securities in the universe
- Top 2 exchanges by security count = "dominant markets"
- All others = "secondary markets"
- Refresh interval: 1 hour

### 2. Dynamic Job Intervals

| Market State | Sync Interval | Description |
|-------------|--------------|-------------|
| Pre-market (30 min before dominant open) | 5 minutes | Prepare for market open |
| Dominant markets open | 5 minutes | Main trading hours - responsive |
| Secondary markets only | 10 minutes | Moderate responsiveness |
| All markets closed | Skip | No execution when markets closed |

### 3. Trade Execution Throttling
- Maximum 1 trade per 15 minutes
- In-memory throttle lock (non-blocking)
- Prevents worker starvation
- Reduces market impact

### 4. Retry Logic
- Maximum 3 retry attempts per recommendation
- Tracks retry count, last attempt timestamp, and failure reason
- Permanent "failed" status after max retries exceeded
- Prevents infinite retry loops

## Files Modified

### Core Implementation
- `internal/market_regime/market_state.go` - Market state detector (NEW)
- `internal/queue/scheduler.go` - Market-aware sync cycle scheduling
- `internal/scheduler/event_based_trading.go` - Trade throttling and retry logic
- `internal/modules/planning/recommendation_repository.go` - Retry tracking

### Dependency Injection
- `internal/di/services.go` - Wired MarketStateDetector
- `internal/di/types.go` - Added MarketStateDetector to container

### Database
- `internal/database/migrations_archive/036_add_recommendation_retry_tracking.sql` - Retry tracking schema

## Bug Fixes

### 1. Race Condition in MarketStateDetector
**Problem**: Concurrent access to cache fields without synchronization
**Fix**: Added `sync.RWMutex` with proper read/write lock patterns

### 2. Blocking Sleep in Trade Execution
**Problem**: `time.Sleep(15 * time.Minute)` blocked worker for hours
**Fix**: Replaced with in-memory throttle lock - job completes in seconds

### 3. Missing Retry Limit
**Problem**: Failed trades would retry infinitely
**Fix**: Added retry tracking with 3-attempt maximum and permanent failure status

### 4. Timezone Parsing Bug
**Problem**: `time.Parse()` caused incorrect pre-market calculations across timezones
**Fix**: Use `time.ParseInLocation()` with proper timezone from MarketStatus

### 5. Scheduler Stop Channel Race
**Problem**: Multiple `Start()` calls without `Stop()` could create zombie goroutines
**Fix**: Added `sync.WaitGroup` to track goroutine lifecycle, `Stop()` now waits for all goroutines to finish

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     TimeScheduler                           │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Market-Aware Sync Cycle (checks every minute)      │   │
│  │  ┌─────────────────────────────────────────────┐    │   │
│  │  │ MarketStateDetector.GetSyncInterval(now)    │    │   │
│  │  │  - Returns 5 min (dominant/pre-market)      │    │   │
│  │  │  - Returns 10 min (secondary only)          │    │   │
│  │  │  - Returns 0 (all closed - SKIP)            │    │   │
│  │  └─────────────────────────────────────────────┘    │   │
│  │                                                       │   │
│  │  Enqueues sync_cycle if interval matches             │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    QueueManager                             │
│  - Stores JobTypeSyncCycle in queue                         │
│  - WorkerPool picks up job                                  │
│  - Executes SyncCycleJob                                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    SyncCycleJob                             │
│  1. Fetch prices                                            │
│  2. Update positions                                        │
│  3. Run planning                                            │
│  4. Emit RecommendationsReady event                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Event: RecommendationsReady                    │
│  Triggers: EventBasedTradingJob (CRITICAL priority)         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              EventBasedTradingJob                           │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Check in-memory throttle lock                      │    │
│  │  - Skip if < 15 minutes since last execution       │    │
│  └────────────────────────────────────────────────────┘    │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Get ONE pending recommendation                     │    │
│  │  - Check retry count (max 3)                       │    │
│  │  - Execute trade                                   │    │
│  │  - Mark executed OR record failed attempt          │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## Database Schema Changes

```sql
ALTER TABLE recommendations ADD COLUMN retry_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE recommendations ADD COLUMN last_attempt_at INTEGER;
ALTER TABLE recommendations ADD COLUMN failure_reason TEXT;

-- Updated status constraint to include 'failed'
status TEXT NOT NULL DEFAULT 'pending'
  CHECK (status IN ('pending', 'executed', 'rejected', 'expired', 'failed'))
```

## Thread Safety

### MarketStateDetector Cache
```go
type MarketStateDetector struct {
    // ...
    mu                 sync.RWMutex  // Protects all fields below
    lastExchangeUpdate time.Time
    dominantExchanges  []string
    secondaryExchanges []string
}

// Pattern 1: Read lock for checking
d.mu.RLock()
needsUpdate := d.lastExchangeUpdate.IsZero()
d.mu.RUnlock()

// Pattern 2: Write lock for updating
d.mu.Lock()
d.dominantExchanges = newDominant
d.lastExchangeUpdate = time.Now()
d.mu.Unlock()

// Pattern 3: Defensive copies when returning
d.mu.RLock()
result := make([]string, len(d.dominantExchanges))
copy(result, d.dominantExchanges)
d.mu.RUnlock()
return result
```

### Trade Throttle Lock
```go
type EventBasedTradingJob struct {
    mu                sync.Mutex
    lastExecutionTime time.Time
    throttleDuration  time.Duration
}

j.mu.Lock()
timeSinceLastExecution := time.Since(j.lastExecutionTime)
if !j.lastExecutionTime.IsZero() && timeSinceLastExecution < j.throttleDuration {
    j.mu.Unlock()
    // Skip execution
    return nil
}
j.lastExecutionTime = time.Now()
j.mu.Unlock()
```

### Scheduler Goroutine Tracking
```go
type Scheduler struct {
    // ...
    mu      sync.Mutex
    wg      sync.WaitGroup // Track goroutine lifecycle
    stop    chan struct{}
    stopped bool
    started bool
}

// Start: Add to WaitGroup before each goroutine
s.wg.Add(1)
go func() {
    defer s.wg.Done()
    for {
        select {
        case <-s.stop:
            ticker.Stop()
            return
        case <-ticker.C:
            // Do work...
        }
    }
}()

// Stop: Wait for all goroutines to finish
func (s *Scheduler) Stop() {
    s.mu.Lock()
    if s.stopped {
        s.mu.Unlock()
        return
    }
    close(s.stop)
    s.stopped = true
    s.mu.Unlock()

    s.wg.Wait() // Block until all goroutines exit
    s.log.Info().Msg("Time scheduler stopped")
}
```

## Testing Recommendations

### Unit Tests
```bash
# Test market state detection
go test ./internal/market_regime -v

# Test scheduler interval calculation
go test ./internal/queue -run TestScheduler -v

# Test trade throttling
go test ./internal/scheduler -run TestEventBasedTrading -v
```

### Integration Tests
1. **Market State Transitions**: Verify correct intervals across market state changes
2. **Timezone Handling**: Test pre-market detection across different timezones
3. **Trade Throttling**: Verify max 1 trade per 15 minutes under load
4. **Retry Logic**: Verify failed trades mark as permanently failed after 3 attempts

### Manual Testing
```bash
# Watch scheduler logs
journalctl -u sentinel -f | grep -E "(market_state|sync_cycle|event_based_trading)"

# Check current market state
curl http://localhost:8080/api/market-state

# View pending recommendations
curl http://localhost:8080/api/recommendations?status=pending

# Trigger sync cycle manually
curl -X POST http://localhost:8080/api/jobs/sync_cycle
```

## Performance Characteristics

- **Market state cache refresh**: 1 hour
- **Scheduler tick interval**: 1 minute (low overhead)
- **Sync cycle interval**: 5-10 minutes during market hours (adaptive)
- **Trade execution**: Non-blocking, completes in seconds
- **Worker efficiency**: No blocked workers, full pool utilization

## Deployment Notes

1. Run migration 036 before deploying:
   ```bash
   # Migration is in migrations_archive/ and will auto-apply on startup
   ```

2. Monitor logs for market state transitions:
   ```bash
   # Should see messages like:
   # "Market state changed" old_state="all_closed" new_state="pre_market"
   # "Sync cycle enqueued (market-aware)" state="dominant_open" interval="5m"
   ```

3. Verify trade throttling is working:
   ```bash
   # Should see messages like:
   # "Trade throttle active - skipping execution" remaining="13m42s"
   ```

## Known Limitations

1. **Cache Staleness**: Exchange counts refresh hourly. New securities won't affect dominant/secondary classification until next refresh.

2. **Single Trade Per Cycle**: Job processes one trade per execution. Multiple pending recommendations will be processed across multiple cycles (15-min intervals).

## Future Enhancements

- Exponential backoff for failed trade retries (currently fixed 15-min intervals)
- Configurable retry limits per recommendation type
- Circuit breaker for repeated failures
- Metrics for trade execution success rate
