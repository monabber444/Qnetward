# 17. Continuous Attack Surface Management (CASM) & Sandboxed DAST

## Status

Accepted

## Context

Enterprise networks and modern distributed infrastructures face continuous perimeter threats from external attack surfaces:
1. **Brand Impersonation & Lookalike Phishing Domains**: Attackers frequently register typosquatting or visually confusable homoglyph domains (e.g. `qnetvvard.com`, `qnetward.secure-portal.net`) and provision TLS certificates to launch credential harvesting campaigns against enterprise employees and clients.
2. **Defensive Posture & Web Security Misconfigurations**: Public-facing APIs, ingress controllers, and management portals often suffer from missing defense-in-depth headers (CSP, HSTS, X-Frame-Options, X-Content-Type-Options), permissive cross-origin resource sharing (CORS wildcard with credentials), insecure session cookies (missing Secure or HttpOnly flags), and banner version leakage (Server, X-Powered-By).
3. **Blind Vulnerabilities Requiring Out-of-Band (OAST) Correlation**: Testing for blind SSRF or asynchronous injections requires dedicated out-of-band correlation listeners capable of capturing asynchronous HTTP and DNS interactions tied to unique audit tokens.

## Decision

1. **Certificate Transparency (CT) Stream & Typosquatting Watcher (`pkg/ctwatcher`)**:
   - **`BrandDetector` (`pkg/ctwatcher/detector.go`)**: Evaluates newly issued TLS certificates against protected brand domains using:
     - Levenshtein edit distance with normalized similarity scoring.
     - Homoglyph and visual confusable character normalization (`vv` $\to$ `w`, `rn` $\to$ `m`, `0` $\to$ `o`, `1` $\to$ `l`).
     - Subdomain and hyphenated impersonation checks (`brand.evil-site.com`, `brand-portal.org`).
   - **`StreamConsumer` (`pkg/ctwatcher/stream.go`)**: Ingests certificate transparency feeds, processes SANs in parallel, and dispatches `security.typosquat` events over the distributed `eventbus.Bus`.

2. **Automated Defensive DAST & OAST Correlation Engine (`pkg/dast`)**:
   - **`Auditor` (`pkg/dast/auditor.go`)**: Inspects HTTP/HTTPS endpoints for defense-in-depth header compliance, wildcard CORS credentials exposure, cookie flag hygiene, and banner information disclosure. Assigns standardized security grades (`A+` to `F`).
   - **`OASTManager` (`pkg/dast/oast_listener.go`)**: Generates cryptographically secure correlation tokens, associates them with audit sessions, and captures inbound asynchronous HTTP/DNS callback interactions, broadcasting `oast.interaction` events across the platform.

3. **Next.js CASM Cockpit (`/casm`)**:
   - Added bilingual support in `src/lib/i18n.ts` and sidebar entry in `src/components/app/sidebar.tsx`.
   - Built interactive CASM workbench (`src/features/casm/casm-workspace.tsx` & `src/app/(protected)/casm/page.tsx`):
     - Real-time CT Log Typosquat Radar with match scoring, one-click takedown requests, and status tracking.
     - Interactive DAST scanner console with instant posture evaluation and remediation guidance.
     - OAST Hub displaying captured out-of-band callbacks with token correlation.

## Consequences

- Provides proactive threat intelligence against domain impersonation before phishing campaigns launch.
- Provides immediate automated visibility into endpoint security posture and header hardening across all enterprise ingress points.
- Enables safe, sandboxed detection of out-of-band blind vulnerabilities without offensive exploitation payloads.
