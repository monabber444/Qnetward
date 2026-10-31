package predictive

import "time"

// ComponentType identifies the hardware or resource component analyzed.
type ComponentType string

const (
	ComponentOpticalTransceiver ComponentType = "OPTICAL_TRANSCEIVER"
	ComponentSwitchBufferPool   ComponentType = "SWITCH_BUFFER_POOL"
	ComponentFanModule          ComponentType = "FAN_MODULE"
	ComponentPowerSupply        ComponentType = "POWER_SUPPLY"
)

// OpticalReading represents SFP+/QSFP optical telemetry readings.
type OpticalReading struct {
	InterfaceID   string    `json:"interface_id"`
	TxPowerDbm    float64   `json:"tx_power_dbm"`
	RxPowerDbm    float64   `json:"rx_power_dbm"`
	BiasCurrentMa float64   `json:"bias_current_ma"`
	TemperatureC  float64   `json:"temperature_c"`
	Timestamp     time.Time `json:"timestamp"`
}

// BufferReading captures switch packet buffer headroom and micro-burst trends.
type BufferReading struct {
	SwitchID       string    `json:"switch_id"`
	IngressUtilPct float64   `json:"ingress_util_pct"`
	MicroBurstsSec int       `json:"micro_bursts_sec"`
	DroppedPackets int64     `json:"dropped_packets"`
	Timestamp      time.Time `json:"timestamp"`
}

// FailurePrediction details an impending component breakdown predicted before it occurs.
type FailurePrediction struct {
	ID                 string        `json:"id"`
	TargetID           string        `json:"target_id"`
	Component          ComponentType `json:"component"`
	Probability        float64       `json:"probability"` // e.g. 0.94 = 94% confidence
	FailureHorizon     string        `json:"failure_horizon"` // "12-24h", "24-48h", "3-5d"
	RootCauseIndicator string        `json:"root_cause_indicator"`
	RecommendedAction  string        `json:"recommended_action"`
	DetectedAt         time.Time     `json:"detected_at"`
}
