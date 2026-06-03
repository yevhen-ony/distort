package storage

import (
	t "dos/internal/common/types"
)

type ChunkState uint8

const (
	ChunkStateStaged ChunkState = iota
	ChunkStateActive
)

func (cs ChunkState) String() string {
	switch cs {
	case ChunkStateStaged:
		return "staged"
	case ChunkStateActive:
		return "active"
	default:
		return ""
	}
}

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

func (r *ChunkRecord) Clone() *ChunkRecord {
	return &ChunkRecord{
		Meta: *r.Meta.Clone(),
		State: r.State,
	}
}

type ChunkCatalog map[t.ChunkID]*ChunkRecord
