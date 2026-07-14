import { test, expect } from "@playwright/test";

test("final: dashboard with continuous multi-agent activity", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "上帝视角" })).toBeVisible();
  // Wait long enough to capture ticker motion + ambient drift.
  await page.waitForTimeout(5000);
  await page.screenshot({ path: "test-results/final-screenshot.png", fullPage: true });
});
