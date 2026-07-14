import { test, expect, type Page } from "@playwright/test";

/**
 * ISS-030 supervisor detail coverage.
 *
 * The CI e2e job seeds only agents, so to test the navigation from dashboard
 * to /supervisor/[id] we POST a run through the backend first. When the
 * world already has runs (e.g. a populated demo world), the POST is
 * idempotent enough that we keep it as the canonical setup.
 */

test.describe.configure({ mode: "serial" });

const TIMINGS: { name: string; ms: number }[] = [];

async function timed(page: Page, name: string, fn: () => Promise<void>) {
  const t0 = Date.now();
  await fn();
  const ms = Date.now() - t0;
  TIMINGS.push({ name, ms });
  // eslint-disable-next-line no-console
  console.log(`[timing] ${name}: ${ms}ms`);
}

const backend = process.env.NEXT_PUBLIC_BACKEND_URL ?? "http://localhost:8080";

interface SeededRun {
  id: number;
  goal: string;
}

async function seedSupervisorRun(): Promise<SeededRun> {
  const started = new Date(Date.now() - 60_000).toISOString();
  const finished = new Date().toISOString();
  const body = {
    goal: `e2e ISS-030 ${Date.now()}`,
    status: "finished",
    error: "",
    warnings: ["e2e seed warning"],
    subtasks: [
      {
        subtask: "trade",
        target_agent: "agent_miner_01",
        target_job_type: "miner",
        reasoning: "e2e seed",
      },
    ],
    worker_results: [
      {
        agent_id: "agent_miner_01",
        receipt: { tx_id: "tx_e2e_seed", amount: 12 },
      },
    ],
    final_summary: "e2e seed run",
    tokens_used: 50,
    tokens_budget: 200,
    started_at: started,
    finished_at: finished,
    duration_ms: 1500,
  };
  const resp = await fetch(`${backend}/v1/supervisor/runs`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!resp.ok) {
    throw new Error(`seed supervisor run failed: ${resp.status} ${await resp.text()}`);
  }
  const created = (await resp.json()) as SeededRun;
  return created;
}

test.describe("EcoMatrix supervisor detail (ISS-030)", () => {
  test("supervisor link in dashboard navigates to detail page", async ({ page }, testInfo) => {
    const consoleErrors: string[] = [];
    page.on("console", (m) => { if (m.type() === "error") consoleErrors.push(m.text()); });
    page.on("pageerror", (e) => consoleErrors.push(`pageerror: ${e.message}`));

    // Seed a run through the backend so the dashboard has a link to follow.
    const seeded = await seedSupervisorRun();

    await timed(page, "goto-dashboard", async () => {
      await page.goto("/");
    });
    await expect(page.getByText("Supervisor 任务日志")).toBeVisible();

    await timed(page, "navigate-to-detail", async () => {
      // Click directly into the seeded run by URL; the dashboard link in the
      // primary panel sets a strong precedent, but a direct nav removes the
      // dependency on the global store hydrating before our click.
      await page.locator(`a[href="/supervisor/${seeded.id}"]`).first().click();
    });
    await expect(page).toHaveURL(new RegExp(`/supervisor/${seeded.id}$`));
    await expect(page.getByText(/Supervisor 运行 #/)).toBeVisible();
    await expect(page.getByText(seeded.goal)).toBeVisible();
    await page.screenshot({ path: `test-results/supervisor-detail-${testInfo.project.name}.png`, fullPage: true });
    expect(consoleErrors, consoleErrors.join("\n")).toEqual([]);
  });

  test("agent detail exposes the recent supervisor section", async ({ page }, testInfo) => {
    const consoleErrors: string[] = [];
    page.on("console", (m) => { if (m.type() === "error") consoleErrors.push(m.text()); });
    page.on("pageerror", (e) => consoleErrors.push(`pageerror: ${e.message}`));

    // Seed a run that mentions agent_miner_01 so the section has content.
    await seedSupervisorRun();

    await timed(page, "navigate-to-agent", async () => {
      await page.goto("/agents/agent_miner_01");
    });
    await expect(page.getByRole("heading", { name: "agent_miner_01" })).toBeVisible();
    await expect(page.getByText(/近期 Supervisor 运行/)).toBeVisible();
    await page.screenshot({ path: `test-results/agent-supervisor-history-${testInfo.project.name}.png`, fullPage: true });
    expect(consoleErrors, consoleErrors.join("\n")).toEqual([]);
  });

  test.afterAll(async () => {
    // eslint-disable-next-line no-console
    console.log("[timings]", JSON.stringify(TIMINGS));
  });
});
