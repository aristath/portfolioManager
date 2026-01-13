package alphavantage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFundamentalsParseCompanyOverview(t *testing.T) {
	jsonData := `{
		"Symbol": "IBM",
		"AssetType": "Common Stock",
		"Name": "International Business Machines Corporation",
		"Description": "IBM is a global technology company.",
		"CIK": "51143",
		"Exchange": "NYSE",
		"Currency": "USD",
		"Country": "USA",
		"Sector": "Technology",
		"Industry": "Information Technology Services",
		"Address": "One New Orchard Road, Armonk, NY, US",
		"MarketCapitalization": "125000000000",
		"EBITDA": "15000000000",
		"PERatio": "20.5",
		"PEGRatio": "1.5",
		"BookValue": "50.25",
		"DividendPerShare": "6.60",
		"DividendYield": "0.0485",
		"EPS": "9.05",
		"RevenuePerShareTTM": "110.50",
		"ProfitMargin": "0.12",
		"OperatingMarginTTM": "0.11",
		"ReturnOnAssetsTTM": "0.05",
		"ReturnOnEquityTTM": "0.25",
		"RevenueTTM": "60000000000",
		"GrossProfitTTM": "30000000000",
		"DilutedEPSTTM": "9.00",
		"QuarterlyEarningsGrowthYOY": "0.05",
		"QuarterlyRevenueGrowthYOY": "0.03",
		"AnalystTargetPrice": "150.00",
		"AnalystRatingStrongBuy": "5",
		"AnalystRatingBuy": "10",
		"AnalystRatingHold": "8",
		"AnalystRatingSell": "2",
		"AnalystRatingStrongSell": "1",
		"TrailingPE": "20.5",
		"ForwardPE": "18.0",
		"PriceToSalesRatioTTM": "2.1",
		"PriceToBookRatio": "3.5",
		"EVToRevenue": "2.5",
		"EVToEBITDA": "10.0",
		"Beta": "0.95",
		"52WeekHigh": "200.00",
		"52WeekLow": "120.00",
		"50DayMovingAverage": "175.00",
		"200DayMovingAverage": "165.00",
		"SharesOutstanding": "900000000",
		"DividendDate": "2024-03-15",
		"ExDividendDate": "2024-02-15"
	}`

	overview, err := parseCompanyOverview([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, "IBM", overview.Symbol)
	assert.Equal(t, "Common Stock", overview.AssetType)
	assert.Equal(t, "International Business Machines Corporation", overview.Name)
	assert.Equal(t, "NYSE", overview.Exchange)
	assert.Equal(t, "USD", overview.Currency)
	assert.Equal(t, "Technology", overview.Sector)
	assert.Equal(t, int64(125000000000), overview.MarketCapitalization)
	require.NotNil(t, overview.PERatio)
	assert.Equal(t, 20.5, *overview.PERatio)
	require.NotNil(t, overview.EPS)
	assert.Equal(t, 9.05, *overview.EPS)
	require.NotNil(t, overview.FiftyTwoWeekHigh)
	assert.Equal(t, 200.0, *overview.FiftyTwoWeekHigh)
	require.NotNil(t, overview.FiftyTwoWeekLow)
	assert.Equal(t, 120.0, *overview.FiftyTwoWeekLow)
}

func TestFundamentalsParseIncomeStatement(t *testing.T) {
	jsonData := `{
		"symbol": "IBM",
		"annualReports": [
			{
				"fiscalDateEnding": "2023-12-31",
				"reportedCurrency": "USD",
				"grossProfit": "30000000000",
				"totalRevenue": "60000000000",
				"costOfRevenue": "30000000000",
				"costofGoodsAndServicesSold": "28000000000",
				"operatingIncome": "9000000000",
				"sellingGeneralAndAdministrative": "15000000000",
				"researchAndDevelopment": "6000000000",
				"operatingExpenses": "21000000000",
				"netInterestIncome": "-500000000",
				"interestIncome": "100000000",
				"interestExpense": "600000000",
				"nonInterestIncome": "60000000000",
				"netIncome": "7200000000",
				"netIncomeFromContinuingOperations": "7200000000",
				"ebit": "8500000000",
				"ebitda": "12000000000",
				"incomeBeforeTax": "8000000000",
				"incomeTaxExpense": "800000000",
				"comprehensiveIncomeNetOfTax": "7000000000"
			}
		],
		"quarterlyReports": []
	}`

	stmt, err := parseIncomeStatement([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, "IBM", stmt.Symbol)
	require.Len(t, stmt.AnnualReports, 1)
	assert.Equal(t, "2023-12-31", stmt.AnnualReports[0].FiscalDateEnding)
	assert.Equal(t, "USD", stmt.AnnualReports[0].ReportedCurrency)
	assert.Equal(t, int64(60000000000), stmt.AnnualReports[0].TotalRevenue)
	assert.Equal(t, int64(30000000000), stmt.AnnualReports[0].GrossProfit)
	assert.Equal(t, int64(7200000000), stmt.AnnualReports[0].NetIncome)
}

func TestParseBalanceSheet(t *testing.T) {
	jsonData := `{
		"symbol": "IBM",
		"annualReports": [
			{
				"fiscalDateEnding": "2023-12-31",
				"reportedCurrency": "USD",
				"totalAssets": "130000000000",
				"totalCurrentAssets": "30000000000",
				"cashAndCashEquivalentsAtCarryingValue": "10000000000",
				"cashAndShortTermInvestments": "12000000000",
				"inventory": "2000000000",
				"currentNetReceivables": "15000000000",
				"totalNonCurrentAssets": "100000000000",
				"propertyPlantEquipment": "20000000000",
				"goodwill": "50000000000",
				"intangibleAssets": "15000000000",
				"totalLiabilities": "85000000000",
				"totalCurrentLiabilities": "25000000000",
				"currentAccountsPayable": "5000000000",
				"shortTermDebt": "8000000000",
				"totalNonCurrentLiabilities": "60000000000",
				"longTermDebt": "45000000000",
				"totalShareholderEquity": "45000000000",
				"retainedEarnings": "150000000000",
				"commonStock": "60000000000",
				"commonStockSharesOutstanding": "900000000"
			}
		],
		"quarterlyReports": []
	}`

	bs, err := parseBalanceSheet([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, "IBM", bs.Symbol)
	require.Len(t, bs.AnnualReports, 1)
	assert.Equal(t, int64(130000000000), bs.AnnualReports[0].TotalAssets)
	assert.Equal(t, int64(45000000000), bs.AnnualReports[0].TotalShareholderEquity)
}

func TestParseCashFlow(t *testing.T) {
	jsonData := `{
		"symbol": "IBM",
		"annualReports": [
			{
				"fiscalDateEnding": "2023-12-31",
				"reportedCurrency": "USD",
				"operatingCashflow": "12000000000",
				"paymentsForOperatingActivities": "30000000000",
				"capitalExpenditures": "3000000000",
				"cashflowFromInvestment": "-5000000000",
				"cashflowFromFinancing": "-8000000000",
				"dividendPayout": "6000000000",
				"dividendPayoutCommonStock": "6000000000",
				"proceedsFromIssuanceOfLongTermDebtAndCapitalSecuritiesNet": "5000000000",
				"proceedsFromRepurchaseOfEquity": "-4000000000",
				"netIncome": "7200000000"
			}
		],
		"quarterlyReports": []
	}`

	cf, err := parseCashFlow([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, "IBM", cf.Symbol)
	require.Len(t, cf.AnnualReports, 1)
	assert.Equal(t, int64(12000000000), cf.AnnualReports[0].OperatingCashflow)
	assert.Equal(t, int64(6000000000), cf.AnnualReports[0].DividendPayout)
}

func TestParseEarnings(t *testing.T) {
	jsonData := `{
		"symbol": "IBM",
		"annualEarnings": [
			{
				"fiscalDateEnding": "2023-12-31",
				"reportedEPS": "9.05"
			}
		],
		"quarterlyEarnings": [
			{
				"fiscalDateEnding": "2023-12-31",
				"reportedDate": "2024-01-25",
				"reportedEPS": "2.50",
				"estimatedEPS": "2.40",
				"surprise": "0.10",
				"surprisePercentage": "4.17"
			}
		]
	}`

	earnings, err := parseEarnings([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, "IBM", earnings.Symbol)
	require.Len(t, earnings.AnnualEarnings, 1)
	assert.Equal(t, "2023-12-31", earnings.AnnualEarnings[0].FiscalDateEnding)
	require.NotNil(t, earnings.AnnualEarnings[0].ReportedEPS)
	assert.Equal(t, 9.05, *earnings.AnnualEarnings[0].ReportedEPS)

	require.Len(t, earnings.QuarterlyEarnings, 1)
	require.NotNil(t, earnings.QuarterlyEarnings[0].ReportedEPS)
	assert.Equal(t, 2.50, *earnings.QuarterlyEarnings[0].ReportedEPS)
	require.NotNil(t, earnings.QuarterlyEarnings[0].EstimatedEPS)
	assert.Equal(t, 2.40, *earnings.QuarterlyEarnings[0].EstimatedEPS)
}

func TestParseDividends(t *testing.T) {
	jsonData := `{
		"data": [
			{
				"ex_dividend_date": "2024-02-15",
				"declaration_date": "2024-01-30",
				"record_date": "2024-02-16",
				"payment_date": "2024-03-10",
				"amount": "1.66"
			},
			{
				"ex_dividend_date": "2023-11-09",
				"declaration_date": "2023-10-31",
				"record_date": "2023-11-10",
				"payment_date": "2023-12-09",
				"amount": "1.66"
			}
		]
	}`

	dividends, err := parseDividends([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, dividends, 2)

	assert.Equal(t, 1.66, dividends[0].Amount)
}

func TestParseSplits(t *testing.T) {
	jsonData := `{
		"data": [
			{
				"effective_date": "2020-08-31",
				"split_coefficient": "4.0"
			},
			{
				"effective_date": "2020-08-31",
				"split_coefficient": "5.0"
			}
		]
	}`

	splits, err := parseSplits([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, splits, 2)

	assert.Equal(t, 4.0, splits[0].SplitCoefficient)
}

func TestParseListingStatus(t *testing.T) {
	// CSV format
	csvData := `symbol,name,exchange,assetType,ipoDate,delistingDate,status
IBM,International Business Machines Corp,NYSE,Stock,1962-01-02,,Active
MSFT,Microsoft Corporation,NASDAQ,Stock,1986-03-13,,Active`

	listings, err := parseListingStatus([]byte(csvData))
	require.NoError(t, err)
	require.Len(t, listings, 2)

	assert.Equal(t, "IBM", listings[0].Symbol)
	assert.Equal(t, "NYSE", listings[0].Exchange)
	assert.Equal(t, "active", listings[0].Status)
}

func TestParseEarningsCalendar(t *testing.T) {
	// CSV format
	csvData := `symbol,name,reportDate,fiscalDateEnding,estimate,currency
IBM,International Business Machines Corp,2024-01-25,2023-12-31,2.40,USD
MSFT,Microsoft Corporation,2024-01-30,2023-12-31,2.75,USD`

	events, err := parseEarningsCalendar([]byte(csvData))
	require.NoError(t, err)
	require.Len(t, events, 2)

	assert.Equal(t, "IBM", events[0].Symbol)
	assert.Equal(t, "USD", events[0].Currency)
}

func TestParseIPOCalendar(t *testing.T) {
	// CSV format
	csvData := `symbol,name,ipoDate,priceRangeLow,priceRangeHigh,currency,exchange
NEWCO,New Company Inc,2024-02-15,15.00,18.00,USD,NASDAQ`

	ipos, err := parseIPOCalendar([]byte(csvData))
	require.NoError(t, err)
	require.Len(t, ipos, 1)

	assert.Equal(t, "NEWCO", ipos[0].Symbol)
	require.NotNil(t, ipos[0].PriceRangeLow)
	assert.Equal(t, 15.0, *ipos[0].PriceRangeLow)
	require.NotNil(t, ipos[0].PriceRangeHigh)
	assert.Equal(t, 18.0, *ipos[0].PriceRangeHigh)
}

func TestParseETFProfile(t *testing.T) {
	jsonData := `{
		"symbol": "SPY",
		"asset_type": "ETF",
		"name": "SPDR S&P 500 ETF Trust",
		"description": "The SPDR S&P 500 ETF Trust seeks to provide investment results that correspond to the S&P 500 Index.",
		"inception_date": "1993-01-22",
		"exchange": "NYSE",
		"asset_class": "US Equity",
		"expense_ratio": "0.0945",
		"dividend_yield": "1.45",
		"total_assets": "500000000000",
		"holdings_count": "503",
		"turnover_ratio": "2.5"
	}`

	profile, err := parseETFProfile([]byte(jsonData))
	require.NoError(t, err)

	assert.Equal(t, "SPY", profile.Symbol)
	assert.Equal(t, "ETF", profile.AssetType)
	assert.Equal(t, 0.0945, profile.ExpenseRatio)
	assert.Equal(t, 1.45, profile.DividendYield)
}

func TestParseETFHoldings(t *testing.T) {
	jsonData := `{
		"symbol": "SPY",
		"holdings": [
			{
				"symbol": "AAPL",
				"description": "Apple Inc",
				"weight": "7.5"
			},
			{
				"symbol": "MSFT",
				"description": "Microsoft Corp",
				"weight": "6.8"
			}
		]
	}`

	holdings, err := parseETFHoldings([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, holdings, 2)

	assert.Equal(t, "AAPL", holdings[0].Symbol)
	assert.Equal(t, 7.5, holdings[0].Weight)
}

func TestParseSharesOutstanding(t *testing.T) {
	// JSON format
	jsonData := `{
		"data": [
			{"date": "2024-01-15", "shares_outstanding": "900000000"},
			{"date": "2023-12-15", "shares_outstanding": "905000000"}
		]
	}`

	records, err := parseSharesOutstanding([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, records, 2)

	assert.Equal(t, int64(900000000), records[0].Shares)
}

func TestFundamentalsParseErrors(t *testing.T) {
	tests := []struct {
		name   string
		parser func([]byte) error
	}{
		{"CompanyOverview", func(b []byte) error { _, err := parseCompanyOverview(b); return err }},
		{"IncomeStatement", func(b []byte) error { _, err := parseIncomeStatement(b); return err }},
		{"BalanceSheet", func(b []byte) error { _, err := parseBalanceSheet(b); return err }},
		{"CashFlow", func(b []byte) error { _, err := parseCashFlow(b); return err }},
		{"Earnings", func(b []byte) error { _, err := parseEarnings(b); return err }},
		{"ETFProfile", func(b []byte) error { _, err := parseETFProfile(b); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_InvalidJSON", func(t *testing.T) {
			err := tt.parser([]byte("not json"))
			assert.Error(t, err)
		})
	}
}

// TestClientFundamentalsMethods verifies all fundamental methods exist.
func TestClientFundamentalsMethods(t *testing.T) {
	client := newTestClient("test-key")

	var _ func(string) (*CompanyOverview, error) = client.GetCompanyOverview
	var _ func(string) (*Earnings, error) = client.GetEarnings
	var _ func(string) (*IncomeStatement, error) = client.GetIncomeStatement
	var _ func(string) (*BalanceSheet, error) = client.GetBalanceSheet
	var _ func(string) (*CashFlow, error) = client.GetCashFlow
	var _ func(string) ([]DividendRecord, error) = client.GetDividends
	var _ func(string) ([]SplitRecord, error) = client.GetSplits
	var _ func(string) (*ETFProfile, error) = client.GetETFProfile
	var _ func(string) ([]ETFHolding, error) = client.GetETFHoldings
	var _ func(string) ([]SharesOutstandingRecord, error) = client.GetSharesOutstanding
	var _ func(string) ([]ListingStatus, error) = client.GetListingStatus
	var _ func(string) ([]EarningsEvent, error) = client.GetEarningsCalendar
	var _ func() ([]IPOEvent, error) = client.GetIPOCalendar
}
