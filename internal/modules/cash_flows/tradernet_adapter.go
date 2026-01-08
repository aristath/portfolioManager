package cash_flows

import (
	"github.com/aristath/sentinel/internal/domain"
)

// TradernetAdapter adapts the BrokerClient to the TradernetClient interface
type TradernetAdapter struct {
	brokerClient domain.BrokerClient
}

// NewTradernetAdapter creates a new adapter
func NewTradernetAdapter(brokerClient domain.BrokerClient) *TradernetAdapter {
	return &TradernetAdapter{
		brokerClient: brokerClient,
	}
}

// GetAllCashFlows fetches cash flows and converts to APITransaction format
func (a *TradernetAdapter) GetAllCashFlows(limit int) ([]APITransaction, error) {
	brokerCashFlows, err := a.brokerClient.GetAllCashFlows(limit)
	if err != nil {
		return nil, err
	}

	// Convert broker domain format to our APITransaction format
	apiTransactions := make([]APITransaction, len(brokerCashFlows))
	for i, bcf := range brokerCashFlows {
		// Extract TypeDocID from params if it was preserved by Tradernet adapter
		typeDocID := 0
		if bcf.Params != nil {
			if tid, ok := bcf.Params["tradernet_type_doc_id"]; ok {
				// Handle both int and float64 (JSON unmarshaling may produce float64)
				switch v := tid.(type) {
				case int:
					typeDocID = v
				case float64:
					typeDocID = int(v)
				}
			}
		}

		apiTransactions[i] = APITransaction{
			// Map domain.BrokerCashFlow fields to APITransaction
			TransactionID:   bcf.TransactionID,
			TypeDocID:       typeDocID, // Extracted from params (Tradernet-specific metadata)
			TransactionType: bcf.Type,  // Map Type to TransactionType
			Date:            bcf.Date,
			Amount:          bcf.Amount,
			Currency:        bcf.Currency,
			AmountEUR:       bcf.AmountEUR,
			Status:          bcf.Status,
			StatusC:         bcf.StatusC,
			Description:     bcf.Description,
			Params:          bcf.Params,
		}
	}

	return apiTransactions, nil
}

// IsConnected checks if broker is connected
func (a *TradernetAdapter) IsConnected() bool {
	return a.brokerClient.IsConnected()
}
