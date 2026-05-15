package master

import (
	t "dos/internal/common/types"
	"maps"
	"time"
)

type Object struct {
	ID          t.ObjectID
	Chunks      map[t.ChunkKey]t.ChunkID
	Replication int
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
		Replication: o.Replication,
	}
}

type Chunk struct {
	t.ChunkMeta
	ReplicaCount int
	ObjectID     t.ObjectID
	ChunkKey     t.ChunkKey
}

func (c *Chunk) Clone() *Chunk {
	if c == nil {
		return nil
	}

	return &Chunk{
		ChunkMeta:    *c.ChunkMeta.Clone(),
		ReplicaCount: c.ReplicaCount,
		ObjectID:     c.ObjectID,
	}
}

type Node struct {
	t.NodeRef
	Stats      t.NodeStats
	LastSeenAt time.Time
}
