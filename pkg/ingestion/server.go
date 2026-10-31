package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"

	"netwatcher/pkg/eventbus"
	proto "netwatcher/proto/ingestion"
)

var (
	ErrAgentNotConnected = errors.New("agent not connected")
	ErrStreamClosed      = errors.New("stream closed")
)

type AgentInfo struct {
	AgentID           string            `json:"agent_id"`
	TenantID          int64             `json:"tenant_id"`
	Hostname          string            `json:"hostname"`
	Version           string            `json:"version"`
	OS                string            `json:"os"`
	Arch              string            `json:"arch"`
	Capabilities      []string          `json:"capabilities"`
	Labels            map[string]string `json:"labels"`
	ConnectedAt       time.Time         `json:"connected_at"`
	LastHeartbeatAt   time.Time         `json:"last_heartbeat_at"`
	ActiveProbesCount int32             `json:"active_probes_count"`
	CPUUsagePercent   double            `json:"cpu_usage_percent"`
	MemoryUsedBytes   uint64            `json:"memory_used_bytes"`
}

type double = float64

type agentStream struct {
	info     AgentInfo
	stream   proto.IngestionService_StreamChannelServer
	sendCh   chan *proto.GatewayCommand
	closedCh chan struct{}
}

type ServerConfig struct {
	Bus                      eventbus.Bus
	HeartbeatIntervalSeconds int32
	MeshSeedPeers            []string
}

type Server struct {
	proto.UnimplementedIngestionServiceServer

	cfg     ServerConfig
	mu      sync.RWMutex
	agents  map[string]*agentStream
	grpcSrv *grpc.Server
	closed  bool
}

func NewServer(cfg ServerConfig) *Server {
	if cfg.HeartbeatIntervalSeconds <= 0 {
		cfg.HeartbeatIntervalSeconds = 15
	}
	return &Server{
		cfg:    cfg,
		agents: make(map[string]*agentStream),
	}
}

func (s *Server) StreamChannel(stream proto.IngestionService_StreamChannelServer) error {
	ctx := stream.Context()

	var currentAgentID string
	var currentStream *agentStream

	defer func() {
		if currentAgentID != "" {
			s.mu.Lock()
			if existing, ok := s.agents[currentAgentID]; ok && existing == currentStream {
				delete(s.agents, currentAgentID)
				close(existing.closedCh)
			}
			s.mu.Unlock()
			slog.Info("agent disconnected", "agent_id", currentAgentID)
		}
	}()

	for {
		msg, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}

		if msg == nil {
			continue
		}

		switch payload := msg.Payload.(type) {
		case *proto.AgentMessage_Register:
			reg := payload.Register
			currentAgentID = msg.AgentId
			now := time.Now().UTC()

			info := AgentInfo{
				AgentID:         msg.AgentId,
				TenantID:        msg.TenantId,
				Hostname:        reg.Hostname,
				Version:         reg.Version,
				OS:              reg.Os,
				Arch:            reg.Arch,
				Capabilities:    reg.Capabilities,
				Labels:          reg.Labels,
				ConnectedAt:     now,
				LastHeartbeatAt: now,
			}

			currentStream = &agentStream{
				info:     info,
				stream:   stream,
				sendCh:   make(chan *proto.GatewayCommand, 64),
				closedCh: make(chan struct{}),
			}

			s.mu.Lock()
			s.agents[currentAgentID] = currentStream
			s.mu.Unlock()

			// Start sender goroutine for this agent
			go s.agentSender(currentStream)

			slog.Info("agent registered",
				"agent_id", currentAgentID,
				"hostname", reg.Hostname,
				"tenant_id", msg.TenantId,
			)

			// Publish agent.register event to bus
			if s.cfg.Bus != nil {
				if raw, err := json.Marshal(info); err == nil {
					_ = s.cfg.Bus.Publish(ctx, eventbus.TopicAgentRegister, raw)
				}
			}

			// Send RegistrationAck
			ack := &proto.GatewayCommand{
				TimestampNs: time.Now().UnixNano(),
				CommandId:   fmt.Sprintf("ack-%d", time.Now().UnixNano()),
				Payload: &proto.GatewayCommand_RegisterAck{
					RegisterAck: &proto.RegistrationAck{
						Accepted:                 true,
						HeartbeatIntervalSeconds: s.cfg.HeartbeatIntervalSeconds,
						MeshSeedPeers:            s.cfg.MeshSeedPeers,
					},
				},
			}
			currentStream.sendCh <- ack

		case *proto.AgentMessage_Heartbeat:
			hb := payload.Heartbeat
			now := time.Now().UTC()

			s.mu.Lock()
			if cur, ok := s.agents[msg.AgentId]; ok {
				cur.info.LastHeartbeatAt = now
				cur.info.CPUUsagePercent = hb.CpuUsagePercent
				cur.info.MemoryUsedBytes = hb.MemoryUsedBytes
				cur.info.ActiveProbesCount = hb.ActiveProbesCount
			}
			s.mu.Unlock()

			if s.cfg.Bus != nil {
				data, _ := json.Marshal(map[string]any{
					"agent_id":            msg.AgentId,
					"tenant_id":           msg.TenantId,
					"cpu_usage_percent":   hb.CpuUsagePercent,
					"memory_used_bytes":   hb.MemoryUsedBytes,
					"active_probes_count": hb.ActiveProbesCount,
					"timestamp":           now,
				})
				_ = s.cfg.Bus.Publish(ctx, eventbus.TopicAgentHeartbeat, data)
			}

		case *proto.AgentMessage_ProbeResult:
			res := payload.ProbeResult
			if s.cfg.Bus != nil {
				data, _ := json.Marshal(map[string]any{
					"agent_id":      msg.AgentId,
					"tenant_id":     msg.TenantId,
					"probe_id":      res.ProbeId,
					"protocol":      res.Protocol,
					"target":        res.Target,
					"port":          res.Port,
					"sequence":      res.Sequence,
					"rtt_ms":        res.RttMs,
					"error":         res.ErrorMessage,
					"resolved_ip":   res.ResolvedIp,
					"timestamp_ms":  res.TimestampMs,
				})
				_ = s.cfg.Bus.Publish(ctx, eventbus.TopicTelemetryProbe, data)
			}

		case *proto.AgentMessage_MeshMatrix:
			matrix := payload.MeshMatrix
			if s.cfg.Bus != nil {
				data, _ := json.Marshal(map[string]any{
					"agent_id":            msg.AgentId,
					"tenant_id":           msg.TenantId,
					"source_node":         matrix.SourceNode,
					"target_node":         matrix.TargetNode,
					"p50_rtt_ms":          matrix.P50RttMs,
					"p95_rtt_ms":          matrix.P95RttMs,
					"p99_rtt_ms":          matrix.P99RttMs,
					"jitter_ms":           matrix.JitterMs,
					"packet_loss_percent": matrix.PacketLossPercent,
					"samples_count":       matrix.SamplesCount,
					"recorded_at_ms":      matrix.RecordedAtMs,
				})
				_ = s.cfg.Bus.Publish(ctx, eventbus.TopicTelemetryMesh, data)
			}

		case *proto.AgentMessage_TaskResult:
			tr := payload.TaskResult
			if s.cfg.Bus != nil {
				data, _ := json.Marshal(map[string]any{
					"agent_id":    msg.AgentId,
					"tenant_id":   msg.TenantId,
					"task_id":     tr.TaskId,
					"success":     tr.Success,
					"error":       tr.Error,
					"duration_ms": tr.DurationMs,
				})
				_ = s.cfg.Bus.Publish(ctx, "task.result", data)
			}
		}
	}
}

func (s *Server) agentSender(stream *agentStream) {
	for {
		select {
		case <-stream.closedCh:
			return
		case cmd, ok := <-stream.sendCh:
			if !ok {
				return
			}
			if err := stream.stream.Send(cmd); err != nil {
				slog.Error("failed to send command to agent", "agent_id", stream.info.AgentID, "err", err)
				return
			}
		}
	}
}

// SendCommand dispatches a control command to a connected edge agent.
func (s *Server) SendCommand(agentID string, cmd *proto.GatewayCommand) error {
	s.mu.RLock()
	stream, ok := s.agents[agentID]
	s.mu.RUnlock()

	if !ok {
		return ErrAgentNotConnected
	}

	select {
	case stream.sendCh <- cmd:
		return nil
	case <-stream.closedCh:
		return ErrAgentNotConnected
	default:
		return errors.New("agent command queue full")
	}
}

// ConnectedAgents returns a snapshot of all active edge probe agents.
func (s *Server) ConnectedAgents() []AgentInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]AgentInfo, 0, len(s.agents))
	for _, a := range s.agents {
		list = append(list, a.info)
	}
	return list
}

// StartListener starts the gRPC server on the given network listener.
func (s *Server) Start(lis net.Listener) error {
	s.mu.Lock()
	s.grpcSrv = grpc.NewServer()
	proto.RegisterIngestionServiceServer(s.grpcSrv, s)
	s.mu.Unlock()

	return s.grpcSrv.Serve(lis)
}

// Stop gracefully stops the gRPC server.
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.grpcSrv != nil {
		s.grpcSrv.GracefulStop()
	}
	for _, stream := range s.agents {
		close(stream.closedCh)
	}
	s.agents = make(map[string]*agentStream)
	s.closed = true
}
