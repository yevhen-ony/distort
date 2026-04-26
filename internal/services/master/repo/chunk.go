package repo

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"dos/internal/common/digest"
	m "dos/internal/services/master"
	t "dos/internal/common/types"
)

type InMemChunkRepo struct { 
	chunks map[t.ChunkID]*m.Chunk
}

func MakeInMemChunkRepo() *InMemChunkRepo {
	return &InMemChunkRepo{
		chunks: map[t.ChunkID]*m.Chunk{},
	}
}

func (r *InMemChunkRepo) Create(_ context.Context) (t.ChunkID, error) {
	id := r.pickChunkID()
	r.chunks[id] = &m.Chunk{ ID: id }
	return id, nil
}

func (r *InMemChunkRepo) pickChunkID() t.ChunkID {
	for {
		id := newChunkID()
		if _, ok := r.chunks[id]; !ok {
			return id	
		}
	}
}

func newChunkID() t.ChunkID {
	var b [16]byte
	rand.Read(b[:])
	return t.ChunkID(hex.EncodeToString(b[:]))
}

func (r *InMemChunkRepo) SetDigest(_ context.Context, id t.ChunkID, digest digest.Digest) error {
	chunk, ok := r.chunks[id]	
	if !ok {
		return m.ErrChunkNotFound 
	}
	if chunk.Digest == nil {
		chunk.Digest = digest.Clone() 
		return nil
	}
	if !chunk.Digest.Equal(digest) {
		return m.ErrChunkDigestConflict
	}
	return nil
}

func (r *InMemChunkRepo) Get(_ context.Context, id t.ChunkID) (m.Chunk, error) {
	chunk, ok := r.chunks[id]
	if !ok {
		return m.Chunk{}, m.ErrChunkNotFound
	}
	return *chunk.Clone(), nil
}
