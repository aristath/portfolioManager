import { Component } from 'react';
import { Alert, Button, Stack, Text } from '@mantine/core';

export class ErrorBoundary extends Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
  }

  render() {
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
              <Text className="error-boundary__message" style={{ fontFamily: 'var(--mantine-font-family)' }}>
                An unexpected error occurred. Please refresh the page or contact support if the problem persists.
              </Text>
              {this.state.error && (
                <details className="error-boundary__details" style={{ marginTop: '1rem' }}>
                  <summary className="error-boundary__summary" style={{
                    cursor: 'pointer',
                    marginBottom: '0.5rem',
                    fontFamily: 'var(--mantine-font-family)',
                  }}>
                    Error Details
                  </summary>
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
              <Button
                className="error-boundary__reload-btn"
                onClick={() => {
                  this.setState({ hasError: false, error: null });
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

    return this.props.children;
  }
}
