<script lang="ts">
  import Counter from './lib/Counter.svelte'

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

  // Use 'redirect: "follow"' to force fetch() to process redirects (which it doesn't by default). Alternatively, use traditional HTML form submission to get redirects to work - HTML forms cause browsers to follow redirects, and don't rely on JS, but this means less capacity for responsive feedback to the user upon error
  async function logOutUser() {
    let res = await fetch('/log-out', {
      method: 'POST',
      credentials: 'include',
      redirect: 'follow'
    })

    if (res.redirected) {
      // Get the redirect URL from the Location header
      let redirectUrl = res.url;
      // Redirect the user manually
      window.location.href = redirectUrl;
      return;
    }

    // Check if response is OK (status 200-299)
    if (res.ok) {
      console.log('Successfully logged out.');
    } else {
      // If it's not a redirect or successful response, check for JSON error
      let error = await res.json();
      console.error('Logout error:', error.message);
      alert('Logout failed: ' + error.message);
    }
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

  <div class="card">
    <button onclick={logOutUser}>Log Out</button>
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