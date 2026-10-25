import { expect, test } from '@playwright/test';

test('synthetic demo is watermarked; captures default and injected palette fixtures', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByText(/demo data/i).first()).toBeVisible();
  await expect(page.getByText(/synthetic/i).first()).toBeVisible();
  await page.screenshot({ path: 'artifacts/synthetic-demo-default-theme.png', fullPage: true });
  const lightPaletteFixture = await page.addStyleTag({
    content: ':root { color-scheme: light; --canvas: #f3f7f6; --surface: #ffffff; --surface-raised: #e5efec; --text: #142b2d; --muted: #526466; --accent: #08775d; --border: #c6d6d1; }',
  });
  await expect(page.locator('html')).toHaveCSS('color-scheme', 'light');
  await page.screenshot({ path: 'artifacts/synthetic-demo-injected-light-palette-fixture.png', fullPage: true });
  await lightPaletteFixture.evaluate((style) => style.remove());
});

test('built documentation search works locally without external requests', async ({ page }) => {
  const externalRequests: string[] = [];
  await page.route('**/*', async (route) => {
    const requestURL = new URL(route.request().url());
    if (requestURL.origin === 'http://127.0.0.1:18081') {
      await route.continue();
      return;
    }
    externalRequests.push(requestURL.origin);
    await route.abort();
  });

  await page.goto('http://127.0.0.1:18081/');
  const searchInput = page.getByRole('textbox', { name: 'Search' });
  await expect(searchInput).toBeVisible();
  await searchInput.fill('loopback');
  await expect(page.getByRole('option', { name: /Local quickstart/i }).first()).toBeVisible();
  expect(externalRequests).toEqual([]);
});
