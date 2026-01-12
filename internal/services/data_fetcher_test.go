package services

import (
	"errors"
	"testing"
	"time"

	"github.com/aristath/sentinel/internal/clients/alphavantage"
	"github.com/aristath/sentinel/internal/clients/symbols"
	"github.com/aristath/sentinel/internal/clients/yahoo"
	"github.com/aristath/sentinel/internal/domain"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// Mock Clients
// =============================================================================

// MockBrokerClient mocks the broker client interface
type MockBrokerClient struct {
	mock.Mock
}

func (m *MockBrokerClient) GetHistoricalPrices(symbol string, from, to int64, interval int) ([]domain.BrokerOHLCV, error) {
	args := m.Called(symbol, from, to, interval)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.BrokerOHLCV), args.Error(1)
}

func (m *MockBrokerClient) GetQuotes(symbols []string) (map[string]*domain.BrokerQuote, error) {
	args := m.Called(symbols)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*domain.BrokerQuote), args.Error(1)
}

func (m *MockBrokerClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

// MockYahooClientDF mocks the DataFetcherYahooClient interface
type MockYahooClientDF struct {
	mock.Mock
}

func (m *MockYahooClientDF) GetHistoricalPrices(symbol string, yahooSymbolOverride *string, period string) ([]yahoo.HistoricalPrice, error) {
	args := m.Called(symbol, yahooSymbolOverride, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]yahoo.HistoricalPrice), args.Error(1)
}

func (m *MockYahooClientDF) GetSecurityCountryAndExchange(symbol string, yahooSymbolOverride *string) (*string, *string, error) {
	args := m.Called(symbol, yahooSymbolOverride)
	var country, exchange *string
	if args.Get(0) != nil {
		country = args.Get(0).(*string)
	}
	if args.Get(1) != nil {
		exchange = args.Get(1).(*string)
	}
	return country, exchange, args.Error(2)
}

func (m *MockYahooClientDF) GetSecurityIndustry(symbol string, yahooSymbolOverride *string) (*string, error) {
	args := m.Called(symbol, yahooSymbolOverride)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockYahooClientDF) GetQuoteName(symbol string, yahooSymbolOverride *string) (*string, error) {
	args := m.Called(symbol, yahooSymbolOverride)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockYahooClientDF) GetQuoteType(symbol string, yahooSymbolOverride *string) (string, error) {
	args := m.Called(symbol, yahooSymbolOverride)
	return args.String(0), args.Error(1)
}

func (m *MockYahooClientDF) GetBatchQuotes(symbolMap map[string]*string) (map[string]*float64, error) {
	args := m.Called(symbolMap)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*float64), args.Error(1)
}

// MockAlphaVantageClient mocks the Alpha Vantage client interface
type MockAlphaVantageClient struct {
	mock.Mock
}

func (m *MockAlphaVantageClient) GetDailyPrices(symbol string, full bool) ([]alphavantage.DailyPrice, error) {
	args := m.Called(symbol, full)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]alphavantage.DailyPrice), args.Error(1)
}

func (m *MockAlphaVantageClient) GetCompanyOverview(symbol string) (*alphavantage.CompanyOverview, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alphavantage.CompanyOverview), args.Error(1)
}

func (m *MockAlphaVantageClient) GetGlobalQuote(symbol string) (*alphavantage.GlobalQuote, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alphavantage.GlobalQuote), args.Error(1)
}

func (m *MockAlphaVantageClient) GetExchangeRate(fromCurrency, toCurrency string) (*alphavantage.ExchangeRate, error) {
	args := m.Called(fromCurrency, toCurrency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*alphavantage.ExchangeRate), args.Error(1)
}

// =============================================================================
// Mock Settings Getter
// =============================================================================

// mockSettingsGetter implements SettingsGetter for testing
type mockSettingsGetter struct {
	priorities map[DataType][]DataSource
}

func (m *mockSettingsGetter) Get(key string) (*string, error) {
	// Return nil to use defaults, unless we have specific priorities
	return nil, nil
}

// =============================================================================
// Test Helpers
// =============================================================================

func createTestDataFetcher(broker *MockBrokerClient, yahooMock *MockYahooClientDF, av *MockAlphaVantageClient, priorities map[DataType][]DataSource) *DataFetcherService {
	log := zerolog.New(nil).Level(zerolog.Disabled)

	// Create mock settings
	mockSettings := &mockSettingsGetter{priorities: priorities}

	// Create router with mock settings
	router := NewDataSourceRouter(mockSettings, &DataSourceClients{
		AlphaVantageAPIKey: "test-key",
	})

	// Override priorities if provided
	if priorities != nil {
		for dataType, sources := range priorities {
			router.priorities[dataType] = sources
		}
	}

	// Create symbol mapper
	mapper := symbols.NewMapper()

	return NewDataFetcherService(router, mapper, broker, yahooMock, av, log)
}

// =============================================================================
// Tests: Historical Prices
// =============================================================================

func TestGetHistoricalPrices_TradernetPrimary(t *testing.T) {
	broker := new(MockBrokerClient)
	yahooMock := new(MockYahooClientDF)
	av := new(MockAlphaVantageClient)

	// Configure Tradernet as primary
	priorities := map[DataType][]DataSource{
		DataTypeHistorical: {SourceTradernet, SourceYahoo},
	}

	fetcher := createTestDataFetcher(broker, yahooMock, av, priorities)

	// Setup mock - Tradernet returns data
	now := time.Now()
	broker.On("IsConnected").Return(true)
	broker.On("GetHistoricalPrices", "AAPL.US", mock.Anything, mock.Anything, 86400).Return([]domain.BrokerOHLCV{
		{Timestamp: now.AddDate(0, 0, -1).Unix(), Open: 150.0, High: 155.0, Low: 149.0, Close: 153.0, Volume: 1000000},
		{Timestamp: now.Unix(), Open: 153.0, High: 158.0, Low: 152.0, Close: 157.0, Volume: 1200000},
	}, nil)

	// Execute
	prices, source, err := fetcher.GetHistoricalPrices("AAPL.US", "", 1)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, SourceTradernet, source)
	assert.Len(t, prices, 2)
	assert.Equal(t, 153.0, prices[0].Close)
	assert.Equal(t, 157.0, prices[1].Close)
	broker.AssertExpectations(t)
}

func TestGetHistoricalPrices_FallbackToYahoo(t *testing.T) {
	broker := new(MockBrokerClient)
	yahooMock := new(MockYahooClientDF)
	av := new(MockAlphaVantageClient)

	priorities := map[DataType][]DataSource{
		DataTypeHistorical: {SourceTradernet, SourceYahoo},
	}

	fetcher := createTestDataFetcher(broker, yahooMock, av, priorities)

	// Tradernet fails
	broker.On("IsConnected").Return(false)

	// Yahoo succeeds
	now := time.Now()
	yahooMock.On("GetHistoricalPrices", mock.Anything, mock.Anything, "1y").Return([]yahoo.HistoricalPrice{
		{Date: now.AddDate(0, 0, -1), Open: 150.0, High: 155.0, Low: 149.0, Close: 153.0, Volume: 1000000},
	}, nil)

	// Execute
	prices, source, err := fetcher.GetHistoricalPrices("AAPL.US", "AAPL", 1)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, SourceYahoo, source)
	assert.Len(t, prices, 1)
	yahooMock.AssertExpectations(t)
}

func TestGetHistoricalPrices_AllSourcesFail(t *testing.T) {
	broker := new(MockBrokerClient)
	yahooMock := new(MockYahooClientDF)
	av := new(MockAlphaVantageClient)

	priorities := map[DataType][]DataSource{
		DataTypeHistorical: {SourceTradernet, SourceYahoo},
	}

	fetcher := createTestDataFetcher(broker, yahooMock, av, priorities)

	// Both fail
	broker.On("IsConnected").Return(false)
	yahooMock.On("GetHistoricalPrices", mock.Anything, mock.Anything, "1y").Return(nil, errors.New("yahoo error"))

	// Execute
	_, _, err := fetcher.GetHistoricalPrices("AAPL.US", "AAPL", 1)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all data sources failed")
}

// =============================================================================
// Tests: Fundamentals
// =============================================================================

func TestGetFundamentals_AlphaVantageSuccess(t *testing.T) {
	broker := new(MockBrokerClient)
	yahooMock := new(MockYahooClientDF)
	av := new(MockAlphaVantageClient)

	priorities := map[DataType][]DataSource{
		DataTypeFundamentals: {SourceAlphaVantage},
	}

	fetcher := createTestDataFetcher(broker, yahooMock, av, priorities)

	// Setup mock
	peRatio := 25.5
	av.On("GetCompanyOverview", "AAPL").Return(&alphavantage.CompanyOverview{
		Symbol:  "AAPL",
		Name:    "Apple Inc",
		PERatio: &peRatio,
		Country: "USA",
		Sector:  "Technology",
	}, nil)

	// Execute
	fundamentals, source, err := fetcher.GetFundamentals("AAPL.US", "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, SourceAlphaVantage, source)
	assert.NotNil(t, fundamentals)
	assert.Equal(t, "AAPL", fundamentals.Symbol)
	assert.NotNil(t, fundamentals.PERatio)
	assert.Equal(t, 25.5, *fundamentals.PERatio)
	av.AssertExpectations(t)
}

// =============================================================================
// Tests: Security Metadata
// =============================================================================

func TestGetSecurityMetadata_YahooPrimary(t *testing.T) {
	broker := new(MockBrokerClient)
	yahooMock := new(MockYahooClientDF)
	av := new(MockAlphaVantageClient)

	priorities := map[DataType][]DataSource{
		DataTypeMetadata: {SourceYahoo, SourceAlphaVantage},
	}

	fetcher := createTestDataFetcher(broker, yahooMock, av, priorities)

	// Setup mocks
	country := "United States"
	exchange := "NASDAQ"
	industry := "Consumer Electronics"
	name := "Apple Inc."

	yahooMock.On("GetSecurityCountryAndExchange", "AAPL.US", mock.Anything).Return(&country, &exchange, nil)
	yahooMock.On("GetSecurityIndustry", "AAPL.US", mock.Anything).Return(&industry, nil)
	yahooMock.On("GetQuoteName", "AAPL.US", mock.Anything).Return(&name, nil)
	yahooMock.On("GetQuoteType", "AAPL.US", mock.Anything).Return("EQUITY", nil)

	// Execute
	metadata, source, err := fetcher.GetSecurityMetadata("AAPL.US", "AAPL")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, SourceYahoo, source)
	assert.NotNil(t, metadata)
	assert.Equal(t, "United States", metadata.Country)
	assert.Equal(t, "NASDAQ", metadata.Exchange)
	assert.Equal(t, "Consumer Electronics", metadata.Industry)
	assert.Equal(t, "Apple Inc.", metadata.Name)
	assert.Equal(t, "EQUITY", metadata.ProductType)
	yahooMock.AssertExpectations(t)
}

func TestGetSecurityMetadata_FallbackToAlphaVantage(t *testing.T) {
	broker := new(MockBrokerClient)
	yahooMock := new(MockYahooClientDF)
	av := new(MockAlphaVantageClient)

	priorities := map[DataType][]DataSource{
		DataTypeMetadata: {SourceYahoo, SourceAlphaVantage},
	}

	fetcher := createTestDataFetcher(broker, yahooMock, av, priorities)

	// Yahoo fails
	yahooMock.On("GetSecurityCountryAndExchange", "AAPL.US", mock.Anything).Return(nil, nil, errors.New("yahoo error"))

	// Alpha Vantage succeeds
	av.On("GetCompanyOverview", "AAPL").Return(&alphavantage.CompanyOverview{
		Symbol:   "AAPL",
		Name:     "Apple Inc",
		Country:  "USA",
		Exchange: "NASDAQ",
		Industry: "Consumer Electronics",
		Sector:   "Technology",
	}, nil)

	// Execute
	metadata, source, err := fetcher.GetSecurityMetadata("AAPL.US", "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, SourceAlphaVantage, source)
	assert.NotNil(t, metadata)
	assert.Equal(t, "USA", metadata.Country)
	av.AssertExpectations(t)
}

// =============================================================================
// Tests: Current Price
// =============================================================================

func TestGetCurrentPrice_TradernetPrimary(t *testing.T) {
	broker := new(MockBrokerClient)
	yahooMock := new(MockYahooClientDF)
	av := new(MockAlphaVantageClient)

	priorities := map[DataType][]DataSource{
		DataTypePrices: {SourceTradernet, SourceYahoo},
	}

	fetcher := createTestDataFetcher(broker, yahooMock, av, priorities)

	// Setup mock
	broker.On("IsConnected").Return(true)
	broker.On("GetQuotes", []string{"AAPL.US"}).Return(map[string]*domain.BrokerQuote{
		"AAPL.US": {Symbol: "AAPL.US", Price: 157.50, Change: 2.50, ChangePct: 1.6, Volume: 1500000},
	}, nil)

	// Execute
	quote, source, err := fetcher.GetCurrentPrice("AAPL.US", "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, SourceTradernet, source)
	assert.NotNil(t, quote)
	assert.Equal(t, 157.50, quote.Price)
	assert.Equal(t, 2.50, quote.Change)
	broker.AssertExpectations(t)
}

// =============================================================================
// Tests: Exchange Rate
// =============================================================================

func TestGetExchangeRate_AlphaVantage(t *testing.T) {
	broker := new(MockBrokerClient)
	yahooMock := new(MockYahooClientDF)
	av := new(MockAlphaVantageClient)

	priorities := map[DataType][]DataSource{
		DataTypeExchangeRates: {SourceAlphaVantage},
	}

	fetcher := createTestDataFetcher(broker, yahooMock, av, priorities)

	// Setup mock
	av.On("GetExchangeRate", "USD", "EUR").Return(&alphavantage.ExchangeRate{
		FromCurrency: "USD",
		ToCurrency:   "EUR",
		ExchangeRate: 0.92,
	}, nil)

	// Execute
	rate, source, err := fetcher.GetExchangeRate("USD", "EUR")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, SourceAlphaVantage, source)
	assert.Equal(t, 0.92, rate)
	av.AssertExpectations(t)
}

// =============================================================================
// Tests: Symbol Conversion
// =============================================================================

func TestSymbolConversion_USStock(t *testing.T) {
	broker := new(MockBrokerClient)
	yahooMock := new(MockYahooClientDF)
	av := new(MockAlphaVantageClient)

	priorities := map[DataType][]DataSource{
		DataTypeFundamentals: {SourceAlphaVantage},
	}

	fetcher := createTestDataFetcher(broker, yahooMock, av, priorities)

	// Test that AAPL.US gets converted to AAPL for Alpha Vantage
	peRatio := 25.5
	av.On("GetCompanyOverview", "AAPL").Return(&alphavantage.CompanyOverview{
		Symbol:  "AAPL",
		PERatio: &peRatio,
	}, nil)

	// Execute
	_, source, err := fetcher.GetFundamentals("AAPL.US", "")

	// Assert - verify Alpha Vantage was called with converted symbol
	assert.NoError(t, err)
	assert.Equal(t, SourceAlphaVantage, source)
	av.AssertExpectations(t)
}
