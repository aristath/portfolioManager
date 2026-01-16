/**
 * Trades Table Component
 * 
 * Displays recent trading activity including executed trades and pending orders.
 * Shows trade details: date, symbol, name, side (BUY/SELL), quantity, price, and value.
 * 
 * Features:
 * - Pending orders displayed first (highlighted in yellow)
 * - Executed trades with execution date
 * - Cash flow trades (currency pairs) highlighted differently
 * - Color-coded side badges (green for BUY, red for SELL)
 * - Responsive design (hides name/price columns on small screens)
 * 
 * Used in the Recent Trades view to monitor trading activity.
 */
import { Card, Table, Text, Badge, Group } from '@mantine/core';
import { useTradesStore } from '../../stores/tradesStore';
import { formatCurrency, formatDateTime } from '../../utils/formatters';

/**
 * Trades table component
 * 
 * Displays pending orders and executed trades in a table format.
 * 
 * @returns {JSX.Element} Trades table with pending orders and executed trades
 */
export function TradesTable() {
  const { trades, pendingOrders } = useTradesStore();

  const hasPending = pendingOrders.length > 0;
  const hasData = trades.length > 0 || hasPending;

  return (
    <Card className="trades-table" p="md" style={{ backgroundColor: 'var(--mantine-color-dark-7)', border: '1px solid var(--mantine-color-dark-6)' }}>
      <Group className="trades-table__header" justify="space-between" mb="md">
        <Text className="trades-table__title" size="xs" tt="uppercase" c="dimmed" fw={600} style={{ fontFamily: 'var(--mantine-font-family)' }}>
          Recent Trades
        </Text>
        {hasPending && (
          <Badge className="trades-table__pending-badge" size="sm" color="yellow" variant="light">
            {pendingOrders.length} pending
          </Badge>
        )}
      </Group>

      {!hasData ? (
        <Text className="trades-table__empty" c="dimmed" size="sm" ta="center" py="xl">
          No trades yet
        </Text>
      ) : (
        <div className="trades-table__scroll-container" style={{ overflowX: 'auto' }}>
          <Table className="trades-table__table" highlightOnHover>
            <Table.Thead className="trades-table__thead">
              <Table.Tr className="trades-table__header-row">
                <Table.Th className="trades-table__th trades-table__th--date">Date</Table.Th>
                <Table.Th className="trades-table__th trades-table__th--symbol">Symbol</Table.Th>
                <Table.Th className="trades-table__th trades-table__th--name" visibleFrom="sm">Name</Table.Th>
                <Table.Th className="trades-table__th trades-table__th--side">Side</Table.Th>
                <Table.Th className="trades-table__th trades-table__th--qty" ta="right">Qty</Table.Th>
                <Table.Th className="trades-table__th trades-table__th--price" ta="right" visibleFrom="sm">Price</Table.Th>
                <Table.Th className="trades-table__th trades-table__th--value" ta="right">Value</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody className="trades-table__tbody">
              {/* Pending orders - shown first, highlighted in yellow */}
              {pendingOrders.map((order) => {
                // Check if this is a cash flow trade (currency pair)
                const isCash = order.symbol.includes('/');
                return (
                  <Table.Tr
                    className={`trades-table__row trades-table__row--pending ${isCash ? 'trades-table__row--cash' : ''}`}
                    key={`pending-${order.order_id}`}
                    style={{
                      backgroundColor: 'rgba(255, 193, 7, 0.1)',  // Yellow highlight for pending
                    }}
                  >
                    {/* Status badge - PENDING */}
                    <Table.Td className="trades-table__td trades-table__td--status">
                      <Badge className="trades-table__status-badge" size="xs" color="yellow" variant="filled">
                        PENDING
                      </Badge>
                    </Table.Td>
                    <Table.Td className="trades-table__td trades-table__td--symbol">
                      <Text
                        className="trades-table__symbol"
                        size="sm"
                        style={{ fontFamily: 'var(--mantine-font-family)' }}
                        c={isCash ? 'violet' : 'yellow'}
                      >
                        {order.symbol}
                      </Text>
                    </Table.Td>
                    <Table.Td className="trades-table__td trades-table__td--name" visibleFrom="sm">
                      <Text
                        className="trades-table__name"
                        size="sm"
                        c="dimmed"
                        truncate
                        style={{ maxWidth: '128px', fontFamily: 'var(--mantine-font-family)' }}
                      >
                        {order.symbol}
                      </Text>
                    </Table.Td>
                    <Table.Td className="trades-table__td trades-table__td--side">
                      <Badge
                        className={`trades-table__side-badge trades-table__side-badge--${order.side.toLowerCase()}`}
                        size="sm"
                        color={order.side === 'BUY' ? 'green' : 'red'}
                        variant="light"
                        style={{ fontFamily: 'var(--mantine-font-family)' }}
                      >
                        {order.side}
                      </Badge>
                    </Table.Td>
                    <Table.Td className="trades-table__td trades-table__td--qty" ta="right">
                      <Text className="trades-table__qty" size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }} c="dimmed">
                        {formatCurrency(order.quantity)}
                      </Text>
                    </Table.Td>
                    <Table.Td className="trades-table__td trades-table__td--price" ta="right" visibleFrom="sm">
                      <Text className="trades-table__price" size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }} c="dimmed">
                        {formatCurrency(order.price)}
                      </Text>
                    </Table.Td>
                    <Table.Td className="trades-table__td trades-table__td--value" ta="right">
                      <Text className="trades-table__value" size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }} fw={600}>
                        {formatCurrency(order.quantity * order.price)}
                      </Text>
                    </Table.Td>
                  </Table.Tr>
                );
              })}
              {/* Executed trades - shown after pending orders */}
              {trades.map((trade) => {
                // Check if this is a cash flow trade (currency pair)
                const isCash = trade.symbol.includes('/');
                return (
                  <Table.Tr
                    className={`trades-table__row trades-table__row--executed ${isCash ? 'trades-table__row--cash' : ''}`}
                    key={trade.id}
                    style={{
                      // Cash trades have darker background
                      backgroundColor: isCash ? 'var(--mantine-color-dark-8)' : undefined,
                    }}
                  >
                    {/* Execution date */}
                    <Table.Td className="trades-table__td trades-table__td--date">
                      <Text className="trades-table__date" size="sm" c="dimmed">
                        {formatDateTime(trade.executed_at)}
                      </Text>
                    </Table.Td>
                    <Table.Td className="trades-table__td trades-table__td--symbol">
                      <Text
                        className="trades-table__symbol"
                        size="sm"
                        style={{ fontFamily: 'var(--mantine-font-family)' }}
                        c={isCash ? 'violet' : 'blue'}
                      >
                        {trade.symbol}
                      </Text>
                    </Table.Td>
                    <Table.Td className="trades-table__td trades-table__td--name" visibleFrom="sm">
                      <Text
                        className="trades-table__name"
                        size="sm"
                        c={isCash ? 'violet' : 'dimmed'}
                        truncate
                        style={{ maxWidth: '128px', fontFamily: 'var(--mantine-font-family)' }}
                      >
                        {trade.name || trade.symbol}
                      </Text>
                    </Table.Td>
                    <Table.Td className="trades-table__td trades-table__td--side">
                      <Badge
                        className={`trades-table__side-badge trades-table__side-badge--${trade.side.toLowerCase()}`}
                        size="sm"
                        color={trade.side === 'BUY' ? 'green' : 'red'}
                        variant="light"
                        style={{ fontFamily: 'var(--mantine-font-family)' }}
                      >
                        {trade.side}
                      </Badge>
                    </Table.Td>
                    <Table.Td className="trades-table__td trades-table__td--qty" ta="right">
                      <Text className="trades-table__qty" size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }} c="dimmed">
                        {formatCurrency(trade.quantity)}
                      </Text>
                    </Table.Td>
                    <Table.Td className="trades-table__td trades-table__td--price" ta="right" visibleFrom="sm">
                      <Text className="trades-table__price" size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }} c="dimmed">
                        {formatCurrency(trade.price)}
                      </Text>
                    </Table.Td>
                    <Table.Td className="trades-table__td trades-table__td--value" ta="right">
                      <Text className="trades-table__value" size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }} fw={600}>
                        {formatCurrency(trade.quantity * trade.price)}
                      </Text>
                    </Table.Td>
                  </Table.Tr>
                );
              })}
            </Table.Tbody>
          </Table>
        </div>
      )}
    </Card>
  );
}
