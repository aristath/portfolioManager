/**
 * R2 Backup Modal Component
 * 
 * Modal for managing Cloudflare R2 database backups.
 * Displays list of available backups with actions to download, restore, or delete.
 * 
 * Features:
 * - Lists all available R2 backups with metadata (date, size, age)
 * - Download backups to local system
 * - Restore backups (with confirmation dialog)
 * - Delete backups
 * - Refresh backup list
 * - Formatted file sizes and timestamps
 * 
 * Used in SettingsModal for backup management.
 */
import { useState, useEffect, useCallback } from 'react';
import { Modal, Table, Button, Group, Text, Alert, ActionIcon, Tooltip, Stack, Badge, Divider } from '@mantine/core';
import { IconDownload, IconTrash, IconRefresh, IconRestore } from '@tabler/icons-react';
import { api } from '../../api/client';
import { useNotifications } from '../../hooks/useNotifications';

/**
 * R2 backup modal component
 * 
 * @param {Object} props - Component props
 * @param {boolean} props.opened - Whether the modal is open
 * @param {function} props.onClose - Callback to close the modal
 * @returns {JSX.Element} R2 backup management modal
 */
export function R2BackupModal({ opened, onClose }) {
  const { showNotification } = useNotifications();
  const [backups, setBackups] = useState([]);
  const [loading, setLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState(null); // Track which action is loading (e.g., 'download-filename.tar.gz')
  const [showRestoreConfirm, setShowRestoreConfirm] = useState(null); // Backup object to restore (null if not confirming)

  /**
   * Fetches list of available R2 backups from the backend
   */
  const fetchBackups = useCallback(async () => {
    setLoading(true);
    try {
      const result = await api.listR2Backups();
      setBackups(result.backups || []);
    } catch (error) {
      showNotification(`Failed to load backups: ${error.message}`, 'error');
      setBackups([]);
    } finally {
      setLoading(false);
    }
  }, [showNotification]);

  // Fetch backups when modal opens
  useEffect(() => {
    if (opened) {
      fetchBackups();
    }
  }, [opened, fetchBackups]);

  /**
   * Handles downloading a backup file
   * 
   * @param {string} filename - Name of the backup file to download
   */
  const handleDownload = async (filename) => {
    setActionLoading(`download-${filename}`);
    try {
      await api.downloadR2Backup(filename);
      showNotification('Backup download started', 'success');
    } catch (error) {
      showNotification(`Failed to download backup: ${error.message}`, 'error');
    } finally {
      setActionLoading(null);
    }
  };

  /**
   * Handles deleting a backup file
   * 
   * @param {string} filename - Name of the backup file to delete
   */
  const handleDelete = async (filename) => {
    if (!confirm(`Are you sure you want to delete backup "${filename}"? This action cannot be undone.`)) {
      return;
    }

    setActionLoading(`delete-${filename}`);
    try {
      await api.deleteR2Backup(filename);
      showNotification('Backup deleted successfully', 'success');
      fetchBackups(); // Refresh list after deletion
    } catch (error) {
      showNotification(`Failed to delete backup: ${error.message}`, 'error');
    } finally {
      setActionLoading(null);
    }
  };

  /**
   * Initiates restore confirmation flow
   * 
   * @param {Object} backup - Backup object to restore
   */
  const handleRestoreClick = (backup) => {
    setShowRestoreConfirm(backup);
  };

  /**
   * Confirms and executes backup restore
   * 
   * Restores the selected backup, which will trigger a system restart.
   * Closes the modal after initiating restore since system will restart.
   */
  const handleRestoreConfirm = async () => {
    if (!showRestoreConfirm) return;

    const filename = showRestoreConfirm.filename;
    setActionLoading(`restore-${filename}`);

    try {
      await api.stageR2Restore(filename);
      showNotification('Restore initiated - system will restart automatically', 'success');
      setShowRestoreConfirm(null);
      onClose(); // Close modal since system is restarting
    } catch (error) {
      showNotification(`Failed to restore backup: ${error.message}`, 'error');
    } finally {
      setActionLoading(null);
    }
  };

  /**
   * Formats byte count to human-readable string (B, KB, MB, GB)
   * 
   * @param {number} bytes - Byte count
   * @returns {string} Formatted size string
   */
  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
  };

  /**
   * Formats timestamp as relative age (e.g., "2 days ago", "5 hours ago")
   * 
   * @param {string|number} timestamp - Timestamp to format
   * @returns {string} Formatted age string
   */
  const formatAge = (timestamp) => {
    const now = new Date();
    const backupDate = new Date(timestamp);
    const diffMs = now - backupDate;
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffHours / 24);

    if (diffDays > 0) {
      return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
    }
    return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
  };

  /**
   * Formats timestamp as locale-specific date/time string
   * 
   * @param {string|number} timestamp - Timestamp to format
   * @returns {string} Formatted date/time string
   */
  const formatTimestamp = (timestamp) => {
    return new Date(timestamp).toLocaleString();
  };

  return (
    <>
      <Modal
        className="backup-modal"
        opened={opened}
        onClose={onClose}
        title="Cloudflare R2 Backups"
        size="xl"
      >
        <Stack className="backup-modal__content" gap="md">
          {/* Backup List View */}
          {!showRestoreConfirm && (
            <>
              {/* Header with backup count and refresh button */}
              <Group className="backup-modal__header" justify="space-between">
                <Text className="backup-modal__count" size="sm" c="dimmed">
                  {backups.length} backup{backups.length !== 1 ? 's' : ''} available
                </Text>
                <Button
                  className="backup-modal__refresh-btn"
                  size="xs"
                  variant="light"
                  leftSection={<IconRefresh size={16} />}
                  onClick={fetchBackups}
                  loading={loading}
                >
                  Refresh
                </Button>
              </Group>

              {/* Empty state when no backups found */}
              {backups.length === 0 && !loading && (
                <Alert className="backup-modal__empty-alert" color="blue">
                  <Text className="backup-modal__empty-text" size="sm">No backups found. Create your first backup using the &quot;Backup Now&quot; button in Settings.</Text>
                </Alert>
              )}

              {/* Backup table */}
              {backups.length > 0 && (
                <Table className="backup-modal__table" striped highlightOnHover>
                  <Table.Thead className="backup-modal__table-head">
                    <Table.Tr className="backup-modal__table-header-row">
                      <Table.Th className="backup-modal__table-th backup-modal__table-th--date">Date</Table.Th>
                      <Table.Th className="backup-modal__table-th backup-modal__table-th--age">Age</Table.Th>
                      <Table.Th className="backup-modal__table-th backup-modal__table-th--size">Size</Table.Th>
                      <Table.Th className="backup-modal__table-th backup-modal__table-th--actions" style={{ textAlign: 'right' }}>Actions</Table.Th>
                    </Table.Tr>
                  </Table.Thead>
                  <Table.Tbody className="backup-modal__table-body">
                    {backups.map((backup) => (
                      <Table.Tr className="backup-modal__table-row" key={backup.filename}>
                        {/* Date column with timestamp and filename */}
                        <Table.Td className="backup-modal__table-td backup-modal__table-td--date">
                          <Text className="backup-modal__date" size="sm" fw={500}>{formatTimestamp(backup.timestamp)}</Text>
                          <Text className="backup-modal__filename" size="xs" c="dimmed">{backup.filename}</Text>
                        </Table.Td>
                        {/* Age column with relative time badge */}
                        <Table.Td className="backup-modal__table-td backup-modal__table-td--age">
                          <Badge className="backup-modal__age-badge" size="sm" variant="light">
                            {formatAge(backup.timestamp)}
                          </Badge>
                        </Table.Td>
                        {/* Size column with formatted file size */}
                        <Table.Td className="backup-modal__table-td backup-modal__table-td--size">
                          <Text className="backup-modal__size" size="sm">{formatBytes(backup.size)}</Text>
                        </Table.Td>
                        {/* Actions column with download, restore, and delete buttons */}
                        <Table.Td className="backup-modal__table-td backup-modal__table-td--actions">
                          <Group className="backup-modal__action-buttons" gap="xs" justify="flex-end">
                            <Tooltip label="Download backup">
                              <ActionIcon
                                className="backup-modal__action-btn backup-modal__action-btn--download"
                                variant="light"
                                color="blue"
                                onClick={() => handleDownload(backup.filename)}
                                loading={actionLoading === `download-${backup.filename}`}
                              >
                                <IconDownload size={18} />
                              </ActionIcon>
                            </Tooltip>
                            <Tooltip label="Restore backup">
                              <ActionIcon
                                className="backup-modal__action-btn backup-modal__action-btn--restore"
                                variant="light"
                                color="green"
                                onClick={() => handleRestoreClick(backup)}
                                loading={actionLoading === `restore-${backup.filename}`}
                              >
                                <IconRestore size={18} />
                              </ActionIcon>
                            </Tooltip>
                            <Tooltip label="Delete backup">
                              <ActionIcon
                                className="backup-modal__action-btn backup-modal__action-btn--delete"
                                variant="light"
                                color="red"
                                onClick={() => handleDelete(backup.filename)}
                                loading={actionLoading === `delete-${backup.filename}`}
                              >
                                <IconTrash size={18} />
                              </ActionIcon>
                            </Tooltip>
                          </Group>
                        </Table.Td>
                      </Table.Tr>
                    ))}
                  </Table.Tbody>
                </Table>
              )}
            </>
          )}

          {/* Restore Confirmation View */}
          {showRestoreConfirm && (
            <Stack className="backup-modal__restore-confirm" gap="md">
              <Alert className="backup-modal__restore-alert" color="orange" title="Confirm Restore">
                <Stack className="backup-modal__restore-details" gap="sm">
                  <Text className="backup-modal__restore-text" size="sm">
                    You are about to restore the following backup:
                  </Text>
                  {/* Backup filename */}
                  <Text className="backup-modal__restore-filename" size="sm" fw={500}>
                    {showRestoreConfirm.filename}
                  </Text>
                  {/* Backup metadata */}
                  <Text className="backup-modal__restore-created" size="sm" c="dimmed">
                    Created: {formatTimestamp(showRestoreConfirm.timestamp)}
                  </Text>
                  <Text className="backup-modal__restore-size" size="sm" c="dimmed">
                    Size: {formatBytes(showRestoreConfirm.size)}
                  </Text>
                  <Divider className="backup-modal__restore-divider" />
                  {/* Warning message */}
                  <Text className="backup-modal__restore-warning" size="sm" fw={500} c="red">
                    Warning: This will replace all current databases!
                  </Text>
                  {/* Important notes about restore process */}
                  <Text className="backup-modal__restore-note" size="xs" c="dimmed">
                    • Your current databases will be backed up automatically before restore
                  </Text>
                  <Text className="backup-modal__restore-note" size="xs" c="dimmed">
                    • The system will restart automatically
                  </Text>
                  <Text className="backup-modal__restore-note" size="xs" c="dimmed">
                    • You may lose connection briefly during restart
                  </Text>
                  <Text className="backup-modal__restore-note" size="xs" c="dimmed">
                    • Pre-restore backup will be saved in case recovery is needed
                  </Text>
                </Stack>
              </Alert>

              {/* Action buttons */}
              <Group className="backup-modal__restore-actions" justify="flex-end" gap="sm">
                <Button
                  className="backup-modal__restore-cancel-btn"
                  variant="default"
                  onClick={() => setShowRestoreConfirm(null)}
                  disabled={actionLoading}
                >
                  Cancel
                </Button>
                <Button
                  className="backup-modal__restore-confirm-btn"
                  color="orange"
                  onClick={handleRestoreConfirm}
                  loading={actionLoading === `restore-${showRestoreConfirm.filename}`}
                >
                  Confirm Restore
                </Button>
              </Group>
            </Stack>
          )}
        </Stack>
      </Modal>
    </>
  );
}
