import { useMemo } from 'react';
import { RadarChart } from './RadarChart';
import { usePortfolioStore } from '../../stores/portfolioStore';

/**
 * Allocation Radar Component
 * Displays geography and industry allocations as radar charts
 */
// Helper function to convert weights to percentages
function getTargetPcts(weights, activeItems) {
  let total = 0;
  for (const name of activeItems) {
    const weight = weights[name] || 0;
    total += weight;
  }
  const targets = {};
  for (const name of activeItems) {
    const weight = weights[name] || 0;
    targets[name] = total > 0 ? weight / total : 0;
  }
  return targets;
}

export function AllocationRadar({ type = 'both' }) {
  const { allocation, activeGeographies, activeIndustries } = usePortfolioStore();

  // Calculate geography data
  const geoData = useMemo(() => {
    if (type !== 'geography' && type !== 'both') return null;

    const labels = Array.from(activeGeographies || []);
    if (labels.length === 0 || !allocation.geography || allocation.geography.length === 0) {
      return null;
    }

    const currentData = labels.map(geography => {
      const item = allocation.geography.find(a => a.name === geography);
      return item ? item.current_pct * 100 : 0;
    });

    const weights = {};
    allocation.geography.forEach(a => {
      weights[a.name] = a.target_pct || 0;
    });

    const targetPcts = getTargetPcts(weights, labels);
    const targetData = labels.map(geography => (targetPcts[geography] || 0) * 100);

    const allValues = [...targetData, ...currentData];
    const maxValue = allValues.length > 0 ? Math.max(...allValues) : 100;

    return { labels, targetData, currentData, maxValue };
  }, [type, allocation.geography, activeGeographies]);

  // Calculate industry data
  const industryData = useMemo(() => {
    if (type !== 'industry' && type !== 'both') return null;

    const labels = Array.from(activeIndustries || []);
    if (labels.length === 0 || !allocation.industry || allocation.industry.length === 0) {
      return null;
    }

    const currentData = labels.map(industry => {
      const item = allocation.industry.find(a => a.name === industry);
      return item ? item.current_pct * 100 : 0;
    });

    const weights = {};
    allocation.industry.forEach(a => {
      weights[a.name] = a.target_pct || 0;
    });

    const targetPcts = getTargetPcts(weights, labels);
    const targetData = labels.map(industry => (targetPcts[industry] || 0) * 100);

    const allValues = [...targetData, ...currentData];
    const maxValue = allValues.length > 0 ? Math.max(...allValues) : 100;

    return { labels, targetData, currentData, maxValue };
  }, [type, allocation.industry, activeIndustries]);

  return (
    <div className="allocation-radar">
      {/* Geography Radar */}
      {(type === 'geography' || type === 'both') && geoData && (
        <div className="allocation-radar__geography" style={{ marginBottom: type === 'both' ? '16px' : 0 }}>
          <RadarChart
            labels={geoData.labels}
            targetData={geoData.targetData}
            currentData={geoData.currentData}
            maxValue={geoData.maxValue}
          />
        </div>
      )}

      {/* Industry Radar */}
      {(type === 'industry' || type === 'both') && industryData && (
        <div className="allocation-radar__industry">
          <RadarChart
            labels={industryData.labels}
            targetData={industryData.targetData}
            currentData={industryData.currentData}
            maxValue={industryData.maxValue}
          />
        </div>
      )}
    </div>
  );
}
