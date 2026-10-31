"use client";

import Link from "next/link";
import { AlertOctagon, CheckCircle2, ChevronRight, Network, ShieldAlert, Wrench } from "lucide-react";
import { useShellStore } from "@/store/ui-store";
import { translations } from "@/lib/i18n";

export function OverviewAiRcaPanel() {
  const { locale } = useShellStore();
  const t = translations[locale];

  return (
    <div className="rounded-2xl border border-purple-500/30 bg-gradient-to-r from-card via-card to-purple-950/20 p-5 shadow-sm space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-border/40 pb-3">
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-xl bg-purple-500/20 text-purple-300 border border-purple-500/30 flex items-center justify-center font-bold">
            <AlertOctagon className="h-4 w-4" />
          </div>
          <div>
            <h3 className="text-sm font-bold text-foreground flex items-center gap-2">
              {t.overview.aiRcaTitle}
              <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-purple-500/20 text-purple-300 border border-purple-500/30">
                1 Incident Grouped
              </span>
            </h3>
            <p className="text-xs text-muted-foreground">{t.overview.aiRcaSubtitle}</p>
          </div>
        </div>

        <Link
          href="/topology"
          className="text-xs font-semibold text-cyan-400 hover:text-cyan-300 flex items-center gap-1 transition"
        >
          <span>View Topology Graph</span>
          <ChevronRight className="h-3.5 w-3.5" />
        </Link>
      </div>

      {/* Root Incident Card */}
      <div className="p-4 rounded-xl bg-purple-500/10 border border-purple-500/25 space-y-3">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <span className="px-2 py-0.5 rounded text-[10px] font-black bg-rose-500/20 text-rose-300 border border-rose-500/30">
              ROOT CAUSE
            </span>
            <span className="text-sm font-bold font-mono text-foreground">core-router-1</span>
            <span className="text-xs text-muted-foreground">(Cisco Catalyst 9300 - IOS-XE)</span>
          </div>
          <span className="text-xs font-mono font-bold text-purple-300">
            {t.overview.confidence}: 95% (Z-score: 8.4)
          </span>
        </div>

        <p className="text-xs text-muted-foreground">
          Optical transceiver <code className="text-cyan-300 font-mono">Te0/1/1</code> experienced sudden optical power degradation (<span className="text-rose-300 font-mono font-semibold">-3.12 dBm</span> vs -1.2 dBm baseline), triggering cascading loss across 4 child systems.
        </p>

        {/* Suppression Strip */}
        <div className="p-3 rounded-lg bg-background/80 border border-border/80 flex flex-col md:flex-row items-start md:items-center justify-between gap-3 text-xs">
          <div className="flex items-center gap-2">
            <ShieldAlert className="h-4 w-4 text-purple-400 shrink-0" />
            <span className="text-muted-foreground">
              {t.overview.suppressedAlerts}:{" "}
              <span className="font-mono text-foreground font-semibold">
                dist-switch-1, access-sw-1, web-01, db-01
              </span>
            </span>
          </div>
          <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 whitespace-nowrap">
            Alert Storm Prevented
          </span>
        </div>

        {/* Suggested Mitigation */}
        <div className="flex items-center gap-2 text-xs pt-1">
          <Wrench className="h-3.5 w-3.5 text-amber-400 shrink-0" />
          <span className="text-muted-foreground">{t.overview.mitigation}:</span>
          <span className="text-amber-300 font-medium">
            Inspect SFP optical power level & patch cord on core-router-1 port Te0/1/1.
          </span>
        </div>
      </div>
    </div>
  );
}
