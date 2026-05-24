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
		ID:          o.ID,
		Chunks:      chunks,
		Replication: o.Replication,
	}
}

type Chunk struct {
	Meta          t.ChunkMeta
	ReplicaCount  int
	Slot          t.ObjectSlot
	LastTouchedAt time.Time
}

func (c *Chunk) Clone() *Chunk {
	if c == nil {
		return nil
	}

	return &Chunk{
		Meta:          *c.Meta.Clone(),
		ReplicaCount:  c.ReplicaCount,
		Slot:          c.Slot,
		LastTouchedAt: c.LastTouchedAt,
	}
}

type Node struct {
	t.NodeRef
	Stats      t.NodeStats
	LastSeenAt time.Time
}
