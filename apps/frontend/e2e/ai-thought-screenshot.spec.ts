import { test } from "@playwright/test";

test("capture agent detail with AI thought trace", async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.goto("http://127.0.0.1:3200/agents/agent_miner_01", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(5000);
  await page.screenshot({
    path: "/home/lora/repos/ecomatrix/docs/evidence/PHASE-5-ai/agent-detail-with-ai-trace.png",
    fullPage: true,
  });
});
