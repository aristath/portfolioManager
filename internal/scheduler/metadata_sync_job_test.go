package scheduler

import (
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMetadataSyncService is a mock for MetadataSyncService
type mockMetadataSyncService struct {
	isins        []string
	successCount int
	err          error
}

func (m *mockMetadataSyncService) GetAllActiveISINs() []string {
	return m.isins
}

func (m *mockMetadataSyncService) SyncMetadataBatch(isins []string) (int, error) {
	return m.successCount, m.err
}

func TestMetadataSyncJob_Run_Success(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockService := &mockMetadataSyncService{
		isins:        []string{"US0000000001", "US0000000002", "US0000000003"},
		successCount: 3,
		err:          nil,
	}

	job := NewMetadataSyncJob(mockService, log)

	err := job.Run()
	assert.NoError(t, err, "Job should complete without error")
}

func TestMetadataSyncJob_Run_NoISINs(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockService := &mockMetadataSyncService{
		isins:        []string{},
		successCount: 0,
		err:          nil,
	}

	job := NewMetadataSyncJob(mockService, log)

	err := job.Run()
	assert.NoError(t, err, "Job should complete without error even with no ISINs")
}

func TestMetadataSyncJob_Run_BatchFailure(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockService := &mockMetadataSyncService{
		isins:        []string{"US0000000001", "US0000000002"},
		successCount: 0,
		err:          errors.New("batch API failed"),
	}

	job := NewMetadataSyncJob(mockService, log)

	// Should NOT propagate error - work processor handles retries
	err := job.Run()
	assert.NoError(t, err, "Job should not propagate batch sync errors")
}

func TestMetadataSyncJob_Run_PartialSuccess(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockService := &mockMetadataSyncService{
		isins:        []string{"US0000000001", "US0000000002", "US0000000003"},
		successCount: 2, // Only 2 of 3 succeeded
		err:          nil,
	}

	job := NewMetadataSyncJob(mockService, log)

	err := job.Run()
	assert.NoError(t, err, "Job should complete without error even with partial success")
}

func TestMetadataSyncJob_Name(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockService := &mockMetadataSyncService{}
	job := NewMetadataSyncJob(mockService, log)

	assert.Equal(t, "metadata_batch_sync", job.Name())
}

func TestMetadataSyncJob_Creation(t *testing.T) {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	mockService := &mockMetadataSyncService{}
	job := NewMetadataSyncJob(mockService, log)

	require.NotNil(t, job)
	assert.NotNil(t, job.metadataService)
}
