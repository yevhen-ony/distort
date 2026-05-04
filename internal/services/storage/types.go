package storage 

import (
	"io"

	t "dos/internal/common/types"
)

type ChunkState struct{
	t.ChunkMeta
	Reported bool
}

type ChunkCatalog map[t.ChunkID]*ChunkState


type Chunk struct {
	Meta t.ChunkMeta
	Data io.ReadCloser
}

