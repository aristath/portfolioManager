import { useEffect, useRef } from 'react';
import { useSecuritiesStore } from '../../stores/securitiesStore';

/**
 * Security Sparkline Component
 * Custom SVG-based sparkline chart for security table
 * Supports cost-basis-aware coloring when avgPrice is provided
 */
export function SecuritySparkline({ symbol, hasPosition = false, avgPrice, currentPrice }) {
  const containerRef = useRef(null);
  const { sparklines } = useSecuritiesStore();

  useEffect(() => {
    if (!containerRef.current || !symbol) return;

    const data = sparklines[symbol];
    if (!data || data.length < 2) {
      containerRef.current.innerHTML = '<span class="sparkline__empty" style="color: var(--mantine-color-dimmed); font-size: 0.875rem;">-</span>';
      return;
    }

    const width = 80;
    const height = 32;
    const padding = 1;

    // Extract values and find min/max
    const values = data.map(d => d.value);
    const minValue = Math.min(...values);
    const maxValue = Math.max(...values);
    const valueRange = maxValue - minValue || 1;

    // Calculate scaling factors
    const chartWidth = width - padding * 2;
    const chartHeight = height - padding * 2;

    // Build SVG path
    const points = values.map((value, index) => {
      const x = padding + (index / (values.length - 1)) * chartWidth;
      const y = padding + chartHeight - ((value - minValue) / valueRange) * chartHeight;
      return `${x},${y}`;
    });

    // Build SVG content with cost-basis-aware coloring
    let svgContent;

    if (hasPosition && avgPrice && avgPrice > 0) {
      // Cost-basis-aware coloring: split into segments at average price crossings
      const segments = [];
      let currentSegmentPoints = [points[0]];
      let currentColor = values[0] >= avgPrice ? '#a6e3a1' : '#f38ba8';

      for (let i = 1; i < values.length; i++) {
        const value = values[i];
        const isAbove = value >= avgPrice;
        const segmentColor = isAbove ? '#a6e3a1' : '#f38ba8'; // Green above, red below

        if (segmentColor !== currentColor) {
          // Color changed - save current segment and start new one
          segments.push({
            path: `M ${currentSegmentPoints.join(' L ')}`,
            color: currentColor,
          });
          // Start new segment from previous point to ensure continuity
          currentSegmentPoints = [points[i - 1], points[i]];
          currentColor = segmentColor;
        } else {
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

      // Render multiple path segments
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
      const firstValue = values[0];
      const lastValue = values[values.length - 1];
      const isPositive = lastValue > firstValue;
      const color = hasPosition
        ? (isPositive ? '#89b4fa' : '#6c7086')  // Blue/gray for positions without avg_price
        : (isPositive ? '#89b4fa' : '#6c7086'); // Blue/gray for non-positions
      const pathData = `M ${points.join(' L ')}`;

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

    containerRef.current.innerHTML = svgContent;
  }, [symbol, hasPosition, avgPrice, currentPrice, sparklines]);

  if (!symbol) {
    return <span className="sparkline__empty" style={{ color: 'var(--mantine-color-dimmed)', fontSize: '0.875rem' }}>-</span>;
  }

  return <div className="sparkline" ref={containerRef} style={{ display: 'inline-block' }} />;
}
