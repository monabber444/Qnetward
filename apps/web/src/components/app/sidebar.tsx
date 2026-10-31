"use client";

import Link from "next/link";
import { useState } from "react";
import {
  Activity,
  AlertTriangle,
  Bot,
  Building2,
  ChevronDown,
  Cloud,
  Cpu,
  GitFork,
  Globe2,
  HardDrive,
  History,
  Lock,
  Network,
  Radio,
  Scan,
  Server,
  Settings,
  ShieldAlert,
  ShieldCheck,
  Terminal,
  Wrench,
} from "lucide-react";
import { usePathname } from "next/navigation";
import { useShellStore } from "@/store/ui-store";
import { SessionResponse } from "@/lib/types";
import { translations } from "@/lib/i18n";

export function Sidebar({ session }: { session: SessionResponse }) {
  const pathname = usePathname();
  const { sidebarOpen, setSidebarOpen, locale } = useShellStore();
  const isDemo = session.mode === "demo";
  const t = translations[locale];

  const navSections = [
    {
      label: t.nav.observability,
      icon: Activity,
      defaultOpen: true,
      items: [
        { href: "/", label: t.nav.overview, icon: Radio },
        { href: "/mesh", label: t.nav.pulseMesh, icon: Network },
        { href: "/topology", label: t.nav.topology, icon: Server },
        { href: "/cloud", label: t.nav.cloud, icon: Cloud },
        { href: "/aiops", label: t.nav.aiops, icon: Wrench },
        { href: "/bgp-twin", label: t.nav.bgpTwin, icon: GitFork },
        { href: "/copilot", label: t.nav.copilot, icon: Bot },
        { href: "/alerts", label: t.nav.alerts, icon: AlertTriangle },
      ],
    },
    {
      label: t.nav.security,
      icon: ShieldCheck,
      defaultOpen: true,
      items: [
        { href: "/security", label: t.nav.securityGate, icon: Lock },
        { href: "/casm", label: t.nav.casm, icon: ShieldAlert },
      ],
    },
    {
      label: t.nav.diagnostics,
      icon: Terminal,
      defaultOpen: true,
      items: [
        { href: "/icmp-ping", label: t.nav.icmpPing, icon: Terminal },
        { href: "/tcp-ping", label: t.nav.tcpPing, icon: Radio },
        { href: "/port-scan", label: t.nav.portScan, icon: Scan },
        { href: "/traceroute", label: t.nav.traceroute, icon: Network },
        { href: "/dns-lookup", label: t.nav.dnsLookup, icon: Globe2 },
        { href: "/tls-inspector", label: t.nav.tlsHttp3, icon: ShieldCheck },
      ],
    },
    {
      label: t.nav.infrastructure,
      icon: HardDrive,
      defaultOpen: true,
      items: [
        { href: "/appliances", label: t.nav.appliances, icon: Radio },
      ],
    },
    {
      label: t.nav.main,
      icon: Cpu,
      defaultOpen: false,
      items: [
        { href: "/msp-portal", label: t.nav.mspPortal, icon: Building2 },
        { href: "/history", label: t.nav.history, icon: History },
        { href: "/settings", label: t.nav.settings, icon: Settings },
      ],
    },
  ];

  const [openSections, setOpenSections] = useState<Set<string>>(() => {
    const initial = new Set<string>();
    navSections.forEach((sec) => {
      if (sec.defaultOpen || sec.items.some((item) => item.href === pathname)) {
        initial.add(sec.label);
      }
    });
    return initial;
  });

  const toggleSection = (label: string) => {
    setOpenSections((prev) => {
      const next = new Set(prev);
      if (next.has(label)) next.delete(label);
      else next.add(label);
      return next;
    });
  };

  return (
    <>
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-20 bg-background/60 backdrop-blur-sm lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      <aside
        className={`fixed top-0 left-0 z-30 h-full w-64 transition-transform duration-300 lg:sticky lg:translate-x-0 ${
          sidebarOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        <div className="sidebar-shell flex h-full flex-col overflow-y-auto border-r border-border/40 bg-sidebar">
          {/* Logo Brand Header */}
          <div className="px-5 py-4 border-b border-border/40 flex items-center gap-2.5">
            <div className="w-7 h-7 rounded-lg bg-gradient-to-tr from-cyan-500 to-emerald-400 flex items-center justify-center text-slate-950 font-black text-xs shadow-md">
              Q
            </div>
            <div>
              <div className="text-xs font-bold tracking-tight text-foreground uppercase">
                {t.brand} Mesh
              </div>
              <div className="text-[10px] text-muted-foreground">v2.0 SaaS Enterprise</div>
            </div>
          </div>

          <nav className="flex-1 space-y-1 px-3 py-4">
            {isDemo && (
              <div className="mx-1 mb-3 rounded-xl border border-cyan-400/20 bg-cyan-400/10 px-3 py-2 text-xs text-cyan-300">
                Live Multi-Node Environment
              </div>
            )}

            {navSections.map((sec) => {
              const isOpen = openSections.has(sec.label);
              const SecIcon = sec.icon;

              return (
                <div key={sec.label} className="mb-2">
                  <button
                    type="button"
                    onClick={() => toggleSection(sec.label)}
                    className="flex w-full items-center justify-between px-2.5 py-1.5 text-xs font-semibold text-muted-foreground hover:text-foreground rounded-lg transition-colors"
                  >
                    <span className="flex items-center gap-2">
                      <SecIcon className="h-3.5 w-3.5 text-cyan-400" />
                      <span>{sec.label}</span>
                    </span>
                    <ChevronDown
                      className={`h-3 w-3 transition-transform duration-200 ${
                        isOpen ? "rotate-0" : "-rotate-90"
                      }`}
                    />
                  </button>

                  {isOpen && (
                    <div className="mt-1 space-y-0.5 pl-2 border-l border-border/40 ml-3">
                      {sec.items.map((item) => {
                        const ItemIcon = item.icon;
                        const isActive = pathname === item.href;

                        return (
                          <Link
                            key={item.href}
                            href={item.href}
                            onClick={() => setSidebarOpen(false)}
                            className={`flex items-center gap-2.5 rounded-lg px-2.5 py-1.5 text-xs font-medium transition-all ${
                              isActive
                                ? "bg-cyan-400/15 text-cyan-300 font-semibold shadow-sm border border-cyan-400/20"
                                : "text-muted-foreground hover:bg-secondary/60 hover:text-foreground"
                            }`}
                          >
                            <ItemIcon className="h-3.5 w-3.5" />
                            <span>{item.label}</span>
                          </Link>
                        );
                      })}
                    </div>
                  )}
                </div>
              );
            })}
          </nav>
        </div>
      </aside>
    </>
  );
}
