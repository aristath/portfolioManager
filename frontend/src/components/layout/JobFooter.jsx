/**
 * Job Footer Component
 * 
 * Provides a collapsible panel for manually triggering background jobs.
 * Organized by category (Sync, Planning, Dividends, Health Check, Composite).
 * 
 * Features:
 * - Collapsible panel (click title to expand/collapse)
 * - Jobs organized by category
 * - Loading states for each job button
 * - Success/error messages that auto-dismiss after 5 seconds
 * - Special handling for "hard-update" job (refreshes page on success)
 * 
 * Used for debugging and manual job execution during development/maintenance.
 */
import { Paper, Group, Button, Text, Stack, Divider, Title, Collapse } from '@mantine/core';
import { api } from '../../api/client';
import { useState } from 'react';

/**
 * Job categories with their associated jobs
 * 
 * Each job has:
 * - id: Unique identifier
 * - name: Display name
 * - api: API method name to call (from api client)
 * 
 * @type {Array<Object>}
 */
const jobCategories = [
  {
    name: 'Sync Jobs',
    jobs: [
      { id: 'sync-trades', name: 'Sync Trades', api: 'triggerSyncTrades' },
      { id: 'sync-cash-flows', name: 'Sync Cash Flows', api: 'triggerSyncCashFlows' },
      { id: 'sync-portfolio', name: 'Sync Portfolio', api: 'triggerSyncPortfolio' },
      { id: 'sync-prices', name: 'Sync Prices', api: 'triggerSyncPrices' },
      { id: 'check-negative-balances', name: 'Check Negative Balances', api: 'triggerCheckNegativeBalances' },
      { id: 'update-display-ticker', name: 'Update Display Ticker', api: 'triggerUpdateDisplayTicker' },
    ],
  },
  {
    name: 'Planning Jobs',
    jobs: [
      { id: 'generate-portfolio-hash', name: 'Generate Portfolio Hash', api: 'triggerGeneratePortfolioHash' },
      { id: 'get-optimizer-weights', name: 'Get Optimizer Weights', api: 'triggerGetOptimizerWeights' },
      { id: 'build-opportunity-context', name: 'Build Opportunity Context', api: 'triggerBuildOpportunityContext' },
      { id: 'create-trade-plan', name: 'Create Trade Plan', api: 'triggerCreateTradePlan' },
      { id: 'store-recommendations', name: 'Store Recommendations', api: 'triggerStoreRecommendations' },
    ],
  },
  {
    name: 'Dividend Jobs',
    jobs: [
      { id: 'get-unreinvested-dividends', name: 'Get Unreinvested Dividends', api: 'triggerGetUnreinvestedDividends' },
      { id: 'group-dividends-by-symbol', name: 'Group Dividends By Symbol', api: 'triggerGroupDividendsBySymbol' },
      { id: 'check-dividend-yields', name: 'Check Dividend Yields', api: 'triggerCheckDividendYields' },
      { id: 'create-dividend-recommendations', name: 'Create Dividend Recommendations', api: 'triggerCreateDividendRecommendations' },
      { id: 'set-pending-bonuses', name: 'Set Pending Bonuses', api: 'triggerSetPendingBonuses' },
      { id: 'execute-dividend-trades', name: 'Execute Dividend Trades', api: 'triggerExecuteDividendTrades' },
    ],
  },
  {
    name: 'Health Check Jobs',
    jobs: [
      { id: 'check-core-databases', name: 'Check Core Databases', api: 'triggerCheckCoreDatabases' },
      { id: 'check-history-databases', name: 'Check History Databases', api: 'triggerCheckHistoryDatabases' },
      { id: 'check-wal-checkpoints', name: 'Check WAL Checkpoints', api: 'triggerCheckWALCheckpoints' },
    ],
  },
  {
    name: 'Composite Jobs',
    jobs: [
      { id: 'sync-cycle', name: 'Sync Cycle', api: 'triggerSyncCycle' },
      { id: 'daily-pipeline', name: 'Daily Pipeline', api: 'triggerDailyPipeline' },
      { id: 'dividend-reinvestment', name: 'Dividend Reinvestment', api: 'triggerDividendReinvestment' },
      { id: 'health-check', name: 'Health Check', api: 'triggerHealthCheck' },
      { id: 'planner-batch', name: 'Planner Batch', api: 'triggerPlannerBatch' },
      { id: 'event-based-trading', name: 'Event Based Trading', api: 'triggerEventBasedTrading' },
      { id: 'tag-update', name: 'Tag Update', api: 'triggerTagUpdate' },
      { id: 'hard-update', name: 'Hard Update', api: 'hardUpdate' },
    ],
  },
];

/**
 * Job footer component
 * 
 * Provides manual job triggering interface for debugging and maintenance.
 * 
 * @returns {JSX.Element} Job footer with collapsible job trigger panel
 */
export function JobFooter() {
  // Loading state per job (prevents double-triggering)
  const [loading, setLoading] = useState({});
  // Success/error messages per job (auto-dismiss after 5 seconds)
  const [messages, setMessages] = useState({});
  // Panel expanded/collapsed state
  const [opened, setOpened] = useState(false);

  /**
   * Triggers a background job via API
   * 
   * Handles loading state, success/error messages, and auto-dismiss.
   * Special case: "hard-update" job refreshes the page on success.
   * 
   * @param {Object} job - Job object with id, name, and api method name
   */
  const triggerJob = async (job) => {
    // Prevent double-triggering if already loading
    if (loading[job.id]) return;

    // Set loading state and clear any previous messages
    setLoading((prev) => ({ ...prev, [job.id]: true }));
    setMessages((prev) => ({ ...prev, [job.id]: null }));

    try {
      // Call the API method for this job
      const result = await api[job.api]();
      
      // Set success/error message based on result
      setMessages((prev) => ({
        ...prev,
        [job.id]: {
          type: result.status === 'success' || result.success ? 'success' : 'error',
          text: result.message || result.status || (result.success ? 'Success' : 'Error'),
        },
      }));

      // Special handling for hard-update: refresh page on success
      if (job.id === 'hard-update' && (result.status === 'success' || result.success)) {
        setTimeout(() => {
          window.location.reload();
        }, 2000); // Wait 2 seconds to show success message before reload
        return;
      }

      // Auto-dismiss success/error messages after 5 seconds
      setTimeout(() => {
        setMessages((prev) => {
          const next = { ...prev };
          delete next[job.id];
          return next;
        });
      }, 5000);
    } catch (error) {
      // Handle API errors
      setMessages((prev) => ({
        ...prev,
        [job.id]: {
          type: 'error',
          text: error.message || 'Failed to trigger job',
        },
      }));

      // Auto-dismiss error messages after 5 seconds
      setTimeout(() => {
        setMessages((prev) => {
          const next = { ...prev };
          delete next[job.id];
          return next;
        });
      }, 5000);
    } finally {
      // Always clear loading state
      setLoading((prev) => {
        const next = { ...prev };
        delete next[job.id];
        return next;
      });
    }
  };

  return (
    <Paper
      className="job-footer"
      p="md"
      mt="xl"
      style={{
        borderTop: '1px solid var(--mantine-color-dark-6)',
        backgroundColor: 'var(--mantine-color-dark-7)',
        border: '1px solid var(--mantine-color-dark-6)',
      }}
    >
      {/* Collapsible title - click to expand/collapse */}
      <Title
        className="job-footer__title"
        order={4}
        mb="md"
        onClick={() => setOpened((o) => !o)}
        style={{
          fontFamily: 'var(--mantine-font-family)',
          cursor: 'pointer',
          userSelect: 'none'
        }}
      >
        Manual Job Triggers
      </Title>
      
      {/* Collapsible content */}
      <Collapse className="job-footer__collapse" in={opened}>
        <Stack className="job-footer__categories" gap="lg">
          {jobCategories.map((category) => (
            <Stack className="job-footer__category" key={category.name} gap="xs">
              {/* Category header */}
              <Text className="job-footer__category-name" size="sm" fw={600} c="dimmed" tt="uppercase" style={{ fontFamily: 'var(--mantine-font-family)' }}>
                {category.name}
              </Text>
              
              {/* Job buttons for this category */}
              <Group className="job-footer__jobs" gap="xs" wrap="wrap">
                {category.jobs.map((job) => (
                  <Stack className="job-footer__job" key={job.id} gap="xs" style={{ minWidth: '140px' }}>
                    {/* Job trigger button */}
                    <Button
                      className="job-footer__job-btn"
                      size="xs"
                      variant="light"
                      onClick={() => triggerJob(job)}
                      loading={loading[job.id]}  // Shows loading spinner while job is running
                      fullWidth
                    >
                      {job.name}
                    </Button>
                    
                    {/* Success/error message (auto-dismisses after 5 seconds) */}
                    {messages[job.id] && (
                      <Text
                        className={`job-footer__job-message job-footer__job-message--${messages[job.id].type}`}
                        size="xs"
                        c={messages[job.id].type === 'success' ? 'green' : 'red'}
                        ta="center"
                      >
                        {messages[job.id].text}
                      </Text>
                    )}
                  </Stack>
                ))}
              </Group>
              
              {/* Divider between categories */}
              <Divider className="job-footer__divider" />
            </Stack>
          ))}
        </Stack>
      </Collapse>
    </Paper>
  );
}
