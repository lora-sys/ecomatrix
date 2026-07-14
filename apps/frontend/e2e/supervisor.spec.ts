import { test, expect } from "@playwright/test";

/**
 * ISS-030 supervisor detail coverage.
 *
 * These tests assume the live backend has at least one supervisor run on file;
 * the e2e runner seeds the world with --scenario supervisor beforehand (see
 * the workflow that drives the dashboard). When no run exists, we fall back
 * to navigating to /supervisor/999999, which produces notFound() and is still
 * a meaningful signal that the route exists.
 */

test.describe.configure({ mode: "serial" });

const TIMINGS: { name: string; ms: number }[] = [];

async function timed(page: import("@playwright/test").Page, name: string, fn: () => Promise<void>) {
  const t0 = Date.now();
  await fn();
  const ms = Date.now() - t0;
  TIMINGS.push({ name, ms });
  // eslint-disable-next-line no-console
  console.log(`[timing] ${name}: ${ms}ms`);
}

test.describe("EcoMatrix supervisor detail (ISS-030)", () => {
  test("supervisor link in dashboard navigates to detail page", async ({ page }, testInfo) => {
    const consoleErrors: string[] = [];
    page.on("console", (m) => { if (m.type() === "error") consoleErrors.push(m.text()); });
    page.on("pageerror", (e) => consoleErrors.push(`pageerror: ${e.message}`));

    await timed(page, "goto-dashboard", async () => {
      await page.goto("/");
    });
    // The supervisor log region must be present (server-rendered).
    await expect(page.locator('[aria-label="supervisor task log"]')).toBeVisible();
    // Look for any supervisor detail link. If none exist (no runs yet), we
    // skip the navigation assertion and use the fallback below.
    const linkCount = await page.locator("[data-supervisor-link]").count();
    if (linkCount === 0) {
      // Synthetic navigation to a non-existent run; expect notFound() to render
      // a sane empty state rather than a 500.
      await timed(page, "navigate-fallback", async () => {
        const resp = await page.goto("/supervisor/999999");
        expect(resp?.status() ?? 0).toBeLessThan(500);
      });
      await expect(page.getByText(/返回仪表板/)).toBeVisible();
      await page.screenshot({ path: `test-results/supervisor-empty-${testInfo.project.name}.png`, fullPage: true });
      return;
    }
    await timed(page, "navigate-to-detail", async () => {
      await page.locator("[data-supervisor-link]").first().click();
    });
    await expect(page).toHaveURL(/\/supervisor\/\d+/);
    // Detail page should render the back-to-dashboard link and the title.
    // The detail panel embeds an extra "← 返回仪表板" link inside the card,
    // so disambiguate with .first().
    await expect(page.getByText(/返回仪表板/).first()).toBeVisible();
    await expect(page.getByText(/Supervisor 运行 #/)).toBeVisible();
    await page.screenshot({ path: `test-results/supervisor-detail-${testInfo.project.name}.png`, fullPage: true });
    expect(consoleErrors, consoleErrors.join("\n")).toEqual([]);
  });

  test("agent detail exposes the recent supervisor section", async ({ page }, testInfo) => {
    const consoleErrors: string[] = [];
    page.on("console", (m) => { if (m.type() === "error") consoleErrors.push(m.text()); });
    page.on("pageerror", (e) => consoleErrors.push(`pageerror: ${e.message}`));

    await timed(page, "navigate-to-agent", async () => {
      await page.goto("/agents/agent_miner_01");
    });
    await expect(page.getByRole("heading", { name: "agent_miner_01" })).toBeVisible();
    // The supervisor section is always rendered; when no runs mention the
    // agent, the empty-state copy still shows.
    await expect(page.getByText(/近期 Supervisor 运行/)).toBeVisible();
    await page.screenshot({ path: `test-results/agent-supervisor-history-${testInfo.project.name}.png`, fullPage: true });
    expect(consoleErrors, consoleErrors.join("\n")).toEqual([]);
  });

  test.afterAll(async () => {
    // eslint-disable-next-line no-console
    console.log("[timings]", JSON.stringify(TIMINGS));
  });
});
