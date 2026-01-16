/**
 * Centralized API client for Sentinel frontend
 *
 * This module provides a unified interface for all HTTP requests to the Sentinel
 * backend API. It handles:
 * - Request timeouts (default 30 seconds, configurable per request)
 * - JSON serialization/deserialization
 * - Error handling and response validation
 * - Consistent error messages
 *
 * All API methods return Promises that resolve to the JSON response data
 * or reject with an Error containing a descriptive message.
 */

// Default timeout for API requests (30 seconds)
// This prevents requests from hanging indefinitely if the server is unresponsive
const DEFAULT_TIMEOUT_MS = 30000;

/**
 * Fetches JSON data from the API with timeout and error handling
 *
 * This is the core function used by all API methods. It:
 * 1. Sets up an AbortController for request timeout
 * 2. Makes the HTTP request with JSON headers
 * 3. Handles timeout errors gracefully
 * 4. Validates and parses the JSON response
 *
 * @param {string} url - The API endpoint URL (relative paths are supported)
 * @param {Object} options - Fetch options (method, body, headers, timeout)
 * @param {string} options.method - HTTP method (GET, POST, PUT, DELETE)
 * @param {string|Object} options.body - Request body (will be JSON stringified if object)
 * @param {Object} options.headers - Additional headers to include
 * @param {number} options.timeout - Request timeout in milliseconds (default: 30000)
 * @returns {Promise<any>} Promise that resolves to the parsed JSON response
 * @throws {Error} If request fails, times out, or returns invalid JSON
 */
async function fetchJSON(url, options = {}) {
  // Create abort controller for timeout
  // This allows us to cancel the request if it takes too long
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), options.timeout || DEFAULT_TIMEOUT_MS);

  try {
    const res = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      signal: controller.signal,  // Attach abort signal for timeout
    });

    clearTimeout(timeoutId);
    return await handleResponse(res, url);
  } catch (error) {
    clearTimeout(timeoutId);
    // Check if error was due to timeout
    if (error.name === 'AbortError') {
      throw new Error(`Request timeout after ${options.timeout || DEFAULT_TIMEOUT_MS}ms: ${url}`);
    }
    throw error;
  }
}

/**
 * Handles HTTP response, extracting error messages or parsing JSON
 *
 * This function processes the response from the fetch call:
 * - Checks if response is OK (status 200-299)
 * - Extracts error messages from failed responses (tries multiple field names)
 * - Parses and validates JSON from successful responses
 * - Handles edge cases (null responses, parse errors)
 *
 * @param {Response} res - The fetch Response object
 * @param {string} url - The original URL (for error messages)
 * @returns {Promise<any>} Promise that resolves to parsed JSON data
 * @throws {Error} If response is not OK or JSON parsing fails
 */
async function handleResponse(res, url) {
  // Check if response indicates an error (status code outside 200-299)
  if (!res.ok) {
    let errorMessage = `Request failed with status ${res.status}`;
    try {
      // Try to extract error message from response body
      // Backend may use different field names: detail, message, or error
      const errorData = await res.json();
      errorMessage = errorData.detail || errorData.message || errorData.error || errorMessage;
    } catch (e) {
      // If response body isn't JSON, use status text
      errorMessage = res.statusText || errorMessage;
    }
    throw new Error(errorMessage);
  }

  // Parse JSON from successful response
  try {
    const data = await res.json();
    // Basic validation - ensure we got valid JSON (not null/undefined)
    // Some endpoints may legitimately return null, so we warn but don't error
    if (data === null || data === undefined) {
      console.warn(`API returned null/undefined for ${url}`);
      return null;
    }
    return data;
  } catch (e) {
    // If JSON parsing fails, log error and throw descriptive error
    console.error(`Failed to parse JSON response from ${url}:`, e);
    throw new Error(`Invalid JSON response from ${url}`);
  }
}

/**
 * API client object - provides methods for all backend endpoints
 *
 * All methods return Promises that resolve to the JSON response data.
 * Methods are organized by functional area (status, logs, jobs, etc.)
 */
export const api = {
  // Base HTTP methods - generic wrappers for common HTTP operations
  /**
   * GET request - fetches data from the API
   * @param {string} url - API endpoint URL
   * @returns {Promise<any>} Parsed JSON response
   */
  get: (url) => fetchJSON(url),

  /**
   * POST request - sends data to the API
   * @param {string} url - API endpoint URL
   * @param {Object} data - Data to send (will be JSON stringified)
   * @returns {Promise<any>} Parsed JSON response
   */
  post: (url, data) => fetchJSON(url, {
    method: 'POST',
    body: data ? JSON.stringify(data) : undefined,
  }),

  /**
   * PUT request - updates data on the API
   * @param {string} url - API endpoint URL
   * @param {Object} data - Data to update (will be JSON stringified)
   * @returns {Promise<any>} Parsed JSON response
   */
  put: (url, data) => fetchJSON(url, {
    method: 'PUT',
    body: JSON.stringify(data),
  }),

  /**
   * DELETE request - deletes data on the API
   * @param {string} url - API endpoint URL
   * @returns {Promise<any>} Parsed JSON response
   */
  delete: (url) => fetchJSON(url, { method: 'DELETE' }),

  // ============================================================================
  // System Status & Health
  // ============================================================================

  /**
   * Fetches overall system status (health, uptime, etc.)
   * @returns {Promise<Object>} System status information
   */
  fetchStatus: () => fetchJSON('/api/system/status'),

  /**
   * Tests Tradernet broker connection and returns connection status
   * @returns {Promise<Object>} Tradernet connection status
   */
  fetchTradernet: () => fetchJSON('/api/system/tradernet'),

  /**
   * Triggers immediate price synchronization from broker
   * @returns {Promise<Object>} Sync operation result
   */
  syncPrices: () => fetchJSON('/api/system/sync/prices', { method: 'POST' }),

  /**
   * Triggers historical price data synchronization
   * @returns {Promise<Object>} Sync operation result
   */
  syncHistorical: () => fetchJSON('/api/system/sync/historical', { method: 'POST' }),

  // ============================================================================
  // System Logs
  // ============================================================================

  /**
   * Fetches system logs with optional filtering
   * @param {number} lines - Number of log lines to fetch (default: 100)
   * @param {string} level - Optional log level filter (e.g., 'error', 'warn', 'info')
   * @param {string} search - Optional search term to filter logs
   * @returns {Promise<Array>} Array of log entries
   */
  fetchLogs: (lines = 100, level = null, search = null) => {
    const params = new URLSearchParams({ lines: lines.toString() });
    if (level) params.append('level', level);
    if (search) params.append('search', search);
    return fetchJSON(`/api/system/logs?${params}`);
  },

  /**
   * Fetches only error-level logs
   * @param {number} lines - Number of error log lines to fetch (default: 50)
   * @returns {Promise<Array>} Array of error log entries
   */
  fetchErrorLogs: (lines = 50) => {
    const params = new URLSearchParams({ lines: lines.toString() });
    return fetchJSON(`/api/system/logs/errors?${params}`);
  },

  /**
   * Fetches list of available log files
   * @returns {Promise<Array>} Array of log file names
   */
  fetchAvailableLogFiles: () => fetchJSON('/api/system/logs/list'),

  // ============================================================================
  // Background Jobs - Composite Jobs (multiple steps)
  // ============================================================================

  /**
   * Triggers a complete sync cycle (prices, trades, portfolio, cash flows)
   * @returns {Promise<Object>} Job execution result
   */
  triggerSyncCycle: () => fetchJSON('/api/system/jobs/sync-cycle', { method: 'POST' }),

  /**
   * Triggers daily pipeline (all daily maintenance tasks)
   * @returns {Promise<Object>} Job execution result
   */
  triggerDailyPipeline: () => fetchJSON('/api/system/sync/daily-pipeline', { method: 'POST' }),

  /**
   * Triggers dividend reinvestment process (find, group, create recommendations, execute)
   * @returns {Promise<Object>} Job execution result
   */
  triggerDividendReinvestment: () => fetchJSON('/api/system/jobs/dividend-reinvestment', { method: 'POST' }),

  /**
   * Triggers system health check (database integrity, WAL checkpoints, etc.)
   * @returns {Promise<Object>} Health check results
   */
  triggerHealthCheck: () => fetchJSON('/api/system/jobs/health-check', { method: 'POST' }),

  /**
   * Triggers event-based trading (executes pending recommendations based on events)
   * @returns {Promise<Object>} Job execution result
   */
  triggerEventBasedTrading: () => fetchJSON('/api/system/jobs/event-based-trading', { method: 'POST' }),

  /**
   * Triggers tag update job (recalculates security tags based on current data)
   * @returns {Promise<Object>} Job execution result
   */
  triggerTagUpdate: () => fetchJSON('/api/system/jobs/tag-update', { method: 'POST' }),

  /**
   * Triggers Tradernet metadata synchronization (security metadata from broker)
   * @returns {Promise<Object>} Job execution result
   */
  triggerTradernetMetadataSync: () => fetchJSON('/api/system/jobs/tradernet-metadata-sync', { method: 'POST' }),

  /**
   * Triggers hard update (full system restart and deployment)
   * @returns {Promise<Object>} Deployment result
   */
  hardUpdate: () => fetchJSON('/api/system/deployment/hard-update', { method: 'POST' }),

  // ============================================================================
  // Background Jobs - Individual Sync Jobs
  // ============================================================================

  /**
   * Triggers trade synchronization from broker
   * @returns {Promise<Object>} Sync result with number of trades synced
   */
  triggerSyncTrades: () => fetchJSON('/api/system/jobs/sync-trades', { method: 'POST' }),

  /**
   * Triggers cash flow synchronization (deposits, withdrawals, dividends)
   * @returns {Promise<Object>} Sync result with number of cash flows synced
   */
  triggerSyncCashFlows: () => fetchJSON('/api/system/jobs/sync-cash-flows', { method: 'POST' }),

  /**
   * Triggers portfolio synchronization (current positions from broker)
   * @returns {Promise<Object>} Sync result with portfolio state
   */
  triggerSyncPortfolio: () => fetchJSON('/api/system/jobs/sync-portfolio', { method: 'POST' }),

  /**
   * Triggers price synchronization (current market prices)
   * @returns {Promise<Object>} Sync result with number of prices updated
   */
  triggerSyncPrices: () => fetchJSON('/api/system/jobs/sync-prices', { method: 'POST' }),

  /**
   * Checks for negative account balances (data integrity check)
   * @returns {Promise<Object>} Check result with any issues found
   */
  triggerCheckNegativeBalances: () => fetchJSON('/api/system/jobs/check-negative-balances', { method: 'POST' }),

  /**
   * Updates LED display ticker with current portfolio information
   * @returns {Promise<Object>} Update result
   */
  triggerUpdateDisplayTicker: () => fetchJSON('/api/system/jobs/update-display-ticker', { method: 'POST' }),

  // ============================================================================
  // Background Jobs - Individual Planning Jobs
  // ============================================================================

  /**
   * Generates portfolio state hash (used for change detection)
   * @returns {Promise<Object>} Hash value and state snapshot
   */
  triggerGeneratePortfolioHash: () => fetchJSON('/api/system/jobs/generate-portfolio-hash', { method: 'POST' }),

  /**
   * Calculates optimizer weights for portfolio optimization
   * @returns {Promise<Object>} Optimizer weights for different strategies
   */
  triggerGetOptimizerWeights: () => fetchJSON('/api/system/jobs/get-optimizer-weights', { method: 'POST' }),

  /**
   * Builds opportunity context (market conditions, opportunities, etc.)
   * @returns {Promise<Object>} Opportunity context data
   */
  triggerBuildOpportunityContext: () => fetchJSON('/api/system/jobs/build-opportunity-context', { method: 'POST' }),

  /**
   * Creates a trade plan (generates trade sequences for evaluation)
   * @returns {Promise<Object>} Trade plan with sequences
   */
  triggerCreateTradePlan: () => fetchJSON('/api/system/jobs/create-trade-plan', { method: 'POST' }),

  /**
   * Stores evaluated recommendations in the database
   * @returns {Promise<Object>} Storage result with number of recommendations stored
   */
  triggerStoreRecommendations: () => fetchJSON('/api/system/jobs/store-recommendations', { method: 'POST' }),

  // ============================================================================
  // Background Jobs - Individual Dividend Jobs
  // ============================================================================

  /**
   * Finds unreinvested dividends in the portfolio
   * @returns {Promise<Array>} Array of unreinvested dividend records
   */
  triggerGetUnreinvestedDividends: () => fetchJSON('/api/system/jobs/get-unreinvested-dividends', { method: 'POST' }),

  /**
   * Groups dividends by security symbol for batch processing
   * @returns {Promise<Object>} Grouped dividends by symbol
   */
  triggerGroupDividendsBySymbol: () => fetchJSON('/api/system/jobs/group-dividends-by-symbol', { method: 'POST' }),

  /**
   * Checks dividend yields for all securities
   * @returns {Promise<Object>} Dividend yield data
   */
  triggerCheckDividendYields: () => fetchJSON('/api/system/jobs/check-dividend-yields', { method: 'POST' }),

  /**
   * Creates dividend reinvestment recommendations
   * @returns {Promise<Object>} Recommendations created
   */
  triggerCreateDividendRecommendations: () => fetchJSON('/api/system/jobs/create-dividend-recommendations', { method: 'POST' }),

  /**
   * Sets pending bonus amounts for dividend reinvestment
   * @returns {Promise<Object>} Update result
   */
  triggerSetPendingBonuses: () => fetchJSON('/api/system/jobs/set-pending-bonuses', { method: 'POST' }),

  /**
   * Executes dividend reinvestment trades
   * @returns {Promise<Object>} Execution result with trades executed
   */
  triggerExecuteDividendTrades: () => fetchJSON('/api/system/jobs/execute-dividend-trades', { method: 'POST' }),

  // ============================================================================
  // Background Jobs - Individual Health Check Jobs
  // ============================================================================

  /**
   * Checks integrity of core databases (universe, config, ledger, portfolio)
   * @returns {Promise<Object>} Health check results for core databases
   */
  triggerCheckCoreDatabases: () => fetchJSON('/api/system/jobs/check-core-databases', { method: 'POST' }),

  /**
   * Checks integrity of history databases (price history, etc.)
   * @returns {Promise<Object>} Health check results for history databases
   */
  triggerCheckHistoryDatabases: () => fetchJSON('/api/system/jobs/check-history-databases', { method: 'POST' }),

  /**
   * Checks WAL (Write-Ahead Log) checkpoints for all databases
   * @returns {Promise<Object>} WAL checkpoint status
   */
  triggerCheckWALCheckpoints: () => fetchJSON('/api/system/jobs/check-wal-checkpoints', { method: 'POST' }),

  // ============================================================================
  // Portfolio Allocation
  // ============================================================================

  /**
   * Fetches current portfolio allocation (actual allocation percentages)
   * @returns {Promise<Object>} Current allocation by geography, industry, etc.
   */
  fetchAllocation: () => fetchJSON('/api/allocation/current'),

  /**
   * Fetches allocation targets (desired allocation percentages)
   * @returns {Promise<Object>} Target allocation by geography, industry, etc.
   */
  fetchTargets: () => fetchJSON('/api/allocation/targets'),

  /**
   * Saves geography allocation targets
   * @param {Object} targets - Geography targets object (e.g., { "US": 0.6, "EU": 0.4 })
   * @returns {Promise<Object>} Update result
   */
  saveGeographyTargets: (targets) => fetchJSON('/api/allocation/targets/geography', {
    method: 'PUT',
    body: JSON.stringify({ targets }),
  }),

  /**
   * Saves industry allocation targets
   * @param {Object} targets - Industry targets object (e.g., { "Technology": 0.3, "Finance": 0.2 })
   * @returns {Promise<Object>} Update result
   */
  saveIndustryTargets: (targets) => fetchJSON('/api/allocation/targets/industry', {
    method: 'PUT',
    body: JSON.stringify({ targets }),
  }),

  // ============================================================================
  // Security Management
  // ============================================================================

  /**
   * Fetches all securities in the investment universe
   * @returns {Promise<Array>} Array of security objects
   */
  fetchSecurities: () => fetchJSON('/api/securities'),

  /**
   * Creates a new security in the universe
   * @param {Object} data - Security data (symbol, ISIN, name, etc.)
   * @returns {Promise<Object>} Created security object
   */
  createSecurity: (data) => fetchJSON('/api/securities', {
    method: 'POST',
    body: JSON.stringify(data),
  }),

  /**
   * Adds a security by identifier (symbol or ISIN) - looks up metadata from broker
   * @param {Object} data - Identifier data (symbol or ISIN)
   * @returns {Promise<Object>} Created security object with broker metadata
   */
  addSecurityByIdentifier: (data) => fetchJSON('/api/securities/add-by-identifier', {
    method: 'POST',
    body: JSON.stringify(data),
  }),

  /**
   * Updates an existing security
   * @param {string} isin - Security ISIN identifier
   * @param {Object} data - Updated security data
   * @returns {Promise<Object>} Updated security object
   */
  updateSecurity: (isin, data) => fetchJSON(`/api/securities/${isin}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  }),

  /**
   * Deletes a security from the universe
   * @param {string} isin - Security ISIN identifier
   * @returns {Promise<Object>} Deletion result
   */
  deleteSecurity: (isin) => fetchJSON(`/api/securities/${isin}`, { method: 'DELETE' }),

  /**
   * Refreshes the score for a single security
   * @param {string} isin - Security ISIN identifier
   * @returns {Promise<Object>} Updated security with new score
   */
  refreshScore: (isin) => fetchJSON(`/api/securities/${isin}/refresh`, { method: 'POST' }),

  /**
   * Refreshes scores for all securities in the universe
   * @returns {Promise<Object>} Result with number of scores refreshed
   */
  refreshAllScores: () => fetchJSON('/api/securities/refresh-all', { method: 'POST' }),

  // ============================================================================
  // Security Overrides (User Customizations)
  // ============================================================================
  // Overrides allow users to customize security properties without modifying
  // the base security data. Overrides take precedence over broker-synced data.

  /**
   * Fetches all overrides for a security
   * @param {string} isin - Security ISIN identifier
   * @returns {Promise<Object>} Object with override field names as keys
   */
  getSecurityOverrides: (isin) => fetchJSON(`/api/securities/${isin}/overrides`),

  /**
   * Sets an override value for a security field
   * @param {string} isin - Security ISIN identifier
   * @param {string} field - Field name to override (e.g., 'name', 'sector')
   * @param {any} value - Override value
   * @returns {Promise<Object>} Update result
   */
  setSecurityOverride: (isin, field, value) => fetchJSON(`/api/securities/${isin}/overrides/${field}`, {
    method: 'PUT',
    body: JSON.stringify({ value: String(value) }),
  }),

  /**
   * Deletes an override for a security field (reverts to broker-synced value)
   * @param {string} isin - Security ISIN identifier
   * @param {string} field - Field name to remove override for
   * @returns {Promise<Object>} Deletion result
   */
  deleteSecurityOverride: (isin, field) => fetchJSON(`/api/securities/${isin}/overrides/${field}`, {
    method: 'DELETE',
  }),

  // ============================================================================
  // Trading & Recommendations
  // ============================================================================

  /**
   * Fetches all executed trades
   * @returns {Promise<Array>} Array of trade objects
   */
  fetchTrades: () => fetchJSON('/api/trades'),

  /**
   * Fetches pending orders (orders submitted but not yet executed)
   * @returns {Promise<Array>} Array of pending order objects
   */
  fetchPendingOrders: () => fetchJSON('/api/system/pending-orders'),

  /**
   * Fetches current trade recommendations (generated by planning system)
   * @returns {Promise<Array>} Array of recommendation objects
   */
  fetchRecommendations: () => fetchJSON('/api/trades/recommendations'),

  /**
   * Executes the top recommendation (highest priority/score)
   * @returns {Promise<Object>} Execution result with order details
   */
  executeRecommendation: () => fetchJSON('/api/trades/recommendations/execute', { method: 'POST' }),

  // ============================================================================
  // Charts & Visualization
  // ============================================================================

  /**
   * Fetches chart data for a security
   * @param {string} isin - Security ISIN identifier
   * @param {string} range - Time range (e.g., '1Y', '6M', '3M', '1M')
   * @param {string} source - Data source ('tradernet' or 'historical')
   * @returns {Promise<Array>} Array of price data points
   */
  fetchSecurityChart: (isin, range = '1Y', source = 'tradernet') => {
    const params = new URLSearchParams({ range, source });
    return fetchJSON(`/api/charts/securities/${isin}?${params}`);
  },

  /**
   * Fetches sparkline data for all securities (mini charts for quick overview)
   * @param {string} period - Time period (e.g., '1Y', '6M')
   * @returns {Promise<Object>} Object with ISIN as keys and sparkline data as values
   */
  fetchSparklines: (period = '1Y') => fetchJSON(`/api/charts/sparklines?period=${period}`),

  // ============================================================================
  // Settings & Configuration
  // ============================================================================

  /**
   * Fetches all application settings
   * @returns {Promise<Object>} Object with setting keys and values
   */
  fetchSettings: () => fetchJSON('/api/settings'),

  /**
   * Updates a single setting value
   * Automatically converts numeric strings to numbers for numeric settings.
   * String settings are kept as strings.
   *
   * @param {string} key - Setting key name
   * @param {string|number} value - Setting value (will be converted if needed)
   * @returns {Promise<Object>} Update result
   */
  updateSetting: (key, value) => {
    // List of settings that should remain as strings (not converted to numbers)
    const stringSettings = [
      'tradernet_api_key',
      'tradernet_api_secret',
      'trading_mode',
      'display_mode',
      'security_table_visible_columns',
      'r2_account_id',
      'r2_access_key_id',
      'r2_secret_access_key',
      'r2_bucket_name',
      'r2_backup_schedule',
      'github_token',
    ];
    // Convert to number if not a string setting, otherwise keep as string
    const finalValue = stringSettings.includes(key) ? value : parseFloat(value);
    return fetchJSON(`/api/settings/${key}`, {
      method: 'PUT',
      body: JSON.stringify({ value: finalValue }),
    });
  },

  /**
   * Fetches current trading mode (e.g., 'paper', 'live')
   * @returns {Promise<Object>} Trading mode information
   */
  getTradingMode: () => fetchJSON('/api/settings/trading-mode'),

  /**
   * Toggles trading mode between paper and live trading
   * @returns {Promise<Object>} Updated trading mode
   */
  toggleTradingMode: () => fetchJSON('/api/settings/trading-mode', { method: 'POST' }),

  /**
   * Restarts the Sentinel service
   * @returns {Promise<Object>} Restart result
   */
  restartService: () => fetchJSON('/api/settings/restart-service', { method: 'POST' }),

  /**
   * Restarts the entire system
   * @returns {Promise<Object>} Restart result
   */
  restartSystem: () => fetchJSON('/api/settings/restart', { method: 'POST' }),

  /**
   * Resets all caches (forces fresh data on next requests)
   * @returns {Promise<Object>} Cache reset result
   */
  resetCache: () => fetchJSON('/api/settings/reset-cache', { method: 'POST' }),

  /**
   * Uploads Arduino sketch to the MCU (long timeout for upload process)
   * @returns {Promise<Object>} Upload result
   */
  uploadSketch: () => fetchJSON('/api/system/mcu/upload-sketch', { method: 'POST', timeout: 120000 }),

  /**
   * Reschedules all background jobs (reloads job schedules)
   * @returns {Promise<Object>} Reschedule result
   */
  rescheduleJobs: () => fetchJSON('/api/settings/reschedule-jobs', { method: 'POST' }),

  /**
   * Tests connection to Tradernet broker API
   * @returns {Promise<Object>} Connection test result
   */
  testTradernetConnection: () => fetchJSON('/api/system/tradernet'),

  // ============================================================================
  // R2 Backups (Cloudflare R2 Storage)
  // ============================================================================

  /**
   * Lists all available R2 backups
   * @returns {Promise<Array>} Array of backup file information
   */
  listR2Backups: () => fetchJSON('/api/backups/r2'),

  /**
   * Creates a new backup and uploads it to R2
   * @returns {Promise<Object>} Backup creation result with filename
   */
  createR2Backup: () => fetchJSON('/api/backups/r2', { method: 'POST' }),

  /**
   * Tests connection to R2 storage (validates credentials)
   * @returns {Promise<Object>} Connection test result
   */
  testR2Connection: () => fetchJSON('/api/backups/r2/test', { method: 'POST' }),

  /**
   * Downloads a backup file from R2 and triggers browser download
   * Note: This method handles file download differently (not JSON response)
   *
   * @param {string} filename - Backup filename to download
   * @returns {Promise<void>} Resolves when download is triggered
   * @throws {Error} If download fails
   */
  downloadR2Backup: async (filename) => {
    // Download returns a file, not JSON - use regular fetch
    const response = await fetch(`/api/backups/r2/${encodeURIComponent(filename)}/download`);
    if (!response.ok) {
      throw new Error(`Failed to download backup: ${response.statusText}`);
    }
    // Convert response to blob and trigger browser download
    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    // Clean up
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);
  },

  /**
   * Deletes a backup file from R2
   * @param {string} filename - Backup filename to delete
   * @returns {Promise<Object>} Deletion result
   */
  deleteR2Backup: (filename) => fetchJSON(`/api/backups/r2/${encodeURIComponent(filename)}`, { method: 'DELETE' }),

  /**
   * Stages a backup for restore (prepares restore, executes on next startup)
   * @param {string} filename - Backup filename to restore
   * @returns {Promise<Object>} Staging result
   */
  stageR2Restore: (filename) => fetchJSON('/api/backups/r2/restore', {
    method: 'POST',
    body: JSON.stringify({ filename }),
  }),

  /**
   * Cancels a staged restore (removes staged restore before startup)
   * @returns {Promise<Object>} Cancellation result
   */
  cancelR2Restore: () => fetchJSON('/api/backups/r2/restore/staged', { method: 'DELETE' }),

  // ============================================================================
  // Planning System
  // ============================================================================

  /**
   * Triggers planner batch job (generates and evaluates trade sequences)
   * @returns {Promise<Object>} Job execution result
   */
  triggerPlannerBatch: () => fetchJSON('/api/jobs/planner-batch', { method: 'POST' }),

  // ============================================================================
  // Planner Configuration
  // ============================================================================
  // The planning system uses a single configuration object (not multiple configs)

  /**
   * Fetches the current planner configuration
   * @returns {Promise<Object>} Planner configuration object
   */
  fetchPlannerConfig: () => fetchJSON('/api/planning/config'),

  /**
   * Updates the planner configuration
   * @param {Object} config - New configuration object
   * @param {string} changedBy - User or system that made the change
   * @param {string} changeNote - Optional note describing the change
   * @returns {Promise<Object>} Updated configuration
   */
  updatePlannerConfig: (config, changedBy, changeNote) => fetchJSON('/api/planning/config', {
    method: 'PUT',
    body: JSON.stringify({ config, changed_by: changedBy, change_note: changeNote }),
  }),

  /**
   * Deletes the planner configuration (resets to defaults)
   * @returns {Promise<Object>} Deletion result
   */
  deletePlannerConfig: () => fetchJSON('/api/planning/config', { method: 'DELETE' }),

  /**
   * Validates planner configuration without saving
   * @returns {Promise<Object>} Validation result with any errors
   */
  validatePlannerConfig: () => fetchJSON('/api/planning/config/validate', {
    method: 'POST',
  }),

  // ============================================================================
  // Market Information
  // ============================================================================

  /**
   * Fetches market status information (open/closed, holidays, etc.)
   * @returns {Promise<Object>} Market status data
   */
  fetchMarkets: () => fetchJSON('/api/system/markets'),

  // ============================================================================
  // Portfolio Information
  // ============================================================================

  /**
   * Fetches cash breakdown (available cash, pending, allocated, etc.)
   * @returns {Promise<Object>} Cash breakdown by category
   */
  fetchCashBreakdown: () => fetchJSON('/api/portfolio/cash-breakdown'),

  // ============================================================================
  // Available Options (for autocomplete in forms)
  // ============================================================================

  /**
   * Fetches list of available geographies for allocation targets
   * @returns {Promise<Array>} Array of geography names
   */
  fetchAvailableGeographies: () => fetchJSON('/api/allocation/available/geographies'),

  /**
   * Fetches list of available industries for allocation targets
   * @returns {Promise<Array>} Array of industry names
   */
  fetchAvailableIndustries: () => fetchJSON('/api/allocation/available/industries'),
};
