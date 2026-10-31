# Architecture Decision Records

This directory records the architecturally significant decisions made in
this codebase: the ones that are expensive to reverse, or that a future
contributor needs the *reasoning* for, not just the resulting code.

Format: one Markdown file per decision, numbered sequentially, using the
lightweight template in [0001](0001-record-architecture-decisions.md)
(context / decision / consequences). A decision is superseded by writing a
new ADR that says so and updating the old one's status — never by editing
history in place, since the point is to keep the trail of *why* intact.

## Index

| # | Title | Status |
|---|-------|--------|
| [0001](0001-record-architecture-decisions.md) | Record architecture decisions | Accepted |
| [0002](0002-split-main-into-domain-files.md) | Split cmd/netwatcher/main.go into domain files | Accepted |
| [0003](0003-schema-migrations-with-golang-migrate.md) | Manage schema changes with golang-migrate instead of inline DDL | Accepted |
| [0004](0004-multi-tenant-model-with-row-level-security.md) | Multi-tenant data model enforced with PostgreSQL Row-Level Security | Accepted |
| [0005](0005-defer-rls-on-results-table.md) | Defer RLS enforcement on the results table | Accepted |
| [0006](0006-opentelemetry-disabled-by-default.md) | OpenTelemetry for tracing and metrics, disabled by default | Accepted |
| [0007](0007-replace-minio-with-seaweedfs.md) | Replace MinIO with SeaweedFS in the production stack | Accepted |
| [0008](0008-adopt-apps-and-pkg-top-level-layout.md) | Adopt `apps/` and `pkg/` top-level directories | Accepted |
| [0009](0009-integrate-pulsemesh-and-ssrf-guard.md) | Integrate PulseMesh and Anti-SSRF Guard | Accepted |
| [0010](0010-ingestion-pipeline-and-eventbus.md) | Ingestion Pipeline, Edge Agent Streaming, and Event Bus | Accepted |
| [0011](0011-hardware-collector-driver-sdk-and-topology-alerting.md) | Hardware Collector, Vendor Driver SDK, and Topology-Aware Alerting | Accepted |
| [0012](0012-three-tier-security-engine-and-ai-rca.md) | Three-Tier Controlled Security Engine and AI Root-Cause Analysis | Accepted |
| [0013](0013-unified-frontend-mesh-workbench-and-i18n.md) | Unified Next.js Frontend, PulseMesh Workbench, and Bilingual i18n | Accepted |
| [0014](0014-production-docker-compose-stack.md) | Full Production Docker Compose Stack | Accepted |
| [0015](0015-cloud-native-connectors-and-ebpf-probing.md) | Cloud-Native Multi-Cloud Connectors and Linux eBPF Kernel Probing | Accepted |
| [0016](0016-autonomous-playbooks-and-closed-loop-chatops.md) | Autonomous Self-Healing Playbooks, Closed-Loop Remediation, and ChatOps | Accepted |
| [0017](0017-continuous-attack-surface-management-and-sandboxed-dast.md) | Continuous Attack Surface Management (CASM) and Sandboxed DAST | Accepted |
| [0018](0018-global-bgp-intelligence-and-network-digital-twin.md) | Global BGP Routing Intelligence and Network Digital Twin | Accepted |
| [0019](0019-multitenant-msp-hierarchy-and-cryptographic-audit-ledger.md) | Multi-Tenant MSP Hierarchy and Cryptographic Audit Ledger | Accepted |
| [0020](0020-micro-probe-appliances-ztp-and-5g-out-of-band.md) | Micro-Probe Appliances, Zero-Touch Provisioning, and 5G Out-of-Band | Accepted |
| [0021](0021-genai-netops-copilot-and-predictive-failure-analytics.md) | GenAI NetOps Copilot and Predictive Failure Analytics | Accepted |
| [0022](0022-hyperscale-partitioning-columnar-olap-and-hierarchical-gossip.md) | Hyperscale Partitioning, Columnar OLAP, and Hierarchical Gossip (H-SWIM) | Accepted |
| [0023](0023-distributed-consensus-locks-and-kubernetes-operator.md) | Distributed Consensus Locks and Kubernetes GitOps Operator | Accepted |
