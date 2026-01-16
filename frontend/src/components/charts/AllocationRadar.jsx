/**
 * Allocation Radar Component
 * 
 * Displays portfolio allocation by geography or industry using radar charts.
 * Shows both current allocation and target allocation for comparison.
 * 
 * Features:
 * - Geography allocation radar chart
 * - Industry allocation radar chart
 * - Current vs target comparison
 * - Normalized target percentages (weights converted to percentages)
 * - Auto-scaling based on data range
 * 
 * Used in GeographyRadarCard and IndustryRadarCard for diversification visualization.
 */
import { useMemo } from 'react';
import { RadarChart } from './RadarChart';
import { usePortfolioStore } from '../../stores/portfolioStore';

/**
 * Converts allocation weights to percentages
 * 
 * Normalizes weights so they sum to 100% for display purposes.
 * Only considers active items (items with non-zero weights).
 * 
 * @param {Object} weights - Object mapping item names to weights
 * @param {Array<string>} activeItems - Array of active item names
 * @returns {Object} Object mapping item names to normalized percentages (0-1)
 */
function getTargetPcts(weights, activeItems) {
  // Calculate total weight for active items
  let total = 0;
  for (const name of activeItems) {
    const weight = weights[name] || 0;
    total += weight;
  }
  
  // Normalize weights to percentages
  const targets = {};
  for (const name of activeItems) {
    const weight = weights[name] || 0;
    targets[name] = total > 0 ? weight / total : 0;
  }
  return targets;
}

/**
 * Allocation radar component
 * 
 * Displays radar charts for geography and/or industry allocation.
 * 
 * @param {Object} props - Component props
 * @param {string} props.type - Chart type: 'geography', 'industry', or 'both' (default)
 * @returns {JSX.Element} Allocation radar charts
 */
export function AllocationRadar({ type = 'both' }) {
  const { allocation, activeGeographies, activeIndustries } = usePortfolioStore();

  // Calculate geography allocation data for radar chart
  const geoData = useMemo(() => {
    if (type !== 'geography' && type !== 'both') return null;

    const labels = Array.from(activeGeographies || []);
    if (labels.length === 0 || !allocation.geography || allocation.geography.length === 0) {
      return null;
    }

    // Extract current allocation percentages (convert to 0-100 scale)
    const currentData = labels.map(geography => {
      const item = allocation.geography.find(a => a.name === geography);
      return item ? item.current_pct * 100 : 0;
    });

    // Extract target weights
    const weights = {};
    allocation.geography.forEach(a => {
      weights[a.name] = a.target_pct || 0;
    });

    // Normalize target weights to percentages
    const targetPcts = getTargetPcts(weights, labels);
    const targetData = labels.map(geography => (targetPcts[geography] || 0) * 100);

    // Calculate max value for chart scaling
    const allValues = [...targetData, ...currentData];
    const maxValue = allValues.length > 0 ? Math.max(...allValues) : 100;

    return { labels, targetData, currentData, maxValue };
  }, [type, allocation.geography, activeGeographies]);

  // Calculate industry allocation data for radar chart
  const industryData = useMemo(() => {
    if (type !== 'industry' && type !== 'both') return null;

    const labels = Array.from(activeIndustries || []);
    if (labels.length === 0 || !allocation.industry || allocation.industry.length === 0) {
      return null;
    }

    // Extract current allocation percentages (convert to 0-100 scale)
    const currentData = labels.map(industry => {
      const item = allocation.industry.find(a => a.name === industry);
      return item ? item.current_pct * 100 : 0;
    });

    // Extract target weights
    const weights = {};
    allocation.industry.forEach(a => {
      weights[a.name] = a.target_pct || 0;
    });

    // Normalize target weights to percentages
    const targetPcts = getTargetPcts(weights, labels);
    const targetData = labels.map(industry => (targetPcts[industry] || 0) * 100);

    // Calculate max value for chart scaling
    const allValues = [...targetData, ...currentData];
    const maxValue = allValues.length > 0 ? Math.max(...allValues) : 100;

    return { labels, targetData, currentData, maxValue };
  }, [type, allocation.industry, activeIndustries]);

  return (
    <div className="allocation-radar">
      {/* Geography Radar - shows allocation by geographic region */}
      {(type === 'geography' || type === 'both') && geoData && (
        <div className="allocation-radar__geography" style={{ marginBottom: type === 'both' ? '16px' : 0 }}>
          <RadarChart
            labels={geoData.labels}
            targetData={geoData.targetData}  // Target allocation percentages
            currentData={geoData.currentData}  // Current allocation percentages
            maxValue={geoData.maxValue}  // Max value for chart scaling
          />
        </div>
      )}

      {/* Industry Radar - shows allocation by industry/sector */}
      {(type === 'industry' || type === 'both') && industryData && (
        <div className="allocation-radar__industry">
          <RadarChart
            labels={industryData.labels}
            targetData={industryData.targetData}  // Target allocation percentages
            currentData={industryData.currentData}  // Current allocation percentages
            maxValue={industryData.maxValue}  // Max value for chart scaling
          />
        </div>
      )}
    </div>
  );
}
