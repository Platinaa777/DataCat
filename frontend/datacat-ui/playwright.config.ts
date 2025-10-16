import { defineConfig } from '@playwright/test';

export default defineConfig({
    testDir: './e2e',
    workers: 1,
    timeout: 60000,
    use: {
        baseURL: process.env['E2E_BASE_URL'] || 'http://localhost:14200',
        browserName: 'chromium',
        launchOptions: process.env['PLAYWRIGHT_CHROMIUM_EXECUTABLE']
            ? { executablePath: process.env['PLAYWRIGHT_CHROMIUM_EXECUTABLE'] }
            : {},
        screenshot: 'only-on-failure',
        trace: 'retain-on-failure',
    },
});
