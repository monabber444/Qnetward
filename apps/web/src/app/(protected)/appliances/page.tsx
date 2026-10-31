import { ApplianceWorkspace } from "@/features/appliance/appliance-workspace";

export const metadata = {
  title: "Micro-Probes & 5G OOBM | Qnetward",
  description: "Plug & play hardware micro-probe appliances, Zero-Touch Provisioning (ZTP), and cellular 5G out-of-band management.",
};

export default function AppliancesPage() {
  return <ApplianceWorkspace />;
}
