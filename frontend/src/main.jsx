/**
 * React Application Entry Point
 *
 * This is the entry point for the Sentinel frontend React application.
 * It initializes React, sets up the Mantine UI library, and renders the App component.
 *
 * Component Hierarchy:
 * - React.StrictMode: Enables React strict mode for development warnings
 * - MantineProvider: Provides Mantine UI theme and styling context
 * - Notifications: Global notification system for user feedback
 * - App: Main application component (routing, error boundary)
 *
 * The application is rendered into the DOM element with id="root" (defined in index.html).
 */
import React from 'react';
import ReactDOM from 'react-dom/client';
import { MantineProvider } from '@mantine/core';
import { Notifications } from '@mantine/notifications';
import { theme } from './theme';
import App from './App';
import '@mantine/core/styles.css';
import '@mantine/notifications/styles.css';
import './styles/animations.css';
import './styles/terminal.css';

/**
 * Initialize and render the React application
 *
 * Uses React 18's createRoot API for concurrent rendering.
 * The application is wrapped in:
 * - StrictMode: Helps identify potential problems during development
 * - MantineProvider: Provides theme and dark mode support
 * - Notifications: Enables toast notifications throughout the app
 */
ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    {/* MantineProvider sets up the UI library with custom theme and dark mode */}
    <MantineProvider theme={theme} forceColorScheme="dark">
      {/* Notifications component enables toast notifications (success, error, etc.) */}
      <Notifications />
      {/* Main App component (contains routing and error boundary) */}
      <App />
    </MantineProvider>
  </React.StrictMode>
);
