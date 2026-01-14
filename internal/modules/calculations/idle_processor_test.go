package calculations

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// mockQueueSizer implements QueueSizer for testing
type mockQueueSizer struct {
	size int
}

func (m *mockQueueSizer) Size() int {
	return m.size
}

func (m *mockQueueSizer) SetSize(size int) {
	m.size = size
}

// mockSecurityProvider implements SecurityProvider for testing
type mockSecurityProvider struct {
	securities []SecurityInfo
	err        error
}

func (m *mockSecurityProvider) GetActiveSecurities() ([]SecurityInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.securities, nil
}

// mockSyncProcessor implements SyncProcessor for testing
type mockSyncProcessor struct {
	needsSync    map[string]bool
	processedIDs []string
	err          error
}

func (m *mockSyncProcessor) NeedsSync(security SecurityInfo) bool {
	if m.needsSync == nil {
		return false
	}
	return m.needsSync[security.ISIN]
}

func (m *mockSyncProcessor) ProcessSync(security SecurityInfo) error {
	if m.err != nil {
		return m.err
	}
	m.processedIDs = append(m.processedIDs, security.ISIN)
	return nil
}

// mockTagProcessor implements TagProcessor for testing
type mockTagProcessor struct {
	needsUpdate  map[string]bool
	processedIDs []string
	err          error
}

func (m *mockTagProcessor) NeedsTagUpdate(symbol string) bool {
	if m.needsUpdate == nil {
		return false
	}
	return m.needsUpdate[symbol]
}

func (m *mockTagProcessor) ProcessTagUpdate(symbol string) error {
	if m.err != nil {
		return m.err
	}
	m.processedIDs = append(m.processedIDs, symbol)
	return nil
}

func setupIdleProcessorDeps(t *testing.T) IdleProcessorDeps {
	cache := setupTestCache(t)

	return IdleProcessorDeps{
		Cache: cache,
		Queue: &mockQueueSizer{size: 0}, // Idle by default
		SecurityProvider: &mockSecurityProvider{
			securities: []SecurityInfo{
				{ISIN: "ISIN1", Symbol: "SYM1", LastSynced: nil},
				{ISIN: "ISIN2", Symbol: "SYM2", LastSynced: nil},
			},
		},
		PriceProvider: &mockPriceProvider{
			prices: map[string][]float64{
				"ISIN1": generatePrices(300),
				"ISIN2": generatePrices(300),
			},
		},
		SyncProcessor: &mockSyncProcessor{needsSync: map[string]bool{}},
		TagProcessor:  &mockTagProcessor{needsUpdate: map[string]bool{}},
		Log:           zerolog.Nop(),
	}
}

func TestIdleProcessor_ProcessOne_WhenIdle(t *testing.T) {
	deps := setupIdleProcessorDeps(t)
	processor := NewIdleProcessor(deps)

	// Process one - should process first security needing work
	processed := processor.ProcessOne()
	assert.True(t, processed, "Should process work when idle")
}

func TestIdleProcessor_ProcessOne_WhenBusy(t *testing.T) {
	deps := setupIdleProcessorDeps(t)
	deps.Queue.(*mockQueueSizer).SetSize(5) // Busy

	processor := NewIdleProcessor(deps)

	// Process one - should do nothing when busy
	processed := processor.ProcessOne()
	assert.False(t, processed, "Should not process work when busy")
}

func TestIdleProcessor_ProcessesTechnicalFirst(t *testing.T) {
	deps := setupIdleProcessorDeps(t)
	processor := NewIdleProcessor(deps)

	// Process one
	processed := processor.ProcessOne()
	assert.True(t, processed)

	// Should have processed technical for ISIN1
	_, ok := deps.Cache.GetTechnical("ISIN1", "ema", 200)
	assert.True(t, ok, "Should have calculated technical metrics for ISIN1")
}

func TestIdleProcessor_SkipsFreshTechnical(t *testing.T) {
	deps := setupIdleProcessorDeps(t)

	// Pre-cache ISIN1 technical metrics
	deps.Cache.SetTechnical("ISIN1", "ema", 200, 100.0, time.Hour)

	processor := NewIdleProcessor(deps)

	// Process one - should skip ISIN1 and process ISIN2
	processed := processor.ProcessOne()
	assert.True(t, processed)

	// ISIN2 should now have cached technical metrics
	_, ok := deps.Cache.GetTechnical("ISIN2", "ema", 200)
	assert.True(t, ok, "Should have calculated technical metrics for ISIN2")
}

func TestIdleProcessor_ProcessesSyncAfterTechnical(t *testing.T) {
	deps := setupIdleProcessorDeps(t)

	// Pre-cache all technical metrics so sync is the next work type
	deps.Cache.SetTechnical("ISIN1", "ema", 200, 100.0, time.Hour)
	deps.Cache.SetTechnical("ISIN2", "ema", 200, 100.0, time.Hour)

	// Set up sync processor to need work for ISIN1
	syncProcessor := deps.SyncProcessor.(*mockSyncProcessor)
	syncProcessor.needsSync = map[string]bool{"ISIN1": true}

	processor := NewIdleProcessor(deps)

	// Process one - should process sync for ISIN1
	processed := processor.ProcessOne()
	assert.True(t, processed)
	assert.Contains(t, syncProcessor.processedIDs, "ISIN1", "Should have synced ISIN1")
}

func TestIdleProcessor_ProcessesTagsAfterSync(t *testing.T) {
	deps := setupIdleProcessorDeps(t)

	// Pre-cache all technical metrics
	deps.Cache.SetTechnical("ISIN1", "ema", 200, 100.0, time.Hour)
	deps.Cache.SetTechnical("ISIN2", "ema", 200, 100.0, time.Hour)

	// Sync doesn't need work
	syncProcessor := deps.SyncProcessor.(*mockSyncProcessor)
	syncProcessor.needsSync = map[string]bool{}

	// Tags need work for SYM1
	tagProcessor := deps.TagProcessor.(*mockTagProcessor)
	tagProcessor.needsUpdate = map[string]bool{"SYM1": true}

	processor := NewIdleProcessor(deps)

	// Process one - should process tags for SYM1
	processed := processor.ProcessOne()
	assert.True(t, processed)
	assert.Contains(t, tagProcessor.processedIDs, "SYM1", "Should have updated tags for SYM1")
}

func TestIdleProcessor_NoWorkNeeded(t *testing.T) {
	deps := setupIdleProcessorDeps(t)

	// Pre-cache all technical metrics
	deps.Cache.SetTechnical("ISIN1", "ema", 200, 100.0, time.Hour)
	deps.Cache.SetTechnical("ISIN2", "ema", 200, 100.0, time.Hour)

	// Sync doesn't need work
	syncProcessor := deps.SyncProcessor.(*mockSyncProcessor)
	syncProcessor.needsSync = map[string]bool{}

	// Tags don't need work
	tagProcessor := deps.TagProcessor.(*mockTagProcessor)
	tagProcessor.needsUpdate = map[string]bool{}

	processor := NewIdleProcessor(deps)

	// Process one - should find nothing to do
	processed := processor.ProcessOne()
	assert.False(t, processed, "Should return false when no work needed")
}

func TestIdleProcessor_GetStats(t *testing.T) {
	deps := setupIdleProcessorDeps(t)
	processor := NewIdleProcessor(deps)

	// Process one - technical
	processor.ProcessOne()

	stats := processor.GetStats()
	assert.Equal(t, int64(1), stats.TechnicalProcessed, "Should have processed 1 technical")
	assert.Equal(t, int64(0), stats.SyncProcessed)
	assert.Equal(t, int64(0), stats.TagsProcessed)
}

func TestIdleProcessor_MultipleProcessCalls(t *testing.T) {
	deps := setupIdleProcessorDeps(t)
	processor := NewIdleProcessor(deps)

	// Process multiple times
	processor.ProcessOne() // ISIN1 technical
	processor.ProcessOne() // ISIN2 technical
	processor.ProcessOne() // Nothing left

	stats := processor.GetStats()
	assert.Equal(t, int64(2), stats.TechnicalProcessed, "Should have processed 2 technical")

	// Both should now be cached
	_, ok := deps.Cache.GetTechnical("ISIN1", "ema", 200)
	assert.True(t, ok)
	_, ok = deps.Cache.GetTechnical("ISIN2", "ema", 200)
	assert.True(t, ok)
}

func TestIdleProcessor_StartAndStop(t *testing.T) {
	deps := setupIdleProcessorDeps(t)
	processor := NewIdleProcessor(deps)

	// Start the processor
	processor.Start()

	// Give it a moment to run
	time.Sleep(50 * time.Millisecond)

	// Stop the processor
	processor.Stop()

	// Should not panic and should be stoppable multiple times
	processor.Stop()
}

func TestIdleProcessor_NilProcessors(t *testing.T) {
	deps := setupIdleProcessorDeps(t)
	deps.SyncProcessor = nil
	deps.TagProcessor = nil

	processor := NewIdleProcessor(deps)

	// Should still process technical metrics without sync/tag processors
	processed := processor.ProcessOne()
	assert.True(t, processed)
}
