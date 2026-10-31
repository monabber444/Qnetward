"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Activity, ArrowRight, Network } from "lucide-react";
import { useShellStore } from "@/store/ui-store";
import { translations } from "@/lib/i18n";

export function OverviewMeshPreview() {
  const { locale } = useShellStore();
  const t = translations[locale];

  const [matrix, setMatrix] = useState({
    "node-1->node-2": { p50_ms: 0.408, loss_pct: 0 },
    "node-1->node-3": { p50_ms: 0.421, loss_pct: 0 },
    "node-2->node-1": { p50_ms: 0.403, loss_pct: 0 },
    "node-2->node-3": { p50_ms: 0.519, loss_pct: 0 },
    "node-3->node-1": { p50_ms: 0.386, loss_pct: 0 },
    "node-3->node-2": { p50_ms: 0.379, loss_pct: 0 },
  });
  const [liveActive, setLiveActive] = useState(false);

  useEffect(() => {
    const fetchMatrix = async () => {
      try {
        const res = await fetch("http://127.0.0.1:8081/api/topology", { cache: "no-store" });
        if (res.ok) {
          const json = await res.json();
          if (json && json.matrix) {
            setMatrix(json.matrix);
            setLiveActive(true);
          }
        }
      } catch {
        // Fallback
      }
    };
    fetchMatrix();
    const interval = setInterval(fetchMatrix, 3000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="rounded-2xl border border-border/60 bg-card p-5 shadow-sm space-y-4">
      <div className="flex items-center justify-between border-b border-border/40 pb-3">
        <div className="flex items-center gap-2">
          <Network className="h-4 w-4 text-cyan-400" />
          <h3 className="text-sm font-semibold text-foreground">{t.overview.livePulseMeshTitle}</h3>
        </div>
        <div className="flex items-center gap-3">
          <span
            className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-bold ${
              liveActive
                ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                : "bg-cyan-500/10 text-cyan-400 border border-cyan-500/20"
            }`}
          >
            <span
              className={`w-1.5 h-1.5 rounded-full ${
                liveActive ? "bg-emerald-400 animate-pulse" : "bg-cyan-400"
              }`}
            />
            {liveActive ? "Port 8081 Live Sync" : "3 Nodes Mesh"}
          </span>
          <Link
            href="/mesh"
            className="text-xs font-semibold text-cyan-400 hover:text-cyan-300 flex items-center gap-1 transition"
          >
            <span>Full Matrix</span>
            <ArrowRight className="h-3 w-3" />
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-3 text-center">
        <div className="p-3 rounded-xl bg-secondary/30 border border-border/50">
          <div className="text-[11px] text-muted-foreground font-mono">node-1 ↔ node-2</div>
          <div className="text-base font-bold font-mono text-emerald-400 mt-0.5">
            {matrix["node-1->node-2"]?.p50_ms.toFixed(2) || "0.41"} ms
          </div>
          <div className="text-[10px] text-muted-foreground">0% loss · Sub-ms RTT</div>
        </div>
        <div className="p-3 rounded-xl bg-secondary/30 border border-border/50">
          <div className="text-[11px] text-muted-foreground font-mono">node-1 ↔ node-3</div>
          <div className="text-base font-bold font-mono text-emerald-400 mt-0.5">
            {matrix["node-1->node-3"]?.p50_ms.toFixed(2) || "0.42"} ms
          </div>
          <div className="text-[10px] text-muted-foreground">0% loss · Sub-ms RTT</div>
        </div>
        <div className="p-3 rounded-xl bg-secondary/30 border border-border/50">
          <div className="text-[11px] text-muted-foreground font-mono">node-2 ↔ node-3</div>
          <div className="text-base font-bold font-mono text-emerald-400 mt-0.5">
            {matrix["node-2->node-3"]?.p50_ms.toFixed(2) || "0.52"} ms
          </div>
          <div className="text-[10px] text-muted-foreground">0% loss · Sub-ms RTT</div>
        </div>
      </div>
    </div>
  );
}
