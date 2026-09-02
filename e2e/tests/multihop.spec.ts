import { test, expect } from '@playwright/test';
const API_KEY = 'e2e-test-key';
test.describe('Multi-hop nodes', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.getByPlaceholder('API key').fill(API_KEY);
    await page.getByRole('button', { name: 'Login' }).click();
  });
  test('nodes page loads', async ({ page }) => {
    await page.goto('/nodes');
    await expect(page.getByRole('heading', { name: 'Multi-hop nodes' })).toBeVisible();
    await expect(page.getByTestId('chain-topology')).toBeVisible();
  });
});
