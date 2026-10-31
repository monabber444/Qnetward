import { AiopsWorkspace } from "@/features/aiops/aiops-workspace";

export const metadata = {
  title: "AIOps & Self-Healing Playbooks | Qnetward",
  description: "Autonomous remediation playbooks, approval gates, BGP rerouting, and bidirectional ChatOps.",
};

export default function AiopsPage() {
  return <AiopsWorkspace />;
}
