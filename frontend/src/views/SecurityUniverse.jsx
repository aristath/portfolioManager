import { Stack } from '@mantine/core';
import { SecurityTable } from '../components/portfolio/SecurityTable';

export function SecurityUniverse() {
  return (
    <Stack className="security-universe-view" gap="md">
      <SecurityTable />
    </Stack>
  );
}
