/**
 * Geography Radar Card Component
 * 
 * Displays geographic allocation of the portfolio using radar chart visualization.
 * Shows allocation by geographic region (US, EU, ASIA, etc.) with alerts for over-concentration.
 * 
 * Features:
 * - Radar chart showing geographic allocation
 * - Geographic chart visualization
 * - Concentration alerts (critical/warning)
 * - Alert badge in header
 * - Current percentage vs limit display
 * 
 * Used in the Diversification view to monitor geographic diversification.
 */
import { Card, Group, Text, Badge, Stack, Alert, Divider } from '@mantine/core';
import { AllocationRadar } from './AllocationRadar';
import { GeoChart } from './GeoChart';
import { usePortfolioStore } from '../../stores/portfolioStore';
import { formatPercent } from '../../utils/formatters';

/**
 * Geography radar card component
 * 
 * Displays geographic allocation with radar chart and alerts.
 * 
 * @returns {JSX.Element} Geography radar card with chart and alerts
 */
export function GeographyRadarCard() {
  const { alerts } = usePortfolioStore();

  // Filter alerts for geography type
  const geographyAlerts = alerts.filter(a => a.type === 'geography');
  const hasCritical = geographyAlerts.some(a => a.severity === 'critical');

  return (
    <Card className="geo-radar-card" p="md">
      {/* Header with title and alert badge */}
      <Group className="geo-radar-card__header" justify="space-between" mb="md">
        <Text className="geo-radar-card__title" size="xs" tt="uppercase" c="dimmed" fw={600}>
          Geography Allocation
        </Text>
        {/* Alert badge - red for critical, yellow for warnings */}
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

      {/* Radar chart showing geographic allocation */}
      <AllocationRadar type="geography" />

      <Divider className="geo-radar-card__divider" my="md" />

      {/* Geographic chart visualization */}
      <GeoChart />

      {/* Geography Alerts - shown below charts */}
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
                  {/* Alert name with icon */}
                  <Group className="geo-radar-card__alert-name" gap="xs">
                    <Text className="geo-radar-card__alert-icon" size="xs">{alert.severity === 'critical' ? '🔴' : '⚠️'}</Text>
                    <Text className="geo-radar-card__alert-label" size="sm" fw={500}>
                      {alert.name}
                    </Text>
                  </Group>
                  {/* Current percentage and limit */}
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
