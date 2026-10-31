package sdk

// ProfileDiscovery specifies OIDs or endpoints for device metadata discovery.
type ProfileDiscovery struct {
	SysDescrOID  string `yaml:"sys_descr_oid"`
	SysNameOID   string `yaml:"sys_name_oid"`
	SysUptimeOID string `yaml:"sys_uptime_oid"`
	SysObjectID  string `yaml:"sys_object_id"`
}

// ProfileMetrics maps OIDs for core chassis telemetry.
type ProfileMetrics struct {
	CPUOID         string `yaml:"cpu_oid"`
	MemoryUsedOID  string `yaml:"memory_used_oid"`
	MemoryFreeOID  string `yaml:"memory_free_oid"`
	MemoryTotalOID string `yaml:"memory_total_oid"`
	TemperatureOID string `yaml:"temperature_oid"`
}

// ProfileInterfaces maps standard IF-MIB and IF-MIB-HC table OIDs.
type ProfileInterfaces struct {
	TableOID       string `yaml:"table_oid"`
	NameOID        string `yaml:"name_oid"`
	AdminStatusOID string `yaml:"admin_status_oid"`
	OperStatusOID  string `yaml:"oper_status_oid"`
	InOctetsOID    string `yaml:"in_octets_oid"`
	OutOctetsOID   string `yaml:"out_octets_oid"`
	InErrorsOID    string `yaml:"in_errors_oid"`
	OutErrorsOID   string `yaml:"out_errors_oid"`
}

// ProfileOpticalDOM maps Digital Optical Monitoring telemetry OIDs.
type ProfileOpticalDOM struct {
	TableOID     string `yaml:"table_oid"`
	TxPowerOID   string `yaml:"tx_power_oid"`
	RxPowerOID   string `yaml:"rx_power_oid"`
	TempOID      string `yaml:"temp_oid"`
	VoltageOID   string `yaml:"voltage_oid"`
	BiasCurOID   string `yaml:"bias_current_oid"`
}

// ProfileLLDP maps LLDP-MIB neighbor discovery tables.
type ProfileLLDP struct {
	TableOID         string `yaml:"table_oid"`
	RemoteChassisOID string `yaml:"remote_chassis_oid"`
	RemotePortOID    string `yaml:"remote_port_oid"`
	RemoteSysNameOID string `yaml:"remote_sysname_oid"`
	RemoteMgmtIPOID  string `yaml:"remote_mgmt_ip_oid"`
}

// ProfilePower maps UPS or PDU battery and voltage OIDs.
type ProfilePower struct {
	InputVoltageOID     string `yaml:"input_voltage_oid"`
	OutputVoltageOID    string `yaml:"output_voltage_oid"`
	BatteryChargeOID    string `yaml:"battery_charge_oid"`
	BatteryRuntimeOID   string `yaml:"battery_runtime_oid"`
	OutputLoadOID       string `yaml:"output_load_oid"`
}

// ProfileRedfish specifies Redfish REST endpoints and JSON paths.
type ProfileRedfish struct {
	SystemPath  string `yaml:"system_path"`
	ChassisPath string `yaml:"chassis_path"`
	ThermalPath string `yaml:"thermal_path"`
	PowerPath   string `yaml:"power_path"`
}

// DeviceProfile is the declarative YAML specification for a network or BMC vendor.
type DeviceProfile struct {
	Name              string            `yaml:"name"`
	Vendor            string            `yaml:"vendor"`
	ModelMatch        string            `yaml:"model_match"`
	Protocol          string            `yaml:"protocol"` // "snmp_v2c", "snmp_v3", "redfish"
	SysObjectIDPrefix string            `yaml:"sys_object_id_prefix"`
	Discovery         ProfileDiscovery  `yaml:"discovery"`
	Metrics           ProfileMetrics    `yaml:"metrics"`
	Interfaces        ProfileInterfaces `yaml:"interfaces"`
	OpticalDOM        ProfileOpticalDOM `yaml:"optical_dom"`
	LLDP              ProfileLLDP       `yaml:"lldp"`
	Power             ProfilePower      `yaml:"power"`
	Redfish           ProfileRedfish    `yaml:"redfish"`
}
