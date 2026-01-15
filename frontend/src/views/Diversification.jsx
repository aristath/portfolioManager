import { Grid, Stack } from '@mantine/core';
import { GeographyRadarCard } from '../components/charts/GeographyRadarCard';
import { IndustryRadarCard } from '../components/charts/IndustryRadarCard';
import { ConcentrationAlerts } from '../components/portfolio/ConcentrationAlerts';

export function Diversification() {
  return (
    <Stack className="diversification-view" gap="md">
      <ConcentrationAlerts />
      <Grid className="diversification-view__charts" mt="md">
        <Grid.Col className="diversification-view__geography-col" span={{ base: 12, md: 6 }}>
          <GeographyRadarCard />
        </Grid.Col>
        <Grid.Col className="diversification-view__industry-col" span={{ base: 12, md: 6 }}>
          <IndustryRadarCard />
        </Grid.Col>
      </Grid>
    </Stack>
  );
}
