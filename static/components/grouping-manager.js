/**
 * Grouping Manager Component
 *
 * Allows users to create custom groups for countries and industries.
 * This replaces hardcoded mappings and scales as the portfolio grows.
 */
class GroupingManager extends HTMLElement {
  connectedCallback() {
    this.innerHTML = `
      <div x-data="groupingManager()" x-init="init()" class="space-y-6">
        <!-- Country Groups -->
        <div>
          <h3 class="text-sm font-medium text-gray-300 mb-3">Country Groups</h3>
          <div class="space-y-3">
            <template x-for="(countries, groupName) in countryGroups" :key="groupName">
              <div class="border border-gray-700 rounded p-3 bg-gray-800">
                <div class="flex items-center justify-between mb-2">
                  <input
                    type="text"
                    :value="groupName"
                    @blur="updateCountryGroupName(groupName, $event.target.value)"
                    class="flex-1 mr-2 px-2 py-1 border border-gray-600 rounded bg-gray-900 text-white text-sm"
                    placeholder="Group name">
                  <button
                    @click="deleteCountryGroup(groupName)"
                    class="px-2 py-1 bg-red-600 hover:bg-red-700 text-white rounded text-xs">
                    Delete
                  </button>
                </div>
                <div class="space-y-2">
                  <template x-for="(country, idx) in countries" :key="idx">
                    <div class="flex items-center gap-2">
                      <select
                        :value="country"
                        @change="updateCountryInGroup(groupName, idx, $event.target.value)"
                        class="flex-1 px-2 py-1 border border-gray-600 rounded bg-gray-900 text-white text-sm">
                        <option value="">-- Select Country --</option>
                        <template x-for="c in availableCountries" :key="c">
                          <option :value="c" :selected="c === country" x-text="c"></option>
                        </template>
                      </select>
                      <button
                        @click="removeCountryFromGroup(groupName, idx)"
                        class="px-2 py-1 bg-gray-600 hover:bg-gray-700 text-white rounded text-xs">
                        Remove
                      </button>
                    </div>
                  </template>
                  <button
                    @click="addCountryToGroup(groupName)"
                    class="text-xs text-blue-400 hover:text-blue-300">
                    + Add Country
                  </button>
                </div>
              </div>
            </template>
            <button
              @click="addNewCountryGroup()"
              class="w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded text-sm">
              + Add Country Group
            </button>
            <p x-show="Object.keys(countryGroups).length === 0" class="text-gray-400 text-sm">
              No country groups defined. Add one to get started.
            </p>
          </div>
        </div>

        <!-- Industry Groups -->
        <div>
          <h3 class="text-sm font-medium text-gray-300 mb-3">Industry Groups</h3>
          <div class="space-y-3">
            <template x-for="(industries, groupName) in industryGroups" :key="groupName">
              <div class="border border-gray-700 rounded p-3 bg-gray-800">
                <div class="flex items-center justify-between mb-2">
                  <input
                    type="text"
                    :value="groupName"
                    @blur="updateIndustryGroupName(groupName, $event.target.value)"
                    class="flex-1 mr-2 px-2 py-1 border border-gray-600 rounded bg-gray-900 text-white text-sm"
                    placeholder="Group name">
                  <button
                    @click="deleteIndustryGroup(groupName)"
                    class="px-2 py-1 bg-red-600 hover:bg-red-700 text-white rounded text-xs">
                    Delete
                  </button>
                </div>
                <div class="space-y-2">
                  <template x-for="(industry, idx) in industries" :key="idx">
                    <div class="flex items-center gap-2">
                      <select
                        :value="industry"
                        @change="updateIndustryInGroup(groupName, idx, $event.target.value)"
                        class="flex-1 px-2 py-1 border border-gray-600 rounded bg-gray-900 text-white text-sm">
                        <option value="">-- Select Industry --</option>
                        <template x-for="i in availableIndustries" :key="i">
                          <option :value="i" :selected="i === industry" x-text="i"></option>
                        </template>
                      </select>
                      <button
                        @click="removeIndustryFromGroup(groupName, idx)"
                        class="px-2 py-1 bg-gray-600 hover:bg-gray-700 text-white rounded text-xs">
                        Remove
                      </button>
                    </div>
                  </template>
                  <button
                    @click="addIndustryToGroup(groupName)"
                    class="text-xs text-blue-400 hover:text-blue-300">
                    + Add Industry
                  </button>
                </div>
              </div>
            </template>
            <button
              @click="addNewIndustryGroup()"
              class="w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded text-sm">
              + Add Industry Group
            </button>
            <p x-show="Object.keys(industryGroups).length === 0" class="text-gray-400 text-sm">
              No industry groups defined. Add one to get started.
            </p>
          </div>
        </div>
      </div>
    `;
  }
}

// Alpine.js component
function groupingManager() {
  return {
    availableCountries: [],
    availableIndustries: [],
    countryGroups: {},
    industryGroups: {},
    loading: false,

    async init() {
      await this.loadData();
    },

    async loadData() {
      this.loading = true;
      try {
        const [countriesRes, industriesRes, countryGroupsRes, industryGroupsRes] = await Promise.all([
          fetch('/api/allocation/groups/available/countries'),
          fetch('/api/allocation/groups/available/industries'),
          fetch('/api/allocation/groups/country'),
          fetch('/api/allocation/groups/industry'),
        ]);

        const countries = await countriesRes.json();
        const industries = await industriesRes.json();
        const countryGroups = await countryGroupsRes.json();
        const industryGroups = await industryGroupsRes.json();

        this.availableCountries = countries.countries || [];
        this.availableIndustries = industries.industries || [];
        this.countryGroups = countryGroups.groups || {};
        this.industryGroups = industryGroups.groups || {};
      } catch (error) {
        console.error('Failed to load grouping data:', error);
        this.showError('Failed to load grouping data');
      } finally {
        this.loading = false;
      }
    },

    async addNewCountryGroup() {
      const groupName = prompt('Enter group name (e.g., EU, US, ASIA):');
      if (!groupName || !groupName.trim()) return;

      await this.saveCountryGroup(groupName.trim(), []);
    },

    async addNewIndustryGroup() {
      const groupName = prompt('Enter group name (e.g., Technology, Industrials):');
      if (!groupName || !groupName.trim()) return;

      await this.saveIndustryGroup(groupName.trim(), []);
    },

    async saveCountryGroup(groupName, countries) {
      try {
        const response = await fetch('/api/allocation/groups/country', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ group_name: groupName, country_names: countries }),
        });

        if (response.ok) {
          await this.loadData();
          this.showSuccess('Country group saved');
        } else {
          const error = await response.json();
          this.showError(error.detail || 'Failed to save country group');
        }
      } catch (error) {
        console.error('Error saving country group:', error);
        this.showError('Failed to save country group');
      }
    },

    async saveIndustryGroup(groupName, industries) {
      try {
        const response = await fetch('/api/allocation/groups/industry', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ group_name: groupName, industry_names: industries }),
        });

        if (response.ok) {
          await this.loadData();
          this.showSuccess('Industry group saved');
        } else {
          const error = await response.json();
          this.showError(error.detail || 'Failed to save industry group');
        }
      } catch (error) {
        console.error('Error saving industry group:', error);
        this.showError('Failed to save industry group');
      }
    },

    async updateCountryGroupName(oldName, newName) {
      if (!newName || !newName.trim() || newName === oldName) return;

      const countries = this.countryGroups[oldName] || [];
      // Delete old group and create new one
      await this.deleteCountryGroup(oldName);
      await this.saveCountryGroup(newName.trim(), countries);
    },

    async updateIndustryGroupName(oldName, newName) {
      if (!newName || !newName.trim() || newName === oldName) return;

      const industries = this.industryGroups[oldName] || [];
      // Delete old group and create new one
      await this.deleteIndustryGroup(oldName);
      await this.saveIndustryGroup(newName.trim(), industries);
    },

    async deleteCountryGroup(groupName) {
      if (!confirm(`Delete country group "${groupName}"?`)) return;

      try {
        const response = await fetch(`/api/allocation/groups/country/${encodeURIComponent(groupName)}`, {
          method: 'DELETE',
        });

        if (response.ok) {
          await this.loadData();
          this.showSuccess('Country group deleted');
        } else {
          this.showError('Failed to delete country group');
        }
      } catch (error) {
        console.error('Error deleting country group:', error);
        this.showError('Failed to delete country group');
      }
    },

    async deleteIndustryGroup(groupName) {
      if (!confirm(`Delete industry group "${groupName}"?`)) return;

      try {
        const response = await fetch(`/api/allocation/groups/industry/${encodeURIComponent(groupName)}`, {
          method: 'DELETE',
        });

        if (response.ok) {
          await this.loadData();
          this.showSuccess('Industry group deleted');
        } else {
          this.showError('Failed to delete industry group');
        }
      } catch (error) {
        console.error('Error deleting industry group:', error);
        this.showError('Failed to delete industry group');
      }
    },

    addCountryToGroup(groupName) {
      // Add empty string to trigger new select dropdown
      const countries = [...(this.countryGroups[groupName] || []), ''];
      this.countryGroups[groupName] = countries;
      // Don't save yet - wait for user to select a country
    },

    addIndustryToGroup(groupName) {
      // Add empty string to trigger new select dropdown
      const industries = [...(this.industryGroups[groupName] || []), ''];
      this.industryGroups[groupName] = industries;
      // Don't save yet - wait for user to select an industry
    },

    async updateCountryInGroup(groupName, index, newCountry) {
      const countries = [...(this.countryGroups[groupName] || [])];

      if (!newCountry) {
        // Remove empty entry if user didn't select anything
        if (countries[index] === '') {
          this.removeCountryFromGroup(groupName, index);
        }
        return;
      }

      // Update the country at this index
      countries[index] = newCountry;
      // Filter out empty strings and remove duplicates
      const uniqueCountries = [...new Set(countries.filter(c => c))];
      await this.saveCountryGroup(groupName, uniqueCountries);
    },

    async updateIndustryInGroup(groupName, index, newIndustry) {
      const industries = [...(this.industryGroups[groupName] || [])];

      if (!newIndustry) {
        // Remove empty entry if user didn't select anything
        if (industries[index] === '') {
          this.removeIndustryFromGroup(groupName, index);
        }
        return;
      }

      // Update the industry at this index
      industries[index] = newIndustry;
      // Filter out empty strings and remove duplicates
      const uniqueIndustries = [...new Set(industries.filter(i => i))];
      await this.saveIndustryGroup(groupName, uniqueIndustries);
    },

    async removeCountryFromGroup(groupName, index) {
      const countries = [...(this.countryGroups[groupName] || [])];
      countries.splice(index, 1);
      await this.saveCountryGroup(groupName, countries);
    },

    async removeIndustryFromGroup(groupName, index) {
      const industries = [...(this.industryGroups[groupName] || [])];
      industries.splice(index, 1);
      await this.saveIndustryGroup(groupName, industries);
    },

    showSuccess(message) {
      // Use the app store's message system if available
      if (window.Alpine && window.Alpine.store && window.Alpine.store('app')) {
        window.Alpine.store('app').message = message;
        window.Alpine.store('app').messageType = 'success';
        setTimeout(() => {
          window.Alpine.store('app').message = '';
        }, 3000);
      } else {
        console.log('Success:', message);
      }
    },

    showError(message) {
      // Use the app store's message system if available
      if (window.Alpine && window.Alpine.store && window.Alpine.store('app')) {
        window.Alpine.store('app').message = message;
        window.Alpine.store('app').messageType = 'error';
        setTimeout(() => {
          window.Alpine.store('app').message = '';
        }, 5000);
      } else {
        console.error('Error:', message);
      }
    },
  };
}

customElements.define('grouping-manager', GroupingManager);
