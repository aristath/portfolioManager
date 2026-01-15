import { Card, Group, Text, Badge, Stack, Alert, Divider } from '@mantine/core';
import { AllocationRadar } from './AllocationRadar';
import { IndustryChart } from './IndustryChart';
import { usePortfolioStore } from '../../stores/portfolioStore';
import { formatPercent } from '../../utils/formatters';

export function IndustryRadarCard() {
  const { alerts } = usePortfolioStore();

  const industryAlerts = alerts.filter(a => a.type === 'sector');
  const hasCritical = industryAlerts.some(a => a.severity === 'critical');

  return (
    <Card className="industry-radar-card" p="md">
      <Group className="industry-radar-card__header" justify="space-between" mb="md">
        <Text className="industry-radar-card__title" size="xs" tt="uppercase" c="dimmed" fw={600}>
          Industry Allocation
        </Text>
        {industryAlerts.length > 0 && (
          <Badge
            className="industry-radar-card__alert-badge"
            size="sm"
            color={hasCritical ? 'red' : 'yellow'}
            variant="light"
          >
            {industryAlerts.length} alert{industryAlerts.length > 1 ? 's' : ''}
          </Badge>
        )}
      </Group>

      <AllocationRadar type="industry" />

      <Divider className="industry-radar-card__divider" my="md" />

      <IndustryChart />

      {/* Industry Alerts */}
      {industryAlerts.length > 0 && (
        <Stack className="industry-radar-card__alerts" gap="xs" mt="md" pt="md" style={{ borderTop: '1px solid var(--mantine-color-dark-6)' }}>
          {industryAlerts.map((alert) => (
            <Alert
              className={`industry-radar-card__alert industry-radar-card__alert--${alert.severity}`}
              key={alert.name}
              color={alert.severity === 'critical' ? 'red' : 'yellow'}
              variant="light"
              title={
                <Group className="industry-radar-card__alert-title" justify="space-between" style={{ width: '100%' }}>
                  <Group className="industry-radar-card__alert-name" gap="xs">
                    <Text className="industry-radar-card__alert-icon" size="xs">{alert.severity === 'critical' ? '🔴' : '⚠️'}</Text>
                    <Text className="industry-radar-card__alert-label" size="sm" fw={500} truncate style={{ maxWidth: '200px' }}>
                      {alert.name}
                    </Text>
                  </Group>
                  <Group className="industry-radar-card__alert-values" gap="xs" style={{ flexShrink: 0 }}>
                    <Text className="industry-radar-card__alert-current" size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }} fw={600}>
                      {formatPercent(alert.current_pct)}
                    </Text>
                    <Text className="industry-radar-card__alert-limit" size="xs" c="dimmed">
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
