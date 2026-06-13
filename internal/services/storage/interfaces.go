package storage

import t "dos/internal/common/types"

type CatalogSource interface {
	List() ([]t.ChunkID, error)
	GetMeta(t.ChunkID) (t.ChunkMeta, error)
}
