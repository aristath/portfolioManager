package queue

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Scheduler enqueues time-based jobs
type Scheduler struct {
	manager *Manager
	stop    chan struct{}
	log     zerolog.Logger
	stopped bool
	started bool
	mu      sync.Mutex
}

// NewScheduler creates a new time-based scheduler
func NewScheduler(manager *Manager) *Scheduler {
	return &Scheduler{
		manager: manager,
		stop:    make(chan struct{}),
		log:     zerolog.Nop(),
	}
}

// SetLogger sets the logger for the scheduler
func (s *Scheduler) SetLogger(log zerolog.Logger) {
	s.log = log.With().Str("component", "time_scheduler").Logger()
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Prevent multiple starts
	if s.started && !s.stopped {
		s.log.Warn().Msg("Time scheduler already started, ignoring")
		return
	}

	if s.stopped {
		// Reset stop channel if it was stopped
		s.stop = make(chan struct{})
		s.stopped = false
	}

	s.started = true
	s.log.Info().Msg("Time scheduler started")

	// Hourly jobs (every hour at :00)
	hourlyTicker := time.NewTicker(1 * time.Hour)
	go func() {
		// Run immediately on start, then every hour
		s.enqueueTimeBasedJob(JobTypeHourlyBackup, PriorityMedium, 1*time.Hour)
		for {
			select {
			case <-s.stop:
				hourlyTicker.Stop()
				return
			case <-hourlyTicker.C:
				s.enqueueTimeBasedJob(JobTypeHourlyBackup, PriorityMedium, 1*time.Hour)
			}
		}
	}()

	// Daily jobs (check every minute, enqueue at specific times)
	dailyTicker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-s.stop:
				dailyTicker.Stop()
				return
			case now := <-dailyTicker.C:
				hour := now.Hour()
				minute := now.Minute()

				// Health check: Daily at 4:00 AM
				if hour == 4 && minute == 0 {
					s.enqueueTimeBasedJob(JobTypeHealthCheck, PriorityMedium, 24*time.Hour)
				}

				// Daily backup: Daily at 1:00 AM
				if hour == 1 && minute == 0 {
					s.enqueueTimeBasedJob(JobTypeDailyBackup, PriorityMedium, 24*time.Hour)
				}

				// Daily maintenance: Daily at 2:00 AM
				if hour == 2 && minute == 0 {
					s.enqueueTimeBasedJob(JobTypeDailyMaintenance, PriorityMedium, 24*time.Hour)
				}

				// Adaptive market check: Daily at 6:00 AM
				if hour == 6 && minute == 0 {
					s.enqueueTimeBasedJob(JobTypeAdaptiveMarket, PriorityMedium, 24*time.Hour)
				}

				// History cleanup: Daily at midnight (00:00)
				if hour == 0 && minute == 0 {
					s.enqueueTimeBasedJob(JobTypeHistoryCleanup, PriorityMedium, 24*time.Hour)
				}

				// Dividend reinvestment: Daily at 10:00 AM
				if hour == 10 && minute == 0 {
					s.enqueueTimeBasedJob(JobTypeDividendReinvest, PriorityHigh, 24*time.Hour)
				}

				// Sync cycle fallback: Every 30 minutes
				if minute%30 == 0 {
					s.enqueueTimeBasedJob(JobTypeSyncCycle, PriorityHigh, 30*time.Minute)
				}
			}
		}
	}()

	// Weekly jobs (check every minute, enqueue on Sunday at specific times)
	weeklyTicker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-s.stop:
				weeklyTicker.Stop()
				return
			case now := <-weeklyTicker.C:
				if now.Weekday() == time.Sunday {
					hour := now.Hour()
					minute := now.Minute()

					// Weekly backup: Sunday at 1:00 AM
					if hour == 1 && minute == 0 {
						s.enqueueTimeBasedJob(JobTypeWeeklyBackup, PriorityMedium, 7*24*time.Hour)
					}

					// Weekly maintenance: Sunday at 3:30 AM
					if hour == 3 && minute == 30 {
						s.enqueueTimeBasedJob(JobTypeWeeklyMaintenance, PriorityMedium, 7*24*time.Hour)
					}
				}
			}
		}
	}()

	// Monthly jobs (check every minute, enqueue on 1st at specific times)
	monthlyTicker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-s.stop:
				monthlyTicker.Stop()
				return
			case now := <-monthlyTicker.C:
				if now.Day() == 1 {
					hour := now.Hour()
					minute := now.Minute()

					// Monthly backup: 1st at 1:00 AM
					if hour == 1 && minute == 0 {
						s.enqueueTimeBasedJob(JobTypeMonthlyBackup, PriorityMedium, 30*24*time.Hour)
					}

					// Monthly maintenance: 1st at 4:00 AM
					if hour == 4 && minute == 0 {
						s.enqueueTimeBasedJob(JobTypeMonthlyMaintenance, PriorityMedium, 30*24*time.Hour)
					}

					// Formula discovery: 1st at 5:00 AM
					if hour == 5 && minute == 0 {
						s.enqueueTimeBasedJob(JobTypeFormulaDiscovery, PriorityMedium, 30*24*time.Hour)
					}
				}
			}
		}
	}()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.stopped {
		close(s.stop)
		s.stopped = true
		s.started = false
		s.log.Info().Msg("Time scheduler stopped")
	}
}

// enqueueTimeBasedJob enqueues a job if the interval has passed
func (s *Scheduler) enqueueTimeBasedJob(jobType JobType, priority Priority, interval time.Duration) bool {
	enqueued := s.manager.EnqueueIfShouldRun(jobType, priority, interval, map[string]interface{}{})
	if enqueued {
		s.log.Info().
			Str("job_type", string(jobType)).
			Dur("interval", interval).
			Msg("Enqueued time-based job")
	} else {
		s.log.Debug().
			Str("job_type", string(jobType)).
			Dur("interval", interval).
			Msg("Skipped time-based job (interval not yet passed)")
	}
	return enqueued
}
