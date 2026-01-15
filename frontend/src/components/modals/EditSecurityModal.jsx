import { Modal, TextInput, NumberInput, Switch, Button, Group, Stack, Select } from '@mantine/core';
import { notifications } from '@mantine/notifications';
import { useAppStore } from '../../stores/appStore';
import { useSecuritiesStore } from '../../stores/securitiesStore';
import { useState, useEffect } from 'react';
import { api } from '../../api/client';

export function EditSecurityModal() {
  const { showEditSecurityModal, editingSecurity, closeEditSecurityModal } = useAppStore();
  const { fetchSecurities } = useSecuritiesStore();
  const [formData, setFormData] = useState(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (editingSecurity) {
      setFormData({ ...editingSecurity });
    }
  }, [editingSecurity]);

  const handleSave = async () => {
    if (!formData || !formData.isin) return;

    setLoading(true);
    try {
      // Only send editable fields that are in the backend whitelist
      // Filter out undefined/null values, but keep empty strings and false booleans
      const updateData = {};
      const editableFields = [
        'symbol',
        'name',
        'geography',
        'fullExchangeName',
        'industry',
        'product_type',
        'min_lot',
        'allow_buy',
        'allow_sell',
      ];

      editableFields.forEach(field => {
        if (formData[field] !== undefined && formData[field] !== null) {
          updateData[field] = formData[field];
        }
      });

      await api.updateSecurity(formData.isin, updateData);
      await fetchSecurities();
      closeEditSecurityModal();
      notifications.show({
        title: 'Success',
        message: 'Security updated successfully',
        color: 'green',
      });
    } catch (e) {
      console.error('Failed to update security:', e);
      notifications.show({
        title: 'Error',
        message: `Failed to update security: ${e.message}`,
        color: 'red',
      });
    } finally {
      setLoading(false);
    }
  };

  if (!formData) return null;

  return (
    <Modal
      className="edit-security-modal"
      opened={showEditSecurityModal}
      onClose={closeEditSecurityModal}
      title="Edit Security"
      size="md"
    >
      <Stack className="edit-security-modal__content" gap="md">
        <TextInput
          className="edit-security-modal__input edit-security-modal__input--symbol"
          label="Symbol (Tradernet)"
          value={formData.symbol || ''}
          onChange={(e) => setFormData({ ...formData, symbol: e.target.value })}
          description="Tradernet ticker symbol (e.g., ASML.NL, RHM.DE)"
        />
        <TextInput
          className="edit-security-modal__input edit-security-modal__input--name"
          label="Name"
          value={formData.name || ''}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
        />
        <TextInput
          className="edit-security-modal__input edit-security-modal__input--isin"
          label="ISIN"
          value={formData.isin || ''}
          disabled
          description="Unique security identifier (cannot be changed)"
        />
        <TextInput
          className="edit-security-modal__input edit-security-modal__input--geography"
          label="Geography"
          value={formData.geography || ''}
          onChange={(e) => setFormData({ ...formData, geography: e.target.value })}
          placeholder="e.g., EU, US, ASIA (comma-separated for multiple)"
          description="Geographic region(s) where the security operates"
        />
        <TextInput
          className="edit-security-modal__input edit-security-modal__input--exchange"
          label="Exchange"
          value={formData.fullExchangeName || ''}
          onChange={(e) => setFormData({ ...formData, fullExchangeName: e.target.value })}
          placeholder="e.g., NASDAQ, Euronext Amsterdam, XETRA"
          description="Exchange where the security trades"
        />
        <TextInput
          className="edit-security-modal__input edit-security-modal__input--industry"
          label="Industry"
          value={formData.industry || ''}
          onChange={(e) => setFormData({ ...formData, industry: e.target.value })}
          placeholder="e.g., Technology, Healthcare, Financial Services"
          description="Industry classification"
        />
        <Select
          className="edit-security-modal__select edit-security-modal__select--product-type"
          label="Product Type"
          value={formData.product_type || 'UNKNOWN'}
          onChange={(value) => setFormData({ ...formData, product_type: value })}
          data={[
            { value: 'EQUITY', label: 'EQUITY - Individual stocks/shares' },
            { value: 'ETF', label: 'ETF - Exchange Traded Funds' },
            { value: 'MUTUALFUND', label: 'MUTUALFUND - Mutual funds' },
            { value: 'ETC', label: 'ETC - Exchange Traded Commodities' },
            { value: 'CASH', label: 'CASH - Cash positions' },
            { value: 'UNKNOWN', label: 'UNKNOWN - Unknown type' },
          ]}
          description="Product type classification"
        />
        <NumberInput
          className="edit-security-modal__number-input edit-security-modal__number-input--min-lot"
          label="Min Lot Size"
          value={formData.min_lot || 1}
          onChange={(val) => setFormData({ ...formData, min_lot: Number(val) || 1 })}
          min={1}
          step={1}
          description="Minimum shares per trade (e.g., 100 for Japanese securities)"
        />
        <Switch
          className="edit-security-modal__switch edit-security-modal__switch--allow-buy"
          label="Allow Buy"
          checked={formData.allow_buy !== false}
          onChange={(e) => setFormData({ ...formData, allow_buy: e.currentTarget.checked })}
        />
        <Switch
          className="edit-security-modal__switch edit-security-modal__switch--allow-sell"
          label="Allow Sell"
          checked={formData.allow_sell !== false}
          onChange={(e) => setFormData({ ...formData, allow_sell: e.currentTarget.checked })}
        />
        <Group className="edit-security-modal__actions" justify="flex-end">
          <Button className="edit-security-modal__cancel-btn" variant="subtle" onClick={closeEditSecurityModal}>
            Cancel
          </Button>
          <Button className="edit-security-modal__save-btn" onClick={handleSave} loading={loading}>
            Save
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}
