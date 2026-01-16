/**
 * Main Layout Component
 *
 * Provides the main application layout structure for all views.
 * Handles:
 * - Initial data loading from all stores
 * - Event stream lifecycle management (SSE connection)
 * - Application version display
 * - Global modals (Add Security, Edit Security, Charts, Settings, Planner)
 * - Layout structure (Header, Status Bar, Navigation, Content Area, Footer)
 *
 * This component wraps all main application views and provides shared UI elements.
 */
import { Container, Text } from '@mantine/core';
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
import { useEffect, useState } from 'react';
import { useAppStore } from '../../stores/appStore';
import { usePortfolioStore } from '../../stores/portfolioStore';
import { useSecuritiesStore } from '../../stores/securitiesStore';
import { useSettingsStore } from '../../stores/settingsStore';
import { useTradesStore } from '../../stores/tradesStore';
import { useNotifications } from '../../hooks/useNotifications';

/**
 * Main layout component
 *
 * Sets up the application structure and initializes all data on mount.
 * Manages event stream connection for real-time updates.
 *
 * @returns {JSX.Element} Layout component with header, navigation, content area, and modals
 */
export function Layout() {
  // Display notifications from app store messages
  useNotifications();

  // Get store actions for data fetching and event stream management
  const { fetchAll, startEventStream, stopEventStream } = useAppStore();
  const { fetchAllocation, fetchCashBreakdown, fetchTargets } = usePortfolioStore();
  const { fetchSecurities } = useSecuritiesStore();
  const { fetchSettings } = useSettingsStore();
  const { fetchAll: fetchTradesAndPending } = useTradesStore();

  // Application version state
  const [version, setVersion] = useState('loading...');

  // Fetch application version on mount
  useEffect(() => {
    fetch('/api/version')
      .then(r => r.json())
      .then(data => setVersion(data.version))
      .catch(() => setVersion('unknown'));  // Fallback if version fetch fails
  }, []);

  // Load initial data from all stores on mount
  // Fetches data in parallel for efficiency
  useEffect(() => {
    const loadData = async () => {
      try {
        await Promise.all([
          fetchAll(),                    // App store: system status, recommendations, etc.
          fetchAllocation(),            // Portfolio: current allocation
          fetchCashBreakdown(),         // Portfolio: cash breakdown
          fetchSecurities(),            // Securities: investment universe
          fetchTargets(),               // Portfolio: allocation targets
          fetchSettings(),              // Settings: application settings
          fetchTradesAndPending(),      // Trades: executed trades and pending orders
        ]);
      } catch (error) {
        console.error('Failed to load initial data:', error);
        // Individual store methods already handle their own errors
        // This catch prevents unhandled promise rejection warnings
      }
    };

    loadData();
  }, [fetchAll, fetchAllocation, fetchCashBreakdown, fetchSecurities, fetchTargets, fetchSettings, fetchTradesAndPending]);

  // Manage event stream lifecycle (SSE connection)
  // Starts stream on mount, stops on unmount
  useEffect(() => {
    startEventStream();

    // Cleanup: stop event stream when component unmounts
    return () => {
      stopEventStream();
    };
  }, [startEventStream, stopEventStream]);

  return (
    <div className="layout" style={{ minHeight: '100vh', backgroundColor: 'var(--mantine-color-dark-9)' }}>
      <Container className="layout__container" size="xl" py="md">
        {/* Application header with title and navigation */}
        <AppHeader />

        {/* Market status indicator (open/closed, holidays) */}
        <MarketStatus />

        {/* Status bar with system status and connection indicators */}
        <StatusBar />

        {/* Tab navigation for main views */}
        <TabNavigation />

        {/* Main content area - renders child routes via Outlet */}
        <main className="layout__main" style={{ marginTop: '16px' }}>
          <Outlet />
        </main>

        {/* Job footer showing running background jobs */}
        <JobFooter />

        {/* Application version display */}
        <Text
          className="layout__version"
          size="xs"
          c="dimmed"
          ta="center"
          mt="md"
          pb="md"
          style={{ fontFamily: 'var(--mantine-font-family-monospace)' }}
        >
          Sentinel {version}
        </Text>
      </Container>

      {/* Global modals - available throughout the application */}
      <AddSecurityModal />
      <EditSecurityModal />
      <SecurityChartModal />
      <SettingsModal />
      <PlannerManagementModal />
    </div>
  );
}
