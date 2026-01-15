import { useMemo } from 'react';
import { Group, Text, Button, Slider, Badge, Stack, Divider } from '@mantine/core';
import { usePortfolioStore } from '../../stores/portfolioStore';
import { formatPercent } from '../../utils/formatters';

// Generate consistent color for geography name
function getGeographyColor(name) {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  const hue = Math.abs(hash) % 360;
  return `hsl(${hue}, 70%, 50%)`;
}

// Convert weights to target percentages
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

// Calculate deviation: current% - target%
function getDeviation(name, currentPct, targets) {
  const targetPct = targets[name] || 0;
  return currentPct - targetPct;
}

function formatDeviation(deviation) {
  const pct = (deviation * 100).toFixed(1);
  return (deviation >= 0 ? '+' : '') + pct + '%';
}

function getDeviationBadgeClass(deviation) {
  if (Math.abs(deviation) < 0.02) return { color: 'gray', variant: 'light' };
  return deviation > 0
    ? { color: 'red', variant: 'light' }
    : { color: 'blue', variant: 'light' };
}

function getDeviationBarColor(deviation) {
  if (Math.abs(deviation) < 0.02) return 'gray';
  return deviation > 0 ? 'red' : 'blue';
}

function getDeviationBarStyle(deviation) {
  const maxDev = 0.50;
  const pct = Math.min(Math.abs(deviation), maxDev) / maxDev * 50;

  if (deviation >= 0) {
    return { width: `${pct}%`, left: '50%', right: 'auto' };
  } else {
    return { width: `${pct}%`, right: '50%', left: 'auto' };
  }
}

function formatWeight(weight) {
  if (weight === 0 || weight === undefined) return '0';
  return weight.toFixed(2);
}

function getWeightBadgeClass(weight) {
  if (weight > 0.7) return { color: 'green', variant: 'light' };
  if (weight < 0.3) return { color: 'red', variant: 'light' };
  return { color: 'gray', variant: 'light' };
}

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

  const targets = useMemo(() => {
    return getTargetPcts(geographyTargets, activeGeographies);
  }, [geographyTargets, activeGeographies]);

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
            const deviation = getDeviation(geography.name, geography.current_pct, targets);
            const badgeClass = getDeviationBadgeClass(deviation);
            const barColor = getDeviationBarColor(deviation);
            const barStyle = getDeviationBarStyle(deviation);

            return (
              <div className="geo-chart__item" key={geography.name}>
                <Group className="geo-chart__item-header" justify="space-between" mb="xs">
                  <Group className="geo-chart__item-label" gap="xs">
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
                    <Text className="geo-chart__item-current" size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }}>
                      {formatPercent(geography.current_pct)}
                    </Text>
                    <Badge className="geo-chart__item-badge" size="xs" {...badgeClass} style={{ fontFamily: 'var(--mantine-font-family)' }}>
                      {formatDeviation(deviation)}
                    </Badge>
                  </Group>
                </Group>
                {/* Deviation bar */}
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
          {/* Weight Scale Legend */}
          <Group className="geo-chart__legend" justify="space-between">
            <Text className="geo-chart__legend-avoid" size="xs" c="red">0 Avoid</Text>
            <Text className="geo-chart__legend-neutral" size="xs" c="dimmed">0.5 Neutral</Text>
            <Text className="geo-chart__legend-prioritize" size="xs" c="green">1 Prioritize</Text>
          </Group>

          <Divider className="geo-chart__divider" />

          {/* Dynamic Geography Sliders */}
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
                <Group className="geo-chart__slider-header" justify="space-between" mb="xs">
                  <Group className="geo-chart__slider-label" gap="xs">
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
                  <Badge className="geo-chart__slider-badge" size="xs" {...badgeClass} style={{ fontFamily: 'var(--mantine-font-family)' }}>
                    {formatWeight(weight)}
                  </Badge>
                </Group>
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

          {/* Buttons */}
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
