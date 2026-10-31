import { MspWorkspace } from "@/features/msp/msp-workspace";

export const metadata = {
  title: "MSP Portal & Compliance | Qnetward",
  description: "Multi-tenant MSP partner hierarchy, white-label branding, SHA-256 cryptographic audit ledger, and enterprise SSO.",
};

export default function MspPortalPage() {
  return <MspWorkspace />;
}
