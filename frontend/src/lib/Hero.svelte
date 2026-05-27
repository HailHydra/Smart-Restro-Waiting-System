<script>
  import { createEventDispatcher } from 'svelte';
  import { fade } from 'svelte/transition';

  const dispatch = createEventDispatcher();
</script>

<section class="hero-frame" in:fade={{ duration: 220 }}>
  <div class="hero-copy">
    <div class="hero-pill">LIVE QUEUE TRANSPARENCY</div>
    <h1>WAIT WITH CONFIDENCE, NOT GUESSWORK.</h1>
    <p>
      Smart Restro gives guests a live, self-correcting view of the line so they can plan their
      next few minutes with less stress and fewer surprises.
    </p>

    <div class="hero-actions">
      <button class="cta yellow" on:click={() => dispatch('restaurantlogin')}>
        JOIN THE LIVE QUEUE
      </button>
      <button class="cta lavender" on:click={() => dispatch('restaurantlogin')}>
        RESTAURANT PORTAL
      </button>
    </div>
  </div>

  <div class="status-card">
    <div class="status-top">
      <div>
        <span class="status-label">LIVE STATUS</span>
        <strong class="status-range">20 - 28 MINUTES</strong>
      </div>
      <div class="sync-pill">
        <span class="sync-dot"></span>
        LIVE SYNC
      </div>
    </div>

    <p class="status-copy">
      Your estimate updates as tables open up and service pace changes, so the range stays honest
      instead of pretending to be exact.
    </p>

    <div class="status-grid">
      <div class="status-block">
        <span>YOUR PLACE</span>
        <strong>3RD IN LINE</strong>
      </div>
      <div class="status-block mint">
        <span>PARTY SIZE</span>
        <strong>TABLE FOR 4</strong>
      </div>
      <div class="status-block lavender">
        <span>LAST UPDATE</span>
        <strong>JUST NOW</strong>
      </div>
    </div>
  </div>
</section>

<style>
  .hero-frame {
    border: 4px solid var(--ink);
    box-shadow: 12px 12px 0 0 var(--ink);
    background: linear-gradient(135deg, #ffffff 0%, #f7f2ff 52%, #eef9ff 100%);
    display: grid;
    grid-template-columns: 1.05fr 0.95fr;
    gap: 24px;
    padding: clamp(20px, 4vw, 36px);
  }

  .hero-copy,
  .status-card {
    display: grid;
    gap: 18px;
  }

  .hero-pill,
  .status-label,
  .status-block span {
    font-family: var(--mono);
    font-size: 0.82rem;
    text-transform: uppercase;
  }

  .hero-pill {
    border: 3px solid var(--ink);
    background: var(--mint);
    padding: 0.55rem 0.8rem;
    width: fit-content;
    font-weight: 900;
  }

  h1 {
    font-size: clamp(3rem, 8vw, 6rem);
    line-height: 0.88;
    letter-spacing: -0.08em;
    max-width: 8ch;
  }

  p {
    font-size: clamp(1rem, 2vw, 1.1rem);
    max-width: 60ch;
  }

  .hero-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 14px;
  }

  .cta {
    border: 4px solid var(--ink);
    box-shadow: 7px 7px 0 0 var(--ink);
    padding: 1rem 1.1rem;
    font-weight: 900;
    font-size: 0.95rem;
    text-transform: uppercase;
    flex: 1 1 220px;
  }

  .yellow {
    background: var(--yellow);
  }

  .lavender {
    background: var(--lavender);
  }

  .cta:active {
    transform: translate(4px, 4px);
    box-shadow: 2px 2px 0 0 var(--ink);
  }

  .status-card {
    border: 4px solid var(--ink);
    box-shadow: 8px 8px 0 0 var(--ink);
    background: rgba(255, 255, 255, 0.9);
    padding: 20px;
    align-content: start;
  }

  .status-top {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    align-items: start;
    flex-wrap: wrap;
  }

  .status-range {
    display: block;
    font-size: clamp(2.2rem, 5vw, 3.7rem);
    line-height: 0.92;
    letter-spacing: -0.08em;
  }

  .sync-pill {
    border: 3px solid var(--ink);
    background: var(--yellow);
    padding: 0.6rem 0.8rem;
    font-family: var(--mono);
    font-size: 0.82rem;
    font-weight: 900;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    text-transform: uppercase;
  }

  .sync-dot {
    width: 11px;
    height: 11px;
    border-radius: 999px;
    background: #1faa59;
    border: 2px solid var(--ink);
    animation: pulse 1.4s infinite;
  }

  .status-copy {
    max-width: 42ch;
  }

  .status-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
  }

  .status-block {
    border: 3px solid var(--ink);
    background: var(--panel);
    padding: 14px;
    display: grid;
    gap: 8px;
  }

  .status-block strong {
    font-size: 1.1rem;
    line-height: 1.1;
  }

  .status-block.mint {
    background: var(--mint);
  }

  .status-block.lavender {
    background: var(--lavender);
  }

  @keyframes pulse {
    0%,
    100% {
      transform: scale(1);
    }

    50% {
      transform: scale(1.18);
    }
  }

  @media (max-width: 920px) {
    .hero-frame {
      grid-template-columns: 1fr;
    }

    h1 {
      max-width: none;
    }
  }

  @media (max-width: 560px) {
    .status-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
