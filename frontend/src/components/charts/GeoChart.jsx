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
    <div>
      <Group justify="space-between" mb="md">
        <Text size="xs" fw={500}>Geography Allocation</Text>
        {!editingGeography && (
          <Button
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
        <Stack gap="sm">
          {geographyAllocations.length === 0 ? (
            <Text size="sm" c="dimmed" ta="center" p="md">
              No geography allocation data available
            </Text>
          ) : (
            geographyAllocations.map((geography) => {
            const deviation = getDeviation(geography.name, geography.current_pct, targets);
            const badgeClass = getDeviationBadgeClass(deviation);
            const barColor = getDeviationBarColor(deviation);
            const barStyle = getDeviationBarStyle(deviation);

            return (
              <div key={geography.name}>
                <Group justify="space-between" mb="xs">
                  <Group gap="xs">
                    <div
                      style={{
                        width: '10px',
                        height: '10px',
                        borderRadius: '50%',
                        backgroundColor: getGeographyColor(geography.name),
                      }}
                    />
                    <Text size="sm">{geography.name}</Text>
                  </Group>
                  <Group gap="xs">
                    <Text size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }}>
                      {formatPercent(geography.current_pct)}
                    </Text>
                    <Badge size="xs" {...badgeClass} style={{ fontFamily: 'var(--mantine-font-family)' }}>
                      {formatDeviation(deviation)}
                    </Badge>
                  </Group>
                </Group>
                {/* Deviation bar */}
                <div
                  style={{
                    height: '6px',
                    backgroundColor: 'var(--mantine-color-dark-6)',
                    borderRadius: '999px',
                    position: 'relative',
                    overflow: 'hidden',
                  }}
                >
                  <div
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
        <Stack gap="md">
          {/* Weight Scale Legend */}
          <Group justify="space-between">
            <Text size="xs" c="red">0 Avoid</Text>
            <Text size="xs" c="dimmed">0.5 Neutral</Text>
            <Text size="xs" c="green">1 Prioritize</Text>
          </Group>

          <Divider />

          {/* Dynamic Geography Sliders */}
          {sortedGeographies.length === 0 ? (
            <Text size="sm" c="dimmed" ta="center" p="md">
              No active geographies available
            </Text>
          ) : (
            sortedGeographies.map((name) => {
            const weight = geographyTargets[name] || 0;
            const badgeClass = getWeightBadgeClass(weight);

            return (
              <div key={name}>
                <Group justify="space-between" mb="xs">
                  <Group gap="xs">
                    <div
                      style={{
                        width: '10px',
                        height: '10px',
                        borderRadius: '50%',
                        backgroundColor: getGeographyColor(name),
                      }}
                    />
                    <Text size="sm">{name}</Text>
                  </Group>
                  <Badge size="xs" {...badgeClass} style={{ fontFamily: 'var(--mantine-font-family)' }}>
                    {formatWeight(weight)}
                  </Badge>
                </Group>
                <Slider
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

          <Divider />

          {/* Buttons */}
          <Group grow>
            <Button
              variant="subtle"
              onClick={cancelEditGeography}
            >
              Cancel
            </Button>
            <Button
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
