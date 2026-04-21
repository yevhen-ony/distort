package chunkserver

import (
	"io"
	"time"
)

type ChunkID string
type ChunkCatalog map[ChunkID]ChunkMeta

type ChunkDigest struct {
	Size     int64
	Checksum string
}

type ChunkMeta struct {
	ChunkDigest
	ModifiedAt time.Time
}

type ChunkInfo struct {
	ID ChunkID
	ChunkDigest
}

type Chunk struct {
	ID ChunkID
	Meta *ChunkMeta
	Reader io.ReadCloser
}

