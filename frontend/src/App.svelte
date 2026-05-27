<script>
  import { onMount } from 'svelte';
  import HomePage     from './lib/HomePage.svelte';
  import AuthPage     from './lib/AuthPage.svelte';
  import CustomerDash from './lib/CustomerDash.svelte';
  import StaffDash    from './lib/StaffDash.svelte';
  import AdminDash    from './lib/AdminDash.svelte';

  const STORAGE_KEY = 'smartrestro-auth';
  const THEME_KEY = 'smartrestro-theme';

  let page = 'home'; // home | auth | customer | staff | admin
  let authToken = '';
  let currentUser = null;
  let theme = 'light';

  function restoreAuth() {
    try {
      const s = JSON.parse(localStorage.getItem(STORAGE_KEY) || 'null');
      if (s?.token && s?.user) {
        authToken = s.token;
        currentUser = s.user;
        routeByRole(s.user.role);
      }
    } catch(_) {}
  }

  function routeByRole(role) {
    if (!role) return;
    if (role === 'customer')         page = 'customer';
    else if (role === 'restaurant_staff') page = 'staff';
    else                             page = 'admin';
  }

  function onLoggedIn(e) {
    authToken   = e.detail.token;
    currentUser = e.detail.user;
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ token: authToken, user: currentUser }));
    routeByRole(currentUser?.role);
  }

  function onLogout() {
    authToken   = '';
    currentUser = null;
    localStorage.removeItem(STORAGE_KEY);
    page = 'home';
  }

  function applyTheme(nextTheme) {
    theme = nextTheme;
    document.documentElement.dataset.theme = theme;
    localStorage.setItem(THEME_KEY, theme);
  }

  function toggleTheme() {
    applyTheme(theme === 'light' ? 'dark' : 'light');
  }

  onMount(() => {
    const savedTheme = localStorage.getItem(THEME_KEY);
    applyTheme(savedTheme === 'dark' ? 'dark' : 'light');
    restoreAuth();
  });
</script>

<svelte:head>
  <meta
    name="viewport"
    content="width=device-width, initial-scale=1"
  />
</svelte:head>

<button class="theme-toggle" type="button" on:click={toggleTheme} aria-label="Toggle theme">
  <span class="theme-icon">{theme === 'light' ? 'D' : 'L'}</span>
  <span>{theme === 'light' ? 'Dark' : 'Light'}</span>
</button>

{#if page === 'home'}
  <HomePage on:goauth={() => page = 'auth'} />
{:else if page === 'auth'}
  <AuthPage on:loggedin={onLoggedIn} on:back={() => page = 'home'} />
{:else if page === 'customer'}
  <CustomerDash {authToken} {currentUser} on:logout={onLogout} />
{:else if page === 'staff'}
  <StaffDash {authToken} {currentUser} on:logout={onLogout} />
{:else if page === 'admin'}
  <AdminDash {authToken} {currentUser} on:logout={onLogout} />
{/if}

<style>
  .theme-toggle {
    position: fixed;
    top: 18px;
    right: 18px;
    z-index: 1000;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    border: 3px solid var(--ink);
    background: var(--theme-toggle-bg);
    color: var(--ink);
    box-shadow: 5px 5px 0 0 var(--ink);
    padding: 0.65rem 0.85rem;
    font-weight: 900;
    font-size: 0.82rem;
    text-transform: uppercase;
  }

  .theme-toggle:active {
    transform: translate(3px, 3px);
    box-shadow: 2px 2px 0 0 var(--ink);
  }

  .theme-icon {
    font-size: 1rem;
    line-height: 1;
  }

  @media (max-width: 640px) {
    .theme-toggle {
      top: 12px;
      right: 12px;
      padding: 0.55rem 0.7rem;
    }
  }
</style>
