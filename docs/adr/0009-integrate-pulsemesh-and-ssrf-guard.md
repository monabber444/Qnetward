# 9. Integrate PulseMesh and Anti-SSRF Guard

## Status

Accepted

## Context

NetWatcher previously ran single-location network diagnostics (ICMP ping, TCP ping, port scanning, and DNS lookups) from a single host. To evolve towards the distributed **Qnetward** architecture, two critical capabilities were needed:

1. **Distributed Mesh Probing (PulseMesh)**: An $N \times N$ path-health and latency matrix capability across nodes using the SWIM gossip protocol (`hashicorp/memberlist`), active UDP echo probing, and thread-safe circular ring buffers computing rolling percentiles ($p50, p95, p99$), RFC 3550 jitter, and directional packet loss.
2. **SSRF Guard**: Public-facing or multi-tenant network probing endpoints (`/api/ping`, `/api/tcpping`, `/api/portscan`, `/ws`) must be protected against Server-Side Request Forgery (SSRF) targeting cloud metadata services (e.g., `169.254.169.254`, `fd00:ec2::254`), loopback interfaces, link-local addresses, CGNAT, and unauthorized internal subnets.

## Decision

1. **Integrate PulseMesh as `pkg/mesh` and `cmd/probe-agent`**:
   - `pkg/mesh/cluster`: Memberlist SWIM gossip wrapper with binary metadata exchange (probe/HTTP ports).
   - `pkg/mesh/prober`: High-precision UDP probe engine and wire protocol (`0x504D` magic, sequence ID, nanosecond monotonic timestamps).
   - `pkg/mesh/metrics`: Lock-free/mutex circular ring buffer computing rolling $p50, p95, p99$ RTT, jitter, and packet loss percentage; Prometheus exporter.
   - `pkg/mesh/server`: REST API (`/api/topology`) and live WebSocket hub streaming matrix snapshots at 1Hz.
   - `pkg/mesh/web`: Embedded React + Vite + Tailwind $N \times N$ interactive heatmap SPA (`web.DistFS()`).
   - Add `netwatcher mesh [flags]` subcommand in `cmd/netwatcher/` to run the daemon directly from the primary binary.
   - Add standalone `cmd/probe-agent/` binary for lightweight edge deployments.

2. **Implement Anti-SSRF Guard in `pkg/netguard`**:
   - `pkg/netguard/ssrf.go` provides `ValidateTarget` and `ValidateIP`.
   - Blocks loopback (`127.0.0.0/8`, `::1/128`), cloud metadata (`169.254.169.254`, `fd00:ec2::254`), link-local (`169.254.0.0/16`, `fe80::/10`), multicast, and carrier-grade NAT.
   - Configurable `AllowPrivate` and `AllowLoopback` options with CLI flags (`--allow-private-ips`, `--allow-loopback`) to accommodate self-hosted intranet use cases while defaulting to strict defense-in-depth.
   - All HTTP and WebSocket probe handlers enforce `netguard.ValidateTarget` before initiating network connections.

## Consequences

- NetWatcher and PulseMesh are now combined into a single unified repository and deployable binary.
- All probe routes protect against internal service discovery and cloud metadata leaks.
- Zero breaking changes to existing NetWatcher APIs or frontend assets.
