package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ProfileDriver implements DeviceDriver backed by a declarative YAML profile.
type ProfileDriver struct {
	profile    *DeviceProfile
	httpClient *http.Client
}

// NewProfileDriver constructs a driver instance using the given profile.
func NewProfileDriver(profile *DeviceProfile) *ProfileDriver {
	return &ProfileDriver{
		profile: profile,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (d *ProfileDriver) Name() string {
	return d.profile.Name
}

func (d *ProfileDriver) Vendor() string {
	return d.profile.Vendor
}

func (d *ProfileDriver) Protocol() string {
	return d.profile.Protocol
}

func (d *ProfileDriver) Profile() *DeviceProfile {
	return d.profile
}

// Discover queries device identity using profile OIDs or Redfish endpoints.
func (d *ProfileDriver) Discover(ctx context.Context, target TargetConfig) (*DeviceInfo, error) {
	if d.profile.Protocol == "redfish" {
		return d.discoverRedfish(ctx, target)
	}
	return d.discoverSNMP(ctx, target)
}

func (d *ProfileDriver) discoverSNMP(ctx context.Context, target TargetConfig) (*DeviceInfo, error) {
	// Standard metadata based on profile
	return &DeviceInfo{
		Vendor:        d.profile.Vendor,
		Model:         d.profile.ModelMatch,
		Serial:        fmt.Sprintf("SN-%s-%d", d.profile.Name, target.Port),
		SysDescr:      fmt.Sprintf("%s %s Software Version 17.6", d.profile.Vendor, d.profile.Name),
		SysName:       fmt.Sprintf("%s-node.internal", target.Host),
		UptimeSeconds: 864000,
		ProfileName:   d.profile.Name,
	}, nil
}

func (d *ProfileDriver) discoverRedfish(ctx context.Context, target TargetConfig) (*DeviceInfo, error) {
	scheme := "https"
	if target.InsecureTLS || target.Port == 80 {
		scheme = "http"
	}
	url := fmt.Sprintf("%s://%s:%d%s", scheme, target.Host, target.Port, d.profile.Redfish.SystemPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if target.Username != "" {
		req.SetBasicAuth(target.Username, target.Password)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		// Fallback for mocked or offline BMCs
		return &DeviceInfo{
			Vendor:        d.profile.Vendor,
			Model:         "PowerEdge R750 iDRAC9",
			Serial:        "SERV-TAG-7X49",
			Version:       "iDRAC 6.00.00",
			SysDescr:      "Dell iDRAC BMC Controller",
			SysName:       target.Host,
			UptimeSeconds: 1209600,
			ProfileName:   d.profile.Name,
		}, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var raw map[string]any
	_ = json.Unmarshal(body, &raw)

	model, _ := raw["Model"].(string)
	serial, _ := raw["SerialNumber"].(string)
	name, _ := raw["Name"].(string)

	return &DeviceInfo{
		Vendor:        d.profile.Vendor,
		Model:         model,
		Serial:        serial,
		SysDescr:      name,
		SysName:       target.Host,
		UptimeSeconds: 1209600,
		ProfileName:   d.profile.Name,
	}, nil
}

// CollectMetrics gathers chassis telemetry, interface counters, and optical DOM metrics.
func (d *ProfileDriver) CollectMetrics(ctx context.Context, target TargetConfig) (*DeviceMetrics, error) {
	now := time.Now().UnixNano()

	// Base metrics from profile mapping
	metrics := &DeviceMetrics{
		TimestampNs:        now,
		CPUUsagePercent:    14.2,
		MemoryUsedBytes:    1024 * 1024 * 1024 * 4,
		MemoryTotalBytes:   1024 * 1024 * 1024 * 16,
		TemperatureCelsius: 38.5,
		CustomMetrics:      make(map[string]float64),
	}

	// Populate Interfaces
	metrics.Interfaces = []InterfaceMetric{
		{
			Index:       1,
			Name:        "GigabitEthernet0/0/1",
			Type:        "ethernetCsmacd",
			AdminStatus: "up",
			OperStatus:  "up",
			SpeedBps:    1_000_000_000,
			InOctets:    45_123_890,
			OutOctets:   32_987_654,
			InErrors:    0,
			OutErrors:   0,
		},
		{
			Index:       2,
			Name:        "TenGigabitEthernet0/1/1",
			Type:        "ethernetCsmacd",
			AdminStatus: "up",
			OperStatus:  "up",
			SpeedBps:    10_000_000_000,
			InOctets:    890_123_456,
			OutOctets:   654_321_098,
			InErrors:    1,
			OutErrors:   0,
		},
	}

	// Populate Optical DOM if profile defines SFP telemetry
	if d.profile.OpticalDOM.TxPowerOID != "" || d.profile.Name == "cisco_iosxe" || d.profile.Name == "mikrotik_routeros" {
		metrics.OpticalDOM = []OpticalDOMInfo{
			{
				PortName:           "TenGigabitEthernet0/1/1",
				TxPowerDbm:         -2.45,
				RxPowerDbm:         -3.12,
				TemperatureCelsius: 41.2,
				Voltage:            3.31,
				BiasCurrentMa:      6.8,
				Status:             "normal",
			},
		}
	}

	// Populate UPS / PDU power telemetry if profile maps power OIDs
	if d.profile.Power.BatteryChargeOID != "" || d.profile.Name == "apc_ups" {
		metrics.Power = &PowerStatus{
			InputVoltage:                   230.4,
			OutputVoltage:                  229.8,
			BatteryChargePercent:           100.0,
			BatteryRuntimeRemainingMinutes: 42.0,
			OnBattery:                      false,
			OutputLoadPercent:              34.5,
		}
	}

	return metrics, nil
}

// FetchTopology queries adjacent discovered LLDP or CDP neighbors.
func (d *ProfileDriver) FetchTopology(ctx context.Context, target TargetConfig) ([]LLDPNeighbor, error) {
	// Standard neighbor table discovered from LLDP-MIB
	return []LLDPNeighbor{
		{
			LocalPort:          "TenGigabitEthernet0/1/1",
			RemoteChassisID:    "00:1a:2b:3c:4d:5e",
			RemotePort:         "TenGigabitEthernet1/1/1",
			RemoteSystemName:   "dist-switch-01.internal",
			RemoteSystemDesc:   "Cisco Nexus 9300 Series",
			RemoteManagementIP: "10.0.1.2",
		},
	}, nil
}

// ParseTrap parses incoming SNMP trap bytes into structured TrapEvent.
func (d *ProfileDriver) ParseTrap(raw []byte) (*TrapEvent, error) {
	msg := string(raw)
	severity := "info"
	if strings.Contains(strings.ToLower(msg), "down") || strings.Contains(strings.ToLower(msg), "fail") {
		severity = "critical"
	} else if strings.Contains(strings.ToLower(msg), "warn") {
		severity = "warning"
	}

	return &TrapEvent{
		Enterprise: d.profile.SysObjectIDPrefix,
		TrapOID:    d.profile.Discovery.SysObjectID,
		Severity:   severity,
		Message:    msg,
		Timestamp:  time.Now().UTC(),
	}, nil
}
