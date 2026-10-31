"use client";

import { useEffect, useState } from "react";
import { Activity, ArrowRightLeft, CheckCircle2, RefreshCw, Zap } from "lucide-react";
import { useShellStore } from "@/store/ui-store";
import { translations } from "@/lib/i18n";

type MatrixEntry = {
  p50_ms: number;
  p95_ms: number;
  p99_ms: number;
  jitter_ms: number;
  loss_pct: number;
};

type TopologyData = {
  nodes: { name: string; addr: string; alive: boolean }[];
  matrix: Record<string, MatrixEntry>;
};

const defaultTopology: TopologyData = {
  nodes: [
    { name: "node-1", addr: "192.168.0.163", alive: true },
    { name: "node-2", addr: "192.168.0.163", alive: true },
    { name: "node-3", addr: "192.168.0.163", alive: true },
  ],
  matrix: {
    "node-1->node-2": { p50_ms: 0.408, p95_ms: 1.036, p99_ms: 1.155, jitter_ms: 0.242, loss_pct: 0 },
    "node-1->node-3": { p50_ms: 0.421, p95_ms: 0.994, p99_ms: 1.132, jitter_ms: 0.208, loss_pct: 0 },
    "node-2->node-1": { p50_ms: 0.403, p95_ms: 1.061, p99_ms: 1.248, jitter_ms: 0.348, loss_pct: 0 },
    "node-2->node-3": { p50_ms: 0.519, p95_ms: 1.913, p99_ms: 2.410, jitter_ms: 0.488, loss_pct: 0 },
    "node-3->node-1": { p50_ms: 0.386, p95_ms: 0.716, p99_ms: 0.879, jitter_ms: 0.122, loss_pct: 0 },
    "node-3->node-2": { p50_ms: 0.379, p95_ms: 0.896, p99_ms: 0.925, jitter_ms: 0.135, loss_pct: 0 },
  },
};

export function MeshWorkspace() {
  const { locale } = useShellStore();
  const t = translations[locale];
  const [data, setData] = useState<TopologyData>(defaultTopology);
  const [selectedLink, setSelectedLink] = useState<{ src: string; dst: string } | null>({
    src: "node-1",
    dst: "node-2",
  });
  const [isLiveConnected, setIsLiveConnected] = useState(false);
  const [lastRefreshed, setLastRefreshed] = useState<Date>(new Date());

  const fetchTopology = async () => {
    try {
      const res = await fetch("http://127.0.0.1:8081/api/topology", { cache: "no-store" });
      if (res.ok) {
        const json = await res.json();
        if (json && json.matrix && Object.keys(json.matrix).length > 0) {
          setData(json);
          setIsLiveConnected(true);
          setLastRefreshed(new Date());
          return;
        }
      }
    } catch {
      // Local cluster fetch unavailable in some browsers; maintain last known state
    }
  };

  useEffect(() => {
    fetchTopology();
    const interval = setInterval(fetchTopology, 2500);
    return () => clearInterval(interval);
  }, []);

  const nodes = data.nodes || defaultTopology.nodes;
  const matrix = data.matrix || defaultTopology.matrix;

  const currentEntry =
    selectedLink && matrix[`${selectedLink.src}->${selectedLink.dst}`]
      ? matrix[`${selectedLink.src}->${selectedLink.dst}`]
      : { p50_ms: 0.408, p95_ms: 1.036, p99_ms: 1.155, jitter_ms: 0.242, loss_pct: 0 };

  const reverseEntry =
    selectedLink && matrix[`${selectedLink.dst}->${selectedLink.src}`]
      ? matrix[`${selectedLink.dst}->${selectedLink.src}`]
      : null;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/40 pb-5">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-foreground flex items-center gap-2.5">
            <Activity className="h-5 w-5 text-cyan-400" />
            {t.mesh.title}
          </h1>
          <p className="text-xs text-muted-foreground mt-1">{t.mesh.subtitle}</p>
        </div>
        <div className="flex items-center gap-3">
          <span
            className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium border ${
              isLiveConnected
                ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                : "bg-cyan-500/10 text-cyan-400 border-cyan-500/20"
            }`}
          >
            <span
              className={`w-1.5 h-1.5 rounded-full ${
                isLiveConnected ? "bg-emerald-400 animate-pulse" : "bg-cyan-400"
              }`}
            />
            {isLiveConnected ? "Cluster Live Sync (8081)" : "Simulated Cluster Mesh"}
          </span>
          <button
            onClick={fetchTopology}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-border bg-secondary hover:bg-secondary/80 text-xs font-medium text-foreground transition-colors"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            {t.mesh.refresh}
          </button>
        </div>
      </div>

      {/* Main Grid: N×N Matrix Table & Link Inspector */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* N×N Table */}
        <div className="lg:col-span-2 rounded-2xl border border-border/60 bg-card p-5 shadow-sm space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold text-foreground flex items-center gap-2">
              <Zap className="h-4 w-4 text-emerald-400" />
              PulseMesh N×N Latency Heatmap
            </h2>
            <span className="text-xs font-mono text-muted-foreground">
              Updated: {lastRefreshed.toLocaleTimeString()}
            </span>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-center border-collapse">
              <thead>
                <tr className="text-xs font-semibold text-muted-foreground border-b border-border/60">
                  <th className="p-3 text-left">{t.mesh.source} \ {t.mesh.target}</th>
                  {nodes.map((node) => (
                    <th key={node.name} className="p-3 font-mono">
                      {node.name}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="text-xs">
                {nodes.map((srcNode) => (
                  <tr key={srcNode.name} className="border-b border-border/40 hover:bg-muted/30 transition">
                    <td className="p-3 text-left font-semibold text-foreground font-mono">
                      {srcNode.name}
                    </td>
                    {nodes.map((dstNode) => {
                      if (srcNode.name === dstNode.name) {
                        return (
                          <td key={dstNode.name} className="p-2">
                            <div className="p-2.5 rounded-xl bg-muted/40 text-muted-foreground font-mono text-xs">
                              Self
                            </div>
                          </td>
                        );
                      }

                      const key = `${srcNode.name}->${dstNode.name}`;
                      const metric = matrix[key];
                      const isSelected =
                        selectedLink?.src === srcNode.name && selectedLink?.dst === dstNode.name;

                      return (
                        <td key={dstNode.name} className="p-2">
                          <button
                            type="button"
                            onClick={() => setSelectedLink({ src: srcNode.name, dst: dstNode.name })}
                            className={`w-full text-left p-2.5 rounded-xl border transition cursor-pointer ${
                              isSelected
                                ? "bg-cyan-500/20 border-cyan-400 text-cyan-200 ring-2 ring-cyan-500/30"
                                : "bg-emerald-500/10 hover:bg-emerald-500/20 border-emerald-500/30 text-emerald-300"
                            }`}
                          >
                            <div className="font-mono font-bold text-sm">
                              {metric ? metric.p50_ms.toFixed(2) : "0.40"} ms
                            </div>
                            <div className="text-[10px] text-muted-foreground mt-0.5">
                              p95: {metric ? metric.p95_ms.toFixed(2) : "1.00"}ms · {metric ? metric.loss_pct : 0}% loss
                            </div>
                          </button>
                        </td>
                      );
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Link Inspector & Asymmetry Check */}
        <div className="rounded-2xl border border-border/60 bg-card p-5 shadow-sm flex flex-col justify-between space-y-4">
          <div>
            <div className="flex items-center justify-between pb-3 border-b border-border/50">
              <h3 className="text-sm font-semibold text-foreground">
                {t.mesh.symmetricCheck}
              </h3>
              <CheckCircle2 className="h-4 w-4 text-emerald-400" />
            </div>

            <div className="mt-3">
              <div className="text-xs text-muted-foreground">Inspected Path:</div>
              <div className="text-base font-mono font-bold text-cyan-400 mt-0.5">
                {selectedLink ? `${selectedLink.src} → ${selectedLink.dst}` : "None"}
              </div>
            </div>

            <div className="mt-4 space-y-3 text-xs">
              <div className="flex justify-between items-center py-1.5 border-b border-border/40">
                <span className="text-muted-foreground">{t.mesh.p50}:</span>
                <span className="font-mono font-bold text-foreground">
                  {currentEntry.p50_ms.toFixed(3)} ms
                </span>
              </div>
              <div className="flex justify-between items-center py-1.5 border-b border-border/40">
                <span className="text-muted-foreground">{t.mesh.p95}:</span>
                <span className="font-mono text-amber-400">
                  {currentEntry.p95_ms.toFixed(3)} ms
                </span>
              </div>
              <div className="flex justify-between items-center py-1.5 border-b border-border/40">
                <span className="text-muted-foreground">{t.mesh.p99}:</span>
                <span className="font-mono text-amber-400">
                  {currentEntry.p99_ms.toFixed(3)} ms
                </span>
              </div>
              <div className="flex justify-between items-center py-1.5 border-b border-border/40">
                <span className="text-muted-foreground">{t.mesh.jitter}:</span>
                <span className="font-mono text-foreground">
                  {currentEntry.jitter_ms.toFixed(3)} ms
                </span>
              </div>
              <div className="flex justify-between items-center py-1.5 border-b border-border/40">
                <span className="text-muted-foreground">{t.mesh.loss}:</span>
                <span className="font-mono font-bold text-emerald-400">
                  {currentEntry.loss_pct.toFixed(1)} %
                </span>
              </div>
            </div>

            {/* Asymmetry Comparison Card */}
            {reverseEntry && selectedLink && (
              <div className="mt-4 p-3 rounded-xl bg-secondary/50 border border-border/60 text-xs space-y-1.5">
                <div className="flex items-center gap-1.5 font-semibold text-foreground">
                  <ArrowRightLeft className="h-3.5 w-3.5 text-cyan-400" />
                  <span>Directional Asymmetry Delta:</span>
                </div>
                <div className="flex justify-between text-muted-foreground">
                  <span>Forward ({selectedLink.src}→{selectedLink.dst}):</span>
                  <span className="font-mono text-foreground font-semibold">{currentEntry.p50_ms.toFixed(2)} ms</span>
                </div>
                <div className="flex justify-between text-muted-foreground">
                  <span>Reverse ({selectedLink.dst}→{selectedLink.src}):</span>
                  <span className="font-mono text-foreground font-semibold">{reverseEntry.p50_ms.toFixed(2)} ms</span>
                </div>
                <div className="flex justify-between text-muted-foreground pt-1 border-t border-border/40">
                  <span>Path Delta (|Fwd - Rev|):</span>
                  <span className="font-mono text-emerald-400 font-bold">
                    {Math.abs(currentEntry.p50_ms - reverseEntry.p50_ms).toFixed(3)} ms
                  </span>
                </div>
              </div>
            )}
          </div>

          <div className="p-3 bg-muted/40 rounded-xl border border-border/50 text-[11px] text-muted-foreground">
            <span className="text-emerald-400 font-semibold">{t.mesh.optimal}:</span> Latency &lt; 20ms and 0% packet loss measured with RFC 3550 ring buffer.
          </div>
        </div>
      </div>
    </div>
  );
}
