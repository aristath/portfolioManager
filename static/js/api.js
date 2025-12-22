/**
 * Arduino Trader - API Layer
 * Centralized API calls for the application
 */

const API = {
  // Base fetch with JSON handling
  async _fetch(url, options = {}) {
    const res = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });
    return res.json();
  },

  async _post(url, data) {
    return this._fetch(url, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  },

  async _put(url, data) {
    return this._fetch(url, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  async _delete(url) {
    return this._fetch(url, { method: 'DELETE' });
  },

  // Status
  fetchStatus: () => fetch('/api/status').then(r => r.json()),
  fetchTradernet: () => fetch('/api/status/tradernet').then(r => r.json()),
  syncPrices: () => API._post('/api/status/sync/prices'),

  // Allocation
  fetchAllocation: () => fetch('/api/trades/allocation').then(r => r.json()),
  fetchTargets: () => fetch('/api/allocation/targets').then(r => r.json()),
  saveGeoTargets: (targets) => API._put('/api/allocation/targets/geography', { targets }),
  saveIndustryTargets: (targets) => API._put('/api/allocation/targets/industry', { targets }),

  // Stocks
  fetchStocks: () => fetch('/api/stocks').then(r => r.json()),
  createStock: (data) => API._post('/api/stocks', data),
  updateStock: (symbol, data) => API._put(`/api/stocks/${symbol}`, data),
  deleteStock: (symbol) => API._delete(`/api/stocks/${symbol}`),
  refreshScore: (symbol) => API._post(`/api/stocks/${symbol}/refresh`),
  refreshAllScores: () => API._post('/api/stocks/refresh-all'),

  // Trades
  fetchTrades: () => fetch('/api/trades').then(r => r.json()),
  previewRebalance: () => API._post('/api/trades/rebalance/preview'),
  executeRebalance: () => API._post('/api/trades/rebalance/execute'),

  // Portfolio
  fetchPnl: () => fetch('/api/portfolio/pnl').then(r => r.json()),
  setDeposits: (amount) => API._put('/api/portfolio/deposits', { amount }),

  // Charts
  fetchPortfolioChart: (range = 'all') => {
    const params = new URLSearchParams({ range });
    return fetch(`/api/charts/portfolio?${params}`).then(r => r.json());
  },
  fetchStockChart: (symbol, range = '1Y', source = 'tradernet') => {
    const params = new URLSearchParams({ range, source });
    return fetch(`/api/charts/stocks/${symbol}?${params}`).then(r => r.json());
  },
};

// Make available globally
window.API = API;
