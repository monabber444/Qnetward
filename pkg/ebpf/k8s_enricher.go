package ebpf

import (
	"sync"
	"time"
)

// K8sEnricher maps IP addresses to Kubernetes workload metadata.
type K8sEnricher struct {
	mu      sync.RWMutex
	ipToPod map[string]*cachedK8sMeta
	ttl     time.Duration
}

type cachedK8sMeta struct {
	meta      *K8sMeta
	expiresAt time.Time
}

// NewK8sEnricher creates a new Kubernetes metadata enricher.
func NewK8sEnricher(ttl time.Duration) *K8sEnricher {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &K8sEnricher{
		ipToPod: make(map[string]*cachedK8sMeta),
		ttl:     ttl,
	}
}

// RegisterPod maps an IP to Kubernetes workload metadata.
func (e *K8sEnricher) RegisterPod(ip string, meta K8sMeta) {
	e.mu.Lock()
	defer e.mu.Unlock()

	labelsCopy := make(map[string]string, len(meta.Labels))
	for k, v := range meta.Labels {
		labelsCopy[k] = v
	}

	cloned := meta
	cloned.Labels = labelsCopy

	e.ipToPod[ip] = &cachedK8sMeta{
		meta:      &cloned,
		expiresAt: time.Now().Add(e.ttl),
	}
}

// Lookup retrieves Kubernetes metadata for a given IP address.
func (e *K8sEnricher) Lookup(ip string) *K8sMeta {
	e.mu.RLock()
	defer e.mu.RUnlock()

	entry, ok := e.ipToPod[ip]
	if !ok {
		return nil
	}

	if time.Now().After(entry.expiresAt) {
		return nil
	}

	return entry.meta
}

// CleanExpired removes stale entries from the cache.
func (e *K8sEnricher) CleanExpired() int {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	cleaned := 0
	for ip, entry := range e.ipToPod {
		if now.After(entry.expiresAt) {
			delete(e.ipToPod, ip)
			cleaned++
		}
	}
	return cleaned
}
