package universe

import (
	"github.com/rs/zerolog"
)

// UniverseService handles universe (securities) business logic
type UniverseService struct {
	log zerolog.Logger
}

// NewUniverseService creates a new universe service
func NewUniverseService(log zerolog.Logger) *UniverseService {
	return &UniverseService{
		log: log.With().Str("service", "universe").Logger(),
	}
}

// SyncPrices synchronizes current prices for all securities
// TODO: Implement in future phase
func (s *UniverseService) SyncPrices() error {
	s.log.Debug().Msg("Price sync not yet implemented")
	return nil
}
