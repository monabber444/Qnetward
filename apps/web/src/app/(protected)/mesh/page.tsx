import { MeshWorkspace } from "@/features/mesh/mesh-workspace";

export const metadata = {
  title: "PulseMesh N×N Matrix | Qnetward",
  description: "Distributed edge-to-edge inter-agent latency matrix and path-health inspection.",
};

export default function MeshPage() {
  return <MeshWorkspace />;
}
