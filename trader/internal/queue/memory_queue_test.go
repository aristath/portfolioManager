package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryQueue_EnqueueDequeue(t *testing.T) {
	q := NewMemoryQueue()

	job := &Job{
		ID:          "test-1",
		Type:        JobTypePlannerBatch,
		Priority:    PriorityHigh,
		Payload:     map[string]interface{}{"test": "data"},
		CreatedAt:   time.Now(),
		AvailableAt: time.Now(),
		Retries:     0,
		MaxRetries:  3,
	}

	err := q.Enqueue(job)
	assert.NoError(t, err)
	assert.Equal(t, 1, q.Size())

	dequeued, err := q.Dequeue()
	assert.NoError(t, err)
	assert.NotNil(t, dequeued)
	assert.Equal(t, "test-1", dequeued.ID)
	assert.Equal(t, JobTypePlannerBatch, dequeued.Type)
	assert.Equal(t, 0, q.Size())
}

func TestMemoryQueue_PriorityOrdering(t *testing.T) {
	q := NewMemoryQueue()

	low := &Job{ID: "low", Type: JobTypeTagUpdate, Priority: PriorityLow, AvailableAt: time.Now()}
	high := &Job{ID: "high", Type: JobTypePlannerBatch, Priority: PriorityHigh, AvailableAt: time.Now()}
	critical := &Job{ID: "critical", Type: JobTypeEventBasedTrading, Priority: PriorityCritical, AvailableAt: time.Now()}
	medium := &Job{ID: "medium", Type: JobTypeTagUpdate, Priority: PriorityMedium, AvailableAt: time.Now()}

	q.Enqueue(low)
	q.Enqueue(high)
	q.Enqueue(critical)
	q.Enqueue(medium)

	// Should dequeue in priority order: Critical, High, Medium, Low
	dequeued, _ := q.Dequeue()
	assert.Equal(t, "critical", dequeued.ID)

	dequeued, _ = q.Dequeue()
	assert.Equal(t, "high", dequeued.ID)

	dequeued, _ = q.Dequeue()
	assert.Equal(t, "medium", dequeued.ID)

	dequeued, _ = q.Dequeue()
	assert.Equal(t, "low", dequeued.ID)
}

func TestMemoryQueue_AvailableAt(t *testing.T) {
	q := NewMemoryQueue()

	future := &Job{
		ID:          "future",
		Type:        JobTypePlannerBatch,
		Priority:    PriorityHigh,
		AvailableAt: time.Now().Add(1 * time.Hour),
	}

	now := &Job{
		ID:          "now",
		Type:        JobTypePlannerBatch,
		Priority:    PriorityHigh,
		AvailableAt: time.Now(),
	}

	q.Enqueue(future)
	q.Enqueue(now)

	// Should dequeue "now" first even though "future" was enqueued first
	dequeued, _ := q.Dequeue()
	assert.Equal(t, "now", dequeued.ID)

	// Future job should still be in queue
	assert.Equal(t, 1, q.Size())
}

func TestMemoryQueue_EmptyDequeue(t *testing.T) {
	q := NewMemoryQueue()

	job, err := q.Dequeue()
	assert.Error(t, err)
	assert.Nil(t, job)
}

func TestMemoryQueue_Size(t *testing.T) {
	q := NewMemoryQueue()

	assert.Equal(t, 0, q.Size())

	q.Enqueue(&Job{ID: "1", Type: JobTypePlannerBatch, Priority: PriorityHigh, AvailableAt: time.Now()})
	assert.Equal(t, 1, q.Size())

	q.Enqueue(&Job{ID: "2", Type: JobTypePlannerBatch, Priority: PriorityHigh, AvailableAt: time.Now()})
	assert.Equal(t, 2, q.Size())

	q.Dequeue()
	assert.Equal(t, 1, q.Size())
}
