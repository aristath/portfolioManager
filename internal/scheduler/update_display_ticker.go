package scheduler

import (
	"fmt"

	"github.com/rs/zerolog"
)

// UpdateDisplayTickerJob updates the LED display ticker
type UpdateDisplayTickerJob struct {
	log                 zerolog.Logger
	updateDisplayTicker func() error
}

// UpdateDisplayTickerConfig holds configuration for update display ticker job
type UpdateDisplayTickerConfig struct {
	Log                 zerolog.Logger
	UpdateDisplayTicker func() error
}

// NewUpdateDisplayTickerJob creates a new update display ticker job
func NewUpdateDisplayTickerJob(cfg UpdateDisplayTickerConfig) *UpdateDisplayTickerJob {
	return &UpdateDisplayTickerJob{
		log:                 cfg.Log.With().Str("job", "update_display_ticker").Logger(),
		updateDisplayTicker: cfg.UpdateDisplayTicker,
	}
}

// Name returns the job name
func (j *UpdateDisplayTickerJob) Name() string {
	return "update_display_ticker"
}

// Run executes the update display ticker job
func (j *UpdateDisplayTickerJob) Run() error {
	j.log.Debug().Msg("Updating display ticker")

	if j.updateDisplayTicker != nil {
		if err := j.updateDisplayTicker(); err != nil {
			j.log.Error().Err(err).Msg("Failed to update display ticker")
			return fmt.Errorf("update display ticker failed: %w", err)
		}
	} else {
		j.log.Debug().Msg("Display ticker update not configured")
	}

	return nil
}
