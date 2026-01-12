// Package services provides core application services.
package services

import (
	"fmt"
	"time"

	"github.com/aristath/sentinel/internal/clients/alphavantage"
	"github.com/aristath/sentinel/internal/clients/symbols"
	"github.com/aristath/sentinel/internal/clients/yahoo"
	"github.com/aristath/sentinel/internal/domain"
	"github.com/rs/zerolog"
)

// =============================================================================
// Domain-Agnostic Data Types
// =============================================================================

// HistoricalPrice represents a single day's price data (provider-agnostic).
type HistoricalPrice struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

// SecurityMetadata represents company metadata (provider-agnostic).
type SecurityMetadata struct {
	Name              string
	Country           string
	Exchange          string
	Industry          string
	Sector            string
	Currency          string
	ProductType       string // "EQUITY", "ETF", "ETC", etc.
	MarketCap         int64
	FullTimeEmployees int64
}

// Fundamentals represents fundamental financial data (provider-agnostic).
type Fundamentals struct {
	Symbol                     string
	PERatio                    *float64
	PEGRatio                   *float64
	PriceToBook                *float64
	PriceToSales               *float64
	EPS                        *float64
	DividendYield              *float64
	ProfitMargin               *float64
	OperatingMargin            *float64
	ReturnOnAssets             *float64
	ReturnOnEquity             *float64
	RevenueGrowthYOY           *float64
	EarningsGrowthYOY          *float64
	Beta                       *float64
	FiftyTwoWeekHigh           *float64
	FiftyTwoWeekLow            *float64
	FiftyDayMovingAverage      *float64
	TwoHundredDayMovingAverage *float64
	// Analyst ratings
	AnalystTargetPrice      *float64
	AnalystRatingStrongBuy  int
	AnalystRatingBuy        int
	AnalystRatingHold       int
	AnalystRatingSell       int
	AnalystRatingStrongSell int
}

// CurrentQuote represents a current price quote (provider-agnostic).
type CurrentQuote struct {
	Symbol    string
	Price     float64
	Change    float64
	ChangePct float64
	Volume    int64
	Timestamp time.Time
}

// =============================================================================
// Interfaces for Dependency Injection
// =============================================================================

// BrokerClientInterface defines the broker client methods needed by DataFetcherService.
type BrokerClientInterface interface {
	GetHistoricalPrices(symbol string, from, to int64, interval int) ([]domain.BrokerOHLCV, error)
	GetQuotes(symbols []string) (map[string]*domain.BrokerQuote, error)
	IsConnected() bool
}

// DataFetcherYahooClient defines the Yahoo client methods needed by DataFetcherService.
type DataFetcherYahooClient interface {
	GetHistoricalPrices(symbol string, yahooSymbolOverride *string, period string) ([]yahoo.HistoricalPrice, error)
	GetSecurityCountryAndExchange(symbol string, yahooSymbolOverride *string) (*string, *string, error)
	GetSecurityIndustry(symbol string, yahooSymbolOverride *string) (*string, error)
	GetQuoteName(symbol string, yahooSymbolOverride *string) (*string, error)
	GetQuoteType(symbol string, yahooSymbolOverride *string) (string, error)
	GetBatchQuotes(symbolMap map[string]*string) (map[string]*float64, error)
}

// AlphaVantageClientInterface defines the Alpha Vantage client methods needed.
type AlphaVantageClientInterface interface {
	GetDailyPrices(symbol string, full bool) ([]alphavantage.DailyPrice, error)
	GetCompanyOverview(symbol string) (*alphavantage.CompanyOverview, error)
	GetGlobalQuote(symbol string) (*alphavantage.GlobalQuote, error)
	GetExchangeRate(fromCurrency, toCurrency string) (*alphavantage.ExchangeRate, error)
}

// =============================================================================
// DataFetcherService
// =============================================================================

// DataFetcherService provides unified data fetching with configurable source priorities.
// It abstracts multiple data providers (Tradernet, Yahoo, Alpha Vantage) behind a
// single interface, handling symbol conversion and automatic fallback.
type DataFetcherService struct {
	router       *DataSourceRouter
	symbolMapper *symbols.Mapper

	// Clients
	brokerClient BrokerClientInterface
	yahooClient  DataFetcherYahooClient
	alphaVantage AlphaVantageClientInterface

	log zerolog.Logger
}

// NewDataFetcherService creates a new data fetcher service.
func NewDataFetcherService(
	router *DataSourceRouter,
	symbolMapper *symbols.Mapper,
	brokerClient BrokerClientInterface,
	yahooClient DataFetcherYahooClient,
	alphaVantage AlphaVantageClientInterface,
	log zerolog.Logger,
) *DataFetcherService {
	return &DataFetcherService{
		router:       router,
		symbolMapper: symbolMapper,
		brokerClient: brokerClient,
		yahooClient:  yahooClient,
		alphaVantage: alphaVantage,
		log:          log.With().Str("service", "data_fetcher").Logger(),
	}
}

// =============================================================================
// Historical Prices
// =============================================================================

// GetHistoricalPrices fetches historical price data using configured source priorities.
// The symbol parameter is the Tradernet symbol (e.g., "AAPL.US").
// For initial seeding, use years=10. For updates, use years=1.
func (s *DataFetcherService) GetHistoricalPrices(tradernetSymbol string, yahooSymbol string, years int) ([]HistoricalPrice, DataSource, error) {
	s.log.Debug().
		Str("symbol", tradernetSymbol).
		Int("years", years).
		Msg("Fetching historical prices")

	// Create fetcher functions for each source
	fetchers := map[DataSource]func() (interface{}, error){
		SourceTradernet: func() (interface{}, error) {
			return s.fetchHistoricalFromTradernet(tradernetSymbol, years)
		},
		SourceYahoo: func() (interface{}, error) {
			return s.fetchHistoricalFromYahoo(tradernetSymbol, yahooSymbol, years)
		},
		SourceAlphaVantage: func() (interface{}, error) {
			avSymbol, err := s.convertToAlphaVantageSymbol(tradernetSymbol)
			if err != nil {
				return nil, fmt.Errorf("symbol conversion failed: %w", err)
			}
			return s.fetchHistoricalFromAlphaVantage(avSymbol, years > 5)
		},
	}

	// Use router for fallback
	result, source, err := s.router.FetchWithFallback(DataTypeHistorical, func(src DataSource) (interface{}, error) {
		if fetcher, ok := fetchers[src]; ok {
			return fetcher()
		}
		return nil, fmt.Errorf("unsupported source for historical prices: %s", src)
	})

	if err != nil {
		return nil, "", err
	}

	prices, ok := result.([]HistoricalPrice)
	if !ok {
		return nil, "", fmt.Errorf("unexpected result type from historical price fetch")
	}

	s.log.Info().
		Str("symbol", tradernetSymbol).
		Str("source", string(source)).
		Int("count", len(prices)).
		Msg("Historical prices fetched")

	return prices, source, nil
}

func (s *DataFetcherService) fetchHistoricalFromTradernet(symbol string, years int) ([]HistoricalPrice, error) {
	if s.brokerClient == nil || !s.brokerClient.IsConnected() {
		return nil, fmt.Errorf("tradernet client not available")
	}

	now := time.Now()
	from := now.AddDate(-years, 0, 0)

	ohlcData, err := s.brokerClient.GetHistoricalPrices(symbol, from.Unix(), now.Unix(), 86400)
	if err != nil {
		return nil, err
	}

	prices := make([]HistoricalPrice, len(ohlcData))
	for i, ohlc := range ohlcData {
		prices[i] = HistoricalPrice{
			Date:   time.Unix(ohlc.Timestamp, 0),
			Open:   ohlc.Open,
			High:   ohlc.High,
			Low:    ohlc.Low,
			Close:  ohlc.Close,
			Volume: ohlc.Volume,
		}
	}

	return prices, nil
}

func (s *DataFetcherService) fetchHistoricalFromYahoo(tradernetSymbol, yahooSymbol string, years int) ([]HistoricalPrice, error) {
	if s.yahooClient == nil {
		return nil, fmt.Errorf("yahoo client not available")
	}

	// Use Yahoo symbol if provided, otherwise convert
	symbol := yahooSymbol
	if symbol == "" {
		var err error
		symbol, err = s.convertToYahooSymbol(tradernetSymbol)
		if err != nil {
			symbol = tradernetSymbol // Fallback to tradernet symbol
		}
	}

	period := "1y"
	if years > 5 {
		period = "max"
	} else if years > 1 {
		period = "5y"
	}

	var yahooSymbolPtr *string
	if yahooSymbol != "" {
		yahooSymbolPtr = &yahooSymbol
	}

	yahooData, err := s.yahooClient.GetHistoricalPrices(symbol, yahooSymbolPtr, period)
	if err != nil {
		return nil, err
	}

	prices := make([]HistoricalPrice, len(yahooData))
	for i, yp := range yahooData {
		prices[i] = HistoricalPrice{
			Date:   yp.Date,
			Open:   yp.Open,
			High:   yp.High,
			Low:    yp.Low,
			Close:  yp.Close,
			Volume: yp.Volume,
		}
	}

	return prices, nil
}

func (s *DataFetcherService) fetchHistoricalFromAlphaVantage(symbol string, full bool) ([]HistoricalPrice, error) {
	if s.alphaVantage == nil {
		return nil, fmt.Errorf("alpha vantage client not available")
	}

	avData, err := s.alphaVantage.GetDailyPrices(symbol, full)
	if err != nil {
		return nil, err
	}

	prices := make([]HistoricalPrice, len(avData))
	for i, ap := range avData {
		prices[i] = HistoricalPrice{
			Date:   ap.Date,
			Open:   ap.Open,
			High:   ap.High,
			Low:    ap.Low,
			Close:  ap.Close,
			Volume: ap.Volume,
		}
	}

	return prices, nil
}

// =============================================================================
// Fundamentals
// =============================================================================

// GetFundamentals fetches fundamental data using configured source priorities.
func (s *DataFetcherService) GetFundamentals(tradernetSymbol string, yahooSymbol string) (*Fundamentals, DataSource, error) {
	s.log.Debug().Str("symbol", tradernetSymbol).Msg("Fetching fundamentals")

	// Create fetcher functions for each source
	fetchers := map[DataSource]func() (interface{}, error){
		SourceAlphaVantage: func() (interface{}, error) {
			avSymbol, err := s.convertToAlphaVantageSymbol(tradernetSymbol)
			if err != nil {
				return nil, fmt.Errorf("symbol conversion failed: %w", err)
			}
			return s.fetchFundamentalsFromAlphaVantage(avSymbol)
		},
		// Note: Yahoo doesn't provide the same level of fundamentals data
		// We could add a Yahoo fallback for basic metrics if needed
	}

	result, source, err := s.router.FetchWithFallback(DataTypeFundamentals, func(src DataSource) (interface{}, error) {
		if fetcher, ok := fetchers[src]; ok {
			return fetcher()
		}
		return nil, fmt.Errorf("unsupported source for fundamentals: %s", src)
	})

	if err != nil {
		return nil, "", err
	}

	fundamentals, ok := result.(*Fundamentals)
	if !ok {
		return nil, "", fmt.Errorf("unexpected result type from fundamentals fetch")
	}

	s.log.Info().
		Str("symbol", tradernetSymbol).
		Str("source", string(source)).
		Msg("Fundamentals fetched")

	return fundamentals, source, nil
}

func (s *DataFetcherService) fetchFundamentalsFromAlphaVantage(symbol string) (*Fundamentals, error) {
	if s.alphaVantage == nil {
		return nil, fmt.Errorf("alpha vantage client not available")
	}

	overview, err := s.alphaVantage.GetCompanyOverview(symbol)
	if err != nil {
		return nil, err
	}

	return &Fundamentals{
		Symbol:                     overview.Symbol,
		PERatio:                    overview.PERatio,
		PEGRatio:                   overview.PEGRatio,
		PriceToBook:                overview.PriceToBookRatio,
		PriceToSales:               overview.PriceToSalesRatioTTM,
		EPS:                        overview.EPS,
		DividendYield:              overview.DividendYield,
		ProfitMargin:               overview.ProfitMargin,
		OperatingMargin:            overview.OperatingMarginTTM,
		ReturnOnAssets:             overview.ReturnOnAssetsTTM,
		ReturnOnEquity:             overview.ReturnOnEquityTTM,
		RevenueGrowthYOY:           overview.QuarterlyRevenueGrowthYOY,
		EarningsGrowthYOY:          overview.QuarterlyEarningsGrowthYOY,
		Beta:                       overview.Beta,
		FiftyTwoWeekHigh:           overview.FiftyTwoWeekHigh,
		FiftyTwoWeekLow:            overview.FiftyTwoWeekLow,
		FiftyDayMovingAverage:      overview.FiftyDayMovingAverage,
		TwoHundredDayMovingAverage: overview.TwoHundredDayMovingAverage,
		AnalystTargetPrice:         overview.AnalystTargetPrice,
		AnalystRatingStrongBuy:     overview.AnalystRatingStrongBuy,
		AnalystRatingBuy:           overview.AnalystRatingBuy,
		AnalystRatingHold:          overview.AnalystRatingHold,
		AnalystRatingSell:          overview.AnalystRatingSell,
		AnalystRatingStrongSell:    overview.AnalystRatingStrongSell,
	}, nil
}

// =============================================================================
// Security Metadata
// =============================================================================

// GetSecurityMetadata fetches company metadata using configured source priorities.
func (s *DataFetcherService) GetSecurityMetadata(tradernetSymbol string, yahooSymbol string) (*SecurityMetadata, DataSource, error) {
	s.log.Debug().Str("symbol", tradernetSymbol).Msg("Fetching security metadata")

	fetchers := map[DataSource]func() (interface{}, error){
		SourceYahoo: func() (interface{}, error) {
			return s.fetchMetadataFromYahoo(tradernetSymbol, yahooSymbol)
		},
		SourceAlphaVantage: func() (interface{}, error) {
			avSymbol, err := s.convertToAlphaVantageSymbol(tradernetSymbol)
			if err != nil {
				return nil, fmt.Errorf("symbol conversion failed: %w", err)
			}
			return s.fetchMetadataFromAlphaVantage(avSymbol)
		},
	}

	result, source, err := s.router.FetchWithFallback(DataTypeMetadata, func(src DataSource) (interface{}, error) {
		if fetcher, ok := fetchers[src]; ok {
			return fetcher()
		}
		return nil, fmt.Errorf("unsupported source for metadata: %s", src)
	})

	if err != nil {
		return nil, "", err
	}

	metadata, ok := result.(*SecurityMetadata)
	if !ok {
		return nil, "", fmt.Errorf("unexpected result type from metadata fetch")
	}

	s.log.Info().
		Str("symbol", tradernetSymbol).
		Str("source", string(source)).
		Msg("Security metadata fetched")

	return metadata, source, nil
}

func (s *DataFetcherService) fetchMetadataFromYahoo(tradernetSymbol, yahooSymbol string) (*SecurityMetadata, error) {
	if s.yahooClient == nil {
		return nil, fmt.Errorf("yahoo client not available")
	}

	var yahooSymbolPtr *string
	if yahooSymbol != "" {
		yahooSymbolPtr = &yahooSymbol
	}

	country, exchange, err := s.yahooClient.GetSecurityCountryAndExchange(tradernetSymbol, yahooSymbolPtr)
	if err != nil {
		return nil, fmt.Errorf("failed to get country/exchange: %w", err)
	}

	industry, err := s.yahooClient.GetSecurityIndustry(tradernetSymbol, yahooSymbolPtr)
	if err != nil {
		s.log.Debug().Err(err).Msg("Failed to get industry from Yahoo")
		// Non-fatal - continue with partial data
	}

	name, err := s.yahooClient.GetQuoteName(tradernetSymbol, yahooSymbolPtr)
	if err != nil {
		s.log.Debug().Err(err).Msg("Failed to get name from Yahoo")
	}

	quoteType, err := s.yahooClient.GetQuoteType(tradernetSymbol, yahooSymbolPtr)
	if err != nil {
		s.log.Debug().Err(err).Msg("Failed to get quote type from Yahoo")
	}

	metadata := &SecurityMetadata{}
	if country != nil {
		metadata.Country = *country
	}
	if exchange != nil {
		metadata.Exchange = *exchange
	}
	if industry != nil {
		metadata.Industry = *industry
	}
	if name != nil {
		metadata.Name = *name
	}
	metadata.ProductType = s.mapQuoteTypeToProductType(quoteType)

	return metadata, nil
}

func (s *DataFetcherService) fetchMetadataFromAlphaVantage(symbol string) (*SecurityMetadata, error) {
	if s.alphaVantage == nil {
		return nil, fmt.Errorf("alpha vantage client not available")
	}

	overview, err := s.alphaVantage.GetCompanyOverview(symbol)
	if err != nil {
		return nil, err
	}

	return &SecurityMetadata{
		Name:              overview.Name,
		Country:           overview.Country,
		Exchange:          overview.Exchange,
		Industry:          overview.Industry,
		Sector:            overview.Sector,
		Currency:          overview.Currency,
		ProductType:       overview.AssetType,
		MarketCap:         overview.MarketCapitalization,
		FullTimeEmployees: overview.FullTimeEmployees,
	}, nil
}

func (s *DataFetcherService) mapQuoteTypeToProductType(quoteType string) string {
	switch quoteType {
	case "EQUITY":
		return "EQUITY"
	case "ETF":
		return "ETF"
	case "MUTUALFUND":
		return "MUTUAL_FUND"
	case "CURRENCY":
		return "CURRENCY"
	case "CRYPTOCURRENCY":
		return "CRYPTO"
	case "FUTURE":
		return "FUTURE"
	case "INDEX":
		return "INDEX"
	default:
		return "UNKNOWN"
	}
}

// =============================================================================
// Current Prices
// =============================================================================

// GetCurrentPrice fetches the current price for a symbol.
func (s *DataFetcherService) GetCurrentPrice(tradernetSymbol string, yahooSymbol string) (*CurrentQuote, DataSource, error) {
	s.log.Debug().Str("symbol", tradernetSymbol).Msg("Fetching current price")

	fetchers := map[DataSource]func() (interface{}, error){
		SourceTradernet: func() (interface{}, error) {
			return s.fetchCurrentPriceFromTradernet(tradernetSymbol)
		},
		SourceYahoo: func() (interface{}, error) {
			return s.fetchCurrentPriceFromYahoo(tradernetSymbol, yahooSymbol)
		},
		SourceAlphaVantage: func() (interface{}, error) {
			avSymbol, err := s.convertToAlphaVantageSymbol(tradernetSymbol)
			if err != nil {
				return nil, err
			}
			return s.fetchCurrentPriceFromAlphaVantage(avSymbol)
		},
	}

	result, source, err := s.router.FetchWithFallback(DataTypePrices, func(src DataSource) (interface{}, error) {
		if fetcher, ok := fetchers[src]; ok {
			return fetcher()
		}
		return nil, fmt.Errorf("unsupported source for current price: %s", src)
	})

	if err != nil {
		return nil, "", err
	}

	quote, ok := result.(*CurrentQuote)
	if !ok {
		return nil, "", fmt.Errorf("unexpected result type from price fetch")
	}

	return quote, source, nil
}

func (s *DataFetcherService) fetchCurrentPriceFromTradernet(symbol string) (*CurrentQuote, error) {
	if s.brokerClient == nil || !s.brokerClient.IsConnected() {
		return nil, fmt.Errorf("tradernet client not available")
	}

	quotes, err := s.brokerClient.GetQuotes([]string{symbol})
	if err != nil {
		return nil, err
	}

	quote, ok := quotes[symbol]
	if !ok || quote == nil {
		return nil, fmt.Errorf("no quote returned for %s", symbol)
	}

	return &CurrentQuote{
		Symbol:    symbol,
		Price:     quote.Price,
		Change:    quote.Change,
		ChangePct: quote.ChangePct,
		Volume:    quote.Volume,
	}, nil
}

func (s *DataFetcherService) fetchCurrentPriceFromYahoo(tradernetSymbol, yahooSymbol string) (*CurrentQuote, error) {
	if s.yahooClient == nil {
		return nil, fmt.Errorf("yahoo client not available")
	}

	var yahooPtr *string
	if yahooSymbol != "" {
		yahooPtr = &yahooSymbol
	}

	symbolMap := map[string]*string{tradernetSymbol: yahooPtr}
	prices, err := s.yahooClient.GetBatchQuotes(symbolMap)
	if err != nil {
		return nil, err
	}

	price, ok := prices[tradernetSymbol]
	if !ok || price == nil {
		return nil, fmt.Errorf("no price returned for %s", tradernetSymbol)
	}

	return &CurrentQuote{
		Symbol: tradernetSymbol,
		Price:  *price,
	}, nil
}

func (s *DataFetcherService) fetchCurrentPriceFromAlphaVantage(symbol string) (*CurrentQuote, error) {
	if s.alphaVantage == nil {
		return nil, fmt.Errorf("alpha vantage client not available")
	}

	quote, err := s.alphaVantage.GetGlobalQuote(symbol)
	if err != nil {
		return nil, err
	}

	return &CurrentQuote{
		Symbol:    symbol,
		Price:     quote.Price,
		Change:    quote.Change,
		ChangePct: quote.ChangePercent,
		Volume:    quote.Volume,
		Timestamp: quote.LatestTradingDay,
	}, nil
}

// =============================================================================
// Exchange Rates
// =============================================================================

// GetExchangeRate fetches the exchange rate between two currencies.
func (s *DataFetcherService) GetExchangeRate(fromCurrency, toCurrency string) (float64, DataSource, error) {
	s.log.Debug().
		Str("from", fromCurrency).
		Str("to", toCurrency).
		Msg("Fetching exchange rate")

	fetchers := map[DataSource]func() (interface{}, error){
		SourceAlphaVantage: func() (interface{}, error) {
			if s.alphaVantage == nil {
				return nil, fmt.Errorf("alpha vantage client not available")
			}
			rate, err := s.alphaVantage.GetExchangeRate(fromCurrency, toCurrency)
			if err != nil {
				return nil, err
			}
			return rate.ExchangeRate, nil
		},
		// Other exchange rate sources can be added here
	}

	result, source, err := s.router.FetchWithFallback(DataTypeExchangeRates, func(src DataSource) (interface{}, error) {
		if fetcher, ok := fetchers[src]; ok {
			return fetcher()
		}
		return nil, fmt.Errorf("unsupported source for exchange rates: %s", src)
	})

	if err != nil {
		return 0, "", err
	}

	rate, ok := result.(float64)
	if !ok {
		return 0, "", fmt.Errorf("unexpected result type from exchange rate fetch")
	}

	return rate, source, nil
}

// =============================================================================
// Symbol Conversion Helpers
// =============================================================================

func (s *DataFetcherService) convertToYahooSymbol(tradernetSymbol string) (string, error) {
	if s.symbolMapper == nil {
		return tradernetSymbol, nil // No conversion available
	}
	return s.symbolMapper.ToYahoo(tradernetSymbol)
}

func (s *DataFetcherService) convertToAlphaVantageSymbol(tradernetSymbol string) (string, error) {
	if s.symbolMapper == nil {
		return tradernetSymbol, nil // No conversion available
	}
	return s.symbolMapper.ToAlphaVantage(tradernetSymbol)
}
