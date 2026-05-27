<script>
  import { fade } from 'svelte/transition';
  import Login from './Login.svelte';
  import CreateAccount from './CreateAccount.svelte';

  export let authToken;
  export let currentUser;
  let mode = 'login';
</script>

{#if mode === 'login'}
  <div in:fade={{ duration: 160 }}>
    <Login
      {authToken}
      {currentUser}
      on:back
      on:loggedin
      on:switchsignup={() => (mode = 'signup')}
    />
  </div>
{:else}
  <div in:fade={{ duration: 160 }}>
    <CreateAccount
      on:back={() => (mode = 'login')}
      on:switchlogin={() => (mode = 'login')}
      on:loggedin
    />
  </div>
{/if}
