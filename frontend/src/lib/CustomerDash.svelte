<script>
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  export let authToken = '';
  export let currentUser = null;

  const API     = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
  const ML_API  = 'http://localhost:8000';

  let mapEl, leafletMap, userMarker;
  let mapReady = false;
  let userLat = null, userLng = null;
  let geoError = '';
  let geoStatus = 'Requesting your location…';

  let restaurants = [];
  let recommendations = [];
  let agentRationale = [];
  let agentModel = '';
  let geoComputation = {};
  let loading = false;
  let loadErr = '';
  let partySize = 2;
  let selectedId = '';
  let joinName = '';
  let joinPhone = '';
  let joinMsg = '';

  // restaurant markers
  let markers = [];

  async function apiFetch(path, opts = {}) {
    const res = await fetch(`${API}${path}`, {
      headers: { 'Content-Type':'application/json', Authorization:`Bearer ${authToken}`, ...(opts.headers||{}) },
      ...opts,
    });
    const body = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(body.error || 'Request failed');
    return body;
  }

  // ── Geo ──────────────────────────────────────────────────────────────────
  function haversine(lat1, lng1, lat2, lng2) {
    const R = 6371;
    const dLat = (lat2 - lat1) * Math.PI / 180;
    const dLng = (lng2 - lng1) * Math.PI / 180;
    const a = Math.sin(dLat/2)**2 + Math.cos(lat1*Math.PI/180)*Math.cos(lat2*Math.PI/180)*Math.sin(dLng/2)**2;
    return R * 2 * Math.asin(Math.sqrt(a));
  }

  function requestGeo() {
    if (!navigator.geolocation) { geoError = 'Geolocation not supported.'; geoStatus=''; return; }
    geoStatus = 'Getting GPS fix…';
    navigator.geolocation.getCurrentPosition(
      pos => { userLat = pos.coords.latitude; userLng = pos.coords.longitude; geoStatus=''; onGeo(); },
      err => { geoError = `Location denied: ${err.message}`; geoStatus=''; },
      { enableHighAccuracy: true, timeout: 10000 }
    );
  }

  async function onGeo() {
    if (!leafletMap) return;
    leafletMap.setView([userLat, userLng], 14);
    if (userMarker) userMarker.remove();
    const L = window.L;
    userMarker = L.circleMarker([userLat, userLng], {
      radius: 14, fillColor:'#6366f1', color:'white', weight:3, fillOpacity:0.9,
    }).addTo(leafletMap).bindPopup('<b>📍 You are here</b>');
    await fetchRestaurantsAndRecommend();
  }

  // ── Data ─────────────────────────────────────────────────────────────────
  async function fetchRestaurantsAndRecommend() {
    loading = true; loadErr = '';
    try {
      const res = await apiFetch('/restaurants');
      restaurants = res.data || [];

      // Compute Haversine distance per restaurant
      const candidates = restaurants.map(r => ({
        id: r.id, name: r.name, address: r.address || '',
        latitude: r.latitude || 0, longitude: r.longitude || 0,
        distance_km: (userLat !== null)
          ? haversine(userLat, userLng, r.latitude || 0, r.longitude || 0)
          : 0,
        capacity: r.capacity || 50,
        current_occupancy: r.current_occupancy || 0,
        average_meal_minutes: r.average_meal_minutes || 45,
        predicted_wait_time: r.predicted_wait_time || 0,
      }));

      // Call agent
      const mlRes = await fetch(`${ML_API}/recommend/restaurants`, {
        method:'POST',
        headers:{'Content-Type':'application/json'},
        body: JSON.stringify({
          latitude: userLat || 0,
          longitude: userLng || 0,
          radius_km: 10,
          party_size: partySize,
          candidate_restaurants: candidates,
        }),
      }).then(r => r.json()).catch(() => null);

      if (mlRes?.recommended_restaurants) {
        recommendations = mlRes.recommended_restaurants;
        agentRationale  = mlRes.rationale || [];
        agentModel      = mlRes.agent_model || '';
        geoComputation  = mlRes.geo_computation || {};
      }

      plotRestaurantMarkers(candidates, mlRes?.recommended_restaurants || []);
    } catch(e) { loadErr = e.message; }
    finally { loading = false; }
  }

  function plotRestaurantMarkers(all, recs) {
    if (!leafletMap) return;
    const L = window.L;
    markers.forEach(m => m.remove());
    markers = [];
    const recIds = new Set(recs.map(r => r.id));

    all.forEach(r => {
      if (!r.latitude && !r.longitude) return;
      const isRec = recIds.has(r.id);
      const rec = recs.find(x => x.id === r.id);
      const wait = rec?.predicted_wait_time ?? r.predicted_wait_time ?? 0;
      const color = wait < 15 ? '#10b981' : wait < 30 ? '#f59e0b' : '#ef4444';

      const icon = L.divIcon({
        className: '',
        html: `<div style="
          width:${isRec?'44px':'34px'};height:${isRec?'44px':'34px'};
          border-radius:50%;background:${color};border:3px solid white;
          display:flex;align-items:center;justify-content:center;
          font-size:${isRec?'0.75rem':'0.68rem'};font-weight:800;color:white;
          box-shadow:0 4px 14px rgba(0,0,0,0.45);
          ${isRec?'box-shadow:0 4px 18px '+color+', 0 0 0 6px '+color+'44;':''}
        ">${wait}m</div>`,
        iconSize: [isRec?44:34, isRec?44:34],
        iconAnchor: [isRec?22:17, isRec?22:17],
      });

      const popup = `
        <div style="min-width:180px;font-family:Inter,sans-serif">
          <strong style="font-size:1rem">${r.name}</strong>
          ${isRec ? '<span style="background:#6366f1;color:white;padding:2px 8px;border-radius:999px;font-size:0.78rem;margin-left:6px">⭐ Recommended</span>' : ''}
          <p style="color:#6b7280;margin:4px 0">${r.address || ''}</p>
          <p style="color:#374151"><b>${wait} min</b> predicted wait · ${rec ? round2(rec.distance_km) : '?'} km</p>
          ${rec?.agent_rationale ? `<p style="color:#6366f1;font-style:italic;margin-top:6px;font-size:0.88rem">"${rec.agent_rationale}"</p>` : ''}
        </div>`;

      const m = L.marker([r.latitude, r.longitude], {icon}).addTo(leafletMap).bindPopup(popup);
      m.on('click', () => { selectedId = r.id; });
      markers.push(m);
    });
  }

  function round2(n) { return Math.round((n||0)*100)/100; }

  async function joinQueue() {
    joinMsg = '';
    if (!selectedId || !joinName.trim()) { joinMsg = 'Select a restaurant and enter your name.'; return; }
    try {
      await apiFetch(`/restaurants/${selectedId}/queue/join`, {
        method:'POST',
        body: JSON.stringify({ guest_name:joinName, party_size:partySize, phone_number:joinPhone, source_channel:'customer-app' }),
      });
      joinMsg = `✅ You joined the queue at ${restaurants.find(r=>r.id===selectedId)?.name}!`;
      joinName = ''; joinPhone = '';
    } catch(e) { joinMsg = '❌ ' + e.message; }
  }

  onMount(() => {
    const L = window.L;
    leafletMap = L.map(mapEl).setView([20.5937, 78.9629], 5);
    L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
      attribution:'© OpenStreetMap © CARTO', maxZoom:19,
    }).addTo(leafletMap);
    mapReady = true;
    requestGeo();
  });

  onDestroy(() => { if (leafletMap) leafletMap.remove(); });
</script>

<div class="cust-page">
  <!-- Top bar -->
  <header class="topbar">
    <div class="topbar-left">
      <span class="logo-icon">🍽️</span>
      <span class="topbar-title">SmartRestro <span class="role-badge">Customer</span></span>
    </div>
    <div class="topbar-right">
      <span class="user-name">👤 {currentUser?.name || 'Guest'}</span>
      <button class="logout-btn" on:click={() => dispatch('logout')}>Logout</button>
    </div>
  </header>

  <!-- Map fills the page -->
  <div class="map-wrap" bind:this={mapEl}></div>

  <!-- Side panel -->
  <div class="side-panel">
    <div class="panel-header">
      <h2>🗺️ Nearby Restaurants</h2>
      {#if geoStatus}<p class="geo-status">{geoStatus}</p>{/if}
      {#if geoError}<p class="geo-err">{geoError} <button on:click={requestGeo}>Retry</button></p>{/if}
      {#if userLat}
        <p class="geo-info">📍 <code>{round2(userLat)}°N, {round2(userLng)}°E</code></p>
      {/if}
    </div>

    <!-- Agent info -->
    {#if agentModel}
      <div class="agent-badge">
        🤖 Agent: <strong>{agentModel}</strong>
        {#if geoComputation.formula}
          <span class="formula-pill" title={geoComputation.formula}>Haversine ℹ️</span>
        {/if}
      </div>
    {/if}

    <!-- Party size & Refresh -->
    <div class="controls-row">
      <label>Party size
        <input type="number" bind:value={partySize} min="1" max="20" />
      </label>
      <button class="btn-refresh" on:click={fetchRestaurantsAndRecommend} disabled={loading}>
        {loading ? '…' : '🔄 Refresh'}
      </button>
    </div>

    {#if loadErr}<div class="err-banner">{loadErr}</div>{/if}

    <!-- Recommendations -->
    {#if recommendations.length > 0}
      <div class="rec-section">
        <h3>⭐ Agent Recommendations</h3>
        {#each recommendations as r, i}
          <button class="rec-card" class:selected={selectedId === r.id} on:click={() => selectedId = r.id}>
            <div class="rec-top">
              <span class="rec-rank">#{i+1}</span>
              <span class="rec-name">{r.name}</span>
              <span class="wait-chip">{r.predicted_wait_time}m</span>
            </div>
            <div class="rec-meta">{round2(r.distance_km)} km away</div>
            {#if r.agent_rationale}
              <div class="rec-rationale">"{r.agent_rationale}"</div>
            {/if}
          </button>
        {/each}
      </div>
    {:else if !loading}
      <div class="empty-state">Allow location & click Refresh to see AI-ranked restaurants.</div>
    {/if}

    <!-- Join queue -->
    {#if selectedId}
      <div class="join-section">
        <h3>🪑 Join Queue — {restaurants.find(r=>r.id===selectedId)?.name}</h3>
        <input placeholder="Your name" bind:value={joinName} />
        <input placeholder="Phone (optional)" bind:value={joinPhone} />
        <button class="btn-join" on:click={joinQueue}>Join Queue</button>
        {#if joinMsg}<p class="join-msg">{joinMsg}</p>{/if}
      </div>
    {/if}

    <!-- Agent rationale log -->
    {#if agentRationale.length > 0}
      <div class="rationale-log">
        <h3>🧠 Agent Pipeline Log</h3>
        {#each agentRationale as line}
          <p>• {line}</p>
        {/each}
        {#if geoComputation.note}
          <p class="geo-note">📡 {geoComputation.note}</p>
        {/if}
      </div>
    {/if}
  </div>
</div>

<style>
.cust-page { height:100vh;display:flex;flex-direction:column;overflow:hidden; }

.topbar { position:relative;z-index:20;display:flex;align-items:center;justify-content:space-between;padding:12px 20px;background:rgba(10,15,30,0.95);border-bottom:1px solid rgba(255,255,255,0.08);flex-shrink:0; }
.topbar-left { display:flex;align-items:center;gap:10px;font-weight:800;font-size:1.1rem; }
.logo-icon { font-size:1.6rem; }
.role-badge { background:rgba(99,102,241,0.25);color:#a5b4fc;border-radius:999px;padding:2px 10px;font-size:0.72rem;font-weight:700;margin-left:6px; }
.topbar-right { display:flex;align-items:center;gap:12px; }
.user-name { color:#94a3b8;font-size:0.9rem; }
.logout-btn { background:rgba(239,68,68,0.15);border:1px solid rgba(239,68,68,0.3);color:#fca5a5;border-radius:10px;padding:8px 16px;font-size:0.85rem;font-weight:600; }
.logout-btn:hover { background:rgba(239,68,68,0.25); }

.map-wrap { flex:1;position:relative;z-index:1; }

.side-panel { position:absolute;top:56px;right:0;bottom:0;width:360px;z-index:10;background:rgba(10,15,30,0.97);border-left:1px solid rgba(255,255,255,0.08);overflow-y:auto;padding:16px;display:flex;flex-direction:column;gap:14px; }

.panel-header h2 { font-size:1.1rem;font-weight:800;margin-bottom:6px; }
.geo-status { color:#94a3b8;font-size:0.85rem; }
.geo-err { color:#fca5a5;font-size:0.85rem;display:flex;align-items:center;gap:8px; }
.geo-err button { background:rgba(99,102,241,0.2);border:none;color:#a5b4fc;border-radius:8px;padding:4px 10px;font-size:0.8rem; }
.geo-info { font-size:0.82rem;color:#94a3b8; }
.geo-info code { background:rgba(255,255,255,0.08);padding:2px 6px;border-radius:6px; }

.agent-badge { background:rgba(99,102,241,0.12);border:1px solid rgba(99,102,241,0.25);border-radius:12px;padding:10px 14px;font-size:0.85rem;display:flex;align-items:center;gap:8px;color:#a5b4fc; }
.formula-pill { background:rgba(99,102,241,0.2);padding:2px 8px;border-radius:8px;font-size:0.78rem;cursor:help; }

.controls-row { display:flex;gap:10px;align-items:flex-end; }
.controls-row label { font-size:0.82rem;color:#94a3b8;display:flex;flex-direction:column;gap:4px; }
.controls-row input { background:rgba(255,255,255,0.07);border:1px solid rgba(255,255,255,0.12);border-radius:10px;color:#f1f5f9;padding:8px 12px;width:70px;font-size:0.9rem; }
.btn-refresh { background:rgba(255,255,255,0.08);border:1px solid rgba(255,255,255,0.15);color:#94a3b8;border-radius:10px;padding:9px 14px;font-size:0.85rem;font-weight:600; }
.btn-refresh:hover:not(:disabled) { background:rgba(255,255,255,0.14); }

.err-banner { background:rgba(239,68,68,0.12);border:1px solid rgba(239,68,68,0.3);color:#fca5a5;border-radius:12px;padding:10px 14px;font-size:0.88rem; }

.rec-section h3, .join-section h3, .rationale-log h3 { font-size:0.88rem;font-weight:700;color:#94a3b8;margin-bottom:8px;text-transform:uppercase;letter-spacing:0.04em; }
.rec-card { width:100%;text-align:left;background:rgba(255,255,255,0.04);border:1px solid rgba(255,255,255,0.08);border-radius:14px;padding:12px 14px;cursor:pointer;transition:all 0.2s;display:block; }
.rec-card.selected { border-color:rgba(99,102,241,0.5);background:rgba(99,102,241,0.1); }
.rec-card:hover { border-color:rgba(255,255,255,0.18); }
.rec-top { display:flex;align-items:center;gap:8px;margin-bottom:4px; }
.rec-rank { background:linear-gradient(135deg,#6366f1,#8b5cf6);color:white;border-radius:8px;padding:2px 8px;font-size:0.78rem;font-weight:800; }
.rec-name { font-weight:700;font-size:0.95rem;flex:1; }
.wait-chip { background:rgba(16,185,129,0.15);color:#6ee7b7;border-radius:999px;padding:3px 10px;font-size:0.8rem;font-weight:700; }
.rec-meta { font-size:0.82rem;color:#64748b; }
.rec-rationale { font-size:0.82rem;color:#a5b4fc;font-style:italic;margin-top:6px;line-height:1.5; }

.empty-state { color:#475569;font-size:0.88rem;text-align:center;padding:20px 0; }

.join-section { display:flex;flex-direction:column;gap:8px;background:rgba(16,185,129,0.06);border:1px solid rgba(16,185,129,0.2);border-radius:16px;padding:14px; }
.join-section input { background:rgba(255,255,255,0.07);border:1px solid rgba(255,255,255,0.1);border-radius:10px;color:#f1f5f9;padding:10px 12px;font-size:0.9rem; }
.btn-join { background:linear-gradient(135deg,#10b981,#059669);border:none;border-radius:12px;color:white;padding:12px;font-weight:800;font-size:0.95rem; }
.btn-join:hover { transform:translateY(-1px); }
.join-msg { font-size:0.88rem;color:#6ee7b7;text-align:center; }

.rationale-log { background:rgba(255,255,255,0.03);border:1px solid rgba(255,255,255,0.07);border-radius:14px;padding:14px; }
.rationale-log p { font-size:0.8rem;color:#64748b;line-height:1.6; }
.geo-note { color:#a5b4fc !important;margin-top:8px; }
</style>
