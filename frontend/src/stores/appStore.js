/**
 * Application State Store (Zustand)
 *
 * Centralized state management for the Sentinel frontend application.
 * Manages:
 * - System status and health information
 * - Job execution tracking (running and completed jobs)
 * - Trade recommendations
 * - Event stream connection (Server-Sent Events)
 * - Polling fallback mechanism (when SSE unavailable)
 * - UI state (modals, loading states, active tab)
 * - User notifications
 *
 * Uses Zustand for lightweight, performant state management without
 * the boilerplate of Redux or Context API.
 */
import { create } from 'zustand';
import { notifications } from '@mantine/notifications';
import { api } from '../api/client';
import { handleEvent, clearAllDebounces } from './eventHandlers';
import { formatCurrency } from '../utils/formatters';

/**
 * Application store created with Zustand
 *
 * @type {Function} useAppStore - Hook to access store state and actions
 */
export const useAppStore = create((set, get) => ({
  // ============================================================================
  // System Status & Health
  // ============================================================================

  /**
   * Overall system status (health, uptime, etc.)
   * @type {Object}
   */
  status: {},

  /**
   * Tradernet broker connection status
   * @type {Object} { connected: boolean }
   */
  tradernet: { connected: false },

  /**
   * Latest Tradernet connection test result (null = not tested, true/false = test result)
   * @type {boolean|null}
   */
  tradernetConnectionStatus: null,

  /**
   * Market status information (open/closed, holidays, etc.)
   * @type {Object}
   */
  markets: {},

  // ============================================================================
  // Job Status Tracking
  // ============================================================================

  /**
   * Currently running background jobs
   * Map structure: jobId -> { jobType, status, description, startedAt, progress, stages }
   * @type {Object<string, Object>}
   */
  runningJobs: {},

  /**
   * Recently completed jobs (cleared after 4 seconds)
   * Map structure: jobId -> { jobType, status, description, completedAt, duration }
   * @type {Object<string, Object>}
   */
  completedJobs: {},

  /**
   * Last completed planner batch run (persists for debugging until next run starts)
   * Contains execution summary, stages, duration, and any errors
   * @type {Object|null}
   */
  lastPlannerRun: null,

  // ============================================================================
  // Recommendations
  // ============================================================================

  /**
   * Current trade recommendations from the planning system
   * @type {Array|null}
   */
  recommendations: null,

  // ============================================================================
  // Event Stream (Server-Sent Events)
  // ============================================================================

  /**
   * EventSource instance for Server-Sent Events connection
   * @type {EventSource|null}
   */
  eventStreamSource: null,

  /**
   * Number of consecutive SSE reconnection attempts
   * Used for exponential backoff calculation
   * @type {number}
   */
  eventStreamReconnectAttempts: 0,

  /**
   * Whether an SSE connection attempt is currently in progress
   * Prevents concurrent connection attempts
   * @type {boolean}
   */
  eventStreamConnecting: false,

  // ============================================================================
  // Polling Fallback Mechanism
  // ============================================================================

  /**
   * Interval ID for polling fallback (when SSE unavailable)
   * @type {number|null}
   */
  pollingIntervalId: null,

  /**
   * Interval ID for SSE reconnection attempts while in polling mode
   * @type {number|null}
   */
  sseRetryIntervalId: null,

  /**
   * Whether the app is currently in polling mode (SSE unavailable)
   * @type {boolean}
   */
  isPollingMode: false,

  /**
   * Maximum SSE failures before switching to polling mode
   * @type {number}
   */
  maxSseFailures: 5,

  // ============================================================================
  // UI State
  // ============================================================================

  /**
   * Currently active tab in the navigation
   * @type {string}
   */
  activeTab: 'next-actions',

  /**
   * Temporary message to display (replaced by notifications)
   * @type {string}
   */
  message: '',

  /**
   * Type of message (success, error, info)
   * @type {string}
   */
  messageType: 'success',

  // ============================================================================
  // Modal States
  // ============================================================================

  /**
   * Whether the "Add Security" modal is visible
   * @type {boolean}
   */
  showAddSecurityModal: false,

  /**
   * Whether the "Edit Security" modal is visible
   * @type {boolean}
   */
  showEditSecurityModal: false,

  /**
   * Whether the security chart modal is visible
   * @type {boolean}
   */
  showSecurityChart: false,

  /**
   * Whether the settings modal is visible
   * @type {boolean}
   */
  showSettingsModal: false,

  /**
   * Whether the planner management modal is visible
   * @type {boolean}
   */
  showPlannerManagementModal: false,

  /**
   * Closes the security chart modal and clears selected security
   */
  closeSecurityChartModal: () => set({
    showSecurityChart: false,
    selectedSecuritySymbol: null,
    selectedSecurityIsin: null,
  }),

  // ============================================================================
  // Selected Items
  // ============================================================================

  /**
   * Currently selected security symbol (for chart display)
   * @type {string|null}
   */
  selectedSecuritySymbol: null,

  /**
   * Currently selected security ISIN (for chart display)
   * @type {string|null}
   */
  selectedSecurityIsin: null,

  /**
   * Security being edited (passed to edit modal)
   * @type {Object|null}
   */
  editingSecurity: null,

  /**
   * Selected planner ID (for planner management)
   * @type {string}
   */
  selectedPlannerId: '',

  // ============================================================================
  // Loading States
  // ============================================================================

  /**
   * Loading state flags for various async operations
   * Prevents duplicate requests and shows loading indicators
   * @type {Object}
   */
  loading: {
    recommendations: false,  // Fetching trade recommendations
    scores: false,            // Refreshing security scores
    sync: false,              // Synchronizing data
    historical: false,        // Syncing historical data
    execute: false,           // Executing a trade
    geographySave: false,     // Saving geography targets
    industrySave: false,      // Saving industry targets
    securitySave: false,      // Saving security data
    refreshData: false,       // Refreshing all data
    logs: false,              // Fetching logs
    tradernetTest: false,     // Testing Tradernet connection
  },

  // ============================================================================
  // UI Actions
  // ============================================================================

  /**
   * Sets the active navigation tab
   * @param {string} tab - Tab identifier (e.g., 'next-actions', 'diversification')
   */
  setActiveTab: (tab) => set({ activeTab: tab }),

  /**
   * Shows a notification message to the user
   * Displays both in-store message state and Mantine notification toast
   *
   * @param {string} message - Message text to display
   * @param {string} type - Message type: 'success', 'error', or 'info'
   */
  showMessage: (message, type = 'success') => {
    set({ message, messageType: type });
    // Show Mantine notification toast
    notifications.show({
      title: type === 'error' ? 'Error' : type === 'success' ? 'Success' : 'Info',
      message,
      color: type === 'error' ? 'red' : type === 'success' ? 'green' : 'blue',
      autoClose: 3000,
    });
    // Clear message state after 3 seconds
    setTimeout(() => set({ message: '', messageType: 'success' }), 3000);
  },

  // ============================================================================
  // Modal Actions
  // ============================================================================

  /**
   * Opens the "Add Security" modal
   */
  openAddSecurityModal: () => set({ showAddSecurityModal: true }),

  /**
   * Closes the "Add Security" modal
   */
  closeAddSecurityModal: () => set({ showAddSecurityModal: false }),

  /**
   * Opens the "Edit Security" modal with a security pre-loaded
   * @param {Object} security - Security object to edit
   */
  openEditSecurityModal: (security) => set({
    showEditSecurityModal: true,
    editingSecurity: security,
    selectedSecurityIsin: security?.isin,
  }),

  /**
   * Closes the "Edit Security" modal and clears editing state
   */
  closeEditSecurityModal: () => set({
    showEditSecurityModal: false,
    editingSecurity: null,
    selectedSecurityIsin: null,
  }),

  /**
   * Opens the settings modal
   */
  openSettingsModal: () => set({ showSettingsModal: true }),

  /**
   * Closes the settings modal
   */
  closeSettingsModal: () => set({ showSettingsModal: false }),

  /**
   * Opens the security chart modal for a specific security
   * @param {string} symbol - Security symbol
   * @param {string} isin - Security ISIN
   */
  openSecurityChart: (symbol, isin) => set({
    showSecurityChart: true,
    selectedSecuritySymbol: symbol,
    selectedSecurityIsin: isin,
  }),

  /**
   * Opens the planner management modal
   */
  openPlannerManagementModal: () => set({ showPlannerManagementModal: true }),

  /**
   * Closes the planner management modal
   */
  closePlannerManagementModal: () => set({ showPlannerManagementModal: false }),

  // ============================================================================
  // Data Fetching Actions
  // ============================================================================

  /**
   * Fetches overall system status (health, uptime, etc.)
   * Updates the status state with the response
   */
  fetchStatus: async () => {
    try {
      const status = await api.fetchStatus();
      set({ status });
    } catch (e) {
      console.error('Failed to fetch status:', e);
    }
  },

  /**
   * Fetches Tradernet broker connection status
   * Updates the tradernet state with connection information
   */
  fetchTradernet: async () => {
    try {
      const tradernet = await api.fetchTradernet();
      set({ tradernet });
    } catch (e) {
      console.error('Failed to fetch tradernet status:', e);
    }
  },

  /**
   * Fetches market status information (open/closed, holidays)
   * Updates the markets state with market data
   */
  fetchMarkets: async () => {
    try {
      const data = await api.fetchMarkets();
      set({ markets: data.markets || {} });
    } catch (e) {
      console.error('Failed to fetch market status:', e);
    }
  },

  /**
   * Fetches current trade recommendations from the planning system
   * Sets loading state during fetch and handles errors gracefully
   */
  fetchRecommendations: async () => {
    set({ loading: { ...get().loading, recommendations: true } });
    try {
      const recommendations = await api.fetchRecommendations();
      set({ recommendations });
    } catch (e) {
      console.error('Failed to fetch recommendations:', e);
      set({ recommendations: null });
    } finally {
      set({ loading: { ...get().loading, recommendations: false } });
    }
  },

  /**
   * Executes the top recommendation (highest priority trade)
   * Shows success/error message and refreshes recommendations after execution
   */
  executeRecommendation: async () => {
    set({ loading: { ...get().loading, execute: true } });
    try {
      const result = await api.executeRecommendation();
      get().showMessage(`Executed: ${result.quantity} ${result.symbol} @ ${formatCurrency(result.price)}`, 'success');
      // Refresh recommendations to get updated list
      await get().fetchRecommendations();
    } catch (e) {
      get().showMessage('Failed to execute trade', 'error');
    } finally {
      set({ loading: { ...get().loading, execute: false } });
    }
  },

  /**
   * Tests the connection to Tradernet broker API
   * Updates tradernetConnectionStatus and shows success/error message
   */
  testTradernetConnection: async () => {
    set({
      loading: { ...get().loading, tradernetTest: true },
      tradernetConnectionStatus: null,
    });
    try {
      const result = await api.testTradernetConnection();
      const connected = result.connected || false;
      set({ tradernetConnectionStatus: connected });
      if (connected) {
        get().showMessage('Tradernet connection successful', 'success');
      } else {
        get().showMessage(`Tradernet connection failed: ${result.message || 'check credentials'}`, 'error');
      }
    } catch (e) {
      console.error('Error testing Tradernet connection:', e);
      set({ tradernetConnectionStatus: false });
      get().showMessage(`Failed to test Tradernet connection: ${e.message}`, 'error');
    } finally {
      set({ loading: { ...get().loading, tradernetTest: false } });
    }
  },

  /**
   * Triggers a planner batch job (generates and evaluates trade sequences)
   * Shows success/error message based on result
   */
  triggerPlannerBatch: async () => {
    try {
      await api.triggerPlannerBatch();
      get().showMessage('Planning job triggered', 'success');
    } catch (e) {
      console.error('Failed to trigger planner batch:', e);
      get().showMessage('Failed to trigger planning', 'error');
    }
  },

  // ============================================================================
  // Job Status Management
  // ============================================================================
  // These methods are called by event handlers when job events are received

  /**
   * Adds a new running job to the tracking system
   * Called when a job starts (via event stream)
   *
   * @param {Object} job - Job information
   * @param {string} job.jobId - Unique job identifier
   * @param {string} job.jobType - Type of job (e.g., 'planner_batch', 'sync_prices')
   * @param {string} job.status - Job status ('running')
   * @param {string} job.description - Human-readable job description
   * @param {number} job.startedAt - Timestamp when job started
   */
  addRunningJob: (job) => {
    const { jobId, jobType, status, description, startedAt } = job;
    set((state) => {
      const newState = {
        runningJobs: {
          ...state.runningJobs,
          [jobId]: {
            jobId,
            jobType,
            status,
            description,
            startedAt,
            progress: null,  // Will be updated by updateJobProgress
            stages: null,     // Will be updated by updateJobProgress
          },
        },
      };
      // Clear lastPlannerRun when a new planner_batch starts
      // This ensures we only keep the most recent completed run
      if (jobType === 'planner_batch') {
        newState.lastPlannerRun = null;
      }
      return newState;
    });
  },

  /**
   * Updates progress information for a running job
   * Called when job progress events are received (via event stream)
   *
   * @param {string} jobId - Unique job identifier
   * @param {Object} progress - Progress information
   * @param {number} progress.current - Current progress value
   * @param {number} progress.total - Total progress value
   * @param {string} progress.message - Progress message
   * @param {string} progress.phase - Current phase (optional)
   * @param {string} progress.sub_phase - Current sub-phase (optional)
   * @param {Object} progress.details - Additional progress details (may contain stages)
   */
  updateJobProgress: (jobId, progress) => {
    set((state) => {
      const job = state.runningJobs[jobId];
      if (!job) return state;  // Job not found, ignore update

      // Extract stages from details if present (for planner_batch jobs)
      // Stages provide detailed breakdown of planner execution phases
      const stages = progress?.details?.stages || job.stages;

      return {
        runningJobs: {
          ...state.runningJobs,
          [jobId]: {
            ...job,
            progress: {
              current: progress.current,
              total: progress.total,
              message: progress.message,
              phase: progress.phase || null,
              subPhase: progress.sub_phase || null,
              details: progress.details || null,
            },
            stages,  // Update stages if provided
          },
        },
      };
    });
  },

  /**
   * Marks a job as completed and moves it from running to completed
   * Called when a job finishes (via event stream)
   *
   * For planner_batch jobs, also saves execution summary to lastPlannerRun
   * Completed jobs are automatically cleared after 4 seconds (to show checkmark briefly)
   *
   * @param {string} jobId - Unique job identifier
   * @param {Object} completion - Completion information
   * @param {string} completion.status - Final status ('completed', 'failed', 'cancelled')
   * @param {string|null} completion.error - Error message if job failed
   * @param {number} completion.duration - Job duration in milliseconds
   */
  completeJob: (jobId, completion) => {
    const { status, error, duration } = completion;

    set((state) => {
      const job = state.runningJobs[jobId];
      if (!job) return state;  // Job not found, ignore

      // Move from running to completed
      const { [jobId]: completedJob, ...remainingRunning } = state.runningJobs;

      const completedJobData = {
        ...completedJob,
        status,
        error,
        duration,
        completedAt: Date.now(),
      };

      // Clear completed job after 4 seconds (allows checkmark to be visible briefly)
      setTimeout(() => {
        set((state) => {
          // eslint-disable-next-line no-unused-vars
          const { [jobId]: _, ...remaining } = state.completedJobs;
          return { completedJobs: remaining };
        });
      }, 4000);

      const newState = {
        runningJobs: remainingRunning,
        completedJobs: {
          ...state.completedJobs,
          [jobId]: completedJobData,
        },
      };

      // Save planner_batch runs to lastPlannerRun for debugging
      // This persists until the next planner_batch starts
      if (job.jobType === 'planner_batch') {
        newState.lastPlannerRun = {
          completedAt: new Date().toISOString(),
          totalDuration: duration,
          stages: job.stages || null,
          status,
          error: error || null,
          summary: job.progress?.details || null,
        };
      }

      return newState;
    });
  },

  // ============================================================================
  // Polling Fallback Mechanism
  // ============================================================================
  // When Server-Sent Events (SSE) are unavailable, fall back to HTTP polling
  // to keep the UI updated with latest data.

  /**
   * Starts polling mode as fallback when SSE is unavailable
   *
   * Polls critical data every 10 seconds and attempts SSE reconnection
   * every 60 seconds. Prevents duplicate polling if already active.
   */
  startPolling: () => {
    const { pollingIntervalId, isPollingMode } = get();

    // Prevent duplicate polling
    if (pollingIntervalId || isPollingMode) {
      return;
    }

    console.log('Starting polling mode (SSE unavailable)');
    set({ isPollingMode: true });

    // Poll critical data every 10 seconds
    // This keeps the UI updated even without real-time event stream
    const intervalId = setInterval(async () => {
      try {
        // Fetch app store data (status, tradernet, markets, recommendations)
        await get().fetchAll();

        // Fetch other stores' data (portfolio, securities, trades)
        // Dynamic imports prevent circular dependencies
        const { usePortfolioStore } = await import('./portfolioStore');
        const { useSecuritiesStore } = await import('./securitiesStore');
        const { useTradesStore } = await import('./tradesStore');

        await Promise.all([
          usePortfolioStore.getState().fetchAllocation(),
          usePortfolioStore.getState().fetchCashBreakdown(),
          useSecuritiesStore.getState().fetchSecurities(),
          useTradesStore.getState().fetchTrades(),
        ]);
      } catch (error) {
        console.error('Polling error:', error);
      }
    }, 10000);

    set({ pollingIntervalId: intervalId });

    // Try to reconnect to SSE every 60 seconds
    // If SSE becomes available, polling will stop automatically
    const retryIntervalId = setInterval(() => {
      console.log('Attempting to reconnect to SSE...');
      get().attemptSseReconnect();
    }, 60000);

    set({ sseRetryIntervalId: retryIntervalId });
  },

  /**
   * Stops polling mode and clears all polling intervals
   * Called when SSE connection is successfully established
   */
  stopPolling: () => {
    const { pollingIntervalId, sseRetryIntervalId } = get();

    if (pollingIntervalId) {
      clearInterval(pollingIntervalId);
      set({ pollingIntervalId: null });
    }

    if (sseRetryIntervalId) {
      clearInterval(sseRetryIntervalId);
      set({ sseRetryIntervalId: null });
    }

    set({ isPollingMode: false });
    console.log('Stopped polling mode');
  },

  /**
   * Attempts to reconnect to Server-Sent Events stream
   * Closes any existing connection and resets reconnect attempts before retrying
   */
  attemptSseReconnect: () => {
    const { eventStreamSource } = get();

    // Close existing failed connection if any
    if (eventStreamSource) {
      eventStreamSource.close();
    }

    // Reset reconnect attempts for fresh start
    set({ eventStreamReconnectAttempts: 0 });

    // Try to start SSE again
    get().startEventStream();
  },

  // ============================================================================
  // Unified Event Stream (Server-Sent Events)
  // ============================================================================
  // Establishes a Server-Sent Events connection for real-time updates from the backend.
  // Automatically falls back to polling if SSE is unavailable.

  /**
   * Starts the Server-Sent Events (SSE) connection for real-time updates
   *
   * Establishes a connection to /api/events/stream and handles:
   * - Job status updates (start, progress, completion)
   * - System state changes
   * - Trade execution notifications
   *
   * Automatically reconnects with exponential backoff on errors.
   * Switches to polling mode after maxSseFailures consecutive failures.
   */
  startEventStream: () => {
    const { eventStreamSource, eventStreamConnecting, isPollingMode } = get();

    // Prevent concurrent connection attempts
    if (eventStreamConnecting) {
      return;
    }

    set({ eventStreamConnecting: true });

    // Close any existing connection before creating a new one
    if (eventStreamSource) {
      eventStreamSource.close();
    }

    // Event stream endpoint (logs use HTTP polling instead of SSE)
    const url = '/api/events/stream';
    const eventSource = new EventSource(url);
    set({ eventStreamSource: eventSource, eventStreamConnecting: false });

    // Handle incoming events
    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        // Route event to appropriate handler (job events, state changes, etc.)
        handleEvent(data);
      } catch (e) {
        console.error('Failed to parse event stream message:', e);
      }
    };

    // Handle connection errors
    eventSource.onerror = (error) => {
      console.error('Event stream error:', error);

      // Close the connection and reset connecting flag
      eventSource.close();
      set({ eventStreamConnecting: false });

      // Get current attempts for delay calculation
      const currentAttempts = get().eventStreamReconnectAttempts;
      const maxFailures = get().maxSseFailures;

      // If we've failed too many times, switch to polling mode
      if (currentAttempts >= maxFailures && !isPollingMode) {
        console.warn(`SSE failed ${currentAttempts} times. Switching to polling mode.`);
        set({ eventStreamReconnectAttempts: currentAttempts + 1 });
        get().startPolling();
        return;
      }

      // Otherwise, reconnect with exponential backoff (max 30 seconds)
      // Exponential backoff: 1s, 2s, 4s, 8s, 16s, 30s (capped)
      const delay = Math.min(1000 * Math.pow(2, currentAttempts), 30000);
      setTimeout(() => {
        // Read fresh state to avoid stale closure
        const freshAttempts = get().eventStreamReconnectAttempts;
        set({ eventStreamReconnectAttempts: freshAttempts + 1 });
        get().startEventStream();
      }, delay);
    };

    // Reset reconnect attempts on successful connection
    eventSource.addEventListener('open', () => {
      console.log('SSE connection established');
      set({ eventStreamReconnectAttempts: 0 });

      // If we were in polling mode, stop polling (SSE is now available)
      if (isPollingMode) {
        console.log('SSE reconnected successfully. Stopping polling mode.');
        get().stopPolling();
      }
    });
  },

  /**
   * Stops the event stream connection and cleans up
   *
   * Closes SSE connection, stops polling if active, and clears
   * all pending debounced event handlers. Called on component unmount.
   */
  stopEventStream: () => {
    const { eventStreamSource } = get();
    if (eventStreamSource) {
      eventStreamSource.close();
      set({ eventStreamSource: null, eventStreamReconnectAttempts: 0 });
    }
    // Stop polling if active
    get().stopPolling();
    // Clear all pending debounced event handlers
    clearAllDebounces();
  },

  // ============================================================================
  // Batch Data Fetching
  // ============================================================================

  /**
   * Fetches all initial application data in parallel
   *
   * Called on app startup to load:
   * - System status
   * - Tradernet connection status
   * - Market status
   * - Trade recommendations
   */
  fetchAll: async () => {
    await Promise.all([
      get().fetchStatus(),
      get().fetchTradernet(),
      get().fetchMarkets(),
      get().fetchRecommendations(),
    ]);
  },
}));
