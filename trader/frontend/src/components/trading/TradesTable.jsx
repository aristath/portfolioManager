import { Card, Table, Text, Badge } from '@mantine/core';
import { useTradesStore } from '../../stores/tradesStore';
import { formatCurrency, formatDateTime } from '../../utils/formatters';

export function TradesTable() {
  const { trades } = useTradesStore();

  return (
    <Card p="md" style={{ backgroundColor: 'var(--mantine-color-dark-7)', border: '1px solid var(--mantine-color-dark-6)' }}>
      <Text size="xs" tt="uppercase" c="dimmed" fw={600} mb="md" style={{ fontFamily: 'var(--mantine-font-family)' }}>
        Recent Trades
      </Text>

      {trades.length === 0 ? (
        <Text c="dimmed" size="sm" ta="center" py="xl">
          No trades yet
        </Text>
      ) : (
        <div style={{ overflowX: 'auto' }}>
          <Table highlightOnHover>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Date</Table.Th>
                <Table.Th>Symbol</Table.Th>
                <Table.Th visibleFrom="sm">Name</Table.Th>
                <Table.Th>Side</Table.Th>
                <Table.Th ta="right">Qty</Table.Th>
                <Table.Th ta="right" visibleFrom="sm">Price</Table.Th>
                <Table.Th ta="right">Value</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {trades.map((trade) => {
                const isCash = trade.symbol.includes('/');
                return (
                  <Table.Tr
                    key={trade.id}
                    style={{
                      backgroundColor: isCash ? 'var(--mantine-color-dark-8)' : undefined,
                    }}
                  >
                    <Table.Td>
                      <Text size="sm" c="dimmed">
                        {formatDateTime(trade.executed_at)}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Text
                        size="sm"
                        style={{ fontFamily: 'var(--mantine-font-family)' }}
                        c={isCash ? 'violet' : 'blue'}
                      >
                        {trade.symbol}
                      </Text>
                    </Table.Td>
                    <Table.Td visibleFrom="sm">
                      <Text
                        size="sm"
                        c={isCash ? 'violet' : 'dimmed'}
                        truncate
                        style={{ maxWidth: '128px', fontFamily: 'var(--mantine-font-family)' }}
                      >
                        {trade.name || trade.symbol}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Badge
                        size="sm"
                        color={trade.side === 'BUY' ? 'green' : 'red'}
                        variant="light"
                        style={{ fontFamily: 'var(--mantine-font-family)' }}
                      >
                        {trade.side}
                      </Badge>
                    </Table.Td>
                    <Table.Td ta="right">
                      <Text size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }} c="dimmed">
                        {formatCurrency(trade.quantity)}
                      </Text>
                    </Table.Td>
                    <Table.Td ta="right" visibleFrom="sm">
                      <Text size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }} c="dimmed">
                        {formatCurrency(trade.price)}
                      </Text>
                    </Table.Td>
                    <Table.Td ta="right">
                      <Text size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }} fw={600}>
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
