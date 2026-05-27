<section class="feature-grid">
  <article class="feature-card">
    <div class="card-tag yellow">PARTY-SIZE MATCHING</div>
    <h3>THE RIGHT TABLE FOR YOUR GROUP</h3>
    <div class="visual-stack">
      <div class="group-chip">PARTY OF 4</div>
      <div class="table-row">
        <div class="table-chip active">4-SEAT TABLE</div>
        <div class="table-chip">2-SEAT TABLE</div>
        <div class="table-chip active">BOOTH FOR 4</div>
      </div>
    </div>
    <p>
      Your estimate is based on tables that actually fit your group, which means a two-seat opening
      will not make a larger party look closer than it really is.
    </p>
  </article>

  <article class="feature-card">
    <div class="card-tag lavender">DYNAMIC FLOOR PACING</div>
    <h3>ESTIMATES THAT STAY CURRENT</h3>
    <div class="pace-panel">
      <div>
        <span>EARLIER</span>
        <strong>32-40 MIN</strong>
      </div>
      <div class="arrow">-></div>
      <div class="active">
        <span>NOW</span>
        <strong>20-28 MIN</strong>
      </div>
    </div>
    <p>
      As the dining room speeds up or slows down, your wait range adjusts to reflect what is
      happening right now rather than what happened an hour ago.
    </p>
  </article>

  <article class="feature-card">
    <div class="card-tag mint">ACTIVE LINE CLEANUP</div>
    <h3>NO GHOST SPOTS AHEAD OF YOU</h3>
    <div class="cleanup-panel">
      <div class="line-item">AHEAD OF YOU: 5 PARTIES</div>
      <div class="line-item ghost">STALE CHECK-IN REMOVED</div>
      <div class="line-item active">UPDATED: 4 PARTIES</div>
    </div>
    <p>
      If someone leaves, stops responding, or abandons their place, the line is cleaned up so your
      position does not stay artificially inflated.
    </p>
  </article>
</section>

<style>
  .feature-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 20px;
  }

  .feature-card {
    border: 4px solid var(--ink);
    box-shadow: 8px 8px 0 0 var(--ink);
    background: var(--panel);
    padding: 20px;
    display: grid;
    gap: 14px;
    align-content: start;
  }

  .card-tag {
    width: fit-content;
    border: 3px solid var(--ink);
    padding: 0.45rem 0.7rem;
    font-size: 0.78rem;
    font-family: var(--mono);
    font-weight: 900;
    text-transform: uppercase;
  }

  .yellow {
    background: var(--yellow);
  }

  .lavender {
    background: var(--lavender);
  }

  .mint {
    background: var(--mint);
  }

  h3 {
    font-size: 1.9rem;
    line-height: 0.95;
    letter-spacing: -0.06em;
  }

  p {
    font-size: 0.98rem;
  }

  .visual-stack,
  .pace-panel,
  .cleanup-panel {
    border: 4px solid var(--ink);
    background: var(--ash);
  }

  .visual-stack {
    padding: 18px;
    display: grid;
    gap: 12px;
  }

  .group-chip,
  .table-chip,
  .line-item {
    border: 3px solid var(--ink);
    padding: 0.7rem 0.85rem;
    font-weight: 900;
    text-align: center;
    text-transform: uppercase;
  }

  .group-chip {
    background: var(--yellow);
  }

  .table-row {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 10px;
  }

  .table-chip {
    background: var(--panel);
  }

  .table-chip.active {
    background: var(--mint);
  }

  .pace-panel {
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    align-items: stretch;
  }

  .pace-panel > div {
    padding: 18px 12px;
    display: grid;
    gap: 8px;
    place-items: center;
    text-align: center;
  }

  .pace-panel span,
  .cleanup-panel .line-item {
    font-family: var(--mono);
    font-size: 0.84rem;
  }

  .pace-panel strong {
    font-size: 1.25rem;
    line-height: 1;
  }

  .pace-panel .arrow {
    border-left: 4px solid var(--ink);
    border-right: 4px solid var(--ink);
    font-size: 1.35rem;
    font-weight: 900;
    background: var(--panel);
  }

  .pace-panel .active {
    background: var(--lavender);
  }

  .cleanup-panel {
    padding: 18px;
    display: grid;
    gap: 12px;
  }

  .line-item {
    background: var(--panel);
  }

  .line-item.ghost {
    background: #ffe0e0;
  }

  .line-item.active {
    background: var(--mint);
  }

  @media (max-width: 980px) {
    .feature-grid {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 560px) {
    .table-row,
    .pace-panel {
      grid-template-columns: 1fr;
    }

    .pace-panel .arrow {
      border-left: 0;
      border-right: 0;
      border-top: 4px solid var(--ink);
      border-bottom: 4px solid var(--ink);
      padding: 12px;
    }
  }
</style>
