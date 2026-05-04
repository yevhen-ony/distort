package repo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"dos/internal/common/digest"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
)

type InMemChunkRepo struct { 
	chunks map[t.ChunkID]*m.Chunk
}

func NewInMemChunkRepo() *InMemChunkRepo {
	return &InMemChunkRepo{
		chunks: map[t.ChunkID]*m.Chunk{},
	}
}

func (r *InMemChunkRepo) Create(_ context.Context, id t.ChunkID) error {
	if _, ok := r.chunks[id]; ok {
		return m.ErrChunkExists
	}
	chunkMeta := t.ChunkMeta{ID: id}
	r.chunks[id] = &m.Chunk{ChunkMeta: chunkMeta}
	return nil
}

func (r *InMemChunkRepo) NewChunkID() t.ChunkID {
	for {
		id := genChunkID()
		if _, ok := r.chunks[id]; !ok {
			return id	
		}
	}
}

func genChunkID() t.ChunkID {
	var b [16]byte
	rand.Read(b[:])
	return t.ChunkID(hex.EncodeToString(b[:]))
}

func (r *InMemChunkRepo) SetDigest(_ context.Context, id t.ChunkID, digest *digest.Digest) error {
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
	chunk, ok := r.chunks[id]
	if !ok {
		return m.Chunk{}, m.ErrChunkNotFound
	}
	return *chunk.Clone(), nil
}

func (r *InMemChunkRepo) IncReplication(_ context.Context, id t.ChunkID) error {
	chunk, ok := r.chunks[id]	
	if !ok {
		return m.ErrChunkNotFound 
	}
	chunk.ReplicaCount++
	return nil
}

