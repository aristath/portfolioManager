/**
 * Market Status Component
 * 
 * Displays the status of global stock exchanges grouped by region.
 * Shows whether each market is open or closed, with tooltips showing
 * open/close times.
 * 
 * Features:
 * - Groups exchanges by region (US, EU, ASIA, OTHER)
 * - Color-coded badges (green for open, gray for closed)
 * - Tooltips with open/close times
 * - Sorted by region priority
 */
import { Group, Badge, Text } from '@mantine/core';
import { useAppStore } from '../../stores/appStore';

/**
 * Maps exchange codes to geographic regions
 * 
 * @param {string} code - Exchange code (e.g., 'XNAS', 'XETR', 'XHKG')
 * @returns {string} Region name: 'US', 'EU', 'ASIA', or 'OTHER'
 */
function getRegion(code) {
  const US = ['XNAS', 'XNYS'];
  const EU = ['XETR', 'XLON', 'XPAR', 'XMIL', 'XAMS', 'XCSE', 'ASEX'];
  const ASIA = ['XHKG', 'XSHG', 'XTSE', 'XASX'];

  if (US.includes(code)) return 'US';
  if (EU.includes(code)) return 'EU';
  if (ASIA.includes(code)) return 'ASIA';
  return 'OTHER';
}

/**
 * Market status component
 * 
 * Displays exchange status badges grouped by region.
 * Returns null if no market data is available.
 * 
 * @returns {JSX.Element|null} Market status component or null if no data
 */
export function MarketStatus() {
  const { markets } = useAppStore();

  // Defensive check for markets - return null if invalid data
  if (!markets || typeof markets !== 'object') {
    return null;
  }

  // Group markets by region for organized display
  const marketsByRegion = Object.entries(markets).reduce((acc, [code, market]) => {
    const region = getRegion(code);
    if (!acc[region]) acc[region] = [];
    acc[region].push({ code, ...market });
    return acc;
  }, {});

  // Sort regions by priority order
  const regionOrder = ['US', 'EU', 'ASIA', 'OTHER'];
  const sortedRegions = Object.entries(marketsByRegion).sort(
    ([a], [b]) => regionOrder.indexOf(a) - regionOrder.indexOf(b)
  );

  // Return null if no markets to display
  if (sortedRegions.length === 0) {
    return null;
  }

  return (
    <Group className="market-status" gap="md" wrap="wrap" mb="md">
      {sortedRegions.map(([region, regionMarkets]) => (
        <Group className="market-status__region" key={region} gap="xs" align="center">
          {/* Region label */}
          <Text className="market-status__region-label" size="xs" fw={500} c="dimmed">
            {region}:
          </Text>
          {/* Market badges - green for open, gray for closed */}
          {regionMarkets.map((market) => (
            <Badge
              className={`market-status__market market-status__market--${market.status}`}
              key={market.code}
              color={market.status === 'open' ? 'green' : 'gray'}
              variant="light"
              size="sm"
              title={
                // Tooltip shows open/close times
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
