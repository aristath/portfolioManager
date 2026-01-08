package scheduler

import (
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestUpdateDisplayTickerJob_Name(t *testing.T) {
	job := &UpdateDisplayTickerJob{
		log: zerolog.Nop(),
	}

	assert.Equal(t, "update_display_ticker", job.Name())
}

func TestUpdateDisplayTickerJob_Run_Success(t *testing.T) {
	// Setup
	log := zerolog.New(nil).Level(zerolog.Disabled)
	updateCalled := false

	job := NewUpdateDisplayTickerJob(UpdateDisplayTickerConfig{
		Log: log,
		UpdateDisplayTicker: func() error {
			updateCalled = true
			return nil
		},
	})

	// Execute
	err := job.Run()

	// Assert
	assert.NoError(t, err)
	assert.True(t, updateCalled, "UpdateDisplayTicker should have been called")
}

func TestUpdateDisplayTickerJob_Run_CallbackError(t *testing.T) {
	// Setup
	log := zerolog.New(nil).Level(zerolog.Disabled)
	updateCalled := false

	job := NewUpdateDisplayTickerJob(UpdateDisplayTickerConfig{
		Log: log,
		UpdateDisplayTicker: func() error {
			updateCalled = true
			return errors.New("ticker update failed")
		},
	})

	// Execute - should not panic, just log error
	err := job.Run()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update display ticker failed")
	assert.True(t, updateCalled, "UpdateDisplayTicker should have been called")
}

func TestUpdateDisplayTickerJob_Run_NoCallback(t *testing.T) {
	// Setup
	log := zerolog.New(nil).Level(zerolog.Disabled)

	job := NewUpdateDisplayTickerJob(UpdateDisplayTickerConfig{
		Log:                 log,
		UpdateDisplayTicker: nil,
	})

	// Execute - should not panic
	err := job.Run()

	// Assert
	assert.NoError(t, err) // Non-critical, don't fail
}
