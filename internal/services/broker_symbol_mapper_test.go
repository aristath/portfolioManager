package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBrokerSymbolRepository is a mock implementation of domain.BrokerSymbolRepositoryInterface
type mockBrokerSymbolRepository struct {
	getBrokerSymbolFunc       func(isin, brokerName string) (string, error)
	setBrokerSymbolFunc       func(isin, brokerName, symbol string) error
	getAllBrokerSymbolsFunc   func(isin string) (map[string]string, error)
	getISINByBrokerSymbolFunc func(brokerName, brokerSymbol string) (string, error)
	deleteBrokerSymbolFunc    func(isin, brokerName string) error
}

func (m *mockBrokerSymbolRepository) GetBrokerSymbol(isin, brokerName string) (string, error) {
	if m.getBrokerSymbolFunc != nil {
		return m.getBrokerSymbolFunc(isin, brokerName)
	}
	return "", errors.New("not implemented")
}

func (m *mockBrokerSymbolRepository) SetBrokerSymbol(isin, brokerName, symbol string) error {
	if m.setBrokerSymbolFunc != nil {
		return m.setBrokerSymbolFunc(isin, brokerName, symbol)
	}
	return errors.New("not implemented")
}

func (m *mockBrokerSymbolRepository) GetAllBrokerSymbols(isin string) (map[string]string, error) {
	if m.getAllBrokerSymbolsFunc != nil {
		return m.getAllBrokerSymbolsFunc(isin)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBrokerSymbolRepository) GetISINByBrokerSymbol(brokerName, brokerSymbol string) (string, error) {
	if m.getISINByBrokerSymbolFunc != nil {
		return m.getISINByBrokerSymbolFunc(brokerName, brokerSymbol)
	}
	return "", errors.New("not implemented")
}

func (m *mockBrokerSymbolRepository) DeleteBrokerSymbol(isin, brokerName string) error {
	if m.deleteBrokerSymbolFunc != nil {
		return m.deleteBrokerSymbolFunc(isin, brokerName)
	}
	return errors.New("not implemented")
}

func TestGetBrokerSymbol_Success(t *testing.T) {
	// Arrange
	mockRepo := &mockBrokerSymbolRepository{
		getBrokerSymbolFunc: func(isin, brokerName string) (string, error) {
			if isin == "US0378331005" && brokerName == "tradernet" {
				return "AAPL.US", nil
			}
			return "", errors.New("not found")
		},
	}
	mapper := NewBrokerSymbolMapper(mockRepo)

	// Execute
	symbol, err := mapper.GetBrokerSymbol("US0378331005", "tradernet")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "AAPL.US", symbol)
}

func TestGetBrokerSymbol_NotFound(t *testing.T) {
	// Arrange
	mockRepo := &mockBrokerSymbolRepository{
		getBrokerSymbolFunc: func(isin, brokerName string) (string, error) {
			return "", errors.New("broker symbol mapping not found for ISIN US0378331005 and broker tradernet")
		},
	}
	mapper := NewBrokerSymbolMapper(mockRepo)

	// Execute
	symbol, err := mapper.GetBrokerSymbol("US0378331005", "tradernet")

	// Assert
	require.Error(t, err)
	assert.Empty(t, symbol)
	assert.Contains(t, err.Error(), "not found")
}

func TestSetBrokerSymbol(t *testing.T) {
	// Arrange
	var capturedIsin, capturedBrokerName, capturedSymbol string
	mockRepo := &mockBrokerSymbolRepository{
		setBrokerSymbolFunc: func(isin, brokerName, symbol string) error {
			capturedIsin = isin
			capturedBrokerName = brokerName
			capturedSymbol = symbol
			return nil
		},
	}
	mapper := NewBrokerSymbolMapper(mockRepo)

	// Execute
	err := mapper.SetBrokerSymbol("US0378331005", "tradernet", "AAPL.US")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "US0378331005", capturedIsin)
	assert.Equal(t, "tradernet", capturedBrokerName)
	assert.Equal(t, "AAPL.US", capturedSymbol)
}

func TestSetBrokerSymbol_Error(t *testing.T) {
	// Arrange
	mockRepo := &mockBrokerSymbolRepository{
		setBrokerSymbolFunc: func(isin, brokerName, symbol string) error {
			return errors.New("database error")
		},
	}
	mapper := NewBrokerSymbolMapper(mockRepo)

	// Execute
	err := mapper.SetBrokerSymbol("US0378331005", "tradernet", "AAPL.US")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}
