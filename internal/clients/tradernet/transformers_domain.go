package tradernet

import "github.com/aristath/sentinel/internal/domain"

// transformPositionsToDomain converts Tradernet positions to domain broker positions
func transformPositionsToDomain(tnPositions []Position) []domain.BrokerPosition {
	result := make([]domain.BrokerPosition, len(tnPositions))
	for i, tn := range tnPositions {
		result[i] = domain.BrokerPosition{
			Symbol:         tn.Symbol,
			Quantity:       tn.Quantity,
			AvgPrice:       tn.AvgPrice,
			CurrentPrice:   tn.CurrentPrice,
			MarketValue:    tn.MarketValue,
			MarketValueEUR: tn.MarketValueEUR,
			UnrealizedPnL:  tn.UnrealizedPnL,
			Currency:       tn.Currency,
			CurrencyRate:   tn.CurrencyRate,
		}
	}
	return result
}

// transformCashBalancesToDomain converts Tradernet cash balances to domain broker cash balances
func transformCashBalancesToDomain(tnBalances []CashBalance) []domain.BrokerCashBalance {
	result := make([]domain.BrokerCashBalance, len(tnBalances))
	for i, tn := range tnBalances {
		result[i] = domain.BrokerCashBalance{
			Currency: tn.Currency,
			Amount:   tn.Amount,
		}
	}
	return result
}

// transformOrderResultToDomain converts Tradernet order result to domain broker order result
func transformOrderResultToDomain(tnResult *OrderResult) *domain.BrokerOrderResult {
	if tnResult == nil {
		return nil
	}
	return &domain.BrokerOrderResult{
		OrderID:  tnResult.OrderID,
		Symbol:   tnResult.Symbol,
		Side:     tnResult.Side,
		Quantity: tnResult.Quantity,
		Price:    tnResult.Price,
	}
}

// transformTradesToDomain converts Tradernet trades to domain broker trades
func transformTradesToDomain(tnTrades []Trade) []domain.BrokerTrade {
	result := make([]domain.BrokerTrade, len(tnTrades))
	for i, tn := range tnTrades {
		result[i] = domain.BrokerTrade{
			OrderID:    tn.OrderID,
			Symbol:     tn.Symbol,
			Side:       tn.Side,
			Quantity:   tn.Quantity,
			Price:      tn.Price,
			ExecutedAt: tn.ExecutedAt,
		}
	}
	return result
}

// transformQuoteToDomain converts Tradernet quote to domain broker quote
func transformQuoteToDomain(tnQuote *Quote) *domain.BrokerQuote {
	if tnQuote == nil {
		return nil
	}
	return &domain.BrokerQuote{
		Symbol:    tnQuote.Symbol,
		Price:     tnQuote.Price,
		Change:    tnQuote.Change,
		ChangePct: tnQuote.ChangePct,
		Volume:    tnQuote.Volume,
		Timestamp: tnQuote.Timestamp,
	}
}

// transformPendingOrdersToDomain converts Tradernet pending orders to domain broker pending orders
func transformPendingOrdersToDomain(tnOrders []PendingOrder) []domain.BrokerPendingOrder {
	result := make([]domain.BrokerPendingOrder, len(tnOrders))
	for i, tn := range tnOrders {
		result[i] = domain.BrokerPendingOrder{
			OrderID:  tn.OrderID,
			Symbol:   tn.Symbol,
			Side:     tn.Side,
			Quantity: tn.Quantity,
			Price:    tn.Price,
			Currency: tn.Currency,
		}
	}
	return result
}

// transformSecurityInfoToDomain converts Tradernet security info to domain broker security info
func transformSecurityInfoToDomain(tnSecurities []SecurityInfo) []domain.BrokerSecurityInfo {
	result := make([]domain.BrokerSecurityInfo, len(tnSecurities))
	for i, tn := range tnSecurities {
		result[i] = domain.BrokerSecurityInfo{
			Symbol:       tn.Symbol,
			Name:         tn.Name,
			ISIN:         tn.ISIN,
			Currency:     tn.Currency,
			Market:       tn.Market,
			ExchangeCode: tn.ExchangeCode,
		}
	}
	return result
}

// transformCashFlowsToDomain converts Tradernet cash flows to domain broker cash flows
func transformCashFlowsToDomain(tnFlows []CashFlowTransaction) []domain.BrokerCashFlow {
	result := make([]domain.BrokerCashFlow, len(tnFlows))
	for i, tn := range tnFlows {
		result[i] = domain.BrokerCashFlow{
			ID:              tn.ID,
			TransactionID:   tn.TransactionID,
			TypeDocID:       tn.TypeDocID,
			Type:            tn.Type,
			TransactionType: tn.TransactionType,
			DT:              tn.DT,
			Date:            tn.Date,
			SM:              tn.SM,
			Amount:          tn.Amount,
			Curr:            tn.Curr,
			Currency:        tn.Currency,
			SMEUR:           tn.SMEUR,
			AmountEUR:       tn.AmountEUR,
			Status:          tn.Status,
			StatusC:         tn.StatusC,
			Description:     tn.Description,
			Params:          tn.Params,
		}
	}
	return result
}

// transformCashMovementsToDomain converts Tradernet cash movements to domain broker cash movements
func transformCashMovementsToDomain(tnMovements *CashMovementsResponse) *domain.BrokerCashMovement {
	if tnMovements == nil {
		return nil
	}
	return &domain.BrokerCashMovement{
		TotalWithdrawals: tnMovements.TotalWithdrawals,
		Withdrawals:      tnMovements.Withdrawals,
		Note:             tnMovements.Note,
	}
}

// transformHealthResultToDomain converts Tradernet health result to domain broker health result
func transformHealthResultToDomain(tnHealth *HealthCheckResult) *domain.BrokerHealthResult {
	if tnHealth == nil {
		return nil
	}
	return &domain.BrokerHealthResult{
		Connected: tnHealth.Connected,
		Timestamp: tnHealth.Timestamp,
	}
}
