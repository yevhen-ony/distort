package repo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"maps"
	"slices"
	"sync"

	m "dos/internal/services/master"
	t "dos/internal/common/types"
)


type InMemNodeRegistry struct {
	nodes map[t.NodeID]*m.Node
	addrs map[string]t.NodeID
	nodeChunks map[t.NodeID]map[t.ChunkID]struct{}
	chunkNodes map[t.ChunkID]map[t.NodeID]struct{}

	mu sync.RWMutex
}

func NewInMemNodeRegistry() *InMemNodeRegistry {
	return &InMemNodeRegistry{
		nodes: map[t.NodeID]*m.Node{},
		addrs: map[string]t.NodeID{},
		nodeChunks: map[t.NodeID]map[t.ChunkID]struct{}{},
		chunkNodes: map[t.ChunkID]map[t.NodeID]struct{}{},
	}
}

func (r *InMemNodeRegistry) Register(_ context.Context, report *t.NodeStats) (t.NodeID, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.addrs[report.Addr]
	if ok {
		return "", m.ErrNodeAddrInUse 
	}

	nid := r.pickNodeID()
	r.addrs[report.Addr] = nid 
	r.nodes[nid] = &m.Node{ ID: nid, Stats: *report }
	r.nodeChunks[nid] = map[t.ChunkID]struct{}{}
	return nid, nil 
}

func (r *InMemNodeRegistry) Unregister(_ context.Context, nid t.NodeID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	node, ok := r.nodes[nid]
	if !ok {
		return m.ErrNodeNotFound
	}

	r.cleanupNodeRelations(nid)
	delete(r.nodes, nid)
	delete(r.addrs, node.Stats.Addr)

	return nil
}

func (r *InMemNodeRegistry) AttachChunk(_ context.Context, nid t.NodeID, cid t.ChunkID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, ok := r.nodes[nid]; !ok {
		return m.ErrNodeNotFound	
	}
	if r.chunkNodes[cid] == nil {
		r.chunkNodes[cid] = map[t.NodeID]struct{}{}
	}

	r.nodeChunks[nid][cid] = struct{}{}
	r.chunkNodes[cid][nid] = struct{}{}

	return nil
}

func (r *InMemNodeRegistry) GetNode(ctx context.Context, nid t.NodeID) (m.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	node, ok := r.nodes[nid]
	if !ok {
		return m.Node{}, m.ErrNodeNotFound	
	}
	return *node, nil
}

func (r *InMemNodeRegistry) GetNodeChunks(_ context.Context, nid t.NodeID) ([]t.ChunkID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	chunks, ok := r.nodeChunks[nid]
	if !ok {
		return nil, m.ErrNodeNotFound
	}
	
	result :=  slices.Collect(maps.Keys(chunks))
	return result, nil
}

func (r *InMemNodeRegistry) GetChunkNodes(_ context.Context, cid t.ChunkID) ([]m.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodeIDs := r.chunkNodes[cid]
	
	nodes := make([]m.Node, 0, len(nodeIDs))
	for nodeID := range nodeIDs {
		node, ok := r.nodes[nodeID]
		if !ok {
			return nil, fmt.Errorf("get node %s: %w", nodeID, m.ErrNodeNotFound)
		}
		nodes = append(nodes, *node)	
	}
	return nodes, nil
}

func (r *InMemNodeRegistry) GetCandidateNodes(
	_ context.Context, query *m.CandidateNodesQuery) ([]m.Node, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make([]m.Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		if node.Stats.FreeBytes < query.MinFreeBytes {
			continue
		}
		if query.ExcludeChunk != "" {
			_, ok := r.nodeChunks[node.ID][query.ExcludeChunk]
			if ok {
				continue
			}
		}
		result = append(result, *node)
	}
	return result, nil	
}


func (r *InMemNodeRegistry) pickNodeID() t.NodeID {
	for {
		id := newNodeID()
		if _, ok := r.nodes[id]; !ok {
			return id	
		}
	}
}

func (r *InMemNodeRegistry) cleanupNodeRelations(nid t.NodeID) {
	chunks := r.nodeChunks[nid]
	for cid := range chunks {
		delete(r.chunkNodes[cid], nid)
		if len(r.chunkNodes[cid]) == 0 {
			delete(r.chunkNodes, cid)
		}
	}
	delete(r.nodeChunks, nid)
}

func newNodeID() t.NodeID {
	var b [16]byte
	rand.Read(b[:])
	return t.NodeID(hex.EncodeToString(b[:]))
}

