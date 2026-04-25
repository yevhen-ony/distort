package master

import (
	"dos/internal/common/digest"
	"maps"
)

type ObjectID string
type ChunkID string
type NodeID string
type ChunkKey string 

type Object struct {
	ID     ObjectID
	Chunks map[ChunkKey]ChunkID
}

func (o *Object) Clone() *Object {
	if o == nil {
		return nil
	}

	chunks := make(map[ChunkKey]ChunkID, len(o.Chunks))
	maps.Copy(chunks, o.Chunks)

	return &Object{
		ID:     o.ID,
		Chunks: chunks,
	}
}

type ObjectAccess struct {
	ObjectID  ObjectID
	TotalSize int64
	Chunks    []ChunkPlacement
}

type Chunk struct {
	ID     ChunkID
	Digest *digest.Digest
}

type ChunkPlacement struct {
	ChunkID ChunkID
	ChunkKey ChunkKey
	Nodes   []NodeAccess
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
	ID     NodeID
	Report NodeReport
}

type NodeReport struct {
	Addr       string
	FreeBytes  int64
	UsedBytes  int64
	ChunkCount int
}

type NodeAccess struct {
	NodeID NodeID
	Addr   string
}
