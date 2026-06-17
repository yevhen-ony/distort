package repo

import (
	"context"
	"maps"
	"slices"
	"sync"

	t "dos/internal/common/types"
)

type NodeSet map[t.NodeID]struct{}
type ChunkSet map[t.ChunkID]struct{}

type InMemChunkNodeIndex struct {
	nodeChunks map[t.NodeID]ChunkSet
	chunkNodes map[t.ChunkID]NodeSet
	mu         sync.RWMutex
}

func NewInMemChunkNodeIndex() *InMemChunkNodeIndex {
	return &InMemChunkNodeIndex{
		nodeChunks: map[t.NodeID]ChunkSet{},
		chunkNodes: map[t.ChunkID]NodeSet{},
	}
}

func (i *InMemChunkNodeIndex) GetNodeChunks(_ context.Context, nodeID t.NodeID) []t.ChunkID {
	i.mu.RLock()
	defer i.mu.RUnlock()

	chunks := i.nodeChunks[nodeID]
	if chunks == nil {
		return []t.ChunkID{}
	}

	result := slices.Collect(maps.Keys(chunks))
	return result
}

func (i *InMemChunkNodeIndex) GetChunkNodes(_ context.Context, chunkID t.ChunkID) []t.NodeID {
	i.mu.RLock()
	defer i.mu.RUnlock()

	nodes := i.chunkNodes[chunkID]
	if nodes == nil {
		return []t.NodeID{}
	}

	result := slices.Collect(maps.Keys(nodes))
	return result
}

func (i *InMemChunkNodeIndex) AttachChunk(_ context.Context, nodeID t.NodeID, chunkID t.ChunkID) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	nodes := i.chunkNodes[chunkID]
	if nodes == nil {
		nodes = NodeSet{}
		i.chunkNodes[chunkID] = nodes
	}
	if _, ok := nodes[nodeID]; ok {
		return false
	}

	nodes[nodeID] = struct{}{}

	chunks := i.nodeChunks[nodeID]
	if chunks == nil {
		chunks = ChunkSet{}
		i.nodeChunks[nodeID] = chunks
	}
	chunks[chunkID] = struct{}{}

	return true
}

func (i *InMemChunkNodeIndex) DetachNode(_ context.Context, nodeID t.NodeID) {
	i.mu.Lock()
	defer i.mu.Unlock()

	chunks := slices.Collect(maps.Keys(i.nodeChunks[nodeID]))
	for _, chunkID := range chunks {
		i.detachChunk(nodeID, chunkID)
	}
}

func (i *InMemChunkNodeIndex) DetachChunk(_ context.Context, nodeID t.NodeID, chunkID t.ChunkID) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	if _, ok := i.chunkNodes[chunkID][nodeID]; !ok {
		return false
	}

	i.detachChunk(nodeID, chunkID)
	return true
}

func (i *InMemChunkNodeIndex) detachChunk(nodeID t.NodeID, chunkID t.ChunkID) {
	delete(i.chunkNodes[chunkID], nodeID)
	if len(i.chunkNodes[chunkID]) == 0 {
		delete(i.chunkNodes, chunkID)
	}
	delete(i.nodeChunks[nodeID], chunkID)
	if len(i.nodeChunks[nodeID]) == 0 {
		delete(i.nodeChunks, nodeID)
	}
}
