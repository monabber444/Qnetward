package connector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"netwatcher/pkg/eventbus"
	"netwatcher/pkg/ingestion"
	proto "netwatcher/proto/ingestion"
)

var (
	ErrNoEligibleAgents = errors.New("no eligible online agents found matching criteria")
	ErrTaskTimeout      = errors.New("task execution timed out")
)

type AgentStatus string

const (
	StatusOnline   AgentStatus = "online"
	StatusDegraded AgentStatus = "degraded"
	StatusOffline  AgentStatus = "offline"
)

type ManagedAgent struct {
	Info       ingestion.AgentInfo `json:"info"`
	Status     AgentStatus         `json:"status"`
	LastSeenAt time.Time           `json:"last_seen_at"`
}

type ManagerConfig struct {
	Gateway          *ingestion.Server
	Bus              eventbus.Bus
	HeartbeatTimeout time.Duration
}

type Manager struct {
	gateway          *ingestion.Server
	bus              eventbus.Bus
	heartbeatTimeout time.Duration

	mu     sync.RWMutex
	agents map[string]*ManagedAgent
	tasks  map[string]chan *proto.TaskExecutionResult
	stopCh chan struct{}
}

func NewManager(cfg ManagerConfig) *Manager {
	if cfg.HeartbeatTimeout <= 0 {
		cfg.HeartbeatTimeout = 30 * time.Second
	}
	return &Manager{
		gateway:          cfg.Gateway,
		bus:              cfg.Bus,
		heartbeatTimeout: cfg.HeartbeatTimeout,
		agents:           make(map[string]*ManagedAgent),
		tasks:            make(map[string]chan *proto.TaskExecutionResult),
		stopCh:           make(chan struct{}),
	}
}

// Start begins periodic agent health checks and bus event subscriptions.
func (m *Manager) Start(ctx context.Context) error {
	if m.bus != nil {
		_, _ = m.bus.Subscribe(ctx, eventbus.TopicAgentRegister, func(c context.Context, msg eventbus.Message) error {
			var info ingestion.AgentInfo
			if err := json.Unmarshal(msg.Data, &info); err == nil {
				m.RegisterAgent(info)
			}
			return nil
		})

		_, _ = m.bus.Subscribe(ctx, eventbus.TopicAgentHeartbeat, func(c context.Context, msg eventbus.Message) error {
			var hb struct {
				AgentID string `json:"agent_id"`
			}
			if err := json.Unmarshal(msg.Data, &hb); err == nil && hb.AgentID != "" {
				m.RecordHeartbeat(hb.AgentID)
			}
			return nil
		})

		_, _ = m.bus.Subscribe(ctx, "task.result", func(c context.Context, msg eventbus.Message) error {
			// Handled via DispatchTask waiting channels
			return nil
		})
	}

	go m.reaperLoop()
	return nil
}

func (m *Manager) Stop() {
	close(m.stopCh)
}

// RegisterAgent updates or records an agent as active.
func (m *Manager) RegisterAgent(info ingestion.AgentInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.agents[info.AgentID] = &ManagedAgent{
		Info:       info,
		Status:     StatusOnline,
		LastSeenAt: time.Now().UTC(),
	}
}

// RecordHeartbeat refreshes last-seen timestamp and status.
func (m *Manager) RecordHeartbeat(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if agent, ok := m.agents[agentID]; ok {
		agent.LastSeenAt = time.Now().UTC()
		agent.Status = StatusOnline
	}
}

// ListAgents returns all registered agents for a tenant.
func (m *Manager) ListAgents(tenantID int64) []ManagedAgent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []ManagedAgent
	for _, a := range m.agents {
		if tenantID == 0 || a.Info.TenantID == tenantID {
			list = append(list, *a)
		}
	}
	return list
}

// SelectAgent finds an online agent matching tenant and required capability.
func (m *Manager) SelectAgent(tenantID int64, capability string) (*ManagedAgent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, a := range m.agents {
		if a.Status != StatusOnline {
			continue
		}
		if tenantID > 0 && a.Info.TenantID != tenantID {
			continue
		}
		if capability != "" && !hasCapability(a.Info.Capabilities, capability) {
			continue
		}
		return a, nil
	}

	return nil, ErrNoEligibleAgents
}

// DispatchProbeTask assigns a probe task to an eligible agent and awaits result.
func (m *Manager) DispatchProbeTask(ctx context.Context, tenantID int64, task *proto.ScheduleProbeTask) error {
	agent, err := m.SelectAgent(tenantID, task.Protocol)
	if err != nil {
		return err
	}

	cmd := &proto.GatewayCommand{
		TimestampNs: time.Now().UnixNano(),
		CommandId:   fmt.Sprintf("cmd-%s", task.TaskId),
		Payload: &proto.GatewayCommand_ScheduleTask{
			ScheduleTask: task,
		},
	}

	return m.gateway.SendCommand(agent.Info.AgentID, cmd)
}

func (m *Manager) reaperLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.sweep()
		}
	}
}

func (m *Manager) sweep() {
	now := time.Now().UTC()
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, a := range m.agents {
		if a.Status == StatusOnline && now.Sub(a.LastSeenAt) > m.heartbeatTimeout {
			a.Status = StatusOffline
		}
	}
}

func hasCapability(caps []string, required string) bool {
	for _, c := range caps {
		if c == required {
			return true
		}
	}
	return false
}
