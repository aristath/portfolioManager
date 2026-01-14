package clientdata

import "time"

// TTL constants for different data types.
// These are added to time.Now() when storing to calculate expires_at.
// Note: External data source TTLs (Alpha Vantage, Yahoo, OpenFIGI) removed - only Tradernet is used.
const (
	// Very stable data (rarely changes)
	TTLSymbolToISIN = 30 * 24 * time.Hour // 30 days - Symbol-to-ISIN mappings rarely change

	// Short-lived data (changes frequently)
	TTLExchangeRate = time.Hour        // 1 hour - Currency exchange rates
	TTLCurrentPrice = 10 * time.Minute // 10 minutes - Current price cache for batch operations
)
