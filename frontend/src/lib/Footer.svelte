<script>
  import { onDestroy } from 'svelte';
  import { Browser, Events } from '@wailsio/runtime';
  import { DashboardService } from '../../bindings/github.com/ao-data/albiondata-client/internal/dashboard/index.js';

  let version = $state('');

  DashboardService.GetStatus().then((s) => (version = s.Version));

  const unlisten = Events.On('status:changed', (evt) => {
    version = evt.data.Version;
  });

  onDestroy(unlisten);

  function openLink(e, url) {
    e.preventDefault();
    Browser.OpenURL(url);
  }
</script>

<footer class="footer">
  <nav class="links">
    <a href="https://www.albion-online-data.com" onclick={(e) => openLink(e, 'https://www.albion-online-data.com')}>
      albion-online-data.com
    </a>
    <a href="https://discord.gg/yv5SgytAjX" onclick={(e) => openLink(e, 'https://discord.gg/yv5SgytAjX')}>
      Discord
    </a>
  </nav>
  <span class="version">v{version || 'dev'}</span>
</footer>

<style>
  .footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    padding: 0.5rem 1.25rem;
    background: var(--bg-raised);
    border-top: 1px solid var(--border);
    flex-shrink: 0;
  }
  .links {
    display: flex;
    gap: 1.25rem;
  }
  .links a {
    font-size: 0.75rem;
    color: var(--text-muted);
    text-decoration: none;
  }
  .links a:hover {
    color: var(--blue);
  }
  .version {
    font-family: var(--font-mono);
    font-size: 0.72rem;
    color: var(--text-faint);
  }
</style>
