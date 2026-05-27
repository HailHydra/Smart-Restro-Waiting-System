<script>
  import { createEventDispatcher } from 'svelte';
  import { fade, fly } from 'svelte/transition';

  const dispatch = createEventDispatcher();
  const API = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

  let busy = false;
  let error = '';
  let info = '';
  let form = {
    name: '',
    email: '',
    password: '',
    role: 'restaurant_staff'
  };

  async function apiFetch(path, opts = {}) {
    const res = await fetch(`${API}${path}`, {
      headers: { 'Content-Type': 'application/json', ...(opts.headers || {}) },
      ...opts
    });
    const body = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(body.error || 'Request failed');
    return body;
  }

  async function doSignup() {
    error = '';
    info = '';
    busy = true;

    try {
      const payload = await apiFetch('/auth/signup', {
        method: 'POST',
        body: JSON.stringify(form)
      });

      if (payload.data?.requires_confirmation) {
        info = payload.data.message || 'Account created. Check your email to confirm access.';
        return;
      }

      dispatch('loggedin', { token: payload.data.token, user: payload.data.user });
    } catch (e) {
      error = e.message;
    } finally {
      busy = false;
    }
  }
</script>

<svelte:head>
  <title>Create Account | Smart Restro</title>
</svelte:head>

<div class="signup-shell">
  <div class="atmosphere" aria-hidden="true"></div>
  <div class="backdrop one" aria-hidden="true"></div>
  <div class="backdrop two" aria-hidden="true"></div>

  <button class="back-action" on:click={() => dispatch('back')} in:fly={{ x: -16, duration: 200 }}>
    BACK TO LOGIN
  </button>

  <section class="signup-card" in:fade={{ duration: 220 }}>
    <div class="header">
      <div class="tag">ACCOUNT PROVISIONING</div>
      <h1>CREATE YOUR SMART RESTRO ACCOUNT</h1>
      <p>
        Set up a staff or manager account cleanly, then continue with credentials or Google OAuth
        once your identity provider endpoint is configured.
      </p>
    </div>

    <button class="google-action" type="button" disabled>
      CONTINUE WITH GOOGLE
    </button>
    <div class="oauth-note">Google sign-in UI reserved for backend enablement.</div>

    <div class="divider"><span>OR CREATE WITH EMAIL</span></div>

    <div class="form-grid">
      <label class="field">
        <span>FULL NAME</span>
        <input bind:value={form.name} type="text" placeholder="JANE SMITH" />
      </label>

      <label class="field">
        <span>WORK EMAIL</span>
        <input bind:value={form.email} type="email" placeholder="ops@restaurant.com" />
      </label>

      <label class="field">
        <span>PASSWORD</span>
        <input bind:value={form.password} type="password" placeholder="MIN 8 CHARACTERS" />
      </label>

      <label class="field">
        <span>ACCESS PROFILE</span>
        <select bind:value={form.role}>
          <option value="restaurant_staff">Floor Staff</option>
          <option value="restaurant_admin">Manager</option>
          <option value="customer">Diner</option>
        </select>
      </label>
    </div>

    {#if error}
      <div class="banner error">{error}</div>
    {/if}

    {#if info}
      <div class="banner info">{info}</div>
    {/if}

    <div class="action-row">
      <button class="create-action" on:click={doSignup} disabled={busy}>
        {busy ? 'CREATING ACCOUNT...' : 'CREATE ACCOUNT'}
      </button>
      <button class="secondary-action" on:click={() => dispatch('switchlogin')} type="button">
        I ALREADY HAVE AN ACCOUNT
      </button>
    </div>
  </section>
</div>

<style>
  .signup-shell {
    min-height: 100vh;
    background: var(--app-background);
    display: grid;
    place-items: center;
    padding: 76px 24px 24px;
    position: relative;
    overflow: hidden;
  }

  .atmosphere {
    position: fixed;
    inset: 0;
    background-image:
      linear-gradient(45deg, var(--checker-a) 25%, transparent 25%),
      linear-gradient(-45deg, var(--checker-a) 25%, transparent 25%),
      linear-gradient(45deg, transparent 75%, var(--checker-b) 75%),
      linear-gradient(-45deg, transparent 75%, var(--checker-b) 75%);
    background-position: 0 0, 0 12px, 12px -12px, -12px 0;
    background-size: 24px 24px;
    pointer-events: none;
  }

  .backdrop {
    position: absolute;
    border: 4px solid var(--ink);
    z-index: 0;
  }

  .backdrop.one {
    width: 240px;
    height: 240px;
    background: var(--mint);
    top: 56px;
    right: -32px;
    transform: rotate(8deg);
  }

  .backdrop.two {
    width: 280px;
    height: 180px;
    background: var(--yellow);
    left: -42px;
    bottom: 70px;
    transform: rotate(-10deg);
  }

  .back-action,
  .signup-card {
    position: relative;
    z-index: 1;
  }

  .back-action,
  .google-action,
  .create-action,
  .secondary-action {
    border: 4px solid var(--ink);
    font-weight: 900;
    text-transform: uppercase;
  }

  .back-action {
    position: absolute;
    top: 20px;
    left: 20px;
    box-shadow: 6px 6px 0 0 var(--ink);
    background: var(--panel);
    padding: 0.85rem 1rem;
  }

  .signup-card {
    width: min(760px, 100%);
    background: rgba(255, 255, 255, 0.92);
    border: 4px solid var(--ink);
    box-shadow: 10px 10px 0 0 var(--ink);
    backdrop-filter: blur(8px);
    padding: clamp(22px, 4vw, 34px);
    display: grid;
    gap: 18px;
  }

  .header {
    display: grid;
    gap: 12px;
  }

  .tag {
    width: fit-content;
    border: 3px solid var(--ink);
    background: var(--lavender);
    padding: 0.45rem 0.7rem;
    font-family: var(--mono);
    font-size: 0.78rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  h1 {
    font-size: clamp(2.1rem, 6vw, 4rem);
    line-height: 0.92;
    letter-spacing: -0.08em;
    max-width: 11ch;
  }

  .google-action {
    background: #f3f3f3;
    box-shadow: 8px 8px 0 0 var(--ink);
    color: rgba(0, 0, 0, 0.6);
    cursor: not-allowed;
    padding: 1rem 1.2rem;
    font-size: 1rem;
  }

  .oauth-note {
    border: 3px solid var(--ink);
    background: #fff9d8;
    padding: 0.75rem 0.9rem;
    font-family: var(--mono);
    font-size: 0.8rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  .divider {
    display: grid;
    place-items: center;
  }

  .divider span {
    border: 3px solid var(--ink);
    background: var(--ash);
    padding: 0.4rem 0.8rem;
    font-family: var(--mono);
    font-size: 0.8rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14px;
  }

  .field {
    display: grid;
    gap: 8px;
    font-size: 0.84rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  .field input,
  .field select {
    border: 4px solid var(--ink);
    background: var(--panel);
    padding: 0.95rem 1rem;
    font-size: 1rem;
    outline: none;
  }

  .field input:focus,
  .field select:focus {
    background: #fff9c4;
  }

  .banner {
    border: 4px solid var(--ink);
    padding: 0.9rem 1rem;
    font-weight: 900;
  }

  .banner.error {
    background: #ffb3b3;
  }

  .banner.info {
    background: var(--mint);
  }

  .action-row {
    display: flex;
    flex-wrap: wrap;
    gap: 14px;
  }

  .create-action {
    background: var(--mint);
    box-shadow: 8px 8px 0 0 var(--ink);
    padding: 1rem 1.2rem;
  }

  .secondary-action {
    background: var(--lavender);
    box-shadow: 8px 8px 0 0 var(--ink);
    padding: 1rem 1.2rem;
  }

  .back-action:active,
  .google-action:active,
  .create-action:active,
  .secondary-action:active {
    transform: translate(4px, 4px);
    box-shadow: 2px 2px 0 0 var(--ink);
  }

  .create-action:disabled {
    opacity: 0.7;
    cursor: wait;
  }

  @media (max-width: 720px) {
    .form-grid {
      grid-template-columns: 1fr;
    }

    .back-action {
      position: static;
      margin-bottom: 16px;
      justify-self: start;
    }

    .signup-shell {
      padding: 70px 16px 16px;
      align-content: start;
    }
  }
</style>
