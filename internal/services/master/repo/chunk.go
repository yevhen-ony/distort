package repo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"

	"dos/internal/common/digest"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
)

type InMemChunkRepo struct { 
	chunks map[t.ChunkID]*m.Chunk
	mu sync.RWMutex
}

func NewInMemChunkRepo() *InMemChunkRepo {
	return &InMemChunkRepo{
		chunks: map[t.ChunkID]*m.Chunk{},
	}
}

func (r *InMemChunkRepo) Create(
	_ context.Context, id t.ChunkID, objectID t.ObjectID, chunkKey t.ChunkKey,
) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.chunks[id]; ok {
		return m.ErrChunkExists
	}
	chunkMeta := t.ChunkMeta{ID: id}
	r.chunks[id] = &m.Chunk{
		ChunkMeta: chunkMeta,
		ObjectID: objectID,
		ChunkKey: chunkKey,
	}
	return nil
}

func (r *InMemChunkRepo) NewChunkID() t.ChunkID {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for {
		id := genChunkID()
		if _, ok := r.chunks[id]; !ok {
			return id	
		}
	}
}

func genChunkID() t.ChunkID {
	var b [8]byte
	rand.Read(b[:])
	return t.ChunkID(hex.EncodeToString(b[:]))
}

func (r *InMemChunkRepo) SetDigest(_ context.Context, id t.ChunkID, digest *digest.Digest) error {
	r.mu.Lock()
	defer r.mu.Unlock() 

	if digest == nil {
		return fmt.Errorf("set nil digest")
	}
	chunk, ok := r.chunks[id]	
	if !ok {
		return m.ErrChunkNotFound 
	}
	if chunk.Digest == nil {
		chunk.Digest = digest.Clone() 
		return nil
	}
	if err := chunk.Digest.Match(digest); err != nil {
		return err 
	}
	return nil
}

func (r *InMemChunkRepo) Get(_ context.Context, id t.ChunkID) (m.Chunk, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chunk, ok := r.chunks[id]
	if !ok {
		return m.Chunk{}, m.ErrChunkNotFound
	}
	return *chunk.Clone(), nil
}

func (r *InMemChunkRepo) IncReplication(ctx context.Context, chunkID t.ChunkID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	chunk, ok := r.chunks[chunkID]	
	if !ok {
		slog.WarnContext(ctx, "try increment replication of non-existing chunk", "chunk_id", chunkID)
	}
	chunk.ReplicaCount++
}

func (r *InMemChunkRepo) DecReplication(ctx context.Context, chunkID t.ChunkID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	chunk, ok := r.chunks[chunkID]
	if !ok {
		slog.WarnContext(ctx, "try decrement replication of non-existing chunk", "chunk_id", chunkID)
	}
	if chunk.ReplicaCount == 0 {
		slog.WarnContext(ctx, "try decrement replication below zero", "chunk_id", chunkID)
	}
	chunk.ReplicaCount--
}

func (r *InMemChunkRepo) List(_ context.Context) []m.Chunk {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]m.Chunk, 0, len(r.chunks))
	for _, chunk := range r.chunks {
		result = append(result, *chunk.Clone())
	}
	return result 
}

func (r *InMemChunkRepo) DeleteWithNoReplicas(_ context.Context, chunkID t.ChunkID) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	chunk, ok := r.chunks[chunkID]
	if !ok {
		return true 
	}

	if chunk.ReplicaCount > 0 {
		return false
	}

	delete(r.chunks, chunkID)
	return true
}
