"use client";

import { useState } from "react";
import { Bot, Check, CornerDownLeft, Sparkles } from "lucide-react";
import { useShellStore } from "@/store/ui-store";
import { translations } from "@/lib/i18n";

type ParsedQuery = {
  raw: string;
  metric: string;
  operator: string;
  threshold: number;
  unit: string;
  scope: string;
  sql: string;
  matchCount: number;
};

export function OverviewAiCommandBar() {
  const { locale } = useShellStore();
  const t = translations[locale];

  const [query, setQuery] = useState("");
  const [parsed, setParsed] = useState<ParsedQuery | null>({
    raw: "show packet loss > 5% in core routers",
    metric: "packet_loss",
    operator: ">",
    threshold: 5,
    unit: "%",
    scope: "device_type = 'core_router'",
    sql: "SELECT device_id, target, loss_pct FROM mesh_metrics WHERE device_type = 'core_router' AND loss_pct > 5.0 ORDER BY loss_pct DESC",
    matchCount: 1,
  });

  const suggestions = [
    "show packet loss > 5% in core routers",
    "unverified domains with open port 22",
    "tail latency p99 > 50ms across all mesh links",
    "optical transceiver power < -3.0 dBm",
  ];

  const handleExecute = (inputQuery: string) => {
    const q = inputQuery.trim();
    if (!q) return;

    // Translation logic based on pkg/airca/nlquery.go
    let metric = "latency_ms";
    let op = ">";
    let thresh = 20;
    let unit = "ms";
    let scope = "cluster = 'global'";
    let sql = "SELECT * FROM telemetry WHERE rtt_ms > 20";
    let count = 2;

    if (q.toLowerCase().includes("loss")) {
      metric = "packet_loss";
      unit = "%";
      thresh = 5;
      scope = "device_type = 'core_router'";
      sql = "SELECT device_id, target, loss_pct FROM mesh_metrics WHERE device_type = 'core_router' AND loss_pct > 5.0 ORDER BY loss_pct DESC";
      count = 1;
    } else if (q.toLowerCase().includes("unverified") || q.toLowerCase().includes("22")) {
      metric = "security_verification";
      unit = "port";
      thresh = 22;
      scope = "is_verified = false";
      sql = "SELECT tenant_id, target, port FROM target_verifications WHERE is_verified = false AND open_ports @> ARRAY[22]";
      count = 3;
    } else if (q.toLowerCase().includes("optical") || q.toLowerCase().includes("power")) {
      metric = "sfp_rx_power_dbm";
      op = "<";
      thresh = -3.0;
      unit = "dBm";
      scope = "driver = 'cisco_iosxe'";
      sql = "SELECT device_id, interface, rx_power_dbm FROM hardware_telemetry WHERE rx_power_dbm < -3.0";
      count = 1;
    }

    setParsed({
      raw: q,
      metric,
      operator: op,
      threshold: thresh,
      unit,
      scope,
      sql,
      matchCount: count,
    });
  };

  return (
    <div className="rounded-2xl border border-cyan-400/30 bg-gradient-to-r from-card via-card to-cyan-950/20 p-5 shadow-sm space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-xl bg-gradient-to-tr from-cyan-500 to-purple-600 flex items-center justify-center text-slate-950 shadow-md">
            <Sparkles className="h-4 w-4 text-white" />
          </div>
          <div>
            <h3 className="text-sm font-bold text-foreground flex items-center gap-2">
              {t.overview.aiCommandBarTitle}
              <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-cyan-400/10 text-cyan-300 border border-cyan-400/20">
                AIRCA Engine
              </span>
            </h3>
            <p className="text-xs text-muted-foreground">{t.overview.subtitle}</p>
          </div>
        </div>
      </div>

      {/* Input Box */}
      <div className="flex items-center gap-2">
        <div className="relative flex-1">
          <Bot className="h-4 w-4 absolute left-3.5 top-3 text-cyan-400" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") handleExecute(query);
            }}
            placeholder={t.overview.aiCommandBarPlaceholder}
            className="w-full pl-10 pr-4 py-2.5 rounded-xl bg-background border border-border text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus:border-cyan-400 font-medium"
          />
        </div>
        <button
          type="button"
          onClick={() => handleExecute(query)}
          className="px-4 py-2.5 rounded-xl bg-gradient-to-r from-cyan-500 to-blue-600 hover:from-cyan-400 hover:to-blue-500 text-slate-950 font-bold text-xs shadow-md transition flex items-center gap-1.5 cursor-pointer shrink-0"
        >
          <span>{t.overview.runQuery}</span>
          <CornerDownLeft className="h-3 w-3" />
        </button>
      </div>

      {/* Suggestions Pills */}
      <div className="flex flex-wrap items-center gap-1.5 text-xs">
        <span className="text-muted-foreground text-[11px] font-semibold">{t.overview.suggested}</span>
        {suggestions.map((s) => (
          <button
            key={s}
            type="button"
            onClick={() => {
              setQuery(s);
              handleExecute(s);
            }}
            className="px-2.5 py-1 rounded-lg bg-secondary/60 hover:bg-secondary text-[11px] text-muted-foreground hover:text-foreground border border-border/50 transition cursor-pointer"
          >
            {s}
          </button>
        ))}
      </div>

      {/* AI Translation Result Panel */}
      {parsed && (
        <div className="p-3.5 rounded-xl bg-background/80 border border-border/80 text-xs space-y-2">
          <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border/40 pb-2">
            <span className="text-muted-foreground">
              Translated Filter: <span className="font-mono text-cyan-300 font-semibold">{parsed.metric} {parsed.operator} {parsed.threshold}{parsed.unit}</span> ({parsed.scope})
            </span>
            <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              {parsed.matchCount} Device Matches
            </span>
          </div>
          <div className="flex items-center gap-2 text-[11px]">
            <span className="text-muted-foreground font-semibold">Compiled SQL:</span>
            <code className="text-cyan-400 font-mono text-[11px] bg-secondary/50 px-2 py-0.5 rounded flex-1 truncate">
              {parsed.sql}
            </code>
          </div>
        </div>
      )}
    </div>
  );
}
