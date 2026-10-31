"use client";

import { Activity, BatteryCharging, Cpu, HardDrive, Server, Zap } from "lucide-react";
import { useShellStore } from "@/store/ui-store";
import { translations } from "@/lib/i18n";

export function OverviewHardwareStrip() {
  const { locale } = useShellStore();
  const t = translations[locale];

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {/* SFP+ Optical DOM */}
      <div className="rounded-2xl border border-border/60 bg-card p-4 shadow-sm space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
            {t.overview.opticalDom}
          </span>
          <span className="px-2 py-0.5 rounded text-[9px] font-bold bg-amber-500/20 text-amber-300 border border-amber-500/30">
            DEGRADED
          </span>
        </div>
        <div className="text-sm font-bold font-mono text-foreground">Te0/1/1 (Cisco)</div>
        <div className="space-y-1 text-xs text-muted-foreground pt-1 border-t border-border/40">
          <div className="flex justify-between"><span>Tx Power:</span> <span className="font-mono text-cyan-300">-2.45 dBm</span></div>
          <div className="flex justify-between"><span>Rx Power:</span> <span className="font-mono text-rose-300 font-semibold">-3.12 dBm</span></div>
          <div className="flex justify-between"><span>Laser Temp:</span> <span className="font-mono text-foreground">41.2 °C</span></div>
        </div>
      </div>

      {/* APC Smart-UPS */}
      <div className="rounded-2xl border border-border/60 bg-card p-4 shadow-sm space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
            {t.overview.upsTelemetry}
          </span>
          <span className="px-2 py-0.5 rounded text-[9px] font-bold bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
            ONLINE
          </span>
        </div>
        <div className="text-sm font-bold font-mono text-foreground">Smart-UPS 1500</div>
        <div className="space-y-1 text-xs text-muted-foreground pt-1 border-t border-border/40">
          <div className="flex justify-between"><span>Battery Charge:</span> <span className="font-mono text-emerald-400 font-semibold">100 %</span></div>
          <div className="flex justify-between"><span>Runtime Est.:</span> <span className="font-mono text-cyan-300">48 min</span></div>
          <div className="flex justify-between"><span>Load Level:</span> <span className="font-mono text-foreground">34 %</span></div>
        </div>
      </div>

      {/* Dell iDRAC Redfish */}
      <div className="rounded-2xl border border-border/60 bg-card p-4 shadow-sm space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
            Dell iDRAC Redfish
          </span>
          <span className="px-2 py-0.5 rounded text-[9px] font-bold bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
            HEALTHY
          </span>
        </div>
        <div className="text-sm font-bold font-mono text-foreground">PowerEdge R750</div>
        <div className="space-y-1 text-xs text-muted-foreground pt-1 border-t border-border/40">
          <div className="flex justify-between"><span>Service Tag:</span> <span className="font-mono text-foreground font-semibold">7X49-PROD</span></div>
          <div className="flex justify-between"><span>PSU Redundancy:</span> <span className="font-mono text-cyan-300">2/2 Online</span></div>
          <div className="flex justify-between"><span>Thermal:</span> <span className="font-mono text-emerald-400">OK (32°C)</span></div>
        </div>
      </div>

      {/* Host Agent Telemetry */}
      <div className="rounded-2xl border border-border/60 bg-card p-4 shadow-sm space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
            {t.overview.hostOs}
          </span>
          <span className="px-2 py-0.5 rounded text-[9px] font-bold bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
            RUNNING
          </span>
        </div>
        <div className="text-sm font-bold font-mono text-foreground">Fedora 40 (4 Cores)</div>
        <div className="space-y-1 text-xs text-muted-foreground pt-1 border-t border-border/40">
          <div className="flex justify-between"><span>Total Memory:</span> <span className="font-mono text-foreground font-semibold">7.62 GB</span></div>
          <div className="flex justify-between"><span>Load Avg:</span> <span className="font-mono text-cyan-300">1.85 / 1.70</span></div>
          <div className="flex justify-between"><span>NVMe SMART:</span> <span className="font-mono text-emerald-400">PASSED</span></div>
        </div>
      </div>
    </div>
  );
}
