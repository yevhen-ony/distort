package storage 

import (
	t "dos/internal/common/types"
)

type ChunkState struct{
	t.ChunkMeta
	Reported bool
}

type ChunkCatalog map[t.ChunkID]*ChunkState

