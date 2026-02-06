/**
 * Portfolio P&L Chart Component
 *
 * SVG-based chart with dual Y-axes:
 * Left axis [-1, +1]: Wavelet portfolio quality score (blue line)
 * Right axis [%]: Actual trailing CAGR (green/red area), 4 ML models (toggleable), Target (dashed)
 */
import { useState } from 'react';
import { catppuccin } from '../theme';
import { buildSmoothPath } from '../utils/chartUtils';
import { useResponsiveWidth } from '../hooks/useResponsiveWidth';

const ML_LINES = [
  { key: 'ml_xgboost', label: 'XGBoost', color: catppuccin.yellow },
  { key: 'ml_ridge', label: 'Ridge', color: catppuccin.peach },
  { key: 'ml_rf', label: 'RF', color: catppuccin.mauve },
  { key: 'ml_svr', label: 'SVR', color: catppuccin.teal },
];

/**
 * Renders the portfolio annualized return chart
 *
 * @param {Object} props
 * @param {Array} props.snapshots - Array of snapshot objects with annualized return fields
 * @param {Object} props.summary - Summary with target_ann_return
 * @param {number} props.height - Chart height (default 300)
 */
export function PortfolioPnLChart({
  snapshots = [],
  summary = null,
  height = 300,
}) {
  const [containerRef, width] = useResponsiveWidth(300);
  const [visibleML, setVisibleML] = useState({
    ml_xgboost: true,
    ml_ridge: true,
    ml_rf: true,
    ml_svr: true,
  });

  const toggleML = (key) => {
    setVisibleML((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  const formatPct = (value) => {
    if (value == null) return '';
    const sign = value >= 0 ? '+' : '';
    return `${sign}${value.toFixed(1)}%`;
  };

  const formatEur = (value) => {
    const sign = value >= 0 ? '+' : '';
    return `${sign}${value.toLocaleString('en-US', {
      style: 'currency',
      currency: 'EUR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    })}`;
  };

  // Legend component with clickable ML lines
  const Legend = () => (
    <div
      style={{
        display: 'flex',
        gap: '12px',
        fontSize: '10px',
        color: catppuccin.subtext0,
        padding: '0 4px',
        marginBottom: '2px',
        flexWrap: 'wrap',
      }}
    >
      {/* Static legend items */}
      {[
        { color: catppuccin.green, label: 'Actual' },
        { color: catppuccin.blue, label: 'Wavelet' },
      ].map(({ color, label }) => (
        <span key={label} style={{ display: 'flex', alignItems: 'center', gap: '3px' }}>
          <span
            style={{
              width: '6px',
              height: '6px',
              borderRadius: '50%',
              backgroundColor: color,
              display: 'inline-block',
            }}
          />
          {label}
        </span>
      ))}

      {/* Toggleable ML legend items */}
      {ML_LINES.map(({ key, label, color }) => (
        <span
          key={key}
          onClick={() => toggleML(key)}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '3px',
            cursor: 'pointer',
            opacity: visibleML[key] ? 1 : 0.35,
            userSelect: 'none',
          }}
        >
          <span
            style={{
              width: '6px',
              height: '6px',
              borderRadius: '50%',
              backgroundColor: color,
              display: 'inline-block',
            }}
          />
          {label}
        </span>
      ))}

      {/* Target (dashed) */}
      <span style={{ display: 'flex', alignItems: 'center', gap: '3px' }}>
        <svg width="10" height="10">
          <line x1="0" y1="5" x2="10" y2="5" stroke={catppuccin.overlay0} strokeWidth="1.5" strokeDasharray="2,2" />
        </svg>
        Target
      </span>
    </div>
  );

  const renderChart = () => {
    if (!snapshots || snapshots.length < 2) {
      return (
        <div
          style={{
            width: '100%',
            height,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: catppuccin.overlay1,
            fontSize: '0.875rem',
          }}
        >
          No data available
        </div>
      );
    }

    const padding = { top: 20, right: 60, bottom: 32, left: 30 };
    const chartWidth = width - padding.left - padding.right;
    const chartHeight = height - padding.top - padding.bottom;

    if (chartWidth <= 0 || chartHeight <= 0) return null;

    const targetReturn = summary?.target_ann_return ?? 11.0;

    // Primary Y-axis: return % (actual, ML models, target)
    const allValues = [0, targetReturn];
    snapshots.forEach((s) => {
      if (s.actual_ann_return != null) allValues.push(s.actual_ann_return);
      ML_LINES.forEach(({ key }) => {
        if (visibleML[key] && s[key] != null) allValues.push(s[key]);
      });
    });

    const rawMin = Math.min(...allValues);
    const rawMax = Math.max(...allValues);
    const range = rawMax - rawMin || 1;
    const paddedMin = rawMin - range * 0.1;
    const paddedMax = rawMax + range * 0.1;
    const valueRange = paddedMax - paddedMin;

    const scaleX = (i) => padding.left + (i / (snapshots.length - 1)) * chartWidth;
    const scaleY = (v) => padding.top + chartHeight - ((v - paddedMin) / valueRange) * chartHeight;

    // Secondary Y-axis: wavelet score [-1, +1]
    const scaleYWavelet = (v) => padding.top + chartHeight - ((v - (-1)) / 2) * chartHeight;

    const zeroY = scaleY(0);
    const targetY = scaleY(targetReturn);

    // Build actual return points with area fill (green/red split at zero)
    const actualPoints = [];
    snapshots.forEach((s, i) => {
      if (s.actual_ann_return != null) {
        actualPoints.push({ x: scaleX(i), y: scaleY(s.actual_ann_return), value: s.actual_ann_return });
      }
    });

    // Split actual points into positive/negative segments at zero crossings
    const segments = [];
    if (actualPoints.length >= 2) {
      let currentSegment = [actualPoints[0]];
      let currentIsPositive = actualPoints[0].value >= 0;

      for (let i = 1; i < actualPoints.length; i++) {
        const isPositive = actualPoints[i].value >= 0;
        if (isPositive !== currentIsPositive) {
          const prev = actualPoints[i - 1];
          const curr = actualPoints[i];
          const t = (0 - prev.value) / (curr.value - prev.value);
          const crossX = prev.x + t * (curr.x - prev.x);
          const crossPoint = { x: crossX, y: zeroY, value: 0 };

          currentSegment.push(crossPoint);
          segments.push({ points: currentSegment, isPositive: currentIsPositive });
          currentSegment = [crossPoint, actualPoints[i]];
          currentIsPositive = isPositive;
        } else {
          currentSegment.push(actualPoints[i]);
        }
      }
      segments.push({ points: currentSegment, isPositive: currentIsPositive });
    }

    // Build wavelet line points
    const waveletPoints = [];
    snapshots.forEach((s, i) => {
      if (s.wavelet_ann_return != null) {
        waveletPoints.push({ x: scaleX(i), y: scaleYWavelet(s.wavelet_ann_return) });
      }
    });
    const waveletPath = buildSmoothPath(waveletPoints);

    // Build per-model ML line points
    const mlPaths = {};
    ML_LINES.forEach(({ key }) => {
      if (!visibleML[key]) return;
      const points = [];
      snapshots.forEach((s, i) => {
        if (s[key] != null) {
          points.push({ x: scaleX(i), y: scaleY(s[key]) });
        }
      });
      mlPaths[key] = buildSmoothPath(points);
    });

    // Current actual value for the dot
    const lastActual = actualPoints.length > 0 ? actualPoints[actualPoints.length - 1] : null;

    return (
      <svg width={width} height={height} style={{ display: 'block' }}>
        <defs>
          <linearGradient id="ann-gradient-pos" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stopColor={catppuccin.green} stopOpacity={0.3} />
            <stop offset="100%" stopColor={catppuccin.green} stopOpacity={0.05} />
          </linearGradient>
          <linearGradient id="ann-gradient-neg" x1="0%" y1="100%" x2="0%" y2="0%">
            <stop offset="0%" stopColor={catppuccin.red} stopOpacity={0.3} />
            <stop offset="100%" stopColor={catppuccin.red} stopOpacity={0.05} />
          </linearGradient>
        </defs>

        {/* Zero line */}
        <line
          x1={padding.left}
          y1={zeroY}
          x2={width - padding.right}
          y2={zeroY}
          stroke={catppuccin.overlay0}
          strokeWidth={1}
          strokeDasharray="4,4"
          opacity={0.4}
        />
        <text x={width - padding.right + 5} y={zeroY + 4} fill={catppuccin.subtext0} fontSize="10">
          0%
        </text>

        {/* Target line */}
        <line
          x1={padding.left}
          y1={targetY}
          x2={width - padding.right}
          y2={targetY}
          stroke={catppuccin.overlay0}
          strokeWidth={1}
          strokeDasharray="4,4"
          opacity={0.6}
        />
        <text x={width - padding.right + 5} y={targetY + 4} fill={catppuccin.subtext0} fontSize="10">
          {formatPct(targetReturn)}
        </text>

        {/* Actual: area fills */}
        {segments.map((seg, idx) => {
          const segPath = buildSmoothPath(seg.points);
          if (!segPath || seg.points.length < 2) return null;
          const firstX = seg.points[0].x;
          const lastX = seg.points[seg.points.length - 1].x;
          const areaPath = `${segPath} L ${lastX},${zeroY} L ${firstX},${zeroY} Z`;
          return (
            <path
              key={`area-${idx}`}
              d={areaPath}
              fill={`url(#ann-gradient-${seg.isPositive ? 'pos' : 'neg'})`}
            />
          );
        })}

        {/* Actual: stroke lines */}
        {segments.map((seg, idx) => {
          const segPath = buildSmoothPath(seg.points);
          if (!segPath) return null;
          return (
            <path
              key={`line-${idx}`}
              d={segPath}
              fill="none"
              stroke={seg.isPositive ? catppuccin.green : catppuccin.red}
              strokeWidth={2}
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          );
        })}

        {/* Wavelet line */}
        {waveletPath && (
          <path
            d={waveletPath}
            fill="none"
            stroke={catppuccin.blue}
            strokeWidth={1.5}
            strokeLinecap="round"
            strokeLinejoin="round"
            opacity={0.8}
          />
        )}

        {/* Per-model ML lines */}
        {ML_LINES.map(({ key, color }) => {
          const path = mlPaths[key];
          if (!path) return null;
          return (
            <path
              key={key}
              d={path}
              fill="none"
              stroke={color}
              strokeWidth={1.5}
              strokeLinecap="round"
              strokeLinejoin="round"
              opacity={0.8}
            />
          );
        })}

        {/* Current value dot */}
        {lastActual && (
          <circle
            cx={lastActual.x}
            cy={lastActual.y}
            r={4}
            fill={lastActual.value >= 0 ? catppuccin.green : catppuccin.red}
            stroke={catppuccin.base}
            strokeWidth={2}
          />
        )}

        {/* Right Y-axis: return % labels */}
        <text x={width - padding.right + 5} y={padding.top + 10} fill={catppuccin.subtext0} fontSize="10">
          {formatPct(rawMax)}
        </text>
        <text x={width - padding.right + 5} y={height - padding.bottom} fill={catppuccin.subtext0} fontSize="10">
          {formatPct(rawMin)}
        </text>

        {/* Left Y-axis: wavelet [-1, +1] labels */}
        <text x={2} y={padding.top + 10} fill={catppuccin.blue} fontSize="9" opacity={0.6}>
          +1
        </text>
        <text x={2} y={padding.top + chartHeight / 2 + 3} fill={catppuccin.blue} fontSize="9" opacity={0.6}>
          0
        </text>
        <text x={2} y={height - padding.bottom} fill={catppuccin.blue} fontSize="9" opacity={0.6}>
          -1
        </text>

        {/* X-axis: date labels */}
        {(() => {
          const tickCount = Math.min(6, snapshots.length);
          const step = (snapshots.length - 1) / (tickCount - 1);
          const ticks = [];
          for (let t = 0; t < tickCount; t++) {
            const idx = Math.round(t * step);
            const s = snapshots[idx];
            if (!s?.date) continue;
            const [, mon, day] = s.date.split('-');
            const monthNames = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
            const label = `${monthNames[parseInt(mon, 10) - 1]} ${parseInt(day, 10)}`;
            ticks.push(
              <text
                key={`xtick-${idx}`}
                x={scaleX(idx)}
                y={height - 4}
                fill={catppuccin.subtext0}
                fontSize="9"
                textAnchor="middle"
              >
                {label}
              </text>
            );
          }
          return ticks;
        })()}
      </svg>
    );
  };

  return (
    <div ref={containerRef} style={{ width: '100%' }} className="portfolio-pnl-chart">
      <Legend />
      <div style={{ height }}>
        {renderChart()}
      </div>
      {summary && snapshots && snapshots.length >= 2 && (
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginTop: '4px',
            padding: '0 4px',
            fontSize: '11px',
            color: catppuccin.subtext0,
          }}
        >
          <span>
            Value: <span style={{ color: catppuccin.text }}>{formatEur(summary.end_value).replace('+', '')}</span>
          </span>
          <span>
            P&L:{' '}
            <span style={{ color: summary.pnl_absolute >= 0 ? catppuccin.green : catppuccin.red }}>
              {formatEur(summary.pnl_absolute)}
            </span>
          </span>
        </div>
      )}
    </div>
  );
}
