/**
 * Settings Modal Component
 *
 * Comprehensive settings management modal with multiple tabs for different configuration areas.
 *
 * Features:
 * - Trading Settings: Frequency limits, limit order protection, scoring parameters, market regime detection
 * - Display Settings: LED matrix display mode, ticker configuration, brightness
 * - System Settings: Job scheduling, cache management, historical data sync, system restart, hardware management
 * - Backup Settings: Cloudflare R2 backup configuration and management
 * - Credentials: API keys for Tradernet, GitHub, and other services
 *
 * All settings are stored in the settings database and take precedence over environment variables.
 */
import { useState, useEffect } from 'react';
import { Modal, Tabs, Text, Button, NumberInput, Switch, Group, Stack, Paper, Divider, Alert, TextInput, PasswordInput, Select } from '@mantine/core';
import { useAppStore } from '../../stores/appStore';
import { useSettingsStore } from '../../stores/settingsStore';
import { api } from '../../api/client';
import { useNotifications } from '../../hooks/useNotifications';
import { R2BackupModal } from './R2BackupModal';

/**
 * Settings modal component
 *
 * Provides a comprehensive interface for managing all application settings.
 *
 * @returns {JSX.Element} Settings modal with tabbed interface
 */
export function SettingsModal() {
  const { showSettingsModal, closeSettingsModal } = useAppStore();
  const { settings, fetchSettings, updateSetting } = useSettingsStore();
  const { showNotification } = useNotifications();
  const [activeTab, setActiveTab] = useState('trading');
  const [loading, setLoading] = useState(false);
  const [syncingHistorical, setSyncingHistorical] = useState(false);
  const [testingR2Connection, setTestingR2Connection] = useState(false);
  const [backingUpToR2, setBackingUpToR2] = useState(false);
  const [showR2BackupModal, setShowR2BackupModal] = useState(false);
  const [uploadingSketch, setUploadingSketch] = useState(false);

  // Fetch settings when modal opens
  useEffect(() => {
    if (showSettingsModal) {
      fetchSettings();
    }
  }, [showSettingsModal, fetchSettings]);

  /**
   * Handles updating a setting value
   *
   * @param {string} key - Setting key
   * @param {*} value - Setting value
   */
  const handleUpdateSetting = async (key, value) => {
    try {
      await updateSetting(key, value);
      showNotification('Setting updated', 'success');
    } catch (error) {
      showNotification(`Failed to update setting: ${error.message}`, 'error');
    }
  };

  /**
   * Handles syncing historical data
   */
  const handleSyncHistorical = async () => {
    setSyncingHistorical(true);
    try {
      await api.syncHistorical();
      showNotification('Historical data sync started', 'success');
    } catch (error) {
      showNotification(`Failed to sync historical data: ${error.message}`, 'error');
    } finally {
      setSyncingHistorical(false);
    }
  };

  /**
   * Handles resetting application caches
   */
  const handleResetCache = async () => {
    setLoading(true);
    try {
      await api.resetCache();
      showNotification('Cache reset successfully', 'success');
    } catch (error) {
      showNotification(`Failed to reset cache: ${error.message}`, 'error');
    } finally {
      setLoading(false);
    }
  };

  /**
   * Handles system restart
   */
  const handleRestartSystem = async () => {
    if (!confirm('Are you sure you want to restart the system?')) return;
    setLoading(true);
    try {
      await api.restartSystem();
      showNotification('System restart initiated', 'success');
    } catch (error) {
      showNotification(`Failed to restart system: ${error.message}`, 'error');
    } finally {
      setLoading(false);
    }
  };

  /**
   * Gets a setting value with default fallback
   *
   * @param {string} key - Setting key
   * @param {*} defaultValue - Default value if setting not found
   * @returns {*} Setting value or default
   */
  const getSetting = (key, defaultValue = 0) => {
    return settings[key] ?? defaultValue;
  };

  /**
   * Tests R2 connection with current credentials
   */
  const handleTestR2Connection = async () => {
    setTestingR2Connection(true);
    try {
      const result = await api.testR2Connection();
      if (result.status === 'success') {
        showNotification('R2 connection successful', 'success');
      } else {
        showNotification(`R2 connection failed: ${result.message}`, 'error');
      }
    } catch (error) {
      showNotification(`Failed to test R2 connection: ${error.message}`, 'error');
    } finally {
      setTestingR2Connection(false);
    }
  };

  /**
   * Creates a new R2 backup
   */
  const handleBackupToR2 = async () => {
    setBackingUpToR2(true);
    try {
      await api.createR2Backup();
      showNotification('Backup job started successfully', 'success');
    } catch (error) {
      showNotification(`Failed to create backup: ${error.message}`, 'error');
    } finally {
      setBackingUpToR2(false);
    }
  };

  /**
   * Opens the R2 backup management modal
   */
  const handleViewR2Backups = () => {
    setShowR2BackupModal(true);
  };

  /**
   * Handles uploading Arduino sketch to MCU
   */
  const handleUploadSketch = async () => {
    setUploadingSketch(true);
    try {
      const result = await api.uploadSketch();
      if (result.status === 'success') {
        showNotification('Sketch uploaded successfully', 'success');
      } else {
        showNotification(`Sketch upload failed: ${result.message}`, 'error');
      }
    } catch (error) {
      showNotification(`Failed to upload sketch: ${error.message}`, 'error');
    } finally {
      setUploadingSketch(false);
    }
  };

  return (
    <>
      <Modal
        className="settings-modal"
        opened={showSettingsModal}
        onClose={closeSettingsModal}
        title="Settings"
        size="xl"
        styles={{ body: { padding: 0 } }}
      >
      <Tabs className="settings-modal__tabs" value={activeTab} onChange={setActiveTab}>
        <Tabs.List className="settings-modal__tab-list" grow>
          <Tabs.Tab className="settings-modal__tab settings-modal__tab--trading" value="trading">Trading</Tabs.Tab>
          <Tabs.Tab className="settings-modal__tab settings-modal__tab--display" value="display">Display</Tabs.Tab>
          <Tabs.Tab className="settings-modal__tab settings-modal__tab--system" value="system">System</Tabs.Tab>
          <Tabs.Tab className="settings-modal__tab settings-modal__tab--backup" value="backup">Backup</Tabs.Tab>
          <Tabs.Tab className="settings-modal__tab settings-modal__tab--credentials" value="credentials">Credentials</Tabs.Tab>
        </Tabs.List>

        {/* Trading Settings Tab */}
        <Tabs.Panel className="settings-modal__panel settings-modal__panel--trading" value="trading" p="md">
          <Stack className="settings-modal__trading-content" gap="md">
            {/* Trade Frequency Limits Section */}
            <Paper className="settings-modal__section settings-modal__section--frequency" p="md" withBorder>
              <Text className="settings-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Trade Frequency Limits</Text>
              <Text className="settings-modal__section-desc" size="xs" c="dimmed" mb="md">
                Prevent excessive trading by enforcing minimum time between trades and daily/weekly limits.
              </Text>
              <Stack className="settings-modal__section-content" gap="sm">
                <Switch
                  className="settings-modal__switch settings-modal__switch--frequency-enabled"
                  label="Enable frequency limits"
                  checked={getSetting('trade_frequency_limits_enabled', 1) === 1}
                  onChange={(e) => handleUpdateSetting('trade_frequency_limits_enabled', e.currentTarget.checked ? 1 : 0)}
                />
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Min Time Between Trades</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Minimum minutes between any trades</Text>
                  </div>
                  <Group className="settings-modal__setting-input" gap="xs">
                    <NumberInput
                      className="settings-modal__number-input"
                      value={getSetting('min_time_between_trades_minutes', 60)}
                      onChange={(val) => handleUpdateSetting('min_time_between_trades_minutes', val)}
                      min={0}
                      step={5}
                      w={80}
                      size="sm"
                    />
                    <Text className="settings-modal__setting-unit" size="sm" c="dimmed">min</Text>
                  </Group>
                </Group>
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Max Trades Per Day</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Maximum trades per calendar day</Text>
                  </div>
                  <NumberInput
                    className="settings-modal__number-input"
                    value={getSetting('max_trades_per_day', 4)}
                    onChange={(val) => handleUpdateSetting('max_trades_per_day', val)}
                    min={1}
                    step={1}
                    w={80}
                    size="sm"
                  />
                </Group>
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Max Trades Per Week</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Maximum trades per rolling 7-day window</Text>
                  </div>
                  <NumberInput
                    className="settings-modal__number-input"
                    value={getSetting('max_trades_per_week', 10)}
                    onChange={(val) => handleUpdateSetting('max_trades_per_week', val)}
                    min={1}
                    step={1}
                    w={80}
                    size="sm"
                  />
                </Group>
              </Stack>
            </Paper>

            {/* Trade Safety - Cooloff Periods */}
            <Paper className="settings-modal__section settings-modal__section--trade-safety" p="md" withBorder>
              <Text className="settings-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Trade Safety</Text>
              <Text className="settings-modal__section-desc" size="xs" c="dimmed" mb="md">
                Cooloff periods prevent trading the same security too frequently.
              </Text>
              <Stack className="settings-modal__section-content" gap="sm">
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Buy Cooldown</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Days before re-buying same security</Text>
                  </div>
                  <Group className="settings-modal__setting-input" gap="xs">
                    <NumberInput
                      className="settings-modal__number-input"
                      value={getSetting('buy_cooldown_days', 30)}
                      onChange={(val) => handleUpdateSetting('buy_cooldown_days', val)}
                      min={0}
                      max={365}
                      step={1}
                      w={80}
                      size="sm"
                    />
                    <Text className="settings-modal__setting-unit" size="sm" c="dimmed">days</Text>
                  </Group>
                </Group>
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Sell Cooldown</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Days after buying before selling allowed</Text>
                  </div>
                  <Group className="settings-modal__setting-input" gap="xs">
                    <NumberInput
                      className="settings-modal__number-input"
                      value={getSetting('sell_cooldown_days', 180)}
                      onChange={(val) => handleUpdateSetting('sell_cooldown_days', val)}
                      min={0}
                      max={730}
                      step={1}
                      w={80}
                      size="sm"
                    />
                    <Text className="settings-modal__setting-unit" size="sm" c="dimmed">days</Text>
                  </Group>
                </Group>
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Minimum Hold</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Minimum days to hold before selling</Text>
                  </div>
                  <Group className="settings-modal__setting-input" gap="xs">
                    <NumberInput
                      className="settings-modal__number-input"
                      value={getSetting('min_hold_days', 90)}
                      onChange={(val) => handleUpdateSetting('min_hold_days', val)}
                      min={0}
                      max={365}
                      step={1}
                      w={80}
                      size="sm"
                    />
                    <Text className="settings-modal__setting-unit" size="sm" c="dimmed">days</Text>
                  </Group>
                </Group>
                <Alert className="settings-modal__alert" color="blue" variant="light" styles={{message: {fontSize: '12px'}}}>
                  <Text className="settings-modal__alert-text" size="xs">
                    Set any value to 0 to disable that particular cooloff check.
                  </Text>
                </Alert>
              </Stack>
            </Paper>

            {/* Limit Order Protection Section */}
            <Paper className="settings-modal__section settings-modal__section--limit-order" p="md" withBorder>
              <Text className="settings-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Limit Order Protection</Text>
              <Stack className="settings-modal__section-content" gap="sm">
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Limit Order Buffer</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Price protection buffer for limit orders</Text>
                  </div>
                  <Group className="settings-modal__setting-input" gap="xs">
                    <NumberInput
                      className="settings-modal__number-input"
                      value={(getSetting('limit_order_buffer_percent', 0.05) * 100).toFixed(1)}
                      onChange={(val) => handleUpdateSetting('limit_order_buffer_percent', (val || 0) / 100)}
                      min={1}
                      max={15}
                      step={0.5}
                      precision={1}
                      w={80}
                      size="sm"
                    />
                    <Text className="settings-modal__setting-unit" size="sm" c="dimmed">%</Text>
                  </Group>
                </Group>
                {/* Explanation of limit order buffer */}
                <Alert className="settings-modal__alert" color="blue" variant="light" styles={{message: {fontSize: '12px'}}}>
                  <Text className="settings-modal__alert-text" size="xs">
                    <strong>Example:</strong> If current price is €30 and buffer is 5%, buy limit is €31.50.
                    Protects against order book price discrepancies.
                  </Text>
                </Alert>
              </Stack>
            </Paper>

            {/* Scoring Parameters Section */}
            <Paper className="settings-modal__section settings-modal__section--scoring" p="md" withBorder>
              <Text className="settings-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Scoring Parameters</Text>
              <Stack className="settings-modal__section-content" gap="sm">
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Target Annual Return</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Target CAGR for scoring</Text>
                  </div>
                  <Group className="settings-modal__setting-input" gap="xs">
                    <NumberInput
                      className="settings-modal__number-input"
                      value={(getSetting('target_annual_return', 0.11) * 100).toFixed(0)}
                      onChange={(val) => handleUpdateSetting('target_annual_return', (val || 0) / 100)}
                      min={0}
                      step={1}
                      w={80}
                      size="sm"
                    />
                    <Text className="settings-modal__setting-unit" size="sm" c="dimmed">%</Text>
                  </Group>
                </Group>
              </Stack>
            </Paper>

            {/* Market Regime Detection Section */}
            <Paper className="settings-modal__section settings-modal__section--regime" p="md" withBorder>
              <Text className="settings-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Market Regime Detection</Text>
              <Text className="settings-modal__section-desc" size="xs" c="dimmed" mb="md">
                Cash reserves adjust automatically based on market conditions (SPY/QQQ 200-day MA).
              </Text>
              <Stack className="settings-modal__section-content" gap="sm">
                <Switch
                  className="settings-modal__switch settings-modal__switch--regime-enabled"
                  label="Enable regime-based cash reserves"
                  checked={getSetting('market_regime_detection_enabled', 1) === 1}
                  onChange={(e) => handleUpdateSetting('market_regime_detection_enabled', e.currentTarget.checked ? 1 : 0)}
                />
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Bull Market Reserve</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Cash reserve percentage</Text>
                  </div>
                  <Group className="settings-modal__setting-input" gap="xs">
                    <NumberInput
                      className="settings-modal__number-input"
                      value={(getSetting('market_regime_bull_cash_reserve', 0.02) * 100).toFixed(1)}
                      onChange={(val) => {
                        const v = Math.max(0.01, Math.min(0.40, (val || 0) / 100));
                        handleUpdateSetting('market_regime_bull_cash_reserve', v);
                      }}
                      min={1}
                      max={40}
                      step={0.5}
                      w={80}
                      size="sm"
                    />
                    <Text className="settings-modal__setting-unit" size="sm" c="dimmed">%</Text>
                  </Group>
                </Group>
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Bear Market Reserve</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Cash reserve percentage</Text>
                  </div>
                  <Group className="settings-modal__setting-input" gap="xs">
                    <NumberInput
                      className="settings-modal__number-input"
                      value={(getSetting('market_regime_bear_cash_reserve', 0.05) * 100).toFixed(1)}
                      onChange={(val) => {
                        const v = Math.max(0.01, Math.min(0.40, (val || 0) / 100));
                        handleUpdateSetting('market_regime_bear_cash_reserve', v);
                      }}
                      min={1}
                      max={40}
                      step={0.5}
                      w={80}
                      size="sm"
                    />
                    <Text className="settings-modal__setting-unit" size="sm" c="dimmed">%</Text>
                  </Group>
                </Group>
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Sideways Market Reserve</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Cash reserve percentage</Text>
                  </div>
                  <Group className="settings-modal__setting-input" gap="xs">
                    <NumberInput
                      className="settings-modal__number-input"
                      value={(getSetting('market_regime_sideways_cash_reserve', 0.03) * 100).toFixed(1)}
                      onChange={(val) => {
                        const v = Math.max(0.01, Math.min(0.40, (val || 0) / 100));
                        handleUpdateSetting('market_regime_sideways_cash_reserve', v);
                      }}
                      min={1}
                      max={40}
                      step={0.5}
                      w={80}
                      size="sm"
                    />
                    <Text className="settings-modal__setting-unit" size="sm" c="dimmed">%</Text>
                  </Group>
                </Group>
                <Text className="settings-modal__note" size="xs" c="dimmed" mt="xs">
                  Reserves are calculated as percentage of total portfolio value, with a minimum floor set in Planner Configuration.
                </Text>
              </Stack>
            </Paper>
          </Stack>
        </Tabs.Panel>

        {/* Display Settings Tab */}
        <Tabs.Panel className="settings-modal__panel settings-modal__panel--display" value="display" p="md">
          <Stack className="settings-modal__display-content" gap="md">
            {/* Display Mode Section */}
            <Paper className="settings-modal__section settings-modal__section--display-mode" p="md" withBorder>
              <Text className="settings-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Display Mode</Text>
              <Text className="settings-modal__section-desc" size="xs" c="dimmed" mb="md">
                Choose what to show on the LED matrix display.
              </Text>
              <Select
                className="settings-modal__select"
                label="Display Mode"
                value={getSetting('display_mode', 'TEXT') || 'TEXT'}
                onChange={(val) => handleUpdateSetting('display_mode', val)}
                data={[
                  { value: 'TEXT', label: 'Ticker (Portfolio value, cash, actions)' },
                  { value: 'HEALTH', label: 'Health (Animated portfolio visualization)' },
                  { value: 'STATS', label: 'Stats (Pixel count visualization)' }
                ]}
                description="Select which information to display on the LED matrix"
              />
            </Paper>

            {/* LED Matrix Section */}
            <Paper className="settings-modal__section settings-modal__section--led" p="md" withBorder>
              <Text className="settings-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">LED Matrix</Text>
              <Stack className="settings-modal__section-content" gap="sm">
                {/* Ticker speed setting */}
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Ticker Speed</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Lower = faster scroll</Text>
                  </div>
                  <Group className="settings-modal__setting-input" gap="xs">
                    <NumberInput
                      className="settings-modal__number-input"
                      value={getSetting('ticker_speed', 50)}
                      onChange={(val) => handleUpdateSetting('ticker_speed', val)}
                      min={1}
                      max={100}
                      step={1}
                      w={80}
                      size="sm"
                    />
                    <Text className="settings-modal__setting-unit" size="sm" c="dimmed">ms</Text>
                  </Group>
                </Group>
                {/* Brightness setting */}
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Brightness</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">0-255 (default 150)</Text>
                  </div>
                  <NumberInput
                    className="settings-modal__number-input"
                    value={getSetting('led_brightness', 150)}
                    onChange={(val) => handleUpdateSetting('led_brightness', val)}
                    min={0}
                    max={255}
                    step={10}
                    w={80}
                    size="sm"
                  />
                </Group>
                <Divider className="settings-modal__divider" />
                {/* Ticker content options */}
                <Text className="settings-modal__subsection-title" size="xs" fw={500} tt="uppercase" mb="xs">Ticker Content</Text>
                <Stack className="settings-modal__ticker-options" gap="xs">
                  <Switch
                    className="settings-modal__switch settings-modal__switch--ticker-value"
                    label="Portfolio value"
                    checked={getSetting('ticker_show_value', 1) === 1}
                    onChange={(e) => handleUpdateSetting('ticker_show_value', e.currentTarget.checked ? 1 : 0)}
                  />
                  <Switch
                    className="settings-modal__switch settings-modal__switch--ticker-cash"
                    label="Cash balance"
                    checked={getSetting('ticker_show_cash', 1) === 1}
                    onChange={(e) => handleUpdateSetting('ticker_show_cash', e.currentTarget.checked ? 1 : 0)}
                  />
                  <Switch
                    className="settings-modal__switch settings-modal__switch--ticker-actions"
                    label="Next actions"
                    checked={getSetting('ticker_show_actions', 1) === 1}
                    onChange={(e) => handleUpdateSetting('ticker_show_actions', e.currentTarget.checked ? 1 : 0)}
                  />
                  <Switch
                    className="settings-modal__switch settings-modal__switch--ticker-amounts"
                    label="Show amounts"
                    checked={getSetting('ticker_show_amounts', 1) === 1}
                    onChange={(e) => handleUpdateSetting('ticker_show_amounts', e.currentTarget.checked ? 1 : 0)}
                  />
                  <Group className="settings-modal__setting-row" justify="space-between">
                    <Text className="settings-modal__setting-name" size="sm">Max actions</Text>
                    <NumberInput
                      className="settings-modal__number-input"
                      value={getSetting('ticker_max_actions', 3)}
                      onChange={(val) => handleUpdateSetting('ticker_max_actions', val)}
                      min={1}
                      max={10}
                      step={1}
                      w={80}
                      size="sm"
                    />
                  </Group>
                </Stack>
              </Stack>
            </Paper>
          </Stack>
        </Tabs.Panel>

        {/* System Settings Tab */}
        <Tabs.Panel className="settings-modal__panel settings-modal__panel--system" value="system" p="md">
          <Stack className="settings-modal__system-content" gap="md">
            {/* Job Scheduling Section */}
            <Paper className="settings-modal__section settings-modal__section--scheduling" p="md" withBorder>
              <Text className="settings-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Job Scheduling</Text>
              <Text className="settings-modal__section-desc" size="xs" c="dimmed" mb="md">
                Simplified to 4 consolidated jobs: sync cycle (trading), daily pipeline (data), and maintenance.
              </Text>
              <Stack className="settings-modal__section-content" gap="sm">
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Sync Cycle</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Trades, prices, recommendations, execution</Text>
                  </div>
                  <Group className="settings-modal__setting-input" gap="xs">
                    <NumberInput
                      className="settings-modal__number-input"
                      value={getSetting('job_sync_cycle_minutes', 15)}
                      onChange={(val) => handleUpdateSetting('job_sync_cycle_minutes', val)}
                      min={5}
                      max={60}
                      step={5}
                      w={80}
                      size="sm"
                    />
                    <Text className="settings-modal__setting-unit" size="sm" c="dimmed">min</Text>
                  </Group>
                </Group>
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Maintenance</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Daily backup and cleanup hour</Text>
                  </div>
                  <Group className="settings-modal__setting-input" gap="xs">
                    <NumberInput
                      className="settings-modal__number-input"
                      value={getSetting('job_maintenance_hour', 3)}
                      onChange={(val) => handleUpdateSetting('job_maintenance_hour', val)}
                      min={0}
                      max={23}
                      step={1}
                      w={80}
                      size="sm"
                    />
                    <Text className="settings-modal__setting-unit" size="sm" c="dimmed">h</Text>
                  </Group>
                </Group>
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Auto-Deploy</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Check for updates and deploy changes</Text>
                  </div>
                  <Group className="settings-modal__setting-input" gap="xs">
                    <NumberInput
                      className="settings-modal__number-input"
                      value={getSetting('job_auto_deploy_minutes', 5)}
                      onChange={(val) => handleUpdateSetting('job_auto_deploy_minutes', val)}
                      min={0}
                      step={1}
                      w={80}
                      size="sm"
                    />
                    <Text className="settings-modal__setting-unit" size="sm" c="dimmed">min</Text>
                  </Group>
                </Group>
                <Divider className="settings-modal__divider" />
                <Text className="settings-modal__subsection-title" size="xs" fw={500} tt="uppercase" mb="xs">Fixed Schedules</Text>
                <Text className="settings-modal__fixed-schedule" size="xs" c="dimmed">Daily Pipeline: Hourly (per-symbol data sync)</Text>
                <Text className="settings-modal__fixed-schedule" size="xs" c="dimmed">Weekly Maintenance: Sundays (integrity checks)</Text>
              </Stack>
            </Paper>

            {/* System Actions Section */}
            <Paper className="settings-modal__section settings-modal__section--system-actions" p="md" withBorder>
              <Text className="settings-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">System</Text>
              <Stack className="settings-modal__section-content" gap="sm">
                {/* Cache reset action */}
                <Group className="settings-modal__action-row" justify="space-between">
                  <Text className="settings-modal__action-label" size="sm">Caches</Text>
                  <Button
                    className="settings-modal__action-btn settings-modal__action-btn--reset-cache"
                    size="xs"
                    variant="light"
                    onClick={handleResetCache}
                    loading={loading}
                  >
                    Reset
                  </Button>
                </Group>
                {/* Historical data sync action */}
                <Group className="settings-modal__action-row" justify="space-between">
                  <Text className="settings-modal__action-label" size="sm">Historical Data</Text>
                  <Button
                    className="settings-modal__action-btn settings-modal__action-btn--sync-historical"
                    size="xs"
                    variant="light"
                    onClick={handleSyncHistorical}
                    loading={syncingHistorical}
                  >
                    {syncingHistorical ? 'Syncing...' : 'Sync'}
                  </Button>
                </Group>
                {/* System restart action */}
                <Group className="settings-modal__action-row" justify="space-between">
                  <Text className="settings-modal__action-label" size="sm">System</Text>
                  <Button
                    className="settings-modal__action-btn settings-modal__action-btn--restart"
                    size="xs"
                    color="red"
                    variant="light"
                    onClick={handleRestartSystem}
                    loading={loading}
                  >
                    Restart
                  </Button>
                </Group>
              </Stack>
            </Paper>

            {/* Hardware Actions Section */}
            <Paper className="settings-modal__section settings-modal__section--hardware" p="md" withBorder>
              <Text className="settings-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Hardware</Text>
              <Text className="settings-modal__section-desc" size="xs" c="dimmed" mb="md">
                Manage Arduino MCU hardware. These actions only work when running on Arduino hardware.
              </Text>
              <Stack className="settings-modal__section-content" gap="sm">
                {/* Sketch upload action */}
                <Group className="settings-modal__action-row" justify="space-between">
                  <div className="settings-modal__action-label">
                    <Text className="settings-modal__action-name" size="sm">LED Display Sketch</Text>
                    <Text className="settings-modal__action-hint" size="xs" c="dimmed">Compile and upload sketch to MCU</Text>
                  </div>
                  <Button
                    className="settings-modal__action-btn settings-modal__action-btn--upload-sketch"
                    size="xs"
                    variant="light"
                    onClick={handleUploadSketch}
                    loading={uploadingSketch}
                  >
                    {uploadingSketch ? 'Uploading...' : 'Reflash'}
                  </Button>
                </Group>
              </Stack>
            </Paper>
          </Stack>
        </Tabs.Panel>

        {/* Backup Settings Tab */}
        <Tabs.Panel className="settings-modal__panel settings-modal__panel--backup" value="backup" p="md">
          <Stack className="settings-modal__backup-content" gap="md">
            {/* Cloudflare R2 Backup Section */}
            <Paper className="settings-modal__section settings-modal__section--r2" p="md" withBorder>
              <Text className="settings-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Cloudflare R2 Backup</Text>
              <Text className="settings-modal__section-desc" size="xs" c="dimmed" mb="md">
                Automatically backup databases to Cloudflare R2 cloud storage. Backups include all 7 databases in a single compressed archive.
              </Text>
              <Stack className="settings-modal__section-content" gap="md">
                <Switch
                  className="settings-modal__switch settings-modal__switch--r2-enabled"
                  label="Enable R2 backups"
                  checked={getSetting('r2_backup_enabled', 0) === 1}
                  onChange={(e) => handleUpdateSetting('r2_backup_enabled', e.currentTarget.checked ? 1 : 0)}
                  description="Automatically backup databases to Cloudflare R2 daily at 3:00 AM"
                />
                <Divider className="settings-modal__divider" />
                <Text className="settings-modal__subsection-title" size="xs" fw={500} tt="uppercase" mb="xs">R2 Configuration</Text>
                <TextInput
                  className="settings-modal__text-input settings-modal__text-input--account-id"
                  label="Account ID"
                  value={getSetting('r2_account_id', '') || ''}
                  onChange={(e) => handleUpdateSetting('r2_account_id', e.target.value)}
                  placeholder="a1b2c3d4e5f6g7h8i9j0"
                  description="Your Cloudflare account ID"
                />
                <TextInput
                  className="settings-modal__text-input settings-modal__text-input--access-key"
                  label="Access Key ID"
                  value={getSetting('r2_access_key_id', '') || ''}
                  onChange={(e) => handleUpdateSetting('r2_access_key_id', e.target.value)}
                  placeholder="a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
                  description="R2 access key ID for authentication"
                />
                <PasswordInput
                  className="settings-modal__password-input settings-modal__password-input--secret-key"
                  label="Secret Access Key"
                  value={getSetting('r2_secret_access_key', '') || ''}
                  onChange={(e) => handleUpdateSetting('r2_secret_access_key', e.target.value)}
                  placeholder="Enter your R2 secret access key"
                  description="R2 secret access key (hidden for security)"
                />
                <TextInput
                  className="settings-modal__text-input settings-modal__text-input--bucket"
                  label="Bucket Name"
                  value={getSetting('r2_bucket_name', '') || ''}
                  onChange={(e) => handleUpdateSetting('r2_bucket_name', e.target.value)}
                  placeholder="sentinel-backups"
                  description="Name of your R2 bucket for backups"
                />
                <Select
                  className="settings-modal__select settings-modal__select--backup-schedule"
                  label="Backup Schedule"
                  value={getSetting('r2_backup_schedule', 'daily') || 'daily'}
                  onChange={(val) => handleUpdateSetting('r2_backup_schedule', val)}
                  data={[
                    { value: 'daily', label: 'Daily (recommended)' },
                    { value: 'weekly', label: 'Weekly (Sundays)' },
                    { value: 'monthly', label: 'Monthly (1st of month)' }
                  ]}
                  description="How often to automatically backup to R2"
                />
                <Group className="settings-modal__setting-row" justify="space-between">
                  <div className="settings-modal__setting-label">
                    <Text className="settings-modal__setting-name" size="sm">Retention Days</Text>
                    <Text className="settings-modal__setting-hint" size="xs" c="dimmed">Keep backups for this many days (0 = forever)</Text>
                  </div>
                  <NumberInput
                    className="settings-modal__number-input"
                    value={getSetting('r2_backup_retention_days', 90)}
                    onChange={(val) => handleUpdateSetting('r2_backup_retention_days', val)}
                    min={0}
                    step={1}
                    w={100}
                    size="sm"
                  />
                </Group>
                <Divider className="settings-modal__divider" />
                <Text className="settings-modal__subsection-title" size="xs" fw={500} tt="uppercase" mb="xs">Actions</Text>
                <Group className="settings-modal__r2-actions" gap="sm">
                  <Button
                    className="settings-modal__action-btn settings-modal__action-btn--test-r2"
                    size="sm"
                    variant="light"
                    onClick={handleTestR2Connection}
                    loading={testingR2Connection}
                    disabled={!getSetting('r2_account_id', '') || !getSetting('r2_access_key_id', '')}
                  >
                    Test Connection
                  </Button>
                  <Button
                    className="settings-modal__action-btn settings-modal__action-btn--view-backups"
                    size="sm"
                    variant="light"
                    onClick={handleViewR2Backups}
                    disabled={!getSetting('r2_account_id', '') || !getSetting('r2_access_key_id', '')}
                  >
                    View Backups
                  </Button>
                  <Button
                    className="settings-modal__action-btn settings-modal__action-btn--backup-now"
                    size="sm"
                    variant="filled"
                    onClick={handleBackupToR2}
                    loading={backingUpToR2}
                    disabled={!getSetting('r2_account_id', '') || !getSetting('r2_access_key_id', '')}
                  >
                    Backup Now
                  </Button>
                </Group>
                <Alert className="settings-modal__alert" color="blue" size="sm">
                  <Text className="settings-modal__alert-text" size="xs">
                    Backups are stored as compressed .tar.gz archives containing all databases.
                    Automatic backups run according to your schedule at 3:00 AM. Old backups are rotated based on retention policy.
                  </Text>
                </Alert>
              </Stack>
            </Paper>
          </Stack>
        </Tabs.Panel>

        {/* Credentials Settings Tab */}
        <Tabs.Panel className="settings-modal__panel settings-modal__panel--credentials" value="credentials" p="md">
          <Stack className="settings-modal__credentials-content" gap="md">
            {/* API Credentials Section */}
            <Paper className="settings-modal__section settings-modal__section--api-credentials" p="md" withBorder>
              <Text className="settings-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">API Credentials</Text>
              <Text className="settings-modal__section-desc" size="xs" c="dimmed" mb="md">
                Configure API keys for external services. Credentials are stored securely in the database.
                The .env file is no longer required - all configuration can be managed through this UI.
              </Text>
              <Stack className="settings-modal__section-content" gap="md">
                <TextInput
                  className="settings-modal__text-input settings-modal__text-input--tradernet-key"
                  label="Tradernet API Key"
                  value={getSetting('tradernet_api_key', '') || ''}
                  onChange={(e) => handleUpdateSetting('tradernet_api_key', e.target.value)}
                  placeholder="Enter your Tradernet API key"
                  description="Your Tradernet API key for accessing trading services"
                />
                <TextInput
                  className="settings-modal__text-input settings-modal__text-input--tradernet-secret"
                  label="Tradernet API Secret"
                  type="password"
                  value={getSetting('tradernet_api_secret', '') || ''}
                  onChange={(e) => handleUpdateSetting('tradernet_api_secret', e.target.value)}
                  placeholder="Enter your Tradernet API secret"
                  description="Your Tradernet API secret (hidden for security)"
                />
                <Divider className="settings-modal__divider" />
                <TextInput
                  className="settings-modal__text-input settings-modal__text-input--github-token"
                  label="GitHub Token"
                  type="password"
                  value={getSetting('github_token', '') || ''}
                  onChange={(e) => handleUpdateSetting('github_token', e.target.value)}
                  placeholder="ghp_your_token_here"
                  description="GitHub personal access token for auto-deployment artifact downloads (requires repo and actions:read scopes)"
                />
                <Divider className="settings-modal__divider" />
                <Alert className="settings-modal__alert" color="blue" size="sm">
                  <Text className="settings-modal__alert-text" size="xs">
                    Credentials are stored in the settings database and take precedence over environment variables.
                    Changes are applied immediately - no restart required.
                  </Text>
                </Alert>
              </Stack>
            </Paper>
          </Stack>
        </Tabs.Panel>
      </Tabs>
      </Modal>
      <R2BackupModal
        opened={showR2BackupModal}
        onClose={() => setShowR2BackupModal(false)}
      />
    </>
  );
}
