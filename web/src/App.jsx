import { AppShell, Group, Title, ActionIcon, Badge, Tooltip, Switch, Text, Modal, Button, RingProgress } from '@mantine/core';
import { notifications } from '@mantine/notifications';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { IconSettings, IconClock, IconRefresh, IconChartLine, IconPlanet, IconReceipt, IconBrain } from '@tabler/icons-react';

import UnifiedPage from './pages/UnifiedPage';
import { SchedulerModal } from './components/SchedulerModal';
import { SettingsModal } from './components/SettingsModal';
import { BacktestModal } from './components/BacktestModal';
import { TradesModal } from './components/TradesModal';
import { getSchedulerStatus, refreshAll, getSettings, updateSetting, getLedStatus, setLedEnabled, getVersion, resetAndRetrain, getResetStatus } from './api/client';
import { useState } from 'react';

function App() {
  const [schedulerOpen, setSchedulerOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [backtestOpen, setBacktestOpen] = useState(false);
  const [tradesOpen, setTradesOpen] = useState(false);
  const [resetConfirmOpen, setResetConfirmOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: schedulerStatus } = useQuery({
    queryKey: ['scheduler'],
    queryFn: getSchedulerStatus,
    refetchInterval: 10000,
  });

  const { data: settings } = useQuery({
    queryKey: ['settings'],
    queryFn: getSettings,
  });

  const { data: ledStatus } = useQuery({
    queryKey: ['ledStatus'],
    queryFn: getLedStatus,
    refetchInterval: 30000,
  });

  const { data: versionData } = useQuery({
    queryKey: ['version'],
    queryFn: getVersion,
    staleTime: Infinity,
  });

  const refreshMutation = useMutation({
    mutationFn: refreshAll,
    onSuccess: () => {
      queryClient.invalidateQueries();
    },
  });

  const tradingModeMutation = useMutation({
    mutationFn: (mode) => updateSetting('trading_mode', mode),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] });
    },
  });

  const ledMutation = useMutation({
    mutationFn: setLedEnabled,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ledStatus'] });
    },
  });

  const { data: resetStatus } = useQuery({
    queryKey: ['resetStatus'],
    queryFn: getResetStatus,
    refetchInterval: (query) => {
      // Only poll frequently when a reset is running
      return query.state.data?.running ? 1000 : 10000;
    },
  });

  const resetRetrainMutation = useMutation({
    mutationFn: resetAndRetrain,
    onSuccess: () => {
      notifications.show({
        title: 'Reset & Retrain Started',
        message: 'ML models are being retrained in the background',
        color: 'blue',
      });
      setResetConfirmOpen(false);
      queryClient.invalidateQueries({ queryKey: ['resetStatus'] });
    },
    onError: (error) => {
      notifications.show({
        title: 'Error',
        message: error.message,
        color: 'red',
      });
    },
  });

  const runningJobs = schedulerStatus?.pending?.length || 0;
  const isRefreshing = refreshMutation.isPending || runningJobs > 0;
  const isLive = settings?.trading_mode === 'live';
  const ledEnabled = ledStatus?.enabled || false;
  const ledRunning = ledStatus?.running || false;

  return (
    <>
      <AppShell header={{ height: 50 }} padding="md" className="app">
        <AppShell.Header className="app__header">
          <Group h="100%" px="md" justify="space-between" className="app__header-content">
            <Group gap={8} align="baseline" className="app__logo">
              <Title order={3} c="blue">
                Sentinel
              </Title>
              {versionData?.version && (
                <Text size="xs" c="dimmed">
                  {versionData.version}
                </Text>
              )}
            </Group>

            <Group gap="md" className="app__controls">
              <Group gap="xs" className="app__trading-mode">
                <Text size="sm" c={isLive ? 'red' : 'dimmed'} className={`app__trading-mode-label ${isLive ? 'app__trading-mode-label--live' : 'app__trading-mode-label--research'}`}>
                  {isLive ? 'LIVE' : 'Research'}
                </Text>
                <Switch
                  checked={isLive}
                  onChange={(e) =>
                    tradingModeMutation.mutate(e.currentTarget.checked ? 'live' : 'research')
                  }
                  color="red"
                  size="sm"
                  disabled={tradingModeMutation.isPending}
                  className="app__trading-mode-switch"
                />
              </Group>

              <Group gap="xs" className="app__actions">
                <Tooltip label={ledEnabled ? (ledRunning ? 'LED Display Active' : 'LED Display Enabled') : 'LED Display Off'}>
                  <ActionIcon
                    variant={ledEnabled ? 'light' : 'subtle'}
                    color={ledEnabled ? (ledRunning ? 'teal' : 'blue') : 'gray'}
                    size="lg"
                    onClick={() => ledMutation.mutate(!ledEnabled)}
                    loading={ledMutation.isPending}
                    className="app__action-btn app__action-btn--led"
                  >
                    <IconPlanet size={20} />
                  </ActionIcon>
                </Tooltip>

                {!isLive && (
                  <Tooltip label="Backtest Portfolio">
                    <ActionIcon
                      variant="subtle"
                      size="lg"
                      onClick={() => setBacktestOpen(true)}
                      className="app__action-btn app__action-btn--backtest"
                    >
                      <IconChartLine size={20} />
                    </ActionIcon>
                  </Tooltip>
                )}

                <Tooltip label="Trade History">
                  <ActionIcon
                    variant="subtle"
                    size="lg"
                    onClick={() => setTradesOpen(true)}
                    className="app__action-btn app__action-btn--trades"
                  >
                    <IconReceipt size={20} />
                  </ActionIcon>
                </Tooltip>

                <Tooltip label="Refresh All (sync rates, portfolio, prices, scores)">
                  <ActionIcon
                    variant="subtle"
                    size="lg"
                    onClick={() => refreshMutation.mutate()}
                    loading={refreshMutation.isPending}
                    disabled={isRefreshing}
                    className="app__action-btn app__action-btn--refresh"
                  >
                    <IconRefresh size={20} />
                  </ActionIcon>
                </Tooltip>

                {resetStatus?.running ? (
                  <Tooltip
                    label={
                      resetStatus.models_total
                        ? `Training ${resetStatus.current_symbol} (${resetStatus.models_current}/${resetStatus.models_total})`
                        : `${resetStatus.step_name} (${resetStatus.current_step}/${resetStatus.total_steps})`
                    }
                  >
                    <Group gap={6}>
                      <RingProgress
                        size={32}
                        thickness={3}
                        sections={[
                          {
                            value: resetStatus.models_total
                              ? (resetStatus.models_current / resetStatus.models_total) * 100
                              : (resetStatus.current_step / resetStatus.total_steps) * 100,
                            color: 'orange',
                          },
                        ]}
                        label={
                          <Text size="xs" ta="center" c="orange" fw={500}>
                            {resetStatus.models_total ? resetStatus.models_current : resetStatus.current_step}
                          </Text>
                        }
                      />
                      <Text size="xs" c="dimmed" style={{ maxWidth: 140 }} truncate>
                        {resetStatus.models_total
                          ? `${resetStatus.current_symbol} (${resetStatus.models_current}/${resetStatus.models_total})`
                          : resetStatus.step_name}
                      </Text>
                    </Group>
                  </Tooltip>
                ) : (
                  <Tooltip label="Reset & Retrain All ML Models">
                    <ActionIcon
                      variant="subtle"
                      size="lg"
                      color="orange"
                      onClick={() => setResetConfirmOpen(true)}
                      loading={resetRetrainMutation.isPending}
                      className="app__action-btn app__action-btn--reset-retrain"
                    >
                      <IconBrain size={20} />
                    </ActionIcon>
                  </Tooltip>
                )}

                <Tooltip label="Scheduler">
                  <ActionIcon
                    variant="subtle"
                    size="lg"
                    onClick={() => setSchedulerOpen(true)}
                    pos="relative"
                    className="app__action-btn app__action-btn--scheduler"
                  >
                    <IconClock size={20} />
                    {runningJobs > 0 && (
                      <Badge
                        size="sm"
                        color="blue"
                        circle
                        pos="absolute"
                        top={-4}
                        right={-4}
                        className="app__running-jobs-badge"
                      >
                        {runningJobs}
                      </Badge>
                    )}
                  </ActionIcon>
                </Tooltip>

                <Tooltip label="Settings">
                  <ActionIcon
                    variant="subtle"
                    size="lg"
                    onClick={() => setSettingsOpen(true)}
                    className="app__action-btn app__action-btn--settings"
                  >
                    <IconSettings size={20} />
                  </ActionIcon>
                </Tooltip>
              </Group>
            </Group>
          </Group>
        </AppShell.Header>

        <AppShell.Main className="app__main">
          <UnifiedPage />
        </AppShell.Main>
      </AppShell>

      <SchedulerModal opened={schedulerOpen} onClose={() => setSchedulerOpen(false)} />
      <SettingsModal opened={settingsOpen} onClose={() => setSettingsOpen(false)} />
      <BacktestModal opened={backtestOpen} onClose={() => setBacktestOpen(false)} />
      <TradesModal opened={tradesOpen} onClose={() => setTradesOpen(false)} />

      <Modal
        opened={resetConfirmOpen}
        onClose={() => setResetConfirmOpen(false)}
        title="Reset & Retrain All ML Models"
        centered
      >
        <Text size="sm" mb="md">
          This will delete all ML training data and model files, then retrain
          all models from scratch. This operation may take several minutes.
        </Text>
        <Text size="sm" c="dimmed" mb="lg">
          Use this after changing security geography or industry classifications.
        </Text>
        <Group justify="flex-end">
          <Button variant="default" onClick={() => setResetConfirmOpen(false)}>
            Cancel
          </Button>
          <Button
            color="orange"
            onClick={() => resetRetrainMutation.mutate()}
            loading={resetRetrainMutation.isPending}
          >
            Reset & Retrain
          </Button>
        </Group>
      </Modal>
    </>
  );
}

export default App;
