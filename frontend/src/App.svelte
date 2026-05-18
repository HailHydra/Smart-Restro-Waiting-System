<script>
  import { onMount } from 'svelte';
  import HomePage     from './lib/HomePage.svelte';
  import AuthPage     from './lib/AuthPage.svelte';
  import CustomerDash from './lib/CustomerDash.svelte';
  import StaffDash    from './lib/StaffDash.svelte';
  import AdminDash    from './lib/AdminDash.svelte';

  const STORAGE_KEY = 'smartrestro-auth';

  let page = 'home';          // home | auth | customer | staff | admin
  let authToken = '';
  let currentUser = null;

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

  onMount(restoreAuth);
</script>

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
