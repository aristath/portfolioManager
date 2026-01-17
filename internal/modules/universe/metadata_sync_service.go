package universe

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/rs/zerolog"
)

// MetadataSyncService syncs Tradernet metadata for securities.
// Stores raw Tradernet API response in the data column for field mapping on read.
type MetadataSyncService struct {
	securityRepo *SecurityRepository
	brokerClient domain.BrokerClient
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
		brokerClient: brokerClient,
		log:          log.With().Str("service", "metadata_sync").Logger(),
	}
}

// SyncMetadata syncs Tradernet metadata for a security identified by ISIN.
// Stores raw Tradernet API response (securities[0]) in the data column.
// Field mapping is applied on read via SecurityFromJSON.
// Returns the symbol for progress reporting.
func (s *MetadataSyncService) SyncMetadata(isin string) (string, error) {
	// Get security by ISIN
	security, err := s.securityRepo.GetByISIN(isin)
	if err != nil {
		return "", fmt.Errorf("failed to get security %s: %w", isin, err)
	}
	if security == nil {
		s.log.Debug().Str("isin", isin).Msg("Security not found, skipping")
		return "", nil
	}

	// Call Tradernet API to get raw response
	rawResponse, err := s.brokerClient.GetSecurityMetadataRaw(security.Symbol)
	if err != nil {
		return security.Symbol, fmt.Errorf("failed to fetch metadata for %s: %w", isin, err)
	}
	if rawResponse == nil {
		s.log.Debug().Str("isin", isin).Str("symbol", security.Symbol).Msg("No metadata returned from broker")
		return security.Symbol, nil
	}

	// Extract securities[0] from raw response
	responseMap, ok := rawResponse.(map[string]interface{})
	if !ok {
		return security.Symbol, fmt.Errorf("unexpected response format for %s: expected map", isin)
	}

	securities, ok := responseMap["securities"].([]interface{})
	if !ok || len(securities) == 0 {
		s.log.Debug().Str("isin", isin).Str("symbol", security.Symbol).Msg("No securities in response")
		return security.Symbol, nil
	}

	// Get first security object
	securityData := securities[0]

	// Marshal to JSON
	jsonBytes, err := json.Marshal(securityData)
	if err != nil {
		return security.Symbol, fmt.Errorf("failed to marshal security data for %s: %w", isin, err)
	}

	// Store raw data with last_synced timestamp
	updates := map[string]any{
		"data":        string(jsonBytes),
		"last_synced": time.Now().Unix(),
	}

	if err := s.securityRepo.Update(isin, updates); err != nil {
		return security.Symbol, fmt.Errorf("failed to update security %s: %w", isin, err)
	}

	s.log.Debug().
		Str("isin", isin).
		Str("symbol", security.Symbol).
		Int("data_size", len(jsonBytes)).
		Msg("Synced raw metadata for security")

	return security.Symbol, nil
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

// SyncMetadataBatch syncs metadata for multiple securities in a single batch API call.
// This method eliminates 429 rate limit errors by batching all securities into one request.
//
// Parameters:
//   - isins: Array of security ISINs to sync metadata for
//
// Returns:
//   - int: Number of securities successfully synced
//   - error: Error if batch operation fails completely
//
// Workflow:
//  1. Build ISIN→Symbol mapping from repository
//  2. Make single batch API call for all symbols
//  3. Parse batch response and build symbol→data lookup map
//  4. Update each security in database with matched data
//  5. Return success count
func (s *MetadataSyncService) SyncMetadataBatch(isins []string) (int, error) {
	if len(isins) == 0 {
		return 0, nil
	}

	s.log.Debug().
		Int("count", len(isins)).
		Msg("Starting batch metadata sync")

	// Step 1: Build ISIN→Symbol map
	isinToSymbol := make(map[string]string)
	symbols := make([]string, 0, len(isins))

	for _, isin := range isins {
		security, err := s.securityRepo.GetByISIN(isin)
		if err != nil {
			s.log.Warn().Str("isin", isin).Err(err).Msg("Failed to get security for batch sync")
			continue
		}
		if security == nil {
			s.log.Debug().Str("isin", isin).Msg("Security not found, skipping")
			continue
		}

		isinToSymbol[isin] = security.Symbol
		symbols = append(symbols, security.Symbol)
	}

	if len(symbols) == 0 {
		return 0, fmt.Errorf("no valid symbols found for batch sync")
	}

	// Step 2: Single batch API call
	rawResponse, err := s.brokerClient.GetSecurityMetadataBatch(symbols)
	if err != nil {
		return 0, fmt.Errorf("batch API call failed: %w", err)
	}

	// Step 3: Parse response
	responseMap, ok := rawResponse.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("unexpected batch response format")
	}

	securities, ok := responseMap["securities"].([]interface{})
	if !ok {
		return 0, fmt.Errorf("missing securities array in batch response")
	}

	if len(securities) == 0 {
		s.log.Warn().Msg("Batch API returned no securities")
		return 0, fmt.Errorf("batch API returned empty results")
	}

	// Step 4: Build symbol→data lookup map
	symbolToData := make(map[string]interface{})
	for _, sec := range securities {
		secMap, ok := sec.(map[string]interface{})
		if !ok {
			continue
		}

		ticker, ok := secMap["ticker"].(string)
		if !ok {
			continue
		}

		symbolToData[ticker] = sec
	}

	// Step 5: Update each security in database
	successCount := 0
	now := time.Now().Unix()

	for isin, symbol := range isinToSymbol {
		securityData, found := symbolToData[symbol]
		if !found {
			s.log.Warn().
				Str("isin", isin).
				Str("symbol", symbol).
				Msg("Symbol not found in batch response")
			continue
		}

		// Marshal to JSON
		jsonBytes, err := json.Marshal(securityData)
		if err != nil {
			s.log.Error().
				Str("isin", isin).
				Str("symbol", symbol).
				Err(err).
				Msg("Failed to marshal security data")
			continue
		}

		// Update database
		updates := map[string]any{
			"data":        string(jsonBytes),
			"last_synced": now,
		}

		if err := s.securityRepo.Update(isin, updates); err != nil {
			s.log.Error().
				Str("isin", isin).
				Str("symbol", symbol).
				Err(err).
				Msg("Failed to update security in batch")
			continue
		}

		successCount++
	}

	s.log.Info().
		Int("total", len(isins)).
		Int("success", successCount).
		Int("failed", len(isins)-successCount).
		Int("data_size", len(securities)).
		Msg("Batch metadata sync completed")

	if successCount == 0 {
		return 0, fmt.Errorf("batch sync failed for all %d securities", len(isins))
	}

	return successCount, nil
}
