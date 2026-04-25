package storage 

import (
	"io"
	"time"

	"dos/internal/common/digest"
	t "dos/internal/common/types"
)

type ChunkWriter interface {
	io.WriteCloser
	Digest() digest.Digest
	Commit(t.ChunkID) (time.Time, error)
}

type ChunkStorage interface {
	Get(chunkID t.ChunkID) (io.ReadCloser, error)
	GetMeta(chunkID t.ChunkID) (*ChunkMeta, error)
	NewWriter() (ChunkWriter, error)
	GetAllIDs() ([]t.ChunkID, error)
}
