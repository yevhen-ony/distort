package storage 

import (
	"io"

	t "dos/internal/common/types"
)

type ChunkCatalog map[t.ChunkID]t.ChunkMeta

type Chunk struct {
	t.ChunkMeta
	Data io.ReadCloser
}

