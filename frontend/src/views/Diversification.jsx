import { Grid, Stack } from '@mantine/core';
import { GeographyRadarCard } from '../components/charts/GeographyRadarCard';
import { IndustryRadarCard } from '../components/charts/IndustryRadarCard';
import { ConcentrationAlerts } from '../components/portfolio/ConcentrationAlerts';

export function Diversification() {
  return (
    <Stack gap="md">
      <ConcentrationAlerts />
      <Grid mt="md">
        <Grid.Col span={{ base: 12, md: 6 }}>
          <GeographyRadarCard />
        </Grid.Col>
        <Grid.Col span={{ base: 12, md: 6 }}>
          <IndustryRadarCard />
        </Grid.Col>
      </Grid>
    </Stack>
  );
}
