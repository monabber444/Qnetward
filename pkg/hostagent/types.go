package hostagent

// DiskPartition stores capacity and IOPS telemetry for a mounted filesystem.
type DiskPartition struct {
	MountPoint string  `json:"mount_point"`
	Device     string  `json:"device"`
	FSType     string  `json:"fs_type"`
	TotalBytes uint64  `json:"total_bytes"`
	UsedBytes  uint64  `json:"used_bytes"`
	FreeBytes  uint64  `json:"free_bytes"`
	IOPSRead   float64 `json:"iops_read"`
	IOPSWrite  float64 `json:"iops_write"`
}

// NetInterface stores interface counters and drop/error statistics.
type NetInterface struct {
	Name        string `json:"name"`
	MAC         string `json:"mac"`
	BytesRecv   uint64 `json:"bytes_recv"`
	BytesSent   uint64 `json:"bytes_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	DropsIn     uint64 `json:"drops_in"`
	DropsOut    uint64 `json:"drops_out"`
	ErrorsIn    uint64 `json:"errors_in"`
	ErrorsOut   uint64 `json:"errors_out"`
}

// ProcessInfo captures resource consumption of a running system process.
type ProcessInfo struct {
	PID         int32   `json:"pid"`
	Name        string  `json:"name"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryBytes uint64  `json:"memory_bytes"`
	Status      string  `json:"status"` // "running", "sleeping", "zombie"
}

// ContainerHealth represents health and telemetry of a Docker or K8s container.
type ContainerHealth struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Image       string  `json:"image"`
	State       string  `json:"state"` // "running", "paused", "restarting"
	Status      string  `json:"status"` // "healthy", "unhealthy"
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryBytes uint64  `json:"memory_bytes"`
}

// SMARTDevice holds disk health and temperature metrics.
type SMARTDevice struct {
	Device       string  `json:"device"`
	Model        string  `json:"model"`
	HealthStatus string  `json:"health_status"` // "PASSED", "WARNING", "FAILED"
	TemperatureC float64 `json:"temperature_c"`
	PowerOnHours uint64  `json:"power_on_hours"`
}

// HostTelemetry aggregates full OS-level telemetry.
type HostTelemetry struct {
	TimestampNs   int64             `json:"timestamp_ns"`
	Hostname      string            `json:"hostname"`
	OS            string            `json:"os"`
	Arch          string            `json:"arch"`
	UptimeSeconds uint64            `json:"uptime_seconds"`
	LoadAvg1      float64           `json:"load_avg_1"`
	LoadAvg5      float64           `json:"load_avg_5"`
	LoadAvg15     float64           `json:"load_avg_15"`
	CPUPercent    float64           `json:"cpu_percent"`
	CPUCores      int               `json:"cpu_cores"`
	MemTotalBytes uint64            `json:"mem_total_bytes"`
	MemUsedBytes  uint64            `json:"mem_used_bytes"`
	MemFreeBytes  uint64            `json:"mem_free_bytes"`
	SwapTotal     uint64            `json:"swap_total_bytes"`
	SwapUsed      uint64            `json:"swap_used_bytes"`
	Disks         []DiskPartition   `json:"disks"`
	Interfaces    []NetInterface    `json:"interfaces"`
	TopProcesses  []ProcessInfo     `json:"top_processes"`
	Containers    []ContainerHealth `json:"containers,omitempty"`
	SMARTDevices  []SMARTDevice     `json:"smart_devices,omitempty"`
}
