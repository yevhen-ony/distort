package storage 

import (
	"io"
	"time"

	"dos/internal/common/digest"
	t "dos/internal/common/types"
)

type ChunkCatalog map[t.ChunkID]ChunkMeta

type ChunkDigest struct {
	Size     int64
	Checksum string
}

type ChunkMeta struct {
	Digest digest.Digest
	ModifiedAt time.Time
}

type ChunkInfo struct {
	ID t.ChunkID
	Digest digest.Digest
	
}

type Chunk struct {
	ID t.ChunkID
	Meta *ChunkMeta
	Reader io.ReadCloser
}

