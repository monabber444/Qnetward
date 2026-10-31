"use client";

import { useState } from "react";
import { Activity, ArrowRight, Play, RefreshCw, ShieldAlert, Terminal } from "lucide-react";
import { useShellStore } from "@/store/ui-store";
import { translations } from "@/lib/i18n";

type Hop = {
  hop: number;
  ip: string;
  host: string;
  asn: string;
  lossPct: number;
  sent: number;
  lastMs: number;
  avgMs: number;
  bestMs: number;
  worstMs: number;
  stdDev: number;
};

const defaultHops: Hop[] = [
  { hop: 1, ip: "192.168.0.1", host: "local-gateway.lan", asn: "Private", lossPct: 0, sent: 10, lastMs: 0.42, avgMs: 0.45, bestMs: 0.38, worstMs: 0.62, stdDev: 0.08 },
  { hop: 2, ip: "100.64.0.1", host: "cgnat-gw.isp.net", asn: "AS13335", lossPct: 0, sent: 10, lastMs: 2.15, avgMs: 2.30, bestMs: 1.95, worstMs: 3.10, stdDev: 0.32 },
  { hop: 3, ip: "172.16.10.5", host: "core-agg-01.isp.net", asn: "AS13335", lossPct: 0, sent: 10, lastMs: 4.80, avgMs: 4.92, bestMs: 4.20, worstMs: 6.15, stdDev: 0.45 },
  { hop: 4, ip: "198.32.160.1", host: "ixp-transit.equinix.com", asn: "AS24115", lossPct: 0, sent: 10, lastMs: 8.42, avgMs: 8.65, bestMs: 8.10, worstMs: 11.20, stdDev: 0.85 },
  { hop: 5, ip: "142.250.160.22", host: "google-edge.1e100.net", asn: "AS15169", lossPct: 0, sent: 10, lastMs: 12.10, avgMs: 12.35, bestMs: 11.80, worstMs: 14.50, stdDev: 0.72 },
  { hop: 6, ip: "8.8.8.8", host: "dns.google", asn: "AS15169", lossPct: 0, sent: 10, lastMs: 12.45, avgMs: 12.60, bestMs: 12.15, worstMs: 13.90, stdDev: 0.51 },
];

export function TracerouteClient() {
  const { locale } = useShellStore();
  const t = translations[locale];

  const [target, setTarget] = useState("8.8.8.8");
  const [protocol, setProtocol] = useState<"icmp" | "udp" | "tcp">("icmp");
  const [maxHops, setMaxHops] = useState(15);
  const [isRunning, setIsRunning] = useState(false);
  const [hops, setHops] = useState<Hop[]>(defaultHops);

  const runTraceroute = () => {
    setIsRunning(true);
    setTimeout(() => {
      setIsRunning(false);
    }, 1500);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/40 pb-5">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-foreground flex items-center gap-2.5">
            <Terminal className="h-5 w-5 text-cyan-400" />
            {t.nav.traceroute}
          </h1>
          <p className="text-xs text-muted-foreground mt-1">
            Real-time hop-by-hop latency profiling, autonomous system (AS) discovery, and route flapping detection.
          </p>
        </div>
      </div>

      {/* Control Box */}
      <div className="rounded-2xl border border-border/60 bg-card p-5 shadow-sm space-y-4">
        <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
          <div className="sm:col-span-2">
            <label className="text-xs font-medium text-foreground block mb-1">Target Host / IP</label>
            <input
              type="text"
              value={target}
              onChange={(e) => setTarget(e.target.value)}
              className="w-full px-3 py-2 rounded-xl bg-background border border-border text-xs font-mono text-foreground focus:outline-none focus:border-cyan-400"
              placeholder="e.g. 8.8.8.8 or cloudflare.com"
            />
          </div>

          <div>
            <label className="text-xs font-medium text-foreground block mb-1">Probe Protocol</label>
            <select
              value={protocol}
              onChange={(e) => setProtocol(e.target.value as any)}
              className="w-full px-3 py-2 rounded-xl bg-background border border-border text-xs text-foreground focus:outline-none focus:border-cyan-400"
            >
              <option value="icmp">ICMP Echo (Standard)</option>
              <option value="udp">UDP Dynamic Port</option>
              <option value="tcp">TCP SYN (Port 443)</option>
            </select>
          </div>

          <div className="flex items-end">
            <button
              type="button"
              onClick={runTraceroute}
              disabled={isRunning}
              className="w-full py-2 rounded-xl bg-gradient-to-r from-cyan-500 to-blue-600 hover:from-cyan-400 hover:to-blue-500 text-slate-950 font-bold text-xs shadow-md transition disabled:opacity-50 flex items-center justify-center gap-2 cursor-pointer"
            >
              {isRunning ? <RefreshCw className="h-3.5 w-3.5 animate-spin" /> : <Play className="h-3.5 w-3.5 fill-current" />}
              {isRunning ? "Tracing Route..." : "Start MTR Trace"}
            </button>
          </div>
        </div>
      </div>

      {/* Hop Table */}
      <div className="rounded-2xl border border-border/60 bg-card p-5 shadow-sm space-y-4">
        <div className="flex items-center justify-between border-b border-border/40 pb-3">
          <h2 className="text-sm font-semibold text-foreground flex items-center gap-2">
            <Activity className="h-4 w-4 text-emerald-400" />
            Path Traversal to {target} ({hops.length} Hops Recorded)
          </h2>
          <span className="text-xs font-mono text-muted-foreground">Protocol: {protocol.toUpperCase()}</span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse text-xs">
            <thead>
              <tr className="border-b border-border/60 text-muted-foreground font-semibold">
                <th className="p-2.5">Hop</th>
                <th className="p-2.5">Hostname / IP</th>
                <th className="p-2.5">Autonomous System</th>
                <th className="p-2.5">Loss %</th>
                <th className="p-2.5">Sent</th>
                <th className="p-2.5">Last</th>
                <th className="p-2.5">Avg RTT</th>
                <th className="p-2.5">Best</th>
                <th className="p-2.5">Worst</th>
                <th className="p-2.5">Std Dev</th>
              </tr>
            </thead>
            <tbody>
              {hops.map((h) => (
                <tr key={h.hop} className="border-b border-border/40 hover:bg-muted/30 transition">
                  <td className="p-2.5 font-mono font-bold text-cyan-400">{h.hop}</td>
                  <td className="p-2.5">
                    <div className="font-mono font-medium text-foreground">{h.host}</div>
                    <div className="text-[10px] text-muted-foreground font-mono">{h.ip}</div>
                  </td>
                  <td className="p-2.5 font-mono text-muted-foreground">{h.asn}</td>
                  <td className="p-2.5">
                    <span className="font-mono font-bold text-emerald-400">{h.lossPct}%</span>
                  </td>
                  <td className="p-2.5 font-mono">{h.sent}</td>
                  <td className="p-2.5 font-mono font-bold text-foreground">{h.lastMs.toFixed(2)} ms</td>
                  <td className="p-2.5 font-mono text-cyan-300 font-semibold">{h.avgMs.toFixed(2)} ms</td>
                  <td className="p-2.5 font-mono text-emerald-400">{h.bestMs.toFixed(2)} ms</td>
                  <td className="p-2.5 font-mono text-amber-400">{h.worstMs.toFixed(2)} ms</td>
                  <td className="p-2.5 font-mono text-muted-foreground">±{h.stdDev.toFixed(2)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
