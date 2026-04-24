package repo

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"dos/internal/libraries/digest"
	m "dos/internal/services/master"
)

type InMemChunkRepo struct { 
	chunks map[m.ChunkID]*m.Chunk
}

func (r *InMemChunkRepo) Create(_ context.Context) (m.ChunkID, error) {
	id := r.pickChunkID()
	r.chunks[id] = &m.Chunk{ ID: id }
	return id, nil
}

func (r *InMemChunkRepo) pickChunkID() m.ChunkID {
	for {
		id := newChunkID()
		if _, ok := r.chunks[id]; !ok {
			return id	
		}
	}
}

func newChunkID() m.ChunkID {
	var b [16]byte
	rand.Read(b[:])
	return m.ChunkID(hex.EncodeToString(b[:]))
}

func (r *InMemChunkRepo) SetDigest(_ context.Context, id m.ChunkID, digest *digest.Digest) error {
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

func (r *InMemChunkRepo) Get(_ context.Context, id m.ChunkID) (m.Chunk, error) {
	chunk, ok := r.chunks[id]
	if !ok {
		return m.Chunk{}, m.ErrChunkNotFound
	}
	return *chunk.Clone(), nil
}
