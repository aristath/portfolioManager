import { Alert, Stack, Text } from '@mantine/core';
import { usePortfolioStore } from '../../stores/portfolioStore';
import { formatPercent } from '../../utils/formatters';

export function ConcentrationAlerts() {
  const { alerts } = usePortfolioStore();

  const criticalAlerts = alerts.filter(a => a.severity === 'critical');
  const warningAlerts = alerts.filter(a => a.severity === 'warning');

  if (alerts.length === 0) {
    return null;
  }

  return (
    <Stack className="concentration-alerts" gap="xs">
      {criticalAlerts.length > 0 && (
        <Alert className="concentration-alerts__critical" color="red" variant="light" title={`${criticalAlerts.length} Critical Alert${criticalAlerts.length > 1 ? 's' : ''}`}>
          <Stack className="concentration-alerts__critical-list" gap="xs">
            {criticalAlerts.map((alert) => (
              <div className="concentration-alerts__item concentration-alerts__item--critical" key={alert.name} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Text className="concentration-alerts__item-name" size="sm" fw={500}>
                  {alert.name} ({alert.type})
                </Text>
                <Text className="concentration-alerts__item-value" size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }} fw={600}>
                  {formatPercent(alert.current_pct)} / {formatPercent(alert.limit_pct, 0)} limit
                </Text>
              </div>
            ))}
          </Stack>
        </Alert>
      )}

      {warningAlerts.length > 0 && (
        <Alert className="concentration-alerts__warning" color="yellow" variant="light" title={`${warningAlerts.length} Warning${warningAlerts.length > 1 ? 's' : ''}`}>
          <Stack className="concentration-alerts__warning-list" gap="xs">
            {warningAlerts.map((alert) => (
              <div className="concentration-alerts__item concentration-alerts__item--warning" key={alert.name} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Text className="concentration-alerts__item-name" size="sm" fw={500}>
                  {alert.name} ({alert.type})
                </Text>
                <Text className="concentration-alerts__item-value" size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }} fw={600}>
                  {formatPercent(alert.current_pct)} / {formatPercent(alert.limit_pct, 0)} limit
                </Text>
              </div>
            ))}
          </Stack>
        </Alert>
      )}
    </Stack>
  );
}
