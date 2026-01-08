package testing

import (
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/aristath/sentinel/internal/clients/tradernet"
	"github.com/aristath/sentinel/internal/domain"
	"github.com/aristath/sentinel/internal/modules/allocation"
	"github.com/aristath/sentinel/internal/modules/portfolio"
	"github.com/aristath/sentinel/internal/modules/trading"
	"github.com/aristath/sentinel/internal/modules/universe"
)

// MockPositionRepository is a mock implementation of PositionRepositoryInterface for testing
type MockPositionRepository struct {
	mu        sync.RWMutex
	positions []portfolio.Position
	err       error
}

// NewMockPositionRepository creates a new mock position repository
func NewMockPositionRepository() *MockPositionRepository {
	return &MockPositionRepository{
		positions: make([]portfolio.Position, 0),
	}
}

// SetPositions sets the positions to return
func (m *MockPositionRepository) SetPositions(positions []portfolio.Position) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.positions = positions
}

// SetError sets the error to return
func (m *MockPositionRepository) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

// GetAll returns all positions
func (m *MockPositionRepository) GetAll() ([]portfolio.Position, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	return m.positions, nil
}

// GetWithSecurityInfo returns all positions with security information
func (m *MockPositionRepository) GetWithSecurityInfo() ([]portfolio.PositionWithSecurity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	// Convert to PositionWithSecurity (simplified for mock)
	result := make([]portfolio.PositionWithSecurity, 0, len(m.positions))
	for _, pos := range m.positions {
		result = append(result, portfolio.PositionWithSecurity{
			Symbol:         pos.Symbol,
			Quantity:       pos.Quantity,
			AvgPrice:       pos.AvgPrice,
			CurrentPrice:   pos.CurrentPrice,
			Currency:       pos.Currency,
			CurrencyRate:   pos.CurrencyRate,
			MarketValueEUR: pos.MarketValueEUR,
		})
	}
	return result, nil
}

// GetBySymbol returns a position by symbol
func (m *MockPositionRepository) GetBySymbol(symbol string) (*portfolio.Position, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	for i := range m.positions {
		if m.positions[i].Symbol == symbol {
			return &m.positions[i], nil
		}
	}
	return nil, nil
}

// GetByISIN returns a position by ISIN
func (m *MockPositionRepository) GetByISIN(isin string) (*portfolio.Position, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	for i := range m.positions {
		if m.positions[i].ISIN == isin {
			return &m.positions[i], nil
		}
	}
	return nil, nil
}

// GetByIdentifier returns a position by symbol or ISIN
func (m *MockPositionRepository) GetByIdentifier(identifier string) (*portfolio.Position, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	for i := range m.positions {
		if m.positions[i].Symbol == identifier || m.positions[i].ISIN == identifier {
			return &m.positions[i], nil
		}
	}
	return nil, nil
}

// GetCount returns the total number of positions
func (m *MockPositionRepository) GetCount() (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return 0, m.err
	}
	return len(m.positions), nil
}

// GetTotalValue returns the total value of all positions
func (m *MockPositionRepository) GetTotalValue() (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return 0, m.err
	}
	var total float64
	for _, pos := range m.positions {
		total += pos.MarketValueEUR
	}
	return total, nil
}

// Upsert inserts or updates a position
func (m *MockPositionRepository) Upsert(position portfolio.Position) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	// Update if exists, otherwise append
	for i := range m.positions {
		if m.positions[i].ISIN == position.ISIN {
			m.positions[i] = position
			return nil
		}
	}
	m.positions = append(m.positions, position)
	return nil
}

// Delete deletes a position by ISIN
func (m *MockPositionRepository) Delete(isin string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	for i, pos := range m.positions {
		if pos.ISIN == isin {
			m.positions = append(m.positions[:i], m.positions[i+1:]...)
			return nil
		}
	}
	return nil
}

// DeleteAll deletes all positions
func (m *MockPositionRepository) DeleteAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.positions = make([]portfolio.Position, 0)
	return nil
}

// UpdatePrice updates the price for a position by ISIN
func (m *MockPositionRepository) UpdatePrice(isin string, price float64, currencyRate float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	for i := range m.positions {
		if m.positions[i].ISIN == isin {
			m.positions[i].CurrentPrice = price
			m.positions[i].CurrencyRate = currencyRate
			m.positions[i].MarketValueEUR = m.positions[i].Quantity * price * currencyRate
			return nil
		}
	}
	return nil
}

// UpdateLastSoldAt updates the last sold timestamp for a position by ISIN
func (m *MockPositionRepository) UpdateLastSoldAt(isin string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	// Mock implementation - would update last_sold_at field if it existed
	return nil
}

// MockTradeRepository is a mock implementation of TradeRepositoryInterface for testing
type MockTradeRepository struct {
	mu     sync.RWMutex
	trades []trading.Trade
	err    error
}

// NewMockTradeRepository creates a new mock trade repository
func NewMockTradeRepository() *MockTradeRepository {
	return &MockTradeRepository{
		trades: make([]trading.Trade, 0),
	}
}

// SetTrades sets the trades to return
func (m *MockTradeRepository) SetTrades(trades []trading.Trade) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.trades = trades
}

// SetError sets the error to return
func (m *MockTradeRepository) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

// Create creates a new trade
func (m *MockTradeRepository) Create(trade trading.Trade) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.trades = append(m.trades, trade)
	return nil
}

// GetByOrderID retrieves a trade by broker order ID
func (m *MockTradeRepository) GetByOrderID(orderID string) (*trading.Trade, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	for i := range m.trades {
		if m.trades[i].OrderID == orderID {
			return &m.trades[i], nil
		}
	}
	return nil, nil
}

// Exists checks if a trade with the given order_id already exists
func (m *MockTradeRepository) Exists(orderID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return false, m.err
	}
	for _, trade := range m.trades {
		if trade.OrderID == orderID {
			return true, nil
		}
	}
	return false, nil
}

// GetHistory retrieves trade history, most recent first
func (m *MockTradeRepository) GetHistory(limit int) ([]trading.Trade, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	if limit <= 0 || limit > len(m.trades) {
		limit = len(m.trades)
	}
	// Return last 'limit' trades (simplified - would sort by date in real implementation)
	if limit > len(m.trades) {
		limit = len(m.trades)
	}
	start := len(m.trades) - limit
	if start < 0 {
		start = 0
	}
	return m.trades[start:], nil
}

// GetAllInRange retrieves all trades in the specified date range
func (m *MockTradeRepository) GetAllInRange(startDate, endDate string) ([]trading.Trade, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	// Simplified implementation - would parse dates and filter in real implementation
	return m.trades, nil
}

// GetBySymbol retrieves trades for a symbol
func (m *MockTradeRepository) GetBySymbol(symbol string, limit int) ([]trading.Trade, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	var result []trading.Trade
	for _, trade := range m.trades {
		if trade.Symbol == symbol {
			result = append(result, trade)
		}
	}
	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}
	return result, nil
}

// GetByISIN retrieves trades for an ISIN
func (m *MockTradeRepository) GetByISIN(isin string, limit int) ([]trading.Trade, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	var result []trading.Trade
	for _, trade := range m.trades {
		if trade.ISIN == isin {
			result = append(result, trade)
		}
	}
	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}
	return result, nil
}

// GetByIdentifier retrieves trades by symbol or ISIN
func (m *MockTradeRepository) GetByIdentifier(identifier string, limit int) ([]trading.Trade, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	var result []trading.Trade
	for _, trade := range m.trades {
		if trade.Symbol == identifier || trade.ISIN == identifier {
			result = append(result, trade)
		}
	}
	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}
	return result, nil
}

// GetRecentlyBoughtSymbols returns symbols that were bought recently (within days)
func (m *MockTradeRepository) GetRecentlyBoughtSymbols(days int) (map[string]bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[string]bool)
	for _, trade := range m.trades {
		if trade.Side == trading.TradeSideBuy {
			result[trade.Symbol] = true
		}
	}
	return result, nil
}

// GetRecentlySoldSymbols returns symbols that were sold recently (within days)
func (m *MockTradeRepository) GetRecentlySoldSymbols(days int) (map[string]bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[string]bool)
	for _, trade := range m.trades {
		if trade.Side == trading.TradeSideSell {
			result[trade.Symbol] = true
		}
	}
	return result, nil
}

// HasRecentSellOrder checks if there was a recent sell order for a symbol within specified hours
func (m *MockTradeRepository) HasRecentSellOrder(symbol string, hours float64) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return false, m.err
	}
	for _, trade := range m.trades {
		if trade.Symbol == symbol && trade.Side == trading.TradeSideSell {
			return true, nil
		}
	}
	return false, nil
}

// GetFirstBuyDate retrieves the first buy date for a symbol
func (m *MockTradeRepository) GetFirstBuyDate(symbol string) (*string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	for _, trade := range m.trades {
		if trade.Symbol == symbol && trade.Side == trading.TradeSideBuy {
			dateStr := trade.ExecutedAt.Format("2006-01-02")
			return &dateStr, nil
		}
	}
	return nil, nil
}

// GetLastBuyDate retrieves the last buy date for a symbol
func (m *MockTradeRepository) GetLastBuyDate(symbol string) (*string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	var lastDate *string
	for _, trade := range m.trades {
		if trade.Symbol == symbol && trade.Side == trading.TradeSideBuy {
			dateStr := trade.ExecutedAt.Format("2006-01-02")
			lastDate = &dateStr
		}
	}
	return lastDate, nil
}

// GetLastSellDate retrieves the last sell date for a symbol
func (m *MockTradeRepository) GetLastSellDate(symbol string) (*string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	var lastDate *string
	for _, trade := range m.trades {
		if trade.Symbol == symbol && trade.Side == trading.TradeSideSell {
			dateStr := trade.ExecutedAt.Format("2006-01-02")
			lastDate = &dateStr
		}
	}
	return lastDate, nil
}

// GetLastTransactionDate retrieves the last transaction date for a symbol
func (m *MockTradeRepository) GetLastTransactionDate(symbol string) (*string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	var lastDate *string
	for _, trade := range m.trades {
		if trade.Symbol == symbol {
			dateStr := trade.ExecutedAt.Format("2006-01-02")
			lastDate = &dateStr
		}
	}
	return lastDate, nil
}

// GetTradeDates retrieves first buy, last buy, and last sell dates for symbols
func (m *MockTradeRepository) GetTradeDates() (map[string]map[string]*string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[string]map[string]*string)
	for _, trade := range m.trades {
		if result[trade.Symbol] == nil {
			result[trade.Symbol] = make(map[string]*string)
		}
		dateStr := trade.ExecutedAt.Format("2006-01-02")
		if trade.Side == trading.TradeSideBuy {
			if result[trade.Symbol]["first_buy"] == nil {
				firstBuy := dateStr
				result[trade.Symbol]["first_buy"] = &firstBuy
			}
			lastBuy := dateStr
			result[trade.Symbol]["last_buy"] = &lastBuy
		} else {
			lastSell := dateStr
			result[trade.Symbol]["last_sell"] = &lastSell
		}
	}
	return result, nil
}

// GetRecentTrades retrieves recent trades for a symbol within specified days
func (m *MockTradeRepository) GetRecentTrades(symbol string, days int) ([]trading.Trade, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	var result []trading.Trade
	for _, trade := range m.trades {
		if trade.Symbol == symbol {
			result = append(result, trade)
		}
	}
	return result, nil
}

// GetLastTradeTimestamp retrieves the timestamp of the most recent trade
func (m *MockTradeRepository) GetLastTradeTimestamp() (*time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	if len(m.trades) == 0 {
		return nil, nil
	}
	// Return timestamp of last trade (simplified)
	lastTrade := m.trades[len(m.trades)-1]
	return &lastTrade.ExecutedAt, nil
}

// GetTradeCountToday returns the number of trades executed today
func (m *MockTradeRepository) GetTradeCountToday() (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return 0, m.err
	}
	// Simplified - would filter by today's date in real implementation
	return len(m.trades), nil
}

// GetTradeCountThisWeek returns the number of trades executed this week
func (m *MockTradeRepository) GetTradeCountThisWeek() (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return 0, m.err
	}
	// Simplified - would filter by this week's date range in real implementation
	return len(m.trades), nil
}

// Verify interface implementation
var _ trading.TradeRepositoryInterface = (*MockTradeRepository)(nil)
var _ portfolio.PositionRepositoryInterface = (*MockPositionRepository)(nil)

// MockSecurityRepository is a mock implementation of SecurityRepositoryInterface for testing
type MockSecurityRepository struct {
	mu         sync.RWMutex
	securities map[string]*universe.Security // keyed by ISIN
	err        error
}

// NewMockSecurityRepository creates a new mock security repository
func NewMockSecurityRepository() *MockSecurityRepository {
	return &MockSecurityRepository{
		securities: make(map[string]*universe.Security),
	}
}

// SetSecurity sets a security to return
func (m *MockSecurityRepository) SetSecurity(security *universe.Security) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if security != nil {
		m.securities[security.ISIN] = security
		// Also index by symbol
		if security.Symbol != "" {
			// Create a copy with symbol as key (for GetBySymbol)
			secCopy := *security
			if secCopy.ISIN == "" {
				secCopy.ISIN = security.Symbol // Fallback if ISIN missing
			}
			m.securities[security.Symbol] = &secCopy
		}
	}
}

// SetSecurities sets multiple securities to return
func (m *MockSecurityRepository) SetSecurities(securities []*universe.Security) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, security := range securities {
		if security != nil {
			m.securities[security.ISIN] = security
			if security.Symbol != "" {
				secCopy := *security
				if secCopy.ISIN == "" {
					secCopy.ISIN = security.Symbol
				}
				m.securities[security.Symbol] = &secCopy
			}
		}
	}
}

// SetError sets the error to return
func (m *MockSecurityRepository) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

// GetBySymbol returns a security by symbol
func (m *MockSecurityRepository) GetBySymbol(symbol string) (*universe.Security, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	// First try direct lookup by symbol
	if sec, ok := m.securities[symbol]; ok {
		return sec, nil
	}
	// Fallback: search by symbol in all securities
	for _, sec := range m.securities {
		if sec.Symbol == symbol {
			return sec, nil
		}
	}
	return nil, nil
}

// GetByISIN returns a security by ISIN
func (m *MockSecurityRepository) GetByISIN(isin string) (*universe.Security, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	if sec, ok := m.securities[isin]; ok {
		return sec, nil
	}
	return nil, nil
}

// GetByIdentifier returns a security by symbol or ISIN
func (m *MockSecurityRepository) GetByIdentifier(identifier string) (*universe.Security, error) {
	// Try ISIN first
	if sec, err := m.GetByISIN(identifier); err != nil || sec != nil {
		return sec, err
	}
	// Fallback to symbol
	return m.GetBySymbol(identifier)
}

// GetAllActive returns all active securities
func (m *MockSecurityRepository) GetAllActive() ([]universe.Security, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	var result []universe.Security
	for _, sec := range m.securities {
		if sec.Active {
			result = append(result, *sec)
		}
	}
	return result, nil
}

// GetDistinctExchanges returns all distinct exchange names
func (m *MockSecurityRepository) GetDistinctExchanges() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	exchanges := make(map[string]bool)
	for _, sec := range m.securities {
		if sec.FullExchangeName != "" {
			exchanges[sec.FullExchangeName] = true
		}
	}
	result := make([]string, 0, len(exchanges))
	for exchange := range exchanges {
		result = append(result, exchange)
	}
	return result, nil
}

// GetAllActiveTradable returns all active and tradable securities
func (m *MockSecurityRepository) GetAllActiveTradable() ([]universe.Security, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	var result []universe.Security
	for _, sec := range m.securities {
		if sec.Active && (sec.AllowBuy || sec.AllowSell) {
			result = append(result, *sec)
		}
	}
	return result, nil
}

// GetAll returns all securities (active and inactive)
func (m *MockSecurityRepository) GetAll() ([]universe.Security, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	result := make([]universe.Security, 0, len(m.securities))
	for _, sec := range m.securities {
		result = append(result, *sec)
	}
	return result, nil
}

// Create creates a new security
func (m *MockSecurityRepository) Create(security universe.Security) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.securities[security.ISIN] = &security
	if security.Symbol != "" {
		secCopy := security
		if secCopy.ISIN == "" {
			secCopy.ISIN = security.Symbol
		}
		m.securities[security.Symbol] = &secCopy
	}
	return nil
}

// Update updates a security by ISIN
func (m *MockSecurityRepository) Update(isin string, updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	if sec, ok := m.securities[isin]; ok {
		// Apply updates (simplified implementation)
		for key, value := range updates {
			switch key {
			case "active":
				if val, ok := value.(bool); ok {
					sec.Active = val
				}
			case "name":
				if val, ok := value.(string); ok {
					sec.Name = val
				}
				// Add more fields as needed
			}
		}
	}
	return nil
}

// Delete deletes a security by ISIN
func (m *MockSecurityRepository) Delete(isin string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	if sec, ok := m.securities[isin]; ok {
		delete(m.securities, isin)
		if sec.Symbol != "" {
			delete(m.securities, sec.Symbol)
		}
	}
	return nil
}

// GetWithScores returns securities with their scores joined
func (m *MockSecurityRepository) GetWithScores(portfolioDB *sql.DB) ([]universe.SecurityWithScore, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	result := make([]universe.SecurityWithScore, 0, len(m.securities))
	for _, sec := range m.securities {
		result = append(result, universe.SecurityWithScore{
			Symbol:             sec.Symbol,
			ISIN:               sec.ISIN,
			Name:               sec.Name,
			Industry:           sec.Industry,
			Country:            sec.Country,
			FullExchangeName:   sec.FullExchangeName,
			Currency:           sec.Currency,
			ProductType:        sec.ProductType,
			YahooSymbol:        sec.YahooSymbol,
			PriorityMultiplier: sec.PriorityMultiplier,
			MaxPortfolioTarget: sec.MaxPortfolioTarget,
			MinPortfolioTarget: sec.MinPortfolioTarget,
			MinLot:             sec.MinLot,
			AllowSell:          sec.AllowSell,
			AllowBuy:           sec.AllowBuy,
			Active:             sec.Active,
			Tags:               sec.Tags,
		})
	}
	return result, nil
}

// SetTagsForSecurity replaces all tags for a security
func (m *MockSecurityRepository) SetTagsForSecurity(symbol string, tagIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	// Simplified implementation - would update tags in real implementation
	return nil
}

// GetTagsForSecurity returns all tag IDs for a security
func (m *MockSecurityRepository) GetTagsForSecurity(symbol string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	if sec, ok := m.securities[symbol]; ok {
		return sec.Tags, nil
	}
	return nil, nil
}

// GetTagsWithUpdateTimes returns all tags for a security with their last update times
func (m *MockSecurityRepository) GetTagsWithUpdateTimes(symbol string) (map[string]time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[string]time.Time)
	if sec, ok := m.securities[symbol]; ok {
		for _, tagID := range sec.Tags {
			result[tagID] = time.Now()
		}
	}
	return result, nil
}

// UpdateSpecificTags updates only the specified tags for a security
func (m *MockSecurityRepository) UpdateSpecificTags(symbol string, tagIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	if sec, ok := m.securities[symbol]; ok {
		sec.Tags = tagIDs
	}
	return nil
}

// GetByTags returns active securities matching any of the provided tags
func (m *MockSecurityRepository) GetByTags(tagIDs []string) ([]universe.Security, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	tagSet := make(map[string]bool)
	for _, tagID := range tagIDs {
		tagSet[tagID] = true
	}
	var result []universe.Security
	for _, sec := range m.securities {
		if sec.Active {
			for _, tag := range sec.Tags {
				if tagSet[tag] {
					result = append(result, *sec)
					break
				}
			}
		}
	}
	return result, nil
}

// GetPositionsByTags returns securities that are in the provided position symbols AND have the specified tags
func (m *MockSecurityRepository) GetPositionsByTags(positionSymbols []string, tagIDs []string) ([]universe.Security, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	positionSet := make(map[string]bool)
	for _, symbol := range positionSymbols {
		positionSet[symbol] = true
	}
	tagSet := make(map[string]bool)
	for _, tagID := range tagIDs {
		tagSet[tagID] = true
	}
	var result []universe.Security
	for _, sec := range m.securities {
		if positionSet[sec.Symbol] {
			for _, tag := range sec.Tags {
				if tagSet[tag] {
					result = append(result, *sec)
					break
				}
			}
		}
	}
	return result, nil
}

// Verify interface implementation
var _ universe.SecurityRepositoryInterface = (*MockSecurityRepository)(nil)

// MockCashManager is a mock implementation of CashManager for testing
type MockCashManager struct {
	mu       sync.RWMutex
	balances map[string]float64
	err      error
}

// NewMockCashManager creates a new mock cash manager
func NewMockCashManager() *MockCashManager {
	return &MockCashManager{
		balances: make(map[string]float64),
	}
}

// SetBalances sets the cash balances to return
func (m *MockCashManager) SetBalances(balances map[string]float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.balances = make(map[string]float64)
	for currency, balance := range balances {
		m.balances[currency] = balance
	}
}

// SetError sets the error to return
func (m *MockCashManager) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

// UpdateCashPosition updates or creates a cash balance for the given currency
func (m *MockCashManager) UpdateCashPosition(currency string, balance float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.balances[currency] = balance
	return nil
}

// GetAllCashBalances returns all cash balances as a map of currency -> balance
func (m *MockCashManager) GetAllCashBalances() (map[string]float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[string]float64)
	for currency, balance := range m.balances {
		result[currency] = balance
	}
	return result, nil
}

// GetCashBalance returns the cash balance for the given currency
func (m *MockCashManager) GetCashBalance(currency string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return 0, m.err
	}
	if balance, ok := m.balances[currency]; ok {
		return balance, nil
	}
	return 0, nil
}

// Verify interface implementation
var _ domain.CashManager = (*MockCashManager)(nil)

// MockTradernetClient is a mock implementation of TradernetClientInterface for testing
type MockTradernetClient struct {
	mu           sync.RWMutex
	connected    bool
	portfolio    []tradernet.Position
	cashBalances []tradernet.CashBalance
	trades       []tradernet.Trade
	err          error
}

// NewMockTradernetClient creates a new mock Tradernet client
func NewMockTradernetClient() *MockTradernetClient {
	return &MockTradernetClient{
		connected:    false,
		portfolio:    make([]tradernet.Position, 0),
		cashBalances: make([]tradernet.CashBalance, 0),
		trades:       make([]tradernet.Trade, 0),
	}
}

// SetConnected sets the connection status
func (m *MockTradernetClient) SetConnected(connected bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = connected
}

// SetPortfolio sets the portfolio to return
func (m *MockTradernetClient) SetPortfolio(portfolio []tradernet.Position) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.portfolio = portfolio
}

// SetCashBalances sets the cash balances to return
func (m *MockTradernetClient) SetCashBalances(cashBalances []tradernet.CashBalance) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cashBalances = cashBalances
}

// SetTrades sets the trades to return
func (m *MockTradernetClient) SetTrades(trades []tradernet.Trade) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.trades = trades
}

// SetError sets the error to return
func (m *MockTradernetClient) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

// GetPortfolio retrieves all positions from Tradernet
func (m *MockTradernetClient) GetPortfolio() ([]tradernet.Position, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	return m.portfolio, nil
}

// GetCashBalances retrieves all cash balances from Tradernet
func (m *MockTradernetClient) GetCashBalances() ([]tradernet.CashBalance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	return m.cashBalances, nil
}

// GetExecutedTrades retrieves executed trades from Tradernet
func (m *MockTradernetClient) GetExecutedTrades(limit int) ([]tradernet.Trade, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	result := m.trades
	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}
	return result, nil
}

// PlaceOrder places an order via Tradernet
func (m *MockTradernetClient) PlaceOrder(symbol, side string, quantity float64) (*tradernet.OrderResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	// Return a mock order result
	return &tradernet.OrderResult{
		OrderID:  "mock_order_" + symbol,
		Symbol:   symbol,
		Side:     side,
		Quantity: quantity,
		Price:    100.0, // Mock price
	}, nil
}

// IsConnected checks if the Tradernet client is connected
func (m *MockTradernetClient) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// Verify interface implementation
var _ domain.TradernetClientInterface = (*MockTradernetClient)(nil)

// MockCurrencyExchangeService is a mock implementation of CurrencyExchangeServiceInterface for testing
type MockCurrencyExchangeService struct {
	mu    sync.RWMutex
	rates map[string]map[string]float64 // from -> to -> rate
	err   error
}

// NewMockCurrencyExchangeService creates a new mock currency exchange service
func NewMockCurrencyExchangeService() *MockCurrencyExchangeService {
	return &MockCurrencyExchangeService{
		rates: make(map[string]map[string]float64),
	}
}

// SetRates sets the exchange rates to return
func (m *MockCurrencyExchangeService) SetRates(rates map[string]map[string]float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rates = make(map[string]map[string]float64)
	for from, toRates := range rates {
		m.rates[from] = make(map[string]float64)
		for to, rate := range toRates {
			m.rates[from][to] = rate
		}
	}
}

// SetError sets the error to return
func (m *MockCurrencyExchangeService) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

// GetRate returns the exchange rate from one currency to another
func (m *MockCurrencyExchangeService) GetRate(fromCurrency, toCurrency string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return 0, m.err
	}
	// Same currency
	if fromCurrency == toCurrency {
		return 1.0, nil
	}
	// Direct rate
	if fromRates, ok := m.rates[fromCurrency]; ok {
		if rate, ok := fromRates[toCurrency]; ok {
			return rate, nil
		}
	}
	// Reverse rate
	if toRates, ok := m.rates[toCurrency]; ok {
		if rate, ok := toRates[fromCurrency]; ok {
			return 1.0 / rate, nil
		}
	}
	return 0, errors.New("exchange rate not found")
}

// Verify interface implementation
var _ domain.CurrencyExchangeServiceInterface = (*MockCurrencyExchangeService)(nil)

// MockAllocationTargetProvider is a mock implementation of AllocationTargetProvider for testing
type MockAllocationTargetProvider struct {
	mu      sync.RWMutex
	targets []allocation.AllocationTarget
	err     error
}

// NewMockAllocationTargetProvider creates a new mock allocation target provider
func NewMockAllocationTargetProvider() *MockAllocationTargetProvider {
	return &MockAllocationTargetProvider{
		targets: make([]allocation.AllocationTarget, 0),
	}
}

// SetTargets sets the allocation targets to return
func (m *MockAllocationTargetProvider) SetTargets(targets []allocation.AllocationTarget) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.targets = targets
}

// SetError sets the error to return
func (m *MockAllocationTargetProvider) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

// GetAll returns all allocation targets as a map of name -> target percentage
func (m *MockAllocationTargetProvider) GetAll() (map[string]float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[string]float64)
	for _, target := range m.targets {
		// Use "type:name" as key (matching repository pattern)
		key := target.Type + ":" + target.Name
		result[key] = target.TargetPct
	}
	return result, nil
}

// Verify interface implementation
var _ domain.AllocationTargetProvider = (*MockAllocationTargetProvider)(nil)

// MockPortfolioSummaryProvider is a mock implementation of PortfolioSummaryProvider for testing
type MockPortfolioSummaryProvider struct {
	mu      sync.RWMutex
	summary domain.PortfolioSummary
	err     error
}

// NewMockPortfolioSummaryProvider creates a new mock portfolio summary provider
func NewMockPortfolioSummaryProvider() *MockPortfolioSummaryProvider {
	return &MockPortfolioSummaryProvider{
		summary: domain.PortfolioSummary{
			CountryAllocations:  make([]domain.PortfolioAllocation, 0),
			IndustryAllocations: make([]domain.PortfolioAllocation, 0),
			TotalValue:          0.0,
			CashBalance:         0.0,
		},
	}
}

// SetSummary sets the portfolio summary to return
func (m *MockPortfolioSummaryProvider) SetSummary(summary domain.PortfolioSummary) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.summary = summary
}

// SetError sets the error to return
func (m *MockPortfolioSummaryProvider) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

// GetPortfolioSummary returns the current portfolio summary
func (m *MockPortfolioSummaryProvider) GetPortfolioSummary() (domain.PortfolioSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return domain.PortfolioSummary{}, m.err
	}
	return m.summary, nil
}

// Verify interface implementation
var _ domain.PortfolioSummaryProvider = (*MockPortfolioSummaryProvider)(nil)
