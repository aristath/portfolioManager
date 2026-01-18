/**
 * Package di provides adapter type definitions for work processor.
 *
 * This file contains all adapter types that bridge interface mismatches between
 * the work processor and container services.
 */
package di

import (
	"fmt"
	"time"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/aristath/sentinel/internal/events"
	"github.com/aristath/sentinel/internal/work"
)

// marketHoursAdapter adapts MarketHoursService to work.MarketChecker interface
type marketHoursAdapter struct {
	container *Container
}

func (a *marketHoursAdapter) IsAnyMarketOpen() bool {
	if a.container.MarketHoursService == nil {
		return false
	}
	openMarkets := a.container.MarketHoursService.GetOpenMarkets(time.Now())
	return len(openMarkets) > 0
}

func (a *marketHoursAdapter) IsSecurityMarketOpen(isin string) bool {
	if a.container.MarketHoursService == nil {
		return false
	}
	// Get security to determine market
	sec, err := a.container.SecurityRepo.GetByISIN(isin)
	if err != nil || sec == nil {
		return false
	}
	return a.container.MarketHoursService.IsMarketOpen(sec.FullExchangeName, time.Now())
}

func (a *marketHoursAdapter) AreAllMarketsClosed() bool {
	if a.container.MarketHoursService == nil {
		return true
	}
	openMarkets := a.container.MarketHoursService.GetOpenMarkets(time.Now())
	return len(openMarkets) == 0
}

// eventEmitterAdapter adapts events.Manager to work.EventEmitter interface
type eventEmitterAdapter struct {
	manager *events.Manager
}

func (a *eventEmitterAdapter) Emit(event string, data any) {
	if a.manager == nil {
		return
	}

	// Map work event names to proper EventType constants
	var eventType events.EventType
	switch event {
	case work.EventJobStarted:
		eventType = events.JobStarted
	case work.EventJobProgress:
		eventType = events.JobProgress
	case work.EventJobCompleted:
		eventType = events.JobCompleted
	case work.EventJobFailed:
		eventType = events.JobFailed
	default:
		// Fallback for unknown events
		eventType = events.EventType(event)
	}

	// Convert work event structs to JobStatusData format expected by frontend
	var jobData *events.JobStatusData

	switch event {
	case work.EventJobStarted:
		if startedEvent, ok := data.(work.WorkStartedEvent); ok {
			jobData = &events.JobStatusData{
				JobID:       startedEvent.WorkID,
				JobType:     startedEvent.WorkType,
				Status:      "started",
				Description: getJobDescription(startedEvent.WorkType),
				Timestamp:   time.Now(),
			}
		}
	case work.EventJobProgress:
		if progressEvent, ok := data.(work.ProgressEvent); ok {
			jobData = &events.JobStatusData{
				JobID:   progressEvent.WorkID,
				JobType: progressEvent.WorkType,
				Status:  "progress",
				Progress: &events.JobProgressInfo{
					Current: progressEvent.Current,
					Total:   progressEvent.Total,
					Message: progressEvent.Message,
					Phase:   progressEvent.Phase,
					Details: progressEvent.Details,
				},
				Description: getJobDescription(progressEvent.WorkType),
				Timestamp:   time.Now(),
			}
		}
	case work.EventJobCompleted:
		if completedEvent, ok := data.(work.WorkCompletedEvent); ok {
			jobData = &events.JobStatusData{
				JobID:       completedEvent.WorkID,
				JobType:     completedEvent.WorkType,
				Status:      "completed",
				Description: getJobDescription(completedEvent.WorkType),
				Duration:    float64(completedEvent.Duration.Milliseconds()),
				Timestamp:   time.Now(),
			}
		}
	case work.EventJobFailed:
		if failedEvent, ok := data.(work.WorkFailedEvent); ok {
			jobData = &events.JobStatusData{
				JobID:       failedEvent.WorkID,
				JobType:     failedEvent.WorkType,
				Status:      "failed",
				Description: getJobDescription(failedEvent.WorkType),
				Error:       failedEvent.Error,
				Duration:    float64(failedEvent.Duration.Milliseconds()),
				Timestamp:   time.Now(),
			}
		}
	}

	// Emit with properly formatted JobStatusData
	if jobData != nil {
		a.manager.EmitTyped(eventType, "", jobData)
	}
}

// workBrokerPriceAdapter adapts domain.BrokerClient to scheduler.BrokerClientForPrices interface
type workBrokerPriceAdapter struct {
	client domain.BrokerClient
}

func (a *workBrokerPriceAdapter) GetBatchQuotes(symbolMap map[string]*string) (map[string]*float64, error) {
	if a.client == nil {
		return nil, fmt.Errorf("broker client not available")
	}

	// Extract symbols from map
	symbols := make([]string, 0, len(symbolMap))
	for symbol := range symbolMap {
		symbols = append(symbols, symbol)
	}

	// Get quotes from broker
	quotes, err := a.client.GetQuotes(symbols)
	if err != nil {
		return nil, fmt.Errorf("failed to get broker quotes: %w", err)
	}

	// Convert to price map
	prices := make(map[string]*float64)
	for symbol, quote := range quotes {
		if quote != nil && quote.Price > 0 {
			price := quote.Price
			prices[symbol] = &price
		}
	}

	return prices, nil
}
