# 21. GenAI NetOps Copilot and Predictive Failure Analytics

## Status

Accepted

## Context

Complex enterprise network operations face two chronic bottlenecks:
1. **Cognitive Overload in Emergency Incident Response**: When a security threat (e.g. zero-day CVE-2024-3400) or high-severity outage strikes, human operators waste crucial minutes drafting syntax-precise vendor ACLs (Cisco IOS-XE, Juniper JunOS) and assembling post-incident executive post-mortems from fragmented logs, telemetry metrics, and alert queues.
2. **Reactive Rather Than Predictive Maintenance**: Conventional network observability triggers alarms *after* an optical transceiver dies or a switch buffer pool exhausts and begins dropping customer packets. Multivariate optical telemetry (Tx/Rx power drift, laser bias current inflation) contains early degradation signatures that can predict hardware death 24-48 hours before an actual link cut occurs.

## Decision

1. **Context-Aware NetOps Copilot Engine (`pkg/copilot`)**:
   - **`CopilotEngine` (`pkg/copilot/engine.go`)**: Synthesizes live network state, topology graphs, and telemetry into conversational diagnostics.
   - **Artifact Synthesis**:
     - **Vendor Extended ACLs**: Automatically drafts hardened Cisco IOS-XE extended access lists or Juniper firewall filters to mitigate threat signatures.
     - **Executive Post-Mortems**: Generates structured Markdown post-mortem documents summarizing timelines, root-causes, affected VLANs, blast radius, and preventative action items.
     - **Diagnostic Queries**: Natural language analysis of packet loss spikes, BGP route flapping, and jitter anomalies.
   - Publishes queries to `eventbus.Bus` on topic `copilot.query`.

2. **Multivariate Predictive Failure Engine (`pkg/predictive`)**:
   - **`PredictiveEngine` (`pkg/predictive/model.go`)**:
     - **Optical DOM Laser Drift Modeling**: Tracks baseline vs current Tx optical power and laser bias current. If Tx power decays $\ge 2.0\text{ dBm}$ while bias current spikes $\ge 25\%$, the engine forecasts imminent transceiver laser failure within 24-48 hours with $>90\%$ probability.
     - **Switch Buffer Pool Saturation Forecasting**: Evaluates ingress buffer utilization trends and micro-burst frequency. If headroom drops below $15\%$ under sustained micro-bursts, it forecasts impending buffer exhaustion within 12-24 hours and prescribes dynamic WRED/ECN QoS reallocation.
   - Dispatches predictive failure events over `eventbus.Bus` on topic `predictive.failure`.

3. **Next.js NetOps & Predictive Cockpit (`/copilot`)**:
   - **Conversational Copilot Workbench**: Real-time chat interface with quick prompt pills (*"Generate Cisco ACL"*, *"Draft Executive Post-Mortem"*, *"Investigate Singapore Packet Loss"*), syntax-highlighted artifact boxes, and one-click copy/deploy actions.
   - **Hardware Failure Radar**: Live prediction cards displaying component name, failure probability, time horizon, root-cause signal, and one-click proactive RMA scheduling.

## Consequences

- Compresses Mean Time to Remediation (MTTR) by automating vendor ACL drafting and executive incident reporting.
- Converts reactive hardware failure firefighting into proactive, scheduled hot-swap maintenance before customer-facing outages occur.
