import { test, expect } from '@playwright/test';

const API_KEY = 'e2e-test-key';

test.describe('proxy-mgr panel', () => {
  test('login and view dashboard', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByRole('heading', { name: 'proxy-mgr' })).toBeVisible();

    await page.getByPlaceholder('API key').fill(API_KEY);
    await page.getByRole('button', { name: 'Login' }).click();

    await expect(page).toHaveURL(/\/$/);
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
  });

  test('api docs are reachable via backend proxy in dev', async ({ request }) => {
    const backendPort = process.env.E2E_BACKEND_PORT || '18080';
    const res = await request.get(`http://127.0.0.1:${backendPort}/api/docs`);
    expect(res.ok()).toBeTruthy();
    const body = await res.text();
    expect(body).toContain('swagger-ui');
  });
});
