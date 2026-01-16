/**
 * Logs View
 * 
 * Displays application logs for debugging and monitoring.
 * Shows system events, errors, and operational messages.
 */
import { LogsViewer } from '../components/system/LogsViewer';

/**
 * Logs view component
 * 
 * Wrapper view for the LogsViewer component.
 * 
 * @returns {JSX.Element} Logs view with log viewer component
 */
export function Logs() {
  return (
    <div className="logs-view">
      <LogsViewer />
    </div>
  );
}
