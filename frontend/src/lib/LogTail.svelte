<script>
  import { onDestroy, tick } from 'svelte';
  import { Browser, Events } from '@wailsio/runtime';
  import { DashboardService } from '../../bindings/github.com/ao-data/albiondata-client/internal/dashboard/index.js';

  // Only a UUID immediately after "Identifier: " is a real data
  // identifier (see the various "...to ingest (Identifier: %s)" log
  // sites in client/*.go) - a bare UUID-shaped string elsewhere in a log
  // line is just as likely to be something unrelated, like a Windows
  // network device GUID (e.g. "Will listen to these devices: ...").
  const IDENTIFIER_RE = /Identifier: ([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})/gi;
  const URL_RE = /https?:\/\/\S+/gi;

  function identifierUrl(id) {
    return `https://www.albion-online-data.com/identifier?identifier=${id}`;
  }

  // Finds identifier and URL matches in a log message, each as a
  // {start, end, type, value} span over the original string. Only the
  // UUID itself (not the "Identifier: " prefix) is included in an
  // identifier span, so that prefix stays plain text.
  function findLinkSpans(message) {
    const matches = [];

    for (const m of message.matchAll(IDENTIFIER_RE)) {
      const start = m.index + m[0].indexOf(m[1]);
      matches.push({ start, end: start + m[1].length, type: 'identifier', value: m[1] });
    }
    for (const m of message.matchAll(URL_RE)) {
      // Trim trailing punctuation that's almost certainly sentence/log
      // formatting rather than part of the URL (e.g. the closing paren
      // in "see https://example.com/foo for more").
      let value = m[0];
      let end = m.index + value.length;
      while (value.length && /[).,;:!?]/.test(value[value.length - 1])) {
        value = value.slice(0, -1);
        end -= 1;
      }
      matches.push({ start: m.index, end, type: 'url', value });
    }

    matches.sort((a, b) => a.start - b.start);

    // Identifiers and URLs have disjoint prefixes ("Identifier: " vs
    // "http(s)://"), so overlap shouldn't happen in practice, but stay
    // defensive rather than render a mangled/duplicated span.
    const kept = [];
    let cursor = 0;
    for (const m of matches) {
      if (m.start < cursor) continue;
      kept.push(m);
      cursor = m.end;
    }
    return kept;
  }

  // Splits a log message into plain-text, identifier, and URL segments
  // so links can be rendered as <a> instead of interpolating raw
  // (untrusted) log text as HTML.
  function splitMessage(message) {
    const parts = [];
    let lastIndex = 0;
    for (const span of findLinkSpans(message)) {
      if (span.start > lastIndex) {
        parts.push({ type: 'text', value: message.slice(lastIndex, span.start) });
      }
      parts.push(span);
      lastIndex = span.end;
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

  function openLink(e, url) {
    e.preventDefault();
    Browser.OpenURL(url);
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
          {:else if part.type === 'url'}
            <a
              class="log-link"
              href={part.value}
              onclick={(e) => openLink(e, part.value)}
            >{part.value}</a>
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
  .identifier-link,
  .log-link {
    color: var(--amber-bright);
    text-decoration: none;
    border-bottom: 1px solid rgba(242, 200, 119, 0.35);
  }
  .identifier-link:hover,
  .log-link:hover {
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
