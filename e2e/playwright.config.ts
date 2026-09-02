import { defineConfig } from '@playwright/test';

const backendPort = process.env.E2E_BACKEND_PORT || '18080';
const frontendPort = process.env.E2E_FRONTEND_PORT || '15173';

export default defineConfig({
  testDir: './tests',
  timeout: 60_000,
  retries: process.env.CI ? 1 : 0,
  use: {
    baseURL: `http://127.0.0.1:${frontendPort}`,
    trace: 'on-first-retry',
  },
  webServer: [
    {
      command: `cd ../backend && CONFIG_PATH=../e2e/config.yaml go run ./cmd/main.go migrate && CONFIG_PATH=../e2e/config.yaml go run ./cmd/main.go`,
      url: `http://127.0.0.1:${backendPort}/api/health`,
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
    },
    {
      command: `cd ../frontend && E2E_BACKEND_PORT=${backendPort} npm run dev -- --host 127.0.0.1 --port ${frontendPort}`,
      url: `http://127.0.0.1:${frontendPort}`,
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
    },
  ],
});
