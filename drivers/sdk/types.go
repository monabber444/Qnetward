package sdk

import (
	"context"
	"time"
)

// TargetConfig specifies the connection parameters for a network device or BMC.
type TargetConfig struct {
	Host        string        `json:"host" yaml:"host"`
	Port        int           `json:"port" yaml:"port"`
	Protocol    string        `json:"protocol" yaml:"protocol"` // "snmp_v2c", "snmp_v3", "redfish", "gnmi"
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
	Community   string        `json:"community,omitempty" yaml:"community,omitempty"`     // SNMP v2c
	Username    string        `json:"username,omitempty" yaml:"username,omitempty"`       // SNMP v3 / Redfish
	Password    string        `json:"password,omitempty" yaml:"password,omitempty"`       // Redfish
	AuthProto   string        `json:"auth_proto,omitempty" yaml:"auth_proto,omitempty"`   // SHA, MD5
	AuthPass    string        `json:"auth_pass,omitempty" yaml:"auth_pass,omitempty"`
	PrivProto   string        `json:"priv_proto,omitempty" yaml:"priv_proto,omitempty"`   // AES, DES
	PrivPass    string        `json:"priv_pass,omitempty" yaml:"priv_pass,omitempty"`
	InsecureTLS bool          `json:"insecure_tls,omitempty" yaml:"insecure_tls,omitempty"`
}

// DeviceInfo holds hardware identification metadata.
type DeviceInfo struct {
	Vendor        string `json:"vendor" yaml:"vendor"`
	Model         string `json:"model" yaml:"model"`
	Serial        string `json:"serial" yaml:"serial"`
	Version       string `json:"version" yaml:"version"`
	SysDescr      string `json:"sys_descr" yaml:"sys_descr"`
	SysName       string `json:"sys_name" yaml:"sys_name"`
	UptimeSeconds uint64 `json:"uptime_seconds" yaml:"uptime_seconds"`
	ProfileName   string `json:"profile_name" yaml:"profile_name"`
}

// InterfaceMetric represents status and counters for a single network interface.
type InterfaceMetric struct {
	Index       int32   `json:"index"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	AdminStatus string  `json:"admin_status"` // "up", "down"
	OperStatus  string  `json:"oper_status"`  // "up", "down", "testing"
	SpeedBps    uint64  `json:"speed_bps"`
	InOctets    uint64  `json:"in_octets"`
	OutOctets   uint64  `json:"out_octets"`
	InErrors    uint64  `json:"in_errors"`
	OutErrors   uint64  `json:"out_errors"`
	InDiscards  uint64  `json:"in_discards"`
	OutDiscards uint64  `json:"out_discards"`
}

// OpticalDOMInfo contains Digital Optical Monitoring telemetry from SFP/SFP+/QSFP transceivers.
type OpticalDOMInfo struct {
	PortName           string  `json:"port_name"`
	TxPowerDbm         float64 `json:"tx_power_dbm"`
	RxPowerDbm         float64 `json:"rx_power_dbm"`
	TemperatureCelsius float64 `json:"temperature_celsius"`
	Voltage            float64 `json:"voltage"`
	BiasCurrentMa      float64 `json:"bias_current_ma"`
	Status             string  `json:"status"` // "normal", "warning", "alarm"
}

// PowerStatus represents power supply, PDU, or UPS electrical state.
type PowerStatus struct {
	InputVoltage                  float64 `json:"input_voltage"`
	OutputVoltage                 float64 `json:"output_voltage"`
	BatteryChargePercent          float64 `json:"battery_charge_percent"`
	BatteryRuntimeRemainingMinutes float64 `json:"battery_runtime_remaining_minutes"`
	OnBattery                     bool    `json:"on_battery"`
	OutputLoadPercent             float64 `json:"output_load_percent"`
}

// LLDPNeighbor describes an adjacent discovered node from LLDP or CDP MIBs.
type LLDPNeighbor struct {
	LocalPort          string `json:"local_port"`
	RemoteChassisID    string `json:"remote_chassis_id"`
	RemotePort         string `json:"remote_port"`
	RemoteSystemName   string `json:"remote_system_name"`
	RemoteSystemDesc   string `json:"remote_system_desc"`
	RemoteManagementIP string `json:"remote_management_ip"`
}

// TrapEvent represents an asynchronous alert from an SNMP trap or Syslog entry.
type TrapEvent struct {
	Enterprise string    `json:"enterprise"`
	TrapOID    string    `json:"trap_oid"`
	Severity   string    `json:"severity"` // "info", "warning", "critical"
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
}

// DeviceMetrics aggregates all collected telemetry from a single device poll.
type DeviceMetrics struct {
	TimestampNs        int64             `json:"timestamp_ns"`
	CPUUsagePercent    float64           `json:"cpu_usage_percent"`
	MemoryUsedBytes    uint64            `json:"memory_used_bytes"`
	MemoryTotalBytes   uint64            `json:"memory_total_bytes"`
	TemperatureCelsius float64           `json:"temperature_celsius"`
	Power              *PowerStatus      `json:"power,omitempty"`
	Interfaces         []InterfaceMetric `json:"interfaces,omitempty"`
	OpticalDOM         []OpticalDOMInfo  `json:"optical_dom,omitempty"`
	CustomMetrics      map[string]float64`json:"custom_metrics,omitempty"`
}

// DeviceDriver defines the standard vendor driver interface for physical network infrastructure.
type DeviceDriver interface {
	Name() string
	Vendor() string
	Protocol() string
	Discover(ctx context.Context, target TargetConfig) (*DeviceInfo, error)
	CollectMetrics(ctx context.Context, target TargetConfig) (*DeviceMetrics, error)
	FetchTopology(ctx context.Context, target TargetConfig) ([]LLDPNeighbor, error)
	ParseTrap(raw []byte) (*TrapEvent, error)
}
