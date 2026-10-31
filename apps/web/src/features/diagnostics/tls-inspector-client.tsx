"use client";

import { useState } from "react";
import { CheckCircle2, Globe, Lock, Play, RefreshCw, ShieldCheck, Zap } from "lucide-react";
import { useShellStore } from "@/store/ui-store";
import { translations } from "@/lib/i18n";

export function TlsInspectorClient() {
  const { locale } = useShellStore();
  const t = translations[locale];

  const [host, setHost] = useState("cloudflare.com");
  const [port, setPort] = useState(443);
  const [isAuditing, setIsAuditing] = useState(false);
  const [result, setResult] = useState({
    tlsVersion: "TLS 1.3",
    cipherSuite: "TLS_AES_128_GCM_SHA256",
    alpn: ["h3", "h2", "http/1.1"],
    http3Supported: true,
    quicVersion: "0x00000001 (RFC 9000)",
    forwardSecrecy: true,
    certificate: {
      subject: "CN=cloudflare.com, O=Cloudflare Inc",
      issuer: "CN=Cloudflare Inc ECC CA-3",
      validFrom: "2026-01-10",
      validUntil: "2027-01-10",
      daysRemaining: 71,
      sanList: ["cloudflare.com", "*.cloudflare.com"],
      keyType: "ECDSA (secp256r1 - 256 bits)",
    },
    timing: {
      dnsMs: 4.2,
      tcpMs: 14.5,
      tlsHandshakeMs: 18.1,
      totalMs: 36.8,
    },
    grade: "A+",
  });

  const runAudit = () => {
    setIsAuditing(true);
    setTimeout(() => {
      setIsAuditing(false);
    }, 1200);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/40 pb-5">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-foreground flex items-center gap-2.5">
            <ShieldCheck className="h-5 w-5 text-emerald-400" />
            {t.nav.tlsHttp3}
          </h1>
          <p className="text-xs text-muted-foreground mt-1">
            Synthetic TLS 1.3 cryptographic auditing, ALPN negotiation, and HTTP/3 QUIC stream verification.
          </p>
        </div>
      </div>

      {/* Target Input */}
      <div className="rounded-2xl border border-border/60 bg-card p-5 shadow-sm space-y-4">
        <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
          <div className="sm:col-span-2">
            <label className="text-xs font-medium text-foreground block mb-1">Target FQDN</label>
            <input
              type="text"
              value={host}
              onChange={(e) => setHost(e.target.value)}
              className="w-full px-3 py-2 rounded-xl bg-background border border-border text-xs font-mono text-foreground focus:outline-none focus:border-cyan-400"
              placeholder="e.g. cloudflare.com"
            />
          </div>

          <div>
            <label className="text-xs font-medium text-foreground block mb-1">Port</label>
            <input
              type="number"
              value={port}
              onChange={(e) => setPort(Number(e.target.value))}
              className="w-full px-3 py-2 rounded-xl bg-background border border-border text-xs font-mono text-foreground focus:outline-none focus:border-cyan-400"
            />
          </div>

          <div className="flex items-end">
            <button
              type="button"
              onClick={runAudit}
              disabled={isAuditing}
              className="w-full py-2 rounded-xl bg-gradient-to-r from-cyan-500 to-blue-600 hover:from-cyan-400 hover:to-blue-500 text-slate-950 font-bold text-xs shadow-md transition disabled:opacity-50 flex items-center justify-center gap-2 cursor-pointer"
            >
              {isAuditing ? <RefreshCw className="h-3.5 w-3.5 animate-spin" /> : <Play className="h-3.5 w-3.5 fill-current" />}
              {isAuditing ? "Inspecting Cipher..." : "Inspect TLS & QUIC"}
            </button>
          </div>
        </div>
      </div>

      {/* Results Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Security Grade & Protocol Badges */}
        <div className="rounded-2xl border border-border/60 bg-card p-5 shadow-sm space-y-5">
          <div className="flex items-center justify-between border-b border-border/40 pb-3">
            <h2 className="text-sm font-semibold text-foreground">TLS Rating & ALPN</h2>
            <span className="text-xs font-mono text-muted-foreground">{host}:{port}</span>
          </div>

          <div className="flex items-center justify-center py-2">
            <div className="w-24 h-24 rounded-2xl bg-gradient-to-tr from-emerald-500/20 to-cyan-500/20 border-2 border-emerald-400/50 flex flex-col items-center justify-center text-emerald-400 shadow-lg shadow-emerald-500/10">
              <span className="text-3xl font-black">{result.grade}</span>
              <span className="text-[10px] font-semibold tracking-wider text-muted-foreground uppercase">SSL Score</span>
            </div>
          </div>

          <div className="space-y-2 text-xs">
            <div className="flex justify-between items-center py-1 border-b border-border/40">
              <span className="text-muted-foreground">Protocol Version:</span>
              <span className="font-mono font-bold text-emerald-400">{result.tlsVersion}</span>
            </div>
            <div className="flex justify-between items-center py-1 border-b border-border/40">
              <span className="text-muted-foreground">HTTP/3 (QUIC):</span>
              <span className="font-mono font-bold text-cyan-400">Supported (RFC 9000)</span>
            </div>
            <div className="flex justify-between items-center py-1 border-b border-border/40">
              <span className="text-muted-foreground">Negotiated ALPN:</span>
              <span className="font-mono text-foreground font-semibold">h3, h2, http/1.1</span>
            </div>
            <div className="flex justify-between items-center py-1">
              <span className="text-muted-foreground">Forward Secrecy (PFS):</span>
              <span className="font-mono text-emerald-400 font-semibold">Enabled (ECDHE)</span>
            </div>
          </div>
        </div>

        {/* Certificate Transparency & Chain Details */}
        <div className="lg:col-span-2 rounded-2xl border border-border/60 bg-card p-5 shadow-sm space-y-4">
          <div className="flex items-center justify-between border-b border-border/40 pb-3">
            <h2 className="text-sm font-semibold text-foreground flex items-center gap-2">
              <Lock className="h-4 w-4 text-cyan-400" />
              X.509 Certificate & Handshake Latency Breakdown
            </h2>
            <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              {result.certificate.daysRemaining} Days Remaining
            </span>
          </div>

          {/* Timing Breakdown Bar */}
          <div className="p-3 rounded-xl bg-secondary/30 border border-border/50 space-y-2">
            <div className="text-xs font-semibold text-foreground">Handshake Timing Breakdown (Total: {result.timing.totalMs} ms)</div>
            <div className="grid grid-cols-3 gap-2 text-center text-xs">
              <div className="p-2 rounded-lg bg-background border border-border">
                <div className="text-[10px] text-muted-foreground">DNS Lookup</div>
                <div className="font-mono font-bold text-cyan-300">{result.timing.dnsMs} ms</div>
              </div>
              <div className="p-2 rounded-lg bg-background border border-border">
                <div className="text-[10px] text-muted-foreground">TCP Connect</div>
                <div className="font-mono font-bold text-cyan-300">{result.timing.tcpMs} ms</div>
              </div>
              <div className="p-2 rounded-lg bg-background border border-border">
                <div className="text-[10px] text-muted-foreground">TLS 1.3 Handshake</div>
                <div className="font-mono font-bold text-emerald-400">{result.timing.tlsHandshakeMs} ms</div>
              </div>
            </div>
          </div>

          {/* Certificate Information */}
          <div className="space-y-2 text-xs">
            <div className="flex justify-between py-1 border-b border-border/40">
              <span className="text-muted-foreground">Common Name (Subject):</span>
              <span className="font-mono text-foreground font-semibold">{result.certificate.subject}</span>
            </div>
            <div className="flex justify-between py-1 border-b border-border/40">
              <span className="text-muted-foreground">Issuing Certificate Authority:</span>
              <span className="font-mono text-foreground">{result.certificate.issuer}</span>
            </div>
            <div className="flex justify-between py-1 border-b border-border/40">
              <span className="text-muted-foreground">Cipher Suite:</span>
              <span className="font-mono text-cyan-400">{result.cipherSuite}</span>
            </div>
            <div className="flex justify-between py-1 border-b border-border/40">
              <span className="text-muted-foreground">Public Key Algorithm:</span>
              <span className="font-mono text-foreground">{result.certificate.keyType}</span>
            </div>
            <div className="flex justify-between py-1">
              <span className="text-muted-foreground">Validity Window:</span>
              <span className="font-mono text-muted-foreground">{result.certificate.validFrom} to {result.certificate.validUntil}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
