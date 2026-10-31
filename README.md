# Qnetward

Qnetward (formerly Net-Watcher) is an enterprise-grade, multi-tenant distributed network observability, infrastructure monitoring, and security scanning platform. It merges high-frequency edge mesh probing (PulseMesh), bidirectional gRPC ingestion, hardware SNMP/NetFlow collector drivers, host telemetry, 3-tier controlled security verification, and topology-aware AI root-cause analysis (RCA) with a bilingual Next.js web cockpit.

## Contributors

- MOPARA PAIR AYAT
- MONABBER EAMAN

## Overview

Qnetward operates across two planes (Control Plane and Data Plane):

- **Data Plane (Edge)**: Outbound-only mTLS edge Probe Agents running SWIM gossip and sub-millisecond UDP echo matrices, hardware collectors polling SNMPv1-v3 and NetFlow v5/IPFIX, and host OS telemetry agents.
- **Control Plane**: Ingestion Gateway (gRPC & MQTT), event bus with wildcard routing, multi-tenant PostgreSQL/TimescaleDB with Row-Level Security (RLS), Redis caching, SeaweedFS S3 object storage, and AI root-cause analysis.
- **Unified Cockpit**: Next.js 16 web interface with bilingual support (English & Bengali), dark/light themes, real-time PulseMesh heatmaps, and directed topology visualization with upstream alert storm suppression.

## Production Docker Compose Stack

Launch the entire 11-service distributed cluster with zero manual setup:

```bash
cp .env.example .env
docker compose up -d
```

Services started:
- `edge`: Nginx reverse proxy on ports `80` and `8080`
- `frontend`: Next.js web application on port `3000`
- `control-plane`: API server & gRPC Ingestion Gateway on ports `8080` and `9090`
- `probe-agent-1`: PulseMesh leader (HTTP: 8081, Gossip: 7946, UDP: 9000)
- `probe-agent-2`: PulseMesh peer (HTTP: 8082, Gossip: 7947, UDP: 9001)
- `probe-agent-3`: PulseMesh peer (HTTP: 8083, Gossip: 7948, UDP: 9002)
- `collector`: Hardware telemetry listener (SNMP Trap UDP 1162, NetFlow UDP 2055)
- `host-agent`: Host OS telemetry daemon
- `timescaledb`: PostgreSQL 16 / TimescaleDB with automated migrations
- `redis`: Redis 7 Alpine cache
- `objectstore`: SeaweedFS S3 storage on port `8333`

Open:
- Web Cockpit: `http://localhost:3000`
- PulseMesh API: `http://localhost:8081/api/topology`
- Prometheus Metrics: `http://localhost:8081/metrics`
- S3 SeaweedFS: `http://localhost:8333`

## Architecture

- `apps/web`
  primary Next.js frontend
- `cmd/collector`
  hardware collector daemon for SNMP/Redfish polling, traps, and NetFlow
- `cmd/host-agent`
  host OS telemetry daemon (CPU, RAM, Disks, SMART, Containers)
- `cmd/netwatcher`
  CLI entrypoint and backend API/probe server
- `cmd/probe-agent`
  standalone edge probe daemon powered by PulseMesh
- `drivers/profiles`
  declarative YAML device profiles (Cisco, MikroTik, Ubiquiti, Dell iDRAC, APC UPS)
- `drivers/sdk`
  hardware driver SDK, DeviceDriver interface, and profile loader
- `internal/cache`
  Redis cache integration
- `internal/dnslookup`
  DNS and reverse DNS lookup engine
- `internal/objectstore`
  S3-compatible object storage integration
- `internal/ping`
  ICMP ping implementation
- `internal/tcpping`
  TCP connect latency implementation
- `internal/report`
  result and summary model
- `internal/store`
  PostgreSQL/TimescaleDB persistence, batching, retention, RLS multi-tenancy, and alert rules
- `internal/portscan`
  TCP connect port scanning
- `pkg/airca`
  AI root-cause analysis (RCA), adaptive Z-score/EMA metric baselining, and natural-language query translator
- `pkg/appliance`
  Plug & Play hardware micro-probe firmware, Zero-Touch Provisioning (ZTP), and automatic cellular 5G out-of-band failover (OOBM)
- `pkg/audit`
  SHA-256 tamper-evident immutable cryptographic audit ledger and mathematical non-repudiation verification
- `pkg/bgp`
  Real-time BGP Monitoring Protocol (BMP) feeds, RPKI Route Origin Validation (ROV), prefix hijack & route leak detection
- `pkg/chatops`
  Bidirectional Slack, Microsoft Teams, and Discord slash command interface and webhook dispatcher
- `pkg/cloudconnect`
  Normalized multi-cloud flow log engine (AWS VPC Flow Logs, GCP VPC JSON, Azure NSG tuples) and hybrid interconnect tracking
- `pkg/collector`
  infrastructure hardware collector and NetFlow v5/IPFIX parser
- `pkg/connector`
  Edge agent lifecycle tracking, liveness reaper, tenant isolation, and task dispatching
- `pkg/copilot`
  Context-aware GenAI NetOps conversational pair-engineer for vendor ACL and executive incident post-mortem synthesis
- `pkg/ctwatcher`
  Certificate Transparency log streaming and visual homoglyph / brand typosquatting radar
- `pkg/dast`
  Automated defensive HTTP security auditor and out-of-band application security testing (OAST) correlation listener
- `pkg/digitaltwin`
  In-memory topological graph simulator for "What-If" failure blast radius, link saturation, and latency modeling
- `pkg/ebpf`
  Linux kernel eBPF socket tracing (RTT, handshake latency, resets) and Kubernetes pod/deployment metadata enricher
- `pkg/eventbus`
  Asynchronous decoupled pub/sub event bus with topic wildcard routing
- `pkg/hostagent`
  host OS telemetry collector (load, memory, disks, interfaces, containers)
- `pkg/ingestion`
  gRPC Ingestion Gateway server and resilient streaming agent client
- `pkg/mesh`
  PulseMesh distributed path-health matrix, SWIM gossip (`memberlist`), UDP echo prober, rolling percentiles, and embedded heatmap UI
- `pkg/msp`
  Multi-tenant MSP partner hierarchy (Partner -> Org -> Env), white-label customization, and co-managed access control
- `pkg/netguard`
  Anti-SSRF guard defending against cloud metadata, link-local, and unauthorized internal network probing
- `pkg/playbooks`
  Autonomous self-healing remediation engine, declarative triggers, operator approval gating, and automated rollback
- `pkg/predictive`
  Multivariate predictive failure modeling (SFP+ optical laser degradation and switch buffer pool exhaustion)
- `pkg/securityengine`
  3-tier controlled security policy gate, automated asset ownership verifier (DNS TXT / HTTP), CVE/KEV matcher, and audit logs
- `pkg/telemetry`
  OpenTelemetry tracing and metrics (see [Observability](#observability))
- `pkg/topology`
  directed dependency graph and Topology-Aware Alerting v2 with upstream suppression
- `proto/ingestion`
  Protobuf specifications and generated Go bindings for edge agent telemetry streaming
- `deploy`
  production `docker-compose` stack

Architecturally significant decisions (and the reasoning behind them) are
recorded in [`docs/adr`](docs/adr/README.md).

## Tech Stack

- Go
- Next.js
- TypeScript
- PostgreSQL
- TimescaleDB
- Redis
- S3-compatible object storage
- Gorilla WebSocket
- Tailwind CSS

## Features

### CLI

```bash
./netwatcher ping 8.8.8.8
./netwatcher tcpping example.com --port 443
./netwatcher portscan --ports 22,80,443 example.com
```

### Web UI

- Realtime results
- Live latency chart
- History loading
- Port scan mode
- Mobile-friendly layout
- Stop button for active WebSocket runs

### API

- `POST /api/ping`
- `POST /api/tcpping`
- `POST /api/portscan`
- `GET /api/history`
- `POST /api/export/history`
- `GET /healthz`
- `GET /ws`

## Quick Start

### Windows

```powershell
go mod tidy
go build -o netwatcher.exe .\cmd\netwatcher
.\netwatcher.exe serve --listen 127.0.0.1:8080 --db-dsn "postgres://netwatcher:change-me@127.0.0.1:5432/netwatcher?sslmode=disable" --timescale
```

Open:

```text
backend: http://127.0.0.1:8080
frontend: http://127.0.0.1:3000
```

### Linux

```bash
go mod tidy
go build -o netwatcher ./cmd/netwatcher
./netwatcher serve --listen 127.0.0.1:8080 --db-dsn "postgres://netwatcher:change-me@127.0.0.1:5432/netwatcher?sslmode=disable" --timescale
```

Open:

```text
backend: http://127.0.0.1:8080
frontend: http://127.0.0.1:3000
```

For ICMP on Linux, root or `CAP_NET_RAW` may be required:

```bash
sudo setcap cap_net_raw+ep ./netwatcher
```

## Example API Requests

### TCP Ping

```bash
curl -X POST http://127.0.0.1:8080/api/tcpping \
  -H "Content-Type: application/json" \
  -d '{"host":"example.com","port":443,"count":4,"interval_ms":1000,"timeout_ms":2000}'
```

### History

```bash
curl "http://127.0.0.1:8080/api/history?type=tcpping&host=example.com&port=443&limit=20"
```

### Export History

Download JSON:

```bash
curl -X POST http://127.0.0.1:8080/api/export/history \
  -H "Content-Type: application/json" \
  -d '{"type":"tcpping","host":"example.com","port":443,"limit":50,"format":"json"}'
```

Upload CSV to configured S3-compatible object storage:

```bash
curl -X POST http://127.0.0.1:8080/api/export/history \
  -H "Content-Type: application/json" \
  -d '{"type":"tcpping","host":"example.com","port":443,"limit":50,"format":"csv","destination":"s3"}'
```

## Production Stack

Recommended production stack:

- PostgreSQL + TimescaleDB
- Redis
- S3-compatible object storage

Environment variables:

```bash
NETWATCHER_DB_DSN=postgres://netwatcher:change-me@timescaledb:5432/netwatcher?sslmode=disable
NETWATCHER_TIMESCALE=true
NETWATCHER_REDIS_ADDR=redis:6379
NETWATCHER_S3_ENDPOINT=objectstore:8333
NETWATCHER_S3_ACCESS_KEY=netwatcher
NETWATCHER_S3_SECRET_KEY=change-me
NETWATCHER_S3_BUCKET=netwatcher-artifacts
NETWATCHER_S3_SSL=false
NETWATCHER_DEMO_ENABLED=false
NETWATCHER_OTEL_ENDPOINT=otel-collector:4318
```

Start the production stack:

```bash
cp .env.example .env
docker compose -f deploy/docker-compose.production.yml --env-file .env up --build
```

Health check response now includes dependency status for:

- database
- redis
- object storage

## Observability

NetWatcher exports OpenTelemetry traces and metrics over OTLP/HTTP when an
endpoint is configured. It is off by default and never blocks startup or
requests if the collector is unreachable.

```bash
NETWATCHER_OTEL_ENDPOINT=localhost:4318      # or OTEL_EXPORTER_OTLP_ENDPOINT
NETWATCHER_OTEL_SERVICE_VERSION=1.2.3        # optional, reported as service.version
```

A bare `host:port` (or an explicit `http://` scheme) connects without TLS;
use `https://` for a collector that requires it. When configured:

- Every HTTP request gets a trace span (`METHOD /path`), via
  [`otelhttp`](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp).
- Every database query gets a span and the connection pool exports gauges
  (open/idle/in-use connections), via
  [`otelsql`](https://pkg.go.dev/github.com/XSAM/otelsql).
- `netwatcher.probes.total` (counter) and `netwatcher.probe.duration`
  (histogram), labeled by `protocol` and `success`, cover ping/tcpping
  probes.
- `netwatcher.alerts.evaluated` and `netwatcher.alerts.triggered`
  (counters), labeled by `protocol`, cover alert rule evaluation.

## Contributor Demo Access

For temporary contributor access, enable demo mode and ask contributors to use the login page.

Demo mode:

- adds a `Continue as Demo` button on `/login`
- signs contributors into a shared temporary demo account
- allows the full dashboard and tool routes
- uses the real application data and behavior of that demo account

## Development

Build:

```bash
go build -o netwatcher ./cmd/netwatcher
```

Run internal tests:

```bash
go test ./internal/...
```

Run vet:

```bash
go vet ./...
```

Run the HTTP smoke test against a running local stack:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/smoke-test.ps1 -BaseUrl http://127.0.0.1:8080
```

GitHub Actions now runs:

- `go test ./...`
- `go vet ./...`
- `go test -race ./cmd/netwatcher ./internal/store ./internal/portscan ./internal/dnslookup`
- frontend `pnpm lint`
- frontend `pnpm build`
- frontend Playwright E2E
- Docker compose smoke checks through `scripts/smoke-test.ps1`

## Notes

- This project is built for self-hosted use.
- If you expose it publicly, put a trusted reverse proxy, tunnel, or access control layer in front of it.
- Runtime artifacts such as `*.log` and `*.exe` are ignored by `.gitignore`.
- WebSocket is the primary realtime path. HTTP is used as a fallback path.
- Production mode is designed around PostgreSQL/TimescaleDB + Redis + S3-compatible storage.

## Current Scope

Implemented now:
- ICMP ping
- TCP ping
- TCP port scan
- DNS toolkit & reverse lookup
- Alerts & multi-tenant evaluation engine
- Anti-SSRF guard (`pkg/netguard`)
- PulseMesh distributed matrix (`pkg/mesh`, `netwatcher mesh`, `cmd/probe-agent`)
- PostgreSQL Row-Level Security multi-tenancy
- OpenTelemetry tracing and metrics
- Realtime Web UI & WebSocket streaming
- History storage & exports (JSON, CSV, S3)
- Stop/cancel for active live runs

Planned next:
- Multi-target periodic scheduling daemon
- HTTP/HTTPS synthetic monitor
- Hardware metrics collector (SNMP/gNMI/Redfish)
- Next.js mesh heatmap integration
