"use client";

import { useState } from "react";
import {
  Activity,
  AlertTriangle,
  ArrowRight,
  Bot,
  CheckCircle2,
  ChevronRight,
  Clock,
  Layers,
  Play,
  RotateCcw,
  ShieldAlert,
  Terminal,
  Wrench,
  Zap,
} from "lucide-react";
import { useShellStore } from "@/store/ui-store";
import { translations } from "@/lib/i18n";

type Playbook = {
  id: string;
  name: string;
  trigger: string;
  target: string;
  actions: string[];
  requiresApproval: boolean;
  status: "ACTIVE" | "IDLE";
};

type Execution = {
  id: string;
  playbookName: string;
  target: string;
  cause: string;
  status: "PENDING_APPROVAL" | "SUCCEEDED" | "ROLLED_BACK";
  time: string;
  logs: string[];
};

const defaultPlaybooks: Playbook[] = [
  {
    id: "pb-optical-degradation-reroute",
    name: "Optical DOM Degradation Auto-Reroute",
    trigger: "Optical Rx Power < -3.0 dBm (Z-Score >= 3.0)",
    target: "core-router-1",
    actions: ["Prepend BGP AS-path x3", "Dispatch P1 Ticket in Jira Service Desk"],
    requiresApproval: true,
    status: "ACTIVE",
  },
  {
    id: "pb-port-flapping-quarantine",
    name: "Port Flapping Auto-Quarantine",
    trigger: "Flapping > 5 state changes in 60s",
    target: "access-sw-1",
    actions: ["Administratively disable switch port", "Notify Slack #netops-alerts"],
    requiresApproval: false,
    status: "ACTIVE",
  },
  {
    id: "pb-cloud-sg-quarantine",
    name: "Cloud Security Group Auto-Quarantine",
    trigger: "Excessive REJECT flow burst (>1000 pkts/min)",
    target: "AWS VPC / Azure NSG",
    actions: ["Inject temporary drop rule into ingress SG", "Trigger 15m expiration timer"],
    requiresApproval: true,
    status: "ACTIVE",
  },
];

const defaultExecutions: Execution[] = [
  {
    id: "exec-9421-optical-reroute",
    playbookName: "Optical DOM Degradation Auto-Reroute",
    target: "core-router-1 (Te0/1/1)",
    cause: "Optical Rx Power degraded to -3.12 dBm (Confidence: 95%)",
    status: "PENDING_APPROVAL",
    time: "Just now",
    logs: [
      "AI RCA isolated root-cause anomaly on core-router-1",
      "Playbook triggered. Awaiting operator approval before traffic shift.",
    ],
  },
  {
    id: "exec-9380-port-quarantine",
    playbookName: "Port Flapping Auto-Quarantine",
    target: "access-sw-1 (Gi1/0/24)",
    cause: "Rapid link flapping detected: 7 transitions in 45s",
    status: "SUCCEEDED",
    time: "14 mins ago",
    logs: [
      "Threshold breach detected on access-sw-1 interface Gi1/0/24",
      "Auto-quarantine executed: interface administratively shutdown",
      "Spanning-Tree topology recalculation storm prevented",
    ],
  },
];

export function AiopsWorkspace() {
  const { locale } = useShellStore();
  const t = translations[locale];

  const [playbooks, setPlaybooks] = useState<Playbook[]>(defaultPlaybooks);
  const [executions, setExecutions] = useState<Execution[]>(defaultExecutions);
  const [activeTab, setActiveTab] = useState<"history" | "playbooks" | "chatops">("history");

  const handleApprove = (id: string) => {
    setExecutions((prev) =>
      prev.map((e) =>
        e.id === id
          ? {
              ...e,
              status: "SUCCEEDED",
              logs: [
                ...e.logs,
                "Approved by operator via Web Cockpit.",
                "Action executed: BGP traffic shifted via AS-path prepend x3.",
                "Remediation verified: 0% packet loss restored.",
              ],
            }
          : e
      )
    );
  };

  const handleRollback = (id: string) => {
    setExecutions((prev) =>
      prev.map((e) =>
        e.id === id
          ? {
              ...e,
              status: "ROLLED_BACK",
              logs: [...e.logs, "Rollback initiated. BGP attributes reverted to baseline routing."],
            }
          : e
      )
    );
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/40 pb-5">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-foreground flex items-center gap-2.5">
            <Wrench className="h-5 w-5 text-purple-400" />
            AIOps Closed-Loop Remediation & Self-Healing
          </h1>
          <p className="text-xs text-muted-foreground mt-1">
            Automated network remediation playbooks with approval gates, BGP rerouting, and bidirectional ChatOps.
          </p>
        </div>
        <div className="flex items-center gap-2 text-xs">
          <span className="px-3 py-1 rounded-full bg-purple-500/10 text-purple-400 border border-purple-500/20 font-medium flex items-center gap-1.5">
            <Zap className="h-3.5 w-3.5 text-purple-400" />
            Autonomous Playbook Engine: Ready
          </span>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b border-border/40 pb-2">
        <button
          onClick={() => setActiveTab("history")}
          className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition cursor-pointer ${
            activeTab === "history"
              ? "bg-purple-500/20 text-purple-300 border border-purple-500/30"
              : "text-muted-foreground hover:bg-secondary/60 hover:text-foreground"
          }`}
        >
          Remediation Executions & Approval Gate
        </button>
        <button
          onClick={() => setActiveTab("playbooks")}
          className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition cursor-pointer ${
            activeTab === "playbooks"
              ? "bg-purple-500/20 text-purple-300 border border-purple-500/30"
              : "text-muted-foreground hover:bg-secondary/60 hover:text-foreground"
          }`}
        >
          Active Playbook Definitions
        </button>
        <button
          onClick={() => setActiveTab("chatops")}
          className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition cursor-pointer ${
            activeTab === "chatops"
              ? "bg-purple-500/20 text-purple-300 border border-purple-500/30"
              : "text-muted-foreground hover:bg-secondary/60 hover:text-foreground"
          }`}
        >
          Bidirectional ChatOps & Webhooks
        </button>
      </div>

      {/* Tab 1: Executions & Approval Gate */}
      {activeTab === "history" && (
        <div className="space-y-4">
          {executions.map((exec) => {
            const isPending = exec.status === "PENDING_APPROVAL";

            return (
              <div
                key={exec.id}
                className={`rounded-2xl border p-5 shadow-sm space-y-4 transition ${
                  isPending
                    ? "border-purple-500/40 bg-purple-500/5 ring-1 ring-purple-500/20"
                    : "border-border/60 bg-card"
                }`}
              >
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-border/40 pb-3">
                  <div>
                    <div className="flex items-center gap-2">
                      <span
                        className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                          isPending
                            ? "bg-amber-500/20 text-amber-300 border border-amber-500/30 animate-pulse"
                            : exec.status === "SUCCEEDED"
                            ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                            : "bg-muted/40 text-muted-foreground"
                        }`}
                      >
                        {exec.status}
                      </span>
                      <span className="text-sm font-bold font-mono text-foreground">{exec.id}</span>
                      <span className="text-xs text-muted-foreground">({exec.playbookName})</span>
                    </div>
                    <div className="text-xs text-muted-foreground mt-1">
                      Target: <span className="font-mono text-cyan-300 font-semibold">{exec.target}</span> · Trigger: {exec.cause}
                    </div>
                  </div>

                  {isPending && (
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => handleApprove(exec.id)}
                        className="px-3.5 py-1.5 rounded-xl bg-gradient-to-r from-emerald-500 to-teal-600 hover:from-emerald-400 hover:to-teal-500 text-slate-950 font-bold text-xs shadow-md transition flex items-center gap-1.5 cursor-pointer"
                      >
                        <CheckCircle2 className="h-3.5 w-3.5" />
                        Approve & Execute
                      </button>
                      <button
                        type="button"
                        onClick={() => handleRollback(exec.id)}
                        className="px-3 py-1.5 rounded-xl border border-border bg-secondary hover:bg-secondary/80 text-foreground text-xs font-semibold transition flex items-center gap-1.5 cursor-pointer"
                      >
                        <RotateCcw className="h-3.5 w-3.5" />
                        Dismiss
                      </button>
                    </div>
                  )}

                  {!isPending && exec.status === "SUCCEEDED" && (
                    <button
                      type="button"
                      onClick={() => handleRollback(exec.id)}
                      className="px-3 py-1.5 rounded-xl border border-border bg-secondary hover:bg-secondary/80 text-xs font-medium text-muted-foreground hover:text-foreground transition flex items-center gap-1.5 cursor-pointer"
                    >
                      <RotateCcw className="h-3.5 w-3.5" />
                      Rollback Changes
                    </button>
                  )}
                </div>

                {/* Execution Logs Terminal */}
                <div className="p-3.5 rounded-xl bg-slate-950/80 border border-slate-800 font-mono text-xs space-y-1">
                  <div className="text-slate-500 text-[10px] font-semibold uppercase tracking-wider mb-1">
                    Playbook Execution Trace Logs:
                  </div>
                  {exec.logs.map((log, lIdx) => (
                    <div key={lIdx} className="text-slate-300 flex items-start gap-2">
                      <span className="text-cyan-400 select-none">›</span>
                      <span>{log}</span>
                    </div>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Tab 2: Playbook Definitions */}
      {activeTab === "playbooks" && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
          {playbooks.map((pb) => (
            <div key={pb.id} className="rounded-2xl border border-border/60 bg-card p-5 shadow-sm space-y-3 flex flex-col justify-between">
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold text-cyan-400 uppercase tracking-wider">{pb.name}</span>
                  <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                    {pb.status}
                  </span>
                </div>
                <div className="text-xs text-muted-foreground pt-1">
                  <span className="font-semibold text-foreground">Trigger:</span> {pb.trigger}
                </div>
                <div className="text-xs text-muted-foreground">
                  <span className="font-semibold text-foreground">Approval Gate:</span>{" "}
                  {pb.requiresApproval ? "Requires Operator Approval" : "Zero-Touch Autonomous"}
                </div>
                <div className="pt-2 border-t border-border/40 space-y-1 text-xs">
                  <span className="font-semibold text-foreground">Remediation Steps:</span>
                  {pb.actions.map((act, aIdx) => (
                    <div key={aIdx} className="text-muted-foreground flex items-center gap-1.5 font-mono text-[11px]">
                      <span className="text-emerald-400">✓</span> {act}
                    </div>
                  ))}
                </div>
              </div>

              <div className="pt-3 border-t border-border/40 text-[11px] text-muted-foreground">
                Target Scope: <code className="text-cyan-300 font-mono">{pb.target}</code>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Tab 3: ChatOps & Webhooks */}
      {activeTab === "chatops" && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="rounded-2xl border border-border/60 bg-card p-5 shadow-sm space-y-4">
            <div className="flex items-center gap-2 border-b border-border/40 pb-3">
              <Bot className="h-4 w-4 text-cyan-400" />
              <h3 className="text-sm font-semibold text-foreground">ChatOps Slash Command Interface</h3>
            </div>
            <p className="text-xs text-muted-foreground">
              Operators can interact with Qnetward directly from Slack, Microsoft Teams, and Discord.
            </p>

            <div className="space-y-2.5 text-xs">
              <div className="p-2.5 rounded-xl bg-background border border-border space-y-1">
                <code className="text-cyan-400 font-mono font-bold">/qnetward status</code>
                <p className="text-muted-foreground text-[11px]">Reports cluster health, active probe count, and operational state.</p>
              </div>
              <div className="p-2.5 rounded-xl bg-background border border-border space-y-1">
                <code className="text-cyan-400 font-mono font-bold">/qnetward approve &lt;exec_id&gt;</code>
                <p className="text-muted-foreground text-[11px]">Approves pending remediation playbook execution from chat.</p>
              </div>
              <div className="p-2.5 rounded-xl bg-background border border-border space-y-1">
                <code className="text-cyan-400 font-mono font-bold">/qnetward rollback &lt;exec_id&gt;</code>
                <p className="text-muted-foreground text-[11px]">Rolls back executed remediation actions to baseline configuration.</p>
              </div>
              <div className="p-2.5 rounded-xl bg-background border border-border space-y-1">
                <code className="text-cyan-400 font-mono font-bold">/qnetward mute &lt;target&gt; &lt;duration&gt;</code>
                <p className="text-muted-foreground text-[11px]">Silences cascading threshold alarms on device during maintenance (e.g. `core-router-1 2h`).</p>
              </div>
            </div>
          </div>

          <div className="rounded-2xl border border-border/60 bg-card p-5 shadow-sm space-y-4">
            <div className="flex items-center gap-2 border-b border-border/40 pb-3">
              <Zap className="h-4 w-4 text-purple-400" />
              <h3 className="text-sm font-semibold text-foreground">Active Webhook Channels</h3>
            </div>

            <div className="space-y-3 text-xs">
              <div className="p-3 rounded-xl bg-background border border-border flex justify-between items-center">
                <div>
                  <div className="font-semibold text-foreground">Slack #netops-incidents</div>
                  <div className="text-[10px] text-muted-foreground font-mono">https://hooks.slack.com/services/T0...</div>
                </div>
                <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  CONNECTED
                </span>
              </div>

              <div className="p-3 rounded-xl bg-background border border-border flex justify-between items-center">
                <div>
                  <div className="font-semibold text-foreground">Microsoft Teams IT Ops</div>
                  <div className="text-[10px] text-muted-foreground font-mono">https://outlook.office.com/webhook/...</div>
                </div>
                <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  CONNECTED
                </span>
              </div>

              <div className="p-3 rounded-xl bg-background border border-border flex justify-between items-center">
                <div>
                  <div className="font-semibold text-foreground">PagerDuty Event Orchestration</div>
                  <div className="text-[10px] text-muted-foreground font-mono">https://events.pagerduty.com/v2/enqueue</div>
                </div>
                <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  CONNECTED
                </span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
