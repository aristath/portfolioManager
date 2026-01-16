import React, { useState } from 'react';
import { ActionIcon, Popover, Slider, Text, Group, Tooltip, Stack, Button } from '@mantine/core';
import {
  IconBan,
  IconTrendingDown,
  IconMinus,
  IconTrendingUp,
  IconStar,
  IconTarget,
  IconRefresh
} from '@tabler/icons-react';
import { useSecuritiesStore } from '../../stores/securitiesStore';

// Icon thresholds for visual feedback
const ICON_THRESHOLDS = [
  { min: 0.2, max: 0.4, icon: IconBan, label: 'Awful', color: 'red' },
  { min: 0.5, max: 0.8, icon: IconTrendingDown, label: 'Bad', color: 'orange' },
  { min: 0.9, max: 1.1, icon: IconMinus, label: 'Normal', color: 'gray' },
  { min: 1.2, max: 1.7, icon: IconTrendingUp, label: 'Good', color: 'green' },
  { min: 1.8, max: 2.3, icon: IconStar, label: 'Excellent', color: 'blue' },
  { min: 2.4, max: 3.0, icon: IconTarget, label: 'Strategic', color: 'violet' }
];

function getIconForMultiplier(multiplier) {
  const threshold = ICON_THRESHOLDS.find(
    t => multiplier >= t.min && multiplier <= t.max
  );
  return threshold || ICON_THRESHOLDS[2]; // Default to Normal
}

export function RatingIcon({ isin, currentMultiplier = 1.0 }) {
  const [opened, setOpened] = useState(false);
  const [sliderValue, setSliderValue] = useState(currentMultiplier);
  const [loading, setLoading] = useState(false);
  const setMultiplier = useSecuritiesStore(state => state.setMultiplier);

  const currentIcon = getIconForMultiplier(currentMultiplier);
  const sliderIcon = getIconForMultiplier(sliderValue);
  const Icon = currentIcon.icon;

  const isCustomized = Math.abs(currentMultiplier - 1.0) > 0.01;

  const handleSliderChange = (value) => {
    setSliderValue(value);
  };

  const handleSave = async () => {
    setLoading(true);
    try {
      await setMultiplier(isin, sliderValue);
      setOpened(false);
    } catch (error) {
      console.error('Failed to update multiplier:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleReset = async () => {
    setLoading(true);
    try {
      await setMultiplier(isin, 1.0);
      setSliderValue(1.0);
      setOpened(false);
    } catch (error) {
      console.error('Failed to reset multiplier:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleOpen = () => {
    setSliderValue(currentMultiplier);
    setOpened(true);
  };

  return (
    <Popover
      width={280}
      position="bottom"
      withArrow
      shadow="md"
      opened={opened}
      onChange={setOpened}
    >
      <Popover.Target>
        <Tooltip label={`${currentIcon.label} (${currentMultiplier.toFixed(2)}x)`}>
          <ActionIcon
            variant={isCustomized ? 'light' : 'subtle'}
            color={isCustomized ? currentIcon.color : 'gray'}
            size="lg"
            onClick={handleOpen}
          >
            <Icon size={20} />
          </ActionIcon>
        </Tooltip>
      </Popover.Target>

      <Popover.Dropdown>
        <Stack gap="md">
          <div>
            <Group justify="space-between" mb="xs">
              <Text size="sm" fw={500}>Priority Multiplier</Text>
              <Group gap={4}>
                <Text size="sm" c={sliderIcon.color} fw={600}>
                  {sliderValue.toFixed(2)}x
                </Text>
                <Text size="xs" c="dimmed">
                  ({sliderIcon.label})
                </Text>
              </Group>
            </Group>

            <Slider
              value={sliderValue}
              onChange={handleSliderChange}
              min={0.2}
              max={3.0}
              step={0.1}
              marks={[
                { value: 0.2, label: '0.2' },
                { value: 1.0, label: '1.0' },
                { value: 3.0, label: '3.0' }
              ]}
              color={sliderIcon.color}
              size="md"
            />
          </div>

          <Group justify="space-between">
            <Button
              variant="subtle"
              color="gray"
              size="xs"
              leftSection={<IconRefresh size={16} />}
              onClick={handleReset}
              loading={loading}
            >
              Reset to 1.0
            </Button>

            <Group gap="xs">
              <Button
                variant="subtle"
                size="xs"
                onClick={() => setOpened(false)}
              >
                Cancel
              </Button>
              <Button
                size="xs"
                onClick={handleSave}
                loading={loading}
              >
                Save
              </Button>
            </Group>
          </Group>
        </Stack>
      </Popover.Dropdown>
    </Popover>
  );
}
