# 11. Hardware Collector, Vendor Driver SDK, and Topology-Aware Alerting

## Status

Accepted

## Context

Qnetward previously monitored network endpoints and synthetic path health via software probes (`netwatcher` and `probe-agent`). To support enterprise data centers, campus networks, and multi-cloud infrastructure, the platform requires deep hardware and physical network observability:

1. **Vendor Infrastructure Telemetry**: Support for physical routers, switches, optical transceivers, server BMCs, and power distribution units (UPS/PDU) across heterogeneous hardware vendors (Cisco, MikroTik, Ubiquiti, Dell, APC).
2. **Standardized Driver SDK & Declarative Profiles**: Rather than hardcoding vendor OIDs in Go, telemetry definitions must be declarative YAML profiles mapping SNMP OIDs, Redfish REST endpoints, Digital Optical Monitoring (DOM) thresholds, and LLDP tables.
3. **OS Host Telemetry**: Lightweight host agent daemon capturing CPU, RAM, Disk IOPS/capacity, SMART disk diagnostics, network interface drop counters, and container health (Docker/Kubernetes).
4. **Topology-Aware Alerting v2**: During upstream core network outages (e.g. core router failure), hundreds of downstream switches, servers, and services fail reachability checks simultaneously, triggering massive alert storms. The system must automatically build directed topology graphs and suppress downstream alerts.

## Decision

1. **Vendor Driver SDK & Declarative Profiles (`drivers/sdk` & `drivers/profiles`)**:
   - `drivers/sdk/types.go`: Standardized `DeviceDriver` interface (`Discover`, `CollectMetrics`, `FetchTopology`, `ParseTrap`) and domain models (`DeviceInfo`, `InterfaceMetric`, `OpticalDOMInfo`, `PowerStatus`, `LLDPNeighbor`, `TrapEvent`).
   - `drivers/sdk/loader.go`: Profile registry supporting dynamic YAML profile parsing from filesystem directories or embedded assets, with automatic `sysObjectID` prefix matching.
   - `drivers/sdk/driver.go`: Generic `ProfileDriver` translating declarative YAML definitions into SNMP and Redfish REST operations.
   - Ships production YAML profiles in `drivers/profiles/`:
     - `cisco_iosxe.yaml`: Cisco Catalyst / CSR1000v (CPU, memory, SFP DOM transceivers, LLDP/CDP).
     - `mikrotik_routeros.yaml`: MikroTik RouterOS (CPU, voltage, temperature, SFP DOM).
     - `ubiquiti_unifi.yaml`: Ubiquiti EdgeSwitch / UniFi Switch (CPU, memory, ports).
     - `dell_idrac_redfish.yaml`: Dell PowerEdge iDRAC9 (Redfish thermal, power, disks).
     - `apc_ups.yaml`: APC Smart-UPS (line voltage, battery capacity, runtime remaining).

2. **Network Hardware Collector (`pkg/collector` & `cmd/collector`)**:
   - Standalone collector daemon polling configured network infrastructure via driver SDK.
   - NetFlow v5 / IPFIX packet parser extracting flow records (IP 5-tuple, packet/byte counters, TCP flags).
   - Asynchronous UDP listeners for SNMP traps / Syslog and NetFlow datagrams.

3. **OS Host Agent (`pkg/hostagent` & `cmd/host-agent`)**:
   - Lightweight daemon capturing Linux/cross-platform host telemetry: `/proc/loadavg`, `/proc/meminfo`, root filesystem block usage via `syscall.Statfs`, network interface drop/error counters via `net.Interfaces()`, process statistics, container health, and SMART disk health.

4. **Topology-Aware Alerting v2 (`pkg/topology`)**:
   - `pkg/topology/graph.go`: Directed dependency graph where edges flow from Upstream (Parent) -> Downstream (Child). Automatically populated manually or dynamically from LLDP neighbor discovery.
   - `pkg/topology/suppression.go`: **Dependency-Aware Suppression Engine**:
     - Evaluates incoming alert storms against the graph.
     - Identifies true upstream root-cause nodes (nodes in critical state whose upstream ancestors are healthy).
     - Automatically suppresses all downstream alerts originating from descendants of the failed node, tagging each with `Suppressed: true` and `RootCauseNodeID`.
     - Bundles the root cause and all suppressed alerts into a unified `Incident`, preventing alert fatigue and providing immediate root-cause diagnosis.

## Consequences

- Full multi-vendor infrastructure observability is achieved with zero hardcoded OID logic in application code.
- Downstream alert storms during core network failures are eliminated via deterministic dependency-aware suppression.
- Hardware collector and host agent binaries are modular and can run standalone or alongside the main platform.
