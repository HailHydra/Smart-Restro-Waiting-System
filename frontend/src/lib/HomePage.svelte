<script>
  import { createEventDispatcher } from 'svelte';
  import { fade, fly } from 'svelte/transition';
  import Hero from './Hero.svelte';
  import FeatureGrid from './FeatureGrid.svelte';
  import QueueSimulator from './QueueSimulator.svelte';

  const dispatch = createEventDispatcher();
</script>

<svelte:head>
  <title>Smart Restro | Live Wait Transparency</title>
  <meta
    name="description"
    content="A premium live wait-time experience with honest ranges, proactive alerts, and clear queue transparency."
  />
</svelte:head>

<div class="landing-shell">
  <div class="atmosphere" aria-hidden="true"></div>

  <header class="topbar" in:fly={{ y: -18, duration: 220 }}>
    <div class="brand-block">
      <span class="brand-kicker">SMART RESTRO</span>
      <span class="brand-name">LIVE WAIT, CLEAR EXPECTATIONS</span>
    </div>
    <button class="topbar-action" on:click={() => dispatch('goauth')}>
      RESTAURANT PORTAL
    </button>
  </header>

  <main class="landing-main">
    <Hero on:restaurantlogin={() => dispatch('goauth')} />

    <section class="section-frame" in:fade={{ duration: 250 }}>
      <div class="section-label mint">PROACTIVE ALERTS</div>
      <h2>STEP AWAY FROM THE DOOR WITHOUT LOSING YOUR PLACE.</h2>
      <p>
        Choose how and when you want to be notified, from an early heads-up to a final call when
        it is time to head back to the host stand.
      </p>
      <QueueSimulator />
    </section>

    <section class="section-frame" in:fade={{ duration: 250 }}>
      <div class="section-label lavender">PROXIMITY PRESENCE</div>
      <h2>FAIR TO EVERYONE IN LINE, FORGIVING OF REAL-WORLD GPS NOISE.</h2>
      <p>
        We confirm that guests are reasonably close to the restaurant to keep the line fair, while
        still accounting for indoor signal drift, older phones, and brief location hiccups.
      </p>

      <div class="proximity-layout">
        <div class="map-placeholder">
          <div class="map-ring"></div>
          <div class="map-pin">RESTAURANT</div>
          <span>INTERACTIVE MAP PLACEHOLDER</span>
        </div>

        <div class="proximity-copy">
          <div class="proximity-note">
            <strong>Why proximity matters</strong>
            <p>
              It helps protect the line from remote hold spots while making sure nearby guests are
              treated consistently and fairly.
            </p>
          </div>

          <div class="proximity-note mint">
            <strong>You will not be dropped unfairly</strong>
            <p>
              If you step inside a nearby shop, walk under heavy coverage, or your phone drifts for
              a moment, your place is not removed just because the signal briefly gets messy.
            </p>
          </div>

          <div class="proximity-note yellow">
            <strong>Built for normal movement</strong>
            <p>
              The experience is designed around everyday guest behavior, not perfect outdoor GPS.
            </p>
          </div>
        </div>
      </div>
    </section>

    <section class="section-frame" in:fade={{ duration: 250 }}>
      <div class="section-label">WHY IT'S FAIR</div>
      <h2>THE LINE FEELS CALMER WHEN PEOPLE CAN SEE WHY IT MOVES THE WAY IT DOES.</h2>
      <p>
        Smart Restro replaces vague clipboard estimates with rules that guests can actually
        understand: table fit, current pace, and active cleanup of abandoned spots.
      </p>
      <FeatureGrid />
    </section>
  </main>

  <footer class="site-footer" in:fade={{ duration: 250 }}>
    <div class="footer-brand">
      <strong>Powered by Smart Dynamic Queueing.</strong>
      <span>Premium live wait transparency for modern walk-up dining.</span>
    </div>

    <nav class="footer-links" aria-label="Footer">
      <a href="/">Privacy Policy</a>
      <a href="/">Terms of Service</a>
      <a href="/">Help Center</a>
      <a href="/" on:click|preventDefault={() => dispatch('goauth')}>Restaurant Portal</a>
    </nav>

    <div class="status-badge">
      <span class="status-dot"></span>
      All systems operational. Live connection secure.
    </div>
  </footer>
</div>

<style>
  .landing-shell {
    min-height: 100vh;
    background: var(--app-background);
    color: var(--ink);
    padding: 76px 24px 24px;
    position: relative;
    overflow-x: clip;
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

  .topbar,
  .section-frame,
  .site-footer {
    position: relative;
    z-index: 1;
  }

  .topbar {
    max-width: 1240px;
    margin: 0 auto 20px;
    display: flex;
    justify-content: space-between;
    gap: 18px;
    align-items: center;
    border: 4px solid var(--ink);
    box-shadow: 8px 8px 0 0 var(--ink);
    background: rgba(255, 255, 255, 0.92);
    padding: 18px 20px;
    backdrop-filter: blur(10px);
  }

  .brand-block {
    display: grid;
    gap: 2px;
  }

  .brand-kicker,
  .brand-name,
  .section-label {
    font-weight: 900;
    letter-spacing: -0.04em;
    text-transform: uppercase;
  }

  .brand-kicker {
    font-size: 0.78rem;
  }

  .brand-name {
    font-size: clamp(1.15rem, 3vw, 1.55rem);
  }

  .topbar-action {
    border: 4px solid var(--ink);
    background: var(--lavender);
    box-shadow: 6px 6px 0 0 var(--ink);
    padding: 0.95rem 1.2rem;
    font-weight: 900;
    font-size: 0.9rem;
    text-transform: uppercase;
  }

  .topbar-action:active {
    transform: translate(4px, 4px);
    box-shadow: 2px 2px 0 0 var(--ink);
  }

  .landing-main {
    max-width: 1240px;
    margin: 0 auto;
    display: grid;
    gap: 28px;
  }

  .section-frame {
    border: 4px solid var(--ink);
    box-shadow: 10px 10px 0 0 var(--ink);
    background: rgba(255, 255, 255, 0.9);
    padding: clamp(20px, 3vw, 34px);
    display: grid;
    gap: 18px;
    backdrop-filter: blur(8px);
  }

  .section-frame h2 {
    font-size: clamp(2rem, 5vw, 3.2rem);
    line-height: 0.94;
    max-width: 15ch;
  }

  .section-frame p {
    max-width: 70ch;
    font-size: 1rem;
  }

  .section-label {
    display: inline-flex;
    width: fit-content;
    border: 3px solid var(--ink);
    background: var(--yellow);
    padding: 0.45rem 0.75rem;
    font-size: 0.82rem;
  }

  .section-label.mint {
    background: var(--mint);
  }

  .section-label.lavender {
    background: var(--lavender);
  }

  .proximity-layout {
    display: grid;
    grid-template-columns: minmax(300px, 0.95fr) minmax(0, 1.05fr);
    gap: 20px;
  }

  .map-placeholder,
  .proximity-note,
  .footer-brand,
  .footer-links,
  .status-badge {
    border: 4px solid var(--ink);
  }

  .map-placeholder {
    min-height: 360px;
    background:
      linear-gradient(180deg, #eef7ff 0%, #fdfdfd 100%);
    box-shadow: 8px 8px 0 0 var(--ink);
    position: relative;
    overflow: hidden;
    display: grid;
    place-items: center;
    text-align: center;
    padding: 20px;
  }

  .map-ring {
    position: absolute;
    width: 66%;
    aspect-ratio: 1;
    border-radius: 999px;
    border: 4px solid var(--ink);
    background: rgba(165, 214, 167, 0.28);
  }

  .map-pin {
    position: relative;
    z-index: 1;
    border: 4px solid var(--ink);
    background: var(--yellow);
    box-shadow: 6px 6px 0 0 var(--ink);
    padding: 0.85rem 1rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  .map-placeholder span {
    position: absolute;
    bottom: 18px;
    left: 18px;
    font-family: var(--mono);
    font-size: 0.82rem;
    text-transform: uppercase;
  }

  .proximity-copy {
    display: grid;
    gap: 14px;
    align-content: start;
  }

  .proximity-note {
    background: var(--panel);
    box-shadow: 6px 6px 0 0 var(--ink);
    padding: 18px;
    display: grid;
    gap: 10px;
  }

  .proximity-note strong {
    font-family: var(--display);
    font-size: 1.15rem;
    line-height: 1;
  }

  .proximity-note.mint {
    background: #effbea;
  }

  .proximity-note.yellow {
    background: #fff9d8;
  }

  .site-footer {
    max-width: 1240px;
    margin: 0 auto 28px;
    display: grid;
    grid-template-columns: 1.2fr 1fr auto;
    gap: 20px;
    align-items: center;
  }

  .footer-brand,
  .footer-links,
  .status-badge {
    background: rgba(255, 255, 255, 0.92);
    box-shadow: 8px 8px 0 0 var(--ink);
    padding: 18px;
  }

  .footer-brand {
    display: grid;
    gap: 8px;
  }

  .footer-brand strong {
    font-family: var(--display);
    font-size: 1.1rem;
    line-height: 1;
  }

  .footer-links {
    display: flex;
    flex-wrap: wrap;
    gap: 14px 18px;
    justify-content: center;
    font-weight: 700;
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    font-size: 0.92rem;
    font-weight: 700;
    white-space: nowrap;
  }

  .status-dot {
    width: 12px;
    height: 12px;
    border-radius: 999px;
    background: #1faa59;
    border: 2px solid var(--ink);
  }

  @media (max-width: 900px) {
    .site-footer,
    .proximity-layout {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 720px) {
    .landing-shell {
      padding: 70px 14px 14px;
    }

    .topbar {
      flex-direction: column;
      align-items: stretch;
    }

    .topbar-action {
      width: 100%;
    }

    .footer-links,
    .status-badge {
      justify-content: start;
      white-space: normal;
    }
  }
</style>
