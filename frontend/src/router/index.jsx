/**
 * Router Configuration
 *
 * Defines all application routes using React Router.
 * All routes are nested under the Layout component, which provides
 * the common UI structure (header, navigation, etc.).
 *
 * Route Structure:
 * - / (index) -> NextActions view (default/home page)
 * - /diversification -> Diversification view (allocation analysis)
 * - /security-universe -> SecurityUniverse view (security management)
 * - /recent-trades -> RecentTrades view (trade history)
 * - /logs -> Logs view (system logs)
 */
import { Routes, Route } from 'react-router-dom';
import { NextActions } from '../views/NextActions';
import { Diversification } from '../views/Diversification';
import { SecurityUniverse } from '../views/SecurityUniverse';
import { RecentTrades } from '../views/RecentTrades';
import { Logs } from '../views/Logs';
import { Layout } from '../components/layout/Layout';

/**
 * Router component - defines all application routes
 *
 * Uses React Router's Routes and Route components to define the routing structure.
 * All routes are nested under the Layout component, which provides:
 * - Common header/navigation
 * - Tab navigation between views
 * - Status bar
 * - Consistent layout structure
 *
 * @returns {JSX.Element} Routes configuration
 */
export function Router() {
  return (
    <Routes>
      {/* Root route with Layout wrapper - all child routes inherit the layout */}
      <Route path="/" element={<Layout />}>
        {/* Index route (/) - default page showing next actions/recommendations */}
        <Route index element={<NextActions />} />

        {/* Diversification route - shows portfolio allocation analysis */}
        <Route path="diversification" element={<Diversification />} />

        {/* Security Universe route - manage securities in investment universe */}
        <Route path="security-universe" element={<SecurityUniverse />} />

        {/* Recent Trades route - displays trade history */}
        <Route path="recent-trades" element={<RecentTrades />} />

        {/* Logs route - displays system logs for debugging */}
        <Route path="logs" element={<Logs />} />
      </Route>
    </Routes>
  );
}
