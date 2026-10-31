"use client";

import { useState } from "react";
import {
  AlertCircle,
  BatteryCharging,
  CheckCircle2,
  Cpu,
  HardDrive,
  Network,
  Radio,
  Server,
  ShieldAlert,
  Zap,
} from "lucide-react";
import { useShellStore } from "@/store/ui-store";
import { translations } from "@/lib/i18n";

type DeviceNode = {
  id: string;
  name: string;
  type: "router" | "switch" | "server" | "ups";
  vendor: string;
  ip: string;
  status: "healthy" | "root-cause" | "suppressed";
  parentId?: string;
  metrics: Record<string, string>;
};

const networkTopology: DeviceNode[] = [
  {
    id: "core-router-1",
    name: "core-router-1",
    type: "router",
    vendor: "Cisco Catalyst 9300 (IOS-XE)",
    ip: "10.0.0.1",
    status: "root-cause",
    metrics: {
      "SFP Tx Power": "-2.45 dBm",
      "SFP Rx Power": "-3.12 dBm",
      "Laser Temp": "41.2 °C",
      "Bias Current": "6.8 mA",
      "Driver": "drivers/profiles/cisco_iosxe.yaml",
    },
  },
  {
    id: "dist-sw-1",
    name: "dist-switch-1",
    type: "switch",
    vendor: "MikroTik CRS328 (RouterOS)",
    ip: "10.0.1.1",
    status: "suppressed",
    parentId: "core-router-1",
    metrics: {
      "Ports Active": "24 / 28",
      "CPU Temp": "38.5 °C",
      "Suppressed By": "core-router-1",
      "Driver": "drivers/profiles/mikrotik_routeros.yaml",
    },
  },
  {
    id: "dist-sw-2",
    name: "dist-switch-2",
    type: "switch",
    vendor: "Ubiquiti UniFi Pro (UniFi OS)",
    ip: "10.0.1.2",
    status: "healthy",
    parentId: "core-router-1",
    metrics: {
      "PoE Consumption": "142 Watts",
      "Client Count": "87 active",
      "Driver": "drivers/profiles/ubiquiti_unifi.yaml",
    },
  },
  {
    id: "server-web",
    name: "web-01 (Host Agent)",
    type: "server",
    vendor: "Linux Host Agent (Fedora 40)",
    ip: "10.0.2.10",
    status: "suppressed",
    parentId: "dist-sw-1",
    metrics: {
      "Cores": "4 Cores (Load: 1.85)",
      "RAM": "7.62 GB (37% Used)",
      "NVMe SMART": "Healthy (0 Errors)",
      "Suppressed By": "core-router-1",
    },
  },
  {
    id: "server-db",
    name: "db-01 (Dell iDRAC)",
    type: "server",
    vendor: "Dell PowerEdge R750 (Redfish)",
    ip: "10.0.2.20",
    status: "suppressed",
    parentId: "dist-sw-1",
    metrics: {
      "Service Tag": "7X49-PROD",
      "PSU Redundancy": "2/2 Online",
      "Thermal Status": "OK (32°C)",
      "Suppressed By": "core-router-1",
      "Driver": "drivers/profiles/dell_idrac_redfish.yaml",
    },
  },
  {
    id: "ups-rack-1",
    name: "apc-ups-rack-1",
    type: "ups",
    vendor: "APC Smart-UPS 1500 (SNMP)",
    ip: "10.0.3.5",
    status: "healthy",
    parentId: "core-router-1",
    metrics: {
      "Battery Charge": "100%",
      "Runtime Est.": "48 min",
      "Load Level": "34%",
      "Driver": "drivers/profiles/apc_ups.yaml",
    },
  },
];

export function TopologyWorkspace() {
  const { locale } = useShellStore();
  const t = translations[locale];
  const [selectedDevice, setSelectedDevice] = useState<DeviceNode>(networkTopology[0]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/40 pb-5">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-foreground flex items-center gap-2.5">
            <Server className="h-5 w-5 text-cyan-400" />
            {t.topology.title}
          </h1>
          <p className="text-xs text-muted-foreground mt-1">{t.topology.subtitle}</p>
        </div>
        <div className="flex items-center gap-2 text-xs">
          <span className="px-3 py-1 rounded-full bg-purple-500/10 text-purple-400 border border-purple-500/20 font-medium">
            Dependency Alert Suppression: Active
          </span>
        </div>
      </div>

      {/* Upstream Suppression Banner */}
      <div className="p-4 rounded-2xl bg-purple-500/10 border border-purple-500/20 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <ShieldAlert className="h-5 w-5 text-purple-400 shrink-0" />
          <div>
            <div className="text-sm font-semibold text-purple-300">
              Upstream Root-Cause Identified: <span className="font-mono text-cyan-300">core-router-1</span>
            </div>
            <div className="text-xs text-muted-foreground mt-0.5">
              4 downstream child alerts have been automatically suppressed to eliminate alert storms.
            </div>
          </div>
        </div>
        <div className="px-3 py-1.5 rounded-xl bg-background/80 border border-purple-500/30 text-xs font-mono text-purple-300 font-medium whitespace-nowrap">
          4 Alerts Suppressed
        </div>
      </div>

      {/* Main Grid: Interactive Topology Graph & Hardware Driver Inspector */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Topology Graph / Hierarchy Tree */}
        <div className="lg:col-span-2 rounded-2xl border border-border/60 bg-card p-6 shadow-sm space-y-5">
          <div className="flex items-center justify-between border-b border-border/40 pb-3">
            <h2 className="text-sm font-semibold text-foreground flex items-center gap-2">
              <Network className="h-4 w-4 text-cyan-400" />
              Directed Acyclic Graph (DAG) Hierarchy
            </h2>
            <div className="flex items-center gap-3 text-[11px] text-muted-foreground">
              <span className="flex items-center gap-1">
                <span className="w-2 h-2 rounded-full bg-rose-400"></span> Root Cause
              </span>
              <span className="flex items-center gap-1">
                <span className="w-2 h-2 rounded-full bg-purple-400"></span> Suppressed
              </span>
              <span className="flex items-center gap-1">
                <span className="w-2 h-2 rounded-full bg-emerald-400"></span> Healthy
              </span>
            </div>
          </div>

          {/* Node Cards Tree */}
          <div className="space-y-6">
            {/* Level 1: Root Node */}
            <div className="flex justify-center">
              <button
                type="button"
                onClick={() => setSelectedDevice(networkTopology[0])}
                className={`p-4 rounded-2xl border text-left transition w-72 cursor-pointer ${
                  selectedDevice.id === "core-router-1"
                    ? "border-rose-500 bg-rose-500/15 shadow-md ring-2 ring-rose-500/30"
                    : "border-rose-500/50 bg-rose-500/10 hover:bg-rose-500/20"
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="text-[10px] font-bold uppercase tracking-wider text-rose-400">
                    Root Layer
                  </span>
                  <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-rose-500/20 text-rose-300 border border-rose-500/30">
                    ROOT INCIDENT
                  </span>
                </div>
                <div className="text-sm font-bold font-mono text-foreground mt-1">core-router-1</div>
                <div className="text-xs text-muted-foreground">{networkTopology[0].vendor}</div>
                <div className="text-[11px] font-mono text-rose-300 mt-2">Optical DOM: High Latency / Degraded</div>
              </button>
            </div>

            {/* Tree Branch Connector */}
            <div className="flex justify-center -my-3">
              <div className="w-0.5 h-6 bg-border/80"></div>
            </div>

            {/* Level 2: Distribution Layer */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              {networkTopology.slice(1, 3).map((dev) => {
                const isSelected = selectedDevice.id === dev.id;
                const isSuppressed = dev.status === "suppressed";

                return (
                  <button
                    key={dev.id}
                    type="button"
                    onClick={() => setSelectedDevice(dev)}
                    className={`p-3.5 rounded-2xl border text-left transition cursor-pointer ${
                      isSelected
                        ? "border-cyan-400 bg-cyan-500/15 ring-2 ring-cyan-500/30"
                        : isSuppressed
                        ? "border-purple-500/40 bg-purple-500/10 hover:bg-purple-500/15"
                        : "border-border/60 bg-secondary/30 hover:bg-secondary/60"
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] font-semibold text-muted-foreground uppercase">
                        Distribution
                      </span>
                      <span
                        className={`px-2 py-0.5 rounded text-[9px] font-bold ${
                          isSuppressed
                            ? "bg-purple-500/20 text-purple-300 border border-purple-500/30"
                            : "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                        }`}
                      >
                        {isSuppressed ? "SUPPRESSED" : "HEALTHY"}
                      </span>
                    </div>
                    <div className="text-xs font-bold font-mono text-foreground mt-1">{dev.name}</div>
                    <div className="text-[11px] text-muted-foreground">{dev.vendor}</div>
                  </button>
                );
              })}

              {/* APC UPS Node */}
              <button
                type="button"
                onClick={() => setSelectedDevice(networkTopology[5])}
                className={`p-3.5 rounded-2xl border text-left transition cursor-pointer ${
                  selectedDevice.id === networkTopology[5].id
                    ? "border-cyan-400 bg-cyan-500/15 ring-2 ring-cyan-500/30"
                    : "border-border/60 bg-secondary/30 hover:bg-secondary/60"
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="text-[10px] font-semibold text-muted-foreground uppercase">
                    Power
                  </span>
                  <span className="px-2 py-0.5 rounded text-[9px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                    ONLINE
                  </span>
                </div>
                <div className="text-xs font-bold font-mono text-foreground mt-1">apc-ups-rack-1</div>
                <div className="text-[11px] text-muted-foreground">Smart-UPS 1500 (100% Bat)</div>
              </button>
            </div>

            {/* Level 3: Edge Servers */}
            <div className="pt-2">
              <div className="text-xs font-semibold text-muted-foreground uppercase mb-2">
                Downstream Edge Servers (Parent: dist-switch-1)
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {networkTopology.slice(3, 5).map((dev) => {
                  const isSelected = selectedDevice.id === dev.id;

                  return (
                    <button
                      key={dev.id}
                      type="button"
                      onClick={() => setSelectedDevice(dev)}
                      className={`p-3.5 rounded-2xl border text-left transition cursor-pointer ${
                        isSelected
                          ? "border-cyan-400 bg-cyan-500/15 ring-2 ring-cyan-500/30"
                          : "border-purple-500/40 bg-purple-500/10 hover:bg-purple-500/15"
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <span className="text-[10px] font-semibold text-muted-foreground uppercase">
                          Host Server
                        </span>
                        <span className="px-2 py-0.5 rounded text-[9px] font-bold bg-purple-500/20 text-purple-300 border border-purple-500/30">
                          SUPPRESSED
                        </span>
                      </div>
                      <div className="text-xs font-bold font-mono text-foreground mt-1">{dev.name}</div>
                      <div className="text-[11px] text-muted-foreground">{dev.vendor}</div>
                    </button>
                  );
                })}
              </div>
            </div>
          </div>
        </div>

        {/* Selected Device Telemetry Panel */}
        <div className="rounded-2xl border border-border/60 bg-card p-5 shadow-sm flex flex-col justify-between space-y-4">
          <div>
            <div className="flex items-center justify-between pb-3 border-b border-border/50">
              <h3 className="text-sm font-semibold text-foreground">Hardware Driver Telemetry</h3>
              <span
                className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                  selectedDevice.status === "root-cause"
                    ? "bg-rose-500/20 text-rose-300 border border-rose-500/30"
                    : selectedDevice.status === "suppressed"
                    ? "bg-purple-500/20 text-purple-300 border border-purple-500/30"
                    : "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                }`}
              >
                {selectedDevice.status.toUpperCase()}
              </span>
            </div>

            <div className="mt-3">
              <div className="text-xs text-muted-foreground">Device Name:</div>
              <div className="text-lg font-mono font-bold text-foreground">{selectedDevice.name}</div>
              <div className="text-xs text-muted-foreground mt-0.5">{selectedDevice.vendor}</div>
              <div className="text-xs font-mono text-cyan-400 mt-1">IP: {selectedDevice.ip}</div>
            </div>

            <div className="mt-5 space-y-2.5 text-xs">
              {Object.entries(selectedDevice.metrics).map(([k, v]) => (
                <div key={k} className="flex justify-between items-center py-1.5 border-b border-border/40">
                  <span className="text-muted-foreground">{k}:</span>
                  <span className="font-mono font-semibold text-foreground truncate max-w-[180px]">
                    {v}
                  </span>
                </div>
              ))}
            </div>
          </div>

          <div className="p-3 bg-secondary/40 rounded-xl border border-border/50 text-[11px] text-muted-foreground space-y-1">
            <div className="font-semibold text-foreground">Driver SDK & Profiles:</div>
            <div>Loaded from <code className="text-cyan-400">drivers/profiles/*.yaml</code> via Collector daemon.</div>
          </div>
        </div>
      </div>
    </div>
  );
}
