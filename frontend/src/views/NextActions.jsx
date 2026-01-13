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
    <Stack gap="md">
      {/* Actions */}
      <Group justify="space-between" wrap="wrap">
        {/* Research mode testing controls */}
        {isResearchMode && (
          <Tooltip label="Disable cooloff period checks for testing (only works in research mode)">
            <Group gap="xs">
              <Switch
                checked={isCooloffDisabled}
                onChange={handleCooloffToggle}
                size="sm"
                color="orange"
              />
              <Text size="sm" c="dimmed">Disable Cooloff</Text>
            </Group>
          </Tooltip>
        )}

        {!isResearchMode && <div />}

        <Button
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
