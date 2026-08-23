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
    /* ink surfaces */
    --bg: #131110;
    --bg-raised: #1b1815;
    --bg-sunken: #0d0b0a;
    --border: rgba(237, 230, 216, 0.08);
    --border-strong: rgba(237, 230, 216, 0.16);

    /* parchment text */
    --text: #ede6d8;
    --text-muted: #9c9284;
    --text-faint: #6b6459;

    /* single accent: lantern amber */
    --amber: #e0a94a;
    --amber-bright: #f2c877;
    --amber-soft: rgba(224, 169, 74, 0.14);

    /* status */
    --ready: var(--text-muted);
    --ready-soft: rgba(156, 146, 132, 0.12);
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
