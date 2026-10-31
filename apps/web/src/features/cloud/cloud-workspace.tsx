"use client";

import { useState } from "react";
import {
  Activity,
  ArrowRight,
  CheckCircle2,
  Cloud,
  Cpu,
  Filter,
  Globe,
  Layers,
  Network,
  RefreshCw,
  Server,
  ShieldAlert,
  Zap,
} from "lucide-react";
import { useShellStore } from "@/store/ui-store";
import { translations } from "@/lib/i18n";

type FlowLog = {
  provider: "AWS" | "GCP" | "Azure";
  src: string;
  dst: string;
  srcPort: number;
  dstPort: number;
  proto: string;
  packets: number;
  bytes: number;
  action: "ACCEPT" | "REJECT";
  time: string;
  workload?: string;
};

type EbpfSocket = {
  srcPod: string;
  srcNs: string;
  dstPod: string;
  dstNs: string;
  rttMs: number;
  handshakeMs: number;
  retransmits: number;
  state: string;
};

const sampleFlows: FlowLog[] = [
  { provider: "AWS", src: "10.0.1.50", dst: "172.217.16.206", srcPort: 49152, dstPort: 443, proto: "TCP", packets: 142, bytes: 84200, action: "ACCEPT", time: "19:35:10", workload: "payment-api" },
  { provider: "AWS", src: "198.51.100.4", dst: "10.0.1.50", srcPort: 54321, dstPort: 22, proto: "TCP", packets: 3, bytes: 180, action: "REJECT", time: "19:35:12", workload: "bastion" },
  { provider: "GCP", src: "10.128.0.2", dst: "10.128.0.3", srcPort: 45120, dstPort: 80, proto: "TCP", packets: 280, bytes: 142000, action: "ACCEPT", time: "19:35:14", workload: "ingress-gke" },
  { provider: "Azure", src: "10.0.0.4", dst: "142.250.190.46", srcPort: 49152, dstPort: 443, proto: "TCP", packets: 65, bytes: 32400, action: "ACCEPT", time: "19:35:15", workload: "aks-worker" },
  { provider: "Azure", src: "203.0.113.88", dst: "10.0.0.4", srcPort: 41200, dstPort: 3389, proto: "TCP", packets: 5, bytes: 240, action: "REJECT", time: "19:35:18", workload: "win-vnet" },
  { provider: "GCP", src: "10.128.0.8", dst: "8.8.8.8", srcPort: 53120, dstPort: 53, proto: "UDP", packets: 12, bytes: 960, action: "ACCEPT", time: "19:35:20", workload: "dns-resolver" },
];

const sampleEbpfSockets: EbpfSocket[] = [
  { srcPod: "payment-service-7f8d", srcNs: "production", dstPod: "postgres-primary-0", dstNs: "database", rttMs: 0.42, handshakeMs: 0.85, retransmits: 0, state: "ESTABLISHED" },
  { srcPod: "web-frontend-5b4c", srcNs: "default", dstPod: "api-gateway-3c2a", dstNs: "production", rttMs: 0.65, handshakeMs: 1.10, retransmits: 0, state: "ESTABLISHED" },
  { srcPod: "checkout-worker-1a2b", srcNs: "production", dstPod: "redis-cluster-leader", dstNs: "cache", rttMs: 0.28, handshakeMs: 0.50, retransmits: 0, state: "ESTABLISHED" },
  { srcPod: "auth-service-9e8f", srcNs: "security", dstPod: "ldap-sync-core", dstNs: "internal", rttMs: 1.85, handshakeMs: 2.40, retransmits: 3, state: "WARNING" },
];

export function CloudWorkspace() {
  const { locale } = useShellStore();
  const t = translations[locale];
  const [providerFilter, setProviderFilter] = useState<string>("ALL");
  const [flows, setFlows] = useState<FlowLog[]>(sampleFlows);

  const filteredFlows = flows.filter((f) => providerFilter === "ALL" || f.provider === providerFilter);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/40 pb-5">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-foreground flex items-center gap-2.5">
            <Cloud className="h-5 w-5 text-cyan-400" />
            Multi-Cloud & Linux eBPF Observability
          </h1>
          <p className="text-xs text-muted-foreground mt-1">
            Zero-sidecar kernel TCP/DNS socket tracing and unified AWS, GCP, and Azure VPC flow intelligence.
          </p>
        </div>
        <div className="flex items-center gap-2 text-xs">
          <span className="px-3 py-1 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-medium flex items-center gap-1.5">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
            eBPF Kernel Socket Tracing: Active
          </span>
        </div>
      </div>

      {/* Cloud Providers Status Strip */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {/* AWS */}
        <div className="p-4 rounded-2xl border border-amber-500/30 bg-amber-500/5 shadow-sm space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-amber-400 uppercase tracking-wider">AWS Transit</span>
            <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              CONNECTED
            </span>
          </div>
          <div className="text-sm font-bold font-mono text-foreground">us-east-1 (TGW)</div>
          <div className="space-y-1 text-xs text-muted-foreground pt-1 border-t border-border/40">
            <div className="flex justify-between"><span>VPC Flow Logs:</span> <span className="font-mono text-foreground">Active (v2)</span></div>
            <div className="flex justify-between"><span>DirectConnect:</span> <span className="font-mono text-cyan-300">4.12 ms RTT</span></div>
          </div>
        </div>

        {/* GCP */}
        <div className="p-4 rounded-2xl border border-blue-500/30 bg-blue-500/5 shadow-sm space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-blue-400 uppercase tracking-wider">GCP Interconnect</span>
            <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              CONNECTED
            </span>
          </div>
          <div className="text-sm font-bold font-mono text-foreground">us-central1 (ORD)</div>
          <div className="space-y-1 text-xs text-muted-foreground pt-1 border-t border-border/40">
            <div className="flex justify-between"><span>VPC Egress:</span> <span className="font-mono text-foreground">2.4 Gbps</span></div>
            <div className="flex justify-between"><span>Cloud NAT Ports:</span> <span className="font-mono text-cyan-300">4,280 Alloc</span></div>
          </div>
        </div>

        {/* Azure */}
        <div className="p-4 rounded-2xl border border-cyan-500/30 bg-cyan-500/5 shadow-sm space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-cyan-400 uppercase tracking-wider">Azure ExpressRoute</span>
            <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              CONNECTED
            </span>
          </div>
          <div className="text-sm font-bold font-mono text-foreground">eastus (NSG Flow)</div>
          <div className="space-y-1 text-xs text-muted-foreground pt-1 border-t border-border/40">
            <div className="flex justify-between"><span>Circuit Bandwidth:</span> <span className="font-mono text-foreground">450 Mbps</span></div>
            <div className="flex justify-between"><span>Packet Drop:</span> <span className="font-mono text-emerald-400">0.0 %</span></div>
          </div>
        </div>

        {/* eBPF K8s */}
        <div className="p-4 rounded-2xl border border-purple-500/30 bg-purple-500/5 shadow-sm space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-purple-400 uppercase tracking-wider">eBPF DaemonSet</span>
            <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              CO-RE LOADED
            </span>
          </div>
          <div className="text-sm font-bold font-mono text-foreground">Kernel Tracing</div>
          <div className="space-y-1 text-xs text-muted-foreground pt-1 border-t border-border/40">
            <div className="flex justify-between"><span>Active Pods:</span> <span className="font-mono text-foreground">12 Enriched</span></div>
            <div className="flex justify-between"><span>Retransmit Rate:</span> <span className="font-mono text-emerald-400">0.02 %</span></div>
          </div>
        </div>
      </div>

      {/* eBPF Pod-to-Pod Service Mesh Section */}
      <div className="rounded-2xl border border-border/60 bg-card p-5 shadow-sm space-y-4">
        <div className="flex items-center justify-between border-b border-border/40 pb-3">
          <div className="flex items-center gap-2">
            <Cpu className="h-4 w-4 text-purple-400" />
            <h2 className="text-sm font-semibold text-foreground">
              eBPF Kernel Socket Tracing — Pod-to-Pod Kubernetes Service Mesh
            </h2>
          </div>
          <span className="text-xs font-mono text-muted-foreground">Tracepoints: sock:inet_sock_set_state</span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse text-xs">
            <thead>
              <tr className="border-b border-border/60 text-muted-foreground font-semibold">
                <th className="p-2.5">Source Workload</th>
                <th className="p-2.5">Destination Workload</th>
                <th className="p-2.5">TCP Handshake</th>
                <th className="p-2.5">Kernel RTT</th>
                <th className="p-2.5">Retransmits</th>
                <th className="p-2.5">Socket State</th>
              </tr>
            </thead>
            <tbody>
              {sampleEbpfSockets.map((s, idx) => (
                <tr key={idx} className="border-b border-border/40 hover:bg-muted/30 transition">
                  <td className="p-2.5">
                    <div className="font-mono font-medium text-foreground">{s.srcPod}</div>
                    <div className="text-[10px] text-muted-foreground font-mono">ns: {s.srcNs}</div>
                  </td>
                  <td className="p-2.5">
                    <div className="font-mono font-medium text-foreground">{s.dstPod}</div>
                    <div className="text-[10px] text-muted-foreground font-mono">ns: {s.dstNs}</div>
                  </td>
                  <td className="p-2.5 font-mono text-cyan-300 font-semibold">{s.handshakeMs.toFixed(2)} ms</td>
                  <td className="p-2.5 font-mono text-foreground">{s.rttMs.toFixed(2)} ms</td>
                  <td className="p-2.5">
                    {s.retransmits > 0 ? (
                      <span className="font-mono font-bold text-amber-400">{s.retransmits} Drops</span>
                    ) : (
                      <span className="font-mono text-emerald-400">0</span>
                    )}
                  </td>
                  <td className="p-2.5">
                    <span
                      className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                        s.state === "ESTABLISHED"
                          ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                          : "bg-amber-500/20 text-amber-300 border border-amber-500/30"
                      }`}
                    >
                      {s.state}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Multi-Cloud VPC Flow Logs Stream */}
      <div className="rounded-2xl border border-border/60 bg-card p-5 shadow-sm space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-border/40 pb-3">
          <div className="flex items-center gap-2">
            <Layers className="h-4 w-4 text-cyan-400" />
            <h2 className="text-sm font-semibold text-foreground">
              Normalized Multi-Cloud VPC Flow Ingestion Stream
            </h2>
          </div>

          <div className="flex items-center gap-2">
            <Filter className="h-3.5 w-3.5 text-muted-foreground" />
            {["ALL", "AWS", "GCP", "Azure"].map((p) => (
              <button
                key={p}
                type="button"
                onClick={() => setProviderFilter(p)}
                className={`px-2.5 py-1 rounded-lg text-xs font-semibold transition cursor-pointer ${
                  providerFilter === p
                    ? "bg-cyan-500/20 text-cyan-300 border border-cyan-500/30"
                    : "bg-secondary/40 text-muted-foreground hover:bg-secondary hover:text-foreground"
                }`}
              >
                {p}
              </button>
            ))}
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse text-xs">
            <thead>
              <tr className="border-b border-border/60 text-muted-foreground font-semibold">
                <th className="p-2.5">Provider</th>
                <th className="p-2.5">Timestamp</th>
                <th className="p-2.5">Source IP:Port</th>
                <th className="p-2.5">Destination IP:Port</th>
                <th className="p-2.5">Protocol</th>
                <th className="p-2.5">Packets / Bytes</th>
                <th className="p-2.5">Security Decision</th>
              </tr>
            </thead>
            <tbody>
              {filteredFlows.map((f, idx) => (
                <tr key={idx} className="border-b border-border/40 hover:bg-muted/30 transition">
                  <td className="p-2.5">
                    <span
                      className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                        f.provider === "AWS"
                          ? "bg-amber-500/20 text-amber-300"
                          : f.provider === "GCP"
                          ? "bg-blue-500/20 text-blue-300"
                          : "bg-cyan-500/20 text-cyan-300"
                      }`}
                    >
                      {f.provider}
                    </span>
                  </td>
                  <td className="p-2.5 font-mono text-muted-foreground">{f.time}</td>
                  <td className="p-2.5 font-mono text-foreground font-medium">
                    {f.src}:{f.srcPort}
                  </td>
                  <td className="p-2.5 font-mono text-foreground font-medium">
                    {f.dst}:{f.dstPort}
                  </td>
                  <td className="p-2.5 font-mono">{f.proto}</td>
                  <td className="p-2.5 font-mono text-muted-foreground">
                    {f.packets} pkts ({f.bytes > 1024 ? (f.bytes / 1024).toFixed(1) + " KB" : f.bytes + " B"})
                  </td>
                  <td className="p-2.5">
                    <span
                      className={`px-2 py-0.5 rounded text-[10px] font-black ${
                        f.action === "ACCEPT"
                          ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                          : "bg-rose-500/20 text-rose-300 border border-rose-500/30"
                      }`}
                    >
                      {f.action}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
