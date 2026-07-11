import { fetchAgents } from "../lib/api";
import { DashboardClient } from "./dashboard-client";

export const dynamic = "force-dynamic";

export default async function Page() {
  const agents = await fetchAgents();
  return <DashboardClient initialAgents={agents} />;
}
