/**
 * Recent Trades View
 * 
 * Displays recent trading activity in a table format.
 * Shows executed trades, their details, and status.
 */
import { TradesTable } from '../components/trading/TradesTable';

/**
 * Recent Trades view component
 * 
 * Wrapper view for the TradesTable component.
 * 
 * @returns {JSX.Element} Recent Trades view with trades table
 */
export function RecentTrades() {
  return (
    <div className="recent-trades-view">
      <TradesTable />
    </div>
  );
}
