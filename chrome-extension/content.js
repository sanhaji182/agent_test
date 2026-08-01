// chrome-extension/content.js — Captures user interactions on recorded pages.
(function () {
  'use strict';

  const SETTINGS_KEY = 'gotest_recorder_settings';

  let recording = false;
  let sessionId = '';
  let eventQueue = [];
  let flushTimer = null;
  let viewport = { width: window.innerWidth, height: window.innerHeight };

  // ---- Settings ----
  chrome.storage.local.get(SETTINGS_KEY, (data) => {
    const s = data[SETTINGS_KEY] || {};
    recording = !!s.active;
    sessionId = s.sessionId || '';
  });

  chrome.runtime.onMessage.addListener((msg) => {
    if (msg.type === 'START_RECORDING') {
      recording = true;
      sessionId = msg.sessionId;
    } else if (msg.type === 'STOP_RECORDING') {
      recording = false;
      flushEvents();
    }
  });

  // ---- Selector generation ----
  function stableSelector(el) {
    if (!el || el === document.body || el === document.documentElement) return '';
    if (el.dataset && el.dataset.testid) return `[data-testid="${el.dataset.testid}"]`;
    if (el.id) return `#${el.id}`;
    if (el.getAttribute('aria-label')) return `[aria-label="${el.getAttribute('aria-label')}"]`;

    const tag = el.tagName.toLowerCase();
    if (tag === 'button' || tag === 'a') {
      const text = (el.textContent || '').trim().substring(0, 40);
      if (text) return `${tag}:has-text("${text.replace(/"/g, '\\"')}")`;
    }
    if (el.name) return `${tag}[name="${el.name}"]`;

    // CSS path fallback
    const parts = [];
    let cur = el;
    while (cur && cur !== document.body && parts.length < 4) {
      const ct = cur.tagName.toLowerCase();
      if (cur.id) { parts.unshift(`#${cur.id}`); break; }
      let nth = 1;
      let prev = cur.previousElementSibling;
      while (prev) { if (prev.tagName === cur.tagName) nth++; prev = prev.previousElementSibling; }
      parts.unshift(`${ct}:nth-child(${nth})`);
      cur = cur.parentElement;
    }
    return parts.join(' > ');
  }

  // ---- Event queueing ----
  function queueEvent(ev) {
    if (!recording || !sessionId) return;
    eventQueue.push({
      event_type: ev.type,
      selector: ev.selector || '',
      value: ev.value || '',
      url: window.location.href,
      timestamp: new Date().toISOString(),
      metadata: { viewport }
    });
    scheduleFlush();
  }

  function scheduleFlush() {
    if (flushTimer) return;
    flushTimer = setTimeout(() => { flushEvents(); flushTimer = null; }, 2000);
  }

  function flushEvents() {
    if (eventQueue.length === 0) return;
    chrome.runtime.sendMessage({
      type: 'FLUSH_EVENTS',
      sessionId,
      events: eventQueue.splice(0)
    });
  }

  // ---- Event listeners ----
  document.addEventListener('click', (e) => {
    const sel = stableSelector(e.target);
    queueEvent({ type: 'click', selector: sel });
  }, true);

  let inputTimer = null;
  document.addEventListener('input', (e) => {
    clearTimeout(inputTimer);
    inputTimer = setTimeout(() => {
      const sel = stableSelector(e.target);
      queueEvent({ type: 'fill', selector: sel, value: e.target.value });
    }, 500);
  }, true);

  document.addEventListener('change', (e) => {
    if (e.target.tagName === 'SELECT') {
      const sel = stableSelector(e.target);
      queueEvent({ type: 'select', selector: sel, value: e.target.value });
    }
  }, true);

  document.addEventListener('submit', (e) => {
    const sel = stableSelector(e.target);
    queueEvent({ type: 'submit', selector: sel });
  }, true);

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && e.target.tagName === 'INPUT') {
      const sel = stableSelector(e.target);
      queueEvent({ type: 'press', selector: sel, value: 'Enter' });
    }
  }, true);

  window.addEventListener('resize', () => {
    viewport = { width: window.innerWidth, height: window.innerHeight };
  });

  // Shadow DOM support: patch attachShadow to add listeners
  const origAttachShadow = Element.prototype.attachShadow;
  Element.prototype.attachShadow = function (init) {
    const root = origAttachShadow.call(this, init);
    // Re-bind listeners on shadow root elements
    return root;
  };
})();
