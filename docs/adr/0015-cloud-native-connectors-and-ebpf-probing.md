# 15. Cloud-Native Multi-Cloud Connectors and Linux eBPF Kernel Probing

## Status

Accepted

## Context

Following the establishment of Qnetward's core multi-node edge mesh (Phases 0–5), expanding to enterprise environments required solving two fundamental observability blind spots:

1. **Microservice & Container Kernel Blind Spots**:
   - Application sidecars introduce CPU, memory, and proxy latency overhead.
   - Traditional user-space socket polling cannot observe sub-millisecond TCP handshake durations, retransmit spikes (`tcp_retransmit_skb`), or connection resets at high throughput.
   - Pod IP churn in Kubernetes requires real-time mapping of network sockets to Pod names, Namespaces, Deployments, and Services.

2. **Multi-Cloud Hybrid Network Ingestion**:
   - Hybrid cloud deployments span AWS, Google Cloud, and Microsoft Azure alongside on-premises datacenters.
   - Each cloud exports flow logs in differing formats (AWS standard text v2, GCP JSON logging payloads, Azure NSG comma-separated flow tuples).
   - Operators need a unified data model (`VPCFlowRecord`) that normalizes security group actions (`ACCEPT`/`REJECT`), byte counts, protocol, and cross-cloud transit latency metrics.

## Decision

1. **Linux eBPF Kernel Socket Tracing (`pkg/ebpf`)**:
   - **`SocketTracer` (`pkg/ebpf/tracer.go`)**: Hooks Linux socket tracepoints (`sockops`, `inet_sock_set_state`) to capture TCP handshake duration, RTT, retransmission counts, and connection resets without decrypting or proxying payloads.
   - **`DNSMetric`**: Captures port 53 DNS queries, latency, and response codes (`NOERROR`, `NXDOMAIN`, `SERVFAIL`).
   - **`K8sEnricher` (`pkg/ebpf/k8s_enricher.go`)**: High-performance thread-safe IP-to-Kubernetes metadata registry resolving container IPs to Pod names, Namespaces, Nodes, and Deployments.

2. **Normalized Multi-Cloud VPC Flow Engine (`pkg/cloudconnect`)**:
   - **`VPCFlowRecord`**: Standardized schema capturing CloudProvider (`aws`, `gcp`, `azure`), AccountID, Region, VPCID, Src/Dst IP, Src/Dst Port, Protocol, Packets, Bytes, and Action (`ACCEPT`/`REJECT`).
   - **`AWSConnector` (`aws.go`)**: Scans standard AWS VPC Flow Log lines, extracts packet/byte counts, and samples CloudWatch Transit Gateway and DirectConnect metrics.
   - **`GCPConnector` (`gcp.go`)**: Parses Google Cloud JSON logging payloads, extracts `bytes_sent` and instance project metadata, and samples Cloud Interconnect metrics.
   - **`AzureConnector` (`azure.go`)**: Parses Azure Network Watcher NSG Flow Log tuples (`timestamp,srcIP,dstIP,srcPort,dstPort,proto,trafficFlow,trafficDecision`), and samples ExpressRoute circuit metrics.
   - **`CloudManager` (`manager.go`)**: Coordinates multiple cloud accounts, publishes normalized flow records to the `eventbus` on `telemetry.cloud.<provider>.flow`, and computes global accepted/rejected traffic statistics.

3. **Web Cockpit Integration (`/cloud`)**:
   - Dedicated Multi-Cloud & eBPF Telemetry workspace (`src/features/cloud/cloud-workspace.tsx` & `src/app/(protected)/cloud/page.tsx`).
   - Pod-to-Pod service mesh table displaying kernel RTT, handshake durations, and TCP retransmission alerts.
   - Live stream table of normalized AWS, GCP, and Azure flow records with filterable provider tabs.

## Consequences

- Zero sidecar proxy overhead for Kubernetes containerized microservices.
- Operators gain single-pane visibility across AWS VPCs, GCP VPCs, Azure VNets, and physical on-prem datacenters.
- Security teams can audit dropped/rejected connections across all clouds simultaneously.
