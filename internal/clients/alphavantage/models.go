// Package alphavantage provides a client for the Alpha Vantage API.
package alphavantage

import (
	"time"
)

// =============================================================================
// Error Types
// =============================================================================

// ErrRateLimitExceeded is returned when the daily API limit is reached.
type ErrRateLimitExceeded struct{}

func (e ErrRateLimitExceeded) Error() string {
	return "alpha vantage daily rate limit exceeded (25 requests/day on free tier)"
}

// ErrInvalidAPIKey is returned when the API key is invalid.
type ErrInvalidAPIKey struct{}

func (e ErrInvalidAPIKey) Error() string {
	return "invalid alpha vantage API key"
}

// ErrSymbolNotFound is returned when the requested symbol is not found.
type ErrSymbolNotFound struct {
	Symbol string
}

func (e ErrSymbolNotFound) Error() string {
	return "symbol not found: " + e.Symbol
}

// =============================================================================
// Time Series Models
// =============================================================================

// DailyPrice represents a single day's OHLCV data.
type DailyPrice struct {
	Date   time.Time `json:"date"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume int64     `json:"volume"`
}

// AdjustedPrice extends DailyPrice with adjusted values and dividend/split info.
type AdjustedPrice struct {
	Date             time.Time `json:"date"`
	Open             float64   `json:"open"`
	High             float64   `json:"high"`
	Low              float64   `json:"low"`
	Close            float64   `json:"close"`
	AdjustedClose    float64   `json:"adjusted_close"`
	Volume           int64     `json:"volume"`
	DividendAmount   float64   `json:"dividend_amount"`
	SplitCoefficient float64   `json:"split_coefficient"`
}

// GlobalQuote represents the latest quote for a symbol.
type GlobalQuote struct {
	Symbol           string    `json:"symbol"`
	Open             float64   `json:"open"`
	High             float64   `json:"high"`
	Low              float64   `json:"low"`
	Price            float64   `json:"price"`
	Volume           int64     `json:"volume"`
	LatestTradingDay time.Time `json:"latest_trading_day"`
	PreviousClose    float64   `json:"previous_close"`
	Change           float64   `json:"change"`
	ChangePercent    float64   `json:"change_percent"`
}

// SymbolMatch represents a search result from symbol search.
type SymbolMatch struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Region      string `json:"region"`
	MarketOpen  string `json:"market_open"`
	MarketClose string `json:"market_close"`
	Timezone    string `json:"timezone"`
	Currency    string `json:"currency"`
	MatchScore  string `json:"match_score"`
}

// =============================================================================
// Fundamental Data Models
// =============================================================================

// CompanyOverview contains comprehensive company information and financial metrics.
type CompanyOverview struct {
	Symbol                     string   `json:"symbol"`
	AssetType                  string   `json:"asset_type"`
	Name                       string   `json:"name"`
	Description                string   `json:"description"`
	CIK                        string   `json:"cik"`
	Exchange                   string   `json:"exchange"`
	Currency                   string   `json:"currency"`
	Country                    string   `json:"country"`
	Sector                     string   `json:"sector"`
	Industry                   string   `json:"industry"`
	Address                    string   `json:"address"`
	FullTimeEmployees          int64    `json:"full_time_employees"`
	FiscalYearEnd              string   `json:"fiscal_year_end"`
	LatestQuarter              string   `json:"latest_quarter"`
	MarketCapitalization       int64    `json:"market_capitalization"`
	EBITDA                     int64    `json:"ebitda"`
	PERatio                    *float64 `json:"pe_ratio"`
	PEGRatio                   *float64 `json:"peg_ratio"`
	BookValue                  *float64 `json:"book_value"`
	DividendPerShare           *float64 `json:"dividend_per_share"`
	DividendYield              *float64 `json:"dividend_yield"`
	EPS                        *float64 `json:"eps"`
	RevenuePerShareTTM         *float64 `json:"revenue_per_share_ttm"`
	ProfitMargin               *float64 `json:"profit_margin"`
	OperatingMarginTTM         *float64 `json:"operating_margin_ttm"`
	ReturnOnAssetsTTM          *float64 `json:"return_on_assets_ttm"`
	ReturnOnEquityTTM          *float64 `json:"return_on_equity_ttm"`
	RevenueTTM                 int64    `json:"revenue_ttm"`
	GrossProfitTTM             int64    `json:"gross_profit_ttm"`
	DilutedEPSTTM              *float64 `json:"diluted_eps_ttm"`
	QuarterlyEarningsGrowthYOY *float64 `json:"quarterly_earnings_growth_yoy"`
	QuarterlyRevenueGrowthYOY  *float64 `json:"quarterly_revenue_growth_yoy"`
	AnalystTargetPrice         *float64 `json:"analyst_target_price"`
	AnalystRatingStrongBuy     int      `json:"analyst_rating_strong_buy"`
	AnalystRatingBuy           int      `json:"analyst_rating_buy"`
	AnalystRatingHold          int      `json:"analyst_rating_hold"`
	AnalystRatingSell          int      `json:"analyst_rating_sell"`
	AnalystRatingStrongSell    int      `json:"analyst_rating_strong_sell"`
	TrailingPE                 *float64 `json:"trailing_pe"`
	ForwardPE                  *float64 `json:"forward_pe"`
	PriceToSalesRatioTTM       *float64 `json:"price_to_sales_ratio_ttm"`
	PriceToBookRatio           *float64 `json:"price_to_book_ratio"`
	EVToRevenue                *float64 `json:"ev_to_revenue"`
	EVToEBITDA                 *float64 `json:"ev_to_ebitda"`
	Beta                       *float64 `json:"beta"`
	FiftyTwoWeekHigh           *float64 `json:"fifty_two_week_high"`
	FiftyTwoWeekLow            *float64 `json:"fifty_two_week_low"`
	FiftyDayMovingAverage      *float64 `json:"fifty_day_moving_average"`
	TwoHundredDayMovingAverage *float64 `json:"two_hundred_day_moving_average"`
	SharesOutstanding          int64    `json:"shares_outstanding"`
	DividendDate               string   `json:"dividend_date"`
	ExDividendDate             string   `json:"ex_dividend_date"`
}

// IncomeStatementReport represents a single period's income statement.
type IncomeStatementReport struct {
	FiscalDateEnding                  string `json:"fiscal_date_ending"`
	ReportedCurrency                  string `json:"reported_currency"`
	GrossProfit                       int64  `json:"gross_profit"`
	TotalRevenue                      int64  `json:"total_revenue"`
	CostOfRevenue                     int64  `json:"cost_of_revenue"`
	CostOfGoodsAndServicesSold        int64  `json:"cost_of_goods_and_services_sold"`
	OperatingIncome                   int64  `json:"operating_income"`
	SellingGeneralAndAdministrative   int64  `json:"selling_general_and_administrative"`
	ResearchAndDevelopment            int64  `json:"research_and_development"`
	OperatingExpenses                 int64  `json:"operating_expenses"`
	InvestmentIncomeNet               int64  `json:"investment_income_net"`
	NetInterestIncome                 int64  `json:"net_interest_income"`
	InterestIncome                    int64  `json:"interest_income"`
	InterestExpense                   int64  `json:"interest_expense"`
	NonInterestIncome                 int64  `json:"non_interest_income"`
	OtherNonOperatingIncome           int64  `json:"other_non_operating_income"`
	Depreciation                      int64  `json:"depreciation"`
	DepreciationAndAmortization       int64  `json:"depreciation_and_amortization"`
	IncomeBeforeTax                   int64  `json:"income_before_tax"`
	IncomeTaxExpense                  int64  `json:"income_tax_expense"`
	InterestAndDebtExpense            int64  `json:"interest_and_debt_expense"`
	NetIncomeFromContinuingOperations int64  `json:"net_income_from_continuing_operations"`
	ComprehensiveIncomeNetOfTax       int64  `json:"comprehensive_income_net_of_tax"`
	EBIT                              int64  `json:"ebit"`
	EBITDA                            int64  `json:"ebitda"`
	NetIncome                         int64  `json:"net_income"`
}

// IncomeStatement contains annual and quarterly income statements.
type IncomeStatement struct {
	Symbol           string                  `json:"symbol"`
	AnnualReports    []IncomeStatementReport `json:"annual_reports"`
	QuarterlyReports []IncomeStatementReport `json:"quarterly_reports"`
}

// BalanceSheetReport represents a single period's balance sheet.
type BalanceSheetReport struct {
	FiscalDateEnding                       string `json:"fiscal_date_ending"`
	ReportedCurrency                       string `json:"reported_currency"`
	TotalAssets                            int64  `json:"total_assets"`
	TotalCurrentAssets                     int64  `json:"total_current_assets"`
	CashAndCashEquivalentsAtCarryingValue  int64  `json:"cash_and_cash_equivalents"`
	CashAndShortTermInvestments            int64  `json:"cash_and_short_term_investments"`
	Inventory                              int64  `json:"inventory"`
	CurrentNetReceivables                  int64  `json:"current_net_receivables"`
	TotalNonCurrentAssets                  int64  `json:"total_non_current_assets"`
	PropertyPlantEquipment                 int64  `json:"property_plant_equipment"`
	AccumulatedDepreciationAmortizationPPE int64  `json:"accumulated_depreciation_amortization_ppe"`
	IntangibleAssets                       int64  `json:"intangible_assets"`
	IntangibleAssetsExcludingGoodwill      int64  `json:"intangible_assets_excluding_goodwill"`
	Goodwill                               int64  `json:"goodwill"`
	Investments                            int64  `json:"investments"`
	LongTermInvestments                    int64  `json:"long_term_investments"`
	ShortTermInvestments                   int64  `json:"short_term_investments"`
	OtherCurrentAssets                     int64  `json:"other_current_assets"`
	OtherNonCurrentAssets                  int64  `json:"other_non_current_assets"`
	TotalLiabilities                       int64  `json:"total_liabilities"`
	TotalCurrentLiabilities                int64  `json:"total_current_liabilities"`
	CurrentAccountsPayable                 int64  `json:"current_accounts_payable"`
	DeferredRevenue                        int64  `json:"deferred_revenue"`
	CurrentDebt                            int64  `json:"current_debt"`
	ShortTermDebt                          int64  `json:"short_term_debt"`
	TotalNonCurrentLiabilities             int64  `json:"total_non_current_liabilities"`
	CapitalLeaseObligations                int64  `json:"capital_lease_obligations"`
	LongTermDebt                           int64  `json:"long_term_debt"`
	CurrentLongTermDebt                    int64  `json:"current_long_term_debt"`
	LongTermDebtNoncurrent                 int64  `json:"long_term_debt_noncurrent"`
	ShortLongTermDebtTotal                 int64  `json:"short_long_term_debt_total"`
	OtherCurrentLiabilities                int64  `json:"other_current_liabilities"`
	OtherNonCurrentLiabilities             int64  `json:"other_non_current_liabilities"`
	TotalShareholderEquity                 int64  `json:"total_shareholder_equity"`
	TreasuryStock                          int64  `json:"treasury_stock"`
	RetainedEarnings                       int64  `json:"retained_earnings"`
	CommonStock                            int64  `json:"common_stock"`
	CommonStockSharesOutstanding           int64  `json:"common_stock_shares_outstanding"`
}

// BalanceSheet contains annual and quarterly balance sheets.
type BalanceSheet struct {
	Symbol           string               `json:"symbol"`
	AnnualReports    []BalanceSheetReport `json:"annual_reports"`
	QuarterlyReports []BalanceSheetReport `json:"quarterly_reports"`
}

// CashFlowReport represents a single period's cash flow statement.
type CashFlowReport struct {
	FiscalDateEnding                                   string `json:"fiscal_date_ending"`
	ReportedCurrency                                   string `json:"reported_currency"`
	OperatingCashflow                                  int64  `json:"operating_cashflow"`
	PaymentsForOperatingActivities                     int64  `json:"payments_for_operating_activities"`
	ProceedsFromOperatingActivities                    int64  `json:"proceeds_from_operating_activities"`
	ChangeInOperatingLiabilities                       int64  `json:"change_in_operating_liabilities"`
	ChangeInOperatingAssets                            int64  `json:"change_in_operating_assets"`
	DepreciationDepletionAndAmortization               int64  `json:"depreciation_depletion_and_amortization"`
	CapitalExpenditures                                int64  `json:"capital_expenditures"`
	ChangeInReceivables                                int64  `json:"change_in_receivables"`
	ChangeInInventory                                  int64  `json:"change_in_inventory"`
	ProfitLoss                                         int64  `json:"profit_loss"`
	CashflowFromInvestment                             int64  `json:"cashflow_from_investment"`
	CashflowFromFinancing                              int64  `json:"cashflow_from_financing"`
	ProceedsFromRepaymentsOfShortTermDebt              int64  `json:"proceeds_from_repayments_of_short_term_debt"`
	PaymentsForRepurchaseOfCommonStock                 int64  `json:"payments_for_repurchase_of_common_stock"`
	PaymentsForRepurchaseOfEquity                      int64  `json:"payments_for_repurchase_of_equity"`
	PaymentsForRepurchaseOfPreferredStock              int64  `json:"payments_for_repurchase_of_preferred_stock"`
	DividendPayout                                     int64  `json:"dividend_payout"`
	DividendPayoutCommonStock                          int64  `json:"dividend_payout_common_stock"`
	DividendPayoutPreferredStock                       int64  `json:"dividend_payout_preferred_stock"`
	ProceedsFromIssuanceOfCommonStock                  int64  `json:"proceeds_from_issuance_of_common_stock"`
	ProceedsFromIssuanceOfLongTermDebtAndCapitalSecNet int64  `json:"proceeds_from_issuance_of_long_term_debt"`
	ProceedsFromIssuanceOfPreferredStock               int64  `json:"proceeds_from_issuance_of_preferred_stock"`
	ProceedsFromRepurchaseOfEquity                     int64  `json:"proceeds_from_repurchase_of_equity"`
	ProceedsFromSaleOfTreasuryStock                    int64  `json:"proceeds_from_sale_of_treasury_stock"`
	ChangeInCashAndCashEquivalents                     int64  `json:"change_in_cash_and_cash_equivalents"`
	ChangeInExchangeRate                               int64  `json:"change_in_exchange_rate"`
	NetIncome                                          int64  `json:"net_income"`
}

// CashFlow contains annual and quarterly cash flow statements.
type CashFlow struct {
	Symbol           string           `json:"symbol"`
	AnnualReports    []CashFlowReport `json:"annual_reports"`
	QuarterlyReports []CashFlowReport `json:"quarterly_reports"`
}

// EarningsReport represents earnings for a single period.
type EarningsReport struct {
	FiscalDateEnding   string   `json:"fiscal_date_ending"`
	ReportedDate       string   `json:"reported_date"`
	ReportedEPS        *float64 `json:"reported_eps"`
	EstimatedEPS       *float64 `json:"estimated_eps"`
	Surprise           *float64 `json:"surprise"`
	SurprisePercentage *float64 `json:"surprise_percentage"`
}

// Earnings contains annual and quarterly earnings data.
type Earnings struct {
	Symbol            string           `json:"symbol"`
	AnnualEarnings    []EarningsReport `json:"annual_earnings"`
	QuarterlyEarnings []EarningsReport `json:"quarterly_earnings"`
}

// DividendRecord represents a single dividend payment.
type DividendRecord struct {
	ExDate         time.Time `json:"ex_date"`
	PaymentDate    time.Time `json:"payment_date"`
	RecordDate     time.Time `json:"record_date"`
	DeclaredDate   time.Time `json:"declared_date"`
	Amount         float64   `json:"amount"`
	AdjustedAmount float64   `json:"adjusted_amount"`
}

// SplitRecord represents a stock split event.
type SplitRecord struct {
	EffectiveDate    time.Time `json:"effective_date"`
	SplitCoefficient float64   `json:"split_coefficient"`
}

// ETFProfile contains ETF-specific information.
type ETFProfile struct {
	Symbol          string    `json:"symbol"`
	AssetType       string    `json:"asset_type"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	InceptionDate   time.Time `json:"inception_date"`
	Exchange        string    `json:"exchange"`
	AssetClass      string    `json:"asset_class"`
	ExpenseRatio    float64   `json:"expense_ratio"`
	DividendYield   float64   `json:"dividend_yield"`
	TotalAssets     int64     `json:"total_assets"`
	HoldingsCount   int       `json:"holdings_count"`
	TurnoverRatio   float64   `json:"turnover_ratio"`
	PriceToEarnings *float64  `json:"price_to_earnings"`
	PriceToBook     *float64  `json:"price_to_book"`
}

// ETFHolding represents a single holding in an ETF.
type ETFHolding struct {
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	Weight      float64 `json:"weight"`
	SharesHeld  int64   `json:"shares_held"`
	MarketValue int64   `json:"market_value"`
}

// EarningsEvent represents an upcoming earnings announcement.
type EarningsEvent struct {
	Symbol           string    `json:"symbol"`
	Name             string    `json:"name"`
	ReportDate       time.Time `json:"report_date"`
	FiscalDateEnding string    `json:"fiscal_date_ending"`
	Estimate         *float64  `json:"estimate"`
	Currency         string    `json:"currency"`
}

// IPOEvent represents an upcoming IPO.
type IPOEvent struct {
	Symbol         string    `json:"symbol"`
	Name           string    `json:"name"`
	IPODate        time.Time `json:"ipo_date"`
	PriceRangeLow  *float64  `json:"price_range_low"`
	PriceRangeHigh *float64  `json:"price_range_high"`
	Currency       string    `json:"currency"`
	Exchange       string    `json:"exchange"`
}

// SharesOutstandingRecord represents shares outstanding at a point in time.
type SharesOutstandingRecord struct {
	Date   time.Time `json:"date"`
	Shares int64     `json:"shares"`
}

// ListingStatus represents active/delisted status for securities.
type ListingStatus struct {
	Symbol     string `json:"symbol"`
	Name       string `json:"name"`
	Exchange   string `json:"exchange"`
	AssetType  string `json:"asset_type"`
	IPODate    string `json:"ipo_date"`
	DelistDate string `json:"delist_date"`
	Status     string `json:"status"` // "active" or "delisted"
}

// =============================================================================
// Technical Indicator Models
// =============================================================================

// IndicatorValue represents a single technical indicator data point.
type IndicatorValue struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

// IndicatorData contains the result of a technical indicator calculation.
type IndicatorData struct {
	Symbol   string           `json:"symbol"`
	Interval string           `json:"interval"`
	Values   []IndicatorValue `json:"values"`
}

// MACDValue represents MACD indicator data for a single period.
type MACDValue struct {
	Date      time.Time `json:"date"`
	MACD      float64   `json:"macd"`
	Signal    float64   `json:"signal"`
	Histogram float64   `json:"histogram"`
}

// MACDData contains MACD indicator results.
type MACDData struct {
	Symbol   string      `json:"symbol"`
	Interval string      `json:"interval"`
	Values   []MACDValue `json:"values"`
}

// BollingerValue represents Bollinger Bands data for a single period.
type BollingerValue struct {
	Date       time.Time `json:"date"`
	UpperBand  float64   `json:"upper_band"`
	MiddleBand float64   `json:"middle_band"`
	LowerBand  float64   `json:"lower_band"`
}

// BollingerData contains Bollinger Bands results.
type BollingerData struct {
	Symbol   string           `json:"symbol"`
	Interval string           `json:"interval"`
	Values   []BollingerValue `json:"values"`
}

// StochValue represents Stochastic Oscillator data for a single period.
type StochValue struct {
	Date  time.Time `json:"date"`
	SlowK float64   `json:"slow_k"`
	SlowD float64   `json:"slow_d"`
}

// StochData contains Stochastic Oscillator results.
type StochData struct {
	Symbol   string       `json:"symbol"`
	Interval string       `json:"interval"`
	Values   []StochValue `json:"values"`
}

// AroonValue represents Aroon indicator data for a single period.
type AroonValue struct {
	Date      time.Time `json:"date"`
	AroonUp   float64   `json:"aroon_up"`
	AroonDown float64   `json:"aroon_down"`
}

// AroonData contains Aroon indicator results.
type AroonData struct {
	Symbol   string       `json:"symbol"`
	Interval string       `json:"interval"`
	Values   []AroonValue `json:"values"`
}

// =============================================================================
// Forex Models
// =============================================================================

// ExchangeRate represents a currency exchange rate.
type ExchangeRate struct {
	FromCurrency     string    `json:"from_currency"`
	FromCurrencyName string    `json:"from_currency_name"`
	ToCurrency       string    `json:"to_currency"`
	ToCurrencyName   string    `json:"to_currency_name"`
	ExchangeRate     float64   `json:"exchange_rate"`
	LastRefreshed    time.Time `json:"last_refreshed"`
	Timezone         string    `json:"timezone"`
	BidPrice         float64   `json:"bid_price"`
	AskPrice         float64   `json:"ask_price"`
}

// FXPrice represents forex OHLC data for a single period.
type FXPrice struct {
	Date  time.Time `json:"date"`
	Open  float64   `json:"open"`
	High  float64   `json:"high"`
	Low   float64   `json:"low"`
	Close float64   `json:"close"`
}

// =============================================================================
// Cryptocurrency Models
// =============================================================================

// CryptoPrice represents cryptocurrency OHLCV data for a single period.
type CryptoPrice struct {
	Date      time.Time `json:"date"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	MarketCap int64     `json:"market_cap"`
}

// =============================================================================
// Commodity Models
// =============================================================================

// CommodityPrice represents commodity price data for a single period.
type CommodityPrice struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

// =============================================================================
// Economic Indicator Models
// =============================================================================

// EconomicDataPoint represents a single economic indicator data point.
type EconomicDataPoint struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

// EconomicIndicator contains economic indicator data.
type EconomicIndicator struct {
	Name     string              `json:"name"`
	Interval string              `json:"interval"`
	Unit     string              `json:"unit"`
	Data     []EconomicDataPoint `json:"data"`
}

// =============================================================================
// Alpha Intelligence Models
// =============================================================================

// MarketMover represents a top gainer, loser, or most active stock.
type MarketMover struct {
	Ticker        string  `json:"ticker"`
	Price         float64 `json:"price"`
	ChangeAmount  float64 `json:"change_amount"`
	ChangePercent float64 `json:"change_percent"`
	Volume        int64   `json:"volume"`
}

// MarketMovers contains top gainers, losers, and most active stocks.
type MarketMovers struct {
	LastUpdated time.Time     `json:"last_updated"`
	TopGainers  []MarketMover `json:"top_gainers"`
	TopLosers   []MarketMover `json:"top_losers"`
	MostActive  []MarketMover `json:"most_active"`
}

// Transcript represents an earnings call transcript.
type Transcript struct {
	Symbol     string `json:"symbol"`
	Quarter    int    `json:"quarter"`
	Year       int    `json:"year"`
	Transcript string `json:"transcript"`
}

// InsiderTransaction represents an insider trading transaction.
type InsiderTransaction struct {
	Symbol                 string    `json:"symbol"`
	FilingDate             time.Time `json:"filing_date"`
	TransactionDate        time.Time `json:"transaction_date"`
	OwnerName              string    `json:"owner_name"`
	OwnerTitle             string    `json:"owner_title"`
	TransactionType        string    `json:"transaction_type"`
	AcquisitionDisposition string    `json:"acquisition_disposition"`
	SharesTraded           int64     `json:"shares_traded"`
	Price                  *float64  `json:"price"`
	SharesOwned            int64     `json:"shares_owned"`
}

// AnalyticsWindow represents analytics data for a time window.
type AnalyticsWindow struct {
	Symbol             string    `json:"symbol"`
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
	AveragePrice       float64   `json:"average_price"`
	HighPrice          float64   `json:"high_price"`
	LowPrice           float64   `json:"low_price"`
	AverageVolume      int64     `json:"average_volume"`
	TotalVolume        int64     `json:"total_volume"`
	PriceChange        float64   `json:"price_change"`
	PriceChangePercent float64   `json:"price_change_percent"`
	Volatility         float64   `json:"volatility"`
}

// =============================================================================
// Options Models
// =============================================================================

// OptionContract represents a single options contract.
type OptionContract struct {
	ContractID    string    `json:"contract_id"`
	Symbol        string    `json:"symbol"`
	Expiration    time.Time `json:"expiration"`
	Strike        float64   `json:"strike"`
	Type          string    `json:"type"` // "call" or "put"
	Last          float64   `json:"last"`
	Mark          float64   `json:"mark"`
	Bid           float64   `json:"bid"`
	Ask           float64   `json:"ask"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"change_percent"`
	Volume        int64     `json:"volume"`
	OpenInterest  int64     `json:"open_interest"`
	ImpliedVol    float64   `json:"implied_volatility"`
	Delta         float64   `json:"delta"`
	Gamma         float64   `json:"gamma"`
	Theta         float64   `json:"theta"`
	Vega          float64   `json:"vega"`
	Rho           float64   `json:"rho"`
}

// OptionsChain contains call and put options for a symbol.
type OptionsChain struct {
	Symbol string           `json:"symbol"`
	Date   time.Time        `json:"date"`
	Calls  []OptionContract `json:"calls"`
	Puts   []OptionContract `json:"puts"`
}
