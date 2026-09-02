import { test, expect, type APIRequestContext } from '@playwright/test';

const API_KEY = 'e2e-test-key';
const backendPort = process.env.E2E_BACKEND_PORT || '18080';
const backendBase = `http://127.0.0.1:${backendPort}/api`;

async function createTestUser(request: APIRequestContext) {
  const email = `e2e-${Date.now()}@example.com`;
  const res = await request.post(`${backendBase}/users`, {
    headers: { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' },
    data: { email, traffic_gb: 5, expire_days: 14 },
  });
  expect(res.ok()).toBeTruthy();
  return res.json() as Promise<{ id: number; email: string }>;
}

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
    const res = await request.get(`${backendBase}/docs`);
    expect(res.ok()).toBeTruthy();
    const body = await res.text();
    expect(body).toContain('swagger-ui');
  });
});

test.describe('RioNexTunnel subscription', () => {
  let userId: number;

  test.beforeEach(async ({ request }) => {
    const user = await createTestUser(request);
    userId = user.id;
  });

  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.getByPlaceholder('API key').fill(API_KEY);
    await page.getByRole('button', { name: 'Login' }).click();
    await expect(page).toHaveURL(/\/$/);
  });

  test('user detail page shows subscription section', async ({ page }) => {
    await page.goto(`/users/${userId}`);
    await expect(page.getByRole('heading', { level: 2, name: 'Subscription' })).toBeVisible();
    await expect(page.getByRole('heading', { level: 2, name: 'Registered devices' })).toBeVisible();
    await expect(page.getByTestId('subscription-url')).toBeVisible();
  });

  test('subscription URL copy button is present', async ({ page }) => {
    await page.goto(`/users/${userId}`);
    const copyBtn = page.getByTestId('copy-subscription');
    await expect(copyBtn).toBeVisible();
    await expect(copyBtn).toHaveText('Copy subscription');
  });

  test('subscription endpoint returns base64 with vless links', async ({ request }) => {
    const userRes = await request.get(`${backendBase}/users/${userId}`, {
      headers: { 'X-API-Key': API_KEY },
    });
    expect(userRes.ok()).toBeTruthy();
    const user = await userRes.json();
    expect(user.subscription_token).toBeTruthy();

    const subRes = await request.get(`${backendBase}/subscription/${user.subscription_token}`);
    expect(subRes.status()).toBe(200);
    const body = await subRes.text();
    expect(body.length).toBeGreaterThan(0);
    const decoded = Buffer.from(body.trim(), 'base64').toString('utf-8');
    expect(decoded).toMatch(/vless:/i);
  });

  test('device registration appears in user detail', async ({ request, page }) => {
    const regRes = await request.post(`${backendBase}/client/register`, {
      headers: { 'Content-Type': 'application/json', 'X-API-Version': 'v1' },
      data: { user_id: userId, label: 'e2e-phone' },
    });
    expect(regRes.status()).toBe(201);

    await page.goto(`/users/${userId}`);
    await expect(page.getByText('e2e-phone')).toBeVisible();
    await expect(page.getByText('Revoke')).toBeVisible();
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

  test('stealth settings load from backend API', async ({ page, request }) => {
    const apiRes = await request.get(`${backendBase}/stealth/settings`, {
      headers: { 'X-API-Key': API_KEY },
    });
    expect(apiRes.ok()).toBeTruthy();

    await page.goto('/stealth');
    const destInput = page.getByTestId('reality-dest');
    await expect(destInput).not.toHaveValue('');
  });
});
