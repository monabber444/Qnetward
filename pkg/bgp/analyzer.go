package bgp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"netwatcher/pkg/eventbus"
)

const (
	TopicBGPAnomaly = "bgp.anomaly"
)

// Analyzer inspects streaming BGP updates for prefix hijacks, route leaks, and flapping.
type Analyzer struct {
	mu            sync.RWMutex
	roas          map[string]uint32              // prefix -> authorized origin ASN
	routes        map[string]*RouteTableEntry    // prefix -> active route state
	updateHistory map[string][]time.Time         // prefix -> timestamps of recent updates
	bus           eventbus.Bus
	anomalies     []*BGPAnomaly
}

// NewAnalyzer creates a real-time BGP routing intelligence analyzer.
func NewAnalyzer(bus eventbus.Bus) *Analyzer {
	return &Analyzer{
		roas:          make(map[string]uint32),
		routes:        make(map[string]*RouteTableEntry),
		updateHistory: make(map[string][]time.Time),
		bus:           bus,
		anomalies:     make([]*BGPAnomaly, 0),
	}
}

// RegisterROA registers an authorized Route Origin Authorization (ROA) mapping.
func (a *Analyzer) RegisterROA(prefix string, originASN uint32) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.roas[prefix] = originASN
}

// ProcessUpdate analyzes an announced or withdrawn BGP route update.
func (a *Analyzer) ProcessUpdate(ctx context.Context, u BGPUpdate) (*BGPAnomaly, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	if u.Timestamp.IsZero() {
		u.Timestamp = now
	}

	if u.IsWithdrawal {
		delete(a.routes, u.Prefix)
		return nil, nil
	}

	// 1. Track update frequency for Route Flapping detection
	timestamps := a.updateHistory[u.Prefix]
	// Clean timestamps older than 60 seconds
	cutoff := now.Add(-60 * time.Second)
	var recent []time.Time
	for _, t := range timestamps {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	recent = append(recent, now)
	a.updateHistory[u.Prefix] = recent

	// Check if update frequency exceeds flapping threshold (>= 5 updates in 60s)
	if len(recent) >= 5 {
		anomaly := &BGPAnomaly{
			ID:              fmt.Sprintf("bgp-flap-%d", now.UnixNano()%1000000),
			Type:            AnomalyRouteFlapping,
			Prefix:          u.Prefix,
			AnnouncedOrigin: u.OriginASN,
			ASPath:          u.ASPath,
			Severity:        "HIGH",
			Description:     fmt.Sprintf("Prefix %s is flapping violently (%d updates in past 60s)", u.Prefix, len(recent)),
			DetectedAt:      now,
		}
		a.recordAnomaly(ctx, anomaly)
		return anomaly, nil
	}

	// Determine origin ASN from ASPath if not explicitly set
	origin := u.OriginASN
	if origin == 0 && len(u.ASPath) > 0 {
		origin = u.ASPath[len(u.ASPath)-1]
	}

	// 2. RPKI Route Origin Validation (ROV)
	expectedOrigin, hasROA := a.roas[u.Prefix]
	rpkiStatus := RPKINotFound
	if hasROA {
		if expectedOrigin == origin {
			rpkiStatus = RPKIValid
		} else {
			rpkiStatus = RPKIInvalid
			anomaly := &BGPAnomaly{
				ID:              fmt.Sprintf("bgp-hijack-%d", now.UnixNano()%1000000),
				Type:            AnomalyPrefixHijack,
				Prefix:          u.Prefix,
				AnnouncedOrigin: origin,
				ExpectedOrigin:  expectedOrigin,
				ASPath:          u.ASPath,
				Severity:        "CRITICAL",
				Description:     fmt.Sprintf("BGP Prefix Hijack detected for %s! Announced by AS%d, but RPKI ROA authorizes AS%d", u.Prefix, origin, expectedOrigin),
				DetectedAt:      now,
			}
			a.recordAnomaly(ctx, anomaly)
			return anomaly, nil
		}
	}

	// 3. Excessive AS-Path Prepending check
	if len(u.ASPath) > 0 {
		maxRepeat := 1
		currentRepeat := 1
		for i := 1; i < len(u.ASPath); i++ {
			if u.ASPath[i] == u.ASPath[i-1] {
				currentRepeat++
				if currentRepeat > maxRepeat {
					maxRepeat = currentRepeat
				}
			} else {
				currentRepeat = 1
			}
		}

		if maxRepeat >= 5 {
			anomaly := &BGPAnomaly{
				ID:              fmt.Sprintf("bgp-prepend-%d", now.UnixNano()%1000000),
				Type:            AnomalyExcessivePrepending,
				Prefix:          u.Prefix,
				AnnouncedOrigin: origin,
				ASPath:          u.ASPath,
				Severity:        "MEDIUM",
				Description:     fmt.Sprintf("Excessive AS-Path prepending detected on %s: ASN repeated %d times in path", u.Prefix, maxRepeat),
				DetectedAt:      now,
			}
			a.recordAnomaly(ctx, anomaly)
			return anomaly, nil
		}
	}

	// Store or update route table entry
	entry, exists := a.routes[u.Prefix]
	if !exists {
		entry = &RouteTableEntry{
			Prefix:      u.Prefix,
			BestPath:    u.ASPath,
			OriginASN:   origin,
			NextHop:     u.NextHop,
			RPKIStatus:  rpkiStatus,
			UpdateCount: 1,
			LastUpdated: now,
		}
		a.routes[u.Prefix] = entry
	} else {
		entry.BestPath = u.ASPath
		entry.OriginASN = origin
		entry.NextHop = u.NextHop
		entry.RPKIStatus = rpkiStatus
		entry.UpdateCount++
		entry.LastUpdated = now
	}

	return nil, nil
}

func (a *Analyzer) recordAnomaly(ctx context.Context, anomaly *BGPAnomaly) {
	a.anomalies = append(a.anomalies, anomaly)
	if a.bus != nil {
		if payload, err := json.Marshal(anomaly); err == nil {
			_ = a.bus.Publish(ctx, TopicBGPAnomaly, payload)
		}
	}
}

// GetRoutes returns a snapshot of active prefix route entries.
func (a *Analyzer) GetRoutes() []*RouteTableEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	res := make([]*RouteTableEntry, 0, len(a.routes))
	for _, r := range a.routes {
		res = append(res, r)
	}
	return res
}

// GetRoute returns an active route entry for a prefix.
func (a *Analyzer) GetRoute(prefix string) (*RouteTableEntry, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	r, exists := a.routes[prefix]
	return r, exists
}

// GetAnomalies returns all recorded BGP anomalies.
func (a *Analyzer) GetAnomalies() []*BGPAnomaly {
	a.mu.RLock()
	defer a.mu.RUnlock()
	res := make([]*BGPAnomaly, len(a.anomalies))
	copy(res, a.anomalies)
	return res
}
