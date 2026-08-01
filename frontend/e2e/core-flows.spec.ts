import { expect, test, type Page, type Route } from '@playwright/test'

const API = 'http://127.0.0.1:18080'
const now = '2026-08-01T12:00:00.000Z'

const runFixture = {
  id: 'run-export',
  project_path: 'https://app.example.com',
  requirements: 'Validate checkout and login flows',
  mode: 'simple',
  test_type: 'ui',
  state: 'done',
  test_plan: {
    summary: 'Smoke coverage',
    scenarios: [{ name: 'Login flow', priority: 'high', steps: ['Open app', 'Fill email'] }],
  },
  test_files: [{
    name: 'login.json',
    content: '[{"action":"goto","url":"https://app.example.com"}]',
  }],
  run_result: { passed: 1, failed: 0, total: 1, failures: [] },
  fix_attempts: 0,
  screenshots: [],
  created_at: now,
  updated_at: now,
}

test.beforeEach(async ({ page }) => {
  await page.addInitScript(() => {
    class FakeEventSource extends EventTarget {
      url: string
      readyState = 1
      withCredentials = false
      static CONNECTING = 0
      static OPEN = 1
      static CLOSED = 2
      constructor(url: string) {
        super()
        this.url = url
      }
      close() {
        this.readyState = 2
      }
    }
    // @ts-expect-error test shim for browser E2E without backend SSE server
    window.EventSource = FakeEventSource
  })

  await page.route(`${API}/api/v1/runs/run-export`, json(runFixture))
  await page.route(`${API}/api/v1/runs/run-export/events`, json([
    { id: 'evt-1', run_id: 'run-export', type: 'run_completed', message: 'done', created_at: now },
  ]))
  await page.route(`${API}/api/v1/runs/run-export/recordings`, json([]))
  await page.route(`${API}/api/v1/runs/run-export/visual`, json([]))
  await page.route(`${API}/api/v1/runs/run-export/export-code?language=appium`, json({
    run_id: 'run-export',
    target: 'appium',
    language: 'javascript',
    framework: 'Appium WebdriverIO',
    count: 1,
    scripts: {
      'login.appium.mjs': "import { remote } from 'webdriverio';\n// appium:automationName\n",
    },
    code: "// File: login.appium.mjs\nimport { remote } from 'webdriverio';\n// appium:automationName\n",
  }))
})

test('login exchanges API key and redirects to saved destination', async ({ page }) => {
  await page.route(`${API}/api/v1/auth/login`, async (route) => {
    const body = JSON.parse(route.request().postData() || '{}')
    expect(body.api_key).toBe('e2e-key')
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ status: 'ok' }),
    })
  })

  await page.goto('/login')
  await page.evaluate(() => sessionStorage.setItem('redirect_after_login', '/create'))
  await page.getByLabel(/api key/i).fill('e2e-key')
  await page.getByRole('button', { name: /sign in/i }).click()

  await expect(page).toHaveURL(/\/create$/)
  await expect(page.getByRole('heading', { name: /create tests/i })).toBeVisible()
})

test('create test wizard posts run payload and opens the generated run', async ({ page }) => {
  await page.route(`${API}/api/v1/runs`, async (route) => {
    if (route.request().method() !== 'POST') {
      await route.fallback()
      return
    }
    const body = JSON.parse(route.request().postData() || '{}')
    expect(body.project_path).toBe('https://app.example.com')
    expect(body.test_type).toBe('ui')
    expect(body.mode).toBe('simple')
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ run_id: 'run-export', state: 'idle' }),
    })
  })

  await page.goto('/create')
  await page.getByRole('button', { name: /continue/i }).click()
  await page.getByPlaceholder('https://app.example.com').fill('https://app.example.com')
  await page.getByRole('button', { name: /continue/i }).click()
  await page.getByRole('button', { name: /continue/i }).click()
  await page.getByRole('button', { name: /generate and run/i }).click()

  await expect(page).toHaveURL(/\/runs\/run-export$/)
  await expect(page.getByText(/validate checkout and login flows/i)).toBeVisible()
})

test('run detail exports Appium code through the real browser UI', async ({ page }) => {
  await page.goto('/runs/run-export')

  await expect(page.getByText(/validate checkout and login flows/i)).toBeVisible()
  await page.getByRole('button', { name: /advanced/i }).click()
  await page.locator('select').selectOption('appium')
  await page.getByRole('button', { name: /^export$/i }).click()

  await expect(page.getByText(/javascript - appium webdriverio/i)).toBeVisible()
  await expect(page.getByText(/appium:automationName/i)).toBeVisible()
})

function json(value: unknown) {
  return async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(value),
    })
  }
}
