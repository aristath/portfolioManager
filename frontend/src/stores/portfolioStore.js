import { create } from 'zustand';
import { api } from '../api/client';

export const usePortfolioStore = create((set, get) => ({
  // Portfolio data
  allocation: {
    geography: [],
    industry: [],
    total_value: 0,
    cash_balance: 0,
  },
  alerts: [],
  cashBreakdown: [],

  // Geographies and industries
  geographies: [],
  geographyTargets: {},
  editingGeography: false,
  activeGeographies: [],

  industries: [],
  industryTargets: {},
  editingIndustry: false,
  activeIndustries: [],

  // Loading states
  loading: {
    geographySave: false,
    industrySave: false,
  },

  // Actions
  fetchAllocation: async () => {
    try {
      const data = await api.fetchAllocation();
      set({
        allocation: {
          geography: data.geography || [],
          industry: data.industry || [],
          total_value: data.total_value || 0,
          cash_balance: data.cash_balance || 0,
        },
        alerts: data.alerts || [],
      });
    } catch (e) {
      console.error('Failed to fetch allocation:', e);
    }
  },

  fetchCashBreakdown: async () => {
    try {
      const data = await api.fetchCashBreakdown();
      set({ cashBreakdown: data.balances || [] });
    } catch (e) {
      console.error('Failed to fetch cash breakdown:', e);
    }
  },

  fetchTargets: async () => {
    try {
      // Fetch targets and available options in parallel
      const [targets, availableGeo, availableInd] = await Promise.all([
        api.fetchTargets(),
        api.fetchAvailableGeographies(),
        api.fetchAvailableIndustries(),
      ]);

      // Available geographies/industries come from active securities in the universe
      const activeGeographies = availableGeo.geographies || [];
      const activeIndustries = availableInd.industries || [];

      // Targets are the saved weights
      const geographyTargets = {};
      const industryTargets = {};

      for (const [name, weight] of Object.entries(targets.geography || {})) {
        geographyTargets[name] = weight;
      }
      for (const [name, weight] of Object.entries(targets.industry || {})) {
        industryTargets[name] = weight;
      }

      set({
        geographies: activeGeographies,
        geographyTargets,
        industries: activeIndustries,
        industryTargets,
        activeGeographies,
        activeIndustries,
      });
    } catch (e) {
      console.error('Failed to fetch targets:', e);
    }
  },

  startEditGeography: () => set({ editingGeography: true }),
  cancelEditGeography: () => set({ editingGeography: false }),
  adjustGeographySlider: (name, value) => {
    const { geographyTargets } = get();
    set({ geographyTargets: { ...geographyTargets, [name]: value } });
  },
  saveGeographyTargets: async () => {
    set({ loading: { ...get().loading, geographySave: true } });
    try {
      await api.saveGeographyTargets(get().geographyTargets);
      await get().fetchTargets();
      await get().fetchAllocation();
      set({ editingGeography: false });
      // Notification will be shown via appStore.showMessage if needed
    } catch (e) {
      console.error('Failed to save geography targets:', e);
      throw e; // Re-throw so components can handle it
    } finally {
      // Ensure loading flag is ALWAYS reset, even if state update fails
      try {
        set({ loading: { ...get().loading, geographySave: false } });
      } catch (finallyError) {
        console.error('Failed to reset loading state:', finallyError);
      }
    }
  },

  startEditIndustry: () => set({ editingIndustry: true }),
  cancelEditIndustry: () => set({ editingIndustry: false }),
  adjustIndustrySlider: (name, value) => {
    const { industryTargets } = get();
    set({ industryTargets: { ...industryTargets, [name]: value } });
  },
  saveIndustryTargets: async () => {
    set({ loading: { ...get().loading, industrySave: true } });
    try {
      await api.saveIndustryTargets(get().industryTargets);
      await get().fetchTargets();
      await get().fetchAllocation();
      set({ editingIndustry: false });
      // Notification will be shown via appStore.showMessage if needed
    } catch (e) {
      console.error('Failed to save industry targets:', e);
      throw e; // Re-throw so components can handle it
    } finally {
      // Ensure loading flag is ALWAYS reset, even if state update fails
      try {
        set({ loading: { ...get().loading, industrySave: false } });
      } catch (finallyError) {
        console.error('Failed to reset loading state:', finallyError);
      }
    }
  },

  updateTestCash: async (amount) => {
    try {
      await api.updateSetting('virtual_test_cash', amount);
      // Refresh cash breakdown to reflect the updated value
      await get().fetchCashBreakdown();
      // Also refresh allocation to update total cash balance
      await get().fetchAllocation();
    } catch (e) {
      console.error('Failed to update TEST cash:', e);
      throw e; // Re-throw so components can handle it
    }
  },
}));
