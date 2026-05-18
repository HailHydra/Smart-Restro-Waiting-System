<script>
  import { createEventDispatcher } from 'svelte';
  import heroImg from '../assets/hero.png';
  const dispatch = createEventDispatcher();

  const features = [
    { icon: '🗺️', title: 'Live Geo Map',     desc: 'See real restaurants around you on an interactive map. We use your location to rank by distance.' },
    { icon: '🤖', title: 'AI Agent',          desc: 'LangGraph + Groq llama-3.3-70b generates a personalised recommendation for your party size.' },
    { icon: '⏱️', title: 'Wait Prediction',   desc: 'Heuristic ML model predicts wait time from queue depth, occupancy, and seating velocity.' },
    { icon: '📡', title: 'Real-Time Queue',   desc: 'WebSocket / SSE push updates — no manual refresh. Staff can advance status live.' },
  ];
</script>

<div class="home">
  <!-- Subtle dark background -->
  <div class="bg-gradient"></div>

  <nav class="nav">
    <div class="nav-logo">
      <span class="logo-icon">🍽️</span>
      <span class="logo-text">SmartRestro</span>
    </div>
    <button class="btn-outline" on:click={() => dispatch('goauth')}>Sign In</button>
  </nav>

  <section class="hero-container">
    <div class="hero-visual" style="background-image: url({heroImg})">
      <div class="hero-overlay"></div>
      <div class="hero-content">
        <div class="hero-badge">✨ AI-Powered Restaurant Intelligence</div>
        <h1 class="hero-title">
          Skip the Wait.<br />
          <span class="gradient-text">Find the Best Table.</span>
        </h1>
        <p class="hero-sub">
          Real-time GPS tracking and AI agents recommend the perfect spot with the shortest wait.
        </p>
        <div class="hero-btns">
          <button class="btn-primary" on:click={() => dispatch('goauth')}>
            Get Started →
          </button>
          <a class="btn-ghost" href="#how">Learn More</a>
        </div>
      </div>
    </div>

    <div class="hero-stats">
      <div class="stat"><strong>3-node</strong><span>AI Agent</span></div>
      <div class="stat-div"></div>
      <div class="stat"><strong>Live</strong><span>Real-time</span></div>
      <div class="stat-div"></div>
      <div class="stat"><strong>Haversine</strong><span>Geo Logic</span></div>
    </div>
  </section>

  <section class="features" id="how">
    <h2 class="section-title">Why SmartRestro?</h2>
    <p class="section-sub">Combining spatial intelligence with predictive modeling</p>
    <div class="feature-grid">
      {#each features as f}
        <div class="feature-card">
          <div class="feature-icon">{f.icon}</div>
          <h3>{f.title}</h3>
          <p>{f.desc}</p>
        </div>
      {/each}
    </div>
  </section>

  <section class="geo-explain">
    <div class="geo-card">
      <h2>📍 Geo-Spatial Intelligence</h2>
      <p>Our agent uses your precise coordinates to calculate real-world travel times and seating availability.</p>
      <div class="pipeline">
        <div class="pipe-step">
          <span class="pipe-num">1</span>
          <div>
            <strong>Browser Geolocation</strong>
            <p>Precise decimal-degree coordinates e.g. <code>(18.52, 73.85)</code></p>
          </div>
        </div>
        <div class="pipe-arrow">↓</div>
        <div class="pipe-step">
          <span class="pipe-num">2</span>
          <div>
            <strong>Haversine Calculation</strong>
            <p>Computes great-circle distance between you and every restaurant in km.</p>
          </div>
        </div>
        <div class="pipe-arrow">↓</div>
        <div class="pipe-step">
          <span class="pipe-num">3</span>
          <div>
            <strong>Agent Reasoning</strong>
            <p>Blends distance (25%), wait time (50%), and occupancy (25%) into a master score.</p>
          </div>
        </div>
        <div class="pipe-arrow">↓</div>
        <div class="pipe-step">
          <span class="pipe-num">4</span>
          <div>
            <strong>LLM Rationale</strong>
            <p>Groq explains <em>why</em> a restaurant was picked for your specific party size.</p>
          </div>
        </div>
      </div>
    </div>
  </section>

  <footer class="footer">
    <span class="logo-icon">🍽️</span>
    <span>SmartRestro © 2026 · AI Waiting System</span>
  </footer>
</div>

<style>
.home { min-height:100vh; position:relative; overflow-x:hidden; background: #0a0f1e; color: #f1f5f9; }
.bg-gradient { position:fixed; inset:0; z-index:0; background: radial-gradient(circle at 50% -20%, #1e1b4b 0%, #0a0f1e 70%); }

.nav { position:relative;z-index:10;display:flex;align-items:center;justify-content:space-between;padding:24px 48px; }
.nav-logo { display:flex;align-items:center;gap:12px;font-size:1.5rem;font-weight:900; }
.logo-icon { font-size:2rem; }
.btn-outline { border:1px solid rgba(255,255,255,0.15);border-radius:14px;background:rgba(255,255,255,0.04);color:#f1f5f9;padding:10px 28px;font-weight:700;font-size:0.95rem; backdrop-filter: blur(8px); }
.btn-outline:hover { background:rgba(255,255,255,0.1); border-color: rgba(255,255,255,0.3); }

.hero-container { position:relative; z-index:5; padding: 40px 48px; max-width: 1400px; margin: 0 auto; }
.hero-visual { 
  position: relative;
  height: 600px;
  border-radius: 40px;
  overflow: hidden;
  background-size: cover;
  background-position: center;
  display: flex;
  align-items: center;
  padding: 80px;
  box-shadow: 0 40px 100px rgba(0,0,0,0.5);
  border: 1px solid rgba(255,255,255,0.1);
}
.hero-overlay { 
  position: absolute; 
  inset: 0; 
  background: linear-gradient(90deg, rgba(10,15,30,0.9) 0%, rgba(10,15,30,0.4) 60%, transparent 100%);
}
.hero-content { position: relative; z-index: 10; max-width: 600px; }

.hero-badge { display:inline-block;background:rgba(99,102,241,0.2);border:1px solid rgba(99,102,241,0.4);color:#a5b4fc;border-radius:999px;padding:8px 20px;font-size:0.85rem;font-weight:700;margin-bottom:28px; }
.hero-title { font-size:clamp(3rem,8vw,5rem);font-weight:950;line-height:1.05;margin-bottom:24px; letter-spacing: -0.02em; }
.gradient-text { background:linear-gradient(135deg,#6366f1,#a855f7,#06b6d4);-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text; }
.hero-sub { font-size:1.25rem;color:#94a3b8;margin-bottom:44px; line-height: 1.6; }
.hero-btns { display:flex;gap:20px; flex-wrap:wrap; margin-bottom: 0; }
.btn-primary { background:linear-gradient(135deg,#6366f1,#8b5cf6);border:none;border-radius:16px;color:white;padding:16px 40px;font-size:1.1rem;font-weight:800;box-shadow:0 10px 30px rgba(99,102,241,0.4); transition: all 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275); }
.btn-primary:hover { transform:translateY(-4px) scale(1.02); box-shadow:0 15px 40px rgba(99,102,241,0.6); }
.btn-ghost { border:1px solid rgba(255,255,255,0.2);border-radius:16px;background:rgba(255,255,255,0.05);color:#cbd5e1;padding:16px 36px;font-size:1.1rem;font-weight:700; backdrop-filter: blur(8px); }
.btn-ghost:hover { background:rgba(255,255,255,0.1); border-color: rgba(255,255,255,0.4); }

.hero-stats { display:flex;align-items:center;justify-content:center;gap:48px; margin-top: 60px; }
.stat { text-align:center; }
.stat strong { display:block;font-size:1.8rem;font-weight:900;color:#f1f5f9; }
.stat span { font-size:0.9rem;color:#64748b; font-weight: 600; text-transform: uppercase; letter-spacing: 0.1em; }
.stat-div { width:1px;height:48px;background:rgba(255,255,255,0.1); }

.section-title { font-size:2.5rem;font-weight:900;text-align:center;margin-bottom:12px; }
.section-sub { color:#94a3b8;text-align:center;margin-bottom:60px; font-size: 1.1rem; }
.features { position:relative;z-index:5;padding:120px 48px; }
.feature-grid { display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:32px;max-width:1200px;margin:0 auto; }
.feature-card { background:rgba(255,255,255,0.03);border:1px solid rgba(255,255,255,0.08);border-radius:24px;padding:40px;transition:all 0.3s; backdrop-filter: blur(12px); }
.feature-card:hover { transform:translateY(-8px); border-color:rgba(99,102,241,0.4); background: rgba(255,255,255,0.05); }
.feature-icon { font-size:3rem;margin-bottom:24px; }
.feature-card h3 { font-size:1.4rem;font-weight:800;margin-bottom:14px; color: #f1f5f9; }
.feature-card p { color:#94a3b8;font-size:1rem;line-height:1.7; }

.geo-explain { position:relative;z-index:5;padding:60px 48px 120px; }
.geo-card { max-width:900px;margin:0 auto;background: linear-gradient(145deg, rgba(99,102,241,0.1), rgba(139,92,246,0.05)); border:1px solid rgba(99,102,241,0.2);border-radius:32px;padding:60px; box-shadow: 0 30px 60px rgba(0,0,0,0.3); }
.geo-card h2 { font-size:2rem;font-weight:900;margin-bottom:16px; text-align: center; }
.geo-card > p { color:#94a3b8;margin-bottom:48px;line-height:1.8; text-align: center; font-size: 1.1rem; }
.geo-card code { background:rgba(99,102,241,0.2);padding:4px 10px;border-radius:8px;font-size:0.9em;font-family: monospace; color: #c4b5fd; }
.pipeline { display:grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap:32px; }
.pipe-step { display:flex; flex-direction: column; gap:20px; align-items:center; text-align: center; }
.pipe-step strong { display:block;font-weight:800;margin-bottom:8px; font-size: 1.1rem; }
.pipe-step p { color:#94a3b8;font-size:0.95rem;line-height:1.6; }
.pipe-num { width:44px;height:44px;border-radius:14px;background:linear-gradient(135deg,#6366f1,#8b5cf6);display:flex;align-items:center;justify-content:center;font-weight:900;font-size:1.1rem; box-shadow: 0 8px 20px rgba(99,102,241,0.4); }
.pipe-arrow { display: none; }

.footer { position:relative;z-index:5;text-align:center;padding:60px 20px;color:#475569;display:flex;align-items:center;justify-content:center;gap:12px; border-top: 1px solid rgba(255,255,255,0.05); }

@media(max-width: 900px) {
  .hero-visual { padding: 40px; height: auto; min-height: 500px; }
  .nav { padding: 24px; }
  .hero-container { padding: 20px; }
  .hero-stats { gap: 24px; flex-wrap: wrap; }
}
</style>
