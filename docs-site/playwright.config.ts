import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  fullyParallel: false,
  reporter: 'list',
  outputDir: 'artifacts/test-results',
  use: {
    baseURL: 'http://127.0.0.1:18080',
    ...devices['Desktop Chrome'],
    screenshot: 'off',
    trace: 'retain-on-failure',
  },
  webServer: [
    {
      command: 'go run ./cmd/nomyr demo --listen 127.0.0.1:18080',
      cwd: '..',
      url: 'http://127.0.0.1:18080/api/v1/healthz',
      reuseExistingServer: false,
      timeout: 120_000,
      env: { ...process.env },
    },
    {
      command: 'npm run serve -- --dir build --host 127.0.0.1 --port 18081 --no-open',
      url: 'http://127.0.0.1:18081/',
      reuseExistingServer: false,
      timeout: 120_000,
    },
  ],
});
