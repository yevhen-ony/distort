package storage 

import (
	"sync"

	t "dos/internal/common/types"
)

type ChunkState uint8

const (
	ChunkStateStaged ChunkState = iota
	ChunkStateActive
)

type ChunkRecord struct {
	Meta  t.ChunkMeta
	State ChunkState
}

func NewChunkRecord(meta t.ChunkMeta) *ChunkRecord {
	return &ChunkRecord{
		Meta:  meta,
		State: ChunkStateStaged,
	}
}

type ChunkCatalog map[t.ChunkID]*ChunkRecord

type ChunkCatalogState struct {
	Catalog    ChunkCatalog
	Mu         sync.RWMutex
	TotalBytes int64
}

func NewChunkCatalogState() *ChunkCatalogState {
	return &ChunkCatalogState{
		Catalog: make(ChunkCatalog),
	}
}

