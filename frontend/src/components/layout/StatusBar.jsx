/**
 * Status Bar Component
 *
 * Displays system status, activity, and portfolio summary in a horizontal bar.
 * Features:
 * - System health indicator (online/offline)
 * - Last sync timestamp
 * - Running and completed job badges
 * - Portfolio summary (total value, cash, positions)
 * - Inline editing for TEST cash (click to edit)
 * - Cash breakdown by currency
 *
 * Provides a quick overview of system state and portfolio metrics.
 */
import { Paper, Group, Text, NumberInput, Badge, Loader } from '@mantine/core';
import { IconCheck, IconX } from '@tabler/icons-react';
import { useState } from 'react';
import { useAppStore } from '../../stores/appStore';
import { usePortfolioStore } from '../../stores/portfolioStore';
import { formatCurrency, formatNumber, formatTimestamp } from '../../utils/formatters';

/**
 * Status bar component
 *
 * Displays system status, job activity, and portfolio summary.
 * Supports inline editing of TEST cash for paper trading.
 *
 * @returns {JSX.Element} Status bar with system info, activity, and portfolio summary
 */
export function StatusBar() {
  const { status, showMessage, runningJobs, completedJobs } = useAppStore();
  const { allocation, cashBreakdown, updateTestCash } = usePortfolioStore();

  // State for inline editing of TEST cash
  const [editingTestCash, setEditingTestCash] = useState(false);
  const [testCashValue, setTestCashValue] = useState(null);

  // Check if there's any job activity (running or completed jobs)
  const hasActivity = Object.keys(runningJobs).length > 0 || Object.keys(completedJobs).length > 0;

  return (
    <Paper
      className="status-bar"
      p="md"
      style={{
        backgroundColor: 'var(--mantine-color-dark-7)',
        border: '1px solid var(--mantine-color-dark-6)',
      }}
    >
      {/* System Status Row - health indicator and last sync time */}
      <Group className="status-bar__system" justify="space-between" mb="xs">
        <Group className="status-bar__system-info" gap="md">
          {/* System health indicator: green dot for healthy, red for unhealthy */}
          <Group className="status-bar__health" gap="xs">
            <div
              className={`status-bar__health-indicator status-bar__health-indicator--${status.status === 'healthy' ? 'healthy' : 'unhealthy'}`}
              style={{
                width: '6px',
                height: '6px',
                borderRadius: '50%',
                backgroundColor: status.status === 'healthy' ? 'var(--mantine-color-green-0)' : 'var(--mantine-color-red-0)',
              }}
            />
            <Text className="status-bar__health-text" size="xs" c="dimmed" ff="var(--mantine-font-family)">
              {status.status === 'healthy' ? 'System Online' : 'System Offline'}
            </Text>
          </Group>
          <Text className="status-bar__separator" size="xs" c="dimmed" ff="var(--mantine-font-family)">|</Text>
          {/* Last sync timestamp (when prices/data were last updated) */}
          <Text className="status-bar__sync" size="xs" c="dimmed" ff="var(--mantine-font-family)">
            Last sync: <span className="status-bar__sync-time">{status.last_sync ? formatTimestamp(status.last_sync) : 'Never'}</span>
          </Text>
        </Group>
      </Group>

      {/* Activity Row - shows running and completed background jobs */}
      <Group className="status-bar__activity" gap="sm" wrap="wrap" mb="xs">
        <Text className="status-bar__activity-label" size="xs" c="dimmed" fw={500} ff="var(--mantine-font-family)">
          Activity:
        </Text>
        {/* Show IDLE if no jobs are running or recently completed */}
        {!hasActivity && (
          <Text className="status-bar__activity-idle" size="xs" c="dimmed" ff="var(--mantine-font-family)">
            IDLE
          </Text>
        )}
        {/* Running jobs - blue badges with spinning loader */}
        {Object.values(runningJobs).map((job) => (
          <Badge
            className="status-bar__job status-bar__job--running"
            key={job.jobId}
            size="sm"
            variant="light"
            color="blue"
            leftSection={<Loader size={10} />}
            style={{ fontFamily: 'var(--mantine-font-family)' }}
          >
            {job.description}
            {/* Show progress if available (e.g., "Step 5/10") */}
            {job.progress && job.progress.total > 0 &&
              ` (${job.progress.current}/${job.progress.total})`}
          </Badge>
        ))}
        {/* Completed jobs - green for success, red for failure */}
        {Object.values(completedJobs).map((job) => (
          <Badge
            className={`status-bar__job status-bar__job--${job.status === 'completed' ? 'completed' : 'failed'}`}
            key={job.jobId}
            size="sm"
            variant="light"
            color={job.status === 'completed' ? 'green' : 'red'}
            leftSection={
              job.status === 'completed' ?
                <IconCheck size={12} /> :
                <IconX size={12} />
            }
            style={{ fontFamily: 'var(--mantine-font-family)' }}
          >
            {job.description}
          </Badge>
        ))}
      </Group>

      {/* Portfolio Summary Row - total value, cash, positions */}
      <Group className="status-bar__portfolio" justify="space-between">
        <Group className="status-bar__portfolio-info" gap="md" wrap="wrap">
          {/* Total portfolio value */}
          <Text className="status-bar__total-value" size="xs" c="dimmed" ff="var(--mantine-font-family)">
            Total Value: <span className="status-bar__total-value-amount" style={{ color: 'var(--mantine-color-green-0)' }}>
              {formatCurrency(allocation.total_value)}
            </span>
          </Text>
          <Text className="status-bar__separator" size="xs" c="dimmed" ff="var(--mantine-font-family)">|</Text>
          {/* Total cash balance */}
          <Text className="status-bar__cash" size="xs" c="dimmed" ff="var(--mantine-font-family)">
            Cash: <span className="status-bar__cash-amount">
              {formatCurrency(allocation.cash_balance)}
            </span>
          </Text>
          {/* Cash breakdown by currency (with inline editing for TEST cash) */}
          {cashBreakdown && cashBreakdown.length > 0 && (
            <>
              <Text className="status-bar__cash-breakdown" size="xs" c="dimmed" ff="var(--mantine-font-family)">
                ({cashBreakdown.map((cb, index) => {
                  // Special handling for TEST cash - allows inline editing
                  if (cb.currency === 'TEST') {
                    const displayAmount = cb.amount ?? 0;
                    const isEditing = editingTestCash;
                    // Use edited value if available, otherwise use display amount
                    const currentValue = testCashValue !== null ? testCashValue : displayAmount;

                    return (
                      <span className="status-bar__currency status-bar__currency--test" key={cb.currency}>
                        {/* Highlighted TEST cash with inline editing capability */}
                        <span className="status-bar__currency-highlight" style={{ backgroundColor: 'rgba(166, 227, 161, 0.15)', padding: '2px 4px', borderRadius: '2px', border: '1px solid rgba(166, 227, 161, 0.3)' }}>
                          <span className="status-bar__currency-code" style={{ color: 'var(--mantine-color-green-0)' }}>{cb.currency}</span>:
                          {isEditing ? (
                            // Inline number input for editing TEST cash
                            <NumberInput
                              className="status-bar__test-cash-input"
                              value={currentValue}
                              onChange={(val) => setTestCashValue(val ?? 0)}
                              onBlur={async () => {
                                // Save on blur (when user clicks away)
                                try {
                                  await updateTestCash(currentValue);
                                  setEditingTestCash(false);
                                  setTestCashValue(null);
                                  showMessage('TEST cash updated', 'success');
                                } catch (error) {
                                  showMessage(`Failed to update TEST cash: ${error.message}`, 'error');
                                  setEditingTestCash(false);
                                  setTestCashValue(null);
                                }
                              }}
                              onKeyDown={(e) => {
                                // Enter: save and exit edit mode
                                if (e.key === 'Enter') {
                                  e.currentTarget.blur();
                                }
                                // Escape: cancel editing without saving
                                else if (e.key === 'Escape') {
                                  setEditingTestCash(false);
                                  setTestCashValue(null);
                                }
                              }}
                              size="xs"
                              min={0}
                              step={0.01}
                              precision={2}
                              style={{
                                display: 'inline-block',
                                width: '80px',
                                marginLeft: '4px',
                              }}
                              styles={{
                                input: {
                                  color: 'var(--mantine-color-green-0)',
                                  backgroundColor: 'rgba(166, 227, 161, 0.2)',
                                  border: '1px solid rgba(166, 227, 161, 0.5)',
                                  fontSize: 'var(--mantine-font-size-xs)',
                                  padding: '2px 4px',
                                  height: 'auto',
                                  minHeight: 'unset',
                                },
                              }}
                              autoFocus
                            />
                          ) : (
                            // Clickable display value - click to edit
                            <span
                              className="status-bar__test-cash-value"
                              style={{
                                color: 'var(--mantine-color-green-0)',
                                cursor: 'pointer',
                                textDecoration: 'underline',
                                textDecorationStyle: 'dotted',
                              }}
                              onClick={() => {
                                // Enter edit mode when clicked
                                setEditingTestCash(true);
                                setTestCashValue(displayAmount);
                              }}
                              title="Click to edit"
                            >
                              {formatNumber(displayAmount, 2)}
                            </span>
                          )}
                        </span>
                      </span>
                    );
                  }
                  // Regular currency display (non-TEST currencies)
                  return (
                    <span className="status-bar__currency" key={cb.currency}>
                      <span className="status-bar__currency-item">
                        {cb.currency}: <span className="status-bar__currency-amount">{formatNumber(cb.amount ?? 0, 2)}</span>
                      </span>
                      {index < cashBreakdown.length - 1 && ', '}
                    </span>
                  );
                })})
              </Text>
            </>
          )}
          <Text className="status-bar__separator" size="xs" c="dimmed" ff="var(--mantine-font-family)">|</Text>
          {/* Active positions count */}
          <Text className="status-bar__positions" size="xs" c="dimmed" ff="var(--mantine-font-family)">
            Positions: <span className="status-bar__positions-count">
              {status.active_positions || 0}
            </span>
          </Text>
        </Group>
      </Group>
    </Paper>
  );
}
