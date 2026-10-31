package consensus

import (
	"fmt"
	"sync"
	"time"
)

// Coordinator manages distributed state machine replication, leader election, and mutual exclusion locks.
type Coordinator struct {
	mu          sync.RWMutex
	localID     string
	region      string
	role        NodeRole
	currentTerm uint64
	leaderID    string
	peers       map[string]*ConsensusNode
	locks       map[string]*DistributedLock
	log         []*LogEntry
	stateStore  map[string][]byte
}

// NewCoordinator creates a new consensus coordinator for multi-region high availability.
func NewCoordinator(localID string, region string, peers []ConsensusNode) *Coordinator {
	peerMap := make(map[string]*ConsensusNode)
	for _, p := range peers {
		cp := p
		peerMap[p.ID] = &cp
	}

	return &Coordinator{
		localID:     localID,
		region:      region,
		role:        RoleFollower,
		currentTerm: 1,
		peers:       peerMap,
		locks:       make(map[string]*DistributedLock),
		log:         make([]*LogEntry, 0),
		stateStore:  make(map[string][]byte),
	}
}

// StepDown transitions the coordinator to a Follower role.
func (c *Coordinator) StepDown(term uint64, leaderID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.role = RoleFollower
	if term > c.currentTerm {
		c.currentTerm = term
	}
	c.leaderID = leaderID
}

// ElectAsLeader promotes the node to Leader after achieving quorum votes.
func (c *Coordinator) ElectAsLeader(term uint64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate cluster quorum: (Total nodes / 2) + 1
	totalNodes := len(c.peers) + 1
	quorumRequired := (totalNodes / 2) + 1

	// In test/simulation mode, verify active peers
	activeVotes := 1 // Local vote
	for _, p := range c.peers {
		if p.Role != RoleCandidate {
			activeVotes++
		}
	}

	if activeVotes < quorumRequired {
		return fmt.Errorf("%w: got %d votes, need %d", ErrQuorumNotMet, activeVotes, quorumRequired)
	}

	c.role = RoleLeader
	c.currentTerm = term
	c.leaderID = c.localID
	return nil
}

// AcquireLock attempts to secure an exclusive cluster-wide lock on a resource.
func (c *Coordinator) AcquireLock(resourceKey string, holderNode string, ttl time.Duration) (*DistributedLock, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.role != RoleLeader {
		return nil, ErrNotLeader
	}

	now := time.Now()
	existing, held := c.locks[resourceKey]
	if held && now.Before(existing.ExpiresAt) && existing.HolderNode != holderNode {
		return nil, ErrLockHeld
	}

	lock := &DistributedLock{
		ResourceKey: resourceKey,
		HolderNode:  holderNode,
		AcquiredAt:  now,
		ExpiresAt:   now.Add(ttl),
	}
	c.locks[resourceKey] = lock
	return lock, nil
}

// ReleaseLock releases an exclusive lock if held by the requesting node.
func (c *Coordinator) ReleaseLock(resourceKey string, holderNode string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.role != RoleLeader {
		return ErrNotLeader
	}

	existing, held := c.locks[resourceKey]
	if !held {
		return nil
	}

	if existing.HolderNode != holderNode {
		return fmt.Errorf("cannot release lock held by node %s", existing.HolderNode)
	}

	delete(c.locks, resourceKey)
	return nil
}

// Propose appends an entry to the replicated state log and updates the local state machine.
func (c *Coordinator) Propose(key string, value []byte) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.role != RoleLeader {
		return 0, ErrNotLeader
	}

	entryIndex := uint64(len(c.log) + 1)
	entry := &LogEntry{
		Index: entryIndex,
		Term:  c.currentTerm,
		Key:   key,
		Value: value,
	}

	c.log = append(c.log, entry)
	c.stateStore[key] = value
	return entryIndex, nil
}

// GetState retrieves a committed key-value from the state machine.
func (c *Coordinator) GetState(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.stateStore[key]
	return val, ok
}

// GetRole returns current node role and leader ID.
func (c *Coordinator) GetRole() (NodeRole, string, uint64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.role, c.leaderID, c.currentTerm
}

// GetActiveLocks returns all currently held locks.
func (c *Coordinator) GetActiveLocks() []*DistributedLock {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now()
	var res []*DistributedLock
	for _, l := range c.locks {
		if now.Before(l.ExpiresAt) {
			res = append(res, l)
		}
	}
	return res
}
