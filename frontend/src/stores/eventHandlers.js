import { useAppStore } from './appStore';
import { useSecuritiesStore } from './securitiesStore';
import { usePortfolioStore } from './portfolioStore';
import { useTradesStore } from './tradesStore';
import { useLogsStore } from './logsStore';
import { useSettingsStore } from './settingsStore';

// Store reference to SecurityChart refresh function
let securityChartRefreshFn = null;

export function setSecurityChartRefreshFn(fn) {
  securityChartRefreshFn = fn;
}

// Debounce utility
const debounceMap = new Map();

function debounce(key, fn, delay) {
  if (debounceMap.has(key)) {
    clearTimeout(debounceMap.get(key));
  }
  const timeoutId = setTimeout(() => {
    debounceMap.delete(key);
    fn();
  }, delay);
  debounceMap.set(key, timeoutId);
}

// Event-to-action mapper
export const eventHandlers = {
  // Securities events
  PRICE_UPDATED: (event) => {
    debounce('securities', () => {
      useSecuritiesStore.getState().fetchSecurities();
      useSecuritiesStore.getState().fetchSparklines();

      // Refresh security chart if open for this security
      if (securityChartRefreshFn) {
        const { selectedSecuritySymbol, selectedSecurityIsin } = useAppStore.getState();
        const eventIsin = event.data?.isin;
        const eventSymbol = event.data?.symbol;
        if (selectedSecuritySymbol &&
            (eventSymbol === selectedSecuritySymbol || eventIsin === selectedSecurityIsin)) {
          securityChartRefreshFn();
        }
      }
    }, 500);
  },

  SCORE_UPDATED: () => {
    debounce('securities', () => {
      useSecuritiesStore.getState().fetchSecurities();
    }, 500);
  },

  SECURITY_SYNCED: (event) => {
    useSecuritiesStore.getState().fetchSecurities();
    useSecuritiesStore.getState().fetchSparklines();

    // Refresh security chart if open for this security
    if (securityChartRefreshFn) {
      const { selectedSecuritySymbol, selectedSecurityIsin } = useAppStore.getState();
      const eventIsin = event.data?.isin;
      const eventSymbol = event.data?.symbol;
      if (selectedSecuritySymbol &&
          (eventSymbol === selectedSecuritySymbol || eventIsin === selectedSecurityIsin)) {
        securityChartRefreshFn();
      }
    }
  },

  SECURITY_ADDED: () => {
    useSecuritiesStore.getState().fetchSecurities();
  },

  // Portfolio events
  PORTFOLIO_CHANGED: () => {
    debounce('portfolio', () => {
      usePortfolioStore.getState().fetchAllocation();
      usePortfolioStore.getState().fetchCashBreakdown();
    }, 500);
  },

  TRADE_EXECUTED: () => {
    usePortfolioStore.getState().fetchAllocation();
    usePortfolioStore.getState().fetchCashBreakdown();
    useTradesStore.getState().fetchTrades();
  },

  DEPOSIT_PROCESSED: () => {
    usePortfolioStore.getState().fetchAllocation();
    usePortfolioStore.getState().fetchCashBreakdown();
  },

  DIVIDEND_CREATED: () => {
    usePortfolioStore.getState().fetchAllocation();
    usePortfolioStore.getState().fetchCashBreakdown();
  },

  CASH_UPDATED: () => {
    usePortfolioStore.getState().fetchCashBreakdown();
  },

  ALLOCATION_TARGETS_CHANGED: () => {
    usePortfolioStore.getState().fetchTargets();
    usePortfolioStore.getState().fetchAllocation();
  },

  // Recommendations events
  RECOMMENDATIONS_READY: () => {
    useAppStore.getState().fetchRecommendations();
  },

  PLAN_GENERATED: () => {
    useAppStore.getState().fetchRecommendations();
  },

  PLANNING_STATUS_UPDATED: (event) => {
    const status = event.data || {};
    useAppStore.getState().updatePlannerStatus(status);
  },

  // System status events
  SYSTEM_STATUS_CHANGED: () => {
    useAppStore.getState().fetchStatus();
  },

  TRADERNET_STATUS_CHANGED: () => {
    useAppStore.getState().fetchTradernet();
  },

  MARKETS_STATUS_CHANGED: () => {
    useAppStore.getState().fetchMarkets();
  },

  // Logs events
  LOG_FILE_CHANGED: (event) => {
    const { selectedLogFile } = useLogsStore.getState();
    if (event.data && event.data.log_file === selectedLogFile) {
      useLogsStore.getState().fetchLogs();
    }
  },

  // Settings events
  SETTINGS_CHANGED: () => {
    useSettingsStore.getState().fetchSettings();
  },

  PLANNER_CONFIG_CHANGED: () => {
    // Planner modal will refetch on open, or emit custom event
    // For now, just log it
    console.log('Planner config changed');
  },

  // Heartbeat - no action needed
  heartbeat: () => {
    // Heartbeat received, connection is alive
  },

  // Connected - no action needed
  connected: () => {
    console.log('Connected to unified event stream');
  },
};

// Handle an event from the SSE stream
export function handleEvent(event) {
  const { type, data } = event;

  if (!type) {
    console.warn('Received event without type:', event);
    return;
  }

  const handler = eventHandlers[type];
  if (handler) {
    try {
      handler({ type, data });
    } catch (error) {
      console.error(`Error handling event ${type}:`, error);
    }
  } else {
    console.debug(`No handler for event type: ${type}`);
  }
}
