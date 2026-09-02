import { test, expect } from '@playwright/test';

const API_KEY = 'e2e-test-key';

test.describe('RioNexGate panel', () => {
  test('login and view dashboard', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByRole('heading', { name: 'RioNexGate' })).toBeVisible();

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

test.describe('RioNexTunnel subscription', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.getByPlaceholder('API key').fill(API_KEY);
    await page.getByRole('button', { name: 'Login' }).click();
    await expect(page).toHaveURL(/\/$/);
  });

  test('user detail page shows subscription section', async ({ page }) => {
    await page.goto('/users');
    await expect(page.getByRole('heading', { name: 'Users' })).toBeVisible();

    const detailsLink = page.getByRole('link', { name: 'Details' }).first();
    if (await detailsLink.isVisible()) {
      await detailsLink.click();
      await expect(page.getByRole('heading', { level: 2, name: 'Subscription' })).toBeVisible();
      await expect(page.getByRole('heading', { level: 2, name: 'Registered devices' })).toBeVisible();
    }
  });

  test('subscription URL copy button is present', async ({ page }) => {
    await page.goto('/users');
    const detailsLink = page.getByRole('link', { name: 'Details' }).first();
    if (await detailsLink.isVisible()) {
      await detailsLink.click();
      const copyBtn = page.getByTestId('copy-subscription');
      if (await copyBtn.isVisible()) {
        await copyBtn.click();
        await expect(copyBtn).toHaveText('Copied!');
      }
    }
  });

  test('subscription endpoint returns base64 when token exists', async ({ request }) => {
    const backendPort = process.env.E2E_BACKEND_PORT || '18080';

    const usersRes = await request.get(`http://127.0.0.1:${backendPort}/api/users`, {
      headers: { 'X-API-Key': API_KEY },
    });
    expect(usersRes.ok()).toBeTruthy();
    const users = await usersRes.json();

    if (users.length > 0) {
      const userRes = await request.get(`http://127.0.0.1:${backendPort}/api/users/${users[0].id}`, {
        headers: { 'X-API-Key': API_KEY },
      });
      if (userRes.ok()) {
        const user = await userRes.json();
        if (user.subscription_token) {
          const subRes = await request.get(
            `http://127.0.0.1:${backendPort}/api/subscription/${user.subscription_token}`,
          );
          if (subRes.status() === 200) {
            const body = await subRes.text();
            expect(body.length).toBeGreaterThan(0);
            const decoded = Buffer.from(body.trim(), 'base64').toString('utf-8');
            expect(decoded).toMatch(/vless:|vmess:|trojan:/i);
          }
        }
      }
    }
  });
});

test.describe('Stealth page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.getByPlaceholder('API key').fill(API_KEY);
    await page.getByRole('button', { name: 'Login' }).click();
    await expect(page).toHaveURL(/\/$/);
  });

  test('stealth page loads with transport presets', async ({ page }) => {
    await page.goto('/stealth');
    await expect(page.getByRole('heading', { name: 'Transports / Stealth' })).toBeVisible();
    await expect(page.getByText('VLESS + Reality + XHTTP')).toBeVisible();
    await expect(page.getByTestId('reality-dest')).toBeVisible();
    await expect(page.getByTestId('test-dest')).toBeVisible();
  });
});
