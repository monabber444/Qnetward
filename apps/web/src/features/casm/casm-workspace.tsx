"use client";

import { useState } from "react";
import {
  AlertTriangle,
  ArrowRight,
  CheckCircle2,
  ExternalLink,
  Filter,
  Flame,
  Globe,
  Radio,
  RefreshCw,
  Search,
  Shield,
  ShieldAlert,
  ShieldCheck,
  Terminal,
  Zap,
} from "lucide-react";
import { useShellStore } from "@/store/ui-store";

type TyposquatDomain = {
  id: string;
  brand: string;
  domain: string;
  type: "homoglyph" | "subdomain_impersonation" | "levenshtein";
  similarity: number;
  issuer: string;
  detectedAt: string;
  severity: "HIGH" | "MEDIUM" | "LOW";
  status: "INVESTIGATING" | "TAKEDOWN_REQUESTED" | "CONFIRMED_PHISHING" | "BENIGN";
};

type DASTReport = {
  target: string;
  statusCode: number;
  durationMs: number;
  grade: "A+" | "A" | "B" | "C" | "D" | "F";
  findings: {
    category: string;
    title: string;
    severity: "CRITICAL" | "HIGH" | "MEDIUM" | "LOW" | "INFO";
    description: string;
    remediation: string;
  }[];
  timestamp: string;
};

type OASTEvent = {
  token: string;
  target: string;
  protocol: "HTTP" | "DNS";
  remoteAddr: string;
  payloadType: string;
  timestamp: string;
};

const sampleTyposquats: TyposquatDomain[] = [
  {
    id: "sq-101",
    brand: "qnetward.com",
    domain: "qnetvvard.com",
    type: "homoglyph",
    similarity: 0.98,
    issuer: "Let's Encrypt Authority X3",
    detectedAt: "2 mins ago",
    severity: "HIGH",
    status: "INVESTIGATING",
  },
  {
    id: "sq-102",
    brand: "qnetward.com",
    domain: "qnetward.auth-verify.net",
    type: "subdomain_impersonation",
    similarity: 0.95,
    issuer: "ZeroSSL RSA Domain Secure Site CA",
    detectedAt: "14 mins ago",
    severity: "HIGH",
    status: "TAKEDOWN_REQUESTED",
  },
  {
    id: "sq-103",
    brand: "qnetward.com",
    domain: "qnetward-login.org",
    type: "levenshtein",
    similarity: 0.88,
    issuer: "Cloudflare Inc ECC CA-3",
    detectedAt: "45 mins ago",
    severity: "HIGH",
    status: "CONFIRMED_PHISHING",
  },
  {
    id: "sq-104",
    brand: "qnetward.com",
    domain: "qnetward1.com",
    type: "levenshtein",
    similarity: 0.89,
    issuer: "cPanel Inc Certification Authority",
    detectedAt: "1 hour ago",
    severity: "MEDIUM",
    status: "BENIGN",
  },
];

const sampleDASTReports: DASTReport[] = [
  {
    target: "https://api.qnetward.internal/v1/metrics",
    statusCode: 200,
    durationMs: 42,
    grade: "A+",
    findings: [],
    timestamp: "Just now",
  },
  {
    target: "https://staging-portal.legacy-corp.test",
    statusCode: 200,
    durationMs: 118,
    grade: "D",
    findings: [
      {
        category: "CORS_PERMISSIVE",
        title: "Wildcard CORS with Credentials Allowed",
        severity: "HIGH",
        description: "Access-Control-Allow-Origin is set to untrusted origin while Allow-Credentials is true.",
        remediation: "Restrict Access-Control-Allow-Origin to strictly enumerated, verified origins.",
      },
      {
        category: "HEADER_MISSING",
        title: "Missing Content Security Policy (CSP)",
        severity: "MEDIUM",
        description: "No Content-Security-Policy header returned, raising XSS risk.",
        remediation: "Configure a strict Content-Security-Policy with script-src restrictions.",
      },
      {
        category: "COOKIE_INSECURE",
        title: "Session Cookie Missing Secure Flag",
        severity: "MEDIUM",
        description: "Cookie 'session_id' lacks Secure attribute on HTTPS.",
        remediation: "Add 'Secure; HttpOnly; SameSite=Lax' to Set-Cookie header.",
      },
      {
        category: "INFO_LEAKAGE",
        title: "Detailed Server Banner Disclosed",
        severity: "LOW",
        description: "Server header discloses 'Apache/2.4.52 (Ubuntu)'.",
        remediation: "Set 'ServerTokens Prod' in web server configuration.",
      },
    ],
    timestamp: "12 mins ago",
  },
];

const sampleOASTEvents: OASTEvent[] = [
  {
    token: "e3f9a12c8b41",
    target: "https://partner-gateway.test/webhook",
    protocol: "HTTP",
    remoteAddr: "198.51.100.84:53210",
    payloadType: "BLIND_SSRF",
    timestamp: "3 mins ago",
  },
  {
    token: "9c04df17a29e",
    target: "https://erp.internal.test/query",
    protocol: "DNS",
    remoteAddr: "203.0.113.19:53",
    payloadType: "JNDI_LOOKUP",
    timestamp: "18 mins ago",
  },
];

export function CasmWorkspace() {
  const { locale } = useShellStore();
  const [activeTab, setActiveTab] = useState<"ctwatcher" | "dast" | "oast">("ctwatcher");
  const [typosquats, setTyposquats] = useState<TyposquatDomain[]>(sampleTyposquats);
  const [dastReports, setDASTReports] = useState<DASTReport[]>(sampleDASTReports);
  const [oastEvents] = useState<OASTEvent[]>(sampleOASTEvents);

  // New Audit target input
  const [scanTarget, setScanTarget] = useState("https://staging-billing.app.internal");
  const [isAuditing, setIsAuditing] = useState(false);

  const handleRunAudit = () => {
    if (!scanTarget) return;
    setIsAuditing(true);
    setTimeout(() => {
      const newReport: DASTReport = {
        target: scanTarget,
        statusCode: 200,
        durationMs: 84,
        grade: "B",
        findings: [
          {
            category: "HEADER_MISSING",
            title: "Missing X-Frame-Options Header",
            severity: "MEDIUM",
            description: "No anti-clickjacking frame restrictions configured.",
            remediation: "Set 'X-Frame-Options: SAMEORIGIN'.",
          },
          {
            category: "INFO_LEAKAGE",
            title: "X-Powered-By Header Disclosed",
            severity: "LOW",
            description: "X-Powered-By reveals 'Express'.",
            remediation: "Disable X-Powered-By header via app.disable('x-powered-by').",
          },
        ],
        timestamp: "Just now",
      };
      setDASTReports([newReport, ...dastReports]);
      setIsAuditing(false);
    }, 900);
  };

  const handleUpdateStatus = (id: string, newStatus: TyposquatDomain["status"]) => {
    setTyposquats(prev =>
      prev.map(t => (t.id === id ? { ...t, status: newStatus } : t))
    );
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
            <ShieldAlert className="h-7 w-7 text-amber-500" />
            {locale === "bn" ? "অ্যাটাক সারফেস ম্যানেজমেন্ট (CASM) ও স্যান্ডবক্সড ডাস্ট" : "Continuous Attack Surface Management (CASM) & Sandboxed DAST"}
          </h1>
          <p className="text-sm text-muted-foreground">
            {locale === "bn"
              ? "সার্টিফিকেট ট্রান্সপারেন্সি লাইভ স্ট্রিম, টাইপোস্কোয়াটিং রাডার, ডিফেন্সিভ ডাস্ট হেডার অডিট এবং ওএএসটি কররিলেশন লিসেনার।"
              : "Certificate Transparency firehose, lookalike phishing radar, automated defensive DAST auditor, and asynchronous OAST correlation."}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => {}}
            className="flex items-center gap-1.5 rounded-lg border border-border bg-card px-3 py-1.5 text-xs font-medium text-foreground hover:bg-muted"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            Live Sync
          </button>
        </div>
      </div>

      {/* KPI Stats */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="rounded-xl border border-border bg-card p-4">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-semibold uppercase tracking-wider">Protected Brands</span>
            <Shield className="h-4 w-4 text-primary" />
          </div>
          <div className="mt-2 text-2xl font-bold text-foreground">qnetward.com</div>
          <p className="mt-1 text-xs text-muted-foreground">CT logs monitored in real-time</p>
        </div>

        <div className="rounded-xl border border-border bg-card p-4">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-semibold uppercase tracking-wider">Typosquat Alerts</span>
            <Flame className="h-4 w-4 text-red-500" />
          </div>
          <div className="mt-2 text-2xl font-bold text-foreground">
            {typosquats.filter(t => t.severity === "HIGH").length} High
          </div>
          <p className="mt-1 text-xs text-red-400">Homoglyphs & lookalikes flagged</p>
        </div>

        <div className="rounded-xl border border-border bg-card p-4">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-semibold uppercase tracking-wider">DAST Audits Run</span>
            <ShieldCheck className="h-4 w-4 text-emerald-500" />
          </div>
          <div className="mt-2 text-2xl font-bold text-foreground">{dastReports.length} Endpoints</div>
          <p className="mt-1 text-xs text-emerald-400">Automated posture evaluated</p>
        </div>

        <div className="rounded-xl border border-border bg-card p-4">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="text-xs font-semibold uppercase tracking-wider">OAST Callbacks</span>
            <Radio className="h-4 w-4 text-cyan-500" />
          </div>
          <div className="mt-2 text-2xl font-bold text-foreground">{oastEvents.length} Captured</div>
          <p className="mt-1 text-xs text-cyan-400">Out-of-band correlation events</p>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="flex border-b border-border">
        <button
          onClick={() => setActiveTab("ctwatcher")}
          className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors ${
            activeTab === "ctwatcher"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Flame className="h-4 w-4" />
          {locale === "bn" ? "সার্টিফিকেট ট্রান্সপারেন্সি ও টাইপোস্কোয়াটিং" : "CT Log & Typosquat Radar"}
        </button>
        <button
          onClick={() => setActiveTab("dast")}
          className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors ${
            activeTab === "dast"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <ShieldAlert className="h-4 w-4" />
          {locale === "bn" ? "স্যান্ডবক্সড ডাস্ট স্ক্যানার" : "Sandboxed DAST Scanner"}
        </button>
        <button
          onClick={() => setActiveTab("oast")}
          className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors ${
            activeTab === "oast"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Radio className="h-4 w-4" />
          {locale === "bn" ? "ওএএসটি কররিলেশন লিসেনার" : "OAST Out-of-Band Listener"}
        </button>
      </div>

      {/* Tab 1: CT Log & Typosquat Radar */}
      {activeTab === "ctwatcher" && (
        <div className="space-y-4">
          <div className="rounded-xl border border-border bg-card p-4">
            <h3 className="text-base font-semibold text-foreground flex items-center justify-between">
              <span>Brand Impersonation & Lookalike Domains</span>
              <span className="text-xs font-normal text-muted-foreground">Streaming from CertStream & CT Logs</span>
            </h3>
            <p className="text-xs text-muted-foreground mt-1">
              Automated Levenshtein distance, visual homoglyph confusables (vv/w, 0/o, rn/m), and subdomain phishing detection.
            </p>

            <div className="mt-4 divide-y divide-border">
              {typosquats.map(t => (
                <div key={t.id} className="py-3.5 flex flex-col md:flex-row md:items-center md:justify-between gap-3">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="font-mono font-medium text-foreground text-sm">{t.domain}</span>
                      <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${
                        t.type === "homoglyph"
                          ? "bg-purple-500/10 text-purple-400 border border-purple-500/20"
                          : t.type === "subdomain_impersonation"
                          ? "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                          : "bg-blue-500/10 text-blue-400 border border-blue-500/20"
                      }`}>
                        {t.type.toUpperCase().replace("_", " ")}
                      </span>
                      <span className="text-xs text-muted-foreground font-mono">
                        {(t.similarity * 100).toFixed(0)}% Match
                      </span>
                    </div>
                    <div className="text-xs text-muted-foreground flex items-center gap-3">
                      <span>Monitored Brand: <b className="text-foreground">{t.brand}</b></span>
                      <span>•</span>
                      <span>Issuer: {t.issuer}</span>
                      <span>•</span>
                      <span>{t.detectedAt}</span>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 self-start md:self-auto">
                    <span className={`px-2.5 py-1 rounded text-xs font-semibold ${
                      t.status === "TAKEDOWN_REQUESTED"
                        ? "bg-amber-500/20 text-amber-300"
                        : t.status === "CONFIRMED_PHISHING"
                        ? "bg-red-500/20 text-red-300"
                        : t.status === "BENIGN"
                        ? "bg-emerald-500/20 text-emerald-300"
                        : "bg-blue-500/20 text-blue-300"
                    }`}>
                      {t.status.replace("_", " ")}
                    </span>

                    {t.status === "INVESTIGATING" && (
                      <button
                        onClick={() => handleUpdateStatus(t.id, "TAKEDOWN_REQUESTED")}
                        className="rounded border border-red-500/40 bg-red-500/10 px-2.5 py-1 text-xs font-medium text-red-400 hover:bg-red-500/20"
                      >
                        Request Takedown
                      </button>
                    )}
                    {t.status === "TAKEDOWN_REQUESTED" && (
                      <button
                        onClick={() => handleUpdateStatus(t.id, "CONFIRMED_PHISHING")}
                        className="rounded border border-red-500/40 bg-red-500/20 px-2.5 py-1 text-xs font-medium text-red-300"
                      >
                        Mark Blocked
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Tab 2: Sandboxed DAST Scanner */}
      {activeTab === "dast" && (
        <div className="space-y-6">
          <div className="rounded-xl border border-border bg-card p-4 space-y-3">
            <h3 className="text-base font-semibold text-foreground">Launch Automated DAST Security Audit</h3>
            <p className="text-xs text-muted-foreground">
              Defensively verifies CSP, HSTS, anti-clickjacking headers, permissive CORS credentials, and insecure cookie flags.
            </p>

            <div className="flex flex-col sm:flex-row gap-2">
              <input
                type="text"
                value={scanTarget}
                onChange={e => setScanTarget(e.target.value)}
                placeholder="https://your-service.corp/api"
                className="flex-1 rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary font-mono"
              />
              <button
                onClick={handleRunAudit}
                disabled={isAuditing}
                className="flex items-center justify-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
              >
                {isAuditing ? (
                  <>
                    <RefreshCw className="h-4 w-4 animate-spin" />
                    Scanning...
                  </>
                ) : (
                  <>
                    <Zap className="h-4 w-4" />
                    Audit Posture
                  </>
                )}
              </button>
            </div>
          </div>

          <div className="space-y-4">
            <h3 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Recent Audit Reports</h3>
            {dastReports.map((report, idx) => (
              <div key={idx} className="rounded-xl border border-border bg-card p-4 space-y-3">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                  <div className="space-y-0.5">
                    <div className="font-mono text-sm font-bold text-foreground flex items-center gap-2">
                      <Globe className="h-4 w-4 text-primary" />
                      {report.target}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      Status: {report.statusCode} OK • Latency: {report.durationMs}ms • {report.timestamp}
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <span className="text-xs text-muted-foreground font-semibold">Security Grade:</span>
                    <span className={`text-base font-black px-3 py-1 rounded-lg ${
                      report.grade === "A+" || report.grade === "A"
                        ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/40"
                        : report.grade === "B"
                        ? "bg-blue-500/20 text-blue-400 border border-blue-500/40"
                        : report.grade === "C"
                        ? "bg-amber-500/20 text-amber-400 border border-amber-500/40"
                        : "bg-red-500/20 text-red-400 border border-red-500/40"
                    }`}>
                      {report.grade}
                    </span>
                  </div>
                </div>

                {report.findings.length === 0 ? (
                  <div className="flex items-center gap-2 text-xs text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 rounded-lg p-2.5">
                    <CheckCircle2 className="h-4 w-4" />
                    No posture deficiencies detected. Full defense-in-depth headers and secure flags validated.
                  </div>
                ) : (
                  <div className="space-y-2 pt-2">
                    {report.findings.map((f, fIdx) => (
                      <div key={fIdx} className="rounded-lg border border-border/80 bg-background/50 p-3 text-xs space-y-1">
                        <div className="flex items-center justify-between">
                          <span className="font-semibold text-foreground flex items-center gap-1.5">
                            <AlertTriangle className={`h-3.5 w-3.5 ${
                              f.severity === "HIGH" ? "text-red-400" : f.severity === "MEDIUM" ? "text-amber-400" : "text-blue-400"
                            }`} />
                            {f.title}
                          </span>
                          <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                            f.severity === "HIGH"
                              ? "bg-red-500/10 text-red-400"
                              : f.severity === "MEDIUM"
                              ? "bg-amber-500/10 text-amber-400"
                              : "bg-blue-500/10 text-blue-400"
                          }`}>
                            {f.severity}
                          </span>
                        </div>
                        <p className="text-muted-foreground">{f.description}</p>
                        <p className="text-primary/90 font-mono text-[11px]">Remediation: {f.remediation}</p>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Tab 3: OAST Correlation Listener */}
      {activeTab === "oast" && (
        <div className="space-y-4">
          <div className="rounded-xl border border-border bg-card p-4 space-y-2">
            <h3 className="text-base font-semibold text-foreground flex items-center gap-2">
              <Radio className="h-4 w-4 text-cyan-400" />
              Out-of-Band Application Security Testing (OAST) Hub
            </h3>
            <p className="text-xs text-muted-foreground">
              Correlates asynchronous DNS lookups and HTTP callbacks triggered by blind injection probes (Blind SSRF, asynchronous JNDI/LDAP lookups).
            </p>
          </div>

          <div className="rounded-xl border border-border bg-card overflow-hidden">
            <table className="w-full text-left text-xs">
              <thead className="bg-muted text-muted-foreground border-b border-border font-semibold">
                <tr>
                  <th className="py-2.5 px-4">Correlation Token</th>
                  <th className="py-2.5 px-4">Target Audited</th>
                  <th className="py-2.5 px-4">Protocol</th>
                  <th className="py-2.5 px-4">Payload Type</th>
                  <th className="py-2.5 px-4">Remote Origin</th>
                  <th className="py-2.5 px-4">Captured At</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {oastEvents.map((event, idx) => (
                  <tr key={idx} className="hover:bg-muted/50">
                    <td className="py-3 px-4 font-mono font-bold text-cyan-400">{event.token}</td>
                    <td className="py-3 px-4 font-mono text-foreground">{event.target}</td>
                    <td className="py-3 px-4">
                      <span className="px-2 py-0.5 rounded font-mono font-bold bg-muted text-foreground">
                        {event.protocol}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-red-500/10 text-red-400 border border-red-500/20">
                        {event.payloadType}
                      </span>
                    </td>
                    <td className="py-3 px-4 font-mono text-muted-foreground">{event.remoteAddr}</td>
                    <td className="py-3 px-4 text-muted-foreground">{event.timestamp}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
