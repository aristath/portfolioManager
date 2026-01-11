package queue

import (
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/aristath/sentinel/internal/events"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupWorkerTest(t *testing.T) (*WorkerPool, *Manager, *Registry, *sql.DB) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS job_history (
			job_type TEXT PRIMARY KEY,
			last_run_at TEXT NOT NULL,
			last_status TEXT NOT NULL DEFAULT 'success'
		)
	`)
	require.NoError(t, err)

	queue := NewMemoryQueue()
	history := NewHistory(db)
	manager := NewManager(queue, history)
	registry := NewRegistry()

	pool := NewWorkerPool(manager, registry, 2)

	return pool, manager, registry, db
}

func TestWorkerPool_ProcessJob(t *testing.T) {
	pool, manager, registry, db := setupWorkerTest(t)
	defer db.Close()

	var executed bool
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)

	registry.Register(JobTypePlannerBatch, func(job *Job) error {
		mu.Lock()
		executed = true
		mu.Unlock()
		wg.Done()
		return nil
	})

	job := &Job{
		ID:          "test-1",
		Type:        JobTypePlannerBatch,
		Priority:    PriorityHigh,
		AvailableAt: time.Now(),
	}

	manager.Enqueue(job)
	pool.Start()

	// Wait for handler to complete
	wg.Wait()
	// Give a bit more time for RecordExecution to complete
	time.Sleep(50 * time.Millisecond)
	pool.Stop()

	mu.Lock()
	assert.True(t, executed)
	mu.Unlock()

	// Verify execution was recorded
	var lastStatus string
	err := db.QueryRow("SELECT last_status FROM job_history WHERE job_type = ?", JobTypePlannerBatch).
		Scan(&lastStatus)
	require.NoError(t, err)
	assert.Equal(t, "success", lastStatus)
}

func TestWorkerPool_ProcessJobFailure(t *testing.T) {
	pool, manager, registry, db := setupWorkerTest(t)
	defer db.Close()

	registry.Register(JobTypePlannerBatch, func(job *Job) error {
		return errors.New("test error")
	})

	job := &Job{
		ID:          "test-1",
		Type:        JobTypePlannerBatch,
		Priority:    PriorityHigh,
		AvailableAt: time.Now(),
		Retries:     3, // Already at max retries, so it will record failure
		MaxRetries:  3,
	}

	manager.Enqueue(job)
	pool.Start()

	// Give worker time to process
	time.Sleep(200 * time.Millisecond)
	pool.Stop()

	// Verify failure was recorded
	var lastStatus string
	err := db.QueryRow("SELECT last_status FROM job_history WHERE job_type = ?", JobTypePlannerBatch).
		Scan(&lastStatus)
	require.NoError(t, err)
	assert.Equal(t, "failed", lastStatus)
}

func TestWorkerPool_RetryOnFailure(t *testing.T) {
	pool, manager, registry, db := setupWorkerTest(t)
	defer db.Close()

	var attempts int
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1) // We'll wait for the final successful attempt

	registry.Register(JobTypePlannerBatch, func(job *Job) error {
		mu.Lock()
		attempts++
		currentAttempt := attempts
		mu.Unlock()

		if currentAttempt < 2 {
			return errors.New("temporary error")
		}
		// Final successful attempt
		wg.Done()
		return nil
	})

	job := &Job{
		ID:          "test-1",
		Type:        JobTypePlannerBatch,
		Priority:    PriorityHigh,
		AvailableAt: time.Now(),
		Retries:     0,
		MaxRetries:  3,
	}

	manager.Enqueue(job)
	pool.Start()

	// Wait for final successful attempt
	wg.Wait()
	// Give a bit more time for retry logic to complete
	time.Sleep(100 * time.Millisecond)
	pool.Stop()

	// Should have retried and eventually succeeded
	mu.Lock()
	assert.GreaterOrEqual(t, attempts, 2)
	mu.Unlock()

	// Verify success was recorded
	var lastStatus string
	err := db.QueryRow("SELECT last_status FROM job_history WHERE job_type = ?", JobTypePlannerBatch).
		Scan(&lastStatus)
	require.NoError(t, err)
	assert.Equal(t, "success", lastStatus)
}

// TestWorkerPool_EmitsJobStarted tests that JobStarted event is emitted
func TestWorkerPool_EmitsJobStarted(t *testing.T) {
	pool, manager, registry, db := setupWorkerTest(t)
	defer db.Close()

	log := zerolog.Nop()
	bus := events.NewBus(log)
	eventManager := events.NewManager(bus, log)
	pool.SetEventManager(eventManager)

	eventsChan := make(chan events.Event, 10)
	_ = bus.Subscribe(events.JobStarted, func(event *events.Event) {
		eventsChan <- *event
	})

	var wg sync.WaitGroup
	wg.Add(1)

	registry.Register(JobTypePlannerBatch, func(job *Job) error {
		wg.Done()
		return nil
	})

	job := &Job{
		ID:          "test-started",
		Type:        JobTypePlannerBatch,
		Priority:    PriorityHigh,
		AvailableAt: time.Now(),
	}

	manager.Enqueue(job)
	pool.Start()

	// Wait for JobStarted event
	select {
	case event := <-eventsChan:
		assert.Equal(t, events.JobStarted, event.Type)
		assert.Equal(t, "queue", event.Module)

		typedData := event.GetTypedData()
		require.NotNil(t, typedData, "Event should have typed data")
		data, ok := typedData.(*events.JobStatusData)
		require.True(t, ok, "Event data should be JobStatusData")
		assert.Equal(t, "test-started", data.JobID)
		assert.Equal(t, string(JobTypePlannerBatch), data.JobType)
		assert.Equal(t, "started", data.Status)
		assert.Equal(t, "Generating trading recommendations", data.Description)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Expected JobStarted event not received")
	}

	wg.Wait()
	pool.Stop()
}

// TestWorkerPool_EmitsJobCompleted tests that JobCompleted event is emitted
func TestWorkerPool_EmitsJobCompleted(t *testing.T) {
	pool, manager, registry, db := setupWorkerTest(t)
	defer db.Close()

	log := zerolog.Nop()
	bus := events.NewBus(log)
	eventManager := events.NewManager(bus, log)
	pool.SetEventManager(eventManager)

	eventsChan := make(chan events.Event, 10)
	_ = bus.Subscribe(events.JobCompleted, func(event *events.Event) {
		eventsChan <- *event
	})

	var wg sync.WaitGroup
	wg.Add(1)

	registry.Register(JobTypeSyncCycle, func(job *Job) error {
		time.Sleep(50 * time.Millisecond) // Simulate work
		wg.Done()
		return nil
	})

	job := &Job{
		ID:          "test-completed",
		Type:        JobTypeSyncCycle,
		Priority:    PriorityHigh,
		AvailableAt: time.Now(),
	}

	manager.Enqueue(job)
	pool.Start()

	// Wait for JobCompleted event
	select {
	case event := <-eventsChan:
		assert.Equal(t, events.JobCompleted, event.Type)

		typedData := event.GetTypedData()
		require.NotNil(t, typedData, "Event should have typed data")
		data, ok := typedData.(*events.JobStatusData)
		require.True(t, ok, "Event data should be JobStatusData")
		assert.Equal(t, "test-completed", data.JobID)
		assert.Equal(t, string(JobTypeSyncCycle), data.JobType)
		assert.Equal(t, "completed", data.Status)
		assert.Greater(t, data.Duration, 0.0) // Should have recorded duration
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Expected JobCompleted event not received")
	}

	wg.Wait()
	pool.Stop()
}

// TestWorkerPool_EmitsJobFailed tests that JobFailed event is emitted on error
func TestWorkerPool_EmitsJobFailed(t *testing.T) {
	pool, manager, registry, db := setupWorkerTest(t)
	defer db.Close()

	log := zerolog.Nop()
	bus := events.NewBus(log)
	eventManager := events.NewManager(bus, log)
	pool.SetEventManager(eventManager)

	eventsChan := make(chan events.Event, 10)
	_ = bus.Subscribe(events.JobFailed, func(event *events.Event) {
		eventsChan <- *event
	})

	registry.Register(JobTypeHourlyBackup, func(job *Job) error {
		return errors.New("backup failed: disk full")
	})

	job := &Job{
		ID:          "test-failed",
		Type:        JobTypeHourlyBackup,
		Priority:    PriorityHigh,
		AvailableAt: time.Now(),
		Retries:     3, // Already at max
		MaxRetries:  3,
	}

	manager.Enqueue(job)
	pool.Start()

	// Wait for JobFailed event
	select {
	case event := <-eventsChan:
		assert.Equal(t, events.JobFailed, event.Type)

		typedData := event.GetTypedData()
		require.NotNil(t, typedData, "Event should have typed data")
		data, ok := typedData.(*events.JobStatusData)
		require.True(t, ok, "Event data should be JobStatusData")
		assert.Equal(t, "test-failed", data.JobID)
		assert.Equal(t, string(JobTypeHourlyBackup), data.JobType)
		assert.Equal(t, "failed", data.Status)
		assert.Contains(t, data.Error, "backup failed: disk full")
		assert.Greater(t, data.Duration, 0.0)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Expected JobFailed event not received")
	}

	time.Sleep(50 * time.Millisecond)
	pool.Stop()
}

// TestWorkerPool_EmitsJobFailedOnPanic tests that JobFailed is emitted when job panics
func TestWorkerPool_EmitsJobFailedOnPanic(t *testing.T) {
	pool, manager, registry, db := setupWorkerTest(t)
	defer db.Close()

	log := zerolog.Nop()
	bus := events.NewBus(log)
	eventManager := events.NewManager(bus, log)
	pool.SetEventManager(eventManager)

	eventsChan := make(chan events.Event, 10)
	_ = bus.Subscribe(events.JobFailed, func(event *events.Event) {
		eventsChan <- *event
	})

	registry.Register(JobTypeDeployment, func(job *Job) error {
		panic("unexpected panic in job")
	})

	job := &Job{
		ID:          "test-panic",
		Type:        JobTypeDeployment,
		Priority:    PriorityHigh,
		AvailableAt: time.Now(),
	}

	manager.Enqueue(job)
	pool.Start()

	// Wait for JobFailed event from panic
	select {
	case event := <-eventsChan:
		assert.Equal(t, events.JobFailed, event.Type)

		typedData := event.GetTypedData()
		require.NotNil(t, typedData, "Event should have typed data")
		data, ok := typedData.(*events.JobStatusData)
		require.True(t, ok, "Event data should be JobStatusData")
		assert.Equal(t, "test-panic", data.JobID)
		assert.Contains(t, data.Error, "panic")
		assert.Contains(t, data.Error, "unexpected panic in job")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Expected JobFailed event from panic not received")
	}

	time.Sleep(50 * time.Millisecond)
	pool.Stop()
}

// TestWorkerPool_InjectsProgressReporter tests that ProgressReporter is injected
func TestWorkerPool_InjectsProgressReporter(t *testing.T) {
	pool, manager, registry, db := setupWorkerTest(t)
	defer db.Close()

	log := zerolog.Nop()
	bus := events.NewBus(log)
	eventManager := events.NewManager(bus, log)
	pool.SetEventManager(eventManager)

	var reporterWasInjected bool
	var wg sync.WaitGroup
	wg.Add(1)

	registry.Register(JobTypePlannerBatch, func(job *Job) error {
		// Check if progress reporter was injected
		reporter := job.GetProgressReporter()
		if reporter != nil {
			reporterWasInjected = true
		}
		wg.Done()
		return nil
	})

	job := &Job{
		ID:          "test-reporter",
		Type:        JobTypePlannerBatch,
		Priority:    PriorityHigh,
		AvailableAt: time.Now(),
	}

	manager.Enqueue(job)
	pool.Start()

	wg.Wait()
	pool.Stop()

	assert.True(t, reporterWasInjected, "ProgressReporter should have been injected into job")
}

// TestWorkerPool_NoEventsWithoutEventManager tests graceful handling when EventManager is nil
func TestWorkerPool_NoEventsWithoutEventManager(t *testing.T) {
	pool, manager, registry, db := setupWorkerTest(t)
	defer db.Close()

	// Do NOT set event manager (should be nil)

	var executed bool
	var wg sync.WaitGroup
	wg.Add(1)

	registry.Register(JobTypePlannerBatch, func(job *Job) error {
		executed = true
		wg.Done()
		return nil
	})

	job := &Job{
		ID:          "test-no-events",
		Type:        JobTypePlannerBatch,
		Priority:    PriorityHigh,
		AvailableAt: time.Now(),
	}

	manager.Enqueue(job)
	pool.Start()

	wg.Wait()
	pool.Stop()

	// Should still execute successfully even without events
	assert.True(t, executed)
}

// TestWorkerPool_ProgressReporterNilWithoutEventManager tests reporter is nil without event manager
func TestWorkerPool_ProgressReporterNilWithoutEventManager(t *testing.T) {
	pool, manager, registry, db := setupWorkerTest(t)
	defer db.Close()

	// Do NOT set event manager

	var reporterWasNil bool
	var wg sync.WaitGroup
	wg.Add(1)

	registry.Register(JobTypePlannerBatch, func(job *Job) error {
		reporter := job.GetProgressReporter()
		if reporter == nil {
			reporterWasNil = true
		}
		wg.Done()
		return nil
	})

	job := &Job{
		ID:          "test-nil-reporter",
		Type:        JobTypePlannerBatch,
		Priority:    PriorityHigh,
		AvailableAt: time.Now(),
	}

	manager.Enqueue(job)
	pool.Start()

	wg.Wait()
	pool.Stop()

	assert.True(t, reporterWasNil, "ProgressReporter should be nil without EventManager")
}

// TestWorkerPool_MultipleJobEvents tests multiple jobs emit separate events
func TestWorkerPool_MultipleJobEvents(t *testing.T) {
	pool, manager, registry, db := setupWorkerTest(t)
	defer db.Close()

	log := zerolog.Nop()
	bus := events.NewBus(log)
	eventManager := events.NewManager(bus, log)
	pool.SetEventManager(eventManager)

	startedChan := make(chan events.Event, 10)
	completedChan := make(chan events.Event, 10)
	_ = bus.Subscribe(events.JobStarted, func(event *events.Event) {
		startedChan <- *event
	})
	_ = bus.Subscribe(events.JobCompleted, func(event *events.Event) {
		completedChan <- *event
	})

	var wg sync.WaitGroup
	wg.Add(3)

	registry.Register(JobTypePlannerBatch, func(job *Job) error {
		time.Sleep(10 * time.Millisecond)
		wg.Done()
		return nil
	})

	// Enqueue 3 jobs
	for i := 0; i < 3; i++ {
		job := &Job{
			ID:          "test-multi-" + string(rune('A'+i)),
			Type:        JobTypePlannerBatch,
			Priority:    PriorityHigh,
			AvailableAt: time.Now(),
		}
		manager.Enqueue(job)
	}

	pool.Start()
	wg.Wait()

	// Should receive 3 started and 3 completed events
	startedCount := 0
	completedCount := 0

	timeout := time.After(500 * time.Millisecond)
	for startedCount < 3 || completedCount < 3 {
		select {
		case <-startedChan:
			startedCount++
		case <-completedChan:
			completedCount++
		case <-timeout:
			t.Fatalf("Timeout waiting for events. Started: %d, Completed: %d", startedCount, completedCount)
		}
	}

	assert.Equal(t, 3, startedCount, "Should have received 3 JobStarted events")
	assert.Equal(t, 3, completedCount, "Should have received 3 JobCompleted events")

	pool.Stop()
}
