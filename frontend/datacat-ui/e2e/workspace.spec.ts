import { expect, test } from '@playwright/test';

test('OIDC login, dashboard persistence, and workspace navigation', async ({ page, context, baseURL }) => {
    const errors: string[] = [];
    page.on('pageerror', error => errors.push(error.message));
    page.on('response', response => {
        if (response.url().startsWith(baseURL!) && response.status() >= 500) {
            errors.push(`${response.status()} ${response.url()}`);
        }
    });

    await page.goto('/');
    await page.locator('#username').fill(process.env['E2E_USERNAME'] || 'developer');
    await page.locator('#password').fill(process.env['E2E_PASSWORD'] || 'local-development-only');
    await page.locator('#kc-login').click();
    await page.waitForURL(`${baseURL}/`);
    await expect(page.locator('.screen')).toBeVisible();
    await page.waitForLoadState('networkidle');

    const cookie = (await context.cookies()).find(cookie => cookie.name === 'access_token');
    expect(cookie?.httpOnly).toBe(true);
    expect(await page.evaluate(() => document.cookie)).not.toContain('access_token');

    const headers = { 'X-Namespace': '00000000-0000-0000-0000-000000000000', Origin: baseURL! };
    const name = `E2E dashboard ${Date.now()}`;
    const created = await context.request.post('/api/v1/dashboard/add', { headers, data: { name, tags: ['e2e'] } });
    expect(created.ok()).toBe(true);
    const id: string = await created.json();

    await page.goto('/dashboards');
    await expect(page.getByText(name, { exact: true })).toBeVisible();
    for (const path of ['/alerts', '/notifications', '/data-sources', '/logs', '/traces', '/settings']) {
        await page.goto(path);
        await page.waitForLoadState('networkidle');
        await expect(page.locator('.screen')).toBeVisible();
    }
    await page.goto('/forbidden');
    await expect(page.getByRole('heading', { name: 'Access denied' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Return to dashboards' })).toBeVisible();

    const deleted = await context.request.delete(`/api/v1/dashboard/remove/${id}`, { headers });
    expect(deleted.ok()).toBe(true);
    expect(errors).toEqual([]);
});
