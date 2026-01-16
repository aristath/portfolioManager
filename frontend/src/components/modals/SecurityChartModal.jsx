/**
 * Security Chart Modal Component
 * 
 * Modal wrapper for the SecurityChart component.
 * Displays a full-screen price chart for a selected security.
 * 
 * Features:
 * - Full-screen chart display
 * - Opens when a security is selected from the security table
 * - Passes ISIN and symbol to SecurityChart component
 * 
 * Used for detailed price analysis of individual securities.
 */
import { useRef } from 'react';
import { Modal } from '@mantine/core';
import { SecurityChart } from '../charts/SecurityChart';
import { useAppStore } from '../../stores/appStore';

/**
 * Security chart modal component
 * 
 * Wrapper modal for displaying security price charts.
 * 
 * @returns {JSX.Element} Security chart modal with embedded chart component
 */
export function SecurityChartModal() {
  const { showSecurityChart, selectedSecuritySymbol, selectedSecurityIsin, closeSecurityChartModal } = useAppStore();
  const chartRef = useRef(null);

  /**
   * Closes the modal
   */
  const closeModal = () => {
    closeSecurityChartModal();
  };

  return (
    <Modal
      className="security-chart-modal"
      opened={showSecurityChart}
      onClose={closeModal}
      title="Security Chart"
      size="xl"
    >
      {/* Render chart only if security is selected */}
      {selectedSecurityIsin && (
        <SecurityChart
          className="security-chart-modal__chart"
          ref={chartRef}
          isin={selectedSecurityIsin}
          symbol={selectedSecuritySymbol}
          onClose={closeModal}
        />
      )}
    </Modal>
  );
}
