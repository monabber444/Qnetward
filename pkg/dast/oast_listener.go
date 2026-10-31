package dast

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"netwatcher/pkg/eventbus"
)

const (
	TopicOASTInteraction = "oast.interaction"
)

// OASTManager handles generating unique correlation tokens and tracking out-of-band callbacks.
type OASTManager struct {
	baseDomain   string
	bus          eventbus.Bus
	mu           sync.RWMutex
	tokens       map[string]*OASTTokenRecord
	interactions []*OASTInteraction
}

// OASTTokenRecord tracks metadata for a generated correlation token.
type OASTTokenRecord struct {
	Token       string    `json:"token"`
	TargetURL   string    `json:"target_url"`
	PayloadType string    `json:"payload_type"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewOASTManager creates an Out-of-Band Application Security Testing tracker.
func NewOASTManager(baseDomain string, bus eventbus.Bus) *OASTManager {
	if baseDomain == "" {
		baseDomain = "oast.qnetward.internal"
	}
	return &OASTManager{
		baseDomain:   baseDomain,
		bus:          bus,
		tokens:       make(map[string]*OASTTokenRecord),
		interactions: make([]*OASTInteraction, 0),
	}
}

// GenerateToken creates a cryptographically random token associated with a target audit.
func (m *OASTManager) GenerateToken(targetURL string, payloadType string) (string, string) {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)
	callbackDomain := fmt.Sprintf("%s.%s", token, m.baseDomain)

	record := &OASTTokenRecord{
		Token:       token,
		TargetURL:   targetURL,
		PayloadType: payloadType,
		CreatedAt:   time.Now(),
	}

	m.mu.Lock()
	m.tokens[token] = record
	m.mu.Unlock()

	return token, callbackDomain
}

// RecordInteraction processes an incoming out-of-band callback (HTTP or DNS).
func (m *OASTManager) RecordInteraction(ctx context.Context, token string, protocol string, remoteAddr string, queryParam string) *OASTInteraction {
	m.mu.Lock()
	record, exists := m.tokens[token]
	targetURL := "unknown"
	payloadType := "OUT_OF_BAND_PING"
	if exists {
		targetURL = record.TargetURL
		payloadType = record.PayloadType
	}

	interaction := &OASTInteraction{
		Token:       token,
		TargetURL:   targetURL,
		Protocol:    protocol,
		RemoteAddr:  remoteAddr,
		QueryParam:  queryParam,
		Timestamp:   time.Now(),
		PayloadType: payloadType,
	}

	m.interactions = append(m.interactions, interaction)
	m.mu.Unlock()

	if m.bus != nil {
		if payload, err := json.Marshal(interaction); err == nil {
			_ = m.bus.Publish(ctx, TopicOASTInteraction, payload)
		}
	}

	return interaction
}

// HTTPHandler returns an http.HandlerFunc that accepts inbound OAST HTTP pings.
func (m *OASTManager) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			// Extract token from Host header prefix if present (e.g. "abcdef123.oast.internal")
			hostParts := r.Host
			idx := 0
			for i, c := range hostParts {
				if c == '.' || c == ':' {
					idx = i
					break
				}
			}
			if idx > 0 {
				token = hostParts[:idx]
			}
		}

		if token != "" {
			m.RecordInteraction(r.Context(), token, "HTTP", r.RemoteAddr, r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"captured"}`))
	}
}

// GetInteractions returns all recorded OAST interactions.
func (m *OASTManager) GetInteractions() []*OASTInteraction {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]*OASTInteraction, len(m.interactions))
	copy(res, m.interactions)
	return res
}
