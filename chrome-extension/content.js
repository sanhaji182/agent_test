// chrome-extension/content.js — Captures user interactions on recorded pages.
(function () {
  'use strict';

  const SETTINGS_KEY = 'gotest_recorder_settings';
  const NAV_STORAGE_KEY = 'gotest_recorder_last_url';

  let recording = false;
  let sessionId = '';
  let eventQueue = [];
  let flushTimer = null;
  let viewport = { width: window.innerWidth, height: window.innerHeight };

  // Last known page URL. Persisted in sessionStorage so full-page loads during
  // a recording can be detected (sessionStorage survives navigations within the
  // same tab/origin, unlike this content script instance).
  let lastUrl = '';
  try {
    lastUrl = sessionStorage.getItem(NAV_STORAGE_KEY) || window.location.href;
  } catch (e) {
    lastUrl = window.location.href;
  }

  // Timestamps of recently captured events, so submit can dedupe against them.
  let lastClickAt = 0;
  let lastEnterPressAt = 0;

  // ---- Settings ----
  chrome.storage.local.get(SETTINGS_KEY, (data) => {
    const s = data[SETTINGS_KEY] || {};
    recording = !!s.active;
    sessionId = s.sessionId || '';
    // Recording state is now known: pick up navigations that happened while
    // this script wasn't running yet (e.g. a full page load during recording).
    checkNavigation();
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
  // Minimal CSS escaping: `#id` selectors must escape structural characters,
  // and quoted attribute values must escape `"` and `\`.
  function escapeCSSIdent(value) {
    return String(value)
      .replace(/\\/g, '\\\\')
      .replace(/"/g, '\\"')
      .replace(/'/g, "\\'")
      .replace(/#/g, '\\#')
      .replace(/\./g, '\\.')
      .replace(/:/g, '\\:')
      .replace(/\s/g, '\\ ');
  }

  function escapeCSSAttr(value) {
    return String(value).replace(/\\/g, '\\\\').replace(/"/g, '\\"');
  }

  function stableSelector(el) {
    if (!el || el === document.body || el === document.documentElement) return '';
    if (el.dataset && el.dataset.testid) return `[data-testid="${escapeCSSAttr(el.dataset.testid)}"]`;
    if (el.id) return `#${escapeCSSIdent(el.id)}`;
    if (el.getAttribute('aria-label')) return `[aria-label="${escapeCSSAttr(el.getAttribute('aria-label'))}"]`;

    const tag = el.tagName.toLowerCase();
    if (tag === 'button' || tag === 'a') {
      const text = (el.textContent || '').trim().substring(0, 40);
      if (text) return `${tag}:has-text("${text.replace(/"/g, '\\"')}")`;
    }
    if (el.name) return `${tag}[name="${escapeCSSAttr(el.name)}"]`;

    // CSS path fallback
    const parts = [];
    let cur = el;
    while (cur && cur !== document.body && parts.length < 4) {
      const ct = cur.tagName.toLowerCase();
      if (cur.id) { parts.unshift(`#${escapeCSSIdent(cur.id)}`); break; }
      // :nth-child(n) counts ALL preceding element siblings, not just same-tag ones.
      let nth = 1;
      let prev = cur.previousElementSibling;
      while (prev) { nth++; prev = prev.previousElementSibling; }
      parts.unshift(`${ct}:nth-child(${nth})`);
      cur = cur.parentElement;
    }
    return parts.join(' > ');
  }

  // ---- Navigation tracking ----
  function checkNavigation() {
    const current = window.location.href;
    let stored;
    try { stored = sessionStorage.getItem(NAV_STORAGE_KEY); } catch (e) { stored = lastUrl; }
    if (current === stored) return;
    lastUrl = current;
    try { sessionStorage.setItem(NAV_STORAGE_KEY, current); } catch (e) {}
    if (!recording || !sessionId) return;
    queueEvent({ type: 'navigate', selector: '', value: '', url: current });
  }

  function initNavigationTracking() {
    // SPA navigations through the history API (pushState/replaceState).
    const origPushState = history.pushState;
    const origReplaceState = history.replaceState;
    history.pushState = function (...args) {
      const result = origPushState.apply(this, args);
      checkNavigation();
      return result;
    };
    history.replaceState = function (...args) {
      const result = origReplaceState.apply(this, args);
      checkNavigation();
      return result;
    };
    window.addEventListener('popstate', checkNavigation);
    window.addEventListener('hashchange', checkNavigation);
    // bfcache restore reuses this script instance, so in-memory state is intact.
    window.addEventListener('pageshow', (e) => {
      if (e.persisted) checkNavigation();
    });
  }

  // ---- Event queueing ----
  function queueEvent(ev) {
    if (!recording || !sessionId) return;
    eventQueue.push({
      event_type: ev.type,
      selector: ev.selector || '',
      value: ev.value || '',
      url: ev.url || window.location.href,
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
    }).catch(() => {});
  }

  // ---- Event listeners ----
  document.addEventListener('click', (e) => {
    lastClickAt = Date.now();
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
    // The backend has no `submit` event type. A click on a submit button
    // already produced a `click`, and Enter in an input already produced a
    // `press`; only synthesize `press` Enter for submits that came from
    // neither (e.g. form.submit()).
    if (Date.now() - lastClickAt < 500) return;
    if (Date.now() - lastEnterPressAt < 500) return;
    const submitBtn = e.target.querySelector
      ? e.target.querySelector('button[type="submit"], button:not([type]), input[type="submit"], input[type="image"]')
      : null;
    const sel = submitBtn ? stableSelector(submitBtn) : (stableSelector(e.target) || 'form');
    queueEvent({ type: 'press', selector: sel, value: 'Enter' });
  }, true);

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && e.target.tagName === 'INPUT') {
      lastEnterPressAt = Date.now();
      const sel = stableSelector(e.target);
      queueEvent({ type: 'press', selector: sel, value: 'Enter' });
    }
  }, true);

  window.addEventListener('resize', () => {
    viewport = { width: window.innerWidth, height: window.innerHeight };
  });

  initNavigationTracking();

  // Shadow DOM support: patch attachShadow to add listeners
  const origAttachShadow = Element.prototype.attachShadow;
  Element.prototype.attachShadow = function (init) {
    const root = origAttachShadow.call(this, init);
    // Re-bind listeners on shadow root elements
    return root;
  };
})();
