package alphavantage

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTechnicalIndicator_SMA(t *testing.T) {
	jsonData := `{
		"Meta Data": {
			"1: Symbol": "IBM",
			"2: Indicator": "Simple Moving Average (SMA)"
		},
		"Technical Analysis: SMA": {
			"2024-01-15": {"SMA": "185.5000"},
			"2024-01-12": {"SMA": "184.2500"},
			"2024-01-11": {"SMA": "183.7500"}
		}
	}`

	data, err := parseTechnicalIndicator([]byte(jsonData), "SMA")
	require.NoError(t, err)
	require.Len(t, data.Values, 3)

	// Verify sorting (newest first)
	assert.Equal(t, 15, data.Values[0].Date.Day())
	assert.Equal(t, 185.5, data.Values[0].Value)

	assert.Equal(t, 12, data.Values[1].Date.Day())
	assert.Equal(t, 184.25, data.Values[1].Value)
}

func TestParseTechnicalIndicator_RSI(t *testing.T) {
	jsonData := `{
		"Meta Data": {
			"1: Symbol": "IBM",
			"2: Indicator": "Relative Strength Index (RSI)"
		},
		"Technical Analysis: RSI": {
			"2024-01-15": {"RSI": "65.4321"},
			"2024-01-12": {"RSI": "58.7654"}
		}
	}`

	data, err := parseTechnicalIndicator([]byte(jsonData), "RSI")
	require.NoError(t, err)
	require.Len(t, data.Values, 2)

	assert.Equal(t, 65.4321, data.Values[0].Value)
}

func TestTechnicalsParseMACD(t *testing.T) {
	jsonData := `{
		"Meta Data": {
			"1: Symbol": "IBM",
			"2: Indicator": "Moving Average Convergence/Divergence (MACD)"
		},
		"Technical Analysis: MACD": {
			"2024-01-15": {
				"MACD": "1.5432",
				"MACD_Signal": "1.2345",
				"MACD_Hist": "0.3087"
			},
			"2024-01-12": {
				"MACD": "1.1234",
				"MACD_Signal": "1.0987",
				"MACD_Hist": "0.0247"
			}
		}
	}`

	data, err := parseMACD([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, data.Values, 2)

	// Verify newest first
	assert.Equal(t, 15, data.Values[0].Date.Day())
	assert.Equal(t, 1.5432, data.Values[0].MACD)
	assert.Equal(t, 1.2345, data.Values[0].Signal)
	assert.Equal(t, 0.3087, data.Values[0].Histogram)
}

func TestParseSTOCH(t *testing.T) {
	jsonData := `{
		"Meta Data": {
			"1: Symbol": "IBM"
		},
		"Technical Analysis: STOCH": {
			"2024-01-15": {
				"SlowK": "75.4321",
				"SlowD": "72.1234"
			},
			"2024-01-12": {
				"SlowK": "68.5432",
				"SlowD": "65.8765"
			}
		}
	}`

	data, err := parseSTOCH([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, data.Values, 2)

	assert.Equal(t, 75.4321, data.Values[0].SlowK)
	assert.Equal(t, 72.1234, data.Values[0].SlowD)
}

func TestParseSTOCHF(t *testing.T) {
	jsonData := `{
		"Meta Data": {
			"1: Symbol": "IBM"
		},
		"Technical Analysis: STOCHF": {
			"2024-01-15": {
				"FastK": "82.5432",
				"FastD": "78.1234"
			}
		}
	}`

	data, err := parseSTOCHF([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, data.Values, 1)

	// STOCHF maps FastK/FastD to SlowK/SlowD in the struct
	assert.Equal(t, 82.5432, data.Values[0].SlowK)
	assert.Equal(t, 78.1234, data.Values[0].SlowD)
}

func TestParseSTOCHRSI(t *testing.T) {
	jsonData := `{
		"Meta Data": {
			"1: Symbol": "IBM"
		},
		"Technical Analysis: STOCHRSI": {
			"2024-01-15": {
				"FastK": "45.6789",
				"FastD": "42.3456"
			}
		}
	}`

	data, err := parseSTOCHRSI([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, data.Values, 1)

	assert.Equal(t, 45.6789, data.Values[0].SlowK)
	assert.Equal(t, 42.3456, data.Values[0].SlowD)
}

func TestParseAROON(t *testing.T) {
	jsonData := `{
		"Meta Data": {
			"1: Symbol": "IBM"
		},
		"Technical Analysis: AROON": {
			"2024-01-15": {
				"Aroon Up": "85.7143",
				"Aroon Down": "42.8571"
			},
			"2024-01-12": {
				"Aroon Up": "78.5714",
				"Aroon Down": "50.0000"
			}
		}
	}`

	data, err := parseAROON([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, data.Values, 2)

	assert.Equal(t, 85.7143, data.Values[0].AroonUp)
	assert.Equal(t, 42.8571, data.Values[0].AroonDown)
}

func TestTechnicalsParseBBANDS(t *testing.T) {
	jsonData := `{
		"Meta Data": {
			"1: Symbol": "IBM"
		},
		"Technical Analysis: BBANDS": {
			"2024-01-15": {
				"Real Upper Band": "190.5432",
				"Real Middle Band": "185.2345",
				"Real Lower Band": "179.9258"
			},
			"2024-01-12": {
				"Real Upper Band": "188.1234",
				"Real Middle Band": "183.4567",
				"Real Lower Band": "178.7890"
			}
		}
	}`

	data, err := parseBBANDS([]byte(jsonData))
	require.NoError(t, err)
	require.Len(t, data.Values, 2)

	// Verify newest first
	assert.Equal(t, 15, data.Values[0].Date.Day())
	assert.Equal(t, 190.5432, data.Values[0].UpperBand)
	assert.Equal(t, 185.2345, data.Values[0].MiddleBand)
	assert.Equal(t, 179.9258, data.Values[0].LowerBand)
}

func TestParseTechnicalIndicator_Empty(t *testing.T) {
	jsonData := `{
		"Meta Data": {}
	}`

	_, err := parseTechnicalIndicator([]byte(jsonData), "SMA")
	assert.Error(t, err)
}

func TestParseTechnicalIndicator_InvalidJSON(t *testing.T) {
	_, err := parseTechnicalIndicator([]byte("not json"), "SMA")
	assert.Error(t, err)
}

func TestTechnicalsParseMACD_Empty(t *testing.T) {
	jsonData := `{
		"Meta Data": {}
	}`

	_, err := parseMACD([]byte(jsonData))
	assert.Error(t, err)
}

func TestParseSTOCH_Empty(t *testing.T) {
	jsonData := `{
		"Meta Data": {}
	}`

	_, err := parseSTOCH([]byte(jsonData))
	assert.Error(t, err)
}

func TestParseAROON_Empty(t *testing.T) {
	jsonData := `{
		"Meta Data": {}
	}`

	_, err := parseAROON([]byte(jsonData))
	assert.Error(t, err)
}

func TestTechnicalsParseBBANDS_Empty(t *testing.T) {
	jsonData := `{
		"Meta Data": {}
	}`

	_, err := parseBBANDS([]byte(jsonData))
	assert.Error(t, err)
}

func TestTechnicalIndicatorDateSorting(t *testing.T) {
	jsonData := `{
		"Meta Data": {},
		"Technical Analysis: SMA": {
			"2024-01-01": {"SMA": "100"},
			"2024-01-15": {"SMA": "105"},
			"2024-01-10": {"SMA": "103"}
		}
	}`

	data, err := parseTechnicalIndicator([]byte(jsonData), "SMA")
	require.NoError(t, err)
	require.Len(t, data.Values, 3)

	// Should be sorted newest first: 15, 10, 1
	assert.Equal(t, 15, data.Values[0].Date.Day())
	assert.Equal(t, 10, data.Values[1].Date.Day())
	assert.Equal(t, 1, data.Values[2].Date.Day())
}

func TestParseTechnicalIndicator_WithDatetime(t *testing.T) {
	// Some indicators include time in the datetime
	jsonData := `{
		"Meta Data": {},
		"Technical Analysis: SMA": {
			"2024-01-15 14:30:00": {"SMA": "185.50"},
			"2024-01-15 14:25:00": {"SMA": "185.25"}
		}
	}`

	data, err := parseTechnicalIndicator([]byte(jsonData), "SMA")
	require.NoError(t, err)
	require.Len(t, data.Values, 2)

	// Check datetime parsing
	assert.Equal(t, 14, data.Values[0].Date.Hour())
	assert.Equal(t, 30, data.Values[0].Date.Minute())
}

// TestClientTechnicalMethods verifies all technical indicator methods exist.
func TestClientTechnicalMethods(t *testing.T) {
	client := NewClient("test-key", zerolog.Nop())

	// Moving Averages
	var _ func(string, string, int, string) (*IndicatorData, error) = client.GetSMA
	var _ func(string, string, int, string) (*IndicatorData, error) = client.GetEMA
	var _ func(string, string, int, string) (*IndicatorData, error) = client.GetWMA
	var _ func(string, string, int, string) (*IndicatorData, error) = client.GetDEMA
	var _ func(string, string, int, string) (*IndicatorData, error) = client.GetTEMA
	var _ func(string, string, int, string) (*IndicatorData, error) = client.GetTRIMA
	var _ func(string, string, int, string) (*IndicatorData, error) = client.GetKAMA
	var _ func(string, string, int, string) (*IndicatorData, error) = client.GetT3
	var _ func(string, string) (*IndicatorData, error) = client.GetVWAP

	// Momentum
	var _ func(string, string, int, string) (*IndicatorData, error) = client.GetRSI
	var _ func(string, string, string) (*MACDData, error) = client.GetMACD
	var _ func(string, string, string, map[string]string) (*MACDData, error) = client.GetMACDEXT
	var _ func(string, string) (*StochData, error) = client.GetSTOCH
	var _ func(string, string) (*StochData, error) = client.GetSTOCHF
	var _ func(string, string, int, string) (*StochData, error) = client.GetSTOCHRSI
	var _ func(string, string, int) (*IndicatorData, error) = client.GetWILLR
	var _ func(string, string, int, string) (*IndicatorData, error) = client.GetROC
	var _ func(string, string, int, string) (*IndicatorData, error) = client.GetROCR
	var _ func(string, string, int, string) (*IndicatorData, error) = client.GetMOM
	var _ func(string, string, int) (*IndicatorData, error) = client.GetMFI
	var _ func(string, string, int, string) (*IndicatorData, error) = client.GetCMO
	var _ func(string, string) (*IndicatorData, error) = client.GetULTOSC
	var _ func(string, string, int, string) (*IndicatorData, error) = client.GetTRIX
	var _ func(string, string, string, int, int) (*IndicatorData, error) = client.GetAPO
	var _ func(string, string, string, int, int) (*IndicatorData, error) = client.GetPPO
	var _ func(string, string) (*IndicatorData, error) = client.GetBOP

	// Trend
	var _ func(string, string, int) (*IndicatorData, error) = client.GetADX
	var _ func(string, string, int) (*IndicatorData, error) = client.GetADXR
	var _ func(string, string, int) (*IndicatorData, error) = client.GetDX
	var _ func(string, string, int) (*IndicatorData, error) = client.GetPLUS_DI
	var _ func(string, string, int) (*IndicatorData, error) = client.GetMINUS_DI
	var _ func(string, string, int) (*IndicatorData, error) = client.GetPLUS_DM
	var _ func(string, string, int) (*IndicatorData, error) = client.GetMINUS_DM
	var _ func(string, string, int) (*AroonData, error) = client.GetAROON
	var _ func(string, string, int) (*IndicatorData, error) = client.GetAROONOSC
	var _ func(string, string, float64, float64) (*IndicatorData, error) = client.GetSAR

	// Volatility
	var _ func(string, string, int, string, float64, float64) (*BollingerData, error) = client.GetBBANDS
	var _ func(string, string, int) (*IndicatorData, error) = client.GetATR
	var _ func(string, string, int) (*IndicatorData, error) = client.GetNATR
	var _ func(string, string) (*IndicatorData, error) = client.GetTRANGE
	var _ func(string, string, int, string) (*IndicatorData, error) = client.GetMIDPOINT
	var _ func(string, string, int) (*IndicatorData, error) = client.GetMIDPRICE

	// Volume
	var _ func(string, string) (*IndicatorData, error) = client.GetOBV
	var _ func(string, string) (*IndicatorData, error) = client.GetAD
	var _ func(string, string, int, int) (*IndicatorData, error) = client.GetADOSC

	// Cycle
	var _ func(string, string, string) (*IndicatorData, error) = client.GetHT_TRENDLINE
	var _ func(string, string, string) (*IndicatorData, error) = client.GetHT_SINE
	var _ func(string, string, string) (*IndicatorData, error) = client.GetHT_TRENDMODE
	var _ func(string, string, string) (*IndicatorData, error) = client.GetHT_DCPERIOD
	var _ func(string, string, string) (*IndicatorData, error) = client.GetHT_DCPHASE
	var _ func(string, string, string) (*IndicatorData, error) = client.GetHT_PHASOR

	// Generic
	var _ func(string, string, string, map[string]string) (*IndicatorData, error) = client.GetIndicator
}

// BenchmarkParseTechnicalIndicator benchmarks indicator parsing.
func BenchmarkParseTechnicalIndicator(b *testing.B) {
	jsonData := []byte(`{
		"Meta Data": {"1: Symbol": "IBM"},
		"Technical Analysis: SMA": {
			"2024-01-15": {"SMA": "185.50"},
			"2024-01-14": {"SMA": "185.25"},
			"2024-01-13": {"SMA": "185.00"}
		}
	}`)

	for i := 0; i < b.N; i++ {
		_, _ = parseTechnicalIndicator(jsonData, "SMA")
	}
}

// BenchmarkTechnicalsParseMACD benchmarks MACD parsing.
func BenchmarkTechnicalsParseMACD(b *testing.B) {
	jsonData := []byte(`{
		"Meta Data": {},
		"Technical Analysis: MACD": {
			"2024-01-15": {"MACD": "1.5", "MACD_Signal": "1.2", "MACD_Hist": "0.3"},
			"2024-01-14": {"MACD": "1.4", "MACD_Signal": "1.1", "MACD_Hist": "0.3"}
		}
	}`)

	for i := 0; i < b.N; i++ {
		_, _ = parseMACD(jsonData)
	}
}

// TestIndicatorDataStructure verifies the IndicatorData struct.
func TestIndicatorDataStructure(t *testing.T) {
	data := &IndicatorData{
		Symbol:   "IBM",
		Interval: "daily",
		Values: []IndicatorValue{
			{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Value: 185.5},
		},
	}

	assert.Equal(t, "IBM", data.Symbol)
	assert.Equal(t, "daily", data.Interval)
	assert.Len(t, data.Values, 1)
}

// TestMACDDataStructure verifies the MACDData struct.
func TestMACDDataStructure(t *testing.T) {
	data := &MACDData{
		Symbol:   "IBM",
		Interval: "daily",
		Values: []MACDValue{
			{
				Date:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				MACD:      1.5,
				Signal:    1.2,
				Histogram: 0.3,
			},
		},
	}

	assert.Equal(t, "IBM", data.Symbol)
	assert.Len(t, data.Values, 1)
	assert.Equal(t, 1.5, data.Values[0].MACD)
}

// TestBollingerDataStructure verifies the BollingerData struct.
func TestBollingerDataStructure(t *testing.T) {
	data := &BollingerData{
		Symbol:   "IBM",
		Interval: "daily",
		Values: []BollingerValue{
			{
				Date:       time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				UpperBand:  190.5,
				MiddleBand: 185.0,
				LowerBand:  179.5,
			},
		},
	}

	assert.Equal(t, "IBM", data.Symbol)
	assert.Len(t, data.Values, 1)
	assert.Equal(t, 190.5, data.Values[0].UpperBand)
}
