/**
 * Geo Chart Component
 * 
 * Displays geographic allocation with deviation visualization and weight editing.
 * Shows current allocation vs target allocation with visual deviation bars.
 * 
 * Features:
 * - View mode: Shows current allocation and deviation from target
 * - Edit mode: Weight sliders for adjusting target allocation (0-1 scale)
 * - Deviation bars: Visual representation of over/under-allocation
 * - Color-coded geography indicators
 * - Weight normalization (weights converted to percentages)
 * 
 * Used in GeographyRadarCard for detailed geography allocation management.
 */
import { useMemo } from 'react';
import { Group, Text, Button, Slider, Badge, Stack, Divider } from '@mantine/core';
import { usePortfolioStore } from '../../stores/portfolioStore';
import { formatPercent } from '../../utils/formatters';

/**
 * Generates a consistent color for a geography name using hash-based HSL
 * 
 * @param {string} name - Geography name
 * @returns {string} HSL color string
 */
function getGeographyColor(name) {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  const hue = Math.abs(hash) % 360;
  return `hsl(${hue}, 70%, 50%)`;
}

/**
 * Converts allocation weights to target percentages
 * 
 * Normalizes weights so they sum to 100% for display purposes.
 * 
 * @param {Object} weights - Object mapping geography names to weights
 * @param {Array<string>} activeGeographies - Array of active geography names
 * @returns {Object} Object mapping geography names to normalized percentages (0-1)
 */
function getTargetPcts(weights, activeGeographies) {
  let total = 0;
  for (const name of activeGeographies) {
    const weight = weights[name] || 0;
    total += weight;
  }

  const targets = {};
  for (const name of activeGeographies) {
    const weight = weights[name] || 0;
    targets[name] = total > 0 ? weight / total : 0;
  }
  return targets;
}

/**
 * Calculates deviation from target: current% - target%
 * 
 * @param {string} name - Geography name
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
 * Geo chart component
 * 
 * Displays geographic allocation with view and edit modes.
 * 
 * @returns {JSX.Element} Geography allocation chart with deviation visualization
 */
export function GeoChart() {
  const {
    allocation,
    geographyTargets,
    editingGeography,
    activeGeographies,
    startEditGeography,
    cancelEditGeography,
    adjustGeographySlider,
    saveGeographyTargets,
    loading,
  } = usePortfolioStore();
  const geographyAllocations = allocation.geography || [];

  // Calculate normalized target percentages from weights
  const targets = useMemo(() => {
    return getTargetPcts(geographyTargets, activeGeographies);
  }, [geographyTargets, activeGeographies]);

  // Sort geographies alphabetically for consistent display
  const sortedGeographies = useMemo(() => {
    return [...activeGeographies].sort();
  }, [activeGeographies]);

  return (
    <div className="geo-chart">
      <Group className="geo-chart__header" justify="space-between" mb="md">
        <Text className="geo-chart__title" size="xs" fw={500}>Geography Allocation</Text>
        {!editingGeography && (
          <Button
            className="geo-chart__edit-btn"
            size="xs"
            variant="subtle"
            onClick={startEditGeography}
          >
            Edit Weights
          </Button>
        )}
      </Group>

      {/* View Mode - Show deviation from target allocation */}
      {!editingGeography && (
        <Stack className="geo-chart__view" gap="sm">
          {geographyAllocations.length === 0 ? (
            <Text className="geo-chart__empty" size="sm" c="dimmed" ta="center" p="md">
              No geography allocation data available
            </Text>
          ) : (
            geographyAllocations.map((geography) => {
            // Calculate deviation from target
            const deviation = getDeviation(geography.name, geography.current_pct, targets);
            const badgeClass = getDeviationBadgeClass(deviation);
            const barColor = getDeviationBarColor(deviation);
            const barStyle = getDeviationBarStyle(deviation);

            return (
              <div className="geo-chart__item" key={geography.name}>
                {/* Geography header with name, current %, and deviation badge */}
                <Group className="geo-chart__item-header" justify="space-between" mb="xs">
                  <Group className="geo-chart__item-label" gap="xs">
                    {/* Color-coded geography indicator */}
                    <div
                      className="geo-chart__item-dot"
                      style={{
                        width: '10px',
                        height: '10px',
                        borderRadius: '50%',
                        backgroundColor: getGeographyColor(geography.name),
                      }}
                    />
                    <Text className="geo-chart__item-name" size="sm">{geography.name}</Text>
                  </Group>
                  <Group className="geo-chart__item-values" gap="xs">
                    {/* Current allocation percentage */}
                    <Text className="geo-chart__item-current" size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }}>
                      {formatPercent(geography.current_pct)}
                    </Text>
                    {/* Deviation badge (red for over, blue for under, gray for neutral) */}
                    <Badge className="geo-chart__item-badge" size="xs" {...badgeClass} style={{ fontFamily: 'var(--mantine-font-family)' }}>
                      {formatDeviation(deviation)}
                    </Badge>
                  </Group>
                </Group>
                {/* Deviation bar - visual representation of over/under-allocation */}
                <div
                  className="geo-chart__deviation-bar"
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
                    className="geo-chart__deviation-center"
                    style={{
                      position: 'absolute',
                      top: 0,
                      bottom: 0,
                      left: '50%',
                      width: '1px',
                      backgroundColor: 'var(--mantine-color-dark-5)',
                      zIndex: 10,
                    }}
                  />
                  {/* Deviation fill - extends right for over-allocation, left for under-allocation */}
                  <div
                    className={`geo-chart__deviation-fill geo-chart__deviation-fill--${barColor}`}
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

      {/* Edit Mode - Weight sliders for active geographies */}
      {editingGeography && (
        <Stack className="geo-chart__edit" gap="md">
          {/* Weight Scale Legend - explains 0-1 scale meaning */}
          <Group className="geo-chart__legend" justify="space-between">
            <Text className="geo-chart__legend-avoid" size="xs" c="red">0 Avoid</Text>
            <Text className="geo-chart__legend-neutral" size="xs" c="dimmed">0.5 Neutral</Text>
            <Text className="geo-chart__legend-prioritize" size="xs" c="green">1 Prioritize</Text>
          </Group>

          <Divider className="geo-chart__divider" />

          {/* Dynamic Geography Sliders - one slider per active geography */}
          {sortedGeographies.length === 0 ? (
            <Text className="geo-chart__empty" size="sm" c="dimmed" ta="center" p="md">
              No active geographies available
            </Text>
          ) : (
            sortedGeographies.map((name) => {
            const weight = geographyTargets[name] || 0;
            const badgeClass = getWeightBadgeClass(weight);

            return (
              <div className="geo-chart__slider-item" key={name}>
                {/* Slider header with geography name and current weight */}
                <Group className="geo-chart__slider-header" justify="space-between" mb="xs">
                  <Group className="geo-chart__slider-label" gap="xs">
                    {/* Color-coded geography indicator */}
                    <div
                      className="geo-chart__slider-dot"
                      style={{
                        width: '10px',
                        height: '10px',
                        borderRadius: '50%',
                        backgroundColor: getGeographyColor(name),
                      }}
                    />
                    <Text className="geo-chart__slider-name" size="sm">{name}</Text>
                  </Group>
                  {/* Current weight badge */}
                  <Badge className="geo-chart__slider-badge" size="xs" {...badgeClass} style={{ fontFamily: 'var(--mantine-font-family)' }}>
                    {formatWeight(weight)}
                  </Badge>
                </Group>
                {/* Weight slider (0-1 scale) */}
                <Slider
                  className="geo-chart__slider"
                  value={weight}
                  onChange={(val) => adjustGeographySlider(name, val)}
                  min={0}
                  max={1}
                  step={0.01}
                  marks={[
                    { value: 0, label: '0' },
                    { value: 0.5, label: '0.5' },
                    { value: 1, label: '1' },
                  ]}
                />
              </div>
            );
          }))}

          <Divider className="geo-chart__divider" />

          {/* Action buttons */}
          <Group className="geo-chart__actions" grow>
            <Button
              className="geo-chart__cancel-btn"
              variant="subtle"
              onClick={cancelEditGeography}
            >
              Cancel
            </Button>
            <Button
              className="geo-chart__save-btn"
              onClick={saveGeographyTargets}
              disabled={loading.geographySave}
              loading={loading.geographySave}
            >
              Save
            </Button>
          </Group>
        </Stack>
      )}
    </div>
  );
}
