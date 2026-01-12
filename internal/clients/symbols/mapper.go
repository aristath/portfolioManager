// Package symbols provides unified symbol mapping across data providers.
// It handles conversion of stock ticker symbols between different financial data providers
// (Tradernet, Yahoo Finance, Alpha Vantage, OpenFIGI) which use different symbol formats
// and exchange suffixes.
package symbols

import (
	"fmt"
	"strings"
)

// Provider represents a financial data provider.
type Provider string

const (
	// ProviderTradernet is the Tradernet broker.
	ProviderTradernet Provider = "tradernet"
	// ProviderYahoo is Yahoo Finance.
	ProviderYahoo Provider = "yahoo"
	// ProviderAlphaVantage is Alpha Vantage.
	ProviderAlphaVantage Provider = "alphavantage"
	// ProviderOpenFIGI is Bloomberg's OpenFIGI service.
	ProviderOpenFIGI Provider = "openfigi"
)

// Exchange represents a stock exchange with symbol formats for all providers.
type Exchange struct {
	Code               string // Internal code (e.g., "ATH", "LSE", "HKG")
	Name               string // Full name (e.g., "Athens Stock Exchange")
	Country            string // ISO country code (e.g., "GR", "GB", "HK")
	Currency           string // Primary currency (e.g., "EUR", "GBP", "HKD")
	TradernetSuffix    string // e.g., ".GR", ".EU", ".US"
	YahooSuffix        string // e.g., ".AT", ".L", ".HK"
	AlphaVantageSuffix string // e.g., ".AT", ".LON", ".HKG"
	OpenFIGICode       string // Bloomberg exchange code (e.g., "GA", "LN", "HK")
	Timezone           string // e.g., "Europe/Athens"
}

// ExchangeMap contains all known exchanges with their provider-specific symbol formats.
var ExchangeMap = map[string]Exchange{
	// Greece
	"ATH": {
		Code:               "ATH",
		Name:               "Athens Stock Exchange",
		Country:            "GR",
		Currency:           "EUR",
		TradernetSuffix:    ".GR",
		YahooSuffix:        ".AT",
		AlphaVantageSuffix: ".AT",
		OpenFIGICode:       "GA",
		Timezone:           "Europe/Athens",
	},
	// United Kingdom
	"LSE": {
		Code:               "LSE",
		Name:               "London Stock Exchange",
		Country:            "GB",
		Currency:           "GBP",
		TradernetSuffix:    ".EU",
		YahooSuffix:        ".L",
		AlphaVantageSuffix: ".LON",
		OpenFIGICode:       "LN",
		Timezone:           "Europe/London",
	},
	// Germany - Frankfurt
	"FRA": {
		Code:               "FRA",
		Name:               "Frankfurt Stock Exchange",
		Country:            "DE",
		Currency:           "EUR",
		TradernetSuffix:    ".EU",
		YahooSuffix:        ".F",
		AlphaVantageSuffix: ".FRA",
		OpenFIGICode:       "GR",
		Timezone:           "Europe/Berlin",
	},
	// Germany - Xetra
	"XETRA": {
		Code:               "XETRA",
		Name:               "Deutsche Börse Xetra",
		Country:            "DE",
		Currency:           "EUR",
		TradernetSuffix:    ".EU",
		YahooSuffix:        ".DE",
		AlphaVantageSuffix: ".DEX",
		OpenFIGICode:       "GR",
		Timezone:           "Europe/Berlin",
	},
	// France
	"PAR": {
		Code:               "PAR",
		Name:               "Euronext Paris",
		Country:            "FR",
		Currency:           "EUR",
		TradernetSuffix:    ".EU",
		YahooSuffix:        ".PA",
		AlphaVantageSuffix: ".PAR",
		OpenFIGICode:       "FP",
		Timezone:           "Europe/Paris",
	},
	// Netherlands
	"AMS": {
		Code:               "AMS",
		Name:               "Euronext Amsterdam",
		Country:            "NL",
		Currency:           "EUR",
		TradernetSuffix:    ".EU",
		YahooSuffix:        ".AS",
		AlphaVantageSuffix: ".AMS",
		OpenFIGICode:       "NA",
		Timezone:           "Europe/Amsterdam",
	},
	// Italy
	"MIL": {
		Code:               "MIL",
		Name:               "Borsa Italiana",
		Country:            "IT",
		Currency:           "EUR",
		TradernetSuffix:    ".EU",
		YahooSuffix:        ".MI",
		AlphaVantageSuffix: ".MIL",
		OpenFIGICode:       "IM",
		Timezone:           "Europe/Rome",
	},
	// Spain
	"BME": {
		Code:               "BME",
		Name:               "Bolsa de Madrid",
		Country:            "ES",
		Currency:           "EUR",
		TradernetSuffix:    ".EU",
		YahooSuffix:        ".MC",
		AlphaVantageSuffix: ".MCE",
		OpenFIGICode:       "SM",
		Timezone:           "Europe/Madrid",
	},
	// Switzerland
	"SIX": {
		Code:               "SIX",
		Name:               "SIX Swiss Exchange",
		Country:            "CH",
		Currency:           "CHF",
		TradernetSuffix:    ".EU",
		YahooSuffix:        ".SW",
		AlphaVantageSuffix: ".SWX",
		OpenFIGICode:       "SW",
		Timezone:           "Europe/Zurich",
	},
	// Hong Kong
	"HKG": {
		Code:               "HKG",
		Name:               "Hong Kong Stock Exchange",
		Country:            "HK",
		Currency:           "HKD",
		TradernetSuffix:    ".AS",
		YahooSuffix:        ".HK",
		AlphaVantageSuffix: ".HKG",
		OpenFIGICode:       "HK",
		Timezone:           "Asia/Hong_Kong",
	},
	// Japan
	"TYO": {
		Code:               "TYO",
		Name:               "Tokyo Stock Exchange",
		Country:            "JP",
		Currency:           "JPY",
		TradernetSuffix:    ".JP",
		YahooSuffix:        ".T",
		AlphaVantageSuffix: ".TYO",
		OpenFIGICode:       "JT",
		Timezone:           "Asia/Tokyo",
	},
	// Singapore
	"SGX": {
		Code:               "SGX",
		Name:               "Singapore Exchange",
		Country:            "SG",
		Currency:           "SGD",
		TradernetSuffix:    ".AS",
		YahooSuffix:        ".SI",
		AlphaVantageSuffix: ".SGX",
		OpenFIGICode:       "SP",
		Timezone:           "Asia/Singapore",
	},
	// Australia
	"ASX": {
		Code:               "ASX",
		Name:               "Australian Securities Exchange",
		Country:            "AU",
		Currency:           "AUD",
		TradernetSuffix:    ".AS",
		YahooSuffix:        ".AX",
		AlphaVantageSuffix: ".ASX",
		OpenFIGICode:       "AU",
		Timezone:           "Australia/Sydney",
	},
	// South Korea
	"KRX": {
		Code:               "KRX",
		Name:               "Korea Exchange",
		Country:            "KR",
		Currency:           "KRW",
		TradernetSuffix:    ".AS",
		YahooSuffix:        ".KS",
		AlphaVantageSuffix: ".KRX",
		OpenFIGICode:       "KS",
		Timezone:           "Asia/Seoul",
	},
	// Taiwan
	"TWSE": {
		Code:               "TWSE",
		Name:               "Taiwan Stock Exchange",
		Country:            "TW",
		Currency:           "TWD",
		TradernetSuffix:    ".AS",
		YahooSuffix:        ".TW",
		AlphaVantageSuffix: ".TWO",
		OpenFIGICode:       "TT",
		Timezone:           "Asia/Taipei",
	},
	// India - BSE
	"BSE": {
		Code:               "BSE",
		Name:               "Bombay Stock Exchange",
		Country:            "IN",
		Currency:           "INR",
		TradernetSuffix:    ".AS",
		YahooSuffix:        ".BO",
		AlphaVantageSuffix: ".BSE",
		OpenFIGICode:       "IB",
		Timezone:           "Asia/Kolkata",
	},
	// India - NSE
	"NSE": {
		Code:               "NSE",
		Name:               "National Stock Exchange of India",
		Country:            "IN",
		Currency:           "INR",
		TradernetSuffix:    ".AS",
		YahooSuffix:        ".NS",
		AlphaVantageSuffix: ".NSE",
		OpenFIGICode:       "IN",
		Timezone:           "Asia/Kolkata",
	},
	// Canada - Toronto
	"TSX": {
		Code:               "TSX",
		Name:               "Toronto Stock Exchange",
		Country:            "CA",
		Currency:           "CAD",
		TradernetSuffix:    ".US",
		YahooSuffix:        ".TO",
		AlphaVantageSuffix: ".TRT",
		OpenFIGICode:       "CT",
		Timezone:           "America/Toronto",
	},
	// Brazil
	"B3": {
		Code:               "B3",
		Name:               "B3 - Brasil Bolsa Balcão",
		Country:            "BR",
		Currency:           "BRL",
		TradernetSuffix:    ".US",
		YahooSuffix:        ".SA",
		AlphaVantageSuffix: ".SAO",
		OpenFIGICode:       "BZ",
		Timezone:           "America/Sao_Paulo",
	},
	// USA - NYSE
	"NYSE": {
		Code:               "NYSE",
		Name:               "New York Stock Exchange",
		Country:            "US",
		Currency:           "USD",
		TradernetSuffix:    ".US",
		YahooSuffix:        "",
		AlphaVantageSuffix: "",
		OpenFIGICode:       "US",
		Timezone:           "America/New_York",
	},
	// USA - NASDAQ
	"NASDAQ": {
		Code:               "NASDAQ",
		Name:               "NASDAQ",
		Country:            "US",
		Currency:           "USD",
		TradernetSuffix:    ".US",
		YahooSuffix:        "",
		AlphaVantageSuffix: "",
		OpenFIGICode:       "US",
		Timezone:           "America/New_York",
	},
}

// Lookup maps for reverse lookups by suffix
var (
	tradernetSuffixMap    map[string]*Exchange
	yahooSuffixMap        map[string]*Exchange
	alphaVantageSuffixMap map[string]*Exchange
	openFIGICodeMap       map[string]*Exchange
)

func init() {
	tradernetSuffixMap = make(map[string]*Exchange)
	yahooSuffixMap = make(map[string]*Exchange)
	alphaVantageSuffixMap = make(map[string]*Exchange)
	openFIGICodeMap = make(map[string]*Exchange)

	for code := range ExchangeMap {
		exchange := ExchangeMap[code]
		exchangePtr := &exchange

		// Build reverse lookup maps
		// Note: Some suffixes map to multiple exchanges (e.g., ".EU" for Tradernet)
		// We store the first match and handle ambiguity through explicit mapping

		if exchange.TradernetSuffix != "" {
			// Only store if not already present (avoid overwriting)
			if _, exists := tradernetSuffixMap[exchange.TradernetSuffix]; !exists {
				tradernetSuffixMap[exchange.TradernetSuffix] = exchangePtr
			}
		}

		if exchange.YahooSuffix != "" {
			yahooSuffixMap[exchange.YahooSuffix] = exchangePtr
		}

		if exchange.AlphaVantageSuffix != "" {
			alphaVantageSuffixMap[exchange.AlphaVantageSuffix] = exchangePtr
		}

		if exchange.OpenFIGICode != "" {
			if _, exists := openFIGICodeMap[exchange.OpenFIGICode]; !exists {
				openFIGICodeMap[exchange.OpenFIGICode] = exchangePtr
			}
		}
	}

	// Explicitly set the default mappings for ambiguous suffixes
	// .US -> NYSE (default US exchange)
	if nyse, exists := ExchangeMap["NYSE"]; exists {
		tradernetSuffixMap[".US"] = &nyse
	}
	// Empty Yahoo suffix -> NYSE (US stocks don't have suffix)
	if nyse, exists := ExchangeMap["NYSE"]; exists {
		yahooSuffixMap[""] = &nyse
	}
	// Empty AlphaVantage suffix -> NYSE
	if nyse, exists := ExchangeMap["NYSE"]; exists {
		alphaVantageSuffixMap[""] = &nyse
	}
	// .GR -> ATH
	if ath, exists := ExchangeMap["ATH"]; exists {
		tradernetSuffixMap[".GR"] = &ath
	}
	// .JP -> TYO
	if tyo, exists := ExchangeMap["TYO"]; exists {
		tradernetSuffixMap[".JP"] = &tyo
	}
	// .EU -> LSE (default European, but should use Yahoo symbol for accurate mapping)
	if lse, exists := ExchangeMap["LSE"]; exists {
		tradernetSuffixMap[".EU"] = &lse
	}
	// .AS -> HKG (default Asian, but should use Yahoo symbol for accurate mapping)
	if hkg, exists := ExchangeMap["HKG"]; exists {
		tradernetSuffixMap[".AS"] = &hkg
	}
}

// GetExchangeByCode returns an exchange by its internal code.
func GetExchangeByCode(code string) *Exchange {
	if code == "" {
		return nil
	}
	if exchange, exists := ExchangeMap[code]; exists {
		return &exchange
	}
	return nil
}

// GetExchangeFromTradernetSuffix returns the exchange for a Tradernet symbol suffix.
func GetExchangeFromTradernetSuffix(suffix string) *Exchange {
	return tradernetSuffixMap[suffix]
}

// GetExchangeFromYahooSuffix returns the exchange for a Yahoo symbol suffix.
func GetExchangeFromYahooSuffix(suffix string) *Exchange {
	return yahooSuffixMap[suffix]
}

// GetExchangeFromAlphaVantageSuffix returns the exchange for an Alpha Vantage symbol suffix.
func GetExchangeFromAlphaVantageSuffix(suffix string) *Exchange {
	return alphaVantageSuffixMap[suffix]
}

// GetExchangeFromOpenFIGICode returns the exchange for an OpenFIGI exchange code.
func GetExchangeFromOpenFIGICode(code string) *Exchange {
	return openFIGICodeMap[code]
}

// ExtractSymbolAndSuffix splits a full symbol into its base symbol and suffix.
// For example: "OPAP.GR" -> ("OPAP", ".GR"), "AAPL" -> ("AAPL", "")
func ExtractSymbolAndSuffix(fullSymbol string) (symbol, suffix string) {
	if fullSymbol == "" {
		return "", ""
	}

	// Find the last dot
	lastDot := strings.LastIndex(fullSymbol, ".")
	if lastDot == -1 {
		return fullSymbol, ""
	}

	return fullSymbol[:lastDot], fullSymbol[lastDot:]
}

// ConvertSymbol converts a symbol from one provider format to another.
// Returns an error if the symbol suffix cannot be mapped.
func ConvertSymbol(symbol string, from, to Provider) (string, error) {
	if from == to {
		return symbol, nil
	}

	baseSymbol, suffix := ExtractSymbolAndSuffix(symbol)

	// Get the exchange based on the source provider's suffix
	var exchange *Exchange
	switch from {
	case ProviderTradernet:
		exchange = GetExchangeFromTradernetSuffix(suffix)
	case ProviderYahoo:
		exchange = GetExchangeFromYahooSuffix(suffix)
	case ProviderAlphaVantage:
		exchange = GetExchangeFromAlphaVantageSuffix(suffix)
	case ProviderOpenFIGI:
		exchange = GetExchangeFromOpenFIGICode(suffix)
	default:
		return "", fmt.Errorf("unknown source provider: %s", from)
	}

	if exchange == nil {
		return "", fmt.Errorf("unknown suffix %q for provider %s", suffix, from)
	}

	// Get the target suffix
	var targetSuffix string
	switch to {
	case ProviderTradernet:
		targetSuffix = exchange.TradernetSuffix
	case ProviderYahoo:
		targetSuffix = exchange.YahooSuffix
	case ProviderAlphaVantage:
		targetSuffix = exchange.AlphaVantageSuffix
	case ProviderOpenFIGI:
		targetSuffix = exchange.OpenFIGICode
	default:
		return "", fmt.Errorf("unknown target provider: %s", to)
	}

	return baseSymbol + targetSuffix, nil
}

// GetAllExchanges returns a slice of all known exchanges.
func GetAllExchanges() []Exchange {
	exchanges := make([]Exchange, 0, len(ExchangeMap))
	for _, exchange := range ExchangeMap {
		exchanges = append(exchanges, exchange)
	}
	return exchanges
}

// GetExchangesByCountry returns all exchanges for a given country code.
func GetExchangesByCountry(country string) []Exchange {
	var exchanges []Exchange
	for _, exchange := range ExchangeMap {
		if exchange.Country == country {
			exchanges = append(exchanges, exchange)
		}
	}
	return exchanges
}

// GetExchangesByCurrency returns all exchanges that use a given currency.
func GetExchangesByCurrency(currency string) []Exchange {
	var exchanges []Exchange
	for _, exchange := range ExchangeMap {
		if exchange.Currency == currency {
			exchanges = append(exchanges, exchange)
		}
	}
	return exchanges
}

// Mapper provides symbol conversion services across different providers.
// It wraps the package-level functions for use as a service in dependency injection.
type Mapper struct{}

// NewMapper creates a new symbol mapper instance.
func NewMapper() *Mapper {
	return &Mapper{}
}

// ToAlphaVantage converts a Tradernet symbol to Alpha Vantage format.
func (m *Mapper) ToAlphaVantage(tradernetSymbol string) (string, error) {
	return ConvertSymbol(tradernetSymbol, ProviderTradernet, ProviderAlphaVantage)
}

// ToYahoo converts a Tradernet symbol to Yahoo Finance format.
func (m *Mapper) ToYahoo(tradernetSymbol string) (string, error) {
	return ConvertSymbol(tradernetSymbol, ProviderTradernet, ProviderYahoo)
}

// ToTradernet converts a symbol from the given provider to Tradernet format.
func (m *Mapper) ToTradernet(symbol string, from Provider) (string, error) {
	return ConvertSymbol(symbol, from, ProviderTradernet)
}

// FromYahooToAlphaVantage converts a Yahoo symbol to Alpha Vantage format.
func (m *Mapper) FromYahooToAlphaVantage(yahooSymbol string) (string, error) {
	return ConvertSymbol(yahooSymbol, ProviderYahoo, ProviderAlphaVantage)
}

// Convert converts a symbol between any two providers.
func (m *Mapper) Convert(symbol string, from, to Provider) (string, error) {
	return ConvertSymbol(symbol, from, to)
}

// GetExchange returns the exchange for a Tradernet symbol.
func (m *Mapper) GetExchange(tradernetSymbol string) *Exchange {
	_, suffix := ExtractSymbolAndSuffix(tradernetSymbol)
	return GetExchangeFromTradernetSuffix(suffix)
}
