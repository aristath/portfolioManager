import { Card, Group, Text, Badge, Stack, Alert, Divider } from '@mantine/core';
import { AllocationRadar } from './AllocationRadar';
import { GeoChart } from './GeoChart';
import { usePortfolioStore } from '../../stores/portfolioStore';
import { formatPercent } from '../../utils/formatters';

export function GeographyRadarCard() {
  const { alerts } = usePortfolioStore();

  const geographyAlerts = alerts.filter(a => a.type === 'geography');
  const hasCritical = geographyAlerts.some(a => a.severity === 'critical');

  return (
    <Card className="geo-radar-card" p="md">
      <Group className="geo-radar-card__header" justify="space-between" mb="md">
        <Text className="geo-radar-card__title" size="xs" tt="uppercase" c="dimmed" fw={600}>
          Geography Allocation
        </Text>
        {geographyAlerts.length > 0 && (
          <Badge
            className="geo-radar-card__alert-badge"
            size="sm"
            color={hasCritical ? 'red' : 'yellow'}
            variant="light"
          >
            {geographyAlerts.length} alert{geographyAlerts.length > 1 ? 's' : ''}
          </Badge>
        )}
      </Group>

      <AllocationRadar type="geography" />

      <Divider className="geo-radar-card__divider" my="md" />

      <GeoChart />

      {/* Geography Alerts */}
      {geographyAlerts.length > 0 && (
        <Stack className="geo-radar-card__alerts" gap="xs" mt="md" pt="md" style={{ borderTop: '1px solid var(--mantine-color-dark-6)' }}>
          {geographyAlerts.map((alert) => (
            <Alert
              className={`geo-radar-card__alert geo-radar-card__alert--${alert.severity}`}
              key={alert.name}
              color={alert.severity === 'critical' ? 'red' : 'yellow'}
              variant="light"
              title={
                <Group className="geo-radar-card__alert-title" justify="space-between" style={{ width: '100%' }}>
                  <Group className="geo-radar-card__alert-name" gap="xs">
                    <Text className="geo-radar-card__alert-icon" size="xs">{alert.severity === 'critical' ? '🔴' : '⚠️'}</Text>
                    <Text className="geo-radar-card__alert-label" size="sm" fw={500}>
                      {alert.name}
                    </Text>
                  </Group>
                  <Group className="geo-radar-card__alert-values" gap="xs">
                    <Text className="geo-radar-card__alert-current" size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }} fw={600}>
                      {formatPercent(alert.current_pct)}
                    </Text>
                    <Text className="geo-radar-card__alert-limit" size="xs" c="dimmed">
                      Limit: {formatPercent(alert.limit_pct, 0)}
                    </Text>
                  </Group>
                </Group>
              }
            />
          ))}
        </Stack>
      )}
    </Card>
  );
}
