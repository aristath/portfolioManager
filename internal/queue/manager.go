// Package queue provides an event queue system for asynchronous job processing.
package queue

import (
	"fmt"
	"time"
)

// Manager coordinates queue operations and history tracking
type Manager struct {
	queue   *MemoryQueue
	history *History
}

// NewManager creates a new queue manager
func NewManager(queue *MemoryQueue, history *History) *Manager {
	return &Manager{
		queue:   queue,
		history: history,
	}
}

// Enqueue adds a job to the queue
func (m *Manager) Enqueue(job *Job) error {
	return m.queue.Enqueue(job)
}

// EnqueueIfShouldRun checks history and enqueues if interval has passed
func (m *Manager) EnqueueIfShouldRun(jobType JobType, priority Priority, interval time.Duration, payload map[string]interface{}) bool {
	if !m.history.ShouldRun(jobType, interval) {
		return false
	}

	job := &Job{
		ID:          fmt.Sprintf("%s-%d", jobType, time.Now().UnixNano()),
		Type:        jobType,
		Priority:    priority,
		Payload:     payload,
		CreatedAt:   time.Now(),
		AvailableAt: time.Now(),
		Retries:     0,
		MaxRetries:  3,
	}

	if err := m.queue.Enqueue(job); err != nil {
		return false
	}

	return true
}

// Dequeue removes and returns the highest priority job
func (m *Manager) Dequeue() (*Job, error) {
	return m.queue.Dequeue()
}

// Size returns the number of jobs in the queue
func (m *Manager) Size() int {
	return m.queue.Size()
}

// RecordExecution records a job execution in history
func (m *Manager) RecordExecution(jobType JobType, status string) error {
	return m.history.RecordExecution(jobType, time.Now(), status)
}
