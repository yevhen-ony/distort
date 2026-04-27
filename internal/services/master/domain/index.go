package domain 

import (
	"context"
	t "dos/internal/common/types"
	"maps"
	"slices"
	"sync"
)

type NodeSet map[t.NodeID]struct{}
type ChunkSet map[t.ChunkID]struct{}

type InMemChunkNodeIndex struct {
	nodeChunks map[t.NodeID]ChunkSet 
	chunkNodes map[t.ChunkID]NodeSet 
	mu sync.RWMutex
}

func NewInMemChunkNodeIndex() *InMemChunkNodeIndex {
	return &InMemChunkNodeIndex{
		nodeChunks: map[t.NodeID]ChunkSet{},
		chunkNodes: map[t.ChunkID]NodeSet{},
	}
}

func (i *InMemChunkNodeIndex) GetNodeChunks(_ context.Context, nodeID t.NodeID) []t.ChunkID  {
	i.mu.RLock()
	defer i.mu.RUnlock()
	
	chunks := i.nodeChunks[nodeID]
	if chunks == nil {
		return []t.ChunkID{} 
	}
	
	result :=  slices.Collect(maps.Keys(chunks))
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

func (i *InMemChunkNodeIndex) AttachChunk(_ context.Context, nodeID t.NodeID, chunkID t.ChunkID) {
	i.mu.Lock()	
	defer i.mu.Unlock()

	nodes := i.chunkNodes[chunkID] 
	if nodes == nil {
		nodes = NodeSet{}
		i.chunkNodes[chunkID] = nodes
	}
	nodes[nodeID] = struct{}{}
	

	chunks := i.nodeChunks[nodeID]
	if chunks == nil {
		chunks = ChunkSet{}
		i.nodeChunks[nodeID] = chunks
	}
	chunks[chunkID] = struct{}{}
	
	return
}

func (i *InMemChunkNodeIndex) DetachNode(_ context.Context, nodeID t.NodeID) {
	i.mu.Lock()	
	defer i.mu.Unlock()

	for chunkID := range i.nodeChunks[nodeID] {
		delete(i.chunkNodes[chunkID], nodeID)
		if len(i.chunkNodes[chunkID]) == 0 {
			delete(i.chunkNodes, chunkID)
		}
	}
	delete(i.nodeChunks, nodeID)
}
