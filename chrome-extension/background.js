// chrome-extension/background.js — Service worker managing recording state and backend sync.
'use strict';

const DEFAULT_BASE_URL = 'http://localhost:8080';
const SETTINGS_KEY = 'gotest_recorder_settings';

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
      flushEventsToBackend(msg.sessionId, msg.events);
      break;
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
  activeSession = { id: sessionId, name: name || '', url: targetURL };
  persistSettings({ active: true, sessionId, sessionName: activeSession.name, url: targetURL });
  await broadcastStart(sessionId);
  return activeSession;
}

async function broadcastStart(sessionId) {
  // Notify all tabs to start recording
  const tabs = await chrome.tabs.query({});
  for (const tab of tabs) {
    chrome.tabs.sendMessage(tab.id, { type: 'START_RECORDING', sessionId }).catch(() => {});
  }
}

async function stopRecording() {
  if (activeSession && activeSession.id) {
    try {
      await fetch(`${baseUrl}/api/v1/recording-sessions/${activeSession.id}`, {
        method: 'PATCH',
        headers: jsonHeaders(),
        body: JSON.stringify({ status: 'completed' })
      });
    } catch (e) { /* best effort */ }
  }
  activeSession = null;
  persistSettings({ active: false });
  const tabs = await chrome.tabs.query({});
  for (const tab of tabs) {
    chrome.tabs.sendMessage(tab.id, { type: 'STOP_RECORDING' }).catch(() => {});
  }
}

async function flushEventsToBackend(sessionId, events) {
  if (!sessionId || events.length === 0) return;
  for (const ev of events) {
    try {
      await fetch(`${baseUrl}/api/v1/recording-sessions/${sessionId}/events`, {
        method: 'POST',
        headers: jsonHeaders(),
        body: JSON.stringify(ev)
      });
    } catch (e) {
      console.warn('Failed to send event:', e);
    }
  }
}

async function listSessions() {
  const response = await fetch(`${baseUrl}/api/v1/recording-sessions`, {
    headers: authHeaders()
  });
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

function respondAsync(sendResponse, fn) {
  fn()
    .then((result) => sendResponse(result))
    .catch((e) => sendResponse({ ok: false, error: e.message || String(e) }));
}
