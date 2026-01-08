package market_regime

// MarketRegimeScore represents market condition as continuous score
// Range: -1.0 (extreme bear) to +1.0 (extreme bull)
// 0.0 = neutral/sideways
// Allows gradual transitions: 0.3 = "bull-ish", -0.2 = "bear-ish"
type MarketRegimeScore float64

// Helper constants for reference (but use continuous score)
const (
	ExtremeBearScore MarketRegimeScore = -1.0
	BearScore        MarketRegimeScore = -0.5
	NeutralScore     MarketRegimeScore = 0.0
	BullScore        MarketRegimeScore = 0.5
	ExtremeBullScore MarketRegimeScore = 1.0
)
