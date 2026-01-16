/**
 * Add Security Modal Component
 * 
 * Modal dialog for adding a new security to the investment universe.
 * Accepts symbol or ISIN identifier and automatically fetches all required data.
 * 
 * Features:
 * - Identifier input (symbol or ISIN)
 * - Automatic data fetching (name, geography, exchange, industry, currency, ISIN, historical data, score)
 * - Success/error notifications
 * - Loading state during addition
 * 
 * The backend automatically fetches all security data including 10 years of historical data.
 */
import { Modal, TextInput, Button, Group, Alert, Text } from '@mantine/core';
import { notifications } from '@mantine/notifications';
import { useAppStore } from '../../stores/appStore';
import { useSecuritiesStore } from '../../stores/securitiesStore';
import { useState } from 'react';
import { api } from '../../api/client';

/**
 * Add Security modal component
 * 
 * Provides a form to add a new security to the investment universe.
 * 
 * @returns {JSX.Element} Add Security modal dialog
 */
export function AddSecurityModal() {
  const { showAddSecurityModal, closeAddSecurityModal } = useAppStore();
  const { fetchSecurities } = useSecuritiesStore();
  const [identifier, setIdentifier] = useState('');
  const [loading, setLoading] = useState(false);

  /**
   * Handles adding a security to the universe
   * 
   * Calls the API to add the security, then refreshes the securities list.
   * Shows success/error notifications.
   */
  const handleAdd = async () => {
    if (!identifier.trim()) return;

    setLoading(true);
    try {
      // Add security via API (backend fetches all data automatically)
      await api.addSecurityByIdentifier({ identifier: identifier.trim() });
      // Refresh securities list to show new security
      await fetchSecurities();
      // Clear input and close modal
      setIdentifier('');
      closeAddSecurityModal();
      // Show success notification
      notifications.show({
        title: 'Success',
        message: 'Security added successfully',
        color: 'green',
      });
    } catch (e) {
      console.error('Failed to add security:', e);
      // Show error notification
      notifications.show({
        title: 'Error',
        message: `Failed to add security: ${e.message}`,
        color: 'red',
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      className="add-security-modal"
      opened={showAddSecurityModal}
      onClose={closeAddSecurityModal}
      title="Add Security to Universe"
      size="md"
    >
      <div className="add-security-modal__content">
        {/* Identifier input - accepts symbol (e.g., AAPL.US) or ISIN */}
        <TextInput
          className="add-security-modal__input"
          label="Identifier"
          placeholder="e.g., AAPL.US or US0378331005"
          value={identifier}
          onChange={(e) => setIdentifier(e.target.value)}
          mb="md"
          required
        />
        {/* Info alert explaining automatic data fetching */}
        <Alert className="add-security-modal__alert" color="blue" variant="light" mb="md">
          <Text className="add-security-modal__alert-text" size="xs">
            All data will be automatically fetched: name, geography, exchange, industry, currency, ISIN, historical data (10 years), and initial score calculation.
          </Text>
        </Alert>
        {/* Action buttons */}
        <Group className="add-security-modal__actions" justify="flex-end">
          <Button className="add-security-modal__cancel-btn" variant="subtle" onClick={closeAddSecurityModal}>
            Cancel
          </Button>
          {/* Submit button - disabled if identifier is empty */}
          <Button className="add-security-modal__submit-btn" onClick={handleAdd} loading={loading} disabled={!identifier.trim()}>
            Add Security
          </Button>
        </Group>
      </div>
    </Modal>
  );
}
