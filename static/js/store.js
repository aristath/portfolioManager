/**
 * Arduino Trader - Alpine.js Store
 * Centralized state management for the application
 */

document.addEventListener('alpine:init', () => {
  Alpine.store('app', {
    // Data
    status: {},
    allocation: {
      geographic: [],
      industry: [],
      total_value: 0,
      cash_balance: 0
    },
    stocks: [],
    trades: [],
    tradernet: { connected: false },

    // UI State - Filters
    stockFilter: 'all',
    industryFilter: 'all',
    searchQuery: '',
    minScore: 0,
    sortBy: 'total_score',
    sortDesc: true,
    showRebalanceModal: false,
    showAddStockModal: false,
    showEditStockModal: false,
    editingStock: null,
    rebalancePreview: null,
    message: '',
    messageType: 'success',

    // Loading States
    loading: {
      rebalance: false,
      scores: false,
      sync: false,
      execute: false,
      geoSave: false,
      industrySave: false,
      stockSave: false
    },

    // Edit Mode States
    editingGeo: false,
    geoTargets: { EU: 50, ASIA: 30, US: 20 },
    editingIndustry: false,
    industryTargets: {},  // Dynamic - populated from API

    // Add Stock Form
    newStock: { symbol: '', name: '', geography: 'EU', industry: '' },
    addingStock: false,

    // Fetch All Data
    async fetchAll() {
      await Promise.all([
        this.fetchStatus(),
        this.fetchAllocation(),
        this.fetchStocks(),
        this.fetchTrades(),
        this.fetchTradernet()
      ]);
    },

    async fetchStatus() {
      try {
        const res = await fetch('/api/status');
        this.status = await res.json();
      } catch (e) {
        console.error('Failed to fetch status:', e);
      }
    },

    async fetchAllocation() {
      try {
        const res = await fetch('/api/trades/allocation');
        this.allocation = await res.json();
      } catch (e) {
        console.error('Failed to fetch allocation:', e);
      }
    },

    async fetchStocks() {
      try {
        const res = await fetch('/api/stocks');
        this.stocks = await res.json();
      } catch (e) {
        console.error('Failed to fetch stocks:', e);
      }
    },

    async fetchTrades() {
      try {
        const res = await fetch('/api/trades');
        this.trades = await res.json();
      } catch (e) {
        console.error('Failed to fetch trades:', e);
      }
    },

    async fetchTradernet() {
      try {
        const res = await fetch('/api/status/tradernet');
        this.tradernet = await res.json();
      } catch (e) {
        console.error('Failed to fetch tradernet status:', e);
      }
    },

    // Get unique industries for filter dropdown (handles comma-separated)
    get industries() {
      const set = new Set();
      this.stocks.forEach(s => {
        if (s.industry) {
          s.industry.split(',').forEach(ind => {
            const trimmed = ind.trim();
            if (trimmed) set.add(trimmed);
          });
        }
      });
      return Array.from(set).sort();
    },

    // Filtered & Sorted Stocks
    get filteredStocks() {
      let filtered = this.stocks;

      // Region filter
      if (this.stockFilter !== 'all') {
        filtered = filtered.filter(s => s.geography === this.stockFilter);
      }

      // Industry filter (handles comma-separated industries)
      if (this.industryFilter !== 'all') {
        filtered = filtered.filter(s => {
          if (!s.industry) return false;
          const industries = s.industry.split(',').map(i => i.trim());
          return industries.includes(this.industryFilter);
        });
      }

      // Search filter
      if (this.searchQuery) {
        const q = this.searchQuery.toLowerCase();
        filtered = filtered.filter(s =>
          s.symbol.toLowerCase().includes(q) ||
          s.name.toLowerCase().includes(q)
        );
      }

      // Score threshold
      if (this.minScore > 0) {
        filtered = filtered.filter(s => (s.total_score || 0) >= this.minScore);
      }

      // Sort
      return filtered.sort((a, b) => {
        let aVal = a[this.sortBy];
        let bVal = b[this.sortBy];

        // Handle nulls/undefined
        if (aVal == null) aVal = this.sortDesc ? -Infinity : Infinity;
        if (bVal == null) bVal = this.sortDesc ? -Infinity : Infinity;

        // String comparison for text fields
        if (typeof aVal === 'string' && typeof bVal === 'string') {
          return this.sortDesc
            ? bVal.localeCompare(aVal)
            : aVal.localeCompare(bVal);
        }

        // Numeric comparison
        return this.sortDesc ? bVal - aVal : aVal - bVal;
      });
    },

    sortStocks(field) {
      if (this.sortBy === field) {
        this.sortDesc = !this.sortDesc;
      } else {
        this.sortBy = field;
        this.sortDesc = true;
      }
    },

    // Rebalance Actions
    async previewRebalance() {
      this.loading.rebalance = true;
      this.rebalancePreview = null;
      try {
        const res = await fetch('/api/trades/rebalance/preview', { method: 'POST' });
        this.rebalancePreview = await res.json();
      } catch (e) {
        this.showMessage('Failed to preview rebalance', 'error');
      }
      this.loading.rebalance = false;
    },

    async executeRebalance() {
      this.loading.execute = true;
      try {
        const res = await fetch('/api/trades/rebalance/execute', { method: 'POST' });
        const data = await res.json();
        this.showMessage(`Executed ${data.successful_trades} trades`, 'success');
        this.showRebalanceModal = false;
        await this.fetchAll();
      } catch (e) {
        this.showMessage('Failed to execute rebalance', 'error');
      }
      this.loading.execute = false;
    },

    // Score Actions
    async refreshScores() {
      this.loading.scores = true;
      try {
        const res = await fetch('/api/stocks/refresh-all', { method: 'POST' });
        const data = await res.json();
        this.showMessage(data.message, 'success');
        await this.fetchStocks();
      } catch (e) {
        this.showMessage('Failed to refresh scores', 'error');
      }
      this.loading.scores = false;
    },

    async refreshSingleScore(symbol) {
      try {
        await fetch(`/api/stocks/${symbol}/refresh`, { method: 'POST' });
        await this.fetchStocks();
      } catch (e) {
        this.showMessage('Failed to refresh score', 'error');
      }
    },

    // Price Sync
    async syncPrices() {
      this.loading.sync = true;
      try {
        const res = await fetch('/api/status/sync/prices', { method: 'POST' });
        const data = await res.json();
        this.showMessage(data.message, 'success');
      } catch (e) {
        this.showMessage('Failed to sync prices', 'error');
      }
      this.loading.sync = false;
    },

    // Geographic Allocation Editing
    get geoTotal() {
      return Math.round(this.geoTargets.EU + this.geoTargets.ASIA + this.geoTargets.US);
    },

    startEditGeo() {
      if (this.allocation.geographic) {
        this.allocation.geographic.forEach(g => {
          if (g.name === 'EU') this.geoTargets.EU = Math.round(g.target_pct * 100);
          if (g.name === 'ASIA') this.geoTargets.ASIA = Math.round(g.target_pct * 100);
          if (g.name === 'US') this.geoTargets.US = Math.round(g.target_pct * 100);
        });
      }
      this.editingGeo = true;
    },

    cancelEditGeo() {
      this.editingGeo = false;
    },

    adjustGeoSlider(changed, newValue) {
      const others = ['EU', 'ASIA', 'US'].filter(r => r !== changed);
      const remaining = 100 - newValue;
      const otherTotal = this.geoTargets[others[0]] + this.geoTargets[others[1]];

      if (otherTotal === 0) {
        this.geoTargets[others[0]] = Math.round(remaining / 2);
        this.geoTargets[others[1]] = remaining - Math.round(remaining / 2);
      } else {
        const ratio0 = this.geoTargets[others[0]] / otherTotal;
        this.geoTargets[others[0]] = Math.round(remaining * ratio0);
        this.geoTargets[others[1]] = remaining - this.geoTargets[others[0]];
      }
      this.geoTargets[changed] = newValue;
    },

    async saveGeoTargets() {
      if (this.geoTotal !== 100) {
        this.showMessage('Targets must sum to 100%', 'error');
        return;
      }
      this.loading.geoSave = true;
      try {
        const res = await fetch('/api/allocation/targets/geography', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            EU: this.geoTargets.EU / 100,
            ASIA: this.geoTargets.ASIA / 100,
            US: this.geoTargets.US / 100
          })
        });
        if (res.ok) {
          this.showMessage('Allocation targets updated', 'success');
          this.editingGeo = false;
          await this.fetchAllocation();
        } else {
          this.showMessage('Failed to save targets', 'error');
        }
      } catch (e) {
        this.showMessage('Failed to save targets', 'error');
      }
      this.loading.geoSave = false;
    },

    // Industry Allocation Editing (dynamic industries)
    get industryTotal() {
      return Math.round(
        Object.values(this.industryTargets).reduce((sum, val) => sum + val, 0)
      );
    },

    startEditIndustry() {
      // Load current targets from allocation data
      this.industryTargets = {};
      if (this.allocation.industry) {
        this.allocation.industry.forEach(ind => {
          this.industryTargets[ind.name] = Math.round(ind.target_pct * 100);
        });
      }
      this.editingIndustry = true;
    },

    cancelEditIndustry() {
      this.editingIndustry = false;
    },

    adjustIndustrySlider(changed, newValue) {
      const industries = Object.keys(this.industryTargets);
      const others = industries.filter(i => i !== changed);
      const remaining = 100 - newValue;
      const otherTotal = others.reduce((sum, i) => sum + this.industryTargets[i], 0);

      if (otherTotal === 0) {
        const each = Math.floor(remaining / others.length);
        others.forEach((i, idx) => {
          this.industryTargets[i] = idx === others.length - 1
            ? remaining - each * (others.length - 1)
            : each;
        });
      } else {
        let distributed = 0;
        others.forEach((i, idx) => {
          if (idx === others.length - 1) {
            this.industryTargets[i] = remaining - distributed;
          } else {
            const ratio = this.industryTargets[i] / otherTotal;
            const val = Math.round(remaining * ratio);
            this.industryTargets[i] = val;
            distributed += val;
          }
        });
      }
      this.industryTargets[changed] = newValue;
    },

    async saveIndustryTargets() {
      if (this.industryTotal !== 100) {
        this.showMessage('Targets must sum to 100%', 'error');
        return;
      }
      this.loading.industrySave = true;
      try {
        // Convert percentages to decimals for API
        const targets = {};
        for (const [name, pct] of Object.entries(this.industryTargets)) {
          targets[name] = pct / 100;
        }
        const res = await fetch('/api/allocation/targets/industry', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ targets })
        });
        if (res.ok) {
          this.showMessage('Industry targets updated', 'success');
          this.editingIndustry = false;
          await this.fetchAllocation();
        } else {
          this.showMessage('Failed to save targets', 'error');
        }
      } catch (e) {
        this.showMessage('Failed to save targets', 'error');
      }
      this.loading.industrySave = false;
    },

    // Stock Management
    resetNewStock() {
      this.newStock = { symbol: '', name: '', geography: 'EU', industry: '' };
    },

    async addStock() {
      if (!this.newStock.symbol || !this.newStock.name) {
        this.showMessage('Symbol and name are required', 'error');
        return;
      }
      this.addingStock = true;
      try {
        const payload = {
          symbol: this.newStock.symbol.toUpperCase(),
          name: this.newStock.name,
          geography: this.newStock.geography
        };
        if (this.newStock.industry) {
          payload.industry = this.newStock.industry;
        }
        const res = await fetch('/api/stocks', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        if (res.ok) {
          this.showMessage('Stock added successfully', 'success');
          this.showAddStockModal = false;
          this.resetNewStock();
          await this.fetchStocks();
        } else {
          const data = await res.json();
          this.showMessage(data.detail || 'Failed to add stock', 'error');
        }
      } catch (e) {
        this.showMessage('Failed to add stock', 'error');
      }
      this.addingStock = false;
    },

    async removeStock(symbol) {
      if (!confirm(`Remove ${symbol} from the universe?`)) return;
      try {
        const res = await fetch(`/api/stocks/${symbol}`, { method: 'DELETE' });
        if (res.ok) {
          this.showMessage(`${symbol} removed`, 'success');
          await this.fetchStocks();
        } else {
          this.showMessage('Failed to remove stock', 'error');
        }
      } catch (e) {
        this.showMessage('Failed to remove stock', 'error');
      }
    },

    // Edit Stock
    openEditStock(stock) {
      this.editingStock = {
        symbol: stock.symbol,
        yahoo_symbol: stock.yahoo_symbol || '',
        name: stock.name,
        geography: stock.geography,
        industry: stock.industry || ''
      };
      this.showEditStockModal = true;
    },

    closeEditStock() {
      this.showEditStockModal = false;
      this.editingStock = null;
    },

    async saveStock() {
      if (!this.editingStock) return;

      this.loading.stockSave = true;
      try {
        const payload = {
          name: this.editingStock.name,
          yahoo_symbol: this.editingStock.yahoo_symbol || null,
          geography: this.editingStock.geography,
          industry: this.editingStock.industry || null
        };

        const res = await fetch(`/api/stocks/${this.editingStock.symbol}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });

        if (res.ok) {
          this.showMessage('Stock updated successfully', 'success');
          this.closeEditStock();
          await this.fetchStocks();
          await this.fetchAllocation();
        } else {
          const data = await res.json();
          this.showMessage(data.detail || 'Failed to update stock', 'error');
        }
      } catch (e) {
        this.showMessage('Failed to update stock', 'error');
      }
      this.loading.stockSave = false;
    },

    // Utilities
    showMessage(msg, type) {
      this.message = msg;
      this.messageType = type;
      setTimeout(() => { this.message = ''; }, 5000);
    }
  });
});
