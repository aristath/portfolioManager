/**
 * Industry Allocation Component
 * Displays industry allocation bars and allows editing targets
 */
class IndustryChart extends HTMLElement {
  connectedCallback() {
    this.innerHTML = `
      <div class="card" x-data>
        <div class="card__header">
          <h2 class="card__title">Industry Allocation</h2>
          <button x-show="!$store.app.editingIndustry"
                  @click="$store.app.startEditIndustry()"
                  class="card__action card__action--purple">
            Edit Targets
          </button>
        </div>

        <!-- View Mode -->
        <div x-show="!$store.app.editingIndustry">
          <template x-for="ind in $store.app.allocation.industry" :key="ind.name">
            <div class="industry-item">
              <div class="industry-item__header">
                <span x-text="ind.name"></span>
                <span :class="Math.abs(ind.deviation) < 0.05 ? 'industry-item__value--balanced' : 'industry-item__value--unbalanced'">
                  <span x-text="(ind.current_pct * 100).toFixed(1)"></span>% /
                  <span x-text="(ind.target_pct * 100).toFixed(0)"></span>%
                </span>
              </div>
              <div class="progress">
                <div class="progress__bar"
                     :style="'width: ' + Math.min(ind.current_pct / ind.target_pct * 100, 100) + '%'"></div>
              </div>
            </div>
          </template>
        </div>

        <!-- Edit Mode -->
        <div x-show="$store.app.editingIndustry" x-transition>
          <template x-for="name in Object.keys($store.app.industryTargets).sort()" :key="name">
            <div class="slider-control">
              <div class="slider-control__header">
                <span x-text="name"></span>
                <span class="slider-control__value slider-control__value--purple"
                      x-text="$store.app.industryTargets[name] + '%'"></span>
              </div>
              <input type="range" min="0" max="100"
                     :value="$store.app.industryTargets[name]"
                     @input="$store.app.adjustIndustrySlider(name, parseInt($event.target.value))"
                     class="slider slider--purple">
            </div>
          </template>

          <!-- Total -->
          <div class="total-row">
            <span class="total-row__label">Total</span>
            <span class="total-row__value"
                  :class="$store.app.industryTotal === 100 ? 'total-row__value--valid' : 'total-row__value--invalid'"
                  x-text="$store.app.industryTotal + '%'"></span>
          </div>

          <!-- Buttons -->
          <div class="button-row">
            <button @click="$store.app.cancelEditIndustry()" class="btn btn--secondary">
              Cancel
            </button>
            <button @click="$store.app.saveIndustryTargets()"
                    :disabled="$store.app.industryTotal !== 100 || $store.app.loading.industrySave"
                    class="btn btn--purple">
              <span x-show="$store.app.loading.industrySave" class="btn__spinner">&#9696;</span>
              Save
            </button>
          </div>
        </div>
      </div>
    `;
  }
}

customElements.define('industry-chart', IndustryChart);
