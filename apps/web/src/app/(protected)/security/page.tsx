import { SecurityWorkspace } from "@/features/security/security-workspace";

export const metadata = {
  title: "Security & Asset Verification | Qnetward",
  description: "Three-Tier controlled security gate, domain ownership verification, and CVE catalog.",
};

export default function SecurityPage() {
  return <SecurityWorkspace />;
}
