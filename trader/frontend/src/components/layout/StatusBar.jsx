import { Paper, Group, Text } from '@mantine/core';
import { useAppStore } from '../../stores/appStore';
import { usePortfolioStore } from '../../stores/portfolioStore';
import { formatCurrency, formatNumber } from '../../utils/formatters';

export function StatusBar() {
  const { status } = useAppStore();
  const { allocation, cashBreakdown } = usePortfolioStore();

  return (
    <Paper
      p="md"
      style={{
        backgroundColor: 'var(--mantine-color-dark-7)',
        border: '1px solid var(--mantine-color-dark-6)',
      }}
    >
      {/* System Status Row */}
      <Group justify="space-between" mb="xs">
        <Group gap="md">
          <Group gap="xs">
            <div
              style={{
                width: '6px',
                height: '6px',
                borderRadius: '50%',
                backgroundColor: status.status === 'healthy' ? 'var(--mantine-color-green-0)' : 'var(--mantine-color-red-0)',
              }}
            />
            <Text size="xs" c="dimmed" ff="var(--mantine-font-family)">
              {status.status === 'healthy' ? 'System Online' : 'System Offline'}
            </Text>
          </Group>
          <Text size="xs" c="dimmed" ff="var(--mantine-font-family)">|</Text>
          <Text size="xs" c="dimmed" ff="var(--mantine-font-family)">
            Last sync: <span>{status.last_sync || 'Never'}</span>
          </Text>
        </Group>
      </Group>

      {/* Portfolio Summary Row */}
      <Group justify="space-between">
        <Group gap="md" wrap="wrap">
          <Text size="xs" c="dimmed" ff="var(--mantine-font-family)">
            Total Value: <span style={{ color: 'var(--mantine-color-green-0)' }}>
              {formatCurrency(allocation.total_value)}
            </span>
          </Text>
          <Text size="xs" c="dimmed" ff="var(--mantine-font-family)">|</Text>
          <Text size="xs" c="dimmed" ff="var(--mantine-font-family)">
            Cash: <span>
              {formatCurrency(allocation.cash_balance)}
            </span>
          </Text>
          {cashBreakdown && cashBreakdown.length > 0 && (
            <>
              <Text size="xs" c="dimmed" ff="var(--mantine-font-family)">
                ({cashBreakdown.map((cb, index) => (
                  <span key={cb.currency}>
                    {cb.currency === 'TEST' ? (
                      <span style={{ backgroundColor: 'rgba(166, 227, 161, 0.15)', padding: '2px 4px', borderRadius: '2px', border: '1px solid rgba(166, 227, 161, 0.3)' }}>
                        <span style={{ color: 'var(--mantine-color-green-0)' }}>{cb.currency}</span>:
                        <span style={{ color: 'var(--mantine-color-green-0)' }}>
                          {formatNumber(cb.amount, 2)}
                        </span>
                      </span>
                    ) : (
                      <span>
                        {cb.currency}: <span>{formatNumber(cb.amount, 2)}</span>
                      </span>
                    )}
                    {index < cashBreakdown.length - 1 && ', '}
                  </span>
                ))})
              </Text>
            </>
          )}
          <Text size="xs" c="dimmed" ff="var(--mantine-font-family)">|</Text>
          <Text size="xs" c="dimmed" ff="var(--mantine-font-family)">
            Positions: <span>
              {status.active_positions || 0}
            </span>
          </Text>
        </Group>
      </Group>
    </Paper>
  );
}
