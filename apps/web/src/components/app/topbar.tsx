"use client";

import Link from "next/link";
import { useState } from "react";
import { ChevronDown, Globe, LogOut, Menu, Moon, Settings, Sun, User } from "lucide-react";
import { useRouter } from "next/navigation";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiRequest } from "@/lib/api";
import { SessionResponse } from "@/lib/types";
import { useShellStore } from "@/store/ui-store";
import { translations } from "@/lib/i18n";

export function Topbar({ session }: { session: SessionResponse }) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { toggleSidebar, locale, toggleLocale, theme, toggleTheme } = useShellStore();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const isDemo = session.mode === "demo";
  const t = translations[locale];

  const logout = useMutation({
    mutationFn: () => apiRequest<{ ok: boolean }>("/api/auth/logout", { method: "POST", body: JSON.stringify({}) }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["session"] });
      router.replace("/login");
    },
  });

  return (
    <header className="topbar-shell sticky top-0 z-20 flex h-14 items-center justify-between gap-3 px-4 backdrop-blur-md lg:px-6 border-b border-border/40 bg-background/80">
      <div className="flex items-center gap-3">
        <button
          onClick={toggleSidebar}
          className="rounded-lg p-2 text-muted-foreground transition-colors hover:bg-cyan-300/10 hover:text-foreground lg:hidden"
        >
          <Menu className="h-5 w-5" />
        </button>
        <Link href="/" className="flex items-center gap-2.5">
          <div className="grid h-8 w-8 place-items-center rounded-lg border border-cyan-400/30 bg-gradient-to-tr from-cyan-500 via-blue-600 to-emerald-400 font-display text-xs font-black tracking-wider text-slate-950 shadow-[0_0_18px_rgba(64,240,255,0.2)]">
            QN
          </div>
          <div className="hidden items-baseline gap-2 sm:flex">
            <h1 className="font-display text-lg font-bold tracking-wide text-foreground">
              {t.brand}
            </h1>
            <span className="text-[10px] uppercase tracking-[0.14em] text-muted-foreground">
              {t.tagline}
            </span>
          </div>
        </Link>
      </div>

      <div className="flex items-center gap-2.5">
        {/* Cluster Active Indicator */}
        <div className="hidden md:flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
          <span>PulseMesh: 3 Nodes Active</span>
        </div>

        {/* Locale Switcher Button */}
        <button
          type="button"
          onClick={toggleLocale}
          className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg border border-border/60 hover:border-cyan-400/40 text-xs font-medium text-muted-foreground hover:text-foreground transition-colors bg-secondary/40"
          title="Switch Language (English / বাংলা)"
        >
          <Globe className="h-3.5 w-3.5 text-cyan-400" />
          <span className="font-semibold">{locale === "en" ? "EN" : "বাংলা"}</span>
        </button>

        {/* Theme Toggle Button */}
        <button
          type="button"
          onClick={toggleTheme}
          className="p-1.5 rounded-lg border border-border/60 hover:border-cyan-400/40 text-muted-foreground hover:text-foreground transition-colors bg-secondary/40"
          title="Toggle Theme"
        >
          {theme === "dark" ? <Sun className="h-3.5 w-3.5 text-amber-400" /> : <Moon className="h-3.5 w-3.5 text-cyan-400" />}
        </button>

        {isDemo ? (
          <span className="rounded-full border border-cyan-300/20 bg-cyan-300/10 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.14em] text-cyan-300">
            Demo
          </span>
        ) : null}

        <div className="relative">
          <button
            onClick={() => setDropdownOpen((current) => !current)}
            className="flex items-center gap-2 rounded-full border border-cyan-400/20 bg-cyan-400/5 px-2 py-1.5 transition-all hover:border-cyan-400/40"
          >
            <span className="grid h-7 w-7 place-items-center rounded-full bg-gradient-to-tr from-cyan-400 to-blue-500 text-[10px] font-bold tracking-wider text-slate-950 shadow-sm">
              {(session.user?.name?.trim() || session.user?.email?.trim() || "OP").slice(0, 2).toUpperCase()}
            </span>
            <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
          </button>

          {dropdownOpen ? (
            <>
              <div className="fixed inset-0 z-40" onClick={() => setDropdownOpen(false)} />
              <div className="absolute right-0 top-full z-50 mt-2 w-56 animate-fade-in rounded-xl border border-border bg-popover/95 p-2 shadow-xl backdrop-blur-lg">
                <div className="mb-1 border-b border-border/50 px-3 py-2">
                  <p className="text-sm font-medium text-foreground">{session.user?.name ?? "Operator"}</p>
                  <p className="text-xs text-muted-foreground truncate">{session.user?.email ?? "-"}</p>
                </div>

                <Link href="/settings" className="flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-foreground transition-colors hover:bg-secondary" onClick={() => setDropdownOpen(false)}>
                  <Settings className="h-4 w-4 text-muted-foreground" /> {t.nav.settings}
                </Link>
                <Link href="/profile" className="flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-foreground transition-colors hover:bg-secondary" onClick={() => setDropdownOpen(false)}>
                  <User className="h-4 w-4 text-muted-foreground" /> {t.nav.profile}
                </Link>

                <button
                  type="button"
                  className="flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-rose-400 transition-colors hover:bg-rose-500/10"
                  onClick={() => logout.mutate()}
                  disabled={logout.isPending}
                >
                  <LogOut className="h-4 w-4" /> {logout.isPending ? "..." : t.nav.logOut}
                </button>
              </div>
            </>
          ) : null}
        </div>
      </div>
    </header>
  );
}
