package bgp

import (
	"time"
)

// AnomalyType categorizes detected BGP routing disruptions.
type AnomalyType string

const (
	AnomalyPrefixHijack        AnomalyType = "PREFIX_HIJACK"
	AnomalyRouteLeak           AnomalyType = "ROUTE_LEAK"
	AnomalyExcessivePrepending AnomalyType = "EXCESSIVE_PREPENDING"
	AnomalyRouteFlapping       AnomalyType = "ROUTE_FLAPPING"
)

// RPKIValidationStatus represents the RPKI Route Origin Validation state.
type RPKIValidationStatus string

const (
	RPKIValid    RPKIValidationStatus = "VALID"
	RPKIInvalid  RPKIValidationStatus = "INVALID"
	RPKINotFound RPKIValidationStatus = "NOT_FOUND"
)

// BGPUpdate represents an announced or withdrawn routing prefix update.
type BGPUpdate struct {
	ID           string    `json:"id"`
	Prefix       string    `json:"prefix"`        // e.g. "198.51.100.0/24"
	PeerASN      uint32    `json:"peer_asn"`      // Collector/Peer AS
	PeerIP       string    `json:"peer_ip"`       // BGP peer router IP
	ASPath       []uint32  `json:"as_path"`       // e.g. [174, 3356, 64500]
	OriginASN    uint32    `json:"origin_asn"`    // Last AS in the path
	NextHop      string    `json:"next_hop"`
	Communities  []string  `json:"communities"`
	IsWithdrawal bool      `json:"is_withdrawal"`
	Timestamp    time.Time `json:"timestamp"`
}

// BGPAnomaly details a suspicious or illegitimate BGP route announcement.
type BGPAnomaly struct {
	ID              string      `json:"id"`
	Type            AnomalyType `json:"type"`
	Prefix          string      `json:"prefix"`
	AnnouncedOrigin uint32      `json:"announced_origin"`
	ExpectedOrigin  uint32      `json:"expected_origin"`
	ASPath          []uint32    `json:"as_path"`
	Severity        string      `json:"severity"` // "CRITICAL", "HIGH", "MEDIUM", "LOW"
	Description     string      `json:"description"`
	DetectedAt      time.Time   `json:"detected_at"`
}

// ASInfo represents autonomous system metadata and Tier designation.
type ASInfo struct {
	ASN     uint32 `json:"asn"`
	Name    string `json:"name"`
	Country string `json:"country"`
	Tier    string `json:"tier"` // "Tier-1", "Tier-2", "Content-Provider", "Enterprise"
}

// RouteTableEntry represents the active best path for a prefix.
type RouteTableEntry struct {
	Prefix      string               `json:"prefix"`
	BestPath    []uint32             `json:"best_path"`
	OriginASN   uint32               `json:"origin_asn"`
	NextHop     string               `json:"next_hop"`
	RPKIStatus  RPKIValidationStatus `json:"rpki_status"`
	UpdateCount int                  `json:"update_count"`
	LastUpdated time.Time            `json:"last_updated"`
}
