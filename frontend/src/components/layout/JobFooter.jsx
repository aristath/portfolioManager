import { Paper, Group, Button, Text, Stack, Divider, Title } from '@mantine/core';
import { api } from '../../api/client';
import { useState } from 'react';

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

export function JobFooter() {
  const [loading, setLoading] = useState({});
  const [messages, setMessages] = useState({});

  const triggerJob = async (job) => {
    if (loading[job.id]) return;

    setLoading((prev) => ({ ...prev, [job.id]: true }));
    setMessages((prev) => ({ ...prev, [job.id]: null }));

    try {
      const result = await api[job.api]();
      setMessages((prev) => ({
        ...prev,
        [job.id]: {
          type: result.status === 'success' || result.success ? 'success' : 'error',
          text: result.message || result.status || (result.success ? 'Success' : 'Error'),
        },
      }));

      // For hard update, refresh page on success
      if (job.id === 'hard-update' && (result.status === 'success' || result.success)) {
        setTimeout(() => {
          window.location.reload();
        }, 2000); // Wait 2 seconds to show success message
        return;
      }

      setTimeout(() => {
        setMessages((prev) => {
          const next = { ...prev };
          delete next[job.id];
          return next;
        });
      }, 5000);
    } catch (error) {
      setMessages((prev) => ({
        ...prev,
        [job.id]: {
          type: 'error',
          text: error.message || 'Failed to trigger job',
        },
      }));

      setTimeout(() => {
        setMessages((prev) => {
          const next = { ...prev };
          delete next[job.id];
          return next;
        });
      }, 5000);
    } finally {
      setLoading((prev) => {
        const next = { ...prev };
        delete next[job.id];
        return next;
      });
    }
  };

  return (
    <Paper
      p="md"
      mt="xl"
      style={{
        borderTop: '1px solid var(--mantine-color-dark-6)',
        backgroundColor: 'var(--mantine-color-dark-7)',
        border: '1px solid var(--mantine-color-dark-6)',
      }}
    >
      <Title order={4} mb="md" style={{ fontFamily: 'var(--mantine-font-family)' }}>Manual Job Triggers</Title>
      <Stack gap="lg">
        {jobCategories.map((category) => (
          <Stack key={category.name} gap="xs">
            <Text size="sm" fw={600} c="dimmed" tt="uppercase" style={{ fontFamily: 'var(--mantine-font-family)' }}>
              {category.name}
            </Text>
            <Group gap="xs" wrap="wrap">
              {category.jobs.map((job) => (
                <Stack key={job.id} gap="xs" style={{ minWidth: '140px' }}>
                  <Button
                    size="xs"
                    variant="light"
                    onClick={() => triggerJob(job)}
                    loading={loading[job.id]}
                    fullWidth
                  >
                    {job.name}
                  </Button>
                  {messages[job.id] && (
                    <Text
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
            <Divider />
          </Stack>
        ))}
      </Stack>
    </Paper>
  );
}
