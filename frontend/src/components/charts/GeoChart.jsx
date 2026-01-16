/**
 * Geo Chart Component
 *
 * Provides weight editing interface for geographic allocation targets.
 *
 * Features:
 * - Weight sliders for adjusting target allocation (0-1 scale)
 * - Weight normalization (weights converted to percentages)
 * - Edit/Cancel/Save workflow
 *
 * Used in GeographyRadarCard for geography allocation management.
 */
import { useMemo } from 'react';
import { Group, Text, Button, Slider, Badge, Stack, Divider } from '@mantine/core';
import { usePortfolioStore } from '../../stores/portfolioStore';

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
 * Provides weight editing interface for geographic allocation.
 *
 * @returns {JSX.Element} Geography weight editing interface
 */
export function GeoChart() {
  const {
    geographyTargets,
    editingGeography,
    activeGeographies,
    startEditGeography,
    cancelEditGeography,
    adjustGeographySlider,
    saveGeographyTargets,
    loading,
  } = usePortfolioStore();

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
