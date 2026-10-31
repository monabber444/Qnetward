# 12. Three-Tier Controlled Security Engine and AI Root-Cause Analysis

## Status

Accepted

## Context

As Qnetward expands into security vulnerability observability and operational telemetry, two architectural challenges arise:

1. **Active Scanning Liability & SSRF Prevention**:
   - Unauthorized active scanning or port probes against external targets create legal liability, abuse complaints, and security risks.
   - Attackers could attempt Server-Side Request Forgery (SSRF) targeting cloud metadata services (`169.254.169.254`, `fd00:ec2::254`), loopback addresses (`127.0.0.1`), or unauthorized internal subnets.
   - The platform requires strict asset ownership verification (proving control of the target before unlocking active scans) and an immutable audit trail.

2. **Alert Storms & Incident Root Cause**:
   - During infrastructure anomalies, hundreds of telemetry alarms fire simultaneously across cascading network layers.
   - Operators need statistical metric baselining (Z-score anomaly detection) and automated Root-Cause Analysis (RCA) that correlates alerts with the physical topology to isolate the lead failure, calculate blast radius, and formulate suggested mitigations.

## Decision

1. **Asset Ownership Verification Engine (`pkg/securityengine/verification.go`)**:
   - Generates cryptographically secure verification tokens (`qnetward-verify=<hex>`).
   - Supports automated verification via:
     - **DNS TXT record**: `qnetward-verify=<token>` at target domain.
     - **HTTP well-known token**: `http://<domain>/.well-known/qnetward-verify.txt`.
   - Re-verifies periodically and supports status transitions (`pending`, `verified`, `revoked`).

2. **3-Tier Controlled Security Policy Gate (`pkg/securityengine/policy.go`)**:
   - **Tier 0 (Passive)**: Permitted on any target without prior verification (Certificate Transparency logs, passive DNS, RDAP/WHOIS, SPF/DKIM/DMARC mail authentication auditing).
   - **Tier 1 (Safe Active)**: Enforces `403 Ownership Verification Required` unless the target has an active verified ownership record. Unlocks port scan, service banner detection, TLS cipher grading, and safe CVE/KEV template matching (`pkg/securityengine/cve.go`).
   - **Tier 2 (Deep Authorized)**: Requires Tier 1 verification + `VerifiedBusiness` plan flag + active scheduled scope window + rate limit token bucket + emergency kill switch. Unlocks DAST crawling, OAST callbacks, and attack-path analysis.
   - **SSRF Defense-in-Depth**: Every scan request must pass `netguard.ValidateTarget`, strictly rejecting cloud metadata, loopback, link-local, and unauthorized IP ranges.
   - **Immutable Audit Logging (`pkg/securityengine/audit.go`)**: Records every authorization attempt (`ActorID`, `TenantID`, `Target`, `Tier`, `Action`, `Outcome`, `Reason`, `Timestamp`).

3. **AI Root-Cause Analysis & Natural Language Query (`pkg/airca`)**:
   - **`AdaptiveBaseline` (`baseline.go`)**: Welford-style rolling Exponential Moving Average (EMA) and variance calculation. Computes dynamic Z-scores ($Z = \frac{|x - \mu|}{\sigma}$) to identify true statistical metric deviations ($Z \ge 3.0$) with initial warmup calibration.
   - **`RCAEngine` (`rca.go`)**: Clusters simultaneous anomaly events across nodes, intersects with the network topology graph (`pkg/topology`), isolates the primary suspect node, computes blast radius, and produces root-cause summaries with actionable mitigations.
   - **`NLQueryTranslator` (`nlquery.go`)**: Translates operator natural-language search queries (e.g. `"packet loss > 5% in us-east last 2 hours"`, `"critical alerts on core-router"`) into structured query filters and SQL `WHERE` clauses for TimescaleDB/PostgreSQL.

## Consequences

- Active scanning is mathematically impossible against unverified external infrastructure or cloud metadata endpoints.
- All authorization decisions are permanently logged in an immutable audit trail.
- Operators gain automated root-cause detection during alert storms, cutting MTTR (Mean Time to Resolution).
