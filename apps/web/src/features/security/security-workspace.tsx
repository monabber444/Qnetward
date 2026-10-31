"use client";

import { useState } from "react";
import {
  AlertCircle,
  Check,
  CheckCircle2,
  Copy,
  ExternalLink,
  Globe,
  KeyRound,
  Lock,
  Search,
  Shield,
  ShieldAlert,
  ShieldCheck,
  Slash,
} from "lucide-react";
import { useShellStore } from "@/store/ui-store";
import { translations } from "@/lib/i18n";

type VerificationMethod = "dns_txt" | "http_token";

export function SecurityWorkspace() {
  const { locale } = useShellStore();
  const t = translations[locale];

  const [domain, setDomain] = useState("api.example-enterprise.com");
  const [method, setMethod] = useState<VerificationMethod>("dns_txt");
  const [token, setToken] = useState("qnetward-verify=8f3b20c91e4a5d290b");
  const [isVerifying, setIsVerifying] = useState(false);
  const [isVerified, setIsVerified] = useState(false);
  const [copied, setCopied] = useState(false);
  const [cveSearch, setCveSearch] = useState("");

  const handleCopy = () => {
    navigator.clipboard.writeText(token);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleVerify = () => {
    setIsVerifying(true);
    setTimeout(() => {
      setIsVerifying(false);
      setIsVerified(true);
    }, 1200);
  };

  const sampleCves = [
    {
      cve: "CVE-2023-44487",
      title: "HTTP/2 Rapid Reset DDoS Vulnerability",
      severity: "HIGH",
      epss: "0.892 (89.2% probability)",
      cisaKev: true,
      service: "HTTP/2 Web Servers",
    },
    {
      cve: "CVE-2023-48795",
      title: "SSH Terrapin Prefix Truncation Attack",
      severity: "MEDIUM",
      epss: "0.245 (24.5% probability)",
      cisaKev: false,
      service: "OpenSSH <= 9.5",
    },
    {
      cve: "CVE-2024-3400",
      title: "Palo Alto PAN-OS Command Injection",
      severity: "CRITICAL",
      epss: "0.965 (96.5% probability)",
      cisaKev: true,
      service: "GlobalProtect Gateway",
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/40 pb-5">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-foreground flex items-center gap-2.5">
            <ShieldCheck className="h-5 w-5 text-emerald-400" />
            {t.security.title}
          </h1>
          <p className="text-xs text-muted-foreground mt-1">{t.security.subtitle}</p>
        </div>
        <div className="flex items-center gap-2 text-xs">
          <span className="px-3 py-1 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-medium">
            SSRF Guard: STRICT ENFORCEMENT
          </span>
        </div>
      </div>

      {/* 3-Tier Policy Status Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
        {/* Tier 0 */}
        <div className="rounded-2xl border border-emerald-500/30 bg-emerald-500/5 p-5 shadow-sm space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-emerald-400 uppercase tracking-wider">
              {t.security.tier0Title}
            </span>
            <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/20 text-emerald-300">
              ALWAYS ALLOWED
            </span>
          </div>
          <p className="text-xs text-muted-foreground">{t.security.tier0Desc}</p>
          <div className="pt-2 border-t border-emerald-500/20 text-[11px] text-muted-foreground space-y-1">
            <div className="flex justify-between">
              <span>CT Log Auditing:</span> <span className="font-semibold text-foreground">Enabled</span>
            </div>
            <div className="flex justify-between">
              <span>SPF / DMARC:</span> <span className="font-semibold text-emerald-400">Strict Reject</span>
            </div>
          </div>
        </div>

        {/* Tier 1 */}
        <div className="rounded-2xl border border-cyan-500/30 bg-cyan-500/5 p-5 shadow-sm space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-cyan-400 uppercase tracking-wider">
              {t.security.tier1Title}
            </span>
            <span
              className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                isVerified
                  ? "bg-emerald-500/20 text-emerald-300"
                  : "bg-amber-500/20 text-amber-300"
              }`}
            >
              {isVerified ? "UNLOCKED (VERIFIED)" : "403 GATED"}
            </span>
          </div>
          <p className="text-xs text-muted-foreground">{t.security.tier1Desc}</p>
          <div className="pt-2 border-t border-cyan-500/20 text-[11px] text-muted-foreground space-y-1">
            <div className="flex justify-between">
              <span>Ownership Token:</span>{" "}
              <span className={`font-semibold ${isVerified ? "text-emerald-400" : "text-amber-400"}`}>
                {isVerified ? "Valid TXT Found" : "Requires Verification"}
              </span>
            </div>
            <div className="flex justify-between">
              <span>Active Port Probes:</span>{" "}
              <span className="font-semibold text-foreground">{isVerified ? "Permitted" : "Denied (403)"}</span>
            </div>
          </div>
        </div>

        {/* Tier 2 */}
        <div className="rounded-2xl border border-purple-500/30 bg-purple-500/5 p-5 shadow-sm space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-purple-400 uppercase tracking-wider">
              {t.security.tier2Title}
            </span>
            <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-purple-500/20 text-purple-300">
              ENTERPRISE KYC ONLY
            </span>
          </div>
          <p className="text-xs text-muted-foreground">{t.security.tier2Desc}</p>
          <div className="pt-2 border-t border-purple-500/20 text-[11px] text-muted-foreground space-y-1">
            <div className="flex justify-between">
              <span>Sandboxed Workers:</span> <span className="font-semibold text-foreground">gVisor / Firecracker</span>
            </div>
            <div className="flex justify-between">
              <span>Kill Switch:</span> <span className="font-semibold text-rose-400">Arm Ready</span>
            </div>
          </div>
        </div>
      </div>

      {/* Domain Ownership Verification Workbench */}
      <div className="rounded-2xl border border-border/60 bg-card p-6 shadow-sm space-y-5">
        <div className="border-b border-border/40 pb-4">
          <h2 className="text-sm font-semibold text-foreground flex items-center gap-2">
            <KeyRound className="h-4 w-4 text-cyan-400" />
            {t.security.verifyDomainTitle}
          </h2>
          <p className="text-xs text-muted-foreground mt-0.5">{t.security.verifyDomainDesc}</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          {/* Form Controls */}
          <div className="space-y-4">
            <div>
              <label className="text-xs font-medium text-foreground block mb-1.5">Target Domain</label>
              <input
                type="text"
                value={domain}
                onChange={(e) => setDomain(e.target.value)}
                className="w-full px-3 py-2 rounded-xl bg-background border border-border text-xs font-mono text-foreground focus:outline-none focus:border-cyan-400"
                placeholder="domain.example.com"
              />
            </div>

            <div>
              <label className="text-xs font-medium text-foreground block mb-1.5">{t.security.verifyMethod}</label>
              <div className="grid grid-cols-2 gap-2">
                <button
                  type="button"
                  onClick={() => setMethod("dns_txt")}
                  className={`p-2.5 rounded-xl border text-xs font-medium transition cursor-pointer text-left ${
                    method === "dns_txt"
                      ? "border-cyan-400 bg-cyan-500/10 text-cyan-300"
                      : "border-border bg-secondary/30 text-muted-foreground"
                  }`}
                >
                  <div className="font-semibold">DNS TXT Record</div>
                  <div className="text-[10px] text-muted-foreground mt-0.5">_qnetward-verify.domain</div>
                </button>
                <button
                  type="button"
                  onClick={() => setMethod("http_token")}
                  className={`p-2.5 rounded-xl border text-xs font-medium transition cursor-pointer text-left ${
                    method === "http_token"
                      ? "border-cyan-400 bg-cyan-500/10 text-cyan-300"
                      : "border-border bg-secondary/30 text-muted-foreground"
                  }`}
                >
                  <div className="font-semibold">HTTP Token File</div>
                  <div className="text-[10px] text-muted-foreground mt-0.5">/.well-known/qnetward-verify.txt</div>
                </button>
              </div>
            </div>

            <div>
              <label className="text-xs font-medium text-foreground block mb-1.5">{t.security.verifyToken}</label>
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  readOnly
                  value={token}
                  className="w-full px-3 py-2 rounded-xl bg-background border border-border text-xs font-mono text-cyan-300 select-all"
                />
                <button
                  type="button"
                  onClick={handleCopy}
                  className="p-2 rounded-xl border border-border bg-secondary hover:bg-secondary/80 text-foreground transition"
                  title="Copy verification token"
                >
                  {copied ? <Check className="h-4 w-4 text-emerald-400" /> : <Copy className="h-4 w-4" />}
                </button>
              </div>
            </div>

            <button
              type="button"
              onClick={handleVerify}
              disabled={isVerifying}
              className="w-full py-2.5 rounded-xl bg-gradient-to-r from-cyan-500 to-blue-600 hover:from-cyan-400 hover:to-blue-500 text-slate-950 font-bold text-xs shadow-md transition disabled:opacity-50 cursor-pointer"
            >
              {isVerifying ? "Querying DNS TXT Records..." : t.security.verifyButton}
            </button>
          </div>

          {/* Verification Status Feedback Card */}
          <div className="rounded-2xl border border-border bg-secondary/20 p-5 flex flex-col justify-between space-y-4">
            <div>
              <div className="text-xs font-semibold text-foreground uppercase tracking-wider mb-2">
                Verification Instructions & Result
              </div>
              <div className="text-xs text-muted-foreground space-y-2">
                {method === "dns_txt" ? (
                  <p>
                    Add a TXT record to your DNS zone with name{" "}
                    <code className="text-cyan-400 bg-background px-1.5 py-0.5 rounded">
                      _qnetward-verify.{domain}
                    </code>{" "}
                    and value <code className="text-cyan-400 bg-background px-1.5 py-0.5 rounded">{token}</code>.
                  </p>
                ) : (
                  <p>
                    Serve a plain text file at{" "}
                    <code className="text-cyan-400 bg-background px-1.5 py-0.5 rounded">
                      https://{domain}/.well-known/qnetward-verify.txt
                    </code>{" "}
                    containing the token.
                  </p>
                )}
              </div>

              {isVerified && (
                <div className="mt-4 p-4 rounded-xl bg-emerald-500/10 border border-emerald-500/20 space-y-2">
                  <div className="flex items-center gap-2 text-emerald-400 font-semibold text-xs">
                    <CheckCircle2 className="h-4 w-4" />
                    <span>Domain Successfully Verified!</span>
                  </div>
                  <p className="text-[11px] text-muted-foreground">
                    Target <span className="text-foreground font-mono font-medium">{domain}</span> has been verified for Tenant #1. Tier 1 Safe Active Scans and port probing are now unlocked.
                  </p>
                </div>
              )}
            </div>

            {/* Anti-SSRF Protection Strip */}
            <div className="p-3.5 rounded-xl bg-rose-500/10 border border-rose-500/20 text-xs space-y-1">
              <div className="font-semibold text-rose-300 flex items-center gap-1.5">
                <Slash className="h-3.5 w-3.5" />
                <span>Anti-SSRF Kernel-Level Socket Filter</span>
              </div>
              <p className="text-[11px] text-muted-foreground">
                Even with verified targets, probing cloud metadata (<code className="text-rose-400">169.254.169.254</code>) and loopbacks (<code className="text-rose-400">127.0.0.1</code>) is strictly blocked by <code className="text-cyan-400">netguard.SafeDialer</code>.
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* CVE & CISA KEV Catalog Matcher */}
      <div className="rounded-2xl border border-border/60 bg-card p-6 shadow-sm space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-border/40 pb-4">
          <div>
            <h2 className="text-sm font-semibold text-foreground flex items-center gap-2">
              <Shield className="h-4 w-4 text-purple-400" />
              CVE, FIRST EPSS & CISA KEV Vulnerability Catalog
            </h2>
            <p className="text-xs text-muted-foreground mt-0.5">
              Live mapping of verified target banners against actively exploited CVE catalogs.
            </p>
          </div>
          <div className="relative w-full sm:w-64">
            <Search className="h-3.5 w-3.5 absolute left-3 top-3 text-muted-foreground" />
            <input
              type="text"
              value={cveSearch}
              onChange={(e) => setCveSearch(e.target.value)}
              placeholder="Filter CVE or service..."
              className="w-full pl-9 pr-3 py-1.5 rounded-xl bg-background border border-border text-xs text-foreground focus:outline-none focus:border-cyan-400"
            />
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse text-xs">
            <thead>
              <tr className="border-b border-border/60 text-muted-foreground font-semibold">
                <th className="p-3">CVE Identifier</th>
                <th className="p-3">Vulnerability Title</th>
                <th className="p-3">Severity</th>
                <th className="p-3">FIRST EPSS Score</th>
                <th className="p-3">CISA Known Exploited (KEV)</th>
                <th className="p-3">Affected Service</th>
              </tr>
            </thead>
            <tbody>
              {sampleCves.map((c) => (
                <tr key={c.cve} className="border-b border-border/40 hover:bg-muted/30 transition">
                  <td className="p-3 font-mono font-bold text-cyan-400">{c.cve}</td>
                  <td className="p-3 font-medium text-foreground">{c.title}</td>
                  <td className="p-3">
                    <span
                      className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                        c.severity === "CRITICAL"
                          ? "bg-rose-500/20 text-rose-300"
                          : c.severity === "HIGH"
                          ? "bg-amber-500/20 text-amber-300"
                          : "bg-blue-500/20 text-blue-300"
                      }`}
                    >
                      {c.severity}
                    </span>
                  </td>
                  <td className="p-3 font-mono text-foreground">{c.epss}</td>
                  <td className="p-3">
                    {c.cisaKev ? (
                      <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-rose-500/20 text-rose-300 border border-rose-500/30">
                        IN CISA KEV
                      </span>
                    ) : (
                      <span className="text-muted-foreground">-</span>
                    )}
                  </td>
                  <td className="p-3 font-mono text-muted-foreground">{c.service}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
