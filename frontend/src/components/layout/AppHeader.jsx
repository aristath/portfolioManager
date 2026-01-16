/**
 * Application Header Component
 *
 * Displays the main application header with:
 * - Application branding (title and subtitle)
 * - Tradernet connection status indicator
 * - Trading mode toggle (Research/Live)
 * - Settings button
 *
 * Provides quick access to key system status and controls.
 */
import { Button, Group } from '@mantine/core';
import { IconSettings } from '@tabler/icons-react';
import { useAppStore } from '../../stores/appStore';
import { useSettingsStore } from '../../stores/settingsStore';

/**
 * Application header component
 *
 * Shows branding, connection status, trading mode, and settings access.
 *
 * @returns {JSX.Element} Header component with branding and controls
 */
export function AppHeader() {
  const { tradernet, openSettingsModal } = useAppStore();
  const { tradingMode, toggleTradingMode } = useSettingsStore();

  return (
    <header className="app-header" style={{
      padding: '12px 0',
      borderBottom: '1px solid var(--mantine-color-dark-6)',
      backgroundColor: 'var(--mantine-color-dark-9)',
    }}>
      <Group className="app-header__content" justify="space-between" align="center">
        {/* Application branding */}
        <div className="app-header__branding">
          <h1 className="app-header__title" style={{
            margin: 0,
            fontSize: '1.25rem',
            fontWeight: 'bold',
            color: 'var(--mantine-color-blue-0)',
            fontFamily: 'var(--mantine-font-family)',
            letterSpacing: '0.5px',
          }}>
            Sentinel
          </h1>
          <p className="app-header__subtitle" style={{
            margin: 0,
            fontSize: '0.875rem',
            color: 'var(--mantine-color-dark-2)',
            fontFamily: 'var(--mantine-font-family)',
          }}>
            Automated Portfolio Management
          </p>
        </div>

        {/* Header actions: connection status, trading mode, settings */}
        <Group className="app-header__actions" gap="md">
          {/* Tradernet connection status indicator */}
          <Group className="app-header__connection" gap="xs">
            {/* Status dot: green when connected, red when offline */}
            <div
              className={`app-header__indicator app-header__indicator--${tradernet.connected ? 'online' : 'offline'}`}
              style={{
                width: '8px',
                height: '8px',
                borderRadius: '50%',
                backgroundColor: tradernet.connected ? 'var(--mantine-color-green-5)' : 'var(--mantine-color-red-5)',
              }}
            />
            <span className="app-header__connection-text" style={{ fontSize: '0.875rem', color: tradernet.connected ? 'var(--mantine-color-green-4)' : 'var(--mantine-color-red-4)' }}>
              {tradernet.connected ? 'Tradernet' : 'Offline'}
            </span>
          </Group>

          {/* Trading mode toggle button */}
          {/* Research mode: simulated trades (yellow) | Live mode: real trades (green) */}
          <Button
            className="app-header__mode-toggle"
            variant="light"
            size="xs"
            onClick={toggleTradingMode}
            color={tradingMode === 'research' ? 'yellow' : 'green'}
            title={tradingMode === 'research' ? 'Research Mode: Trades are simulated' : 'Live Mode: Trades are executed'}
          >
            <Group className="app-header__mode-content" gap="xs">
              {/* Mode indicator dot */}
              <div
                className={`app-header__mode-indicator app-header__mode-indicator--${tradingMode}`}
                style={{
                  width: '8px',
                  height: '8px',
                  borderRadius: '50%',
                  backgroundColor: tradingMode === 'research' ? 'var(--mantine-color-yellow-5)' : 'var(--mantine-color-green-5)',
                }}
              />
              <span className="app-header__mode-text" style={{ fontSize: '0.875rem', fontWeight: 500 }}>
                {tradingMode === 'research' ? 'Research' : 'Live'}
              </span>
            </Group>
          </Button>

          {/* Settings button */}
          <Button
            className="app-header__settings-btn"
            variant="subtle"
            size="xs"
            onClick={openSettingsModal}
            title="Settings"
          >
            <IconSettings size={20} />
          </Button>
        </Group>
      </Group>
    </header>
  );
}
