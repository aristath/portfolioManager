package universe

import (
	"fmt"
	"strings"

	"github.com/aristath/sentinel/internal/domain"
	"github.com/aristath/sentinel/internal/modules/portfolio"
	"github.com/rs/zerolog"
)

// SecurityDeletionService handles hard deletion of securities and all related data
// across multiple databases. It validates that no open positions or pending orders
// exist before deletion.
type SecurityDeletionService struct {
	securityRepo        SecurityRepositoryInterface
	positionRepo        portfolio.PositionRepositoryInterface
	scoreRepo           ScoreRepositoryInterface
	historyDB           HistoryDBInterface
	dismissedFilterRepo DismissedFilterClearer
	brokerClient        domain.BrokerClient
	log                 zerolog.Logger
}

// NewSecurityDeletionService creates a new security deletion service
func NewSecurityDeletionService(
	securityRepo SecurityRepositoryInterface,
	positionRepo portfolio.PositionRepositoryInterface,
	scoreRepo ScoreRepositoryInterface,
	historyDB HistoryDBInterface,
	dismissedFilterRepo DismissedFilterClearer,
	brokerClient domain.BrokerClient,
	log zerolog.Logger,
) *SecurityDeletionService {
	return &SecurityDeletionService{
		securityRepo:        securityRepo,
		positionRepo:        positionRepo,
		scoreRepo:           scoreRepo,
		historyDB:           historyDB,
		dismissedFilterRepo: dismissedFilterRepo,
		brokerClient:        brokerClient,
		log:                 log.With().Str("service", "security_deletion").Logger(),
	}
}

// HardDelete permanently removes a security and all related data across databases.
// Returns error if:
// - Security does not exist
// - Security has open positions (quantity > 0)
// - Security has pending orders at the broker
//
// Deletion order (universe first for data integrity):
// 1. universe.db: securities, security_tags, broker_symbols, client_symbols
// 2. portfolio.db: positions, scores (kelly_sizes deleted via CASCADE)
// 3. history.db: daily_prices, monthly_prices
// 4. config.db: dismissed_filters
//
// ledger.db is UNTOUCHED (audit trail preserved)
func (s *SecurityDeletionService) HardDelete(isin string) error {
	isin = strings.ToUpper(strings.TrimSpace(isin))

	// 1. Verify security exists and get details for logging
	security, err := s.securityRepo.GetByISIN(isin)
	if err != nil {
		return fmt.Errorf("failed to lookup security: %w", err)
	}
	if security == nil {
		return fmt.Errorf("security not found: %s", isin)
	}

	symbol := security.Symbol

	// 2. Check for open positions
	position, err := s.positionRepo.GetByISIN(isin)
	if err != nil {
		return fmt.Errorf("failed to check positions: %w", err)
	}
	if position != nil && position.Quantity > 0 {
		return fmt.Errorf("cannot delete security with open position: %.4f shares held", position.Quantity)
	}

	// 3. Check for pending orders at the broker
	pendingOrders, err := s.brokerClient.GetPendingOrders()
	if err != nil {
		return fmt.Errorf("failed to check pending orders: %w", err)
	}
	for _, order := range pendingOrders {
		// Check symbol matches (pending orders only have symbol, not ISIN)
		if strings.EqualFold(order.Symbol, symbol) {
			return fmt.Errorf("cannot delete security with pending order: %s %s %.4f shares",
				order.Side, order.Symbol, order.Quantity)
		}
	}

	s.log.Info().
		Str("isin", isin).
		Str("symbol", symbol).
		Msg("Hard deleting security and all related data")

	// 4. Delete from universe.db FIRST (authoritative source)
	// This ensures if this fails, we haven't orphaned any cleanup operations
	if err := s.securityRepo.HardDelete(isin); err != nil {
		return fmt.Errorf("failed to delete security from universe: %w", err)
	}

	// 5. Clean up related data in other databases
	// These are best-effort - the security is already gone from universe
	// Errors are logged but don't fail the operation

	// Delete position (if exists with zero quantity)
	if position != nil {
		if err := s.positionRepo.Delete(isin); err != nil {
			s.log.Error().Err(err).Str("isin", isin).Msg("Failed to delete position")
		}
	}

	// Delete scores (kelly_sizes deleted via CASCADE)
	if err := s.scoreRepo.Delete(isin); err != nil {
		s.log.Error().Err(err).Str("isin", isin).Msg("Failed to delete scores")
	}

	// Delete price history
	if err := s.historyDB.DeletePricesForSecurity(isin); err != nil {
		s.log.Error().Err(err).Str("isin", isin).Msg("Failed to delete price history")
	}

	// Delete dismissed filters
	if _, err := s.dismissedFilterRepo.ClearForSecurity(isin); err != nil {
		s.log.Error().Err(err).Str("isin", isin).Msg("Failed to delete dismissed filters")
	}

	s.log.Info().
		Str("isin", isin).
		Str("symbol", symbol).
		Msg("Security hard deleted successfully")

	return nil
}
