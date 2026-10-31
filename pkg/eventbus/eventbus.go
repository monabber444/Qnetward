package eventbus

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrClosed       = errors.New("event bus is closed")
	ErrInvalidTopic = errors.New("invalid or empty topic")
	ErrNilHandler   = errors.New("handler cannot be nil")
)

// Standard topic names used across Qnetward
const (
	TopicTelemetryProbe     = "telemetry.probe"
	TopicTelemetryMesh      = "telemetry.mesh"
	TopicAgentHeartbeat     = "agent.heartbeat"
	TopicAgentRegister      = "agent.register"
	TopicTaskCommand        = "task.command"
	TopicAlertEvaluated     = "alert.evaluated"
)

// Message encapsulates a published event payload.
type Message struct {
	Subject string
	Data    []byte
	Headers map[string]string
}

// Handler handles an incoming event message.
type Handler func(ctx context.Context, msg Message) error

// Subscription represents an active subscription that can be cancelled.
type Subscription interface {
	Unsubscribe() error
}

// Bus defines the messaging interface across the distributed control plane.
type Bus interface {
	Publish(ctx context.Context, subject string, data []byte) error
	Subscribe(ctx context.Context, subject string, handler Handler) (Subscription, error)
	Close() error
}

// MemoryBus provides a concurrent in-memory event bus implementation
// suitable for development, embedded modes, and automated testing.
type MemoryBus struct {
	mu     sync.RWMutex
	subs   map[string][]*memSub
	closed bool
}

type memSub struct {
	id      int64
	subject string
	handler Handler
	bus     *MemoryBus
}

func (s *memSub) Unsubscribe() error {
	return s.bus.unsubscribe(s)
}

// NewMemoryBus creates a new in-memory event bus.
func NewMemoryBus() *MemoryBus {
	return &MemoryBus{
		subs: make(map[string][]*memSub),
	}
}

func (b *MemoryBus) Publish(ctx context.Context, subject string, data []byte) error {
	if subject == "" {
		return ErrInvalidTopic
	}

	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return ErrClosed
	}

	// Copy matching subscribers
	var targets []*memSub
	if list, ok := b.subs[subject]; ok {
		targets = append(targets, list...)
	}
	// Wildcard support e.g. "telemetry.*" or ">"
	for pat, list := range b.subs {
		if pat != subject && matchSubject(pat, subject) {
			targets = append(targets, list...)
		}
	}
	b.mu.RUnlock()

	msg := Message{
		Subject: subject,
		Data:    data,
		Headers: make(map[string]string),
	}

	for _, sub := range targets {
		s := sub
		go func() {
			_ = s.handler(ctx, msg)
		}()
	}

	return nil
}

func (b *MemoryBus) Subscribe(ctx context.Context, subject string, handler Handler) (Subscription, error) {
	if subject == "" {
		return nil, ErrInvalidTopic
	}
	if handler == nil {
		return nil, ErrNilHandler
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, ErrClosed
	}

	sub := &memSub{
		subject: subject,
		handler: handler,
		bus:     b,
	}

	b.subs[subject] = append(b.subs[subject], sub)
	return sub, nil
}

func (b *MemoryBus) unsubscribe(s *memSub) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	list := b.subs[s.subject]
	for i, sub := range list {
		if sub == s {
			b.subs[s.subject] = append(list[:i], list[i+1:]...)
			break
		}
	}
	return nil
}

func (b *MemoryBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true
	b.subs = make(map[string][]*memSub)
	return nil
}

func matchSubject(pattern, subject string) bool {
	if pattern == ">" {
		return true
	}
	pParts := splitSubject(pattern)
	sParts := splitSubject(subject)

	for i := 0; i < len(pParts); i++ {
		if pParts[i] == ">" {
			return true
		}
		if i >= len(sParts) {
			return false
		}
		if pParts[i] != "*" && pParts[i] != sParts[i] {
			return false
		}
	}
	return len(pParts) == len(sParts)
}

func splitSubject(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	if start <= len(s) {
		parts = append(parts, s[start:])
	}
	return parts
}
