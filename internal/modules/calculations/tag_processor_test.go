package calculations

import (
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTagUpdateService implements TagUpdateService for testing
type mockTagUpdateService struct {
	tagsWithTimes  map[string]map[string]time.Time // symbol -> (tagID -> updateTime)
	updatedSymbols []string
	getErr         error
	updateErr      error
}

func (m *mockTagUpdateService) GetTagsWithUpdateTimes(symbol string) (map[string]time.Time, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.tagsWithTimes == nil {
		return make(map[string]time.Time), nil
	}
	tags, ok := m.tagsWithTimes[symbol]
	if !ok {
		return make(map[string]time.Time), nil
	}
	return tags, nil
}

func (m *mockTagUpdateService) UpdateTagsForSecurity(symbol string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedSymbols = append(m.updatedSymbols, symbol)
	return nil
}

// mockFrequencyChecker returns a function that checks tag staleness
func mockFrequencyChecker(staleTags map[string]bool) TagFrequencyChecker {
	return func(currentTags map[string]time.Time, now time.Time) map[string]bool {
		return staleTags
	}
}

func TestDefaultTagProcessor_NeedsTagUpdate_NoTags(t *testing.T) {
	service := &mockTagUpdateService{
		tagsWithTimes: map[string]map[string]time.Time{},
	}

	// Frequency checker returns stale tags for empty tag set (new security)
	checker := mockFrequencyChecker(map[string]bool{"tag1": true})

	processor := NewDefaultTagProcessor(service, checker, zerolog.Nop())

	assert.True(t, processor.NeedsTagUpdate("TEST"), "Should need update when no tags exist")
}

func TestDefaultTagProcessor_NeedsTagUpdate_FreshTags(t *testing.T) {
	now := time.Now()
	service := &mockTagUpdateService{
		tagsWithTimes: map[string]map[string]time.Time{
			"TEST": {
				"high-quality": now.Add(-1 * time.Hour), // Recent
				"stable":       now.Add(-2 * time.Hour), // Recent
			},
		},
	}

	// Frequency checker returns no stale tags (all fresh)
	checker := mockFrequencyChecker(map[string]bool{})

	processor := NewDefaultTagProcessor(service, checker, zerolog.Nop())

	assert.False(t, processor.NeedsTagUpdate("TEST"), "Should not need update when all tags fresh")
}

func TestDefaultTagProcessor_NeedsTagUpdate_StaleTags(t *testing.T) {
	now := time.Now()
	service := &mockTagUpdateService{
		tagsWithTimes: map[string]map[string]time.Time{
			"TEST": {
				"high-quality": now.Add(-25 * time.Hour), // Stale (24h frequency)
				"stable":       now.Add(-1 * time.Hour),  // Fresh
			},
		},
	}

	// Frequency checker returns high-quality as stale
	checker := mockFrequencyChecker(map[string]bool{"high-quality": true})

	processor := NewDefaultTagProcessor(service, checker, zerolog.Nop())

	assert.True(t, processor.NeedsTagUpdate("TEST"), "Should need update when some tags are stale")
}

func TestDefaultTagProcessor_NeedsTagUpdate_GetError(t *testing.T) {
	service := &mockTagUpdateService{
		getErr: errors.New("database error"),
	}

	checker := mockFrequencyChecker(map[string]bool{})

	processor := NewDefaultTagProcessor(service, checker, zerolog.Nop())

	// Should return true on error (assume needs update)
	assert.True(t, processor.NeedsTagUpdate("TEST"), "Should assume needs update on error")
}

func TestDefaultTagProcessor_ProcessTagUpdate_Success(t *testing.T) {
	service := &mockTagUpdateService{}
	checker := mockFrequencyChecker(map[string]bool{})

	processor := NewDefaultTagProcessor(service, checker, zerolog.Nop())

	err := processor.ProcessTagUpdate("TEST")
	require.NoError(t, err)
	assert.Contains(t, service.updatedSymbols, "TEST")
}

func TestDefaultTagProcessor_ProcessTagUpdate_Error(t *testing.T) {
	service := &mockTagUpdateService{
		updateErr: errors.New("update failed"),
	}
	checker := mockFrequencyChecker(map[string]bool{})

	processor := NewDefaultTagProcessor(service, checker, zerolog.Nop())

	err := processor.ProcessTagUpdate("TEST")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
}

func TestDefaultTagProcessor_ProcessTagUpdate_MultipleSecurities(t *testing.T) {
	service := &mockTagUpdateService{}
	checker := mockFrequencyChecker(map[string]bool{})

	processor := NewDefaultTagProcessor(service, checker, zerolog.Nop())

	symbols := []string{"SYM1", "SYM2", "SYM3"}
	for _, sym := range symbols {
		err := processor.ProcessTagUpdate(sym)
		require.NoError(t, err)
	}

	assert.Len(t, service.updatedSymbols, 3)
	assert.Contains(t, service.updatedSymbols, "SYM1")
	assert.Contains(t, service.updatedSymbols, "SYM2")
	assert.Contains(t, service.updatedSymbols, "SYM3")
}

func TestDefaultTagProcessor_NeedsTagUpdate_MixedStaleness(t *testing.T) {
	now := time.Now()

	// Create multiple securities with different staleness
	service := &mockTagUpdateService{
		tagsWithTimes: map[string]map[string]time.Time{
			"FRESH": {
				"high-quality": now.Add(-1 * time.Hour),
			},
			"STALE": {
				"high-quality": now.Add(-25 * time.Hour),
			},
		},
	}

	// Create different checker responses per call
	callCount := 0
	checker := func(currentTags map[string]time.Time, now time.Time) map[string]bool {
		callCount++
		// First call (FRESH) - no stale tags
		if callCount == 1 {
			return map[string]bool{}
		}
		// Second call (STALE) - has stale tags
		return map[string]bool{"high-quality": true}
	}

	processor := NewDefaultTagProcessor(service, checker, zerolog.Nop())

	assert.False(t, processor.NeedsTagUpdate("FRESH"), "FRESH should not need update")
	assert.True(t, processor.NeedsTagUpdate("STALE"), "STALE should need update")
}
