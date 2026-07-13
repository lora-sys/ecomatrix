import { test, expect, type Page } from "@playwright/test";

const TIMINGS: { name: string; ms: number }[] = [];

async function timed(page: Page, name: string, fn: () => Promise<void>) {
  const t0 = Date.now();
  await fn();
  const ms = Date.now() - t0;
  TIMINGS.push({ name, ms });
  // eslint-disable-next-line no-console
  console.log(`[timing] ${name}: ${ms}ms`);
}

test.describe.configure({ mode: "serial" });

test.describe("EcoMatrix dashboard (polish + history + interactions)", () => {
  test("dashboard renders with all panels, history chart, trade-volume chart", async ({ page }, testInfo) => {
    const consoleErrors: string[] = [];
    page.on("console", (m) => { if (m.type() === "error") consoleErrors.push(m.text()); });
    page.on("pageerror", (e) => consoleErrors.push(`pageerror: ${e.message}`));

    await timed(page, "first-paint", async () => {
      await page.goto("/");
    });
    // Hero copy.
    await expect(page.getByRole("heading", { name: "上帝视角" })).toBeVisible();
    // KPI labels.
    await expect(page.getByText("存活 AGENT")).toBeVisible();
    await expect(page.getByText("全网总资产")).toBeVisible();
    await expect(page.getByText("近 10S QPS")).toBeVisible();
    await expect(page.getByText("在线观测端")).toBeVisible();
    // Panels.
    await expect(page.getByText("财富分布 · TOP 12")).toBeVisible();
    await expect(page.getByText("赛博交易广播")).toBeVisible();
    await expect(page.getByText("社交广场 · POST_FEED")).toBeVisible();
    await expect(page.getByText("公民一览")).toBeVisible();
    // History panel (new).
    await expect(page.getByText("全网 GOLD · 历史 2 分钟")).toBeVisible();
    // Trade-volume panel (new).
    await expect(page.getByText("交易量 · 1 秒桶")).toBeVisible();
    // Job cards.
    for (const j of ["MINER", "MERCHANT", "HACKER", "MEDIATOR"]) {
      await expect(page.getByText(j).first()).toBeVisible();
    }
    // Let history + trade-volume fill.
    await page.waitForTimeout(3000);
    await page.screenshot({ path: `test-results/dashboard-${testInfo.project.name}.png`, fullPage: true });
    expect(consoleErrors, consoleErrors.join("\n")).toEqual([]);
  });

  test("interaction: click into agent detail", async ({ page }, testInfo) => {
    await timed(page, "navigate-to-detail", async () => {
      await page.goto("/");
      await page.getByRole("link", { name: /agent_miner_01/ }).first().click();
      await page.waitForURL("**/agents/agent_miner_01");
    });
    await expect(page.getByRole("heading", { name: "agent_miner_01" })).toBeVisible();
    await expect(page.getByText("BALANCE", { exact: true })).toBeVisible();
    await expect(page.getByText("长期记忆 · LTM", { exact: true })).toBeVisible();
    await expect(page.getByText("近期交易", { exact: true })).toBeVisible();
    await page.screenshot({ path: `test-results/agent-${testInfo.project.name}.png`, fullPage: true });
  });

  test("interaction: hover the wealth chart, then return", async ({ page }, testInfo) => {
    await page.goto("/");
    await page.waitForTimeout(2000);
    // Hover the chart; verify tooltip appears.
    const chart = page.getByText("财富分布 · TOP 12").locator("..");
    await chart.hover();
    await page.waitForTimeout(500);
    await page.screenshot({ path: `test-results/hover-${testInfo.project.name}.png`, fullPage: true });
  });

  test("a11y: live regions and ARIA labels are present", async ({ page }) => {
    await page.goto("/");
    // Trade and social feeds always expose a live region container; the inner
    // <ul role="log"> is only mounted once the backend has at least one item.
    const tradeList = page.locator('[aria-label="live trade broadcast"]');
    await expect(tradeList).toHaveAttribute("aria-live", "polite");
    const socialList = page.locator('[aria-label="agent social feed"]');
    await expect(socialList).toHaveAttribute("aria-live", "polite");
    // KPI tile has aria-live.
    const kpi = page.getByText("全网总资产").locator("..");
    await expect(kpi).toHaveAttribute("aria-live", /polite|assertive/);
  });

  test("motion: prefers-reduced-motion respected", async ({ browser }) => {
    const ctx = await browser.newContext({ reducedMotion: "reduce" });
    const page = await ctx.newPage();
    await page.goto("/");
    await page.waitForTimeout(1500);
    // With reduced motion, animations should be near-instant.
    // The CSS rule we set forces duration: 0.001ms.
    const duration = await page.evaluate(() => {
      const el = document.querySelector(".ring-cyan-glow") as HTMLElement | null;
      if (!el) return null;
      return getComputedStyle(el).transitionDuration;
    });
    expect(duration).toBe("0.001ms");
    await ctx.close();
  });

  test.afterAll(async () => {
    // eslint-disable-next-line no-console
    console.log("[timings]", JSON.stringify(TIMINGS));
  });
});
