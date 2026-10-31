package consensus

import (
	"testing"
	"time"
)

func TestCoordinator_LeaderElectionAndLocks(t *testing.T) {
	peers := []ConsensusNode{
		{ID: "coord-lon-1", Region: "eu-west", Role: RoleFollower},
		{ID: "coord-fra-1", Region: "eu-central", Role: RoleFollower},
	}

	coord := NewCoordinator("coord-nyc-1", "us-east", peers)

	// 1. Follower node should reject lock acquisition and state proposals
	_, err := coord.AcquireLock("bgp:core-fra-01", "coord-nyc-1", 5*time.Second)
	if err != ErrNotLeader {
		t.Errorf("Expected ErrNotLeader when acquiring lock as follower, got %v", err)
	}

	_, err = coord.Propose("topology:version", []byte("v2.4.1"))
	if err != ErrNotLeader {
		t.Errorf("Expected ErrNotLeader when proposing state as follower, got %v", err)
	}

	// 2. Elect as Leader (1 local + 2 followers = 3/3 votes >= 2 quorum)
	if err := coord.ElectAsLeader(2); err != nil {
		t.Fatalf("ElectAsLeader failed: %v", err)
	}

	role, leaderID, term := coord.GetRole()
	if role != RoleLeader || leaderID != "coord-nyc-1" || term != 2 {
		t.Fatalf("Unexpected leader state: role=%s leader=%s term=%d", role, leaderID, term)
	}

	// 3. Leader proposes state machine update
	idx, err := coord.Propose("topology:version", []byte("v2.4.1"))
	if err != nil {
		t.Fatalf("Propose failed on leader: %v", err)
	}
	if idx != 1 {
		t.Errorf("Expected log index 1, got %d", idx)
	}

	val, found := coord.GetState("topology:version")
	if !found || string(val) != "v2.4.1" {
		t.Errorf("Failed to read committed state: found=%v val=%s", found, string(val))
	}

	// 4. Acquire distributed lock for BGP playbook mutation
	lock, err := coord.AcquireLock("bgp:core-fra-01", "playbook-runner-1", 5*time.Second)
	if err != nil {
		t.Fatalf("AcquireLock failed: %v", err)
	}
	if lock.ResourceKey != "bgp:core-fra-01" || lock.HolderNode != "playbook-runner-1" {
		t.Errorf("Unexpected lock fields: %+v", lock)
	}

	// 5. Concurrent acquisition of same resource by another runner must fail
	_, err = coord.AcquireLock("bgp:core-fra-01", "playbook-runner-2", 5*time.Second)
	if err != ErrLockHeld {
		t.Errorf("Expected ErrLockHeld for concurrent lock attempt, got %v", err)
	}

	// 6. Release lock
	if err := coord.ReleaseLock("bgp:core-fra-01", "playbook-runner-1"); err != nil {
		t.Fatalf("ReleaseLock failed: %v", err)
	}

	// 7. Re-acquisition should succeed now
	lock2, err := coord.AcquireLock("bgp:core-fra-01", "playbook-runner-2", 5*time.Second)
	if err != nil {
		t.Fatalf("Re-acquisition failed after release: %v", err)
	}
	if lock2.HolderNode != "playbook-runner-2" {
		t.Errorf("Expected lock holder to be playbook-runner-2")
	}
}
