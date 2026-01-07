package portfolio

import (
	"database/sql"

	"github.com/aristath/sentinel/internal/market_regime"
	"github.com/rs/zerolog"
)

// NOTE: These aliases preserve the historical `portfolio.*` API while the codebase is migrated
// to `internal/market_regime`. They will be removed once all imports are updated.

// MarketRegimeScore represents market condition as continuous score.
type MarketRegimeScore = market_regime.MarketRegimeScore

const (
	ExtremeBearScore = market_regime.ExtremeBearScore
	BearScore        = market_regime.BearScore
	NeutralScore     = market_regime.NeutralScore
	BullScore        = market_regime.BullScore
	ExtremeBullScore = market_regime.ExtremeBullScore
)

const TanhCompressionFactor = market_regime.TanhCompressionFactor

type MarketRegimeDetector = market_regime.MarketRegimeDetector

func NewMarketRegimeDetector(log zerolog.Logger) *MarketRegimeDetector {
	return market_regime.NewMarketRegimeDetector(log)
}

type RegimePersistence = market_regime.RegimePersistence
type RegimeHistoryEntry = market_regime.RegimeHistoryEntry

func NewRegimePersistence(db *sql.DB, log zerolog.Logger) *RegimePersistence {
	return market_regime.NewRegimePersistence(db, log)
}

type MarketIndex = market_regime.MarketIndex
type MarketIndexService = market_regime.MarketIndexService

func NewMarketIndexService(universeDB *sql.DB, historyDB *sql.DB, tradernet interface{}, log zerolog.Logger) *MarketIndexService {
	return market_regime.NewMarketIndexService(universeDB, historyDB, tradernet, log)
}
