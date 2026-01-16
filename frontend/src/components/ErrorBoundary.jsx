/**
 * ErrorBoundary Component
 *
 * React error boundary that catches JavaScript errors anywhere in the child component tree,
 * logs those errors, and displays a fallback UI instead of crashing the entire application.
 *
 * This component implements the React error boundary pattern:
 * - Catches errors during rendering, in lifecycle methods, and in constructors
 * - Does NOT catch errors in event handlers, async code, or during server-side rendering
 *
 * When an error is caught:
 * 1. Updates state to show error UI
 * 2. Logs error to console for debugging
 * 3. Displays user-friendly error message with option to reload
 * 4. Shows error details in expandable section for debugging
 */
import { Component } from 'react';
import { Alert, Button, Stack, Text } from '@mantine/core';

/**
 * ErrorBoundary class component
 *
 * Wraps the application to catch and handle React errors gracefully.
 * Prevents the entire app from crashing when a component throws an error.
 */
export class ErrorBoundary extends Component {
  /**
   * Constructor - initializes error state
   * @param {Object} props - Component props (children will be rendered normally if no error)
   */
  constructor(props) {
    super(props);
    // State tracks whether an error has occurred and stores the error object
    this.state = { hasError: false, error: null };
  }

  /**
   * Static lifecycle method called when an error is thrown
   *
   * This method is called during the "render" phase, so side-effects are not allowed.
   * It should return an object to update state, or null to update nothing.
   *
   * @param {Error} error - The error that was thrown
   * @returns {Object} State update object with hasError and error
   */
  static getDerivedStateFromError(error) {
    // Update state so the next render will show the fallback UI
    return { hasError: true, error };
  }

  /**
   * Lifecycle method called after an error has been thrown
   *
   * This method is called during the "commit" phase, so side-effects are allowed.
   * Use this for logging errors to an error reporting service.
   *
   * @param {Error} error - The error that was thrown
   * @param {Object} errorInfo - Component stack trace information
   */
  componentDidCatch(error, errorInfo) {
    // Log error to console for debugging
    // In production, this could be sent to an error tracking service (e.g., Sentry)
    console.error('Error caught by boundary:', error, errorInfo);
  }

  /**
   * Render method - displays error UI or children
   *
   * If an error has been caught, displays a user-friendly error message.
   * Otherwise, renders children normally.
   *
   * @returns {JSX.Element} Error UI or children
   */
  render() {
    // If an error was caught, show error UI instead of crashing
    if (this.state.hasError) {
      return (
        <div className="error-boundary" style={{
          padding: '2rem',
          maxWidth: '800px',
          margin: '0 auto',
          fontFamily: 'var(--mantine-font-family)',
        }}>
          <Alert
            className="error-boundary__alert"
            color="red"
            title="Something went wrong"
            variant="filled"
            style={{ fontFamily: 'var(--mantine-font-family)' }}
          >
            <Stack className="error-boundary__content" gap="md">
              {/* User-friendly error message */}
              <Text className="error-boundary__message" style={{ fontFamily: 'var(--mantine-font-family)' }}>
                An unexpected error occurred. Please refresh the page or contact support if the problem persists.
              </Text>

              {/* Expandable error details for debugging */}
              {this.state.error && (
                <details className="error-boundary__details" style={{ marginTop: '1rem' }}>
                  <summary className="error-boundary__summary" style={{
                    cursor: 'pointer',
                    marginBottom: '0.5rem',
                    fontFamily: 'var(--mantine-font-family)',
                  }}>
                    Error Details
                  </summary>
                  {/* Error stack trace in a scrollable pre block */}
                  <pre className="error-boundary__stack" style={{
                    backgroundColor: 'var(--mantine-color-dark-7)',
                    border: '1px solid var(--mantine-color-dark-6)',
                    padding: '1rem',
                    borderRadius: '2px',
                    overflow: 'auto',
                    fontSize: '0.875rem',
                    fontFamily: 'var(--mantine-font-family)',
                    color: 'var(--mantine-color-dark-0)',
                  }}>
                    {this.state.error.toString()}
                    {this.state.error.stack && `\n${this.state.error.stack}`}
                  </pre>
                </details>
              )}

              {/* Reload button to reset error state and refresh page */}
              <Button
                className="error-boundary__reload-btn"
                onClick={() => {
                  // Clear error state (though page will reload anyway)
                  this.setState({ hasError: false, error: null });
                  // Reload the entire page to reset application state
                  window.location.reload();
                }}
                style={{ fontFamily: 'var(--mantine-font-family)' }}
              >
                Reload Page
              </Button>
            </Stack>
          </Alert>
        </div>
      );
    }

    // No error - render children normally
    return this.props.children;
  }
}
