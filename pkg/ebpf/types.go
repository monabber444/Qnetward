package ebpf

import "time"

// SocketMetric represents kernel-level TCP socket performance metrics.
type SocketMetric struct {
	Timestamp      time.Time `json:"timestamp"`
	SrcIP          string    `json:"src_ip"`
	SrcPort        uint16    `json:"src_port"`
	DstIP          string    `json:"dst_ip"`
	DstPort        uint16    `json:"dst_port"`
	Protocol       string    `json:"protocol"` // "tcp", "udp"
	HandshakeMs    float64   `json:"handshake_ms"`
	RttMs          float64   `json:"rtt_ms"`
	Retransmits    uint32    `json:"retransmits"`
	BytesSent      uint64    `json:"bytes_sent"`
	BytesReceived  uint64    `json:"bytes_received"`
	State          string    `json:"state"` // "ESTABLISHED", "CLOSED", "RESET", "TIME_WAIT"
	IsReset        bool      `json:"is_reset"`
	K8sSource      *K8sMeta  `json:"k8s_source,omitempty"`
	K8sDestination *K8sMeta  `json:"k8s_destination,omitempty"`
}

// DNSMetric represents kernel-level DNS resolution performance.
type DNSMetric struct {
	Timestamp    time.Time `json:"timestamp"`
	ClientIP     string    `json:"client_ip"`
	ServerIP     string    `json:"server_ip"`
	Question     string    `json:"question"`
	RecordType   string    `json:"record_type"` // "A", "AAAA", "CNAME", "SRV"
	LatencyMs    float64   `json:"latency_ms"`
	RCode        string    `json:"rcode"` // "NOERROR", "NXDOMAIN", "SERVFAIL"
	IsError      bool      `json:"is_error"`
	K8sSource    *K8sMeta  `json:"k8s_source,omitempty"`
}

// K8sMeta holds enriched Kubernetes context for a container IP.
type K8sMeta struct {
	PodName    string            `json:"pod_name"`
	Namespace  string            `json:"namespace"`
	NodeName   string            `json:"node_name"`
	Deployment string            `json:"deployment,omitempty"`
	Service    string            `json:"service,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// TracerConfig configures the eBPF socket tracer.
type TracerConfig struct {
	InterfaceName    string        // e.g. "eth0" or empty for all
	SamplingRate     float64       // 0.0 to 1.0 (default 1.0)
	TraceDNS         bool          // monitor port 53 DNS queries
	RetransmitAlert  uint32        // threshold of retransmits to flag
	EnrichKubernetes bool          // enable K8s pod IP mapping
	FlushInterval    time.Duration // batching interval
}
