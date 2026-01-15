import { Card, Select, TextInput, NumberInput, Checkbox, Button, Group, Stack, ScrollArea, Text, Code } from '@mantine/core';
import { useLogsStore } from '../../stores/logsStore';
import { useEffect, useRef, useState } from 'react';
import { formatDateTime } from '../../utils/formatters';
import { useDebouncedValue } from '@mantine/hooks';

export function LogsViewer() {
  const {
    entries,
    filterLevel,
    searchQuery,
    lineCount,
    showErrorsOnly,
    autoRefresh,
    loading,
    fetchLogs,
    setFilterLevel,
    setLineCount,
    setShowErrorsOnly,
    setAutoRefresh,
    startAutoRefresh,
    stopAutoRefresh,
  } = useLogsStore();

  const [debouncedSearch] = useDebouncedValue(searchQuery, 300);
  const [autoScroll, setAutoScroll] = useState(true);
  const scrollAreaRef = useRef(null);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs, debouncedSearch, filterLevel, lineCount, showErrorsOnly]);

  useEffect(() => {
    if (autoRefresh) {
      startAutoRefresh();
    } else {
      stopAutoRefresh();
    }
    return () => {
      stopAutoRefresh();
    };
  }, [autoRefresh, startAutoRefresh, stopAutoRefresh]);

  useEffect(() => {
    if (autoScroll && scrollAreaRef.current) {
      // Mantine ScrollArea exposes viewport property
      const scrollElement = scrollAreaRef.current.viewport || scrollAreaRef.current;
      if (scrollElement && typeof scrollElement.scrollTo === 'function') {
        scrollElement.scrollTo({ top: scrollElement.scrollHeight, behavior: 'smooth' });
      }
    }
  }, [entries, autoScroll]);

  const getLevelColor = (level) => {
    switch (level?.toUpperCase()) {
      case 'DEBUG': return 'dimmed';
      case 'INFO': return 'blue';
      case 'WARNING': return 'yellow';
      case 'ERROR': return 'orange';
      case 'CRITICAL': return 'red';
      default: return 'dimmed';
    }
  };

  return (
    <Card className="logs-viewer" p="md" style={{ backgroundColor: 'var(--mantine-color-dark-7)', border: '1px solid var(--mantine-color-dark-6)' }}>
      <Stack className="logs-viewer__content" gap="md">
        {/* Controls */}
        <Card className="logs-viewer__controls" p="md" style={{ backgroundColor: 'var(--mantine-color-dark-8)', border: '1px solid var(--mantine-color-dark-6)' }}>
          <Group className="logs-viewer__filters" gap="md" wrap="wrap">
            <Select
              className="logs-viewer__select logs-viewer__select--level"
              label="Level"
              data={[
                { value: 'all', label: 'All' },
                { value: 'DEBUG', label: 'DEBUG' },
                { value: 'INFO', label: 'INFO' },
                { value: 'WARNING', label: 'WARNING' },
                { value: 'ERROR', label: 'ERROR' },
                { value: 'CRITICAL', label: 'CRITICAL' },
              ]}
              value={filterLevel || 'all'}
              onChange={setFilterLevel}
              size="xs"
              style={{ width: '120px' }}
            />
            <TextInput
              className="logs-viewer__search"
              label="Search"
              placeholder="Search logs..."
              value={searchQuery}
              onChange={(e) => useLogsStore.getState().setSearchQuery(e.target.value)}
              size="xs"
              style={{ flex: 1, minWidth: '150px' }}
            />
            <NumberInput
              className="logs-viewer__lines-input"
              label="Lines"
              value={lineCount}
              onChange={(val) => setLineCount(Number(val))}
              min={50}
              max={1000}
              step={50}
              size="xs"
              style={{ width: '100px' }}
            />
            <Stack className="logs-viewer__checkboxes" gap="xs" mt="xl">
              <Checkbox
                className="logs-viewer__checkbox logs-viewer__checkbox--errors"
                label="Errors Only"
                checked={showErrorsOnly}
                onChange={(e) => setShowErrorsOnly(e.currentTarget.checked)}
                size="xs"
              />
              <Checkbox
                className="logs-viewer__checkbox logs-viewer__checkbox--auto-refresh"
                label="Auto-refresh"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.currentTarget.checked)}
                size="xs"
              />
              <Checkbox
                className="logs-viewer__checkbox logs-viewer__checkbox--auto-scroll"
                label="Auto-scroll"
                checked={autoScroll}
                onChange={(e) => setAutoScroll(e.currentTarget.checked)}
                size="xs"
              />
            </Stack>
          </Group>

          <Group className="logs-viewer__actions" gap="xs" mt="md">
            <Button className="logs-viewer__refresh-btn" size="xs" onClick={fetchLogs} loading={loading}>
              Refresh
            </Button>
            <Text className="logs-viewer__count" size="xs" c="dimmed">
              {entries.length} lines
            </Text>
          </Group>
        </Card>

        {/* Log Entries */}
        <ScrollArea className="logs-viewer__scroll-area" h={600} ref={scrollAreaRef}>
          <Code
            className="logs-viewer__code"
            block
            style={{
              padding: '12px',
              backgroundColor: 'var(--mantine-color-dark-9)',
              border: '1px solid var(--mantine-color-dark-6)',
              fontFamily: 'var(--mantine-font-family)',
            }}
          >
            {entries.length === 0 ? (
              <Text className="logs-viewer__empty" c="dimmed" size="sm" style={{ fontFamily: 'var(--mantine-font-family)' }}>No log entries</Text>
            ) : (
              entries.map((entry, index) => (
                <div className={`logs-viewer__entry logs-viewer__entry--${entry.level?.toLowerCase() || 'unknown'}`} key={index} style={{ marginBottom: '4px' }}>
                  <Text
                    className="logs-viewer__entry-text"
                    size="xs"
                    c={getLevelColor(entry.level)}
                    span
                    style={{ fontFamily: 'var(--mantine-font-family)' }}
                  >
                    [{formatDateTime(entry.timestamp)}] [{entry.level}] {entry.message}
                  </Text>
                </div>
              ))
            )}
          </Code>
        </ScrollArea>
      </Stack>
    </Card>
  );
}
