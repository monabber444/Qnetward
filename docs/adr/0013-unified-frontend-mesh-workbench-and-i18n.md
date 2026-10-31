# 13. Unified Next.js Frontend, PulseMesh Workbench, and Bilingual i18n

## Status

Accepted

## Context

Phase 0 through Phase 3 established the multi-tenant Go backend architecture: PulseMesh distributed edge gossip, bidirectional gRPC ingestion, hardware SNMP/NetFlow collector drivers, 3-tier security policy verification, and topology-aware AI RCA.

To deliver these capabilities to operators, the web interface (`apps/web`) required comprehensive modernization:
1. **PulseMesh Operational Matrix**: Operators need an interactive $N \times N$ matrix visualizing sub-millisecond rolling latency percentiles ($p_{50}, p_{95}, p_{99}$), RFC 3550 jitter, and asymmetric path deltas ($A \rightarrow B$ vs $B \rightarrow A$).
2. **Topology & Upstream Alert Suppression**: Visualizing the directed graph hierarchy and clearly distinguishing root-cause incidents from suppressed downstream victim alerts.
3. **Security Gate & Asset Verification Workbench**: Visual controls for the 3-Tier Security Policy Gate, domain ownership token generation (DNS TXT / HTTP token), Anti-SSRF socket dialer enforcement, and FIRST EPSS / CISA KEV catalog search.
4. **AI Command Bar & Telemetry Query**: Natural language telemetry query bar translating plain text into structured metric filters and SQL expressions.
5. **Internationalization (i18n)**: Bilingual interface supporting English and Bengali (`en` / `bn`) for global operations teams.

## Decision

1. **PulseMesh Distributed Matrix Workspace (`src/features/mesh/mesh-workspace.tsx` & `/mesh`)**:
   - Interactive matrix table with real-time polling to `http://127.0.0.1:8081/api/topology`.
   - Directional Link Inspector displaying rolling percentiles and RFC 3550 jitter.
   - Forward vs. Reverse path asymmetry calculation ($|Fwd - Rev|$ delta).

2. **Network Topology & Hardware Workspace (`src/features/topology/topology-workspace.tsx` & `/topology`)**:
   - Directed Acyclic Graph (DAG) hierarchy from root core routers down through distribution switches to host servers.
   - Visual alert suppression indicators demonstrating how downstream victim alerts are collapsed under upstream root causes.
   - Hardware Driver Telemetry panel for Cisco SFP+ optical DOM, MikroTik, Ubiquiti UniFi, Dell iDRAC Redfish, and APC Smart-UPS.

3. **Three-Tier Security & Verification Workbench (`src/features/security/security-workspace.tsx` & `/security`)**:
   - Visual 3-tier policy status cards (Tier 0 OSINT allowed, Tier 1 gated by 403 Forbidden until verified, Tier 2 deep authorized).
   - Domain verification wizard with DNS TXT record (`_qnetward-verify.<domain>`) and HTTP token verification.
   - Anti-SSRF dialer shield indicator.
   - Vulnerability catalog matcher with FIRST EPSS and CISA KEV status.

4. **Network Tools Workbench Additions**:
   - Traceroute / MTR Workbench (`/traceroute`) with hop-by-hop latency and Autonomous System (AS) discovery.
   - TLS & HTTP/3 Synthetics Inspector (`/tls-inspector`) evaluating TLS 1.3 ciphers, ALPN negotiation, and QUIC stream performance.

5. **AI Telemetry Command Bar (`src/features/overview/overview-ai-command-bar.tsx`)**:
   - Embedded in Overview workspace to parse natural language queries into structured parameters and compiled SQL.

6. **Bilingual i18n & Theme Store (`src/lib/i18n.ts` & `src/store/ui-store.ts`)**:
   - Complete dictionary mappings for English (`en`) and Bengali (`bn`).
   - One-click language switcher (`[EN | বাংলা]`) and theme toggle in Topbar.

## Consequences

- Operators have a unified, responsive SaaS cockpit covering all 5 architectural planes.
- Multi-region teams can operate the platform in English or Bengali.
- Full Next.js 16 build compiles statically with zero errors and Turbopack optimization.
