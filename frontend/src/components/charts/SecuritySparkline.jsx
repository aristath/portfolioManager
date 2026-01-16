/**
 * Security Sparkline Component
 * 
 * Custom SVG-based mini line chart (sparkline) for displaying price trends.
 * Used in the security table to show price history at a glance.
 * 
 * Features:
 * - Cost-basis-aware coloring (green above avg price, red below) for positions
 * - Trend-based coloring (blue/gray) for non-positions or missing avg price
 * - Segmented coloring when price crosses average price
 * - Responsive to sparkline data updates
 * 
 * The sparkline provides a quick visual indicator of price performance.
 */
import { useEffect, useRef } from 'react';
import { useSecuritiesStore } from '../../stores/securitiesStore';

/**
 * Security sparkline component
 * 
 * Renders a mini line chart showing price trend over time.
 * Uses cost-basis-aware coloring when average price is available.
 * 
 * @param {Object} props - Component props
 * @param {string} props.symbol - Security symbol
 * @param {boolean} props.hasPosition - Whether user has a position in this security
 * @param {number|null} props.avgPrice - Average purchase price (for cost-basis coloring)
 * @param {number|null} props.currentPrice - Current price (not used in rendering, but triggers re-render)
 * @returns {JSX.Element} Sparkline chart or empty indicator
 */
export function SecuritySparkline({ symbol, hasPosition = false, avgPrice, currentPrice }) {
  const containerRef = useRef(null);
  const { sparklines } = useSecuritiesStore();

  // Render sparkline when data changes
  useEffect(() => {
    if (!containerRef.current || !symbol) return;

    const data = sparklines[symbol];
    // Show empty indicator if no data or insufficient data points
    if (!data || data.length < 2) {
      containerRef.current.innerHTML = '<span class="sparkline__empty" style="color: var(--mantine-color-dimmed); font-size: 0.875rem;">-</span>';
      return;
    }

    // Chart dimensions
    const width = 80;
    const height = 32;
    const padding = 1;

    // Extract values and find min/max for scaling
    const values = data.map(d => d.value);
    const minValue = Math.min(...values);
    const maxValue = Math.max(...values);
    const valueRange = maxValue - minValue || 1;  // Prevent division by zero

    // Calculate scaling factors for chart area
    const chartWidth = width - padding * 2;
    const chartHeight = height - padding * 2;

    // Build SVG path points (x, y coordinates)
    const points = values.map((value, index) => {
      const x = padding + (index / (values.length - 1)) * chartWidth;
      const y = padding + chartHeight - ((value - minValue) / valueRange) * chartHeight;
      return `${x},${y}`;
    });

    // Build SVG content with cost-basis-aware coloring
    let svgContent;

    if (hasPosition && avgPrice && avgPrice > 0) {
      // Cost-basis-aware coloring: split into segments at average price crossings
      // Green segments = price above average (profit), Red segments = price below average (loss)
      const segments = [];
      let currentSegmentPoints = [points[0]];
      let currentColor = values[0] >= avgPrice ? '#a6e3a1' : '#f38ba8';  // Green or red

      // Build segments by detecting when price crosses average price
      for (let i = 1; i < values.length; i++) {
        const value = values[i];
        const isAbove = value >= avgPrice;
        const segmentColor = isAbove ? '#a6e3a1' : '#f38ba8'; // Green above, red below

        if (segmentColor !== currentColor) {
          // Color changed - price crossed average price
          // Save current segment and start new one
          segments.push({
            path: `M ${currentSegmentPoints.join(' L ')}`,
            color: currentColor,
          });
          // Start new segment from previous point to ensure continuity
          currentSegmentPoints = [points[i - 1], points[i]];
          currentColor = segmentColor;
        } else {
          // Same color - continue current segment
          currentSegmentPoints.push(points[i]);
        }
      }

      // Save last segment
      if (currentSegmentPoints.length > 0) {
        segments.push({
          path: `M ${currentSegmentPoints.join(' L ')}`,
          color: currentColor,
        });
      }

      // Render multiple path segments with different colors
      const svgPaths = segments.map(seg =>
        `<path
           class="sparkline__segment"
           d="${seg.path}"
           fill="none"
           stroke="${seg.color}"
           stroke-width="1.5"
           stroke-linecap="round"
           stroke-linejoin="round"
         />`
      ).join('');

      svgContent = `<svg class="sparkline__svg" width="${width}" height="${height}" style="display: block;">
        ${svgPaths}
      </svg>`;
    } else {
      // Default trend-based coloring for non-positions or missing avg_price
      // Blue for positive trend, gray for negative trend
      const firstValue = values[0];
      const lastValue = values[values.length - 1];
      const isPositive = lastValue > firstValue;
      const color = hasPosition
        ? (isPositive ? '#89b4fa' : '#6c7086')  // Blue/gray for positions without avg_price
        : (isPositive ? '#89b4fa' : '#6c7086'); // Blue/gray for non-positions
      const pathData = `M ${points.join(' L ')}`;

      // Single path for simple trend coloring
      svgContent = `<svg class="sparkline__svg" width="${width}" height="${height}" style="display: block;">
        <path
          class="sparkline__path"
          d="${pathData}"
          fill="none"
          stroke="${color}"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
      </svg>`;
    }

    // Render SVG into container
    containerRef.current.innerHTML = svgContent;
  }, [symbol, hasPosition, avgPrice, currentPrice, sparklines]);

  // Return empty indicator if no symbol
  if (!symbol) {
    return <span className="sparkline__empty" style={{ color: 'var(--mantine-color-dimmed)', fontSize: '0.875rem' }}>-</span>;
  }

  // Return container div - SVG is rendered via innerHTML in useEffect
  return <div className="sparkline" ref={containerRef} style={{ display: 'inline-block' }} />;
}
