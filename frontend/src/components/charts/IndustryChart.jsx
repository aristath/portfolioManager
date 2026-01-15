import { useMemo } from 'react';
import { Group, Text, Button, Slider, Badge, Stack, Divider } from '@mantine/core';
import { usePortfolioStore } from '../../stores/portfolioStore';
import { useNotifications } from '../../hooks/useNotifications';
import { formatPercent } from '../../utils/formatters';

// Convert weights to target percentages
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

  const handleSave = async () => {
    try {
      await saveIndustryTargets();
      showNotification('Industry targets saved successfully', 'success');
    } catch (error) {
      showNotification(`Failed to save industry targets: ${error.message}`, 'error');
    }
  };

  const industryAllocations = allocation.industry || [];

  const targets = useMemo(() => {
    return getTargetPcts(industryTargets, activeIndustries);
  }, [industryTargets, activeIndustries]);

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
            const deviation = getDeviation(industry.name, industry.current_pct, targets);
            const badgeClass = getDeviationBadgeClass(deviation);
            const barColor = getDeviationBarColor(deviation);
            const barStyle = getDeviationBarStyle(deviation);

            return (
              <div className="industry-chart__item" key={industry.name}>
                <Group className="industry-chart__item-header" justify="space-between" mb="xs">
                  <Text className="industry-chart__item-name" size="sm" truncate style={{ maxWidth: '200px' }}>
                    {industry.name}
                  </Text>
                  <Group className="industry-chart__item-values" gap="xs" style={{ flexShrink: 0 }}>
                    <Text className="industry-chart__item-current" size="xs" style={{ fontFamily: 'var(--mantine-font-family)' }}>
                      {formatPercent(industry.current_pct)}
                    </Text>
                    <Badge className="industry-chart__item-badge" size="xs" {...badgeClass} style={{ fontFamily: 'var(--mantine-font-family)' }}>
                      {formatDeviation(deviation)}
                    </Badge>
                  </Group>
                </Group>
                {/* Deviation bar */}
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
          {/* Weight Scale Legend */}
          <Group className="industry-chart__legend" justify="space-between">
            <Text className="industry-chart__legend-avoid" size="xs" c="red">0 Avoid</Text>
            <Text className="industry-chart__legend-neutral" size="xs" c="dimmed">0.5 Neutral</Text>
            <Text className="industry-chart__legend-prioritize" size="xs" c="green">1 Prioritize</Text>
          </Group>

          <Divider className="industry-chart__divider" />

          {/* Dynamic Industry Sliders */}
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
                <Group className="industry-chart__slider-header" justify="space-between" mb="xs">
                  <Text className="industry-chart__slider-name" size="sm" truncate style={{ maxWidth: '200px' }}>
                    {name}
                  </Text>
                  <Badge className="industry-chart__slider-badge" size="xs" {...badgeClass} style={{ flexShrink: 0, fontFamily: 'var(--mantine-font-family)' }}>
                    {formatWeight(weight)}
                  </Badge>
                </Group>
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

          {/* Buttons */}
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
