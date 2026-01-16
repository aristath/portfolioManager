/**
 * Radar Chart Component
 * 
 * SVG-based radar/spider chart for displaying multi-dimensional data.
 * Shows both target and current allocation percentages for comparison.
 * 
 * Features:
 * - Circular radar chart with configurable number of axes
 * - Target allocation (dashed blue line)
 * - Current allocation (solid green line with filled area)
 * - Grid circles (0%, 25%, 50%, 75%, 100%)
 * - Radial lines from center to each axis
 * - Tick marks and labels on 0° axis
 * - Point labels (geography/industry names) outside chart
 * - Legend showing Target vs Current
 * - Auto-scaling based on data range
 * - Catppuccin Mocha color scheme
 * 
 * Used by AllocationRadar component for geography and industry visualization.
 */
import { useEffect, useRef } from 'react';

/**
 * Radar chart component
 * 
 * Renders an SVG-based radar chart comparing target vs current allocation.
 * 
 * @param {Object} props - Component props
 * @param {Array<string>} props.labels - Axis labels (geography/industry names)
 * @param {Array<number>} props.targetData - Target allocation percentages (0-100)
 * @param {Array<number>} props.currentData - Current allocation percentages (0-100)
 * @param {number|null} props.maxValue - Maximum value for scaling (auto-calculated if null)
 * @returns {JSX.Element} Radar chart SVG component
 */
export function RadarChart({ labels = [], targetData = [], currentData = [], maxValue = null }) {
  const svgRef = useRef(null);

  // Render SVG chart when data changes
  useEffect(() => {
    if (!svgRef.current) return;

    // Validate data - clear chart if no labels
    if (!labels || labels.length === 0) {
      svgRef.current.innerHTML = '';
      return;
    }

    // Validate data length matches
    if (targetData.length !== labels.length || currentData.length !== labels.length) {
      console.warn('RadarChart: Data length mismatch');
      return;
    }

    // Calculate max value if not provided
    const allValues = [...targetData, ...currentData];
    let maxVal = maxValue;
    if (!maxVal || maxVal <= 0) {
      maxVal = allValues.length > 0 ? Math.max(...allValues) : 100;
    }

    // Add 25% padding and round to nearest step of 25 for clean tick marks
    const paddedMax = Math.ceil(maxVal * 1.25);
    const roundedMax = Math.ceil(paddedMax / 25) * 25;

    // Chart dimensions and constants
    const centerX = 250;
    const centerY = 250;
    const radius = 180;
    const numPoints = labels.length;

    // Clear previous content
    svgRef.current.innerHTML = '';

    // Create SVG namespace
    const svgNS = 'http://www.w3.org/2000/svg';

    // Create SVG groups for organized rendering
    const gridGroup = document.createElementNS(svgNS, 'g');
    gridGroup.setAttribute('class', 'radar-chart__grid');

    const radialGroup = document.createElementNS(svgNS, 'g');
    radialGroup.setAttribute('class', 'radar-chart__radial-lines');

    const tickGroup = document.createElementNS(svgNS, 'g');
    tickGroup.setAttribute('class', 'radar-chart__ticks');

    const dataGroup = document.createElementNS(svgNS, 'g');
    dataGroup.setAttribute('class', 'radar-chart__data');

    const labelGroup = document.createElementNS(svgNS, 'g');
    labelGroup.setAttribute('class', 'radar-chart__labels');

    const legendGroup = document.createElementNS(svgNS, 'g');
    legendGroup.setAttribute('class', 'radar-chart__legend');

    // Draw grid circles (5 levels: 0%, 25%, 50%, 75%, 100%)
    // These provide visual reference for allocation percentages
    for (let i = 0; i <= 4; i++) {
      const level = i / 4;
      const circleRadius = radius * level;
      const circle = document.createElementNS(svgNS, 'circle');
      circle.setAttribute('cx', centerX);
      circle.setAttribute('cy', centerY);
      circle.setAttribute('r', circleRadius);
      circle.setAttribute('fill', 'none');
      circle.setAttribute('stroke', '#313244'); // Catppuccin Mocha Surface 0
      circle.setAttribute('stroke-width', '1');
      circle.setAttribute('class', 'radar-chart__grid-circle');
      gridGroup.appendChild(circle);
    }

    // Calculate angles and coordinates for each data point
    // Angles start at -90° (top) and distribute evenly around circle
    const angles = [];
    const coordinates = [];

    for (let i = 0; i < numPoints; i++) {
      const angle = (360 / numPoints) * i - 90;  // Start at top (-90°)
      const angleRad = (angle * Math.PI) / 180;
      angles.push(angleRad);

      // Calculate endpoint coordinates for each axis
      const x = centerX + radius * Math.cos(angleRad);
      const y = centerY + radius * Math.sin(angleRad);
      coordinates.push({ x, y, angle: angleRad });
    }

    // Draw radial lines from center to each axis endpoint
    coordinates.forEach(coord => {
      const line = document.createElementNS(svgNS, 'line');
      line.setAttribute('x1', centerX);
      line.setAttribute('y1', centerY);
      line.setAttribute('x2', coord.x);
      line.setAttribute('y2', coord.y);
      line.setAttribute('stroke', '#313244'); // Catppuccin Mocha Surface 0
      line.setAttribute('stroke-width', '1');
      line.setAttribute('class', 'radar-chart__radial-line');
      radialGroup.appendChild(line);
    });

    // Draw tick marks and labels on top radial line (0° axis)
    // Shows percentage values at each grid level
    const tickLevels = [0, 0.25, 0.5, 0.75, 1.0];
    const tickLength = 5;
    const tickLabelOffset = 15;

    tickLevels.forEach((level) => {
      const tickRadius = radius * level;
      const tickX = centerX + tickRadius;  // On 0° axis (right side)
      const tickY = centerY;

      // Tick mark (vertical line)
      const tick = document.createElementNS(svgNS, 'line');
      tick.setAttribute('x1', tickX);
      tick.setAttribute('y1', tickY - tickLength);
      tick.setAttribute('x2', tickX);
      tick.setAttribute('y2', tickY + tickLength);
      tick.setAttribute('stroke', '#6c7086'); // Catppuccin Mocha Overlay 0
      tick.setAttribute('stroke-width', '1');
      tick.setAttribute('class', 'radar-chart__tick');
      tickGroup.appendChild(tick);

      // Tick label (percentage value)
      const tickValue = Math.round(roundedMax * level);
      const label = document.createElementNS(svgNS, 'text');
      label.setAttribute('x', tickX);
      label.setAttribute('y', tickY - tickLabelOffset);
      label.setAttribute('text-anchor', 'middle');
      label.setAttribute('fill', '#6c7086'); // Catppuccin Mocha Overlay 0
      label.setAttribute('font-size', '9');
      label.setAttribute('font-family', 'JetBrains Mono, Fira Code, IBM Plex Mono, monospace');
      label.setAttribute('class', 'radar-chart__tick-label');
      label.textContent = tickValue.toString();
      tickGroup.appendChild(label);
    });

    // Calculate data point coordinates for target and current allocations
    // Values are normalized to 0-1 range based on roundedMax
    const targetPoints = [];
    const currentPoints = [];

    coordinates.forEach((coord, i) => {
      // Normalize values to 0-1 range (clamped)
      const targetValue = Math.max(0, Math.min(targetData[i] / roundedMax, 1));
      const currentValue = Math.max(0, Math.min(currentData[i] / roundedMax, 1));

      // Calculate radius for each point based on normalized value
      const targetRadius = radius * targetValue;
      const currentRadius = radius * currentValue;

      // Calculate x,y coordinates for each point
      targetPoints.push({
        x: centerX + targetRadius * Math.cos(coord.angle),
        y: centerY + targetRadius * Math.sin(coord.angle)
      });

      currentPoints.push({
        x: centerX + currentRadius * Math.cos(coord.angle),
        y: centerY + currentRadius * Math.sin(coord.angle)
      });
    });

    // Draw Current dataset (filled polygon + solid polyline)
    // Green color indicates current allocation
    if (currentPoints.length > 0) {
      // Filled polygon for area visualization
      const currentPolygon = document.createElementNS(svgNS, 'polygon');
      const currentPolygonPoints = currentPoints.map(p => `${p.x},${p.y}`).join(' ');
      currentPolygon.setAttribute('points', currentPolygonPoints);
      currentPolygon.setAttribute('fill', '#a6e3a1'); // Catppuccin Mocha Green
      currentPolygon.setAttribute('fill-opacity', '0.2');
      currentPolygon.setAttribute('stroke', 'none');
      currentPolygon.setAttribute('class', 'radar-chart__current-polygon');
      dataGroup.appendChild(currentPolygon);

      // Solid polyline connecting all points
      const currentPolyline = document.createElementNS(svgNS, 'polyline');
      const currentPolylinePoints = [...currentPoints, currentPoints[0]].map(p => `${p.x},${p.y}`).join(' ');  // Close polygon
      currentPolyline.setAttribute('points', currentPolylinePoints);
      currentPolyline.setAttribute('fill', 'none');
      currentPolyline.setAttribute('stroke', '#a6e3a1'); // Catppuccin Mocha Green
      currentPolyline.setAttribute('stroke-opacity', '0.8');
      currentPolyline.setAttribute('stroke-width', '2');
      currentPolyline.setAttribute('class', 'radar-chart__current-line');
      dataGroup.appendChild(currentPolyline);

      // Data point markers (circles)
      currentPoints.forEach(point => {
        const circle = document.createElementNS(svgNS, 'circle');
        circle.setAttribute('cx', point.x);
        circle.setAttribute('cy', point.y);
        circle.setAttribute('r', '3');
        circle.setAttribute('fill', '#a6e3a1'); // Catppuccin Mocha Green
        circle.setAttribute('fill-opacity', '0.8');
        circle.setAttribute('class', 'radar-chart__current-point');
        dataGroup.appendChild(circle);
      });
    }

    // Draw Target dataset (dashed polyline)
    // Blue dashed line indicates target allocation
    if (targetPoints.length > 0) {
      const targetPolyline = document.createElementNS(svgNS, 'polyline');
      const targetPolylinePoints = [...targetPoints, targetPoints[0]].map(p => `${p.x},${p.y}`).join(' ');  // Close polygon
      targetPolyline.setAttribute('points', targetPolylinePoints);
      targetPolyline.setAttribute('fill', 'none');
      targetPolyline.setAttribute('stroke', '#89b4fa'); // Catppuccin Mocha Blue
      targetPolyline.setAttribute('stroke-opacity', '0.8');
      targetPolyline.setAttribute('stroke-width', '2');
      targetPolyline.setAttribute('stroke-dasharray', '5,5');  // Dashed line
      targetPolyline.setAttribute('class', 'radar-chart__target-line');
      dataGroup.appendChild(targetPolyline);

      // Data point markers (circles)
      targetPoints.forEach(point => {
        const circle = document.createElementNS(svgNS, 'circle');
        circle.setAttribute('cx', point.x);
        circle.setAttribute('cy', point.y);
        circle.setAttribute('r', '3');
        circle.setAttribute('fill', '#89b4fa'); // Catppuccin Mocha Blue
        circle.setAttribute('fill-opacity', '0.8');
        circle.setAttribute('class', 'radar-chart__target-point');
        dataGroup.appendChild(circle);
      });
    }

    // Draw point labels (geography/industry names) outside chart area
    // Positioned along each axis, offset from the endpoint
    const labelOffset = 20;
    coordinates.forEach((coord, i) => {
      const labelX = coord.x + labelOffset * Math.cos(coord.angle);
      const labelY = coord.y + labelOffset * Math.sin(coord.angle);

      const label = document.createElementNS(svgNS, 'text');
      label.setAttribute('x', labelX);
      label.setAttribute('y', labelY);
      // Text anchor based on position (left/right/middle)
      label.setAttribute('text-anchor', coord.x > centerX ? 'start' : coord.x < centerX ? 'end' : 'middle');
      label.setAttribute('dominant-baseline', 'middle');
      label.setAttribute('fill', '#cdd6f4'); // Catppuccin Mocha Text
      label.setAttribute('font-size', '10');
      label.setAttribute('font-family', 'JetBrains Mono, Fira Code, IBM Plex Mono, monospace');
      label.setAttribute('class', 'radar-chart__point-label');
      label.textContent = labels[i];
      labelGroup.appendChild(label);
    });

    // Draw legend at bottom of chart
    // Shows Target (dashed blue) and Current (solid green) line styles
    const legendY = 480;
    const legendX = centerX;

    // Target legend item (dashed blue line)
    const targetLegendLine = document.createElementNS(svgNS, 'line');
    targetLegendLine.setAttribute('x1', legendX - 60);
    targetLegendLine.setAttribute('y1', legendY);
    targetLegendLine.setAttribute('x2', legendX - 40);
    targetLegendLine.setAttribute('y2', legendY);
    targetLegendLine.setAttribute('stroke', '#89b4fa'); // Catppuccin Mocha Blue
    targetLegendLine.setAttribute('stroke-opacity', '0.8');
    targetLegendLine.setAttribute('stroke-width', '2');
    targetLegendLine.setAttribute('stroke-dasharray', '5,5');  // Dashed
    targetLegendLine.setAttribute('class', 'radar-chart__legend-line radar-chart__legend-line--target');
    legendGroup.appendChild(targetLegendLine);

    const targetLegendText = document.createElementNS(svgNS, 'text');
    targetLegendText.setAttribute('x', legendX - 35);
    targetLegendText.setAttribute('y', legendY);
    targetLegendText.setAttribute('dominant-baseline', 'middle');
    targetLegendText.setAttribute('fill', '#a6adc8'); // Catppuccin Mocha Subtext 0
    targetLegendText.setAttribute('font-size', '10');
    targetLegendText.setAttribute('font-family', 'JetBrains Mono, Fira Code, IBM Plex Mono, monospace');
    targetLegendText.setAttribute('class', 'radar-chart__legend-text radar-chart__legend-text--target');
    targetLegendText.textContent = 'Target';
    legendGroup.appendChild(targetLegendText);

    // Current legend item (solid green line)
    const currentLegendLine = document.createElementNS(svgNS, 'line');
    currentLegendLine.setAttribute('x1', legendX + 10);
    currentLegendLine.setAttribute('y1', legendY);
    currentLegendLine.setAttribute('x2', legendX + 30);
    currentLegendLine.setAttribute('y2', legendY);
    currentLegendLine.setAttribute('stroke', '#a6e3a1'); // Catppuccin Mocha Green
    currentLegendLine.setAttribute('stroke-opacity', '0.8');
    currentLegendLine.setAttribute('stroke-width', '2');
    currentLegendLine.setAttribute('class', 'radar-chart__legend-line radar-chart__legend-line--current');
    legendGroup.appendChild(currentLegendLine);

    const currentLegendText = document.createElementNS(svgNS, 'text');
    currentLegendText.setAttribute('x', legendX + 35);
    currentLegendText.setAttribute('y', legendY);
    currentLegendText.setAttribute('dominant-baseline', 'middle');
    currentLegendText.setAttribute('fill', '#a6adc8'); // Catppuccin Mocha Subtext 0
    currentLegendText.setAttribute('font-size', '10');
    currentLegendText.setAttribute('font-family', 'JetBrains Mono, Fira Code, IBM Plex Mono, monospace');
    currentLegendText.setAttribute('class', 'radar-chart__legend-text radar-chart__legend-text--current');
    currentLegendText.textContent = 'Current';
    legendGroup.appendChild(currentLegendText);

    // Append all groups to SVG in render order (background to foreground)
    svgRef.current.appendChild(gridGroup);      // Grid circles (background)
    svgRef.current.appendChild(radialGroup);    // Radial lines
    svgRef.current.appendChild(tickGroup);       // Tick marks and labels
    svgRef.current.appendChild(dataGroup);      // Data polygons and lines
    svgRef.current.appendChild(labelGroup);      // Point labels
    svgRef.current.appendChild(legendGroup);    // Legend (foreground)
  }, [labels, targetData, currentData, maxValue]);

  return (
    <div className="radar-chart" style={{ position: 'relative', width: '100%', aspectRatio: '1' }}>
      {/* SVG container with fixed viewBox for consistent scaling */}
      <svg
        className="radar-chart__svg"
        ref={svgRef}
        viewBox="0 0 500 500"
        style={{ width: '100%', height: '100%' }}
        preserveAspectRatio="xMidYMid meet"
      />
    </div>
  );
}
