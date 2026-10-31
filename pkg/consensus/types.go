package consensus

import (
	"errors"
	"time"
)

var (
	ErrNotLeader       = errors.New("node is not the elected cluster leader")
	ErrLockHeld        = errors.New("distributed lock is already held by another coordinator")
	ErrQuorumNotMet    = errors.New("consensus quorum not achieved across regions")
	ErrNodeUnreachable = errors.New("remote consensus peer unreachable")
)

// NodeRole represents the consensus state of a control plane coordinator node.
type NodeRole string

const (
	RoleLeader    NodeRole = "LEADER"
	RoleFollower  NodeRole = "FOLLOWER"
	RoleCandidate NodeRole = "CANDIDATE"
)

// ConsensusNode describes a control plane node participating in the Raft quorum.
type ConsensusNode struct {
	ID         string    `json:"id"`
	Region     string    `json:"region"`
	RPCAddress string    `json:"rpc_address"`
	Role       NodeRole  `json:"role"`
	Term       uint64    `json:"term"`
	LastSeen   time.Time `json:"last_seen"`
}

// DistributedLock represents a cluster-wide mutual exclusion lock for playbook or topology mutations.
type DistributedLock struct {
	ResourceKey string    `json:"resource_key"` // e.g. "bgp:fra-core-01", "port:te0/1/1"
	HolderNode  string    `json:"holder_node"`
	AcquiredAt  time.Time `json:"acquired_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// LogEntry encapsulates a replicated state machine mutation.
type LogEntry struct {
	Index uint64 `json:"index"`
	Term  uint64 `json:"term"`
	Key   string `json:"key"`
	Value []byte `json:"value"`
}
