/**
 * Concentration Alerts Component
 * 
 * Displays portfolio concentration warnings and critical alerts.
 * Shows when portfolio allocation exceeds limits for geography, industry, or individual securities.
 * 
 * Features:
 * - Critical alerts (red) - severe concentration issues
 * - Warning alerts (yellow) - approaching concentration limits
 * - Shows current percentage vs limit for each alert
 * - Grouped by severity for easy scanning
 */
import { Alert, Stack, Text } from '@mantine/core';
import { usePortfolioStore } from '../../stores/portfolioStore';
import { formatPercent } from '../../utils/formatters';

/**
 * Concentration alerts component
 * 
 * Displays portfolio concentration warnings grouped by severity.
 * Returns null if no alerts are present.
 * 
 * @returns {JSX.Element|null} Concentration alerts or null if no alerts
 */
export function ConcentrationAlerts() {
  const { alerts } = usePortfolioStore();

  // Separate alerts by severity
  const criticalAlerts = alerts.filter(a => a.severity === 'critical');
  const warningAlerts = alerts.filter(a => a.severity === 'warning');

  // Don't render if no alerts
  if (alerts.length === 0) {
    return null;
  }

  return (
    <Stack className="concentration-alerts" gap="xs">
      {/* Critical alerts - red, shown first */}
      {criticalAlerts.length > 0 && (
        <Alert className="concentration-alerts__critical" color="red" variant="light" title={`${criticalAlerts.length} Critical Alert${criticalAlerts.length > 1 ? 's' : ''}`}>
          <Stack className="concentration-alerts__critical-list" gap="xs">
            {criticalAlerts.map((alert) => (
              <div className="concentration-alerts__item concentration-alerts__item--critical" key={alert.name} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                {/* Alert name and type (geography, industry, or security) */}
                <Text className="concentration-alerts__item-name" size="sm" fw={500}>
                  {alert.name} ({alert.type})
                </Text>
                {/* Current percentage vs limit */}
                <Text className="concentration-alerts__item-value" size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }} fw={600}>
                  {formatPercent(alert.current_pct)} / {formatPercent(alert.limit_pct, 0)} limit
                </Text>
              </div>
            ))}
          </Stack>
        </Alert>
      )}

      {/* Warning alerts - yellow, shown after critical */}
      {warningAlerts.length > 0 && (
        <Alert className="concentration-alerts__warning" color="yellow" variant="light" title={`${warningAlerts.length} Warning${warningAlerts.length > 1 ? 's' : ''}`}>
          <Stack className="concentration-alerts__warning-list" gap="xs">
            {warningAlerts.map((alert) => (
              <div className="concentration-alerts__item concentration-alerts__item--warning" key={alert.name} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                {/* Alert name and type */}
                <Text className="concentration-alerts__item-name" size="sm" fw={500}>
                  {alert.name} ({alert.type})
                </Text>
                {/* Current percentage vs limit */}
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
