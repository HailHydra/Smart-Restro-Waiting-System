<script>
  let selectedAlert = 'almost';
  let updatesOn = true;
  let expiryOn = true;

  $: alertCopy =
    selectedAlert === 'almost'
      ? "You're up next! Please begin making your way to the host stand."
      : selectedAlert === 'soon'
        ? 'Your table is getting closer. You still have a few minutes to wrap up nearby.'
        : 'We will alert you the moment your table is ready.';
</script>

<div class="alert-grid">
  <div class="control-card">
    <div class="card-head">
      <span>PROACTIVE ALERT OPTIONS</span>
      <strong>CHOOSE YOUR COMFORT LEVEL</strong>
    </div>

    <div class="alert-options">
      <button class:active={selectedAlert === 'ready'} on:click={() => (selectedAlert = 'ready')}>
        TABLE READY
      </button>
      <button class:active={selectedAlert === 'soon'} on:click={() => (selectedAlert = 'soon')}>
        A FEW MINUTES AWAY
      </button>
      <button class:active={selectedAlert === 'almost'} on:click={() => (selectedAlert = 'almost')}>
        YOU'RE UP NEXT
      </button>
    </div>

    <div class="toggle-list">
      <label class:enabled={updatesOn}>
        <div>
          <strong>STATUS CHANGES</strong>
          <span>Instant updates if your slot is called or your time changes.</span>
        </div>
        <input bind:checked={updatesOn} type="checkbox" />
      </label>

      <label class:enabled={expiryOn}>
        <div>
          <strong>EXPIRATION WARNING</strong>
          <span>A heads-up before your hold window runs out.</span>
        </div>
        <input bind:checked={expiryOn} type="checkbox" />
      </label>
    </div>
  </div>

  <div class="result-card">
    <div class="result-label">ALERT PREVIEW</div>
    <div class="phone-card">
      <span class="phone-badge">SMART RESTRO</span>
      <strong>{alertCopy}</strong>
      <p>
        You can minimize the app, check nearby shops, or wait comfortably without worrying about
        losing your place.
      </p>
    </div>
  </div>
</div>

<style>
  .alert-grid {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(280px, 360px);
    gap: 20px;
  }

  .control-card,
  .result-card {
    border: 4px solid var(--ink);
    box-shadow: 8px 8px 0 0 var(--ink);
    background: var(--panel);
    padding: 20px;
    display: grid;
    gap: 16px;
  }

  .card-head,
  .toggle-list {
    display: grid;
    gap: 12px;
  }

  .card-head span,
  .result-label,
  .phone-badge {
    font-family: var(--mono);
    text-transform: uppercase;
    font-size: 0.82rem;
  }

  .card-head strong {
    font-size: 2rem;
    line-height: 0.95;
    letter-spacing: -0.06em;
  }

  .alert-options {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
  }

  .alert-options button {
    border: 4px solid var(--ink);
    background: var(--ash);
    box-shadow: 6px 6px 0 0 var(--ink);
    padding: 0.95rem 0.8rem;
    font-weight: 900;
    text-transform: uppercase;
  }

  .alert-options button.active {
    background: var(--yellow);
  }

  .alert-options button:active {
    transform: translate(4px, 4px);
    box-shadow: 2px 2px 0 0 var(--ink);
  }

  .toggle-list label {
    border: 4px solid var(--ink);
    background: var(--panel);
    padding: 16px;
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 16px;
    align-items: center;
  }

  .toggle-list label.enabled {
    background: #f8fff5;
  }

  .toggle-list strong {
    display: block;
    font-size: 1rem;
    margin-bottom: 6px;
  }

  .toggle-list span {
    font-size: 0.95rem;
  }

  .toggle-list input {
    width: 22px;
    height: 22px;
    accent-color: var(--ink);
  }

  .result-card {
    background: linear-gradient(180deg, #fef7db 0%, #ffffff 100%);
    align-content: start;
  }

  .phone-card {
    border: 4px solid var(--ink);
    background: var(--panel);
    box-shadow: 8px 8px 0 0 var(--ink);
    padding: 18px;
    display: grid;
    gap: 12px;
  }

  .phone-badge {
    border: 3px solid var(--ink);
    background: var(--mint);
    width: fit-content;
    padding: 0.4rem 0.65rem;
    font-weight: 900;
  }

  .phone-card strong {
    font-size: 1.4rem;
    line-height: 1.05;
    letter-spacing: -0.04em;
  }

  .phone-card p {
    font-size: 0.98rem;
  }

  @media (max-width: 860px) {
    .alert-grid {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 620px) {
    .alert-options {
      grid-template-columns: 1fr;
    }

    .toggle-list label {
      grid-template-columns: 1fr;
    }
  }
</style>
