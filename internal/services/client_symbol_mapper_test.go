package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClientSymbolRepository is a mock implementation of domain.ClientSymbolRepositoryInterface
type mockClientSymbolRepository struct {
	getClientSymbolFunc       func(isin, clientName string) (string, error)
	setClientSymbolFunc       func(isin, clientName, symbol string) error
	getAllClientSymbolsFunc   func(isin string) (map[string]string, error)
	getISINByClientSymbolFunc func(clientName, clientSymbol string) (string, error)
	deleteClientSymbolFunc    func(isin, clientName string) error
}

func (m *mockClientSymbolRepository) GetClientSymbol(isin, clientName string) (string, error) {
	if m.getClientSymbolFunc != nil {
		return m.getClientSymbolFunc(isin, clientName)
	}
	return "", errors.New("not implemented")
}

func (m *mockClientSymbolRepository) SetClientSymbol(isin, clientName, symbol string) error {
	if m.setClientSymbolFunc != nil {
		return m.setClientSymbolFunc(isin, clientName, symbol)
	}
	return errors.New("not implemented")
}

func (m *mockClientSymbolRepository) GetAllClientSymbols(isin string) (map[string]string, error) {
	if m.getAllClientSymbolsFunc != nil {
		return m.getAllClientSymbolsFunc(isin)
	}
	return nil, errors.New("not implemented")
}

func (m *mockClientSymbolRepository) GetISINByClientSymbol(clientName, clientSymbol string) (string, error) {
	if m.getISINByClientSymbolFunc != nil {
		return m.getISINByClientSymbolFunc(clientName, clientSymbol)
	}
	return "", errors.New("not implemented")
}

func (m *mockClientSymbolRepository) DeleteClientSymbol(isin, clientName string) error {
	if m.deleteClientSymbolFunc != nil {
		return m.deleteClientSymbolFunc(isin, clientName)
	}
	return errors.New("not implemented")
}

func TestGetClientSymbol_Success(t *testing.T) {
	// Arrange
	mockRepo := &mockClientSymbolRepository{
		getClientSymbolFunc: func(isin, clientName string) (string, error) {
			if isin == "US0378331005" && clientName == "tradernet" {
				return "AAPL.US", nil
			}
			return "", errors.New("not found")
		},
	}
	mapper := NewClientSymbolMapper(mockRepo)

	// Execute
	symbol, err := mapper.GetClientSymbol("US0378331005", "tradernet")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "AAPL.US", symbol)
}

func TestGetClientSymbol_NotFound(t *testing.T) {
	// Arrange
	mockRepo := &mockClientSymbolRepository{
		getClientSymbolFunc: func(isin, clientName string) (string, error) {
			return "", errors.New("client symbol mapping not found for ISIN US0378331005 and client tradernet")
		},
	}
	mapper := NewClientSymbolMapper(mockRepo)

	// Execute
	symbol, err := mapper.GetClientSymbol("US0378331005", "tradernet")

	// Assert
	require.Error(t, err)
	assert.Empty(t, symbol)
	assert.Contains(t, err.Error(), "not found")
}

func TestSetClientSymbol(t *testing.T) {
	// Arrange
	var capturedIsin, capturedClientName, capturedSymbol string
	mockRepo := &mockClientSymbolRepository{
		setClientSymbolFunc: func(isin, clientName, symbol string) error {
			capturedIsin = isin
			capturedClientName = clientName
			capturedSymbol = symbol
			return nil
		},
	}
	mapper := NewClientSymbolMapper(mockRepo)

	// Execute
	err := mapper.SetClientSymbol("US0378331005", "tradernet", "AAPL.US")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "US0378331005", capturedIsin)
	assert.Equal(t, "tradernet", capturedClientName)
	assert.Equal(t, "AAPL.US", capturedSymbol)
}

func TestSetClientSymbol_Error(t *testing.T) {
	// Arrange
	mockRepo := &mockClientSymbolRepository{
		setClientSymbolFunc: func(isin, clientName, symbol string) error {
			return errors.New("database error")
		},
	}
	mapper := NewClientSymbolMapper(mockRepo)

	// Execute
	err := mapper.SetClientSymbol("US0378331005", "tradernet", "AAPL.US")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}
