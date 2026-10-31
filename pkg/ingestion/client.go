package ingestion

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	proto "netwatcher/proto/ingestion"
)

type ClientCallbacks struct {
	OnRegistrationAck func(ack *proto.RegistrationAck)
	OnScheduleTask    func(task *proto.ScheduleProbeTask)
	OnCancelTask      func(task *proto.CancelProbeTask)
	OnMeshConfig      func(config *proto.MeshConfigUpdate)
}

type ClientConfig struct {
	TargetAddr   string
	AgentID      string
	TenantID     int64
	Hostname     string
	Version      string
	OS           string
	Arch         string
	Capabilities []string
	Labels       map[string]string
	Callbacks    ClientCallbacks
}

type Client struct {
	cfg ClientConfig

	mu        sync.Mutex
	conn      *grpc.ClientConn
	stream    proto.IngestionService_StreamChannelClient
	sendCh    chan *proto.AgentMessage
	ctx       context.Context
	cancel    context.CancelFunc
	closed    bool
	connected bool
}

func NewClient(cfg ClientConfig) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		cfg:    cfg,
		sendCh: make(chan *proto.AgentMessage, 256),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start begins the background connection loop.
func (c *Client) Start() error {
	conn, err := grpc.NewClient(c.cfg.TargetAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("dialing gateway %s: %w", c.cfg.TargetAddr, err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	go c.runLoop()
	return nil
}

func (c *Client) runLoop() {
	backoff := time.Second
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		client := proto.NewIngestionServiceClient(c.conn)
		stream, err := client.StreamChannel(c.ctx)
		if err != nil {
			slog.Warn("failed to open ingestion stream, retrying", "err", err, "backoff", backoff)
			select {
			case <-c.ctx.Done():
				return
			case <-time.After(backoff):
			}
			if backoff < 15*time.Second {
				backoff *= 2
			}
			continue
		}

		backoff = time.Second
		c.mu.Lock()
		c.stream = stream
		c.connected = true
		c.mu.Unlock()

		// 1. Send registration
		regMsg := &proto.AgentMessage{
			TimestampNs: time.Now().UnixNano(),
			AgentId:     c.cfg.AgentID,
			TenantId:    c.cfg.TenantID,
			Payload: &proto.AgentMessage_Register{
				Register: &proto.AgentRegister{
					Hostname:     c.cfg.Hostname,
					Version:      c.cfg.Version,
					Os:           c.cfg.OS,
					Arch:         c.cfg.Arch,
					Capabilities: c.cfg.Capabilities,
					Labels:       c.cfg.Labels,
				},
			},
		}

		if err := stream.Send(regMsg); err != nil {
			slog.Error("failed to send registration message", "err", err)
			continue
		}

		slog.Info("connected and registered with Ingestion Gateway", "target", c.cfg.TargetAddr)

		// 2. Run read/write loops
		errCh := make(chan error, 2)
		go c.readPump(stream, errCh)
		go c.writePump(stream, errCh)

		// Wait for failure
		err = <-errCh
		slog.Warn("stream disconnected, reconnecting...", "err", err)

		c.mu.Lock()
		c.stream = nil
		c.connected = false
		c.mu.Unlock()

		select {
		case <-c.ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

func (c *Client) readPump(stream proto.IngestionService_StreamChannelClient, errCh chan error) {
	for {
		cmd, err := stream.Recv()
		if err != nil {
			errCh <- err
			return
		}
		if cmd == nil {
			continue
		}

		switch p := cmd.Payload.(type) {
		case *proto.GatewayCommand_RegisterAck:
			if c.cfg.Callbacks.OnRegistrationAck != nil {
				c.cfg.Callbacks.OnRegistrationAck(p.RegisterAck)
			}
		case *proto.GatewayCommand_ScheduleTask:
			if c.cfg.Callbacks.OnScheduleTask != nil {
				c.cfg.Callbacks.OnScheduleTask(p.ScheduleTask)
			}
		case *proto.GatewayCommand_CancelTask:
			if c.cfg.Callbacks.OnCancelTask != nil {
				c.cfg.Callbacks.OnCancelTask(p.CancelTask)
			}
		case *proto.GatewayCommand_MeshConfig:
			if c.cfg.Callbacks.OnMeshConfig != nil {
				c.cfg.Callbacks.OnMeshConfig(p.MeshConfig)
			}
		}
	}
}

func (c *Client) writePump(stream proto.IngestionService_StreamChannelClient, errCh chan error) {
	for {
		select {
		case <-c.ctx.Done():
			errCh <- c.ctx.Err()
			return
		case msg := <-c.sendCh:
			if err := stream.Send(msg); err != nil {
				errCh <- err
				return
			}
		}
	}
}

// SendHeartbeat queues a heartbeat message for transmission.
func (c *Client) SendHeartbeat(cpuPercent float64, memBytes uint64, activeProbes int32) error {
	return c.enqueue(&proto.AgentMessage{
		TimestampNs: time.Now().UnixNano(),
		AgentId:     c.cfg.AgentID,
		TenantId:    c.cfg.TenantID,
		Payload: &proto.AgentMessage_Heartbeat{
			Heartbeat: &proto.AgentHeartbeat{
				CpuUsagePercent:   cpuPercent,
				MemoryUsedBytes:   memBytes,
				ActiveProbesCount: activeProbes,
			},
		},
	})
}

// SendProbeResult queues an individual probe result for transmission.
func (c *Client) SendProbeResult(res *proto.ProbeTelemetry) error {
	return c.enqueue(&proto.AgentMessage{
		TimestampNs: time.Now().UnixNano(),
		AgentId:     c.cfg.AgentID,
		TenantId:    c.cfg.TenantID,
		Payload: &proto.AgentMessage_ProbeResult{
			ProbeResult: res,
		},
	})
}

// SendMeshMatrix queues a PulseMesh matrix snapshot for transmission.
func (c *Client) SendMeshMatrix(matrix *proto.MeshMatrixTelemetry) error {
	return c.enqueue(&proto.AgentMessage{
		TimestampNs: time.Now().UnixNano(),
		AgentId:     c.cfg.AgentID,
		TenantId:    c.cfg.TenantID,
		Payload: &proto.AgentMessage_MeshMatrix{
			MeshMatrix: matrix,
		},
	})
}

// SendTaskResult queues a task execution response for transmission.
func (c *Client) SendTaskResult(result *proto.TaskExecutionResult) error {
	return c.enqueue(&proto.AgentMessage{
		TimestampNs: time.Now().UnixNano(),
		AgentId:     c.cfg.AgentID,
		TenantId:    c.cfg.TenantID,
		Payload: &proto.AgentMessage_TaskResult{
			TaskResult: result,
		},
	})
}

func (c *Client) enqueue(msg *proto.AgentMessage) error {
	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()

	if closed {
		return errors.New("client closed")
	}

	select {
	case c.sendCh <- msg:
		return nil
	default:
		return errors.New("client send queue full, dropping message")
	}
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true
	c.cancel()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
