package repo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
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
	_ context.Context, chunkID t.ChunkID, objectSlot t.ObjectSlot) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.chunks[chunkID]; ok {
		return m.ErrChunkExists
	}
	r.chunks[chunkID] = &m.Chunk{
		Meta:          t.ChunkMeta{ID: chunkID},
		Slot:          objectSlot,
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

func (r *InMemChunkRepo) SetDigest(_ context.Context, id t.ChunkID, digest digest.Digest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	chunk, ok := r.chunks[id]
	if !ok {
		return m.ErrChunkNotFound
	}
	if chunk.ReplicaCount == 0 {
		chunk.Meta.Digest = digest.Clone()
		chunk.LastTouchedAt = time.Now()
		return nil
	}
	if err := chunk.Meta.Digest.Match(&digest); err != nil {
		return err
	}
	chunk.LastTouchedAt = time.Now()
	return nil
}

func (r *InMemChunkRepo) GetDigest(_ context.Context, id t.ChunkID) (digest.Digest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chunk, ok := r.chunks[id]
	if !ok {
		return digest.Digest{}, m.ErrChunkNotFound
	}
	return chunk.Meta.Digest.Clone(), nil
}

func (r *InMemChunkRepo) Exists(_ context.Context, chunkID t.ChunkID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.chunks[chunkID]
	return ok, nil
}

func (r *InMemChunkRepo) Get(_ context.Context, chunkID t.ChunkID) (m.Chunk, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chunk, ok := r.chunks[chunkID]
	if !ok {
		return m.Chunk{}, m.ErrChunkNotFound
	}
	return *chunk.Clone(), nil
}

func (r *InMemChunkRepo) IncReplicaCount(ctx context.Context, chunkID t.ChunkID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	chunk, ok := r.chunks[chunkID]
	if !ok {
		slog.WarnContext(ctx, "try increment replication of non-existing chunk", "chunk_id", chunkID)
		return m.ErrChunkNotFound
	}
	chunk.ReplicaCount++
	chunk.LastTouchedAt = time.Now()
	return nil
}

func (r *InMemChunkRepo) DecReplicaCount(ctx context.Context, chunkID t.ChunkID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	chunk, ok := r.chunks[chunkID]
	if !ok {
		slog.WarnContext(ctx, "try decrement replication of non-existing chunk", "chunk_id", chunkID)
		return m.ErrChunkNotFound
	}
	if chunk.ReplicaCount == 0 {
		slog.WarnContext(ctx, "try decrement replication below zero", "chunk_id", chunkID)
		return errors.New("decrement replicas below zero")
	}
	chunk.ReplicaCount--
	chunk.LastTouchedAt = time.Now()
	return nil
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

func (r *InMemChunkRepo) Drop(_ context.Context, chunkID t.ChunkID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.chunks, chunkID)
}

func (r *InMemChunkRepo) Delete(_ context.Context, chunkID t.ChunkID) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	chunk, ok := r.chunks[chunkID]
	if !ok {
		return false, nil
	}

	if chunk.ReplicaCount > 0 {
		return false, errors.New("delete not empty chunk")
	}

	delete(r.chunks, chunkID)
	return true, nil
}

func (r *InMemChunkRepo) Touch(_ context.Context, chunkID t.ChunkID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	chunk, ok := r.chunks[chunkID]
	if !ok {
		return m.ErrChunkNotFound
	}
	chunk.LastTouchedAt = time.Now()
	return nil
}
