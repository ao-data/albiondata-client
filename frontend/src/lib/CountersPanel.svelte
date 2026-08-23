<script>
  import { onDestroy } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import { DashboardService } from '../../bindings/github.com/ao-data/albiondata-client/internal/dashboard/index.js';

  const labels = {
    'marketorders.ingest': 'Market Orders',
    'goldprices.ingest': 'Gold Prices',
    'markethistories.ingest': 'Market Histories',
    'festivities.ingest': 'Festivities',
    'banditevent.ingest': 'Bandit Events',
  };

  let counts = $state({});

  DashboardService.GetUploadCounts().then((c) => (counts = c));

  const unlisten = Events.On('counters:snapshot', (evt) => {
    counts = evt.data;
  });

  onDestroy(unlisten);
</script>

<section class="counters">
  {#each Object.entries(labels) as [topic, label]}
    <div class="counter">
      <span class="label">{label}</span>
      <span class="value">{counts[topic] ?? 0}</span>
    </div>
  {/each}
</section>

<style>
  .counters {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    padding-top: 1.15rem;
    border-top: 1px solid var(--border);
  }
  .counter {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 0.75rem;
  }
  .label {
    font-size: 0.66rem;
    font-weight: 600;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    color: var(--text-faint);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .value {
    font-family: var(--font-mono);
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--amber-bright);
    font-variant-numeric: tabular-nums;
    flex-shrink: 0;
  }
</style>
