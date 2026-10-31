package cluster

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrCellCapacityExceeded = errors.New("cell capacity exceeded max limit (250 nodes)")
	ErrNodeNotFound         = errors.New("node not found in cell")
)

// CellNode describes an edge node mapped to a regional topological cell.
type CellNode struct {
	Name         string `json:"name"`
	CellID       string `json:"cell_id"`
	Region       string `json:"region"`
	Addr         string `json:"addr"`
	ProbePort    int    `json:"probe_port"`
	IsSuperProbe bool   `json:"is_super_probe"`
}

// CellManager implements Hierarchical SWIM (H-SWIM) partitioning to prevent O(N^2) gossip explosions.
type CellManager struct {
	mu            sync.RWMutex
	localCellID   string
	localRegion   string
	maxCellNodes  int
	cellNodes     map[string]*CellNode // Local cell peers
	superProbes   map[string]*CellNode // Remote cell leaders (one per cell)
}

// NewCellManager creates a hierarchical gossip cell manager.
func NewCellManager(localCellID string, localRegion string, maxCellNodes int) *CellManager {
	if maxCellNodes <= 0 {
		maxCellNodes = 250
	}
	return &CellManager{
		localCellID:  localCellID,
		localRegion:  localRegion,
		maxCellNodes: maxCellNodes,
		cellNodes:    make(map[string]*CellNode),
		superProbes:  make(map[string]*CellNode),
	}
}

// RegisterNode assigns a node to either the local cell or remote super-probe registry.
func (cm *CellManager) RegisterNode(node CellNode) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if node.CellID == cm.localCellID {
		if len(cm.cellNodes) >= cm.maxCellNodes {
			return fmt.Errorf("%w (%d max)", ErrCellCapacityExceeded, cm.maxCellNodes)
		}
		cm.cellNodes[node.Name] = &node
		if node.IsSuperProbe {
			cm.superProbes[node.CellID] = &node
		}
		return nil
	}

	// Remote cell node: register as super probe if marked
	if node.IsSuperProbe {
		cm.superProbes[node.CellID] = &node
	}
	return nil
}

// PromoteSuperProbe designates a local node as the inter-cell representative.
func (cm *CellManager) PromoteSuperProbe(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	node, exists := cm.cellNodes[name]
	if !exists {
		return ErrNodeNotFound
	}

	node.IsSuperProbe = true
	cm.superProbes[cm.localCellID] = node
	return nil
}

// GetIntraCellPeers returns all peers within the local cell for high-frequency UDP echo matrix probing.
func (cm *CellManager) GetIntraCellPeers() []*CellNode {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	res := make([]*CellNode, 0, len(cm.cellNodes))
	for _, n := range cm.cellNodes {
		res = append(res, n)
	}
	return res
}

// GetInterCellSuperProbes returns designated super-probes for cross-region WAN probing.
func (cm *CellManager) GetInterCellSuperProbes() []*CellNode {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	res := make([]*CellNode, 0, len(cm.superProbes))
	for _, n := range cm.superProbes {
		res = append(res, n)
	}
	return res
}

// CellCount returns the total number of active cells observed.
func (cm *CellManager) CellCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.superProbes)
}
