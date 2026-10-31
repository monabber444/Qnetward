import { BgpTwinWorkspace } from "@/features/bgp/bgp-twin-workspace";

export const metadata = {
  title: "BGP Routing & Digital Twin | Qnetward",
  description: "Global BGP monitoring, RPKI route validation, AS-path visualizer, and in-memory failure blast radius simulation.",
};

export default function BgpTwinPage() {
  return <BgpTwinWorkspace />;
}
