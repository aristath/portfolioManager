// Package portfolio provides portfolio management functionality.
package portfolio

import (
	"github.com/aristath/sentinel/internal/domain"
)

// PortfolioSummaryAdapter adapts PortfolioService to domain.PortfolioSummaryProvider
// This adapter breaks the circular dependency: allocation → portfolio
type PortfolioSummaryAdapter struct {
	service *PortfolioService
}

// NewPortfolioSummaryAdapter creates a new adapter
func NewPortfolioSummaryAdapter(service *PortfolioService) domain.PortfolioSummaryProvider {
	return &PortfolioSummaryAdapter{service: service}
}

// GetPortfolioSummary implements domain.PortfolioSummaryProvider
func (a *PortfolioSummaryAdapter) GetPortfolioSummary() (domain.PortfolioSummary, error) {
	portfolioSummary, err := a.service.GetPortfolioSummary()
	if err != nil {
		return domain.PortfolioSummary{}, err
	}

	// Convert portfolio.PortfolioSummary to domain.PortfolioSummary
	return domain.PortfolioSummary{
		GeographyAllocations: convertAllocationsToDomain(portfolioSummary.GeographyAllocations),
		IndustryAllocations:  convertAllocationsToDomain(portfolioSummary.IndustryAllocations),
		TotalValue:           portfolioSummary.TotalValue,
		CashBalance:          portfolioSummary.CashBalance,
	}, nil
}

// convertAllocationsToDomain converts []AllocationStatus to []domain.PortfolioAllocation
func convertAllocationsToDomain(src []AllocationStatus) []domain.PortfolioAllocation {
	result := make([]domain.PortfolioAllocation, len(src))
	for i, a := range src {
		result[i] = domain.PortfolioAllocation{
			Name:         a.Name,
			TargetPct:    a.TargetPct,
			CurrentPct:   a.CurrentPct,
			CurrentValue: a.CurrentValue,
			Deviation:    a.Deviation,
		}
	}
	return result
}
