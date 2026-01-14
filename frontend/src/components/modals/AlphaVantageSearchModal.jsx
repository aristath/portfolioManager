import { Modal, TextInput, Button, Group, Stack, Table, Text, Loader, Alert, ScrollArea } from '@mantine/core';
import { useSettingsStore } from '../../stores/settingsStore';
import { useState } from 'react';

export function AlphaVantageSearchModal({ opened, onClose, onSelectSymbol }) {
  const { settings } = useSettingsStore();
  const [searchQuery, setSearchQuery] = useState('');
  const [results, setResults] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Get next API key using cycling logic
  const getNextApiKey = () => {
    const apiKeysString = settings.alphavantage_api_key || '';
    if (!apiKeysString.trim()) {
      return null;
    }

    const keys = apiKeysString.split(',').map(k => k.trim()).filter(k => k);
    if (keys.length === 0) {
      return null;
    }

    // Get last used index from localStorage
    const lastIndexKey = 'alphavantage_last_key_index';
    let lastIndex = -1; // -1 means no previous index (first call)
    try {
      const stored = localStorage.getItem(lastIndexKey);
      if (stored !== null) {
        lastIndex = parseInt(stored, 10);
        if (isNaN(lastIndex)) {
          lastIndex = -1;
        }
      }
    } catch (e) {
      // localStorage not available or error, use first key
      return keys[0];
    }

    // Calculate next index (increment and wrap around)
    // If lastIndex is -1 (first call), use 0. Otherwise use (lastIndex + 1) % keys.length
    const nextIndex = lastIndex === -1 ? 0 : (lastIndex + 1) % keys.length;

    // Store new index
    try {
      localStorage.setItem(lastIndexKey, nextIndex.toString());
    } catch (e) {
      // localStorage quota exceeded or error, continue anyway
    }

    return keys[nextIndex];
  };

  // Get cached response or make API call
  const searchSymbols = async () => {
    if (!searchQuery.trim()) {
      setError('Please enter a search query');
      return;
    }

    // Build cache key from search query only (not API key)
    const normalizedQuery = searchQuery.trim().toLowerCase();
    const cacheKey = `alphavantage_cache_SYMBOL_SEARCH_${normalizedQuery}`;

    // Check cache first
    try {
      const cached = localStorage.getItem(cacheKey);
      if (cached) {
        const cachedData = JSON.parse(cached);
        if (cachedData.bestMatches) {
          setResults(cachedData.bestMatches);
          setError(null);
          return;
        }
      }
    } catch (e) {
      // Cache read error, continue with API call
    }

    // Get API key for the request
    const apiKey = getNextApiKey();
    if (!apiKey) {
      setError('Alpha Vantage API key not configured. Please set it in Settings.');
      return;
    }

    const apiUrl = `https://www.alphavantage.co/query?function=SYMBOL_SEARCH&keywords=${encodeURIComponent(searchQuery.trim())}&apikey=${apiKey}`;

    // Make API call
    setLoading(true);
    setError(null);
    try {
      const response = await fetch(apiUrl);
      if (!response.ok) {
        throw new Error(`API request failed with status ${response.status}`);
      }

      const data = await response.json();

      // Check for API error messages
      if (data['Error Message']) {
        throw new Error(data['Error Message']);
      }
      if (data['Note']) {
        throw new Error(data['Note']);
      }

      if (!data.bestMatches || !Array.isArray(data.bestMatches)) {
        setResults([]);
        setError('No results found');
      } else {
        setResults(data.bestMatches);

        // Cache the response (keyed by search query only)
        try {
          localStorage.setItem(cacheKey, JSON.stringify(data));
        } catch (e) {
          // Cache write error (quota exceeded), continue anyway
        }
      }
    } catch (e) {
      console.error('Alpha Vantage API error:', e);
      setError(e.message || 'Failed to search symbols');
      setResults([]);
    } finally {
      setLoading(false);
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !loading) {
      searchSymbols();
    }
  };

  const handleSelectSymbol = (symbol) => {
    onSelectSymbol(symbol);
  };

  return (
    <Modal
      opened={opened}
      onClose={onClose}
      title="Search Alpha Vantage Symbols"
      size="lg"
      centered
    >
      <Stack gap="md">
        <Group gap="xs">
          <TextInput
            placeholder="Enter symbol or company name (e.g., BMW, ASML)"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            style={{ flex: 1 }}
          />
          <Button onClick={searchSymbols} loading={loading} disabled={!searchQuery.trim()}>
            Search
          </Button>
        </Group>

        {error && (
          <Alert color="red" variant="light">
            {error}
          </Alert>
        )}

        {loading && (
          <Group justify="center" py="xl">
            <Loader size="sm" />
            <Text size="sm" c="dimmed">Searching...</Text>
          </Group>
        )}

        {!loading && results.length > 0 && (
          <ScrollArea.Autosize mah={400}>
            <Table highlightOnHover>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Symbol</Table.Th>
                  <Table.Th>Name</Table.Th>
                  <Table.Th>Type</Table.Th>
                  <Table.Th>Region</Table.Th>
                  <Table.Th>Currency</Table.Th>
                  <Table.Th ta="right">Match Score</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {results.map((match) => (
                  <Table.Tr
                    key={match['1. symbol']}
                    style={{ cursor: 'pointer' }}
                    onClick={() => handleSelectSymbol(match['1. symbol'])}
                  >
                    <Table.Td>
                      <Text size="sm" fw={500}>
                        {match['1. symbol']}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" truncate style={{ maxWidth: '200px' }}>
                        {match['2. name']}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" c="dimmed">
                        {match['3. type']}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" c="dimmed">
                        {match['4. region']}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" c="dimmed">
                        {match['8. currency']}
                      </Text>
                    </Table.Td>
                    <Table.Td ta="right">
                      <Text size="sm" c="dimmed">
                        {parseFloat(match['9. matchScore']).toFixed(4)}
                      </Text>
                    </Table.Td>
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>
          </ScrollArea.Autosize>
        )}

        {!loading && !error && results.length === 0 && searchQuery && (
          <Text size="sm" c="dimmed" ta="center" py="xl">
            No results found. Try a different search query.
          </Text>
        )}

        <Group justify="flex-end">
          <Button variant="subtle" onClick={onClose}>
            Close
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}
