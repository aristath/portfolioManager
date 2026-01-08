package scheduler

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestCheckHistoryDatabasesJob_Name(t *testing.T) {
	job := &CheckHistoryDatabasesJob{
		log: zerolog.Nop(),
	}
	assert.Equal(t, "check_history_databases", job.Name())
}

func TestCheckHistoryDatabasesJob_Run_NoDatabase(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)
	job := NewCheckHistoryDatabasesJob(nil)
	job.SetLogger(log)

	err := job.Run()
	assert.NoError(t, err) // Should handle nil database gracefully
}
