/**
 * Main React application component
 *
 * This is the root component of the Sentinel frontend application. It sets up
 * the essential infrastructure:
 * - ErrorBoundary: Catches and handles React errors gracefully, preventing
 *   the entire app from crashing on component errors
 * - BrowserRouter: Provides React Router context for client-side routing
 * - Router: Defines all application routes and navigation
 *
 * The component hierarchy:
 * App
 *   └─ ErrorBoundary (error handling)
 *       └─ BrowserRouter (routing context)
 *           └─ Router (route definitions)
 */
import { BrowserRouter } from 'react-router-dom';
import { Router } from './router';
import { ErrorBoundary } from './components/ErrorBoundary';

/**
 * App component - Root of the React application
 *
 * Wraps the application in error boundary and routing context.
 * This ensures that:
 * 1. Any unhandled errors in child components are caught and displayed
 *    gracefully instead of crashing the entire app
 * 2. Client-side routing works correctly throughout the application
 *
 * @returns {JSX.Element} The root application component
 */
function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <Router />
      </BrowserRouter>
    </ErrorBoundary>
  );
}

export default App;
