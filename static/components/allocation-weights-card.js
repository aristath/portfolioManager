/**
 * Allocation Weights Card Component
 * Container for geographic and industry weight charts
 */
class AllocationWeightsCard extends HTMLElement {
  connectedCallback() {
    this.innerHTML = `
      <div class="bg-gray-800 border border-gray-700 rounded p-3" x-data>
        <h2 class="text-xs text-gray-400 uppercase tracking-wide mb-3">Allocation Weights</h2>
        <div class="space-y-4">
          <geo-chart></geo-chart>
          <industry-chart></industry-chart>
        </div>
      </div>
    `;
  }
}

customElements.define('allocation-weights-card', AllocationWeightsCard);
