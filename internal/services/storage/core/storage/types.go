package storage

import (
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
