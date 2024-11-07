package server

import (
	"fmt"
	"net/rpc"
	"sync"
	"time"
)

type NodeState string

const (
	NodeStateHealthy  NodeState = "healthy"
	NodeStateDegraded NodeState = "degraded"
	NodeStateDown     NodeState = "down"
)

type NodeInfo struct {
	ID       string      `json:"id"`
	RPCAddr  string      `json:"rpc_addr"`
	RespAddr string      `json:"resp_addr"`
	State    NodeState   `json:"state"`
	LastSeen time.Time   `json:"last_seen"`
	IsLeader bool        `json:"is_leader"`
	Stats    ServerStats `json:"stats"`
}

type ClusterInfo struct {
	mu           sync.RWMutex
	nodes        map[string]*NodeInfo
	localNode    *NodeInfo
	server       *Server
	healthTicker *time.Ticker
	stopCh       chan struct{}
}

func (s *Server) initCluster(rpcAddr, respAddr string) error {
	nodeID := fmt.Sprintf("node-%s", rpcAddr)

	s.cluster = &ClusterInfo{
		nodes: make(map[string]*NodeInfo),
		localNode: &NodeInfo{
			ID:       nodeID,
			RPCAddr:  rpcAddr,
			RespAddr: respAddr,
			State:    NodeStateHealthy,
			LastSeen: time.Now(),
			Stats:    ServerStats{},
		},
		server:       s,
		healthTicker: time.NewTicker(5 * time.Second),
		stopCh:       make(chan struct{}),
	}

	// Register local node
	s.cluster.nodes[nodeID] = s.cluster.localNode

	// Start health check routine
	go s.cluster.healthCheckLoop()

	return nil
}

func (s *Server) JoinCluster(knownAddr string) error {
	client, err := rpc.Dial("tcp", knownAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to cluster at %s: %w", knownAddr, err)
	}
	defer client.Close()

	var nodes map[string]*NodeInfo
	if err := client.Call("Store.GetClusterNodes", struct{}{}, &nodes); err != nil {
		return fmt.Errorf("failed to get cluster nodes: %w", err)
	}

	s.cluster.mu.Lock()
	for id, node := range nodes {
		if id != s.cluster.localNode.ID {
			s.cluster.nodes[id] = node
			// Add node's RESP server as a replica
			if err := s.addNodeAsReplica(node); err != nil {
				fmt.Printf("[warning] failed to add replica for node %s: %v\n", node.RPCAddr, err)
			}
		}
	}
	s.cluster.mu.Unlock()

	// Announce ourselves to all nodes
	return s.cluster.announceToCluster()
}

func (ci *ClusterInfo) announceToCluster() error {
	ci.mu.RLock()
	nodes := make(map[string]*NodeInfo)
	for k, v := range ci.nodes {
		nodes[k] = v
	}
	ci.mu.RUnlock()

	for _, node := range nodes {
		if node.ID == ci.localNode.ID {
			continue
		}

		client, err := rpc.Dial("tcp", node.RPCAddr)
		if err != nil {
			fmt.Printf("[warning] Failed to connect to node %s: %v\n", node.RPCAddr, err)
			continue
		}

		var reply struct{}
		if err := client.Call("Store.RegisterNode", ci.localNode, &reply); err != nil {
			client.Close()
			fmt.Printf("[warning] Failed to register with node %s: %v\n", node.RPCAddr, err)
			continue
		}
		client.Close()
	}

	return nil
}

// RegisterNode handles a new node joining the cluster
func (s *Server) RegisterNode(node *NodeInfo, reply *struct{}) error {
	s.cluster.mu.Lock()
	defer s.cluster.mu.Unlock()

	node.LastSeen = time.Now()
	s.cluster.nodes[node.ID] = node

	// Add node's RESP server as a replica
	if err := s.addNodeAsReplica(node); err != nil {
		return fmt.Errorf("failed to add replica: %w", err)
	}

	return nil
}

func (s *Server) GetClusterNodes(args struct{}, reply *map[string]*NodeInfo) error {
	s.cluster.mu.RLock()
	defer s.cluster.mu.RUnlock()

	*reply = make(map[string]*NodeInfo)
	for k, v := range s.cluster.nodes {
		(*reply)[k] = v
	}

	return nil
}

func (ci *ClusterInfo) updateNodeStats(stats ServerStats) {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	ci.localNode.Stats = stats
	ci.localNode.LastSeen = time.Now()
}

func (ci *ClusterInfo) healthCheckLoop() {
	for {
		select {
		case <-ci.stopCh:
			return
		case <-ci.healthTicker.C:
			ci.checkNodesHealth()
			// Update our stats
			ci.updateNodeStats(ci.server.Stats())
		}
	}
}

func (ci *ClusterInfo) checkNodesHealth() {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	now := time.Now()
	for id, node := range ci.nodes {
		if id == ci.localNode.ID {
			continue
		}

		// Check last seen timestamp
		oldState := node.State
		if now.Sub(node.LastSeen) > 15*time.Second {
			node.State = NodeStateDown
		} else if now.Sub(node.LastSeen) > 10*time.Second {
			node.State = NodeStateDegraded
		} else {
			node.State = NodeStateHealthy
		}

		// If node went down, remove its replica
		if oldState != NodeStateDown && node.State == NodeStateDown {
			if err := ci.server.removeNodeReplica(node); err != nil {
				fmt.Printf("[warning] failed to remove replica for down node %s: %v\n", node.RPCAddr, err)
			}
		}
	}
}

func (ci *ClusterInfo) stopCluster() {
	close(ci.stopCh)
	ci.healthTicker.Stop()
}
