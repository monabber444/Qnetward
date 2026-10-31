package appliance

import (
	"time"
)

// HardwareModel identifies the physical micro-probe edge appliance.
type HardwareModel string

const (
	ModelRaspberryPi5  HardwareModel = "RASPBERRY_PI_5"
	ModelIntelNUCEdge  HardwareModel = "INTEL_NUC_EDGE"
	ModelOdroidM1      HardwareModel = "ODROID_M1"
	ModelX86MiniPC     HardwareModel = "X86_MINI_PC"
)

// UplinkState indicates whether primary WAN or cellular 5G is currently carrying traffic.
type UplinkState string

const (
	UplinkPrimaryWAN UplinkState = "PRIMARY_WAN"
	UplinkCellular5G UplinkState = "CELLULAR_5G"
)

// NetworkInterface represents an active network card on the appliance.
type NetworkInterface struct {
	Name         string `json:"name"`          // e.g. "eth0", "wwan0"
	Type         string `json:"type"`          // "ETHERNET_WAN", "ETHERNET_LAN", "CELLULAR_5G"
	IPAddress    string `json:"ip_address"`
	MACAddress   string `json:"mac_address"`
	CarrierName  string `json:"carrier_name"`  // e.g. "Verizon 5G", "Vodafone IoT", "T-Mobile"
	SignalRSRP   int    `json:"signal_rsrp"`   // RSRP in dBm (e.g. -78 dBm)
	IsConnected  bool   `json:"is_connected"`
}

// Appliance describes a deployed hardware micro-probe.
type Appliance struct {
	ID              string             `json:"id"`
	SerialNumber    string             `json:"serial_number"`
	TenantID        string             `json:"tenant_id"`
	LocationName    string             `json:"location_name"`
	HardwareModel   HardwareModel      `json:"hardware_model"`
	FirmwareVersion string             `json:"firmware_version"`
	Interfaces      []NetworkInterface `json:"interfaces"`
	ActiveUplink    UplinkState        `json:"active_uplink"`
	Status          string             `json:"status"` // "ONLINE", "FAILOVER", "OFFLINE"
	LastHeartbeat   time.Time          `json:"last_heartbeat"`
}

// ZTPEnrollmentToken encapsulates a zero-touch provisioning token.
type ZTPEnrollmentToken struct {
	Token          string    `json:"token"`
	TenantID       string    `json:"tenant_id"`
	TargetDeviceID string    `json:"target_device_id"`
	ExpiresAt      time.Time `json:"expires_at"`
	Claimed        bool      `json:"claimed"`
	ClaimedAt      time.Time `json:"claimed_at"`
}

// WANFailoverEvent records when a branch office automatically fails over to cellular 5G.
type WANFailoverEvent struct {
	ID                     string      `json:"id"`
	ApplianceID            string      `json:"appliance_id"`
	LocationName           string      `json:"location_name"`
	Timestamp              time.Time   `json:"timestamp"`
	PreviousUplink         UplinkState `json:"previous_uplink"`
	NewUplink              UplinkState `json:"new_uplink"`
	FailureCause           string      `json:"failure_cause"` // "PRIMARY_WAN_UNREACHABLE", "FIBER_CUT", "RECOVERED"
	CellularSignalRSRP     int         `json:"cellular_signal_rsrp"`
}
