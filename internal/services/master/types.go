package master

import (
	"dos/internal/common/digest"
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
	ID     t.ChunkID
	Digest *digest.Digest
}

func (c *Chunk) Clone() *Chunk {
	if c == nil {
		return nil
	}

	clone := Chunk{ID: c.ID}
	if c.Digest != nil {
		clone.Digest = c.Digest.Clone()
	}
	return &clone
}

type Node struct {
	t.NodeRef
	Stats      t.NodeStats
	LastSeenAt time.Time
}
