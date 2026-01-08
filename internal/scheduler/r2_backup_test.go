package scheduler

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
)

// TestR2BackupJob_TimeChecking tests time-based backup logic
func TestR2BackupJob_TimeChecking(t *testing.T) {
	log := zerolog.Nop()

	tests := []struct {
		name        string
		lastBackup  time.Time
		expectDaily bool
	}{
		{
			name:        "never backed up should run",
			lastBackup:  time.Time{}, // Zero time
			expectDaily: true,
		},
		{
			name:        "backed up 1 hour ago should not run",
			lastBackup:  time.Now().Add(-1 * time.Hour),
			expectDaily: false,
		},
		{
			name:        "backed up 24 hours ago should run",
			lastBackup:  time.Now().Add(-24 * time.Hour),
			expectDaily: true,
		},
		{
			name:        "backed up 48 hours ago should run",
			lastBackup:  time.Now().Add(-48 * time.Hour),
			expectDaily: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &R2BackupJob{
				log:        log,
				lastBackup: tt.lastBackup,
			}

			shouldRun := job.shouldRunDaily()

			if shouldRun != tt.expectDaily {
				t.Errorf("Expected shouldRunDaily()=%v, got %v", tt.expectDaily, shouldRun)
			}
		})
	}
}

// TestR2BackupJob_HasBeenMoreThan tests duration checking
func TestR2BackupJob_HasBeenMoreThan(t *testing.T) {
	log := zerolog.Nop()

	tests := []struct {
		name       string
		lastBackup time.Time
		duration   time.Duration
		expected   bool
	}{
		{
			name:       "zero time always returns true",
			lastBackup: time.Time{},
			duration:   1 * time.Hour,
			expected:   true,
		},
		{
			name:       "30 minutes ago, checking for 1 hour should be false",
			lastBackup: time.Now().Add(-30 * time.Minute),
			duration:   1 * time.Hour,
			expected:   false,
		},
		{
			name:       "2 hours ago, checking for 1 hour should be true",
			lastBackup: time.Now().Add(-2 * time.Hour),
			duration:   1 * time.Hour,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &R2BackupJob{
				log:        log,
				lastBackup: tt.lastBackup,
			}

			result := job.hasBeenMoreThan(tt.duration)

			if result != tt.expected {
				t.Errorf("Expected hasBeenMoreThan()=%v, got %v", tt.expected, result)
			}
		})
	}
}

// TestR2BackupJob_Name tests job name
func TestR2BackupJob_Name(t *testing.T) {
	job := &R2BackupJob{}
	name := job.Name()

	if name != "r2_backup" {
		t.Errorf("Expected job name 'r2_backup', got '%s'", name)
	}
}

// TestR2BackupRotationJob_Name tests rotation job name
func TestR2BackupRotationJob_Name(t *testing.T) {
	job := &R2BackupRotationJob{}
	name := job.Name()

	if name != "r2_backup_rotation" {
		t.Errorf("Expected job name 'r2_backup_rotation', got '%s'", name)
	}
}
