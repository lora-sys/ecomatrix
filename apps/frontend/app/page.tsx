import { fetchAgents, fetchMetrics } from "../lib/api";
import { DashboardClient } from "./dashboard-client";

export const dynamic = "force-dynamic";

export default async function Page() {
  // Fetch in parallel: agents and the initial metrics snapshot. The dashboard
  // would otherwise show zeros on first paint while LiveProvider's first poll
  // is still in flight.
  const [agents, initialMetrics] = await Promise.all([
    fetchAgents(),
    fetchMetrics().catch((e) => {
      console.error("initial metrics fetch failed:", e);
      return null;
    }),
  ]);
  return <DashboardClient initialAgents={agents} initialMetrics={initialMetrics} />;
}
