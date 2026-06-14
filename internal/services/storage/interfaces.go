package storage

import (
	"context"
	t "dos/internal/common/types"
)

type CatalogSource interface {
	List() ([]t.ChunkID, error)
	GetMeta(t.ChunkID) (t.ChunkMeta, error)
}

type UploadSession interface {
	Write([]byte) (int, error)
	Commit(context.Context) error
	Close() error
}
