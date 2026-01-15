import { Group, Badge, Text } from '@mantine/core';
import { useAppStore } from '../../stores/appStore';

// Map exchange codes to regions
function getRegion(code) {
  const US = ['XNAS', 'XNYS'];
  const EU = ['XETR', 'XLON', 'XPAR', 'XMIL', 'XAMS', 'XCSE', 'ASEX'];
  const ASIA = ['XHKG', 'XSHG', 'XTSE', 'XASX'];

  if (US.includes(code)) return 'US';
  if (EU.includes(code)) return 'EU';
  if (ASIA.includes(code)) return 'ASIA';
  return 'OTHER';
}

export function MarketStatus() {
  const { markets } = useAppStore();

  // Defensive check for markets
  if (!markets || typeof markets !== 'object') {
    return null;
  }

  // Group markets by region
  const marketsByRegion = Object.entries(markets).reduce((acc, [code, market]) => {
    const region = getRegion(code);
    if (!acc[region]) acc[region] = [];
    acc[region].push({ code, ...market });
    return acc;
  }, {});

  // Sort regions
  const regionOrder = ['US', 'EU', 'ASIA', 'OTHER'];
  const sortedRegions = Object.entries(marketsByRegion).sort(
    ([a], [b]) => regionOrder.indexOf(a) - regionOrder.indexOf(b)
  );

  if (sortedRegions.length === 0) {
    return null;
  }

  return (
    <Group className="market-status" gap="md" wrap="wrap" mb="md">
      {sortedRegions.map(([region, regionMarkets]) => (
        <Group className="market-status__region" key={region} gap="xs" align="center">
          <Text className="market-status__region-label" size="xs" fw={500} c="dimmed">
            {region}:
          </Text>
          {regionMarkets.map((market) => (
            <Badge
              className={`market-status__market market-status__market--${market.status}`}
              key={market.code}
              color={market.status === 'open' ? 'green' : 'gray'}
              variant="light"
              size="sm"
              title={
                market.status === 'open'
                  ? `${market.name} open (closes ${market.close_time})`
                  : `${market.name} closed${market.open_time ? ` (opens ${market.open_time})` : ''}`
              }
            >
              {market.name}
            </Badge>
          ))}
        </Group>
      ))}
    </Group>
  );
}
