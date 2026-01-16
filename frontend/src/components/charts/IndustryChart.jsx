/**
 * Industry Chart Component
 * 
 * Displays industry/sector allocation with deviation visualization and weight editing.
 * Shows current allocation vs target allocation with visual deviation bars.
 * 
 * Features:
 * - View mode: Shows current allocation and deviation from target
 * - Edit mode: Weight sliders for adjusting target allocation (0-1 scale)
 * - Deviation bars: Visual representation of over/under-allocation
 * - Weight normalization (weights converted to percentages)
 * 
 * Used in IndustryRadarCard for detailed industry allocation management.
 */
import { useMemo } from 'react';
import { Group, Text, Button, Slider, Badge, Stack, Divider } from '@mantine/core';
import { usePortfolioStore } from '../../stores/portfolioStore';
import { useNotifications } from '../../hooks/useNotifications';
import { formatPercent } from '../../utils/formatters';

/**
 * Converts allocation weights to target percentages
 * 
 * Normalizes weights so they sum to 100% for display purposes.
 * 
 * @param {Object} weights - Object mapping industry names to weights
 * @param {Array<string>} activeIndustries - Array of active industry names
 * @returns {Object} Object mapping industry names to normalized percentages (0-1)
 */
function getTargetPcts(weights, activeIndustries) {
  let total = 0;
  for (const name of activeIndustries) {
    const weight = weights[name] || 0;
    total += weight;
  }

  const targets = {};
  for (const name of activeIndustries) {
    const weight = weights[name] || 0;
    targets[name] = total > 0 ? weight / total : 0;
  }
  return targets;
}

/**
 * Calculates deviation from target: current% - target%
 * 
 * @param {string} name - Industry name
 * @param {number} currentPct - Current allocation percentage (0-1)
 * @param {Object} targets - Target percentages object
 * @returns {number} Deviation value (positive = over-allocated, negative = under-allocated)
 */
function getDeviation(name, currentPct, targets) {
  const targetPct = targets[name] || 0;
  return currentPct - targetPct;
}

/**
 * Formats deviation as percentage string with sign
 * 
 * @param {number} deviation - Deviation value
 * @returns {string} Formatted deviation (e.g., "+5.2%" or "-3.1%")
 */
function formatDeviation(deviation) {
  const pct = (deviation * 100).toFixed(1);
  return (deviation >= 0 ? '+' : '') + pct + '%';
}

/**
 * Gets badge color class based on deviation magnitude
 * 
 * @param {number} deviation - Deviation value
 * @returns {Object} Badge props with color and variant
 */
function getDeviationBadgeClass(deviation) {
  if (Math.abs(deviation) < 0.02) return { color: 'gray', variant: 'light' };  // Within 2% = neutral
  return deviation > 0
    ? { color: 'red', variant: 'light' }   // Over-allocated
    : { color: 'blue', variant: 'light' }; // Under-allocated
}

/**
 * Gets deviation bar color based on deviation direction
 * 
 * @param {number} deviation - Deviation value
 * @returns {string} Color name for deviation bar
 */
function getDeviationBarColor(deviation) {
  if (Math.abs(deviation) < 0.02) return 'gray';  // Within 2% = neutral
  return deviation > 0 ? 'red' : 'blue';  // Red for over, blue for under
}

/**
 * Gets deviation bar style (position and width)
 * 
 * Bar extends from center (50%) based on deviation magnitude.
 * 
 * @param {number} deviation - Deviation value
 * @returns {Object} CSS style object for deviation bar
 */
function getDeviationBarStyle(deviation) {
  const maxDev = 0.50;  // Maximum deviation to display (50%)
  const pct = Math.min(Math.abs(deviation), maxDev) / maxDev * 50;  // Scale to 0-50%

  if (deviation >= 0) {
    // Positive deviation: bar extends right from center
    return { width: `${pct}%`, left: '50%', right: 'auto' };
  } else {
    // Negative deviation: bar extends left from center
    return { width: `${pct}%`, right: '50%', left: 'auto' };
  }
}

/**
 * Formats weight value for display
 * 
 * @param {number} weight - Weight value (0-1)
 * @returns {string} Formatted weight string
 */
function formatWeight(weight) {
  if (weight === 0 || weight === undefined) return '0';
  return weight.toFixed(2);
}

/**
 * Gets badge color class based on weight value
 * 
 * @param {number} weight - Weight value (0-1)
 * @returns {Object} Badge props with color and variant
 */
function getWeightBadgeClass(weight) {
  if (weight > 0.7) return { color: 'green', variant: 'light' };   // High priority
  if (weight < 0.3) return { color: 'red', variant: 'light' };     // Low priority
  return { color: 'gray', variant: 'light' };                      // Neutral
}

/**
 * Industry chart component
 * 
 * Displays industry/sector allocation with view and edit modes.
 * 
 * @returns {JSX.Element} Industry allocation chart with deviation visualization
 */
export function IndustryChart() {
  const {
    allocation,
    industryTargets,
    editingIndustry,
    activeIndustries,
    startEditIndustry,
    cancelEditIndustry,
    adjustIndustrySlider,
    saveIndustryTargets,
    loading,
  } = usePortfolioStore();
  const { showNotification } = useNotifications();

  /**
   * Handles saving industry targets with notification
   */
  const handleSave = async () => {
    try {
      await saveIndustryTargets();
      showNotification('Industry targets saved successfully', 'success');
    } catch (error) {
      showNotification(`Failed to save industry targets: ${error.message}`, 'error');
    }
  };

  const industryAllocations = allocation.industry || [];

  // Calculate normalized target percentages from weights
  const targets = useMemo(() => {
    return getTargetPcts(industryTargets, activeIndustries);
  }, [industryTargets, activeIndustries]);

  // Sort industries alphabetically for consistent display
  const sortedIndustries = useMemo(() => {
    return [...activeIndustries].sort();
  }, [activeIndustries]);

  return (
    <div className="industry-chart">
      <Group className="industry-chart__header" justify="space-between" mb="md">
        <Text className="industry-chart__title" size="xs" fw={500}>Industry Groups</Text>
        {!editingIndustry && (
          <Button
            className="industry-chart__edit-btn"
            size="xs"
            variant="subtle"
            color="violet"
            onClick={startEditIndustry}
          >
            Edit Weights
          </Button>
        )}
      </Group>

      {/* View Mode - Show deviation from target allocation */}
      {!editingIndustry && (
        <Stack className="industry-chart__view" gap="sm">
          {industryAllocations.length === 0 ? (
            <Text className="industry-chart__empty" size="sm" c="dimmed" ta="center" p="md">
              No industry allocation data available
            </Text>
          ) : (
            industryAllocations.map((industry) => {
            // Calculate deviation from target
            const deviation = getDeviation(industry.name, industry.current_pct, targets);
            const badgeClass = getDeviationBadgeClass(deviation);
            const barColor = getDeviationBarColor(deviation);
            const barStyle = getDeviationBarStyle(deviation);

            return (
              <div className="industry-chart__item" key={industry.name}>
                {/* Industry header with name, current %, and deviation badge */}
                <Group className="industry-chart__item-header" justify="space-between" mb="xs">
                  <Text className="industry-chart__item-name" size="sm" truncate style={{ maxWidth: '200px' }}>
                    {industry.name}
                  </Text>
                  <Group className="industry-chart__item-values" gap="xs" style={{ flexShrink: 0 }}>
                    {/* Current allocation percentage */}
                    <Text className="industry-chart__item-current" size="xs" style={{ fontFamily: 'var(--mantine-font-family)' }}>
                      {formatPercent(industry.current_pct)}
                    </Text>
                    {/* Deviation badge (red for over, blue for under, gray for neutral) */}
                    <Badge className="industry-chart__item-badge" size="xs" {...badgeClass} style={{ fontFamily: 'var(--mantine-font-family)' }}>
                      {formatDeviation(deviation)}
                    </Badge>
                  </Group>
                </Group>
                {/* Deviation bar - visual representation of over/under-allocation */}
                <div
                  className="industry-chart__deviation-bar"
                  style={{
                    height: '6px',
                    backgroundColor: 'var(--mantine-color-dark-6)',
                    borderRadius: '999px',
                    position: 'relative',
                    overflow: 'hidden',
                  }}
                >
                  {/* Center line (target position) */}
                  <div
                    className="industry-chart__deviation-center"
                    style={{
                      position: 'absolute',
                      top: 0,
                      bottom: 0,
                      left: '50%',
                      width: '1px',
                      backgroundColor: 'var(--mantine-color-dark-4)',
                      zIndex: 10,
                    }}
                  />
                  {/* Deviation fill - extends right for over-allocation, left for under-allocation */}
                  <div
                    className={`industry-chart__deviation-fill industry-chart__deviation-fill--${barColor}`}
                    style={{
                      position: 'absolute',
                      top: 0,
                      bottom: 0,
                      borderRadius: '999px',
                      backgroundColor: `var(--mantine-color-${barColor}-5)`,
                      ...barStyle,
                    }}
                  />
                </div>
              </div>
            );
          }))}
        </Stack>
      )}

      {/* Edit Mode - Weight sliders for active industries */}
      {editingIndustry && (
        <Stack className="industry-chart__edit" gap="md">
          {/* Weight Scale Legend - explains 0-1 scale meaning */}
          <Group className="industry-chart__legend" justify="space-between">
            <Text className="industry-chart__legend-avoid" size="xs" c="red">0 Avoid</Text>
            <Text className="industry-chart__legend-neutral" size="xs" c="dimmed">0.5 Neutral</Text>
            <Text className="industry-chart__legend-prioritize" size="xs" c="green">1 Prioritize</Text>
          </Group>

          <Divider className="industry-chart__divider" />

          {/* Dynamic Industry Sliders - one slider per active industry */}
          {sortedIndustries.length === 0 ? (
            <Text className="industry-chart__empty" size="sm" c="dimmed" ta="center" p="md">
              No active industries available
            </Text>
          ) : (
            sortedIndustries.map((name) => {
            const weight = industryTargets[name] || 0;
            const badgeClass = getWeightBadgeClass(weight);

            return (
              <div className="industry-chart__slider-item" key={name}>
                {/* Slider header with industry name and current weight */}
                <Group className="industry-chart__slider-header" justify="space-between" mb="xs">
                  <Text className="industry-chart__slider-name" size="sm" truncate style={{ maxWidth: '200px' }}>
                    {name}
                  </Text>
                  {/* Current weight badge */}
                  <Badge className="industry-chart__slider-badge" size="xs" {...badgeClass} style={{ flexShrink: 0, fontFamily: 'var(--mantine-font-family)' }}>
                    {formatWeight(weight)}
                  </Badge>
                </Group>
                {/* Weight slider (0-1 scale) */}
                <Slider
                  className="industry-chart__slider"
                  value={weight}
                  onChange={(val) => adjustIndustrySlider(name, val)}
                  min={0}
                  max={1}
                  step={0.01}
                  color="violet"
                  marks={[
                    { value: 0, label: '0' },
                    { value: 0.5, label: '0.5' },
                    { value: 1, label: '1' },
                  ]}
                />
              </div>
            );
          }))}

          <Divider className="industry-chart__divider" />

          {/* Action buttons */}
          <Group className="industry-chart__actions" grow>
            <Button
              className="industry-chart__cancel-btn"
              variant="subtle"
              onClick={cancelEditIndustry}
            >
              Cancel
            </Button>
            <Button
              className="industry-chart__save-btn"
              color="violet"
              onClick={handleSave}
              disabled={loading.industrySave}
              loading={loading.industrySave}
            >
              Save
            </Button>
          </Group>
        </Stack>
      )}
    </div>
  );
}
