package allocation

import "github.com/aristath/sentinel/internal/domain"

// Note: PortfolioSummaryProvider and ConcentrationAlertProvider have been moved to domain/interfaces.go
// They are now available as domain.PortfolioSummaryProvider and domain.ConcentrationAlertProvider
// The PortfolioSummary and ConcentrationAlert types are also in domain/interfaces.go

// PortfolioSummaryProvider is an alias for domain.PortfolioSummaryProvider for backward compatibility
type PortfolioSummaryProvider = domain.PortfolioSummaryProvider

// ConcentrationAlertProvider is an alias for domain.ConcentrationAlertProvider for backward compatibility
type ConcentrationAlertProvider = domain.ConcentrationAlertProvider

// PortfolioSummary is an alias for domain.PortfolioSummary for backward compatibility
type PortfolioSummary = domain.PortfolioSummary

// PortfolioAllocation is an alias for domain.PortfolioAllocation for backward compatibility
type PortfolioAllocation = domain.PortfolioAllocation

// ConcentrationAlert is an alias for domain.ConcentrationAlert for backward compatibility
type ConcentrationAlert = domain.ConcentrationAlert
