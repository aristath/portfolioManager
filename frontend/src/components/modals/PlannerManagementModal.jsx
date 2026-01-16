/**
 * Planner Management Modal Component
 * 
 * Comprehensive modal for configuring the portfolio planning system.
 * 
 * Features:
 * - Temperament Settings: Three sliders (Risk Tolerance, Aggression, Patience) that control 150+ parameters
 * - General Settings: Batch processing, trade permissions
 * - Planner Settings: Sequence selection, portfolio optimizer configuration
 * - Transaction Costs: Fixed and variable cost settings
 * - Opportunity Calculators: Enable/disable individual opportunity calculators
 * - Filters: Post-generation filters and tag-based filtering
 * 
 * The temperament sliders are stored in global settings, while planner-specific
 * configuration is stored in the planner config database.
 */
import { useState, useEffect, useCallback } from 'react';
import { Modal, Tabs, Text, Button, Switch, NumberInput, Slider, Group, Stack, Paper, Alert, Loader, Divider } from '@mantine/core';
import { IconInfoCircle } from '@tabler/icons-react';
import { useAppStore } from '../../stores/appStore';
import { api } from '../../api/client';
import { useNotifications } from '../../hooks/useNotifications';

/**
 * Default planner configuration values
 * Used as fallback when loading configuration fails
 */
const DEFAULT_CONFIG = {
  enable_batch_generation: true,
  enable_diverse_selection: true,
  diversity_weight: 0.3,
  transaction_cost_fixed: 5.0,
  transaction_cost_percent: 0.001,
  allow_sell: true,
  allow_buy: true,
  // Portfolio optimizer
  optimizer_target_return: 0.11,
  min_cash_reserve: 500.0,
  // Opportunity Calculators
  enable_profit_taking_calc: true,
  enable_averaging_down_calc: true,
  enable_opportunity_buys_calc: true,
  enable_rebalance_sells_calc: true,
  enable_rebalance_buys_calc: true,
  enable_weight_based_calc: true,
  // Portfolio optimizer
  optimizer_blend: 0.5,
  // Post-generation Filters (eligibility & recently_traded are now handled during generation)
  enable_correlation_aware_filter: true,
  enable_diversity_filter: true,
  // Tag filtering
  enable_tag_filtering: true,
};

/**
 * Default temperament settings (stored in global settings, not planner config)
 * These three sliders control 150+ parameters across the system
 */
const DEFAULT_TEMPERAMENT = {
  risk_tolerance: 0.5,       // Conservative (0) to Risk-Taking (1)
  temperament_aggression: 0.5, // Passive (0) to Aggressive (1)
  temperament_patience: 0.5,   // Impatient (0) to Patient (1)
};

/**
 * Planner management modal component
 * 
 * Provides comprehensive configuration interface for the portfolio planning system.
 * 
 * @returns {JSX.Element} Planner configuration modal with tabbed interface
 */
export function PlannerManagementModal() {
  const { showPlannerManagementModal, closePlannerManagementModal } = useAppStore();
  const { showNotification } = useNotifications();
  const [activeTab, setActiveTab] = useState('temperament');
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [config, setConfig] = useState(DEFAULT_CONFIG);
  const [temperament, setTemperament] = useState(DEFAULT_TEMPERAMENT);

  /**
   * Loads planner configuration and temperament settings from the backend
   */
  const loadConfig = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      // Fetch planner config and temperament settings in parallel
      const [plannerResponse, settingsResponse] = await Promise.all([
        api.fetchPlannerConfig(),
        api.fetchSettings(),
      ]);
      setConfig(plannerResponse.config || DEFAULT_CONFIG);

      // Extract temperament settings from global settings
      const settings = settingsResponse.settings || {};
      setTemperament({
        risk_tolerance: settings.risk_tolerance ?? 0.5,
        temperament_aggression: settings.temperament_aggression ?? 0.5,
        temperament_patience: settings.temperament_patience ?? 0.5,
      });
    } catch (error) {
      setError(`Failed to load configuration: ${error.message}`);
      showNotification(`Failed to load configuration: ${error.message}`, 'error');
      // Use defaults on error
      setConfig(DEFAULT_CONFIG);
      setTemperament(DEFAULT_TEMPERAMENT);
    } finally {
      setLoading(false);
    }
  }, [showNotification]);

  // Load configuration when modal opens
  useEffect(() => {
    if (showPlannerManagementModal) {
      loadConfig();
    }
  }, [showPlannerManagementModal, loadConfig]);

  /**
   * Saves planner configuration and temperament settings to the backend
   */
  const handleSave = async () => {
    setSaving(true);
    setError(null);

    try {
      // Save planner config and temperament settings in parallel
      await Promise.all([
        api.updatePlannerConfig(config, 'ui', 'Updated via UI'),
        api.updateSetting('risk_tolerance', temperament.risk_tolerance),
        api.updateSetting('temperament_aggression', temperament.temperament_aggression),
        api.updateSetting('temperament_patience', temperament.temperament_patience),
      ]);
      showNotification('Configuration saved successfully', 'success');
    } catch (error) {
      const errorMsg = error.message || 'Failed to save configuration';
      setError(errorMsg);
      showNotification(errorMsg, 'error');
    } finally {
      setSaving(false);
    }
  };

  /**
   * Updates a planner configuration field
   * 
   * @param {string} field - Configuration field name
   * @param {*} value - New value
   */
  const updateConfig = (field, value) => {
    setConfig({ ...config, [field]: value });
  };

  /**
   * Updates a temperament setting
   * 
   * @param {string} field - Temperament field name
   * @param {*} value - New value
   */
  const updateTemperament = (field, value) => {
    setTemperament({ ...temperament, [field]: value });
  };

  /**
   * Gets a configuration value with default fallback
   * 
   * @param {string} field - Configuration field name
   * @param {*} defaultValue - Default value if not found
   * @returns {*} Configuration value or default
   */
  const getConfigValue = (field, defaultValue) => {
    return config[field] ?? defaultValue;
  };

  /**
   * Gets a temperament value with default fallback
   * 
   * @param {string} field - Temperament field name
   * @param {*} defaultValue - Default value if not found
   * @returns {*} Temperament value or default
   */
  const getTemperamentValue = (field, defaultValue) => {
    return temperament[field] ?? defaultValue;
  };

  return (
    <Modal
      className="planner-modal"
      opened={showPlannerManagementModal}
      onClose={closePlannerManagementModal}
      title="Planner Configuration"
      size="xl"
      styles={{ body: { padding: 0 } }}
    >
      {loading ? (
        <Group className="planner-modal__loading" justify="center" p="xl">
          <Loader className="planner-modal__loader" />
          <Text className="planner-modal__loading-text" c="dimmed">Loading configuration...</Text>
        </Group>
      ) : (
        <>
          {error && (
            <Alert className="planner-modal__error" color="red" title="Error" m="md">
              {error}
            </Alert>
          )}

          <Tabs className="planner-modal__tabs" value={activeTab} onChange={setActiveTab}>
            <Tabs.List className="planner-modal__tab-list" grow>
              <Tabs.Tab className="planner-modal__tab planner-modal__tab--temperament" value="temperament">Temperament</Tabs.Tab>
              <Tabs.Tab className="planner-modal__tab planner-modal__tab--general" value="general">General</Tabs.Tab>
              <Tabs.Tab className="planner-modal__tab planner-modal__tab--planner" value="planner">Planner</Tabs.Tab>
              <Tabs.Tab className="planner-modal__tab planner-modal__tab--transaction" value="transaction">Costs</Tabs.Tab>
              <Tabs.Tab className="planner-modal__tab planner-modal__tab--calculators" value="calculators">Calculators</Tabs.Tab>
              <Tabs.Tab className="planner-modal__tab planner-modal__tab--filters" value="filters">Filters</Tabs.Tab>
            </Tabs.List>

            {/* Temperament Tab - Three sliders that control 150+ parameters */}
            <Tabs.Panel className="planner-modal__panel planner-modal__panel--temperament" value="temperament" p="md">
              <Stack className="planner-modal__temperament-content" gap="md">
                {/* Info alert explaining temperament system */}
                <Alert className="planner-modal__alert" color="blue" title="Investment Temperament" icon={<IconInfoCircle />}>
                  These three sliders control 150+ parameters across the system, defining how the planner behaves.
                  Move sliders to adjust your investment philosophy. Changes affect evaluation weights, thresholds,
                  hold periods, position sizing, and more.
                </Alert>

                {/* Risk Tolerance Slider */}
                <Paper className="planner-modal__section planner-modal__section--risk" p="md" withBorder>
                  <Text className="planner-modal__section-title" size="sm" fw={500} mb="md" tt="uppercase">Risk Tolerance</Text>
                  <Text className="planner-modal__section-desc" size="xs" c="dimmed" mb="md">
                    Controls volatility acceptance, drawdown tolerance, position concentration, and quality floors.
                    Conservative investors prefer stable, high-quality positions while risk-takers accept more volatility for higher returns.
                  </Text>
                  <div className="planner-modal__slider-container">
                    <Group className="planner-modal__slider-labels" justify="space-between" mb="xs">
                      <Text className="planner-modal__slider-min" size="sm">Conservative</Text>
                      <Text className="planner-modal__slider-value" size="sm" fw={500}>
                        {(getTemperamentValue('risk_tolerance', 0.5) * 100).toFixed(0)}%
                      </Text>
                      <Text className="planner-modal__slider-max" size="sm">Risk-Taking</Text>
                    </Group>
                    <Slider
                      className="planner-modal__slider planner-modal__slider--risk"
                      value={getTemperamentValue('risk_tolerance', 0.5)}
                      onChange={(val) => updateTemperament('risk_tolerance', val)}
                      min={0}
                      max={1}
                      step={0.01}
                      marks={[
                        { value: 0, label: '0' },
                        { value: 0.25, label: '25' },
                        { value: 0.5, label: '50' },
                        { value: 0.75, label: '75' },
                        { value: 1, label: '100' },
                      ]}
                      mb="xl"
                    />
                  </div>
                </Paper>

                {/* Aggression Slider */}
                <Paper className="planner-modal__section planner-modal__section--aggression" p="md" withBorder>
                  <Text className="planner-modal__section-title" size="sm" fw={500} mb="md" tt="uppercase">Aggression</Text>
                  <Text className="planner-modal__section-desc" size="xs" c="dimmed" mb="md">
                    Controls scoring thresholds, action frequency, evaluation weights, position sizing, and opportunity pursuit.
                    Passive investors wait for clear opportunities while aggressive investors act more readily on signals.
                  </Text>
                  <div className="planner-modal__slider-container">
                    <Group className="planner-modal__slider-labels" justify="space-between" mb="xs">
                      <Text className="planner-modal__slider-min" size="sm">Passive</Text>
                      <Text className="planner-modal__slider-value" size="sm" fw={500}>
                        {(getTemperamentValue('temperament_aggression', 0.5) * 100).toFixed(0)}%
                      </Text>
                      <Text className="planner-modal__slider-max" size="sm">Aggressive</Text>
                    </Group>
                    <Slider
                      className="planner-modal__slider planner-modal__slider--aggression"
                      value={getTemperamentValue('temperament_aggression', 0.5)}
                      onChange={(val) => updateTemperament('temperament_aggression', val)}
                      min={0}
                      max={1}
                      step={0.01}
                      marks={[
                        { value: 0, label: '0' },
                        { value: 0.25, label: '25' },
                        { value: 0.5, label: '50' },
                        { value: 0.75, label: '75' },
                        { value: 1, label: '100' },
                      ]}
                      mb="xl"
                    />
                  </div>
                </Paper>

                {/* Patience Slider */}
                <Paper className="planner-modal__section planner-modal__section--patience" p="md" withBorder>
                  <Text className="planner-modal__section-title" size="sm" fw={500} mb="md" tt="uppercase">Patience</Text>
                  <Text className="planner-modal__section-desc" size="xs" c="dimmed" mb="md">
                    Controls hold periods, cooldowns, windfall thresholds, rebalance triggers, and dividend focus.
                    Impatient investors seek quick wins while patient investors let positions mature.
                  </Text>
                  <div className="planner-modal__slider-container">
                    <Group className="planner-modal__slider-labels" justify="space-between" mb="xs">
                      <Text className="planner-modal__slider-min" size="sm">Impatient</Text>
                      <Text className="planner-modal__slider-value" size="sm" fw={500}>
                        {(getTemperamentValue('temperament_patience', 0.5) * 100).toFixed(0)}%
                      </Text>
                      <Text className="planner-modal__slider-max" size="sm">Patient</Text>
                    </Group>
                    <Slider
                      className="planner-modal__slider planner-modal__slider--patience"
                      value={getTemperamentValue('temperament_patience', 0.5)}
                      onChange={(val) => updateTemperament('temperament_patience', val)}
                      min={0}
                      max={1}
                      step={0.01}
                      marks={[
                        { value: 0, label: '0' },
                        { value: 0.25, label: '25' },
                        { value: 0.5, label: '50' },
                        { value: 0.75, label: '75' },
                        { value: 1, label: '100' },
                      ]}
                      mb="xl"
                    />
                  </div>
                </Paper>
              </Stack>
            </Tabs.Panel>

            {/* General Tab - Basic planner settings */}
            <Tabs.Panel className="planner-modal__panel planner-modal__panel--general" value="general" p="md">
              <Stack className="planner-modal__general-content" gap="md">
                {/* Batch Processing Section */}
                <Paper className="planner-modal__section planner-modal__section--batch" p="md" withBorder>
                  <Text className="planner-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Batch Processing</Text>
                  <Stack className="planner-modal__section-content" gap="sm">
                    <Switch
                      className="planner-modal__switch planner-modal__switch--batch"
                      label="Enable Batch Generation"
                      checked={getConfigValue('enable_batch_generation', true)}
                      onChange={(e) => updateConfig('enable_batch_generation', e.currentTarget.checked)}
                      description="Generate sequences in batches for better performance"
                    />
                  </Stack>
                </Paper>

                {/* Trade Permissions Section */}
                <Paper className="planner-modal__section planner-modal__section--permissions" p="md" withBorder>
                  <Text className="planner-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Trade Permissions</Text>
                  <Stack className="planner-modal__section-content" gap="sm">
                    <Switch
                      className="planner-modal__switch planner-modal__switch--allow-buy"
                      label="Allow Buy Orders"
                      checked={getConfigValue('allow_buy', true)}
                      onChange={(e) => updateConfig('allow_buy', e.currentTarget.checked)}
                    />
                    <Switch
                      className="planner-modal__switch planner-modal__switch--allow-sell"
                      label="Allow Sell Orders"
                      checked={getConfigValue('allow_sell', true)}
                      onChange={(e) => updateConfig('allow_sell', e.currentTarget.checked)}
                    />
                  </Stack>
                </Paper>
              </Stack>
            </Tabs.Panel>

            {/* Planner Settings Tab - Sequence selection and optimizer configuration */}
            <Tabs.Panel className="planner-modal__panel planner-modal__panel--planner" value="planner" p="md">
              <Stack className="planner-modal__planner-content" gap="md">
                {/* Sequence Selection Section */}
                <Paper className="planner-modal__section planner-modal__section--sequence" p="md" withBorder>
                  <Text className="planner-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Sequence Selection</Text>
                  <Stack className="planner-modal__section-content" gap="md">
                    <Switch
                      className="planner-modal__switch planner-modal__switch--diverse"
                      label="Enable Diverse Selection"
                      checked={getConfigValue('enable_diverse_selection', true)}
                      onChange={(e) => updateConfig('enable_diverse_selection', e.currentTarget.checked)}
                      description="Select diverse sequences to avoid redundancy"
                    />

                    {/* Diversity weight slider */}
                    <div className="planner-modal__slider-container">
                      <Group className="planner-modal__slider-labels" justify="space-between" mb="xs">
                        <Text className="planner-modal__setting-name" size="sm">Diversity Weight</Text>
                        <Text className="planner-modal__slider-value" size="sm" fw={500}>
                          {getConfigValue('diversity_weight', 0.3).toFixed(2)}
                        </Text>
                      </Group>
                      <Slider
                        className="planner-modal__slider planner-modal__slider--diversity"
                        value={getConfigValue('diversity_weight', 0.3)}
                        onChange={(val) => updateConfig('diversity_weight', val)}
                        min={0}
                        max={1}
                        step={0.01}
                        mb="xs"
                      />
                      <Text className="planner-modal__setting-hint" size="xs" c="dimmed">Weight for diversity in sequence selection (0.0 - 1.0)</Text>
                    </div>
                  </Stack>
                </Paper>

                {/* Portfolio Optimizer Section */}
                <Paper className="planner-modal__section planner-modal__section--optimizer" p="md" withBorder>
                  <Text className="planner-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Portfolio Optimizer</Text>
                  <Stack className="planner-modal__section-content" gap="md">
                    {/* Target return setting */}
                    <Group className="planner-modal__setting-row" justify="space-between">
                      <div className="planner-modal__setting-label">
                        <Text className="planner-modal__setting-name" size="sm">Target Return</Text>
                        <Text className="planner-modal__setting-hint" size="xs" c="dimmed">Annual return goal for optimizer</Text>
                      </div>
                      <Group className="planner-modal__setting-input" gap="xs">
                        <NumberInput
                          className="planner-modal__number-input"
                          value={(getConfigValue('optimizer_target_return', 0.11) * 100).toFixed(0)}
                          onChange={(val) => updateConfig('optimizer_target_return', (val || 0) / 100)}
                          min={0}
                          step={1}
                          w={80}
                          size="sm"
                        />
                        <Text className="planner-modal__setting-unit" size="sm" c="dimmed">%</Text>
                      </Group>
                    </Group>

                    {/* Strategy blend slider (read-only, algorithm-controlled) */}
                    <div className="planner-modal__slider-container">
                      <Group className="planner-modal__slider-labels" justify="space-between" mb="xs">
                        <Text className="planner-modal__setting-name" size="sm">Strategy Blend</Text>
                        <Text className="planner-modal__slider-value" size="sm" fw={500}>
                          {(getConfigValue('optimizer_blend', 0.5) * 100).toFixed(0)}%
                        </Text>
                      </Group>
                      <Group className="planner-modal__blend-slider" gap="xs" mb="xs">
                        <Text className="planner-modal__blend-label" size="xs" c="dimmed">MV</Text>
                        <Slider
                          className="planner-modal__slider planner-modal__slider--blend"
                          value={getConfigValue('optimizer_blend', 0.5)}
                          onChange={() => {}} // Read-only: algorithm-controlled
                          min={0}
                          max={1}
                          step={0.05}
                          style={{ flex: 1 }}
                          disabled
                        />
                        <Text className="planner-modal__blend-label" size="xs" c="dimmed">HRP</Text>
                      </Group>
                      <Text className="planner-modal__setting-hint" size="xs" c="dimmed">
                        Algorithm-controlled based on market regime. 0% = Goal-directed (Mean-Variance), 100% = Robust (HRP)
                      </Text>
                    </div>

                    {/* Minimum cash reserve setting */}
                    <Group className="planner-modal__setting-row" justify="space-between">
                      <div className="planner-modal__setting-label">
                        <Text className="planner-modal__setting-name" size="sm">Min Cash Reserve</Text>
                        <Text className="planner-modal__setting-hint" size="xs" c="dimmed">Never deploy below this amount</Text>
                      </div>
                      <Group className="planner-modal__setting-input" gap="xs">
                        <Text className="planner-modal__setting-unit" size="sm" c="dimmed">EUR</Text>
                        <NumberInput
                          className="planner-modal__number-input"
                          value={getConfigValue('min_cash_reserve', 500)}
                          onChange={(val) => updateConfig('min_cash_reserve', val)}
                          min={0}
                          step={50}
                          w={100}
                          size="sm"
                        />
                      </Group>
                    </Group>
                  </Stack>
                </Paper>

                {/* Info alert about risk management settings */}
                <Alert className="planner-modal__alert planner-modal__alert--info" color="gray" variant="light">
                  <Text className="planner-modal__alert-text" size="xs">
                    Risk management settings (hold periods, cooldowns, loss thresholds, sell percentages)
                    are now controlled by the Temperament sliders. Adjust the Patience and Risk Tolerance
                    sliders to change these behaviors.
                  </Text>
                </Alert>
              </Stack>
            </Tabs.Panel>

            {/* Transaction Costs Tab */}
            <Tabs.Panel className="planner-modal__panel planner-modal__panel--transaction" value="transaction" p="md">
              <Stack className="planner-modal__transaction-content" gap="md">
                {/* Transaction Costs Section */}
                <Paper className="planner-modal__section planner-modal__section--costs" p="md" withBorder>
                  <Text className="planner-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Transaction Costs</Text>
                  <Text className="planner-modal__section-desc" size="xs" c="dimmed" mb="md">
                    Transaction costs are used to evaluate sequence quality.
                  </Text>
                  <Stack className="planner-modal__section-content" gap="sm">
                    <Group className="planner-modal__setting-row" justify="space-between">
                      <div className="planner-modal__setting-label">
                        <Text className="planner-modal__setting-name" size="sm">Fixed Cost</Text>
                        <Text className="planner-modal__setting-hint" size="xs" c="dimmed">Fixed cost per trade</Text>
                      </div>
                      <Group className="planner-modal__setting-input" gap="xs">
                        <NumberInput
                          className="planner-modal__number-input"
                          value={getConfigValue('transaction_cost_fixed', 5.0)}
                          onChange={(val) => updateConfig('transaction_cost_fixed', val)}
                          min={0}
                          step={0.5}
                          precision={2}
                          w={100}
                          size="sm"
                        />
                      </Group>
                    </Group>

                    <Group className="planner-modal__setting-row" justify="space-between">
                      <div className="planner-modal__setting-label">
                        <Text className="planner-modal__setting-name" size="sm">Variable Cost</Text>
                        <Text className="planner-modal__setting-hint" size="xs" c="dimmed">Percentage of trade value</Text>
                      </div>
                      <Group className="planner-modal__setting-input" gap="xs">
                        <NumberInput
                          className="planner-modal__number-input"
                          value={(getConfigValue('transaction_cost_percent', 0.001) * 100).toFixed(3)}
                          onChange={(val) => updateConfig('transaction_cost_percent', (val || 0) / 100)}
                          min={0}
                          step={0.001}
                          precision={3}
                          w={100}
                          size="sm"
                        />
                        <Text className="planner-modal__setting-unit" size="sm" c="dimmed">%</Text>
                      </Group>
                    </Group>
                  </Stack>
                </Paper>
              </Stack>
            </Tabs.Panel>

            {/* Opportunity Calculators Tab - Enable/disable individual calculators */}
            <Tabs.Panel className="planner-modal__panel planner-modal__panel--calculators" value="calculators" p="md">
              <Stack className="planner-modal__calculators-content" gap="md">
                {/* Opportunity Calculators Section */}
                <Paper className="planner-modal__section planner-modal__section--calculators" p="md" withBorder>
                  <Text className="planner-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Opportunity Calculators</Text>
                  <Text className="planner-modal__section-desc" size="xs" c="dimmed" mb="md">
                    Enable or disable opportunity calculators that identify trading opportunities.
                  </Text>
                  <Stack className="planner-modal__section-content" gap="sm">
                    <Switch
                      className="planner-modal__switch planner-modal__switch--profit-taking"
                      label="Profit Taking Calculator"
                      checked={getConfigValue('enable_profit_taking_calc', true)}
                      onChange={(e) => updateConfig('enable_profit_taking_calc', e.currentTarget.checked)}
                    />
                    <Switch
                      className="planner-modal__switch planner-modal__switch--averaging-down"
                      label="Averaging Down Calculator"
                      checked={getConfigValue('enable_averaging_down_calc', true)}
                      onChange={(e) => updateConfig('enable_averaging_down_calc', e.currentTarget.checked)}
                    />
                    <Switch
                      className="planner-modal__switch planner-modal__switch--opportunity-buys"
                      label="Opportunity Buys Calculator"
                      checked={getConfigValue('enable_opportunity_buys_calc', true)}
                      onChange={(e) => updateConfig('enable_opportunity_buys_calc', e.currentTarget.checked)}
                    />
                    <Switch
                      className="planner-modal__switch planner-modal__switch--rebalance-sells"
                      label="Rebalance Sells Calculator"
                      checked={getConfigValue('enable_rebalance_sells_calc', true)}
                      onChange={(e) => updateConfig('enable_rebalance_sells_calc', e.currentTarget.checked)}
                    />
                    <Switch
                      className="planner-modal__switch planner-modal__switch--rebalance-buys"
                      label="Rebalance Buys Calculator"
                      checked={getConfigValue('enable_rebalance_buys_calc', true)}
                      onChange={(e) => updateConfig('enable_rebalance_buys_calc', e.currentTarget.checked)}
                    />
                    <Switch
                      className="planner-modal__switch planner-modal__switch--weight-based"
                      label="Weight Based Calculator"
                      checked={getConfigValue('enable_weight_based_calc', true)}
                      onChange={(e) => updateConfig('enable_weight_based_calc', e.currentTarget.checked)}
                    />
                  </Stack>
                </Paper>
              </Stack>
            </Tabs.Panel>

            {/* Filters Tab - Post-generation filters and tag-based filtering */}
            <Tabs.Panel className="planner-modal__panel planner-modal__panel--filters" value="filters" p="md">
              <Stack className="planner-modal__filters-content" gap="md">
                {/* Post-Generation Filters Section */}
                <Paper className="planner-modal__section planner-modal__section--filters" p="md" withBorder>
                  <Text className="planner-modal__section-title" size="sm" fw={500} mb="xs" tt="uppercase">Post-Generation Filters</Text>
                  <Text className="planner-modal__section-desc" size="xs" c="dimmed" mb="md">
                    Filters that refine generated sequences after generation.
                    Eligibility and cooloff checks are now performed during generation for early pruning.
                  </Text>
                  <Stack className="planner-modal__section-content" gap="sm">
                    <Switch
                      className="planner-modal__switch planner-modal__switch--correlation"
                      label="Correlation Aware Filter"
                      checked={getConfigValue('enable_correlation_aware_filter', true)}
                      onChange={(e) => updateConfig('enable_correlation_aware_filter', e.currentTarget.checked)}
                      description="Filters sequences with highly correlated actions"
                    />
                    <Switch
                      className="planner-modal__switch planner-modal__switch--diversity"
                      label="Diversity Filter"
                      checked={getConfigValue('enable_diversity_filter', true)}
                      onChange={(e) => updateConfig('enable_diversity_filter', e.currentTarget.checked)}
                      description="Ensures sequences include diverse actions"
                    />
                    <Divider className="planner-modal__divider" my="sm" />
                    <Switch
                      className="planner-modal__switch planner-modal__switch--tag-filtering"
                      label="Tag-Based Filtering"
                      checked={getConfigValue('enable_tag_filtering', true)}
                      onChange={(e) => updateConfig('enable_tag_filtering', e.currentTarget.checked)}
                      description="Enable tag-based pre-filtering for opportunity identification. When disabled, all active securities are considered (uses score-based quality checks instead)."
                    />
                  </Stack>
                </Paper>
              </Stack>
            </Tabs.Panel>
          </Tabs>

          <Divider className="planner-modal__footer-divider" />

          <Group className="planner-modal__actions" justify="flex-end" p="md">
            <Button
              className="planner-modal__cancel-btn"
              variant="subtle"
              onClick={closePlannerManagementModal}
            >
              Cancel
            </Button>
            <Button
              className="planner-modal__save-btn"
              onClick={handleSave}
              disabled={saving}
              loading={saving}
            >
              Save Configuration
            </Button>
          </Group>
        </>
      )}
    </Modal>
  );
}
