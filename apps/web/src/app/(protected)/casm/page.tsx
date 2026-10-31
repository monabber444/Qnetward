import { CasmWorkspace } from "@/features/casm/casm-workspace";

export const metadata = {
  title: "Attack Surface Management & DAST | Qnetward",
  description: "Continuous Attack Surface Management, Certificate Transparency monitoring, Typosquat radar, and sandboxed DAST.",
};

export default function CasmPage() {
  return <CasmWorkspace />;
}
