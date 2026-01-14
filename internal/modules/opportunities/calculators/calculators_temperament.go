package calculators

import (
	"github.com/aristath/sentinel/internal/modules/settings"
)

// CalculatorConfig holds all configuration parameters for opportunity calculators.
// These are organized by calculator type and include priority boost multipliers.
type CalculatorConfig struct {
	// ProfitTaking holds profit-taking calculator parameters
	ProfitTaking ProfitTakingConfig

	// AveragingDown holds averaging-down calculator parameters
	AveragingDown AveragingDownConfig

	// OpportunityBuys holds opportunity buy calculator parameters
	OpportunityBuys OpportunityBuysConfig

	// Rebalance holds rebalance calculator parameters
	Rebalance RebalanceConfig

	// TransactionEfficiency holds transaction cost parameters
	TransactionEfficiency TransactionEfficiencyConfig

	// Boosts holds all priority boost multipliers
	Boosts CalculatorBoosts
}

// ProfitTakingConfig holds profit-taking calculator configuration
type ProfitTakingConfig struct {
	MinGainThreshold  float64 // Minimum gain to consider profit taking (default: 0.15 = 15%)
	WindfallThreshold float64 // Threshold for windfall classification (default: 0.30 = 30%)
	MinHoldDays       int     // Minimum holding period before profit taking (default: 90)
	SellPercentage    float64 // Default percentage of position to sell (default: 1.0 = 100%)
	MaxSellPercentage float64 // Maximum percentage cap for risk management (default: 1.0 = 100%)
	MaxPositions      int     // Maximum positions to include (default: 0 = unlimited)
}

// AveragingDownConfig holds averaging-down calculator configuration
type AveragingDownConfig struct {
	MaxLossThreshold     float64 // Maximum loss threshold (default: -0.20 = -20%)
	MinLossThreshold     float64 // Minimum loss threshold (default: -0.05 = -5%)
	MaxValuePerPosition  float64 // Maximum value per position (default: 500.0)
	AveragingDownPercent float64 // Percentage of position to add (default: 0.10 = 10%)
	MaxPositions         int     // Maximum positions to average down (default: 3)
}

// OpportunityBuysConfig holds opportunity buy calculator configuration
type OpportunityBuysConfig struct {
	MinScore            float64 // Minimum score threshold (default: 0.65)
	MaxValuePerPosition float64 // Maximum value per position (default: 500.0)
	MaxPositions        int     // Maximum positions to buy (default: 5)
	ExcludeExisting     bool    // Exclude existing positions (default: false)
}

// RebalanceConfig holds rebalance calculator configuration
type RebalanceConfig struct {
	MinOverweightThreshold float64 // Minimum overweight to trigger sell (default: 0.05 = 5%)
	MaxSellPercentage      float64 // Maximum percentage to sell (default: 0.50 = 50%)
	MaxPositions           int     // Maximum positions to rebalance (default: 0 = unlimited)
}

// TransactionEfficiencyConfig holds transaction cost parameters
type TransactionEfficiencyConfig struct {
	MaxCostRatio float64 // Maximum acceptable cost ratio (default: 0.01 = 1%)
}

// CalculatorBoosts holds all priority boost multipliers across calculators
type CalculatorBoosts struct {
	// Windfall/Sell boosts (Category 10)
	WindfallBoost       float64 // Boost for windfall gains (default: 1.5)
	BubbleRiskBoost     float64 // Boost for bubble risk detection (default: 1.4)
	NeedsRebalanceBoost float64 // Boost for rebalance alignment (default: 1.3)
	OverweightBoost     float64 // Boost for overweight positions (default: 1.2)
	OvervaluedBoost     float64 // Boost for overvalued positions (default: 1.15)
	Near52wHighBoost    float64 // Boost for near 52-week high (default: 1.1)

	// Quality-Value boosts (Category 11)
	QualityValueBoost      float64 // Boost for quality value (default: 1.5)
	RecoveryCandidateBoost float64 // Boost for recovery candidates (default: 1.3)
	HighQualityBoost       float64 // Boost for high quality (default: 1.15)
	ValueOpportunityBoost  float64 // Boost for value opportunity (default: 1.1)
	DeepValueBoost         float64 // Boost for deep value (default: 1.2)

	// Risk Profile boosts (Category 12)
	LowRiskBoost    float64 // Boost for low risk (default: 1.15)
	MediumRiskBoost float64 // Boost for medium risk (default: 1.05)
	HighRiskPenalty float64 // Penalty for high risk (default: 0.90)

	// Regime-Aware boosts (Category 13)
	BullGrowthBoost       float64 // Boost for growth in bull market (default: 1.15)
	BearValueBoost        float64 // Boost for value in bear market (default: 1.15)
	SidewaysDividendBoost float64 // Boost for dividends in sideways market (default: 1.12)
	NeutralGrowthBoost    float64 // Boost for growth in neutral market (default: 1.08)
	NeutralValueBoost     float64 // Boost for value in neutral market (default: 1.08)
	NeutralDividendBoost  float64 // Boost for dividends in neutral market (default: 1.10)

	// Quality boosts
	HighStabilityBoost       float64 // Boost for high stability (default: 1.12)
	ConsistentGrowerBoost    float64 // Boost for consistent grower (default: 1.10)
	StableBoost              float64 // Boost for stable (default: 1.08)
	DividendTotalReturnBoost float64 // Boost for dividend total return (default: 1.12)

	// Performance-based sell boosts
	UnsustainableGainsBoost float64 // Boost to sell unsustainable gains (default: 1.25)
	StagnantBoost           float64 // Boost to sell stagnant (default: 1.15)
	UnderperformingBoost    float64 // Boost to sell underperforming (default: 1.20)
	MeetsTargetReturnBoost  float64 // Boost for meeting target return (default: 1.10)

	// Return-based boosts
	ExcellentReturnBoost float64 // Boost for excellent total return (default: 1.25)
	HighReturnBoost      float64 // Boost for high total return (default: 1.15)
	QualityHighCAGRBoost float64 // Boost for quality high CAGR (default: 1.2)
	DividendGrowerBoost  float64 // Boost for dividend grower (default: 1.15)
	HighDividendBoost    float64 // Boost for high dividend (default: 1.1)

	// Quantum warning penalties (Category 14)
	QuantumPenaltyMild   float64 // Mild penalty for quantum warnings (default: 0.90)
	QuantumPenaltySevere float64 // Severe penalty for quantum warnings (default: 0.70)
}

// NewCalculatorConfig creates a CalculatorConfig from the settings service.
// This is the recommended way to create config as it respects temperament settings.
func NewCalculatorConfig(settingsService *settings.Service) CalculatorConfig {
	profitTaking := settingsService.GetAdjustedProfitTakingParams()
	avgDown := settingsService.GetAdjustedAveragingDownParams()
	oppBuys := settingsService.GetAdjustedOpportunityBuysParams()
	rebalancing := settingsService.GetAdjustedRebalancingParams()
	transaction := settingsService.GetAdjustedTransactionParams()
	riskMgmt := settingsService.GetAdjustedRiskManagementParams()
	ptBoosts := settingsService.GetAdjustedProfitTakingBoosts()
	avgBoosts := settingsService.GetAdjustedAveragingDownBoosts()
	oppBoosts := settingsService.GetAdjustedOpportunityBuyBoosts()
	regimeBoosts := settingsService.GetAdjustedRegimeBoosts()

	return CalculatorConfig{
		ProfitTaking: ProfitTakingConfig{
			MinGainThreshold:  profitTaking.MinGainThreshold,
			WindfallThreshold: profitTaking.WindfallThreshold,
			MinHoldDays:       riskMgmt.MinHoldDays,
			SellPercentage:    profitTaking.SellPercentage,
			MaxSellPercentage: riskMgmt.MaxSellPercentage,
			MaxPositions:      0, // Always unlimited
		},
		AveragingDown: AveragingDownConfig{
			MaxLossThreshold:     avgDown.MaxLossThreshold,
			MinLossThreshold:     avgDown.MinLossThreshold,
			MaxValuePerPosition:  oppBuys.MaxValuePerPosition, // Use same value from opportunity buys
			AveragingDownPercent: avgDown.Percent,
			MaxPositions:         3, // Hardcoded conservative default
		},
		OpportunityBuys: OpportunityBuysConfig{
			MinScore:            oppBuys.MinScore,
			MaxValuePerPosition: oppBuys.MaxValuePerPosition,
			MaxPositions:        oppBuys.MaxPositions,
			ExcludeExisting:     false, // Always include existing
		},
		Rebalance: RebalanceConfig{
			MinOverweightThreshold: rebalancing.MinOverweightThreshold,
			MaxSellPercentage:      riskMgmt.MaxSellPercentage,
			MaxPositions:           0, // Always unlimited
		},
		TransactionEfficiency: TransactionEfficiencyConfig{
			MaxCostRatio: transaction.MaxCostRatio,
		},
		Boosts: CalculatorBoosts{
			// Windfall/Profit-taking boosts
			WindfallBoost:       ptBoosts.WindfallPriority,
			BubbleRiskBoost:     ptBoosts.BubbleRisk,
			NeedsRebalanceBoost: ptBoosts.NeedsRebalance,
			OverweightBoost:     ptBoosts.Overweight,
			OvervaluedBoost:     ptBoosts.Overvalued,
			Near52wHighBoost:    ptBoosts.Near52wHigh,

			// Quality-Value boosts (from averaging down)
			QualityValueBoost:      avgBoosts.QualityValue,
			RecoveryCandidateBoost: avgBoosts.RecoveryCandidate,
			HighQualityBoost:       avgBoosts.HighQuality,
			ValueOpportunityBoost:  avgBoosts.ValueOpportunity,
			DeepValueBoost:         oppBoosts.DeepValue,

			// Risk Profile boosts
			LowRiskBoost:    regimeBoosts.LowRisk,
			MediumRiskBoost: regimeBoosts.MediumRisk,
			HighRiskPenalty: regimeBoosts.HighRiskPenalty,

			// Regime boosts
			BullGrowthBoost:       regimeBoosts.GrowthBull,
			BearValueBoost:        regimeBoosts.ValueBear,
			SidewaysDividendBoost: regimeBoosts.DividendSideways,
			NeutralGrowthBoost:    1.08, // Hardcoded for now
			NeutralValueBoost:     1.08, // Hardcoded for now
			NeutralDividendBoost:  1.10, // Hardcoded for now

			// Quality boosts
			HighStabilityBoost:       regimeBoosts.HighStability,
			ConsistentGrowerBoost:    1.10,
			StableBoost:              1.08,
			DividendTotalReturnBoost: 1.12,

			// Performance boosts
			UnsustainableGainsBoost: 1.25,
			StagnantBoost:           1.15,
			UnderperformingBoost:    1.20,
			MeetsTargetReturnBoost:  1.10,

			// Return boosts
			ExcellentReturnBoost: oppBoosts.ExcellentReturns,
			HighReturnBoost:      oppBoosts.HighReturns,
			QualityHighCAGRBoost: oppBoosts.QualityHighCAGR,
			DividendGrowerBoost:  oppBoosts.DividendGrower,
			HighDividendBoost:    oppBoosts.HighDividend,

			// Quantum penalties
			QuantumPenaltyMild:   0.90,
			QuantumPenaltySevere: oppBoosts.QuantumWarningPenalty,
		},
	}
}

// NewDefaultCalculatorConfig creates a CalculatorConfig with default values.
// These match the current hardcoded values in the calculator files.
func NewDefaultCalculatorConfig() CalculatorConfig {
	return CalculatorConfig{
		ProfitTaking: ProfitTakingConfig{
			MinGainThreshold:  0.15, // Line 48: 15% minimum gain
			WindfallThreshold: 0.30, // Line 49: 30% for windfall
			MinHoldDays:       90,   // Line 50: 90 days minimum
			SellPercentage:    1.0,  // Line 51: 100% default
			MaxSellPercentage: 1.0,  // Line 52: 100% max
			MaxPositions:      0,    // Line 53: unlimited
		},
		AveragingDown: AveragingDownConfig{
			MaxLossThreshold:     -0.20, // Line 48: -20% max loss
			MinLossThreshold:     -0.05, // Line 49: -5% min loss
			MaxValuePerPosition:  500.0, // Line 50: 500 EUR max
			AveragingDownPercent: 0.10,  // Line 51: 10% of position
			MaxPositions:         3,     // Line 52: max 3 positions
		},
		OpportunityBuys: OpportunityBuysConfig{
			MinScore:            0.65,  // Line 50: 0.65 min score
			MaxValuePerPosition: 500.0, // Line 51: 500 EUR max
			MaxPositions:        5,     // Line 52: max 5 positions
			ExcludeExisting:     false, // Line 53: include existing
		},
		Rebalance: RebalanceConfig{
			MinOverweightThreshold: 0.05, // Line 47: 5% overweight
			MaxSellPercentage:      0.50, // Line 48: 50% max sell
			MaxPositions:           0,    // Line 49: unlimited
		},
		TransactionEfficiency: TransactionEfficiencyConfig{
			MaxCostRatio: 0.01, // Line 55/56: 1% max cost ratio
		},
		Boosts: CalculatorBoosts{
			// Windfall boosts (profit_taking.go calculatePriority)
			WindfallBoost:       1.5,  // Line 287
			BubbleRiskBoost:     1.4,  // Line 297
			NeedsRebalanceBoost: 1.3,  // Line 301
			OverweightBoost:     1.2,  // Line 306
			OvervaluedBoost:     1.15, // Line 311
			Near52wHighBoost:    1.1,  // Line 317

			// Quality-Value boosts (averaging_down.go, opportunity_buys.go calculatePriority)
			QualityValueBoost:      1.5,  // Line 406 (avg_down), Line 555 (opp_buys: 1.4)
			RecoveryCandidateBoost: 1.3,  // Line 411
			HighQualityBoost:       1.15, // Line 416
			ValueOpportunityBoost:  1.1,  // Line 421
			DeepValueBoost:         1.2,  // Line 562 (opp_buys)

			// Risk Profile boosts (base.go ApplyTagBasedPriorityBoosts)
			LowRiskBoost:    1.15, // Line 204
			MediumRiskBoost: 1.05, // Line 206
			HighRiskPenalty: 0.90, // Line 208

			// Regime-Aware boosts (base.go ApplyTagBasedPriorityBoosts)
			BullGrowthBoost:       1.15, // Line 222
			BearValueBoost:        1.15, // Line 224
			SidewaysDividendBoost: 1.12, // Line 226
			NeutralGrowthBoost:    1.08, // Line 229, 241
			NeutralValueBoost:     1.08, // Line 232, 244
			NeutralDividendBoost:  1.10, // Line 235, 247

			// Quality boosts (base.go ApplyTagBasedPriorityBoosts)
			HighStabilityBoost:       1.12, // Line 253
			ConsistentGrowerBoost:    1.10, // Line 256
			StableBoost:              1.08, // Line 259
			DividendTotalReturnBoost: 1.12, // Line 264

			// Performance-based sell boosts (base.go ApplyTagBasedPriorityBoosts)
			UnsustainableGainsBoost: 1.25, // Line 271
			StagnantBoost:           1.15, // Line 274
			UnderperformingBoost:    1.20, // Line 277
			MeetsTargetReturnBoost:  1.10, // Line 282

			// Return boosts (opportunity_buys.go calculatePriority)
			ExcellentReturnBoost: 1.25, // Line 572
			HighReturnBoost:      1.15, // Line 574
			QualityHighCAGRBoost: 1.2,  // Line 579
			DividendGrowerBoost:  1.15, // Line 589
			HighDividendBoost:    1.1,  // Line 591

			// Quantum warning penalties (base.go ApplyQuantumWarningPenalty)
			QuantumPenaltyMild:   0.90, // Line 135: 10% reduction
			QuantumPenaltySevere: 0.70, // Line 137, 141: 30% reduction
		},
	}
}
