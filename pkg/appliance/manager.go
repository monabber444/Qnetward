package appliance

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"netwatcher/pkg/eventbus"
)

const (
	TopicApplianceFailover  = "appliance.failover"
	TopicApplianceRecovered = "appliance.recovered"
)

var (
	ErrInvalidToken   = errors.New("invalid or expired enrollment token")
	ErrTokenClaimed   = errors.New("enrollment token already claimed")
	ErrApplianceFound = errors.New("appliance not found")
)

// Manager orchestrates zero-touch provisioning and 5G out-of-band failover.
type Manager struct {
	mu         sync.RWMutex
	appliances map[string]*Appliance
	tokens     map[string]*ZTPEnrollmentToken
	failovers  []*WANFailoverEvent
	bus        eventbus.Bus
}

// NewManager creates a micro-probe fleet manager.
func NewManager(bus eventbus.Bus) *Manager {
	return &Manager{
		appliances: make(map[string]*Appliance),
		tokens:     make(map[string]*ZTPEnrollmentToken),
		failovers:  make([]*WANFailoverEvent, 0),
		bus:        bus,
	}
}

// GenerateZTPToken produces a cryptographically secure one-time enrollment token.
func (m *Manager) GenerateZTPToken(tenantID, targetDeviceID string, validFor time.Duration) (*ZTPEnrollmentToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to generate random token: %w", err)
	}
	tokenStr := hex.EncodeToString(b)

	token := &ZTPEnrollmentToken{
		Token:          tokenStr,
		TenantID:       tenantID,
		TargetDeviceID: targetDeviceID,
		ExpiresAt:      time.Now().Add(validFor),
		Claimed:        false,
	}

	m.tokens[tokenStr] = token
	return token, nil
}

// ClaimZTPToken provisions and enrolls a new appliance using a valid ZTP token.
func (m *Manager) ClaimZTPToken(tokenStr, serialNumber, locationName string, model HardwareModel) (*Appliance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	token, exists := m.tokens[tokenStr]
	if !exists || time.Now().After(token.ExpiresAt) {
		return nil, ErrInvalidToken
	}
	if token.Claimed {
		return nil, ErrTokenClaimed
	}

	token.Claimed = true
	token.ClaimedAt = time.Now()

	applianceID := token.TargetDeviceID
	if applianceID == "" {
		applianceID = fmt.Sprintf("probe-%s", serialNumber)
	}

	app := &Appliance{
		ID:              applianceID,
		SerialNumber:    serialNumber,
		TenantID:        token.TenantID,
		LocationName:    locationName,
		HardwareModel:   model,
		FirmwareVersion: "qnetward-probe-v1.4.2-alpine",
		Interfaces: []NetworkInterface{
			{
				Name:        "eth0",
				Type:        "ETHERNET_WAN",
				IPAddress:   "192.168.1.100",
				MACAddress:  "b8:27:eb:42:01:aa",
				IsConnected: true,
			},
			{
				Name:        "wwan0",
				Type:        "CELLULAR_5G",
				IPAddress:   "10.64.12.89",
				CarrierName: "Verizon 5G Enterprise",
				SignalRSRP:  -78,
				IsConnected: true,
			},
		},
		ActiveUplink:  UplinkPrimaryWAN,
		Status:        "ONLINE",
		LastHeartbeat: time.Now(),
	}

	m.appliances[app.ID] = app
	return app, nil
}

// ProcessUplinkTelemetry evaluates primary WAN vs 5G uplink health, executing failover if needed.
func (m *Manager) ProcessUplinkTelemetry(applianceID string, primaryWANUp bool, cellularUp bool, rsrp int) (*WANFailoverEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	app, exists := m.appliances[applianceID]
	if !exists {
		return nil, ErrApplianceFound
	}

	now := time.Now()
	app.LastHeartbeat = now

	// Update wwan0 cellular signal
	for i := range app.Interfaces {
		if app.Interfaces[i].Type == "CELLULAR_5G" {
			app.Interfaces[i].SignalRSRP = rsrp
			app.Interfaces[i].IsConnected = cellularUp
		} else if app.Interfaces[i].Type == "ETHERNET_WAN" {
			app.Interfaces[i].IsConnected = primaryWANUp
		}
	}

	// 1. Failover condition: Primary WAN goes down, Cellular 5G is available
	if app.ActiveUplink == UplinkPrimaryWAN && !primaryWANUp && cellularUp {
		app.ActiveUplink = UplinkCellular5G
		app.Status = "FAILOVER"

		event := &WANFailoverEvent{
			ID:                 fmt.Sprintf("fo-%d", now.UnixNano()%1000000),
			ApplianceID:        app.ID,
			LocationName:       app.LocationName,
			Timestamp:          now,
			PreviousUplink:     UplinkPrimaryWAN,
			NewUplink:          UplinkCellular5G,
			FailureCause:       "PRIMARY_WAN_UNREACHABLE",
			CellularSignalRSRP: rsrp,
		}

		m.failovers = append(m.failovers, event)

		if m.bus != nil {
			if data, err := json.Marshal(event); err == nil {
				_ = m.bus.Publish(context.Background(), TopicApplianceFailover, data)
			}
		}

		return event, nil
	}

	// 2. Failback condition: Primary WAN recovers
	if app.ActiveUplink == UplinkCellular5G && primaryWANUp {
		app.ActiveUplink = UplinkPrimaryWAN
		app.Status = "ONLINE"

		event := &WANFailoverEvent{
			ID:                 fmt.Sprintf("fb-%d", now.UnixNano()%1000000),
			ApplianceID:        app.ID,
			LocationName:       app.LocationName,
			Timestamp:          now,
			PreviousUplink:     UplinkCellular5G,
			NewUplink:          UplinkPrimaryWAN,
			FailureCause:       "RECOVERED",
			CellularSignalRSRP: rsrp,
		}

		m.failovers = append(m.failovers, event)

		if m.bus != nil {
			if data, err := json.Marshal(event); err == nil {
				_ = m.bus.Publish(context.Background(), TopicApplianceRecovered, data)
			}
		}

		return event, nil
	}

	return nil, nil
}

// GetAppliances returns appliances registered to a tenant.
func (m *Manager) GetAppliances(tenantID string) []*Appliance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*Appliance
	for _, a := range m.appliances {
		if tenantID == "" || a.TenantID == tenantID {
			res = append(res, a)
		}
	}
	return res
}

// GetFailovers returns recorded uplink failover history.
func (m *Manager) GetFailovers() []*WANFailoverEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]*WANFailoverEvent, len(m.failovers))
	copy(res, m.failovers)
	return res
}
