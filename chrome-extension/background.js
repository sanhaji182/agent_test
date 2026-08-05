// chrome-extension/background.js — Service worker managing recording state and backend sync.
'use strict';

const DEFAULT_BASE_URL = 'http://localhost:8080';
const SETTINGS_KEY = 'gotest_recorder_settings';
const STOP_FLUSH_WAIT_MS = 800; // time for content scripts to flush after STOP_RECORDING
const POST_RETRY_DELAY_MS = 500;

let activeSession = null;
let baseUrl = DEFAULT_BASE_URL;
let apiKey = '';

// Restore settings on startup
chrome.storage.local.get(SETTINGS_KEY, (data) => {
  const s = data[SETTINGS_KEY] || {};
  baseUrl = s.baseUrl || DEFAULT_BASE_URL;
  apiKey = s.apiKey || '';
  if (s.active && s.sessionId) {
    activeSession = { id: s.sessionId, name: s.sessionName || '', url: s.url || '' };
  }
});

// ---- Message handling ----
chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  switch (msg.type) {
    case 'START_RECORDING':
      respondAsync(sendResponse, () => startRecording(msg.name, msg.url));
      return true; // async response
    case 'STOP_RECORDING':
      respondAsync(sendResponse, async () => {
        await stopRecording();
        return { ok: true };
      });
      return true;
    case 'ATTACH_SESSION':
      // Rekam ke session yang sudah dibuat dari dashboard wizard (auto-link).
      respondAsync(sendResponse, () => attachSession(msg.sessionId, msg.name, msg.url));
      return true;
    case 'FLUSH_EVENTS':
      // Keep the service worker alive until the batch is posted: without
      // respondAsync + return true the async fetch loop can be killed on idle.
      respondAsync(sendResponse, () => flushEventsToBackend(msg.sessionId, msg.events));
      return true;
    case 'GET_STATUS':
      sendResponse({
        recording: !!activeSession,
        session: activeSession,
        baseUrl,
        apiKeySet: !!apiKey
      });
      break;
    case 'UPDATE_SETTINGS':
      try {
        updateSettings(msg.settings || {});
        sendResponse({ ok: true });
      } catch (e) {
        sendResponse({ ok: false, error: e.message });
      }
      break;
    case 'LIST_SESSIONS':
      respondAsync(sendResponse, listSessions);
      return true;
  }
});

// ---- Recording lifecycle ----
async function startRecording(name, url) {
  const targetURL = normalizeHTTPURL(url, 'target URL');
  const response = await fetch(`${baseUrl}/api/v1/recording-sessions`, {
    method: 'POST',
    headers: jsonHeaders(),
    body: JSON.stringify({ name, project_path: targetURL, base_url: targetURL, metadata: {} })
  });
  if (!response.ok) throw new Error(`Failed to create session: ${response.statusText}`);
  const session = await response.json();
  activeSession = { id: session.id, name, url: targetURL };
  persistSettings({ active: true, sessionId: session.id, sessionName: name, url: targetURL });
  await broadcastStart(session.id);
  return session;
}

// attachSession mulai merekam ke session yang sudah ada (dibuat dari dashboard
// wizard) — events masuk ke session yang sama, jadi wizard melihatnya live.
async function attachSession(sessionId, name, url) {
  if (!sessionId) throw new Error('session id required');
  const targetURL = url ? normalizeHTTPURL(url, 'target URL') : '';
  // Verify the session exists server-side and is still recording before attaching.
  let remote;
  try {
    const response = await fetch(`${baseUrl}/api/v1/recording-sessions/${sessionId}`, {
      headers: authHeaders()
    });
    if (!response.ok) {
      return { ok: false, error: `Session not found or unreachable (HTTP ${response.status})` };
    }
    const body = await response.json();
    remote = body && body.session ? body.session : body;
  } catch (e) {
    return { ok: false, error: `Could not verify session: ${e.message || e}` };
  }
  if (remote.status !== 'recording') {
    return { ok: false, error: `Session is not recording (status: ${remote.status || 'unknown'})` };
  }
  activeSession = { id: sessionId, name: name || '', url: targetURL };
  persistSettings({ active: true, sessionId, sessionName: activeSession.name, url: targetURL });
  await broadcastStart(sessionId);
  return activeSession;
}

async function broadcastStart(sessionId) {
  // Notify all tabs to start recording. Cross-origin recording is by design:
  // tabs that cannot receive the message are logged, not filtered out.
  const tabs = await chrome.tabs.query({});
  for (const tab of tabs) {
    chrome.tabs.sendMessage(tab.id, { type: 'START_RECORDING', sessionId }).catch(() => {
      console.warn(`Recording: content script not present in tab ${tab.id} (${tab.url || 'unknown url'}), skipped`);
    });
  }
}

async function stopRecording() {
  const session = activeSession;
  activeSession = null;
  persistSettings({ active: false });
  // Broadcast STOP_RECORDING first so content scripts flush their pending
  // events, wait for the flush to reach the backend, THEN mark the session
  // completed — otherwise the flush races the "completed" status.
  const tabs = await chrome.tabs.query({});
  for (const tab of tabs) {
    chrome.tabs.sendMessage(tab.id, { type: 'STOP_RECORDING' }).catch(() => {});
  }
  if (session && session.id) {
    await sleep(STOP_FLUSH_WAIT_MS);
    try {
      await fetch(`${baseUrl}/api/v1/recording-sessions/${session.id}`, {
        method: 'PATCH',
        headers: jsonHeaders(),
        body: JSON.stringify({ status: 'completed' })
      });
    } catch (e) { /* best effort */ }
  }
}

async function flushEventsToBackend(sessionId, events) {
  if (!sessionId || events.length === 0) return;
  for (const ev of events) {
    // Retry each failed POST once after a short delay.
    for (let attempt = 0; attempt < 2; attempt++) {
      try {
        const response = await fetch(`${baseUrl}/api/v1/recording-sessions/${sessionId}/events`, {
          method: 'POST',
          headers: jsonHeaders(),
          body: JSON.stringify(ev)
        });
        if (response.ok) break;
        if (attempt === 0) {
          await sleep(POST_RETRY_DELAY_MS);
          continue;
        }
        console.warn('Failed to send event:', response.status, response.statusText);
      } catch (e) {
        if (attempt === 0) {
          await sleep(POST_RETRY_DELAY_MS);
          continue;
        }
        console.warn('Failed to send event:', e);
      }
    }
  }
}

async function listSessions() {
  const response = await fetch(`${baseUrl}/api/v1/recording-sessions`, {
    headers: authHeaders()
  });
  // Surface auth failures so the popup can prompt for the API key instead of
  // silently showing an empty list.
  if (response.status === 401) return { error: 'unauthorized' };
  if (!response.ok) return [];
  return response.json();
}

// ---- Settings persistence ----
function persistSettings(partial) {
  chrome.storage.local.get(SETTINGS_KEY, (data) => {
    const existing = data[SETTINGS_KEY] || {};
    const updated = { ...existing, ...partial, baseUrl, apiKey };
    chrome.storage.local.set({ [SETTINGS_KEY]: updated });
  });
}

function updateSettings(settings) {
  if (settings.baseUrl) baseUrl = normalizeHTTPURL(settings.baseUrl, 'backend URL').replace(/\/$/, '');
  if (settings.apiKey !== undefined) apiKey = settings.apiKey;
  persistSettings({});
}

function authHeaders() {
  return apiKey ? { 'X-Api-Key': apiKey } : {};
}

function jsonHeaders() {
  return { 'Content-Type': 'application/json', ...authHeaders() };
}

function normalizeHTTPURL(rawURL, label) {
  let parsed;
  try {
    parsed = new URL(String(rawURL || '').trim());
  } catch {
    throw new Error(`Invalid ${label}`);
  }
  if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
    throw new Error(`${label} must start with http:// or https://`);
  }
  parsed.hash = '';
  return parsed.toString();
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function respondAsync(sendResponse, fn) {
  fn()
    .then((result) => { try { sendResponse(result); } catch (e) {} })
    .catch((e) => { try { sendResponse({ ok: false, error: e.message || String(e) }); } catch (e2) {} });
}
