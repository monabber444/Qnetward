import { TopologyWorkspace } from "@/features/topology/topology-workspace";

export const metadata = {
  title: "Network Topology Map | Qnetward",
  description: "Dependency-aware network topology graph and upstream alert suppression.",
};

export default function TopologyPage() {
  return <TopologyWorkspace />;
}
