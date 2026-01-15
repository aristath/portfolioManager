import { Group, Button, Stack, Switch, Text, Tooltip } from '@mantine/core';
import { NextActionsCard } from '../components/portfolio/NextActionsCard';
import { useAppStore } from '../stores/appStore';
import { useSettingsStore } from '../stores/settingsStore';

export function NextActions() {
  const { openPlannerManagementModal } = useAppStore();
  const { settings, tradingMode, updateSetting } = useSettingsStore();

  const isResearchMode = tradingMode === 'research';
  const isCooloffDisabled = settings.disable_cooloff_checks >= 1.0;

  const handleCooloffToggle = async (event) => {
    const newValue = event.currentTarget.checked ? 1.0 : 0.0;
    try {
      await updateSetting('disable_cooloff_checks', newValue);
    } catch (error) {
      console.error('Failed to toggle cooloff checks:', error);
    }
  };

  return (
    <Stack className="next-actions-view" gap="md">
      {/* Actions */}
      <Group className="next-actions-view__actions" justify="space-between" wrap="wrap">
        {/* Research mode testing controls */}
        {isResearchMode && (
          <Tooltip label="Disable cooloff period checks for testing (only works in research mode)">
            <Group className="next-actions-view__cooloff-toggle" gap="xs">
              <Switch
                className="next-actions-view__cooloff-switch"
                checked={isCooloffDisabled}
                onChange={handleCooloffToggle}
                size="sm"
                color="orange"
              />
              <Text className="next-actions-view__cooloff-label" size="sm" c="dimmed">Disable Cooloff</Text>
            </Group>
          </Tooltip>
        )}

        {!isResearchMode && <div className="next-actions-view__spacer" />}

        <Button
          className="next-actions-view__configure-btn"
          variant="light"
          color="green"
          size="sm"
          onClick={() => openPlannerManagementModal()}
        >
          ⚙️ Configure Planner
        </Button>
      </Group>

      <NextActionsCard />
    </Stack>
  );
}
