import { fetchAgents, fetchMetrics, fetchSupervisorRuns } from "../lib/api";
import { DashboardClient } from "./dashboard-client";

export const dynamic = "force-dynamic";

export default async function Page() {
  const [agents, initialMetrics, supervisorRuns] = await Promise.all([
    fetchAgents(),
    fetchMetrics().catch((e) => {
      console.error("initial metrics fetch failed:", e);
      return null;
    }),
    fetchSupervisorRuns(6).catch((e) => {
      console.error("initial supervisor fetch failed:", e);
      return [];
    }),
  ]);
  return (
    <DashboardClient
      initialAgents={agents}
      initialMetrics={initialMetrics}
      initialSupervisorRuns={supervisorRuns}
    />
  );
}
