import { defineConfig } from '@playwright/test'

const TEST_DB = 'postgres://postgres:postgres@localhost:5433/xffl?sslmode=disable'
const TEST_GW = 'http://localhost:8190'
// Fixed clock for live-round tests: falls inside Round 3 window (2026-01-15 Adelaide time)
const CLOCK_OVERRIDE = '2026-01-15T10:00:00+10:30'

export default defineConfig({
  testDir: './e2e',
  globalSetup: './e2e/global-setup.ts',
  timeout: 30_000,
  // Tests share one Postgres + AFL/FFL stack (spawned via webServer below). Per-worker
  // DB isolation per Playwright docs would require per-worker stacks too — not warranted
  // at this scale. See https://playwright.dev/docs/test-parallel for the worker-isolation
  // pattern we'd adopt when this gets too slow.
  workers: 1,
  reporter: [['html', { open: 'never' }], ['list']],
  use: {
    baseURL: 'http://localhost:3001',
    headless: true,
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    trace: 'on-first-retry',
  },
  projects: [
    { name: 'chromium', use: { browserName: 'chromium' } },
  ],
  webServer: [
    {
      command: `DATABASE_URL="${TEST_DB}" CLOCK_OVERRIDE="${CLOCK_OVERRIDE}" PORT=8180 go run ./cmd/main.go`,
      cwd: '../../services/afl',
      port: 8180,
      timeout: 60_000,
      reuseExistingServer: false,
    },
    {
      command: `DATABASE_URL="${TEST_DB}" CLOCK_OVERRIDE="${CLOCK_OVERRIDE}" PORT=8181 go run ./cmd/main.go`,
      cwd: '../../services/ffl',
      port: 8181,
      timeout: 60_000,
      reuseExistingServer: false,
    },
    {
      command: 'AFL_SERVICE_URL=http://localhost:8180 FFL_SERVICE_URL=http://localhost:8181 CORS_ORIGIN=http://localhost:3001 PORT=8190 go run ./cmd/main.go',
      cwd: '../../services/gateway',
      port: 8190,
      timeout: 60_000,
      reuseExistingServer: false,
    },
    {
      command: `VITE_GATEWAY_URL="${TEST_GW}" npm run dev -- --port 3001`,
      cwd: '.',
      port: 3001,
      timeout: 60_000,
      reuseExistingServer: false,
    },
  ],
})
