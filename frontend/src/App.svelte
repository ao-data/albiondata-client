<script>
  import Sidebar from './lib/Sidebar.svelte';
  import LogTail from './lib/LogTail.svelte';
  import Footer from './lib/Footer.svelte';
</script>

<main>
  <div class="body">
    <Sidebar />
    <div class="content">
      <LogTail />
    </div>
  </div>
  <Footer />
</main>

<style>
  :global(:root) {
    /* charcoal surfaces */
    --bg: #14171c;
    --bg-raised: #1c2128;
    --bg-sunken: #0d0f13;
    --border: rgba(231, 234, 239, 0.08);
    --border-strong: rgba(231, 234, 239, 0.16);

    /* cool text */
    --text: #e7eaef;
    --text-muted: #8d94a3;
    --text-faint: #5a6070;

    /* single accent: electric blue */
    --blue: #2f8cff;
    --blue-bright: #5fb4ff;
    --blue-soft: rgba(47, 140, 255, 0.14);

    /* status */
    --ready: var(--text-muted);
    --ready-soft: rgba(141, 148, 163, 0.12);
    --ember: #c1543f;
    --ember-soft: rgba(193, 84, 63, 0.16);
    --moss: #6d9c5f;
    --moss-soft: rgba(109, 156, 95, 0.16);

    --radius: 7px;
    --radius-sm: 5px;

    --font-display: 'Iowan Old Style', 'Palatino Linotype', Palatino, Georgia, serif;
    --font-body: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    --font-mono: ui-monospace, 'SF Mono', SFMono-Regular, Menlo, monospace;
  }
  :global(html, body) {
    margin: 0;
    height: 100%;
    background: var(--bg);
    color: var(--text);
    font-family: var(--font-body);
    -webkit-font-smoothing: antialiased;
  }
  main {
    display: flex;
    flex-direction: column;
    height: 100vh;
  }
  .body {
    display: flex;
    flex: 1;
    min-height: 0;
  }
  .content {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
    background: var(--bg);
  }
</style>
