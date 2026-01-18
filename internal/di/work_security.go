/**
 * Package di provides dependency injection for security work registration.
 *
 * Security work types handle historical price sync, technical calculations,
 * formula discovery, tag updates, and metadata synchronization.
 */
package di

import (
	"github.com/aristath/sentinel/internal/modules/universe"
	"github.com/aristath/sentinel/internal/work"
	"github.com/rs/zerolog"
)

// Security work type adapters
type securityHistorySyncAdapter struct {
	container *Container
}

func (a *securityHistorySyncAdapter) SyncSecurityHistory(isin string) (string, error) {
	// Get security to get the symbol
	sec, err := a.container.SecurityRepo.GetByISIN(isin)
	if err != nil || sec == nil {
		return "", err
	}
	err = a.container.HistoricalSyncService.SyncHistoricalPrices(sec.Symbol)
	return sec.Symbol, err
}

func (a *securityHistorySyncAdapter) GetStaleSecurities() []string {
	// Get ISINs that need historical data sync
	// For now, return all active securities - the work processor handles staleness via completion tracker
	securities, err := a.container.SecurityRepo.GetAllActive()
	if err != nil {
		return nil
	}
	var isins []string
	for _, sec := range securities {
		isins = append(isins, sec.ISIN)
	}
	return isins
}

type securityTechnicalAdapter struct {
	container *Container
}

func (a *securityTechnicalAdapter) CalculateTechnicals(isin string) (string, error) {
	// Technical calculations are done during historical sync
	// Get security to return the symbol for progress reporting
	sec, err := a.container.SecurityRepo.GetByISIN(isin)
	if err != nil || sec == nil {
		return "", err
	}
	return sec.Symbol, nil
}

func (a *securityTechnicalAdapter) GetSecuritiesNeedingTechnicals() []string {
	// Already handled by historical sync
	return nil
}

type securityFormulaAdapter struct {
	container *Container
}

func (a *securityFormulaAdapter) RunDiscovery(isin string) (string, error) {
	// Formula discovery - placeholder for now
	// Get security to return the symbol for progress reporting
	sec, err := a.container.SecurityRepo.GetByISIN(isin)
	if err != nil || sec == nil {
		return "", err
	}
	return sec.Symbol, nil
}

func (a *securityFormulaAdapter) GetSecuritiesNeedingDiscovery() []string {
	return nil
}

type securityTagAdapter struct {
	container *Container
}

func (a *securityTagAdapter) UpdateTags(isin string) (string, error) {
	// Tag assignment requires full AssignTagsInput - for now this is a no-op
	// Tags are assigned as part of the scoring workflow
	// Get security to return the symbol for progress reporting
	sec, err := a.container.SecurityRepo.GetByISIN(isin)
	if err != nil || sec == nil {
		return "", err
	}
	return sec.Symbol, nil
}

func (a *securityTagAdapter) GetSecuritiesNeedingTagUpdate() []string {
	// Tag updates happen as part of scoring workflow
	// Return empty list for now
	return nil
}

type metadataSyncAdapter struct {
	service *universe.MetadataSyncService
}

func (a *metadataSyncAdapter) SyncMetadata(isin string) (string, error) {
	return a.service.SyncMetadata(isin)
}

func (a *metadataSyncAdapter) GetAllActiveISINs() []string {
	return a.service.GetAllActiveISINs()
}

func (a *metadataSyncAdapter) SyncMetadataBatch(isins []string) (int, error) {
	return a.service.SyncMetadataBatch(isins)
}

func registerSecurityWork(registry *work.Registry, container *Container, log zerolog.Logger) {
	// Use container's MetadataSyncService (used by work system batch sync)
	deps := &work.SecurityDeps{
		HistorySyncService:  &securityHistorySyncAdapter{container: container},
		TechnicalService:    &securityTechnicalAdapter{container: container},
		FormulaService:      &securityFormulaAdapter{container: container},
		TagService:          &securityTagAdapter{container: container},
		MetadataSyncService: &metadataSyncAdapter{service: container.MetadataSyncService},
	}

	work.RegisterSecurityWorkTypes(registry, deps)
	log.Debug().Msg("Security work types registered")
}
