package calculations

import (
	"time"

	"github.com/rs/zerolog"
)

// TagUpdateService defines the interface for checking and updating tags.
// This is implemented by scheduler.TagUpdateJob (will need method exposure).
type TagUpdateService interface {
	// GetTagsWithUpdateTimes returns current tag IDs with their last update times for a security
	GetTagsWithUpdateTimes(symbol string) (map[string]time.Time, error)

	// UpdateTagsForSecurity updates all tags for a single security
	UpdateTagsForSecurity(symbol string) error
}

// TagFrequencyChecker checks tag staleness based on per-tag update frequencies.
// This is a standalone function type to avoid coupling to scheduler package.
type TagFrequencyChecker func(currentTags map[string]time.Time, now time.Time) map[string]bool

// DefaultTagProcessor implements TagProcessor using existing tag update logic.
// It checks if a security has any stale tags and delegates updates to the
// underlying TagUpdateService.
type DefaultTagProcessor struct {
	tagService     TagUpdateService
	frequencyCheck TagFrequencyChecker
	log            zerolog.Logger
}

// NewDefaultTagProcessor creates a new tag processor
func NewDefaultTagProcessor(tagService TagUpdateService, frequencyCheck TagFrequencyChecker, log zerolog.Logger) *DefaultTagProcessor {
	return &DefaultTagProcessor{
		tagService:     tagService,
		frequencyCheck: frequencyCheck,
		log:            log.With().Str("component", "tag_processor").Logger(),
	}
}

// NeedsTagUpdate checks if a security has any tags that need updating.
// Uses per-tag update frequencies defined in tag_update_frequencies.go.
func (tp *DefaultTagProcessor) NeedsTagUpdate(symbol string) bool {
	// Get current tags with their update times
	currentTags, err := tp.tagService.GetTagsWithUpdateTimes(symbol)
	if err != nil {
		tp.log.Debug().
			Err(err).
			Str("symbol", symbol).
			Msg("Failed to get tags with update times, assuming needs update")
		return true // Assume needs update on error
	}

	// Check which tags need updating based on their frequencies
	tagsNeedingUpdate := tp.frequencyCheck(currentTags, time.Now())

	return len(tagsNeedingUpdate) > 0
}

// ProcessTagUpdate updates tags for a single security.
// This delegates to TagUpdateService.UpdateTagsForSecurity which handles:
// - Getting score and market data
// - Assigning tags based on current conditions
// - Updating security_tags table
func (tp *DefaultTagProcessor) ProcessTagUpdate(symbol string) error {
	return tp.tagService.UpdateTagsForSecurity(symbol)
}
