<script>
  import { onMount, createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  export let authToken = '';
  export let currentUser = null;

  const API = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

  let restaurants = [], selectedId = '';
  let queue = [], summary = null;
  let loading = false, error = '';
  let joinForm = { guest_name:'', party_size:2, phone_number:'' };
  let activeTab = 'queue'; // queue | join

  async function apiFetch(path, opts = {}) {
    const res = await fetch(`${API}${path}`, {
      headers: { 'Content-Type':'application/json', Authorization:`Bearer ${authToken}`, ...(opts.headers||{}) },
      ...opts,
    });
    const body = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(body.error || 'Request failed');
    return body;
  }

  async function loadRestaurants() {
    const r = await apiFetch('/restaurants');
    restaurants = r.data || [];
    if (!selectedId && restaurants.length) selectedId = restaurants[0].id;
    await loadQueue();
  }

  async function loadQueue() {
    if (!selectedId) return;
    loading = true; error = '';
    try {
      const [qr, sr] = await Promise.all([
        apiFetch(`/restaurants/${selectedId}/queue`),
        apiFetch(`/restaurants/${selectedId}/queue-summary`),
      ]);
      queue   = qr.data || [];
      summary = sr.data || null;
    } catch(e) { error = e.message; }
    finally { loading = false; }
  }

  async function advance(entry) {
    const map = { WAITING:'CONFIRMED', CONFIRMED:'ARRIVED', ARRIVED:'SEATED' };
    const next = map[entry.status];
    if (!next) return;
    await apiFetch(`/queue/${entry.id}/status`, { method:'PATCH', body: JSON.stringify({status:next}) });
    await loadQueue();
  }

  async function joinQueue() {
    if (!selectedId || !joinForm.guest_name.trim()) return;
    try {
      await apiFetch(`/restaurants/${selectedId}/queue/join`, {
        method:'POST',
        body: JSON.stringify({ ...joinForm, party_size:Number(joinForm.party_size), source_channel:'staff-app' }),
      });
      joinForm = { guest_name:'', party_size:2, phone_number:'' };
      await loadQueue();
      activeTab = 'queue';
    } catch(e) { error = e.message; }
  }

  function statusColor(s) {
    if (s==='WAITING')   return '#3b82f6';
    if (s==='CONFIRMED') return '#10b981';
    if (s==='ARRIVED')   return '#f59e0b';
    if (s==='SEATED')    return '#8b5cf6';
    return '#64748b';
  }

  onMount(loadRestaurants);

  let refreshInterval = setInterval(loadQueue, 8000);
  import { onDestroy } from 'svelte';
  onDestroy(() => clearInterval(refreshInterval));
</script>

<div class="staff-page">
  <header class="topbar">
    <div class="tl">
      <span class="logo">🍽️</span>
      <span class="title">SmartRestro <span class="badge staff">Staff</span></span>
    </div>
    <div class="tr">
      <select bind:value={selectedId} on:change={loadQueue}>
        {#each restaurants as r}<option value={r.id}>{r.name}</option>{/each}
      </select>
      <span class="uname">👤 {currentUser?.name}</span>
      <button class="logout" on:click={() => dispatch('logout')}>Logout</button>
    </div>
  </header>

  <div class="content">
    <!-- Summary bar -->
    {#if summary}
      <div class="summary-bar">
        <div class="scard"><strong>{summary.active_queue_count ?? 0}</strong><span>Active</span></div>
        <div class="scard"><strong>{summary.waiting_count ?? 0}</strong><span>Waiting</span></div>
        <div class="scard"><strong>{summary.average_wait_minutes ?? 0}m</strong><span>Avg Wait</span></div>
        <div class="scard"><strong>{summary.total_served_today ?? 0}</strong><span>Served Today</span></div>
      </div>
    {/if}

    <!-- Tab nav -->
    <div class="tabs">
      <button class:active={activeTab==='queue'} on:click={() => activeTab='queue'}>📋 Queue</button>
      <button class:active={activeTab==='join'}  on:click={() => activeTab='join'} >➕ Add Guest</button>
    </div>

    {#if error}<div class="err">{error}</div>{/if}

    {#if activeTab === 'queue'}
      <div class="queue-grid">
        {#if loading}
          <p class="loading">Loading queue…</p>
        {:else if queue.length === 0}
          <div class="empty">No active entries. Use "Add Guest" to enqueue.</div>
        {:else}
          {#each queue as entry}
            <div class="qcard">
              <div class="qcard-top">
                <div>
                  <p class="guest-name">{entry.guest_name}</p>
                  <p class="guest-meta">{entry.party_size} pax · pos #{entry.position || '—'}</p>
                </div>
                <span class="status-pill" style="--c:{statusColor(entry.status)}">{entry.status}</span>
              </div>
              <div class="qcard-bottom">
                <span class="wait-est">⏱ {entry.estimated_wait_time ?? '?'} min</span>
                {#if entry.status !== 'SEATED'}
                  <button class="adv-btn" on:click={() => advance(entry)}>
                    Move → {({WAITING:'Confirmed',CONFIRMED:'Arrived',ARRIVED:'Seated'})[entry.status]}
                  </button>
                {:else}
                  <span class="seated-label">✓ Seated</span>
                {/if}
              </div>
            </div>
          {/each}
        {/if}
      </div>
    {:else}
      <div class="join-form">
        <h2>Add Guest to Queue</h2>
        <div class="fg"><label>Guest Name</label><input bind:value={joinForm.guest_name} placeholder="John Doe" /></div>
        <div class="fg"><label>Party Size</label><input type="number" bind:value={joinForm.party_size} min="1" /></div>
        <div class="fg"><label>Phone</label><input bind:value={joinForm.phone_number} placeholder="+91…" /></div>
        <button class="join-btn" on:click={joinQueue}>Add to Queue</button>
      </div>
    {/if}
  </div>
</div>

<style>
.staff-page { min-height:100vh;display:flex;flex-direction:column;padding-top:64px;background:var(--app-background);position:relative; }
.staff-page::before { content:"";position:fixed;inset:0;z-index:0;pointer-events:none;background-image:linear-gradient(45deg,var(--checker-a) 25%,transparent 25%),linear-gradient(-45deg,var(--checker-a) 25%,transparent 25%),linear-gradient(45deg,transparent 75%,var(--checker-b) 75%),linear-gradient(-45deg,transparent 75%,var(--checker-b) 75%);background-position:0 0,0 12px,12px -12px,-12px 0;background-size:24px 24px; }
.staff-page > * { position:relative;z-index:1; }
.topbar { display:flex;align-items:center;justify-content:space-between;padding:14px 24px;background:rgba(255,255,255,0.04);border-bottom:1px solid rgba(255,255,255,0.08);flex-wrap:wrap;gap:12px; }
.tl { display:flex;align-items:center;gap:10px;font-weight:800;font-size:1.1rem; }
.logo { font-size:1.7rem; }
.badge { border-radius:999px;padding:2px 10px;font-size:0.72rem;font-weight:700;margin-left:6px; }
.badge.staff { background:rgba(245,158,11,0.2);color:#fcd34d; }
.tr { display:flex;align-items:center;gap:12px;flex-wrap:wrap; }
.tr select { background:rgba(255,255,255,0.07);border:1px solid rgba(255,255,255,0.12);color:#f1f5f9;border-radius:10px;padding:8px 12px;font-size:0.9rem; }
.tr select option { background:#1e293b; }
.uname { color:#94a3b8;font-size:0.88rem; }
.logout { background:rgba(239,68,68,0.15);border:1px solid rgba(239,68,68,0.3);color:#fca5a5;border-radius:10px;padding:8px 16px;font-size:0.85rem;font-weight:600; }

.content { flex:1;padding:24px;max-width:1100px;margin:0 auto;width:100%; }

.summary-bar { display:grid;grid-template-columns:repeat(4,1fr);gap:16px;margin-bottom:24px; }
.scard { background:rgba(255,255,255,0.05);border:1px solid rgba(255,255,255,0.09);border-radius:16px;padding:20px;text-align:center; }
.scard strong { display:block;font-size:1.8rem;font-weight:800;color:#a5b4fc; }
.scard span { font-size:0.82rem;color:#64748b; }

.tabs { display:flex;gap:8px;margin-bottom:20px; }
.tabs button { border:1px solid rgba(255,255,255,0.1);border-radius:12px;background:transparent;color:#64748b;padding:10px 20px;font-weight:700;font-size:0.92rem; }
.tabs button.active { background:linear-gradient(135deg,#6366f1,#8b5cf6);border-color:transparent;color:white; }

.err { background:rgba(239,68,68,0.12);border:1px solid rgba(239,68,68,0.3);color:#fca5a5;border-radius:12px;padding:12px 16px;margin-bottom:16px;font-size:0.9rem; }

.queue-grid { display:grid;grid-template-columns:repeat(auto-fill,minmax(280px,1fr));gap:16px; }
.loading, .empty { color:#64748b;text-align:center;padding:40px; }
.qcard { background:rgba(255,255,255,0.05);border:1px solid rgba(255,255,255,0.09);border-radius:18px;padding:18px;transition:border-color 0.2s; }
.qcard:hover { border-color:rgba(99,102,241,0.3); }
.qcard-top { display:flex;align-items:flex-start;justify-content:space-between;margin-bottom:12px; }
.guest-name { font-weight:700;font-size:1rem;margin-bottom:4px; }
.guest-meta { font-size:0.82rem;color:#64748b; }
.status-pill { background:color-mix(in srgb,var(--c) 15%,transparent);color:var(--c);border-radius:999px;padding:4px 12px;font-size:0.78rem;font-weight:700;white-space:nowrap; }
.qcard-bottom { display:flex;align-items:center;justify-content:space-between; }
.wait-est { font-size:0.88rem;color:#94a3b8; }
.adv-btn { background:linear-gradient(135deg,#6366f1,#8b5cf6);border:none;border-radius:10px;color:white;padding:8px 14px;font-size:0.82rem;font-weight:700; }
.adv-btn:hover { transform:translateY(-1px); }
.seated-label { color:#10b981;font-size:0.88rem;font-weight:600; }

.join-form { max-width:420px;background:rgba(255,255,255,0.04);border:1px solid rgba(255,255,255,0.09);border-radius:20px;padding:32px;display:flex;flex-direction:column;gap:16px; }
.join-form h2 { font-size:1.3rem;font-weight:800;margin-bottom:4px; }
.fg { display:flex;flex-direction:column;gap:6px; }
.fg label { font-size:0.85rem;color:#94a3b8;font-weight:600; }
.fg input { background:rgba(255,255,255,0.07);border:1px solid rgba(255,255,255,0.12);border-radius:12px;color:#f1f5f9;padding:12px 14px;font-size:0.95rem; }
.join-btn { background:linear-gradient(135deg,#10b981,#059669);border:none;border-radius:14px;color:white;padding:14px;font-weight:800;font-size:1rem; }
.join-btn:hover { transform:translateY(-2px); }

@media(max-width:600px) { .summary-bar { grid-template-columns:1fr 1fr; } }
</style>
