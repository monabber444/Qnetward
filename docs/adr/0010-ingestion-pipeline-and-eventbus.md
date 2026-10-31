# 10. Ingestion Pipeline, Edge Agent Streaming, and Event Bus

## Status

Accepted

## Context

As Qnetward transitions from single-host probing to a distributed multi-tenant observability platform, edge probe daemons (`cmd/probe-agent`) deployed across multi-cloud environments, customer VPCs, and on-premises LANs must securely stream real-time telemetry back to the centralized control plane without exposing inbound ports.

Furthermore, the centralized control plane requires an asynchronous decoupled event bus architecture to fan out incoming telemetry streams (raw ICMP/TCP probes, PulseMesh matrices, agent heartbeats, and alert evaluations) to time-series storage, alert evaluation engines, and real-time frontend dashboards without blocking network I/O.

## Decision

1. **Protobuf & Bidirectional gRPC Streaming Protocol (`proto/ingestion/ingestion.proto`)**:
   - Implemented `IngestionService` with bidirectional streaming RPC `StreamChannel`.
   - Outbound-only edge agent architecture: Edge agents initiate an outbound TLS/mTLS gRPC connection to the Ingestion Gateway, maintaining a persistent long-lived stream.
   - Message payloads support:
     - `AgentRegister`: Handshake transmitting hostname, tenant ID, OS, architecture, capabilities, and key-value tags.
     - `AgentHeartbeat`: Periodic health and utilization metrics (CPU percent, memory usage, active probe count).
     - `ProbeTelemetry`: Micro-telemetry for individual synthetic probes (sequence, RTT ms, error strings).
     - `MeshMatrixTelemetry`: Aggregated PulseMesh $N \times N$ rolling path-health percentiles ($p50, p95, p99$), jitter, and loss rates.
     - `TaskExecutionResult`: Asynchronous completion responses for commands dispatched by the control plane.
   - Control plane commands support:
     - `RegistrationAck`: Acknowledgment returning assigned heartbeat intervals and gossip seed peers.
     - `ScheduleProbeTask`: Ad-hoc and scheduled probe tasks routed dynamically to eligible edge agents.
     - `CancelProbeTask`: Cancellation of active long-running probe loops.
     - `MeshConfigUpdate`: Dynamic mesh topology parameter reconfiguration.

2. **Decoupled Event Bus (`pkg/eventbus`)**:
   - Defined `eventbus.Bus` interface supporting subject-based pub/sub, pattern/wildcard matching (`*` and `>`), and concurrent delivery.
   - Built an in-memory lock-protected implementation (`MemoryBus`) for single-process deployments and unit testing, designed to swap with NATS JetStream or Kafka for distributed cluster scaling.
   - Standardized topic taxonomy: `telemetry.probe`, `telemetry.mesh`, `agent.heartbeat`, `agent.register`, `task.command`, and `alert.evaluated`.

3. **Ingestion Gateway (`pkg/ingestion`) & Connector Manager (`pkg/connector`)**:
   - `pkg/ingestion/server.go`: Thread-safe gRPC server managing bidirectional client streams, dispatching commands to specific agent channels, and publishing incoming telemetry directly to the event bus.
   - `pkg/ingestion/client.go`: Resilient edge agent gRPC client with auto-reconnect, exponential backoff, send buffering, and typed callback hooks.
   - `pkg/connector/manager.go`: Agent lifecycle tracking, periodic heartbeat timeout reaping (marking stale agents degraded/offline), multi-tenant capability filtering, and task routing.

4. **Integration with Daemon & Probe Agent**:
   - `cmd/netwatcher`: Added `--ingestion-listen` flag to host the gRPC Ingestion Gateway alongside HTTP/REST/WebSocket APIs, with `/api/agents` endpoint exposing connected cluster agents.
   - `cmd/probe-agent`: Integrated `--gateway-addr`, `--tenant-id`, and `--agent-id` flags to connect edge agents to the gateway, automatically streaming PulseMesh matrix snapshots and system health.

## Consequences

- Edge agents require zero inbound open firewall ports, operating entirely over outbound gRPC connections.
- Control plane components communicate asynchronously via event topics, decoupling ingestion throughput from database write latency.
- Full multi-tenant isolation is enforced at the wire level through `tenant_id` propagation on all messages and agent registry lookups.
- Verified through multi-node integration tests exercising 3 concurrent edge agents streaming telemetry and executing control tasks.
