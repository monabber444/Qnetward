# 19. Multi-Tenant MSP Hierarchy and Cryptographic Audit Ledger

## Status

Accepted

## Context

Scaling Qnetward to enterprise service providers and regulated environments requires:
1. **Multi-Tenant MSP Hierarchy**: Managed Service Providers (MSPs) manage hundreds of client enterprises across diverse regions and environments (Production, Staging, DR). They require strict tenant isolation, co-managed access control (differentiating MSP super admins, partner admins, client admins, and read-only auditors), and white-label branding.
2. **Cryptographic Non-Repudiation & Compliance**: For SOC 2 Type II, ISO 27001, HIPAA, and PCI-DSS compliance audits, traditional database logs are insufficient because an administrator with database write access could alter or delete historical action records. An immutable, tamper-evident cryptographic hash chain is required.
3. **Enterprise Zero-Trust IAM**: Identity federation via SAML 2.0 / OIDC (Okta, Microsoft Entra ID) and passwordless hardware security keys (FIDO2 / WebAuthn / YubiKey).

## Decision

1. **MSP Multi-Tenant Engine (`pkg/msp`)**:
   - **`Manager` (`pkg/msp/manager.go`)**: Enforces multi-tier tenancy (`Partner` $\to$ `Organization` $\to$ `Environment`).
   - **White-Label Customization**: Partners configure custom domains (`net.partner.io`), brand names, primary color schemes, and PDF executive report headers.
   - **Co-Managed Scoped Authorization**:
     - `RoleMSPSuperAdmin`: Global fleet control.
     - `RolePartnerAdmin`: Full management scoped strictly to client organizations within their partner tenant.
     - `RoleClientAdmin`: Scoped strictly to their enterprise organization.
     - `RoleReadOnlyOperator`: Scoped read-only access for compliance auditors.

2. **Tamper-Evident Cryptographic Audit Ledger (`pkg/audit`)**:
   - **`ImmutableLedger` (`pkg/audit/ledger.go`)**: Each event computes a deterministic SHA-256 hash chaining sequence, UTC timestamp, tenant ID, actor ID, action type, resource ID, payload hash, and preceding block hash:
     $$\text{Hash}_n = \text{SHA256}(\text{Seq} \parallel \text{Timestamp} \parallel \text{Tenant} \parallel \text{Actor} \parallel \text{Action} \parallel \text{Resource} \parallel \text{PayloadHash} \parallel \text{PrevHash})$$
   - **Automated Chain Verification**: `VerifyChain()` scans the ledger sequentially from the Genesis block to the Head block, immediately flagging broken links or modified payloads (`TamperedSequence`).
   - Broadcasts audit events over `eventbus.Bus` on topic `audit.event`.

3. **Unified Next.js Cockpit (`/msp-portal`)**:
   - Client organization management table with real-time SLA metrics (99.999%), active edge node counters, and health scores.
   - Cryptographic Audit Ledger viewer with live **"Verify Hash Chain"** button displaying mathematical proof of non-repudiation.
   - Enterprise SSO & FIDO2 passkey governance cards.

## Consequences

- Guarantees strict multi-tenant data isolation and role delegation for MSP partners and enterprise customers.
- Delivers mathematically provable, tamper-evident audit trails ready for external SOC 2 Type II and ISO 27001 compliance auditors.
