package universe

import (
	"github.com/aristath/sentinel/internal/modules/settings"
)

// TagAssignerConfig holds all thresholds used by the tag assigner.
// These thresholds control when various tags are assigned to securities.
type TagAssignerConfig struct {
	// Value thresholds for value opportunity detection
	Value settings.ValueThresholds

	// Quality thresholds for quality tag assignment
	Quality settings.QualityThresholds

	// Technical thresholds for technical indicator tags
	Technical settings.TechnicalThresholds

	// Dividend thresholds for dividend-related tags
	Dividend settings.DividendThresholds

	// Danger thresholds for warning/danger tags
	Danger DangerConfig

	// PortfolioRisk thresholds for portfolio risk tags
	PortfolioRisk settings.PortfolioRiskThresholds

	// RiskProfile thresholds for risk classification
	RiskProfile settings.RiskProfileThresholds

	// QualityGate thresholds for multi-path quality gates
	QualityGate QualityGateConfig

	// BubbleTrap thresholds for bubble and value trap detection
	BubbleTrap settings.BubbleTrapThresholds

	// TotalReturn thresholds for return classification
	TotalReturn settings.TotalReturnThresholds

	// Regime thresholds for regime-specific tags
	Regime settings.RegimeThresholds
}

// DangerConfig holds danger/warning thresholds not covered by settings.DangerThresholds
type DangerConfig struct {
	VolatileThreshold        float64 // Threshold for "volatile" tag
	HighVolatilityThreshold  float64 // Threshold for "high-volatility" tag
	UnsustainableGainsReturn float64 // Return threshold for "unsustainable-gains"
	ValuationStretchEMA      float64 // EMA stretch for "valuation-stretch"
	UnderperformingDays      int     // Days for "underperforming" tag
	StagnantReturnThreshold  float64 // Return threshold for "stagnant"
	StagnantDaysThreshold    int     // Days threshold for "stagnant"
	HighDrawdownThreshold    float64 // Drawdown threshold for "high-drawdown"
}

// QualityGateConfig holds multi-path quality gate thresholds
type QualityGateConfig struct {
	StabilityThreshold float64 // Path 1: Balanced stability threshold
	LongTermThreshold     float64 // Path 1: Balanced long-term threshold
	ExceptionalThreshold  float64 // Path 2: Exceptional excellence threshold
}

// NewTagAssignerConfig creates a TagAssignerConfig from the settings service.
// This is the recommended way to create config as it respects temperament settings.
func NewTagAssignerConfig(settingsService *settings.Service) TagAssignerConfig {
	return TagAssignerConfig{
		Value:         settingsService.GetAdjustedValueThresholds(),
		Quality:       settingsService.GetAdjustedQualityThresholds(),
		Technical:     settingsService.GetAdjustedTechnicalThresholds(),
		Dividend:      settingsService.GetAdjustedDividendThresholds(),
		Danger:        newDangerConfigFromSettings(settingsService),
		PortfolioRisk: settingsService.GetAdjustedPortfolioRiskThresholds(),
		RiskProfile:   settingsService.GetAdjustedRiskProfileThresholds(),
		QualityGate:   newQualityGateConfigFromSettings(settingsService),
		BubbleTrap:    settingsService.GetAdjustedBubbleTrapThresholds(),
		TotalReturn:   settingsService.GetAdjustedTotalReturnThresholds(),
		Regime:        settingsService.GetAdjustedRegimeThresholds(),
	}
}

// newDangerConfigFromSettings creates a DangerConfig from settings service
func newDangerConfigFromSettings(settingsService *settings.Service) DangerConfig {
	danger := settingsService.GetAdjustedDangerThresholds()
	vol := settingsService.GetAdjustedVolatilityParams()

	return DangerConfig{
		VolatileThreshold:        vol.VolatileThreshold,
		HighVolatilityThreshold:  vol.HighThreshold,
		UnsustainableGainsReturn: danger.UnsustainableGainsReturn,
		ValuationStretchEMA:      danger.ValuationStretchEMA,
		UnderperformingDays:      danger.UnderperformingDays,
		StagnantReturnThreshold:  danger.StagnantReturnThreshold,
		StagnantDaysThreshold:    danger.StagnantDaysThreshold,
		HighDrawdownThreshold:    vol.MaxAcceptableDrawdown, // Use max acceptable as high threshold
	}
}

// newQualityGateConfigFromSettings creates a QualityGateConfig from settings service
func newQualityGateConfigFromSettings(settingsService *settings.Service) QualityGateConfig {
	quality := settingsService.GetAdjustedQualityGateParams()
	return QualityGateConfig{
		StabilityThreshold: quality.StabilityThreshold,
		LongTermThreshold:     quality.LongTermThreshold,
		ExceptionalThreshold:  quality.ExceptionalThreshold,
	}
}

// NewDefaultTagAssignerConfig creates a TagAssignerConfig with default values.
// These match the current hardcoded values in tag_assigner.go.
func NewDefaultTagAssignerConfig() TagAssignerConfig {
	return TagAssignerConfig{
		Value: settings.ValueThresholds{
			ValueOpportunityDiscountPct: 0.15, // Line 171: below52wHighPct > 15.0
			DeepValueDiscountPct:        0.25, // Line 176: below52wHighPct > 25.0
			DeepValueExtremePct:         0.30, // Line 176: below52wHighPct > 30.0
			Below52wHighThreshold:       0.10, // Line 180: below52wHighPct > 10.0
		},
		Quality: settings.QualityThresholds{
			HighQualityStability:     0.70, // Line 189: stabilityScore > 0.7
			HighQualityLongTerm:         0.70, // Line 189: longTermScore > 0.7
			StableStability:          0.75, // Line 193: stabilityScore > 0.75
			StableVolatilityMax:         0.25, // Line 193: volatility < 0.25
			StableConsistency:           0.75, // Line 193: consistencyScore > 0.75
			ConsistentGrowerConsistency: 0.75, // Line 199: consistencyScore > 0.75
			ConsistentGrowerCAGR:        0.09, // Line 199: cagrRaw > 0.09
			HighStabilityThreshold: 0.75, // Line 203: stabilityScore > 0.75
		},
		Technical: settings.TechnicalThresholds{
			RSIOversold:               30.0,  // Line 210: *input.RSI < 30
			RSIOverbought:             70.0,  // Line 269: *input.RSI > 70
			RecoveryMomentumThreshold: -0.05, // Line 234: momentumScore < 0 (using -0.05 as threshold)
			RecoveryStabilityMin:   0.65,  // Line 234: stabilityScore > 0.65
			RecoveryDiscountMin:       0.12,  // Line 234: below52wHighPct > 12.0 (as decimal)
		},
		Dividend: settings.DividendThresholds{
			HighDividendYield:        0.04,  // Line 217: dividendYield > 0.04
			DividendOpportunityScore: 0.55,  // Line 221: dividendScore > 0.55
			DividendOpportunityYield: 0.025, // Line 221: dividendYield > 0.025
			DividendConsistencyScore: 0.70,  // Line 225: dividendConsistencyScore > 0.7
		},
		Danger: DangerConfig{
			VolatileThreshold:        0.30, // Line 248: volatility > 0.30
			HighVolatilityThreshold:  0.40, // Line 256: volatility > 0.40
			UnsustainableGainsReturn: 0.50, // Line 276: annualizedReturn > 0.50
			ValuationStretchEMA:      0.30, // Line 280: math.Abs(distanceFromEMA) > 0.30
			UnderperformingDays:      180,  // Line 285: daysHeld > 180
			StagnantReturnThreshold:  0.05, // Line 289: annualizedReturn < 0.05
			StagnantDaysThreshold:    365,  // Line 289: daysHeld > 365
			HighDrawdownThreshold:    0.30, // Line 293: maxDrawdown > 30.0 (as decimal)
		},
		PortfolioRisk: settings.PortfolioRiskThresholds{
			OverweightDeviation:        0.02, // Line 298: targetWeight+0.02
			OverweightAbsolute:         0.10, // Line 298: positionWeight > 0.10
			ConcentrationRiskThreshold: 0.15, // Line 302: positionWeight > 0.15
			NeedsRebalanceDeviation:    0.03, // Line 312: deviation > 0.03
		},
		RiskProfile: settings.RiskProfileThresholds{
			LowRiskVolatilityMax:          0.15, // Line 320: volatility < 0.15
			LowRiskStabilityMin:        0.70, // Line 320: stabilityScore > 0.7
			LowRiskDrawdownMax:            0.20, // Line 320: maxDrawdown < 20.0
			MediumRiskVolatilityMin:       0.15, // Line 324: volatility >= 0.15
			MediumRiskVolatilityMax:       0.30, // Line 324: volatility <= 0.30
			MediumRiskStabilityMin:     0.55, // Line 324: stabilityScore > 0.55
			HighRiskVolatilityThreshold:   0.30, // Line 328: volatility > 0.30
			HighRiskStabilityThreshold: 0.50, // Line 328: stabilityScore < 0.5
		},
		QualityGate: QualityGateConfig{
			StabilityThreshold: 0.55, // Line 348: stabilityThreshold := 0.55
			LongTermThreshold:     0.45, // Line 349: longTermThreshold := 0.45
			ExceptionalThreshold:  0.75, // Line 736: 0.75 (Path 2)
		},
		BubbleTrap: settings.BubbleTrapThresholds{
			BubbleCAGRThreshold:         0.15,  // Line 461: cagrRaw > 0.15
			BubbleSharpeThreshold:       0.50,  // Line 464: sharpeRatio < 0.5
			BubbleVolatilityThreshold:   0.40,  // Line 464: volatility > 0.40
			BubbleStabilityThreshold: 0.55,  // Line 464: stabilityScore < 0.55
			ValueTrapStability:       0.55,  // Line 525: stabilityScore < 0.55
			ValueTrapLongTerm:           0.45,  // Line 525: longTermScore < 0.45
			ValueTrapMomentum:           -0.05, // Line 525: momentumScore < -0.05
			ValueTrapVolatility:         0.35,  // Line 525: volatility > 0.35
			QuantumBubbleHighProb:       0.70,  // Line 511: quantumBubbleProb > 0.7
			QuantumBubbleWarningProb:    0.50,  // Line 515: quantumBubbleProb > 0.5
			QuantumTrapHighProb:         0.70,  // Line 552: quantumTrapProb > 0.7
			QuantumTrapWarningProb:      0.50,  // Line 556: quantumTrapProb > 0.5
		},
		TotalReturn: settings.TotalReturnThresholds{
			ExcellentTotalReturn:     0.18, // Line 567: totalReturn >= 0.18
			HighTotalReturn:          0.15, // Line 569: totalReturn >= 0.15
			ModerateTotalReturn:      0.12, // Line 571: totalReturn >= 0.12
			DividendTotalReturnYield: 0.08, // Line 576: dividendYield >= 0.08
			DividendTotalReturnCAGR:  0.05, // Line 576: cagrRaw >= 0.05
		},
		Regime: settings.RegimeThresholds{
			BearSafeVolatility:       0.20, // Line 583: volatility < 0.20
			BearSafeStability:     0.70, // Line 583: stabilityScore > 0.70
			BearSafeDrawdown:         0.20, // Line 583: maxDrawdown < 20.0
			BullGrowthCAGR:           0.12, // Line 588: cagrRaw > 0.12
			BullGrowthStability:   0.70, // Line 588: stabilityScore > 0.7
			RegimeVolatileVolatility: 0.30, // Line 606: volatility > 0.30
		},
	}
}
