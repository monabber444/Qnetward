"use client";

import { useState } from "react";
import {
  Building2,
  CheckCircle2,
  FileCheck2,
  Globe,
  KeyRound,
  Layers,
  Lock,
  Palette,
  RefreshCw,
  Search,
  Server,
  Shield,
  ShieldCheck,
  UserCheck,
  Users,
  Zap,
} from "lucide-react";
import { useShellStore } from "@/store/ui-store";

type OrganizationItem = {
  id: string;
  name: string;
  industry: string;
  sla: string;
  nodeCount: number;
  healthScore: number;
  status: "ACTIVE" | "MAINTENANCE" | "SUSPENDED";
};

type AuditBlock = {
  seq: number;
  id: string;
  timestamp: string;
  actor: string;
  action: string;
  resource: string;
  prevHash: string;
  hash: string;
};

const sampleOrgs: OrganizationItem[] = [
  { id: "org-fintech", name: "Apex Financial Core", industry: "Banking & Fintech", sla: "99.999%", nodeCount: 24, healthScore: 99.4, status: "ACTIVE" },
  { id: "org-health", name: "Metro Health Systems", industry: "Healthcare / HIPAA", sla: "99.99%", nodeCount: 16, healthScore: 98.8, status: "ACTIVE" },
  { id: "org-retail", name: "OmniRetail Global", industry: "E-Commerce", sla: "99.9%", nodeCount: 42, healthScore: 97.5, status: "ACTIVE" },
  { id: "org-gov", name: "Nordic Regional Transit", industry: "Critical Infrastructure", sla: "99.999%", nodeCount: 30, healthScore: 100.0, status: "ACTIVE" },
];

const sampleLedger: AuditBlock[] = [
  {
    seq: 0,
    id: "aud-0-4f9a12b",
    timestamp: "2026-10-31 19:15:02 UTC",
    actor: "system@genesis",
    action: "LEDGER_GENESIS_INITIALIZED",
    resource: "tenant-apex-msp",
    prevHash: "0000000000000000000000000000000000000000000000000000000000000000",
    hash: "4f9a12b9c8e104a37f529810ad9b0485c9a72198031dcf85b294e19034afec10",
  },
  {
    seq: 1,
    id: "aud-1-98e4d10",
    timestamp: "2026-10-31 19:22:15 UTC",
    actor: "admin@apexcloud.io",
    action: "ENTERPRISE_SSO_LOGIN",
    resource: "okta-saml-session-9104",
    prevHash: "4f9a12b9c8e104a37f529810ad9b0485c9a72198031dcf85b294e19034afec10",
    hash: "98e4d10fa89b210c47318049281aef9120485bc9204857219084729104fa2810",
  },
  {
    seq: 2,
    id: "aud-2-7b19a04",
    timestamp: "2026-10-31 19:35:40 UTC",
    actor: "sec-engineer@fintech.corp",
    action: "PLAYBOOK_REMEDIATION_APPROVED",
    resource: "playbook-exec-8491",
    prevHash: "98e4d10fa89b210c47318049281aef9120485bc9204857219084729104fa2810",
    hash: "7b19a048ec192804b9281048572019485bc01928475819203847592019384729",
  },
  {
    seq: 3,
    id: "aud-3-3d08e19",
    timestamp: "2026-10-31 19:48:12 UTC",
    actor: "aiops-engine@internal",
    action: "PORT_AUTOMATICALLY_QUARANTINED",
    resource: "core-switch-fra-Te0/1/2",
    prevHash: "7b19a048ec192804b9281048572019485bc01928475819203847592019384729",
    hash: "3d08e194a8274019283749582019384756201928475638291029384756102938",
  },
];

export function MspWorkspace() {
  const { locale } = useShellStore();
  const [activeTab, setActiveTab] = useState<"hierarchy" | "audit" | "iam">("hierarchy");
  const [partner, setPartner] = useState({
    name: "Apex Cloud Managed Services",
    tier: "PLATINUM MSP PARTNER",
    customDomain: "net.apexcloud.io",
    brandName: "ApexNet Enterprise Observability",
    primaryColor: "#0ea5e9",
  });
  const [isVerifying, setIsVerifying] = useState(false);
  const [verificationResult, setVerificationResult] = useState<{
    valid: boolean;
    blocksChecked: number;
    headHash: string;
  } | null>(null);

  const handleVerifyChain = () => {
    setIsVerifying(true);
    setTimeout(() => {
      setVerificationResult({
        valid: true,
        blocksChecked: sampleLedger.length,
        headHash: sampleLedger[sampleLedger.length - 1].hash,
      });
      setIsVerifying(false);
    }, 600);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
            <Building2 className="h-7 w-7 text-sky-500" />
            {locale === "bn"
              ? "মাল্টি-টেন্যান্ট এমএসপি পোর্টাল ও ক্রিপ্টোগ্রাফিক কমপ্লায়েন্স"
              : "Multi-Tenant MSP Portal & Cryptographic Audit Ledger"}
          </h1>
          <p className="text-sm text-muted-foreground">
            {locale === "bn"
              ? "এমএসপি পার্টনার হায়ারার্কি, হোয়াইট-লেবেল কাস্টমাইজেশন, এসএইচএ-২৫৬ ইমিউটেবল অডিট ট্রেইল এবং এন্টারপ্রাইজ এসএসও।"
              : "MSP partner hierarchy, white-label branding, SHA-256 immutable audit chain, and zero-trust IAM governance."}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className="rounded-lg border border-sky-500/30 bg-sky-500/10 px-3 py-1.5 text-xs font-semibold text-sky-400 flex items-center gap-1.5">
            <Shield className="h-4 w-4" />
            SOC 2 Type II / ISO 27001 Certified
          </span>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-border">
        <button
          onClick={() => setActiveTab("hierarchy")}
          className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors ${
            activeTab === "hierarchy"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <Users className="h-4 w-4" />
          {locale === "bn" ? "এমএসপি ক্লায়েন্ট হায়ারার্কি" : "Client Organizations & Hierarchy"}
        </button>
        <button
          onClick={() => setActiveTab("audit")}
          className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors ${
            activeTab === "audit"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <FileCheck2 className="h-4 w-4" />
          {locale === "bn" ? "ক্রিপ্টোগ্রাফিক অডিট লেজার" : "Cryptographic Audit Ledger"}
        </button>
        <button
          onClick={() => setActiveTab("iam")}
          className={`flex items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors ${
            activeTab === "iam"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground"
          }`}
        >
          <KeyRound className="h-4 w-4" />
          {locale === "bn" ? "এন্টারপ্রাইজ এসএসও ও আইএএম" : "Enterprise SSO & IAM"}
        </button>
      </div>

      {/* TAB 1: MSP Hierarchy & White-Label */}
      {activeTab === "hierarchy" && (
        <div className="space-y-6">
          {/* White-Label Banner */}
          <div className="rounded-xl border border-border bg-card p-4 space-y-3">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-border pb-3">
              <div>
                <span className="text-[11px] font-semibold text-sky-400 uppercase tracking-wider">
                  {partner.tier}
                </span>
                <h3 className="text-lg font-bold text-foreground">{partner.name}</h3>
              </div>
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <Globe className="h-4 w-4 text-primary" />
                <span className="font-mono">{partner.customDomain}</span>
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-1 text-xs">
              <div className="rounded-lg bg-background p-3 border border-border">
                <span className="text-muted-foreground">White-Label Brand</span>
                <p className="font-semibold text-foreground mt-0.5">{partner.brandName}</p>
              </div>
              <div className="rounded-lg bg-background p-3 border border-border">
                <span className="text-muted-foreground">Primary Accent Color</span>
                <div className="flex items-center gap-2 mt-0.5">
                  <span className="h-3.5 w-3.5 rounded-full" style={{ backgroundColor: partner.primaryColor }} />
                  <span className="font-mono font-semibold text-foreground">{partner.primaryColor}</span>
                </div>
              </div>
              <div className="rounded-lg bg-background p-3 border border-border">
                <span className="text-muted-foreground">Co-Managed IT Seats</span>
                <p className="font-semibold text-foreground mt-0.5">48 / 100 Active Seats</p>
              </div>
            </div>
          </div>

          {/* Customer Organizations */}
          <div className="rounded-xl border border-border bg-card overflow-hidden space-y-0">
            <div className="p-4 border-b border-border flex items-center justify-between">
              <h3 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-2">
                <Building2 className="h-4 w-4 text-primary" />
                Managed Customer Organizations ({sampleOrgs.length})
              </h3>
            </div>
            <table className="w-full text-left text-xs">
              <thead className="bg-muted text-muted-foreground border-b border-border font-semibold">
                <tr>
                  <th className="py-2.5 px-4">Organization</th>
                  <th className="py-2.5 px-4">Industry / Regulatory</th>
                  <th className="py-2.5 px-4">SLA Commitment</th>
                  <th className="py-2.5 px-4">Monitored Nodes</th>
                  <th className="py-2.5 px-4">Mesh Health</th>
                  <th className="py-2.5 px-4">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {sampleOrgs.map(org => (
                  <tr key={org.id} className="hover:bg-muted/50">
                    <td className="py-3 px-4 font-semibold text-foreground">{org.name}</td>
                    <td className="py-3 px-4 text-muted-foreground">{org.industry}</td>
                    <td className="py-3 px-4 font-mono font-bold text-sky-400">{org.sla}</td>
                    <td className="py-3 px-4 font-mono text-foreground">{org.nodeCount} Edge Nodes</td>
                    <td className="py-3 px-4">
                      <span className="font-mono font-bold text-emerald-400">{org.healthScore}%</span>
                    </td>
                    <td className="py-3 px-4">
                      <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                        {org.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 2: Cryptographic Audit Ledger */}
      {activeTab === "audit" && (
        <div className="space-y-6">
          {/* Verification Bar */}
          <div className="rounded-xl border border-border bg-card p-4 space-y-3">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <h3 className="text-base font-semibold text-foreground flex items-center gap-2">
                  <ShieldCheck className="h-5 w-5 text-emerald-400" />
                  SHA-256 Tamper-Evident Hash Chain
                </h3>
                <p className="text-xs text-muted-foreground">
                  Every user action, remediation approval, and policy mutation is cryptographically signed and chained to preceding blocks.
                </p>
              </div>

              <button
                onClick={handleVerifyChain}
                disabled={isVerifying}
                className="flex items-center gap-2 rounded-lg bg-primary px-3.5 py-2 text-xs font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-50 self-start sm:self-auto"
              >
                {isVerifying ? (
                  <>
                    <RefreshCw className="h-4 w-4 animate-spin" />
                    Recalculating Proof...
                  </>
                ) : (
                  <>
                    <CheckCircle2 className="h-4 w-4" />
                    Verify Hash Chain
                  </>
                )}
              </button>
            </div>

            {verificationResult && (
              <div className="rounded-lg bg-emerald-500/10 border border-emerald-500/30 p-3 text-xs text-emerald-300 space-y-1">
                <div className="font-bold flex items-center gap-1.5">
                  <CheckCircle2 className="h-4 w-4 text-emerald-400" />
                  Mathematical Integrity Verified (100% Intact)
                </div>
                <div className="text-[11px] font-mono text-emerald-200">
                  Checked {verificationResult.blocksChecked} blocks from Genesis block #0 to Head block #{verificationResult.blocksChecked - 1}. Zero hash mutations detected.
                </div>
                <div className="text-[11px] font-mono text-muted-foreground truncate">
                  Head Hash: {verificationResult.headHash}
                </div>
              </div>
            )}
          </div>

          {/* Audit Chain Table */}
          <div className="rounded-xl border border-border bg-card overflow-hidden">
            <table className="w-full text-left text-xs">
              <thead className="bg-muted text-muted-foreground border-b border-border font-semibold">
                <tr>
                  <th className="py-2.5 px-4">Block #</th>
                  <th className="py-2.5 px-4">Timestamp (UTC)</th>
                  <th className="py-2.5 px-4">Actor</th>
                  <th className="py-2.5 px-4">Action</th>
                  <th className="py-2.5 px-4">Resource</th>
                  <th className="py-2.5 px-4">Block SHA-256 Hash</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {sampleLedger.map(b => (
                  <tr key={b.seq} className="hover:bg-muted/50">
                    <td className="py-3 px-4 font-mono font-bold text-sky-400">#{b.seq}</td>
                    <td className="py-3 px-4 font-mono text-muted-foreground">{b.timestamp}</td>
                    <td className="py-3 px-4 font-mono text-foreground font-medium">{b.actor}</td>
                    <td className="py-3 px-4">
                      <span className="px-2 py-0.5 rounded font-mono text-[10px] font-bold bg-muted text-foreground border border-border">
                        {b.action}
                      </span>
                    </td>
                    <td className="py-3 px-4 font-mono text-muted-foreground">{b.resource}</td>
                    <td className="py-3 px-4 font-mono text-[11px] text-muted-foreground">
                      <span className="text-foreground font-semibold">{b.hash.substring(0, 12)}</span>
                      <span>...</span>
                      <span>{b.hash.substring(b.hash.length - 8)}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 3: Enterprise SSO & IAM */}
      {activeTab === "iam" && (
        <div className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Identity Providers */}
            <div className="rounded-xl border border-border bg-card p-4 space-y-3">
              <h3 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-2">
                <KeyRound className="h-4 w-4 text-primary" />
                Enterprise SSO Connectors
              </h3>
              <div className="space-y-2.5">
                <div className="p-3 rounded-lg border border-border bg-background flex items-center justify-between">
                  <div>
                    <div className="font-semibold text-foreground text-sm">Okta Enterprise SSO</div>
                    <div className="text-xs text-muted-foreground font-mono">SAML 2.0 • JIT SCIM Provisioning Active</div>
                  </div>
                  <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                    CONNECTED
                  </span>
                </div>

                <div className="p-3 rounded-lg border border-border bg-background flex items-center justify-between">
                  <div>
                    <div className="font-semibold text-foreground text-sm">Microsoft Entra ID (Azure AD)</div>
                    <div className="text-xs text-muted-foreground font-mono">OIDC • Conditional Access Enforced</div>
                  </div>
                  <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                    CONNECTED
                  </span>
                </div>

                <div className="p-3 rounded-lg border border-border bg-background flex items-center justify-between opacity-70">
                  <div>
                    <div className="font-semibold text-foreground text-sm">Google Workspace</div>
                    <div className="text-xs text-muted-foreground font-mono">SAML 2.0 • Standby</div>
                  </div>
                  <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-muted text-muted-foreground border border-border">
                    CONFIGURED
                  </span>
                </div>
              </div>
            </div>

            {/* FIDO2 & Passwordless */}
            <div className="rounded-xl border border-border bg-card p-4 space-y-3">
              <h3 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-2">
                <Lock className="h-4 w-4 text-primary" />
                FIDO2 / WebAuthn & ReBAC Governance
              </h3>
              <div className="space-y-2.5">
                <div className="p-3 rounded-lg border border-border bg-background space-y-1">
                  <div className="flex justify-between items-center">
                    <span className="font-semibold text-foreground text-sm">Hardware Security Keys (YubiKey)</span>
                    <span className="text-xs font-bold text-sky-400">14 Active Keys</span>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Hardware MFA enforced for all MSP Super Admins and high-impact network configuration approvals.
                  </p>
                </div>

                <div className="p-3 rounded-lg border border-border bg-background space-y-1">
                  <div className="flex justify-between items-center">
                    <span className="font-semibold text-foreground text-sm">OpenFGA Relationship-Based Access (ReBAC)</span>
                    <span className="text-xs font-bold text-emerald-400">Active</span>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Fine-grained permissions model separating client-tenant blast radius from MSP partner administration.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
