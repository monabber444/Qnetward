# 23. Distributed Consensus Locks and Kubernetes GitOps Operator

## Status

Accepted

## Context

Running Qnetward across multiple global geographic regions with thousands of automated playbooks requires:
1. **Multi-Region Distributed State Coordination**: Ensuring that only one coordinator acts as leader, state machine mutations are committed with quorum ($N/2 + 1$), and conflicting AIOps playbooks cannot execute simultaneously against the same router or switch.
2. **Kubernetes-Native GitOps Automation**: Managing massive fleets of edge probe cells, clickhouse shards, and control plane pods requires declarative Custom Resource Definitions (`QnetwardCluster`, `ProbeMesh`) managed via GitOps (ArgoCD, Flux, Helm).

## Decision

1. **Distributed Consensus & Mutex Engine (`pkg/consensus`)**:
   - **`Coordinator` (`pkg/consensus/coordinator.go`)**:
     - Participates in Raft leader election across multi-region nodes.
     - **Distributed Mutex (`AcquireLock` / `ReleaseLock`)**: Prevents conflicting topology mutations by granting time-bounded exclusive resource locks (`resourceKey` with TTL).
     - **Replicated Log (`Propose`)**: Synchronously replicates state updates across nodes before committing to the local state machine.

2. **Kubernetes Operator & CRDs (`deploy/k8s/operator`)**:
   - **`QnetwardCluster` CRD**: Declaratively configures control-plane replicas, multi-region Anycast topology, hot storage engines (ClickHouse / TimescaleDB), and cold S3 archive policies.
   - **`ProbeMesh` CRD**: Defines regional cell boundaries ($N \le 250$), super-probe redundancies, target CIDR subnets, and UDP probe frequencies.
   - **`Deployment` & RBAC**: Provides self-healing operator controller with least-privilege RBAC bindings and non-root security contexts.

## Consequences

- Prevents race conditions and split-brain execution across multi-region AIOps remediation workflows.
- Enables automated multi-cluster deployment via GitOps pipelines with declarative Kubernetes manifests.
