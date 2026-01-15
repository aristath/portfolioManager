import { useRef } from 'react';
import { Modal } from '@mantine/core';
import { SecurityChart } from '../charts/SecurityChart';
import { useAppStore } from '../../stores/appStore';

export function SecurityChartModal() {
  const { showSecurityChart, selectedSecuritySymbol, selectedSecurityIsin, closeSecurityChartModal } = useAppStore();
  const chartRef = useRef(null);

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
