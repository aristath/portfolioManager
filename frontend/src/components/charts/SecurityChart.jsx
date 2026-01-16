/**
 * Security Chart Component
 * 
 * Full-featured price chart component using TradingView Lightweight Charts library.
 * Displays historical price data for a security with interactive controls.
 * 
 * Features:
 * - Interactive price chart (line series)
 * - Time range selection (1M, 3M, 6M, 1Y, 2Y, 5Y, 10Y)
 * - Data source selection (currently Tradernet only)
 * - Responsive to window resize
 * - Catppuccin Mocha color scheme
 * - Refresh function exposed via ref for external updates
 * - Auto-refresh registration with event handlers
 * 
 * Used in SecurityChartModal for detailed price analysis.
 */
import { useEffect, useRef, useState, useImperativeHandle, forwardRef, useCallback } from 'react';
import { createChart, ColorType } from 'lightweight-charts';
import { api } from '../../api/client';
import { Select, Button, Group, Loader, Text } from '@mantine/core';
import { setSecurityChartRefreshFn } from '../../stores/eventHandlers';

/**
 * Security chart component (forwardRef for external refresh control)
 * 
 * Creates and manages a TradingView Lightweight Charts instance.
 * Exposes refresh function via ref for external updates.
 * 
 * @param {Object} props - Component props
 * @param {string} props.isin - Security ISIN identifier
 * @param {string} props.symbol - Security symbol (for display)
 * @param {Function} props.onClose - Optional close handler
 * @param {Object} ref - Ref object to expose refresh function
 * @returns {JSX.Element} Security chart with controls and chart container
 */
const SecurityChartComponent = forwardRef(({ isin, symbol, onClose }, ref) => {
  const chartContainerRef = useRef(null);
  const chartRef = useRef(null);
  const lineSeriesRef = useRef(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [selectedRange, setSelectedRange] = useState('1Y');
  const [selectedSource, setSelectedSource] = useState('tradernet');

  /**
   * Loads chart data from the API
   * 
   * Fetches historical price data and updates the chart.
   * Handles loading states and errors.
   */
  const loadChartData = useCallback(async () => {
    if (!isin) return;

    setLoading(true);
    setError(null);

    try {
      // Fetch chart data from API
      const data = await api.fetchSecurityChart(isin, selectedRange, selectedSource);

      if (!data || data.length === 0) {
        setError('No chart data available');
        if (lineSeriesRef.current) {
          lineSeriesRef.current.setData([]);
        }
        return;
      }

      // Transform data for Lightweight Charts format
      // Supports multiple field names: value, close, price
      const chartData = data.map(item => ({
        time: item.time,
        value: item.value || item.close || item.price,
      }));

      // Update chart with new data
      if (lineSeriesRef.current) {
        lineSeriesRef.current.setData(chartData);
        // Auto-fit time scale to show all data
        chartRef.current.timeScale().fitContent();
      }
    } catch (err) {
      console.error('Failed to load security chart:', err);
      setError('Failed to load chart data');
    } finally {
      setLoading(false);
    }
  }, [isin, selectedRange, selectedSource]);

  // Expose refresh function via ref for external control
  useImperativeHandle(ref, () => ({
    refresh: loadChartData,
  }), [loadChartData]);

  // Initialize chart on mount or when ISIN changes
  useEffect(() => {
    if (!chartContainerRef.current || !isin) return;

    // Create chart with Catppuccin Mocha color scheme
    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: '#1e1e2e' }, // Catppuccin Mocha Base
        textColor: '#cdd6f4', // Catppuccin Mocha Text
      },
      grid: {
        vertLines: { color: '#313244' }, // Catppuccin Mocha Surface 0
        horzLines: { color: '#313244' }, // Catppuccin Mocha Surface 0
      },
      width: chartContainerRef.current.clientWidth,
      height: 400,
    });

    // Add line series for price data
    const lineSeries = chart.addLineSeries({
      color: '#89b4fa', // Catppuccin Mocha Blue
      lineWidth: 2,
    });

    // Store references for data updates
    chartRef.current = chart;
    lineSeriesRef.current = lineSeries;

    // Handle window resize to maintain chart responsiveness
    const handleResize = () => {
      if (chartContainerRef.current && chartRef.current) {
        chartRef.current.applyOptions({
          width: chartContainerRef.current.clientWidth,
        });
      }
    };

    window.addEventListener('resize', handleResize);

    // Cleanup: remove event listener and destroy chart
    return () => {
      window.removeEventListener('resize', handleResize);
      if (chartRef.current) {
        chartRef.current.remove();
        chartRef.current = null;
        lineSeriesRef.current = null;
      }
    };
  }, [isin]);

  // Load chart data when range, source, or ISIN changes
  useEffect(() => {
    if (isin && lineSeriesRef.current) {
      loadChartData();
    }
  }, [isin, selectedRange, selectedSource, loadChartData]);

  // Register refresh function with event handler for external updates
  // Allows event handlers to trigger chart refresh when price data updates
  useEffect(() => {
    if (isin) {
      setSecurityChartRefreshFn(() => {
        if (lineSeriesRef.current) {
          loadChartData();
        }
      });

      // Cleanup: unregister refresh function on unmount
      return () => {
        setSecurityChartRefreshFn(null);
      };
    }
  }, [isin, loadChartData]);

  return (
    <div className="security-chart">
      {/* Header with title and controls */}
      <Group className="security-chart__header" justify="space-between" mb="md">
        <Text className="security-chart__title" size="lg" fw={600}>
          {symbol} Chart
        </Text>
        <Group className="security-chart__controls" gap="xs">
          {/* Time range selector */}
          <Select
            className="security-chart__range-select"
            size="xs"
            data={[
              { value: '1M', label: '1 Month' },
              { value: '3M', label: '3 Months' },
              { value: '6M', label: '6 Months' },
              { value: '1Y', label: '1 Year' },
              { value: '2Y', label: '2 Years' },
              { value: '5Y', label: '5 Years' },
              { value: '10Y', label: '10 Years' },
            ]}
            value={selectedRange}
            onChange={(val) => setSelectedRange(val || '1Y')}
            style={{ width: '120px' }}
          />
          {/* Data source selector (currently only Tradernet) */}
          <Select
            className="security-chart__source-select"
            size="xs"
            data={[
              { value: 'tradernet', label: 'Tradernet' },
            ]}
            value={selectedSource}
            onChange={(val) => setSelectedSource(val || 'tradernet')}
            disabled  // Only one source available currently
            style={{ width: '140px' }}
          />
          {/* Close button (if onClose handler provided) */}
          {onClose && (
            <Button className="security-chart__close-btn" size="xs" variant="subtle" onClick={onClose}>
              Close
            </Button>
          )}
        </Group>
      </Group>

      {/* Loading indicator */}
      {loading && (
        <div className="security-chart__loading" style={{ display: 'flex', justifyContent: 'center', padding: '40px' }}>
          <Loader />
        </div>
      )}

      {/* Error message */}
      {error && (
        <Text className="security-chart__error" c="red" ta="center" py="xl">
          {error}
        </Text>
      )}

      {/* Chart container - Lightweight Charts renders here */}
      <div
        className="security-chart__container"
        ref={chartContainerRef}
        style={{ width: '100%', height: '400px' }}
      />
    </div>
  );
});

// Set display name for React DevTools
SecurityChartComponent.displayName = 'SecurityChart';

export const SecurityChart = SecurityChartComponent;
