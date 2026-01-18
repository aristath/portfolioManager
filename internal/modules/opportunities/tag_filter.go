package opportunities

import (
	"github.com/aristath/sentinel/internal/modules/planning/domain"
	"github.com/rs/zerolog"
)

// TagBasedFilter provides intelligent tag-based pre-filtering for opportunity identification.
// It uses tags to quickly reduce the candidate set from 100+ securities to 10-20,
// enabling focused calculations on a smaller, higher-quality set.
// Follows Dependency Inversion Principle - depends on SecurityRepository interface.
type TagBasedFilter struct {
	securityRepo SecurityRepository
	log          zerolog.Logger
}

// NewTagBasedFilter creates a new tag-based filter.
// Accepts SecurityRepository interface, not concrete implementation.
func NewTagBasedFilter(securityRepo SecurityRepository, log zerolog.Logger) *TagBasedFilter {
	return &TagBasedFilter{
		securityRepo: securityRepo,
		log:          log.With().Str("component", "tag_filter").Logger(),
	}
}

// GetOpportunityCandidates uses exclusion-based tag filtering to identify buying opportunity candidates.
// Returns all active securities EXCEPT those with exclusion tags (bad tags).
// This is more inclusive than inclusion-based filtering and aligns with "tag what's wrong, not what's right" philosophy.
// If config is provided and EnableTagFiltering is false, returns all active securities.
func (f *TagBasedFilter) GetOpportunityCandidates(ctx *domain.OpportunityContext, config *domain.PlannerConfiguration) ([]string, error) {
	if ctx == nil {
		return nil, nil
	}

	// Get all active securities first
	allSecurities, err := f.securityRepo.GetAllActive()
	if err != nil {
		return nil, err
	}

	// If tag filtering is disabled, return all active securities
	if config != nil && !config.EnableTagFiltering {
		f.log.Debug().Msg("Tag filtering disabled, returning all active securities")
		symbols := make([]string, 0, len(allSecurities))
		for _, sec := range allSecurities {
			if sec.Symbol != "" {
				symbols = append(symbols, sec.Symbol)
			}
		}

		f.log.Debug().
			Int("candidates", len(symbols)).
			Msg("Returned all active securities (tag filtering disabled)")

		return symbols, nil
	}

	// Get exclusion tags (bad tags to exclude)
	exclusionTags := f.selectExclusionTags(ctx, config)
	if len(exclusionTags) == 0 {
		// No exclusions - return all active securities
		symbols := make([]string, 0, len(allSecurities))
		for _, sec := range allSecurities {
			if sec.Symbol != "" {
				symbols = append(symbols, sec.Symbol)
			}
		}

		f.log.Debug().
			Int("candidates", len(symbols)).
			Msg("Tag-based pre-filtering complete (no exclusions)")

		return symbols, nil
	}

	f.log.Debug().
		Strs("exclusion_tags", exclusionTags).
		Msg("Excluding securities with bad tags")

	// Get securities with exclusion tags
	excludedSecurities, err := f.securityRepo.GetByTags(exclusionTags)
	if err != nil {
		return nil, err
	}

	// Build map of excluded symbols for fast lookup
	excludedSymbols := make(map[string]bool)
	for _, sec := range excludedSecurities {
		if sec.Symbol != "" {
			excludedSymbols[sec.Symbol] = true
		}
	}

	// Filter out excluded securities
	symbols := make([]string, 0, len(allSecurities))
	for _, sec := range allSecurities {
		if sec.Symbol != "" && !excludedSymbols[sec.Symbol] {
			symbols = append(symbols, sec.Symbol)
		}
	}

	f.log.Debug().
		Int("candidates", len(symbols)).
		Int("excluded", len(excludedSymbols)).
		Int("total_active", len(allSecurities)).
		Msg("Tag-based pre-filtering complete (exclusion-based)")

	return symbols, nil
}

// GetSellCandidates uses tags to quickly identify selling opportunity candidates.
// Returns a list of security symbols from positions that match sell-related tags.
// If config is provided and EnableTagFiltering is false, returns all position symbols.
func (f *TagBasedFilter) GetSellCandidates(ctx *domain.OpportunityContext, config *domain.PlannerConfiguration) ([]string, error) {
	if ctx == nil || len(ctx.EnrichedPositions) == 0 {
		return []string{}, nil
	}

	// Get position symbols
	positionSymbols := make([]string, 0, len(ctx.EnrichedPositions))
	for _, pos := range ctx.EnrichedPositions {
		if pos.Symbol != "" {
			positionSymbols = append(positionSymbols, pos.Symbol)
		}
	}

	if len(positionSymbols) == 0 {
		return []string{}, nil
	}

	// If tag filtering is disabled, return all position symbols
	if config != nil && !config.EnableTagFiltering {
		f.log.Debug().
			Int("candidates", len(positionSymbols)).
			Msg("Returned all position symbols (tag filtering disabled)")
		return positionSymbols, nil
	}

	tags := f.selectSellTags(ctx)
	if len(tags) == 0 {
		f.log.Debug().Msg("No sell tags selected")
		return []string{}, nil
	}

	f.log.Debug().
		Strs("tags", tags).
		Int("positions", len(positionSymbols)).
		Msg("Selecting sell candidates by tags")

	candidates, err := f.securityRepo.GetPositionsByTags(positionSymbols, tags)
	if err != nil {
		return nil, err
	}

	symbols := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if c.Symbol != "" {
			symbols = append(symbols, c.Symbol)
		}
	}

	f.log.Debug().
		Int("candidates", len(symbols)).
		Msg("Tag-based sell pre-filtering complete")

	return symbols, nil
}

// selectExclusionTags returns tags that should be excluded from opportunity candidates.
// Reversed logic: Instead of including securities with "good" tags, we exclude securities with "bad" tags.
// This is more inclusive and aligns with the "tag what's wrong, not what's right" philosophy.
func (f *TagBasedFilter) selectExclusionTags(ctx *domain.OpportunityContext, config *domain.PlannerConfiguration) []string {
	exclusionTags := []string{}

	// Core exclusion tags - always exclude these bad qualities
	// These are fundamental filters that prevent investing in poor-quality securities

	// Quality gate failures - securities that don't meet minimum quality standards
	exclusionTags = append(exclusionTags, "quality-gate-fail")

	// Below minimum return - securities that can't meet minimum return requirements
	exclusionTags = append(exclusionTags, "below-minimum-return")

	// Bubble risks - avoid securities with detected bubble characteristics
	exclusionTags = append(exclusionTags, "bubble-risk")
	exclusionTags = append(exclusionTags, "quantum-bubble-risk")
	exclusionTags = append(exclusionTags, "ensemble-bubble-risk")

	// Optional: Exclude regime-volatile in very volatile markets
	// This is optional as volatile securities might still be acceptable in some contexts
	// Uncomment if you want stricter filtering:
	// exclusionTags = append(exclusionTags, "regime-volatile")

	return exclusionTags
}

// selectSellTags intelligently selects tags for identifying sell opportunities.
func (f *TagBasedFilter) selectSellTags(ctx *domain.OpportunityContext) []string {
	tags := []string{}

	// Price-based sell signals
	tags = append(tags, "overvalued", "near-52w-high", "overbought")

	// Portfolio-based sell signals
	tags = append(tags, "overweight", "concentration-risk")

	// Optimizer alignment sell signals (enhanced tags)
	tags = append(tags, "needs-rebalance", "slightly-overweight")

	// Bubble detection sell signals (enhanced tags)
	tags = append(tags, "bubble-risk")

	return tags
}

// IsMarketVolatile determines if market conditions are volatile.
// Checks if many securities have volatility-spike tag as a proxy for market volatility.
// Falls back to checking all securities if tags are disabled.
func (f *TagBasedFilter) IsMarketVolatile(ctx *domain.OpportunityContext, config *domain.PlannerConfiguration) bool {
	// If tag filtering is disabled, we can't use tags to check volatility
	// Return false (conservative: assume market is not volatile)
	if config != nil && !config.EnableTagFiltering {
		f.log.Debug().Msg("Tag filtering disabled, cannot check market volatility via tags")
		return false
	}

	// Check if many securities have volatility-spike tag
	volatileSecurities, err := f.securityRepo.GetByTags([]string{"volatility-spike"})
	if err != nil {
		f.log.Warn().Err(err).Msg("Failed to check market volatility")
		return false
	}

	// Threshold for "volatile market" - if 5+ securities have volatility spike, market is volatile
	return len(volatileSecurities) >= 5
}
