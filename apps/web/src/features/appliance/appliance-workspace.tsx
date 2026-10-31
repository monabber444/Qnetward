"use client";

import { useState } from "react";
import {
  Activity,
  AlertTriangle,
  ArrowRight,
  CheckCircle2,
  Cpu,
  HardDrive,
  QrCode,
  Radio,
  RefreshCw,
  Signal,
  Terminal,
  Wifi,
  Zap,
} from "lucide-react";
import { useShellStore } from "@/store/ui-store";

type ApplianceItem = {
  id: string;
  serial: string;
  model: string;
  location: string;
  firmware: string;
  wanStatus: "CONNECTED" | "DOWN";
  cellularStatus: "STANDBY" | "ACTIVE_FAILOVER";
  carrier: string;
  rsrpDbm: number;
  activeUplink: "PRIMARY_WAN" | "CELLULAR_5G";
  lastPing: string;
};

const sampleAppliances: ApplianceItem[] = [
  {
    id: "probe-lon-01",
    serial: "SN-RPI5-98412",
    model: "Raspberry Pi 5 CM4",
    location: "London Oxford St Retail Branch",
    firmware: "qnetward-probe-v1.4.2",
    wanStatus: "CONNECTED",
    cellularStatus: "STANDBY",
    carrier: "Vodafone UK 5G",
    rsrpDbm: -74,
    activeUplink: "PRIMARY_WAN",
    lastPing: "12s ago",
  },
  {
    id: "probe-fra-02",
    serial: "SN-NUC-41092",
    model: "Intel NUC Edge Rugged",
    location: "Frankfurt Logistics Hub",
    firmware: "qnetward-probe-v1.4.2",
    wanStatus: "DOWN",
    cellularStatus: "ACTIVE_FAILOVER",
    carrier: "Telekom 5G Ultra",
    rsrpDbm: -82,
    activeUplink: "CELLULAR_5G",
    lastPing: "Just now",
  },
  {
    id: "probe-sin-03",
    serial: "SN-ODROID-10492",
    model: "Odroid M1 Industrial",
    location: "Singapore Jurong Plant",
    firmware: "qnetward-probe-v1.4.2",
    wanStatus: "CONNECTED",
    cellularStatus: "STANDBY",
    carrier: "Singtel 5G IoT",
    rsrpDbm: -68,
    activeUplink: "PRIMARY_WAN",
    lastPing: "8s ago",
  },
  {
    id: "probe-nyc-04",
    serial: "SN-X86-77192",
    model: "X86 Mini Edge PC",
    location: "New York Midtown Branch",
    firmware: "qnetward-probe-v1.4.2",
    wanStatus: "CONNECTED",
    cellularStatus: "STANDBY",
    carrier: "Verizon 5G Enterprise",
    rsrpDbm: -71,
    activeUplink: "PRIMARY_WAN",
    lastPing: "15s ago",
  },
];

export function ApplianceWorkspace() {
  const { locale } = useShellStore();
  const [appliances, setAppliances] = useState<ApplianceItem[]>(sampleAppliances);
  const [activeTab, setActiveTab] = useState<"fleet" | "ztp" | "oobm">("fleet");

  // ZTP state
  const [ztpDeviceName, setZtpDeviceName] = useState("probe-tokyo-05");
  const [generatedToken, setGeneratedToken] = useState<string | null>(null);

  // OOBM serial console simulation
  const [consoleLogs, setConsoleLogs] = useState<string[]>([
    "[OOBM-SYS] 20:28:14 wwan0 connected to Telekom 5G (RSRP: -82 dBm, Band n78)",
    "[OOBM-WAN] 20:28:15 eth0 Primary WAN gateway unreachable (100% packet loss)",
    "[OOBM-AIOPS] 20:28:16 Automatic uplink failover executed -> Traffic routed via wwan0",
    "[OOBM-DIAG] 20:28:17 Root cause: Upstream ISP fiber cut outside building perimeter",
    "[OOBM-SERIAL] 20:28:18 Emergency RS-232 serial console connected at /dev/ttyUSB0 (9600 8N1)",
  ]);

  const handleGenerateZTP = () => {
    const token = "ztp_" + Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
    setGeneratedToken(token);
  };

  const handleSendConsoleCommand = (cmd: string) => {
    setConsoleLogs(prev => [
      ...prev,
      `admin@probe-fra-02:~$ ${cmd}`,
      `[SYS-RESP] Executed: ${cmd} (Success, exit 0)`,
    ]);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
            <Radio className="h-7 w-7 text-cyan-500" />
            {locale === "bn"
              ? "মাইক্রো-প্রোব অ্যাপ্লায়েন্স ও ৫জি আউট-অব-ব্যান্ড (OOBM)"
              : "Plug & Play Micro-Probe Appliances & 5G Out-of-Band"}
          </h1>
          <p className="text-sm text-muted-foreground">
            {locale === "bn"
              ? "জিরো-টাচ প্রভিশনিং (ZTP), র‍্যাস্পবেরি পাই/ইন্টেল নাক ফার্মওয়্যার এবং সেলুলার ৫জি ইমার্জেন্সি সিরিয়াল কনসোল।"
              : "Zero-Touch Provisioning (ZTP), Raspberry Pi & Intel NUC firmware, and automatic 5G cellular out-of-band failover."}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className="rounded-lg border border-cyan-500/30 bg-cyan-500/10 px-3 py-1.5 text-xs font-semibold text-cyan-400 flex items-center gap-1.5">
            <Signal className="h-4 w-4" />
            Active 5G Cellular Failover Available
          </span>
        </div>
      </div>

      {/* KPI Stats */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="rounded-xl border border-border bg-card p-4">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-semibold uppercase tracking-wider">Active Micro-Probes</span>
            <HardDrive className="h-4 w-4 text-primary" />
          </div>
          <div className="mt-2 text-2xl font-bold text-foreground">{appliances.length} Deployed</div>
          <p className="mt-1 text-xs text-muted-foreground">Branch offices & edge locations</p>
        </div>

        <div className="rounded-xl border border-border bg-card p-4">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-semibold uppercase tracking-wider">Active 5G Failover</span>
            <Radio className="h-4 w-4 text-amber-500" />
          </div>
          <div className="mt-2 text-2xl font-bold text-amber-400">
            {appliances.filter(a => a.activeUplink === "CELLULAR_5G").length} Branch
          </div>
          <p className="mt-1 text-xs text-amber-400/80">Primary WAN fiber cut active</p>
        </div>

        <div className="rounded-xl border border-border bg-card p-4">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-semibold uppercase tracking-wider">Firmware Fleet</span>
            <Cpu className="h-4 w-4 text-emerald-500" />
          </div>
          <div className="mt-2 text-2xl font-bold text-foreground">v1.4.2 Alpine</div>
          <p className="mt-1 text-xs text-emerald-400">100% fleet synchronized</p>
        </div>

        <div className="rounded-xl border border-border bg-card p-4">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-semibold uppercase tracking-wider">ZTP Status</span>
            <QrCode className="h-4 w-4 text-cyan-500" />
          </div>
          <div className="mt-2 text-2xl font-bold text-foreground">Ready</div>
          <p className="mt-1 text-xs text-cyan-400">QR / DHCP Zero-Touch enrollment</p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-border">
        <button
          onClick={() => setActiveTab("fleet")}
          className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors ${
            activeTab === "fleet"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <HardDrive className="h-4 w-4" />
          {locale === "bn" ? "অ্যাপ্লায়েন্স ফ্লিট" : "Micro-Probe Fleet"}
        </button>
        <button
          onClick={() => setActiveTab("ztp")}
          className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors ${
            activeTab === "ztp"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <QrCode className="h-4 w-4" />
          {locale === "bn" ? "জিরো-টাচ প্রভিশনিং (ZTP)" : "Zero-Touch Provisioning (ZTP)"}
        </button>
        <button
          onClick={() => setActiveTab("oobm")}
          className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors ${
            activeTab === "oobm"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Terminal className="h-4 w-4" />
          {locale === "bn" ? "৫জি আউট-অব-ব্যান্ড সিরিয়াল কনসোল" : "5G Out-of-Band Serial Console"}
        </button>
      </div>

      {/* TAB 1: Fleet */}
      {activeTab === "fleet" && (
        <div className="space-y-4">
          <div className="rounded-xl border border-border bg-card overflow-hidden">
            <table className="w-full text-left text-xs">
              <thead className="bg-muted text-muted-foreground border-b border-border font-semibold">
                <tr>
                  <th className="py-2.5 px-4">Appliance</th>
                  <th className="py-2.5 px-4">Hardware Model</th>
                  <th className="py-2.5 px-4">Location</th>
                  <th className="py-2.5 px-4">Primary WAN</th>
                  <th className="py-2.5 px-4">5G Cellular OOBM</th>
                  <th className="py-2.5 px-4">Active Uplink</th>
                  <th className="py-2.5 px-4">Last Ping</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {appliances.map(app => (
                  <tr key={app.id} className="hover:bg-muted/50">
                    <td className="py-3 px-4">
                      <div className="font-mono font-bold text-foreground">{app.id}</div>
                      <div className="text-[11px] text-muted-foreground">{app.serial}</div>
                    </td>
                    <td className="py-3 px-4 font-medium text-foreground">{app.model}</td>
                    <td className="py-3 px-4 text-muted-foreground">{app.location}</td>
                    <td className="py-3 px-4">
                      <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                        app.wanStatus === "CONNECTED"
                          ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                          : "bg-red-500/10 text-red-400 border border-red-500/20"
                      }`}>
                        {app.wanStatus}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-1.5 font-mono text-[11px]">
                        <Signal className="h-3.5 w-3.5 text-cyan-400" />
                        <span className="text-foreground">{app.carrier}</span>
                        <span className="text-muted-foreground">({app.rsrpDbm} dBm)</span>
                      </div>
                    </td>
                    <td className="py-3 px-4">
                      <span className={`px-2.5 py-1 rounded text-xs font-bold ${
                        app.activeUplink === "PRIMARY_WAN"
                          ? "bg-emerald-500/20 text-emerald-300 border border-emerald-500/40"
                          : "bg-amber-500/20 text-amber-300 border border-amber-500/40"
                      }`}>
                        {app.activeUplink === "PRIMARY_WAN" ? "PRIMARY WAN" : "5G FAILOVER"}
                      </span>
                    </td>
                    <td className="py-3 px-4 font-mono text-muted-foreground">{app.lastPing}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 2: Zero-Touch Provisioning */}
      {activeTab === "ztp" && (
        <div className="space-y-6">
          <div className="rounded-xl border border-border bg-card p-4 space-y-3">
            <h3 className="text-base font-semibold text-foreground flex items-center gap-2">
              <QrCode className="h-5 w-5 text-cyan-400" />
              Enroll New Hardware Micro-Probe via ZTP
            </h3>
            <p className="text-xs text-muted-foreground">
              Generate a one-time cryptographic enrollment token. When the appliance boots the minimal Alpine Linux image and grabs DHCP, it will automatically connect, provision its mTLS certificates, and join the PulseMesh cluster.
            </p>

            <div className="flex flex-col sm:flex-row gap-2 max-w-md">
              <input
                type="text"
                value={ztpDeviceName}
                onChange={e => setZtpDeviceName(e.target.value)}
                placeholder="probe-target-name"
                className="flex-1 rounded-lg border border-border bg-background px-3 py-2 text-sm font-mono text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
              />
              <button
                onClick={handleGenerateZTP}
                className="flex items-center justify-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90"
              >
                <Zap className="h-4 w-4" />
                Generate ZTP Token
              </button>
            </div>

            {generatedToken && (
              <div className="mt-4 rounded-lg bg-background p-4 border border-border space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-semibold text-cyan-400 uppercase">Cryptographic Enrollment Payload</span>
                  <span className="text-xs text-muted-foreground">Expires in 15 minutes</span>
                </div>
                <div className="p-3 bg-muted rounded font-mono text-xs text-foreground break-all border border-border">
                  {generatedToken}
                </div>

                <div className="text-xs text-muted-foreground space-y-1">
                  <p className="font-semibold text-foreground">Zero-Touch Provisioning Instructions:</p>
                  <ol className="list-decimal pl-4 space-y-1">
                    <li>Flash the Qnetward Embedded Alpine firmware to micro-SD / USB drive.</li>
                    <li>Copy enrollment token into <code>/boot/ztp.conf</code> (or scan generated QR code).</li>
                    <li>Connect appliance to Ethernet or insert 5G SIM card, then power on.</li>
                    <li>Device establishes outbound mTLS tunnel to Qnetward Ingestion Gateway.</li>
                  </ol>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* TAB 3: 5G Out-of-Band Serial Console */}
      {activeTab === "oobm" && (
        <div className="space-y-4">
          <div className="rounded-xl border border-border bg-card p-4 space-y-3">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-border pb-3">
              <div>
                <h3 className="text-base font-semibold text-foreground flex items-center gap-2">
                  <Terminal className="h-5 w-5 text-amber-400" />
                  Remote Serial Console (probe-fra-02 • 5G Out-of-Band)
                </h3>
                <p className="text-xs text-muted-foreground">
                  Primary WAN link is down. Emergency diagnostics session tunneled securely over cellular 5G.
                </p>
              </div>

              <span className="px-2.5 py-1 rounded text-xs font-bold bg-amber-500/20 text-amber-300 border border-amber-500/40 self-start sm:self-auto">
                ACTIVE 5G OOBM SESSION
              </span>
            </div>

            {/* Terminal Window */}
            <div className="rounded-lg bg-black border border-border p-4 font-mono text-xs text-emerald-400 space-y-1 h-64 overflow-y-auto">
              {consoleLogs.map((log, idx) => (
                <div key={idx}>{log}</div>
              ))}
            </div>

            {/* Quick Actions */}
            <div className="flex flex-wrap gap-2 pt-1">
              <button
                onClick={() => handleSendConsoleCommand("show ip interface brief")}
                className="rounded border border-border bg-background px-3 py-1.5 text-xs font-medium text-foreground hover:bg-muted font-mono"
              >
                show ip interface brief
              </button>
              <button
                onClick={() => handleSendConsoleCommand("ping 8.8.8.8 -I wwan0")}
                className="rounded border border-border bg-background px-3 py-1.5 text-xs font-medium text-foreground hover:bg-muted font-mono"
              >
                ping via 5G wwan0
              </button>
              <button
                onClick={() => handleSendConsoleCommand("modem-cli at+csq")}
                className="rounded border border-border bg-background px-3 py-1.5 text-xs font-medium text-foreground hover:bg-muted font-mono"
              >
                check cellular signal (AT+CSQ)
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
