package alphavantage

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/aristath/sentinel/internal/clientdata"
)

// =============================================================================
// Fundamental Data Endpoints
// =============================================================================

// GetCompanyOverview returns comprehensive company information and financial metrics.
func (c *Client) GetCompanyOverview(symbol string) (*CompanyOverview, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	// Check cache
	table := getTableForFunction("OVERVIEW")
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, isin); err == nil && data != nil {
			var overview CompanyOverview
			if err := json.Unmarshal(data, &overview); err == nil {
				return &overview, nil
			}
		}
	}

	// Fetch from API
	body, err := c.doRequest("OVERVIEW", map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}

	overview, err := parseCompanyOverview(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse company overview: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		ttl := getTTLForFunction("OVERVIEW")
		if err := c.cacheRepo.Store(table, isin, overview, ttl); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache company overview")
		}
	}

	return overview, nil
}

// GetEarnings returns annual and quarterly earnings data.
func (c *Client) GetEarnings(symbol string) (*Earnings, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	// Check cache
	table := getTableForFunction("EARNINGS")
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, isin); err == nil && data != nil {
			var earnings Earnings
			if err := json.Unmarshal(data, &earnings); err == nil {
				return &earnings, nil
			}
		}
	}

	// Fetch from API
	body, err := c.doRequest("EARNINGS", map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}

	earnings, err := parseEarnings(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse earnings: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		ttl := getTTLForFunction("EARNINGS")
		if err := c.cacheRepo.Store(table, isin, earnings, ttl); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache earnings")
		}
	}

	return earnings, nil
}

// GetIncomeStatement returns annual and quarterly income statements.
func (c *Client) GetIncomeStatement(symbol string) (*IncomeStatement, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	// Check cache (uses overview table per mapping)
	table := getTableForFunction("INCOME_STATEMENT")
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, isin); err == nil && data != nil {
			var stmt IncomeStatement
			if err := json.Unmarshal(data, &stmt); err == nil {
				return &stmt, nil
			}
		}
	}

	// Fetch from API
	body, err := c.doRequest("INCOME_STATEMENT", map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}

	stmt, err := parseIncomeStatement(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse income statement: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		ttl := getTTLForFunction("INCOME_STATEMENT")
		if err := c.cacheRepo.Store(table, isin, stmt, ttl); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache income statement")
		}
	}

	return stmt, nil
}

// GetBalanceSheet returns annual and quarterly balance sheets.
func (c *Client) GetBalanceSheet(symbol string) (*BalanceSheet, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	// Check cache
	table := getTableForFunction("BALANCE_SHEET")
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, isin); err == nil && data != nil {
			var sheet BalanceSheet
			if err := json.Unmarshal(data, &sheet); err == nil {
				return &sheet, nil
			}
		}
	}

	// Fetch from API
	body, err := c.doRequest("BALANCE_SHEET", map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}

	sheet, err := parseBalanceSheet(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance sheet: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		ttl := getTTLForFunction("BALANCE_SHEET")
		if err := c.cacheRepo.Store(table, isin, sheet, ttl); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache balance sheet")
		}
	}

	return sheet, nil
}

// GetCashFlow returns annual and quarterly cash flow statements.
func (c *Client) GetCashFlow(symbol string) (*CashFlow, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	// Check cache
	table := getTableForFunction("CASH_FLOW")
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, isin); err == nil && data != nil {
			var cf CashFlow
			if err := json.Unmarshal(data, &cf); err == nil {
				return &cf, nil
			}
		}
	}

	// Fetch from API
	body, err := c.doRequest("CASH_FLOW", map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}

	cf, err := parseCashFlow(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cash flow: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		ttl := getTTLForFunction("CASH_FLOW")
		if err := c.cacheRepo.Store(table, isin, cf, ttl); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache cash flow")
		}
	}

	return cf, nil
}

// GetDividends returns historical dividend data.
func (c *Client) GetDividends(symbol string) ([]DividendRecord, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	// Check cache
	table := getTableForFunction("DIVIDENDS")
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, isin); err == nil && data != nil {
			var dividends []DividendRecord
			if err := json.Unmarshal(data, &dividends); err == nil {
				return dividends, nil
			}
		}
	}

	// Fetch from API
	body, err := c.doRequest("DIVIDENDS", map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}

	dividends, err := parseDividends(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dividends: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		ttl := getTTLForFunction("DIVIDENDS")
		if err := c.cacheRepo.Store(table, isin, dividends, ttl); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache dividends")
		}
	}

	return dividends, nil
}

// GetSplits returns historical stock split data.
func (c *Client) GetSplits(symbol string) ([]SplitRecord, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	cacheKey := isin + ":SPLITS"

	// Check cache (using current_prices table)
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var splits []SplitRecord
			if err := json.Unmarshal(data, &splits); err == nil {
				return splits, nil
			}
		}
	}

	body, err := c.doRequest("SPLITS", map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}

	splits, err := parseSplits(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse splits: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, splits, clientdata.TTLAVOverview); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache splits")
		}
	}

	return splits, nil
}

// GetETFProfile returns ETF-specific information.
func (c *Client) GetETFProfile(symbol string) (*ETFProfile, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	// Check cache
	table := getTableForFunction("ETF_PROFILE")
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, isin); err == nil && data != nil {
			var profile ETFProfile
			if err := json.Unmarshal(data, &profile); err == nil {
				return &profile, nil
			}
		}
	}

	// Fetch from API
	body, err := c.doRequest("ETF_PROFILE", map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}

	profile, err := parseETFProfile(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ETF profile: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		ttl := getTTLForFunction("ETF_PROFILE")
		if err := c.cacheRepo.Store(table, isin, profile, ttl); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache ETF profile")
		}
	}

	return profile, nil
}

// GetETFHoldings returns the holdings of an ETF.
func (c *Client) GetETFHoldings(symbol string) ([]ETFHolding, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	cacheKey := isin + ":ETF_HOLDINGS"

	// Check cache (using current_prices table)
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var holdings []ETFHolding
			if err := json.Unmarshal(data, &holdings); err == nil {
				return holdings, nil
			}
		}
	}

	body, err := c.doRequest("ETF_HOLDINGS", map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}

	holdings, err := parseETFHoldings(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ETF holdings: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, holdings, clientdata.TTLETFProfile); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache ETF holdings")
		}
	}

	return holdings, nil
}

// GetSharesOutstanding returns historical shares outstanding data.
func (c *Client) GetSharesOutstanding(symbol string) ([]SharesOutstandingRecord, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	cacheKey := isin + ":SHARES_OUTSTANDING"

	// Check cache (using current_prices table)
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var shares []SharesOutstandingRecord
			if err := json.Unmarshal(data, &shares); err == nil {
				return shares, nil
			}
		}
	}

	body, err := c.doRequest("SHARES_OUTSTANDING", map[string]string{"symbol": symbol})
	if err != nil {
		return nil, err
	}

	shares, err := parseSharesOutstanding(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse shares outstanding: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, shares, clientdata.TTLAVOverview); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache shares outstanding")
		}
	}

	return shares, nil
}

// GetListingStatus returns listing status for securities.
// Market-wide data, doesn't require ISIN resolution.
func (c *Client) GetListingStatus(status string) ([]ListingStatus, error) {
	cacheKey := "LISTING_STATUS"
	if status != "" {
		cacheKey = cacheKey + ":" + status
	}

	// Check cache (using current_prices table for market-wide data)
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var listings []ListingStatus
			if err := json.Unmarshal(data, &listings); err == nil {
				return listings, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{}
	if status != "" {
		params["state"] = status
	}

	body, err := c.doRequest("LISTING_STATUS", params)
	if err != nil {
		return nil, err
	}

	listings, err := parseListingStatus(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse listing status: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, listings, clientdata.TTLAVOverview); err != nil {
			c.log.Warn().Err(err).Msg("Failed to cache listing status")
		}
	}

	return listings, nil
}

// GetEarningsCalendar returns upcoming earnings announcements.
// Market-wide data, doesn't require ISIN resolution.
func (c *Client) GetEarningsCalendar(horizon string) ([]EarningsEvent, error) {
	cacheKey := "EARNINGS_CALENDAR"
	if horizon != "" {
		cacheKey = cacheKey + ":" + horizon
	}

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var events []EarningsEvent
			if err := json.Unmarshal(data, &events); err == nil {
				return events, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{}
	if horizon != "" {
		params["horizon"] = horizon
	}

	body, err := c.doRequest("EARNINGS_CALENDAR", params)
	if err != nil {
		return nil, err
	}

	events, err := parseEarningsCalendar(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse earnings calendar: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, events, clientdata.TTLAVOverview); err != nil {
			c.log.Warn().Err(err).Msg("Failed to cache earnings calendar")
		}
	}

	return events, nil
}

// GetIPOCalendar returns upcoming IPO events.
// Market-wide data, doesn't require ISIN resolution.
func (c *Client) GetIPOCalendar() ([]IPOEvent, error) {
	cacheKey := "IPO_CALENDAR"

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var events []IPOEvent
			if err := json.Unmarshal(data, &events); err == nil {
				return events, nil
			}
		}
	}

	body, err := c.doRequest("IPO_CALENDAR", nil)
	if err != nil {
		return nil, err
	}

	events, err := parseIPOCalendar(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse IPO calendar: %w", err)
	}

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, events, clientdata.TTLAVOverview); err != nil {
			c.log.Warn().Err(err).Msg("Failed to cache IPO calendar")
		}
	}

	return events, nil
}

// =============================================================================
// Parsing Functions
// =============================================================================

func parseCompanyOverview(body []byte) (*CompanyOverview, error) {
	var raw map[string]string
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	if len(raw) == 0 {
		return nil, fmt.Errorf("empty overview response")
	}

	return &CompanyOverview{
		Symbol:                     raw["Symbol"],
		AssetType:                  raw["AssetType"],
		Name:                       raw["Name"],
		Description:                raw["Description"],
		CIK:                        raw["CIK"],
		Exchange:                   raw["Exchange"],
		Currency:                   raw["Currency"],
		Country:                    raw["Country"],
		Sector:                     raw["Sector"],
		Industry:                   raw["Industry"],
		Address:                    raw["Address"],
		FullTimeEmployees:          parseInt64(raw["FullTimeEmployees"]),
		FiscalYearEnd:              raw["FiscalYearEnd"],
		LatestQuarter:              raw["LatestQuarter"],
		MarketCapitalization:       parseInt64(raw["MarketCapitalization"]),
		EBITDA:                     parseInt64(raw["EBITDA"]),
		PERatio:                    parseFloat64Ptr(raw["PERatio"]),
		PEGRatio:                   parseFloat64Ptr(raw["PEGRatio"]),
		BookValue:                  parseFloat64Ptr(raw["BookValue"]),
		DividendPerShare:           parseFloat64Ptr(raw["DividendPerShare"]),
		DividendYield:              parseFloat64Ptr(raw["DividendYield"]),
		EPS:                        parseFloat64Ptr(raw["EPS"]),
		RevenuePerShareTTM:         parseFloat64Ptr(raw["RevenuePerShareTTM"]),
		ProfitMargin:               parseFloat64Ptr(raw["ProfitMargin"]),
		OperatingMarginTTM:         parseFloat64Ptr(raw["OperatingMarginTTM"]),
		ReturnOnAssetsTTM:          parseFloat64Ptr(raw["ReturnOnAssetsTTM"]),
		ReturnOnEquityTTM:          parseFloat64Ptr(raw["ReturnOnEquityTTM"]),
		RevenueTTM:                 parseInt64(raw["RevenueTTM"]),
		GrossProfitTTM:             parseInt64(raw["GrossProfitTTM"]),
		DilutedEPSTTM:              parseFloat64Ptr(raw["DilutedEPSTTM"]),
		QuarterlyEarningsGrowthYOY: parseFloat64Ptr(raw["QuarterlyEarningsGrowthYOY"]),
		QuarterlyRevenueGrowthYOY:  parseFloat64Ptr(raw["QuarterlyRevenueGrowthYOY"]),
		AnalystTargetPrice:         parseFloat64Ptr(raw["AnalystTargetPrice"]),
		AnalystRatingStrongBuy:     parseInt(raw["AnalystRatingStrongBuy"]),
		AnalystRatingBuy:           parseInt(raw["AnalystRatingBuy"]),
		AnalystRatingHold:          parseInt(raw["AnalystRatingHold"]),
		AnalystRatingSell:          parseInt(raw["AnalystRatingSell"]),
		AnalystRatingStrongSell:    parseInt(raw["AnalystRatingStrongSell"]),
		TrailingPE:                 parseFloat64Ptr(raw["TrailingPE"]),
		ForwardPE:                  parseFloat64Ptr(raw["ForwardPE"]),
		PriceToSalesRatioTTM:       parseFloat64Ptr(raw["PriceToSalesRatioTTM"]),
		PriceToBookRatio:           parseFloat64Ptr(raw["PriceToBookRatio"]),
		EVToRevenue:                parseFloat64Ptr(raw["EVToRevenue"]),
		EVToEBITDA:                 parseFloat64Ptr(raw["EVToEBITDA"]),
		Beta:                       parseFloat64Ptr(raw["Beta"]),
		FiftyTwoWeekHigh:           parseFloat64Ptr(raw["52WeekHigh"]),
		FiftyTwoWeekLow:            parseFloat64Ptr(raw["52WeekLow"]),
		FiftyDayMovingAverage:      parseFloat64Ptr(raw["50DayMovingAverage"]),
		TwoHundredDayMovingAverage: parseFloat64Ptr(raw["200DayMovingAverage"]),
		SharesOutstanding:          parseInt64(raw["SharesOutstanding"]),
		DividendDate:               raw["DividendDate"],
		ExDividendDate:             raw["ExDividendDate"],
	}, nil
}

func parseEarnings(body []byte) (*Earnings, error) {
	var response struct {
		Symbol            string              `json:"symbol"`
		AnnualEarnings    []map[string]string `json:"annualEarnings"`
		QuarterlyEarnings []map[string]string `json:"quarterlyEarnings"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	earnings := &Earnings{
		Symbol:            response.Symbol,
		AnnualEarnings:    make([]EarningsReport, 0, len(response.AnnualEarnings)),
		QuarterlyEarnings: make([]EarningsReport, 0, len(response.QuarterlyEarnings)),
	}

	for _, e := range response.AnnualEarnings {
		earnings.AnnualEarnings = append(earnings.AnnualEarnings, EarningsReport{
			FiscalDateEnding: e["fiscalDateEnding"],
			ReportedEPS:      parseFloat64Ptr(e["reportedEPS"]),
		})
	}

	for _, e := range response.QuarterlyEarnings {
		earnings.QuarterlyEarnings = append(earnings.QuarterlyEarnings, EarningsReport{
			FiscalDateEnding:   e["fiscalDateEnding"],
			ReportedDate:       e["reportedDate"],
			ReportedEPS:        parseFloat64Ptr(e["reportedEPS"]),
			EstimatedEPS:       parseFloat64Ptr(e["estimatedEPS"]),
			Surprise:           parseFloat64Ptr(e["surprise"]),
			SurprisePercentage: parseFloat64Ptr(e["surprisePercentage"]),
		})
	}

	return earnings, nil
}

func parseIncomeStatement(body []byte) (*IncomeStatement, error) {
	var response struct {
		Symbol           string              `json:"symbol"`
		AnnualReports    []map[string]string `json:"annualReports"`
		QuarterlyReports []map[string]string `json:"quarterlyReports"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	stmt := &IncomeStatement{
		Symbol:           response.Symbol,
		AnnualReports:    make([]IncomeStatementReport, 0, len(response.AnnualReports)),
		QuarterlyReports: make([]IncomeStatementReport, 0, len(response.QuarterlyReports)),
	}

	for _, r := range response.AnnualReports {
		stmt.AnnualReports = append(stmt.AnnualReports, parseIncomeStatementReport(r))
	}

	for _, r := range response.QuarterlyReports {
		stmt.QuarterlyReports = append(stmt.QuarterlyReports, parseIncomeStatementReport(r))
	}

	return stmt, nil
}

func parseIncomeStatementReport(r map[string]string) IncomeStatementReport {
	return IncomeStatementReport{
		FiscalDateEnding:                  r["fiscalDateEnding"],
		ReportedCurrency:                  r["reportedCurrency"],
		GrossProfit:                       parseInt64(r["grossProfit"]),
		TotalRevenue:                      parseInt64(r["totalRevenue"]),
		CostOfRevenue:                     parseInt64(r["costOfRevenue"]),
		CostOfGoodsAndServicesSold:        parseInt64(r["costofGoodsAndServicesSold"]),
		OperatingIncome:                   parseInt64(r["operatingIncome"]),
		SellingGeneralAndAdministrative:   parseInt64(r["sellingGeneralAndAdministrative"]),
		ResearchAndDevelopment:            parseInt64(r["researchAndDevelopment"]),
		OperatingExpenses:                 parseInt64(r["operatingExpenses"]),
		InvestmentIncomeNet:               parseInt64(r["investmentIncomeNet"]),
		NetInterestIncome:                 parseInt64(r["netInterestIncome"]),
		InterestIncome:                    parseInt64(r["interestIncome"]),
		InterestExpense:                   parseInt64(r["interestExpense"]),
		NonInterestIncome:                 parseInt64(r["nonInterestIncome"]),
		OtherNonOperatingIncome:           parseInt64(r["otherNonOperatingIncome"]),
		Depreciation:                      parseInt64(r["depreciation"]),
		DepreciationAndAmortization:       parseInt64(r["depreciationAndAmortization"]),
		IncomeBeforeTax:                   parseInt64(r["incomeBeforeTax"]),
		IncomeTaxExpense:                  parseInt64(r["incomeTaxExpense"]),
		InterestAndDebtExpense:            parseInt64(r["interestAndDebtExpense"]),
		NetIncomeFromContinuingOperations: parseInt64(r["netIncomeFromContinuingOperations"]),
		ComprehensiveIncomeNetOfTax:       parseInt64(r["comprehensiveIncomeNetOfTax"]),
		EBIT:                              parseInt64(r["ebit"]),
		EBITDA:                            parseInt64(r["ebitda"]),
		NetIncome:                         parseInt64(r["netIncome"]),
	}
}

func parseBalanceSheet(body []byte) (*BalanceSheet, error) {
	var response struct {
		Symbol           string              `json:"symbol"`
		AnnualReports    []map[string]string `json:"annualReports"`
		QuarterlyReports []map[string]string `json:"quarterlyReports"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	sheet := &BalanceSheet{
		Symbol:           response.Symbol,
		AnnualReports:    make([]BalanceSheetReport, 0, len(response.AnnualReports)),
		QuarterlyReports: make([]BalanceSheetReport, 0, len(response.QuarterlyReports)),
	}

	for _, r := range response.AnnualReports {
		sheet.AnnualReports = append(sheet.AnnualReports, parseBalanceSheetReport(r))
	}

	for _, r := range response.QuarterlyReports {
		sheet.QuarterlyReports = append(sheet.QuarterlyReports, parseBalanceSheetReport(r))
	}

	return sheet, nil
}

func parseBalanceSheetReport(r map[string]string) BalanceSheetReport {
	return BalanceSheetReport{
		FiscalDateEnding:                       r["fiscalDateEnding"],
		ReportedCurrency:                       r["reportedCurrency"],
		TotalAssets:                            parseInt64(r["totalAssets"]),
		TotalCurrentAssets:                     parseInt64(r["totalCurrentAssets"]),
		CashAndCashEquivalentsAtCarryingValue:  parseInt64(r["cashAndCashEquivalentsAtCarryingValue"]),
		CashAndShortTermInvestments:            parseInt64(r["cashAndShortTermInvestments"]),
		Inventory:                              parseInt64(r["inventory"]),
		CurrentNetReceivables:                  parseInt64(r["currentNetReceivables"]),
		TotalNonCurrentAssets:                  parseInt64(r["totalNonCurrentAssets"]),
		PropertyPlantEquipment:                 parseInt64(r["propertyPlantEquipment"]),
		AccumulatedDepreciationAmortizationPPE: parseInt64(r["accumulatedDepreciationAmortizationPPE"]),
		IntangibleAssets:                       parseInt64(r["intangibleAssets"]),
		IntangibleAssetsExcludingGoodwill:      parseInt64(r["intangibleAssetsExcludingGoodwill"]),
		Goodwill:                               parseInt64(r["goodwill"]),
		Investments:                            parseInt64(r["investments"]),
		LongTermInvestments:                    parseInt64(r["longTermInvestments"]),
		ShortTermInvestments:                   parseInt64(r["shortTermInvestments"]),
		OtherCurrentAssets:                     parseInt64(r["otherCurrentAssets"]),
		OtherNonCurrentAssets:                  parseInt64(r["otherNonCurrentAssets"]),
		TotalLiabilities:                       parseInt64(r["totalLiabilities"]),
		TotalCurrentLiabilities:                parseInt64(r["totalCurrentLiabilities"]),
		CurrentAccountsPayable:                 parseInt64(r["currentAccountsPayable"]),
		DeferredRevenue:                        parseInt64(r["deferredRevenue"]),
		CurrentDebt:                            parseInt64(r["currentDebt"]),
		ShortTermDebt:                          parseInt64(r["shortTermDebt"]),
		TotalNonCurrentLiabilities:             parseInt64(r["totalNonCurrentLiabilities"]),
		CapitalLeaseObligations:                parseInt64(r["capitalLeaseObligations"]),
		LongTermDebt:                           parseInt64(r["longTermDebt"]),
		CurrentLongTermDebt:                    parseInt64(r["currentLongTermDebt"]),
		LongTermDebtNoncurrent:                 parseInt64(r["longTermDebtNoncurrent"]),
		ShortLongTermDebtTotal:                 parseInt64(r["shortLongTermDebtTotal"]),
		OtherCurrentLiabilities:                parseInt64(r["otherCurrentLiabilities"]),
		OtherNonCurrentLiabilities:             parseInt64(r["otherNonCurrentLiabilities"]),
		TotalShareholderEquity:                 parseInt64(r["totalShareholderEquity"]),
		TreasuryStock:                          parseInt64(r["treasuryStock"]),
		RetainedEarnings:                       parseInt64(r["retainedEarnings"]),
		CommonStock:                            parseInt64(r["commonStock"]),
		CommonStockSharesOutstanding:           parseInt64(r["commonStockSharesOutstanding"]),
	}
}

func parseCashFlow(body []byte) (*CashFlow, error) {
	var response struct {
		Symbol           string              `json:"symbol"`
		AnnualReports    []map[string]string `json:"annualReports"`
		QuarterlyReports []map[string]string `json:"quarterlyReports"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	cf := &CashFlow{
		Symbol:           response.Symbol,
		AnnualReports:    make([]CashFlowReport, 0, len(response.AnnualReports)),
		QuarterlyReports: make([]CashFlowReport, 0, len(response.QuarterlyReports)),
	}

	for _, r := range response.AnnualReports {
		cf.AnnualReports = append(cf.AnnualReports, parseCashFlowReport(r))
	}

	for _, r := range response.QuarterlyReports {
		cf.QuarterlyReports = append(cf.QuarterlyReports, parseCashFlowReport(r))
	}

	return cf, nil
}

func parseCashFlowReport(r map[string]string) CashFlowReport {
	return CashFlowReport{
		FiscalDateEnding:                                   r["fiscalDateEnding"],
		ReportedCurrency:                                   r["reportedCurrency"],
		OperatingCashflow:                                  parseInt64(r["operatingCashflow"]),
		PaymentsForOperatingActivities:                     parseInt64(r["paymentsForOperatingActivities"]),
		ProceedsFromOperatingActivities:                    parseInt64(r["proceedsFromOperatingActivities"]),
		ChangeInOperatingLiabilities:                       parseInt64(r["changeInOperatingLiabilities"]),
		ChangeInOperatingAssets:                            parseInt64(r["changeInOperatingAssets"]),
		DepreciationDepletionAndAmortization:               parseInt64(r["depreciationDepletionAndAmortization"]),
		CapitalExpenditures:                                parseInt64(r["capitalExpenditures"]),
		ChangeInReceivables:                                parseInt64(r["changeInReceivables"]),
		ChangeInInventory:                                  parseInt64(r["changeInInventory"]),
		ProfitLoss:                                         parseInt64(r["profitLoss"]),
		CashflowFromInvestment:                             parseInt64(r["cashflowFromInvestment"]),
		CashflowFromFinancing:                              parseInt64(r["cashflowFromFinancing"]),
		ProceedsFromRepaymentsOfShortTermDebt:              parseInt64(r["proceedsFromRepaymentsOfShortTermDebt"]),
		PaymentsForRepurchaseOfCommonStock:                 parseInt64(r["paymentsForRepurchaseOfCommonStock"]),
		PaymentsForRepurchaseOfEquity:                      parseInt64(r["paymentsForRepurchaseOfEquity"]),
		PaymentsForRepurchaseOfPreferredStock:              parseInt64(r["paymentsForRepurchaseOfPreferredStock"]),
		DividendPayout:                                     parseInt64(r["dividendPayout"]),
		DividendPayoutCommonStock:                          parseInt64(r["dividendPayoutCommonStock"]),
		DividendPayoutPreferredStock:                       parseInt64(r["dividendPayoutPreferredStock"]),
		ProceedsFromIssuanceOfCommonStock:                  parseInt64(r["proceedsFromIssuanceOfCommonStock"]),
		ProceedsFromIssuanceOfLongTermDebtAndCapitalSecNet: parseInt64(r["proceedsFromIssuanceOfLongTermDebtAndCapitalSecuritiesNet"]),
		ProceedsFromIssuanceOfPreferredStock:               parseInt64(r["proceedsFromIssuanceOfPreferredStock"]),
		ProceedsFromRepurchaseOfEquity:                     parseInt64(r["proceedsFromRepurchaseOfEquity"]),
		ProceedsFromSaleOfTreasuryStock:                    parseInt64(r["proceedsFromSaleOfTreasuryStock"]),
		ChangeInCashAndCashEquivalents:                     parseInt64(r["changeInCashAndCashEquivalents"]),
		ChangeInExchangeRate:                               parseInt64(r["changeInExchangeRate"]),
		NetIncome:                                          parseInt64(r["netIncome"]),
	}
}

func parseDividends(body []byte) ([]DividendRecord, error) {
	var response struct {
		Data []map[string]string `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	dividends := make([]DividendRecord, 0, len(response.Data))
	for _, d := range response.Data {
		dividends = append(dividends, DividendRecord{
			ExDate:         parseDate(d["ex_dividend_date"]),
			PaymentDate:    parseDate(d["payment_date"]),
			RecordDate:     parseDate(d["record_date"]),
			DeclaredDate:   parseDate(d["declaration_date"]),
			Amount:         parseFloat64(d["amount"]),
			AdjustedAmount: parseFloat64(d["adjusted_amount"]),
		})
	}

	sort.Slice(dividends, func(i, j int) bool {
		return dividends[i].ExDate.After(dividends[j].ExDate)
	})

	return dividends, nil
}

func parseSplits(body []byte) ([]SplitRecord, error) {
	var response struct {
		Data []map[string]string `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	splits := make([]SplitRecord, 0, len(response.Data))
	for _, s := range response.Data {
		splits = append(splits, SplitRecord{
			EffectiveDate:    parseDate(s["effective_date"]),
			SplitCoefficient: parseFloat64(s["split_coefficient"]),
		})
	}

	sort.Slice(splits, func(i, j int) bool {
		return splits[i].EffectiveDate.After(splits[j].EffectiveDate)
	})

	return splits, nil
}

func parseETFProfile(body []byte) (*ETFProfile, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	getString := func(key string) string {
		if v, ok := raw[key].(string); ok {
			return v
		}
		return ""
	}

	getFloat := func(key string) float64 {
		switch v := raw[key].(type) {
		case float64:
			return v
		case string:
			return parseFloat64(v)
		}
		return 0
	}

	getFloatPtr := func(key string) *float64 {
		switch v := raw[key].(type) {
		case float64:
			return &v
		case string:
			return parseFloat64Ptr(v)
		}
		return nil
	}

	getInt := func(key string) int {
		switch v := raw[key].(type) {
		case float64:
			return int(v)
		case string:
			return parseInt(v)
		}
		return 0
	}

	getInt64 := func(key string) int64 {
		switch v := raw[key].(type) {
		case float64:
			return int64(v)
		case string:
			return parseInt64(v)
		}
		return 0
	}

	return &ETFProfile{
		Symbol:          getString("symbol"),
		AssetType:       getString("asset_type"),
		Name:            getString("name"),
		Description:     getString("description"),
		InceptionDate:   parseDate(getString("inception_date")),
		Exchange:        getString("exchange"),
		AssetClass:      getString("asset_class"),
		ExpenseRatio:    getFloat("expense_ratio"),
		DividendYield:   getFloat("dividend_yield"),
		TotalAssets:     getInt64("total_assets"),
		HoldingsCount:   getInt("holdings_count"),
		TurnoverRatio:   getFloat("turnover_ratio"),
		PriceToEarnings: getFloatPtr("price_to_earnings"),
		PriceToBook:     getFloatPtr("price_to_book"),
	}, nil
}

func parseETFHoldings(body []byte) ([]ETFHolding, error) {
	var response struct {
		Holdings []map[string]interface{} `json:"holdings"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	holdings := make([]ETFHolding, 0, len(response.Holdings))
	for _, h := range response.Holdings {
		getString := func(key string) string {
			if v, ok := h[key].(string); ok {
				return v
			}
			return ""
		}

		getFloat := func(key string) float64 {
			switch v := h[key].(type) {
			case float64:
				return v
			case string:
				return parseFloat64(v)
			}
			return 0
		}

		getInt64 := func(key string) int64 {
			switch v := h[key].(type) {
			case float64:
				return int64(v)
			case string:
				return parseInt64(v)
			}
			return 0
		}

		holdings = append(holdings, ETFHolding{
			Symbol:      getString("symbol"),
			Name:        getString("name"),
			Weight:      getFloat("weight"),
			SharesHeld:  getInt64("shares"),
			MarketValue: getInt64("market_value"),
		})
	}

	return holdings, nil
}

func parseSharesOutstanding(body []byte) ([]SharesOutstandingRecord, error) {
	var response struct {
		Data []map[string]string `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	records := make([]SharesOutstandingRecord, 0, len(response.Data))
	for _, r := range response.Data {
		records = append(records, SharesOutstandingRecord{
			Date:   parseDate(r["date"]),
			Shares: parseInt64(r["shares_outstanding"]),
		})
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Date.After(records[j].Date)
	})

	return records, nil
}

func parseListingStatus(body []byte) ([]ListingStatus, error) {
	// Listing status returns CSV data
	reader := csv.NewReader(strings.NewReader(string(body)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return []ListingStatus{}, nil
	}

	listings := make([]ListingStatus, 0, len(records)-1)
	for _, r := range records[1:] { // Skip header row
		if len(r) < 6 {
			continue
		}
		listings = append(listings, ListingStatus{
			Symbol:     r[0],
			Name:       r[1],
			Exchange:   r[2],
			AssetType:  r[3],
			IPODate:    r[4],
			DelistDate: r[5],
			Status:     "active",
		})
	}

	return listings, nil
}

func parseEarningsCalendar(body []byte) ([]EarningsEvent, error) {
	// Earnings calendar returns CSV data
	reader := csv.NewReader(strings.NewReader(string(body)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return []EarningsEvent{}, nil
	}

	events := make([]EarningsEvent, 0, len(records)-1)
	for _, r := range records[1:] { // Skip header row
		if len(r) < 5 {
			continue
		}
		events = append(events, EarningsEvent{
			Symbol:           r[0],
			Name:             r[1],
			ReportDate:       parseDate(r[2]),
			FiscalDateEnding: r[3],
			Estimate:         parseFloat64Ptr(r[4]),
			Currency:         "USD", // Default
		})
	}

	return events, nil
}

func parseIPOCalendar(body []byte) ([]IPOEvent, error) {
	// IPO calendar returns CSV data
	reader := csv.NewReader(strings.NewReader(string(body)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return []IPOEvent{}, nil
	}

	events := make([]IPOEvent, 0, len(records)-1)
	for _, r := range records[1:] { // Skip header row
		if len(r) < 6 {
			continue
		}
		events = append(events, IPOEvent{
			Symbol:         r[0],
			Name:           r[1],
			IPODate:        parseDate(r[2]),
			PriceRangeLow:  parseFloat64Ptr(r[3]),
			PriceRangeHigh: parseFloat64Ptr(r[4]),
			Currency:       r[5],
			Exchange:       "",
		})
	}

	return events, nil
}
