package queue

import (
	"github.com/aristath/portfolioManager/internal/scheduler"
)

// JobToHandler converts a scheduler.Job to a queue.Handler
func JobToHandler(job scheduler.Job) Handler {
	return func(queueJob *Job) error {
		return job.Run()
	}
}
