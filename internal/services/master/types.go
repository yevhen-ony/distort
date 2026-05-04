package master

import (
	t "dos/internal/common/types"
	"maps"
	"time"
)

type Object struct {
	ID     t.ObjectID
	Chunks map[t.ChunkKey]t.ChunkID
}

func (o *Object) Clone() *Object {
	if o == nil {
		return nil
	}

	chunks := make(map[t.ChunkKey]t.ChunkID, len(o.Chunks))
	maps.Copy(chunks, o.Chunks)

	return &Object{
		ID:     o.ID,
		Chunks: chunks,
	}
}

type Chunk struct {
	t.ChunkMeta
	ReplicaCount int	
}

func (c *Chunk) Clone() *Chunk {
	if c == nil {
		return nil
	}

	return &Chunk{
		ChunkMeta: *c.ChunkMeta.Clone(),
		ReplicaCount: c.ReplicaCount,
	}
}

type Node struct {
	t.NodeRef
	Stats      t.NodeStats
	LastSeenAt time.Time
}
