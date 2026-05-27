<script>
  import { onMount, createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  export let authToken = '';
  export let currentUser = null;

  const API    = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
  const ML_API = 'http://localhost:8000';

  let restaurants = [], activeTab = 'overview';
  let allQueues = [];      // flat list of all entries across restaurants
  let summaries = {};      // id → summary
  let agentHealth = null;
  let loading = false;

  async function apiFetch(path, opts = {}) {
    const res = await fetch(`${API}${path}`, {
      headers: { 'Content-Type':'application/json', Authorization:`Bearer ${authToken}`, ...(opts.headers||{}) },
      ...opts,
    });
    const body = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(body.error || 'Request failed');
    return body;
  }

  async function loadAll() {
    loading = true;
    try {
      const r = await apiFetch('/restaurants');
      restaurants = r.data || [];

      const results = await Promise.allSettled(
        restaurants.map(async rest => {
          const [qr, sr] = await Promise.all([
            apiFetch(`/restaurants/${rest.id}/queue`),
            apiFetch(`/restaurants/${rest.id}/queue-summary`),
          ]);
          return { id: rest.id, queue: qr.data||[], summary: sr.data||null };
        })
      );

      allQueues = [];
      results.forEach(res => {
        if (res.status === 'fulfilled') {
          const { id, queue, summary } = res.value;
          summaries[id] = summary;
          allQueues.push(...queue.map(e => ({ ...e, restaurant_id: id })));
        }
      });

      // Check ML agent health
      agentHealth = await fetch(`${ML_API}/health`).then(r=>r.json()).catch(()=>null);
    } finally { loading = false; }
  }

  // Totals across all restaurants
  $: totalActive  = Object.values(summaries).reduce((s,v) => s + (v?.active_queue_count||0), 0);
  $: totalWaiting = Object.values(summaries).reduce((s,v) => s + (v?.waiting_count||0), 0);
  $: totalServed  = Object.values(summaries).reduce((s,v) => s + (v?.total_served_today||0), 0);
  $: avgWait = (() => {
    const counts = Object.values(summaries).filter(v=>v?.average_wait_minutes);
    if (!counts.length) return 0;
    return Math.round(counts.reduce((s,v)=>s+(v.average_wait_minutes||0),0)/counts.length);
  })();

  onMount(loadAll);
</script>

<div class="admin-page">
  <header class="topbar">
    <div class="tl">
      <span class="logo">🍽️</span>
      <span class="title">SmartRestro <span class="badge admin">Admin</span></span>
    </div>
    <div class="tr">
      <button class="refresh-btn" on:click={loadAll} disabled={loading}>🔄 Refresh</button>
      <span class="uname">👤 {currentUser?.name}</span>
      <button class="logout" on:click={() => dispatch('logout')}>Logout</button>
    </div>
  </header>

  <!-- Global KPIs -->
  <div class="kpi-strip">
    <div class="kpi"><span class="kpi-val">{restaurants.length}</span><span class="kpi-lbl">Restaurants</span></div>
    <div class="kpi accent"><span class="kpi-val">{totalActive}</span><span class="kpi-lbl">Active Queues</span></div>
    <div class="kpi"><span class="kpi-val">{totalWaiting}</span><span class="kpi-lbl">Waiting Groups</span></div>
    <div class="kpi green"><span class="kpi-val">{totalServed}</span><span class="kpi-lbl">Served Today</span></div>
    <div class="kpi orange"><span class="kpi-val">{avgWait}m</span><span class="kpi-lbl">Avg Wait</span></div>
    <div class="kpi purple"><span class="kpi-val">{allQueues.length}</span><span class="kpi-lbl">Total Entries</span></div>
  </div>

  <!-- Tabs -->
  <div class="tabs-bar">
    <button class:active={activeTab==='overview'} on:click={()=>activeTab='overview'}>🏠 Overview</button>
    <button class:active={activeTab==='queues'}   on:click={()=>activeTab='queues'}>📋 All Queues</button>
    <button class:active={activeTab==='agent'}    on:click={()=>activeTab='agent'}>🤖 Agent Status</button>
  </div>

  <div class="tab-content">
    {#if activeTab === 'overview'}
      <div class="rest-grid">
        {#each restaurants as rest}
          {@const sum = summaries[rest.id]}
          <div class="rest-card">
            <div class="rest-card-head">
              <div class="rest-dot" style="background:{sum?.active_queue_count > 0 ? '#10b981' : '#475569'}"></div>
              <h3>{rest.name}</h3>
            </div>
            <p class="rest-addr">{rest.address || '—'}</p>
            <div class="rest-stats">
              <div class="rs"><strong>{sum?.active_queue_count ?? '—'}</strong><span>Active</span></div>
              <div class="rs"><strong>{sum?.average_wait_minutes ?? '—'}m</strong><span>Avg Wait</span></div>
              <div class="rs"><strong>{sum?.total_served_today ?? '—'}</strong><span>Served</span></div>
            </div>
            {#if rest.latitude && rest.longitude}
              <p class="coords">📍 {rest.latitude.toFixed(4)}, {rest.longitude.toFixed(4)}</p>
            {/if}
          </div>
        {/each}
      </div>

    {:else if activeTab === 'queues'}
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Guest</th><th>Party</th><th>Status</th><th>Wait</th><th>Position</th><th>Restaurant</th>
            </tr>
          </thead>
          <tbody>
            {#each allQueues as e}
              <tr>
                <td class="bold">{e.guest_name}</td>
                <td>{e.party_size}</td>
                <td><span class="spill" data-s={e.status}>{e.status}</span></td>
                <td>{e.estimated_wait_time ?? '?'}m</td>
                <td>#{e.position ?? '—'}</td>
                <td class="muted">{restaurants.find(r=>r.id===e.restaurant_id)?.name || e.restaurant_id}</td>
              </tr>
            {:else}
              <tr><td colspan="6" class="muted" style="text-align:center;padding:30px">No queue data loaded.</td></tr>
            {/each}
          </tbody>
        </table>
      </div>

    {:else}
      <!-- Agent / ML status -->
      <div class="agent-panel">
        <h2>🤖 ML Agent Status</h2>
        {#if agentHealth}
          <div class="agent-status {agentHealth.status === 'healthy' ? 'ok' : 'fail'}">
            <span class="ah-dot"></span>
            Service: <strong>{agentHealth.status}</strong> · Version: <strong>{agentHealth.version || '?'}</strong>
          </div>
          <div class="agent-details">
            <div class="ad"><span>Groq Enabled</span><strong>{agentHealth.groq_enabled ? '✅ Yes' : '❌ No (set GROQ_API_KEY)'}</strong></div>
            <div class="ad"><span>LangGraph Enabled</span><strong>{agentHealth.langgraph_enabled ? '✅ Yes' : '❌ No'}</strong></div>
            <div class="ad"><span>Pipeline</span><strong>{agentHealth.agent_pipeline || '—'}</strong></div>
          </div>
        {:else}
          <p class="muted-p">ML Engine not reachable at http://localhost:8000 — start it with <code>uvicorn app:app --reload</code></p>
        {/if}

        <div class="geo-explainer">
          <h3>📍 Geo-Coordinate Computation Pipeline</h3>
          <div class="pipeline">
            <div class="pstep"><span class="pnum">1</span><div><b>Browser Geolocation</b><br/><code>navigator.geolocation.getCurrentPosition()</code> → (lat, lng)</div></div>
            <div class="parrow">↓</div>
            <div class="pstep"><span class="pnum">2</span><div><b>Haversine Distance</b><br/><code>d = 2R·arcsin(√(sin²(Δlat/2) + cos(lat₁)·cos(lat₂)·sin²(Δlng/2)))</code><br/>Result: <em>distance_km</em> per restaurant</div></div>
            <div class="parrow">↓</div>
            <div class="pstep"><span class="pnum">3</span><div><b>Agent Node 1 — Enrich</b><br/>No-show probability, occupancy ratio, party-fit penalty computed per candidate</div></div>
            <div class="parrow">↓</div>
            <div class="pstep"><span class="pnum">4</span><div><b>Agent Node 2 — Rank</b><br/><code>score = wait×0.5 + dist×0.25 + occ×0.15 + party×0.05 − noshow×0.05</code><br/>Top-5 sorted by score ascending</div></div>
            <div class="parrow">↓</div>
            <div class="pstep"><span class="pnum">5</span><div><b>Agent Node 3 — LLM Rationale (Groq)</b><br/>llama-3.3-70b writes one sentence explaining each recommendation in plain English</div></div>
            <div class="parrow">↓</div>
            <div class="pstep"><span class="pnum">6</span><div><b>Map Markers + Side Cards</b><br/>Displayed on Leaflet map — pin colour = wait time (green &lt;15m, amber &lt;30m, red ≥30m)</div></div>
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
.admin-page { min-height:100vh;display:flex;flex-direction:column;padding-top:64px;background:var(--app-background);position:relative; }
.admin-page::before { content:"";position:fixed;inset:0;z-index:0;pointer-events:none;background-image:linear-gradient(45deg,var(--checker-a) 25%,transparent 25%),linear-gradient(-45deg,var(--checker-a) 25%,transparent 25%),linear-gradient(45deg,transparent 75%,var(--checker-b) 75%),linear-gradient(-45deg,transparent 75%,var(--checker-b) 75%);background-position:0 0,0 12px,12px -12px,-12px 0;background-size:24px 24px; }
.admin-page > * { position:relative;z-index:1; }
.topbar { display:flex;align-items:center;justify-content:space-between;padding:14px 28px;background:rgba(255,255,255,0.04);border-bottom:1px solid rgba(255,255,255,0.08);flex-wrap:wrap;gap:12px; }
.tl { display:flex;align-items:center;gap:10px;font-weight:800;font-size:1.1rem; }
.logo { font-size:1.7rem; }
.badge { border-radius:999px;padding:2px 10px;font-size:0.72rem;font-weight:700;margin-left:6px; }
.badge.admin { background:rgba(99,102,241,0.25);color:#a5b4fc; }
.tr { display:flex;align-items:center;gap:12px; }
.refresh-btn { background:rgba(255,255,255,0.08);border:1px solid rgba(255,255,255,0.15);color:#94a3b8;border-radius:10px;padding:8px 16px;font-size:0.88rem;font-weight:600; }
.uname { color:#94a3b8;font-size:0.88rem; }
.logout { background:rgba(239,68,68,0.15);border:1px solid rgba(239,68,68,0.3);color:#fca5a5;border-radius:10px;padding:8px 16px;font-size:0.85rem;font-weight:600; }

.kpi-strip { display:flex;gap:1px;background:rgba(255,255,255,0.06);border-bottom:1px solid rgba(255,255,255,0.07); }
.kpi { flex:1;display:flex;flex-direction:column;align-items:center;padding:20px 10px;background:#0a0f1e; }
.kpi.accent { background:rgba(99,102,241,0.08); }
.kpi.green  { background:rgba(16,185,129,0.06); }
.kpi.orange { background:rgba(245,158,11,0.06); }
.kpi.purple { background:rgba(139,92,246,0.06); }
.kpi-val { font-size:2rem;font-weight:900;color:#f1f5f9; }
.kpi-lbl { font-size:0.78rem;color:#64748b;margin-top:2px; }

.tabs-bar { display:flex;gap:4px;padding:16px 28px 0;border-bottom:1px solid rgba(255,255,255,0.07); }
.tabs-bar button { border:none;background:transparent;color:#64748b;padding:10px 20px;font-weight:700;font-size:0.92rem;border-bottom:2px solid transparent;border-radius:0; }
.tabs-bar button.active { color:#a5b4fc;border-bottom-color:#6366f1; }

.tab-content { flex:1;padding:28px;overflow: auto; }

.rest-grid { display:grid;grid-template-columns:repeat(auto-fill,minmax(260px,1fr));gap:20px; }
.rest-card { background:rgba(255,255,255,0.04);border:1px solid rgba(255,255,255,0.08);border-radius:20px;padding:22px;transition:border-color 0.2s; }
.rest-card:hover { border-color:rgba(99,102,241,0.3); }
.rest-card-head { display:flex;align-items:center;gap:10px;margin-bottom:6px; }
.rest-dot { width:10px;height:10px;border-radius:50%;flex-shrink:0; }
.rest-card h3 { font-size:1rem;font-weight:700; }
.rest-addr { font-size:0.82rem;color:#64748b;margin-bottom:14px; }
.rest-stats { display:grid;grid-template-columns:repeat(3,1fr);gap:8px;margin-bottom:10px; }
.rs { text-align:center;background:rgba(255,255,255,0.04);border-radius:12px;padding:10px 6px; }
.rs strong { display:block;font-size:1.1rem;font-weight:800;color:#a5b4fc; }
.rs span { font-size:0.76rem;color:#64748b; }
.coords { font-size:0.78rem;color:#475569; }

.table-wrap { overflow-x:auto; }
table { width:100%;border-collapse:collapse; }
th { text-align:left;padding:10px 14px;font-size:0.8rem;font-weight:700;color:#64748b;border-bottom:1px solid rgba(255,255,255,0.08); }
td { padding:12px 14px;border-bottom:1px solid rgba(255,255,255,0.05);font-size:0.9rem; }
.bold { font-weight:600; }
.muted { color:#64748b; }
.spill { border-radius:999px;padding:3px 10px;font-size:0.78rem;font-weight:700;background:rgba(99,102,241,0.15);color:#a5b4fc; }
.spill[data-s="WAITING"]   { background:rgba(59,130,246,0.15);color:#93c5fd; }
.spill[data-s="CONFIRMED"] { background:rgba(16,185,129,0.15);color:#6ee7b7; }
.spill[data-s="ARRIVED"]   { background:rgba(245,158,11,0.15);color:#fcd34d; }
.spill[data-s="SEATED"]    { background:rgba(139,92,246,0.15);color:#c4b5fd; }

.agent-panel { max-width:760px; }
.agent-panel h2 { font-size:1.6rem;font-weight:800;margin-bottom:16px; }
.agent-status { display:flex;align-items:center;gap:10px;padding:14px 18px;border-radius:14px;font-size:0.95rem;margin-bottom:20px;border:1px solid; }
.agent-status.ok   { background:rgba(16,185,129,0.1);border-color:rgba(16,185,129,0.3);color:#6ee7b7; }
.agent-status.fail { background:rgba(239,68,68,0.1);border-color:rgba(239,68,68,0.3);color:#fca5a5; }
.ah-dot { width:10px;height:10px;border-radius:50%;background:currentColor;flex-shrink:0; }
.agent-details { display:flex;flex-direction:column;gap:10px;margin-bottom:28px; }
.ad { display:flex;justify-content:space-between;align-items:center;background:rgba(255,255,255,0.04);border-radius:12px;padding:12px 16px;font-size:0.9rem; }
.ad span { color:#94a3b8; }
.muted-p { color:#64748b;font-size:0.92rem;margin-bottom:24px; }
.muted-p code { background:rgba(255,255,255,0.08);padding:2px 8px;border-radius:6px; }

.geo-explainer { background:rgba(99,102,241,0.06);border:1px solid rgba(99,102,241,0.2);border-radius:20px;padding:28px; }
.geo-explainer h3 { font-size:1.1rem;font-weight:800;margin-bottom:20px; }
.pipeline { display:flex;flex-direction:column;gap:0; }
.pstep { display:flex;gap:14px;align-items:flex-start; }
.pstep b { display:block;font-weight:700;margin-bottom:4px; }
.pstep div { font-size:0.88rem;color:#94a3b8;line-height:1.6; }
.pstep code { background:rgba(99,102,241,0.15);padding:2px 6px;border-radius:6px;font-family:monospace;font-size:0.82em;color:#c4b5fd; }
.pnum { min-width:28px;height:28px;border-radius:50%;background:linear-gradient(135deg,#6366f1,#8b5cf6);display:flex;align-items:center;justify-content:center;font-weight:800;font-size:0.85rem;flex-shrink:0;margin-top:2px; }
.parrow { color:#6366f1;font-size:1.2rem;padding:4px 0 4px 13px; }

@media(max-width:700px) {
  .kpi-strip { display:grid;grid-template-columns:repeat(3,1fr); }
  .kpi { padding:14px 6px; }
  .kpi-val { font-size:1.5rem; }
}
</style>
