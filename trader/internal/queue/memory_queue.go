package queue

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var (
	ErrQueueEmpty = errors.New("queue is empty")
)

// MemoryQueue is an in-memory priority queue
type MemoryQueue struct {
	jobs []*Job
	mu   sync.Mutex
}

// NewMemoryQueue creates a new in-memory queue
func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		jobs: make([]*Job, 0),
	}
}

// Enqueue adds a job to the queue
func (q *MemoryQueue) Enqueue(job *Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.jobs = append(q.jobs, job)
	return nil
}

// Dequeue removes and returns the highest priority available job
func (q *MemoryQueue) Dequeue() (*Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.jobs) == 0 {
		return nil, ErrQueueEmpty
	}

	now := time.Now()

	// Find available jobs (AvailableAt <= now)
	available := make([]*Job, 0)
	for _, job := range q.jobs {
		if job.AvailableAt.Before(now) || job.AvailableAt.Equal(now) {
			available = append(available, job)
		}
	}

	if len(available) == 0 {
		return nil, ErrQueueEmpty
	}

	// Sort by priority (higher priority first), then by AvailableAt
	sort.Slice(available, func(i, j int) bool {
		if available[i].Priority != available[j].Priority {
			return available[i].Priority > available[j].Priority
		}
		return available[i].AvailableAt.Before(available[j].AvailableAt)
	})

	// Remove the highest priority job from the queue
	selected := available[0]
	for i, job := range q.jobs {
		if job.ID == selected.ID {
			q.jobs = append(q.jobs[:i], q.jobs[i+1:]...)
			break
		}
	}

	return selected, nil
}

// Size returns the number of jobs in the queue
func (q *MemoryQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.jobs)
}
