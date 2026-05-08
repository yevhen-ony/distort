package storage 

import (
	t "dos/internal/common/types"
)

type ChunkState uint8 

const (
	ChunkStateStaged ChunkState = iota;
	ChunkStateActive
)

type ChunkRecord struct{
	Meta t.ChunkMeta
	State ChunkState
}

type ChunkCatalog map[t.ChunkID]*ChunkRecord


