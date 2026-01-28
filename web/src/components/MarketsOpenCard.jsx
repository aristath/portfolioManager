/**
 * Markets Open Card Component
 *
 * Shows which markets are open, filtered to markets that have securities in our universe.
 */
import { useQuery } from '@tanstack/react-query';
import { Card, Group, Text, Badge } from '@mantine/core';
import { getMarketsStatus } from '../api/client';

export function MarketsOpenCard() {
  const { data } = useQuery({
    queryKey: ['markets-status'],
    queryFn: getMarketsStatus,
    refetchInterval: 60000, // Refresh every minute
  });

  const markets = data?.markets || [];

  return (
    <Card className="markets-card" p="sm" withBorder>
      <Group gap="xs" wrap="wrap">
        <Text size="xs" c="dimmed" fw={600} tt="uppercase">Markets</Text>
        {markets.length === 0 && (
          <Text size="xs" c="dimmed" fs="italic">No data</Text>
        )}
        {markets.map((market) => (
          <Badge
            key={market.name}
            size="sm"
            color={market.is_open ? 'green' : 'gray'}
            variant={market.is_open ? 'light' : 'outline'}
          >
            {market.name}
          </Badge>
        ))}
      </Group>
    </Card>
  );
}
