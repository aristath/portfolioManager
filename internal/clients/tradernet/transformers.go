package tradernet

import (
	"fmt"
)

// transformPositions transforms SDK AccountSummary positions to []Position
func transformPositions(sdkResult interface{}) ([]Position, error) {
	resultMap, ok := sdkResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: expected map[string]interface{}")
	}

	result, ok := resultMap["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: missing 'result' field")
	}

	ps, ok := result["ps"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: missing 'ps' field")
	}

	posArray, ok := ps["pos"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: missing or invalid 'pos' array")
	}

	positions := make([]Position, 0, len(posArray))
	for _, posItem := range posArray {
		posMap, ok := posItem.(map[string]interface{})
		if !ok {
			continue
		}

		position := Position{
			Symbol:        getString(posMap, "i"),
			Quantity:      getFloat64(posMap, "q"),
			AvgPrice:      getFloat64(posMap, "bal_price_a"),
			CurrentPrice:  getFloat64(posMap, "mkt_price"),
			UnrealizedPnL: getFloat64(posMap, "profit_close"),
			Currency:      getString(posMap, "curr"),
			CurrencyRate:  1.0, // Default, will be calculated if needed
		}

		// Calculate MarketValue = Quantity * CurrentPrice
		position.MarketValue = position.Quantity * position.CurrentPrice
		// MarketValueEUR will be calculated by client if needed
		position.MarketValueEUR = position.MarketValue

		positions = append(positions, position)
	}

	return positions, nil
}

// transformCashBalances transforms SDK AccountSummary cash accounts to []CashBalance
func transformCashBalances(sdkResult interface{}) ([]CashBalance, error) {
	resultMap, ok := sdkResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: expected map[string]interface{}")
	}

	result, ok := resultMap["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: missing 'result' field")
	}

	ps, ok := result["ps"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: missing 'ps' field")
	}

	accArray, ok := ps["acc"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: missing or invalid 'acc' array")
	}

	balances := make([]CashBalance, 0, len(accArray))
	for _, accItem := range accArray {
		accMap, ok := accItem.(map[string]interface{})
		if !ok {
			continue
		}

		balance := CashBalance{
			Currency: getString(accMap, "curr"),
			Amount:   getFloat64(accMap, "s"),
		}

		balances = append(balances, balance)
	}

	return balances, nil
}

// transformOrderResult transforms SDK Buy/Sell response to OrderResult
func transformOrderResult(sdkResult interface{}, symbol, side string, quantity float64) (*OrderResult, error) {
	resultMap, ok := sdkResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: expected map[string]interface{}")
	}

	// Extract order ID - check both 'id' and 'order_id' fields
	var orderID string
	if idVal, exists := resultMap["order_id"]; exists {
		orderID = fmt.Sprintf("%v", idVal)
	} else if idVal, exists := resultMap["id"]; exists {
		orderID = fmt.Sprintf("%v", idVal)
	} else {
		return nil, fmt.Errorf("invalid SDK result format: missing 'id' or 'order_id' field")
	}

	// Extract price - check both 'price' and 'p' fields
	var price float64
	if pVal, exists := resultMap["price"]; exists {
		price = getFloat64FromValue(pVal)
	} else if pVal, exists := resultMap["p"]; exists {
		price = getFloat64FromValue(pVal)
	} else {
		price = 0.0
	}

	return &OrderResult{
		OrderID:  orderID,
		Symbol:   symbol,
		Side:     side,
		Quantity: quantity,
		Price:    price,
	}, nil
}

// transformPendingOrders transforms SDK GetPlaced response to []PendingOrder
func transformPendingOrders(sdkResult interface{}) ([]PendingOrder, error) {
	resultMap, ok := sdkResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: expected map[string]interface{}")
	}

	// Handle empty or null result
	result, ok := resultMap["result"]
	if !ok || result == nil {
		// Empty result - return empty array
		return []PendingOrder{}, nil
	}

	resultArray, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: 'result' must be array, got %T", result)
	}

	orders := make([]PendingOrder, 0, len(resultArray))
	for _, orderItem := range resultArray {
		orderMap, ok := orderItem.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract order ID - check both 'id' and 'orderId' fields
		var orderID string
		if idVal, exists := orderMap["orderId"]; exists {
			orderID = fmt.Sprintf("%v", idVal)
		} else if idVal, exists := orderMap["id"]; exists {
			orderID = fmt.Sprintf("%v", idVal)
		} else {
			continue // Skip orders without ID
		}

		order := PendingOrder{
			OrderID:  orderID,
			Symbol:   getString(orderMap, "i"),
			Quantity: getFloat64(orderMap, "q"),
			Price:    getFloat64(orderMap, "p"),
			Currency: getString(orderMap, "curr"),
		}

		orders = append(orders, order)
	}

	return orders, nil
}

// transformCashMovements transforms SDK GetClientCpsHistory to CashMovementsResponse
func transformCashMovements(sdkResult interface{}) (*CashMovementsResponse, error) {
	resultMap, ok := sdkResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: expected map[string]interface{}")
	}

	// Handle empty or null result
	result, ok := resultMap["result"]
	if !ok || result == nil {
		// Empty result - return empty response
		return &CashMovementsResponse{
			Withdrawals: []map[string]interface{}{},
		}, nil
	}

	resultArray, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: 'result' must be array, got %T", result)
	}

	withdrawals := make([]map[string]interface{}, 0, len(resultArray))
	var totalWithdrawals float64

	for _, item := range resultArray {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		withdrawals = append(withdrawals, itemMap)

		// Sum up withdrawal amounts if available
		if amount, exists := itemMap["amount"]; exists {
			if amtFloat, ok := amount.(float64); ok {
				totalWithdrawals += amtFloat
			}
		}
	}

	return &CashMovementsResponse{
		TotalWithdrawals: totalWithdrawals,
		Withdrawals:      withdrawals,
		Note:             "",
	}, nil
}

// transformCashFlows transforms SDK responses to []CashFlowTransaction
func transformCashFlows(sdkResult interface{}) ([]CashFlowTransaction, error) {
	resultMap, ok := sdkResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: expected map[string]interface{}")
	}

	// Handle empty or null result
	result, ok := resultMap["result"]
	if !ok || result == nil {
		// Empty result - return empty array
		return []CashFlowTransaction{}, nil
	}

	resultArray, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: 'result' must be array, got %T", result)
	}

	transactions := make([]CashFlowTransaction, 0, len(resultArray))
	for _, item := range resultArray {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		tx := CashFlowTransaction{
			ID:              getString(itemMap, "id"),
			TransactionID:   getString(itemMap, "transaction_id"),
			TypeDocID:       int(getFloat64(itemMap, "type_doc_id")),
			Type:            getString(itemMap, "type"),
			TransactionType: getString(itemMap, "transaction_type"),
			DT:              getString(itemMap, "dt"),
			Date:            getString(itemMap, "date"),
			SM:              getFloat64(itemMap, "sm"),
			Amount:          getFloat64(itemMap, "amount"),
			Curr:            getString(itemMap, "curr"),
			Currency:        getString(itemMap, "currency"),
			SMEUR:           getFloat64(itemMap, "sm_eur"),
			AmountEUR:       getFloat64(itemMap, "amount_eur"),
			Status:          getString(itemMap, "status"),
			StatusC:         int(getFloat64(itemMap, "status_c")),
			Description:     getString(itemMap, "description"),
		}

		// Handle params field
		if params, exists := itemMap["params"]; exists {
			if paramsMap, ok := params.(map[string]interface{}); ok {
				tx.Params = paramsMap
			} else {
				tx.Params = make(map[string]interface{})
			}
		} else {
			tx.Params = make(map[string]interface{})
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// transformTrades transforms SDK GetTradesHistory to []Trade
func transformTrades(sdkResult interface{}) ([]Trade, error) {
	resultMap, ok := sdkResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: expected map[string]interface{}")
	}

	// Handle empty or null result
	result, ok := resultMap["result"]
	if !ok || result == nil {
		// Empty result - return empty array
		return []Trade{}, nil
	}

	resultArray, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: 'result' must be array, got %T", result)
	}

	trades := make([]Trade, 0, len(resultArray))
	for _, item := range resultArray {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract order ID - check both 'order_id' and 'id' fields
		var orderID string
		if idVal, exists := itemMap["order_id"]; exists {
			orderID = fmt.Sprintf("%v", idVal)
		} else if idVal, exists := itemMap["id"]; exists {
			orderID = fmt.Sprintf("%v", idVal)
		} else {
			continue // Skip trades without ID
		}

		trade := Trade{
			OrderID:    orderID,
			Symbol:     getString(itemMap, "i"),
			Side:       getString(itemMap, "side"),
			Quantity:   getFloat64(itemMap, "q"),
			Price:      getFloat64(itemMap, "p"),
			ExecutedAt: getString(itemMap, "executed_at"),
		}

		trades = append(trades, trade)
	}

	return trades, nil
}

// transformSecurityInfo transforms SDK FindSymbol to []SecurityInfo
// Handles both normalized format ({"result": [...]}) and raw API format ({"found": [...]})
// Maps short field names from API ("t", "nm", "x_curr", etc.) to expected field names
func transformSecurityInfo(sdkResult interface{}) ([]SecurityInfo, error) {
	resultMap, ok := sdkResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: expected map[string]interface{}")
	}

	// Handle both "result" (normalized) and "found" (raw API response from tickerFinder)
	var result interface{}
	var okResult bool
	if result, okResult = resultMap["found"]; !okResult || result == nil {
		// Fallback to "result" for normalized responses
		result, okResult = resultMap["result"]
	}
	if !okResult || result == nil {
		// Empty result - return empty array
		return []SecurityInfo{}, nil
	}

	resultArray, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: 'found'/'result' must be array, got %T", result)
	}

	securities := make([]SecurityInfo, 0, len(resultArray))
	for _, item := range resultArray {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Map field names: API uses short names ("t", "nm", "x_curr") but also supports full names
		// Try short names first (raw API format), fallback to full names (normalized format)
		symbol := getString(itemMap, "t") // Short form
		if symbol == "" {
			symbol = getString(itemMap, "symbol") // Full form (normalized)
		}

		if symbol == "" {
			continue // Skip items without symbol
		}

		sec := SecurityInfo{
			Symbol: symbol,
		}

		// Name: "nm" (short) or "name" (full)
		if nameVal, exists := itemMap["nm"]; exists && nameVal != nil {
			if nameStr, ok := nameVal.(string); ok && nameStr != "" {
				sec.Name = &nameStr
			}
		}
		if sec.Name == nil {
			if nameVal, exists := itemMap["name"]; exists && nameVal != nil {
				if nameStr, ok := nameVal.(string); ok && nameStr != "" {
					sec.Name = &nameStr
				}
			}
		}

		// ISIN: same in both formats
		if isin, exists := itemMap["isin"]; exists && isin != nil {
			if isinStr, ok := isin.(string); ok && isinStr != "" {
				sec.ISIN = &isinStr
			}
		}

		// Currency: "x_curr" (short) or "currency" (full)
		if currVal, exists := itemMap["x_curr"]; exists && currVal != nil {
			if currStr, ok := currVal.(string); ok && currStr != "" {
				sec.Currency = &currStr
			}
		}
		if sec.Currency == nil {
			if currVal, exists := itemMap["currency"]; exists && currVal != nil {
				if currStr, ok := currVal.(string); ok && currStr != "" {
					sec.Currency = &currStr
				}
			}
		}

		// Market: "mkt" (short) or "market" (full)
		if mktVal, exists := itemMap["mkt"]; exists && mktVal != nil {
			if mktStr, ok := mktVal.(string); ok && mktStr != "" {
				sec.Market = &mktStr
			}
		}
		if sec.Market == nil {
			if mktVal, exists := itemMap["market"]; exists && mktVal != nil {
				if mktStr, ok := mktVal.(string); ok && mktStr != "" {
					sec.Market = &mktStr
				}
			}
		}

		// Exchange code: "codesub" (short) or "exchange_code" (full)
		if exVal, exists := itemMap["codesub"]; exists && exVal != nil {
			if exStr, ok := exVal.(string); ok && exStr != "" {
				sec.ExchangeCode = &exStr
			}
		}
		if sec.ExchangeCode == nil {
			if exVal, exists := itemMap["exchange_code"]; exists && exVal != nil {
				if exStr, ok := exVal.(string); ok && exStr != "" {
					sec.ExchangeCode = &exStr
				}
			}
		}

		securities = append(securities, sec)
	}

	return securities, nil
}

// transformQuote transforms SDK GetQuotes to Quote
// Handles both array and map response formats from getStockQuotesJson
func transformQuote(sdkResult interface{}, symbol string) (*Quote, error) {
	resultMap, ok := sdkResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: expected map[string]interface{}")
	}

	result, ok := resultMap["result"]
	if !ok {
		return nil, fmt.Errorf("invalid SDK result format: missing 'result' field")
	}

	var symbolData map[string]interface{}

	// Handle array format: result is an array of quote objects
	if resultArray, ok := result.([]interface{}); ok {
		// Search for the quote with matching symbol
		found := false
		for _, item := range resultArray {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			// Check if this item matches the symbol
			// The symbol might be in different fields: "symbol", "i", "ticker", etc.
			itemSymbol := getString(itemMap, "symbol")
			if itemSymbol == "" {
				itemSymbol = getString(itemMap, "i")
			}
			if itemSymbol == "" {
				itemSymbol = getString(itemMap, "ticker")
			}
			if itemSymbol == symbol {
				symbolData = itemMap
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("quote not found for symbol: %s", symbol)
		}
	} else if resultMapData, ok := result.(map[string]interface{}); ok {
		// Handle map format: result is a map keyed by symbol
		var found bool
		symbolData, found = resultMapData[symbol].(map[string]interface{})
		if !found {
			return nil, fmt.Errorf("quote not found for symbol: %s", symbol)
		}
	} else {
		return nil, fmt.Errorf("invalid SDK result format: 'result' must be array or map, got %T", result)
	}

	quote := &Quote{
		Symbol:    symbol,
		Price:     getFloat64(symbolData, "p"),
		Change:    getFloat64(symbolData, "change"),
		ChangePct: getFloat64(symbolData, "change_pct"),
		Volume:    int64(getFloat64(symbolData, "volume")),
		Timestamp: getString(symbolData, "timestamp"),
	}

	// Handle alternative field names (fallback)
	if quote.Price == 0 {
		quote.Price = getFloat64(symbolData, "ltp")
	}
	if quote.Price == 0 {
		quote.Price = getFloat64(symbolData, "last_price")
	}
	if quote.Change == 0 {
		quote.Change = getFloat64(symbolData, "chg")
	}
	if quote.ChangePct == 0 {
		quote.ChangePct = getFloat64(symbolData, "chg_pc")
	}
	if quote.Volume == 0 {
		quote.Volume = int64(getFloat64(symbolData, "v"))
	}

	return quote, nil
}

// Helper functions

// getString safely extracts a string value from a map
func getString(m map[string]interface{}, key string) string {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
		// Try to convert other types to string
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// getFloat64 safely extracts a float64 value from a map
func getFloat64(m map[string]interface{}, key string) float64 {
	if val, exists := m[key]; exists {
		return getFloat64FromValue(val)
	}
	return 0.0
}

// getFloat64FromValue safely converts a value to float64
func getFloat64FromValue(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	default:
		return 0.0
	}
}
