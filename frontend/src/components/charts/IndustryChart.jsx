/**
 * Industry Chart Component
 *
 * Provides weight editing interface for industry allocation targets.
 *
 * Features:
 * - Weight sliders for adjusting target allocation (0-1 scale)
 * - Weight normalization (weights converted to percentages)
 * - Edit/Cancel/Save workflow
 * - Success/error notifications
 *
 * Used in IndustryRadarCard for industry allocation management.
 */
import { useMemo } from 'react';
import { Group, Text, Button, Slider, Badge, Stack, Divider } from '@mantine/core';
import { usePortfolioStore } from '../../stores/portfolioStore';
import { useNotifications } from '../../hooks/useNotifications';

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
 * Provides weight editing interface for industry allocation.
 *
 * @returns {JSX.Element} Industry weight editing interface
 */
export function IndustryChart() {
  const {
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
