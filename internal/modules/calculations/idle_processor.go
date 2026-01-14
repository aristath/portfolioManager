package calculations

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	// IdleQueueThreshold is the queue size below which the system is considered idle
	IdleQueueThreshold = 2

	// CheckInterval is how often the idle processor checks for work
	CheckInterval = 30 * time.Second
)

// IdleJobType represents the type of idle work for event emission
type IdleJobType string

const (
	IdleJobTypeTechnical IdleJobType = "idle_technical"
	IdleJobTypeSync      IdleJobType = "idle_sync"
	IdleJobTypeTags      IdleJobType = "idle_tags"
)

// JobEventEmitter emits job lifecycle events for visibility in the Activity UI.
// This interface matches the queue.WorkerPool event emission pattern.
type JobEventEmitter interface {
	EmitJobStarted(jobType string, jobID string, description string)
	EmitJobCompleted(jobType string, jobID string, description string)
	EmitJobFailed(jobType string, jobID string, description string, err error)
}

// QueueSizer returns the current queue size
type QueueSizer interface {
	Size() int
}

// SecurityProvider provides list of active securities
type SecurityProvider interface {
	GetActiveSecurities() ([]SecurityInfo, error)
}

// SecurityInfo holds minimal security data for idle processing
type SecurityInfo struct {
	ISIN       string
	Symbol     string
	LastSynced *int64
}

// SyncProcessor handles per-security historical sync
type SyncProcessor interface {
	NeedsSync(security SecurityInfo) bool
	ProcessSync(security SecurityInfo) error
}

// TagProcessor handles per-security tag updates
type TagProcessor interface {
	NeedsTagUpdate(symbol string) bool
	ProcessTagUpdate(symbol string) error
}

// IdleProcessorDeps holds dependencies for the idle processor
type IdleProcessorDeps struct {
	Cache            *Cache
	Queue            QueueSizer
	SecurityProvider SecurityProvider
	PriceProvider    PriceProvider
	SyncProcessor    SyncProcessor
	TagProcessor     TagProcessor
	EventEmitter     JobEventEmitter // Optional: for Activity UI visibility
	Log              zerolog.Logger
}

// IdleProcessorStats tracks processing statistics
type IdleProcessorStats struct {
	TechnicalProcessed int64
	SyncProcessed      int64
	TagsProcessed      int64
}

// IdleProcessor performs staggered per-security work during idle time.
// It handles three types of work in priority order:
// 1. Technical indicator calculations (EMA, RSI, Sharpe, etc.)
// 2. Securities data sync (historical prices + scores)
// 3. Tag updates
type IdleProcessor struct {
	cache            *Cache
	technicalCalc    *TechnicalCalculator
	queue            QueueSizer
	securityProvider SecurityProvider
	syncProcessor    SyncProcessor
	tagProcessor     TagProcessor
	eventEmitter     JobEventEmitter
	log              zerolog.Logger

	stats    IdleProcessorStats
	stopChan chan struct{}
	wg       sync.WaitGroup
	running  bool
	mu       sync.Mutex
}

// NewIdleProcessor creates a new idle processor
func NewIdleProcessor(deps IdleProcessorDeps) *IdleProcessor {
	return &IdleProcessor{
		cache:            deps.Cache,
		technicalCalc:    NewTechnicalCalculator(deps.Cache, deps.PriceProvider, deps.Log),
		queue:            deps.Queue,
		securityProvider: deps.SecurityProvider,
		syncProcessor:    deps.SyncProcessor,
		tagProcessor:     deps.TagProcessor,
		eventEmitter:     deps.EventEmitter,
		log:              deps.Log.With().Str("component", "idle_processor").Logger(),
		stopChan:         make(chan struct{}),
	}
}

// SetEventEmitter sets the event emitter for Activity UI visibility.
// This allows deferred wiring when the EventManager isn't available at IdleProcessor creation time.
func (ip *IdleProcessor) SetEventEmitter(emitter JobEventEmitter) {
	ip.mu.Lock()
	defer ip.mu.Unlock()
	ip.eventEmitter = emitter
	ip.log.Info().Msg("Event emitter wired to idle processor")
}

// SetTagProcessor sets the tag processor after initialization.
// This allows deferred wiring when the TagUpdateJob isn't available at IdleProcessor creation time.
func (ip *IdleProcessor) SetTagProcessor(tp TagProcessor) {
	ip.mu.Lock()
	defer ip.mu.Unlock()
	ip.tagProcessor = tp
	ip.log.Info().Msg("Tag processor wired to idle processor")
}

// Start begins the idle processing loop
func (ip *IdleProcessor) Start() {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	if ip.running {
		return
	}

	ip.running = true
	ip.stopChan = make(chan struct{})
	ip.wg.Add(1)

	go func() {
		defer ip.wg.Done()
		ticker := time.NewTicker(CheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ip.stopChan:
				return
			case <-ticker.C:
				ip.ProcessOne()
			}
		}
	}()

	ip.log.Info().Msg("Idle processor started")
}

// Stop halts the idle processor
func (ip *IdleProcessor) Stop() {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	if !ip.running {
		return
	}

	close(ip.stopChan)
	ip.wg.Wait()
	ip.running = false

	ip.log.Info().Msg("Idle processor stopped")
}

// ProcessOne processes one unit of work if system is idle.
// Returns true if work was done, false if system busy or nothing to do.
func (ip *IdleProcessor) ProcessOne() bool {
	if ip.queue.Size() >= IdleQueueThreshold {
		ip.log.Debug().Int("queue_size", ip.queue.Size()).Msg("System busy, skipping idle work")
		return false
	}

	securities, err := ip.securityProvider.GetActiveSecurities()
	if err != nil {
		ip.log.Error().Err(err).Msg("Failed to get active securities")
		return false
	}

	// Try each work type until we find something to do.
	// This ensures all work types get processed eventually.

	// 1. Technical metrics (highest priority - affects scoring and optimization)
	for _, sec := range securities {
		if ip.needsTechnical(sec.ISIN) {
			jobID := fmt.Sprintf("%s-%s-%d", IdleJobTypeTechnical, sec.ISIN, time.Now().UnixNano())
			desc := fmt.Sprintf("Calculating technical indicators for %s", sec.Symbol)

			ip.emitJobStarted(string(IdleJobTypeTechnical), jobID, desc)

			if err := ip.technicalCalc.CalculateForISIN(sec.ISIN); err != nil {
				ip.log.Error().Err(err).Str("isin", sec.ISIN).Msg("Technical calculation failed")
				ip.emitJobFailed(string(IdleJobTypeTechnical), jobID, desc, err)
			} else {
				ip.mu.Lock()
				ip.stats.TechnicalProcessed++
				ip.mu.Unlock()
				ip.log.Debug().Str("isin", sec.ISIN).Msg("Processed technical metrics")
				ip.emitJobCompleted(string(IdleJobTypeTechnical), jobID, desc)
			}
			return true
		}
	}

	// 2. Securities sync (historical prices + scores)
	if ip.syncProcessor != nil {
		for _, sec := range securities {
			if ip.syncProcessor.NeedsSync(sec) {
				jobID := fmt.Sprintf("%s-%s-%d", IdleJobTypeSync, sec.Symbol, time.Now().UnixNano())
				desc := fmt.Sprintf("Syncing security data for %s", sec.Symbol)

				ip.emitJobStarted(string(IdleJobTypeSync), jobID, desc)

				if err := ip.syncProcessor.ProcessSync(sec); err != nil {
					ip.log.Error().Err(err).Str("symbol", sec.Symbol).Msg("Sync failed")
					ip.emitJobFailed(string(IdleJobTypeSync), jobID, desc, err)
				} else {
					ip.mu.Lock()
					ip.stats.SyncProcessed++
					ip.mu.Unlock()
					ip.log.Debug().Str("symbol", sec.Symbol).Msg("Processed sync")
					ip.emitJobCompleted(string(IdleJobTypeSync), jobID, desc)
				}
				return true
			}
		}
	}

	// 3. Tag updates (lowest priority)
	if ip.tagProcessor != nil {
		for _, sec := range securities {
			if ip.tagProcessor.NeedsTagUpdate(sec.Symbol) {
				jobID := fmt.Sprintf("%s-%s-%d", IdleJobTypeTags, sec.Symbol, time.Now().UnixNano())
				desc := fmt.Sprintf("Updating security tags for %s", sec.Symbol)

				ip.emitJobStarted(string(IdleJobTypeTags), jobID, desc)

				if err := ip.tagProcessor.ProcessTagUpdate(sec.Symbol); err != nil {
					ip.log.Error().Err(err).Str("symbol", sec.Symbol).Msg("Tag update failed")
					ip.emitJobFailed(string(IdleJobTypeTags), jobID, desc, err)
				} else {
					ip.mu.Lock()
					ip.stats.TagsProcessed++
					ip.mu.Unlock()
					ip.log.Debug().Str("symbol", sec.Symbol).Msg("Processed tags")
					ip.emitJobCompleted(string(IdleJobTypeTags), jobID, desc)
				}
				return true
			}
		}
	}

	ip.log.Debug().Msg("All securities up to date")
	return false
}

// emitJobStarted emits a job started event if an event emitter is configured
func (ip *IdleProcessor) emitJobStarted(jobType, jobID, description string) {
	if ip.eventEmitter != nil {
		ip.eventEmitter.EmitJobStarted(jobType, jobID, description)
	}
}

// emitJobCompleted emits a job completed event if an event emitter is configured
func (ip *IdleProcessor) emitJobCompleted(jobType, jobID, description string) {
	if ip.eventEmitter != nil {
		ip.eventEmitter.EmitJobCompleted(jobType, jobID, description)
	}
}

// emitJobFailed emits a job failed event if an event emitter is configured
func (ip *IdleProcessor) emitJobFailed(jobType, jobID, description string, err error) {
	if ip.eventEmitter != nil {
		ip.eventEmitter.EmitJobFailed(jobType, jobID, description, err)
	}
}

// needsTechnical checks if technical metrics need calculation.
// Uses EMA-200 as the indicator metric since it requires the most data.
func (ip *IdleProcessor) needsTechnical(isin string) bool {
	_, ok := ip.cache.GetTechnical(isin, "ema", 200)
	return !ok
}

// GetStats returns processing statistics
func (ip *IdleProcessor) GetStats() IdleProcessorStats {
	ip.mu.Lock()
	defer ip.mu.Unlock()
	return ip.stats
}
