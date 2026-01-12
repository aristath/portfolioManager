package services

// YahooNativeClientInterface defines the full Yahoo native client interface
// This is used to wrap the actual Yahoo client for the PriceValidator
type YahooNativeClientInterface interface {
	GetCurrentPrice(symbol string, yahooSymbolOverride *string, maxRetries int) (*float64, error)
}

// YahooPriceAdapter adapts the full Yahoo client interface to the simple PriceValidator interface
type YahooPriceAdapter struct {
	client YahooNativeClientInterface
}

// NewYahooPriceAdapter creates a new Yahoo price adapter
func NewYahooPriceAdapter(client YahooNativeClientInterface) *YahooPriceAdapter {
	return &YahooPriceAdapter{client: client}
}

// GetCurrentPrice implements YahooClientInterface for PriceValidator
// Takes a Yahoo symbol directly and uses it as the override (not as Tradernet symbol)
// This is correct because PriceValidator passes the yahoo_symbol field from securities table
func (a *YahooPriceAdapter) GetCurrentPrice(yahooSymbol string) (float64, error) {
	// Pass the Yahoo symbol as the override, not as the first argument
	// The first argument is for Tradernet symbols; we have a Yahoo symbol
	price, err := a.client.GetCurrentPrice("", &yahooSymbol, 3)
	if err != nil {
		return 0, err
	}
	if price == nil {
		return 0, nil
	}
	return *price, nil
}
