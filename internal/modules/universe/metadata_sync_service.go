package universe

import (
	"fmt"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/rs/zerolog"
)

// MetadataSyncService syncs Tradernet metadata for securities.
// It uses the MetadataEnricher to fetch and apply metadata from the broker API.
type MetadataSyncService struct {
	securityRepo *SecurityRepository
	enricher     *MetadataEnricher
	log          zerolog.Logger
}

// NewMetadataSyncService creates a new metadata sync service.
func NewMetadataSyncService(
	securityRepo *SecurityRepository,
	brokerClient domain.BrokerClient,
	log zerolog.Logger,
) *MetadataSyncService {
	return &MetadataSyncService{
		securityRepo: securityRepo,
		enricher:     NewMetadataEnricher(brokerClient, log),
		log:          log.With().Str("service", "metadata_sync").Logger(),
	}
}

// SyncMetadata syncs Tradernet metadata for a security identified by ISIN.
// Uses the MetadataEnricher which:
// - Always overwrites: geography (CntryOfRisk), industry (raw sector code), min_lot (quotes.x_lot)
// - Only fills if empty: name, currency, fullExchangeName, market_code
func (s *MetadataSyncService) SyncMetadata(isin string) error {
	// Get security by ISIN
	security, err := s.securityRepo.GetByISIN(isin)
	if err != nil {
		return fmt.Errorf("failed to get security %s: %w", isin, err)
	}
	if security == nil {
		s.log.Debug().Str("isin", isin).Msg("Security not found, skipping")
		return nil
	}

	// Enrich metadata from broker
	if err := s.enricher.Enrich(security); err != nil {
		return fmt.Errorf("failed to enrich metadata for %s: %w", isin, err)
	}

	// Build update map with enriched fields
	updates := map[string]any{
		"geography":        security.Geography,
		"industry":         security.Industry,
		"min_lot":          security.MinLot,
		"name":             security.Name,
		"currency":         security.Currency,
		"fullExchangeName": security.FullExchangeName,
		"market_code":      security.MarketCode,
	}

	// Update security in database
	if err := s.securityRepo.Update(isin, updates); err != nil {
		return fmt.Errorf("failed to update security %s: %w", isin, err)
	}

	s.log.Debug().
		Str("isin", isin).
		Str("symbol", security.Symbol).
		Str("geography", security.Geography).
		Str("industry", security.Industry).
		Int("min_lot", security.MinLot).
		Msg("Synced metadata for security")

	return nil
}

// GetAllActiveISINs returns all active security ISINs for metadata sync.
func (s *MetadataSyncService) GetAllActiveISINs() []string {
	securities, err := s.securityRepo.GetAllActive()
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get active securities")
		return nil
	}

	isins := make([]string, 0, len(securities))
	for _, sec := range securities {
		if sec.ISIN != "" {
			isins = append(isins, sec.ISIN)
		}
	}

	return isins
}
