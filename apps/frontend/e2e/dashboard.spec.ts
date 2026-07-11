import { test, expect } from "@playwright/test";

test.describe("EcoMatrix dashboard", () => {
  test("renders KPI tiles + chart + feed on desktop", async ({ page }, testInfo) => {
    const consoleErrors: string[] = [];
    page.on("console", (m) => {
      if (m.type() === "error") consoleErrors.push(m.text());
    });
    page.on("pageerror", (e) => consoleErrors.push(`pageerror: ${e.message}`));

    await page.goto("/");
    // Hero copy.
    await expect(page.getByRole("heading", { name: "上帝视角" })).toBeVisible();
    // KPI labels.
    await expect(page.getByText("存活 Agent")).toBeVisible();
    await expect(page.getByText("全网总资产")).toBeVisible();
    await expect(page.getByText("近 10s QPS")).toBeVisible();
    await expect(page.getByText("在线观测端")).toBeVisible();
    // Agents list / chart presence (initial RSC fetch from /v1/agents).
    await expect(page.getByText("财富分布 · TOP 12")).toBeVisible();
    await expect(page.getByText("赛博交易广播")).toBeVisible();
    await expect(page.getByText("公民一览")).toBeVisible();
    // Wait for WS to connect or until timeout (the dev backend might be offline).
    // If no backend, the page still renders; tiles default to 0.
    await page.waitForTimeout(800);
    await page.screenshot({
      path: `test-results/dashboard-${testInfo.project.name}.png`,
      fullPage: true,
    });
    expect(consoleErrors, consoleErrors.join("\n")).toEqual([]);
  });

  test("agent detail page renders vitals + recent trades", async ({ page }, testInfo) => {
    await page.goto("/agents/agent_miner_01");
    await expect(page.getByRole("heading", { name: "agent_miner_01" })).toBeVisible();
    await expect(page.getByText("基础状态")).toBeVisible();
    await expect(page.getByText("BALANCE")).toBeVisible();
    await page.screenshot({
      path: `test-results/agent-${testInfo.project.name}.png`,
      fullPage: true,
    });
  });
});
