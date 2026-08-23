<script>
  import { onDestroy, tick } from 'svelte';
  import { Browser, Events } from '@wailsio/runtime';
  import { DashboardService } from '../../bindings/github.com/ao-data/albiondata-client/internal/dashboard/index.js';

  const UUID_RE = /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/gi;

  function identifierUrl(id) {
    return `https://www.albion-online-data.com/identifier?identifier=${id}`;
  }

  // Splits a log message into plain-text and identifier segments so
  // identifiers can be rendered as links instead of interpolating raw
  // (untrusted) log text as HTML.
  function splitMessage(message) {
    const parts = [];
    let lastIndex = 0;
    for (const match of message.matchAll(UUID_RE)) {
      if (match.index > lastIndex) {
        parts.push({ type: 'text', value: message.slice(lastIndex, match.index) });
      }
      parts.push({ type: 'identifier', value: match[0] });
      lastIndex = match.index + match[0].length;
    }
    if (lastIndex < message.length) {
      parts.push({ type: 'text', value: message.slice(lastIndex) });
    }
    return parts;
  }

  function openIdentifier(e, id) {
    e.preventDefault();
    Browser.OpenURL(identifierUrl(id));
  }

  let copiedId = $state(null);
  let copiedTimer;

  function copyIdentifier(id) {
    navigator.clipboard.writeText(identifierUrl(id));
    copiedId = id;
    clearTimeout(copiedTimer);
    copiedTimer = setTimeout(() => (copiedId = null), 1200);
  }

  let lines = $state([]);
  let container = $state();
  let stickToBottom = $state(true);

  DashboardService.GetRecentLogs().then((initial) => {
    lines = initial;
    scrollToBottom();
  });

  const unlisten = Events.On('log:line', (evt) => {
    lines = [...lines, evt.data].slice(-500);
    scrollToBottom();
  });

  onDestroy(() => {
    unlisten();
    clearTimeout(copiedTimer);
  });

  async function scrollToBottom() {
    if (!stickToBottom) return;
    await tick();
    if (container) container.scrollTop = container.scrollHeight;
  }

  function handleScroll() {
    if (!container) return;
    const atBottom =
      container.scrollHeight - container.scrollTop - container.clientHeight < 20;
    stickToBottom = atBottom;
  }
</script>

<div class="log-tail" bind:this={container} onscroll={handleScroll}>
  {#each lines as line}
    <div class="line level-{line.Level}">
      <span class="time">{line.Time}</span>
      <span class="level">{line.Level}</span>
      <span class="message">
        {#each splitMessage(line.Message) as part}
          {#if part.type === 'identifier'}
            <a
              class="identifier-link"
              href={identifierUrl(part.value)}
              onclick={(e) => openIdentifier(e, part.value)}
            >{part.value}</a
            ><button
              class="copy-btn"
              type="button"
              onclick={() => copyIdentifier(part.value)}
              aria-label="Copy identifier link"
              title="Copy identifier link"
            >{copiedId === part.value ? 'copied' : 'copy'}</button
            >
          {:else}{part.value}{/if}
        {/each}
      </span>
    </div>
  {/each}
</div>

<style>
  .log-tail {
    flex: 1;
    overflow-y: auto;
    padding: 0.9rem 1.25rem;
    font-family: var(--font-mono);
    font-size: 0.8rem;
    line-height: 1.7;
    background: var(--bg-sunken);
  }
  .line {
    display: flex;
    gap: 0.6rem;
    white-space: pre-wrap;
    word-break: break-word;
  }
  .time {
    flex-shrink: 0;
    color: var(--text-faint);
  }
  .level {
    flex-shrink: 0;
    width: 4.5em;
    text-transform: uppercase;
    font-weight: 600;
    color: var(--text-muted);
  }
  .message {
    color: var(--text);
  }
  .identifier-link {
    color: var(--amber-bright);
    text-decoration: none;
    border-bottom: 1px solid rgba(242, 200, 119, 0.35);
  }
  .identifier-link:hover {
    border-bottom-color: var(--amber-bright);
  }
  .copy-btn {
    margin-left: 0.35rem;
    padding: 0 0.35rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--text-faint);
    font-family: inherit;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.03em;
    line-height: 1.4;
    cursor: pointer;
  }
  .copy-btn:hover {
    color: var(--amber);
    border-color: var(--border-strong);
  }
  .level-error .level,
  .level-fatal .level,
  .level-panic .level {
    color: var(--ember);
  }
  .level-error .message,
  .level-fatal .message,
  .level-panic .message {
    color: var(--ember);
  }
  .level-warning .level {
    color: var(--amber);
  }
  .level-warning .message {
    color: var(--amber);
  }
</style>
