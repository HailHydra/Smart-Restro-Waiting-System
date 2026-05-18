<script>
  import { createEventDispatcher, onMount } from 'svelte';
  const dispatch = createEventDispatcher();

  const API  = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
  let mode   = 'login';   // login | signup
  let error  = '';
  let info   = '';
  let busy   = false;
  let mapEl;

  let login  = { email:'admin@smartrestro.local', password:'admin123' };
  let signup = { name:'', email:'', password:'', role:'customer' };

  // Leaflet decorative background map
  onMount(() => {
    if (!window.L || !mapEl) return;
    const map = window.L.map(mapEl, {
      center: [20.5937, 78.9629],
      zoom: 5,
      zoomControl: false,
      attributionControl: false,
      dragging: false,
      scrollWheelZoom: false,
    });
    window.L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
      maxZoom: 19,
    }).addTo(map);
  });

  async function apiFetch(path, opts = {}) {
    const res = await fetch(`${API}${path}`, {
      headers: { 'Content-Type': 'application/json', ...(opts.headers||{}) },
      ...opts,
    });
    const body = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(body.error || 'Request failed');
    return body;
  }

  async function doLogin() {
    error = ''; info = ''; busy = true;
    try {
      const payload = await apiFetch('/auth/login', { method:'POST', body: JSON.stringify(login) });
      dispatch('loggedin', { token: payload.data.token, user: payload.data.user });
    } catch(e) { error = e.message; }
    finally { busy = false; }
  }

  async function doSignup() {
    error = ''; info = ''; busy = true;
    try {
      const payload = await apiFetch('/auth/signup', { method:'POST', body: JSON.stringify(signup) });
      if (payload.data.requires_confirmation) {
        info = payload.data.message || 'Check your email to confirm.';
        mode = 'login'; return;
      }
      dispatch('loggedin', { token: payload.data.token, user: payload.data.user });
    } catch(e) { error = e.message; }
    finally { busy = false; }
  }
</script>

<div class="auth-page">
  <!-- Full-screen map background -->
  <div class="map-bg" bind:this={mapEl}></div>
  <div class="map-overlay"></div>

  <button class="back-btn" on:click={() => dispatch('back')}>← Back</button>

  <div class="auth-container">
    <!-- Left branding -->
    <div class="auth-brand">
      <div class="brand-icon">🍽️</div>
      <h1>SmartRestro</h1>
      <p>AI-powered queue management with real-time geo recommendations.</p>
      <div class="brand-features">
        <div class="bf">✦ LangGraph Agent Pipeline</div>
        <div class="bf">✦ Groq LLM Rationale</div>
        <div class="bf">✦ Haversine Geo Ranking</div>
        <div class="bf">✦ Live WebSocket Queue</div>
      </div>
    </div>

    <!-- Right form panel -->
    <div class="auth-form-panel">
      <div class="demo-banner">
        <strong>Demo Credentials:</strong>
        <div class="demo-grid">
          <span>Customer:</span> <code>demo_customer@example.com</code> / <code>customer123</code>
          <span>Staff:</span> <code>demo_staff@example.com</code> / <code>staff123</code>
          <span>Admin:</span> <code>demo_admin@example.com</code> / <code>admin123</code>
        </div>
      </div>

      <div class="auth-tabs">
        <button class:active={mode==='login'}  on:click={() => { mode='login';  error=''; }}>Login</button>
        <button class:active={mode==='signup'} on:click={() => { mode='signup'; error=''; }}>Sign Up</button>
      </div>

      {#if error}<div class="banner err">{error}</div>{/if}
      {#if info} <div class="banner ok">{info}</div>{/if}

      {#if mode === 'login'}
        <div class="form-group">
          <label>Email</label>
          <input bind:value={login.email} type="email" placeholder="you@example.com" />
        </div>
        <div class="form-group">
          <label>Password</label>
          <input bind:value={login.password} type="password" placeholder="••••••••" />
        </div>
        <button class="submit-btn" on:click={doLogin} disabled={busy}>
          {busy ? 'Signing in…' : 'Enter Dashboard →'}
        </button>
        <p class="demo-hint">Use credentials from the banner above or sign up.</p>
      {:else}
        <div class="form-group">
          <label>Full Name</label>
          <input bind:value={signup.name} placeholder="Jane Smith" />
        </div>
        <div class="form-group">
          <label>Email</label>
          <input bind:value={signup.email} type="email" placeholder="you@example.com" />
        </div>
        <div class="form-group">
          <label>Password</label>
          <input bind:value={signup.password} type="password" placeholder="Min 8 chars" />
        </div>
        <div class="form-group">
          <label>I am a…</label>
          <select bind:value={signup.role}>
            <option value="customer">Customer</option>
            <option value="restaurant_staff">Restaurant Staff</option>
            <option value="restaurant_admin">Restaurant Admin / Manager</option>
          </select>
        </div>
        <button class="submit-btn" on:click={doSignup} disabled={busy}>
          {busy ? 'Creating…' : 'Create Account →'}
        </button>
      {/if}
    </div>
  </div>
</div>

<style>
.auth-page { position:relative;min-height:100vh;display:flex;align-items:center;justify-content:center;padding:20px; }
.map-bg { position:fixed;inset:0;z-index:0; }
.map-overlay { position:fixed;inset:0;z-index:1;background:linear-gradient(135deg,rgba(10,15,30,0.88),rgba(14,20,40,0.82)); }

.back-btn { position:fixed;top:20px;left:20px;z-index:20;background:rgba(255,255,255,0.08);border:1px solid rgba(255,255,255,0.15);color:#94a3b8;border-radius:12px;padding:10px 18px;font-size:0.9rem;font-weight:600; }
.back-btn:hover { background:rgba(255,255,255,0.14); }

.auth-container { position:relative;z-index:10;display:grid;grid-template-columns:1fr 1fr;gap:0;max-width:900px;width:100%;border-radius:28px;overflow:hidden;box-shadow:0 40px 80px rgba(0,0,0,0.6);border:1px solid rgba(255,255,255,0.1); }

.auth-brand { background:linear-gradient(135deg,#1e1b4b,#312e81);padding:60px 40px;display:flex;flex-direction:column;gap:20px; }
.brand-icon { font-size:3rem; }
.auth-brand h1 { font-size:2.2rem;font-weight:900; }
.auth-brand > p { color:#a5b4fc;line-height:1.7;font-size:1rem; }
.brand-features { display:flex;flex-direction:column;gap:10px;margin-top:auto; }
.bf { color:#c4b5fd;font-size:0.92rem;font-weight:500; }

.auth-form-panel { background:rgba(15,20,40,0.97);padding:50px 40px;display:flex;flex-direction:column;gap:16px; }
.auth-tabs { display:grid;grid-template-columns:1fr 1fr;gap:8px;margin-bottom:4px; }
.auth-tabs button { border:1px solid rgba(255,255,255,0.12);border-radius:12px;background:transparent;color:#64748b;padding:12px;font-weight:700;font-size:0.95rem; }
.auth-tabs button.active { background:linear-gradient(135deg,#6366f1,#8b5cf6);border-color:transparent;color:white; }

.banner { padding:12px 16px;border-radius:12px;font-size:0.9rem; }
.banner.err { background:rgba(239,68,68,0.12);border:1px solid rgba(239,68,68,0.3);color:#fca5a5; }
.banner.ok  { background:rgba(16,185,129,0.1);border:1px solid rgba(16,185,129,0.3);color:#6ee7b7; }

.form-group { display:flex;flex-direction:column;gap:6px; }
.form-group label { font-size:0.85rem;font-weight:600;color:#94a3b8; }
.form-group input, .form-group select { background:rgba(255,255,255,0.06);border:1px solid rgba(255,255,255,0.12);border-radius:12px;color:#f1f5f9;padding:13px 16px;font-size:0.95rem;outline:none; }
.form-group input:focus, .form-group select:focus { border-color:rgba(99,102,241,0.6);background:rgba(99,102,241,0.08); }
.form-group select option { background:#1e293b; }

.submit-btn { background:linear-gradient(135deg,#6366f1,#8b5cf6);border:none;border-radius:14px;color:white;padding:15px;font-size:1rem;font-weight:800;box-shadow:0 6px 24px rgba(99,102,241,0.4);margin-top:4px; }
.submit-btn:hover:not(:disabled) { transform:translateY(-2px);box-shadow:0 10px 30px rgba(99,102,241,0.55); }
.submit-btn:disabled { opacity:0.6;cursor:not-allowed; }

.demo-hint { font-size:0.82rem;color:#475569;text-align:center; }
.demo-hint code { background:rgba(255,255,255,0.08);padding:2px 6px;border-radius:6px; }

.demo-banner {
  background: rgba(99,102,241,0.12);
  border: 1px solid rgba(99,102,241,0.3);
  color: #c4b5fd;
  padding: 14px 18px;
  border-radius: 14px;
  margin-bottom: 20px;
  font-size: 0.85rem;
}
.demo-grid {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 4px 12px;
  margin-top: 8px;
  align-items: center;
}
.demo-grid code {
  background: rgba(255,255,255,0.08);
  padding: 2px 6px;
  border-radius: 6px;
  font-family: monospace;
  color: #a5b4fc;
}

@media(max-width:640px) {
  .auth-container { grid-template-columns:1fr; }
  .auth-brand { padding:40px 28px; }
  .auth-form-panel { padding:36px 24px; }
}
</style>
