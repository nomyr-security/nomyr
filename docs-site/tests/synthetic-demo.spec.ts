import { expect, test } from '@playwright/test';

test('synthetic demo is watermarked and screenshots light and dark modes', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByText(/demo data/i).first()).toBeVisible();
  await expect(page.getByText(/synthetic/i).first()).toBeVisible();
  const lightTheme = await page.addStyleTag({
    content: ':root { color-scheme: light; --canvas: #f3f7f6; --surface: #ffffff; --surface-raised: #e5efec; --text: #142b2d; --muted: #526466; --accent: #08775d; --border: #c6d6d1; }',
  });
  const lightBackground = await page.locator('body').evaluate((element) => getComputedStyle(element).backgroundColor);
  await page.screenshot({ path: 'artifacts/synthetic-demo-light.png', fullPage: true });
  await lightTheme.evaluate((style) => style.remove());
  const darkBackground = await page.locator('body').evaluate((element) => getComputedStyle(element).backgroundColor);
  expect(darkBackground).not.toBe(lightBackground);
  await page.screenshot({ path: 'artifacts/synthetic-demo-dark.png', fullPage: true });
});
