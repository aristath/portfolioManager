// Package yahoo provides client functionality for fetching data from Yahoo Finance API.
package yahoo

// FullClientInterface defines all methods that a Yahoo Finance client must implement
type FullClientInterface interface {
	// Batch operations
	GetBatchQuotes(symbolMap map[string]*string) (map[string]*float64, error)

	// Single quote operations
	GetCurrentPrice(symbol string, yahooSymbolOverride *string, maxRetries int) (*float64, error)

	// Historical data
	GetHistoricalPrices(symbol string, yahooSymbolOverride *string, period string) ([]HistoricalPrice, error)

	// Fundamental and analyst data
	GetFundamentalData(symbol string, yahooSymbolOverride *string) (*FundamentalData, error)
	GetAnalystData(symbol string, yahooSymbolOverride *string) (*AnalystData, error)

	// Security metadata
	GetSecurityIndustry(symbol string, yahooSymbolOverride *string) (*string, error)
	GetSecurityCountryAndExchange(symbol string, yahooSymbolOverride *string) (*string, *string, error)
	GetQuoteName(symbol string, yahooSymbolOverride *string) (*string, error)
	GetQuoteType(symbol string, yahooSymbolOverride *string) (string, error)

	// ISIN lookup
	LookupTickerFromISIN(isin string) (string, error)
}

// Ensure NativeClient implements FullClientInterface
var _ FullClientInterface = (*NativeClient)(nil)
