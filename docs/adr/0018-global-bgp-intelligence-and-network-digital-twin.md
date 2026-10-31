# 18. Global BGP Routing Intelligence and Network Digital Twin

## Status

Accepted

## Context

Global enterprise connectivity and hybrid cloud backbones rely heavily on the Border Gateway Protocol (BGP). However:
1. **BGP Vulnerabilities**: BGP lacks built-in encryption or cryptographic route validation by default, making networks susceptible to malicious or accidental prefix hijacks, route leaks, and flapping alert storms.
2. **Lack of Pre-Commit Blast Radius Modeling**: Network operators frequently commit configuration changes (e.g. taking a core router down for firmware maintenance, or shedding traffic during transoceanic fiber incidents) without predictive modeling, risking unforeseen link saturation, routing loops, and critical packet drops.

## Decision

1. **Global BGP Routing Intelligence (`pkg/bgp`)**:
   - **`Analyzer` (`pkg/bgp/analyzer.go`)**: Ingests streaming BGP route updates and validates routing paths against:
     - **RPKI Route Origin Validation (ROV)**: Cross-checks announced origin AS against authorized Route Origin Authorizations (ROAs), immediately flagging unauthorized origins as `AnomalyPrefixHijack` (Severity: `CRITICAL`).
     - **Excessive Prepending Detection**: Identifies AS-Path stuffing / padding anomalies ($\ge 5$ repeated ASNs) that distort shortest-path selection.
     - **Route Flapping Watcher**: Monitors sliding-window update rates, suppressing flapping routes ($\ge 5$ updates per 60 seconds) and alerting operators.
   - Dispatches routing anomalies to `eventbus.Bus` over topic `bgp.anomaly`.

2. **In-Memory Network Digital Twin Engine (`pkg/digitaltwin`)**:
   - **`TwinGraph` (`pkg/digitaltwin/graph.go`)**: Thread-safe in-memory graph representing routers, switches, subsea cables, and steady-state traffic flows.
   - **Predictive Dijkstra Recomputation**: Dynamically simulates "What-If" scenarios:
     - Cable cuts / link severances.
     - Planned router and switch maintenance windows.
     - Traffic surge multipliers (e.g. 2.5x peak events).
   - **Quantitative Blast Radius Metrics**:
     - Calculates affected and rerouted flows with exact path shifts.
     - Quantifies end-to-end latency degradation ($\Delta\text{ms}$).
     - Detects bandwidth saturation hotspots where simulated load exceeds link capacity ($> 90\%$).
     - Flags dropped traffic and isolated/disconnected network partitions.

3. **Unified Next.js Cockpit (`/bgp-twin`)**:
   - Live BGP update feed with RPKI validation status badges and AS-Path carrier hop visualization (Tier-1 carriers: Cogent, Lumen, Telia/Arelion, NTT, HE).
   - Interactive Digital Twin Simulator workbench with one-click failure injection (*"Sever Transatlantic TAT-14"*, *"Frankfurt Maintenance"*) and instant blast radius impact cards.

## Consequences

- Delivers real-time detection of BGP prefix hijacks and transit route leaks before customer traffic is intercepted or blackholed.
- Provides network architects and DevOps engineers with pre-change verification, guaranteeing that planned maintenance or link failures will not cause unexpected cascading link saturation.
