package cloudconnect

import "time"

// CloudProvider represents a supported public cloud vendor.
type CloudProvider string

const (
	ProviderAWS   CloudProvider = "aws"
	ProviderGCP   CloudProvider = "gcp"
	ProviderAzure CloudProvider = "azure"
)

// FlowAction represents whether a packet was permitted or dropped.
type FlowAction string

const (
	ActionAccept FlowAction = "ACCEPT"
	ActionReject FlowAction = "REJECT"
)

// VPCFlowRecord is the unified normalized cloud flow record across AWS, GCP, and Azure.
type VPCFlowRecord struct {
	Provider    CloudProvider `json:"provider"`
	AccountID   string        `json:"account_id"`
	Region      string        `json:"region"`
	VPCID       string        `json:"vpc_id"`
	InterfaceID string        `json:"interface_id"`
	SrcIP       string        `json:"src_ip"`
	DstIP       string        `json:"dst_ip"`
	SrcPort     uint16        `json:"src_port"`
	DstPort     uint16        `json:"dst_port"`
	Protocol    int           `json:"protocol"` // 6 for TCP, 17 for UDP, 1 for ICMP
	Packets     uint64        `json:"packets"`
	Bytes       uint64        `json:"bytes"`
	Action      FlowAction    `json:"action"`
	Timestamp   time.Time     `json:"timestamp"`
	TCPFlags    uint8         `json:"tcp_flags,omitempty"`
}

// CloudNetworkMetric represents infrastructure performance telemetry from cloud APIs.
type CloudNetworkMetric struct {
	Provider    CloudProvider     `json:"provider"`
	Region      string            `json:"region"`
	MetricName  string            `json:"metric_name"` // e.g. "TransitGatewayBytes", "DirectConnectLatencyMs"
	Value       float64           `json:"value"`
	Unit        string            `json:"unit"`
	Timestamp   time.Time         `json:"timestamp"`
	ResourceID  string            `json:"resource_id"`
	Dimensions  map[string]string `json:"dimensions,omitempty"`
}

// CloudAccountConfig defines credentials and configuration for cloud connectors.
type CloudAccountConfig struct {
	ID          string            `json:"id"`
	Provider    CloudProvider     `json:"provider"`
	AccountID   string            `json:"account_id"`
	RoleARN     string            `json:"role_arn,omitempty"`
	Regions     []string          `json:"regions"`
	PollInterval time.Duration    `json:"poll_interval"`
	Options     map[string]string `json:"options,omitempty"`
}
