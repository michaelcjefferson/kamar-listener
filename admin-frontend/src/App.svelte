<script lang="ts">
  import Home from "./pages/Home.svelte"

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
  <nav>
    <ul>
      <li class="nav-item">
        <button class="nav-link">Home</button>
      </li>
      <li class="nav-item">
        <button class="nav-link">Config</button>
      </li>
      <li class="nav-item">
        <button class="nav-link">Logs</button>
      </li>
    </ul>
  </nav>

  <h1>KAMAR Refresh Dashboard</h1>

  <Home />

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