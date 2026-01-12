import { create } from 'zustand';
import { api } from '../api/client';

// Data source types with human-readable names
export const DATA_SOURCE_TYPES = {
  fundamentals: 'Financial Fundamentals',
  current_prices: 'Current Prices',
  historical: 'Historical Data',
  technicals: 'Technical Indicators',
  exchange_rates: 'Exchange Rates',
  isin_lookup: 'ISIN Lookup',
  company_metadata: 'Company Metadata',
};

// Available data sources with descriptions
export const AVAILABLE_SOURCES = {
  alphavantage: { name: 'Alpha Vantage', description: 'Financial data provider with fundamentals, technicals, and more' },
  yahoo: { name: 'Yahoo Finance', description: 'Unofficial Yahoo Finance API for quotes and historical data' },
  tradernet: { name: 'Tradernet', description: 'Primary broker with real-time quotes and order execution' },
  exchangerate: { name: 'ExchangeRate API', description: 'Currency exchange rates provider' },
  openfigi: { name: 'OpenFIGI', description: 'ISIN to ticker symbol mapping service' },
};

export const useSettingsStore = create((set, get) => ({
  // Settings
  settings: {
    limit_order_buffer_percent: 0.05,
  },
  tradingMode: 'research',
  
  // Data sources state
  dataSources: {
    sources: [],
    availableSources: [],
    loading: false,
    error: null,
  },

  // Actions
  fetchSettings: async () => {
    try {
      const settings = await api.fetchSettings();
      set({
        settings,
        tradingMode: settings.trading_mode || 'research',
      });
    } catch (e) {
      console.error('Failed to fetch settings:', e);
    }
  },

  updateSetting: async (key, value) => {
    const { settings, tradingMode } = get();
    const oldValue = settings[key];
    const oldTradingMode = tradingMode;

    // Optimistic update - apply immediately for better UX
    set({ settings: { ...settings, [key]: value } });
    if (key === 'trading_mode') {
      set({ tradingMode: value });
    }

    try {
      await api.updateSetting(key, value);
    } catch (e) {
      console.error(`Failed to update setting ${key}:`, e);
      // Rollback on error
      set({ settings: { ...settings, [key]: oldValue } });
      if (key === 'trading_mode') {
        set({ tradingMode: oldTradingMode });
      }
      throw e;
    }
  },

  toggleTradingMode: async () => {
    const { tradingMode } = get();
    const oldMode = tradingMode;
    const newMode = tradingMode === 'live' ? 'research' : 'live';

    // Optimistic update - toggle immediately
    set({ tradingMode: newMode });

    try {
      await api.toggleTradingMode();
    } catch (e) {
      console.error('Failed to toggle trading mode:', e);
      // Rollback on error
      set({ tradingMode: oldMode });
      throw e;
    }
  },

  // Data Sources Actions
  fetchDataSources: async () => {
    set((state) => ({
      dataSources: { ...state.dataSources, loading: true, error: null },
    }));

    try {
      const data = await api.fetchDataSources();
      set({
        dataSources: {
          sources: data.sources || [],
          availableSources: data.available_sources || [],
          loading: false,
          error: null,
        },
      });
    } catch (e) {
      console.error('Failed to fetch data sources:', e);
      set((state) => ({
        dataSources: {
          ...state.dataSources,
          loading: false,
          error: e.message || 'Failed to fetch data sources',
        },
      }));
    }
  },

  updateDataSourcePriority: async (dataType, priorities) => {
    const { dataSources } = get();
    const oldSources = dataSources.sources;

    // Optimistic update
    const updatedSources = dataSources.sources.map((source) =>
      source.type === dataType ? { ...source, priorities } : source
    );
    set((state) => ({
      dataSources: { ...state.dataSources, sources: updatedSources },
    }));

    try {
      await api.updateDataSourcePriority(dataType, priorities);
    } catch (e) {
      console.error(`Failed to update data source priority for ${dataType}:`, e);
      // Rollback on error
      set((state) => ({
        dataSources: { ...state.dataSources, sources: oldSources },
      }));
      throw e;
    }
  },

  // Move a source up in the priority list
  moveSourceUp: async (dataType, sourceIndex) => {
    if (sourceIndex <= 0) return;

    const { dataSources } = get();
    const sourceConfig = dataSources.sources.find((s) => s.type === dataType);
    if (!sourceConfig) return;

    const newPriorities = [...sourceConfig.priorities];
    [newPriorities[sourceIndex - 1], newPriorities[sourceIndex]] = [
      newPriorities[sourceIndex],
      newPriorities[sourceIndex - 1],
    ];

    await get().updateDataSourcePriority(dataType, newPriorities);
  },

  // Move a source down in the priority list
  moveSourceDown: async (dataType, sourceIndex) => {
    const { dataSources } = get();
    const sourceConfig = dataSources.sources.find((s) => s.type === dataType);
    if (!sourceConfig) return;

    if (sourceIndex >= sourceConfig.priorities.length - 1) return;

    const newPriorities = [...sourceConfig.priorities];
    [newPriorities[sourceIndex], newPriorities[sourceIndex + 1]] = [
      newPriorities[sourceIndex + 1],
      newPriorities[sourceIndex],
    ];

    await get().updateDataSourcePriority(dataType, newPriorities);
  },

  // Add a source to a data type's priority list
  addSource: async (dataType, source) => {
    const { dataSources } = get();
    const sourceConfig = dataSources.sources.find((s) => s.type === dataType);
    if (!sourceConfig) return;

    if (sourceConfig.priorities.includes(source)) return; // Already exists

    const newPriorities = [...sourceConfig.priorities, source];
    await get().updateDataSourcePriority(dataType, newPriorities);
  },

  // Remove a source from a data type's priority list
  removeSource: async (dataType, sourceIndex) => {
    const { dataSources } = get();
    const sourceConfig = dataSources.sources.find((s) => s.type === dataType);
    if (!sourceConfig) return;

    if (sourceConfig.priorities.length <= 1) return; // Must have at least one source

    const newPriorities = sourceConfig.priorities.filter((_, i) => i !== sourceIndex);
    await get().updateDataSourcePriority(dataType, newPriorities);
  },
}));
