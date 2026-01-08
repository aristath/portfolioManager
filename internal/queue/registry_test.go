package queue

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	registry := NewRegistry()

	handler := func(job *Job) error {
		return nil
	}

	registry.Register(JobTypePlannerBatch, handler)

	retrieved, exists := registry.Get(JobTypePlannerBatch)
	assert.True(t, exists)
	assert.NotNil(t, retrieved)
}

func TestRegistry_GetNonExistent(t *testing.T) {
	registry := NewRegistry()

	handler, exists := registry.Get(JobTypePlannerBatch)
	assert.False(t, exists)
	assert.Nil(t, handler)
}

func TestRegistry_MultipleHandlers(t *testing.T) {
	registry := NewRegistry()

	handler1 := func(job *Job) error { return nil }
	handler2 := func(job *Job) error { return errors.New("test") }

	registry.Register(JobTypePlannerBatch, handler1)
	registry.Register(JobTypeEventBasedTrading, handler2)

	h1, exists1 := registry.Get(JobTypePlannerBatch)
	h2, exists2 := registry.Get(JobTypeEventBasedTrading)

	assert.True(t, exists1)
	assert.True(t, exists2)
	assert.NotNil(t, h1)
	assert.NotNil(t, h2)

	// Test that handlers work differently
	err1 := h1(&Job{})
	err2 := h2(&Job{})

	assert.NoError(t, err1)
	assert.Error(t, err2)
}
