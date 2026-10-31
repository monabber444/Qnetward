package cluster

import (
	"fmt"
	"testing"
)

func TestCellManager_IntraAndInterCellRouting(t *testing.T) {
	cm := NewCellManager("cell-fra-dc1", "eu-central", 10)

	// 1. Register 5 nodes in local cell
	for i := 1; i <= 5; i++ {
		node := CellNode{
			Name:      fmt.Sprintf("fra-node-%d", i),
			CellID:    "cell-fra-dc1",
			Region:    "eu-central",
			Addr:      fmt.Sprintf("10.0.1.%d", i),
			ProbePort: 9000 + i,
		}
		if err := cm.RegisterNode(node); err != nil {
			t.Fatalf("RegisterNode local %d failed: %v", i, err)
		}
	}

	// Promote node 1 as local super-probe
	if err := cm.PromoteSuperProbe("fra-node-1"); err != nil {
		t.Fatalf("PromoteSuperProbe failed: %v", err)
	}

	// 2. Register 3 remote super probes from other regions
	remoteCells := []struct {
		cellID string
		region string
		name   string
	}{
		{"cell-lon-dc1", "eu-west", "lon-super-probe"},
		{"cell-nyc-dc1", "us-east", "nyc-super-probe"},
		{"cell-sin-dc1", "ap-southeast", "sin-super-probe"},
	}

	for _, rc := range remoteCells {
		node := CellNode{
			Name:         rc.name,
			CellID:       rc.cellID,
			Region:       rc.region,
			Addr:         "198.51.100.1",
			ProbePort:    9000,
			IsSuperProbe: true,
		}
		if err := cm.RegisterNode(node); err != nil {
			t.Fatalf("RegisterNode remote failed: %v", err)
		}
	}

	// 3. Verify Intra-Cell Peers: exactly 5 nodes
	intra := cm.GetIntraCellPeers()
	if len(intra) != 5 {
		t.Errorf("Expected 5 intra-cell peers, got %d", len(intra))
	}

	// 4. Verify Inter-Cell Super Probes: 1 local + 3 remote = 4 super probes
	inter := cm.GetInterCellSuperProbes()
	if len(inter) != 4 {
		t.Errorf("Expected 4 super probes across cells, got %d", len(inter))
	}

	// 5. Test Cell capacity bounds (max 10)
	for i := 6; i <= 10; i++ {
		_ = cm.RegisterNode(CellNode{
			Name:   fmt.Sprintf("fra-node-%d", i),
			CellID: "cell-fra-dc1",
		})
	}

	// 11th node should fail with ErrCellCapacityExceeded
	err := cm.RegisterNode(CellNode{
		Name:   "fra-node-11",
		CellID: "cell-fra-dc1",
	})
	if err == nil {
		t.Errorf("Expected ErrCellCapacityExceeded for node 11, got nil")
	}
}
