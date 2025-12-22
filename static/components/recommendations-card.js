/**
 * Recommendations Card Component
 * Shows top 3 trade recommendations with execute buttons
 */
class RecommendationsCard extends HTMLElement {
  connectedCallback() {
    this.innerHTML = `
      <div class="bg-gray-800 border border-gray-700 rounded p-3" x-data>
        <div class="flex items-center justify-between mb-3">
          <h2 class="text-xs text-gray-400 uppercase tracking-wide">Next Actions</h2>
          <button @click="$store.app.fetchRecommendations()"
                  class="p-1 text-gray-400 hover:text-gray-200 rounded hover:bg-gray-700 transition-colors"
                  :disabled="$store.app.loading.recommendations"
                  title="Refresh recommendations">
            <span x-show="$store.app.loading.recommendations" class="inline-block animate-spin">&#9696;</span>
            <span x-show="!$store.app.loading.recommendations">&#8635;</span>
          </button>
        </div>

        <!-- Loading state -->
        <template x-if="$store.app.loading.recommendations && $store.app.recommendations.length === 0">
          <div class="text-gray-500 text-sm py-4 text-center">Loading recommendations...</div>
        </template>

        <!-- Empty state -->
        <template x-if="!$store.app.loading.recommendations && $store.app.recommendations.length === 0">
          <div class="text-gray-500 text-sm py-4 text-center">No recommendations available</div>
        </template>

        <!-- Recommendations list -->
        <div class="space-y-2">
          <template x-for="(rec, index) in $store.app.recommendations" :key="rec.symbol">
            <div class="bg-gray-900 rounded p-2 border border-gray-700">
              <div class="flex items-start justify-between gap-2">
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2">
                    <span class="text-xs font-mono bg-gray-700 px-1.5 py-0.5 rounded" x-text="'#' + (index + 1)"></span>
                    <span class="font-mono text-blue-400 font-bold" x-text="rec.symbol"></span>
                    <span class="text-xs px-1.5 py-0.5 rounded"
                          :class="rec.geography === 'EU' ? 'bg-blue-900 text-blue-300' : rec.geography === 'US' ? 'bg-green-900 text-green-300' : 'bg-red-900 text-red-300'"
                          x-text="rec.geography"></span>
                  </div>
                  <div class="text-sm text-gray-300 truncate mt-0.5" x-text="rec.name"></div>
                  <div class="text-xs text-gray-500 mt-1" x-text="rec.reason"></div>
                </div>
                <div class="text-right flex-shrink-0">
                  <div class="text-sm font-mono font-bold text-green-400" x-text="'€' + rec.amount.toLocaleString()"></div>
                  <div class="text-xs text-gray-400" x-text="rec.quantity ? rec.quantity + ' @ €' + rec.current_price : ''"></div>
                  <div class="flex items-center justify-end gap-1 mt-1">
                    <span class="text-xs text-gray-500" x-text="rec.current_portfolio_score"></span>
                    <span x-show="rec.score_change > 0" class="text-green-400 text-xs">→</span>
                    <span x-show="rec.score_change < 0" class="text-red-400 text-xs">→</span>
                    <span x-show="rec.score_change === 0" class="text-gray-400 text-xs">→</span>
                    <span class="text-xs" :class="rec.score_change > 0 ? 'text-green-400' : rec.score_change < 0 ? 'text-red-400' : 'text-gray-400'" x-text="rec.new_portfolio_score"></span>
                    <span class="text-xs px-1 rounded"
                          :class="rec.score_change > 0 ? 'bg-green-900/50 text-green-400' : rec.score_change < 0 ? 'bg-red-900/50 text-red-400' : 'bg-gray-700 text-gray-400'"
                          x-text="(rec.score_change > 0 ? '+' : '') + rec.score_change"></span>
                  </div>
                </div>
              </div>
              <button @click="$store.app.executeRecommendation(rec.symbol)"
                      class="w-full mt-2 px-2 py-1.5 text-xs rounded transition-colors"
                      :class="index === 0 ? 'bg-green-600 hover:bg-green-500 text-white' : 'bg-gray-700 hover:bg-gray-600 text-gray-300'"
                      :disabled="$store.app.loading.execute || !rec.quantity">
                <span x-show="$store.app.executingSymbol === rec.symbol" class="inline-block animate-spin mr-1">&#9696;</span>
                <span x-text="index === 0 ? 'Execute Now' : 'Execute'"></span>
              </button>
            </div>
          </template>
        </div>
      </div>
    `;
  }
}

customElements.define('recommendations-card', RecommendationsCard);
