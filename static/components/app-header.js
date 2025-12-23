/**
 * App Header Component
 * Displays the application title and Tradernet connection status
 */
class AppHeader extends HTMLElement {
  connectedCallback() {
    this.innerHTML = `
      <header class="flex items-center justify-between py-3 border-b border-gray-800" x-data>
        <div>
          <h1 class="text-xl font-bold text-blue-400">Arduino Trader</h1>
          <p class="text-xs text-gray-500">Automated Portfolio Management</p>
        </div>
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-1.5"
               :class="$store.app.tradernet.connected ? 'text-green-400' : 'text-red-400'">
            <span class="w-2 h-2 rounded-full"
                  :class="$store.app.tradernet.connected ? 'bg-green-500' : 'bg-red-500'"></span>
            <span class="text-xs" x-text="$store.app.tradernet.connected ? 'Tradernet Connected' : 'Tradernet Offline'"></span>
          </div>
          <button @click="$store.app.showSettingsModal = true"
                  class="p-1.5 text-gray-400 hover:text-gray-200 hover:bg-gray-700 rounded transition-colors"
                  title="Settings">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" style="height:1em" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </button>
        </div>
      </header>
    `;
  }
}

customElements.define('app-header', AppHeader);
