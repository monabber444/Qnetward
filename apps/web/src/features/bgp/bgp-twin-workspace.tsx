"use client";

import { useState } from "react";
import {
  Activity,
  AlertOctagon,
  AlertTriangle,
  ArrowRight,
  CheckCircle2,
  Cpu,
  GitFork,
  Globe2,
  Layers,
  Network,
  Play,
  Radio,
  RefreshCw,
  RotateCcw,
  Server,
  ShieldAlert,
  Zap,
} from "lucide-react";
import { useShellStore } from "@/store/ui-store";

type BGPRouteItem = {
  prefix: string;
  originASN: number;
  originName: string;
  asPath: { asn: number; name: string; tier: string }[];
  nextHop: string;
  rpkiStatus: "VALID" | "INVALID" | "NOT_FOUND";
  status: "NORMAL" | "HIJACK" | "FLAPPING" | "PREPENDED";
  updatedAt: string;
};

type TwinNode = {
  id: string;
  name: string;
  region: string;
  type: string;
  status: "UP" | "MAINTENANCE" | "DOWN";
};

type TwinLink = {
  id: string;
  name: string;
  src: string;
  dst: string;
  latencyMs: number;
  capacityGbps: number;
  currentLoadGbps: number;
  status: "UP" | "SEVERED";
};

type SimulationOutput = {
  scenarioName: string;
  severity: "CRITICAL" | "HIGH" | "MEDIUM" | "LOW";
  totalTrafficGbps: number;
  droppedTrafficGbps: number;
  avgLatencyDeltaMs: number;
  saturatedLinks: {
    linkName: string;
    capacity: number;
    load: number;
    utilization: number;
  }[];
  reroutedFlows: {
    flowName: string;
    from: string;
    to: string;
    originalPath: string[];
    newPath: string[];
    latencyDelta: number;
    isDropped: boolean;
  }[];
};

const initialBgpRoutes: BGPRouteItem[] = [
  {
    prefix: "198.51.100.0/24",
    originASN: 64500,
    originName: "Enterprise Corp Net",
    asPath: [
      { asn: 174, name: "Cogent", tier: "Tier-1" },
      { asn: 3356, name: "Lumen", tier: "Tier-1" },
      { asn: 64500, name: "Enterprise Corp", tier: "Enterprise" },
    ],
    nextHop: "198.51.100.1",
    rpkiStatus: "VALID",
    status: "NORMAL",
    updatedAt: "2 mins ago",
  },
  {
    prefix: "203.0.113.0/24",
    originASN: 99999,
    originName: "Suspicious Rogue ISP",
    asPath: [
      { asn: 1299, name: "Telia / Arelion", tier: "Tier-1" },
      { asn: 99999, name: "Rogue ISP", tier: "Transit" },
    ],
    nextHop: "203.0.113.5",
    rpkiStatus: "INVALID",
    status: "HIJACK",
    updatedAt: "4 mins ago",
  },
  {
    prefix: "142.250.190.0/24",
    originASN: 15169,
    originName: "Google LLC",
    asPath: [
      { asn: 2914, name: "NTT Global", tier: "Tier-1" },
      { asn: 15169, name: "Google Cloud", tier: "Hyperscaler" },
    ],
    nextHop: "142.250.190.1",
    rpkiStatus: "VALID",
    status: "NORMAL",
    updatedAt: "12 mins ago",
  },
  {
    prefix: "192.0.2.0/24",
    originASN: 65001,
    originName: "Cloud Hosting Core",
    asPath: [
      { asn: 6939, name: "Hurricane Electric", tier: "Tier-1" },
      { asn: 65001, name: "Hosting Edge", tier: "ISP" },
    ],
    nextHop: "192.0.2.254",
    rpkiStatus: "NOT_FOUND",
    status: "FLAPPING",
    updatedAt: "Just now",
  },
];

const baselineNodes: TwinNode[] = [
  { id: "nyc", name: "NYC Core Gateway", region: "us-east", type: "CoreRouter", status: "UP" },
  { id: "lon", name: "London Internet Exchange", region: "eu-west", type: "CoreRouter", status: "UP" },
  { id: "fra", name: "Frankfurt Transit Hub", region: "eu-central", type: "CoreRouter", status: "UP" },
  { id: "sin", name: "Singapore Equinix Core", region: "ap-southeast", type: "CoreRouter", status: "UP" },
  { id: "tyo", name: "Tokyo Otemachi Pop", region: "ap-northeast", type: "CoreRouter", status: "UP" },
];

const baselineLinks: TwinLink[] = [
  { id: "tat14", name: "Transatlantic TAT-14 (NYC ↔ LON)", src: "nyc", dst: "lon", latencyMs: 65, capacityGbps: 100, currentLoadGbps: 42, status: "UP" },
  { id: "trans-south", name: "Subsea Cable South (NYC ↔ FRA)", src: "nyc", dst: "fra", latencyMs: 85, capacityGbps: 50, currentLoadGbps: 24, status: "UP" },
  { id: "euro-backhaul", name: "Euro Core Terrestrial (LON ↔ FRA)", src: "lon", dst: "fra", latencyMs: 14, capacityGbps: 100, currentLoadGbps: 36, status: "UP" },
  { id: "eurasia", name: "SMW-5 Eurasia Cable (FRA ↔ SIN)", src: "fra", dst: "sin", latencyMs: 124, capacityGbps: 80, currentLoadGbps: 48, status: "UP" },
  { id: "apac-ring", name: "APCN-2 Subsea Ring (SIN ↔ TYO)", src: "sin", dst: "tyo", latencyMs: 68, capacityGbps: 60, currentLoadGbps: 28, status: "UP" },
  { id: "transpacific", name: "FASTER Transpacific (TYO ↔ NYC)", src: "tyo", dst: "nyc", latencyMs: 148, capacityGbps: 40, currentLoadGbps: 18, status: "UP" },
];

export function BgpTwinWorkspace() {
  const { locale } = useShellStore();
  const [activeTab, setActiveTab] = useState<"bgp" | "twin">("twin");
  const [bgpRoutes] = useState<BGPRouteItem[]>(initialBgpRoutes);

  // Digital Twin Simulator State
  const [nodes, setNodes] = useState<TwinNode[]>(baselineNodes);
  const [links, setLinks] = useState<TwinLink[]>(baselineLinks);
  const [activeSim, setActiveSim] = useState<SimulationOutput | null>(null);

  // Simulate Cable Severance
  const handleSimulateSeverCable = () => {
    setLinks(prev =>
      prev.map(l => (l.id === "tat14" ? { ...l, status: "SEVERED" } : l))
    );
    setActiveSim({
      scenarioName: "Sever Transatlantic Cable TAT-14 (NYC ↔ LON)",
      severity: "HIGH",
      totalTrafficGbps: 196,
      droppedTrafficGbps: 0,
      avgLatencyDeltaMs: 24.0,
      saturatedLinks: [
        {
          linkName: "Subsea Cable South (NYC ↔ FRA)",
          capacity: 50,
          load: 48.5,
          utilization: 97.0,
        },
      ],
      reroutedFlows: [
        {
          flowName: "US-East to London Enterprise Flow",
          from: "nyc",
          to: "lon",
          originalPath: ["nyc", "lon"],
          newPath: ["nyc", "fra", "lon"],
          latencyDelta: 34.0,
          isDropped: false,
        },
        {
          flowName: "US-East to Singapore Financial Flow",
          from: "nyc",
          to: "sin",
          originalPath: ["nyc", "lon", "fra", "sin"],
          newPath: ["nyc", "fra", "sin"],
          latencyDelta: -6.0,
          isDropped: false,
        },
      ],
    });
  };

  // Simulate Node Maintenance
  const handleSimulateNodeMaintenance = () => {
    setNodes(prev =>
      prev.map(n => (n.id === "fra" ? { ...n, status: "MAINTENANCE" } : n))
    );
    setActiveSim({
      scenarioName: "Frankfurt Core Router Firmware Maintenance",
      severity: "CRITICAL",
      totalTrafficGbps: 196,
      droppedTrafficGbps: 48,
      avgLatencyDeltaMs: 46.5,
      saturatedLinks: [
        {
          linkName: "FASTER Transpacific (TYO ↔ NYC)",
          capacity: 40,
          load: 39.2,
          utilization: 98.0,
        },
      ],
      reroutedFlows: [
        {
          flowName: "Euro to Asia Smw-5 Transit",
          from: "lon",
          to: "sin",
          originalPath: ["lon", "fra", "sin"],
          newPath: ["lon", "nyc", "tyo", "sin"],
          latencyDelta: 122.0,
          isDropped: false,
        },
        {
          flowName: "Continental Central European Ingress",
          from: "fra",
          to: "nyc",
          originalPath: ["fra", "nyc"],
          newPath: [],
          latencyDelta: 0,
          isDropped: true,
        },
      ],
    });
  };

  // Reset to Baseline
  const handleResetBaseline = () => {
    setNodes(baselineNodes);
    setLinks(baselineLinks);
    setActiveSim(null);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
            <GitFork className="h-7 w-7 text-indigo-500" />
            {locale === "bn"
              ? "গ্লোবাল বিজিপি রাউটিং ইন্টেলিজেন্স ও নেটওয়ার্ক ডিজিটাল টুইন"
              : "Global BGP Routing Intelligence & Network Digital Twin"}
          </h1>
          <p className="text-sm text-muted-foreground">
            {locale === "bn"
              ? "লাইভ বিজিপি রুট হিজ্যাক ও লিক ডিটেকশন, আরপিকেআই ভ্যালিডেশন এবং ইন-মেমোরি ব্লাস্ট রেডিয়াস সিমুলেটর।"
              : "Live BGP route leak & hijack detection, RPKI validation, AS-path visualizer, and in-memory failure blast radius simulation."}
          </p>
        </div>

        <div className="flex items-center gap-2">
          {activeSim && (
            <button
              onClick={handleResetBaseline}
              className="flex items-center gap-1.5 rounded-lg border border-border bg-card px-3 py-1.5 text-xs font-medium text-foreground hover:bg-muted"
            >
              <RotateCcw className="h-3.5 w-3.5" />
              Reset Baseline
            </button>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-border">
        <button
          onClick={() => setActiveTab("twin")}
          className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors ${
            activeTab === "twin"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Network className="h-4 w-4" />
          {locale === "bn" ? "নেটওয়ার্ক ডিজিটাল টুইন সিমুলেটর" : "Network Digital Twin Simulator"}
        </button>
        <button
          onClick={() => setActiveTab("bgp")}
          className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors ${
            activeTab === "bgp"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Globe2 className="h-4 w-4" />
          {locale === "bn" ? "লাইভ বিজিপি রাউট মনিটর ও আরপিকেআই" : "Live BGP Routes & RPKI Monitor"}
        </button>
      </div>

      {/* TAB 1: Digital Twin Simulator */}
      {activeTab === "twin" && (
        <div className="space-y-6">
          {/* Action Trigger Toolbar */}
          <div className="rounded-xl border border-border bg-card p-4 space-y-3">
            <h3 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
              Execute "What-If" Failure Simulations
            </h3>
            <p className="text-xs text-muted-foreground">
              Simulate link severances, scheduled router maintenance windows, or peak traffic surges to evaluate rerouted paths and link saturation before committing changes to production.
            </p>

            <div className="flex flex-wrap gap-2.5 pt-1">
              <button
                onClick={handleSimulateSeverCable}
                className="flex items-center gap-2 rounded-lg border border-red-500/40 bg-red-500/10 px-3 py-2 text-xs font-medium text-red-400 hover:bg-red-500/20"
              >
                <AlertOctagon className="h-4 w-4" />
                Sever Transatlantic TAT-14 (NYC ↔ LON)
              </button>
              <button
                onClick={handleSimulateNodeMaintenance}
                className="flex items-center gap-2 rounded-lg border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-xs font-medium text-amber-400 hover:bg-amber-500/20"
              >
                <Server className="h-4 w-4" />
                Simulate Frankfurt Router Maintenance
              </button>
              <button
                onClick={handleResetBaseline}
                className="flex items-center gap-2 rounded-lg border border-border bg-background px-3 py-2 text-xs font-medium text-foreground hover:bg-muted"
              >
                <RotateCcw className="h-4 w-4" />
                Reset Topology
              </button>
            </div>
          </div>

          {/* Simulation Output Banner */}
          {activeSim && (
            <div className="rounded-xl border border-border bg-card p-4 space-y-4">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-border pb-3">
                <div className="space-y-0.5">
                  <span className="text-xs font-semibold text-muted-foreground uppercase">Simulation Result</span>
                  <h3 className="text-base font-bold text-foreground">{activeSim.scenarioName}</h3>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-xs font-semibold text-muted-foreground">Blast Radius:</span>
                  <span className={`px-2.5 py-1 rounded text-xs font-bold ${
                    activeSim.severity === "CRITICAL"
                      ? "bg-red-500/20 text-red-300 border border-red-500/40"
                      : "bg-amber-500/20 text-amber-300 border border-amber-500/40"
                  }`}>
                    {activeSim.severity}
                  </span>
                </div>
              </div>

              {/* Metrics Grid */}
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                <div className="rounded-lg bg-background p-3 border border-border">
                  <span className="text-[11px] text-muted-foreground">Total Traffic</span>
                  <p className="text-lg font-bold text-foreground">{activeSim.totalTrafficGbps} Gbps</p>
                </div>
                <div className="rounded-lg bg-background p-3 border border-border">
                  <span className="text-[11px] text-muted-foreground">Dropped Traffic</span>
                  <p className={`text-lg font-bold ${activeSim.droppedTrafficGbps > 0 ? "text-red-400" : "text-emerald-400"}`}>
                    {activeSim.droppedTrafficGbps} Gbps
                  </p>
                </div>
                <div className="rounded-lg bg-background p-3 border border-border">
                  <span className="text-[11px] text-muted-foreground">Avg Latency Shift</span>
                  <p className="text-lg font-bold text-amber-400">+{activeSim.avgLatencyDeltaMs.toFixed(1)} ms</p>
                </div>
                <div className="rounded-lg bg-background p-3 border border-border">
                  <span className="text-[11px] text-muted-foreground">Saturated Links (&gt;90%)</span>
                  <p className="text-lg font-bold text-red-400">{activeSim.saturatedLinks.length} Hotspots</p>
                </div>
              </div>

              {/* Saturated Links Warning */}
              {activeSim.saturatedLinks.length > 0 && (
                <div className="space-y-2">
                  <span className="text-xs font-semibold text-red-400 flex items-center gap-1.5">
                    <AlertTriangle className="h-3.5 w-3.5" />
                    Critical Bandwidth Saturation Hotspots
                  </span>
                  {activeSim.saturatedLinks.map((sat, idx) => (
                    <div key={idx} className="rounded-lg border border-red-500/30 bg-red-500/5 p-3 text-xs space-y-1.5">
                      <div className="flex justify-between items-center">
                        <span className="font-semibold text-foreground">{sat.linkName}</span>
                        <span className="font-mono text-red-400 font-bold">{sat.utilization.toFixed(1)}% Capacity</span>
                      </div>
                      <div className="w-full bg-background rounded-full h-2 overflow-hidden">
                        <div
                          className="bg-red-500 h-2 rounded-full"
                          style={{ width: `${Math.min(sat.utilization, 100)}%` }}
                        />
                      </div>
                      <span className="text-muted-foreground text-[11px]">
                        Projected Load: {sat.load.toFixed(1)} Gbps / {sat.capacity} Gbps
                      </span>
                    </div>
                  ))}
                </div>
              )}

              {/* Rerouted Flows Table */}
              <div className="space-y-2">
                <span className="text-xs font-semibold text-muted-foreground">Rerouted Traffic Paths</span>
                <div className="rounded-lg border border-border overflow-hidden">
                  <table className="w-full text-left text-xs">
                    <thead className="bg-muted text-muted-foreground border-b border-border">
                      <tr>
                        <th className="py-2 px-3">Flow</th>
                        <th className="py-2 px-3">Original Path</th>
                        <th className="py-2 px-3">New Recomputed Path</th>
                        <th className="py-2 px-3">Latency Delta</th>
                        <th className="py-2 px-3">Status</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border">
                      {activeSim.reroutedFlows.map((f, idx) => (
                        <tr key={idx} className="hover:bg-muted/40">
                          <td className="py-2.5 px-3 font-medium text-foreground">{f.flowName}</td>
                          <td className="py-2.5 px-3 font-mono text-muted-foreground">
                            {f.originalPath.join(" → ")}
                          </td>
                          <td className="py-2.5 px-3 font-mono text-foreground font-semibold">
                            {f.newPath.length > 0 ? f.newPath.join(" → ") : "—"}
                          </td>
                          <td className="py-2.5 px-3 font-mono">
                            {f.isDropped ? "N/A" : `+${f.latencyDelta}ms`}
                          </td>
                          <td className="py-2.5 px-3">
                            {f.isDropped ? (
                              <span className="text-red-400 font-bold">DROPPED</span>
                            ) : (
                              <span className="text-emerald-400 font-bold">REROUTED</span>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}

          {/* Topology Model View */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* Core Nodes */}
            <div className="rounded-xl border border-border bg-card p-4 space-y-3">
              <h3 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-2">
                <Server className="h-4 w-4 text-primary" />
                Network Digital Twin Nodes ({nodes.length})
              </h3>
              <div className="divide-y divide-border">
                {nodes.map(n => (
                  <div key={n.id} className="py-2.5 flex items-center justify-between">
                    <div>
                      <div className="font-semibold text-foreground text-sm">{n.name}</div>
                      <div className="text-xs text-muted-foreground font-mono">
                        ID: {n.id} • Region: {n.region} • Type: {n.type}
                      </div>
                    </div>
                    <span className={`px-2 py-0.5 rounded text-[11px] font-bold ${
                      n.status === "UP"
                        ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                        : "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                    }`}>
                      {n.status}
                    </span>
                  </div>
                ))}
              </div>
            </div>

            {/* Core Links */}
            <div className="rounded-xl border border-border bg-card p-4 space-y-3">
              <h3 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-2">
                <Network className="h-4 w-4 text-primary" />
                Subsea Cables & Transit Circuits ({links.length})
              </h3>
              <div className="divide-y divide-border">
                {links.map(l => (
                  <div key={l.id} className="py-2.5 flex items-center justify-between">
                    <div>
                      <div className="font-semibold text-foreground text-sm flex items-center gap-1.5">
                        {l.name}
                      </div>
                      <div className="text-xs text-muted-foreground font-mono">
                        {l.latencyMs}ms baseline • Capacity: {l.capacityGbps} Gbps • Load: {l.currentLoadGbps} Gbps
                      </div>
                    </div>
                    <span className={`px-2 py-0.5 rounded text-[11px] font-bold ${
                      l.status === "UP"
                        ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                        : "bg-red-500/10 text-red-400 border border-red-500/20"
                    }`}>
                      {l.status}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* TAB 2: Live BGP Routes & RPKI */}
      {activeTab === "bgp" && (
        <div className="space-y-4">
          <div className="rounded-xl border border-border bg-card p-4 space-y-2">
            <h3 className="text-base font-semibold text-foreground flex items-center gap-2">
              <Globe2 className="h-5 w-5 text-indigo-400" />
              Real-Time BGP Monitoring Protocol (BMP) Feeds
            </h3>
            <p className="text-xs text-muted-foreground">
              Continuous route analysis detecting Prefix Hijacking via RPKI Route Origin Validation (ROV), AS-Path prepending anomalies, and route flapping.
            </p>
          </div>

          <div className="rounded-xl border border-border bg-card overflow-hidden">
            <table className="w-full text-left text-xs">
              <thead className="bg-muted text-muted-foreground border-b border-border font-semibold">
                <tr>
                  <th className="py-2.5 px-4">Prefix</th>
                  <th className="py-2.5 px-4">Origin AS</th>
                  <th className="py-2.5 px-4">AS-Path Flow (Tier-1 Carrier Hops)</th>
                  <th className="py-2.5 px-4">Next-Hop</th>
                  <th className="py-2.5 px-4">RPKI Status</th>
                  <th className="py-2.5 px-4">Health State</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {bgpRoutes.map((r, idx) => (
                  <tr key={idx} className="hover:bg-muted/50">
                    <td className="py-3 px-4 font-mono font-bold text-foreground">{r.prefix}</td>
                    <td className="py-3 px-4">
                      <div className="font-mono text-foreground font-semibold">AS{r.originASN}</div>
                      <div className="text-[11px] text-muted-foreground">{r.originName}</div>
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-1 flex-wrap">
                        {r.asPath.map((hop, hIdx) => (
                          <span key={hIdx} className="flex items-center gap-1">
                            <span className="rounded bg-muted px-2 py-0.5 font-mono text-[11px] text-foreground border border-border">
                              AS{hop.asn} ({hop.name})
                            </span>
                            {hIdx < r.asPath.length - 1 && (
                              <ArrowRight className="h-3 w-3 text-muted-foreground" />
                            )}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td className="py-3 px-4 font-mono text-muted-foreground">{r.nextHop}</td>
                    <td className="py-3 px-4">
                      <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                        r.rpkiStatus === "VALID"
                          ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                          : r.rpkiStatus === "INVALID"
                          ? "bg-red-500/10 text-red-400 border border-red-500/20"
                          : "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                      }`}>
                        RPKI {r.rpkiStatus}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                        r.status === "NORMAL"
                          ? "bg-emerald-500/10 text-emerald-300"
                          : r.status === "HIJACK"
                          ? "bg-red-500/20 text-red-300 border border-red-500/40"
                          : "bg-amber-500/10 text-amber-300"
                      }`}>
                        {r.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
