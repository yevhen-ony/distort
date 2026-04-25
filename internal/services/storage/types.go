package storage 

import (
	"io"
	"time"

	"dos/internal/common/digest"
)

type ChunkID string
type ChunkCatalog map[ChunkID]ChunkMeta

type ChunkDigest struct {
	Size     int64
	Checksum string
}

type ChunkMeta struct {
	Digest digest.Digest
	ModifiedAt time.Time
}

type ChunkInfo struct {
	ID ChunkID
	Digest digest.Digest
	
}

type Chunk struct {
	ID ChunkID
	Meta *ChunkMeta
	Reader io.ReadCloser
}

