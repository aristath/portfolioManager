package alphavantage

// ClientInterface defines all methods for the Alpha Vantage API client.
// This interface covers all free-tier endpoints organized by category.
type ClientInterface interface {
	// =========================================================================
	// Time Series
	// =========================================================================

	// GetDailyPrices returns daily OHLCV data for a symbol.
	// If full is true, returns the complete historical time series (20+ years).
	// If full is false, returns the latest 100 data points.
	GetDailyPrices(symbol string, full bool) ([]DailyPrice, error)

	// GetDailyAdjustedPrices returns daily adjusted OHLCV data including dividends and splits.
	// If full is true, returns the complete historical time series.
	GetDailyAdjustedPrices(symbol string, full bool) ([]AdjustedPrice, error)

	// GetWeeklyPrices returns weekly OHLCV data for a symbol.
	GetWeeklyPrices(symbol string) ([]DailyPrice, error)

	// GetWeeklyAdjustedPrices returns weekly adjusted OHLCV data.
	GetWeeklyAdjustedPrices(symbol string) ([]AdjustedPrice, error)

	// GetMonthlyPrices returns monthly OHLCV data for a symbol.
	GetMonthlyPrices(symbol string) ([]DailyPrice, error)

	// GetMonthlyAdjustedPrices returns monthly adjusted OHLCV data.
	GetMonthlyAdjustedPrices(symbol string) ([]AdjustedPrice, error)

	// GetGlobalQuote returns the latest price and volume information for a symbol.
	GetGlobalQuote(symbol string) (*GlobalQuote, error)

	// SearchSymbol searches for symbols matching the given keywords.
	SearchSymbol(keywords string) ([]SymbolMatch, error)

	// =========================================================================
	// Fundamental Data
	// =========================================================================

	// GetCompanyOverview returns comprehensive company information and financial metrics.
	GetCompanyOverview(symbol string) (*CompanyOverview, error)

	// GetEarnings returns annual and quarterly earnings data.
	GetEarnings(symbol string) (*Earnings, error)

	// GetIncomeStatement returns annual and quarterly income statements.
	GetIncomeStatement(symbol string) (*IncomeStatement, error)

	// GetBalanceSheet returns annual and quarterly balance sheets.
	GetBalanceSheet(symbol string) (*BalanceSheet, error)

	// GetCashFlow returns annual and quarterly cash flow statements.
	GetCashFlow(symbol string) (*CashFlow, error)

	// GetDividends returns historical dividend data.
	GetDividends(symbol string) ([]DividendRecord, error)

	// GetSplits returns historical stock split data.
	GetSplits(symbol string) ([]SplitRecord, error)

	// GetETFProfile returns ETF-specific information.
	GetETFProfile(symbol string) (*ETFProfile, error)

	// GetETFHoldings returns the holdings of an ETF.
	GetETFHoldings(symbol string) ([]ETFHolding, error)

	// GetSharesOutstanding returns historical shares outstanding data.
	GetSharesOutstanding(symbol string) ([]SharesOutstandingRecord, error)

	// GetListingStatus returns listing status for securities.
	// status can be "active", "delisted", or empty for all.
	GetListingStatus(status string) ([]ListingStatus, error)

	// GetEarningsCalendar returns upcoming earnings announcements.
	// horizon can be "3month", "6month", or "12month".
	GetEarningsCalendar(horizon string) ([]EarningsEvent, error)

	// GetIPOCalendar returns upcoming IPO events.
	GetIPOCalendar() ([]IPOEvent, error)

	// =========================================================================
	// Technical Indicators - Generic
	// =========================================================================

	// GetIndicator fetches any technical indicator with custom parameters.
	// function is the indicator name (e.g., "SMA", "RSI", "MACD").
	// interval can be "1min", "5min", "15min", "30min", "60min", "daily", "weekly", "monthly".
	// params contains additional parameters specific to the indicator.
	GetIndicator(function, symbol, interval string, params map[string]string) (*IndicatorData, error)

	// =========================================================================
	// Technical Indicators - Moving Averages
	// =========================================================================

	// GetSMA returns Simple Moving Average values.
	GetSMA(symbol, interval string, period int, seriesType string) (*IndicatorData, error)

	// GetEMA returns Exponential Moving Average values.
	GetEMA(symbol, interval string, period int, seriesType string) (*IndicatorData, error)

	// GetWMA returns Weighted Moving Average values.
	GetWMA(symbol, interval string, period int, seriesType string) (*IndicatorData, error)

	// GetDEMA returns Double Exponential Moving Average values.
	GetDEMA(symbol, interval string, period int, seriesType string) (*IndicatorData, error)

	// GetTEMA returns Triple Exponential Moving Average values.
	GetTEMA(symbol, interval string, period int, seriesType string) (*IndicatorData, error)

	// GetTRIMA returns Triangular Moving Average values.
	GetTRIMA(symbol, interval string, period int, seriesType string) (*IndicatorData, error)

	// GetKAMA returns Kaufman Adaptive Moving Average values.
	GetKAMA(symbol, interval string, period int, seriesType string) (*IndicatorData, error)

	// GetT3 returns Triple Exponential Moving Average (T3) values.
	GetT3(symbol, interval string, period int, seriesType string) (*IndicatorData, error)

	// GetVWAP returns Volume Weighted Average Price values.
	GetVWAP(symbol, interval string) (*IndicatorData, error)

	// =========================================================================
	// Technical Indicators - Momentum
	// =========================================================================

	// GetRSI returns Relative Strength Index values.
	GetRSI(symbol, interval string, period int, seriesType string) (*IndicatorData, error)

	// GetMACD returns Moving Average Convergence Divergence values.
	GetMACD(symbol, interval string, seriesType string) (*MACDData, error)

	// GetMACDEXT returns MACD with controllable MA types.
	GetMACDEXT(symbol, interval string, seriesType string, params map[string]string) (*MACDData, error)

	// GetSTOCH returns Stochastic Oscillator values.
	GetSTOCH(symbol, interval string) (*StochData, error)

	// GetSTOCHF returns Fast Stochastic Oscillator values.
	GetSTOCHF(symbol, interval string) (*StochData, error)

	// GetSTOCHRSI returns Stochastic RSI values.
	GetSTOCHRSI(symbol, interval string, period int, seriesType string) (*StochData, error)

	// GetWILLR returns Williams' %R values.
	GetWILLR(symbol, interval string, period int) (*IndicatorData, error)

	// GetROC returns Rate of Change values.
	GetROC(symbol, interval string, period int, seriesType string) (*IndicatorData, error)

	// GetROCR returns Rate of Change Ratio values.
	GetROCR(symbol, interval string, period int, seriesType string) (*IndicatorData, error)

	// GetMOM returns Momentum values.
	GetMOM(symbol, interval string, period int, seriesType string) (*IndicatorData, error)

	// GetMFI returns Money Flow Index values.
	GetMFI(symbol, interval string, period int) (*IndicatorData, error)

	// GetCMO returns Chande Momentum Oscillator values.
	GetCMO(symbol, interval string, period int, seriesType string) (*IndicatorData, error)

	// GetULTOSC returns Ultimate Oscillator values.
	GetULTOSC(symbol, interval string) (*IndicatorData, error)

	// GetTRIX returns TRIX values.
	GetTRIX(symbol, interval string, period int, seriesType string) (*IndicatorData, error)

	// GetAPO returns Absolute Price Oscillator values.
	GetAPO(symbol, interval string, seriesType string, fastPeriod, slowPeriod int) (*IndicatorData, error)

	// GetPPO returns Percentage Price Oscillator values.
	GetPPO(symbol, interval string, seriesType string, fastPeriod, slowPeriod int) (*IndicatorData, error)

	// GetBOP returns Balance of Power values.
	GetBOP(symbol, interval string) (*IndicatorData, error)

	// =========================================================================
	// Technical Indicators - Trend
	// =========================================================================

	// GetADX returns Average Directional Movement Index values.
	GetADX(symbol, interval string, period int) (*IndicatorData, error)

	// GetADXR returns Average Directional Movement Index Rating values.
	GetADXR(symbol, interval string, period int) (*IndicatorData, error)

	// GetDX returns Directional Movement Index values.
	GetDX(symbol, interval string, period int) (*IndicatorData, error)

	// GetPLUS_DI returns Plus Directional Indicator values.
	GetPLUS_DI(symbol, interval string, period int) (*IndicatorData, error)

	// GetMINUS_DI returns Minus Directional Indicator values.
	GetMINUS_DI(symbol, interval string, period int) (*IndicatorData, error)

	// GetPLUS_DM returns Plus Directional Movement values.
	GetPLUS_DM(symbol, interval string, period int) (*IndicatorData, error)

	// GetMINUS_DM returns Minus Directional Movement values.
	GetMINUS_DM(symbol, interval string, period int) (*IndicatorData, error)

	// GetAROON returns Aroon indicator values.
	GetAROON(symbol, interval string, period int) (*AroonData, error)

	// GetAROONOSC returns Aroon Oscillator values.
	GetAROONOSC(symbol, interval string, period int) (*IndicatorData, error)

	// GetSAR returns Parabolic SAR values.
	GetSAR(symbol, interval string, acceleration, maximum float64) (*IndicatorData, error)

	// =========================================================================
	// Technical Indicators - Volatility
	// =========================================================================

	// GetBBANDS returns Bollinger Bands values.
	GetBBANDS(symbol, interval string, period int, seriesType string, nbdevup, nbdevdn float64) (*BollingerData, error)

	// GetATR returns Average True Range values.
	GetATR(symbol, interval string, period int) (*IndicatorData, error)

	// GetNATR returns Normalized Average True Range values.
	GetNATR(symbol, interval string, period int) (*IndicatorData, error)

	// GetTRANGE returns True Range values.
	GetTRANGE(symbol, interval string) (*IndicatorData, error)

	// GetMIDPOINT returns Midpoint values.
	GetMIDPOINT(symbol, interval string, period int, seriesType string) (*IndicatorData, error)

	// GetMIDPRICE returns Midpoint Price values.
	GetMIDPRICE(symbol, interval string, period int) (*IndicatorData, error)

	// =========================================================================
	// Technical Indicators - Volume
	// =========================================================================

	// GetOBV returns On Balance Volume values.
	GetOBV(symbol, interval string) (*IndicatorData, error)

	// GetAD returns Chaikin A/D Line values.
	GetAD(symbol, interval string) (*IndicatorData, error)

	// GetADOSC returns Chaikin A/D Oscillator values.
	GetADOSC(symbol, interval string, fastPeriod, slowPeriod int) (*IndicatorData, error)

	// =========================================================================
	// Technical Indicators - Cycle (Hilbert Transform)
	// =========================================================================

	// GetHT_TRENDLINE returns Hilbert Transform - Instantaneous Trendline values.
	GetHT_TRENDLINE(symbol, interval string, seriesType string) (*IndicatorData, error)

	// GetHT_SINE returns Hilbert Transform - Sine Wave values.
	GetHT_SINE(symbol, interval string, seriesType string) (*IndicatorData, error)

	// GetHT_TRENDMODE returns Hilbert Transform - Trend vs Cycle Mode values.
	GetHT_TRENDMODE(symbol, interval string, seriesType string) (*IndicatorData, error)

	// GetHT_DCPERIOD returns Hilbert Transform - Dominant Cycle Period values.
	GetHT_DCPERIOD(symbol, interval string, seriesType string) (*IndicatorData, error)

	// GetHT_DCPHASE returns Hilbert Transform - Dominant Cycle Phase values.
	GetHT_DCPHASE(symbol, interval string, seriesType string) (*IndicatorData, error)

	// GetHT_PHASOR returns Hilbert Transform - Phasor Components values.
	GetHT_PHASOR(symbol, interval string, seriesType string) (*IndicatorData, error)

	// =========================================================================
	// Technical Indicators - Other
	// =========================================================================

	// GetCCI returns Commodity Channel Index values.
	GetCCI(symbol, interval string, period int) (*IndicatorData, error)

	// =========================================================================
	// Forex
	// =========================================================================

	// GetExchangeRate returns real-time exchange rate between two currencies.
	GetExchangeRate(fromCurrency, toCurrency string) (*ExchangeRate, error)

	// GetFXDaily returns daily forex data for a currency pair.
	GetFXDaily(fromCurrency, toCurrency string, full bool) ([]FXPrice, error)

	// GetFXWeekly returns weekly forex data for a currency pair.
	GetFXWeekly(fromCurrency, toCurrency string) ([]FXPrice, error)

	// GetFXMonthly returns monthly forex data for a currency pair.
	GetFXMonthly(fromCurrency, toCurrency string) ([]FXPrice, error)

	// =========================================================================
	// Cryptocurrency
	// =========================================================================

	// GetCryptoExchangeRate returns real-time exchange rate for a cryptocurrency.
	GetCryptoExchangeRate(fromCurrency, toCurrency string) (*ExchangeRate, error)

	// GetCryptoDaily returns daily cryptocurrency data.
	// symbol is the cryptocurrency (e.g., "BTC"), market is the fiat currency (e.g., "USD").
	GetCryptoDaily(symbol, market string) ([]CryptoPrice, error)

	// GetCryptoWeekly returns weekly cryptocurrency data.
	GetCryptoWeekly(symbol, market string) ([]CryptoPrice, error)

	// GetCryptoMonthly returns monthly cryptocurrency data.
	GetCryptoMonthly(symbol, market string) ([]CryptoPrice, error)

	// =========================================================================
	// Commodities
	// =========================================================================

	// GetCommodity returns data for a specific commodity.
	// commodity can be: "WTI", "BRENT", "NATURAL_GAS", "COPPER", "ALUMINUM",
	// "WHEAT", "CORN", "COTTON", "SUGAR", "COFFEE", "ALL_COMMODITIES".
	// interval can be "daily", "weekly", or "monthly".
	GetCommodity(commodity, interval string) ([]CommodityPrice, error)

	// GetWTI returns West Texas Intermediate crude oil prices.
	GetWTI(interval string) ([]CommodityPrice, error)

	// GetBrent returns Brent crude oil prices.
	GetBrent(interval string) ([]CommodityPrice, error)

	// GetNaturalGas returns natural gas prices.
	GetNaturalGas(interval string) ([]CommodityPrice, error)

	// GetCopper returns copper prices.
	GetCopper(interval string) ([]CommodityPrice, error)

	// GetAluminum returns aluminum prices.
	GetAluminum(interval string) ([]CommodityPrice, error)

	// GetWheat returns wheat prices.
	GetWheat(interval string) ([]CommodityPrice, error)

	// GetCorn returns corn prices.
	GetCorn(interval string) ([]CommodityPrice, error)

	// GetCotton returns cotton prices.
	GetCotton(interval string) ([]CommodityPrice, error)

	// GetSugar returns sugar prices.
	GetSugar(interval string) ([]CommodityPrice, error)

	// GetCoffee returns coffee prices.
	GetCoffee(interval string) ([]CommodityPrice, error)

	// GetAllCommodities returns global commodities index data.
	GetAllCommodities(interval string) ([]CommodityPrice, error)

	// =========================================================================
	// Economic Indicators
	// =========================================================================

	// GetEconomicIndicator returns data for a specific economic indicator.
	// indicator can be: "REAL_GDP", "REAL_GDP_PER_CAPITA", "TREASURY_YIELD",
	// "FEDERAL_FUNDS_RATE", "CPI", "INFLATION", "RETAIL_SALES", "DURABLES",
	// "UNEMPLOYMENT", "NONFARM_PAYROLL".
	GetEconomicIndicator(indicator, interval string) (*EconomicIndicator, error)

	// GetRealGDP returns real GDP data.
	GetRealGDP(interval string) (*EconomicIndicator, error)

	// GetRealGDPPerCapita returns real GDP per capita data.
	GetRealGDPPerCapita() (*EconomicIndicator, error)

	// GetTreasuryYield returns treasury yield data.
	// maturity can be "3month", "2year", "5year", "7year", "10year", "30year".
	GetTreasuryYield(interval, maturity string) (*EconomicIndicator, error)

	// GetFederalFundsRate returns federal funds rate data.
	GetFederalFundsRate(interval string) (*EconomicIndicator, error)

	// GetCPI returns Consumer Price Index data.
	GetCPI(interval string) (*EconomicIndicator, error)

	// GetInflation returns inflation rate data.
	GetInflation() (*EconomicIndicator, error)

	// GetRetailSales returns retail sales data.
	GetRetailSales() (*EconomicIndicator, error)

	// GetDurableGoodsOrders returns durable goods orders data.
	GetDurableGoodsOrders() (*EconomicIndicator, error)

	// GetUnemployment returns unemployment rate data.
	GetUnemployment() (*EconomicIndicator, error)

	// GetNonfarmPayroll returns nonfarm payroll data.
	GetNonfarmPayroll() (*EconomicIndicator, error)

	// =========================================================================
	// Alpha Intelligence
	// =========================================================================

	// GetTopGainersLosers returns top gainers, losers, and most active stocks.
	GetTopGainersLosers() (*MarketMovers, error)

	// GetEarningsCallTranscript returns the transcript for an earnings call.
	GetEarningsCallTranscript(symbol string, year, quarter int) (*Transcript, error)

	// GetInsiderTransactions returns insider trading data for a symbol.
	GetInsiderTransactions(symbol string) ([]InsiderTransaction, error)

	// GetAnalyticsFixedWindow returns analytics for a fixed time window.
	GetAnalyticsFixedWindow(symbols []string, startDate, endDate string) ([]AnalyticsWindow, error)

	// GetAnalyticsSlidingWindow returns analytics for a sliding time window.
	GetAnalyticsSlidingWindow(symbols []string, windowSize int) ([]AnalyticsWindow, error)

	// =========================================================================
	// Options
	// =========================================================================

	// GetHistoricalOptions returns historical options chain data.
	// date format: "YYYY-MM-DD"
	GetHistoricalOptions(symbol, date string) (*OptionsChain, error)

	// =========================================================================
	// Utility
	// =========================================================================

	// GetRemainingRequests returns the number of remaining API requests for today.
	GetRemainingRequests() int

	// ResetDailyCounter resets the daily request counter (for testing).
	ResetDailyCounter()
}
