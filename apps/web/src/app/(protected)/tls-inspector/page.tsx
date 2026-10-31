import { TlsInspectorClient } from "@/features/diagnostics/tls-inspector-client";

export const metadata = {
  title: "TLS & HTTP/3 Synthetics | Qnetward",
  description: "Cryptographic TLS 1.3 grading, ALPN inspection, and HTTP/3 QUIC validation.",
};

export default function TlsInspectorPage() {
  return <TlsInspectorClient />;
}
