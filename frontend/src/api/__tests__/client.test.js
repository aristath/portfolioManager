import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { api } from '../client';

// Mock fetch globally
const mockFetch = vi.fn();
// eslint-disable-next-line no-undef
global.fetch = mockFetch;

describe('API Client', () => {
  beforeEach(() => {
    mockFetch.mockClear();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('fetchSettings', () => {
    it('should fetch settings successfully', async () => {
      const mockSettings = {
        settings: {
          risk_tolerance: 0.5,
          temperament_aggression: 0.5,
          temperament_patience: 0.5,
          trading_mode: 'research',
        },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockSettings),
      });

      const result = await api.fetchSettings();

      expect(mockFetch).toHaveBeenCalledWith('/api/settings', expect.objectContaining({
        headers: { 'Content-Type': 'application/json' },
        signal: expect.any(AbortSignal),
      }));
      expect(result).toEqual(mockSettings);
    });

    it('should return temperament settings with correct defaults', async () => {
      const mockSettings = {
        settings: {
          risk_tolerance: 0.5,
          temperament_aggression: 0.5,
          temperament_patience: 0.5,
        },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockSettings),
      });

      const result = await api.fetchSettings();

      expect(result.settings.risk_tolerance).toBe(0.5);
      expect(result.settings.temperament_aggression).toBe(0.5);
      expect(result.settings.temperament_patience).toBe(0.5);
    });

    it('should handle fetch error gracefully', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: () => Promise.reject(new Error('No JSON')),
      });

      await expect(api.fetchSettings()).rejects.toThrow('Internal Server Error');
    });
  });

  describe('updateSetting', () => {
    it('should update risk_tolerance as a number', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      });

      await api.updateSetting('risk_tolerance', 0.7);

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/settings/risk_tolerance',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify({ value: 0.7 }),
        })
      );
    });

    it('should update temperament_aggression as a number', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      });

      await api.updateSetting('temperament_aggression', 0.3);

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/settings/temperament_aggression',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify({ value: 0.3 }),
        })
      );
    });

    it('should update temperament_patience as a number', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      });

      await api.updateSetting('temperament_patience', 0.8);

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/settings/temperament_patience',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify({ value: 0.8 }),
        })
      );
    });

    it('should keep string settings as strings', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      });

      await api.updateSetting('trading_mode', 'live');

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/settings/trading_mode',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify({ value: 'live' }),
        })
      );
    });

    it('should handle update error gracefully', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: () => Promise.resolve({ error: 'Invalid value' }),
      });

      await expect(api.updateSetting('risk_tolerance', 2.0)).rejects.toThrow('Invalid value');
    });
  });

  describe('fetchPlannerConfig', () => {
    it('should fetch planner config successfully', async () => {
      const mockConfig = {
        config: {
          enable_batch_generation: true,
          enable_diverse_selection: true,
          diversity_weight: 0.3,
          transaction_cost_fixed: 5.0,
          transaction_cost_percent: 0.001,
          allow_sell: true,
          allow_buy: true,
          optimizer_target_return: 0.11,
          min_cash_reserve: 500.0,
        },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockConfig),
      });

      const result = await api.fetchPlannerConfig();

      expect(mockFetch).toHaveBeenCalledWith('/api/planning/config', expect.objectContaining({
        headers: { 'Content-Type': 'application/json' },
      }));
      expect(result).toEqual(mockConfig);
    });

    it('should return config without legacy name/description fields', async () => {
      const mockConfig = {
        config: {
          enable_batch_generation: true,
          diversity_weight: 0.3,
        },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockConfig),
      });

      const result = await api.fetchPlannerConfig();

      // Config should not require name/description
      expect(result.config.name).toBeUndefined();
      expect(result.config.description).toBeUndefined();
    });

    it('should handle fetch error gracefully', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        json: () => Promise.reject(new Error('No JSON')),
      });

      await expect(api.fetchPlannerConfig()).rejects.toThrow('Not Found');
    });
  });

  describe('updatePlannerConfig', () => {
    it('should update planner config successfully', async () => {
      const config = {
        enable_batch_generation: true,
        diversity_weight: 0.5,
        allow_buy: true,
        allow_sell: true,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      });

      await api.updatePlannerConfig(config, 'ui', 'Updated via UI');

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/planning/config',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify({
            config,
            changed_by: 'ui',
            change_note: 'Updated via UI',
          }),
        })
      );
    });

    it('should update planner config without legacy fields', async () => {
      const config = {
        // No name or description
        enable_batch_generation: true,
        diversity_weight: 0.3,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      });

      await api.updatePlannerConfig(config, 'ui', 'Test update');

      const callBody = JSON.parse(mockFetch.mock.calls[0][1].body);
      expect(callBody.config.name).toBeUndefined();
      expect(callBody.config.description).toBeUndefined();
    });

    it('should handle update error gracefully', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: () => Promise.resolve({ error: 'Invalid configuration' }),
      });

      await expect(
        api.updatePlannerConfig({}, 'ui', 'Test')
      ).rejects.toThrow('Invalid configuration');
    });
  });

  describe('temperament settings integration', () => {
    it('should correctly round-trip temperament values', async () => {
      // First update temperament settings
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      });
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      });
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      });

      // Update all three temperament values
      await api.updateSetting('risk_tolerance', 0.3);
      await api.updateSetting('temperament_aggression', 0.7);
      await api.updateSetting('temperament_patience', 0.9);

      // Verify all three were called with numeric values
      expect(mockFetch).toHaveBeenNthCalledWith(
        1,
        '/api/settings/risk_tolerance',
        expect.objectContaining({
          body: JSON.stringify({ value: 0.3 }),
        })
      );
      expect(mockFetch).toHaveBeenNthCalledWith(
        2,
        '/api/settings/temperament_aggression',
        expect.objectContaining({
          body: JSON.stringify({ value: 0.7 }),
        })
      );
      expect(mockFetch).toHaveBeenNthCalledWith(
        3,
        '/api/settings/temperament_patience',
        expect.objectContaining({
          body: JSON.stringify({ value: 0.9 }),
        })
      );
    });

    it('should handle boundary values for temperament', async () => {
      // Test boundary values (0 and 1)
      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      });

      // Min boundary
      await api.updateSetting('risk_tolerance', 0);
      expect(mockFetch).toHaveBeenLastCalledWith(
        '/api/settings/risk_tolerance',
        expect.objectContaining({
          body: JSON.stringify({ value: 0 }),
        })
      );

      // Max boundary
      await api.updateSetting('risk_tolerance', 1);
      expect(mockFetch).toHaveBeenLastCalledWith(
        '/api/settings/risk_tolerance',
        expect.objectContaining({
          body: JSON.stringify({ value: 1 }),
        })
      );
    });
  });
});
