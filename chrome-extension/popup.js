// chrome-extension/popup.js — Popup UI for recording control.
'use strict';

const startBtn = document.getElementById('startBtn');
const stopBtn = document.getElementById('stopBtn');
const statusEl = document.getElementById('status');
const sessionNameInput = document.getElementById('sessionName');
const baseUrlInput = document.getElementById('baseUrl');
const backendUrlInput = document.getElementById('backendUrl');
const apiKeyInput = document.getElementById('apiKey');
const saveSettingsBtn = document.getElementById('saveSettings');
const sessionInfoDiv = document.getElementById('sessionInfo');
const sessionNameDisplay = document.getElementById('sessionNameDisplay');
const sessionIdDisplay = document.getElementById('sessionId');
const sessionListDiv = document.getElementById('sessionList');

// ---- Init ----
chrome.runtime.sendMessage({ type: 'GET_STATUS' }, (status) => {
  updateUI(status || { recording: false });
  if (status && status.baseUrl) backendUrlInput.value = status.baseUrl;
});

loadSessions();

// ---- Events ----
startBtn.addEventListener('click', async () => {
  const name = sessionNameInput.value.trim() || 'Unnamed Session';
  const url = baseUrlInput.value.trim();

  try {
    const session = await chrome.runtime.sendMessage({
      type: 'START_RECORDING',
      name,
      url
    });
    if (session && session.ok === false) throw new Error(session.error || 'Unknown error');
    updateUI({ recording: true, session });
  } catch (e) {
    alert('Failed to start recording: ' + e.message);
  }
});

stopBtn.addEventListener('click', async () => {
  const result = await chrome.runtime.sendMessage({ type: 'STOP_RECORDING' });
  if (result && result.ok === false) {
    alert('Failed to stop recording: ' + (result.error || 'Unknown error'));
    return;
  }
  updateUI({ recording: false, session: null });
  loadSessions();
});

saveSettingsBtn.addEventListener('click', async () => {
  const result = await chrome.runtime.sendMessage({
    type: 'UPDATE_SETTINGS',
    settings: {
      baseUrl: backendUrlInput.value.trim() || 'http://localhost:8080',
      apiKey: apiKeyInput.value.trim()
    }
  });
  if (result && result.ok === false) {
    alert('Failed to save settings: ' + (result.error || 'Unknown error'));
    return;
  }
  saveSettingsBtn.textContent = 'Saved!';
  setTimeout(() => { saveSettingsBtn.textContent = 'Save Settings'; }, 1500);
});

// ---- UI helpers ----
function updateUI(status) {
  if (status.recording && status.session) {
    startBtn.style.display = 'none';
    stopBtn.style.display = 'block';
    statusEl.textContent = 'Recording';
    statusEl.className = 'status recording';
    sessionInfoDiv.style.display = 'block';
    sessionNameDisplay.textContent = status.session.name || '';
    sessionIdDisplay.textContent = status.session.id || '';
  } else {
    startBtn.style.display = 'block';
    stopBtn.style.display = 'none';
    statusEl.textContent = 'Idle';
    statusEl.className = 'status idle';
    sessionInfoDiv.style.display = 'none';
  }
}

async function loadSessions() {
  try {
    const sessions = await chrome.runtime.sendMessage({ type: 'LIST_SESSIONS' });
    if (sessions && sessions.error) {
      sessionListDiv.innerHTML = sessions.error === 'unauthorized'
        ? '<p class="empty">Unauthorized — cek API key di Settings</p>'
        : '<p class="empty">Could not load sessions</p>';
      return;
    }
    if (!sessions || sessions.length === 0) {
      sessionListDiv.innerHTML = '<p class="empty">No sessions yet. Start recording!</p>';
      return;
    }
    sessionListDiv.innerHTML = sessions.map((s) => {
      const attachable = s.status === 'recording' ? `<button class="btn btn-attach" data-attach="${s.id}" data-name="${escapeHtml(s.name)}" data-url="${escapeHtml(s.base_url || '')}">Attach</button>` : '';
      return `<div class="session-item" data-id="${s.id}" title="Click to copy ID">
        <strong>${escapeHtml(s.name)}</strong>${attachable}<br/>
        <span style="color:#64748b">${s.status}</span> &middot;
        <code style="font-size:10px">${s.id}</code>
      </div>`;
    }).join('');

    // Copy session ID on click
    document.querySelectorAll('.session-item').forEach((el) => {
      el.addEventListener('click', (e) => {
        if (e.target.classList.contains('btn-attach')) return; // attach handled below
        const id = el.dataset.id;
        navigator.clipboard.writeText(id).then(() => {
          el.style.background = '#d1fae5';
          setTimeout(() => { el.style.background = ''; }, 800);
        });
      });
    });

    // Attach buttons: rekam ke session yang dibuat dari dashboard wizard.
    document.querySelectorAll('[data-attach]').forEach((btn) => {
      btn.addEventListener('click', async (e) => {
        e.stopPropagation();
        try {
          const session = await chrome.runtime.sendMessage({
            type: 'ATTACH_SESSION',
            sessionId: btn.dataset.attach,
            name: btn.dataset.name,
            url: btn.dataset.url
          });
          if (session && session.ok === false) throw new Error(session.error || 'Unknown error');
          updateUI({ recording: true, session });
        } catch (err) {
          alert('Failed to attach: ' + err.message);
        }
      });
    });
  } catch (e) {
    sessionListDiv.innerHTML = '<p class="empty">Could not load sessions</p>';
  }
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}
