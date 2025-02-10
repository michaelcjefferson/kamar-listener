<script lang="ts">
  import Counter from './lib/Counter.svelte'

  let healthcheckInfo = $state({
    "status": "",
    "system_info": {
      "environment": ""
    }
  });

  async function pingListenerService() {
    let m = await fetch('/healthcheck')
    healthcheckInfo = await m.json()
  }
</script>

<main>
  <h1>KAMAR Refresh Dashboard</h1>

  <div class="healthcheck-card">
    <button onclick={pingListenerService}>Ping Listener Service</button>
    <p><strong>RESPONSE: </strong>{healthcheckInfo.status}</p>
  </div>

  <div class="card">
    <Counter />
  </div>
</main>

<style>
  .logo {
    height: 6em;
    padding: 1.5em;
    will-change: filter;
    transition: filter 300ms;
  }
  .logo:hover {
    filter: drop-shadow(0 0 2em #646cffaa);
  }
  .logo.svelte:hover {
    filter: drop-shadow(0 0 2em #ff3e00aa);
  }
  .read-the-docs {
    color: #888;
  }
</style>