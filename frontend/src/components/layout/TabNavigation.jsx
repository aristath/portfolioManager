/**
 * Tab Navigation Component
 * 
 * Provides tab-based navigation between main application views.
 * Features:
 * - Tab navigation with keyboard shortcuts (1-5 keys)
 * - Badge showing pending recommendations count on "Next Actions" tab
 * - Synchronizes with React Router for URL-based navigation
 * - Keyboard shortcuts work globally (except when typing in inputs)
 * 
 * Tabs:
 * - Next Actions (1): Trading recommendations and actions
 * - Diversification (2): Portfolio allocation and targets
 * - Security Universe (3): Investment universe management
 * - Recent Trades (4): Trade history
 * - Logs (5): System logs
 */
import { Tabs, Badge, Group } from '@mantine/core';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAppStore } from '../../stores/appStore';
import { useEffect, useCallback } from 'react';

/**
 * Tab navigation component
 * 
 * Provides tab-based navigation with keyboard shortcuts.
 * Synchronizes tab state with React Router location.
 * 
 * @returns {JSX.Element} Tab navigation component
 */
export function TabNavigation() {
  const navigate = useNavigate();
  const location = useLocation();
  const { recommendations, setActiveTab } = useAppStore();

  /**
   * Maps route paths to tab values
   * 
   * @param {string} path - Current route path
   * @returns {string} Tab value corresponding to the path
   */
  const getTabFromPath = (path) => {
    if (path === '/' || path === '/next-actions') return 'next-actions';
    if (path === '/diversification') return 'diversification';
    if (path === '/security-universe') return 'security-universe';
    if (path === '/recent-trades') return 'recent-trades';
    if (path === '/logs') return 'logs';
    return 'next-actions';  // Default to next-actions
  };

  // Determine active tab from current route
  const activeTab = getTabFromPath(location.pathname);

  /**
   * Handles tab change - updates store and navigates to corresponding route
   * 
   * @param {string} value - Tab value to navigate to
   */
  const handleTabChange = useCallback((value) => {
    setActiveTab(value);
    // Map tab values to route paths
    const routes = {
      'next-actions': '/',
      'diversification': '/diversification',
      'security-universe': '/security-universe',
      'recent-trades': '/recent-trades',
      'logs': '/logs',
    };
    navigate(routes[value] || '/');
  }, [navigate, setActiveTab]);

  // Keyboard shortcuts for quick navigation (1-5 keys)
  useEffect(() => {
    const handleKeydown = (e) => {
      // Don't trigger shortcuts when typing in inputs, textareas, or contenteditable elements
      if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.isContentEditable) return;
      // Don't trigger shortcuts when modifier keys are pressed
      if (e.ctrlKey || e.metaKey || e.altKey || e.shiftKey) return;

      // Map number keys to tab values
      const shortcuts = {
        '1': 'next-actions',
        '2': 'diversification',
        '3': 'security-universe',
        '4': 'recent-trades',
        '5': 'logs',
      };

      if (shortcuts[e.key]) {
        e.preventDefault();
        handleTabChange(shortcuts[e.key]);
      }
    };

    document.addEventListener('keydown', handleKeydown);
    // Cleanup: remove event listener on unmount
    return () => document.removeEventListener('keydown', handleKeydown);
  }, [handleTabChange]);

  // Count of pending recommendations for badge display
  const pendingCount = recommendations?.steps?.length || 0;

  return (
    <Tabs className="tab-nav" value={activeTab} onChange={handleTabChange}>
      <Tabs.List className="tab-nav__list">
        {/* Next Actions tab - shows badge with pending recommendations count */}
        <Tabs.Tab className="tab-nav__tab tab-nav__tab--next-actions" value="next-actions" style={{ fontFamily: 'var(--mantine-font-family)' }}>
          <Group className="tab-nav__tab-content" gap="xs">
            <span className="tab-nav__tab-label">Next Actions</span>
            {/* Badge shows count of pending recommendations */}
            {pendingCount > 0 && (
              <Badge className="tab-nav__badge pulse" size="xs" color="blue" variant="filled" style={{ fontFamily: 'var(--mantine-font-family)' }}>
                {pendingCount}
              </Badge>
            )}
          </Group>
        </Tabs.Tab>
        
        {/* Other tabs */}
        <Tabs.Tab className="tab-nav__tab tab-nav__tab--diversification" value="diversification" style={{ fontFamily: 'var(--mantine-font-family)' }}>Diversification</Tabs.Tab>
        <Tabs.Tab className="tab-nav__tab tab-nav__tab--security-universe" value="security-universe" style={{ fontFamily: 'var(--mantine-font-family)' }}>Security Universe</Tabs.Tab>
        <Tabs.Tab className="tab-nav__tab tab-nav__tab--recent-trades" value="recent-trades" style={{ fontFamily: 'var(--mantine-font-family)' }}>Recent Trades</Tabs.Tab>
        <Tabs.Tab className="tab-nav__tab tab-nav__tab--logs" value="logs" style={{ fontFamily: 'var(--mantine-font-family)' }}>Logs</Tabs.Tab>
        
        {/* Keyboard shortcut hint */}
        <div className="tab-nav__hint" style={{
          marginLeft: 'auto',
          fontSize: '0.875rem',
          color: 'var(--mantine-color-dimmed)',
          fontFamily: 'var(--mantine-font-family)',
        }}>
          Press <kbd className="tab-nav__kbd" style={{
            padding: '2px 6px',
            backgroundColor: 'var(--mantine-color-dark-7)',
            border: '1px solid var(--mantine-color-dark-6)',
            borderRadius: '2px',
            fontFamily: 'var(--mantine-font-family)',
          }}>1-5</kbd>
        </div>
      </Tabs.List>
    </Tabs>
  );
}
