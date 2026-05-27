<script>
  import { createEventDispatcher } from 'svelte';
  import { fade, fly } from 'svelte/transition';

  const dispatch = createEventDispatcher();
  const API = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

  let busy = false;
  let error = '';
  let username = 'demo_staff@example.com';
  let password = 'staff123';
  let gateway = 'floor';

  $: scopeText =
    gateway === 'floor'
      ? 'Scope: Call, Seat, Recalculate Queue'
      : 'Scope: View Historical Analytics, Override Pacing Multipliers';

  async function apiFetch(path, opts = {}) {
    const res = await fetch(`${API}${path}`, {
      headers: { 'Content-Type': 'application/json', ...(opts.headers || {}) },
      ...opts
    });
    const body = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(body.error || 'Request failed');
    return body;
  }

  async function doLogin() {
    error = '';
    busy = true;

    try {
      const payload = await apiFetch('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email: username, password })
      });

      dispatch('loggedin', { token: payload.data.token, user: payload.data.user });
    } catch (e) {
      error = e.message;
    } finally {
      busy = false;
    }
  }
</script>

<svelte:head>
  <title>Restaurant Portal Authentication</title>
</svelte:head>

<div class="login-shell">
  <div class="atmosphere" aria-hidden="true"></div>
  <div class="background-block one" aria-hidden="true"></div>
  <div class="background-block two" aria-hidden="true"></div>

  <button class="back-action" on:click={() => dispatch('back')} in:fly={{ x: -16, duration: 200 }}>
    BACK TO LANDING
  </button>

  <section class="login-card" in:fade={{ duration: 220 }}>
    <div class="login-header">
      <div class="card-tag">SECURE EDGE GATEWAY</div>
      <h1>RESTAURANT PORTAL AUTHENTICATION</h1>
      <p>
        Operational access is segmented by role context before credential exchange so floor actions
        and override controls stay explicit.
      </p>
    </div>

    <div class="form-grid">
      <label class="field">
        <span>USERNAME / EMAIL</span>
        <input bind:value={username} type="email" placeholder="ops@restaurant.com" />
      </label>

      <label class="field">
        <span>PASSWORD</span>
        <input bind:value={password} type="password" placeholder="ENTER PASSWORD" />
        <small>Session tokens expire strictly at 15 minutes with rolling cryptographic rotation.</small>
      </label>
    </div>

    <div class="selector-block">
      <div class="selector-title">RBAC GATEWAY CONTEXT</div>
      <div class="selector-grid">
        <label class:active={gateway === 'floor'}>
          <input bind:group={gateway} type="radio" name="gateway" value="floor" />
          <span>FLOOR STAFF</span>
          <strong>Scope: Call, Seat, Recalculate Queue</strong>
        </label>

        <label class:active={gateway === 'manager'}>
          <input bind:group={gateway} type="radio" name="gateway" value="manager" />
          <span>MANAGER</span>
          <strong>Scope: View Historical Analytics, Override Pacing Multipliers</strong>
        </label>
      </div>
      <div class="selector-meta">{scopeText}</div>
    </div>

    <button class="google-action" type="button" disabled>
      CONTINUE WITH GOOGLE
    </button>
    <div class="oauth-note">Google sign-in UI reserved for backend enablement.</div>

    {#if error}
      <div class="error-banner">{error}</div>
    {/if}

    <div class="action-row">
      <button class="login-action" on:click={doLogin} disabled={busy}>
        {busy ? 'AUTHENTICATING...' : 'ENTER PORTAL'}
      </button>
      <button class="create-link" on:click={() => dispatch('switchsignup')} type="button">
        CREATE ACCOUNT
      </button>
      <div class="demo-chip">DEMO: demo_staff@example.com / staff123</div>
    </div>

    <div class="security-badge">
      <div>[x] Composite Leaky Bucket Enabled (IP + Device + Auth ID)</div>
      <div>[x] HMAC Coordinate Spoof Verification Active</div>
    </div>
  </section>
</div>

<style>
  .login-shell {
    min-height: 100vh;
    background: var(--app-background);
    position: relative;
    display: grid;
    place-items: center;
    padding: 76px 24px 24px;
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

  .background-block {
    position: absolute;
    border: 4px solid var(--ink);
    z-index: 0;
  }

  .background-block.one {
    width: 280px;
    height: 280px;
    background: var(--yellow);
    top: 40px;
    left: -40px;
    transform: rotate(-8deg);
  }

  .background-block.two {
    width: 320px;
    height: 220px;
    background: var(--mint);
    right: -40px;
    bottom: 60px;
    transform: rotate(11deg);
  }

  .back-action,
  .login-card {
    position: relative;
    z-index: 1;
  }

  .back-action {
    position: absolute;
    top: 20px;
    left: 20px;
    border: 4px solid var(--ink);
    box-shadow: 6px 6px 0 0 var(--ink);
    background: var(--panel);
    padding: 0.85rem 1rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  .back-action:active,
  .login-action:active {
    transform: translate(4px, 4px);
    box-shadow: 2px 2px 0 0 var(--ink);
  }

  .login-card {
    width: min(760px, 100%);
    border: 4px solid var(--ink);
    box-shadow: 10px 10px 0 0 var(--ink);
    background: rgba(255, 255, 255, 0.92);
    backdrop-filter: blur(8px);
    padding: clamp(22px, 4vw, 34px);
    display: grid;
    gap: 20px;
  }

  .login-header {
    display: grid;
    gap: 12px;
  }

  .card-tag {
    width: fit-content;
    border: 3px solid var(--ink);
    background: var(--yellow);
    padding: 0.45rem 0.7rem;
    font-family: var(--mono);
    font-size: 0.78rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  h1 {
    font-size: clamp(2.2rem, 6vw, 4rem);
    line-height: 0.92;
    letter-spacing: -0.08em;
    max-width: 12ch;
  }

  .form-grid,
  .selector-block {
    display: grid;
    gap: 14px;
  }

  .field {
    display: grid;
    gap: 8px;
    font-weight: 900;
    text-transform: uppercase;
    font-size: 0.85rem;
  }

  .field input {
    border: 4px solid var(--ink);
    background: var(--panel);
    padding: 0.95rem 1rem;
    font-size: 1rem;
    font-weight: 700;
    outline: none;
  }

  .field input:focus {
    background: #fff9c4;
  }

  .field small {
    text-transform: none;
    font-size: 0.82rem;
    font-weight: 700;
  }

  .selector-title,
  .selector-meta {
    font-family: var(--mono);
    font-size: 0.82rem;
    text-transform: uppercase;
    font-weight: 900;
  }

  .selector-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 14px;
  }

  .selector-grid label {
    border: 4px solid var(--ink);
    box-shadow: 6px 6px 0 0 var(--ink);
    background: var(--panel);
    padding: 16px;
    display: grid;
    gap: 8px;
    cursor: pointer;
  }

  .selector-grid label.active {
    background: var(--mint);
    transform: rotate(-1deg);
  }

  .selector-grid input {
    display: none;
  }

  .selector-grid span,
  .selector-grid strong {
    display: block;
  }

  .selector-grid span {
    font-size: 1.1rem;
    font-weight: 900;
  }

  .selector-grid strong {
    font-size: 0.88rem;
    line-height: 1.4;
  }

  .google-action,
  .create-link {
    border: 4px solid var(--ink);
    box-shadow: 8px 8px 0 0 var(--ink);
    padding: 1rem 1.2rem;
    font-weight: 900;
    font-size: 1rem;
    text-transform: uppercase;
  }

  .google-action {
    background: var(--panel);
  }

  .google-action:disabled {
    background: #f3f3f3;
    color: rgba(0, 0, 0, 0.6);
    cursor: not-allowed;
    box-shadow: 6px 6px 0 0 var(--ink);
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

  .error-banner {
    border: 4px solid var(--ink);
    background: #ffb3b3;
    padding: 0.9rem 1rem;
    font-weight: 900;
  }

  .action-row {
    display: flex;
    flex-wrap: wrap;
    gap: 14px;
    align-items: center;
  }

  .login-action {
    border: 4px solid var(--ink);
    box-shadow: 8px 8px 0 0 var(--ink);
    background: var(--yellow);
    padding: 1rem 1.2rem;
    font-weight: 900;
    font-size: 1rem;
    text-transform: uppercase;
  }

  .login-action:disabled {
    opacity: 0.7;
    cursor: wait;
  }

  .create-link {
    background: var(--mint);
  }

  .google-action:active,
  .create-link:active {
    transform: translate(4px, 4px);
    box-shadow: 2px 2px 0 0 var(--ink);
  }

  .demo-chip {
    border: 3px solid var(--ink);
    background: var(--panel);
    padding: 0.8rem 1rem;
    font-family: var(--mono);
    font-size: 0.84rem;
    font-weight: 700;
  }

  .security-badge {
    border: 4px solid var(--ink);
    box-shadow: 8px 8px 0 0 var(--ink);
    background: var(--mint);
    padding: 16px;
    display: grid;
    gap: 10px;
    font-family: var(--mono);
    font-size: 0.86rem;
    font-weight: 900;
  }

  @media (max-width: 700px) {
    .selector-grid {
      grid-template-columns: 1fr;
    }

    .back-action {
      position: static;
      margin-bottom: 16px;
      justify-self: start;
    }

    .login-shell {
      padding: 70px 16px 16px;
      align-content: start;
    }
  }
</style>
