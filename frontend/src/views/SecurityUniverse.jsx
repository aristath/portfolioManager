/**
 * Security Universe View
 * 
 * Displays the investment universe (all securities) in a table format.
 * This view provides access to the SecurityTable component for managing
 * and viewing all securities in the portfolio.
 */
import { Stack } from '@mantine/core';
import { SecurityTable } from '../components/portfolio/SecurityTable';

/**
 * Security Universe view component
 * 
 * Wrapper view for the SecurityTable component.
 * 
 * @returns {JSX.Element} Security Universe view with security table
 */
export function SecurityUniverse() {
  return (
    <Stack className="security-universe-view" gap="md">
      <SecurityTable />
    </Stack>
  );
}
