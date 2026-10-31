package topology

import (
	"sync"

	"netwatcher/drivers/sdk"
)

type NodeType string

const (
	NodeCoreRouter   NodeType = "core_router"
	NodeDistSwitch   NodeType = "dist_switch"
	NodeAccessSwitch NodeType = "access_switch"
	NodeHost         NodeType = "host"
	NodeServer       NodeType = "server"
	NodePDU          NodeType = "pdu"
)

type NodeStatus string

const (
	StatusUp       NodeStatus = "up"
	StatusDegraded NodeStatus = "degraded"
	StatusDown     NodeStatus = "down"
	StatusUnknown  NodeStatus = "unknown"
)

type Node struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     NodeType          `json:"type"`
	Status   NodeStatus        `json:"status"`
	TenantID int64             `json:"tenant_id"`
	IP       string            `json:"ip"`
	Labels   map[string]string `json:"labels"`
}

type Edge struct {
	SourceID   string `json:"source_id"` // Upstream parent
	TargetID   string `json:"target_id"` // Downstream child
	LocalPort  string `json:"local_port"`
	RemotePort string `json:"remote_port"`
	LinkType   string `json:"link_type"` // "lldp", "cdp", "gateway", "static"
}

// Graph represents a directed acyclic network dependency topology.
// Edges point from Upstream (Parent) -> Downstream (Child).
type Graph struct {
	mu          sync.RWMutex
	nodes       map[string]*Node
	parents     map[string][]string // childID -> []parentID
	children    map[string][]string // parentID -> []childID
	edgesByPair map[string]*Edge
}

// NewGraph creates an empty topology graph.
func NewGraph() *Graph {
	return &Graph{
		nodes:       make(map[string]*Node),
		parents:     make(map[string][]string),
		children:    make(map[string][]string),
		edgesByPair: make(map[string]*Edge),
	}
}

// AddNode adds or updates a network node.
func (g *Graph) AddNode(node *Node) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nodes[node.ID] = node
}

// GetNode retrieves a node by its ID.
func (g *Graph) GetNode(id string) (*Node, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	n, ok := g.nodes[id]
	return n, ok
}

// AddEdge registers a dependency from upstream (source) to downstream (target).
func (g *Graph) AddEdge(edge Edge) {
	g.mu.Lock()
	defer g.mu.Unlock()

	pairKey := edge.SourceID + "->" + edge.TargetID
	g.edgesByPair[pairKey] = &edge

	// Append parent if not present
	if !contains(g.parents[edge.TargetID], edge.SourceID) {
		g.parents[edge.TargetID] = append(g.parents[edge.TargetID], edge.SourceID)
	}

	// Append child if not present
	if !contains(g.children[edge.SourceID], edge.TargetID) {
		g.children[edge.SourceID] = append(g.children[edge.SourceID], edge.TargetID)
	}
}

// GetChildren returns the direct downstream children of a node.
func (g *Graph) GetChildren(nodeID string) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	childIDs := g.children[nodeID]
	res := make([]*Node, 0, len(childIDs))
	for _, cid := range childIDs {
		if n, ok := g.nodes[cid]; ok {
			res = append(res, n)
		}
	}
	return res
}

// GetParents returns the direct upstream parents of a node.
func (g *Graph) GetParents(nodeID string) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	parentIDs := g.parents[nodeID]
	res := make([]*Node, 0, len(parentIDs))
	for _, pid := range parentIDs {
		if n, ok := g.nodes[pid]; ok {
			res = append(res, n)
		}
	}
	return res
}

// GetDownstreamDescendants returns all transitive downstream nodes reachable from nodeID.
func (g *Graph) GetDownstreamDescendants(nodeID string) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	visited := make(map[string]bool)
	var queue []string
	queue = append(queue, g.children[nodeID]...)

	var descendants []*Node
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if visited[curr] {
			continue
		}
		visited[curr] = true

		if n, ok := g.nodes[curr]; ok {
			descendants = append(descendants, n)
		}
		queue = append(queue, g.children[curr]...)
	}

	return descendants
}

// GetUpstreamAncestors returns all transitive upstream nodes that nodeID depends on.
func (g *Graph) GetUpstreamAncestors(nodeID string) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	visited := make(map[string]bool)
	var queue []string
	queue = append(queue, g.parents[nodeID]...)

	var ancestors []*Node
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if visited[curr] {
			continue
		}
		visited[curr] = true

		if n, ok := g.nodes[curr]; ok {
			ancestors = append(ancestors, n)
		}
		queue = append(queue, g.parents[curr]...)
	}

	return ancestors
}

// BuildFromLLDP ingests discovered LLDP neighbor records and automatically constructs edges.
func (g *Graph) BuildFromLLDP(localNodeID string, neighbors []sdk.LLDPNeighbor) {
	for _, n := range neighbors {
		remoteID := n.RemoteSystemName
		if remoteID == "" {
			remoteID = n.RemoteChassisID
		}
		if remoteID == "" {
			continue
		}

		// Ensure remote node exists in graph
		if _, ok := g.GetNode(remoteID); !ok {
			g.AddNode(&Node{
				ID:     remoteID,
				Name:   n.RemoteSystemName,
				Type:   NodeAccessSwitch,
				Status: StatusUp,
				IP:     n.RemoteManagementIP,
			})
		}

		// Local node is upstream of discovered neighbor
		g.AddEdge(Edge{
			SourceID:   localNodeID,
			TargetID:   remoteID,
			LocalPort:  n.LocalPort,
			RemotePort: n.RemotePort,
			LinkType:   "lldp",
		})
	}
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
