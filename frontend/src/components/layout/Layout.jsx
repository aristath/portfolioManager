import { Container } from '@mantine/core';
import { Outlet } from 'react-router-dom';
import { AppHeader } from './AppHeader';
import { StatusBar } from './StatusBar';
import { TabNavigation } from './TabNavigation';
import { MarketStatus } from './MarketStatus';
import { JobFooter } from './JobFooter';
import { AddSecurityModal } from '../modals/AddSecurityModal';
import { EditSecurityModal } from '../modals/EditSecurityModal';
import { SecurityChartModal } from '../modals/SecurityChartModal';
import { SettingsModal } from '../modals/SettingsModal';
import { PlannerManagementModal } from '../modals/PlannerManagementModal';
import { useEffect } from 'react';
import { useAppStore } from '../../stores/appStore';
import { usePortfolioStore } from '../../stores/portfolioStore';
import { useSecuritiesStore } from '../../stores/securitiesStore';
import { useSettingsStore } from '../../stores/settingsStore';
import { useTradesStore } from '../../stores/tradesStore';
import { useLogsStore } from '../../stores/logsStore';
import { useNotifications } from '../../hooks/useNotifications';
import { ColorSchemeToggle } from './ColorSchemeToggle';

export function Layout() {
  // Display notifications from app store
  useNotifications();
  const { fetchAll, startEventStream, stopEventStream } = useAppStore();
  const { fetchAllocation, fetchCashBreakdown, fetchTargets } = usePortfolioStore();
  const { fetchSecurities, fetchSparklines } = useSecuritiesStore();
  const { fetchSettings } = useSettingsStore();
  const { fetchTrades } = useTradesStore();
  const { fetchAvailableLogFiles, selectedLogFile } = useLogsStore();

  useEffect(() => {
    // Fetch all initial data
    const loadData = async () => {
      await Promise.all([
        fetchAll(),
        fetchAllocation(),
        fetchCashBreakdown(),
        fetchSecurities(),
        fetchTargets(),
        fetchSparklines(),
        fetchSettings(),
        fetchTrades(),
        fetchAvailableLogFiles(),
      ]);
    };

    loadData();

    // Start unified event stream with log_file param if logs view is active
    startEventStream(selectedLogFile);

    // Cleanup on unmount
    return () => {
      stopEventStream();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Restart event stream when log file changes
  useEffect(() => {
    if (selectedLogFile) {
      stopEventStream();
      startEventStream(selectedLogFile);
    }
    return () => {
      stopEventStream();
    };
  }, [selectedLogFile, startEventStream, stopEventStream]);

  return (
    <div style={{ minHeight: '100vh', backgroundColor: 'var(--mantine-color-dark-9)' }}>
      <Container size="xl" py="md">
        <AppHeader />
        <MarketStatus />
        <StatusBar />
        <TabNavigation />
        <div style={{ marginTop: '16px' }}>
          <Outlet />
        </div>
        <JobFooter />
      </Container>

      {/* Modals */}
      <AddSecurityModal />
      <EditSecurityModal />
      <SecurityChartModal />
      <SettingsModal />
      <PlannerManagementModal />

      {/* Color Scheme Toggle */}
      <ColorSchemeToggle />
    </div>
  );
}
