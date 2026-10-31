import { TracerouteClient } from "@/features/diagnostics/traceroute-client";

export const metadata = {
  title: "Traceroute & MTR | Qnetward",
  description: "Hop-by-hop latency profiling, autonomous system discovery, and path analysis.",
};

export default function TraceroutePage() {
  return <TracerouteClient />;
}
