# 22. Hyperscale Partitioning, Columnar OLAP, and Hierarchical Gossip (H-SWIM)

## Status

Accepted

## Context

Scaling Qnetward to 1,000,000+ edge probes, 100M telemetry metrics per second, and global multi-region deployments introduced three critical architectural challenges:
1. **Streaming Bus & Ingestion In-Order Contention**: Single-channel event buses and unkeyed pub/sub cannot scale beyond 100,000 metrics/sec without thread contention and loss of per-device sequential ordering.
2. **$O(N^2)$ Flat Gossip Saturation**: Traditional flat SWIM gossip matrices flood local and wide-area network links as node counts grow beyond 10,000 nodes ($N \times (N-1)$ echo packets).
3. **High-Throughput Time-Series OLAP**: Row-oriented relational databases struggle under sustained writes of 100M rows/second, requiring high-throughput columnar storage with automatic S3 cold archiving.

## Decision

1. **Partitioned Event Bus & Adaptive Micro-Batching (`pkg/eventbus`)**:
   - **`PartitionedStreamBus` (`pkg/eventbus/partitioned.go`)**: Key-based deterministic hash partitioning:
     $$\text{Partition} = \text{fnv64a}(\text{key}) \pmod{\text{NumPartitions}}$$
     Guarantees in-order delivery per edge device while load-balancing across independent partition workers.
   - **`AdaptiveBatcher` (`pkg/eventbus/batcher.go`)**: Non-blocking micro-batching mechanism that dynamically flushes based on size (`maxBatchSize`) or latency timer (`flushInterval`), preventing syscall overhead under heavy burst loads.

2. **Hierarchical Gossip Clustering: H-SWIM (`pkg/mesh/cluster/hswim.go`)**:
   - **Topological Cells**: Probes are divided into regional cells ($N \le 250$). Probes only execute high-frequency UDP echo matrix probing within their local cell.
   - **Super-Probe Aggregation**: Designated Cell Super-Probes summarize regional matrices and execute cross-cell WAN probing, cutting global packet overhead by over $95\%$.

3. **Columnar OLAP Engine & Parquet Offloader (`pkg/olap`)**:
   - **`ClickHouseEngine` (`pkg/olap/engine.go`)**: Columnar time-series batch insertion and percentile analytics (Min, Max, Avg, P50, P95, P99).
   - **Automated S3 Offloading**: As partitions age past the hot retention threshold, data is serialized into compressed Parquet-equivalent binary blocks for long-term storage in SeaweedFS/Ceph S3 object stores.

## Consequences

- Unlocks horizontal scale to 1,000,000+ concurrent edge devices.
- Prevents WAN link saturation by bounding peer gossip matrices within local cells ($N \le 250$).
- Guarantees sub-50ms analytics query latency across millions of telemetry metrics via columnar aggregation.
