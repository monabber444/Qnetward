# 14. Full Production Docker Compose Stack

## Status

Accepted

## Context

With all platform layers built across Phases 0–4 (Control plane, Polyglot storage, gRPC ingestion gateway, PulseMesh distributed probe matrix, hardware collectors, host agents, 3-tier security policy verification, AI root-cause analysis, and the Next.js unified frontend), operations teams require a single unified Docker Compose topology to deploy, orchestrate, and run the entire platform.

Key deployment requirements:
1. **Zero-Configuration Initialization**: Automated database schema migrations (`golang-migrate`), Redis cache provisioning, and S3 SeaweedFS bucket seeding on initial launch.
2. **Polyglot Service Orchestration**: Starting the TimescaleDB persistence layer and Redis cache first with deterministic healthchecks (`pg_isready`, `redis-cli ping`) before launching the Control Plane.
3. **Cluster Ingestion & Probing Mesh**: Launching the gRPC Ingestion Gateway (`0.0.0.0:9090`) alongside a 3-node PulseMesh edge cluster with automatic gossip peering (`node-2` and `node-3` joining `node-1:7946`).
4. **Hardware & Host Observability**: Launching the Hardware Collector listening for SNMP traps (UDP 1162) and NetFlow packets (UDP 2055), and the Host OS telemetry daemon.
5. **Ingress Routing & SSR/SPA Serving**: Reverse proxying through Nginx with dynamic upstream routing for REST APIs, WebSocket streams, PulseMesh metrics, and Next.js frontend pages.

## Decision

1. **Unified Multi-Binary Go Container Image (`Dockerfile`)**:
   - Single multi-stage build creating all 4 system binaries:
     - `/app/netwatcher` (Control Plane, Ingestion Gateway, REST, WebSockets)
     - `/app/probe-agent` (PulseMesh SWIM Gossip, UDP active probers, Prometheus metrics)
     - `/app/collector` (SNMPv1-v3 trap receiver, NetFlow/IPFIX listener, vendor driver SDK)
     - `/app/host-agent` (Host OS telemetry daemon)
   - Packages vendor device driver profiles (`drivers/profiles/*.yaml`) inside the runtime image.
   - Installs `ca-certificates`, `iproute2`, and `iputils-ping` for network diagnostics.

2. **Unified Compose Topologies (`docker-compose.yml` & `deploy/docker-compose.production.yml`)**:
   - **`timescaledb`**: TimescaleDB PG16 with volume persistence (`timescaledb-data`). Healthchecked via `pg_isready`.
   - **`redis`**: Redis 7 Alpine with append-only persistence. Healthchecked via `redis-cli ping`.
   - **`objectstore`**: SeaweedFS S3 gateway (`chrislusf/seaweedfs:latest`) with static S3 keys.
   - **`control-plane`**: Launches `netwatcher serve --ingestion-listen 0.0.0.0:9090` with `NET_RAW` capability. Healthchecked via `/healthz`.
   - **`probe-agent-1`, `probe-agent-2`, `probe-agent-3`**: 3-node distributed mesh automatically joining on port 7946 and establishing bidirectional UDP probing matrices.
   - **`collector`**: Listening on UDP 1162 (SNMP/Syslog) and UDP 2055 (NetFlow/IPFIX).
   - **`host-agent`**: Inspects `/host/proc` and `/host/sys` metrics.
   - **`frontend`**: Next.js standalone container serving port 3000.
   - **`edge`**: Nginx 1.27 Alpine routing `/api/` and `/ws` to the control plane, `/mesh-api/` to `probe-agent-1`, and root traffic to Next.js.

3. **Environment Configuration (`.env.example`)**:
   - Provides default credentials, S3 keys, and allowed CORS origins for local and production deployments.

## Consequences

- The entire 11-service distributed observability platform can be launched with a single command (`docker compose up -d`).
- Deterministic healthcheck dependencies eliminate race conditions on startup.
- Edge agents immediately establish gossip mesh connectivity and stream telemetry to the Ingestion Gateway.
