# GoTest Agent Chrome Extension

Record browser interactions and automatically generate Playwright tests.

## Installation

1. Open Chrome and go to `chrome://extensions/`
2. Enable **Developer mode** (toggle in top-right)
3. Click **Load unpacked** and select this `chrome-extension/` folder

## Usage

1. Click the GoTest Agent icon in your toolbar
2. Enter a **Session Name** and the **Target URL** of your app
3. Click **Start Recording**
4. Interact with your web app — clicks, inputs, navigation are captured
5. Click **Stop Recording** when done
6. Go to your GoTest dashboard to view and generate tests from the recording

## Configuration

- **Backend URL**: Default `http://localhost:8080`. Change if your GoTest Agent runs elsewhere.
- **API Key**: Required if API_KEY is set on the backend.

## Selector Priority

The extension captures stable selectors in this order:
1. `data-testid` attributes
2. Element `id`
3. `aria-label` attributes
4. Text content (for buttons/links)
5. CSS path fallback

## Files

| File | Purpose |
|------|---------|
| `manifest.json` | Chrome Manifest V3 config |
| `background.js` | Service worker — manages state, backend sync |
| `content.js` | Content script — captures interactions |
| `popup.html/css/js` | Extension popup UI |
