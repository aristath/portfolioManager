package alphavantage

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/aristath/sentinel/internal/clientdata"
)

// =============================================================================
// Technical Indicator Endpoints - Generic
// =============================================================================

// GetIndicator fetches any technical indicator with custom parameters.
func (c *Client) GetIndicator(function, symbol, interval string, params map[string]string) (*IndicatorData, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	// Build composite cache key: isin:function:interval:params
	var keyParts []string
	keyParts = append(keyParts, isin, function, interval)
	// Sort params for consistent cache keys
	paramKeys := make([]string, 0, len(params))
	for k := range params {
		paramKeys = append(paramKeys, k)
	}
	sort.Strings(paramKeys)
	for _, k := range paramKeys {
		keyParts = append(keyParts, k+"="+params[k])
	}
	cacheKey := strings.Join(keyParts, ":")

	// Check cache (using current_prices table for technical indicators)
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var indicatorData IndicatorData
			if err := json.Unmarshal(data, &indicatorData); err == nil {
				return &indicatorData, nil
			}
		}
	}

	// Fetch from API
	allParams := map[string]string{
		"symbol":   symbol,
		"interval": interval,
	}
	for k, v := range params {
		allParams[k] = v
	}

	body, err := c.doRequest(function, allParams)
	if err != nil {
		return nil, err
	}

	data, err := parseTechnicalIndicator(body, function)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", function, err)
	}

	data.Symbol = symbol
	data.Interval = interval

	// Store in cache (using 1 hour TTL for technical indicators - need to add constant)
	// For now using TTLCurrentPrice, but should have a TTLTechnicalIndicator constant
	if c.cacheRepo != nil {
		ttl := clientdata.TTLCurrentPrice // Technical indicators change frequently (10 minutes)
		if err := c.cacheRepo.Store(table, cacheKey, data, ttl); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Str("function", function).Msg("Failed to cache technical indicator")
		}
	}

	return data, nil
}

// =============================================================================
// Moving Averages
// =============================================================================

// GetSMA returns Simple Moving Average values.
func (c *Client) GetSMA(symbol, interval string, period int, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("SMA", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	})
}

// GetEMA returns Exponential Moving Average values.
func (c *Client) GetEMA(symbol, interval string, period int, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("EMA", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	})
}

// GetWMA returns Weighted Moving Average values.
func (c *Client) GetWMA(symbol, interval string, period int, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("WMA", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	})
}

// GetDEMA returns Double Exponential Moving Average values.
func (c *Client) GetDEMA(symbol, interval string, period int, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("DEMA", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	})
}

// GetTEMA returns Triple Exponential Moving Average values.
func (c *Client) GetTEMA(symbol, interval string, period int, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("TEMA", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	})
}

// GetTRIMA returns Triangular Moving Average values.
func (c *Client) GetTRIMA(symbol, interval string, period int, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("TRIMA", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	})
}

// GetKAMA returns Kaufman Adaptive Moving Average values.
func (c *Client) GetKAMA(symbol, interval string, period int, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("KAMA", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	})
}

// GetT3 returns Triple Exponential Moving Average (T3) values.
func (c *Client) GetT3(symbol, interval string, period int, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("T3", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	})
}

// GetVWAP returns Volume Weighted Average Price values.
func (c *Client) GetVWAP(symbol, interval string) (*IndicatorData, error) {
	return c.GetIndicator("VWAP", symbol, interval, nil)
}

// =============================================================================
// Momentum Indicators
// =============================================================================

// GetRSI returns Relative Strength Index values.
func (c *Client) GetRSI(symbol, interval string, period int, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("RSI", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	})
}

// GetMACD returns Moving Average Convergence Divergence values.
func (c *Client) GetMACD(symbol, interval string, seriesType string) (*MACDData, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	// Build cache key
	cacheKey := isin + ":MACD:" + interval + ":series_type=" + seriesType

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var macdData MACDData
			if err := json.Unmarshal(data, &macdData); err == nil {
				return &macdData, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{
		"symbol":      symbol,
		"interval":    interval,
		"series_type": seriesType,
	}

	body, err := c.doRequest("MACD", params)
	if err != nil {
		return nil, err
	}

	data, err := parseMACD(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MACD: %w", err)
	}

	data.Symbol = symbol
	data.Interval = interval

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, data, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache MACD")
		}
	}

	return data, nil
}

// GetMACDEXT returns MACD with controllable MA types.
func (c *Client) GetMACDEXT(symbol, interval string, seriesType string, params map[string]string) (*MACDData, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	// Build cache key
	var keyParts []string
	keyParts = append(keyParts, isin, "MACDEXT", interval, "series_type="+seriesType)
	paramKeys := make([]string, 0, len(params))
	for k := range params {
		paramKeys = append(paramKeys, k)
	}
	sort.Strings(paramKeys)
	for _, k := range paramKeys {
		keyParts = append(keyParts, k+"="+params[k])
	}
	cacheKey := strings.Join(keyParts, ":")

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var macdData MACDData
			if err := json.Unmarshal(data, &macdData); err == nil {
				return &macdData, nil
			}
		}
	}

	// Fetch from API
	allParams := map[string]string{
		"symbol":      symbol,
		"interval":    interval,
		"series_type": seriesType,
	}
	for k, v := range params {
		allParams[k] = v
	}

	body, err := c.doRequest("MACDEXT", allParams)
	if err != nil {
		return nil, err
	}

	data, err := parseMACD(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MACDEXT: %w", err)
	}

	data.Symbol = symbol
	data.Interval = interval

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, data, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache MACDEXT")
		}
	}

	return data, nil
}

// GetSTOCH returns Stochastic Oscillator values.
func (c *Client) GetSTOCH(symbol, interval string) (*StochData, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	cacheKey := isin + ":STOCH:" + interval

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var stochData StochData
			if err := json.Unmarshal(data, &stochData); err == nil {
				return &stochData, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{
		"symbol":   symbol,
		"interval": interval,
	}

	body, err := c.doRequest("STOCH", params)
	if err != nil {
		return nil, err
	}

	data, err := parseSTOCH(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse STOCH: %w", err)
	}

	data.Symbol = symbol
	data.Interval = interval

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, data, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache STOCH")
		}
	}

	return data, nil
}

// GetSTOCHF returns Fast Stochastic Oscillator values.
func (c *Client) GetSTOCHF(symbol, interval string) (*StochData, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	cacheKey := isin + ":STOCHF:" + interval

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var stochData StochData
			if err := json.Unmarshal(data, &stochData); err == nil {
				return &stochData, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{
		"symbol":   symbol,
		"interval": interval,
	}

	body, err := c.doRequest("STOCHF", params)
	if err != nil {
		return nil, err
	}

	data, err := parseSTOCHF(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse STOCHF: %w", err)
	}

	data.Symbol = symbol
	data.Interval = interval

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, data, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache STOCHF")
		}
	}

	return data, nil
}

// GetSTOCHRSI returns Stochastic RSI values.
func (c *Client) GetSTOCHRSI(symbol, interval string, period int, seriesType string) (*StochData, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	cacheKey := isin + ":STOCHRSI:" + interval + ":period=" + strconv.Itoa(period) + ":series_type=" + seriesType

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var stochData StochData
			if err := json.Unmarshal(data, &stochData); err == nil {
				return &stochData, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{
		"symbol":      symbol,
		"interval":    interval,
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	}

	body, err := c.doRequest("STOCHRSI", params)
	if err != nil {
		return nil, err
	}

	data, err := parseSTOCHRSI(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse STOCHRSI: %w", err)
	}

	data.Symbol = symbol
	data.Interval = interval

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, data, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache STOCHRSI")
		}
	}

	return data, nil
}

// GetWILLR returns Williams' %R values.
func (c *Client) GetWILLR(symbol, interval string, period int) (*IndicatorData, error) {
	return c.GetIndicator("WILLR", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
	})
}

// GetROC returns Rate of Change values.
func (c *Client) GetROC(symbol, interval string, period int, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("ROC", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	})
}

// GetROCR returns Rate of Change Ratio values.
func (c *Client) GetROCR(symbol, interval string, period int, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("ROCR", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	})
}

// GetMOM returns Momentum values.
func (c *Client) GetMOM(symbol, interval string, period int, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("MOM", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	})
}

// GetMFI returns Money Flow Index values.
func (c *Client) GetMFI(symbol, interval string, period int) (*IndicatorData, error) {
	return c.GetIndicator("MFI", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
	})
}

// GetCMO returns Chande Momentum Oscillator values.
func (c *Client) GetCMO(symbol, interval string, period int, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("CMO", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	})
}

// GetULTOSC returns Ultimate Oscillator values.
func (c *Client) GetULTOSC(symbol, interval string) (*IndicatorData, error) {
	return c.GetIndicator("ULTOSC", symbol, interval, nil)
}

// GetTRIX returns TRIX values.
func (c *Client) GetTRIX(symbol, interval string, period int, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("TRIX", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	})
}

// GetAPO returns Absolute Price Oscillator values.
func (c *Client) GetAPO(symbol, interval string, seriesType string, fastPeriod, slowPeriod int) (*IndicatorData, error) {
	return c.GetIndicator("APO", symbol, interval, map[string]string{
		"series_type": seriesType,
		"fastperiod":  strconv.Itoa(fastPeriod),
		"slowperiod":  strconv.Itoa(slowPeriod),
	})
}

// GetPPO returns Percentage Price Oscillator values.
func (c *Client) GetPPO(symbol, interval string, seriesType string, fastPeriod, slowPeriod int) (*IndicatorData, error) {
	return c.GetIndicator("PPO", symbol, interval, map[string]string{
		"series_type": seriesType,
		"fastperiod":  strconv.Itoa(fastPeriod),
		"slowperiod":  strconv.Itoa(slowPeriod),
	})
}

// GetBOP returns Balance of Power values.
func (c *Client) GetBOP(symbol, interval string) (*IndicatorData, error) {
	return c.GetIndicator("BOP", symbol, interval, nil)
}

// GetCCI returns Commodity Channel Index values.
func (c *Client) GetCCI(symbol, interval string, period int) (*IndicatorData, error) {
	return c.GetIndicator("CCI", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
	})
}

// =============================================================================
// Trend Indicators
// =============================================================================

// GetADX returns Average Directional Movement Index values.
func (c *Client) GetADX(symbol, interval string, period int) (*IndicatorData, error) {
	return c.GetIndicator("ADX", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
	})
}

// GetADXR returns Average Directional Movement Index Rating values.
func (c *Client) GetADXR(symbol, interval string, period int) (*IndicatorData, error) {
	return c.GetIndicator("ADXR", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
	})
}

// GetDX returns Directional Movement Index values.
func (c *Client) GetDX(symbol, interval string, period int) (*IndicatorData, error) {
	return c.GetIndicator("DX", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
	})
}

// GetPLUS_DI returns Plus Directional Indicator values.
func (c *Client) GetPLUS_DI(symbol, interval string, period int) (*IndicatorData, error) {
	return c.GetIndicator("PLUS_DI", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
	})
}

// GetMINUS_DI returns Minus Directional Indicator values.
func (c *Client) GetMINUS_DI(symbol, interval string, period int) (*IndicatorData, error) {
	return c.GetIndicator("MINUS_DI", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
	})
}

// GetPLUS_DM returns Plus Directional Movement values.
func (c *Client) GetPLUS_DM(symbol, interval string, period int) (*IndicatorData, error) {
	return c.GetIndicator("PLUS_DM", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
	})
}

// GetMINUS_DM returns Minus Directional Movement values.
func (c *Client) GetMINUS_DM(symbol, interval string, period int) (*IndicatorData, error) {
	return c.GetIndicator("MINUS_DM", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
	})
}

// GetAROON returns Aroon indicator values.
func (c *Client) GetAROON(symbol, interval string, period int) (*AroonData, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	cacheKey := isin + ":AROON:" + interval + ":period=" + strconv.Itoa(period)

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var aroonData AroonData
			if err := json.Unmarshal(data, &aroonData); err == nil {
				return &aroonData, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{
		"symbol":      symbol,
		"interval":    interval,
		"time_period": strconv.Itoa(period),
	}

	body, err := c.doRequest("AROON", params)
	if err != nil {
		return nil, err
	}

	data, err := parseAROON(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AROON: %w", err)
	}

	data.Symbol = symbol
	data.Interval = interval

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, data, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache AROON")
		}
	}

	return data, nil
}

// GetAROONOSC returns Aroon Oscillator values.
func (c *Client) GetAROONOSC(symbol, interval string, period int) (*IndicatorData, error) {
	return c.GetIndicator("AROONOSC", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
	})
}

// GetSAR returns Parabolic SAR values.
func (c *Client) GetSAR(symbol, interval string, acceleration, maximum float64) (*IndicatorData, error) {
	return c.GetIndicator("SAR", symbol, interval, map[string]string{
		"acceleration": strconv.FormatFloat(acceleration, 'f', -1, 64),
		"maximum":      strconv.FormatFloat(maximum, 'f', -1, 64),
	})
}

// =============================================================================
// Volatility Indicators
// =============================================================================

// GetBBANDS returns Bollinger Bands values.
func (c *Client) GetBBANDS(symbol, interval string, period int, seriesType string, nbdevup, nbdevdn float64) (*BollingerData, error) {
	// Resolve symbol to ISIN
	isin, err := c.resolveISIN(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ISIN for symbol %s: %w", symbol, err)
	}

	cacheKey := isin + ":BBANDS:" + interval + ":period=" + strconv.Itoa(period) + ":series_type=" + seriesType + ":nbdevup=" + strconv.FormatFloat(nbdevup, 'f', -1, 64) + ":nbdevdn=" + strconv.FormatFloat(nbdevdn, 'f', -1, 64)

	// Check cache
	table := "current_prices"
	if c.cacheRepo != nil {
		if data, err := c.cacheRepo.GetIfFresh(table, cacheKey); err == nil && data != nil {
			var bbandsData BollingerData
			if err := json.Unmarshal(data, &bbandsData); err == nil {
				return &bbandsData, nil
			}
		}
	}

	// Fetch from API
	params := map[string]string{
		"symbol":      symbol,
		"interval":    interval,
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
		"nbdevup":     strconv.FormatFloat(nbdevup, 'f', -1, 64),
		"nbdevdn":     strconv.FormatFloat(nbdevdn, 'f', -1, 64),
	}

	body, err := c.doRequest("BBANDS", params)
	if err != nil {
		return nil, err
	}

	data, err := parseBBANDS(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse BBANDS: %w", err)
	}

	data.Symbol = symbol
	data.Interval = interval

	// Store in cache
	if c.cacheRepo != nil {
		if err := c.cacheRepo.Store(table, cacheKey, data, clientdata.TTLCurrentPrice); err != nil {
			c.log.Warn().Err(err).Str("isin", isin).Msg("Failed to cache BBANDS")
		}
	}

	return data, nil
}

// GetATR returns Average True Range values.
func (c *Client) GetATR(symbol, interval string, period int) (*IndicatorData, error) {
	return c.GetIndicator("ATR", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
	})
}

// GetNATR returns Normalized Average True Range values.
func (c *Client) GetNATR(symbol, interval string, period int) (*IndicatorData, error) {
	return c.GetIndicator("NATR", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
	})
}

// GetTRANGE returns True Range values.
func (c *Client) GetTRANGE(symbol, interval string) (*IndicatorData, error) {
	return c.GetIndicator("TRANGE", symbol, interval, nil)
}

// GetMIDPOINT returns Midpoint values.
func (c *Client) GetMIDPOINT(symbol, interval string, period int, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("MIDPOINT", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
		"series_type": seriesType,
	})
}

// GetMIDPRICE returns Midpoint Price values.
func (c *Client) GetMIDPRICE(symbol, interval string, period int) (*IndicatorData, error) {
	return c.GetIndicator("MIDPRICE", symbol, interval, map[string]string{
		"time_period": strconv.Itoa(period),
	})
}

// =============================================================================
// Volume Indicators
// =============================================================================

// GetOBV returns On Balance Volume values.
func (c *Client) GetOBV(symbol, interval string) (*IndicatorData, error) {
	return c.GetIndicator("OBV", symbol, interval, nil)
}

// GetAD returns Chaikin A/D Line values.
func (c *Client) GetAD(symbol, interval string) (*IndicatorData, error) {
	return c.GetIndicator("AD", symbol, interval, nil)
}

// GetADOSC returns Chaikin A/D Oscillator values.
func (c *Client) GetADOSC(symbol, interval string, fastPeriod, slowPeriod int) (*IndicatorData, error) {
	return c.GetIndicator("ADOSC", symbol, interval, map[string]string{
		"fastperiod": strconv.Itoa(fastPeriod),
		"slowperiod": strconv.Itoa(slowPeriod),
	})
}

// =============================================================================
// Cycle Indicators (Hilbert Transform)
// =============================================================================

// GetHT_TRENDLINE returns Hilbert Transform - Instantaneous Trendline values.
func (c *Client) GetHT_TRENDLINE(symbol, interval string, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("HT_TRENDLINE", symbol, interval, map[string]string{
		"series_type": seriesType,
	})
}

// GetHT_SINE returns Hilbert Transform - Sine Wave values.
func (c *Client) GetHT_SINE(symbol, interval string, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("HT_SINE", symbol, interval, map[string]string{
		"series_type": seriesType,
	})
}

// GetHT_TRENDMODE returns Hilbert Transform - Trend vs Cycle Mode values.
func (c *Client) GetHT_TRENDMODE(symbol, interval string, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("HT_TRENDMODE", symbol, interval, map[string]string{
		"series_type": seriesType,
	})
}

// GetHT_DCPERIOD returns Hilbert Transform - Dominant Cycle Period values.
func (c *Client) GetHT_DCPERIOD(symbol, interval string, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("HT_DCPERIOD", symbol, interval, map[string]string{
		"series_type": seriesType,
	})
}

// GetHT_DCPHASE returns Hilbert Transform - Dominant Cycle Phase values.
func (c *Client) GetHT_DCPHASE(symbol, interval string, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("HT_DCPHASE", symbol, interval, map[string]string{
		"series_type": seriesType,
	})
}

// GetHT_PHASOR returns Hilbert Transform - Phasor Components values.
func (c *Client) GetHT_PHASOR(symbol, interval string, seriesType string) (*IndicatorData, error) {
	return c.GetIndicator("HT_PHASOR", symbol, interval, map[string]string{
		"series_type": seriesType,
	})
}

// =============================================================================
// Parsing Functions
// =============================================================================

func parseTechnicalIndicator(body []byte, function string) (*IndicatorData, error) {
	var rawResponse map[string]json.RawMessage
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, err
	}

	// Find the technical analysis key
	var technicalData map[string]map[string]string
	for key, value := range rawResponse {
		if key != "Meta Data" {
			if err := json.Unmarshal(value, &technicalData); err != nil {
				continue
			}
			break
		}
	}

	if technicalData == nil {
		return nil, fmt.Errorf("no technical indicator data found")
	}

	values := make([]IndicatorValue, 0, len(technicalData))
	for dateStr, data := range technicalData {
		date := parseDateTime(dateStr)
		// Get the first (and usually only) value from the data map
		var value float64
		for _, v := range data {
			value = parseFloat64(v)
			break
		}
		values = append(values, IndicatorValue{
			Date:  date,
			Value: value,
		})
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].Date.After(values[j].Date)
	})

	return &IndicatorData{Values: values}, nil
}

func parseMACD(body []byte) (*MACDData, error) {
	var rawResponse map[string]json.RawMessage
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, err
	}

	var technicalData map[string]map[string]string
	for key, value := range rawResponse {
		if key != "Meta Data" {
			if err := json.Unmarshal(value, &technicalData); err != nil {
				continue
			}
			break
		}
	}

	if technicalData == nil {
		return nil, fmt.Errorf("no MACD data found")
	}

	values := make([]MACDValue, 0, len(technicalData))
	for dateStr, data := range technicalData {
		date := parseDateTime(dateStr)
		values = append(values, MACDValue{
			Date:      date,
			MACD:      parseFloat64(data["MACD"]),
			Signal:    parseFloat64(data["MACD_Signal"]),
			Histogram: parseFloat64(data["MACD_Hist"]),
		})
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].Date.After(values[j].Date)
	})

	return &MACDData{Values: values}, nil
}

func parseSTOCH(body []byte) (*StochData, error) {
	var rawResponse map[string]json.RawMessage
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, err
	}

	var technicalData map[string]map[string]string
	for key, value := range rawResponse {
		if key != "Meta Data" {
			if err := json.Unmarshal(value, &technicalData); err != nil {
				continue
			}
			break
		}
	}

	if technicalData == nil {
		return nil, fmt.Errorf("no STOCH data found")
	}

	values := make([]StochValue, 0, len(technicalData))
	for dateStr, data := range technicalData {
		date := parseDateTime(dateStr)
		values = append(values, StochValue{
			Date:  date,
			SlowK: parseFloat64(data["SlowK"]),
			SlowD: parseFloat64(data["SlowD"]),
		})
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].Date.After(values[j].Date)
	})

	return &StochData{Values: values}, nil
}

func parseSTOCHF(body []byte) (*StochData, error) {
	var rawResponse map[string]json.RawMessage
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, err
	}

	var technicalData map[string]map[string]string
	for key, value := range rawResponse {
		if key != "Meta Data" {
			if err := json.Unmarshal(value, &technicalData); err != nil {
				continue
			}
			break
		}
	}

	if technicalData == nil {
		return nil, fmt.Errorf("no STOCHF data found")
	}

	values := make([]StochValue, 0, len(technicalData))
	for dateStr, data := range technicalData {
		date := parseDateTime(dateStr)
		values = append(values, StochValue{
			Date:  date,
			SlowK: parseFloat64(data["FastK"]),
			SlowD: parseFloat64(data["FastD"]),
		})
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].Date.After(values[j].Date)
	})

	return &StochData{Values: values}, nil
}

func parseSTOCHRSI(body []byte) (*StochData, error) {
	var rawResponse map[string]json.RawMessage
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, err
	}

	var technicalData map[string]map[string]string
	for key, value := range rawResponse {
		if key != "Meta Data" {
			if err := json.Unmarshal(value, &technicalData); err != nil {
				continue
			}
			break
		}
	}

	if technicalData == nil {
		return nil, fmt.Errorf("no STOCHRSI data found")
	}

	values := make([]StochValue, 0, len(technicalData))
	for dateStr, data := range technicalData {
		date := parseDateTime(dateStr)
		values = append(values, StochValue{
			Date:  date,
			SlowK: parseFloat64(data["FastK"]),
			SlowD: parseFloat64(data["FastD"]),
		})
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].Date.After(values[j].Date)
	})

	return &StochData{Values: values}, nil
}

func parseAROON(body []byte) (*AroonData, error) {
	var rawResponse map[string]json.RawMessage
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, err
	}

	var technicalData map[string]map[string]string
	for key, value := range rawResponse {
		if key != "Meta Data" {
			if err := json.Unmarshal(value, &technicalData); err != nil {
				continue
			}
			break
		}
	}

	if technicalData == nil {
		return nil, fmt.Errorf("no AROON data found")
	}

	values := make([]AroonValue, 0, len(technicalData))
	for dateStr, data := range technicalData {
		date := parseDateTime(dateStr)
		values = append(values, AroonValue{
			Date:      date,
			AroonUp:   parseFloat64(data["Aroon Up"]),
			AroonDown: parseFloat64(data["Aroon Down"]),
		})
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].Date.After(values[j].Date)
	})

	return &AroonData{Values: values}, nil
}

func parseBBANDS(body []byte) (*BollingerData, error) {
	var rawResponse map[string]json.RawMessage
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, err
	}

	var technicalData map[string]map[string]string
	for key, value := range rawResponse {
		if key != "Meta Data" {
			if err := json.Unmarshal(value, &technicalData); err != nil {
				continue
			}
			break
		}
	}

	if technicalData == nil {
		return nil, fmt.Errorf("no BBANDS data found")
	}

	values := make([]BollingerValue, 0, len(technicalData))
	for dateStr, data := range technicalData {
		date := parseDateTime(dateStr)
		values = append(values, BollingerValue{
			Date:       date,
			UpperBand:  parseFloat64(data["Real Upper Band"]),
			MiddleBand: parseFloat64(data["Real Middle Band"]),
			LowerBand:  parseFloat64(data["Real Lower Band"]),
		})
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].Date.After(values[j].Date)
	})

	return &BollingerData{Values: values}, nil
}
