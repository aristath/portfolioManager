/**
 * Diversification View
 * 
 * Displays portfolio diversification analysis with radar charts and concentration alerts.
 * 
 * Features:
 * - Concentration alerts (warnings for over-concentration)
 * - Geography radar chart (geographic allocation)
 * - Industry radar chart (sector/industry allocation)
 * 
 * This view helps identify diversification issues and monitor allocation across
 * different dimensions (geography, industry).
 */
import { Grid, Stack } from '@mantine/core';
import { GeographyRadarCard } from '../components/charts/GeographyRadarCard';
import { IndustryRadarCard } from '../components/charts/IndustryRadarCard';
import { ConcentrationAlerts } from '../components/portfolio/ConcentrationAlerts';

/**
 * Diversification view component
 * 
 * Displays portfolio diversification metrics and alerts.
 * 
 * @returns {JSX.Element} Diversification view with alerts and radar charts
 */
export function Diversification() {
  return (
    <Stack className="diversification-view" gap="md">
      {/* Concentration alerts - shown at top for visibility */}
      <ConcentrationAlerts />
      
      {/* Radar charts in responsive grid - side by side on desktop, stacked on mobile */}
      <Grid className="diversification-view__charts" mt="md">
        {/* Geography radar chart - shows allocation by geographic region */}
        <Grid.Col className="diversification-view__geography-col" span={{ base: 12, md: 6 }}>
          <GeographyRadarCard />
        </Grid.Col>
        
        {/* Industry radar chart - shows allocation by sector/industry */}
        <Grid.Col className="diversification-view__industry-col" span={{ base: 12, md: 6 }}>
          <IndustryRadarCard />
        </Grid.Col>
      </Grid>
    </Stack>
  );
}
