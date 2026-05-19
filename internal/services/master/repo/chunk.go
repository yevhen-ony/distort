package repo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"dos/internal/common/digest"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
)

type InMemChunkRepo struct {
	chunks map[t.ChunkID]*m.Chunk
	mu     sync.RWMutex
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
	r.chunks[id] = &m.Chunk{
		Meta:          t.ChunkMeta{ID: id},
		ObjectID:      objectID,
		ChunkKey:      chunkKey,
		LastTouchedAt: time.Now(),
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
	if chunk.Meta.Digest == nil {
		chunk.Meta.Digest = digest.Clone()
		chunk.LastTouchedAt = time.Now()
		return nil
	}
	if err := chunk.Meta.Digest.Match(digest); err != nil {
		return err
	}
	chunk.LastTouchedAt = time.Now()
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
		return
	}
	chunk.ReplicaCount++
	chunk.LastTouchedAt = time.Now()
}

func (r *InMemChunkRepo) DecReplication(ctx context.Context, chunkID t.ChunkID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	chunk, ok := r.chunks[chunkID]
	if !ok {
		slog.WarnContext(ctx, "try decrement replication of Chnon-existing chunk", "chunk_id", chunkID)
		return
	}
	if chunk.ReplicaCount == 0 {
		slog.WarnContext(ctx, "try decrement replication below zero", "chunk_id", chunkID)
	}
	chunk.ReplicaCount--
	chunk.LastTouchedAt = time.Now()
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

func (r *InMemChunkRepo) Touch(_ context.Context, chunkID t.ChunkID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	chunk, ok := r.chunks[chunkID]
	if !ok {
		return
	}
	chunk.LastTouchedAt = time.Now()
}

func (r *InMemChunkRepo) ForEach(_ context.Context, fn func(m.Chunk)) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for _, chunk := range r.chunks {
		fn(*chunk.Clone())
	}
}
