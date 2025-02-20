<script>
  import Counter from '../lib/Counter.svelte'

  let healthcheckInfo = $state({
    "status": "",
    "system_info": {
      "environment": ""
    }
  });

  async function pingListenerService() {
    let res = await fetch('/healthcheck')
    healthcheckInfo = await res.json()
  }
</script>

<div class="healthcheck-card">
  <button onclick={pingListenerService}>Ping Listener Service</button>
  <p><strong>RESPONSE: </strong>{healthcheckInfo.status}</p>
</div>

<div class="card">
  <Counter />
</div>