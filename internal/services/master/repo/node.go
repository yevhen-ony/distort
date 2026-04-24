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
)


type InMemNodeRegistry struct {
	nodes map[m.NodeID]*m.Node
	addrs map[string]m.NodeID
	nodeChunks map[m.NodeID]map[m.ChunkID]struct{}
	chunkNodes map[m.ChunkID]map[m.NodeID]struct{}

	mu sync.RWMutex
}

func (r *InMemNodeRegistry) Register(_ context.Context, report m.NodeReport) (m.NodeID, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.addrs[report.Addr]
	if ok {
		return "", m.ErrNodeAddrInUse 
	}

	nid := r.pickNodeID()
	r.addrs[report.Addr] = nid 
	r.nodes[nid] = &m.Node{ ID: nid, Report: report }
	r.nodeChunks[nid] = map[m.ChunkID]struct{}{}
	return nid, nil 
}

func (r *InMemNodeRegistry) Unregister(_ context.Context, nid m.NodeID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	node, ok := r.nodes[nid]
	if !ok {
		return m.ErrNodeNotFound
	}

	r.cleanupNodeRelations(nid)
	delete(r.nodes, nid)
	delete(r.addrs, node.Report.Addr)

	return nil
}

func (r *InMemNodeRegistry) AttachChunk(_ context.Context, nid m.NodeID, cid m.ChunkID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, ok := r.nodes[nid]; !ok {
		return m.ErrNodeNotFound	
	}
	if r.chunkNodes[cid] == nil {
		r.chunkNodes[cid] = map[m.NodeID]struct{}{}
	}

	r.nodeChunks[nid][cid] = struct{}{}
	r.chunkNodes[cid][nid] = struct{}{}

	return nil
}

func (r *InMemNodeRegistry) GetNode(ctx context.Context, nid m.NodeID) (m.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	node, ok := r.nodes[nid]
	if !ok {
		return m.Node{}, m.ErrNodeNotFound	
	}
	return *node, nil
}

func (r *InMemNodeRegistry) GetNodeChunks(_ context.Context, nid m.NodeID) ([]m.ChunkID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	chunks, ok := r.nodeChunks[nid]
	if !ok {
		return nil, m.ErrNodeNotFound
	}
	
	result :=  slices.Collect(maps.Keys(chunks))
	return result, nil
}

func (r *InMemNodeRegistry) GetChunkNodes(_ context.Context, cid m.ChunkID) ([]m.Node, error) {
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
	_ context.Context, query m.CandidateNodesQuery) ([]m.Node, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make([]m.Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		if node.Report.FreeBytes < query.MinFreeBytes {
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


func (r *InMemNodeRegistry) pickNodeID() m.NodeID {
	for {
		id := newNodeID()
		if _, ok := r.nodes[id]; !ok {
			return id	
		}
	}
}

func (r *InMemNodeRegistry) cleanupNodeRelations(nid m.NodeID) {
	chunks := r.nodeChunks[nid]
	for cid := range chunks {
		delete(r.chunkNodes[cid], nid)
		if len(r.chunkNodes[cid]) == 0 {
			delete(r.chunkNodes, cid)
		}
	}
	delete(r.nodeChunks, nid)
}

func newNodeID() m.NodeID {
	var b [16]byte
	rand.Read(b[:])
	return m.NodeID(hex.EncodeToString(b[:]))
}

