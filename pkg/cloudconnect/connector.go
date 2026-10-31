package cloudconnect

import "context"

// CloudConnector defines the behavior of a cloud provider telemetry driver.
type CloudConnector interface {
	Provider() CloudProvider
	ParseFlowLogs(raw []byte) ([]VPCFlowRecord, error)
	PollMetrics(ctx context.Context) ([]CloudNetworkMetric, error)
	GetHealth() bool
}
