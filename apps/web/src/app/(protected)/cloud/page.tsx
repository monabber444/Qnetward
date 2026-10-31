import { CloudWorkspace } from "@/features/cloud/cloud-workspace";

export const metadata = {
  title: "Multi-Cloud & eBPF Telemetry | Qnetward",
  description: "Kernel-level socket tracing, Kubernetes pod metrics, and AWS/GCP/Azure VPC flow intelligence.",
};

export default function CloudPage() {
  return <CloudWorkspace />;
}
